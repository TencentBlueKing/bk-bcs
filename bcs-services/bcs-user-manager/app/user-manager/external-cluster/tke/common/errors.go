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

package common

const (
	// NoErr flag
	NoErr         = 0
	// NoErrCodeDesc description
	NoErrCodeDesc = "Success"
)

// LegacyAPIError for LegacyAPI error
type LegacyAPIError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	CodeDesc string `json:"codeDesc"`
}

func (lae LegacyAPIError) Error() string {
	return lae.Message
}

// VersionAPIError for version api error
type VersionAPIError struct {
	Response struct {
		Error apiErrorResponse `json:"Error"`
	} `json:"Response"`
}

type apiErrorResponse struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

func (vae VersionAPIError) Error() string {
	return vae.Response.Error.Message
}

// ClientError for client error
type ClientError struct {
	Message string
}

func (ce ClientError) Error() string {
	return ce.Message
}

func makeClientError(err error) ClientError {
	return ClientError{
		Message: err.Error(),
	}
}
