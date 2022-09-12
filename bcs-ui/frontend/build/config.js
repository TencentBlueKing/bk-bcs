

const path = require('path')
const prodEnv = require('./prod.env')
const devEnv = require('./dev.env')

// 打包的版本
const VERSION = process.env.VERSION

module.exports = {
    build: {
        env: prodEnv,
        assetsRoot: path.resolve(__dirname, '../dist'),
        assetsSubDirectory: `${VERSION}/static`,
        // assetsPublicPath: '{{ STATIC_URL }}',
        assetsPublicPath: '{{STATIC_URL}}',
        // assetsPublicPath: '/',
        productionSourceMap: true,
        productionGzip: false,
        productionGzipExtensions: ['js', 'css'],
        bundleAnalyzerReport: process.env.npm_config_report
    },
    dev: {
        env: devEnv,
        port: 8004,
        assetsSubDirectory: 'static',
        assetsPublicPath: '/',
        proxyTable: {},
        cssSourceMap: false,
        autoOpenBrowser: false
    }
}
