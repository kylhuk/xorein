A brief description of the v2.3 regression harness files.

`scenarios.conf` defines the Podman regression matrix row-by-row as `slug|label|package|pattern|required_output` so the scripts can map scenario IDs to specific packages and expected artifacts. `scripts/v23-regression-scenarios.sh` reads that file, exports the environment variables for each scenario, and drives the Podman jobs listed in `containers/v2.3/*`. Treat `scenarios.conf` as the single source of truth for scenario names, packages, and outputs when updating the regression harness.
