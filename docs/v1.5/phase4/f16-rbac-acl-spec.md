# v16 RBAC + ACL Specification

- Roles: founder, admin, moderator, member, guest, and extensible custom roles.
- ACLs evaluate allow/deny statements per channel, with deny taking precedence when equal weight.
- Legacy permissions map to default role sets; overrides are idempotent when merged with custom channel policies.
