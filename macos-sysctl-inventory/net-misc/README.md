# net — Miscellaneous Network

Remaining `net.*` keys (~300) covering IPv6, link layer, routing, local sockets, and Apple-specific networking.

## Categories

- **`net.inet6.*`** — IPv6 configuration (addresses, NDP, autoconf)
- **`net.link.*`** — Link layer (Ethernet, generic, loopback, bond)
- **`net.local.*`** — Unix domain sockets
- **`net.route.*`** — Routing table stats
- **`net.key.*`** — IPsec key management
- **`net.ipsec.*`** — IPsec policy
- **`net.necp.*`** — Network Extension Control Policy
- **`net.netagent.*`** — Network agent framework
- **`net.cfil.*`** — Content filter (Screen Time, parental controls)
- **`net.classq.*`** — Packet scheduling (QoS)
- **`net.pktsched.*`** — Packet scheduler config
- **`net.qos.*`** — Quality of service
- **`net.restricted_port.*`** — Restricted port ranges
- **`net.smb.*`** — SMB client/server
- **`net.stats.*`** — Network statistics framework
- **`net.tracker.*`** — App tracking transparency network
- **`net.utun.*`** — User tunnel interfaces (VPN)
- **`net.vsock.*`** — Virtual socket (VM communication)
- **`net.soflow.*`** — Socket flow tracking
- **`net.mpklog.*`** — Multipath kernel log

## Key performance metrics

| Key | Type | Description |
|-----|------|-------------|
| `net.local.pcbcount` | int32 | Unix domain socket count |
| `net.classq.target_qdelay` | int32 | Target queue delay |
| `net.classq.update_interval` | int32 | Queue update interval |

## Notes

- Most keys are configuration knobs, not live counters.
- For per-connection stats, `nettop` or `networksetup` provide more detail.
- Apple's Skywalk/channel framework keys are under `kern.skywalk.*`, not `net.*`.
