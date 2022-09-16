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

module.exports = function (api) {
    api.cache(true)

    const presets = [
        [
            '@babel/preset-env',
            {
                'modules': 'commonjs',
                'corejs': '3.1.4',
                'targets': {
                    'browsers': ['> 1%', 'last 2 versions', 'not ie <= 8']
                },
                'debug': false,
                'useBuiltIns': 'usage'
            }
        ],
        ['@vue/babel-preset-jsx']
    ]

    const plugins = [
        'lodash',
        '@babel/plugin-transform-runtime',
        '@babel/plugin-transform-async-to-generator',
        '@babel/plugin-transform-object-assign',
        '@babel/plugin-syntax-dynamic-import',
        'date-fns',
        '@vue/babel-plugin-transform-vue-jsx',
        '@babel/plugin-syntax-jsx',
        '@babel/plugin-proposal-optional-chaining',
        '@babel/plugin-proposal-class-properties'
    ]
    const comments = true

    return {
        compact: false,
        presets,
        plugins,
        comments,
        babelrcRoots: ['./src', './bk-bcs-saas/bcs-app/frontend/src/'],
        exclude: /node_modules/
    }
}
