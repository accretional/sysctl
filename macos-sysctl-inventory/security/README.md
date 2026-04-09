# security — Code Signing & MAC

Security framework settings. 6 keys.

| Key | Bytes | Type | Description |
|-----|-------|------|-------------|
| `security.codesigning.monitor` | 4 | int32 | Code signing monitor active |
| `security.mac.sandbox.sentinel` | varies | string | Sandbox sentinel path |
| `security.mac.amfi.hsp_enable` | 4 | int32 | AMFI HSP enabled |
| `security.mac.amfi.force_policy` | 4 | int32 | AMFI force policy |
| `security.mac.proc_enforce` | 4 | int32 | Process enforcement |
| `security.mac.vnode_enforce` | 4 | int32 | Vnode enforcement |

## Notes

- AMFI = Apple Mobile File Integrity (code signing enforcement).
- MAC = Mandatory Access Control framework (TrustedBSD-derived).
- These are mostly read-only indicators of security policy state.
