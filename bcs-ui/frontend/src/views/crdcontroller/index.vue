<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-app-title">
                {{crdKind === 'BcsLog' ? $t('日志采集') : $t('组件库')}}
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper" style="padding: 0;">
            <template>
                <div class="biz-panel-header">
                    <div class="right">
                        <bk-data-searcher
                            :scope-list="searchScopeList"
                            :search-key.sync="searchKeyword"
                            :search-scope.sync="searchScope"
                            :cluster-fixed="!!curClusterId"
                            @search="search"
                            @refresh="refresh">
                        </bk-data-searcher>
                    </div>
                </div>
                <div class="biz-crdcontroller" v-bkloading="{ isLoading: isPageLoading }" style="min-height: 180px;">
                    <svg style="display: none;">
                        <title>{{$t('模板集默认图标')}}</title>
                        <symbol id="biz-set-icon" viewBox="0 0 60 60">
                            <g id="图层_6">
                                <g id="图层_32_1_">
                                    <path class="st0" d="M12,8v4H8c-1.1,0-2,0.9-2,2v42c0,1.1,0.9,2,2,2h42c1.1,0,2-0.9,2-2v-4h4c1.1,0,2-0.9,2-2V8c0-1.1-0.9-2-2-2
                                        H14C12.9,6,12,6.9,12,8z M48,48v4v2H10V16h2h4h32V48z M54,48h-2V14c0-1.1-0.9-2-2-2H16v-2h38V48z" />
                                </g>
                                <path class="st1" d="M45.7,33.7h-1.8l-3.4-8.3l1.3-1.3c0.5-0.5,0.5-1.3,0-1.8l0,0c-0.5-0.5-1.3-0.5-1.8,0l-1.3,1.3l-8.4-3.5v-1.8
                                    c0-0.7-0.6-1.3-1.3-1.3l0,0c-0.7,0-1.3,0.6-1.3,1.3V20l-8.4,3.5l-1.2-1.2c-0.5-0.5-1.3-0.5-1.8,0l0,0c-0.5,0.5-0.5,1.3,0,1.8
                                    l1.2,1.2L14,33.7h-1.8c-0.7,0-1.3,0.6-1.3,1.3l0,0c0,0.7,0.6,1.3,1.3,1.3H14l3.5,8.4L16.2,46c-0.5,0.5-0.5,1.3,0,1.8l0,0
                                    c0.5,0.5,1.3,0.5,1.8,0l1.3-1.3l8.3,3.4v1.8c0,0.7,0.6,1.3,1.3,1.3l0,0c0.7,0,1.3-0.6,1.3-1.3v-1.9l8.3-3.4l1.3,1.3
                                    c0.5,0.5,1.3,0.5,1.8,0l0,0c0.5-0.5,0.5-1.3,0-1.8l-1.3-1.3l3.4-8.3h1.9c0.7,0,1.3-0.6,1.3-1.3l0,0C47,34.3,46.4,33.7,45.7,33.7z
                                     M30.3,23.4l6,2.5l-4.6,4.6c-0.4-0.2-0.9-0.4-1.3-0.6v-6.5H30.3z M27.7,23.4V30c-0.5,0.1-0.9,0.3-1.4,0.6l-4.7-4.7L27.7,23.4z
                                     M19.9,27.7l4.7,4.7c-0.2,0.4-0.4,0.9-0.5,1.3h-6.6L19.9,27.7z M17.4,36.3H24c0.1,0.5,0.3,0.9,0.6,1.3l-4.7,4.7L17.4,36.3z
                                     M27.7,46.5l-6-2.5l4.7-4.7c0.4,0.2,0.8,0.4,1.3,0.5V46.5z M29,37.5c-1.4,0-2.6-1.2-2.6-2.6c0-1.4,1.2-2.6,2.6-2.6s2.6,1.2,2.6,2.6
                                    C31.6,36.4,30.4,37.5,29,37.5z M30.3,46.5v-6.6c0.5-0.1,0.9-0.3,1.3-0.5l4.6,4.6L30.3,46.5z M38,42.2l-4.6-4.6
                                    c0.2-0.4,0.4-0.8,0.6-1.3h6.5L38,42.2z M34,33.7c-0.1-0.5-0.3-0.9-0.5-1.3l4.6-4.6l2.5,6H34V33.7z" />
                                <g class="st2">
                                    <path class="st3" d="M41,49H17c-1.1,0-2-0.9-2-2V23c0-1.1,0.9-2,2-2h24c1.1,0,2,0.9,2,2v24C43,48.1,42.1,49,41,49z" />
                                </g>
                                <g>
                                    <path class="st0" d="M42.2,25c-1.9,0-2.9,0.5-2.9,1.5v17.1c0,1,1,1.5,2.9,1.5v1.8H31.4V45c2,0,3-0.5,3-1.5v-8H23.6v8
                                        c0,1,1,1.5,3,1.5v1.8H15.8V45c1.9,0,2.8-0.5,2.8-1.5V26.4c0-1-0.9-1.5-2.8-1.5V23h10.8v2c-2,0-3,0.5-3,1.5v6.8h10.8v-6.8
                                        c0-1-1-1.5-3-1.5v-1.9h10.8V25z" />
                                </g>
                            </g>
                        </symbol>
                    </svg>
                    <table class="bk-table biz-templateset-table mb20" v-if="crdKind !== 'BcsLog'">
                        <thead>
                            <tr>
                                <th style="width: 120px; padding-left: 0;" class="center">{{$t('图标')}}</th>
                                <th style="width: 250px; padding-left: 20px;">{{$t('组件名称')}}</th>
                                <th style="width: 150px; padding-left: 20px;">{{$t('状态')}}</th>
                                <th style="padding-left: 0;">{{$t('描述')}}</th>
                                <th style="width: 170px; padding-left: 0;">{{$t('操作')}}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="crdControllerList.length">
                                <tr
                                    v-for="crdcontroller of crdControllerList"
                                    :key="crdcontroller.id">
                                    <td colspan="5">
                                        <table class="biz-inner-table">
                                            <tr>
                                                <td class="logo">
                                                    <div class="logo-wrapper" v-if="logMap[crdcontroller.name]">
                                                        <i :class="logMap[crdcontroller.name]"></i>
                                                    </div>
                                                    <svg class="biz-set-icon" v-else>
                                                        <use xlink:href="#biz-set-icon"></use>
                                                    </svg>
                                                </td>
                                                <td class="name" style="width: 250px;">
                                                    <p class="text">{{crdcontroller.display_name || '--'}}</p>
                                                </td>
                                                <td class="status">
                                                    <span class="biz-mark" v-if="crdcontroller.status === 'deployed'">
                                                        <bk-tag type="filled" theme="success">{{$t('已部署')}}</bk-tag>
                                                    </span>
                                                    <span class="biz-mark" v-else-if="!crdcontroller.status || crdcontroller.status === 'not_deployed'">
                                                        <bk-tag type="filled">{{$t('未启用')}}</bk-tag>
                                                    </span>
                                                    <span class="biz-mark" v-else-if="crdcontroller.status === 'failed'">
                                                        <bcs-popover :width="500" :content="crdcontroller.message" placement="top">
                                                            <bk-tag type="filled" theme="danger">{{$t('异常')}}</bk-tag>
                                                        </bcs-popover>
                                                    </span>
                                                    <span class="biz-mark" v-else-if="crdcontroller.status === 'unknown'">
                                                        <bcs-popover :content="$t('请联系蓝鲸容器助手')" placement="top">
                                                            <bk-tag type="filled" theme="warning">{{$t('未知')}}</bk-tag>
                                                        </bcs-popover>
                                                    </span>
                                                    <template v-else-if="crdcontroller.status === 'pending'">
                                                        <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-primary vm" style="margin-right: 3px;">
                                                            <div class="rotate rotate1"></div>
                                                            <div class="rotate rotate2"></div>
                                                            <div class="rotate rotate3"></div>
                                                            <div class="rotate rotate4"></div>
                                                            <div class="rotate rotate5"></div>
                                                            <div class="rotate rotate6"></div>
                                                            <div class="rotate rotate7"></div>
                                                            <div class="rotate rotate8"></div>
                                                        </div>
                                                        <span class="vm">{{$t('启用中...')}}</span>
                                                    </template>
                                                </td>
                                                <td class="description">
                                                    <p class="text">
                                                        {{crdcontroller.description}}
                                                        <a :href="crdcontroller.help_link" class="bk-text-button f12" target="_blank" v-if="crdcontroller.help_link">{{$t('详情查看文档')}}</a>
                                                    </p>
                                                </td>
                                                <td class="action">
                                                    <template v-if="crdcontroller.status === 'deployed'">
                                                        <template v-if="crdcontroller.name === 'DbPrivilege'">
                                                            <bk-button type="primary" @click="goControllerInstances(crdcontroller)">{{$t('前往配置')}}</bk-button>
                                                        </template>
                                                        <template v-else>
                                                            <bk-dropdown-menu
                                                                class="dropdown-menu"
                                                                :align="'left'"
                                                                ref="dropdown">
                                                                <bk-button :class="['bk-button bk-default btn']" slot="dropdown-trigger" style="position: relative; width: 88px;">
                                                                    <span>{{$t('操作')}}</span>
                                                                    <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down ml5" style="font-size: 10px;"></i>
                                                                </bk-button>

                                                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                                                    <li v-if="crdcontroller.name === 'BcsPolaris'">
                                                                        <a href="javascript:void(0)" @click="goControllerInstances(crdcontroller)">{{$t('前往配置')}}</a>
                                                                    </li>
                                                                    <li>
                                                                        <a href="javascript:void(0)" @click="showInstanceDetail(crdcontroller)">{{$t('更新组件')}}</a>
                                                                    </li>
                                                                </ul>
                                                            </bk-dropdown-menu>
                                                        </template>
                                                    </template>
                                                    <template v-else-if="!crdcontroller.status || crdcontroller.status === 'not_deployed'">
                                                        <bk-button :loading="crdcontroller.status === 'pending'" type="primary" @click="haneldEnableCrdController(crdcontroller)">{{$t('启用')}}</bk-button>
                                                    </template>
                                                    <template v-else-if="crdcontroller.status === 'failed'">
                                                        <template v-if="crdcontroller.name === 'DbPrivilege'">
                                                            <bk-button :loading="crdcontroller.status === 'pending'" type="primary" @click="haneldEnableCrdController(crdcontroller)">{{$t('重新启用')}}</bk-button>
                                                        </template>
                                                        <template v-else>
                                                            <bk-dropdown-menu
                                                                class="dropdown-menu"
                                                                :align="'left'"
                                                                ref="dropdown">
                                                                <bk-button :class="['bk-button bk-default btn']" slot="dropdown-trigger" style="position: relative; width: 88px;">
                                                                    <span>{{$t('操作')}}</span>
                                                                    <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down ml5" style="font-size: 10px;"></i>
                                                                </bk-button>
                                                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                                                    <li>
                                                                        <a href="javascript:void(0)" @click="showInstanceDetail(crdcontroller)">{{$t('更新组件')}}</a>
                                                                    </li>
                                                                </ul>
                                                            </bk-dropdown-menu>
                                                        </template>
                                                    </template>
                                                    <template v-else-if="crdcontroller.status === 'unknown'">
                                                        <span v-bk-tooltips="$t('请联系蓝鲸容器助手')">
                                                            <bk-button :disabled="true">{{$t('启用')}}</bk-button>
                                                        </span>
                                                    </template>
                                                    <template v-else-if="crdcontroller.status === 'pending'">
                                                        <bk-button :disabled="true">{{$t('启用中...')}}</bk-button>
                                                    </template>
                                                </td>
                                            </tr>
                                        </table>
                                    </td>
                                </tr>
                            </template>
                            <template v-if="!crdControllerList.length && !showLoading">
                                <tr>
                                    <td colspan="5">
                                        <bcs-exception type="empty" scene="part"></bcs-exception>
                                    </td>
                                </tr>
                            </template>
                        </tbody>
                    </table>

                    <table class="bk-table biz-templateset-table mb20" v-else>
                        <thead>
                            <tr>
                                <th style="width: 120px; padding-left: 0;" class="center">{{$t('图标')}}</th>
                                <th style="width: 110px; padding-left: 20px;">{{$t('组件名称')}}</th>
                                <th style="width: 100px; padding-left: 20px;">{{$t('状态')}}</th>
                                <th style="width: 390px; padding-left: 20px;">{{$t('数据源信息')}}</th>
                                <th style="padding-left: 0;">{{$t('描述')}}</th>
                                <th style="width: 170px; padding-left: 0;">{{$t('操作')}}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="crdControllerList.length">
                                <tr
                                    v-for="crdcontroller of crdControllerList"
                                    :key="crdcontroller.id">
                                    <td colspan="6">
                                        <table class="biz-inner-table">
                                            <tr>
                                                <td class="logo">
                                                    <div class="logo-wrapper" v-if="crdcontroller.logo && isImage(crdcontroller.logo)">
                                                        <img :src="crdcontroller.logo">
                                                    </div>
                                                    <svg class="biz-set-icon" v-else>
                                                        <use xlink:href="#biz-set-icon"></use>
                                                    </svg>
                                                </td>
                                                <td class="log-name">
                                                    <p class="text">{{crdcontroller.display_name || '--'}}</p>
                                                </td>
                                                <td class="log-status">
                                                    <span class="biz-mark" v-if="crdcontroller.status === 'deployed'">
                                                        <bk-tag type="filled" theme="success">{{$t('已启用')}}</bk-tag>
                                                    </span>
                                                    <span class="biz-mark" v-else-if="!crdcontroller.status || crdcontroller.status === 'not_deployed'">
                                                        <bk-tag type="filled">{{$t('未启用')}}</bk-tag>
                                                    </span>
                                                    <span class="biz-mark" v-else-if="crdcontroller.status === 'failed'">
                                                        <bcs-popover :content="crdcontroller.message || '--'" placement="top" width="500">
                                                            <bk-tag type="filled" theme="danger">{{$t('启用失败')}}</bk-tag>
                                                        </bcs-popover>
                                                    </span>
                                                </td>
                                                <td class="log-source" v-if="crdKind === 'BcsLog'">
                                                    <p>{{$t('标准日志')}}：{{dataSource.std_data_name || '--'}}</p>
                                                    <p>{{$t('文件路径日志')}}：{{dataSource.file_data_name || '--'}}</p>
                                                    <p>{{$t('系统日志')}}：{{dataSource.sys_data_name || '--'}}</p>
                                                </td>
                                                <td class="description">
                                                    <p class="text">
                                                        {{crdcontroller.description}}
                                                        <a :href="crdcontroller.help_link" class="bk-text-button f12" target="_blank" v-if="crdcontroller.help_link">{{$t('详情查看文档')}}</a>
                                                    </p>
                                                </td>
                                                <td class="action">
                                                    <template v-if="crdcontroller.status === 'deployed'">
                                                        <bk-button type="primary" @click="goControllerInstances(crdcontroller)">{{$t('前往配置')}}</bk-button>
                                                    </template>
                                                    <template v-else>
                                                        <bk-button type="primary" :loading="crdcontroller.status === 'pending'" @click="haneldEnableCrdController(crdcontroller)">{{$t('启用')}}</bk-button>
                                                    </template>
                                                </td>
                                            </tr>
                                        </table>
                                    </td>
                                </tr>
                            </template>
                            <template v-if="!crdControllerList.length && !showLoading">
                                <tr>
                                    <td colspan="6">
                                        <bcs-exception type="empty" scene="part"></bcs-exception>
                                    </td>
                                </tr>
                            </template>
                        </tbody>
                    </table>
                </div>
            </template>
        </div>

        <bk-sideslider
            class="editor-slider"
            :quick-close="false"
            :is-show.sync="valueSlider.isShow"
            :title="valueSlider.title"
            :width="'900'">
            <div class="p0" slot="content">
                <bk-button class="bk-button bk-primary save-crd-btn" @click.stop.prevent="enableCrdController">{{$t('启用')}}</bk-button>
                <bk-button class="bk-button bk-default hide-crd-btn" @click.stop.prevent="hideApplicationJson">{{$t('取消')}}</bk-button>
                <div :class="['diff-editor-box', { 'editor-fullscreen': editorOptions.fullScreen }]" style="position: relative;">
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
            </div>
        </bk-sideslider>
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
                fiexedStatus: ['not_deployed', 'deployed', 'failed', 'unknown'],
                isInitLoading: true,
                isPageLoading: false,
                crdControllerList: [],
                crdControllerListCache: [],
                curCrdcontroller: null,
                searchKeyword: '',
                searchScope: '',
                statusTimer: {},
                valueSlider: {
                    isShow: false,
                    fullScreen: false,
                    title: ''
                },
                renderEditorKey: 0,
                editorOptions: {
                    readOnly: false,
                    fontSize: 14,
                    fullScreen: false,
                    content: '',
                    originContent: '',
                    isDiff: false
                },
                dataSource: {
                    std_data_name: '',
                    file_data_name: '',
                    sys_data_name: ''
                },
                logMap: {
                    DbPrivilege: 'bcs-icon bcs-icon-db-auth',
                    BcsLog: 'bcs-icon bcs-icon-log',
                    GameStatefulSet: 'bcs-icon bcs-icon-gss',
                    GameDeployment: 'bcs-icon bcs-icon-gd',
                    PrometheusAdapter: 'bcs-icon bcs-icon-prom',
                    BcsIngressController: 'bcs-icon bcs-icon-bi-2',
                    BcsHookOperator: 'bcs-icon bcs-icon-bh',
                    BcsPolaris: 'bcs-icon bcs-icon-pol'
                }
            }
        },
        computed: {
            isEn () {
                return this.$store.state.isEn
            },
            projectId () {
                return this.$route.params.projectId
            },
            curProject () {
                return this.$store.state.curProject
            },
            crdKind () {
                return this.$route.meta.crdKind
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            },
            searchScopeList () {
                const clusterList = this.clusterList
                let results = []
                if (clusterList.length) {
                    results = []
                    clusterList.forEach(item => {
                        results.push({
                            id: item.cluster_id,
                            name: item.name
                        })
                    })
                }

                return results
            },
            editorHeight () {
                const height = window.innerHeight
                return this.editorOptions.fullScreen ? height : height - 80
            },
            curClusterId () {
                return this.$store.state.curClusterId
            }
        },
        watch: {
            curClusterId () {
                this.searchScope = this.curClusterId
                this.search()
            }
        },
        mounted () {
            this.init()
        },
        beforeRouteLeave (to, from, next) {
            this.clearAllInterval()
            next()
        },
        methods: {
            async init () {
                try {
                    if (this.clusterList.length) {
                        // if (this.curClusterId) {
                        //     this.searchScope = this.curClusterId
                        // } else {
                        //     this.searchScope = this.clusterList[0].cluster_id
                        // }
                        this.getCrdControllersByCluster()
                    } else {
                        this.isInitLoading = false
                        this.isPageLoading = false
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                    this.isInitLoading = false
                    this.isPageLoading = false
                }
                if (this.crdKind === 'BcsLog') {
                    this.getLogPlans()
                }
            },

            async haneldEnableCrdController (crdcontroller) {
                // 清空数据
                this.editorOptions.content = ''
                this.editorOptions.originContent = ''

                this.curCrdcontroller = crdcontroller
                if (crdcontroller.default_values_content) {
                    this.valueSlider.title = `${this.$t('启用组件：')}${crdcontroller.name}`
                    this.editorOptions.content = crdcontroller.default_values_content
                    this.editorOptions.originContent = crdcontroller.default_values_content
                    this.renderEditorKey++
                    this.valueSlider.isShow = true
                } else {
                    this.enableCrdController()
                }
            },

            async enableCrdController () {
                try {
                    const crdcontroller = this.curCrdcontroller
                    const projectId = this.projectId
                    const clusterId = this.searchScope
                    const id = crdcontroller.crd_ctr_id
                    const data = {
                        values_content: this.editorOptions.content
                    }
                    crdcontroller.status = 'pending'
                    this.valueSlider.isShow = false
                    await this.$store.dispatch('crdcontroller/enableCrdController', { projectId, clusterId, id, data })
                    // this.getCrdControllersByCluster()
                    this.getCrdcontrollerStatus(crdcontroller)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.editorOptions.readOnly = false
                }
            },

            hideApplicationJson () {
                this.valueSlider.isShow = false
                // 清空数据
                this.editorOptions.content = ''
                this.editorOptions.originContent = ''
            },

            async getCrdControllersByCluster () {
                if (this.isPageLoading) {
                    return false
                }
                if (!this.searchScope) {
                    return false
                }
                const projectId = this.projectId
                const clusterId = this.searchScope

                this.isPageLoading = true
                try {
                    const res = await this.$store.dispatch('crdcontroller/getCrdControllersByCluster', { projectId, clusterId })
                    res.data.forEach(item => {
                        const instance = item.crd_ctr_inst
                        item.cluster_id = instance.cluster_id || ''
                        item.message = instance.message || ''
                        item.status = instance.status || ''
                        item.values_content = instance.values_content || ''
                    })
                    // 搜索
                    let results = res.data.filter(item => {
                        if (this.crdKind === 'BcsLog') {
                            return item.name === 'BcsLog'
                        } else {
                            return item.name !== 'BcsLog'
                        }
                    })
                    if (this.searchKeyword.trim()) {
                        results = []
                        const keyword = this.searchKeyword.trim()
                        const keyList = ['display_name']
                        const list = res.data

                        list.forEach(item => {
                            item.isChecked = false
                            for (const key of keyList) {
                                if (item[key].indexOf(keyword) > -1) {
                                    results.push(item)
                                    return true
                                }
                            }
                        })
                    }
                    // results[0].status = 'pending'
                    this.crdControllerList = results
                    this.clearAllInterval()
                    this.crdControllerList.forEach(item => {
                        if (item.status === 'pending') {
                            this.getCrdcontrollerStatus(item)
                        }
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    setTimeout(() => {
                        this.isInitLoading = false
                        this.isPageLoading = false
                    }, 200)
                }
            },

            clearAllInterval () {
                for (const key in this.statusTimer) {
                    clearInterval(this.statusTimer[key])
                }
                this.statusTimer = {}
            },

            async getLogPlans () {
                const projectId = this.projectId

                try {
                    const res = await this.$store.dispatch('getLogPlans', projectId)
                    this.dataSource = res.data
                } catch (e) {
                    if (e.code !== 404) {
                        catchErrorHandler(e, this)
                    }
                }
            },

            async enableLogPlans () {
                const projectId = this.projectId
                try {
                    await this.$store.dispatch('enableLogPlans', projectId)
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            showInstanceDetail (crdcontroller) {
                if (window.sessionStorage) {
                    window.sessionStorage['bcs-cluster'] = this.searchScope
                    window.sessionStorage['bcs-crdcontroller'] = JSON.stringify(crdcontroller)
                }
                this.$router.push({
                    name: 'crdcontrollerInstanceDetail',
                    params: {
                        clusterId: this.searchScope,
                        name: crdcontroller.name,
                        id: crdcontroller.crd_ctr_id,
                        chartName: crdcontroller.chart_name
                    }
                })
            },

            async goControllerInstances (crdcontroller) {
                if (window.sessionStorage) {
                    window.sessionStorage['bcs-cluster'] = this.searchScope
                }

                if (this.crdKind === 'BcsLog') {
                    try {
                        const projectId = this.projectId
                        await this.$store.dispatch('enableLogPlans', projectId)
                        this.$router.push({
                            name: 'crdcontrollerLogInstances',
                            params: {
                                crdKind: crdcontroller.name,
                                clusterId: this.searchScope
                            }
                        })
                    } catch (e) {
                        if (e.code !== 404) {
                            catchErrorHandler(e, this)
                        }
                    }
                } else {
                    if (crdcontroller.name === 'DbPrivilege') {
                        this.$router.push({
                            name: 'crdcontrollerDBInstances',
                            params: {
                                crdKind: crdcontroller.name,
                                clusterId: this.searchScope
                            }
                        })
                    } else if (crdcontroller.name === 'BcsPolaris') {
                        this.$router.push({
                            name: 'crdcontrollerPolarisInstances',
                            params: {
                                crdKind: crdcontroller.name,
                                clusterId: this.searchScope
                            }
                        })
                    }
                }
            },

            search () {
                this.getCrdControllersByCluster()
            },

            refresh () {
                this.searchKeyword = ''
                this.getCrdControllersByCluster()
            },

            /**
             * 简单判断是否为图片
             * @param  {string} img 图片url
             * @return {Boolean} true/false
             */
            isImage (img) {
                if (!img) {
                    return false
                }
                if (img.startsWith('http://') || img.startsWith('https://') || img.startsWith('data:image/')) {
                    return true
                }
                return false
            },

            /**
             * 获取crdcontroller状态
             * @param  {object} crdcontroller crdcontroller
             * @param  {number} index 索引
             */
            getCrdcontrollerStatus (crdcontroller, index) {
                if (crdcontroller.crd_ctr_id === undefined) {
                    return false
                }
                const projectId = this.projectId
                const crdcontrollerId = crdcontroller.crd_ctr_id
                const clusterId = crdcontroller.cluster_id || this.searchScope
                const self = this

                clearInterval(this.statusTimer[crdcontroller.crd_ctr_id])
                this.statusTimer[crdcontroller.crd_ctr_id] = setInterval(async () => {
                    try {
                        const res = await this.$store.dispatch('crdcontroller/getCrdcontrollerStatus', {
                            projectId,
                            clusterId,
                            crdcontrollerId
                        })
                        if (this.fiexedStatus.includes(res.data.status)) {
                            clearInterval(self.statusTimer[crdcontroller.crd_ctr_id])
                            this.crdControllerList.forEach(item => {
                                if (item.crd_ctr_id === crdcontrollerId) {
                                    item.status = res.data.status
                                    item.message = res.data.message
                                }
                            })
                        }
                    } catch (e) {
                        catchErrorHandler(e, this)
                    }
                }, 2000)
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
