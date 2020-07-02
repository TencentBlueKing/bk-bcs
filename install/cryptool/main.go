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

package main

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("command line arguments lost!\n")
		fmt.Printf("try:\n")
		fmt.Printf("    cryptool decrypt [string]\n")
		fmt.Printf("    cryptool encrypt [string]\n")
		return
	}
	switch os.Args[1] {
	case "encrypt":
		out, err := encrypt.DesEncryptToBase([]byte(os.Args[2]))
		if err != nil {
			fmt.Printf("encrypt from original failed: %s\n", err.Error())
			return
		}
		fmt.Printf("Encrypt text: %s\n", string(out))
	case "decrypt":
		out, err := encrypt.DesDecryptFromBase([]byte(os.Args[2]))
		if err != nil {
			fmt.Printf("Decrypt from Base failed: %s\n", err.Error())
			return
		}
		fmt.Printf("Original text: %s\n", string(out))
	default:
		fmt.Printf("Unkown action...\n")
	}
}
