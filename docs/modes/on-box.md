# Mode A — ASR on the device

The agent and an OpenVoiceStream instance run on the same board; the agent's
ASR endpoint points at localhost. One device, one engine.

**Suited to:** a reRouter CM4 or a Raspberry Pi with a reSpeaker mic array —
somewhere audio must not leave the box, or there is no LAN server to share.

**Cost:** every device needs enough CPU to decode. On Pi-class hardware that
is a real constraint; see `docs/operations/capacity.md`.

## Deploy

```bash
cp .env.example .env      # set OVS_IMAGE, OVS_API_KEYS, cloud address
docker compose -f deploy/on-box/docker-compose.yml up -d
```

Then open `http://DEVICE_IP:8090/` — the agent's own config page — and point
the device at the cloud. On a freshly flashed device that page is the only way
in; nothing else is configured yet.

## Two things that bite

**Health checks must probe `/livez`, not `/readyz`.** Readiness includes "the
limiter has spare capacity". On a one-slot board that means every in-flight
decode reports 503, and an orchestrator watching `/readyz` restarts the
container mid-request. The compose file already does this correctly.

**Keep admission at one session.** `cpu.sherpa_asr` declares support for four,
but a single CPU decode already uses most of a Pi's cores. The shipped
profiles pin it to one; raise it only with measurements from your own board.
