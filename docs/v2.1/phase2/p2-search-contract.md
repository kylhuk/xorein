# P2 Search Contract

- **Indexing scope**: plaintext timelines are indexed locally by `pkg/v21/search` with explicit channel, sender, and time-filter inputs. No remote keyword search or server-assisted ranking is permitted.
- **Coverage reporting**: each channel exposes a deterministic coverage window (machine `Status` + human `Label`) derived from the earliest/latest timestamps in the index. Empty channels report `COVERAGE_EMPTY` with a descriptive label.
- **Query limits & timeouts**: search queries obey a fixed limit (default 64 results) and abort with the sentinel error `SEARCH_QUERY_TIMEOUT` whenever the limit is exceeded or the caller cancels the request. No partial result sets are returned when the guard triggers.
