# Contributing

## Build

```bash
cd cloud/service && go build ./... && go vet ./...
cd cloud/console && npm ci && npm run build
cd device/agent  && CGO_ENABLED=1 go build ./...   # cgo required: audio bindings
```

The device agent needs `CGO_ENABLED=1` — its capture path binds to a C audio
library, so a pure-Go build will not link.

## Database migrations

`AutoMigrate` creates missing **tables** and skips any table that already
exists — it will not add a column. Adding a column to a live table therefore
needs a script under `cloud/service/docs/migration_*.sql` **and** the matching
struct change. Shipping only the struct change silently does nothing.

## Measurements in documentation

Performance and capacity claims here are measured, and the docs say where the
number came from. If you change one, replace it with your own measurement and
say what hardware produced it. A plausible estimate that turns out wrong is
more expensive than an absent one, because the next person builds on it.

## Things that fail silently

This codebase has a few places where a mistake produces no error:

* a fixed-shape ASR engine truncates over-long audio and returns HTTP 200
* a version pinned in two files (Dockerfile ARG and `requirements.txt`) where
  changing only one is undone by the other
* a capability lookup that falls back to a conservative default when it raises,
  so a mis-wired concurrency setting looks configured and does nothing

When you touch any of them, verify the *effect* — inside the built image, on
the device — not the diff.
