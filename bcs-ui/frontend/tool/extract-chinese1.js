const { readdirSync, readFileSync, writeFileSync, statSync } = require('fs')
const { resolve, basename, extname, relative } = require('path')
// const originZh = require('./translate-zh.json')
// const origin = require('./translate.json')
// const originPath = require('./extract-chinese.json')
const vueFiles = {}
;(function walkVue (filePaths) {
    filePaths.forEach(filePath => {
        const dirList = readdirSync(filePath)
        dirList.forEach(item => {
            if (statSync(filePath + '/' + item).isDirectory()) {
                walkVue([filePath + '/' + item])
            } else {
                const ext = extname(item)
                if (ext === '.vue' || ext === '.js' || ext === '.html') {
                    if (!vueFiles[basename(filePath)]) {
                        vueFiles[basename(filePath)] = []
                    }
                    vueFiles[basename(filePath)].push(relative('.', filePath + '/' + item))
                }
            }
        })
    })
})([resolve(__dirname, '../src/views/app/k8s/11')])

const JS_COMMENT_REG = /(\/\*([\s\S]*?)\*\/|([^:]|^)\/\/(.*)$)/mg
const HTML_COMMENT_REG = /(<!--[\s\S]*?-->)/mg
const CHINESE_MARK = `[\\u4E00-\\u9FA5\\uF900-\\uFA2D]`
const CHINESE_TEXT = `[\\u3002\\uff1f\\uff01\\uff0c\\u3001\\uff1b\\uff1a\\u201c\\u201d\\u2018\\u2019\\uff08\\uff09\\u300a\\u300b\\u3008\\u3009\\u3010\\u3011\\u300e\\u300f\\u300c\\u300d\\ufe43\\ufe44\\u3014\\u3015\\u2026\\u2014\\uff5e\\ufe4f\\uffe5\\u4E00-\\u9FA5\\uF900-\\uFA2D]+[^\'\"\`\<]*[\\u3002\\uff1f\\uff01\\uff0c\\u3001\\uff1b\\uff1a\\u201c\\u201d\\u2018\\u2019\\uff08\\uff09\\u300a\\u300b\\u3008\\u3009\\u3010\\u3011\\u300e\\u300f\\u300c\\u300d\\ufe43\\ufe44\\u3014\\u3015\\u2026\\u2014\\uff5e\\ufe4f\\uffe5\\u4E00-\\u9FA5\\uF900-\\uFA2D]+`
const CHINESE_REG = /([【】`（）》《])*[\u3400-\u4DB5\u4E00-\u9FEA\uFA0E\uFA0F\uFA11\uFA13\uFA14\uFA1F\uFA21\uFA23\uFA24\uFA27-\uFA29\u{20000}-\u{2A6D6}\u{2A700}-\u{2B734}\u{2B740}-\u{2B81D}\u{2B820}-\u{2CEA1}\u{2CEB0}-\u{2EBE0}]([【】`（）》《])*[^\n`'"<]*/umg
const ret = {}
// let x = /\s([a-zA-Z]\w+)=['"]+(\w*[\u3002\uff1f\uff01\uff0c\u3001\uff1b\uff1a\u201c\u201d\u2018\u2019\uff08\uff09\u300a\u300b\u3008\u3009\u3010\u3011\u300e\u300f\u300c\u300d\ufe43\ufe44\u3014\u3015\u2026\u2014\uff5e\ufe4f\uffe5\u4E00-\u9FA5\uF900-\uFA2D]+[^'"`<]*[\(\)\,\/\.\*%\$@\!\-\+\;\?]*[\u3002\uff1f\uff01\uff0c\u3001\uff1b\uff1a\u201c\u201d\u2018\u2019\uff08\uff09\u300a\u300b\u3008\u3009\u3010\u3011\u300e\u300f\u300c\u300d\ufe43\ufe44\u3014\u3015\u2026\u2014\uff5e\ufe4f\uffe5\u4E00-\u9FA5\uF900-\uFA2D]+\w*)['"]+/
const translate = {}
const translateZh = {}
Object.keys(vueFiles).forEach(key => {
    ret[key] = {}
    const text = CHINESE_TEXT
    const reg1 = new RegExp('\\s+([a-zA-Z][\\w\\-]*)\\s*=\\s*([\'\"\`])+\\s*(\\w*' + text + '\\w*)\\s*(\\2)', 'gu')
    const reg2 = new RegExp('>\\s*([\\w\\(\\)...]*' + text + '[\\w\\(\\)...!/\\s\\?]*)\\s*([<\\{])', 'gu')
    const reg3 = new RegExp('\\s*\\}\\s*(\\w*' + text + '\\w*)\\s*[\"\'\`]*\\s*\\{', 'gu')
    const reg4 = new RegExp('\\s*\\}\\s*([\\w]*' + text + '\\w*)\\s*<', 'gu')
    const reg5 = new RegExp('\\s([a-zA-Z]+):\\s+[\'\"\`]\\s*([a-zA-Z0-9]*' + text + '[a-zA-Z0-9]*)\\s*[\'\"\`]', 'gu')
    const reg6 = new RegExp('([\'\"\`])\\s*([\\w\\(\\)...]*' + text + '[\\w\\(\\)...]*)\\s*(\\1)', 'gu')
    const reg7 = new RegExp('([\'\"\`])\\s*(\\w*' + CHINESE_MARK + '\\w*)\\s*(\\1)', 'gu')
    const reg8 = new RegExp('\`\\s*([^\`<\'\"]' + text + '[^\`<\'\"]*)\\s*\\$\\{[^\`<]*\`')
    const markReg = /([\!\$\?\+\-\]\[\{\}\(\)])/gmi
    const is$t = (text, val) => new RegExp('\\$t\\(([\'\"])' + text.replace(markReg, '\\$1') + '(\\1)', 'gm').test(val)
    vueFiles[key].forEach(file => {
        if (!ret[key][file]) {
            ret[key][file] = []
        }
        const content = readFileSync(resolve(file), 'UTF-8')
        if (!content.match(/<i18n>/gmi)) {
            let tmpContent = content
            const noCommentContent = content.replace(JS_COMMENT_REG, '').replace(HTML_COMMENT_REG, '')

            let result = null

            let noCommentTemplate = noCommentContent.replace(/(<script>[\s\S]*?<\/script>)/gm, '').replace(/<style[\s\S]*<\/style>/, '')
            if (noCommentTemplate.length) {
                // eslint-disable-next-line no-cond-assign
                while (result = reg1.exec(noCommentTemplate)) {
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                    if (!new RegExp(':' + result[1] + '\\s*=\\s*"\\$t\\(\\s*([\'\"\`])' + result[3] + '(\\1)\\)"').test(tmpContent)) {
                        tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), ` :${result[1]}="$t('${result[3]}')"`)
                    }
                    ret[key][file].push(result[3])
                    translate[result[3]] = ''
                    translateZh[result[3]] = result[3]
                }
                // eslint-disable-next-line no-cond-assign
                while (result = reg4.exec(noCommentTemplate)) {
                    const check = /([^\{]*)(\{\{[\s\S]*\}\})([^\}]*)/gm.exec(result[1])
                    if (check) {
                        if (!is$t(result[1], tmpContent)) {
                            tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), `}{{$t('${check[1]}')}}${check[2]}{{$t('${check[3]}')}}<`)
                        }
                        ret[key][file].push(check[1])
                        translate[check[1]] = ''
                        translateZh[check[1]] = check[1]
                        ret[key][file].push(check[3])
                        translate[check[3]] = ''
                        translateZh[check[3]] = check[3]
                    } else {
                        if (!new RegExp('}\\s+{{\\s*\\$t\\(\\s*([\'\"\`])' + result[1] + '(\\1)\\s*}}\\s+<', 'gm').test(tmpContent)) {
                            tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), `}{{$t('${result[1]}')}}<`)
                        }
                        ret[key][file].push(result[1])
                        translate[result[1]] = ''
                        translateZh[result[1]] = result[1]
                    }
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                }

                const checkResult2 = noCommentTemplate.match(reg2)
                // eslint-disable-next-line no-cond-assign
                while (result = reg2.exec(noCommentTemplate)) {
                    const check = /([^\{]*)(\{\{[\s\S]*\}\})([^\}]*)/gm.exec(result[1])
                    if (check) {
                        if (!is$t(result[1], tmpContent)) {
                            tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), `>{{$t('${check[1]}')}}${check[2]}{{$t('${check[3]}')}}${result[2]}`)
                        }
                        ret[key][file].push(check[1])
                        translate[check[1]] = ''
                        translateZh[check[1]] = check[1]
                        ret[key][file].push(check[3])
                        translate[check[3]] = ''
                        translateZh[check[3]] = check[3]
                    } else {
                        if (!new RegExp('>\\s+{{\\s*\\$t\\(\\s*([\'\"\`])' + result[1] + '(\\1)\\s*}}\\s+' + result[2], 'gm').test(tmpContent)) {
                            tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), `>{{$t('${result[1]}')}}${result[2]}`)
                        }
                        ret[key][file].push(result[1])
                        translate[result[1]] = ''
                        translateZh[result[1]] = result[1]
                    }
                    if (checkResult2 && checkResult2.length) {
                        const index = checkResult2.findIndex(item => item === result[0])
                        if (index > -1) {
                            checkResult2.splice(index, 1)
                        }
                    }
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                }
                if (checkResult2 && checkResult2.length) {
                    checkResult2.forEach(item => {
                        const check = new RegExp('>\\s*([\\w\\(\\)...]*' + text + '[\\w\\(\\)...]*)\\s*([<\\{])', 'gu').exec(item)
                        if (check) {
                            tmpContent = tmpContent.replace(new RegExp(item.replace(markReg, '\\$1'), 'gm'), `>{{$t('${check[1]}')}}<`)
                            ret[key][file].push(check[1])
                            translate[check[1]] = ''
                            translateZh[check[1]] = check[1]
                        }
                    })
                }
                // eslint-disable-next-line no-cond-assign
                while (result = reg3.exec(noCommentTemplate)) {
                    if (!new RegExp('}\\s+{{\\s*\\$t\\(\\s*([\'\"\`])' + result[1] + '(\\1)\\s*}}\\s+{', 'gm').test(tmpContent)) {
                        tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), `}{{$t('${result[1]}')}}{`)
                    }
                    ret[key][file].push(result[1])
                    translate[result[1]] = ''
                    translateZh[result[1]] = result[1]
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                }
                // eslint-disable-next-line no-cond-assign
                while (result = reg5.exec(noCommentTemplate)) {
                    if (!new RegExp('\\s+' + result[1] + ':\\s+this\\.\\$t\\(([\'\"\`])' + result[2] + '(\\1)\\)', 'gm').test(tmpContent)) {
                        tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), ` ${result[1]}: $t('${result[2]}') `)
                    }
                    ret[key][file].push(result[2])
                    translate[result[2]] = ''
                    translateZh[result[2]] = result[2]
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                }

                const checkResult = noCommentTemplate.match(reg6)
                // eslint-disable-next-line no-cond-assign
                while (result = reg6.exec(noCommentTemplate)) {
                    if (!is$t(result[2], tmpContent)) {
                        if (result[0].includes('`')) {
                            tmpContent = tmpContent.replace(new RegExp(result[2], 'gm'), `\$\{$t('${result[2]}')\}`)
                        } else {
                            if (new RegExp('\\s*([\'\"\`])\\s*' + result[2] + '\\s*(\\1)\\s*').test(result[0])) {
                                tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), `$t('${result[2]}')`)
                            } else {
                                tmpContent = tmpContent.replace(new RegExp(result[2], 'gm'), `$t('${result[2]}')`)
                            }
                        }
                    }
                    if (checkResult && checkResult.length) {
                        const index = checkResult.findIndex(item => item === result[0])
                        if (index > -1) {
                            checkResult.splice(index, 1)
                        }
                    }
                    // console.log(result[0])
                    ret[key][file].push(result[2])
                    translate[result[2]] = ''
                    translateZh[result[2]] = result[2]
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                }

                if (checkResult && checkResult.length) {
                    checkResult.forEach(item => {
                        if (item.includes('`')) {
                            tmpContent = tmpContent.replace(new RegExp(item, 'gm'), `\$\{$t('${item}')\}`)
                        } else {
                            tmpContent = tmpContent.replace(new RegExp(item, 'gm'), `$t(${item})`)
                        }
                        ret[key][file].push(item)
                        translate[item] = ''
                        translateZh[item] = item
                    })
                }

                // eslint-disable-next-line no-cond-assign
                while (result = reg7.exec(noCommentTemplate)) {
                    if (!is$t(result[2], tmpContent)) {
                        if (result[2].includes('`')) {
                            tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), `\$\{$t('${result[2]}')\}`)
                        } else {
                            tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), `$t('${result[2]}')`)
                        }
                    }
                    ret[key][file].push(result[2])
                    translate[result[2]] = ''
                    translateZh[result[2]] = result[2]
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                }
                // eslint-disable-next-line no-cond-assign
                while (result = reg8.exec(noCommentTemplate)) {
                    const check = /([^\$]*)(\$\{[^\}]*\})([^\$\`]*)([^\`]*)\`/gm.exec(result[0])
                    if (check) {
                        if (!is$t(check[1], tmpContent)) {
                            if (!check[3]) {
                                tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), ` \`\$\{$t('${check[1]}')\}${check[2]}\``)
                                ret[key][file].push(check[1])
                                translate[check[1]] = ''
                                translateZh[check[1]] = check[1]
                            } else {
                                tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), ` \`\$\{$t('${check[1]}')\}${check[2]}\$\{$t('${check[3]}')\}${check[4]}\``)
                                ret[key][file].push(check[1])
                                translate[check[1]] = ''
                                translateZh[check[1]] = check[1]
                                ret[key][file].push(check[3])
                                translate[check[3]] = ''
                                translateZh[check[3]] = check[3]
                            }
                        }
                    }
                    noCommentTemplate = noCommentTemplate.replace(result[0], '')
                }
            }

            let noCommentScrpit = noCommentContent.match(/(<script>[\s\S]*?<\/script>)/gm)
            if (noCommentScrpit) {
                noCommentScrpit = noCommentScrpit[0]
                // eslint-disable-next-line no-cond-assign
                while (result = reg5.exec(noCommentScrpit)) {
                    if (!new RegExp('\\s+' + result[1] + ':\\s+this\\.\\$t\\(([\'\"\`])' + result[2] + '(\\1)\\)', 'gm').test(tmpContent)) {
                        tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), ` ${result[1]}: this.$t('${result[2]}') `)
                    }
                    ret[key][file].push(result[2])
                    translate[result[2]] = ''
                    translateZh[result[2]] = result[2]
                    noCommentScrpit = noCommentScrpit.replace(result[0], '')
                }
                // console.info(reg6)
                const checkResult = noCommentScrpit.match(reg6)

                // eslint-disable-next-line no-cond-assign
                while (result = reg6.exec(noCommentScrpit)) {
                    if (result[2].match(/\$\{.*\}/gm)) {
                        const check = /([^\$]*)(\$\{[^\}]*\})([^\$\`]*)([^\`]*)\`/gm.exec(result[0])
                        if (check) {
                            if (!is$t(check[1], tmpContent)) {
                                if (!check[3]) {
                                    tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), ` \`\$\{this.$t('${check[1]}')\}${check[2]}\``)
                                    ret[key][file].push(check[1])
                                    translate[check[1]] = ''
                                    translateZh[check[1]] = check[1]
                                } else {
                                    tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), ` \`\$\{this.$t('${check[1]}')\}${check[2]}\$\{this.$t('${check[3]}')\}${check[4]}\``)
                                    ret[key][file].push(check[1])
                                    translate[check[1]] = ''
                                    translateZh[check[1]] = check[1]
                                    ret[key][file].push(check[3])
                                    translate[check[3]] = ''
                                    translateZh[check[3]] = check[3]
                                }
                            }
                        }
                    } else {
                        if (!is$t(result[2], tmpContent)) {
                            if (result[0].includes('`')) {
                                tmpContent = tmpContent.replace(new RegExp(result[2], 'gm'), `\$\{this.$t('${result[2]}')\}`)
                            } else {
                                if (new RegExp('\\s*([\'\"\`])\\s*' + result[2] + '\\s*(\\1)\\s*').test(result[0])) {
                                    tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), `this.$t('${result[2]}')`)
                                } else {
                                    tmpContent = tmpContent.replace(new RegExp(result[2], 'gm'), `this.$t('${result[2]}')`)
                                }
                            }
                        }
                        ret[key][file].push(result[2])
                        translate[result[2]] = ''
                        translateZh[result[2]] = result[2]
                    }
                    if (checkResult && checkResult.length) {
                        const index = checkResult.findIndex(item => item === result[0])
                        if (index > -1) {
                            checkResult.splice(index, 1)
                        }
                    }
                    noCommentScrpit = noCommentScrpit.replace(result[0], '')
                }

                if (checkResult && checkResult.length) {
                    checkResult.forEach(item => {
                        if (item.includes('`')) {
                            tmpContent = tmpContent.replace(new RegExp(item, 'gm'), `\$\{this.$t('${item}')\}`)
                        } else {
                            tmpContent = tmpContent.replace(new RegExp(item, 'gm'), `this.$t(${item})`)
                        }
                        ret[key][file].push(item)
                        translate[item] = ''
                        translateZh[item] = item
                    })
                }
                // eslint-disable-next-line no-cond-assign
                while (result = reg7.exec(noCommentScrpit)) {
                    if (!is$t(result[2], tmpContent)) {
                        if (result[2].includes('`')) {
                            tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), `\$\{this.$t('${result[2]}')\}`)
                        } else {
                            tmpContent = tmpContent.replace(new RegExp(result[0], 'gm'), `this.$t('${result[2]}')`)
                        }
                    }
                    ret[key][file].push(result[2])
                    translate[result[2]] = ''
                    translateZh[result[2]] = result[2]
                    noCommentScrpit = noCommentScrpit.replace(result[0], '')
                }
                // eslint-disable-next-line no-cond-assign
                while (result = reg8.exec(noCommentScrpit)) {
                    const check = /([^\$]*)(\$\{[^\}]*\})([^\$\`]*)([^\`]*)\`/gm.exec(result[0])
                    if (check) {
                        if (!is$t(check[1], tmpContent)) {
                            if (!check[3]) {
                                tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), ` \`\$\{this.$t('${check[1]}')\}${check[2]}\``)
                                ret[key][file].push(check[1])
                                translate[check[1]] = ''
                                translateZh[check[1]] = check[1]
                            } else {
                                tmpContent = tmpContent.replace(new RegExp(result[0].replace(markReg, '\\$1'), 'gm'), ` \`\$\{this.$t('${check[1]}')\}${check[2]}\$\{this.$t('${check[3]}')\}${check[4]}\``)
                                ret[key][file].push(check[1])
                                translate[check[1]] = ''
                                translateZh[check[1]] = check[1]
                                ret[key][file].push(check[3])
                                translate[check[3]] = ''
                                translateZh[check[3]] = check[3]
                            }
                        }
                    }
                    noCommentScrpit = noCommentScrpit.replace(result[0], '')
                }
                // eslint-disable-next-line no-cond-assign
                if (result = CHINESE_REG.exec(noCommentScrpit)) {
                    console.info(result[1], '----')
                }
            }
            writeFileSync(resolve(file), tmpContent)
        }
    })
})
// const mergePathObject = (oldObj, newObj) => {
//     // Object.keys(newObj).forEach(key => {
//     //     if (Array.isArray(oldObj[key])) {
//     //         if (!oldObj[key] || oldObj[key].length < 1) {
//     //             oldObj[key] = newObj[key]
//     //         } else {
//     //             oldObj[key] = Array.from(new Set(oldObj[key].concat(newObj[key])))
//     //         }
//     //     }
//     // })
//     return oldObj
// }
// const retFileName = 'extract-chinese.json'
// const absolutePath = resolve(__dirname, retFileName)
// writeFileSync(absolutePath, JSON.stringify(mergePathObject(ret, ret), null, 4), 'UTF-8')

// const translateFileName = 'translate.json'
// const translateFileZHName = 'translate-zh.json'
// const absolutePath1 = resolve(__dirname, translateFileName)
// const absolutePath2 = resolve(__dirname, translateFileZHName)
// writeFileSync(absolutePath1, JSON.stringify(Object.assign(origin, translate), null, 4), 'UTF-8')
// writeFileSync(absolutePath2, JSON.stringify(Object.assign(originZh, translateZh), null, 4), 'UTF-8')
