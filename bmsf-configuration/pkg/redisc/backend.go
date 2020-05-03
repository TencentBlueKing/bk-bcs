/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package redisc

import (
	"sync/atomic"
	"time"

	redis "gopkg.in/redis.v5"
	tomb "gopkg.in/tomb.v2"
)

// backend is redis client wraper as a backend.
type backend struct {
	// real redis client.
	cli *redis.Client

	// options for redis client.
	opt *Options

	// alive.
	alive int32

	// status count.
	succs int32
	fails int32

	// redis client closer.
	closer tomb.Tomb
}

// newBackend returns a new redis backend.
func newBackend(opt *Options) *backend {
	backend := &backend{
		cli:   redis.NewClient(&opt.Options),
		opt:   opt,
		alive: 1,
	}
	backend.start()
	return backend
}

// alive returns ture if the redis client alive.
func (b *backend) isAlive() bool {
	return atomic.LoadInt32(&b.alive) > 0
}

// close closes the redis backend.
func (b *backend) close() error {
	b.closer.Kill(nil)
	return b.closer.Wait()
}

// ping send redis PING and wait for a PONG back.
func (b *backend) ping() {
	v, err := b.cli.Ping().Result()
	if err != nil {
		// update the client status.
		b.status(false)
		return
	}

	if v != pong {
		// update the client status.
		b.status(false)
		return
	}

	// success, update the client status.
	b.status(true)
}

// status updates the redis client stat count.
func (b *backend) status(ok bool) {
	if ok {
		// PING success.
		atomic.StoreInt32(&b.fails, 0)

		rise := b.opt.getRise()
		n := int(atomic.AddInt32(&b.succs, 1))

		if n > rise {
			atomic.AddInt32(&b.succs, -1)
		} else if n == rise {
			// count to rise, client alive.
			atomic.CompareAndSwapInt32(&b.alive, 0, 1)
		}
	} else {
		// PING failed.
		atomic.StoreInt32(&b.succs, 0)

		fall := b.opt.getFall()
		n := int(atomic.AddInt32(&b.fails, 1))

		if n > fall {
			atomic.AddInt32(&b.fails, -1)
		} else if n == fall {
			// count to fall, client non-alive.
			atomic.CompareAndSwapInt32(&b.alive, 1, 0)
		}
	}
}

// start starts keeping check the redis client by PING to get
// the health information of redis client.
func (b *backend) start() {
	// send redis PING at first.
	b.ping()

	// loop keep check the redis client and closer would
	// kill the redis backend.
	loop := func() error {
		for {
			select {
			case <-b.closer.Dying():
				// closing now.
				return b.cli.Close()

			case <-time.After(b.opt.getInterval()):
				// check by redis PING.
				b.ping()
			}
		}
	}

	// run the loop now.
	b.closer.Go(loop)
}
