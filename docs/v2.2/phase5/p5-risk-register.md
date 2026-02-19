# P5 Risk Register

This register now describes the as-built risk posture with completed mitigations and traceable evidence per gate.

| Risk | Impact | Likelihood | Mitigation | Evidence | Status |
| --- | --- | --- | --- | --- | --- |
| Undetected replay gaps | History integrity could degrade if missing backfill is not surfaced | Low | Podman replay scenarios plus history backfill/perf suites exercise missing-backfill, coverage-gap detection, and deterministic failure modes. | `EV-v22-G4-001`, `EV-v22-G6-001` | Mitigated |
| Proto delta breaking change | Older clients may drop support if optional fields become mandatory | Low | Buf lint + breaking checks validate the optional metadata and enum; the `DEFAULT` category deprecation warning is documented but non-fatal. | `EV-v22-G7-001`, `EV-v22-G7-002` | Mitigated |
| Evidence bundle misalignment | Release signoff stalls if evidence entries lack trace | Low | Evidence catalog, index, and gate signoffs now reference every mandatory command output and handshake gate. | `EV-v22-G4-001`..`EV-v22-G8-004` | Closed |

Trivy emitted a warning in `make check-full` that `--security-checks` is deprecated; the scan completed cleanly and is recorded in `EV-v22-G5-002`.
