package broker

import (
	"context"
	"fmt"

	eventsv1 "zntr.io/hexagonal-bazel/api/system/events/v1"
)

func Console() EventPublisher {
	return &consolePublisher{}
}

type consolePublisher struct{}

func (p *consolePublisher) Publish(ctx context.Context, e *eventsv1.Event) error {
	fmt.Printf("Publishing event: %+v", e)
	return nil
}
