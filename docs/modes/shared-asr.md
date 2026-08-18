# Mode B — one ASR server, many devices

Devices capture audio and send it to a shared OpenVoiceStream instance on the
LAN. Devices stay cheap; the engine runs once on hardware that suits it.

**Suited to:** a store or floor with several devices and one Jetson or RK box.

## Deploy

On the ASR server:
```bash
docker compose -f deploy/shared-asr/docker-compose.yml up -d
```

On each device:
```bash
docker compose -f deploy/shared-asr/agent.yml up -d
```

Then register the ASR server in the console and assign devices to it. The
endpoint reaches each device through its heartbeat — no per-device edit.

## Sizing

A server decodes **one clip at a time**; that is a property of the backend, not
a configuration mistake. Raising `ASR_MAX_SLOTS` does not add parallelism, it
adds a queue: callers wait instead of receiving a 429.

Measured on an Orin Nano with SenseVoice, 3 s of audio, admission at 4:

| Concurrency | 429s | Latency |
|---|---|---|
| 1 | 0 | ~54 ms |
| 4 (before) | 15 of 20 | — |
| 4 (after) | **0** | 60 / 112 / 162 / 212 ms |

Each step of that ladder is one serialized decode. Queueing is what buys
device count: at one slot with no queue the server must sit ~99% idle to keep
rejections rare; with a queue it can run far hotter. See
`docs/operations/capacity.md`.
