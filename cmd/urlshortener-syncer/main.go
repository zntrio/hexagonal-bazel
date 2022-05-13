package main

import (
	"log"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	eventsv1 "zntr.io/hexagonal-bazel/api/system/events/v1"
	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
)

const (
	subSubjectName = "urlshortener.v1.link.created"
)

func main() {
	// Connect to jetstream
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}

	// Get jetstream handle
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}

	// Create a subscriber
	sub, err := js.PullSubscribe(subSubjectName, "urlshortener-syncer", nats.PullMaxWaiting(128))
	if err != nil {
		log.Fatal(err)
	}

	for {
		msgs, _ := sub.Fetch(5)
		for _, msg := range msgs {
			msg.Ack()

			// Decode event
			var event eventsv1.Event
			err := proto.Unmarshal(msg.Data, &event)
			if err != nil {
				log.Fatal(err)
			}

			// Decode payload
			var link urlshortenerv1.Link
			if err := event.Payload.UnmarshalTo(&link); err != nil {
				log.Fatal(err)
			}

			log.Println("urlshortener-syncer service")
			log.Printf("EventID:%s, LinkID: %s, URL:%s\n", event.EventId, link.Id, *link.Url)
		}
	}
}
