# sysctl
Kernel performance metrics for macOS + Linux (currently only targeting macOS)

## Objective

We need exportable/accessible telemetry for our macOS machines interacting with our agent mesh, and a standard format for accessing, reporting, and representing them. Apple exposes kernel metrics through the sysctl API [https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man3/sysctl.3.html](https://developer.apple.com/documentation/kernel/sys#3571071). Note that there are a very large number of APIs in https://developer.apple.com/documentation/kernel/sys and these also include proc/ but unlike Linux, these are not the units through which performance gets reported via procfs. Sysctl is Apple's canonical way of getting the metrics. So we are going to try to use that and build a grpc service (mostly in Go, because all our other stuff is using that right now) around it.

Later we'll think about whether a kernel extension for automated reporting (https://developer.apple.com/documentation/apple-silicon/installing-a-custom-kernel-extension) or some other approach is better. For now we'll use sysctlbyname (https://developer.apple.com/documentation/kernel/1387446-sysctlbyname).

IMPORTANT: we are probably going to have to be architecture-aware about this super low level stuff anyway. So we want to use Go's asm facilities to call the underlying syscalls directly, first on Apple silicon because that's what we're using. This way we won't need to fight dependencies and libraries to get to the underlying information from the kernel. The first task is enumerating all of these and making sure we can get some kind of e2e reporting of grpc/proto->go server for accretional/sysctl -> sysctl asm on apple silicon -> back to the client working. Then we can scale this out horizontally.

First let's just write a package called like macos-asm-sysctl (or formatted however golang wants) and  get the basic core working there. Then we'll build up and out around that.
