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

// Package main xxx
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// matches i18n.T 、 i18n.Tf 、Translate 、TranslateFormat
	i18nPattern = regexp.MustCompile(`i18n\.(T|Tf|Translate|TranslateFormat)\(.*?, "(.*?)"`)
	// variable matching rules
	varPattern       = regexp.MustCompile(`{{\.(.*?)}}`)
	symbolsToCheck   = []string{"#", "@", "{"}
	varNames         = make(map[string]bool)
	separator        = string(filepath.Separator)
	defaultTrimChars = string([]byte{
		'\t', // Tab.
		'\v', // Vertical tab.
		'\n', // New line (line feed).
		'\r', // Carriage return.
		'\f', // New page.
		' ',  // Ordinary space.
		0x00, // NUL-byte.
		0x85, // Delete.
		0xA0, // Non-breaking space.
	})
)

func main() {
	projectPath := flag.String("p", "", "项目根目录, 默认当前根目录")
	output := flag.String("o", "", "输出yaml文件目录")
	lang := flag.String("l", "", "指定导出语种, 多个逗号拼接 示例：-l en,zh, 语种文件存在追加写不存在创建, 默认为空")

	flag.Parse()

	rootPath := pwd()
	if *projectPath != "" {
		rootPath = *projectPath
	}

	if *output == "" {
		log.Fatal("Output directory is required")
	}
	langSlice := []string{}
	if *lang != "" {
		langSlice = strings.Split(*lang, ",")
	}

	if err := parseTemplate(rootPath); err != nil {
		log.Fatal(err)
	}

	uniqueVarNames := make([]string, 0, len(varNames))
	for varName := range varNames {
		uniqueVarNames = append(uniqueVarNames, varName)
	}

	var paths []string
	if len(langSlice) > 0 {
		for _, item := range langSlice {
			filePath := join([]string{*output, item + ".yaml"}...)
			paths = append(paths, filePath)
		}
	} else {
		err := filepath.Walk(*output, func(path string, info os.FileInfo, err error) error {
			ext := strings.ToLower(filepath.Ext(path))
			if !info.IsDir() && (ext == ".yaml" || ext == ".yml") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	err := updateYamlFile(paths, uniqueVarNames)
	if err != nil {
		log.Fatal(err)
	}
}

// parse template variables
// Example:
// hello -> hello
// {{.hello}} -> hello
// {#hello} -> hello
// {@hello} -> hello
func parseTemplate(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				// 跳过带有注释的行
				if strings.Contains(line, "//") {
					continue
				}
				matches := i18nPattern.FindAllStringSubmatch(line, -1)
				for _, match := range matches {
					i18nKey := match[2]
					varMatches := varPattern.FindAllStringSubmatch(i18nKey, -1)
					if varMatches != nil {
						for _, varMatch := range varMatches {
							varName := varMatch[1]
							varNames[varName] = true
						}
					} else if !containsSymbols(i18nKey, symbolsToCheck) {
						varNames[i18nKey] = true
					}
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func pwd() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return path
}

func join(paths ...string) string {
	var s string
	for _, path := range paths {
		if s != "" {
			s += separator
		}
		s += trimRight(path, separator)
	}
	return s
}

func trimRight(str string, characterMask ...string) string {
	trimChars := defaultTrimChars
	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	return strings.TrimRight(str, trimChars)
}

func containsSymbols(text string, symbols []string) bool {
	for _, symbol := range symbols {
		if containsSymbol(text, symbol) {
			return true
		}
	}
	return false
}

func containsSymbol(text, symbol string) bool {
	return regexp.MustCompile(regexp.QuoteMeta(symbol)).MatchString(text)
}

func updateYamlFile(paths, varNames []string) error {
	for _, item := range paths {
		yamlData, _ := readYAMLFile(item)
		key := []string{}
		for _, varName := range varNames {
			_, exists := yamlData[varName]
			if !exists {
				key = append(key, varName)
			}
		}
		if len(key) == 0 {
			fmt.Printf("no data update %s\n", item)
			continue
		}

		newData := make(map[string]interface{})
		for _, v := range key {
			newData[v] = "nil"
		}

		err := writeYAMLFile(item, newData)
		if err != nil {
			return fmt.Errorf("error writing YAML %s %s", item, err.Error())
		}
		fmt.Printf("update completed %s %v\n", item, newData)
	}
	return nil
}

func readYAMLFile(filePath string) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	file, err := os.Open(filePath)
	if err != nil {
		return data, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func writeYAMLFile(filePath string, dataToAppend map[string]interface{}) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("fail to open the file %s %s", filePath, err.Error())
	}
	defer file.Close()

	_, err = file.WriteString("\n")
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buf)

	err = encoder.Encode(dataToAppend)
	if err != nil {
		return err
	}

	_, err = file.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("fail write %s %s", filePath, err.Error())
	}

	return nil
}
