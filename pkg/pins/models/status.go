package models

import (
	"fmt"
)

const (
	StatusInit Status = iota
	StatusQueued
	StatusPinQueued
	StatusUnpinQueued
	StatusPinning
	StatusPinned
	StatusUnpinning
	StatusUnpinned
	StatusError
	StatusUndefined
)

type Status int

var StatusStrings = []string{
	"initialized",
	"queued",
	"pin_queued",
	"unpin_queued",
	"pinning",
	"pinned",
	"unpinning",
	"unpinned",
	"error",
	"undefined",
}

func (s Status) String() (status string) {
	if int(s) > -1 && len(StatusStrings) > int(s) {
		status = StatusStrings[s]
	}
	return
}

func ToStatus(s string) (Status, error) {
	for i, ss := range StatusStrings {
		if ss == s {
			return Status(i), nil
		}
	}
	return Status(-1), fmt.Errorf("unknown status string")
}
