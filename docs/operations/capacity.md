# Capacity

Numbers measured on an Orin Nano with SenseVoice on TensorRT. Treat them as a
starting point for your own board, not as a spec.

## One request

`W(D) ≈ 0.058 + 0.0032 · D` seconds, for D seconds of audio. A 3 s clip takes
about 54 ms end to end over HTTP.

The fixed ~58 ms dominates. Where it goes, for a 3 s clip:

| | ms |
|---|---|
| GPU forward | 37.0 |
| CPU feature extraction (fbank + LFR) | 8.4 |
| CTC decode (argmax) | ~0.6 |
| HTTP / server overhead | ~4.9 |

Because the fixed cost dominates, **longer chunks are far more efficient**:
3 s chunks sustain ~42 audio-seconds per second, 10 s chunks ~98. Use the
longest chunk your accuracy constraints allow — but see the ceiling in
`asr-constraints.md`.

## Concurrency

One server decodes **one clip at a time**. That is the backend: a single
execution context, serialized. It is not a misconfiguration and it is not
worth "fixing" with a context pool — measured, a pool bought 1.13× throughput
for +302 MB per slot and multiplied latency, because the GPU is already at 98%
with one stream and the bottleneck is enqueue, not parallelism.

What raising `ASR_MAX_SLOTS` buys is a **queue**. Requests wait instead of
being rejected, and that is what drives device count:

| Admission depth | Devices at 1% rejection | Mean queue wait | Worst case |
|---|---|---|---|
| 1 | ~25 | 0 ms | 84 ms |
| 4 | ~880 | 40 ms | 336 ms |
| **8** | **~1600** | 129 ms | 672 ms |
| 16 | ~2100 | 349 ms | 1.34 s |

Assumes 3 minutes of speech per device per hour in 10.5 s chunks. **8 is the
recommended default**: still sub-second at worst, and past it the return
shrinks while latency crosses a second.

Throughput barely moves across that table — utilisation does. At one slot with
no queue, the server has to sit ~99% idle to keep rejections rare.

Two caveats: the model assumes rejected requests vanish, but a real client
retries and adds load, so treat these as upper bounds; and the figures cover
ASR only, not any LLM or TTS stage.
