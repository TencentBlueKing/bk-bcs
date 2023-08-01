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

package audit

import (
	"strings"

	"github.com/pborman/ansi"
)

type cmdParse struct {
	CmdOutput  chan []map[*ansi.S][]*ansi.S
	Cmd        *ansi.S
	InputSlice []*ansi.S
	CmdResult  map[*ansi.S]*ansi.S
}

func NewCmdParse() *cmdParse {
	c := new(cmdParse)
	c.CmdOutput = make(chan []map[*ansi.S][]*ansi.S)
	c.CmdResult = make(map[*ansi.S]*ansi.S)
	c.InputSlice = make([]*ansi.S, 0)

	return c
}

// ReceivedCmdOutChan 从outChannel读取数据
func ReceivedCmdOutChan(c *cmdParse) {
	for inOutSlice := range c.CmdOutput {
		readyParse := make([]*ansi.S, 0)
		cmdMap := make(map[*ansi.S]*ansi.S)
		for _, inOutMap := range inOutSlice {
			for in, outSlice := range inOutMap {
				for _, out := range outSlice {
					if in.Code == "\r" {
						if !notInVim(outSlice) {
							readyParse = []*ansi.S{}
						}
					} else {
						readyParse = append(readyParse, in)
						cmdMap[in] = out
					}
				}

			}
		}

		udCmd := parseUpDown(readyParse, cmdMap)
		tabCmd := parseTab(udCmd, cmdMap)
		lCmd, rCmd := parseShortCut(tabCmd)
		lrCmd := parseLeftRight(lCmd, rCmd)
		dCmd := parseDelete(lrCmd)
		//for _, v := range dCmd {
		//	fmt.Printf("---dCmd---%#v\n", v)
		//}

		cmd := ""
		for _, v := range dCmd {
			cmd += string(v.Code)
		}

	}
}

// 将命令字符串转换结构体
func str2Struct(s string) []*ansi.S {
	result := make([]*ansi.S, 0)
	for _, appendCmd := range s {
		newS := new(ansi.S)
		newS.Code = ansi.Name(string(appendCmd))
		result = append(result, newS)
	}
	return result
}

// 是否在vim状态
func notInVim(a []*ansi.S) bool {
	if len(a) == 1 && a[0].Code == "\r\n" {
		return true
	}
	if len(a) >= 2 {
		t1 := contains(a, "\r\n")
		t2 := contains(a, ansi.SM)
		return t1 && t2
	}

	return false
}

// contains 切片是否包含某元素
func contains(s []*ansi.S, params ansi.Name) bool {
	for _, v := range s {
		if v.Code == params {
			return true
		}
	}
	return false
}

// insertByIndex 指定位置插入
func insertByIndex(slice []*ansi.S, index int, element *ansi.S) []*ansi.S {
	slice = append(slice[:index], append([]*ansi.S{element}, slice[index:]...)...)
	return slice

}

// parseUpDown 解析上下键,历史命令
func parseUpDown(s []*ansi.S, m map[*ansi.S]*ansi.S) []*ansi.S {
	result := make([]*ansi.S, 0)
	//获取最后一个
	var last *ansi.S
	for i := len(s) - 1; i >= 0; i-- {
		if s[i].Code == ansi.CUD || s[i].Code == ansi.CUU {
			last = s[i]
			cmdS := m[last]
			cmd := strings.TrimLeft(string(cmdS.Code), "\b")
			//完整命令解析拆分字节流形式
			lis := str2Struct(cmd)
			result = append(append(result, lis...), s[i+1:]...)
			return result
		}
	}
	return s
}

// parseTab 解析tab键 命令补充
func parseTab(s []*ansi.S, m map[*ansi.S]*ansi.S) []*ansi.S {
	result := make([]*ansi.S, 0)
	for _, v := range s {
		if v.Code == "\t" {
			outCmd := m[v]
			outCmdStr := string(outCmd.Code)
			switch {
			//情形一:有多个结果展示,但不补全
			case strings.Contains(outCmdStr, "\r\n"):
				continue
			//情形二:后面有字符的补全 eg:j/t(obs)h (命令不能是中文)
			case strings.HasSuffix(outCmdStr, "\b"):
				i := 0
				for _, v := range outCmdStr {
					if string(v) == "\b" {
						i += 1
					}
				}
				outCmdStr = outCmdStr[:len(outCmdStr)-i*2]
				tmplist := str2Struct(outCmdStr)
				result = append(result, tmplist...)
			//和情形二一样,只是多余\a字符
			case strings.HasPrefix(outCmdStr, "\a"):
				outCmdStr = strings.TrimPrefix(outCmdStr, "\a")
				tmplist := str2Struct(outCmdStr)
				result = append(result, tmplist...)
			default:
				//情形三:直接补全
				result = append(result, outCmd)

			}
		} else {
			result = append(result, v)
		}
	}

	return result
}

// parseShortCut 解析快捷键
func parseShortCut(s []*ansi.S) ([]*ansi.S, []*ansi.S) {
	result1 := make([]*ansi.S, 0)
	result2 := make([]*ansi.S, 0)
	for _, v := range s {
		switch v.Code {
		case "\x01": //CTRL A
			result2 = result1
			result1 = []*ansi.S{}
		default:
			result1 = append(result1, v)
		}
	}
	return result1, result2
}

// parseLeftRight 包含左右键时解析
func parseLeftRight(lc []*ansi.S, rc []*ansi.S) []*ansi.S {
	result := make([]*ansi.S, 0)
	cursor := 0
	p := 0
	for _, v := range lc {
		switch {
		//左键
		case v.Code == ansi.CUB:
			cursor -= 1
		//右键
		case v.Code == ansi.CUF:
			//在行尾就不移动
			if len(rc) > p {
				result = append(result, rc[p])
				p += 1
			}
			cursor += 1
		default:
			result = insertByIndex(result, cursor, v)
			cursor = cursor + 1
		}
	}
	if len(rc) > p {
		result = append(result, rc[p:]...)
	}
	return result
}

// parseDelete 解析删除键
func parseDelete(s []*ansi.S) []*ansi.S {
	result := make([]*ansi.S, 0)
	for _, v := range s {
		if v.Code == "\x7f" && (len(result)-1) >= 0 {
			result = result[:len(result)-1]
		} else {
			result = append(result, v)
		}
	}
	return result
}

func ResolveInOut(c *cmdParse) string {
	readyParseIn := make([]*ansi.S, 0)
	readyParseMap := make(map[*ansi.S]*ansi.S)
	for _, v := range c.InputSlice {
		//TODO:输入输出循序不一致,导致nil存在
		if c.CmdResult[v] != nil {
			// in vim
			if v.Code == "\r" && c.CmdResult[v].Code != "\r\n" {
				readyParseIn = []*ansi.S{}
				readyParseMap = map[*ansi.S]*ansi.S{}
			}

			//无效信号
			if c.CmdResult[v].Code == "\a" || v.Code == "\r" {
				continue
			}
		}
		readyParseIn = append(readyParseIn, v)
		readyParseMap[v] = c.CmdResult[v]
	}

	udCmd := parseUpDown(readyParseIn, readyParseMap)
	tabCmd := parseTab(udCmd, readyParseMap)
	lCmd, rCmd := parseShortCut(tabCmd)
	lrCmd := parseLeftRight(lCmd, rCmd)
	dCmd := parseDelete(lrCmd)
	cmd := ""
	for _, v := range dCmd {
		cmd += string(v.Code)
	}

	c.InputSlice = []*ansi.S{}
	c.CmdResult = map[*ansi.S]*ansi.S{}

	return cmd
}
