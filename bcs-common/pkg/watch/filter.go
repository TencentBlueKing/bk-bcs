package watch

import (
	"bk-bcs/bcs-common/pkg/meta"

	"golang.org/x/net/context"
)

//SelectFunc custom function to verify how filter acts
type SelectFunc func(meta.Object) (bool, error)

//NewSelectWatch wrap watcher with filter func
func NewSelectWatch(w Interface, fn SelectFunc) Interface {
	cxt, canceler := context.WithCancel(context.Background())
	f := &SelectorWatch{
		watch:        w,
		cxt:          cxt,
		selectFn:     fn,
		stopFn:       canceler,
		eventChannel: make(chan Event, DefaultChannelBuffer),
	}
	go f.selectWatchEvent()
	return f
}

//SelectorWatch watcher wraper offer filter function to filter data object if needed
type SelectorWatch struct {
	watch        Interface          //inner watch for original data to filte
	selectFn     SelectFunc         //filter for watch
	cxt          context.Context    //context for stop
	stopFn       context.CancelFunc //stopFn for context
	eventChannel chan Event         //event channel for data already filtered
}

//Stop stop watch channel
func (fw *SelectorWatch) Stop() {
	fw.stopFn()
}

//WatchEvent get watch events
func (fw *SelectorWatch) WatchEvent() <-chan Event {
	return fw.eventChannel
}

//filterWatchEvent handler for filter
func (fw *SelectorWatch) selectWatchEvent() {
	tunnel := fw.watch.WatchEvent()
	if tunnel == nil {
		fw.watch.Stop()
		close(fw.eventChannel)
		return
	}
	defer func() {
		fw.watch.Stop()
		close(fw.eventChannel)
	}()
	for {
		select {
		case event, ok := <-tunnel:
			if !ok {
				return
			}
			matched, err := fw.selectFn(event.Data)
			if err != nil || !matched {
				continue
			}
			fw.eventChannel <- event
		case <-fw.cxt.Done():
			return
		}
	}
}
