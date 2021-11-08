/**
 * Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
 * Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

const { extname } = require('path')

class ReplaceStaticUrlPlugin {
    apply (compiler, callback) {
        // emit: 在生成资源并输出到目录之前
        compiler.hooks.emit.tapAsync('ReplaceCSSStaticUrlPlugin', (compilation, callback) => {
            const assets = Object.keys(compilation.assets)
            const assetsLen = assets.length

            for (let i = 0; i < assetsLen; i++) {
                const fileName = assets[i]
                if (extname(fileName) === '.js') {
                    if (fileName.indexOf('manifest') > -1) {
                        const asset = compilation.assets[fileName]

                        const minifyFileContent = asset.source().replace(
                            // /\"\{\{\s{1}STATIC_URL\s{1}\}\}\"/,
                            /\"\{\{STATIC_URL\}\}\"/g,
                            () => 'window.STATIC_URL + "/"'
                        )
                        // 设置输出资源
                        compilation.assets[fileName] = {
                            // 返回文件内容
                            source: () => minifyFileContent,
                            // 返回文件大小
                            size: () => Buffer.byteLength(minifyFileContent, 'utf8')
                        }
                        break
                    }
                }
            }

            callback()
        })
    }
}

module.exports = ReplaceStaticUrlPlugin
