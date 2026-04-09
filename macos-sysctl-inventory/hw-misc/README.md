# hw — Miscellaneous Hardware

Remaining `hw.*` keys not covered by CPU topology, memory, cache, perflevel, or ARM features.

| Key | Bytes | Type | Example | Description |
|-----|-------|------|---------|-------------|
| `hw.targettype` | varies | string | "J773" | Apple hardware target type code |
| `hw.serialdebugmode` | 4 | int32 | 0 | Serial debug mode enabled |
| `hw.use_kernelmanagerd` | 4 | int32 | 1 | Kernel manager daemon in use |
| `hw.use_recovery_securityd` | 4 | int32 | 0 | Recovery security daemon |
| `hw.osenvironment` | 32 | raw | | OS environment data |
| `hw.ephemeral_storage` | 4 | int32 | 0 | Ephemeral storage available |
| `hw.features.allows_security_research` | 4 | int32 | 0 | Security research device |
