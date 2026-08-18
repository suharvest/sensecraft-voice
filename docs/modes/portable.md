# Mode C — portable device via a phone

The device pairs with a phone over BLE. The phone app has an ASR endpoint
setting; point it at the control plane, and the backend takes over
transcription, storage and display.

**The app needs no protocol change.** The backend sits where an ASR vendor
would, speaking the standard OpenAI transcription API, so only the address and
key change.

## What the device is, to the rest of the system

A row in the same `devices` table. `api_key` identifies it, so there is no
enrolment handshake and no heartbeat: the last upload time is the liveness
signal. Everything downstream — storage, transcripts, keyword matching,
analysis, dashboards — is shared with modes A and B.

## Two constraints that follow from BLE

**Store and forward, not streaming.** 16 kHz PCM is about 32 KB/s; BLE
throughput in practice is a fraction of that. The device records, then
transfers.

**A LAN ASR server is unreachable.** The phone is usually on cellular, so
mode B's `192.168.x.x` endpoint does not apply. Transcription happens in the
cloud.
