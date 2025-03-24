package elkproducer

import "time"

type LogDocument struct {
	Timestamp time.Time   `json:"timestamp"`
	Log       interface{} `json:"log"`
	Url       string      `json:"url"`
}
