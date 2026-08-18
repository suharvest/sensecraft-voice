# Spec — portable-device transcription intake

**Status: designed, not built.** See `docs/status.md`.

A portable device pairs with a phone over BLE. The phone app already has a
setting for an ASR endpoint. Point it at the control plane and the backend
takes over transcription, storage and display.

---

## The shape of it

The app's ASR configuration has two layers, and only the inner one changes:

| Setting | Meaning |
|---|---|
| The app's own server address | app → its backend |
| **The configured ASR endpoint** | that backend → an ASR provider |

Pointing the **second** at the control plane means the control plane occupies
the position of an ASR provider. The app keeps its protocol; only an address
and a key change. Job submission, polling and their state machine belong to
the app's own layer and are not our concern.

## What to implement

One endpoint, speaking the standard OpenAI transcription API:

```
POST {control-plane}/v1/audio/transcriptions
Authorization: Bearer <api_key>
Content-Type: multipart/form-data

file=<audio.wav>          # 16 kHz mono PCM16
model=<accepted, ignored>
language=zh|en|ja|ko|yue|auto
response_format=json      # -> {"text": "..."}
```

`json` returns `{"text": …}`, `text` returns the bare string, `verbose_json`
adds task, language, duration and segments. Mirror the error semantics of the
ASR server: 401 for a bad key, 503 when no backend is ready, 429 with
`Retry-After` under load.

Confirm which provider name the app will use before starting. If it is
configured as an OpenAI-compatible provider the above is it; some apps offer a
FunASR-shaped provider instead, which posts multipart `file` to a different
path. It is a one-line difference and the only thing needing agreement with
the app side.

## Identity

**Issue one api_key per device from the console.** It authenticates and
identifies in one step, so there is no enrolment handshake and no heartbeat —
`last_seen_at` is refreshed from the last upload.

Everything else about the device is configured server-side: which ASR server
it uses, the language, ownership. The app holds an address and a key, nothing
more.

Two ways to carry the key, depending on how many fields the app exposes:

* **Two fields** — address plus key, sent as `Authorization: Bearer`.
  Standard, and the key stays out of access logs. Prefer this.
* **One field** — key embedded in the path. Works where the app has a single
  input, at the cost of the key appearing in gateway and proxy logs. If you do
  this, make the key revocable and do not log full URLs.

## Forwarding

Take the audio, forward it to the ASR server bound to that device, store the
transcript against the device. Four things the ASR server already handles —
do not reimplement them here:

* **Long audio.** It segments over-long clips itself. Forward the clip whole;
  cutting it first only makes the VAD's job harder.
* **Overload.** It returns 429 with `Retry-After`, and it does not queue past
  its admission depth. **Backoff is required, not optional** — several devices
  sharing one server will hit it.
* **Language.** Pass it through.
* **Health.** The registry already distinguishes a busy server from a broken
  one; skip servers marked down.

## Storage

Write the transcript through the existing recording path so keyword matching,
analysis and the dashboards work unchanged. Store the app's recording id as an
idempotency key — a retry after a network failure should not produce a second
row.

## Out of scope

**Real-time transcription.** 16 kHz PCM is about 32 KB/s and BLE in practice
carries a fraction of that, so the device records and then transfers. Live
transcription would need on-device compression and a different design.
