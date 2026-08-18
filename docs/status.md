# Status

What works, what does not, and what has been measured versus assumed. Written
for someone picking this up without the history.

Measurements come from an Orin Nano running SenseVoice on TensorRT unless
stated otherwise. Where a number has not been measured, it says so.

---

## Working

| Capability | Where | How it was checked |
|---|---|---|
| Automatic segmentation of long audio | OpenVoiceStream, `asr_segmenter` | Against a backend that emulates the fixed-shape truncation: a 30 s clip loses 10 of 30 words unsegmented and recovers all 30 segmented |
| Decode moved off the event loop | OpenVoiceStream, `api_execution` | A probe against `/readyz` during a decode; reverting the change fails the test |
| OpenAI-compatible transcription | OpenVoiceStream | Contract tests upstream |
| `asr_max_slots` reaching the backend | `cloud/service` config injection | On device: rejections went 15 → 0 at four slots |
| Admission queue (not parallelism) | voxedge SenseVoice backend | Latency ladder 60 / 112 / 162 / 212 ms — one serialized decode per step |
| Buffer reuse, valid-frame transfer | voxedge ≥ 0.0.11a0 | 3 s clip end to end: 68 ms → 53.7 ms |
| Engine build parameterised, cache keyed on it | OpenVoiceStream `model_downloader` | Changing any knob forces a rebuild; the ONNX is identified by content digest |
| Device enrolment and tokens | `cloud/service` middleware | A mock device walks the full lifecycle against a live service |
| Online status, computed at query time | `cloud/service` | Filtered in SQL; no status column, no sweeper |
| ASR-server registry and health probing | `cloud/service` | Busy and broken are separated; six unit tests |
| Config delivery over the heartbeat | `cloud/service` | Version bump → fetch → apply → report, end to end |
| Console JWT | `cloud/service` | Matches the frontend's 401 handling |
| Credential encryption at rest | `cloud/service`, `pkg/util/crypto` | Unit tests |

---

## Not built

**Portable-device transcription intake.** The design is written and the
surrounding pieces exist, but no code. Needed: the transcription endpoint,
api_key issuance and validation, a forwarding client with backoff on 429,
persistence tied to the device, and console management for the keys.

**Console screens for fleet management.** Device online state, the ASR-server
page, assignment and applied-state display. The backend APIs are done and
exercised; only the UI is missing.

**Streaming.** The architecture selects transports from advertised
capabilities, so the extension point exists. Note the offline SenseVoice
backends do not advertise streaming — a streaming deployment needs a different
backend, and the streaming models measured so far were noticeably worse on
English and Cantonese than the offline path.

**Failover.** An unhealthy ASR server is reported, not routed around.

---

## Built, but inert until you do something

These will look implemented and quietly do nothing if the prerequisite is
missed.

**`device_auth_enforce` ships as `false`.** Device-facing endpoints accept
unauthenticated calls while it is false — that default exists so a fleet
without credential-capable firmware keeps working. Set it to `true` once your
rollout completes.

**Placeholder credentials.** `enrollment_key`, `crypto_master_key` and
`jwt_key` are all `CHANGE_ME`. `jwt_key` in particular signs console sessions.

**The database migration is manual.** `AutoMigrate` creates missing tables and
skips existing ones — it will not add a column. Run
`cloud/service/docs/migration_*.sql` before deploying a build that expects the
new device columns, or queries against them fail outright.

**Health checks in your own compose.** The files under `deploy/` probe
`/livez`. If you write your own, do the same — `/readyz` includes spare
capacity, so a one-slot device reads as 503 during every decode.

---

## Measured versus assumed

Worth knowing before you quote any of this.

| Claim | Basis |
|---|---|
| 3 s clip ≈ 54 ms; `W(D) ≈ 0.058 + 0.0032·D` | Measured, Orin Nano |
| One server decodes one clip at a time | Measured; a context pool bought 1.13× for +302 MB per slot and was rejected |
| Safe chunk 10.5 s with English, 15 s Chinese-only | Measured by bisection. An earlier 12 s figure was published and withdrawn — it came from a multilingual fixture, and the engine takes one language prompt per utterance, so that measured language switching rather than length |
| Hard truncation at ~20.4 s | Measured: a 26.57 s and a 21.21 s clip returned byte-identical text |
| Device counts per admission depth | **Modelled**, not measured. Queueing theory over the measured per-request cost, assuming rejected requests vanish — a retrying client adds load, so treat them as upper bounds |
| Service time for 10.5 s chunks | **Extrapolated** from the 3 s measurement. The capacity table rests on it; measure it on your hardware before committing to a device count |

---

## Not yet verified anywhere

- **SenseVoice on RK hardware.** Every figure here is from Jetson. The RK
  backend segments long audio internally, so its behaviour differs by
  construction — it needs its own pass.
- **Segmentation on real hardware.** Covered by 45 tests against a fake
  backend; the VAD cut quality on real speech has not been checked on a
  device.
- **On-box mode on CM4-class hardware.** CPU figures come from a Pi 5; a CM4
  is roughly 2–3× slower per core. The margin looks adequate on paper and has
  never been run.

---

## Things that fail silently

Each of these has bitten during development. They produce no error.

1. **A fixed-shape engine truncates over-long audio** and returns HTTP 200.
   Handled inside OpenVoiceStream; any other path to the engine must chunk.
2. **A version pinned in two places.** The image installs
   `voxedge==${VOXEDGE_VERSION}` and then `pip install -r requirements.txt`.
   Bumping only the first is undone by the second, and the build stays green.
   Verify with `pip show` inside the built image.
3. **A capability lookup that falls back on error.** A mis-wired concurrency
   setting resolves to the conservative default, so the knob appears connected
   and does nothing. The fallback now logs at warning level.
4. **`include_router` no longer flattens sub-routes** in current FastAPI, so
   code that scans `app.routes` for a path silently stops finding it.

When you touch any of these, verify the effect — in the built image, on the
device — not the diff.
