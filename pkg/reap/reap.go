package reap

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gabeduke/reap/pkg/influx"
	"github.com/gabeduke/reap/pkg/mqttc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

type Reap struct {
	*mqttc.Mqttc
	*influx.Influx
}

func NewReaper(broker, db, clientID string) (*Reap, error) {
	reaper := &Reap{}
	var err error
	cfg := mqttc.Config{
		BrokerUrl: broker,
		ClientID:  clientID,
	}

	reaper.Mqttc, err = mqttc.NewMqttc(cfg, watermill.NewStdLogger(false, false))
	if err != nil {
		return reaper, err
	}

	reaper.Influx, err = influx.NewClient(db)

	return reaper, nil
}

func (r *Reap) InfluxHandler(messages <-chan *message.Message) {
	for msg := range messages {
		topic := msg.Metadata[mqttc.MQTTC_TOPIC]
		payload := string(msg.Payload)

		log.WithFields(log.Fields{
			"topic":   topic,
			"payload": payload,
		}).Info("received message")

		topicParts := strings.Split(topic, "/")

		if !strings.Contains(topic, "state") && len(topicParts) >= 3 {
			group := topicParts[0]
			device := topicParts[1]
			measurement := strings.Join(topicParts[2:len(topicParts)], "-")

			convertedPayload, err := strconv.ParseFloat(payload, 32)

			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Info("Failed to parse number for payload")
				continue
			}

			tags := map[string]string{
				"group":  group,
				"device": device,
				"sensor": measurement,
			}
			bucket := viper.GetString("bucket")
			err = r.WritePoint(measurement, bucket, convertedPayload, tags)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err,
					"bucket": bucket,
				}).Info("Failed to write point to influx")
				continue
			}
		} else {
			log.Info("Message did not contain expected topic parts, skipping it")
		}

		// we need to Acknowledge that we received and processed the message,
		// otherwise, it will be resent over and over again.
		msg.Ack()
	}
}
