package bbolt

import (
	"errors"
	"fmt"

	"github.com/hallgren/eventsourcing/base"
	eventstore "github.com/hallgren/eventsourcing/eventstore"
	"go.etcd.io/bbolt"
)

type iterator struct {
	tx              *bbolt.Tx
	bucketName      string
	firstEventIndex uint64
	cursor          *bbolt.Cursor
	serializer      eventstore.Serializer
}

// Close closes the iterator
func (i *iterator) Close() {
	i.tx.Rollback()
}

// Next return the next event
func (i *iterator) Next() (base.Event, error) {
	var k, obj []byte
	if i.cursor == nil {
		bucket := i.tx.Bucket([]byte(i.bucketName))
		if bucket == nil {
			return base.Event{}, base.ErrNoMoreEvents
		}
		i.cursor = bucket.Cursor()
		k, obj = i.cursor.Seek(itob(i.firstEventIndex))
		if k == nil {
			return base.Event{}, base.ErrNoMoreEvents
		}
	} else {
		k, obj = i.cursor.Next()
	}
	if k == nil {
		return base.Event{}, base.ErrNoMoreEvents
	}
	bEvent := boltEvent{}
	err := i.serializer.Unmarshal(obj, &bEvent)
	if err != nil {
		return base.Event{}, errors.New(fmt.Sprintf("could not deserialize event, %v", err))
	}
	f, ok := i.serializer.Type(bEvent.AggregateType, bEvent.Reason)
	if !ok {
		// if the typ/reason is not register jump over the event
		return i.Next()
	}
	eventData := f()
	err = i.serializer.Unmarshal(bEvent.Data, &eventData)
	if err != nil {
		return base.Event{}, errors.New(fmt.Sprintf("could not deserialize event data, %v", err))
	}
	event := base.Event{
		AggregateID:   bEvent.AggregateID,
		AggregateType: bEvent.AggregateType,
		Version:       base.Version(bEvent.Version),
		GlobalVersion: base.Version(bEvent.GlobalVersion),
		Timestamp:     bEvent.Timestamp,
		Metadata:      bEvent.Metadata,
		Data:          eventData,
	}
	return event, nil
}
