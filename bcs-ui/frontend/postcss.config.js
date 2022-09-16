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

const path = require('path')
const ResolverFactory = require('enhanced-resolve/lib/ResolverFactory')
const NodeJsInputFileSystem = require('enhanced-resolve/lib/NodeJsInputFileSystem')
const CachedInputFileSystem = require('enhanced-resolve/lib/CachedInputFileSystem')

const CACHED_DURATION = 60000
const fileSystem = new CachedInputFileSystem(new NodeJsInputFileSystem(), CACHED_DURATION)

const resolver = ResolverFactory.createResolver({
    alias: {
        '@': path.resolve('src')
    },
    extensions: ['.css'],
    modules: ['src', 'node_modules'],
    useSyncFileSystemCalls: true,
    fileSystem
})

// https://github.com/michael-ciniawsky/postcss-load-config
module.exports = {
    plugins: {
        // 把 import 的内容转换为 inline
        // @see https://github.com/postcss/postcss-import#postcss-import
        'postcss-import': {
            resolve: (id, basedir) => resolver.resolveSync({}, basedir, id)
        },

        // mixins，本插件需要放在 postcss-simple-vars 和 postcss-nested 插件前面
        // @see https://github.com/postcss/postcss-mixins#postcss-mixins-
        'postcss-mixins': {
            // mixins: require('./src/css/mixins')
        },

        // 用于在 URL ( )上重新定位、内嵌或复制。
        // @see https://github.com/postcss/postcss-url#postcss-url
        'postcss-url': {
            url: 'rebase'
        },

        // cssnext 已经不再维护，推荐使用 postcss-preset-env
        'postcss-preset-env': {
            // see https://github.com/csstools/postcss-preset-env#options
            stage: 0,
            autoprefixer: {
                grid: true
            }
        },
        // 这个插件可以在写 nested 样式时省略开头的 &
        // @see https://github.com/postcss/postcss-nested#postcss-nested-
        'postcss-nested': {},

        // 将 @at-root 里的规则放入到根节点
        // @see https://github.com/OEvgeny/postcss-atroot#postcss-at-root-
        'postcss-atroot': {},

        // 提供 @extend 语法
        // @see https://github.com/jonathantneal/postcss-extend-rule#postcss-extend-rule-
        'postcss-extend-rule': {},

        // 变量相关
        // @see https://github.com/jonathantneal/postcss-advanced-variables#postcss-advanced-variables-
        'postcss-advanced-variables': {
            // variables 属性内的变量为全局变量
            // variables: require('./src/css/variable.js')
        },

        // 类似于 stylus，直接引用属性而不需要变量定义
        // @see https://github.com/simonsmith/postcss-property-lookup#postcss-property-lookup-
        'postcss-property-lookup': {}
    }
}
