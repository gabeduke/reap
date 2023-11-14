package main

import (
	"context"
	"fmt"
	"github.com/adampresley/sigint"
	"github.com/gabeduke/reap/pkg/reap"
	"github.com/lithammer/shortuuid/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

var brokerURL string
var dbURL string

func main() {
	readConfiguration()

	clientId := fmt.Sprintf("reap-%d", shortuuid.New())

	// build watermill client
	reaper, err := reap.NewReaper(brokerURL, dbURL, clientId)
	if err != nil {
		log.Fatal(err)
	}

	// start signal handler
	sigint.Listen(func() {
		err := reaper.Close()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	})

	// start message processing
	msgStream, err := reaper.Subscribe(context.Background(), "#")

	reaper.InfluxHandler(msgStream)
}

func readConfiguration() {
	// allow setting values also with commandline flags
	pflag.String("broker", "tcp://mqtt.leetserve.com:1883", "MQTT-broker url")
	viper.BindPFlag("broker", pflag.Lookup("broker"))

	pflag.String("bucket", "home", "InfluxDB bucket")
	viper.BindPFlag("bucket", pflag.Lookup("bucket"))

	pflag.String("db", "http://localhost:8086", "Influx database connection url")
	viper.BindPFlag("db", pflag.Lookup("db"))

	// parse values from environment variables
	viper.AutomaticEnv()
	pflag.Parse()

	brokerURL = viper.GetString("broker")
	dbBucket := viper.GetString("bucket")
	dbURL = viper.GetString("db")

	log.Infof("Using broker %s", brokerURL)
	log.Infof("Using bucket %s", dbBucket)
	log.Infof("Using database %s", dbURL)
}
