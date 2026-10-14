package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"webform-builder/internal/config"
	"webform-builder/internal/database"
	"webform-builder/internal/messaging"
	"webform-builder/internal/submission"
)

func main() {

	// Load configuration.
	cfg, err := config.Load()

	if err != nil {
		log.Fatal(
			"failed to load config:",
			err,
		)
	}

	// Connect PostgreSQL.
	db, err := database.ConnectDB(
		cfg.DatabaseURL,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Connect NATS.
	natsClient, err := messaging.NewNATS(
		cfg.NATSURL,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer natsClient.Close()

	// Make sure stream exists.
	err = messaging.CreateSubmissionStream(
		natsClient.JS,
	)

	if err != nil {
		log.Fatal(err)
	}

	// Create repository.
	repo := submission.NewRepository(db)

	// Create durable pull consumer.
	consumer, err := natsClient.JS.PullSubscribe(
		messaging.SubmissionSubject,
		"submission-workers",

		nats.AckExplicit(),

		nats.AckWait(30*time.Second),

		nats.MaxDeliver(5),
	)

	if err != nil {
		log.Fatal(
			"failed to create submission consumer:",
			err,
		)
	}

	log.Println(
		"submission worker started",
	)

	for {

		// Pull subscriptions must use Fetch,
		// NextMsg only works on push subscriptions.
		msgs, err := consumer.Fetch(
			1,
			nats.MaxWait(5*time.Second),
		)

		if errors.Is(err, nats.ErrTimeout) {
			continue
		}

		if err != nil {
			log.Println(
				"failed to receive message:",
				err,
			)

			// Avoid a hot loop on persistent errors.
			time.Sleep(time.Second)

			continue
		}

		if len(msgs) == 0 {
			continue
		}

		msg := msgs[0]

		err = processMessage(
			msg,
			repo,
		)

		if err != nil {

			log.Println(
				"failed to process message:",
				err,
			)

			// Don't ACK.
			//
			// JetStream will redeliver
			// according to MaxDeliver.
			continue
		}

		// ACK ONLY after successful DB write.
		if err := msg.Ack(); err != nil {

			log.Println(
				"failed to ACK message:",
				err,
			)
		}
	}
}

func processMessage(
	msg *nats.Msg,
	repo *submission.Repository,
) error {

	var event submission.SubmissionMessage

	err := json.Unmarshal(
		msg.Data,
		&event,
	)

	if err != nil {
		return err
	}

	if event.SubmissionID == "" {
		return &InvalidMessageError{
			Message: "submission ID is missing",
		}
	}

	if event.TenantID == "" {
		return &InvalidMessageError{
			Message: "tenant ID is missing",
		}
	}

	if event.FormID == "" {
		return &InvalidMessageError{
			Message: "form ID is missing",
		}
	}

	if event.FormVersionID == "" {
		return &InvalidMessageError{
			Message: "form version ID is missing",
		}
	}

	_, err = repo.SubmitForm(
		context.Background(),
		event.TenantID,
		event.FormID,
		event.FormVersionID,
		event.SubmissionID,
		event.Data,
		event.CreatedAt,
	)

	return err
}

type InvalidMessageError struct {
	Message string
}

func (e *InvalidMessageError) Error() string {
	return e.Message
}