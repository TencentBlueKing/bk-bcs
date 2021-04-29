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

package safeviper

import (
	"sync"
	"time"

	"github.com/spf13/viper"
)

// SafeViper is a coroutine safe viper common wrapper.
// We writes this on account of the spf13 viper dose not support
// coroutine safe. In some module, we need to set and get configs in viper
// in same time, and we can't find a better com to replace viper.
// Only parts of the spf13 viper funcs are rewriten here, you need to
// add more funcs if necessary.
type SafeViper struct {
	// real viper.
	viper *viper.Viper
	mu    sync.RWMutex
}

// NewSafeViper creates a new safe viper.
func NewSafeViper(viper *viper.Viper) *SafeViper {
	return &SafeViper{viper: viper}
}

// Set sets the value for the key in the override register.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func (sv *SafeViper) Set(key string, value interface{}) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.viper.Set(key, value)
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
// Get returns an interface. For a specific value use one of the Get____ methods.
func (sv *SafeViper) Get(key string) interface{} {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.Get(key)
}

// GetString returns the value associated with the key as a string.
func (sv *SafeViper) GetString(key string) string {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetString(key)
}

// GetBool returns the value associated with the key as a boolean.
func (sv *SafeViper) GetBool(key string) bool {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetBool(key)
}

// GetInt returns the value associated with the key as an integer.
func (sv *SafeViper) GetInt(key string) int {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetInt(key)
}

// GetInt32 returns the value associated with the key as an integer.
func (sv *SafeViper) GetInt32(key string) int32 {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetInt32(key)
}

// GetInt64 returns the value associated with the key as an integer.
func (sv *SafeViper) GetInt64(key string) int64 {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetInt64(key)
}

// GetUint returns the value associated with the key as an unsigned integer.
func (sv *SafeViper) GetUint(key string) uint {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetUint(key)
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func (sv *SafeViper) GetUint32(key string) uint32 {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetUint32(key)
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func (sv *SafeViper) GetUint64(key string) uint64 {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetUint64(key)
}

// GetFloat64 returns the value associated with the key as a float64.
func (sv *SafeViper) GetFloat64(key string) float64 {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetFloat64(key)
}

// GetTime returns the value associated with the key as time.
func (sv *SafeViper) GetTime(key string) time.Time {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetTime(key)
}

// GetDuration returns the value associated with the key as a duration.
func (sv *SafeViper) GetDuration(key string) time.Duration {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetDuration(key)
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func (sv *SafeViper) GetIntSlice(key string) []int {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetIntSlice(key)
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (sv *SafeViper) GetStringSlice(key string) []string {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetStringSlice(key)
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (sv *SafeViper) GetStringMap(key string) map[string]interface{} {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetStringMap(key)
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (sv *SafeViper) GetStringMapString(key string) map[string]string {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetStringMapString(key)
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (sv *SafeViper) GetStringMapStringSlice(key string) map[string][]string {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return sv.viper.GetStringMapStringSlice(key)
}
