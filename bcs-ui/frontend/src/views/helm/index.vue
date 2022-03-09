<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                {{$t('Helm Release列表')}}
            </div>
            <bk-guide>
                <a class="bk-text-button" :href="PROJECT_CONFIG.doc.serviceAccess" target="_blank">{{$t('如何使用Helm？')}}</a>
            </bk-guide>
        </div>
        <div class="biz-content-wrapper biz-helm-wrapper m0 p0" v-bkloading="{ isLoading: showLoading, opacity: 0.1 }">
            <template v-if="!showLoading">
                <app-exception
                    v-if="exceptionCode && !showLoading"
                    :type="exceptionCode.code"
                    :text="exceptionCode.msg">
                </app-exception>

                <div class="biz-panel-header p20">
                    <div class="left">
                        <!-- <router-link class="bk-button bk-primary" :to="{ name: 'helmTplList' }">
                            <i class="bcs-icon bcs-icon-plus"></i>
                            <span>{{$t('部署Helm Chart')}}</span>
                        </router-link> -->
                        <!-- <bcs-button>批量下载</bcs-button>
                        <bcs-button>批量删除</bcs-button> -->
                    </div>
                    <div class="right">
                        <search
                            :width-refresh="false"
                            :scope-list="searchScopeList"
                            :namespace-list="namespaceList"
                            :search-key.sync="searchKeyword"
                            :search-scope.sync="searchScope"
                            :search-namespace.sync="searchNamespace"
                            :cluster-fixed="!!curClusterId"
                            @cluster-change="handleClusterChange"
                            @search="handleSearch"
                            @refresh="handleRefresh">
                        </search>
                    </div>
                </div>

                <svg style="display: none;">
                    <title>{{$t('模板集默认图标')}}</title>
                    <symbol id="biz-set-icon" viewBox="0 0 60 60">
                        <path class="st0" d="M54.8,16.5L34,4.5C33.4,4.2,32.7,4,32,4s-1.4,0.2-2,0.5l-20.8,12c-1.2,0.7-2,2-2,3.5v24c0,1.4,0.8,2.7,2,3.5
                            l20.8,12c0.6,0.4,1.3,0.5,2,0.5s1.4-0.2,2-0.5l20.8-12c1.2-0.7,2-2,2-3.5V20C56.8,18.6,56,17.3,54.8,16.5z M11.2,20L11.2,20L11.2,20
                            L11.2,20z M30,54.8L11.2,44V22.3L30,33.2V54.8z M32,29.7L13.2,18.8L32,8l18.8,10.8L32,29.7z M52,28.1c-1.2,0.7-1.8,1.3-1.8,2v10.7
                            c0,0.6,0.6,0.6,1.8-0.1v1.1l-6.7,3.9v-1.1c1.3-0.7,1.9-1.4,1.9-2v-5l-6.8,3.9v5c0,0.6,0.6,0.6,1.9-0.2v1.1l-6.7,3.9v-1.1
                            c1.2-0.7,1.8-1.3,1.8-1.9V37.5c0-0.6-0.6-0.6-1.8,0.1v-1.2l6.7-3.9v1.2c-1.3,0.7-1.9,1.4-1.9,2V40l6.8-3.9v-4.2
                            c0-0.6-0.6-0.6-1.9,0.2v-1.2L52,27V28.1z M52.8,20L52.8,20L52.8,20L52.8,20z" />
                    </symbol>
                </svg>

                <div class="biz-namespace" style="padding-bottom: 100px;" v-bkloading="{ isLoading: isPageLoading }">
                    <bk-table
                        :data="curPageData"
                        size="small"
                        :pagination="pagination"
                        @page-change="handlePageChange"
                        @page-limit-change="handlePageLimitChange">
                        <!-- <bk-table-column key="selection" :render-header="renderSelectionHeader" width="50">
                            <template slot-scope="{ row }">
                                <bk-checkbox name="check-strategy" v-model="row.isChecked" @change="checkApp(row)" />
                            </template>
                        </bk-table-column> -->
                        <bk-table-column :label="$t('Release名称')" min-width="160">
                            <template slot-scope="{ row }">
                                <div>
                                    <span v-if="row.transitioning_on" class="f14 fb app-name">
                                        {{ row.name }}
                                    </span>
                                    <a @click="showAppDetail(row)"
                                        href="javascript:void(0)"
                                        class="bk-text-button app-name f14"
                                        v-authority="{
                                            clickable: webAnnotationsPerms[row.iam_ns_id]
                                                && webAnnotationsPerms[row.iam_ns_id].namespace_scoped_view,
                                            actionId: 'namespace_scoped_view',
                                            resourceName: row.namespace,
                                            disablePerms: true,
                                            permCtx: {
                                                project_id: projectId,
                                                cluster_id: row.cluster_id,
                                                name: row.namespace
                                            }
                                        }"
                                        v-else>
                                        {{ row.name }}
                                    </a>
                                </div>
                                <template v-if="row.transitioning_on">
                                    <bk-tag theme="warning mt5" style="margin-left: -5px;">
                                        <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-warning">
                                            <div class="rotate rotate1"></div>
                                            <div class="rotate rotate2"></div>
                                            <div class="rotate rotate3"></div>
                                            <div class="rotate rotate4"></div>
                                            <div class="rotate rotate5"></div>
                                            <div class="rotate rotate6"></div>
                                            <div class="rotate rotate7"></div>
                                            <div class="rotate rotate8"></div>
                                        </div>
                                        {{appAction[row.transitioning_action]}}中...
                                    </bk-tag>
                                </template>
                                <template v-else-if="!row.transitioning_result && row.transitioning_action !== 'noop'">
                                    <bcs-popover :content="$t('点击查看原因')" placement="top" style="margin-left: -5px;">
                                        <bk-tag class="m0 mt5" type="filled" theme="danger" style="cursor: pointer;" @click.native="showAppError(row)">
                                            <i class="bcs-icon bcs-icon-order"></i>
                                            {{appAction[row.transitioning_action]}}{{$t('失败')}}
                                        </bk-tag>
                                    </bcs-popover>
                                </template>
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Chart" prop="source" min-width="160">
                            <template slot-scope="{ row }">
                                {{ `${row.chart_name}:${row.current_version}` }}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('集群')" prop="status" :show-overflow-tooltip="false" min-width="250">
                            <template slot-scope="{ row }">
                                <div class="col-cluster">
                                    {{$t('所属集群')}}：
                                    <bcs-popover :content="row.cluster_id || '--'" placement="top">
                                        <span>{{row.cluster_name ? row.cluster_name : '--'}}</span>
                                    </bcs-popover>
                                    <template v-if="row.cluster_env === 'stag'">
                                        <bk-tag type="filled" theme="warning" class="biz-small-tag m0">{{$t('测试')}}</bk-tag>
                                    </template>
                                    <template v-else-if="row.cluster_env === 'prod'">
                                        <bk-tag type="filled" theme="success" class="biz-small-tag m0">{{$t('正式')}}</bk-tag>
                                    </template>
                                </div>
                                <p>
                                    {{$t('命名空间')}}：<span class="biz-text-wrapper ml5">{{row.namespace}}</span>
                                </p>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('操作记录')" prop="create_time" width="260">
                            <template slot-scope="{ row }">
                                <p class="updator">{{$t('操作者')}}：{{ row.updator }}</p>
                                <p class="updated">{{$t('更新时间')}}：{{ row.updated }}</p>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('操作')" width="230">
                            <template slot-scope="{ row }">
                                <bk-button class="ml5"
                                    text
                                    v-authority="{
                                        clickable: webAnnotationsPerms[row.iam_ns_id]
                                            && webAnnotationsPerms[row.iam_ns_id].namespace_scoped_view,
                                        actionId: 'namespace_scoped_view',
                                        resourceName: row.namespace,
                                        disablePerms: true,
                                        permCtx: {
                                            project_id: projectId,
                                            cluster_id: row.cluster_id,
                                            name: row.namespace
                                        }
                                    }"
                                    @click="showAppInfoSlider(row)"
                                >{{ $t('查看状态') }}</bk-button>
                                <bk-button class="ml5"
                                    text
                                    v-authority="{
                                        clickable: webAnnotationsPerms[row.iam_ns_id]
                                            && webAnnotationsPerms[row.iam_ns_id].namespace_scoped_update,
                                        actionId: 'namespace_scoped_update',
                                        resourceName: row.namespace,
                                        disablePerms: true,
                                        permCtx: {
                                            project_id: projectId,
                                            cluster_id: row.cluster_id,
                                            name: row.namespace
                                        }
                                    }"
                                    @click="showAppDetail(row)"
                                >{{ $t('更新') }}</bk-button>
                                <bk-button class="ml5"
                                    text
                                    v-authority="{
                                        clickable: webAnnotationsPerms[row.iam_ns_id]
                                            && webAnnotationsPerms[row.iam_ns_id].namespace_scoped_update,
                                        actionId: 'namespace_scoped_update',
                                        resourceName: row.namespace,
                                        disablePerms: true,
                                        permCtx: {
                                            project_id: projectId,
                                            cluster_id: row.cluster_id,
                                            name: row.namespace
                                        }
                                    }"
                                    @click="showRebackDialog(row)"
                                >{{ $t('回滚') }}</bk-button>
                                <bk-button class="ml5"
                                    text
                                    v-authority="{
                                        clickable: webAnnotationsPerms[row.iam_ns_id]
                                            && webAnnotationsPerms[row.iam_ns_id].namespace_scoped_delete,
                                        actionId: 'namespace_scoped_delete',
                                        resourceName: row.namespace,
                                        disablePerms: true,
                                        permCtx: {
                                            project_id: projectId,
                                            cluster_id: row.cluster_id,
                                            name: row.namespace
                                        }
                                    }"
                                    @click="deleteApp(row)"
                                >{{ $t('删除') }}</bk-button>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </div>
            </template>
        </div>

        <bk-dialog
            width="800"
            :title="rebackDialogConf.title"
            :close-icon="!isRebackLoading"
            :quick-close="false"
            :is-show.sync="rebackDialogConf.isShow"
            @cancel="cancelReback">
            <template slot="content">
                <div class="flex" style="margin-top: -15px;">
                    <div class="bk-form bk-form-vertical" style="width: 760px;">
                        <div class="bk-form-item mb20">
                            <label for="" class="bk-label">
                                {{$t('回滚到版本')}} <span class="error-tip" v-if="!isRebackListLoading && !rebackList.length">{{$t('（Release当前没有可切换的版本，无法回滚）')}}</span>
                            </label>
                            <div class="bk-form-content mb10">
                                <bk-selector
                                    :placeholder="$t('请选择')"
                                    :selected.sync="versionId"
                                    :list="rebackList"
                                    :setting-key="'id'"
                                    :disabled="isRebackLoading"
                                    :display-key="'version'"
                                    @item-selected="showRebackPreview">
                                </bk-selector>
                            </div>

                            <div style="height: 370px;" v-bkloading="{ isLoading: isRebackVersionLoading }" v-if="versionId">
                                <bk-tab
                                    class="biz-special-tab"
                                    :type="'fill'"
                                    :size="'small'"
                                    :active-name="'Difference'"
                                    :key="rebackPreviewList.length">
                                    <bk-tab-panel :name="'Difference'" :title="$t('版本对比')">
                                        <div style="height: 320px;">
                                            <ace
                                                :value="difference"
                                                :width="rebackEditorConfig.width"
                                                :height="rebackEditorConfig.height"
                                                :lang="rebackEditorConfig.lang"
                                                :read-only="rebackEditorConfig.readOnly"
                                                :full-screen="rebackEditorConfig.fullScreen">
                                            </ace>
                                        </div>
                                    </bk-tab-panel>
                                    <bk-tab-panel :key="index" :name="item.name" :title="item.name" v-for="(item, index) in rebackPreviewList">
                                        <div style="height: 320px;">
                                            <ace
                                                :value="item.value"
                                                :width="rebackEditorConfig.width"
                                                :height="rebackEditorConfig.height"
                                                :lang="rebackEditorConfig.lang"
                                                :read-only="rebackEditorConfig.readOnly"
                                                :full-screen="rebackEditorConfig.fullScreen">
                                            </ace>
                                        </div>
                                    </bk-tab-panel>
                                </bk-tab>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
            <div slot="footer">
                <div class="bk-dialog-outer">
                    <template>
                        <bk-button theme="primary" :loading="isRebackLoading" :disabled="isRebackVersionLoading || !versionId" @click="submitRebackData">
                            {{$t('确定')}}
                        </bk-button>
                        <bk-button :disabled="isRebackLoading" @click="cancelReback">
                            {{$t('取消')}}
                        </bk-button>
                    </template>
                </div>
            </div>
        </bk-dialog>

        <bk-dialog
            :is-show.sync="errorDialogConf.isShow"
            :width="750"
            :has-footet="false"
            :title="errorDialogConf.title"
            @cancel="hideErrorDialog">
            <template slot="content">
                <div class="bk-intro bk-danger pb30 mb15" v-if="errorDialogConf.message" style="position: relative;">
                    <pre class="biz-error-message">
                        {{errorDialogConf.message}}
                    </pre>
                    <bk-button size="small" type="default" id="error-copy-btn" :data-clipboard-text="errorDialogConf.message"><i class="bcs-icon bcs-icon-clipboard mr5"></i>{{$t('复制')}}</bk-button>
                </div>
            </template>
            <div slot="footer">
                <div class="biz-footer">
                    <bk-button type="primary" @click="hideErrorDialog">{{$t('知道了')}}</bk-button>
                </div>
            </div>
        </bk-dialog>

        <bk-sideslider
            :is-show.sync="appInfoConf.isShow"
            :title="appInfoConf.title"
            :quick-close="true"
            :width="800"
            @hidden="hideAppInfoSlider">
            <div slot="content" :style="{ height: `${winHeight - 100}px`, padding: '20px' }" v-bkloading="{ isLoading: isAppInfoLoading }">
                <div class="biz-search-input" style="width: 240px; float: right; margin-top: -68px;" v-if="!isAppInfoLoading">
                    <bk-input right-icon="bk-icon icon-search"
                        clearable
                        :placeholder="$t('输入关键字，按Enter搜索')"
                        v-model="resourceSearchKey"
                        @enter="searchResource"
                        @clear="clearResoureceSearch" />
                </div>
                <table
                    class="bk-table has-table-hover biz-data-table"
                    v-if="!isAppInfoLoading"
                    style="border: 1px solid #e6e6e6; border-bottom: none;">
                    <thead>
                        <tr>
                            <th style="width: 10px; padding: 0;"></th>
                            <th>{{$t('名称')}}</th>
                            <th style="width: 130px;">{{$t('类型')}}</th>
                            <th style="width: 100px;">
                                Pods
                                <bcs-popover :content="$t('实际实例数/期望数')" placement="right">
                                    <i class="bcs-icon bcs-icon-info-circle tip-trigger"></i>
                                </bcs-popover>
                            </th>
                        </tr>
                    </thead>
                    <tbody>
                        <template v-if="curAppResources.length">
                            <template v-for="resource of curAppResources">
                                <tr
                                    @click="showErrorInfo(resource)"
                                    :class="{ 'has-error': resource.pods.warnings }"
                                    :key="resource.id">
                                    <td style="padding: 0;">
                                        <template v-if="resource.pods.warnings">
                                            <bcs-popover :content="$t('点击查看原因')" placement="left">
                                                <i class="biz-status-icon bcs-icon bcs-icon-info-circle tip-trigger biz-danger-text f13"></i>
                                            </bcs-popover>
                                        </template>
                                        <i class="biz-status-icon bcs-icon bcs-icon-check-circle biz-success-text f13" v-else></i>
                                    </td>
                                    <td>
                                        <template v-if="resource.link">
                                            <a href="javascript:void(0);" class="bk-text-button" @click.stop.prevent="goResourceInfo(resource.link)">{{resource.name}}</a>
                                        </template>
                                        <template v-else>
                                            {{resource.name}}
                                        </template>
                                    </td>
                                    <td>
                                        <bk-tag type="filled" theme="info">{{resource.kind}}</bk-tag>
                                    </td>
                                    <td>
                                        <template v-if="resource.pods.running !== 0 || resource.pods.desired !== 0">
                                            {{resource.pods.running}}/{{resource.pods.desired}}
                                        </template>
                                        <template v-else>
                                            --
                                        </template>
                                    </td>
                                </tr>
                                <tr v-if="resource.isOpened && resource.pods.warnings" :key="resource.id">
                                    <td colspan="4">
                                        <pre class="bk-intro bk-danger biz-error-message mb0">
                                            {{resource.pods.warnings}}
                                        </pre>
                                    </td>
                                </tr>
                            </template>
                        </template>
                        <template v-if="curAppResources.length === 0">
                            <tr>
                                <td colspan="4">
                                    <bcs-exception type="empty" scene="part"></bcs-exception>
                                </td>
                            </tr>
                        </template>
                    </tbody>
                </table>
            </div>
        </bk-sideslider>
    </div>
</template>

<script>
    import ace from '@/components/ace-editor'
    import { catchErrorHandler } from '@/common/util'
    import Clipboard from 'clipboard'
    import search from './search.vue'
    import { mapGetters } from 'vuex'

    const FAST_TIME = 3000
    const SLOW_TIME = 10000

    export default {
        components: {
            ace,
            search
        },
        data () {
            return {
                curApp: {},
                curAppResources: [],
                curAppResourcesCache: [],
                namespaceList: [],
                statusTimer: 0,
                isRebackLoading: false,
                isRebackListLoading: false,
                isRebackVersionLoading: false,
                isRouterLeave: false,
                isAppInfoLoading: false,
                appList: [],
                appListCache: [],
                showLoading: true,
                isPageLoading: false,
                exceptionCode: null,
                versionId: '',
                difference: '',
                versionList: [],
                rebackPreviewList: [],
                rebackList: [],
                editor: null,
                searchKeyword: '',
                searchScope: '',
                searchNamespace: '',
                previewLoading: true,
                rebackDialogConf: {
                    title: '',
                    isShow: false
                },
                appInfoConf: {
                    isShow: false,
                    title: ''
                },
                curAppDetail: {
                    created: '',
                    namespace_id: '',
                    release: {
                        id: '',
                        customs: [],
                        answers: {}
                    }
                },
                errorDialogConf: {
                    title: '',
                    isShow: false,
                    message: '',
                    errorCode: 0
                },
                curProjectId: '',
                winHeight: 0,
                editorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'json',
                    readOnly: true,
                    fullScreen: false,
                    value: '',
                    editors: []
                },
                resourceSearchKey: '',
                rebackEditorConfig: {
                    width: '100%',
                    height: '100%',
                    lang: 'yaml',
                    readOnly: true,
                    fullScreen: false,
                    value: '',
                    editors: []
                },
                operaRunningApp: {}, // 缓存操作更新中的app状态信息
                appAction: {
                    create: this.$t('部署'),
                    noop: '',
                    update: this.$t('更新'),
                    rollback: this.$t('回滚'),
                    delete: this.$t('删除'),
                    destroy: this.$t('删除')
                },
                isOperaLayerShow: false, // 操作弹层显示，包括删除和回滚
                appCheckTime: FAST_TIME,
                pagination: {
                    current: 1,
                    count: 0,
                    limit: 10
                },
                selectLists: [],
                isCheckAll: false, // 表格全选状态
                timeOutFlag: false,
                webAnnotationsPerms: {}
            }
        },
        computed: {
            curProject () {
                const project = this.$store.state.curProject
                return project
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            projectId () {
                this.curProjectId = this.$route.params.projectId
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            searchScopeList () {
                const clusterList = this.$store.state.cluster.clusterList
                const results = []
                if (clusterList.length) {
                    clusterList.forEach(item => {
                        results.push({
                            id: item.cluster_id,
                            name: item.name
                        })
                    })
                }

                return results
            },
            ...mapGetters('cluster', ['isSharedCluster'])
        },
        watch: {
            curProjectId () {
                // 如果不是k8s类型的项目，无法访问些页面，重定向回集群首页
                if (this.curProject && (this.curProject.kind !== PROJECT_K8S && this.curProject.kind !== PROJECT_TKE)) {
                    this.$router.push({
                        name: 'clusterMain',
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                }
            },
            curClusterId () {
                this.searchScope = this.curClusterId
                this.searchNamespace = ''
                this.handleSearch()
            }
        },
        created () {
            this.searchScope = this.searchScopeList[0]?.id
        },
        mounted () {
            this.isRouterLeave = false
            this.winHeight = window.innerHeight
            if (window.sessionStorage && window.sessionStorage['bcs-cluster']) {
                this.searchScope = window.sessionStorage['bcs-cluster']
            }
            if (window.sessionStorage && window.sessionStorage['bcs-helm-namespace']) {
                this.searchNamespace = window.sessionStorage['bcs-helm-namespace']
            }

            this.getAppList()
            this.getNamespaces()
        },
        beforeRouteLeave (to, from, next) {
            this.isRouterLeave = true
            // 如果不是到详情内页，清空搜索条件
            if (to.name !== 'helmAppDetail') {
                // window.sessionStorage['bcs-helm-cluster'] = ''
                window.sessionStorage['bcs-helm-namespace'] = ''
            }
            clearTimeout(this.statusTimer)
            next()
        },
        beforeDestroy () {
            this.isRouterLeave = true
            this.timeOutFlag = true
            clearTimeout(this.statusTimer)
            // window.sessionStorage['bcs-helm-cluster'] = ''
            window.sessionStorage['bcs-helm-namespace'] = ''
        },
        methods: {
            /**
             * 刷新列表
             */
            handleRefresh () {
                this.getAppList()
            },

            /**
             * 搜索列表
             */
            handleSearch () {
                window.sessionStorage['bcs-cluster'] = this.searchScope
                window.sessionStorage['bcs-helm-namespace'] = this.searchNamespace
                this.pagination.count = 0
                this.pagination.current = 1
                this.getAppList()
            },

            /**
             * 查看资源详情
             * @param  {string} link 资源链接
             */
            goResourceInfo (link) {
                const clusterId = this.curApp.cluster_id
                if (this.isSharedCluster) {
                    const route = this.$router.resolve({ name: 'dashboardWorkload' })
                    window.open(route.href)
                } else {
                    const url = `${window.location.origin}${link}&cluster_id=${clusterId}`
                    window.open(url)
                }
            },

            /**
             * 显示资源异常信息
             * @param  {object} resource 资源
             */
            showErrorInfo (resource) {
                resource.isOpened = !resource.isOpened
            },

            /**
             * 显示应用状态信息
             * @params {object} app 应用对象
             */
            async showAppInfoSlider (app) {
                this.curAppResources = []
                this.appInfoConf.isShow = true
                this.isOperaLayerShow = true
                this.appInfoConf.title = app.name
                this.curApp = app

                const projectId = this.projectId
                const appId = app.id

                this.isAppInfoLoading = true
                try {
                    const res = await this.$store.dispatch('helm/getAppInfo', { projectId, appId })
                    const resources = []
                    const appResources = res.data.status
                    for (const key in appResources) {
                        const resource = appResources[key]
                        const metedata = {
                            name: resource.name,
                            kind: resource.kind,
                            link: resource.link,
                            isOpened: false,
                            pods: {
                                desired: resource.status_sumary.desired_pods, // 期望数
                                running: resource.status_sumary.ready_pods, // 实例实例数
                                warnings: resource.status_sumary.messages
                            }
                        }
                        resources.push(metedata)
                    }
                    this.curAppResources = this.sortResource(resources)
                    this.curAppResourcesCache = JSON.parse(JSON.stringify(this.curAppResources))
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isAppInfoLoading = false
                }
            },

            /**
             * 按顺序对资源进行靠近排序（相同类型临近一起）
             * @param  {array} resources 资源
             * @return {array} arrayCache 排序结果
             */
            sortResource (resources) {
                let sortCache = []
                const sortKey = {
                    'Deployment': [],
                    'DaemonSet': [],
                    'Job': [],
                    'StatefulSet': [],
                    'Service': [],
                    'Ingress': [],
                    'ConfigMap': [],
                    'Secret': [],
                    'other': []
                }

                resources.forEach(resource => {
                    const kind = resource.kind
                    if (sortKey[kind]) {
                        sortKey[kind].push(resource)
                    } else {
                        sortKey['other'].push(resource)
                    }
                })

                sortKey['other'].sort((a, b) => {
                    return a.kind.toLowerCase() < b.kind.toLowerCase()
                })

                for (const key in sortKey) {
                    sortCache = sortCache.concat(...sortKey[key])
                }
                return sortCache
            },

            /**
             * 显示应用详情
             * @param {object} app 应用
             */
            async showAppDetail (app) {
                this.$router.push({
                    name: 'helmAppDetail',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        appId: app.id
                    }
                })
            },

            /**
             * 隐藏错误提示弹层
             */
            hideErrorDialog () {
                this.errorDialogConf.isShow = false
                this.isOperaLayerShow = false
                this.appCheckTime = FAST_TIME
                this.getAppsStatus()
            },

            /**
             * 确认删除应用
             * @param {object} app 应用
             */
            async deleteApp (app) {
                const projectId = this.projectId
                const appId = app.id
                const me = this
                // const boxStyle = {
                //     'margin-top': '-20px',
                //     'margin-bottom': '-20px'
                // }
                // const titleStyle = {
                //     style: {
                //         'text-align': 'left',
                //         'font-size': '14px',
                //         'margin-bottom': '10px',
                //         'color': '#313238'
                //     }
                // }
                // const itemStyle = {
                //     style: {
                //         'text-align': 'left',
                //         'font-size': '14px',
                //         'margin-bottom': '3px',
                //         'color': '#71747c'
                //     }
                // }

                clearTimeout(this.statusTimer)
                this.isOperaLayerShow = true
                this.$bkInfo({
                    title: this.$t('确定删除该应用?'),
                    clsName: 'biz-remove-dialog',
                    defaultInfo: true,
                    // content: me.$createElement('p', {
                    //     class: 'biz-confirm-desc'
                    // }, `确定要删除Release【${app.name}】？`),
                    // content: me.$createElement('div', { class: 'biz-confirm-desc', style: boxStyle }, [
                    //     // me.$createElement('p', titleStyle, this.$t('确定要删除Release？')),
                    //     me.$createElement('p', itemStyle, `${this.$t('名称')}：${app.name}`),
                    //     me.$createElement('p', itemStyle, `${this.$t('所属集群')}：${app.cluster_name}`),
                    //     me.$createElement('p', itemStyle, `${this.$t('命名空间')}：${app.namespace}`)
                    // ]),
                    async confirmFn () {
                        app.transitioning_action = 'delete'
                        app.transitioning_on = true
                        try {
                            await me.$store.dispatch('helm/deleteApp', { projectId, appId }, { cancelPrevious: true })
                            me.checkingAppStatus(app, 'delete')
                        } catch (e) {
                            catchErrorHandler(e, this)
                        } finally {
                            me.isOperaLayerShow = false
                            me.showLoading = false
                        }
                    },
                    cancelFn (close) {
                        me.appCheckTime = FAST_TIME
                        me.isOperaLayerShow = false
                        me.getAppsStatus()
                        close()
                    }
                })
            },

            /**
             * 展示App异常信息
             * @param  {object} app 应用对象
             */
            showAppError (app) {
                let actionType = ''
                const res = {
                    code: 500,
                    message: ''
                }

                res.message = app.transitioning_message
                actionType = app.transitioning_action

                const title = `${app.name}${this.appAction[app.transitioning_action]}${this.$t('失败')}`
                this.showErrorDialog(res, title, actionType)
            },

            /**
             * 显示错误弹层
             * @param  {object} res ajax数据对象
             * @param  {string} title 错误提示
             * @param  {string} actionType 操作
             */
            showErrorDialog (res, title, actionType) {
                // 先检查集群是否注册到 BKE server。未注册则返回 code: 40031
                this.errorDialogConf.errorCode = res.code
                this.errorDialogConf.message = res.message || res.data.msg || res.statusText
                this.errorDialogConf.isShow = true
                this.isOperaLayerShow = true
                this.errorDialogConf.title = title
                this.rebackDialogConf.isShow = false
                this.errorDialogConf.actionType = actionType

                if (this.clipboardInstance && this.clipboardInstance.off) {
                    this.clipboardInstance.off('success')
                }
                if (this.errorDialogConf.message) {
                    this.$nextTick(() => {
                        this.clipboardInstance = new Clipboard('#error-copy-btn')
                        this.clipboardInstance.on('success', e => {
                            this.$bkMessage({
                                theme: 'success',
                                message: this.$t('复制成功')
                            })
                            this.isVarPanelShow = false
                        })
                    })
                }
            },

            /**
             * 搜索Helm app
             */
            search () {
                const keyword = this.searchKeyword
                const keyList = ['name', 'namespace', 'cluster_name']
                const list = JSON.parse(JSON.stringify(this.appListCache))
                if (keyword) {
                    const results = list.filter(item => {
                        for (const key of keyList) {
                            if (item[key].indexOf(keyword) > -1) {
                                return true
                            }
                        }
                        return false
                    })
                    this.appList.splice(0, this.appList.length, ...results)
                    this.curPageData = this.getDataByPage(this.pagination.current, false)
                    this.pagination.count = this.appList.length
                } else {
                    // 没有搜索关键字，直接从缓存返回列表
                    this.appList.splice(0, this.appList.length, ...list)
                    this.curPageData = this.getDataByPage(this.pagination.current, false)
                    this.pagination.count = this.appList.length
                }
            },

            /**
             * 搜索resource
             */
            searchResource () {
                const keyword = this.resourceSearchKey
                if (keyword) {
                    const results = this.curAppResourcesCache.filter(item => {
                        if (item.name.indexOf(keyword) > -1 || item.kind.indexOf(keyword) > -1) {
                            return true
                        } else {
                            return false
                        }
                    })
                    this.curAppResources.splice(0, this.curAppResources.length, ...results)
                } else {
                    // 没有搜索关键字，直接从缓存返回列表
                    this.curAppResources.splice(0, this.curAppResources.length, ...this.curAppResourcesCache)
                }
            },

            /**
             * 清除Helm app搜索
             */
            clearSearch () {
                this.searchKeyword = ''
                this.search()
            },

            /**
             * 清空资源搜索
             */
            clearResoureceSearch () {
                this.resourceSearchKey = ''
                this.searchResource()
            },

            /**
             * ace编辑器初始化成功回调
             * @param  {object} editor ace
             */
            handlerEditorInit (editor) {
                this.editor = editor
            },

            /**
             *  显示回滚弹层
             * @param  {object} app 应用
             */
            async showRebackDialog (app) {
                clearTimeout(this.statusTimer)

                this.curApp = app
                this.versionId = ''
                this.isRebackVersionLoading = false
                this.rebackDialogConf.isShow = true
                this.isOperaLayerShow = true
                this.rebackDialogConf.title = `${this.$t('回滚')} ${app.name}`
                this.rebackPreviewList = []
                this.rebackList = []
                this.isRebackListLoading = true
                this.getRebackList(app.id)
            },

            /**
             * 获取回滚版本列表
             * @param  {number} appId 应用ID
             */
            async getRebackList (appId) {
                const projectId = this.projectId
                this.isRebackListLoading = true

                try {
                    const res = await this.$store.dispatch('helm/getRebackList', { projectId, appId })

                    if (res.data.results) {
                        res.data.results.forEach(item => {
                            item.version = `${this.$t('版本')}：${item.version} （${this.$t('部署时间')}：${item.created_at}） `
                        })
                        this.rebackList = res.data.results
                    } else {
                        this.rebackList = []
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isRebackListLoading = false
                }
            },

            getParams () {
                const data = {
                    projectId: this.projectId,
                    params: {
                        limit: this.pagination.limit,
                        page: this.pagination.current,
                        offset: 0,
                        cluster_id: this.searchScope,
                        namespace: '',
                        keyword: this.keyword
                    }
                }
                if (this.searchNamespace) {
                    const args = this.searchNamespace.split(':')
                    data.params.cluster_id = args[0]
                    data.params.namespace = args[1]
                }
                return data
            },

            /**
             * 获取应用列表
             */
            async getAppList (reload) {
                if (reload) {
                    this.searchKeyword = ''
                }
                this.isPageLoading = true
                try {
                    clearTimeout(this.statusTimer)
                    const data = this.getParams()
                    const res = await this.$store.dispatch('helm/getAppList', data)
                    this.searchScope = data.params.cluster_id
                    this.pagination.count = res.data.results.length
                    this.appList = res.data.results
                    this.webAnnotationsPerms = Object.assign(this.webAnnotationsPerms, res.web_annotations?.perms || {})
                    this.curPageData = this.getDataByPage(this.pagination.current, false)

                    this.appListCache = JSON.parse(JSON.stringify(res.data.results))

                    this.getAppsStatus()

                    // 按原关键字再搜索
                    if (this.searchKeyword) {
                        this.search()
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.showLoading = false
                    this.isPageLoading = false
                }
            },

            /**
             * 获取所有命名空间列表
             */
            async getNamespaces (reload) {
                try {
                    clearTimeout(this.statusTimer)
                    const res = await this.$store.dispatch('helm/getNamespaceList', {
                        projectId: this.projectId,
                        params: {
                            cluster_id: this.searchScope
                        }
                    })
                    this.namespaceList = (res.data || []).map(item => {
                        return {
                            ...item,
                            namespace_id: `${item.cluster_id}:${item.name}`
                        }
                    })
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            handleClusterChange () {
                this.searchNamespace = ''
                this.getNamespaces()
            },

            /**
             * 提交回滚数据
             * @return {[type]} [description]
             */
            async submitRebackData () {
                const projectId = this.projectId
                const appId = this.curApp.id
                const params = {
                    release: this.versionId
                }

                if (this.isRebackLoading || !this.versionId) {
                    return false
                } else {
                    this.isRebackLoading = true
                }

                try {
                    await this.$store.dispatch('helm/reback', { projectId, appId, params })
                    this.rebackDialogConf.isShow = false
                    this.isOperaLayerShow = false
                    this.checkingAppStatus(this.curApp, 'rollback')
                } catch (e) {
                    this.showErrorDialog(e, this.$t('回滚失败'), 'reback')
                } finally {
                    this.isRebackLoading = false
                }
            },

            updateApp (appId, status) {
                this.appList.forEach((app, index) => {
                    if (app.id === appId) {
                        app.transitioning_action = status.transitioning_action
                        app.transitioning_message = status.transitioning_message
                        app.transitioning_on = status.transitioning_on
                        app.transitioning_result = status.transitioning_result
                    }
                })
                this.appListCache.forEach((app, index) => {
                    if (app.id === appId) {
                        app.transitioning_action = status.transitioning_action
                        app.transitioning_message = status.transitioning_message
                        app.transitioning_on = status.transitioning_on
                        app.transitioning_result = status.transitioning_result
                    }
                })
            },

            /**
             * 查询应用的状态
             * @param  {object} app 应用对象
             */
            checkingAppStatus (app, action) {
                if (app) {
                    const status = {
                        name: app.name,
                        transitioning_on: true,
                        transitioning_action: action,
                        transitioning_result: false,
                        transitioning_message: ''
                    }
                    this.operaRunningApp[app.id] = status
                    this.updateApp(app.id, status)
                }

                this.appCheckTime = FAST_TIME
                this.getAppsStatus()
            },

            /**
             * 查看app状态，包括创建、更新、回滚、删除
             * @param  {object} app 应用对象
             */
            getAppsStatus () {
                clearTimeout(this.statusTimer)
                this.statusTimer = setTimeout(async () => {
                    if (this.isOperaLayerShow) {
                        return false
                    }
                    try {
                        const data = this.getParams()
                        const res = await this.$store.dispatch('helm/getAppList', data)
                        const count = res.data.results.length
                        const loading = this.appList.length !== count
                        this.appList = res.data.results
                        this.pagination.count = count
                        this.appListCache = JSON.parse(JSON.stringify(res.data.results))
                        // 轮询接口,保持选中状态
                        this.appList.forEach(appItem => {
                            this.selectLists.forEach(selectAppItem => {
                                if (appItem.id === selectAppItem.id) {
                                    this.$set(appItem, 'isChecked', true)
                                }
                            })
                        })
                        this.curPageData = this.getDataByPage(this.pagination.current, loading)

                        this.appCheckTime = SLOW_TIME
                        this.appList.forEach(app => {
                            if (app.transitioning_on) {
                                this.appCheckTime = FAST_TIME // 如果有更新中的app，继续快速轮询
                            }
                        })

                        // 按原关键字再搜索
                        if (this.searchKeyword) {
                            this.search()
                        }

                        this.diffAppStatus()
                        if (!this.timeOutFlag) {
                            this.getAppsStatus()
                        } else {
                            clearTimeout(this.statusTimer)
                        }
                    } catch (e) {
                        catchErrorHandler(e, this)
                    } finally {
                        this.showLoading = false
                    }
                }, this.appCheckTime)
            },

            /**
             * 遍历appList 获取应用状态
             * @param {number} appId appId
             * @return {object} app状态数据
             */
            getAppStatusById (appId) {
                let result = null
                this.appList.forEach(item => {
                    if (String(item.id) === appId) {
                        result = {
                            name: item.name,
                            transitioning_on: item.transitioning_on,
                            transitioning_action: item.transitioning_action,
                            transitioning_result: item.transitioning_result,
                            transitioning_message: item.transitioning_message
                        }
                    }
                })
                return result
            },

            /**
             * 对比各个应用发生变化的状态
             */
            diffAppStatus () {
                const continueRunningApps = {}

                this.appList.forEach(app => {
                    if (app.transitioning_on) {
                        continueRunningApps[app.id] = {
                            name: app.name,
                            transitioning_on: app.transitioning_on,
                            transitioning_action: app.transitioning_action,
                            transitioning_result: app.transitioning_result,
                            transitioning_message: app.transitioning_message
                        }
                    }
                })

                for (const appId in this.operaRunningApp) {
                    const appStatus = this.getAppStatusById(appId)

                    // 和上次对比，如果应用不存在，则已经删除成功
                    if (!appStatus) {
                        const app = this.operaRunningApp[appId]
                        this.$bkMessage({
                            theme: 'success',
                            message: `${app.name}${this.$t('删除成功')}`
                        })
                        delete this.operaRunningApp[appId]
                        return true
                    }

                    // 如果操作状态结束
                    if (!appStatus.transitioning_on) {
                        const action = this.appAction[appStatus.transitioning_action]

                        if (appStatus.transitioning_result) {
                            this.$bkMessage({
                                theme: 'success',
                                message: `${appStatus.name}${action}${this.$t('成功')}`
                            })
                        } else {
                            this.$bkMessage({
                                theme: 'error',
                                message: `${appStatus.name}${action}${this.$t('失败')}`
                            })
                        }
                        delete this.operaRunningApp[appId]
                    }
                }

                this.operaRunningApp = continueRunningApps // 保存当前更新中的应用
            },

            /**
             * 显示回滚相应的预览对比列表
             * @param  {object} app 应用
             */
            async showRebackPreview (app) {
                const projectId = this.projectId
                const appId = this.curApp.id
                const params = {
                    release: this.versionId
                }

                this.isRebackVersionLoading = true
                this.difference = ''
                this.rebackPreviewList = []

                try {
                    const res = await this.$store.dispatch('helm/previewReback', {
                        projectId,
                        appId,
                        params
                    })

                    this.rebackEditorConfig.value = res.data.notes

                    for (const key in res.data.content) {
                        this.rebackPreviewList.push({
                            name: key,
                            value: res.data.content[key]
                        })
                    }
                    if (res.data.difference) {
                        this.difference = res.data.difference
                    } else {
                        this.difference = this.$t('与当前线上版本没有内容差异')
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isRebackVersionLoading = false
                }
            },

            /**
             *  获取应用
             * @param  {number} appId 应用ID
             */
            async getAppById (appId) {
                let result = {}
                const projectId = this.projectId

                this.previewLoading = true
                try {
                    const res = await this.$store.dispatch('helm/getAppById', { projectId, appId })
                    result = res.data
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.previewLoading = false
                }

                return result
            },

            /**
             * 取消回滚操作
             */
            cancelReback () {
                this.appCheckTime = FAST_TIME
                this.rebackDialogConf.isShow = false
                this.isOperaLayerShow = false
                this.getAppsStatus()
            },

            /**
             * 隐藏应用详情面板回调
             */
            hideAppInfoSlider () {
                this.isOperaLayerShow = false
                this.appCheckTime = FAST_TIME
                this.getAppsStatus()
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            handlePageLimitChange (pageSize) {
                this.appList.forEach(item => {
                    item.isChecked = false
                })
                this.pagination.limit = pageSize
                this.pagination.current = 1
                this.handlePageChange(this.pagination.current)
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            handlePageChange (page) {
                this.isCheckAll = false
                this.pagination.current = page
                this.curPageData = this.getDataByPage(page)
                this.pagination.count = this.appList.length
            },

            /**
             * 获取分页数据
             * @param  {number} page 第几页
             * @return {object} data 数据
             */
            getDataByPage (page, loading = true) {
                let startIndex = (page - 1) * this.pagination.limit
                let endIndex = page * this.pagination.limit
                this.isPageLoading = loading
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.appList.length) {
                    endIndex = this.appList.length
                }
                setTimeout(() => {
                    this.isPageLoading = false
                }, 200)
                this.selectLists = []
                return this.appList.slice(startIndex, endIndex)
            }

            // /**
            //  * 自定义checkbox表格头
            //  */
            // renderSelectionHeader () {
            //     return <bk-checkbox name="check-all-strategy" v-model={this.isCheckAll} onChange={this.checkAllApp} />
            // },

            // /**
            //  * 列表每一行的 checkbox 点击
            //  *
            //  * @param {Object} row 当前对象
            //  */
            // checkApp (row) {
            //     this.selectLists = this.curPageData.filter(item => item.isChecked === true)
            //     this.isCheckAll = this.selectLists.length === this.curPageData.length
            // }

            // /**
            //  * 列表全选
            //  */
            // checkAllApp (value) {
            //     const isChecked = value
            //     this.curPageData.forEach(item => {
            //         this.$set(item, 'isChecked', isChecked)
            //     })
            //     this.selectLists = isChecked ? this.curPageData : []
            // }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
