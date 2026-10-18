package reap

import (
	"errors"
	"log/slog"
)

// Writer stores a point. Implementations must not block the MQTT callback.
type Writer interface {
	Write(Point)
}

// Handler routes messages to a Writer.
type Handler struct {
	Writer Writer
	Ignore map[string]struct{}
	Log    *slog.Logger
}

// Handle processes one message. It never blocks and never fails the caller:
// a bad message is logged and dropped so one payload cannot stall ingestion.
func (h *Handler) Handle(topic, payload string) {
	if Ignored(topic, h.Ignore) {
		h.Log.Debug("ignored topic", "topic", topic)
		return
	}
	p, err := Parse(topic, payload)
	switch {
	case errors.Is(err, ErrNotNumeric):
		h.Log.Debug("skipping non-numeric payload", "topic", topic)
	case err != nil:
		h.Log.Debug("skipping message", "topic", topic, "err", err)
	default:
		h.Writer.Write(p)
	}
}
