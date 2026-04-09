# hw — ARM Feature Flags

ARM architecture feature detection via `hw.optional.arm.*`. All are 4 bytes / `int32`
(0 = not supported, 1 = supported) unless noted.

## Cryptography & Security

| Key | M4 | Description |
|-----|----|-------------|
| `hw.optional.arm.FEAT_AES` | 1 | AES instructions |
| `hw.optional.arm.FEAT_SHA1` | 1 | SHA-1 instructions |
| `hw.optional.arm.FEAT_SHA256` | 1 | SHA-256 instructions |
| `hw.optional.arm.FEAT_SHA512` | 1 | SHA-512 instructions |
| `hw.optional.arm.FEAT_SHA3` | 1 | SHA-3 instructions |
| `hw.optional.arm.FEAT_PMULL` | 1 | Polynomial multiply (for GCM) |
| `hw.optional.arm.FEAT_CRC32` | 1 | CRC32 instructions |
| `hw.optional.arm.FEAT_PAuth` | 1 | Pointer authentication |
| `hw.optional.arm.FEAT_PAuth2` | 1 | Pointer authentication v2 |
| `hw.optional.arm.FEAT_FPAC` | 1 | Faulting PAC |
| `hw.optional.arm.FEAT_FPACCOMBINE` | 1 | Combined PAC |
| `hw.optional.arm.FEAT_PACIMP` | 1 | PAC implementation defined |
| `hw.optional.arm.FEAT_BTI` | 1 | Branch target identification |

## SIMD & Matrix

| Key | M4 | Description |
|-----|----|-------------|
| `hw.optional.arm.AdvSIMD` | 1 | Advanced SIMD (NEON) |
| `hw.optional.arm.AdvSIMD_HPFPCvt` | 1 | Half-precision FP conversion |
| `hw.optional.arm.FEAT_FP16` | 1 | Half-precision FP |
| `hw.optional.arm.FEAT_BF16` | 1 | BFloat16 |
| `hw.optional.arm.FEAT_EBF16` | 0 | Extended BFloat16 |
| `hw.optional.arm.FEAT_I8MM` | 1 | Int8 matrix multiply |
| `hw.optional.arm.FEAT_DotProd` | 1 | Dot product instructions |
| `hw.optional.arm.FEAT_FHM` | 1 | FP half-precision multiply |
| `hw.optional.arm.FEAT_RDM` | 1 | Rounding double multiply |
| `hw.optional.arm.FEAT_FCMA` | 1 | Complex number multiply |
| `hw.optional.arm.FEAT_JSCVT` | 1 | JavaScript FP conversion |
| `hw.optional.arm.FEAT_FRINTTS` | 1 | FP round to int instructions |

## SME (Scalable Matrix Extension) — Apple M4+

| Key | M4 | Description |
|-----|----|-------------|
| `hw.optional.arm.FEAT_SME` | 1 | SME base |
| `hw.optional.arm.FEAT_SME2` | 1 | SME version 2 |
| `hw.optional.arm.SME_F32F32` | 1 | FP32 outer product |
| `hw.optional.arm.SME_F16F32` | 1 | FP16→FP32 outer product |
| `hw.optional.arm.SME_B16F32` | 1 | BF16→FP32 outer product |
| `hw.optional.arm.SME_I8I32` | 1 | Int8→Int32 outer product |
| `hw.optional.arm.SME_I16I32` | 1 | Int16→Int32 outer product |
| `hw.optional.arm.SME_BI32I32` | 1 | Binary Int32 outer product |
| `hw.optional.arm.FEAT_SME_F64F64` | 1 | FP64 outer product |
| `hw.optional.arm.FEAT_SME_I16I64` | 1 | Int16→Int64 outer product |
| `hw.optional.arm.sme_max_svl_b` | 64 | Max streaming vector length in bytes |

## Atomics & Barriers

| Key | M4 | Description |
|-----|----|-------------|
| `hw.optional.arm.FEAT_LSE` | 1 | Large System Extensions (atomic ops) |
| `hw.optional.arm.FEAT_LSE2` | 1 | LSE v2 (128-bit atomics) |
| `hw.optional.arm.FEAT_SB` | 1 | Speculation barrier |
| `hw.optional.arm.FEAT_DPB` | 1 | Data cache clean to PoP |
| `hw.optional.arm.FEAT_DPB2` | 1 | Data cache clean to PoDP |
| `hw.optional.arm.FEAT_LRCPC` | 1 | Load-acquire RCpc |
| `hw.optional.arm.FEAT_LRCPC2` | 1 | Load-acquire RCpc v2 |

## Speculation & Side-Channel

| Key | M4 | Description |
|-----|----|-------------|
| `hw.optional.arm.FEAT_CSV2` | 1 | Cache speculation variant 2 |
| `hw.optional.arm.FEAT_CSV3` | 1 | Cache speculation variant 3 |
| `hw.optional.arm.FEAT_DIT` | 1 | Data-independent timing |
| `hw.optional.arm.FEAT_SSBS` | 0 | Speculative store bypass safe |
| `hw.optional.arm.FEAT_SPECRES` | 0 | Speculation restriction |
| `hw.optional.arm.FEAT_SPECRES2` | 0 | Speculation restriction v2 |

## Misc

| Key | M4 | Description |
|-----|----|-------------|
| `hw.optional.arm.FEAT_FlagM` | 1 | Flag manipulation |
| `hw.optional.arm.FEAT_FlagM2` | 1 | Flag manipulation v2 |
| `hw.optional.arm.FEAT_WFxT` | 1 | WFE/WFI with timeout |
| `hw.optional.arm.FEAT_RPRES` | 1 | Reciprocal precision |
| `hw.optional.arm.FEAT_AFP` | 1 | Alternate FP mode |
| `hw.optional.arm.FEAT_ECV` | 1 | Enhanced counter virtualization |
| `hw.optional.arm.FEAT_CSSC` | 0 | Common short-sequence compression |
| `hw.optional.arm.FEAT_HBC` | 0 | Hinted conditional branches |
| `hw.optional.arm.FP_SyncExceptions` | 1 | Synchronous FP exceptions |
| `hw.optional.arm.caps` | raw (10 bytes) | Capability bitmask |
| `hw.optional.floatingpoint` | 1 | FP support (always 1) |

## Notes

- Apple M4 notably adds **SME and SME2** — the first Apple chip with scalable matrix extensions.
- `FEAT_SSBS` = 0: Apple handles speculative store bypass differently (hardware mitigation).
- `hw.optional.arm.caps` is a 10-byte raw bitmask, not a simple integer.
