

const { join } = require('path')
const ora = require('ora')
const chalk = require('chalk')
const webpack = require('webpack')
const rm = require('rimraf')
const fse = require('fs-extra')

const checkVer = require('./check-versions')
const config = require('./config')
const webpackConf = require('./webpack.prod.conf')

checkVer()

// 打包的版本
const VERSION = process.env.VERSION

const isCleanHardSourceCache = process.env.CLEAN_HARD_SOURCE_CACHE

const spinner = ora(`building for ${chalk.green(VERSION)}...`)
spinner.start()

if (isCleanHardSourceCache === '1') {
    fse.removeSync(join(__dirname, '../.hard-source-cache'))
}

rm(join(config.build.assetsRoot, VERSION), e => {
    if (e) {
        throw e
    }
    webpack(webpackConf, (err, stats) => {
        spinner.stop()
        if (err) {
            throw err
        }

        process.stdout.write(stats.toString({
            colors: true,
            modules: false,
            children: false,
            chunks: false,
            chunkModules: false
        }) + '\n\n')

        if (stats.hasErrors()) {
            console.log(chalk.red('  Build failed with errors.\n'))
            process.exit(1)
        }

        console.log(chalk.cyan('  Build complete.\n'))
        console.log(chalk.yellow(
            '  Tip: built files are meant to be served over an HTTP server.\n'
            + '  Opening index.html over file:// won\'t work.\n'
        ))

        process.exit(0)
    })
})
