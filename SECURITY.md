# Security

## Reporting

Please report vulnerabilities privately to the maintainers rather than opening
a public issue.

## Before you deploy this

Every credential and address in this repository is a placeholder. Nothing here
is a working deployment until you supply your own.

* **`jwt_key`** in the config files is `CHANGE_ME`. It signs console sessions;
  a deployment that leaves it at any published value can have its tokens
  forged.
* **`enrollment_key`** gates device registration. Anything holding it can
  enrol a device.
* **`crypto_master_key`** encrypts ASR-server API keys at rest.
* **`device_auth_enforce` ships as `false`** so a fleet without credential-
  capable firmware keeps working. While it is false the device-facing
  endpoints accept unauthenticated calls. Set it to `true` once your firmware
  rollout is complete — otherwise this layer is decorative.

## Notes

The config loader reads YAML only; it has no environment-variable injection.
If you need credentials from a secret manager, add that first rather than
committing real values — see `.env.example` for the full list of what a
deployment must supply.
