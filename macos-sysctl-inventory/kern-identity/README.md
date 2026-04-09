# kern — System Identity

OS and system identification strings.

| Key | Type | Example | Description |
|-----|------|---------|-------------|
| `kern.ostype` | string | "Darwin" | OS type |
| `kern.osrelease` | string | "24.6.0" | Darwin kernel release |
| `kern.osversion` | string | "24G90" | macOS build version |
| `kern.osproductversion` | string | "15.6.1" | macOS product version |
| `kern.osproductversioncompat` | string | "15.6" | Compatibility version |
| `kern.osreleasetype` | string | "User" | Release type |
| `kern.osrevision` | int32 | 199506 | OS revision number |
| `kern.version` | string | "Darwin Kernel Version 24.6.0:..." | Full kernel version string |
| `kern.hostname` | string | "Freds-Mac-mini.local" | System hostname |
| `kern.nisdomainname` | string | "" | NIS domain name (usually empty) |
| `kern.uuid` | string | | System UUID |
| `kern.bootuuid` | string | | Boot session UUID |
| `kern.bootsessionuuid` | string | | Boot session UUID (alt) |
| `kern.kernelcacheuuid` | string | | Kernel cache UUID |
| `kern.filesetuuid` | string | | Fileset UUID |
| `kern.apfsprebootuuid` | string | | APFS preboot volume UUID |
| `kern.bootargs` | string | "" | Kernel boot arguments |
| `kern.iossupportversion` | string | "18.5" | iOS compatibility version |

## Notes

- `kern.osproductversion` gives the user-facing macOS version (15.6.1).
- `kern.osrelease` gives the Darwin kernel version (24.6.0).
- `kern.osversion` is the build identifier (24G90).
- `kern.iossupportversion` indicates Catalyst/iOS compatibility level.
