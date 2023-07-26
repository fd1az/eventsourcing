package base

import (
	"context"
	"errors"
)

// ErrEventMultipleAggregates when events holds different id
var ErrEventMultipleAggregates = errors.New("events holds events for more than one aggregate")

// ErrEventMultipleAggregateTypes when events holds different aggregate types
var ErrEventMultipleAggregateTypes = errors.New("events holds events for more than one aggregate type")

// ErrConcurrency when the currently saved version of the aggregate differs from the new ones
var ErrConcurrency = errors.New("concurrency error")

// ErrReasonMissing when the reason is not present in the events
var ErrReasonMissing = errors.New("event holds no reason")

// EventStore interface expose the methods an event store must uphold
type EventStore interface {
	Save(events []Event) error
	Get(ctx context.Context, id string, aggregateType string, afterVersion Version) (Iterator, error)
}

// ValidateEvents make sure the incoming events are valid
func ValidateEvents(aggregateID string, currentVersion Version, events []Event) error {
	aggregateType := events[0].AggregateType

	for _, event := range events {
		if event.AggregateID != aggregateID {
			return ErrEventMultipleAggregates
		}

		if event.AggregateType != aggregateType {
			return ErrEventMultipleAggregateTypes
		}

		if currentVersion+1 != event.Version {
			return ErrConcurrency
		}

		if event.Reason == "" {
			return ErrReasonMissing
		}

		currentVersion = event.Version
	}
	return nil
}

// ValidateEventsNoVersionCheck make sure the incoming events are valid
func ValidateEventsNoVersionCheck(aggregateID string, events []Event) error {
	aggregateType := events[0].AggregateType
	currentVersion := events[0].Version - 1

	for _, event := range events {
		if event.AggregateID != aggregateID {
			return ErrEventMultipleAggregates
		}

		if event.AggregateType != aggregateType {
			return ErrEventMultipleAggregateTypes
		}

		if currentVersion+1 != event.Version {
			return ErrConcurrency
		}
		if event.Reason == "" {
			return ErrReasonMissing
		}
		currentVersion = event.Version
	}
	return nil
}
