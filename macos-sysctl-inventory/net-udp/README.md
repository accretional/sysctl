# net — UDP

UDP configuration under `net.inet.udp.*`. ~50 keys, all `int32`.

## Key metrics

| Key | Example | Description |
|-----|---------|-------------|
| `net.inet.udp.pcbcount` | | Active UDP PCB count |
| `net.inet.udp.maxdgram` | 9216 | Max datagram size |
| `net.inet.udp.recvspace` | 196724 | Default receive buffer |
| `net.inet.udp.checksum` | 1 | Checksum enabled |
| `net.inet.udp.blackhole` | 0 | Blackhole mode |
| `net.inet.udp.log_in_vain` | 0 | Log to closed ports |

## Notes

- UDP is stateless, so `pcbcount` shows bound sockets, not "connections."
- `net.inet.udp.maxdgram` = 9216 bytes default (enough for jumbo frames).
