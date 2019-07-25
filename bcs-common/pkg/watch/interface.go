package watch

import (
	"bk-bcs/bcs-common/pkg/meta"
)

//EventType definition for watch
type EventType string

const (
	//EventSync sync event, reserved for force synchronization
	EventSync EventType = "SYNC"
	//EventAdded add event
	EventAdded EventType = "ADDED"
	//EventUpdated updated/modified event
	EventUpdated EventType = "UPDATED"
	//EventDeleted deleted event
	EventDeleted EventType = "DELETED"
	//EventErr error event for watch, error occured, but watch still works
	EventErr EventType = "ERROR"
	//DefaultChannelBuffer buffer for watch event channel
	DefaultChannelBuffer = 128
)

//Interface define watch channel
type Interface interface {
	//stop watch channel
	Stop()
	//get watch events, if watch stopped/error, watch must close
	// channel and exit, watch user must read channel like
	// e, ok := <-channel
	WatchEvent() <-chan Event
}

//Event holding event info for data object
type Event struct {
	Type EventType   `json:"type"`
	Data meta.Object `json:"data"`
}
