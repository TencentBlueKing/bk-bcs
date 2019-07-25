package storage

import (
	"bk-bcs/bcs-common/pkg/meta"
	"bk-bcs/bcs-common/pkg/watch"
	"errors"

	"golang.org/x/net/context"
)

var (
	//ErrNotFound define err for no data in storage
	ErrNotFound = errors.New("Data Not Found")
)

//Storage offer one common interface for serialization of meta.Object and hide
//all event storage operation
type Storage interface {
	//Create add new object data to storage, if data already exists, force update
	//and return data already exists
	//param ctx: reserved
	//param key: update data key in event storage
	//param obj: data object
	//param ttl: time-to-live in seconds, 0 means forever
	//return obj: nil or holding data if data already exists
	Create(ctx context.Context, key string, obj meta.Object, ttl int) (out meta.Object, err error)
	//Delete clean data by key, return object already delete
	//if data do not exist, return
	//param ctx: reserved
	//param key: delete data key in event storage
	Delete(ctx context.Context, key string) (obj meta.Object, err error)
	//Watch begin to watch key
	//param ctx: reserved
	//param key: specified key for watching
	//param version: specified version for watching, empty means latest version
	//param selector: filter for origin data in storage, nil means no filter
	//return watch.Interface: interface instance for watch
	Watch(ctx context.Context, key, version string, selector Selector) (watch.Interface, error)
	//Watch begin to watch all items under key directory
	//param ctx: reserved
	//param key: specified key for watching
	//param version: specified version for watching, empty means latest version
	//param selector: filter for origin data in storage, nil means no filter
	//return watch.Interface: interface instance for watch
	WatchList(ctx context.Context, key, version string, selector Selector) (watch.Interface, error)
	//Get data object by key
	//param ctx: reserved
	//param key: data key
	//param ignoreNotFound: nil object and no error when setting true
	Get(ctx context.Context, key, version string, ignoreNotFound bool) (obj meta.Object, err error)
	//List data object under key directory
	//param ctx: reserved
	//param key: data  directory
	//param ignoreNotFound: no error returns when setting true even nil object
	//param selector: filter for origin data in storage, nil means no filter
	List(ctx context.Context, key string, selector Selector) (objs []meta.Object, err error)
	//Close storage conenction, clean resource
	Close()
}
