# Third-party notices

This project is MIT licensed. It depends on, but does not vendor, the
following components. Each keeps its own licence and its own release cadence.

## Runtime dependencies (not included in this repository)

| Component | Role | Licence |
|---|---|---|
| [OpenVoiceStream](https://github.com/suharvest/openvoicestream) | ASR/TTS serving layer. Every deployment mode talks to it; it is never vendored here. | MIT |
| [voxedge](https://github.com/suharvest/voxedge) | Edge inference backends (TensorRT / RKNN / sherpa-onnx) used by OpenVoiceStream. | MIT |

## Models and inference libraries

Reached through OpenVoiceStream rather than directly, but listed because a
deployment ships them and their terms apply to that deployment:

| Component | Licence / terms |
|---|---|
| [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx) | Apache-2.0 |
| SenseVoice (FunAudioLLM) | See the model card; the weights carry their own terms, separate from this repository's MIT licence |
| Silero VAD | MIT |

Check the licence of any model you deploy. An MIT codebase does not make the
weights it loads MIT.
