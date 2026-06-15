package sse

import "time"

// Event is a single Server-Sent Events event.
type Event struct {
	// ID is the value of the id field. It can be empty if the field is not set.
	//
	// Use [Decoder.LastEventID] to get the current last event ID of the stream.
	ID string
	// Type is the value of the event field. It defaults to "message".
	Type string
	// Data is the event data accumulated from the data fields.
	Data string
	// Retry is the reconnection delay. It is nil if the event has no retry field.
	//
	// Use [Decoder.Retry] to get the current retry interval of the stream.
	Retry *time.Duration
}
