

const { extname } = require('path')

class ReplaceCSSStaticUrlPlugin {
    apply (compiler, callback) {
        // emit: 在生成资源并输出到目录之前
        compiler.hooks.emit.tapAsync('ReplaceCSSStaticUrlPlugin', (compilation, callback) => {
            const assets = Object.keys(compilation.assets)
            const assetsLen = assets.length

            for (let i = 0; i < assetsLen; i++) {
                const fileName = assets[i]
                if (extname(fileName) !== '.css') {
                    continue
                }

                const asset = compilation.assets[fileName]

                const minifyFileContent = asset.source().replace(
                    /\{\{STATIC_URL\}\}/g,
                    () => '../../../'
                )
                // 设置输出资源
                compilation.assets[fileName] = {
                    // 返回文件内容
                    source: () => minifyFileContent,
                    // 返回文件大小
                    size: () => Buffer.byteLength(minifyFileContent, 'utf8')
                }
            }

            callback()
        })
    }
}

module.exports = ReplaceCSSStaticUrlPlugin
