package eventbus

import (
	"context"

	eventsv1 "zntr.io/hexagonal-bazel/api/system/events/v1"
)

type EventPublisher interface {
	// Publish the event throught the eventbus implementation.
	Publish(ctx context.Context, event *eventsv1.Event)
}
