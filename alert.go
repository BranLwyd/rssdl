package alert

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// Code describes a class of alerts.
type Code uint8

const (
	ERROR Code = iota
	NEW_ITEM
)

func (c Code) String() string {
	switch c {
	case ERROR:
		return "ERROR"
	case NEW_ITEM:
		return "NEW_ITEM"
	default:
		return "UNKNOWN"
	}
}

// Alerter indicates the ability to take an alert and act on it in some way.
// (e.g. running a command, logging, etc)
type Alerter interface {
	// Alert causes an alert to be fired. The `code` describes the class of
	// alert, and `details` is a human-readable description of the event that
	// caused the alert to be fired.
	Alert(ctx context.Context, code Code, details string) error
}

type cmdAlerter struct {
	cmd string
}

// NewCommand creates a new alerter that runs a specified command when an alert
// is fired. The subprocess has its ALERT_CODE environment variable set to the
// alert code, and its ALERT_DETAILS environment variable set to the alert
// details.
func NewCommand(cmd string) Alerter {
	return &cmdAlerter{cmd}
}

func (ca cmdAlerter) Alert(ctx context.Context, code Code, details string) error {
	cmd := exec.CommandContext(ctx, ca.cmd)
	cmd.Env = append(os.Environ(), fmt.Sprintf("ALERT_CODE=%s", code), fmt.Sprintf("ALERT_DETAILS=%s", details))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("alert command %q failed: %v", ca.cmd, err)
	}
	return nil
}
