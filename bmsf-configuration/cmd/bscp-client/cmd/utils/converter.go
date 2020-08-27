/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sergi/go-diff/diffmatchpatch"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
)

//PrintBusiness print business list to ouput
func PrintBusiness(business *common.Business) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", business.Name)
	lineColor.Print("BusinessID: ")
	fmt.Printf("\t%s\n", business.Bid)
	lineColor.Print("Department: ")
	fmt.Printf("\t%s\n", business.Depid)
	lineColor.Print("State:\t\t")
	if business.State == 0 {
		color.Cyan("Affectived\n")
	} else {
		color.Red("Deleted\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", business.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", business.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", business.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", business.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n\n", business.UpdatedAt)
}

//PrintBusiness print business list to ouput
func PrintBusinesses(businesses []*common.Business) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Department", "Creator", "State"})
	table.SetHeaderColor(setTitleColorToRed(5)...)
	for _, business := range businesses {
		var line []string
		line = append(line, business.Bid)
		line = append(line, business.Name)
		line = append(line, business.Depid)
		line = append(line, business.Creator)
		line = append(line, strconv.Itoa(int(business.State)))
		table.Append(line)
	}
	table.Render()
}

//PrintShardingDB print business list to ouput
func PrintShardingDBList(shardingDBs []*common.ShardingDB) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Host", "Port", "User", "State", "LastUpdated"})
	table.SetHeaderColor(setTitleColorToRed(6)...)
	for _, db := range shardingDBs {
		var line []string
		line = append(line, db.Dbid)
		line = append(line, db.Host)
		line = append(line, strconv.Itoa(int(db.Port)))
		line = append(line, db.User)
		line = append(line, strconv.Itoa(int(db.State)))
		line = append(line, db.UpdatedAt)
		table.Append(line)
	}
	table.Render()
}

// PrintShardingDB print list to ouput
func PrintShardingDB(shardingDB *common.ShardingDB) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("DBID: ")
	fmt.Printf("\t\t%s\n", shardingDB.Dbid)
	lineColor.Print("Host: ")
	fmt.Printf("\t\t%s\n", shardingDB.Host)
	lineColor.Print("Port: ")
	fmt.Printf("\t\t%s\n", strconv.Itoa(int(shardingDB.Port)))
	lineColor.Print("User: ")
	fmt.Printf("\t\t%s\n", shardingDB.User)
	lineColor.Print("Password: ")
	fmt.Printf("\t%s\n", shardingDB.Password)
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", shardingDB.Memo)
	lineColor.Print("State:\t\t")
	if shardingDB.State == 0 {
		color.Cyan("Affectived\n")
	} else {
		color.Red("Deleted\n")
	}
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", shardingDB.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n\n", shardingDB.UpdatedAt)
}

// PrintSharding print list to ouput
func PrintSharding(sharding *common.Sharding) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Key: ")
	fmt.Printf("\t\t%s\n", sharding.Key)
	lineColor.Print("DBID: ")
	fmt.Printf("\t\t%s\n", sharding.Dbid)

	lineColor.Print("DBName: ")
	fmt.Printf("\t%s\n", sharding.Dbname)
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", sharding.Memo)
	lineColor.Print("State:\t\t")
	if sharding.State == 0 {
		color.Cyan("Affectived\n")
	} else {
		color.Red("Deleted\n")
	}
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", sharding.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n\n", sharding.UpdatedAt)
}

//PrintApp print app list to ouput
func PrintAppList(apps []*common.App, business string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Type", "State", "Business"})
	table.SetHeaderColor(setTitleColorToRed(5)...)
	for _, app := range apps {
		var line []string
		line = append(line, app.Appid)
		line = append(line, app.Name)
		if app.DeployType == 0 {
			line = append(line, "container")
		} else {
			line = append(line, "process")
		}
		if app.State == 0 {
			line = append(line, "AFFECTIVED")
		} else {
			line = append(line, "DELETE")
		}
		line = append(line, business)
		table.Append(line)
	}
	table.Render()
}

//PrintApp print app to ouput
func PrintApplication(business string, app *common.App) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", app.Name)
	lineColor.Print("ApplicationID: ")
	fmt.Printf("\t%s\n", app.Appid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s - %s\n", app.Bid, business)
	lineColor.Print("DeployType: \t")
	if app.DeployType == 0 {
		fmt.Printf("Container\n")
	} else {
		fmt.Printf("process\n")
	}
	lineColor.Print("State:\t\t")
	if app.State == 0 {
		color.Cyan("Affectived\n")
	} else {
		color.Red("Deleted\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", app.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", app.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", app.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", app.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n\n", app.UpdatedAt)
}

//PrintApp print cluster to ouput
func PrintCluster(business, app string, cluster *common.Cluster) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", cluster.Name)
	lineColor.Print("ClusterID: ")
	fmt.Printf("\t%s\n", cluster.Clusterid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s - %s\n", cluster.Bid, business)
	lineColor.Print("Application: ")
	fmt.Printf("\t%s - %s\n", cluster.Appid, app)
	lineColor.Print("RClusterid: ")
	fmt.Printf("\t%s\n", cluster.RClusterid)
	lineColor.Print("State:\t\t")
	if cluster.State == 0 {
		color.Cyan("Affectived\n")
	} else {
		color.Red("Deleted\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", cluster.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", cluster.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", cluster.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", cluster.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n\n", cluster.UpdatedAt)
}

//PrintApp print zone to ouput
func PrintZone(business, app, cluster string, zone *common.Zone) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", zone.Name)
	lineColor.Print("ZoneID: ")
	fmt.Printf("\t%s\n", zone.Zoneid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s - %s\n", zone.Bid, business)
	lineColor.Print("Application: ")
	fmt.Printf("\t%s - %s\n", zone.Appid, app)
	lineColor.Print("Cluster: ")
	fmt.Printf("\t%s - %s\n", zone.Clusterid, cluster)
	lineColor.Print("State:\t\t")
	if zone.State == 0 {
		color.Cyan("Affectived\n")
	} else {
		color.Red("Deleted\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", zone.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", zone.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", zone.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", zone.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n\n", zone.UpdatedAt)
}

//PrintClusters print cluster list to ouput
func PrintClusters(clusters []*common.Cluster, business, app string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "State", "Business", "App", "RClusterid"})
	table.SetHeaderColor(setTitleColorToRed(6)...)
	for _, cluster := range clusters {
		var line []string
		line = append(line, cluster.Clusterid)
		line = append(line, cluster.Name)
		if cluster.State == 0 {
			line = append(line, "AFFECTIVED")
		} else {
			line = append(line, "DELETE")
		}
		line = append(line, business)
		line = append(line, app)
		line = append(line, cluster.RClusterid)
		table.Append(line)
	}
	table.Render()
}

//PrintZones print zone list to ouput
func PrintZones(zones []*common.Zone, business, app, cluster string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "State", "Business", "App", "Cluster"})
	table.SetHeaderColor(setTitleColorToRed(6)...)
	for _, zone := range zones {
		var line []string
		line = append(line, zone.Zoneid)
		line = append(line, zone.Name)
		if zone.State == 0 {
			line = append(line, "AFFECTIVED")
		} else {
			line = append(line, "DELETE")
		}
		line = append(line, business)
		line = append(line, app)
		line = append(line, cluster)
		table.Append(line)
	}
	table.Render()
}

//PrintConfigSet print zone list to ouput
func PrintConfigSet(configSets []*common.ConfigSet, business *common.Business, app *common.App) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Fpath", "Name", "State", "Business", "App"})
	table.SetHeaderColor(setTitleColorToRed(6)...)
	for _, cfg := range configSets {
		var line []string
		line = append(line, cfg.Cfgsetid)
		line = append(line, cfg.Fpath)
		line = append(line, cfg.Name)
		if cfg.State == 0 {
			line = append(line, "AFFECTIVED")
		} else {
			line = append(line, "DELETE")
		}
		line = append(line, business.Name)
		line = append(line, app.Name)
		table.Append(line)
	}
	table.Render()
}

//PrintMultiCommits print commit list to ouput
func PrintMultiCommits(multiCommits []*common.MultiCommit, business *common.Business, app *common.App) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"CommitID", "ReleaseID", "State", "Memo", "Business", "App", "Creator"})
	table.SetHeaderColor(setTitleColorToRed(7)...)
	for _, multiCommit := range multiCommits {
		var line []string
		line = append(line, multiCommit.MultiCommitid)
		if len(multiCommit.MultiReleaseid) != 0 {
			line = append(line, multiCommit.MultiReleaseid)
		} else {
			line = append(line, "Not Released")
		}
		if multiCommit.State == 0 {
			line = append(line, "INIT")
		} else if multiCommit.State == 1 {
			line = append(line, "CONFIRMED")
		} else {
			line = append(line, "CANCELED")
		}
		line = append(line, multiCommit.Memo)
		line = append(line, business.Name)
		line = append(line, app.Name)
		line = append(line, multiCommit.Operator)
		table.Append(line)
	}
	table.Render()
}

func setTitleColorToRed(column int) []tablewriter.Colors {
	var colors []tablewriter.Colors
	for ; column > 0; column-- {
		colors = append(colors, tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold})
	}
	return colors
}

//PrintMultiCommits print multi-commit list to ouput
func PrintCommitMetaData(operator *service.AccessOperator, commitMetaData []*common.CommitMetadata, bid, appid string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ModuleId", "Cfgset", "Template", "TemplateRule"})
	table.SetHeaderColor(setTitleColorToRed(4)...)
	for _, c := range commitMetaData {
		var line []string
		line = append(line, c.Commitid)
		cfgset, _ := GetConfigSet(operator, &common.ConfigSet{Bid: bid, Appid: appid, Cfgsetid: c.Cfgsetid})
		line = append(line, path.Clean(cfgset.Fpath+"/"+cfgset.Name))
		line = append(line, c.Template)
		line = append(line, c.TemplateRule)
		table.Append(line)
	}
	table.Render()
}

//PrintMultiCommits print multi-commit list to ouput
func PrintMultiReleaseMetadatas(operator *service.AccessOperator, releaseMetaData []*common.ReleaseMetadata, bid, appid string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ModuleId", "Cfgset", "CommitModuleId"})
	table.SetHeaderColor(setTitleColorToRed(3)...)
	for _, r := range releaseMetaData {
		var line []string
		line = append(line, r.Releaseid)
		cfgset, _ := GetConfigSet(operator, &common.ConfigSet{Bid: bid, Appid: appid, Cfgsetid: r.Cfgsetid})
		line = append(line, path.Clean(cfgset.Fpath+"/"+cfgset.Name))
		line = append(line, r.Commitid)
		table.Append(line)
	}
	table.Render()
}

//PrintDetailsCommit print commit list to ouput
func PrintDetailsCommit(commit *common.Commit, business, app, cfgset string) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("ModuleID: ")
	fmt.Printf("\t%s\n", commit.Commitid)
	lineColor.Print("CommitID: ")
	fmt.Printf("\t%s\n", commit.MultiCommitid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s  -  %s\n", commit.Bid, business)
	lineColor.Print("Application: ")
	fmt.Printf("\t%s  -  %s\n", commit.Appid, app)
	lineColor.Print("ConfigSet: ")
	fmt.Printf("\t%s  -  %s\n", commit.Cfgsetid, cfgset)
	lineColor.Print("ReleaseID: ")
	if len(commit.Releaseid) != 0 {
		fmt.Printf("\t%s\n", commit.Releaseid)
	} else {
		fmt.Println("\tNot Released")
	}
	lineColor.Print("State:\t\t")
	if commit.State == 0 {
		color.Red("INIT\n")
	} else if commit.State == 1 {
		color.Cyan("CONFIRMED\n")
	} else {
		color.Magenta("CANCELED\n")
	}
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", commit.Operator)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", commit.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n", commit.UpdatedAt)

	//lineColor.Println("PrevConfigs:")
	//prevLines := strings.Split(string(commit.PrevConfigs), "\n")
	//for _, line := range prevLines {
	//	fmt.Printf("    %s\n", line)
	//}
	lineColor.Println("Configs:")
	cfgLines := strings.Split(string(commit.Configs), "\n")
	for _, line := range cfgLines {
		fmt.Printf("    %s\n", line)
	}
	//lineColor.Println("Changes:")
	//fmt.Printf("%s\n", commit.Changes)
}

//PrintDetailsMultiCommit print commit list to ouput
func PrintDetailsMultiCommit(multiCommit *common.MultiCommit, business, app string) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("CommitID: ")
	fmt.Printf("\t%s\n", multiCommit.MultiCommitid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s  -  %s\n", multiCommit.Bid, business)
	lineColor.Print("Application: ")
	fmt.Printf("\t%s  -  %s\n", multiCommit.Appid, app)
	lineColor.Print("ReleaseID: ")
	if len(multiCommit.MultiReleaseid) != 0 {
		fmt.Printf("\t%s\n", multiCommit.MultiReleaseid)
	} else {
		fmt.Println("\tNot Released")
	}
	lineColor.Print("State:\t\t")
	if multiCommit.State == 0 {
		color.Red("INIT\n")
	} else if multiCommit.State == 1 {
		color.Cyan("CONFIRMED\n")
	} else {
		color.Magenta("CANCELED\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", multiCommit.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", multiCommit.Operator)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", multiCommit.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n", multiCommit.UpdatedAt)
	lineColor.Print("Metadatas: \n")
}

//PrintRelease print commit list to ouput
func PrintRelease(release *common.Release, business, app, cfgset, strategy string) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", release.Name)
	lineColor.Print("ModuleId: ")
	fmt.Printf("\t%s\n", release.Releaseid)
	lineColor.Print("ReleaseID: ")
	fmt.Printf("\t%s\n", release.MultiReleaseid)
	lineColor.Print("CommitModuleID: ")
	fmt.Printf("%s\n", release.Commitid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s  -  %s\n", release.Bid, business)
	lineColor.Print("Application: ")
	fmt.Printf("\t%s  -  %s\n", release.Appid, app)
	lineColor.Print("ConfigSet: ")
	fmt.Printf("\t%s  -  %s\n", release.Cfgsetid, cfgset)
	lineColor.Print("Strategy: ")
	if len(release.Strategyid) != 0 {
		fmt.Printf("\t%s  -  %s\n", release.Strategyid, strategy)
	} else {
		fmt.Printf("\tNo Strategy!\n")
	}
	lineColor.Print("State:\t\t")
	if release.State == 0 {
		color.Red("INIT\n")
	} else {
		color.Cyan("PUBLISHED\n")
	}
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", release.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", release.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", release.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n", release.UpdatedAt)
}

//PrintDetailsCommit print commit list to ouput
func PrintDetailsCfgset(cfgset *common.ConfigSet, business, app string) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("ConfigSetId: ")
	fmt.Printf("\t%s\n", cfgset.Cfgsetid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s - %s\n", cfgset.Bid, business)
	lineColor.Print("Application: ")
	fmt.Printf("\t%s - %s\n", cfgset.Appid, app)
	lineColor.Print("Fpath: ")
	fmt.Printf("\t\t%s\n", cfgset.Fpath)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", cfgset.Name)
	lineColor.Print("State:\t\t")
	if cfgset.State == 0 {
		color.Cyan("Affectived\n")
	} else {
		color.Red("Deleted\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t%s\n", cfgset.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", cfgset.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", cfgset.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", cfgset.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n\n", cfgset.UpdatedAt)
}

//PrintInstanceList print instance list to ouput
func PrintInstanceList(instances []*common.AppInstance) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Instanceid", "Appid", "Clusterid", "Zoneid", "Dc", "IP", "Labels", "State", "CreatedAt", "UpdatedAt"})
	table.SetHeaderColor(setTitleColorToRed(10)...)
	for _, instance := range instances {
		var line []string
		line = append(line, string(instance.Instanceid))
		line = append(line, instance.Appid)
		line = append(line, instance.Clusterid)
		line = append(line, instance.Zoneid)
		line = append(line, instance.Dc)
		line = append(line, instance.IP)
		line = append(line, instance.Labels)
		if instance.State == 0 {
			line = append(line, "OFFLINE")
		} else {
			line = append(line, "ONLINE")
		}
		line = append(line, instance.CreatedAt)
		line = append(line, instance.UpdatedAt)
		table.Append(line)
	}
	table.Render()
}

//PrintStrategies print commit list to ouput
func PrintStrategies(strategies []*common.Strategy, app *common.App) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "State", "Memo", "App", "Creator"})
	table.SetHeaderColor(setTitleColorToRed(6)...)
	for _, c := range strategies {
		var line []string
		line = append(line, c.Strategyid)
		line = append(line, c.Name)
		if c.State == 0 {
			line = append(line, "AFFECTIVED")
		} else {
			line = append(line, "DELETE")
		}
		line = append(line, c.Memo)
		line = append(line, app.Name)
		line = append(line, c.Creator)
		table.Append(line)
	}
	table.Render()
}

//PrintStrategy print commit list to ouput
func PrintStrategy(strategy *common.Strategy, app *common.App) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", strategy.Name)
	lineColor.Print("StrategyID: ")
	fmt.Printf("\t%s\n", strategy.Strategyid)
	lineColor.Print("App: ")
	fmt.Printf("\t\t%s - %s\n", strategy.Appid, app.Name)
	lineColor.Print("Status:\t\t")
	if strategy.State == 0 {
		color.Cyan("AFFECTIVED\n")
	} else {
		color.Red("DELETE\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", strategy.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", strategy.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", strategy.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", strategy.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n", strategy.UpdatedAt)
	lineColor.Print("Content:\n")
	printJsonFormat([]byte(strategy.Content))
}

//PrintMultiReleases print commit list to ouput
func PrintMultiReleases(operator *service.AccessOperator, mreleases []*common.MultiRelease, business *common.Business, app *common.App) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "State", "Strategy", "Business", "App", "Creator"})
	table.SetHeaderColor(setTitleColorToRed(7)...)
	for _, mrelease := range mreleases {
		var line []string
		line = append(line, mrelease.MultiReleaseid)
		line = append(line, mrelease.Name)
		if mrelease.State == 0 {
			line = append(line, "INIT")
		} else if mrelease.State == 1 {
			line = append(line, "PUBLISHED")
		} else if mrelease.State == 2 {
			line = append(line, "CANCELED")
		} else {
			line = append(line, "ROLLBACKED")
		}
		strategy, _ := operator.GetStrategyById(context.TODO(), mrelease.Strategyid)
		if strategy != nil && len(strategy.Strategyid) != 0 {
			line = append(line, strategy.Name)
		} else {
			line = append(line, "No Strategy!")
		}
		line = append(line, business.Name)
		line = append(line, app.Name)
		line = append(line, mrelease.Creator)
		table.Append(line)
	}
	table.Render()
}

//PrintMultiRelease print commit list to ouput
func PrintMultiRelease(release *common.MultiRelease, business, app, strategy string) {
	//format output
	lineColor := color.New(color.FgYellow)
	lineColor.Print("Name: ")
	fmt.Printf("\t\t%s\n", release.Name)
	lineColor.Print("ReleaseID: ")
	fmt.Printf("\t%s\n", release.MultiReleaseid)
	lineColor.Print("CommitID: ")
	fmt.Printf("\t%s\n", release.MultiCommitid)
	lineColor.Print("Business: ")
	fmt.Printf("\t%s  -  %s\n", release.Bid, business)
	lineColor.Print("Application: ")
	fmt.Printf("\t%s  -  %s\n", release.Appid, app)
	lineColor.Print("Strategy: ")
	if len(release.Strategyid) != 0 {
		fmt.Printf("\t%s  -  %s\n", release.Strategyid, strategy)
	} else {
		fmt.Printf("\tNo Strategy!\n")
	}
	lineColor.Print("State:\t\t")
	if release.State == 0 {
		color.Red("INIT\n")
	} else if release.State == 1 {
		color.Cyan("PUBLISHED\n")
	} else if release.State == 2 {
		color.Magenta("CANCELED\n")
	} else {
		color.Cyan("ROLLBACKED\n")
	}
	lineColor.Print("Memo: ")
	fmt.Printf("\t\t%s\n", release.Memo)
	lineColor.Print("Creator: ")
	fmt.Printf("\t%s\n", release.Creator)
	lineColor.Print("LastModifyBy: ")
	fmt.Printf("\t%s\n", release.LastModifyBy)
	lineColor.Print("CreatedAt: ")
	fmt.Printf("\t%s\n", release.CreatedAt)
	lineColor.Print("UpdatedAt: ")
	fmt.Printf("\t%s\n", release.UpdatedAt)
	lineColor.Print("Metadatas: \n")
}

//PrintAppInstances print commit list to ouput
func PrintAppInstances(instances []*common.AppInstance, business, app string, clusters map[string]*common.Cluster, zones map[string]*common.Zone) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "IP", "Business", "App", "Cluster", "Zone", "DataCenter", "Lables", "State"})
	table.SetHeaderColor(setTitleColorToRed(9)...)
	for _, c := range instances {
		var line []string
		line = append(line, strconv.FormatUint(c.Instanceid, 10))
		line = append(line, c.IP)
		line = append(line, business)
		line = append(line, app)
		if clu, ok := clusters[c.Clusterid]; ok {
			line = append(line, clu.Name)
		} else {
			line = append(line, c.Clusterid)
		}
		if z, ok := zones[c.Zoneid]; ok {
			line = append(line, z.Name)
		} else {
			line = append(line, c.Zoneid)
		}
		line = append(line, c.Dc)
		line = append(line, handleLablesFormat(c.Labels))
		if c.State == 1 {
			line = append(line, "Online")
		} else {
			line = append(line, "Offline")
		}
		table.Append(line)
	}
	table.Render()
}

//PrintAppInstances print commit list to ouput
func PrintAppInstancesByRelease(instances []*common.AppInstance, business string, app string, clusters map[string]*common.Cluster, zones map[string]*common.Zone,
	queryType int32) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "IP", "Business", "App", "Cluster", "Zone", "DataCenter", "Lables", "State", "EffectStatus - Time"})
	table.SetHeaderColor(setTitleColorToRed(10)...)
	for _, instance := range instances {
		// judge show online instance data / offline instance data (0 is all， 1 is online， 2 is offline)
		if queryType == 1 && instance.State != 1 {
			continue
		} else if queryType == 2 && instance.State != 0 {
			continue
		}
		var line []string
		line = append(line, strconv.FormatUint(instance.Instanceid, 10))
		line = append(line, instance.IP)
		line = append(line, business)
		line = append(line, app)
		if clu, ok := clusters[instance.Clusterid]; ok {
			line = append(line, clu.Name)
		} else {
			line = append(line, instance.Clusterid)
		}
		if z, ok := zones[instance.Zoneid]; ok {
			line = append(line, z.Name)
		} else {
			line = append(line, instance.Zoneid)
		}
		line = append(line, instance.Dc)
		line = append(line, handleLablesFormat(instance.Labels))
		if instance.State == 1 {
			line = append(line, "Online")
		} else {
			line = append(line, "Offline")
		}

		if instance.ReloadCode != 0 {
			if instance.ReloadCode == 1 {
				line = append(line, fmt.Sprintf("PublishReload - %s", instance.ReloadTime))
			} else {
				line = append(line, fmt.Sprintf("RollBackReload - %s", instance.ReloadTime))
			}
		} else if instance.EffectCode == 0 && instance.EffectMsg == "SUCCESS" {
			line = append(line, fmt.Sprintf("PublishEffected - %s", instance.EffectTime))
		} else {
			line = append(line, "UnEffected")
		}
		table.Append(line)
	}
	table.Render()
}

func PrintInitInfo(conf *option.CurrentDirConf) {
	pwd, _ := os.Getwd()
	fmt.Printf("Current repository[%s] init info:\n", pwd)
	fmt.Println("Business: " + conf.Business)
	fmt.Println("App: " + conf.App)
	fmt.Println("Operator: " + conf.Operator)
	fmt.Println("Token: " + conf.Token)
}

// PrintCommitDiffsTextMode prints diffs string between two commits in normal text mode.
func PrintCommitDiffsTextMode(commit1, commit2 *common.Commit) {
	if commit1 == nil || commit2 == nil {
		return
	}
	result := ""

	// commit1 metadata.
	configs1 := commit1.Configs
	template1 := commit1.Template
	templateRule1 := commit1.TemplateRule

	// commit2 metadata.
	configs2 := commit2.Configs
	template2 := commit2.Template
	templateRule2 := commit2.TemplateRule

	// diffs.
	dmp := diffmatchpatch.New()
	configsDiffs := dmp.DiffPrettyText(dmp.DiffMain(string(configs1), string(configs2), true))
	if len(configsDiffs) != 0 {
		result += `
Configs Diffs:
-------------
` + configsDiffs
	}

	dmp = diffmatchpatch.New()
	templateDiffs := dmp.DiffPrettyText(dmp.DiffMain(template1, template2, true))
	if len(templateDiffs) != 0 {
		result += `
Template Diffs:
--------------
` + templateDiffs
	}

	dmp = diffmatchpatch.New()
	templateRuleDiffs := dmp.DiffPrettyText(dmp.DiffMain(templateRule1, templateRule2, true))
	if len(templateRuleDiffs) != 0 {
		result += `
TemplateRule Diffs:
------------------
` + templateRuleDiffs
	}

	fmt.Println(result)
}

func PrintFileDiffGit(fileName1 string, fileContent1 []byte, fileName2 string, fileContent2 []byte) {
	result := ""
	// diffs.
	configsDiffs, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(fileContent1)),
		B:        difflib.SplitLines(string(fileContent2)),
		FromFile: fileName1,
		ToFile:   fileName2,
	})
	if err != nil {
		return
	}
	if len(configsDiffs) != 0 {
		result += `diff --bk-bscp-client new:` + fileName1 + ` old:` + fileName2 + ` 
------------------------------------------------------------------
` + configsDiffs
	}
	fmt.Println(result)
}

// PrintCommitDiffsGitMode prints diffs string between two commits in git mode.
func PrintCommitDiffsGitMode(commit1, commit2 *common.Commit) {
	if commit1 == nil || commit2 == nil {
		return
	}
	result := ""

	// commit1 metadata.
	configs1 := commit1.Configs
	template1 := commit1.Template
	templateRule1 := commit1.TemplateRule

	// commit2 metadata.
	configs2 := commit2.Configs
	template2 := commit2.Template
	templateRule2 := commit2.TemplateRule

	// diffs.
	configsDiffs, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(configs1)),
		B:        difflib.SplitLines(string(configs2)),
		FromFile: commit1.Commitid,
		ToFile:   commit2.Commitid,
	})
	if err != nil {
		return
	}

	if len(configsDiffs) != 0 {
		result += `
Configs Diffs:
-------------
` + configsDiffs
	}

	templateDiffs, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(template1),
		B:        difflib.SplitLines(template2),
		FromFile: commit1.Commitid,
		ToFile:   commit2.Commitid,
	})
	if err != nil {
		return
	}

	if len(templateDiffs) != 0 {
		result += `
Template Diffs:
--------------
` + templateDiffs
	}

	templateRuleDiffs, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(templateRule1),
		B:        difflib.SplitLines(templateRule2),
		FromFile: commit1.Commitid,
		ToFile:   commit2.Commitid,
	})
	if err != nil {
		return
	}
	if len(templateRuleDiffs) != 0 {
		result += `
TemplateRule Diffs:
------------------
` + templateRuleDiffs
	}

	fmt.Println(result)
}

func handleLablesFormat(lables string) string {
	pos := strings.Index(lables, ":")
	lables = lables[(pos + 1):]
	if lables[len(lables)-1] == '}' || lables[len(lables)-1] == '}' {
		lables = lables[:len(lables)-1]
	}
	return lables
}

func printJsonFormat(content []byte) {
	var out bytes.Buffer
	json.Indent(&out, content, "", "    ")
	out.WriteTo(os.Stdout)
}
