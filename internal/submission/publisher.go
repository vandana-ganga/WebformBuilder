package submission

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"

	"webform-builder/internal/messaging"
)

type Publisher struct {
	JS nats.JetStreamContext
}

func NewPublisher(
	js nats.JetStreamContext,
) *Publisher {
	return &Publisher{
		JS: js,
	}
}

func (p *Publisher) Publish(
	ctx context.Context,
	message SubmissionMessage,
) error {

	data, err := json.Marshal(message)

	if err != nil {
		return fmt.Errorf(
			"failed to marshal submission message: %w",
			err,
		)
	}

	// Use the submission ID as NATS message ID.
	//
	// This helps JetStream detect duplicate publishes
	// if the same message is retried.
	msg := &nats.Msg{
		Subject: messaging.SubmissionSubject,
		Data:    data,
		Header: nats.Header{
			"Nats-Msg-Id": []string{
				message.SubmissionID,
			},
		},
	}

	// PublishMsg waits for a JetStream acknowledgement.
	_, err = p.JS.PublishMsg(msg)

	if err != nil {
		return fmt.Errorf(
			"failed to publish submission to JetStream: %w",
			err,
		)
	}

	return nil
}