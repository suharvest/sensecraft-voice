# What a fixed-shape ASR engine will not tell you

SenseVoice on TensorRT/RKNN takes a **fixed-length input**: 344 LFR frames,
about 20.4 s of audio. The preprocessor truncates anything longer. There is no
error, no warning, and the response is a normal HTTP 200 carrying a shortened
transcript.

Measured on hardware: a 26.57 s clip and a 21.21 s clip returned
**byte-identical** text. Everything past 20.4 s simply did not exist.

## Chunk lengths, measured

Quality degrades before the hard cut, and how early depends on the language:

| | Safe chunk | Set by |
|---|---|---|
| Anywhere English may appear | **10.5 s** | degradation onset, bisected: clean at 10.65 s, dropping words at 12.15 s |
| Chinese-only | **15 s** | measured clean at 15.45 s; no degradation observed below the hard cut |
| Absolute ceiling, any language | **20.4 s** | silent truncation |

Every one of those is the last **verified-good** measurement, not an
interpolation. An earlier reading of 12 s was published and then withdrawn: it
came from a fixture that concatenated Chinese, English and Cantonese, and
SenseVoice takes one language prompt per utterance — so that test measured
language switching, not length. Bisecting single-language audio put the
English onset between 10.65 s and 12.15 s, meaning 12 s sat inside the
degraded region.

**Mixed languages in one clip cannot be fixed by chunking on length.** Chunk at
language turns, or decode per language.

## What handles this for you

OpenVoiceStream segments long clips automatically — VAD-aligned cuts, per-piece
decode, rejoined text, and it logs what it did. Set
`OVS_ASR_AUTO_SEGMENT=0` to restore single-pass behaviour.

If you send audio to a fixed-shape engine by any other path, chunk it
yourself. Do not rely on an upload-size limit to do it: a 25 MB threshold is
roughly 13 minutes of 16 kHz mono, about 65× over.

## Health checks

Probe **`/livez`**, never `/readyz`. Readiness includes "the limiter has spare
capacity", so on a one-slot device every in-flight decode reads as 503 and an
orchestrator will restart the container mid-request. `/livez` is independent of
backend, GPU and capacity — which is what a restart policy should key on.

The same distinction matters when monitoring a fleet: a 503 from `/readyz`
whose `reasons` is exactly `["sessions_full"]` means *busy*, not *broken*.
Counting it as a failure marks every working server down under load.
