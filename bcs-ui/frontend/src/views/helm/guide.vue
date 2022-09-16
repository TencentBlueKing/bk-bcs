<template>
    <bk-sideslider
        :is-show.sync="visibility"
        :title="$t('如何推送Helm Chart到项目仓库？')"
        :width="900"
        :quick-close="true">
        <div slot="content">
            <div v-html="markdown" class="biz-markdown-content" id="markdown"></div>
        </div>
    </bk-sideslider>
</template>

<script>
    import MarkdownIt from 'markdown-it'

    export default {
        props: {
            isShow: {
                type: Boolean,
                default: false
            }
        },
        data () {
            return {
                visibility: this.isShow,
                markdown: ''
            }
        },
        mounted () {
            this.init()
        },
        methods: {
            /**
             * 显示
             */
            show () {
                this.visibility = true
                this.$emit('status-change', this.visibility)
            },

            /**
             * 隐藏
             */
            hide () {
                this.visibility = false
                this.$emit('status-change', this.visibility)
            },

            /**
             * 初始化
             */
            async init () {
                const projectId = this.$route.params.projectId
                const res = await this.$store.dispatch('helm/getQuestionsMD', projectId).catch(() => ({ data: { content: '' } }))
                const markdown = res.data.content
                const md = new MarkdownIt({
                    linkify: false
                })
                // create render rules
                const defaultRender = md.renderer.rules.link_open || function (tokens, idx, options, env, self) {
                    return self.renderToken(tokens, idx, options)
                }
                md.renderer.rules.link_open = function (tokens, idx, options, env, self) {
                    const aIndex = tokens[idx].attrIndex('target')

                    if (aIndex < 0) {
                        tokens[idx].attrPush(['target', '_blank'])
                    } else {
                        tokens[idx].attrs[aIndex][1] = '_blank'
                    }
                    return defaultRender(tokens, idx, options, env, self)
                }

                this.markdown = md.render(markdown)
            }
        }
    }
</script>

<style>
    .biz-markdown-content {
        pre {
            padding: 0;
            position: relative;

            code {
                word-break: break-all;
            }

            .code-box {
                padding: 10px;
                width: 100%;
                min-height: 30px;
                overflow: auto;
            }

            .copy-btn {
                display: none;
                position: absolute;
                right: 8px;
                top: 8px;
            }

            &:hover {
                .copy-btn {
                    display: inline-block;
                }
            }
        }
    }
</style>
