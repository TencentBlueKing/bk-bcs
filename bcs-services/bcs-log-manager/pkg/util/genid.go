package util

import "sync"

var id int64
var mutex sync.Mutex

func GenerateID() int64 {
	var ret int64
	mutex.Lock()
	ret = id
	id++
	mutex.Unlock()
	return ret
}
