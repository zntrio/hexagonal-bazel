package link

import (
	eventsv1 "zntr.io/hexagonal-bazel/api/system/events/v1"
	urlshortenerv1 "zntr.io/hexagonal-bazel/api/urlshortener/v1"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/GoWebProd/uuid7"
)

// LinkCreated generates an event to notify shortened link creation.
func LinkCreated(obj *urlshortenerv1.Link) *eventsv1.Event {
	payload, err := anypb.New(obj)
	if err != nil {
		panic(err)
	}
	return &eventsv1.Event{
		EventId:   uuid7.New().Next().String(),
		EventType: "urlshortener.v1.link.created",
		Payload:   payload,
	}
}
