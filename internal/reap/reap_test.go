package reap

import (
	"errors"
	"io"
	"log/slog"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		topic, payload string
		wantM          string
		wantV          float64
		wantErr        error
		anyErr         bool
	}{
		{topic: "grow/tent1/temperature", payload: "21.7", wantM: "temperature", wantV: 21.7},
		{topic: "grow/tent1/temperature", payload: " 21.7\n", wantM: "temperature", wantV: 21.7},
		{topic: "a/b/c/d", payload: "1", wantM: "c-d", wantV: 1},
		{topic: "grow/tent1/status", payload: "online", wantErr: ErrNotNumeric},
		{topic: "grow/tent1/config/applied", payload: `{"mode":"flower"}`, wantErr: ErrNotNumeric},
		{topic: "grow/tent1", payload: "1", anyErr: true},
		{topic: "grow//temperature", payload: "1", anyErr: true},
	}
	for _, c := range cases {
		p, err := Parse(c.topic, c.payload)
		switch {
		case c.wantErr != nil:
			if !errors.Is(err, c.wantErr) {
				t.Errorf("%s %q: err = %v, want %v", c.topic, c.payload, err, c.wantErr)
			}
		case c.anyErr:
			if err == nil {
				t.Errorf("%s: expected an error", c.topic)
			}
		default:
			if err != nil {
				t.Errorf("%s: %v", c.topic, err)
				continue
			}
			if p.Measurement != c.wantM || p.Value != c.wantV {
				t.Errorf("%s: got %s=%v, want %s=%v", c.topic, p.Measurement, p.Value, c.wantM, c.wantV)
			}
			if p.Tags["group"] == "" || p.Tags["device"] == "" || p.Tags["sensor"] != c.wantM {
				t.Errorf("%s: bad tags %v", c.topic, p.Tags)
			}
		}
	}
}

func TestIgnoredMatchesWholeSegments(t *testing.T) {
	ig := map[string]struct{}{"log": {}, "metadata": {}, "state": {}}
	for topic, want := range map[string]bool{
		"grow/tent1/state":         true,
		"dev/x/log/errors":         true,
		"dining/state-of-charge/x": false, // substring only; the old filter dropped this
		"grow/tent1/temperature":   false,
	} {
		if got := Ignored(topic, ig); got != want {
			t.Errorf("Ignored(%q) = %v, want %v", topic, got, want)
		}
	}
}

type fake struct{ got []Point }

func (f *fake) Write(p Point) { f.got = append(f.got, p) }

// A non-numeric message must not prevent later numeric ones from being written
// (the old consumer blocked forever on the first one).
func TestHandlerSurvivesBadPayloads(t *testing.T) {
	w := &fake{}
	h := &Handler{Writer: w, Ignore: map[string]struct{}{"state": {}}, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	h.Handle("grow/tent1/status", "online")
	h.Handle("grow/tent1/config/applied", "{}")
	h.Handle("grow/tent1/state", `{"temperature":1}`)
	h.Handle("bad", "1")
	h.Handle("grow/tent1/temperature", "11.2")
	if len(w.got) != 1 || w.got[0].Measurement != "temperature" || w.got[0].Value != 11.2 {
		t.Fatalf("got %+v", w.got)
	}
}
