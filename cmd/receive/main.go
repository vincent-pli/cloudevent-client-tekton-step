package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func main() {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :8080\n")
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, receive))
}

func receive(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, error) {
	fmt.Printf("%s", event)

	data := `{"hello": "you!"}`
	var untyped map[string]interface{}
	if err := json.Unmarshal([]byte(data), &untyped); err != nil {
		fmt.Println("Currently sendevent only supports JSON event data")
		return nil, err
	}

	respEvent := cloudevents.NewEvent()
	if err := respEvent.SetData(cloudevents.ApplicationJSON, untyped); err != nil {
		log.Printf("failed to set data, %v", err)
		return nil, err
	}
	return &respEvent, nil
}
