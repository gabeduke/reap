# reap

Subscribes to an MQTT broker and writes numeric sensor messages to InfluxDB v2.

A message on `<group>/<device>/<name...>` with a numeric payload becomes a point:

| | |
|---|---|
| measurement | `<name...>` joined with `-` |
| tags | `group`, `device`, `sensor` (= measurement) |
| field | `value` (float) |

Non-numeric payloads (status, mode, JSON), topics with fewer than three segments, and any topic with a
segment in `IGNORE` (default `log,metadata,state`) are skipped. Skipping never blocks later messages.

## Config

Flags, or the environment variable in brackets:

| Flag | Env | Default |
|---|---|---|
| `--broker` | `BROKER` | `tcp://mqtt.leetserve.com:1883` |
| `--topic` | `TOPIC` | `#` |
| `--db` | `DB` | `http://localhost:8086` |
| `--org` | `ORG` | `leetserve` |
| `--bucket` | `BUCKET` | `home` |
| `--ignore` | `IGNORE` | `log,metadata,state` |
| `--debug` | `DEBUG` | off (logs every skipped message) |
| | `INFLUXDB_TOKEN` | required |

## Develop / release

`make test`, `make run`. CI (`.github/workflows/ci.yaml`) vets and tests every PR and, on `main`, builds a
multi-arch image to `ghcr.io/gabeduke/reap` tagged `latest` and `sha-<short>`. Deploy with
`kubectl apply -f deploy/reap.yaml` (needs a `reap-creds` secret with `INFLUXDB_TOKEN`).
