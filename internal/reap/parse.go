// Package reap turns MQTT sensor messages into InfluxDB points.
package reap

import (
	"fmt"
	"strconv"
	"strings"
)

// Point is one numeric sample derived from an MQTT message.
type Point struct {
	Measurement string
	Tags        map[string]string
	Value       float64
}

// ErrNotNumeric marks a payload that is not a number. It is expected for
// topics such as status/mode/config, so callers should not treat it as a fault.
var ErrNotNumeric = fmt.Errorf("payload is not numeric")

// Parse maps "<group>/<device>/<name...>" with a numeric payload to a Point.
// The measurement is the remainder of the topic joined with "-", and the tags
// are group, device and sensor (the same as the measurement).
func Parse(topic, payload string) (Point, error) {
	parts := strings.Split(topic, "/")
	if len(parts) < 3 {
		return Point{}, fmt.Errorf("topic %q needs at least group/device/name", topic)
	}
	for _, p := range parts {
		if p == "" {
			return Point{}, fmt.Errorf("topic %q has an empty segment", topic)
		}
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(payload), 64)
	if err != nil {
		return Point{}, ErrNotNumeric
	}
	m := strings.Join(parts[2:], "-")
	return Point{
		Measurement: m,
		Value:       v,
		Tags:        map[string]string{"group": parts[0], "device": parts[1], "sensor": m},
	}, nil
}

// Ignored reports whether any topic segment equals one of the ignored names.
// Matching is per segment, so "dining/state-of-charge" is not ignored by "state".
func Ignored(topic string, ignore map[string]struct{}) bool {
	for _, seg := range strings.Split(topic, "/") {
		if _, ok := ignore[seg]; ok {
			return true
		}
	}
	return false
}
