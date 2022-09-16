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

const webpack = require('webpack')
const { VueLoaderPlugin } = require('vue-loader')
const threadLoader = require('thread-loader')
const friendlyFormatter = require('eslint-friendly-formatter')

const config = require('./config')
const { assetsPath, resolve } = require('./util')

const isProd = process.env.NODE_ENV === 'production'

const jsWorkerPool = {
    // options

    // 产生的 worker 的数量，默认是 (cpu 核心数 - 1)
    // 当 require('os').cpus() 是 undefined 时，则为 1
    workers: 2,

    // 闲置时定时删除 worker 进程
    // 默认为 500ms
    // 可以设置为无穷大， 这样在监视模式(--watch)下可以保持 worker 持续存在
    poolTimeout: 2000
}

threadLoader.warmup(jsWorkerPool, ['babel-loader'])

module.exports = {
    output: {
        path: config.build.assetsRoot,
        filename: '[name].js',
        chunkFilename: '[name].[chunkhash].js',
        publicPath: isProd ? config.build.assetsPublicPath : config.dev.assetsPublicPath
    },
    resolve: {
        // 避免层层查找，默认值为 ['node_modules']，会依次查找./node_modules、../node_modules、../../node_modules
        modules: [resolve('./src/components'), resolve('./src/components/bk-magic'), resolve('node_modules')],
        extensions: ['.ts', '.tsx', '.js', '.vue', '.json'],
        alias: {
            'vue$': 'vue/dist/vue.esm.js',
            '@': resolve('src'),
            'echarts$': 'echarts/dist/echarts.min.js',
            'echarts': 'echarts',
            'zrender$': 'zrender/dist/zrender.min.js',
            'zrender': 'zrender'
        }
    },
    cache: true,
    plugins: [
        new VueLoaderPlugin(),
        // moment 优化，只提取本地包
        new webpack.ContextReplacementPlugin(/moment\/locale$/, /zh-cn/),
        // brace 优化，只提取需要的语法
        new webpack.ContextReplacementPlugin(/brace\/mode$/, /^\.\/(json|yaml|python|sh|text)$/),
        // brace 优化，只提取需要的 theme
        new webpack.ContextReplacementPlugin(/brace\/theme$/, /^\.\/(monokai)$/),
        new webpack.ProgressPlugin()
    ],
    module: {
        noParse: [
            /\/node_modules\/echarts\/dist\/echarts\.min\.js$/,
            /\/node_modules\/zrender\/dist\/zrender\.min\.js$/
        ],
        rules: [
            {
                test: /\.(js|vue)$/,
                enforce: 'pre',
                loader: 'eslint-loader',
                include: [resolve('src'), resolve('build')],
                exclude: /node_modules\/|lib\.bundle\.js|ip-selector\.js/,
                options: {
                    formatter: friendlyFormatter
                }
            },
            {
                test: /\.tsx?$/,
                include: resolve('src'),
                exclude: /node_modules/,
                use: [
                    {
                        loader: 'babel-loader',
                        options: {
                            cacheDirectory: './webpack_cache/'
                        }
                    },
                    {
                        loader: 'ts-loader',
                        options: {
                            appendTsxSuffixTo: [/\.vue$/],
                            transpileOnly: true
                        }
                    }
                ]
            },
            {
                test: /\.vue$/,
                loader: 'vue-loader',
                options: {
                    include: [resolve('src')],
                    transformToRequire: {
                        video: 'src',
                        source: 'src',
                        img: 'src',
                        image: 'xlink:href'
                    }
                }
            },
            {
                test: /\.js$/,
                exclude: /node_modules/,
                // use: {
                //     loader: 'babel-loader',
                //     options: {
                //         include: [
                //             resolve('src'),
                //             resolve('bk-bcs-saas/bcs-app/frontend/src'),
                //             resolve('/node_modules/monaco-editor/esm')
                //         ],
                //         cacheDirectory: './webpack_cache/',
                //         plugins: [
                //             'dynamic-import-webpack'
                //         ]
                //     }
                // }
                use: [
                    {
                        loader: 'thread-loader',
                        options: jsWorkerPool
                    },
                    {
                        loader: 'babel-loader',
                        options: {
                            include: [
                                resolve('src'),
                                resolve('/node_modules/monaco-editor/esm')
                            ],
                            cacheDirectory: './webpack_cache/',
                            plugins: [
                                'dynamic-import-webpack'
                            ]
                        }
                    }
                ]
            },
            {
                test: /\.(png|jpe?g|gif|svg)(\?.*)?$/,
                loader: 'url-loader',
                options: {
                    limit: 10000,
                    name: assetsPath('img/[name].[hash:7].[ext]')
                }
            },
            {
                test: /\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/,
                loader: 'url-loader',
                options: {
                    limit: 10000,
                    name: assetsPath('media/[name].[hash:7].[ext]')
                }
            },
            {
                test: /\.(woff2?|eot|ttf|otf)(\?.*)?$/,
                // test: /\.(woff2?|eot|ttf|otf)(\?\S*)?$/,
                loader: 'url-loader',
                options: {
                    limit: 10000,
                    name: assetsPath('fonts/[name].[hash:7].[ext]')
                }
            },
            {
                test: /\.md$/,
                loader: "text-loader"
            }
        ]
    }
}
