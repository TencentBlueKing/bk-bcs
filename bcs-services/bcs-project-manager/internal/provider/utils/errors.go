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

// Package utils xxx
package utils

import (
	"fmt"
)

const (
	null = "nil"
)

type errNoProvider struct {
	provider string
}

func (err *errNoProvider) Error() string {
	if err == nil {
		return null
	}
	return fmt.Sprintf("no such provider %s", err.provider)
}

// NewNoProviderError create a new no provider error
func NewNoProviderError(provider string) error {
	return &errNoProvider{provider}
}

// IsNoProviderError judges error is errNoProvider
func IsNoProviderError(err error) bool {
	if _, ok := err.(*errNoProvider); ok {
		return true
	}
	return false
}

type errNotImplemented struct {
	handler string
}

func (err *errNotImplemented) Error() string {
	if err == nil {
		return null
	}
	return fmt.Sprintf("handler %s not implemented", err.handler)
}

// NewNotImplemented create a new not implemented error
func NewNotImplemented(msg string) error {
	return &errNotImplemented{msg}
}

// IsNotImplementedError judges error is errNotImplemented
func IsNotImplementedError(err error) bool {
	if _, ok := err.(*errNotImplemented); ok {
		return true
	}
	return false
}

type errServerNil struct {
	server string
}

func (err *errServerNil) Error() string {
	if err == nil {
		return null
	}
	return fmt.Sprintf("server %s is not found", err.server)
}

// NewServerNil create a new server nil error
func NewServerNil(server string) error {
	return &errServerNil{server}
}

// IsServerNilError judges error is errServerNil
func IsServerNilError(err error) bool {
	if _, ok := err.(*errServerNil); ok {
		return true
	}
	return false
}
