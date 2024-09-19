package logger

// INSTALL: go install golang.org/x/tools/cmd/stringer@latest
//go:generate stringer -type Level -linecomment --output level-string.go

type Level int

const (
	Fatal Level = iota // FATAL
	Error              // ERROR
	Warn               // WARN
	Info               // INFO
	Debug              // DEBUG
	Trace              // TRACE
)
