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
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
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

// genTraceTopologyCmd generates topology trace command.
func genTraceTopologyCmd() *cobra.Command {
	// #lizard forgives
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var bizID, appID string
	f.StringVar(&bizID, "biz_id", "", "Business id.")
	f.StringVar(&appID, "app_id", "", "Application id.")

	cmd := &cobra.Command{
		Use:   "TraceTopology",
		Short: "Trace topology command.",
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

			// application config.
			configs := []database.Config{}
			if len(appID) != 0 {
				if err := sd.DB().
					Order("Fupdate_time DESC, Fid DESC").
					Where(&database.Config{BizID: bizID, AppID: appID}).
					Find(&configs).Error; err != nil {
					panic(err)
				}
			}

			// format output.
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"TemplateID", "Name", "CfgName", "CfgFpath", "User", "UserGroup", "FilePrivilege",
				"FileFormat", "FileMode", "EngineType", "State", "Creator", "CreatedAt", "LastModifyBy", "UpdatedAt",
				"Memo"})

			fmt.Printf("\nTemplate Count: %d\n", len(templates))
			for _, template := range templates {
				var line []string
				line = append(line, template.TemplateID)
				line = append(line, template.Name)
				line = append(line, template.CfgName)
				line = append(line, template.CfgFpath)
				line = append(line, template.User)
				line = append(line, template.UserGroup)
				line = append(line, template.FilePrivilege)
				line = append(line, template.FileFormat)
				line = append(line, strconv.Itoa(int(template.FileMode)))
				line = append(line, strconv.Itoa(int(template.EngineType)))
				line = append(line, strconv.Itoa(int(template.State)))
				line = append(line, template.Creator)
				line = append(line, template.CreatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, template.LastModifyBy)
				line = append(line, template.UpdatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, template.Memo)
				table.Append(line)
			}
			table.Render()

			table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"AppID", "Name", "DeployType", "State", "Creator", "CreatedAt", "LastModifyBy",
				"UpdatedAt", "Memo"})

			fmt.Printf("\nApplication Count: %d\n", len(apps))
			for _, app := range apps {
				var line []string
				line = append(line, app.AppID)
				line = append(line, app.Name)
				line = append(line, strconv.Itoa(int(app.DeployType)))
				line = append(line, strconv.Itoa(int(app.State)))
				line = append(line, app.Creator)
				line = append(line, app.CreatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, app.LastModifyBy)
				line = append(line, app.UpdatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, app.Memo)
				table.Append(line)
			}
			table.Render()

			if len(appID) != 0 {
				table = tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"CfgID", "Name", "Fpath", "User", "UserGroup", "FilePrivilege", "FileFormat",
					"FileMode", "State", "Creator", "CreatedAt", "LastModifyBy", "UpdatedAt", "Memo"})

				fmt.Printf("\nApplication[%s] Config Count: %d\n", appID, len(configs))
				for _, config := range configs {
					var line []string
					line = append(line, config.CfgID)
					line = append(line, config.Name)
					line = append(line, config.Fpath)
					line = append(line, config.User)
					line = append(line, config.UserGroup)
					line = append(line, config.FilePrivilege)
					line = append(line, config.FileFormat)
					line = append(line, strconv.Itoa(int(config.FileMode)))
					line = append(line, strconv.Itoa(int(config.State)))
					line = append(line, config.Creator)
					line = append(line, config.CreatedAt.Format("2006-01-02 15:04:05"))
					line = append(line, config.LastModifyBy)
					line = append(line, config.UpdatedAt.Format("2006-01-02 15:04:05"))
					line = append(line, config.Memo)
					table.Append(line)
				}
				table.Render()
			}
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
	cmd.Flags().AddGoFlag(f.Lookup("app_id"))

	return cmd
}

// genTraceMultiCommitCmd generates multi commit trace command.
func genTraceMultiCommitCmd() *cobra.Command {
	// #lizard forgives
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var bizID, multiCommitID, commitID string
	f.StringVar(&bizID, "biz_id", "", "Business id.")
	f.StringVar(&multiCommitID, "multi_commit_id", "", "Multi commit id.")
	f.StringVar(&commitID, "commit_id", "", "Commit id.")

	cmd := &cobra.Command{
		Use:   "TraceMultiCommit",
		Short: "Trace multi commit command.",
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

			multiCommit := database.MultiCommit{}
			if err := sd.DB().
				Where(&database.MultiCommit{BizID: bizID, MultiCommitID: multiCommitID}).
				Last(&multiCommit).Error; err != nil {
				panic(err)
			}

			commits := []database.Commit{}
			if err := sd.DB().
				Order("Fupdate_time DESC, Fid DESC").
				Where(&database.Commit{BizID: bizID, MultiCommitID: multiCommitID}).
				Find(&commits).Error; err != nil {
				panic(err)
			}

			// format output.
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"MultiCommitID", "AppID", "MultiReleaseID", "Operator", "State", "CreatedAt",
				"UpdatedAt", "Memo"})

			fmt.Printf("\nMultiCommit:\n")
			var line []string
			line = append(line, multiCommit.MultiCommitID)
			line = append(line, multiCommit.AppID)
			line = append(line, multiCommit.MultiReleaseID)
			line = append(line, multiCommit.Operator)
			line = append(line, strconv.Itoa(int(multiCommit.State)))
			line = append(line, multiCommit.CreatedAt.Format("2006-01-02 15:04:05"))
			line = append(line, multiCommit.UpdatedAt.Format("2006-01-02 15:04:05"))
			line = append(line, multiCommit.Memo)
			table.Append(line)
			table.Render()

			table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"CommitID", "AppID", "CfgID", "CommitMode", "ReleaseID", "MultiCommitID",
				"CfgName", "CfgFpath", "User", "UserGroup", "FilePrivilege", "FileFormat", "FileMode", "Operator",
				"State", "CreatedAt", "UpdatedAt", "Memo"})

			fmt.Printf("\nSub Commit Count: %d\n", len(commits))
			for _, commit := range commits {
				config := database.Config{}
				if err := sd.DB().
					Where(&database.Config{BizID: bizID, CfgID: commit.CfgID}).
					Last(&config).Error; err != nil {
					panic(err)
				}

				var line []string
				line = append(line, commit.CommitID)
				line = append(line, commit.AppID)
				line = append(line, commit.CfgID)
				line = append(line, strconv.Itoa(int(commit.CommitMode)))
				line = append(line, commit.ReleaseID)
				line = append(line, commit.MultiCommitID)
				line = append(line, config.Name)
				line = append(line, config.Fpath)
				line = append(line, config.User)
				line = append(line, config.UserGroup)
				line = append(line, config.FilePrivilege)
				line = append(line, config.FileFormat)
				line = append(line, strconv.Itoa(int(config.FileMode)))
				line = append(line, commit.Operator)
				line = append(line, strconv.Itoa(int(commit.State)))
				line = append(line, commit.CreatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, commit.UpdatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, commit.Memo)
				table.Append(line)
			}
			table.Render()

			if len(commitID) != 0 {
				contentList := []database.Content{}
				if err := sd.DB().
					Order("Fupdate_time DESC, Fid DESC").
					Where(&database.Content{BizID: bizID, CommitID: commitID}).
					Find(&contentList).Error; err != nil {
					panic(err)
				}

				table = tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"CommitID", "AppID", "CfgID", "ContentID", "ContentSize", "Index", "State",
					"Creator", "CreatedAt", "LastModifyBy", "UpdatedAt", "Memo"})

				fmt.Printf("\nContents of Target Commit:\n")
				for _, content := range contentList {
					var line []string
					line = append(line, content.CommitID)
					line = append(line, content.AppID)
					line = append(line, content.CfgID)
					line = append(line, content.ContentID)
					line = append(line, strconv.Itoa(int(content.ContentSize)))
					line = append(line, content.Index)
					line = append(line, strconv.Itoa(int(content.State)))
					line = append(line, content.Creator)
					line = append(line, content.CreatedAt.Format("2006-01-02 15:04:05"))
					line = append(line, content.LastModifyBy)
					line = append(line, content.UpdatedAt.Format("2006-01-02 15:04:05"))
					line = append(line, content.Memo)
					table.Append(line)
				}
				table.Render()
			}
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
	cmd.Flags().AddGoFlag(f.Lookup("multi_commit_id"))
	cmd.MarkFlagRequired("multi_commit_id")
	cmd.Flags().AddGoFlag(f.Lookup("commit_id"))

	return cmd
}

// genTraceMultiReleaseCmd generates multi release trace command.
func genTraceMultiReleaseCmd() *cobra.Command {
	// #lizard forgives
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	var bizID, multiReleaseID, releaseID, strategyID string
	f.StringVar(&bizID, "biz_id", "", "Business id.")
	f.StringVar(&multiReleaseID, "multi_release_id", "", "Multi release id.")
	f.StringVar(&releaseID, "release_id", "", "Release id.")
	f.StringVar(&strategyID, "strategy_id", "", "Strategy id.")

	cmd := &cobra.Command{
		Use:   "TraceMultiRelease",
		Short: "Trace multi release command.",
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

			multiRelease := database.MultiRelease{}
			if err := sd.DB().
				Where(&database.MultiRelease{BizID: bizID, MultiReleaseID: multiReleaseID}).
				Last(&multiRelease).Error; err != nil {
				panic(err)
			}

			multiCommit := database.MultiCommit{}
			if err := sd.DB().
				Where(&database.MultiCommit{BizID: bizID, MultiCommitID: multiRelease.MultiCommitID}).
				Last(&multiCommit).Error; err != nil {
				panic(err)
			}

			// format output.
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"MultiReleaseID", "Name", "AppID", "MultiCommitID", "StrategyID", "Strategies",
				"State", "Creator", "CreatedAt", "LastModifyBy", "UpdatedAt", "Memo"})

			fmt.Printf("\nMultiRelease:\n")
			var line []string
			line = append(line, multiRelease.MultiReleaseID)
			line = append(line, multiRelease.Name)
			line = append(line, multiRelease.AppID)
			line = append(line, multiRelease.MultiCommitID)
			line = append(line, multiRelease.StrategyID)
			line = append(line, multiRelease.Strategies)
			line = append(line, strconv.Itoa(int(multiCommit.State)))
			line = append(line, multiRelease.Creator)
			line = append(line, multiCommit.CreatedAt.Format("2006-01-02 15:04:05"))
			line = append(line, multiRelease.LastModifyBy)
			line = append(line, multiCommit.UpdatedAt.Format("2006-01-02 15:04:05"))
			line = append(line, multiCommit.Memo)
			table.Append(line)
			table.Render()

			table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"MultiCommitID", "AppID", "MultiReleaseID", "Operator", "State", "CreatedAt",
				"UpdatedAt", "Memo"})

			fmt.Printf("\nMultiCommit:\n")
			line = []string{}
			line = append(line, multiCommit.MultiCommitID)
			line = append(line, multiCommit.AppID)
			line = append(line, multiCommit.MultiReleaseID)
			line = append(line, multiCommit.Operator)
			line = append(line, strconv.Itoa(int(multiCommit.State)))
			line = append(line, multiCommit.CreatedAt.Format("2006-01-02 15:04:05"))
			line = append(line, multiCommit.UpdatedAt.Format("2006-01-02 15:04:05"))
			line = append(line, multiCommit.Memo)
			table.Append(line)
			table.Render()

			releases := []database.Release{}
			if err := sd.DB().
				Where(&database.Release{BizID: bizID, MultiReleaseID: multiReleaseID}).
				Find(&releases).Error; err != nil {
				panic(err)
			}

			table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ReleaseID", "Name", "AppID", "CommitID", "CfgID", "CfgName", "CfgFpath", "User",
				"UserGroup", "FilePrivilege", "FileFormat", "FileMode", "StrategyID", "Strategies", "MultiReleaseID",
				"State", "Creator", "CreatedAt", "LastModifyBy", "UpdatedAt", "Memo"})

			fmt.Printf("\nSub Release Count: %d\n", len(releases))
			for _, release := range releases {
				var line []string
				line = append(line, release.ReleaseID)
				line = append(line, release.Name)
				line = append(line, release.AppID)
				line = append(line, release.CommitID)
				line = append(line, release.CfgID)
				line = append(line, release.CfgName)
				line = append(line, release.CfgFpath)
				line = append(line, release.User)
				line = append(line, release.UserGroup)
				line = append(line, release.FilePrivilege)
				line = append(line, release.FileFormat)
				line = append(line, strconv.Itoa(int(release.FileMode)))
				line = append(line, release.StrategyID)
				line = append(line, release.Strategies)
				line = append(line, release.MultiReleaseID)
				line = append(line, strconv.Itoa(int(release.State)))
				line = append(line, release.Creator)
				line = append(line, release.CreatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, release.LastModifyBy)
				line = append(line, release.UpdatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, release.Memo)
				table.Append(line)
			}
			table.Render()

			if len(strategyID) != 0 {
				strategy := database.Strategy{}
				if err := sd.DB().
					Where(&database.Strategy{BizID: bizID, StrategyID: strategyID}).
					Last(&strategy).Error; err != nil {
					panic(err)
				}

				table = tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"StrategyID", "Name", "AppID", "Content", "State", "Creator", "CreatedAt",
					"LastModifyBy", "UpdatedAt", "Memo"})

				fmt.Printf("\nTarget Strategy:\n")
				var line []string
				line = append(line, strategy.StrategyID)
				line = append(line, strategy.Name)
				line = append(line, strategy.AppID)
				line = append(line, strategy.Content)
				line = append(line, strconv.Itoa(int(strategy.State)))
				line = append(line, strategy.Creator)
				line = append(line, strategy.CreatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, strategy.LastModifyBy)
				line = append(line, strategy.UpdatedAt.Format("2006-01-02 15:04:05"))
				line = append(line, strategy.Memo)
				table.Append(line)
				table.Render()
			}

			if len(releaseID) != 0 &&
				(multiRelease.State == int32(pbcommon.ReleaseState_RS_PUBLISHED) ||
					multiRelease.State == int32(pbcommon.ReleaseState_RS_ROLLBACKED)) {

				release := database.Release{}
				if err := sd.DB().
					Where(&database.Release{BizID: bizID, ReleaseID: releaseID}).
					Last(&release).Error; err != nil {
					panic(err)
				}
				if release.MultiReleaseID != multiReleaseID {
					return
				}

				appInstanceReleases := []database.AppInstanceRelease{}
				if err := sd.DB().
					Order("Fupdate_time DESC, Fid DESC").
					Where(&database.AppInstanceRelease{BizID: bizID, ReleaseID: releaseID}).
					Find(&appInstanceReleases).Error; err != nil {
					panic(err)
				}

				table = tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"InstanceID", "CloudID", "IP", "Path", "Labels", "EffectTime", "EffectCode",
					"EffectMsg", "ReloadTime", "ReloadCode", "ReloadMsg", "CreatedAt", "UpdatedAt"})

				fmt.Printf("\nEffect AppInstance:\n")
				for _, appInstanceRelease := range appInstanceReleases {
					appInstance := database.AppInstance{}
					if err := sd.DB().
						Where(&database.AppInstance{BizID: bizID, ID: appInstanceRelease.InstanceID}).
						Last(&appInstance).Error; err != nil {
						panic(err)
					}

					var line []string
					line = append(line, strconv.Itoa(int(appInstanceRelease.InstanceID)))
					line = append(line, appInstance.CloudID)
					line = append(line, appInstance.IP)
					line = append(line, appInstance.Path)
					line = append(line, appInstance.Labels)
					if appInstanceRelease.EffectTime != nil {
						line = append(line, appInstanceRelease.EffectTime.Format("2006-01-02 15:04:05"))
					}
					line = append(line, strconv.Itoa(int(appInstanceRelease.EffectCode)))
					line = append(line, appInstanceRelease.EffectMsg)
					if appInstanceRelease.ReloadTime != nil {
						line = append(line, appInstanceRelease.ReloadTime.Format("2006-01-02 15:04:05"))
					}
					line = append(line, strconv.Itoa(int(appInstanceRelease.ReloadCode)))
					line = append(line, appInstanceRelease.ReloadMsg)
					line = append(line, appInstanceRelease.CreatedAt.Format("2006-01-02 15:04:05"))
					line = append(line, appInstanceRelease.UpdatedAt.Format("2006-01-02 15:04:05"))
					table.Append(line)
				}
				table.Render()
			}
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
	cmd.Flags().AddGoFlag(f.Lookup("multi_release_id"))
	cmd.MarkFlagRequired("multi_release_id")
	cmd.Flags().AddGoFlag(f.Lookup("release_id"))
	cmd.Flags().AddGoFlag(f.Lookup("strategy_id"))

	return cmd
}

// genSubCmds returns sub commands.
func genSubCmds() []*cobra.Command {
	cmds := []*cobra.Command{}
	cmds = append(cmds, genTraceTopologyCmd())
	cmds = append(cmds, genTraceMultiCommitCmd())
	cmds = append(cmds, genTraceMultiReleaseCmd())
	return cmds
}

// bscp trace tool.
func main() {
	// root command.
	rootCmd := &cobra.Command{Use: "bk-bscp-trace-tool"}

	// sub commands.
	subCmds := genSubCmds()

	// add sub commands.
	rootCmd.AddCommand(subCmds...)

	// run root command.
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
