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
	"io/ioutil"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"knative.dev/pkg/apis/duck"
	"knative.dev/pkg/apis/duck/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	BaseDir  = "/tekton/results/"
	FileName = "response"
)

var (
	c   client.Client
	err error
)

func init() {
	c, err = client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		log.Infof("Create client failed: %+v", err)
		os.Exit(1)
	}
}

func main() {
	var target string
	var eventID string
	var eventType string
	var source string
	var data string
	var slink string
	flag.StringVar(&target, "target", "", "Target")
	flag.StringVar(&eventID, "event-id", "", "Event ID to use. Defaults to a generated UUID")
	flag.StringVar(&eventType, "event-type", "google.events.action.demo", "The Event Type to use.")
	flag.StringVar(&source, "source", "", "Source URI to use. Defaults to the current machine's hostname")
	flag.StringVar(&data, "data", `{"hello": "world!"}`, "Event data")
	flag.StringVar(&slink, "slink", "", "Slink")
	flag.Parse()

	var sinkOjb corev1.ObjectReference
	if slink != "" {
		reader := strings.NewReader(slink)
		decoder := yaml.NewYAMLToJSONDecoder(reader)
		err := decoder.Decode(&sinkOjb)
		if err != nil {
			log.Infof("unmarshal slink failed: %+v", err)
			os.Exit(1)
		}

		target, err = getSinkURI(context.Background(), c, &sinkOjb)
		if err != nil {
			log.Infof("Parse slink failed: %+v", err)
			os.Exit(1)
		}
	}

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
	reps, result := c.Request(ctx, event)
	if cloudevents.IsNACK(result) {
		log.Fatalf("failed to send, %v", result)
		os.Exit(1)
	}
	fmt.Printf("%s", reps)
	err = writeFiles(reps.Data())
	if err != nil {
		log.Fatalf("failed to write response to file, %v", err)
		os.Exit(1)
	}
}

// GetSinkURI retrieves the sink URI from the object referenced by the given
// ObjectReference.
func getSinkURI(ctx context.Context, c client.Client, sink *corev1.ObjectReference) (string, error) {
	if sink == nil {
		return "", fmt.Errorf("sink ref is nil")
	}
	objIdentifier := fmt.Sprintf("\"%s/%s\" (%s)", sink.Namespace, sink.Name, sink.GroupVersionKind())

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(sink.GroupVersionKind())
	err := c.Get(context.TODO(), client.ObjectKey{Name: sink.Name, Namespace: sink.Namespace}, u)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize sink %s: %v", objIdentifier, err)
	}

	t := v1beta1.AddressableType{}
	err = duck.FromUnstructured(u, &t)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize sink %s: %v", objIdentifier, err)
	}

	if t.Status.Address == nil {
		return "", fmt.Errorf("sink %s does not contain address", objIdentifier)
	}

	if t.Status.Address.URL == nil {
		return "", fmt.Errorf("sink %s contains an empty URL", objIdentifier)
	}

	return t.Status.Address.URL.String(), nil
}

// Write file to the disk
func writeFiles(respData []byte) error {
	outputFile := BaseDir + FileName
	err = ioutil.WriteFile(outputFile, respData, 0644)
	if err != nil {
		log.Errorf("Write response to file failed: %+v:", err)
		return err
	}

	return nil
}
