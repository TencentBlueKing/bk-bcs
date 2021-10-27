/**
 * @file 提取 .vue, .js 文件里的中文（非注释中的）
 * @author ielgnaw <wuji0223@gmail.com>
 */

const { readdirSync, readFileSync, writeFileSync, statSync } = require('fs')
const { resolve, basename, extname, relative } = require('path')

const vueFiles = {}
;(function walkVue (filePath) {
    const dirList = readdirSync(filePath)
    dirList.forEach(item => {
        if (statSync(filePath + '/' + item).isDirectory()) {
            walkVue(filePath + '/' + item)
        } else {
            const ext = extname(item)
            if (ext === '.vue' || ext === '.js') {
                if (!vueFiles[basename(filePath)]) {
                    vueFiles[basename(filePath)] = []
                }
                vueFiles[basename(filePath)].push(relative('.', filePath + '/' + item))
            }
        }
    })
})(resolve(__dirname, '../src'))

const JS_COMMENT_REG = /(\/\*([\s\S]*?)\*\/|([^:]|^)\/\/(.*)$)/mg
const HTML_COMMENT_REG = /(<!--[\s\S]*?-->)/mg
const CHINESE_REG = /([【】`（）》《])*[\u3400-\u4DB5\u4E00-\u9FEA\uFA0E\uFA0F\uFA11\uFA13\uFA14\uFA1F\uFA21\uFA23\uFA24\uFA27-\uFA29\u{20000}-\u{2A6D6}\u{2A700}-\u{2B734}\u{2B740}-\u{2B81D}\u{2B820}-\u{2CEA1}\u{2CEB0}-\u{2EBE0}]([【】`（）》《])*[^\n'"<]*/umg

const ret = {}
const translate = {}
let match = null

Object.keys(vueFiles).forEach(key => {
    ret[key] = {}
    vueFiles[key].forEach(file => {
        if (!ret[key][file]) {
            ret[key][file] = []
        }
        const content = readFileSync(resolve(file), 'UTF-8')
        const noCommentContent = content.replace(JS_COMMENT_REG, '').replace(HTML_COMMENT_REG, '')
        // eslint-disable-next-line no-cond-assign
        while (match = CHINESE_REG.exec(noCommentContent)) {
            ret[key][file].push(match[0])
            translate[match[0]] = ''
        }
    })
})

const retFileName = 'extract-chinese.json'
const absolutePath = resolve(__dirname, retFileName)
writeFileSync(absolutePath, JSON.stringify(ret, null, 4), 'UTF-8')

const translateFileName = 'translate.json'
const absolutePath1 = resolve(__dirname, translateFileName)
writeFileSync(absolutePath1, JSON.stringify(translate, null, 4), 'UTF-8')

// test
// const content = readFileSync(resolve(vueFileList[0].path), 'UTF-8')
// const noCommentContent = content.replace(JS_COMMENT_REG, '').replace(HTML_COMMENT_REG, '')
// console.error(noCommentContent)

// let match = null
// const ret = []
// // const r = /[\u4E00-\u9FA5\uf900-\ufa2d]\S+/gum
// // console.error(noCommentContent.match(r))
// // const r = /\p{Unified_Ideograph}/u
// while (!!(match = CHINESE_REG.exec(noCommentContent))) {
//     console.log(111, match[0])
// }
