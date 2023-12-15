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

// Package translations is used to generate translations for different languages
package translations

// 通过gotext命令行工具，能够自动从源码中提取要翻译的message以及合并message，并生成对应的目录文件
// -srclang 指定在源码代中使用的语言，这里是英语en, 语言名称需符合BCP 47规范https://en.wikipedia.org/wiki/IETF_language_tag
// update 该子命令用来从源码中提取要翻译的message以及进行合并操作，并生成对应的目录文件
// -out 指定要生成的message catalog文件，主要是存放不同语言翻译前后的映射关系
// -lang 指定要翻译的目标语言，这里是英语en和中文zh，多个语言之间以逗号分隔
// 最后的参数表示要提取翻译message的包路径，多个包路径以空格分隔（这里使用需要翻译的模块入口所在package，会自动加载这些包的依赖包）
// gotext命令行工具更多使用，见gotext help输出
// gotext源码使用example：https://cs.opensource.google/go/x/text/+/refs/tags/v0.14.0:cmd/gotext/examples/
// 更多实现细节见gotext源码：https://cs.opensource.google/go/x/text/+/refs/tags/v0.14.0:cmd/gotext/main.go

//go:generate gotext -srclang=zh-Hans update -out=catalog.go -lang=en,zh-Hans ../../
