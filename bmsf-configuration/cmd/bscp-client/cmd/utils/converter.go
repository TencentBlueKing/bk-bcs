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
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/sergi/go-diff/diffmatchpatch"

	"bk-bscp/internal/protocol/common"
)

//PrintBusiness print business list to ouput
func PrintBusiness(businesses []*common.Business) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Department", "Creator", "State"})
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
func PrintShardingDB(shardingDBs []*common.ShardingDB) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Host", "Port", "User", "State", "LastUpdated"})
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

//PrintApp print app list to ouput
func PrintApp(apps []*common.App, business string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Business", "Type", "Creator", "State", "LastModifyBy"})
	for _, app := range apps {
		var line []string
		line = append(line, app.Appid)
		line = append(line, app.Name)
		line = append(line, business)
		line = append(line, strconv.Itoa(int(app.DeployType)))
		line = append(line, app.Creator)
		line = append(line, strconv.Itoa(int(app.State)))
		line = append(line, app.LastModifyBy)
		table.Append(line)
	}
	table.Render()
}

//PrintClusters print cluster list to ouput
func PrintClusters(clusters []*common.Cluster, business, app string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Business", "App", "Creator", "State", "LastModifyBy", "RClusterid"})
	for _, cluster := range clusters {
		var line []string
		line = append(line, cluster.Clusterid)
		line = append(line, cluster.Name)
		line = append(line, business)
		line = append(line, app)
		line = append(line, cluster.Creator)
		line = append(line, strconv.Itoa(int(cluster.State)))
		line = append(line, cluster.LastModifyBy)
		line = append(line, cluster.RClusterid)
		table.Append(line)
	}
	table.Render()
}

//PrintZones print zone list to ouput
func PrintZones(zones []*common.Zone, business, app, cluster string) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Business", "App", "Cluster", "Creator", "State", "LastModifyBy", "UpdateAt"})
	for _, zone := range zones {
		var line []string
		line = append(line, zone.Clusterid)
		line = append(line, zone.Name)
		line = append(line, business)
		line = append(line, app)
		line = append(line, cluster)
		line = append(line, zone.Creator)
		line = append(line, strconv.Itoa(int(zone.State)))
		line = append(line, zone.LastModifyBy)
		line = append(line, zone.UpdatedAt)
		table.Append(line)
	}
	table.Render()
}

//PrintConfigSet print zone list to ouput
func PrintConfigSet(configSets []*common.ConfigSet, business *common.Business, app *common.App) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Fpath", "Business", "App", "Creator", "State", "LastModifyBy", "UpdateAt"})
	for _, cfg := range configSets {
		var line []string
		line = append(line, cfg.Cfgsetid)
		line = append(line, cfg.Name)
		line = append(line, cfg.Fpath)
		line = append(line, business.Name)
		line = append(line, app.Name)
		line = append(line, cfg.Creator)
		line = append(line, strconv.Itoa(int(cfg.State)))
		line = append(line, cfg.LastModifyBy)
		line = append(line, cfg.UpdatedAt)
		table.Append(line)
	}
	table.Render()
}

//PrintCommits print commit list to ouput
func PrintCommits(commits []*common.Commit, business *common.Business, app *common.App, cfgset *common.ConfigSet) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Business", "Application", "ConfigSet", "Creator", "State", "CreatedAt", "UpdatedAt"})
	for _, c := range commits {
		var line []string
		line = append(line, c.Commitid)
		line = append(line, business.Name)
		line = append(line, app.Name)
		line = append(line, cfgset.Name)
		line = append(line, c.Operator)
		line = append(line, strconv.Itoa(int(c.State)))
		line = append(line, c.CreatedAt)
		line = append(line, c.UpdatedAt)
		table.Append(line)
	}
	table.Render()
}

//PrintDetailsCommit print commit list to ouput
func PrintDetailsCommit(commit *common.Commit, business string) {
	//format output
	fmt.Printf("CommitID: %s\n", commit.Commitid)
	fmt.Printf("BusinessID: %s/%s\n", commit.Bid, business)
	fmt.Printf("AppID: %s\n", commit.Appid)
	fmt.Printf("ConfigSetID: %s\n", commit.Cfgsetid)
	fmt.Printf("Creator: %s\n", commit.Operator)
	if len(commit.Releaseid) != 0 {
		fmt.Printf("ReleaseID: %s\n", commit.Releaseid)
	} else {
		fmt.Println("ReleaseID: Not Released")
	}
	if commit.State == 0 {
		fmt.Printf("State: NotAffectived\n")
	} else {
		fmt.Printf("State: Affectived\n")
	}
	fmt.Printf("CreatedAt: %s\n", commit.CreatedAt)
	fmt.Printf("UpdatedAt: %s\n", commit.UpdatedAt)
	fmt.Println("PrevConfigs:")
	prevLines := strings.Split(string(commit.PrevConfigs), "\n")
	for _, line := range prevLines {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println("Configs:")
	cfgLines := strings.Split(string(commit.Configs), "\n")
	for _, line := range cfgLines {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println("Changes:")
	fmt.Printf("%s\n", commit.Changes)
}

//PrintStrategies print commit list to ouput
func PrintStrategies(strategies []*common.Strategy, app *common.App) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "App", "State", "Creator", "CreatedAt", "LastModifyBy", "UpdatedAt"})
	for _, c := range strategies {
		var line []string
		line = append(line, c.Strategyid)
		line = append(line, c.Name)
		line = append(line, app.Name)
		line = append(line, strconv.Itoa(int(c.State)))
		line = append(line, c.Creator)
		line = append(line, c.CreatedAt)
		line = append(line, c.LastModifyBy)
		line = append(line, c.UpdatedAt)
		table.Append(line)
	}
	table.Render()
}

//PrintStrategy print commit list to ouput
func PrintStrategy(strategy *common.Strategy, app *common.App) {
	//format output
	fmt.Printf("StrategyID: %s\n", strategy.Strategyid)
	fmt.Printf("Name: %s\n", strategy.Name)
	fmt.Printf("App: %s/%s\n", strategy.Appid, app.Name)
	fmt.Printf("State: %d\n", strategy.State)
	fmt.Printf("Creator: %s\n", strategy.Creator)
	fmt.Printf("CreatedAt: %s\n", strategy.CreatedAt)
	fmt.Printf("LastModifyBy: %s\n", strategy.LastModifyBy)
	fmt.Printf("UpdatedAt: %s\n", strategy.UpdatedAt)
	fmt.Printf("Content:\n")
	lines := strings.Split(strategy.Content, "\n")
	for _, line := range lines {
		fmt.Printf("    %s\n", line)
	}
}

//PrintReleases print commit list to ouput
func PrintReleases(releases []*common.Release, business *common.Business, app *common.App) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Business", "App", "ConfigSet", "ConfigSetFpath", "CommitID", "State", "Creator", "CreatedAt", "LastModifyBy", "UpdatedAt"})
	for _, c := range releases {
		var line []string
		line = append(line, c.Releaseid)
		line = append(line, c.Name)
		line = append(line, business.Name)
		line = append(line, app.Name)
		line = append(line, c.CfgsetName)
		line = append(line, c.CfgsetFpath)
		line = append(line, c.Commitid)
		if c.State == 0 {
			line = append(line, "NotAffected")
		} else {
			line = append(line, "Published")
		}
		line = append(line, c.Creator)
		line = append(line, c.CreatedAt)
		line = append(line, c.LastModifyBy)
		line = append(line, c.UpdatedAt)
		table.Append(line)
	}
	table.Render()
}

//PrintRelease print commit list to ouput
func PrintRelease(release *common.Release) {
	//format output
	fmt.Printf("ReleaseID: %s\n", release.Releaseid)
	fmt.Printf("Name: %s\n", release.Name)
	fmt.Printf("BusinessID: %s\n", release.Bid)
	fmt.Printf("AppID: %s\n", release.Appid)
	if release.State == 0 {
		fmt.Printf("State: Not affected\n")
	} else {
		fmt.Printf("State: Published\n")
	}
	fmt.Printf("ConfigSetID: %s\n", release.Cfgsetid)
	fmt.Printf("ConfigSetName: %s\n", release.CfgsetName)
	fmt.Printf("ConfigSetFpath: %s\n", release.CfgsetFpath)
	fmt.Printf("CommitID: %s\n", release.Commitid)
	fmt.Printf("StrategyID: %s\n", release.Strategyid)
	fmt.Printf("Strategies:\n")
	strLines := strings.Split(release.Strategies, "\n")
	for _, line := range strLines {
		fmt.Printf("    %s\n", line)
	}
	fmt.Printf("Creator: %s\n", release.Creator)
	fmt.Printf("CreatedAt: %s\n", release.CreatedAt)
	fmt.Printf("LastModifyBy: %s\n", release.LastModifyBy)
	fmt.Printf("UpdatedAt: %s\n", release.UpdatedAt)
}

//PrintAppInstances print commit list to ouput
func PrintAppInstances(instances []*common.AppInstance, business, app string, clusters map[string]*common.Cluster, zones map[string]*common.Zone) {
	//format output
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "IP", "Business", "App", "Cluster", "Zone", "DataCenter", "Lables", "State", "EffectTime", "CreatedAt", "UpdatedAt"})
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
		line = append(line, c.Labels)
		if c.State == 1 {
			line = append(line, "Online")
		} else {
			line = append(line, "Offline")
		}
		line = append(line, c.EffectTime)
		line = append(line, c.CreatedAt)
		line = append(line, c.UpdatedAt)
		table.Append(line)
	}
	table.Render()
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
