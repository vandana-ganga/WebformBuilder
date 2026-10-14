package messaging

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

const SubmissionStream = "SUBMISSIONS"

const SubmissionSubject = "submissions.created"

func CreateSubmissionStream(
	js nats.JetStreamContext,
) error {

	// Check whether stream already exists.
	_, err := js.StreamInfo(SubmissionStream)

	if err == nil {
		return nil
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:      SubmissionStream,
		Subjects:  []string{SubmissionSubject},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,

		// Keep messages for 24 hours.
		MaxAge: 24 * time.Hour,
	})

	if err != nil {
		return fmt.Errorf(
			"failed to create submission stream: %w",
			err,
		)
	}

	return nil
}