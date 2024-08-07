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

package mongo

import (
	"context"
	"time"

	bcsmongo "github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	driver "go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
)

// NewMongoCli ...
func NewMongoCli(opt *bcsmongo.Options) (*driver.Client, error) {
	credential := mopt.Credential{
		AuthMechanism: opt.AuthMechanism,
		AuthSource:    opt.AuthDatabase,
		Username:      opt.Username,
		Password:      opt.Password,
		PasswordSet:   true,
	}
	if len(credential.AuthMechanism) == 0 {
		credential.AuthMechanism = "SCRAM-SHA-256"
	}
	// construct mongo client options
	mCliOpt := &mopt.ClientOptions{
		Auth:  &credential,
		Hosts: opt.Hosts,
	}
	if opt.MaxPoolSize != 0 {
		mCliOpt.MaxPoolSize = &opt.MaxPoolSize
	}
	if opt.MinPoolSize != 0 {
		mCliOpt.MinPoolSize = &opt.MinPoolSize
	}
	var timeoutDuration time.Duration
	if opt.ConnectTimeoutSeconds != 0 {
		timeoutDuration = time.Duration(opt.ConnectTimeoutSeconds) * time.Second
	}
	mCliOpt.ConnectTimeout = &timeoutDuration

	// create mongo client
	mCli, err := driver.NewClient(mCliOpt) // nolint
	if err != nil {
		return nil, err
	}
	// connect to mongo
	if err = mCli.Connect(context.TODO()); err != nil { // nolint
		return nil, err
	}

	if err = mCli.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	return mCli, nil
}
