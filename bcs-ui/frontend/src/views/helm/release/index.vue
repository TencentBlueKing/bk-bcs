<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-helm-title">
                Release
            </div>
            <bk-guide>
                <a class="bk-text-button" :href="PROJECT_CONFIG.doc.serviceAccess" target="_blank">{{$t('如何使用Helm？')}}</a>
            </bk-guide>
        </div>
        <div class="biz-content-wrapper">
            <div class="biz-panel-header p20">
                <div class="left">
                    <bcs-button @click="handleBatchDelete">{{ $t('批量删除') }}</bcs-button>
                    <span v-if="selections.length">
                        <span class="selected-num">
                            <i18n path="已选择 {num} 项">
                                <span place="num" class="tips-num">{{ selections.length }}</span>
                            </i18n>
                        </span>
                        <span class="clear-btn" @click="handleClearSelection()">{{ $t('清空') }}</span>
                    </span>
                </div>
                <div class="right">
                    <search
                        :width-refresh="false"
                        :scope-list="searchScopeList"
                        :namespace-list="namespacesList"
                        :search-key.sync="searchKeyword"
                        :search-scope.sync="searchScope"
                        :search-namespace.sync="searchNamespace"
                        :cluster-fixed="!!curClusterId"
                        @cluster-change="handleClusterChange"
                        @search="handleSearch"
                        @refresh="handleRefresh">
                    </search>
                </div>

                <section>
                    <bcs-table
                        ref="releaseTable"
                        ext-cls="release-table"
                        :data="appList"
                        @selection-change="handleSelectionChange"
                        :pagination="pagination"
                        @page-change="pageChange"
                        @page-limit-change="pageSizeChange"
                        v-bkloading="{ isLoading: isPageLoading }">
                        <bcs-table-column type="selection"></bcs-table-column>
                        <bk-table-column :label="$t('名称')" prop="name" width="250">
                            <template slot-scope="{ row }">
                                <span v-if="row.transitioning_on" class="f14 app-name">
                                    {{ row.name }}
                                </span>
                                <a v-else
                                    @click="showAppStatus(row)"
                                    href="javascript:void(0)"
                                    class="bk-text-button app-name f14"
                                    
                                >
                                    {{ row.name }}
                                </a>
                                <!-- v-authority="{
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
                                    }" -->
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('状态')" prop="id" width="120">
                            <template slot-scope="{ row }">
                                <template v-if="row.transitioning_on">
                                    <div>
                                        <div class="bk-spin-loading bk-spin-loading-mini bk-spin-loading-warning mr5 status-loading">
                                            <div class="rotate rotate1"></div>
                                            <div class="rotate rotate2"></div>
                                            <div class="rotate rotate3"></div>
                                            <div class="rotate rotate4"></div>
                                            <div class="rotate rotate5"></div>
                                            <div class="rotate rotate6"></div>
                                            <div class="rotate rotate7"></div>
                                            <div class="rotate rotate8"></div>
                                        </div>
                                        {{ appAction[row.transitioning_action] }}中...
                                    </div>
                                </template>
                                <template v-else>
                                    <div>
                                        <i :class="['status-icon', row.status === 'failed' ? 'error-icon' : 'normal-icon']"></i>
                                        <span>{{ row.status === 'failed' ? $t('失败') : $t('正常') }}</span>
                                    </div>
                                </template>
                            </template>
                        </bk-table-column>
                        <bk-table-column label="chart" prop="chart_name" width="250">
                            <template slot-scope="{ row }">
                                {{ `${row.name}:${row.chartVersion}` }}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('集群')" prop="cluster_name"></bk-table-column>
                        <bk-table-column :label="$t('命名空间')" prop="namespace"></bk-table-column>
                        <bk-table-column :label="$t('操作者')" prop="updator"></bk-table-column>
                        <bk-table-column :label="$t('更新时间')" prop="updated"></bk-table-column>
                        <bk-table-column :label="$t('操作')" width="200">
                            <template slot-scope="{ row }">
                                <!-- v-authority="{
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
                                }" -->
                                <bk-button class="ml5"
                                    text
                                    @click="showUpdateApp(row)"
                                >{{ $t('升级') }}</bk-button>
                                <!-- v-authority="{
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
                                }" -->
                                <bk-button class="ml5"
                                    text
                                    @click="showRebackDialog(row)"
                                >{{ $t('回滚') }}</bk-button>
                                <!-- v-authority="{
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
                                    }" -->
                                <bk-button class="ml5"
                                    text
                                    @click="showDeleteDialog(row)"
                                >{{ $t('删除') }}</bk-button>
                            </template>
                        </bk-table-column>
                    </bcs-table>
                </section>
            </div>
        </div>
        
        <!-- 回滚弹框 -->
        <bk-dialog
            width="800"
            :title="rebackDialogConf.title"
            :close-icon="!isRebackLoading"
            :quick-close="false"
            key="rebackDialog"
            :is-show.sync="rebackDialogConf.isShow">
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

        <!-- 删除Release弹框 -->
        <bk-dialog
            width="480"
            :title="deleteDialogConf.title"
            :quick-close="false"
            key="deleteDialog"
            :is-show.sync="deleteDialogConf.isShow"
            @cancel="cancelDelete"
            @confirm="deleteApp">
            <template slot="content">
                {{ $t('确认要删除')}} {{ curApp.name }}
            </template>
        </bk-dialog>
    </div>
</template>

<script>
    import { onMounted, watch, computed, ref, reactive, onBeforeUnmount } from '@vue/composition-api'
    import search from './search.vue'
    import ace from '@/components/ace-editor'

    const FAST_TIME = 3000
    const SLOW_TIME = 10000

    export default {
        name: 'Release',
        components: {
            ace,
            search
        },
        setup (props, ctx) {
            const { $i18n, $router, $bkMessage, $store, $route } = ctx.root
            const releaseTable = ref(null)
            const appList = ref([])
            const namespacesList = ref([])
            const webAnnotationsPerms = ref({})
            const searchNamespace = ref('')
            const searchKeyword = ref('')
            const searchScope = ref('')
            const appListCache = ref([])
            const isPageLoading = ref(false)
            const showLoading = ref(false)
            const statusTimer = ref(null)
            const appCheckTime = ref(null)
            const isRebackLoading = ref(false)
            const timeOutFlag = ref(false)
            const operaRunningApp = ref({}) // 缓存操作更新中的app状态信息
            const curApp = ref({})
            const difference = ref('')
            const versionId = ref('')
            const isRebackVersionLoading = ref(false)
            const rebackList = ref([])
            const isOperaLayerShow = ref(false)
            const rebackPreviewList = ref([])
            const selections = ref([])
            const pagination = reactive({
                current: 1,
                limit: 10,
                count: 0
            })
            const rebackDialogConf = reactive({
                title: '',
                isShow: false
            })
            const deleteDialogConf = reactive({
                title: $i18n.t('删除'),
                isShow: false
            })
            const isRebackListLoading = ref(false)
            const rebackEditorConfig = reactive({
                width: '100%',
                height: '100%',
                lang: 'yaml',
                readOnly: true,
                fullScreen: false,
                value: '',
                editors: []
            })
            const appAction = reactive({
                create: $i18n.t('部署'),
                noop: '',
                update: $i18n.t('更新'),
                rollback: $i18n.t('回滚'),
                delete: $i18n.t('删除'),
                destroy: $i18n.t('删除')
            })

            const searchScopeList = computed(() => {
                const clusterList = $store.state.cluster.clusterList
                return clusterList.map(item => {
                    return {
                        id: item.cluster_id,
                        name: item.name
                    }
                })
            })
            const projectId = computed(() => {
                return $route.params.projectId
            })
            const projectCode = computed(() => {
                return $route.params.projectCode
            })
            const curProjectId = computed(() => {
                return $route.params.projectId
            })
            const curProject = computed(() => {
                return $store.state.curProject
            })
            const curClusterId = computed(() => {
                return $store.state.curClusterId
            })

            watch(curClusterId, (val) => {
                searchScope.value = val
                searchNamespace.value = ''
                handleSearch()
            })

            watch(curProjectId, () => {
                // 如果不是k8s类型的项目，无法访问些页面，重定向回集群首页
                if (curProject && (curProject.kind !== PROJECT_K8S && curProject.kind !== PROJECT_TKE)) {
                    $router.push({
                        name: 'clusterMain',
                        params: {
                            projectId: projectId.value,
                            projectCode: projectCode.value
                        }
                    })
                }
            })

            /**
             * 获取所有命名空间列表
             */
            const getNamespacesList = async () => {
                const res = await $store.dispatch('helm/getNamespaceList', {
                    projectId: projectId.value,
                    params: {
                        cluster_id: searchScope.value
                    }
                }).catch(() => false)

                if (!res) return
                namespacesList.value = (res.data || []).map(item => {
                    return {
                        ...item,
                        namespace_id: `${item.cluster_id}:${item.name}`
                    }
                })
            }

            const handleSearch = () => {
                window.sessionStorage['bcs-cluster'] = searchScope.value
                window.sessionStorage['bcs-helm-namespace'] = searchNamespace.value
                getAppList()
                handleClearSelection()
            }

            const getAppList = async () => {
                isPageLoading.value = true
                clearTimeout(statusTimer.value)

                const data = getParams()
                const res = await $store.dispatch('helm/getAppList', data).catch(() => false)
                showLoading.value = false
                isPageLoading.value = false

                if (!res) return
                appList.value = res.data
                pagination.count = res.total
                webAnnotationsPerms.value = Object.assign(webAnnotationsPerms.value, res.web_annotations?.perms || {})
                appListCache.value = JSON.parse(JSON.stringify(res.data))

                getAppsStatus()

                // 按原关键字再搜索
                if (searchKeyword.value) {
                    search()
                }
            }

            const getAppsStatus = () => {
                clearTimeout(statusTimer.value)
                statusTimer.value = setTimeout(async () => {
                    if (isOperaLayerShow.value) {
                        return false
                    }
                    const data = getParams()
                    const res = await $store.dispatch('helm/getAppList', data).catch(() => false)
                    showLoading.value = false

                    if (!res) return

                    appList.value = res.data
                    appListCache.value = JSON.parse(JSON.stringify(res.data))

                    appCheckTime.value = SLOW_TIME
                    appList.value.forEach(app => {
                        if (app.transitioning_on) {
                            appCheckTime.value = FAST_TIME // 如果有更新中的app，继续快速轮询
                        }
                    })

                    // 按原关键字再搜索
                    if (searchKeyword.value) {
                        search()
                    }

                    diffAppStatus()
                    if (!timeOutFlag.value) {
                        getAppsStatus()
                    } else {
                        clearTimeout(statusTimer.value)
                    }
                }, appCheckTime.value)
            }

            const diffAppStatus = () => {
                const continueRunningApps = {}

                appList.value.forEach(app => {
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

                for (const appId in operaRunningApp.value) {
                    const appStatus = getAppStatusById(appId)

                    // 和上次对比，如果应用不存在，则已经删除成功
                    if (!appStatus) {
                        const app = operaRunningApp.value[appId]
                        $bkMessage({
                            theme: 'success',
                            message: `${app.name}${$i18n.t('删除成功')}`
                        })
                        delete operaRunningApp.value[appId]
                        return true
                    }

                    // 如果操作状态结束
                    if (!appStatus.transitioning_on) {
                        const action = appAction[appStatus.transitioning_action]

                        if (appStatus.transitioning_result) {
                            $bkMessage({
                                theme: 'success',
                                message: `${appStatus.name}${action}${$i18n.t('成功')}`
                            })
                        } else {
                            $bkMessage({
                                theme: 'error',
                                message: `${appStatus.name}${action}${$i18n.t('失败')}`
                            })
                        }
                        delete operaRunningApp.value[appId]
                    }
                }

                operaRunningApp.value = continueRunningApps // 保存当前更新中的应用
            }

            /**
             * 遍历appList 获取应用状态
             * @param {number} appId appId
             * @return {object} app状态数据
             */
            const getAppStatusById = (appId) => {
                let result = null
                appList.value.forEach(item => {
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
            }

            const getParams = () => {
                const data = {
                    $clusterId: searchScope.value,
                    size: pagination.limit,
                    page: pagination.current - 1,
                    namespace: ''
                }
                if (searchNamespace.value) {
                    const args = searchNamespace.value.split(':')
                    data.$namespace = args[1]
                }
                return data
            }

            /**
             * 搜索Helm app
             */
            const search = () => {
                const keyword = searchKeyword.value
                const keyList = ['name', 'namespace', 'cluster_name']
                const list = JSON.parse(JSON.stringify(appListCache.value))
                if (keyword) {
                    const results = list.filter(item => {
                        for (const key of keyList) {
                            if (item[key] && item[key].indexOf(keyword) > -1) {
                                return true
                            }
                        }
                        return false
                    })
                    appList.value.splice(0, appList.value.length, ...results)
                } else {
                    // 没有搜索关键字，直接从缓存返回列表
                    appList.value.splice(0, appList.value.length, ...list)
                }
            }

            /**
             * 批量删除Release app
             */
            const handleBatchDelete = () => {
            }

            const handleSelectionChange = (selection) => {
                selections.value = selection
            }

            const pageChange = (page) => {
                pagination.current = page
                getAppList()
            }
            const pageSizeChange = (size) => {
                pagination.current = 1
                pagination.limit = size
                getAppList()
            }

            /**
             * 清空已选app
             */
            const handleClearSelection = () => {
                releaseTable.value.clearSelection()
            }

            const handleClusterChange = () => {
                searchNamespace.value = ''
                getNamespacesList()
            }

            /**
             * 显示回滚弹层
             * @param  {object} app 应用
             */
            const showRebackDialog = async (app) => {
                clearTimeout(statusTimer.value)

                curApp.value = app
                versionId.value = ''
                isRebackVersionLoading.value = false
                rebackDialogConf.isShow = true
                rebackDialogConf.title = `${$i18n.t('回滚')} ${app.name}`
                isOperaLayerShow.value = true
                rebackPreviewList.value = []
                rebackList.value = []
                isRebackListLoading.value = true
                getRebackList(app.id)
            }

            /**
             * 获取回滚版本列表
             * @param  {number} appId 应用ID
             */
            const getRebackList = async (appId) => {
                isRebackListLoading.value = true

                const res = await $store.dispatch('helm/getRebackList', { projectId: projectId.value, appId }).catch(() => false)
                isRebackListLoading.value = false

                if (!res) {
                    rebackList.value = []
                    return
                }

                res.data.results.forEach(item => {
                    item.version = `${$i18n.t('版本')}：${item.version} （${$i18n.t('部署时间')}：${item.created_at}） `
                })
                rebackList.value = res.data.results
            }

            /**
             * 取消回滚操作
             */
            const cancelReback = () => {
                appCheckTime.value = FAST_TIME
                rebackDialogConf.isShow = false
                isOperaLayerShow.value = false
                getAppsStatus()
            }
            
            /**
             * 显示回滚相应的预览对比列表
             */
            const showRebackPreview = async () => {
                isRebackVersionLoading.value = true
                difference.value = ''
                rebackPreviewList.value = []

                const res = await $store.dispatch('helm/previewReback', {
                    projectId: projectId.value,
                    appId: curApp.value.id,
                    params: {
                        release: versionId.value
                    }
                }).catch(() => false)
                isRebackVersionLoading.value = false
                
                if (!res) return

                rebackEditorConfig.value = res.data.notes

                for (const key in res.data.content) {
                    rebackPreviewList.value.push({
                        name: key,
                        value: res.data.content[key]
                    })
                }
                if (res.data.difference) {
                    difference.value = res.data.difference
                } else {
                    difference.value = $i18n.t('与当前线上版本没有内容差异')
                }
            }

            /**
             * 确认删除应用
             * @param {object} app 应用
             */
            const deleteApp = async () => {
                curApp.value.transitioning_action = 'delete'
                curApp.value.transitioning_on = true

                try {
                    await $store.dispatch('helm/deleteApp', {
                        projectId: projectId.value,
                        appId: curApp.value.id
                    }, {
                        cancelPrevious: true
                    })
                    checkingAppStatus(curApp.value, 'delete')
                } catch (e) {
                    console.log(e)
                } finally {
                    isOperaLayerShow.value = false
                    showLoading.value = false
                }
            }

            /**
             * 显示删除弹层
             * @param  {object} app 应用
             */
            const showDeleteDialog = (app) => {
                curApp.value = app
                clearTimeout(statusTimer.value)
                deleteDialogConf.isShow = true
                isOperaLayerShow.value = true
            }

            /**
             * 提交回滚数据
             * @return {[type]} [description]
             */
            const submitRebackData = async () => {
                if (isRebackLoading.value || !versionId.value) {
                    return false
                } else {
                    isRebackLoading.value = true
                }

                try {
                    await $store.dispatch('helm/reback', {
                        $clusterId: curApp.value.cluster_id,
                        $namespace: curApp.value.namespace,
                        $name: curApp.value.name
                    })
                    rebackDialogConf.isShow = false
                    isOperaLayerShow.value = false
                    checkingAppStatus(curApp.value, 'rollback')
                } catch (e) {
                    console.log(e)
                } finally {
                    isRebackLoading.value = false
                }
            }

            /**
             * 取消删除操作
             */
            const cancelDelete = () => {
                appCheckTime.value = FAST_TIME
                isOperaLayerShow.value = false
                deleteDialogConf.isShow = false
                curApp.value = {}
                getAppsStatus()
            }

            /**
             * 查询应用的状态
             * @param  {object} app 应用对象
             */
            const checkingAppStatus = (app, action) => {
                if (app) {
                    const status = {
                        name: app.name,
                        transitioning_on: true,
                        transitioning_action: action,
                        transitioning_result: false,
                        transitioning_message: ''
                    }
                    operaRunningApp.value[app.id] = status
                    updateApp(app.id, status)
                }

                appCheckTime.value = FAST_TIME
                getAppsStatus()
            }

            const updateApp = (appId, status) => {
                appList.value.forEach((app) => {
                    if (app.id === appId) {
                        app.transitioning_action = status.transitioning_action
                        app.transitioning_message = status.transitioning_message
                        app.transitioning_on = status.transitioning_on
                        app.transitioning_result = status.transitioning_result
                    }
                })
                appListCache.value.forEach((app) => {
                    if (app.id === appId) {
                        app.transitioning_action = status.transitioning_action
                        app.transitioning_message = status.transitioning_message
                        app.transitioning_on = status.transitioning_on
                        app.transitioning_result = status.transitioning_result
                    }
                })
            }

            /**
             * 升级应用页面
             * @param {object} app 应用
             */
            const showUpdateApp = (app) => {
                $router.push({
                    name: 'helmUpdateApp',
                    params: {
                        projectId: projectId.value,
                        projectCode: projectCode.value,
                        clusterId: searchScope.value,
                        namespace: app.namespace,
                        name: app.name
                    }
                })
            }

            /**
             * 显示应用详情
             * @param {object} app 应用
             */
            const showAppStatus = (app) => {
                $router.push({
                    name: 'helmAppStatus',
                    params: {
                        clusterId: searchScope.value,
                        namespace: app.namespace,
                        name: app.name
                    }
                })
            }

            onMounted(() => {
                searchScope.value = searchScopeList.value[0]?.id
                
                if (window.sessionStorage && window.sessionStorage['bcs-cluster']) {
                    searchScope.value = window.sessionStorage['bcs-cluster']
                }
                if (window.sessionStorage && window.sessionStorage['bcs-helm-namespace']) {
                    searchNamespace.value = window.sessionStorage['bcs-helm-namespace']
                }
                getAppList()
                getNamespacesList()
            })

            onBeforeUnmount(() => {
                timeOutFlag.value = true
                clearTimeout(statusTimer.value)
                // window.sessionStorage['bcs-helm-cluster'] = ''
                window.sessionStorage['bcs-helm-namespace'] = ''
            })

            return {
                releaseTable,
                appList,
                selections,
                pagination,
                searchScopeList,
                namespacesList,
                searchScope,
                searchNamespace,
                searchKeyword,
                curClusterId,
                webAnnotationsPerms,
                isPageLoading,
                appAction,
                rebackDialogConf,
                curApp,
                versionId,
                isRebackVersionLoading,
                rebackPreviewList,
                rebackList,
                isRebackListLoading,
                rebackEditorConfig,
                deleteDialogConf,
                difference,
                handleClearSelection,
                handleSelectionChange,
                pageChange,
                pageSizeChange,
                handleSearch,
                showDeleteDialog,
                showRebackPreview,
                cancelDelete,
                cancelReback,
                deleteApp,
                handleBatchDelete,
                handleClusterChange,
                showRebackDialog,
                showUpdateApp,
                showAppStatus,
                submitRebackData
            }
        }
    }
</script>

<style lang="postcss" scoped>
    @import '@/css/variable.css';
    @import '@/css/mixins/ellipsis.css';
    .biz-helm-title {
        display: inline-block;
        height: 60px;
        line-height: 60px;
        font-size: 16px;
        margin-left: 20px;
    }

    .release-table {
        margin-top: 60px;
        /deep/ .bk-page-selection-count-left {
            display: none;
        }
    }

    .selected-num {
        font-size: 12px;
        padding: 0 12px;
        .tips-num {
            font-weight: 700;
        }
    }
    .status-icon {
        display: inline-block;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        margin-right: 5px;
    }
    .error-icon {
        background: #fd9c9c;
        border: 1px solid #ea3636;
    }

    .normal-icon {
        background: #94f5a4;
        border: 1px solid #2dcb56;
    }
    .status-loading {
        width: 10px !important;
        height: 10px !important;
    }
    .clear-btn {
        font-size: 12px;
        color: #3a84ff;
        cursor: pointer;
    }
    .delete-release-content {
        font-size: 16px;
        margin-bottom: 20px;
    }
</style>
