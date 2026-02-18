# Podman Scenarios

Run the deterministic podman scenario script and review logs:

```
./scripts/v20-podman-scenarios.sh
cat artifacts/generated/v20-podman-scenarios/result-manifest.json
```

The script performs the following concrete commands in order:
1. `podman pod create --name v20-operator`
2. `podman container run --rm --pod v20-operator registry.access.redhat.com/ubi8/ubi:latest echo ready`
3. `podman pod ps --filter name=v20-operator`

Failures or missing manifest entries block gate G5.
