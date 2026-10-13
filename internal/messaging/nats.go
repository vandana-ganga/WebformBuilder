package messaging

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

type NATS struct {
	Conn *nats.Conn
	JS   nats.JetStreamContext
}

func NewNATS(url string) (*NATS, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()

		return nil, fmt.Errorf(
			"failed to initialize JetStream: %w",
			err,
		)
	}

	return &NATS{
		Conn: conn,
		JS:   js,
	}, nil
}

func (n *NATS) Close() {
	if n.Conn != nil {
		n.Conn.Close()
	}
}