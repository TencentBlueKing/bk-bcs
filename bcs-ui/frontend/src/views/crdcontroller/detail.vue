<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-crd-title">
                <a class="bcs-icon bcs-icon-arrows-left back" href="javascript:void(0);" @click="goBack"></a>
                <span>{{curChartInfo.display_name || curCrdName}}</span>
            </div>
        </div>

        <div class="biz-content-wrapper">
            <div>
                <div class="biz-crd-header">
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
                            <svg class="logo" style="cursor: pointer;">
                                <use xlink:href="#biz-set-icon"></use>
                            </svg>
                            <div class="desc" :title="curApp.description">
                                <span>{{$t('简介')}}：</span>
                                {{curChartInfo.description || '--'}}
                            </div>
                        </div>
                    </div>

                    <div class="right">
                        <div class="bk-collapse-item bk-collapse-item-active">
                            <div class="biz-item-header" style="cursor: default;">
                                {{$t('配置选项')}}
                            </div>
                            <div class="bk-collapse-item-content f12" style="padding: 15px;">
                                <div class="config-box" style="min-width: 580px;">
                                    <div class="inner">
                                        <div class="inner-item">
                                            <label class="title">{{$t('名称')}}</label>
                                            <bkbcs-input :value="curApp.release_name" :disabled="true" style="width: 250px;" />
                                        </div>

                                        <div class="inner-item">
                                            <label class="title">{{$t('版本')}}</label>
                                            <div>
                                                <bcs-select v-model="curApp.chart_version" style="width: 250px;">
                                                    <bcs-option
                                                        v-for="(opt, index) in chartVersionsList"
                                                        :key="index"
                                                        :name="opt.version"
                                                        :id="opt.version"
                                                    ></bcs-option>
                                                </bcs-select>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="inner">
                                        <div class="inner-item">
                                            <label class="title">{{$t('所属集群')}}</label>
                                            <bkbcs-input :value="curApp.cluster_id" :disabled="true" style="width: 250px;" />
                                        </div>

                                        <div class="inner-item">
                                            <label class="title">{{$t('命名空间')}}</label>
                                            <div>
                                                <bkbcs-input v-model="curApp.namespace" style="width: 250px;" disabled></bkbcs-input>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="action-box">
                    <div class="title mb10">
                        Values {{$t('内容')}}
                    </div>
                </div>

                <div class="editor-box">
                    <monaco-editor
                        ref="yamlEditor"
                        class="editor"
                        theme="monokai"
                        language="yaml"
                        :style="{ height: `${editorHeight}px`, width: '100%' }"
                        v-model="editorOptions.content"
                        :diff-editor="editorOptions.isDiff"
                        :key="renderEditorKey"
                        :options="editorOptions"
                        :original="editorOptions.originContent">
                    </monaco-editor>
                </div>

                <div class="create-wrapper">
                    <bk-button type="primary" :title="$t('更新')" @click="handleUpdate">
                        {{$t('更新')}}
                    </bk-button>
                    <bk-button type="default" :title="$t('取消')" @click="goBack">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </div>

        <bk-dialog
            :width="1100"
            :title="updateConfirmDialog.title"
            :close-icon="!updateInstanceLoading"
            :is-show.sync="updateConfirmDialog.isShow"
            @cancel="hideConfirmDialog">
            <template slot="content">
                <p class="biz-tip mb5 tl" style="color: #666;" v-if="yamlDiffEditorOptions.isDiff">{{$t('Values 内容发生如下变化，请确认后再点击“确定”更新')}}</p>
                <div class="difference-code">
                    <div class="editor-header" v-if="yamlDiffEditorOptions.isDiff">
                        <div>当前内容</div>
                        <div>更新内容</div>
                    </div>

                    <div :class="['diff-editor-box', { 'editor-fullscreen': yamlDiffEditorOptions.fullScreen }]" style="position: relative;">
                        <monaco-editor
                            ref="yamlEditor"
                            class="editor"
                            theme="monokai"
                            language="yaml"
                            :style="{ height: `${diffEditorHeight}px`, width: '100%' }"
                            v-model="curAppDifference.content"
                            :diff-editor="yamlDiffEditorOptions.isDiff"
                            :key="differenceKey"
                            :options="yamlDiffEditorOptions"
                            :original="curAppDifference.originContent">
                        </monaco-editor>
                    </div>
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer mt10">
                    <template>
                        <bk-button
                            :class="['bk-button bk-dialog-btn-confirm bk-primary', { 'is-disabled': updateInstanceLoading }]"
                            @click="updateCrdController">
                            {{updateInstanceLoading ? $t('更新中...') : $t('确定')}}
                        </bk-button>
                        <bk-button
                            :class="['bk-button bk-dialog-btn-cancel bk-default']"
                            @click="hideConfirmDialog">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>
    </div>
</template>

<script>
    import { catchErrorHandler } from '@/common/util'
    import MonacoEditor from '@/components/monaco-editor/editor.vue'

    export default {
        components: {
            MonacoEditor
        },
        data () {
            return {
                editorOptions: {
                    readOnly: false,
                    fontSize: 14,
                    fullScreen: false,
                    content: '',
                    originContent: '',
                    isDiff: false
                },
                curAppDifference: {
                    content: '',
                    originContent: ''
                },
                updateInstanceLoading: false,
                differenceKey: 0,
                yamlDiffEditorOptions: {
                    readOnly: true,
                    fontSize: 14,
                    fullScreen: false,
                    isDiff: false
                },
                updateConfirmDialog: {
                    title: this.$t('确认更新'),
                    isShow: false,
                    width: 1060,
                    height: 350,
                    lang: 'yaml',
                    closeIcon: true,
                    readOnly: true,
                    fullScreen: false,
                    values: [],
                    editors: []
                },
                diffEditorHeight: 350,
                editorHeight: 500,
                curChartInfo: {},
                curApp: {
                    namespace: ''
                },
                namespaceList: [],
                chartVersionsList: []
            }
        },
        computed: {
            curProject () {
                return this.$store.state.curProject
            },
            projectId () {
                return this.$route.params.projectId
            }
        },

        created () {
            this.curClusterId = this.$route.params.clusterId
            this.curCrdId = this.$route.params.id
            this.curCrdName = this.$route.params.name
            this.chartName = this.$route.params.chartName

            if (window.sessionStorage['bcs-crdcontroller']) {
                const obj = JSON.parse(window.sessionStorage['bcs-crdcontroller'])
                if (String(obj.crd_ctr_id) === String(this.curCrdId)) {
                    this.curChartInfo = obj
                }
            }
            this.getCommonCrdInstanceDetail()
            this.fetchChartVersionsList()
        },
        methods: {
            async getCommonCrdInstanceDetail () {
                try {
                    const projectId = this.projectId
                    const clusterId = this.curClusterId
                    const crdId = this.curCrdId
                    const res = await this.$store.dispatch('crdcontroller/getCommonCrdInstanceDetail', { projectId, clusterId, crdId })
                    this.curApp = res.data
                    this.editorOptions.content = res.data.values_content
                    this.editorOptions.originContent = res.data.values_content
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },
            async fetchChartVersionsList () {
                const projectId = this.projectId
                const chartName = this.chartName
                const res = await this.$store.dispatch('crdcontroller/getChartVersionsList', { projectId, chartName })
                this.chartVersionsList = res.data
            },

            goBack () {
                this.$router.push({
                    name: 'dbCrdcontroller',
                    params: {
                        projectId: this.projectId
                    }
                })
            },

            handleUpdate () {
                this.curAppDifference.content = this.editorOptions.content
                this.curAppDifference.originContent = this.editorOptions.originContent
                if (this.curAppDifference.content === this.curAppDifference.originContent) {
                    this.curAppDifference.content = this.$t('本次更新没有内容变化')
                    this.yamlDiffEditorOptions.isDiff = false
                } else {
                    this.yamlDiffEditorOptions.isDiff = true
                }
                this.updateConfirmDialog.isShow = true
                setTimeout(() => {
                    this.differenceKey++
                }, 0)
            },

            checkData () {
                if (!this.editorOptions.content) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请填写Values 内容')
                    })
                    return false
                }
                return true
            },

            hideConfirmDialog () {
                if (this.updateInstanceLoading) {
                    return false
                }
                this.updateConfirmDialog.isShow = false
            },

            async updateCrdController () {
                if (this.updateInstanceLoading) {
                    return false
                }

                this.updateInstanceLoading = true
                try {
                    const curVersion = this.chartVersionsList.find(i => i.version === this.curApp.chart_version)
                    const chartUrl = curVersion && curVersion.urls[0]
                    const projectId = this.projectId
                    const clusterId = this.curClusterId
                    const crdId = this.curApp.crd_ctr_id
                    const data = {
                        values_content: this.editorOptions.content,
                        chart_url: chartUrl
                    }
                    await this.$store.dispatch('crdcontroller/updateCommonCrdInstance', { projectId, clusterId, crdId, data })
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('配置已下发成功')
                    })
                    this.goBack()
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.updateInstanceLoading = false
                }
            }
        }
    }
</script>

<style scoped>
    @import './detail.css';
</style>
