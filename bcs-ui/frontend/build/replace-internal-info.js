/**
 * @file 替换 asset js 中的敏感信息
 * @author ielgnaw <wuji0223@gmail.com>
 */

const fs = require('fs')
const path = require('path')
const stream = require('stream')
const fse = require('fs-extra')
const chalk = require('chalk')

console.log(chalk.cyan('  Start Replace...\n'))

const Transform = stream.Transform

// 打包的版本
const VERSION = process.env.VERSION

// 临时目录
const TMP_DIR = path.resolve(__dirname, '..', 'dist/js-tmp')

// 构建后的 js 目录
const DIST_DIR = path.resolve(__dirname, '..', 'dist', VERSION, 'static', 'js')
const distJSFiles = []
;(function walkTpl (filePath) {
    fs.readdirSync(filePath).forEach(item => {
        if (fs.statSync(filePath + '/' + item).isDirectory()) {
            walkTpl(filePath + '/' + item)
        } else {
            const ext = path.extname(item)
            if (ext === '.js') {
                distJSFiles.push({
                    fileName: item,
                    filePath: path.resolve(__dirname, '..', 'dist', filePath + '/' + item)
                })
            }
        }
    })
})(DIST_DIR)

class ReplaceContent extends Transform {
    _transform (chunk, enc, done) {
        const reg = /((http:\/\/|ftp:\/\/|https:\/\/|\/\/)?(([^./"' \u4e00-\u9fa5（]+\.)*(oa.com|ied.com)+))/ig
        this.push(chunk.toString('utf-8').replace(reg, 'http://bking.com'))
        done()
    }
}

// const arr = [
//     {
//         fileName: 'ee.cf0a6ab228cd129dffc9.js',
//         filePath: '/Users/ielgnaw/Workspace/tencent-git/paas-bcs-webfe/package_vue/dist/ce.bak/static/js/ee.cf0a6ab228cd129dffc9.js'
//     },
//     {
//         fileName: 'vendor.b5488250a85ed8ad7ffc.js',
//         filePath: '/Users/ielgnaw/Workspace/tencent-git/paas-bcs-webfe/package_vue/dist/ce.bak/static/js/vendor.b5488250a85ed8ad7ffc.js'
//     }
// ]

const doTransform = file => {
    return new Promise((resolve, reject) => {
        const read = fs.createReadStream(file.filePath)
        read.setEncoding('utf-8').resume().pipe(new ReplaceContent(file))
            .pipe(
                fs.createWriteStream(
                    path.resolve(TMP_DIR, `${file.fileName}`)
                )
            ).on('finish', async () => {
                resolve(1)
            }).on('error', e => {
                console.error(e)
                reject(e)
            })
    })
}

fse.ensureDir(TMP_DIR, err => {
    if (err) {
        throw err
    }

    Promise.all(distJSFiles.map(async f => {
        await doTransform(f)
    })).then(async ret => {
        // await fse.copy(
        //     DIST_DIR,
        //     path.resolve(__dirname, '..', `dist/${VERSION}/static/js-origin`)
        // )
        await fse.move(
            TMP_DIR,
            DIST_DIR,
            { overwrite: true }
        )

        console.log(chalk.cyan('  Replace complete.\n'))
    })
})
