# F13 Proto Delta (v1.3 planning)

## Status
Planning artifact only.

## Additive wire plan
- Add `SpaceState` message with immutable `space_id` and lifecycle enum.
- Add `ChannelState` message for text-channel baseline metadata.
- Add `MessageDeliveryState` reason enum extensions for transient/policy/auth failure classes.
- Add optional read-marker metadata fields to channel message sync payloads.

## Compatibility constraints
- No field renumbering or type mutation.
- Removed fields must be reserved by number and name.
- Unknown fields are ignored by older clients; defaults preserve behavior.

## Validation requirements
- `buf lint`
- `buf breaking`
- Additive changelog reviewed by protocol steward.

## Planned vs implemented
- No v1.3 proto generation is performed in v1.2 for this spec artifact.
