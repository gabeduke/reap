package reap

import (
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Subscribe connects to broker and feeds every message on topic to h. The
// subscription is re-established on each (re)connect, which is what keeps the
// consumer alive across broker restarts and network drops.
func Subscribe(broker, clientID, topic string, h *Handler, log *slog.Logger) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetMaxReconnectInterval(30 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetCleanSession(true)
	opts.OnConnectionLost = func(_ mqtt.Client, err error) { log.Warn("mqtt connection lost", "err", err) }
	opts.OnConnect = func(c mqtt.Client) {
		log.Info("mqtt connected", "broker", broker, "topic", topic)
		t := c.Subscribe(topic, 0, func(_ mqtt.Client, m mqtt.Message) {
			h.Handle(m.Topic(), string(m.Payload()))
		})
		if t.Wait() && t.Error() != nil {
			log.Error("mqtt subscribe failed", "topic", topic, "err", t.Error())
		}
	}

	c := mqtt.NewClient(opts)
	if t := c.Connect(); t.Wait() && t.Error() != nil {
		return nil, fmt.Errorf("connect %s: %w", broker, t.Error())
	}
	return c, nil
}
