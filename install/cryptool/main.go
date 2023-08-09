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
	if len(os.Args) <= 3 {
		fmt.Printf("command line arguments lost!\n")
		fmt.Printf("try:\n")
		fmt.Printf("normal Algorithm\n")
		fmt.Printf("    cryptool normal priKey decrypt [string]\n")
		fmt.Printf("    cryptool normal priKey encrypt [string]\n")
		fmt.Printf("SM4 Algorithm\n")
		fmt.Printf("    cryptool SM4 key iv decrypt [string]\n")
		fmt.Printf("    cryptool SM4 key iv encrypt [string]\n")
		fmt.Printf("AES-GCM Algorithm\n")
		fmt.Printf("    cryptool AES-GCM key nonce decrypt [string]\n")
		fmt.Printf("    cryptool normal key nonce encrypt [string]\n")
		return
	}

	switch os.Args[1] {
	case encrypt.Normal.String():
		if len(os.Args) != 5 {
			fmt.Printf("normal Algorithm style:\n")
			fmt.Printf("    cryptool normal priKey decrypt [string]\n")
			fmt.Printf("    cryptool normal priKey encrypt [string]\n")
			return
		}
		priKey := os.Args[2]

		cryptor, err := encrypt.NewCrypto(&encrypt.Config{
			Enabled:   true,
			Algorithm: encrypt.Normal,
			PriKey:    priKey,
		})
		if err != nil {
			fmt.Printf("normal cryptor init failed: %s\n", err.Error())
			return
		}
		switch os.Args[3] {
		case "encrypt":
			out, err := cryptor.Encrypt(os.Args[4])
			if err != nil {
				fmt.Printf("normal encrypt from original failed: %s\n", err.Error())
				return
			}
			fmt.Printf("normal Encrypt text: %s\n", string(out))
		case "decrypt":
			out, err := cryptor.Decrypt(os.Args[4])
			if err != nil {
				fmt.Printf("normal Decrypt from Base failed: %s\n", err.Error())
				return
			}
			fmt.Printf("normal Original text: %s\n", string(out))
		default:
			fmt.Printf("Unknown action...\n")
		}
	case encrypt.Sm4.String(), encrypt.AesGcm.String():
		if len(os.Args) != 6 {
			fmt.Printf("[SM4|AES-GCM] Algorithm style:\n")
			fmt.Printf("    cryptool [SM4|AES-GCM] key iv decrypt [string]\n")
			fmt.Printf("    cryptool [SM4|AES-GCM] key iv encrypt [string]\n")
			return
		}
		key := os.Args[2]
		iv := os.Args[3]

		cryptor, err := encrypt.NewCrypto(&encrypt.Config{
			Enabled:   true,
			Algorithm: encrypt.Algorithm(os.Args[1]),
			Sm4: &encrypt.Sm4Conf{
				Key: key,
				Iv:  iv,
			},
			AesGcm: &encrypt.AesGcmConf{
				Key:   key,
				Nonce: iv,
			},
		})
		if err != nil {
			fmt.Printf("[SM4|AES-GCM] cryptor init failed: %s\n", err.Error())
			return
		}
		switch os.Args[4] {
		case "encrypt":
			out, err := cryptor.Encrypt(os.Args[5])
			if err != nil {
				fmt.Printf("%s encrypt from original failed: %s\n", os.Args[1], err.Error())
				return
			}
			fmt.Printf("%s Encrypt text: %s\n", os.Args[1], string(out))
		case "decrypt":
			out, err := cryptor.Decrypt(os.Args[5])
			if err != nil {
				fmt.Printf("%s Decrypt from Base failed: %s\n", os.Args[1], err.Error())
				return
			}
			fmt.Printf("%s Original text: %s\n", os.Args[1], string(out))
		default:
			fmt.Printf("Unknown action...\n")
		}
	default:
		fmt.Printf("Unknown Algorithm, please [normal|SM4|AES-GCM]...\n")
	}

	return
}
