# P2-Private Space Anti-Enumeration

Private Space history endpoints derive retrieval keys from membership secrets so unauthorized callers never learn whether history exists. Failure responses always return the generic `HISTORY_RETRIEVAL_FAILURE` even when segments/head/manifests are missing.
