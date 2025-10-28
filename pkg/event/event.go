package event

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	correlationId string
	stream        string
	source        string
	version       string
}

func NewEvent(correlationId string, stream string, version string) Event {
	return Event{
		correlationId: correlationId,
		source:        "dss",
		version:       version,
		stream:        stream,
	}
}

func (e Event) ToJson() []byte {
	output := struct {
		CorrelationId string `json:"correlation_id"`
		Stream        string `json:"stream"`
		Source        string `json:"source"`
		Version       string `json:"version"`
	}{
		CorrelationId: e.correlationId,
		Stream:        e.stream,
		Source:        e.source,
		Version:       e.version,
	}

	json, err := json.Marshal(output)

	if err != nil {
		fmt.Println("Event object can not be converted to JSON format")
	}

	return json
}
