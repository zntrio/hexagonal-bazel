package broker

import (
	"context"

	eventsv1 "zntr.io/hexagonal-bazel/api/system/events/v1"
	"zntr.io/hexagonal-bazel/pkg/eventbus"
)

func NoOP() eventbus.EventPublisher {
	return &noopPublisher{}
}

type noopPublisher struct {
}

func (p *noopPublisher) Publish(ctx context.Context, e *eventsv1.Event) error {
	// Do nothing
	return nil
}
