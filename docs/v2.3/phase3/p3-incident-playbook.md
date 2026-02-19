# Archivist Abuse/Privacy Incident Playbook (G4/G5)

This playbook keeps operators honest with respect to G4 (operational readiness) and G5 (regression/abuse gate) when handling potential abuse or privacy violations in Archivist. It stays planning-only until the first actual incident triage is recorded; any references to communications or escalation are intended to match the documented gate owners and SOC/Privacy partners.

## Gate mapping
- **G4**: Ensures that the operator runbook’s alerts map to documented abuse/privacy response actions.
- **G5**: Confirms that regression expectations (restore/backfill/durability) remain satisfied while mitigating the abuse/privacy event.

## Detection
1. **Monitoring alerts:** the storage growth, quota exhaustion, prune lag, or replica-target alarms may flag abuse or privacy-induced degradation. Pair them with the abuse metrics that feed into `archivist_reputation_event_total`.
2. **Telemetry spikes:** watch for surges in `archivist_backfill_requests` with repeated keyword patterns, repeated `quota_refusal` events, or `archive_manifest_leak` signals.
3. **User reports:** funnel direct abuse/privacy reports into the SOC queue; treat any report referencing private Spaces as high priority and begin the play immediately.
4. **Automated audits:** scheduled scans (e.g., `scripts/archivist-privacy-audit.sh`) that detect unexpected egress or manifest exposures should trigger the same incident flow.

## Containment
1. **Isolate the scope:** mark affected Archivist nodes read-only with `./scripts/archivist-control.sh --drain`, pause Pods, and block ingest if needed.
2. **Snapshot evidence:** capture `journalctl`, `podman logs archivist-*-replica`, and any relevant metrics dashboards before altering state.
3. **Limit the blast radius:** disable the offending client/key, revoke tokens, and, if necessary, remove the problematic volume from the service mesh until the investigation completes.
4. **Engage the privacy guardrails:** confirm `privacy.py` (or policy gating) is enforcing classified spaces, adjust gating if new information indicates a misconfiguration.

## Escalation
1. **Notify gate owners:** update the `G4` and `G5` gate owners via the RACI template (`docs/templates/roadmap-signoff-raci.md`) and note the EV evidence IDs that will capture the incident.
2. **Involve SOC/Privacy:** route to the security operations center and the privacy lead, attaching logs and metric snapshots from containment.
3. **Executive summary:** if the incident affects customer data or compliance, brief the governance board with the initial impact assessment and the planned mitigation steps.

## Communications
1. **Internal update:** log the incident in the TL;DR channel, include EV references (e.g., `EV-v23-G4-incident-01`), and document decisions, actions, and next steps.
2. **Operator artifacts:** update `docs/v2.3/phase5/p5-evidence-index.md` with the alert timeline and remediation notes so downstream reviewers can trace the response.
3. **External messaging:** coordinate with communications/legal for any necessary notify notices. Public statements should emphasize that the incident was contained according to the documented gate playbooks, referencing the relevant gating status if it affects release timelines.

## Post-incident review
1. **Lessons learned:** convene a blameless review that covers the trigger, what worked, what failed (monitoring thresholds, runbook clarity), and assign actions.
2. **Gate adjustments:** if the incident exposed missing coverage in G4/G5, update the runbook or regression matrix and note the change in the gate log (e.g., `docs/v2.3/phase3/p3-regression-report.md`).
3. **Evidence wrap-up:** append the final timeline, impact assessment, and any communications artifacts to the EV log (`EV-v23-G4-###`/`EV-v23-G5-###`) and mark the G5 regression row accordingly.

## Planning honesty note
This playbook is complete as a plan; no live incident run is claimed yet. When the next incident occurs, link to these sections in the incident report and update the relevant evidence IDs with actual outputs and communications notes.
