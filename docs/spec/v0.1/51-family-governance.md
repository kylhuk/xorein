# 51 — Family: Governance (`/aether/governance/0.2.0`)

This document specifies the Governance family, which provides role-based access
control (RBAC) management for Xorein servers: role assignment, revocation,
custom role creation, and role-table synchronisation.

## 1. Overview

The Governance family manages the server role hierarchy. Roles determine what
moderation and management actions a member may perform. The role table is the
authority consulted by the Moderation family (`50-family-moderation.md`) before
executing any action.

**Protocol ID:** `/aether/governance/0.2.0`

**Required capability:** `cap.rbac`

**Roles that use this family:** client (owner, admin, or moderator performing
assignments; any member receiving a sync); relay (forwarding signed deliveries
only, never processing role state).

**Design principle:** Role assignments form a signed append-log. Every
assignment or revocation MUST be delivered as a signed `GovernanceMetadata`
`Delivery` to all current server members so that the role table converges
deterministically across a distributed peer set.

## 2. Capability requirements

| Capability | Meaning |
|------------|---------|
| `cap.rbac` | Required for all Governance family operations |

The `cap.moderation` capability is NOT required to receive
`governance.sync`; any member with `cap.rbac` MUST accept sync deliveries.

Operations that mutate the role table (`governance.assign_role`,
`governance.revoke_role`, `governance.create_role`, `governance.delete_role`)
additionally require the actor to hold a sufficiently high role (§3.1).

## 3. Operations

| Operation | Required caps | Minimum actor role | Request payload | Response payload | Description |
|-----------|--------------|-------------------|-----------------|------------------|-------------|
| `governance.assign_role` | `cap.rbac` | admin (for moderator/member); owner (for admin) | `RoleAssignRequest` | `RoleAssignResponse` | Assign a role to a member |
| `governance.revoke_role` | `cap.rbac` | admin (for moderator/member); owner (for admin) | `RoleRevokeRequest` | `RoleAssignResponse` | Revoke a role, returning the member to `member` |
| `governance.create_role` | `cap.rbac` | admin | `RoleCreateRequest` | `RoleCreateResponse` | Create a custom role with a permissions bitfield |
| `governance.delete_role` | `cap.rbac` | admin | `RoleDeleteRequest` | `RoleAssignResponse` | Delete a custom role; members holding it revert to `member` |
| `governance.sync` | `cap.rbac` | any | `RoleSyncRequest` | `RoleSyncResponse` | Push the current role table to a peer (used on join or reconnect) |

### 3.1 Authorization rules

The Governance family uses the role hierarchy defined in `pkg/v02/rbac/rbac.go`:

```
owner > admin > moderator > member
```

No `READ_ONLY` role is defined in v0.1; `member` is the lowest ranked role.

The receiver MUST enforce these constraints before applying any mutation:

1. **Role exists:** The actor's `RoleState` for this server MUST be present in
   the receiver's local role table.
2. **Actor outranks target:** `rbac.CanActOnTarget(actor.role, target.role)`
   MUST return `true`. An actor CANNOT assign or revoke a role equal to or
   higher than their own.
3. **Assignment ceiling:**
   - Assigning `admin`: MUST be `owner` only.
   - Assigning `moderator` or `member`: MUST be `admin` or above.
   - Assigning `owner` is NOT permitted via this operation (ownership transfer
     is a manifest-level operation outside the scope of v0.1).
4. **Custom role creation / deletion:** MUST be `admin` or above.

Failure modes return `GOVERNANCE_UNAUTHORIZED` or `GOVERNANCE_FORBIDDEN_TARGET`
(§7).

### 3.2 `governance.assign_role`

Assigns a role to a member:

1. Receiver verifies authorization (§3.1).
2. Receiver upserts the `RoleState` for `(server_id, target_peer_id)`:
   `role = assigned_role`, `version += 1`, `updated_at_unix = now`.
3. Receiver resolves any concurrent views using `rbac.ResolveRoleState()` (see
   §6 for convergence semantics).
4. Receiver constructs and hybrid-signs a `GovernanceMetadata` delivery.
5. Receiver broadcasts the signed delivery to all server members.

### 3.3 `governance.revoke_role`

Revokes a role, returning the member to the base `member` role:

1. Receiver verifies authorization (§3.1).
2. Receiver sets `RoleState.role = member`, increments `version`, updates
   `updated_at_unix`.
3. Receiver broadcasts the signed `GovernanceMetadata` delivery to all members.

The `owner` role MUST NOT be revocable via `governance.revoke_role`. Ownership
transfer is out of scope for v0.1.

### 3.4 `governance.create_role`

Creates a custom named role with a 64-bit permissions bitfield:

1. Receiver verifies authorization (§3.1); MUST be `admin` or above.
2. Receiver stores the new custom role definition:
   `(server_id, role_name, permissions_bitfield)`.
3. Receiver broadcasts the signed `GovernanceMetadata` delivery.

Custom roles are additive extensions; they do not replace the base hierarchy.
The base roles (`owner`, `admin`, `moderator`, `member`) MUST NOT be overridden
by custom roles.

### 3.5 `governance.delete_role`

Deletes a custom role:

1. Receiver verifies authorization (§3.1); MUST be `admin` or above.
2. All members currently holding the deleted custom role MUST have their
   `RoleState` updated to `member`.
3. Receiver broadcasts the signed `GovernanceMetadata` delivery.

Base roles (`owner`, `admin`, `moderator`, `member`) MUST NOT be deleted.

### 3.6 `governance.sync`

Pushes the full role table to a peer. This operation is initiated by:
- The server owner or admin when a new member joins.
- Any member that reconnects after an absence.

Sync procedure:

1. Initiator sends `RoleSyncRequest` containing its local role table version
   (`GovernanceMetadata.policy_version`).
2. Responder compares its `policy_version` with the initiator's.
3. If the responder's version is higher, it sends all `RoleState` records for
   the server.
4. If the initiator's version is higher or equal, the responder sends an empty
   table and `RoleSyncResponse.up_to_date = true`.
5. The initiator applies received `RoleState` records using
   `rbac.ResolveRoleState()` for any conflicts.

`governance.sync` MUST be performed after a new member joins (before the member
can send any messages) and after a reconnect when the `policy_version` in the
member's local state is behind the server's.

## 4. Wire format details

All payloads are JSON-encoded in `PeerStreamRequest.payload` /
`PeerStreamResponse.payload`.

### 4.1 `RoleAssignRequest`

```
{
  "actor_peer_id":    string,    // peer ID of the admin/owner performing the assignment
  "target_peer_id":  string,    // peer ID of the member being assigned the role
  "server_id":       string,
  "role":            string,    // "owner" | "admin" | "moderator" | "member"
  "policy_version":  uint64,    // current role table version known to the actor
  "signature":       string     // base64url hybrid sig over canonical JSON (excl. this field)
}
```

### 4.2 `RoleRevokeRequest`

Identical structure to `RoleAssignRequest`. The `role` field specifies the role
being revoked (the member reverts to `member` after revocation).

### 4.3 `RoleCreateRequest`

```
{
  "actor_peer_id":         string,
  "server_id":             string,
  "role_name":             string,    // unique within the server; max 64 bytes; must match [a-z0-9_-]+
  "permissions_bitfield":  uint64,    // see §4.5 for bit definitions
  "policy_version":        uint64,
  "signature":             string
}
```

### 4.4 `RoleDeleteRequest`

```
{
  "actor_peer_id":  string,
  "server_id":      string,
  "role_name":      string,
  "policy_version": uint64,
  "signature":      string
}
```

### 4.5 Permissions bitfield (64-bit)

Custom roles use a 64-bit permissions bitfield. The following bits are defined:

| Bit | Name | Description |
|-----|------|-------------|
| 0 | `SEND_MESSAGES` | Can send messages in text channels |
| 1 | `MANAGE_CHANNELS` | Can create, edit, and delete channels |
| 2 | `MANAGE_ROLES` | Can assign/revoke roles below own rank |
| 3 | `KICK_MEMBERS` | Can kick members below own rank |
| 4 | `BAN_MEMBERS` | Can ban members below own rank |
| 5 | `MANAGE_MESSAGES` | Can delete messages from other members |
| 6 | `MUTE_MEMBERS` | Can apply timed mutes to members below own rank |
| 7 | `VIEW_CHANNEL` | Can view channel history and receive messages |
| 8 | `USE_VOICE` | Can join and speak in voice channels |
| 9 | `MANAGE_SERVER` | Can update server manifest settings |
| 10–63 | (reserved) | MUST be zero in v0.1; reserved for future protocol extensions |

Reserved bits MUST be zero when creating a custom role. Receivers MUST ignore
unknown set bits for forward compatibility.

Base roles have implicitly defined permission sets and MUST NOT be expressed as
a bitfield in `RoleCreateRequest`.

### 4.6 Common response: `RoleAssignResponse`

```
{
  "accepted":       bool,
  "policy_version": uint64,    // updated policy_version after the mutation
  "error":          string     // human-readable; absent on success
}
```

### 4.7 `RoleCreateResponse`

```
{
  "accepted":       bool,
  "role_name":      string,
  "policy_version": uint64,
  "error":          string
}
```

### 4.8 `RoleSyncRequest`

```
{
  "requester_peer_id": string,
  "server_id":         string,
  "known_policy_version": uint64
}
```

### 4.9 `RoleSyncResponse`

```
{
  "up_to_date":     bool,
  "policy_version": uint64,
  "roles": [
    {
      "identity_id":      string,
      "role":             string,
      "version":          uint64,
      "updated_at_unix":  uint64
    }
  ]
}
```

### 4.10 `RoleState` (proto message)

Defined in `proto/aether.proto` as `message RoleState`:

| Field | Type | Notes |
|-------|------|-------|
| `identity_id` | string | Peer ID of the member |
| `role` | `V02Role` | `OWNER`, `ADMIN`, `MODERATOR`, or `MEMBER` |
| `version` | uint64 | Monotonically increasing assignment counter |
| `updated_at_unix` | uint64 | Unix seconds of the last role change |

### 4.11 `GovernanceMetadata` (proto message)

Defined in `proto/aether.proto` as `message GovernanceMetadata`:

| Field | Type | Notes |
|-------|------|-------|
| `protocol_id` | string | `/aether/governance/0.2.0` |
| `protocol_major` | uint32 | `0` |
| `protocol_minor` | uint32 | `2` |
| `required_flags` | repeated string | `["cap.rbac"]` |
| `advertised_flags` | repeated string | Capabilities of the broadcasting node |
| `security_mode` | `SecurityMode` | Active server security mode |
| `mode_epoch_id` | string | Current epoch ID for the security mode |
| `policy_version` | uint64 | Role table version after this update |

`GovernanceMetadata` is broadcast as the `data` field (protobuf-encoded) of a
`Delivery` with `kind = "governance"` and `scope_type = "server"`. The
`Delivery` MUST carry a valid hybrid signature from the acting node.

### 4.12 `V02Role` enum values

Defined in `proto/aether.proto`:

| Value | Integer | Notes |
|-------|---------|-------|
| `V02_ROLE_UNSPECIFIED` | 0 | Invalid; MUST NOT appear in wire messages |
| `V02_ROLE_MEMBER` | 1 | Base role; default for new members |
| `V02_ROLE_MODERATOR` | 2 | Can kick, mute, delete messages |
| `V02_ROLE_ADMIN` | 3 | Can ban, create/delete roles, assign moderators |
| `V02_ROLE_OWNER` | 4 | All permissions; can assign/revoke admins |

## 5. Security mode binding

Governance deliveries MUST be encrypted under the server's active security mode.
The `GovernanceMetadata` bytes are placed in `Delivery.data` (protobuf-encoded),
which is then encrypted with the server scope key.

Receivers MUST verify the hybrid signature on the `Delivery` before applying
any role state change. A `GovernanceMetadata` delivery with a missing or
invalid signature MUST be rejected and MUST NOT update local role state.

The `GovernanceMetadata.policy_version` monotonically increases. Receivers MUST
reject `GovernanceMetadata` with a `policy_version` lower than or equal to the
locally stored `policy_version` for the same server (replay protection).

## 6. State persistence

| State bucket | Key | Value | Description |
|-------------|-----|-------|-------------|
| `servers` | `server_id` | `ServerRecord` (JSON) | Members list; role state embedded per member |
| `governance` | `(server_id, peer_id)` | `RoleState` (JSON) | Per-member role with convergence metadata |

Go types from `pkg/v02/rbac/rbac.go`:

- `RoleState`: `Role rbac.Role`, `Version uint64`, `UpdatedAt time.Time`

Convergence: when two `RoleState` records exist for the same `(server_id,
peer_id)`, `rbac.ResolveRoleState(a, b)` deterministically selects the
canonical state:

1. Higher `Version` wins.
2. If versions are equal, the most recent `UpdatedAt` wins.
3. If timestamps are also equal, the lower-ranked role wins (conservative
   resolution).

`policy_version` for a server is the maximum `RoleState.Version` across all
members and is stored in `GovernanceMetadata.policy_version`. It increments on
every successful assignment, revocation, or role table modification.

Custom role definitions are stored in a `custom_roles` bucket:
`(server_id, role_name)` → `(permissions_bitfield uint64)`.

## 7. Error codes

| Code | Trigger |
|------|---------|
| `GOVERNANCE_UNAUTHORIZED` | Actor's role does not permit the requested mutation |
| `GOVERNANCE_FORBIDDEN_TARGET` | Actor cannot act on the target (target role >= actor role) |
| `GOVERNANCE_STALE_VERSION` | `policy_version` in the request is behind the server's current version |
| `GOVERNANCE_ROLE_NOT_FOUND` | Target role name does not exist (for delete/assign custom role) |
| `GOVERNANCE_ROLE_CONFLICT` | `governance.create_role` — role name already exists |
| `GOVERNANCE_BASE_ROLE_PROTECTED` | Attempt to delete or override a base role |
| `GOVERNANCE_INVALID_BITFIELD` | Reserved bits set in `permissions_bitfield` |
| `GOVERNANCE_MISSING_SIGNATURE` | `GovernanceMetadata` delivery signature absent |
| `GOVERNANCE_INVALID_SIGNATURE` | Hybrid signature verification failed |
| `GOVERNANCE_OWNER_IMMUTABLE` | Attempt to revoke or reassign the owner role |

## 8. Conformance

Implementations claiming Governance family conformance MUST pass the following
KATs in `pkg/spectest/governance/`:

| KAT file | Covers |
|----------|--------|
| `assign_moderator_kat.json` | Admin assigns moderator role; verify `RoleState` update and broadcast |
| `assign_admin_kat.json` | Owner assigns admin; verify `policy_version` increment |
| `revoke_role_kat.json` | Admin revokes moderator; member reverts to `member` |
| `create_role_kat.json` | Admin creates custom role with bitfield `0x1A`; verify storage |
| `delete_role_kat.json` | Admin deletes custom role; affected members revert to `member` |
| `sync_kat.json` | New member receives full role table via `governance.sync` |
| `convergence_kat.json` | Two concurrent role assignments for the same member; `ResolveRoleState` picks deterministic winner |
| `unauthorized_kat.json` | Moderator attempts to assign admin; expect `GOVERNANCE_UNAUTHORIZED` |
| `forbidden_target_kat.json` | Admin attempts to revoke another admin; expect `GOVERNANCE_FORBIDDEN_TARGET` |
| `stale_version_kat.json` | Request with `policy_version` behind current; expect `GOVERNANCE_STALE_VERSION` |
| `replay_kat.json` | `GovernanceMetadata` with duplicate `policy_version`; MUST be rejected |

Each KAT MUST include the full `PeerStreamRequest` and `PeerStreamResponse`
bytes in the format defined by `90-conformance-harness.md`.
