

const express = require('express')
const path = require('path')
const webpack = require('webpack')
const bodyParser = require('body-parser')
const proxyMiddleware = require('http-proxy-middleware')
const webpackHotMiddleware = require('webpack-hot-middleware')
const webpackDevMiddleware = require('webpack-dev-middleware')
const history = require('connect-history-api-fallback')

const checkVer = require('./check-versions')
const config = require('./config')
const devConf = require('./webpack.dev.conf')
const { getIP } = require('./util')

checkVer()

if (!process.env.NODE_ENV) {
    process.env.NODE_ENV = JSON.parse(config.dev.env.NODE_ENV)
}

const webpackConfig = devConf

const port = process.env.PORT || config.dev.port

const proxyTable = config.dev.proxyTable

const app = express()
const compiler = webpack(webpackConfig)

const devMiddleware = webpackDevMiddleware(compiler, {
    publicPath: webpackConfig.output.publicPath,
    quiet: true
})

const hotMiddleware = webpackHotMiddleware(compiler, {
    log: false,
    heartbeat: 2000
})

// compiler.plugin('compilation', compilation => {
//     compilation.plugin('html-webpack-plugin-after-emit', (data, cb) => {
//         // hotMiddleware.publish({action: 'reload'})
//         cb()
//     })
// })

Object.keys(proxyTable).forEach(context => {
    let options = proxyTable[context]
    if (typeof options === 'string') {
        options = {
            target: options
        }
    }
    app.use(proxyMiddleware(context, options))
})

app.use(history({
    verbose: false,
    rewrites: [
        {
            from: /(\d+\.)*\d+$/,
            to: '/'
        },
        {
            from: /\/+.*\..*\//,
            to: '/'
        }
    ]
}))

app.use(devMiddleware)

app.use(hotMiddleware)

app.use(bodyParser.json())

app.use(bodyParser.urlencoded({
    extended: true
}))

const staticPath = path.posix.join(config.dev.assetsPublicPath, config.dev.assetsSubDirectory)
app.use(staticPath, express.static('./static'))

let _resolve
const readyPromise = new Promise(resolve => {
    _resolve = resolve
})

console.log('> Starting dev server...')
devMiddleware.waitUntilValid(() => {
    console.log('Other available url: ')
    console.log(`http://${getIP()}:${port}`)
    console.log(`http://localhost:${port}\n`)
    _resolve()
})

const server = app.listen(port)

module.exports = {
    ready: readyPromise,
    close: () => {
        server.close()
    }
}
