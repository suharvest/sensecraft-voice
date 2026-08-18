# Architecture

Three deployment modes, one control plane. What changes between them is
**where the ASR engine runs** — not the code.

```
                        ┌──────────────────────────────┐
                        │  cloud/service  (:3008)      │
   mode A ──────────┐   │  devices · auth · recordings │
   device + local   ├──▶│  ASR-server registry         │◀── cloud/console
   OVS              │   │  keywords · analysis         │      (:3000)
                    │   └──────────────────────────────┘
   mode B ──────────┤              ▲
   device ──▶ LAN   │              │ transcripts, heartbeats
   OVS server       │              │
                    │   ┌──────────┴───────────┐
   mode C ──────────┘   │  MySQL · object store │
   device --BLE--> phone└───────────────────────┘
   --HTTPS--> cloud
```

## Why not split the repository by mode

The device agent and the control plane are identical across A and B. The only
difference is the value of one delivered setting — the ASR endpoint, which the
cloud hands to the device via `GET /api/v1/devices/me/asr-config`. Mode A
points it at localhost, mode B at a LAN server.

Splitting by mode would fork the control plane into two copies that drift.
Splitting by layer keeps one copy, and the mode lives in `deploy/`.

## Where ASR happens

The engine is [OpenVoiceStream](https://github.com/suharvest/openvoicestream),
kept as a separate project and never vendored here: it has its own release
cadence and is useful on its own. Every mode talks to it over the same
OpenAI-compatible HTTP surface.

## Identity and liveness

| | modes A / B | mode C |
|---|---|---|
| Identity | device token, issued at enrolment against a shared enrolment key; only the SHA-256 hash is stored | `api_key` issued from the console; one per device |
| Liveness | registration doubles as a 60 s heartbeat; `online` is computed at query time from `last_seen_at`, so there is no status column to keep consistent and no sweeper job | no heartbeat — the last upload time *is* the liveness signal |

A portable device has no route by which to heartbeat: it is only reachable
while a phone is nearby with the app open. Reporting it as "offline" the rest
of the time would say nothing about the device, so it is not modelled that way.
