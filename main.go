package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gabeduke/reap/internal/reap"
)

// env returns $key, or def when unset. Flags default to the environment so the
// same names (BROKER, DB, BUCKET, ...) work either way.
func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func main() {
	broker := flag.String("broker", env("BROKER", "tcp://mqtt.leetserve.com:1883"), "MQTT broker URL [BROKER]")
	topic := flag.String("topic", env("TOPIC", "#"), "MQTT topic filter [TOPIC]")
	db := flag.String("db", env("DB", "http://localhost:8086"), "InfluxDB URL [DB]")
	org := flag.String("org", env("ORG", "leetserve"), "InfluxDB org [ORG]")
	bucket := flag.String("bucket", env("BUCKET", "home"), "InfluxDB bucket [BUCKET]")
	ignore := flag.String("ignore", env("IGNORE", "log,metadata,state"), "comma-separated topic segments to skip [IGNORE]")
	debug := flag.Bool("debug", env("DEBUG", "") != "", "log every skipped message [DEBUG]")
	flag.Parse()

	lvl := slog.LevelInfo
	if *debug {
		lvl = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))

	token := os.Getenv("INFLUXDB_TOKEN")
	if token == "" {
		log.Error("INFLUXDB_TOKEN is not set")
		os.Exit(1)
	}

	skip := map[string]struct{}{}
	for _, s := range strings.Split(*ignore, ",") {
		if s = strings.TrimSpace(s); s != "" {
			skip[s] = struct{}{}
		}
	}

	log.Info("starting", "broker", *broker, "topic", *topic, "db", *db, "org", *org, "bucket", *bucket)
	out := reap.NewInflux(*db, token, *org, *bucket, log)
	h := &reap.Handler{Writer: out, Ignore: skip, Log: log}

	host, _ := os.Hostname()
	client, err := reap.Subscribe(*broker, fmt.Sprintf("reap-%s", host), *topic, h, log)
	if err != nil {
		log.Error("cannot start", "err", err)
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Info("shutting down")
	client.Disconnect(250)
	out.Close()
}
