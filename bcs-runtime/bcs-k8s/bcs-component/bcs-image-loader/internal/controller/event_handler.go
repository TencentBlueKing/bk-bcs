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
 *
 */

package controller

import (
	"context"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

var _ handler.EventHandler = &EmptyEventHandler{}

// EmptyEventHandler implements EventHandler but does nothing.
type EmptyEventHandler struct {
}

// Create is called in response to an create event
func (h *EmptyEventHandler) Create(ctx context.Context, evt event.CreateEvent, q workqueue.RateLimitingInterface) {
}

// Update is called in response to an update event
func (h *EmptyEventHandler) Update(ctx context.Context, evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
}

// Delete is called in response to a delete event
func (h *EmptyEventHandler) Delete(ctx context.Context, evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
}

// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
// external trigger request
func (h *EmptyEventHandler) Generic(ctx context.Context, evt event.GenericEvent, q workqueue.RateLimitingInterface) {
}
