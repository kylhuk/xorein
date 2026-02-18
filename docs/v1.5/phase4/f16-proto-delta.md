# v16 Proto Delta

- Add RBAC capability field to `SessionMetadata` without renumbering existing fields.
- Introduce `AclEntry` message with optional `role`, `allow`, `deny`, and `channel` selectors.
- Ensure all additions are optional to preserve wire-compatibility for v15 clients.
