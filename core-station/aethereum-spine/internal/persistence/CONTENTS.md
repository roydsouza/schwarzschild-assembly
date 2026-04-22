# Aethereum-Spine: Persistence & State

This sector manages the station's durable state, including database connections and mission metadata storage.

## Sections
- [migrations/](./migrations/CONTENTS.md) — SQL migration scripts for the station's audit database.

## Key Files
- [db.go](./db.go) — Database connection and pooling logic.
- [mission_store.go](./mission_store.go) — CRUD operations for mission and spacecraft metadata.
- [CONTENTS.md](./CONTENTS.md) — This atlas file.
