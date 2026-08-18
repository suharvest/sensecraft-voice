# SenseCraft Voice

Speech capture, transcription and analysis for edge devices, with a cloud
control plane that manages the fleet.

The ASR engine itself is **not** in this repository — it lives in
[OpenVoiceStream](https://github.com/suharvest/openvoicestream), and every
deployment mode below talks to it. What differs between the modes is only
**where that engine runs**.

---

## Three deployment modes

| | **A. On-box ASR** | **B. Shared ASR** | **C. Portable device** |
|---|---|---|---|
| Hardware | reRouter CM4 + reSpeaker mic array | Thin device | Portable device + phone |
| Audio path | device → local ASR | device → LAN ASR server | device --BLE--> phone --HTTPS--> cloud |
| ASR runs on | the device itself | one shared LAN server | a server the cloud manages |
| Devices per ASR server | 1:1 | 1:N | 1:N |
| Identity | device token | device token | api_key issued from the console |
| Heartbeat | every 60 s | every 60 s | none; last upload time is the liveness signal |

All three share the same control plane: device registry, ownership, recording
storage, transcripts, keyword matching, analysis and dashboards. **Only the
audio entry point differs.**

That is why this repository is laid out by layer rather than by mode — the
mode shows up in `deploy/`, not in the code.

---

## Layout

```
cloud/
  service/      Go backend (:3008) — devices, auth, recordings, ASR-server management
  console/      React admin UI
device/
  agent/        Go agent — capture, push to ASR, report results
    web/        the device's own config page (see below)
deploy/
  on-box/       mode A
  shared-asr/   mode B
  cloud/        control plane (MySQL, object storage)
docs/
examples/
tools/          mock device, load-test client
```

### The device config page is not optional

`device/agent/web/` is served by the agent itself on port 8090. On a freshly
flashed device it is the **only** way to point the device at a cloud address —
there is no connection yet through which anything else could configure it. It
is shipped in the same binary and image as the agent, and the two must stay in
step.

It is a different thing from `cloud/console/`, which is the fleet-wide admin
UI. One runs on the device, the other in the cloud.

---

## Quick start

Pick the mode first; each has its own compose file under `deploy/`.

```bash
cp .env.example .env      # fill in addresses and credentials
docker compose -f deploy/cloud/docker-compose.yml up -d       # control plane
docker compose -f deploy/shared-asr/docker-compose.yml up -d  # mode B, for example
```

Every address and credential in this repository is a placeholder
(`*.example.com`, `CHANGE_ME`). Nothing here is a working deployment until you
supply your own.

---

## Hardware

| Board | Modes | Notes |
|---|---|---|
| reRouter CM4 | A | Pi4-class, no NPU. ASR on CPU |
| Raspberry Pi 4 / 5 | A, B | Admission pinned to one session — a single CPU decode already uses most of the cores |
| RK3576 / RK3588 | B | NPU; the RK backend segments long audio internally |
| Jetson Orin Nano / NX | B | TensorRT; measured ~54 ms for a 3 s clip |

---

## A constraint worth reading before you build on this

SenseVoice takes a **fixed-length input** — about 20.4 s — and silently drops
whatever does not fit. No error, no log line, HTTP 200 with a truncated
transcript. Measured on hardware: a 26.57 s clip and a 21.21 s clip returned
byte-identical text.

OpenVoiceStream now segments long clips automatically, so this is handled for
you on that path. But if you send audio to a fixed-shape engine by any other
route, chunk it yourself: **≤10.5 s where English may appear, ≤15 s for
Chinese-only**. Those numbers are measured, not guessed — English starts
dropping words between 10.65 s and 12.15 s while Chinese stays clean to
15.45 s.

---

## Licence

MIT — see [LICENSE](LICENSE). Third-party components and model terms are in
[NOTICE.md](NOTICE.md); an MIT codebase does not make the model weights it
loads MIT.
