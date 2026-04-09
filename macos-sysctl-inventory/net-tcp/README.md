# net — TCP

TCP configuration and performance metrics under `net.inet.tcp.*`. ~180 keys, all `int32` unless noted.

## Connection tuning

| Key | Example | Description |
|-----|---------|-------------|
| `net.inet.tcp.mssdflt` | 512 | Default MSS |
| `net.inet.tcp.sendspace` | 131072 | Default send buffer |
| `net.inet.tcp.recvspace` | 131072 | Default receive buffer |
| `net.inet.tcp.keepidle` | 7200000 | Keepalive idle time (ms) |
| `net.inet.tcp.keepintvl` | 75000 | Keepalive interval (ms) |
| `net.inet.tcp.keepcnt` | 8 | Keepalive probe count |
| `net.inet.tcp.keepinit` | 75000 | Initial keepalive timeout |
| `net.inet.tcp.v6mssdflt` | 1024 | Default IPv6 MSS |
| `net.inet.tcp.maxseg_unacked` | 8 | Max unacked segments |
| `net.inet.tcp.rfc1323` | 1 | Window scaling enabled |
| `net.inet.tcp.rfc3390` | 1 | Increasing initial window |

## Connection state

| Key | Description |
|-----|-------------|
| `net.inet.tcp.pcbcount` | Active TCP PCB count (connections) |
| `net.inet.tcp.tw_pcbcount` | TIME_WAIT PCB count |
| `net.inet.tcp.cubic_sockets` | Sockets using CUBIC |

## SACK

| Key | Description |
|-----|-------------|
| `net.inet.tcp.sack` | SACK enabled |
| `net.inet.tcp.sack_maxholes` | Max SACK holes per connection |
| `net.inet.tcp.sack_globalmaxholes` | Max SACK holes global |
| `net.inet.tcp.sack_globalholes` | Current global SACK holes |

## Congestion control

| Key | Description |
|-----|-------------|
| `net.inet.tcp.cc_algo` | Default CC algorithm |
| `net.inet.tcp.ecn_initiate_out` | ECN for outgoing connections |
| `net.inet.tcp.ecn_negotiate_in` | ECN for incoming connections |
| `net.inet.tcp.ecn_timeout` | ECN negotiation timeout |
| `net.inet.tcp.cubic_tcp_friendliness` | CUBIC TCP friendliness |
| `net.inet.tcp.cubic_fast_convergence` | CUBIC fast convergence |
| `net.inet.tcp.cubic_use_newreno` | Use NewReno with CUBIC |

## Delayed ACK

| Key | Description |
|-----|-------------|
| `net.inet.tcp.delayed_ack` | Delayed ACK enabled |
| `net.inet.tcp.ack_prioritize` | Prioritize ACKs |
| `net.inet.tcp.maxdelack` | Max delayed ACK count |
| `net.inet.tcp.delack_timeout` | Delayed ACK timeout |

## Security

| Key | Description |
|-----|-------------|
| `net.inet.tcp.always_keepalive` | Force keepalive on all connections |
| `net.inet.tcp.drop_synfin` | Drop SYN+FIN packets |
| `net.inet.tcp.blackhole` | Blackhole mode (drop RST) |
| `net.inet.tcp.log_in_vain` | Log incoming connections to closed ports |

## Retransmission

| Key | Description |
|-----|-------------|
| `net.inet.tcp.rexmt_thresh` | Fast retransmit threshold |
| `net.inet.tcp.rexmt_slop` | Retransmit timer slop |
| `net.inet.tcp.rxt_seg_drop` | int64 — Segments dropped on retransmit |
| `net.inet.tcp.rxt_findpeergwroute` | Find peer gateway on retransmit |

## Miscellaneous

| Key | Description |
|-----|-------------|
| `net.inet.tcp.fastopen` | TCP Fast Open enabled |
| `net.inet.tcp.fastopen_backlog` | TFO backlog |
| `net.inet.tcp.now_init` | TCP clock init value |
| `net.inet.tcp.microuptime_init` | Microtime init |
| `net.inet.tcp.minmss` | Minimum MSS |
| `net.inet.tcp.do_tcpdrain` | Enable TCP drain |
| `net.inet.tcp.socket_unlocked_on_output` | Unlocked socket output |

## Notes

- `net.inet.tcp.pcbcount` is the primary "how many TCP connections" metric.
- `sendspace`/`recvspace` = default socket buffer sizes (128 KB default on macOS).
- SACK holes > 0 indicates packet loss / reordering.
- TCP Fast Open and ECN are both configurable here.
