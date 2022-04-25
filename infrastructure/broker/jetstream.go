package broker

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"

	eventsv1 "zntr.io/hexagonal-bazel/api/system/events/v1"
	"zntr.io/hexagonal-bazel/pkg/eventbus"
)

func Jetstream(js nats.JetStreamContext) eventbus.EventPublisher {
	return &jetstreamPublisher{
		js: js,
	}
}

type jetstreamPublisher struct {
	js nats.JetStreamContext
}

func (p *jetstreamPublisher) Publish(ctx context.Context, e *eventsv1.Event) error {
	// Check argument
	if e == nil {
		return errors.New("jetstream: unable to publish nil event")
	}

	// Marshal to protobuf
	payload, err := proto.Marshal(e)
	if err != nil {
		return fmt.Errorf("jetstream: unable to marshal event as a protobuf payload: %w", err)
	}

	// Publish to subject
	if _, err := p.js.Publish(e.EventType, payload); err != nil {
		return fmt.Errorf("jetstream: unable to publish event to %q subject: %w", e.EventType, err)
	}

	return nil
}
