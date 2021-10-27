<template>
    <div ref="result">
    </div>
</template>

<script>
    import diffview from './diffview'

    export default {
        name: 'js-diff',
        props: {
            srcContent: {
                type: String
            },
            targetContent: {
                type: String
            }
        },
        data () {
            return {
                source: '',
                target: '',
                sm: {},
                opcodes: []
            }
        },
        watch: {
            srcContent (val) {
                this.source = val
                this.render()
            },
            targetContent (val) {
                this.target = val
                this.render()
            }
        },
        mounted () {
            this.prepare()
        },
        methods: {
            prepare () {
                this.source = this.srcContent
                this.target = this.targetContent
                this.render()
            },

            render () {
                const options = {
                    api: 'dom',
                    content: false,
                    context: -1,
                    source: this.source,
                    diff: this.target,
                    diffcli: false,
                    diffcomments: false,
                    diffspaceignore: false,
                    diffview: 'sidebyside',
                    functions: {
                        binaryCheck: /\u0000|\u0001|\u0002|\u0003|\u0004|\u0005|\u0006|\u0007|\u000b|\u000e|\u000f|\u0010|\u0011|\u0012|\u0013|\u0014|\u0015|\u0016|\u0017|\u0018|\u0019|\u001a|\u001c|\u001d|\u001e|\u001f|\u007f|\u0080|\u0081|\u0082|\u0083|\u0084|\u0085|\u0086|\u0087|\u0088|\u0089|\u008a|\u008b|\u008c|\u008d|\u008e|\u008f|\u0090|\u0091|\u0092|\u0093|\u0094|\u0095|\u0096|\u0097|\u0098|\u0099|\u009a|\u009b|\u009c|\u009d|\u009e|\u009f/g
                    },
                    inchar: ' ',
                    insize: 4,
                    quote: false,
                    semicolon: false
                }
                const diffRet = diffview(options)
                this.$refs.result.innerHTML = diffRet[0]
                this.$emit('change-count', diffRet[2] + diffRet[1])
            }
        }
    }
</script>
<style lang="postcss">
    .diff {
        display: flex;
        max-height: 400px;
        overflow-y: scroll;
        overflow-x: hidden;
        border: 1px solid #e5e5e5;
        border-left: none;
        &::-webkit-scrollbar {
            width: 4px;
            background-color: lighten(transparent, 80%);
        }
        &::-webkit-scrollbar-thumb {
            height: 5px;
            border-radius: 2px;
            background-color: #e6e9ea;
        }

        .diff-left,
        .diff-right {
            flex: 1;
            width: 400px;
        }

        ol {
            margin: 0;
            padding: 0;
            overflow-x: scroll;
            &::-webkit-scrollbar {
                width: 4px;
                height: 4px;
                background-color: lighten(transparent, 80%);
            }
            &::-webkit-scrollbar-thumb {
                height: 5px;
                border-radius: 2px;
                background-color: #e6e9ea;
            }
        }

        li {
            height: 20px;
            line-height: 20px;
            list-style-type: none;
            letter-spacing: 1px;
            padding: 0 7px;
            margin: 1px 0;
        }

        em {
            font-style: normal;
            font-weight: bold;
            margin: 1px;
        }

        .count {
            text-align: right;
            float: left;
            border: 1px solid #e5e5e5;
            border-top: none;
            border-bottom: none;
            min-height: 398px;
            li {
                text-align: right;
                &.fold {
                    cursor: pointer;
                    font-weight: bold;
                    padding-left: 7px;
                    color: #900;
                }
            }
        }
        .data {
            text-align: left;
            white-space: pre;
            min-height: 398px;
            li {
                border-color: #ccc;
                min-width: 16.5em;
            }

            .replace {
                background: #ffdf8f;
                em {
                    background: #ffe;
                    border-color: #a86;
                    color: #852;
                }
            }
            .insert {
                background: #ecfdf0;
            }
            .delete {
                background: #fbe9eb;
            }
            .empty {
                background: #ddd;
            }
        }

        .replace {
            em {
                border-style: solid;
                border-width: 1px;
            }
        }
    }
</style>
