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

package orderedmap

import (
	"container/list"
)

// Element is map element.
type Element struct {
	// Key map element key.
	Key interface{}

	// Value map element value.
	Value interface{}

	// element list.
	element *list.Element
}

type orderedMapElement struct {
	key   interface{}
	value interface{}
}

// newElement creates a new map element struct.
func newElement(element *list.Element) *Element {
	if element == nil {
		return nil
	}
	orderedMapElement := element.Value.(*orderedMapElement)

	return &Element{
		Key:     orderedMapElement.key,
		Value:   orderedMapElement.value,
		element: element,
	}
}

// Prev returns previous map element.
func (ele *Element) Prev() *Element {
	return newElement(ele.element.Prev())
}

// Next returns next map element.
func (ele *Element) Next() *Element {
	return newElement(ele.element.Next())
}

// OrderedMap is ordered map type.
type OrderedMap struct {
	// kv mapping.
	kv map[interface{}]*list.Element

	// map element list.
	ll *list.List
}

// New creates a new OrderedMap instance.
func New() *OrderedMap {
	return &OrderedMap{kv: make(map[interface{}]*list.Element), ll: list.New()}
}

// Len returns the number of kv elements in the map.
func (m *OrderedMap) Len() int {
	return len(m.kv)
}

// Get returns the value for a key in ordered map.
func (m *OrderedMap) Get(key interface{}) (interface{}, bool) {
	value, isExist := m.kv[key]
	if isExist {
		return value.Value.(*orderedMapElement).value, true
	}
	return nil, false
}

// Set sets the value for target key.
// Return true if the target key is a new element, else return false.
func (m *OrderedMap) Set(key, value interface{}) bool {
	_, isExist := m.kv[key]
	if !isExist {
		element := m.ll.PushBack(&orderedMapElement{key, value})
		m.kv[key] = element
	} else {
		m.kv[key].Value.(*orderedMapElement).value = value
	}
	return !isExist
}

// Keys returns all keys in the order set.
// It would retain the same position when replaced.
func (m *OrderedMap) Keys() (keys []interface{}) {
	keys = make([]interface{}, m.Len())

	element := m.ll.Front()
	for i := 0; element != nil; i++ {
		keys[i] = element.Value.(*orderedMapElement).key
		element = element.Next()
	}
	return keys
}

// GetElement returns the element for target key.
func (m *OrderedMap) GetElement(key interface{}) *Element {
	value, isExist := m.kv[key]
	if isExist {
		element := value.Value.(*orderedMapElement)

		return &Element{
			element: value,
			Key:     element.key,
			Value:   element.value,
		}
	}
	return nil
}

// Delete removes a key from the order map.
// It would return true if the key exist before delete action.
func (m *OrderedMap) Delete(key interface{}) (didDelete bool) {
	element, isExist := m.kv[key]
	if isExist {
		m.ll.Remove(element)
		delete(m.kv, key)
	}
	return isExist
}

// Front returns the element that the first element.
// Return nil if no elements there.
func (m *OrderedMap) Front() *Element {
	front := m.ll.Front()
	if front == nil {
		return nil
	}
	element := front.Value.(*orderedMapElement)

	return &Element{
		element: front,
		Key:     element.key,
		Value:   element.value,
	}
}

// Back returns the element that the last element.
// Return nil if no elements there.
func (m *OrderedMap) Back() *Element {
	back := m.ll.Back()
	if back == nil {
		return nil
	}
	element := back.Value.(*orderedMapElement)

	return &Element{
		element: back,
		Key:     element.key,
		Value:   element.value,
	}
}

// Copy returns a new OrderedMap with the same elements.
func (m *OrderedMap) Copy() *OrderedMap {
	newMap := New()
	for el := m.Front(); el != nil; el = el.Next() {
		newMap.Set(el.Key, el.Value)
	}
	return newMap
}
