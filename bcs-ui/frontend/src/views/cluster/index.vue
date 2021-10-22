<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-cluster-title">
                {{$t('集群')}}
                <span class="cc-info">
                    （{{$t('业务')}}: {{curProject.cc_app_name}}&nbsp;&nbsp;&nbsp;{{$t('编排类型')}}: {{kindMap[curProject.kind]}}）
                </span>
                <span class="bk-text-button bk-default f12" @click="handleShowProjectConf">
                    <i class="bcs-icon bcs-icon-edit"></i>
                </span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper biz-cluster-wrapper" v-bkloading="{ isLoading, color: '#fafbfd' }">
            <template v-if="clusterList.length">
                <div class="cluster-btns">
                    <bk-button theme="primary" icon="plus" @click="goCreateCluster">{{$t('新建集群')}}</bk-button>
                    <apply-host class="ml10" v-if="$INTERNAL" />
                </div>
                <!-- 集群面板 -->
                <div class="biz-cluster-list">
                    <div class="biz-cluster" v-for="cluster in clusterList" :key="cluster.cluster_id">
                        <!-- 异常角标 -->
                        <div class="bk-mark-corner bk-warning" v-if="showCorner(cluster)"><p>!</p></div>
                        <!-- 集群信息 -->
                        <div class="biz-cluster-header">
                            <h2 :class="['cluster-title', { clickable: cluster.status === 'normal' }]"
                                v-bk-tooltips.top="{ content: cluster.name, delay: 500 }" @click="goOverview(cluster)">
                                {{ cluster.name }}
                            </h2>
                            <p class="cluster-metadata">
                                <span class="cluster-id" v-bk-tooltips.top="{ content: cluster.cluster_id, delay: 500 }">
                                    {{cluster.cluster_id}}
                                </span>
                                <template v-if="$INTERNAL">
                                    <span v-if="cluster.environment === 'stag'" class="stag">
                                        {{$t('测试')}}
                                    </span>
                                    <span v-else-if="cluster.environment === 'prod'" class="prod">
                                        {{$t('正式')}}
                                    </span>
                                </template>
                                <span v-if="cluster.state === 'existing'" class="prod">{{$t('自有集群')}}</span>
                            </p>
                            <!-- 集群操作菜单 -->
                            <bk-dropdown-menu v-if="cluster.status === 'normal'">
                                <bk-button class="cluster-opera-btn" slot="dropdown-trigger">
                                    <i class="bcs-icon bcs-icon-more"></i>
                                </bk-button>
                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                    <li @click="goOverview(cluster)"><a href="javascript:;">{{$t('总览')}}</a></li>
                                    <li @click="goClusterInfo(cluster)"><a href="javascript:;">{{$t('集群信息')}}</a></li>
                                    <li v-if="cluster.type === 'k8s' && $INTERNAL" @click="handleUpdateCluster(cluster)">
                                        <a href="javascript:;">{{$t('集群升级')}}</a>
                                    </li>
                                    <li :class="{ disabled: !cluster.allow }"
                                        v-bk-tooltips="{
                                            content: $t('您需要删除集群内所有节点后，再进行集群删除操作'),
                                            placement: 'right',
                                            boundary: window,
                                            interactive: false,
                                            disabled: cluster.allow
                                        }" @click="handleDeleteCluster(cluster)">
                                        <a href="javascript:;">{{$t('删除')}}</a>
                                    </li>
                                    <li v-if="!cluster.permissions.use">
                                        <a :href="createApplyPermUrl({
                                            policy: 'use',
                                            projectCode: projectCode,
                                            idx: `cluster_${cluster.environment === 'stag' ? 'test' : 'prod'}:${cluster.cluster_id}`
                                        })" target="_blank">{{$t('申请使用权限')}}</a>
                                    </li>
                                </ul>
                            </bk-dropdown-menu>
                            <bk-dropdown-menu v-else-if="cluster.allow">
                                <bk-button class="cluster-opera-btn" slot="dropdown-trigger">
                                    <i class="bcs-icon bcs-icon-more"></i>
                                </bk-button>
                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                    <li @click="handleDeleteCluster(cluster)">
                                        <a href="javascript:;">{{$t('删除')}}</a>
                                    </li>
                                </ul>
                            </bk-dropdown-menu>
                        </div>
                        <!-- 集群状态 -->
                        <div class="biz-cluster-content">
                            <!-- 进行中（移除中、升级中、初始化中） -->
                            <div class="biz-status-box" v-if="['removing', 'upgrading', 'initializing', 'so_initializing', 'initial_checking'].includes(cluster.status)">
                                <div class="status-icon" v-bkloading="{ isLoading: true, theme: 'primary', mode: 'spin' }"></div>
                                <p class="status-text">{{ statusTextMap[cluster.status] }}</p>
                                <div class="status-opera">
                                    <bk-button text @click="handleShowLog(cluster)">{{$t('查看日志')}}</bk-button>
                                </div>
                            </div>
                            <!-- 失败（升级失败、删除失败、初始化失败） -->
                            <div class="biz-status-box" v-else-if="['upgrade_failed', 'remove_failed', 'initial_failed', 'so_init_failed', 'check_failed'].includes(cluster.status)">
                                <div class="status-icon danger">
                                    <i class="bcs-icon bcs-icon-close-circle"></i>
                                </div>
                                <p class="status-text">{{ statusTextMap[cluster.status] }}</p>
                                <div class="status-opera">
                                    <bk-button text @click="handleShowLog(cluster)">{{$t('查看日志')}}</bk-button> |
                                    <bk-button text @click="handleRedo(cluster)">{{ redoTextMap[cluster.status] }}</bk-button>
                                </div>
                            </div>
                            <!-- 正常 -->
                            <template v-else>
                                <!-- 指标信息 -->
                                <div class="biz-progress-box" v-for="item in clusterMetricList" :key="item.id">
                                    <div class="progress-header">
                                        <span class="title">{{item.title}}</span>
                                        <span class="percent" v-if="clusterOverviewMap[cluster.cluster_id]">
                                            {{getMetricPercent(cluster, item)}}%
                                        </span>
                                    </div>
                                    <div class="progress" :class="!clusterOverviewMap[cluster.cluster_id] ? 'loading' : ''">
                                        <div :class="['progress-bar', item.theme]"
                                            :style="{ width: !clusterOverviewMap[cluster.cluster_id] ? '0%' : `${getMetricPercent(cluster, item)}%` }"></div>
                                    </div>
                                </div>
                                <bk-button class="add-node-btn" @click="goOverview(cluster)">
                                    <span>{{$t('添加节点')}}</span>
                                </bk-button>
                            </template>
                        </div>
                    </div>
                    <div class="biz-cluster biz-cluster-add" @click="goCreateCluster">
                        <div class="add-btn">
                            <i class="bcs-icon bcs-icon-plus"></i>
                            <strong>{{$t('点击新建集群')}}</strong>
                        </div>
                    </div>
                </div>
            </template>
            <!-- 集群创建引导 -->
            <template v-else-if="!isLoading">
                <div class="biz-guide-box">
                    <p class="title">{{$t('欢迎使用容器服务')}}</p>
                    <p class="desc">{{$t('使用容器服务，蓝鲸将为您快速搭建、运维和管理容器集群，您可以轻松对容器进行启动、停止等操作，也可以查看集群、容器及服务的状态，以及使用各种组件服务。')}}</p>
                    <p class="desc">
                        <a :href="PROJECT_CONFIG.doc.quickStart" class="guide-link" target="_blank">{{$t('请点击了解更多')}}<i class="bcs-icon bcs-icon-angle-double-right ml5"></i></a>
                    </p>
                    <div class="guide-btn-group">
                        <a href="javascript:void(0);" class="bk-button bk-primary bk-button-large" @click="goCreateCluster">
                            <span style="margin-left: 0;">{{$t('创建容器集群')}}</span>
                        </a>

                        <a class="bk-button bk-default bk-button-large" :href="PROJECT_CONFIG.doc.quickStart" target="_blank">
                            <span style="margin-left: 0;">{{$t('快速入门指引')}}</span>
                        </a>
                        <apply-host class="apply-host ml5" v-if="$INTERNAL" />
                    </div>
                </div>
            </template>
        </div>
        <!-- 集群日志（保留的旧逻辑） -->
        <bk-sideslider
            :is-show.sync="showLogDialog"
            :title="curOperateCluster && curOperateCluster.cluster_id"
            :quick-close="true"
            @hidden="handleCloseLog">
            <div slot="content" style="margin: 0 0 0 20px;">
                <template v-if="logEndState === 'none'">
                    <div class="biz-no-data">
                        {{$t('暂无日志信息')}}
                    </div>
                </template>
                <template v-else>
                    <div class="biz-log-box">
                        <template v-if="logList && logList.length">
                            <div class="operation-item">
                                <p class="log-message title">
                                    {{logList[0].prefix_message}}
                                </p>
                                <div class="log-message item" v-for="(task, taskIndex) in logList[0].log.node_tasks" :key="taskIndex">
                                    {{task.name}} -
                                    <span v-if="task.state.toLowerCase() === 'failure'" class="biz-danger-text">
                                        {{task.state}}
                                    </span>
                                    <span v-else-if="task.state.toLowerCase() === 'success'" class="biz-success-text">
                                        {{task.state}}
                                    </span>
                                    <span v-else-if="task.state.toLowerCase() === 'running'" class="biz-warning-text">
                                        {{task.state}}
                                    </span>
                                    <div v-else-if="task.state.indexOf('html-tag') > -1" v-html="task.state">
                                    </div>
                                    <span v-else>
                                        {{task.state}}
                                    </span>
                                </div>
                                <div v-if="logList[0].status.toLowerCase() === 'success'" class="biz-success-text f14" style="margin: 0 0 5px 0; font-weight: 700; margin-left: 20px;">
                                    {{$t('操作成功')}}
                                </div>
                                <div v-else-if="logList[0].status.toLowerCase() === 'failed'" class="biz-danger-text f14" style="margin: 0 0 5px 0; font-weight: 700; margin-left: 20px;">
                                    {{$t('操作失败')}}
                                    <template v-if="logList[0].errorMsgList && logList[0].errorMsgList.length">
                                        <span>{{logList[0].errorMsgList[0]}}</span>
                                        <template v-for="(msg, msgIndex) in logList[0].errorMsgList">
                                            <div :key="msgIndex" v-if="msgIndex > 0">{{msg}}</div>
                                        </template>
                                    </template>
                                    <template v-else>
                                        <i18n path="请联系“{user}”解决">
                                            <a place="user" :href="PROJECT_CONFIG.doc.contact" class="bk-text-button">{{$t('蓝鲸容器助手')}}</a>
                                        </i18n>
                                    </template>
                                </div>
                                <div style="margin: 10px 0px 5px 13px; font-size: 10px;" v-else>
                                    <div class="bk-spin-loading bk-spin-loading-small bk-spin-loading-primary">
                                        <div class="rotate rotate1"></div>
                                        <div class="rotate rotate2"></div>
                                        <div class="rotate rotate3"></div>
                                        <div class="rotate rotate4"></div>
                                        <div class="rotate rotate5"></div>
                                        <div class="rotate rotate6"></div>
                                        <div class="rotate rotate7"></div>
                                        <div class="rotate rotate8"></div>
                                    </div>
                                    {{$t('正在加载中...')}}
                                </div>
                            </div>
                        </template>

                        <template v-for="(op, index) in logList">
                            <div class="operation-item" :key="index" v-if="index > 0">
                                <p class="log-message title">
                                    {{op.prefix_message}}
                                </p>
                                <div class="log-message item" v-for="(task, taskIndex) in op.log.node_tasks" :key="taskIndex">
                                    {{task.name}} -
                                    <span v-if="task.state.toLowerCase() === 'failure'" class="biz-danger-text">
                                        {{task.state}}
                                    </span>
                                    <span v-else-if="task.state.toLowerCase() === 'success'" class="biz-success-text">
                                        {{task.state}}
                                    </span>
                                    <span v-else-if="task.state.toLowerCase() === 'running'" class="biz-warning-text">
                                        {{task.state}}
                                    </span>
                                    <div v-else-if="task.state.indexOf('html-tag') > -1" v-html="task.state">
                                    </div>
                                    <span v-else>
                                        {{task.state}}
                                    </span>
                                </div>
                                <div v-if="op.status.toLowerCase() === 'success'" class="biz-success-text f14" style="margin: 0 0 5px 0; font-weight: 700; margin-left: 20px;">
                                    {{$t('操作成功')}}
                                </div>
                                <div v-else-if="op.status.toLowerCase() === 'failed'" class="biz-danger-text f14" style="margin: 0 0 5px 0; font-weight: 700; margin-left: 20px;">
                                    {{$t('操作失败')}}
                                    <template v-if="op.errorMsgList && op.errorMsgList.length">
                                        <span>{{op.errorMsgList[0]}}</span>
                                        <template v-for="(msg, msgIndex) in op.errorMsgList">
                                            <div :key="msgIndex" v-if="msgIndex > 0">{{msg}}</div>
                                        </template>
                                    </template>
                                    <template v-else>
                                        <i18n path="请联系“{user}”解决">
                                            <a place="user" :href="PROJECT_CONFIG.doc.contact" class="bk-text-button">{{$t('蓝鲸容器助手')}}</a>
                                        </i18n>
                                    </template>
                                </div>
                                <div style="margin: 10px 0px 5px 13px; font-size: 10px;" v-else>
                                    <div class="bk-spin-loading bk-spin-loading-small bk-spin-loading-primary">
                                        <div class="rotate rotate1"></div>
                                        <div class="rotate rotate2"></div>
                                        <div class="rotate rotate3"></div>
                                        <div class="rotate rotate4"></div>
                                        <div class="rotate rotate5"></div>
                                        <div class="rotate rotate6"></div>
                                        <div class="rotate rotate7"></div>
                                        <div class="rotate rotate8"></div>
                                    </div>
                                    {{$t('正在加载中...')}}
                                </div>
                            </div>
                        </template>
                    </div>
                </template>
            </div>
        </bk-sideslider>
        <!-- 集群升级 -->
        <bcs-dialog v-model="showUpdateDialog" header-position="left" width="448"
            :loading="isUpgrading" :close-icon="false" :mask-close="false" :title="$t('集群升级')"
            @confirm="handleConfirmUpdateCluster" @cancel="handleCancelUpdateCluster">
            <main class="bk-form update-cluster-form" v-bkloading="{ isLoading: versionLoading, opacity: 1 }">
                <template v-if="versionList.length">
                    <bk-alert type="error" class="mb15" :title="$t('升级过程需要业务停机；升级完后暂不支持版本回退')"></bk-alert>
                    <div class="form-item">
                        <label>{{$t('集群版本')}}：<span class="red">*</span></label>
                        <div class="form-item-inner mt10">
                            <div style="display: inline-block;" class="mr5">
                                <bk-selector
                                    style="width: 400px;"
                                    :placeholder="$t('请选择')"
                                    :searchable="true"
                                    :setting-key="'id'"
                                    :display-key="'name'"
                                    :selected.sync="version"
                                    :list="versionList">
                                </bk-selector>
                            </div>
                        </div>
                    </div>
                </template>
                <template v-else>
                    <bk-exception type="empty" scene="part" style="margin: 0 0 30px 0;">
                        <p>{{$t('当前集群暂无可用的升级版本')}}</p>
                    </bk-exception>
                </template>
            </main>
        </bcs-dialog>
        <!-- 集群删除确认弹窗 -->
        <tip-dialog
            ref="clusterNoticeDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :title="$t('确定删除集群？')"
            :sub-title="$t('此操作无法撤回，请确认：')"
            :check-list="clusterNoticeList"
            :confirm-btn-text="$t('确定')"
            :cancel-btn-text="$t('取消')"
            :confirm-callback="confirmDeleteCluster">
        </tip-dialog>
        <!-- 编辑项目集群信息 -->
        <ProjectConfig v-model="isProjectConfDialogShow"></ProjectConfig>
    </div>
</template>

<script lang="ts">
    /* eslint-disable camelcase */
    import { computed, defineComponent, ref } from '@vue/composition-api'
    import ApplyHost from './apply-host.vue'
    import ProjectConfig from '@/views/project/project-config.vue'
    import tipDialog from '@/components/tip-dialog/index.vue'
    import applyPerm from '@/mixins/apply-perm'
    import { useClusterList, useClusterOverview, useClusterOperate } from './use-cluster'

    export default defineComponent({
        components: {
            ApplyHost,
            ProjectConfig,
            tipDialog
        },
        mixins: [applyPerm],
        setup (props, ctx) {
            const { $store, $router, $i18n, $bkInfo } = ctx.root
            const curProject = computed(() => {
                return $store.state.curProject
            })
            const kindMap = ref({
                1: 'K8S',
                2: 'Mesos'
            })
            // 获取集群状态的中文
            const statusTextMap = {
                'removing': $i18n.t('正在删除中，请稍等···'),
                'upgrading': $i18n.t('正在升级中，请稍等···'),
                'upgrade_failed': $i18n.t('升级失败，请重试'),
                'remove_failed': $i18n.t('删除失败，请重试'),
                'initializing': $i18n.t('正在初始化中，请稍等···'),
                'so_initializing': $i18n.t('正在初始化中，请稍等···'),
                'initial_checking': $i18n.t('正在初始化中，请稍等···'),
                'initial_failed': $i18n.t('初始化失败，请重试'),
                'so_init_failed': $i18n.t('初始化失败，请重试'),
                'check_failed': $i18n.t('初始化失败，请重试')
            }
            // 获取失败重试按钮文案
            const redoTextMap = {
                'upgrade_failed': $i18n.t('重新升级'),
                'remove_failed': $i18n.t('重新删除'),
                'initial_failed': $i18n.t('重新初始化'),
                'so_init_failed': $i18n.t('重新初始化'),
                'check_failed': $i18n.t('重新初始化')
            }
            // 指标信息配置
            const clusterMetricList = [
                {
                    id: 'cpu_usage',
                    title: $i18n.t('CPU使用率'),
                    theme: 'primary'
                },
                {
                    id: 'memory_usage',
                    title: $i18n.t('内存使用率'),
                    theme: 'success'
                },
                {
                    id: 'disk_usage',
                    title: $i18n.t('磁盘使用率'),
                    theme: 'warning'
                }
            ]
            // 集群列表
            const { clusterList, getClusterList, permissions, curProjectId } = useClusterList(ctx)
            const isLoading = ref(false)
            const handleGetClusterList = async () => {
                isLoading.value = true
                await getClusterList()
                isLoading.value = false
            }
            handleGetClusterList()
            // 集群指标
            const { getClusterOverview, clusterOverviewMap } = useClusterOverview(ctx, clusterList)
            // 获取集群指标项百分比
            const getMetricPercent = (cluster, item) => {
                const data = getClusterOverview(cluster.cluster_id)
                if (!data) return 0

                let used = 0
                let total = 0
                if (item.id === 'cpu_usage') {
                    used = data?.[item.id]?.used
                    total = data?.[item.id]?.total
                } else {
                    used = data?.[item.id]?.used_bytes
                    total = data?.[item.id]?.total_bytes
                }

                if (!Number(total)) {
                    return 0
                }
                let ret = Number(used) / Number(total) * 100
                if (ret !== 0 && ret !== 100) {
                    ret = Number(ret.toFixed(2))
                }

                return ret
            }
            // 指标超出 80 时显示异常角标
            const showCorner = (cluster) => {
                return clusterMetricList.some(item => getMetricPercent(cluster, item) >= 80)
            }

            // 集群信息编辑
            const isProjectConfDialogShow = ref(false)
            const handleShowProjectConf = () => {
                isProjectConfDialogShow.value = true
            }
            // todo 权限校验（后面要删除）
            const validatePermission = async (action: string, resourceList) => {
                if (permissions.value[action]) return

                await $store.dispatch('getMultiResourcePermissions', {
                    project_id: curProjectId.value,
                    operator: 'or',
                    resource_list: resourceList
                })
            }
            // 跳转创建集群界面
            const goCreateCluster = async () => {
                await validatePermission('create', [
                    {
                        policy_code: 'create',
                        resource_type: 'cluster_test'
                    },
                    {
                        policy_code: 'create',
                        resource_type: 'cluster_prod'
                    }])
                $router.push({ name: 'clusterCreate' })
            }
            // 跳转预览界面
            const goOverview = async (cluster) => {
                if (cluster.status !== 'normal') return

                // todo
                if (!cluster.permissions.view) {
                    await $store.dispatch('getResourcePermissions', {
                        project_id: curProjectId.value,
                        policy_code: 'view',
                        resource_code: cluster.cluster_id,
                        resource_name: cluster.name,
                        resource_type: `cluster_${cluster.environment === 'stag' ? 'test' : 'prod'}`
                    })
                }
                $router.push({
                    name: 'clusterOverview',
                    params: {
                        clusterId: cluster.cluster_id
                    }
                })
            }
            // 跳转集群信息界面
            const goClusterInfo = async (cluster) => {
                if (!cluster.permissions.view) {
                    const type = `cluster_${cluster.environment === 'stag' ? 'test' : 'prod'}`
                    const params = {
                        project_id: curProjectId.value,
                        policy_code: 'view',
                        resource_code: cluster.cluster_id,
                        resource_name: cluster.name,
                        resource_type: type
                    }
                    await $store.dispatch('getResourcePermissions', params)
                }
                $router.push({
                    name: 'clusterInfo',
                    params: {
                        clusterId: cluster.cluster_id
                    }
                })
            }
            const { deleteCluster, upgradeCluster, reUpgradeCluster, reInitializationCluster } = useClusterOperate(ctx)
            const curOperateCluster = ref<any>(null)
            // 集群删除
            const clusterNoticeDialog = ref<any>(null)
            const clusterNoticeList = ref([
                {
                    id: 1,
                    text: $i18n.t('将master主机归还到你业务的空闲机模块'),
                    isChecked: false
                },
                {
                    id: 2,
                    text: $i18n.t('清理其它容器服务相关组件'),
                    isChecked: false
                }
            ])
            const confirmDeleteCluster = async () => {
                // todo
                if (!curOperateCluster.value.permissions.delete) {
                    await $store.dispatch('getResourcePermissions', {
                        project_id: curProjectId.value,
                        policy_code: 'delete',
                        resource_code: curOperateCluster.value.cluster_id,
                        resource_name: curOperateCluster.value.name,
                        resource_type: `cluster_${curOperateCluster.value.environment === 'stag' ? 'test' : 'prod'}`
                    })
                }
                const result = await deleteCluster(curOperateCluster.value)
                result && handleGetClusterList()
            }
            const handleDeleteCluster = (cluster) => {
                if (!cluster.allow) return

                curOperateCluster.value = cluster
                clusterNoticeDialog.value && clusterNoticeDialog.value.show()
            }
            // 集群升级
            const showUpdateDialog = ref(false)
            const isUpgrading = ref(false)
            const versionLoading = ref(false)
            const version = ref('')
            const versionList = ref([])
            const handleConfirmUpdateCluster = async () => {
                if (!version.value || !curOperateCluster.value) return

                isUpgrading.value = true
                const result = await upgradeCluster(curOperateCluster.value, version.value)
                if (result) {
                    await handleGetClusterList()
                    showUpdateDialog.value = false
                }
                isUpgrading.value = false
            }
            const handleCancelUpdateCluster = async () => {
                versionList.value = []
                version.value = ''
                curOperateCluster.value = null
            }
            const handleUpdateCluster = async (cluster) => {
                showUpdateDialog.value = true
                versionLoading.value = true
                const res = await $store.dispatch('cluster/getClusterVersion', {
                    projectId: cluster.project_id,
                    clusterId: cluster.cluster_id
                }).catch(() => ({ data: [] }))
                versionList.value = res.data.map(item => ({
                    id: item,
                    name: item
                }))
                version.value = res.data[0] // 默认选取第一个
                curOperateCluster.value = cluster
                versionLoading.value = false
            }
            // 集群日志
            const showLogDialog = ref(false)
            const logEndState = ref('')
            const logList = ref([])
            const logTimer = ref<any>(null)
            const fetchLogData = async (cluster) => {
                const res = await $store.dispatch('cluster/getClusterLogs', {
                    projectId: cluster.project_id,
                    clusterId: cluster.cluster_id
                }).catch(() => {
                    return {
                        data: {
                            status: 'failed'
                        }
                    }
                })
                const { status, log = [], error_msg_list: errorMsgList = [] } = res.data
                // 最终的状态 running / failed / success
                logEndState.value = status
                logList.value = log.map(operation => {
                    if (operation.log.node_tasks) {
                        operation.log.node_tasks.forEach(task => {
                            task.state = task.state.replace(/\|/ig, '<p class="html-tag"></p>')
                            task.state = task.state.replace(/(Failed)/ig, '<span class="biz-danger-text">$1</span>')
                            task.state = task.state.replace(/(OK)/ig, '<span class="biz-success-text">$1</span>')
                        })
                    }
                    operation.errorMsgList = errorMsgList
                    return operation
                })

                if (logEndState.value === 'running' && showLogDialog.value) {
                    logTimer.value = setTimeout(() => {
                        fetchLogData(cluster)
                    }, 5000)
                } else {
                    clearTimeout(logTimer.value)
                    logTimer.value = null
                }
            }
            const handleShowLog = (cluster) => {
                showLogDialog.value = true
                curOperateCluster.value = cluster
                fetchLogData(cluster)
            }
            const handleCloseLog = () => {
                curOperateCluster.value = null
                clearTimeout(logTimer.value)
            }
            // 重新升级
            const handleReUpgrade = (cluster) => {
                $bkInfo({
                    type: 'warning',
                    clsName: 'custom-info-confirm',
                    title: $i18n.t('确认操作'),
                    subTitle: $i18n.t('确定重新升级？'),
                    defaultInfo: true,
                    confirmFn: async () => {
                        const result = await reUpgradeCluster(cluster)
                        result && await handleGetClusterList()
                    }
                })
            }
            // 重新初始化
            const handleReInitialization = async (cluster) => {
                $bkInfo({
                    type: 'warning',
                    clsName: 'custom-info-confirm',
                    title: $i18n.t('确认操作'),
                    subTitle: $i18n.t('确定重新初始化？'),
                    defaultInfo: true,
                    confirmFn: async () => {
                        const result = await reInitializationCluster(cluster)
                        result && await handleGetClusterList()
                    }
                })
            }
            // 失败重试
            const handleRedo = (cluster) => {
                switch (cluster.status) {
                    case 'upgrade_failed':
                        handleReUpgrade(cluster)
                        break
                    case 'remove_failed':
                        handleDeleteCluster(cluster)
                        break
                    case 'initial_failed':
                        handleReInitialization(cluster)
                        break
                    case 'so_init_failed':
                        handleReInitialization(cluster)
                        break
                    case 'check_failed':
                        handleReInitialization(cluster)
                        break
                }
            }

            return {
                isLoading,
                clusterList,
                curProject,
                kindMap,
                statusTextMap,
                redoTextMap,
                clusterMetricList,
                getClusterOverview,
                clusterOverviewMap,
                isProjectConfDialogShow,
                curOperateCluster,
                getMetricPercent,
                handleShowProjectConf,
                goCreateCluster,
                goOverview,
                goClusterInfo,
                clusterNoticeList,
                clusterNoticeDialog,
                confirmDeleteCluster,
                handleDeleteCluster,
                showLogDialog,
                logEndState,
                logList,
                handleShowLog,
                handleCloseLog,
                handleRedo,
                showCorner,
                version,
                versionLoading,
                showUpdateDialog,
                isUpgrading,
                versionList,
                handleCancelUpdateCluster,
                handleConfirmUpdateCluster,
                handleUpdateCluster
            }
        }
    })
</script>

<style lang="postcss" scoped>
    @import './index.css';
    @import './status-mark-corner.css';
    @import './status-progress.css';
    .biz-error-message {
        white-space: normal;
        text-align: left;
        max-height: 200px;
        overflow: auto;
        margin: 0 0 15px 0;
    }

    .biz-message {
        margin-bottom: 0;
        h3 {
            text-align: left;
            font-size: 14px;
        }
        p {
            text-align: left;
            font-size: 13px;
        }
    }
    .guide-link {
        display: flex;
        align-items: center;
        justify-content: center;
        i {
            font-size: 12px;
        }
    }
    .guide-btn-group {
        display: flex;
        align-items: center;
        justify-content: center;
        /deep/ .bk-button-normal {
            height: 38px;
        }
    }
    .apply-host {
        /deep/ .bk-button-normal {
            line-height: 38px;
            font-size: 16px;
        }
    }
    .bk-dropdown-list {
        li.disabled a {
            width: 100%;
            cursor: not-allowed;
            color: rgb(204, 204, 204);
        }
    }
</style>
