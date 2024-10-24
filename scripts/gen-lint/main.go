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

// gen-lint is a program for auto generate .golangci.yml for mod
package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/bitfield/script"
	goyaml "github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/parser"
)

func getWhiteList() ([]string, error) {
	file, err := os.Open("./scripts/gen-lint/modules_white_list")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建Scanner来逐行读取
	scanner := bufio.NewScanner(file)

	// 创建一个string类型的切片来存储文件中的每一行
	var lines []string

	// 逐行扫描
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "#") {
			continue
		}
		// 将扫描到的行添加到切片中
		lines = append(lines, scanner.Text())
	}

	// 检查是否有可能的错误
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func matchWhiteList(line string, blackList []string) bool {
	for _, m := range blackList {
		if line == m {
			return true
		}
	}
	return false
}

func getGolangciYmlTemplate() (string, error) {
	content, err := os.ReadFile("./scripts/gen-lint/.golangci.yml.tpl")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// get skip files from code cc
func getSkipFiles() ([]string, error) {
	content, err := os.ReadFile(".code.yml")
	if err != nil {
		return nil, err
	}

	filePath, err := goyaml.PathString("$.source.test_source.filepath_regex")
	if err != nil {
		return nil, err
	}
	skipFiles := []string{}
	_ = filePath.Read(strings.NewReader(string(content)), &skipFiles)
	return skipFiles, nil
}

func findModuleSkipFilesAndDirs(moduleDir string, files []string) ([]string, []string) {
	skipDirs := []string{}
	skipFiles := []string{}
	for _, file := range files {
		file = strings.TrimPrefix(file, "/")
		if strings.HasPrefix(file, moduleDir) {
			file = strings.Replace(file, moduleDir, "", 1)
			file = strings.TrimPrefix(file, "/")
			if strings.HasSuffix(file, ".go") {
				skipFiles = append(skipFiles, file)
			}
			if strings.HasSuffix(file, ".*") {
				file = strings.ReplaceAll(file, ".*", "*")
				skipDirs = append(skipDirs, file)
			}
		}
	}
	return skipFiles, skipDirs
}

func main() {
	whiteList, err := getWhiteList()
	if err != nil {
		panic(err)
	}
	mods, err := script.FindFiles("./").Match("go.mod").String()
	if err != nil {
		panic(err)
	}
	tpl, err := getGolangciYmlTemplate()
	if err != nil {
		panic(err)
	}
	skipFiles, err := getSkipFiles()
	if err != nil {
		panic(err)
	}
	// parse origin yaml
	f, err := parser.ParseBytes([]byte(tpl), 0)
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`module\s+(.*)`)
	for _, line := range strings.Split(mods, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		dir := path.Dir(line)
		if !matchWhiteList(dir, whiteList) {
			continue
		}

		v, err := script.File(line).String()
		if err != nil {
			continue
		}
		match := re.FindStringSubmatch(v)
		if len(match) <= 1 {
			continue
		}
		module := match[1]

		// get skip files and dirs
		gotSkipFiles, gotSkipDirs := findModuleSkipFilesAndDirs(dir, skipFiles)
		skipDirsPath, err := goyaml.PathString("$.issues.exclude-dirs")
		if err != nil {
			continue
		}
		skipFilesPath, err := goyaml.PathString("$.issues.exclude-files")
		if err != nil {
			continue
		}

		// append skip files and dirs
		skipDirs := []string{}
		skipFiles := []string{}
		_ = skipDirsPath.Read(strings.NewReader(tpl), &skipDirs)
		_ = skipFilesPath.Read(strings.NewReader(tpl), &skipFiles)
		_ = skipDirsPath.ReplaceWithReader(f, strings.NewReader(stringsToYAML(append(skipDirs, gotSkipDirs...))))
		_ = skipFilesPath.ReplaceWithReader(f, strings.NewReader(stringsToYAML(append(skipFiles, gotSkipFiles...))))

		// append gci local-prefixes
		localPrefixesPath, err := goyaml.PathString("$.linters-settings.goimports.local-prefixes")
		if err != nil {
			continue
		}
		_ = localPrefixesPath.ReplaceWithReader(f, strings.NewReader(module))

		// append gci prefix
		gciPath, err := goyaml.PathString("$.linters-settings.gci.sections")
		if err != nil {
			continue
		}
		_ = gciPath.ReplaceWithReader(f, strings.NewReader(stringsToYAML([]string{"standard", "default",
			fmt.Sprintf("prefix(%s)", module)})))

		_ = os.WriteFile(path.Join(dir, "./.golangci.yml"), []byte(f.String()), 0644)
		fmt.Printf("generate .golangci.yml to %s done.\n", dir)
	}
}

func stringsToYAML(v []string) string {
	s := ""
	for _, i := range v {
		s += fmt.Sprintf("- %s\n", i)
	}
	return s
}
