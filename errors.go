package hydrafetch

import "fmt"

type Error struct {
	Code      string
	Message   string
	Status    int
	RequestID string
	Details   any
}

func (e *Error) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("hydrafetch: %s (%s, request %s)", e.Message, e.Code, e.RequestID)
	}
	return fmt.Sprintf("hydrafetch: %s (%s)", e.Message, e.Code)
}

func (e *Error) IsAuth() bool { return e.Status == 401 || e.Status == 403 }

func (e *Error) IsOutOfCredits() bool { return e.Status == 402 }

func (e *Error) IsInvalidRequest() bool { return e.Status == 400 || e.Status == 422 }

func (e *Error) IsRetryable() bool { return e.Status == 429 || e.Status >= 500 }

func (e *Error) IsTimeout() bool { return e.Code == "TIMEOUT" }
