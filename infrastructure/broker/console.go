package broker

import (
	"context"
	"fmt"

	eventsv1 "zntr.io/hexagonal-bazel/api/system/events/v1"
	"zntr.io/hexagonal-bazel/pkg/eventbus"
)

func Console() eventbus.EventPublisher {
	return &consolePublisher{}
}

type consolePublisher struct {
}

func (p *consolePublisher) Publish(ctx context.Context, e *eventsv1.Event) {
	fmt.Printf("Publishing event: %+v", e)
}
