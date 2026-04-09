# net — IP

IP layer configuration under `net.inet.ip.*`. ~60 keys, all `int32`.

## Key metrics

| Key | Example | Description |
|-----|---------|-------------|
| `net.inet.ip.forwarding` | 0 | IP forwarding enabled |
| `net.inet.ip.ttl` | 64 | Default TTL |
| `net.inet.ip.maxfragpackets` | 2048 | Max fragmented packets |
| `net.inet.ip.maxfrags` | 4096 | Max fragments |
| `net.inet.ip.maxchainsent` | | Max chain sent |
| `net.inet.ip.accept_sourceroute` | 0 | Accept source-routed packets |
| `net.inet.ip.redirect` | 1 | ICMP redirect enabled |
| `net.inet.ip.gifttl` | 30 | GIF tunnel TTL |
| `net.inet.ip.subnets_are_local` | 0 | Treat subnets as local |
| `net.inet.ip.random_id` | 1 | Random IP ID |

## Notes

- `net.inet.ip.forwarding` = 0 is default (not a router).
- Fragment limits protect against fragmentation attacks.
