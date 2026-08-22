package sla

import "errors"

var (
	errFlightMissing = errors.New("flight does not exist")
	errTaskMissing   = errors.New("task does not exist")
)
