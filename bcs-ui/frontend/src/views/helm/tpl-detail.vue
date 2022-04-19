<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                <a class="bcs-icon bcs-icon-arrows-left back" @click="goTplList"></a>
                <span>{{$t('Chart详情')}}</span>
            </div>
            <div class="biz-actions" style="margin-top: 11px; ">
                <router-link :to="{ name: 'helmTplInstance', params: { tplId: curTpl.id } }" :class="['bk-button bk-primary']">
                    {{$t('部署')}}
                </router-link>
            </div>
        </div>

        <div class="biz-content-wrapper" v-bkloading="{ isLoading: createInstanceLoading }">
            <div>
                <div class="biz-helm-header">
                    <div class="left">
                        <svg style="display: none;">
                            <title>{{$t('模板集默认图标')}}</title>
                            <symbol id="biz-set-icon" viewBox="0 0 32 32">
                                <path d="M6 3v3h-3v23h23v-3h3v-23h-23zM24 24v3h-19v-19h19v16zM27 24h-1v-18h-18v-1h19v19z"></path>
                                <path d="M13.688 18.313h-6v6h6v-6z"></path>
                                <path d="M21.313 10.688h-6v13.625h6v-13.625z"></path>
                                <path d="M13.688 10.688h-6v6h6v-6z"></path>
                            </symbol>
                        </svg>
                        <div class="info">
                            <div class="logo-wrapper" v-if="curTpl.icon && isImage(curTpl.icon)">
                                <img :src="curTpl.icon" style="width: 100px;">
                            </div>
                            <svg class="logo" v-else>
                                <use xlink:href="#biz-set-icon"></use>
                            </svg>
                            <div class="title">{{curTpl.name}}</div>
                            <div class="desc" :title="curTpl.description">
                                <span>{{$t('简介')}}：</span>
                                {{curTpl.description || '--'}}
                            </div>
                        </div>
                    </div>

                    <div class="right">
                        <div class="bk-collapse biz-collapse">
                            <div class="bk-collapse-item bk-collapse-item-active">
                                <div class="biz-item-header">
                                    {{$t('版本')}}
                                </div>
                                <!-- <div class="bk-collapse-item-header" style="cursor: default; border-color: #dfe0e5;">
                                    {{$t('版本')}}
                                </div> -->
                                <div class="bk-collapse-item-content f13" style="padding: 15px;">
                                    <div class="config-box">
                                        <div class="inner">
                                            <label class="title">{{$t('Chart版本')}}：</label>
                                            <bk-selector
                                                :placeholder="$t('请选择')"
                                                style="max-width: 800px;"
                                                :selected.sync="tplsetVerIndex"
                                                :list="curTplVersions"
                                                :setting-key="'version'"
                                                :display-key="'version'"
                                                @item-selected="getTplDetail">
                                            </bk-selector>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div v-if="tplsetVerIndex" v-bkloading="{ isLoading: isVersionLoading }">
                    <bk-tab :active-name="'files'" class="mt20 biz-tab-container">
                        <bk-tab-panel name="files" :title="$t('资源文件')" class="biz-tree-tab">
                            <template v-if="previewList.length">
                                <div class="biz-resource-wrapper" style="display: flex;">
                                    <resizer :class="['resize-layout fl']"
                                        direction="right"
                                        :handler-offset="3"
                                        :min="250"
                                        :max="500">
                                        <div class="tree-box" style="max-height: 500px;">
                                            <bcs-tree
                                                ref="tree1"
                                                :class="'biz-helm-tree'"
                                                :data="treeData"
                                                :node-key="'id'"
                                                :has-border="true"
                                                @on-click="getFileDetail">
                                            </bcs-tree>
                                        </div>
                                    </resizer>

                                    <div class="resource-box">
                                        <div class="biz-code-wrapper">
                                            <ace
                                                ref="codeViewer"
                                                :value="curReourceFile.value"
                                                :width="editorConfig.width"
                                                :height="editorConfig.height"
                                                :lang="editorConfig.lang"
                                                :read-only="editorConfig.readOnly"
                                                :full-screen="editorConfig.fullScreen">
                                            </ace>
                                        </div>
                                    </div>
                                </div>
                            </template>
                            <template v-else>
                                <div class="bk-message-box">
                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                </div>
                            </template>
                        </bk-tab-panel>
                        <bk-tab-panel name="readme" :title="$t('详细说明')">
                            <template v-if="curTplReadme">
                                <div class="p20">
                                    <div class="biz-scroller-container">
                                        <div v-html="markdown" class="biz-markdown-content" id="markdown"></div>
                                        <!-- <pre style="white-space: pre-line;">{{curTplReadme}}</pre> -->
                                    </div>
                                </div>
                            </template>
                            <template v-else>
                                <div class="bk-message-box">
                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                </div>
                            </template>
                        </bk-tab-panel>
                    </bk-tab>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import MarkdownIt from 'markdown-it'
    import path2tree from '@/common/path2tree'
    import baseMixin from '@/mixins/helm/mixin-base'
    import { catchErrorHandler } from '@/common/util'
    import resizer from '@/components/resize'

    export default {
        components: {
            resizer
        },
        mixins: [baseMixin],
        data () {
            return {
                markdown: '',
                curTplReadme: '',
                yamlEditor: null,
                yamlFile: '',
                curTplYaml: '',
                activeName: ['config'],
                collapseName: ['var'],
                tplsetVerList: [],
                formData: {},
                createInstanceLoading: false,
                previewList: [],
                curReourceFile: {
                    name: '',
                    value: ''
                },
                isVersionLoading: true,
                tplPreviewList: [],
                difference: '',
                previewInstanceLoading: true,
                editor: null,
                curTpl: {
                    data: {
                        name: ''
                    }
                },
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: true,
                    fullScreen: false,
                    values: [],
                    editors: []
                },
                curTplVersions: [],
                tplsetVerIndex: '',
                answers: {},
                curLabelList: [
                    {
                        key: '',
                        value: ''
                    }
                ],
                treeData: []
            }
        },
        computed: {
            curProject () {
                return this.$store.state.curProject
            },
            projectId () {
                return this.$route.params.projectId
            },
            tplList () {
                return this.$store.state.helm.tplList
            }
        },
        async mounted () {
            const tplId = this.$route.params.tplId
            this.curTpl = await this.getTplById(tplId)
            this.getTplVersions()
        },
        methods: {
            /**
             * 返回chart 模版列表
             */
            goTplList () {
                const projectCode = this.$route.params.projectCode
                this.$router.push({
                    name: 'helmTplList',
                    params: {
                        projectCode: projectCode
                    }
                })
            },

            /**
             * 获取文件详情
             * @param  {object} file 文件
             */
            getFileDetail (file) {
                if (file.hasOwnProperty('value')) {
                    this.curReourceFile = file
                    this.$nextTick(() => {
                        this.$refs.codeViewer && this.$refs.codeViewer.$ace && this.$refs.codeViewer.$ace.scrollToLine(1, true, true)
                    })
                }
            },

            /**
             * 获取模板
             * @param  {number} id 模板ID
             * @return {object} result 模板
             */
            async getTplById (id) {
                let result = {}
                let list = this.tplList
                // 如果没有缓存，获取远程数据
                if (!list.length) {
                    try {
                        const projectId = this.projectId
                        const res = await this.$store.dispatch('helm/asyncGetTplList', projectId)
                        list = res.data
                    } catch (e) {
                        catchErrorHandler(e, this)
                    }
                }

                list.forEach(item => {
                    // 跟由获取的id为string，转number
                    if (item.id === Number(id)) {
                        result = item
                    }
                })
                return result
            },

            /**
             * 根据版本号获取模板详情
             * @param  {number} index 索引
             * @param  {object} data 数据
             */
            async getTplDetail (index, data) {
                const list = []
                const projectId = this.projectId
                const version = index
                const chartId = this.curTpl.name

                this.isVersionLoading = true
                this.treeData = []

                try {
                    const isPublic = this.curTpl.repository.name === 'public-repo'
                    const res = await this.$store.dispatch('helm/getChartVersionDetail', {
                        projectId,
                        chartId,
                        version,
                        isPublic
                    })

                    const tplData = res.data
                    const files = tplData.data.files
                    const tplName = tplData.name
                    this.formData = tplData.data.questions

                    for (const key in files) {
                        list.push({
                            name: key,
                            value: files[key]
                        })
                    }

                    this.previewList.splice(0, this.previewList.length, ...list)
                    const tree = path2tree(this.previewList, { expandIndex: 0 })
                    this.treeData.push(tree)
                    this.curTplReadme = files[`${tplName}/README.md`]
                    this.curTplYaml = files[`${tplName}/values.yaml`]
                    // default: 显示第一个
                    if (this.previewList.length) {
                        this.curReourceFile = this.previewList[0]

                        this.$nextTick(() => {
                            this.$refs.codeViewer && this.$refs.codeViewer.$ace && this.$refs.codeViewer.$ace.scrollToLine(1, true, true)
                        })
                    }

                    const md = new MarkdownIt({
                        linkify: false
                    })
                    this.markdown = md.render(this.curTplReadme)
                    this.$nextTick(() => {
                        // 点击链接新开窗口打开
                        const markdownDom = document.getElementById('markdown')
                        if (!markdownDom) return

                        markdownDom.querySelectorAll('a').forEach(item => {
                            item.target = '_blank'
                            item.className = 'bk-text-button'
                        })
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isVersionLoading = false
                }
            },

            /**
             * 获取模板版本列表
             * @param  {number} tplId 模板ID
             */
            async getTplVersions () {
                const projectId = this.projectId
                const isPublic = this.curTpl.repository.name === 'public-repo'
                try {
                    if (this.$INTERNAL) {
                        // 内部版本
                        const tplId = this.curTpl.name
                        const res = await this.$store.dispatch('helm/getTplVersionList', { projectId, tplId, isPublic })
                        this.curTplVersions = res.data || []
                    } else {
                        // 外部版本
                        const tplId = this.curTpl.id
                        const res = await this.$store.dispatch('helm/getTplVersions', { projectId, tplId })
                        this.curTplVersions = res.data.results || []
                    }
                    if (this.curTplVersions.length) {
                        const firstVersion = this.curTplVersions[0]
                        const versionId = firstVersion.version

                        this.tplsetVerIndex = versionId
                        this.getTplDetail(versionId, firstVersion)
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            }
        }
    }
</script>

<style scoped>
    @import './common.css';
    @import './tpl-detail.css';
</style>
