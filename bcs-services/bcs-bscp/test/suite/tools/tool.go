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

// Package main is used as the entry of the suit-test tool
package main

import (
	"bufio"
	"flag"
	"log"
	"os"
)

// inputDir goconvey json file save dir.
var inputDir string

// outputPath statistics result save file path.
var outputPath string

func main() {
	flag.StringVar(&inputDir, "input-dir", "./result", "go convey test result export json file "+
		"dir, no other files can exist in this dir")
	flag.StringVar(&outputPath, "output-path", "./statistics.html", "statistics result html "+
		"file that by go convey test result, save file path")
	flag.Parse()

	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalln(err)
	}

	outFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE, 0644) //nolint
	if err != nil {
		log.Fatalf("open file failed, err: %v\n", err)
	}
	defer outFile.Close()
	write := bufio.NewWriter(outFile)

	for _, f := range files {
		path := inputDir + "/" + f.Name()

		results, err := statistics(path)
		if err != nil {
			log.Fatalln(err)
		}

		tp.render(write, results)
		write.Flush()
	}

	return
}
