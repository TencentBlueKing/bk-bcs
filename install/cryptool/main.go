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

// Package main is the main package of cryptool
package main

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2" // nolint
)

func main() {
	if len(os.Args) != 6 {
		fmt.Printf("command line arguments lost!\n")
		fmt.Printf("try:\n")
		fmt.Printf("normal Algorithm\n")
		fmt.Printf("    cryptools normal compileKey priKey decrypt [string]\n")
		fmt.Printf("    cryptools normal compileKey priKey encrypt [string]\n")
		fmt.Printf("SM4 Algorithm\n")
		fmt.Printf("    cryptools SM4 key iv decrypt [string]\n")
		fmt.Printf("    cryptools SM4 key iv encrypt [string]\n")
		fmt.Printf("AES-GCM Algorithm\n")
		fmt.Printf("    cryptools AES-GCM key nonce decrypt [string]\n")
		fmt.Printf("    cryptools normal key nonce encrypt [string]\n")
		return
	}

	switch os.Args[1] {
	case encryptv2.Normal.String(), encryptv2.Sm4.String(), encryptv2.AesGcm.String():
		key1 := os.Args[2]
		key2 := os.Args[3]

		cryptor, err := encryptv2.NewCrypto(&encryptv2.Config{
			Enabled:   true,
			Algorithm: encryptv2.Algorithm(os.Args[1]),
			Sm4: &encryptv2.Sm4Conf{
				Key: key1,
				Iv:  key2,
			},
			AesGcm: &encryptv2.AesGcmConf{
				Key:   key1,
				Nonce: key2,
			},
			Normal: &encryptv2.NormalConf{
				CompileKey: key1,
				PriKey:     key2,
			},
		})
		if err != nil {
			fmt.Printf("[normal|SM4|AES-GCM] cryptor init failed: %s\n", err.Error())
			return
		}
		switch os.Args[4] {
		case "encrypt":
			out, err := cryptor.Encrypt(os.Args[5])
			if err != nil {
				fmt.Printf("%s encrypt from original failed: %s\n", os.Args[1], err.Error())
				return
			}
			fmt.Printf("%s Encrypt text: %s\n", os.Args[1], string(out)) // nolint
		case "decrypt":
			out, err := cryptor.Decrypt(os.Args[5])
			if err != nil {
				fmt.Printf("%s Decrypt from Base failed: %s\n", os.Args[1], err.Error())
				return
			}
			fmt.Printf("%s Original text: %s\n", os.Args[1], string(out)) // nolint
		default:
			fmt.Printf("Unknown action...\n")
		}
	default:
		fmt.Printf("Unknown Algorithm, please [normal|SM4|AES-GCM]...\n")
	}

	return // nolint
}
