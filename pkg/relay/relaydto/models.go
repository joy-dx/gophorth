package relaydto

import (
	"encoding/json"
	"time"
)

type EventChannel string
type EventRef string

type RelaySinkConfig struct {
	Ref   string     `yaml:"ref"`
	Level RelayLevel `json:"level"`
}

type Event struct {
	Channel   EventChannel    `json:"channel"`
	Ref       EventRef        `json:"ref"` // The event key (used for routing)
	Level     RelayLevel      `json:"level"`
	Timestamp time.Time       `json:"timestamp" ts_type:"string"` // Event creation time
	Data      json.RawMessage `json:"data"`                       // The marshaled JSON data
}
