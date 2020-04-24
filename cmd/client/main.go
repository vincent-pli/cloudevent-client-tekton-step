/*
Copyright 2018 The whatever Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Implements a simple utility for sending a JSON-encoded sample event.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)
const(
	Tese = "xxx"
)
func main() {
	
	var target string
	var eventID string
	var eventType string
	var source string
	var data string
	flag.StringVar(&target, "target", "", "Target")
        flag.StringVar(&eventID, "event-id", "", "Event ID to use. Defaults to a generated UUID")
        flag.StringVar(&eventType, "event-type", "google.events.action.demo", "The Event Type to use.")
        flag.StringVar(&source, "source", "", "Source URI to use. Defaults to the current machine's hostname")
        flag.StringVar(&data, "data", `{"hello": "world!"}`, "Event data")
	flag.Parse()
	
	var untyped map[string]interface{}
	if err := json.Unmarshal([]byte(data), &untyped); err != nil {
		fmt.Println("Currently sendevent only supports JSON event data")
		os.Exit(1)
	}

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Printf("failed to create client, %v", err)
		os.Exit(1)
	}
	
	event := cloudevents.NewEvent()
	if eventID != "" {
		event.SetID(eventID)
	}
	event.SetType(eventType)
	event.SetSource(source)
	if err := event.SetData(cloudevents.ApplicationJSON, untyped); err != nil {
		log.Printf("failed to set data, %v", err)
		os.Exit(1)
	}

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), target)

	// Send that Event.
	if result := c.Send(ctx, event); !cloudevents.IsACK(result) {
		log.Fatalf("failed to send, %v", result)
	}
}
