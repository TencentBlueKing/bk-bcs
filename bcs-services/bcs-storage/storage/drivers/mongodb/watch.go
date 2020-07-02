/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package mongodb

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"gopkg.in/mgo.v2/bson"
)

type watchHandler struct {
	opts  *operator.WatchOptions
	event chan *operator.Event

	listenerName string
	dbName       string
	cName        string
	diffTree     []string
}

func newWatchHandler(opts *operator.WatchOptions, tank *mongoTank) *watchHandler {
	wh := &watchHandler{
		opts:         opts,
		listenerName: tank.listenerName,
		dbName:       tank.dbName,
		cName:        tank.cName,
	}
	if opts.MustDiff != "" {
		wh.diffTree = strings.Split(opts.MustDiff, ".")
	}
	return wh
}

func (wh *watchHandler) isDiff(op *opLog) bool {
	if op.OP == opDeleteValue {
		return true
	}
	var d interface{} = op.O
	for _, t := range wh.diffTree {
		md, ok := d.(map[string]interface{})
		if !ok {
			return false
		}
		if d = md[t]; d == nil {
			return false
		}
	}
	return true
}

func (wh *watchHandler) ns() string {
	return fmt.Sprintf("%s.%s", wh.dbName, wh.cName)
}

func (wh *watchHandler) watch() (event chan *operator.Event, cancel context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	event = make(chan *operator.Event, 1000)
	wh.event = event
	go wh.watching(ctx)
	return
}

func (wh *watchHandler) watching(pCtx context.Context) {
	isValid := true

	listenerPoolLock.RLock()
	listener := listenerPool[wh.listenerName]
	listenerPoolLock.RUnlock()
	if listener == nil {
		blog.Errorf("mongodb watching | watcherName does not exists: %s", wh.listenerName)
		isValid = false
	}

	if wh.dbName == "" || wh.cName == "" {
		blog.Errorf("mongodb watching | dbName or cName is empty")
		isValid = false
	}

	if !isValid {
		wh.event <- operator.EventWatchBreak
		return
	}

	ns := wh.ns()
	we := listener.wr.subscribe(ns)
	blog.Infof("mongodb watching | begin to watch: %s", ns)

	var ctx context.Context
	var cancel context.CancelFunc
	var op *opLog
	var eventsNumber uint

	if wh.opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(pCtx, wh.opts.Timeout)
	} else {
		ctx, cancel = context.WithCancel(pCtx) //nolint
	}

	for {
		if (wh.opts.MaxEvents > 0) && (eventsNumber >= wh.opts.MaxEvents) {
			cancel()
		}

		select {
		case <-ctx.Done():
			wh.event <- operator.EventWatchBreak
			listener.wr.unsubscribe(we)
			blog.Infof("mongodb watching | end watch: %s", ns)
			cancel()
			return //nolint
		case op = <-we.w:
			eventsNumber++
			var eventType operator.EventType
			switch op.OP {
			case opNopValue:
				continue
			case opInsertValue:
				eventType = operator.Add
			case opDeleteValue:
				eventType = operator.Del
			case opUpdateValue:
				eventType = operator.Chg
			default:
				continue
			}

			// If SelfOnly is true means the watcher only concern the change of node itself,
			// and its eventType should be EventSelfChange.
			// Others such as children change, children add, children delete will be ignored.
			if wh.opts.SelfOnly && eventType != operator.SChg {
				continue
			}

			// If MustDiff is set and the change part does not contain the keys of diffTree, then continue
			if !wh.isDiff(op) {
				continue
			}

			wh.event <- &operator.Event{Type: eventType, Value: op.RAW}
		}
	}
}

type watcher chan *opLog

type watcherElement struct {
	ns string
	w  watcher
	l  *watcherList
	e  *list.Element
}

type watcherList struct {
	*list.List
}

func (wl *watcherList) do(f func(*list.Element)) {
	if wl == nil {
		return
	}
	for e := wl.Front(); e != nil; e = e.Next() {
		f(e)
	}
}

type watcherRouter struct {
	mutexLock sync.RWMutex
	routeLock sync.RWMutex
	mutex     map[string]*sync.RWMutex
	route     map[string]*watcherList
}

func (wr *watcherRouter) getMutex(ns string) *sync.RWMutex {
	wr.mutexLock.RLock()
	mutex := wr.mutex[ns]
	wr.mutexLock.RUnlock()
	if mutex == nil {
		return wr.newMutex(ns)
	}
	return mutex
}

func (wr *watcherRouter) getWL(ns string) *watcherList {
	wr.routeLock.RLock()
	wl := wr.route[ns]
	wr.routeLock.RUnlock()
	if wl == nil {
		return wr.newWL(ns)
	}
	return wl
}

func (wr *watcherRouter) newMutex(ns string) *sync.RWMutex {
	wr.mutexLock.Lock()
	defer wr.mutexLock.Unlock()

	if wr.mutex[ns] == nil {
		wr.mutex[ns] = new(sync.RWMutex)
	}
	return wr.mutex[ns]
}

func (wr *watcherRouter) newWL(ns string) *watcherList {
	wr.routeLock.Lock()
	defer wr.routeLock.Unlock()

	if wr.route[ns] == nil {
		mutex := wr.getMutex(ns)
		mutex.Lock()
		wr.route[ns] = &watcherList{list.New()}
		mutex.Unlock()
	}
	return wr.route[ns]
}

func (wr *watcherRouter) subscribe(ns string) (we *watcherElement) {
	w := make(watcher)

	wList := wr.getWL(ns)

	mutex := wr.getMutex(ns)
	mutex.Lock()
	defer mutex.Unlock()
	return &watcherElement{
		ns: ns,
		w:  w,
		l:  wList,
		e:  wList.PushBack(w),
	}
}

func (wr *watcherRouter) unsubscribe(we *watcherElement) {
	mutex := wr.getMutex(we.ns)
	mutex.Lock()
	defer mutex.Unlock()
	we.l.Remove(we.e)
}

const (
	opLogTimestampKey = "ts"
	opNopValue        = "n"
	opInsertValue     = "i"
	opUpdateValue     = "u"
	opDeleteValue     = "d"
	opLogRetryGap     = 5 * time.Second
	opIdKey           = "_id"
)

type opLogListener struct {
	name string
	wr   *watcherRouter
	tank *mongoTank
}

func (ol *opLogListener) init(tank *mongoTank, name string) *opLogListener {
	ol.wr = &watcherRouter{route: make(map[string]*watcherList), mutex: make(map[string]*sync.RWMutex)}
	ol.tank = tank
	ol.name = name
	return ol
}

func (ol *opLogListener) listen() {
	blog.Infof(ol.sprint(fmt.Sprintf("get last timestamp of oplog")))
	tank := ol.tank

	iter := tank.Copy().Filter(operator.BaseCondition.AddOp(operator.Gt, opLogTimestampKey, ol.getLastTimestamp()).AddOp(operator.Ne, "op", opNopValue)).(*mongoTank).Tail()
	for {
		var op *opLog
		for iter.Next(&op) {

			if strings.HasSuffix(op.OP, ".$cmd") {
				continue
			}

			if !op.preProcess(ol.tank.Copy()) {
				continue
			}

			wl := ol.wr.getWL(op.NS)
			mutex := ol.wr.getMutex(op.NS)
			mutex.RLock()

			op.RAW, _ = dotRecover([]interface{}{op.RAW})[0].(map[string]interface{})
			wl.do(func(element *list.Element) {
				w, ok := element.Value.(watcher)
				if !ok {
					blog.Errorf(ol.sprint("watcher list value is not a chan *watcher"))
					return
				}
				subOp := &opLog{}
				*subOp = *op
				w <- subOp
			})
			mutex.RUnlock()
		}
		if err := iter.Err(); err != nil {
			blog.Errorf(ol.sprint("get mongodb tail Next() failed: %v"), err)
			_ = iter.Close()
			time.Sleep(opLogRetryGap)
		}
		iter = tank.Copy().Filter(operator.BaseCondition.AddOp(operator.Gt, opLogTimestampKey, ol.getLastTimestamp()).AddOp(operator.Ne, "op", opNopValue)).(*mongoTank).Tail()
	}
}

// get the last timestamp of oplog
func (ol *opLogListener) getLastTimestamp() bson.MongoTimestamp {
	tank := ol.tank.Copy().OrderBy("-" + opLogTimestampKey).Select(opLogTimestampKey).Limit(1)
	for ; ; time.Sleep(opLogRetryGap) {
		t := tank.Query()
		if err := t.GetError(); err != nil {
			blog.Errorf(ol.sprint(fmt.Sprintf("getLastTimestamp failed! %v", err)))
			continue
		}
		r := t.GetValue()
		if t.GetLen() == 0 {
			blog.Errorf(ol.sprint(fmt.Sprintf("getLastTimestamp failed! no oplog found.")))
			continue
		}
		op, err := ol.getOpLog(r[0])
		if err != nil {
			blog.Errorf(ol.sprint(fmt.Sprintf("getLastTimestamp failed! %v", err)))
			continue
		}
		return bson.MongoTimestamp(op.TS)
	}
}

// get oplog
func (ol *opLogListener) getOpLog(v interface{}) (*opLog, error) {
	vv, ok := v.(map[string]interface{})
	if !ok {
		return nil, errors.New(ol.sprint(fmt.Sprintf("getLastTimestamp failed! oplog format error: %v", vv)))
	}

	op := &opLog{}
	if err := op.fillUp(vv); err != nil {
		return nil, err
	}
	return op, nil
}

func (ol *opLogListener) sprint(s string) string {
	return fmt.Sprintf("mongodb listener | %s | %s", ol.name, s)
}

type opLog struct {
	NS   string
	OP   string
	O    map[string]interface{}
	O2   map[string]interface{}
	TS   uint64
	WALL time.Time
	H    int64
	V    int64
	RAW  map[string]interface{}
}

func (opl *opLog) fillUp(value map[string]interface{}) error {
	for k, v := range value {
		err := setField(opl, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (opl *opLog) preProcess(tank operator.Tank) bool {
	defer tank.Close()
	if opl.OP == opNopValue {
		return false
	}

	var ok bool
	if opl.O["$set"] != nil {
		opl.O, ok = opl.O["$set"].(map[string]interface{})
		if !ok {
			return false
		}
	}

	if opl.OP == opInsertValue || opl.OP == opDeleteValue {
		opl.RAW = opl.O
		return true
	}

	objectId, ok := opl.O2[opIdKey].(bson.ObjectId)
	if !ok {
		return false
	}
	ns := strings.Split(opl.NS, ".")
	if len(ns) < 2 {
		return false
	}

	defer func() {
		if r := recover(); r != nil {
			// prevent bson.ObjectIdHex() panic, return false
		}
	}()
	t := tank.Using(ns[0]).From(ns[1]).Filter(operator.BaseCondition.AddOp(operator.Eq, opIdKey, objectId)).Query()
	if t.GetError() != nil {
		return false
	}
	if v := t.GetValue(); len(v) > 0 {
		opl.RAW, ok = v[0].(map[string]interface{})
		return ok
	}
	return false
}

func setField(obj interface{}, name string, value interface{}) error {
	name = strings.ToUpper(name)
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}

var (
	listenerPool     map[string]*opLogListener
	listenerPoolLock sync.RWMutex
)

// StartWatch start storage operator unit watch
func StartWatch(tank operator.Tank, name string) {
	t, ok := tank.(*mongoTank)
	if !ok {
		return
	}

	listenerPoolLock.Lock()
	listener := new(opLogListener).init(t, name)
	if listenerPool == nil {
		listenerPool = make(map[string]*opLogListener)
	}
	if listenerPool[name] != nil {
		blog.Errorf("watcherName duplicated: %s", name)
		listenerPoolLock.Unlock()
		return
	}
	listenerPool[name] = listener
	listenerPoolLock.Unlock()
	listener.listen()
}
