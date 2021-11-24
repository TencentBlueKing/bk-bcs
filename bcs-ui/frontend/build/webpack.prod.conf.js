/**
 * @file webpack prod config
 * @author ielgnaw <wuji0223@gmail.com>
 */

const { resolve, sep, join } = require('path')
const webpack = require('webpack')
const merge = require('webpack-merge')
const CopyWebpackPlugin = require('copy-webpack-plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const OptimizeCSSPlugin = require('optimize-css-assets-webpack-plugin')
const CompressionWebpackPlugin = require('compression-webpack-plugin')
const bundleAnalyzer = require('webpack-bundle-analyzer')
const LodashModuleReplacementPlugin = require('lodash-webpack-plugin')
const TerserPlugin = require('terser-webpack-plugin')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const MonacoEditorPlugin = require('monaco-editor-webpack-plugin')
// const SpeedMeasurePlugin = require('speed-measure-webpack-plugin')
// const SentryPlugin = require('webpack-sentry-plugin')
const threadLoader = require('thread-loader')
const HardSourceWebpackPlugin = require('hard-source-webpack-plugin')

const config = require('./config')
const baseWebpackConfig = require('./webpack.base.conf')
const { assetsPath } = require('./util')
const manifest = require('../static/lib-manifest.json')
const ReplaceJSStaticUrlPlugin = require('./replace-js-static-url-plugin')
const ReplaceCSSStaticUrlPlugin = require('./replace-css-static-url-plugin')
const MomentLocalesPlugin = require('moment-locales-webpack-plugin')

const cssWorkerPool = {
    // 一个 worker 进程中并行执行工作的数量
    // 默认为 20
    workerParallelJobs: 2,
    poolTimeout: 2000
}
threadLoader.warmup(cssWorkerPool, ['css-loader', 'postcss-loader'])

// const smp = new SpeedMeasurePlugin()

const NOW = new Date()
const RELEASE_VERSION = [NOW.getFullYear(), '-', (NOW.getMonth() + 1), '-', NOW.getDate(), '_', NOW.getHours(), ':', NOW.getMinutes(), ':', NOW.getSeconds()].join('')

// 打包的版本
const VERSION = process.env.VERSION
console.log('VERSION', VERSION)

const webpackConfig = merge(baseWebpackConfig, {
    mode: 'production',
    entry: {
        [`${VERSION}`]: `./src/main.js`
        // 'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
        // 'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
        // 'css.worker': 'monaco-editor/esm/vs/language/css/css.worker',
        // 'html.worker': 'monaco-editor/esm/vs/language/html/html.worker',
        // 'ts.worker': 'monaco-editor/esm/vs/language/typescript/ts.worker'
    },
    output: {
        path: config.build.assetsRoot,
        filename: assetsPath('js/[name].[chunkhash].js'),
        chunkFilename: assetsPath('js/[name].[chunkhash].js')
    },
    module: {
        rules: [
            {
                test: /\.(css|postcss)?$/,
                // use: [
                //     MiniCssExtractPlugin.loader,
                //     {
                //         loader: 'css-loader',
                //         options: {
                //             sourceMap: config.build.cssSourceMap,
                //             importLoaders: 1
                //         }
                //     },
                //     {
                //         loader: 'postcss-loader',
                //         options: {
                //             sourceMap: config.build.cssSourceMap,
                //             config: {
                //                 path: resolve(__dirname, '..', 'postcss.config.js')
                //             }
                //         }
                //     }
                // ]
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: 'thread-loader',
                        options: cssWorkerPool
                    },
                    {
                        loader: 'css-loader',
                        options: {
                            sourceMap: config.build.cssSourceMap,
                            importLoaders: 1
                        }
                    },
                    {
                        loader: 'postcss-loader',
                        options: {
                            sourceMap: config.build.cssSourceMap,
                            postcssOptions: {
                                config: resolve(__dirname, '..', 'postcss.config.js')
                            }
                        }
                    }
                ]
            },
            {
                test: /\.s[ac]ss$/i,
                use: [
                    MiniCssExtractPlugin.loader,
                    // Translates CSS into CommonJS
                    {
                        loader: 'css-loader',
                        options: {
                            importLoaders: 1
                        }
                    },
                    // Compiles Sass to CSS
                    'sass-loader'
                ]
            }
        ]
    },
    optimization: {
        runtimeChunk: {
            name: 'manifest'
        },
        minimizer: [
            new TerserPlugin({
                terserOptions: {
                    compress: false,
                    mangle: true,
                    output: {
                        comments: false
                    }
                },
                cache: true,
                parallel: true,
                sourceMap: true
            }),
            new OptimizeCSSPlugin({
                cssProcessorOptions: {
                    safe: true
                }
            })
        ],
        splitChunks: {
            // 表示从哪些 chunks 里面提取代码，除了三个可选字符串值 initial、async、all 之外，还可以通过函数来过滤所需的 chunks
            // async: 针对异步加载的 chunk 做分割，默认值
            // initial: 针对同步 chunk
            // all: 针对所有 chunk
            chunks: 'all',
            // 表示提取出来的文件在压缩前的最小大小，默认为 30kb
            minSize: 30000,
            // 表示提取出来的文件在压缩前的最大大小，默认为 0，表示不限制最大大小
            maxSize: 0,
            // 表示被引用次数，默认为 1
            minChunks: 1,
            // 最多有 5 个异步加载请求该 module
            maxAsyncRequests: 5,
            // 初始化的时候最多有 3 个请求该 module
            maxInitialRequests: 3,
            // 名字中间的间隔符
            automaticNameDelimiter: '~',
            // chunk 的名字，如果设成 true，会根据被提取的 chunk 自动生成
            name: true,
            // 要切割成的每一个新 chunk 就是一个 cache group，缓存组会继承 splitChunks 的配置，但是 test, priorty 和 reuseExistingChunk 只能用于配置缓存组。
            // test: 和 CommonsChunkPlugin 里的 minChunks 非常像，用来决定提取哪些 module，可以接受字符串，正则表达式，或者函数
            //      函数的一个参数为 module，第二个参数为引用这个 module 的 chunk（数组）
            // priority: 表示提取优先级，数字越大表示优先级越高。因为一个 module 可能会满足多个 cacheGroups 的条件，那么提取到哪个就由权重最高的说了算；
            //          优先级高的 chunk 为被优先选择，优先级一样的话，size 大的优先被选择
            // reuseExistingChunk: 表示是否使用已有的 chunk，如果为 true 则表示如果当前的 chunk 包含的模块已经被提取出去了，那么将不会重新生成新的。
            cacheGroups: {
                // 提取 chunk-bk-magic-vue 代码块
                bkMagic: {
                    chunks: 'all',
                    // 单独将 bkMagic 拆包
                    name: 'chunk-bk-magic-vue',
                    // 权重
                    priority: 5,
                    // 表示是否使用已有的 chunk，如果为 true 则表示如果当前的 chunk 包含的模块已经被提取出去了，那么将不会重新生成新的。
                    reuseExistingChunk: true,
                    test: module => {
                        return /bk-magic-vue/.test(module.context)
                    }
                },
                // 所有 node_modules 的模块被不同的 chunk 引入超过 1 次的提取为 twice
                // 如果去掉 test 那么提取的就是所有模块被不同的 chunk 引入超过 1 次的
                twice: {
                    // test: /[\\/]node_modules[\\/]/,
                    chunks: 'all',
                    name: 'twice',
                    priority: 6,
                    minChunks: 2,
                    reuseExistingChunk: true
                },
                // default 和 vendors 是默认缓存组，可通过 optimization.splitChunks.cacheGroups.default: false 来禁用
                default: {
                    minChunks: 2,
                    priority: -20,
                    reuseExistingChunk: true
                },
                vendors: {
                    test: /[\\/]node_modules[\\/]/,
                    priority: -10,
                    reuseExistingChunk: true
                },
                // 单独拆分monaco
                monaco: {
                    name: 'chunk-monaco-editor',
                    priority: 7,
                    reuseExistingChunk: true,
                    test: module => {
                        return /monaco-editor/.test(module.context)
                    }
                },
                // 拆分echarts
                echarts: {
                    name: 'chunk-echarts',
                    priority: 7,
                    reuseExistingChunk: true,
                    test: module => {
                        return /echarts/.test(module.context)
                    }
                }
            }
        }
    },
    devtool: (VERSION === 'ieod' && config.build.productionSourceMap) ? '#source-map' : false,
    plugins: [
        new webpack.DefinePlugin(config.build.env),

        new webpack.DllReferencePlugin({
            context: __dirname,
            manifest: manifest
        }),

        new MiniCssExtractPlugin({
            filename: assetsPath('css/[name].[contenthash].css')
        }),

        new LodashModuleReplacementPlugin(),

        new HtmlWebpackPlugin({
            filename: resolve(config.build.assetsRoot + sep + config.build.assetsSubDirectory, '..') + '/index.html',
            template: 'index.html',
            inject: true,
            minify: {
                removeComments: true,
                collapseWhitespace: true
                // removeAttributeQuotes: true
            },
            // 如果打开 vendor 和 manifest 那么需要配置 chunksSortMode 保证引入 script 的顺序
            // chunksSortMode: 'dependency',
            staticUrl: config.build.env.staticUrl,
            releaseVersion: RELEASE_VERSION
        }),

        new MonacoEditorPlugin({
            // https://github.com/Microsoft/monaco-editor-webpack-plugin#options
            // Include a subset of languages support
            // Some language extensions like typescript are so huge that may impact build performance
            // e.g. Build full languages support with webpack 4.0 takes over 80 seconds
            // Languages are loaded on demand at runtime
            output: `${VERSION}/static/`,
            languages: ['javascript', 'html', 'css', 'json', 'shell', 'yaml']
        }),

        new CopyWebpackPlugin([
            {
                from: resolve(__dirname, '../static'),
                to: config.build.assetsSubDirectory,
                ignore: ['.*']
            }
        ]),

        new CopyWebpackPlugin([
            {
                from: resolve(__dirname, '../login_success.html'),
                to: `${VERSION}`,
                ignore: ['.*']
            }
        ]),

        new ReplaceJSStaticUrlPlugin({}),
        new ReplaceCSSStaticUrlPlugin({}),
        // new ReplaceInternalInfo({})

        new HardSourceWebpackPlugin({
            // cacheDirectory是在高速缓存写入。默认情况下，将缓存存储在node_modules下的目录中
            // 'node_modules/.cache/hard-source/[confighash]'
            cacheDirectory: join(__dirname, '../.hard-source-cache/[confighash]'),
            // configHash在启动webpack实例时转换webpack配置，
            // 并用于cacheDirectory为不同的webpack配置构建不同的缓存
            configHash: (webpackConfig) => {
                // node-object-hash on npm can be used to build this.
                return require('node-object-hash')({ sort: false }).hash(webpackConfig)
            },
            // 当加载器、插件、其他构建时脚本或其他动态依赖项发生更改时，
            // hard-source需要替换缓存以确保输出正确。
            // environmentHash被用来确定这一点。如果散列与先前的构建不同，则将使用新的缓存
            environmentHash: {
                root: process.cwd(),
                directories: [],
                files: ['package-lock.json', 'yarn.lock']
            },
            // An object. 控制来源
            info: {
                // 'none' or 'test'.
                mode: 'none',
                // 'debug', 'log', 'info', 'warn', or 'error'.
                level: 'debug'
            },
            // Clean up large, old caches automatically.
            cachePrune: {
                // Caches younger than `maxAge` are not considered for deletion. They must
                // be at least this (default: 2 days) old in milliseconds.
                maxAge: 2 * 24 * 60 * 60 * 1000,
                // All caches together must be larger than `sizeThreshold` before any
                // caches will be deleted. Together they must be at least this
                // (default: 50 MB) big in bytes.
                // 要删除缓存，所有缓存的总和必须超过此阈值
                sizeThreshold: 50 * 1024 * 1024
            }
        }),
        new HardSourceWebpackPlugin.ExcludeModulePlugin([
            {
                // HardSource works with mini-css-extract-plugin but due to how
                // mini-css emits assets, assets are not emitted on repeated builds with
                // mini-css and hard-source together. Ignoring the mini-css loader
                // modules, but not the other css loader modules, excludes the modules
                // that mini-css needs rebuilt to output assets every time.
                test: /mini-css-extract-plugin[\\/]dist[\\/]loader/
            },
            {
                test: /url-loader/
            },
            {
                test: /.*\.DS_Store/
            }
        ]),
        new MomentLocalesPlugin({
            localesToKeep: ['zh-cn']
        })
    ]
})

if (config.build.productionGzip) {
    webpackConfig.plugins.push(
        new CompressionWebpackPlugin({
            asset: '[path].gz[query]',
            algorithm: 'gzip',
            test: new RegExp('\\.(' + config.build.productionGzipExtensions.join('|') + ')$'),
            threshold: 10240,
            minRatio: 0.8
        })
    )
}

if (config.build.bundleAnalyzerReport) {
    const BundleAnalyzerPlugin = bundleAnalyzer.BundleAnalyzerPlugin
    webpackConfig.plugins.push(new BundleAnalyzerPlugin(
        {
            //  可以是`server`，`static`或`disabled`。
            //  在`server`模式下，分析器将启动HTTP服务器来显示软件包报告。
            //  在“静态”模式下，会生成带有报告的单个HTML文件。
            //  在`disabled`模式下，你可以使用这个插件来将`generateStatsFile`设置为`true`来生成Webpack Stats JSON文件。
            analyzerMode: 'server',
            //  将在“服务器”模式下使用的主机启动HTTP服务器。
            analyzerHost: '127.0.0.1',
            //  将在“服务器”模式下使用的端口启动HTTP服务器。
            analyzerPort: 8888,
            //  路径捆绑，将在`static`模式下生成的报告文件。
            //  相对于捆绑输出目录。
            reportFilename: 'report.html',
            //  模块大小默认显示在报告中。
            //  应该是`stat`，`parsed`或者`gzip`中的一个。
            //  有关更多信息，请参见“定义”一节。
            defaultSizes: 'parsed',
            //  在默认浏览器中自动打开报告
            openAnalyzer: true,
            //  如果为true，则Webpack Stats JSON文件将在bundle输出目录中生成
            generateStatsFile: true,
            //  如果`generateStatsFile`为`true`，将会生成Webpack Stats JSON文件的名字。
            //  相对于捆绑输出目录。
            statsFilename: 'stats.json',
            //  stats.toJson（）方法的选项。
            //  例如，您可以使用`source：false`选项排除统计文件中模块的来源。
            //  在这里查看更多选项：https：  //github.com/webpack/webpack/blob/webpack-1/lib/Stats.js#L21
            statsOptions: null,
            logLevel: 'info' //日志级别。可以是'信息'，'警告'，'错误'或'沉默'。
        }
    ))
}

module.exports = webpackConfig
// module.exports = smp.wrap(webpackConfig)
