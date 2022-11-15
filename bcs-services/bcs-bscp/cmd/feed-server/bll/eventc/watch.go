/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package eventc

// Watcher defines all the supported operations by watch.
// which is used to watch the resource's change events.
type Watcher interface {
	Subscribe(currentRelease uint32, currentCursorID uint32, subSpec *SubscribeSpec) (uint64, error)
	Unsubscribe(appID uint32, sn uint64, uid string)
}
