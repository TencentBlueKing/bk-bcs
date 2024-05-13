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
 */

package grpclb

import (
	"fmt"

	"google.golang.org/grpc/naming"
)

// NewPseudoResolver creates a new pseudo resolver which returns fixed addrs.
func NewPseudoResolver(addrs []string) naming.Resolver {
	return &pseudoResolver{addrs}
}

// pseudoResolver is simple name resolver
type pseudoResolver struct {
	addrs []string
}

// Resolve resolve target
func (r *pseudoResolver) Resolve(target string) (naming.Watcher, error) {
	w := &pseudoWatcher{
		updatesChan: make(chan []*naming.Update, 1),
	}
	updates := []*naming.Update{}
	for _, addr := range r.addrs {
		updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
	}
	w.updatesChan <- updates
	return w, nil
}

// pseudoWatcher watcher for update event
type pseudoWatcher struct {
	updatesChan chan []*naming.Update
}

// Next get next update event
func (w *pseudoWatcher) Next() ([]*naming.Update, error) {
	us, ok := <-w.updatesChan
	if !ok {
		return nil, fmt.Errorf("error watch close")
	}
	return us, nil
}

// Close close watcher
func (w *pseudoWatcher) Close() {
	close(w.updatesChan)
}
