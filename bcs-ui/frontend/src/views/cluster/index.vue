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
                            <h2 :class="['cluster-title', { clickable: cluster.status === 'RUNNING' }]"
                                v-bk-tooltips.top="{ content: cluster.name, delay: 500 }" @click="goOverview(cluster)">
                                {{ cluster.name }}
                            </h2>
                            <p class="cluster-metadata">
                                <span class="cluster-id" v-bk-tooltips.top="{ content: cluster.cluster_id, delay: 500 }">
                                    {{cluster.cluster_id}}
                                </span>
                                <span v-if="['stag', 'debug'].includes(cluster.environment)" class="stag">
                                    {{$t('测试')}}
                                </span>
                                <span v-else-if="cluster.environment === 'prod'" class="prod">
                                    {{$t('正式')}}
                                </span>
                            </p>
                            <!-- 集群操作菜单 -->
                            <bk-dropdown-menu v-if="cluster.status === 'RUNNING'">
                                <bk-button class="cluster-opera-btn" slot="dropdown-trigger">
                                    <i class="bcs-icon bcs-icon-more"></i>
                                </bk-button>
                                <ul class="bk-dropdown-list" slot="dropdown-content">
                                    <li @click="goOverview(cluster)"><a href="javascript:;">{{$t('总览')}}</a></li>
                                    <li @click="goClusterInfo(cluster)"><a href="javascript:;">{{$t('集群信息')}}</a></li>
                                    <li :class="{ disabled: !allowDelete(cluster) }"
                                        v-bk-tooltips="{
                                            content: $t('您需要删除集群内所有节点后，再进行集群删除操作'),
                                            placement: 'right',
                                            boundary: window,
                                            interactive: false,
                                            disabled: allowDelete(cluster)
                                        }" @click="handleDeleteCluster(cluster)">
                                        <a href="javascript:;">{{$t('删除')}}</a>
                                    </li>
                                    <li v-if="!(clusterPerm[cluster.clusterID] && clusterPerm[cluster.clusterID].policy.use)">
                                        <a :href="createApplyPermUrl({
                                            policy: 'use',
                                            projectCode: curProject.project_code,
                                            idx: `cluster_${cluster.environment === 'prod' ? 'prod' : 'test'}:${cluster.cluster_id}`
                                        })" target="_blank">{{$t('申请使用权限')}}</a>
                                    </li>
                                </ul>
                            </bk-dropdown-menu>
                            <bk-dropdown-menu v-else-if="allowDelete(cluster)">
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
                            <!-- 进行中（移除中、初始化中） -->
                            <div class="biz-status-box" v-if="['INITIALIZATION', 'DELETING'].includes(cluster.status)">
                                <div class="status-icon" v-bkloading="{ isLoading: true, theme: 'primary', mode: 'spin' }"></div>
                                <p class="status-text">{{ statusTextMap[cluster.status] }}</p>
                                <div class="status-opera">
                                    <bk-button text @click="handleShowLog(cluster)">{{$t('查看日志')}}</bk-button>
                                </div>
                            </div>
                            <!-- 失败 -->
                            <div class="biz-status-box" v-else-if="['CREATE-FAILURE', 'DELETE-FAILURE'].includes(cluster.status)">
                                <div class="status-icon danger">
                                    <i class="bcs-icon bcs-icon-close-circle"></i>
                                </div>
                                <p class="status-text">{{ statusTextMap[cluster.status] }}</p>
                                <div class="status-opera">
                                    <bk-button text @click="handleShowLog(cluster)">{{$t('查看日志')}}</bk-button> |
                                    <bk-button text @click="handleRetry(cluster)">{{ $t('重试') }}</bk-button>
                                </div>
                            </div>
                            <!-- 正常 -->
                            <template v-else-if="cluster.status === 'RUNNING'">
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
                                <bk-button class="add-node-btn" @click="goNodeInfo(cluster)">
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
            <div class="biz-guide-box" v-else-if="!isLoading">
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
        </div>
        <!-- 集群日志 -->
        <bcs-sideslider
            :is-show.sync="showLogDialog"
            :title="curOperateCluster && curOperateCluster.cluster_id"
            :width="640"
            quick-close
            @hidden="handleCloseLog">
            <template #content>
                <div class="log-wrapper">
                    <bk-table :data="taskData">
                        <bk-table-column :label="$t('步骤')" prop="taskName"></bk-table-column>
                        <bk-table-column :label="$t('状态')" prop="status">
                            <template #default="{ row }">
                                <StatusIcon :status="row.status" :status-color-map="statusColorMap">
                                    {{ taskStatusTextMap[row.status.toLowerCase()] }}
                                </StatusIcon>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('内容')" prop="message"></bk-table-column>
                    </bk-table>
                </div>
            </template>
        </bcs-sideslider>
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
    import { useClusterList, useClusterOverview, useClusterOperate, useTask } from './use-cluster'
    import StatusIcon from '@/views/dashboard/common/status-icon'

    export default defineComponent({
        components: {
            ApplyHost,
            ProjectConfig,
            tipDialog,
            StatusIcon
        },
        mixins: [applyPerm],
        setup (props, ctx) {
            const { $store, $router, $i18n, $bkMessage, $bkInfo } = ctx.root
            const curProject = computed(() => {
                return $store.state.curProject
            })
            const kindMap = ref({
                1: 'K8S',
                2: 'Mesos'
            })
            // 获取集群状态的中文
            const statusTextMap = {
                'INITIALIZATION': $i18n.t('正在初始化中，请稍等···'),
                'DELETING': $i18n.t('正在删除中，请稍等···'),
                'CREATE-FAILURE': $i18n.t('创建失败，请重试'),
                'DELETE-FAILURE': $i18n.t('删除失败，请重试')
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
            const { clusterList, getClusterList, clusterPerm, curProjectId, clusterExtraInfo, permissions } = useClusterList(ctx)
            const allowDelete = (cluster) => {
                return !!clusterExtraInfo.value[cluster.clusterID]?.canDeleted
            }
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
                    }]
                )
                $router.push({ name: 'clusterCreate' })
            }
            // 跳转预览界面
            const goOverview = async (cluster) => {
                if (cluster.status !== 'RUNNING') return

                // todo
                if (!clusterPerm.value[cluster.clusterID]?.policy?.view) {
                    await $store.dispatch('getResourcePermissions', {
                        project_id: curProjectId.value,
                        policy_code: 'view',
                        resource_code: cluster.cluster_id,
                        resource_name: cluster.name,
                        resource_type: `cluster_${cluster.environment === 'prod' ? 'prod' : 'test'}`
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
                if (!clusterPerm.value[cluster.clusterID]?.policy?.view) {
                    const type = `cluster_${cluster.environment === 'prod' ? 'prod' : 'test'}`
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
            // 跳转添加节点界面
            const goNodeInfo = async (cluster) => {
                if (!clusterPerm.value[cluster.clusterID]?.policy?.view) {
                    await $store.dispatch('getResourcePermissions', {
                        project_id: curProjectId.value,
                        policy_code: 'view',
                        resource_code: cluster.cluster_id,
                        resource_name: cluster.name,
                        resource_type: `cluster_${cluster.environment === 'prod' ? 'prod' : 'test'}`
                    })
                }
                $router.push({
                    name: 'clusterNode',
                    params: {
                        clusterId: cluster.cluster_id
                    }
                })
            }
            const { deleteCluster, retryCluster } = useClusterOperate(ctx)
            const curOperateCluster = ref<any>(null)
            // 集群删除
            const clusterNoticeDialog = ref<any>(null)
            const clusterNoticeList = computed(() => {
                return [
                    {
                        id: 1,
                        text: $i18n.t('您正在尝试删除集群 {clusterName}，此操作不可逆，请谨慎操作', { clusterName: curOperateCluster.value?.clusterID }),
                        isChecked: false
                    },
                    {
                        id: 2,
                        text: $i18n.t('请确认已清理该集群下的所有应用与节点'),
                        isChecked: false
                    },
                    {
                        id: 3,
                        text: $i18n.t('集群删除时会清理集群上的工作负载、服务、路由等集群上的所有资源'),
                        isChecked: false
                    },
                    {
                        id: 4,
                        text: $i18n.t('集群删除后服务器如不再使用请尽快回收，避免产生不必要的成本'),
                        isChecked: false
                    }
                ]
            })
            const confirmDeleteCluster = async () => {
                isLoading.value = true
                // todo
                if (!clusterPerm.value[curOperateCluster.value.clusterID]?.policy?.delete) {
                    await $store.dispatch('getResourcePermissions', {
                        project_id: curProjectId.value,
                        policy_code: 'delete',
                        resource_code: curOperateCluster.value.cluster_id,
                        resource_name: curOperateCluster.value.name,
                        resource_type: `cluster_${curOperateCluster.value.environment === 'prod' ? 'prod' : 'test'}`
                    })
                }
                const result = await deleteCluster(curOperateCluster.value)
                if (result) {
                    await handleGetClusterList()
                    $bkMessage({
                        theme: 'success',
                        message: $i18n.t('任务下发成功')
                    })
                }
                isLoading.value = false
            }
            const handleDeleteCluster = (cluster) => {
                if (!allowDelete(cluster)) return

                curOperateCluster.value = cluster
                setTimeout(() => {
                    clusterNoticeDialog.value && clusterNoticeDialog.value.show()
                }, 0)
            }
            // 集群日志
            const { taskList } = useTask(ctx)
            const showLogDialog = ref(false)
            const latestTask = ref<any>(null)
            const taskTimer = ref<any>(null)
            const statusColorMap = ref({
                initialzing: 'blue',
                running: 'blue',
                success: 'green',
                failure: 'red',
                timeout: 'red',
                notstarted: 'blue'
            })
            const taskStatusTextMap = ref({
                initialzing: $i18n.t('初始化中'),
                running: $i18n.t('运行中'),
                success: $i18n.t('成功'),
                failure: $i18n.t('失败'),
                timeout: $i18n.t('超时'),
                notstarted: $i18n.t('未执行')
            })
            const taskData = computed(() => {
                const steps = latestTask.value?.stepSequence || []
                return steps.map(step => {
                    return latestTask.value?.steps[step]
                })
            })
            const fetchLogData = async (cluster) => {
                const res = await taskList(cluster)
                latestTask.value = res.latestTask
                if (['RUNNING', 'INITIALZING'].includes(latestTask.value?.status)) {
                    taskTimer.value = setTimeout(() => {
                        fetchLogData(cluster)
                    }, 5000)
                } else {
                    clearTimeout(taskTimer.value)
                    taskTimer.value = null
                }
            }
            const handleShowLog = (cluster) => {
                showLogDialog.value = true
                curOperateCluster.value = cluster
                fetchLogData(cluster)
            }
            const handleCloseLog = () => {
                curOperateCluster.value = null
                clearTimeout(taskTimer.value)
            }
            // 失败重试
            const handleRetry = async (cluster) => {
                isLoading.value = true
                if (cluster.status === 'CREATE-FAILURE') {
                    // 创建重试
                    $bkInfo({
                        type: 'warning',
                        title: $i18n.t('确认重新创建集群'),
                        clsName: 'custom-info-confirm default-info',
                        subTitle: cluster.clusterName,
                        confirmFn: async () => {
                            isLoading.value = true
                            const result = await retryCluster(cluster)
                            if (result) {
                                await handleGetClusterList()
                                $bkMessage({
                                    theme: 'success',
                                    message: $i18n.t('任务下发成功')
                                })
                            }
                            isLoading.value = false
                        }
                    })
                } else if (cluster.status === 'DELETE-FAILURE') {
                    // 删除重试
                    $bkInfo({
                        type: 'warning',
                        title: $i18n.t('确认删除集群'),
                        clsName: 'custom-info-confirm default-info',
                        subTitle: cluster.clusterName,
                        confirmFn: async () => {
                            isLoading.value = true
                            const result = await deleteCluster(cluster)
                            if (result) {
                                await handleGetClusterList()
                                $bkMessage({
                                    theme: 'success',
                                    message: $i18n.t('任务下发成功')
                                })
                            }
                            isLoading.value = false
                        }
                    })
                } else {
                    $bkMessage({
                        theme: 'error',
                        message: $i18n.t('未知状态')
                    })
                }
                isLoading.value = false
            }

            return {
                isLoading,
                clusterList,
                curProject,
                kindMap,
                statusTextMap,
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
                latestTask,
                taskData,
                statusColorMap,
                taskStatusTextMap,
                handleShowLog,
                handleCloseLog,
                handleRetry,
                showCorner,
                goNodeInfo,
                clusterPerm,
                allowDelete
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
    .log-wrapper {
        padding: 20px 30px 0 30px;
    }
</style>
