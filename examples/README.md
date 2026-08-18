# Examples

## Transcribe a file through your ASR server

```bash
curl -sS -X POST "http://ASR_HOST:8621/v1/audio/transcriptions" \
  -H "Authorization: Bearer $OVS_API_KEY" \
  -F "file=@sample.wav" -F "model=whisper-1" -F "language=auto"
# {"text":"..."}
```

Audio should be 16 kHz mono PCM16. Keep a clip under 10.5 s where English may
appear — see `docs/operations/asr-constraints.md` for why that number, and
what happens silently if you ignore it.

## Enrol a device and deliver its ASR config

```bash
# 1. the device enrols once, with the shared enrolment key, and gets a token
curl -sS -X POST "http://CLOUD_HOST:3008/api/v1/devices/register" \
  -H "X-Enrollment-Key: $ENROLLMENT_KEY" \
  -d '{"mac_address":"aa:bb:cc:dd:ee:01","name":"floor-1"}'
# {"result":{"token":"...","asr_config_version":0,...}}

# 2. it heartbeats with that token; the response carries asr_config_version
curl -sS -X POST "http://CLOUD_HOST:3008/api/v1/devices/register" \
  -H "Authorization: Bearer $DEVICE_TOKEN" \
  -d '{"mac_address":"aa:bb:cc:dd:ee:01"}'

# 3. when the version changes, it fetches the config and applies it
curl -sS "http://CLOUD_HOST:3008/api/v1/devices/me/asr-config" \
  -H "Authorization: Bearer $DEVICE_TOKEN"
# {"result":{"code":"openai_whisper","config_json":{"base_url":"...","api_key":"..."}}}
```

`tools/mock_device.py` runs this loop end to end without any hardware.

## Measure how many channels a server really serves

```bash
python3 tools/mock_openai_client.py --base-url http://ASR_HOST:8621 \
        --api-key "$OVS_API_KEY" --concurrency 4 --burst --interval 0 --count 3
# STATS: ... peak_inflight=4 peak_accepted_inflight=1
```

`peak_accepted_inflight` is the answer. A server that admits four requests and
serves them one at a time reports 4 and 1 — which is the expected shape, not a
fault. See `docs/operations/capacity.md`.
