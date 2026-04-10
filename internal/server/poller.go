//go:build darwin

package server

import (
	"log"
	"sync"
	"time"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
)

// polledMetric tracks a single metric's polling schedule.
type polledMetric struct {
	name          string
	ttlNanos      int64 // TTL in nanoseconds
	nextGatherNs  int64 // unix nanos of next scheduled gather
}

// metricStore holds cached metric values behind a read-write lock.
type metricStore struct {
	mu     sync.RWMutex
	values map[string]*pb.Metric
}

func newMetricStore() *metricStore {
	return &metricStore{values: make(map[string]*pb.Metric)}
}

func (s *metricStore) get(name string) (*pb.Metric, bool) {
	s.mu.RLock()
	m, ok := s.values[name]
	s.mu.RUnlock()
	return m, ok
}

func (s *metricStore) put(name string, m *pb.Metric) {
	s.mu.Lock()
	s.values[name] = m
	s.mu.Unlock()
}

// putBatch writes multiple metrics under a single lock acquisition.
func (s *metricStore) putBatch(batch map[string]*pb.Metric) {
	s.mu.Lock()
	for name, m := range batch {
		s.values[name] = m
	}
	s.mu.Unlock()
}

// poller runs a single goroutine that periodically gathers metrics
// based on their recommended access patterns.
type poller struct {
	server             *SysctlServer
	store              *metricStore
	frequencyIncrement time.Duration
	polledMetrics      []polledMetric
	stopCh             chan struct{}
	doneCh             chan struct{}
}

// newPoller creates a poller that ticks at frequencyIncrement.
// It reads the full registry to determine which metrics are STATIC, POLLED, or CONSTRAINED.
// STATIC metrics are read immediately into the store.
// POLLED and CONSTRAINED metrics are scheduled for periodic gathering at their TTL.
func newPoller(srv *SysctlServer, frequencyIncrement time.Duration) *poller {
	p := &poller{
		server:             srv,
		store:              newMetricStore(),
		frequencyIncrement: frequencyIncrement,
		stopCh:             make(chan struct{}),
		doneCh:             make(chan struct{}),
	}

	if srv.fullRegistry == nil {
		return p
	}

	now := time.Now().UnixNano()
	staticCount := 0
	polledCount := 0
	constrainedCount := 0

	for _, km := range srv.fullRegistry.Metrics {
		rec := km.RecommendedAccessPattern
		if rec == nil {
			continue
		}

		switch rec.Pattern {
		case pb.AccessPattern_STATIC:
			// Read once now, store forever.
			m := srv.readMetricLive(km.Name)
			p.store.put(km.Name, m)
			staticCount++

		case pb.AccessPattern_POLLED, pb.AccessPattern_CONSTRAINED:
			ttlNanos := int64(10 * time.Second) // default 10s
			if rec.Ttl != nil {
				ttlNanos = rec.Ttl.Seconds*int64(time.Second) + int64(rec.Ttl.Nanos)
			}
			p.polledMetrics = append(p.polledMetrics, polledMetric{
				name:         km.Name,
				ttlNanos:     ttlNanos,
				nextGatherNs: now, // gather immediately on first tick
			})
			if rec.Pattern == pb.AccessPattern_CONSTRAINED {
				constrainedCount++
			} else {
				polledCount++
			}

			// Also do an initial read so the store has values before the first tick.
			m := srv.readMetricLive(km.Name)
			p.store.put(km.Name, m)
		}
	}

	log.Printf("poller: %d static, %d polled, %d constrained (tick every %v)", staticCount, polledCount, constrainedCount, frequencyIncrement)
	return p
}

// start begins the polling goroutine.
func (p *poller) start() {
	go p.loop()
}

// stop signals the polling goroutine to exit and waits for it.
func (p *poller) stop() {
	close(p.stopCh)
	<-p.doneCh
}

func (p *poller) loop() {
	defer close(p.doneCh)

	ticker := time.NewTicker(p.frequencyIncrement)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.tick()
		}
	}
}

func (p *poller) tick() {
	now := time.Now().UnixNano()
	batch := make(map[string]*pb.Metric)

	for i := range p.polledMetrics {
		pm := &p.polledMetrics[i]
		if pm.nextGatherNs <= now {
			m := p.server.readMetricLive(pm.name)
			batch[pm.name] = m
			pm.nextGatherNs = now + pm.ttlNanos
		}
	}

	if len(batch) > 0 {
		p.store.putBatch(batch)
	}
}
