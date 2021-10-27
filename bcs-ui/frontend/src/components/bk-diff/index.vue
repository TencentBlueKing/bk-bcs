<template>
    <div class="bk-diff">
        <div v-html="html" v-highlight></div>
    </div>
</template>

<script>
    import { createPatch } from 'diff'
    import { Diff2Html } from 'diff2html'
    import hljs from 'highlight.js/lib/highlight'

    export default {
        name: 'bk-diff',
        directives: {
            highlight: el => {
                const hljsLanguageConfig = [
                    'javascript',
                    'json',
                    'shell',
                    'bash',
                    'xml',
                    'vim'
                ]

                hljsLanguageConfig.forEach(lang => {
                    require([`highlight.js/lib/languages/${lang}`], langModule => {
                        hljs.registerLanguage(
                            lang,
                            (langModule.default && typeof langModule.default === 'function')
                                ? langModule.default
                                : langModule
                        )
                    })

                    // import(
                    //     /* webpackChunkName: 'hljs' */
                    //     `highlight.js/lib/languages/${lang}`
                    // ).then(langModule => {
                    //     hljs.registerLanguage(
                    //         lang,
                    //         (langModule.default && typeof langModule.default === 'function')
                    //             ? langModule.default
                    //             : langModule
                    //     )
                    // })
                })

                const blocks = el.querySelectorAll('code')

                blocks.forEach(block => {
                    hljs.highlightBlock(block)
                })
            }
        },
        props: {
            oldContent: {
                type: String,
                default: ''
            },
            newContent: {
                type: String,
                default: ''
            },
            context: {
                type: Number,
                default: 5
            },
            format: {
                type: String,
                default: 'line-by-line'
            },
            minHeight: {
                type: Number,
                default: 100
            }
        },
        computed: {
            html () {
                return this.createdHtml(this.oldContent, this.newContent, this.context, this.format)
            }
        },

        methods: {
            getDiffJson (oldContent, newContent, context, outputFormat) {
                const args = ['', oldContent, newContent, '', '', { context: context }]
                const patch = createPatch(...args)

                const outStr = Diff2Html.getJsonFromDiff(patch, {
                    inputFormat: 'diff',
                    outputFormat: outputFormat,
                    showFiles: true,
                    matching: 'lines'
                })

                const addedLines = outStr[0].addedLines
                const deleteLines = outStr[0].deletedLines
                const changeLines = Math.max(addedLines, deleteLines)
                outStr.changeLines = changeLines

                return outStr
            },
            createdHtml (oldContent, newContent, context, outputFormat) {
                function htmlReplace (html) {
                    return html.replace(
                        /<span class="d2h-code-line-ctn">(.+?)<\/span>/g,
                        '<span class="d2h-code-line-ctn"><code>$1</code></span>'
                    )
                }

                let diffJsonConf = this.getDiffJson(oldContent, newContent, context, outputFormat)

                this.$emit('change-count', diffJsonConf.changeLines)
                if (diffJsonConf.changeLines) {
                    const html = Diff2Html.getPrettyHtml(diffJsonConf, {
                        inputFormat: 'json',
                        outputFormat: outputFormat,
                        showFiles: false,
                        matching: 'lines'
                    })
                    return htmlReplace(html)
                } else {
                    diffJsonConf = this.getDiffJson(oldContent, newContent + '\n', context, outputFormat)

                    diffJsonConf.changeLines = 0
                    diffJsonConf[0].addedLines = 0
                    diffJsonConf[0].deletedLines = 0
                    const firstBlock = diffJsonConf[0].blocks[0]
                    firstBlock.header = ''
                    firstBlock.newStartLine = '0'
                    firstBlock.oldStartLine = '0'
                    firstBlock.lines.splice(firstBlock.lines.length - 1, 1)

                    const html = Diff2Html.getPrettyHtml(diffJsonConf, {
                        inputFormat: 'json',
                        outputFormat: outputFormat,
                        showFiles: false,
                        matching: 'lines'
                    })
                    return htmlReplace(html)
                    // return `<div class="diff-tip-box" style="line-height: ${this.minHeight}px;">数据没有差异</div>`
                }
            }
        }
    }
</script>

<style>
    @import './diff.css';
</style>
