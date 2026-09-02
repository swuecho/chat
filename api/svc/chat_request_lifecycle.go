package svc

import (
	"context"
	"errors"
	"fmt"
)

type ChatRequestStatus string

const (
	ChatRequestPending   ChatRequestStatus = "pending"
	ChatRequestStreaming ChatRequestStatus = "streaming"
	ChatRequestCompleted ChatRequestStatus = "completed"
	ChatRequestFailed    ChatRequestStatus = "failed"
	ChatRequestCanceled  ChatRequestStatus = "canceled"
)

var ErrInvalidChatRequestTransition = errors.New("invalid chat request transition")

func transitionError(requestUUID string, target ChatRequestStatus) error {
	return fmt.Errorf("%w: request %s cannot transition to %s", ErrInvalidChatRequestTransition, requestUUID, target)
}

type ChatRequestLifecycle interface {
	Claim(context.Context, string, string, int32) error
	State(context.Context, string, string, int32) (ChatRequestState, error)
	StartStreaming(context.Context, string, string, int32) error
	Fail(context.Context, string, string, int32, string) error
	Cancel(context.Context, string, string, int32, string) error
}

var _ ChatRequestLifecycle = (*ChatService)(nil)
