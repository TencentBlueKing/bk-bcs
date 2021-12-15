/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bce

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Error implements the error interface
//
// Most methods in the SDK will return bce.Error instance.
type Error struct {
	StatusCode               int
	Code, Message, RequestID string
}

// Error returns the formatted error message.
func (err *Error) Error() string {
	return fmt.Sprintf("Error Message: \"%s\", Error Code: \"%s\", Status Code: %d, Request Id: \"%s\"",
		err.Message, err.Code, err.StatusCode, err.RequestID)
}

func buildError(resp *Response) error {
	bodyContent, err := resp.GetBodyContent()

	if err == nil {
		if bodyContent == nil || string(bodyContent) == "" {
			return errors.New("Unknown Error")
		}
		var bceError *Error
		err := json.Unmarshal(bodyContent, &bceError)
		if err != nil {
			return errors.New(string(bodyContent))
		}
		bceError.StatusCode = resp.StatusCode
		return bceError
	}

	return err
}
