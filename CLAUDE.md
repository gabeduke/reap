# CLAUDE.md

## Project

`reap` subscribes to `#` on an MQTT broker (`mqtt.leetserve.com`) and writes numeric messages to the `home` bucket of the in-cluster InfluxDB v2 (`influxdb-influxdb2`). See README.md for the topic-to-point mapping and config.

## Layout

```
main.go                  # flags/env, wiring, signal handling
internal/reap/parse.go   # topic+payload -> Point; segment-based ignore filter (pure, tested)
internal/reap/handler.go # never blocks/fails on a bad message
internal/reap/mqtt.go    # paho client; resubscribes in OnConnect
internal/reap/influx.go  # non-blocking batched write API; errors logged from a channel
deploy/reap.yaml         # Deployment (expects reap-creds secret with INFLUXDB_TOKEN)
.github/workflows/ci.yaml# vet+test, then multi-arch image to ghcr.io/gabeduke/reap on main
```

## Commands

`make test`, `make build`, `make run` (needs `INFLUXDB_TOKEN`).

## Gotchas

- The MQTT callback must never block: the old Watermill adapter waited for an Ack, and one unparseable message (e.g. the retained `status=online`) stalled the whole consumer. Keep `Handle` non-blocking and error-tolerant.
- The subscription is made in `OnConnect` because sessions are clean; otherwise a broker restart silently ends ingestion.
- Influx API tokens other than the admin token are lost when the cluster's Influx pod restarts (its volume is an `emptyDir`); a bare `unauthorized` in the logs means the `reap-creds` token is stale.
- The old Homie (`homie/...`) handler was never finished and was removed. Pre-refresh WIP is in `git stash` on this clone only.
