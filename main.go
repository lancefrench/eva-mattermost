package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/lunny/html2md"
)

//https://docs.pivotal.io/event-alerts/1-2/using.html#webhook_targets
type IncomingWebhookMessage struct {
	Publisher string            `json:"publisher"`
	Topic     string            `json:"topic"`
	Timestamp string            `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
	Subject   string            `json:"subject"`
	Body      string            `json:"body"`
}

type OutgoingWebhookMessage struct {
	Text string `json:"text"`
}

func main() {
	url := os.Getenv("URL")
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println("ERROR reading: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		fmt.Println("body: ", string(bodyBytes))

		// What event alerts sends
		var incomingWebhookMessage IncomingWebhookMessage
		err = json.Unmarshal(bodyBytes, &incomingWebhookMessage)
		if err != nil {
			fmt.Println("ERROR unmarshalling: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		markdown := html2md.Convert(incomingWebhookMessage.Body)

		// Conversion
		outgoingMessage := &OutgoingWebhookMessage{
			Text: "Converted: " + markdown + " " + incomingWebhookMessage.Subject,
		}

		// Marshal into required style
		messageBytes, marshalErr := json.Marshal(outgoingMessage)
		if marshalErr != nil {
			fmt.Println("ERROR marshalling: ", marshalErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("sending: ", string(messageBytes))

		// Send to final destination
		_, err = http.DefaultClient.Post(url, "application/json", bytes.NewBuffer(messageBytes))
		if err != nil {
			fmt.Println("ERROR sending: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
