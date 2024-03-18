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

		// Check if topic should be ignored
		if shouldIgnoreTopic(topic) {
			log.Info("Ignoring topic based on filter criteria")
			msg.Ack()
			continue
			// Check if the message conforms to the Homie convention
		} else if strings.HasPrefix(topic, "homie/") {
			r.handleHomieMessage(topic, payload)
			continue
		} else {
			// Handle non-Homie messages (simple floats)

			topicParts := strings.Split(topic, "/")

			if len(topicParts) >= 3 {
				group := topicParts[0]
				device := topicParts[1]
				measurement := strings.Join(topicParts[2:], "-")
				convertedPayload, err := strconv.ParseFloat(payload, 64)
				if err != nil {
					log.WithError(err).WithField("payload", payload).Error("Failed to parse float payload")
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
					msg.Ack()
					continue
				}
			} else {
				log.Info("Message did not contain expected topic parts, skipping it")
			}
			msg.Ack()
		}
	}
}

func (r *Reap) handleHomieMessage(topic, payload string) {
	topicParts := strings.Split(topic, "/")
	if len(topicParts) < 4 {
		log.WithField("topic", topic).Error("Homie topic does not conform to expected structure")
		return
	}

	deviceID := topicParts[1]
	nodeID := topicParts[2]
	property := topicParts[3]

	// Parse the payload as float, assuming Homie convention for property values
	value, err := strconv.ParseFloat(payload, 64)
	if err != nil {
		log.WithError(err).WithField("payload", payload).Error("Failed to parse Homie payload as float")
		return
	}

	bucket := viper.GetString("bucket")
	measurement := property // Use property as measurement, or customize as needed
	tags := map[string]string{
		"device": deviceID,
		"node":   nodeID,
	}

	err = r.WritePoint(measurement, bucket, value, tags)
	if err != nil {
		log.WithError(err).WithField("bucket", bucket).Error("Failed to write Homie data point to InfluxDB")
	}
}

func shouldIgnoreTopic(topic string) bool {
	// Define topics to ignore
	ignoredTopics := []string{"log", "metadata", "state"}

	for _, ignored := range ignoredTopics {
		if strings.Contains(topic, ignored) {
			return true
		}
	}
	return false
}
