package conference

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	CreateConferenceLink(ctx context.Context, slotID, userID uuid.UUID) (joinURL string, err error)
	CancelConferenceLink(ctx context.Context, joinURL string) error
}

type Mock struct {
	Delay      time.Duration
	ShouldFail bool
	BadLink    string
}

func (m *Mock) CreateConferenceLink(ctx context.Context, slotID, userID uuid.UUID) (string, error) {
	if m.Delay > 0 {
		t := time.NewTimer(m.Delay)
		defer t.Stop()
		select {
		case <-ctx.Done():
			err := ctx.Err()
			log.Printf("conference mock: CreateConferenceLink %v", err)
			return "", err
		case <-t.C:
		}
	}
	if m.ShouldFail {
		err := errors.New("mock: conference unavailable")
		log.Printf("conference mock: CreateConferenceLink %v", err)
		return "", err
	}
	switch strings.ToLower(strings.TrimSpace(m.BadLink)) {
	case "empty":
		return "", nil
	case "invalid":
		return "not-a-valid-url", nil
	}
	link := fmt.Sprintf("https://mock-conference.invalid/join/%s/%s", slotID.String(), userID.String())
	return link, nil
}

func (m *Mock) CancelConferenceLink(_ context.Context, joinURL string) error {
	if strings.TrimSpace(joinURL) != "" {
		log.Print("conference mock: conference link cancelled")
	}
	return nil
}
