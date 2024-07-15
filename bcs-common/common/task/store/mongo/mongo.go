package mongo

import (
	"context"
	"time"

	bcsmongo "github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	driver "go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
)

func newMongoCli(opt *bcsmongo.Options) (*driver.Client, error) {
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
