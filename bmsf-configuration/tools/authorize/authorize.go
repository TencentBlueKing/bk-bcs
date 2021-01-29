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

package main

import (
	"flag"
	"os"
	"time"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/middle-services/bscp-authserver/modules/auth"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
)

var (
	// db sharding manager.
	smgr *dbsharding.ShardingManager
)

var (
	// host is mysql host.
	host string

	// port is mysql port.
	port int

	// userName is mysql username.
	userName string

	// password is mysql password.
	password string

	// timeout is database read/write timeout.
	timeout time.Duration
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "Host of target bscp main database.")
	flag.IntVar(&port, "port", 3306, "Port of target bscp main database.")
	flag.StringVar(&userName, "user", "root", "User name of target bscp main database.")
	flag.StringVar(&password, "password", "123456", "Password of target bscp main database.")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Read/Write timeout of database.")
}

// connect sharding database.
func connectDB() {
	smgr = dbsharding.NewShardingMgr(&dbsharding.Config{
		DBHost:        host,
		DBPort:        port,
		DBUser:        userName,
		DBPasswd:      password,
		Size:          10,
		PurgeInterval: time.Hour,
	}, &dbsharding.DBConfigTemplate{
		ConnTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		MaxOpenConns: 50,
		MaxIdleConns: 1,
		KeepAlive:    time.Minute,
	})

	if err := smgr.Init(); err != nil {
		panic(err)
	}
}

// close sharding database.
func closeDB() {
	smgr.Close()
}

// genCompensatePolicyCmd generates compensate policy command.
func genCompensatePolicyCmd() *cobra.Command {
	// #lizard forgives
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var bizID string
	f.StringVar(&bizID, "biz_id", "", "Business id.")

	cmd := &cobra.Command{
		Use:   "CompensatePolicy",
		Short: "Compensate policy command.",
		PreRun: func(cmd *cobra.Command, args []string) {
			connectDB()
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			closeDB()
		},
		Run: func(cmd *cobra.Command, args []string) {
			// business sharding db.
			sd, err := smgr.ShardingDB(bizID)
			if err != nil {
				panic(err)
			}

			// application.
			apps := []database.App{}
			if err := sd.DB().
				Order("Fupdate_time DESC, Fid DESC").
				Where(&database.App{BizID: bizID}).
				Find(&apps).Error; err != nil {
				panic(err)
			}

			// template.
			templates := []database.ConfigTemplate{}
			if err := sd.DB().
				Order("Fupdate_time DESC, Fid DESC").
				Where(&database.ConfigTemplate{BizID: bizID}).
				Find(&templates).Error; err != nil {
				panic(err)
			}

			// compensate local auth policies.
			sd, err = smgr.ShardingDB(dbsharding.BSCPDBKEY)
			if err != nil {
				panic(err)
			}

			// application policies.
			for _, app := range apps {
				// V0: subject  V1: object  V2: action.
				localAuth := database.LocalAuth{}

				err := sd.DB().
					Where(&database.LocalAuth{V1: app.AppID}).
					Last(&localAuth).Error

				if err == nil {
					// not need to compensate.
					continue
				}

				if err != dbsharding.RECORDNOTFOUND {
					panic(err)
				}

				// missing local auth policy, need to compensate.
				if err := sd.DB().
					Create(&database.LocalAuth{
						PType: "p",
						V0:    app.Creator,
						V1:    app.AppID,
						V2:    auth.LocalAuthAction,
					}).Error; err != nil {
					panic(err)
				}
			}

			// templates policies.
			for _, template := range templates {
				// V0: subject  V1: object  V2: action.
				localAuth := database.LocalAuth{}

				err := sd.DB().
					Where(&database.LocalAuth{V1: template.TemplateID}).
					Last(&localAuth).Error

				if err == nil {
					// not need to compensate.
					continue
				}

				if err != dbsharding.RECORDNOTFOUND {
					panic(err)
				}

				// missing local auth policy, need to compensate.
				if err := sd.DB().
					Create(&database.LocalAuth{
						PType: "p",
						V0:    template.Creator,
						V1:    template.TemplateID,
						V2:    auth.LocalAuthAction,
					}).Error; err != nil {
					panic(err)
				}
			}

			// TODO compensate BKIAM policies.
		},
	}

	// flags.
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("host"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("port"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("user"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("password"))
	cmd.Flags().AddGoFlag(flag.CommandLine.Lookup("timeout"))
	cmd.Flags().AddGoFlag(f.Lookup("biz_id"))
	cmd.MarkFlagRequired("biz_id")

	return cmd
}

// genSubCmds returns sub commands.
func genSubCmds() []*cobra.Command {
	cmds := []*cobra.Command{}
	cmds = append(cmds, genCompensatePolicyCmd())
	return cmds
}

// bscp authorize tool.
func main() {
	// root command.
	rootCmd := &cobra.Command{Use: "bk-bscp-authorize-tool"}

	// sub commands.
	subCmds := genSubCmds()

	// add sub commands.
	rootCmd.AddCommand(subCmds...)

	// run root command.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
