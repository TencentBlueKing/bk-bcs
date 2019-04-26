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

package utils

type ClusterAreadyExistError struct {
	message string
}

func (e *ClusterAreadyExistError) Error() string {
	return e.message
}

func NewClusterAreadyExistError(message string) *ClusterAreadyExistError {
	return &ClusterAreadyExistError{
		message: message,
	}
}

type CannotCreateClusterError struct {
	message string
}

func (e *CannotCreateClusterError) Error() string {
	return e.message
}

func NewCannotCreateClusterError(message string) *CannotCreateClusterError {
	return &CannotCreateClusterError{
		message: message,
	}
}

type ClusterInitFailedError struct {
	message string
}

func (e *ClusterInitFailedError) Error() string {
	return e.message
}

func NewClusterInitFailedError(message string) *ClusterInitFailedError {
	return &ClusterInitFailedError{
		message: message,
	}
}
