# Asset Coverage Map (G0)

| Asset class | BlobRef-backed in v25 | Notes |
| --- | --- | --- |
| Attachments | ✅ (mandatory) | Upload pipeline stores ciphertext-only blobs; manifest references embedded in message payloads and tied to BlobRef hash/size/mime. |
| User avatars | ✅ (mandatory) | Avatar updates create BlobRef; UI fetches lazy-download with `avatar_manifest` verification. |
| Space icons | ✅ (mandatory) | Mirrors avatar treatment; Space metadata records BlobRef and enforcement of private Space gating. |
| Emojis/stickers | ✅ if already present, otherwise deferred | If emoji/sticker catalog exists in v25, each entry references BlobRef; otherwise, document as v25-ready placeholder and defer UI rollout. |

## Frozen data-class attributes
- All covered assets require manifest hash, size, mime, and optional chunk index metadata.
- Non-covered data (e.g., credential secrets, logs) remain outside v25 scope and logged as No-go in `p0-scope-lock`.

## MIME/size tiers per device
- Desktop tier (250 MiB, image/video/audio/archives) applies to attachments, avatars, icons, emojis.
- Mobile tier (64 MiB, multimedia subset) applies to avatar/avatar updates, mobile-side attachments; UI surfaces default to Wi-Fi-only downloads for files above 64 MiB.
- Any deviation requires gate re-evaluation and evidence entry in `EV-v25-G0-005`.

## Evidence placeholders
- `EV-v25-G0-005`: Asset map review and MIME-tier compliance signature.
