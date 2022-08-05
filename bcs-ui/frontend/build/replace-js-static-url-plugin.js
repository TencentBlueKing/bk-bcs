

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
