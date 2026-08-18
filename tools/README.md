# Tools

Both scripts are standard-library only and need no device.

## `mock_device.py`
Walks a device through its whole lifecycle against a running control plane:
enrol → receive token → heartbeat → notice an `asr_config_version` change →
fetch the config → report it applied. Use it to validate the flow before any
firmware exists.

## `mock_openai_client.py`
Drives an OpenAI-compatible transcription endpoint.

```bash
python3 tools/mock_openai_client.py --base-url http://ASR_HOST:8621 \
        --api-key KEY --concurrency 4 --burst --interval 0 --count 3
```

Two flags matter when measuring capacity:

* `--burst` synchronises the workers on a barrier so requests genuinely
  overlap. Without it a fast decode never coincides with the next request and
  any concurrency figure is fiction.
* `peak_accepted_inflight` in the STATS line is the real channel count — the
  maximum overlap among requests the server actually *served*, swept from
  their time spans. `peak_inflight` is only what the client had in flight,
  which counts requests that are about to be rejected.
