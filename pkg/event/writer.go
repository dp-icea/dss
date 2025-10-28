package event

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type EventWriter struct {
	cli *resty.Client
}

func NewEventWriter(cli *resty.Client) *EventWriter {
	return &EventWriter{cli: cli}
}

func (ew *EventWriter) ConfigEventWriter(url string) {
	ew.cli.SetTimeout(1 * time.Minute)
	ew.cli.SetHeaders(map[string]string{
		"Content-Type": "application/json",
	})
}

func (ew *EventWriter) Create(event Event) {
	requestBody := event.ToJson()
	// TODO: change this to env variables
	resp, err := ew.cli.R().SetBody(requestBody).Post("http://event-store-api:8003/api/v1/events/")

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Event created - Response Status Code (", resp.StatusCode(), ") - { ", resp, " }")
}
