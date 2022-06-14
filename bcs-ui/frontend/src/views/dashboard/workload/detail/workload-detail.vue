<template>
    <div class="workload-detail" v-bkloading="{ isLoading }">
        <div class="workload-detail-info">
            <div class="workload-info-basic">
                <div class="basic-left">
                    <span class="name mr20">{{ metadata.name }}</span>
                    <div class="basic-wrapper">
                        <div v-for="item in basicInfoList"
                            :key="item.label"
                            class="basic-item">
                            <span class="label">{{ item.label }}</span>
                            <span class="value">{{ item.value }}</span>
                        </div>
                    </div>
                </div>
                <div class="btns">
                    <bk-button theme="primary" @click="handleShowYamlPanel">To YAML</bk-button>
                    <template v-if="!hiddenOperate">
                        <bk-button theme="primary"
                            v-authority="{ clickable: pagePerms.update.clickable, content: pagePerms.update.tip }"
                            @click="handleUpdateResource">{{$t('更新')}}</bk-button>
                        <bk-button theme="danger"
                            v-authority="{ clickable: pagePerms.delete.clickable, content: pagePerms.delete.tip }"
                            @click="handleDeleteResource">{{$t('删除')}}</bk-button>
                    </template>
                </div>
            </div>
            <div class="workload-main-info">
                <div class="info-item">
                    <span class="label">{{ $t('命名空间') }}</span>
                    <span class="value">{{ metadata.namespace }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('镜像') }}</span>
                    <span class="value" v-bk-overflow-tips="getImagesTips(manifestExt.images)">{{ manifestExt.images && manifestExt.images.join(', ') }}</span>
                </div>
                <div class="info-item">
                    <span class="label">UID</span>
                    <span class="value">{{ metadata.uid }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('创建时间') }}</span>
                    <span class="value">{{ manifestExt.createTime }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('存在时间') }}</span>
                    <span class="value">{{ manifestExt.age }}</span>
                </div>
            </div>
        </div>
        <div class="workload-detail-body">
            <div class="workload-metric" v-bkloading="{ isLoading: podLoading }">
                <Metric :title="$t('CPU使用率')" metric="cpu_usage" :params="params" category="pods" colors="#30d878"></Metric>
                <Metric :title="$t('内存使用率')" metric="memory_usage" :params="params" unit="byte" category="pods" colors="#3a84ff"></Metric>
                <Metric :title="$t('网络')"
                    :metric="['network_receive', 'network_transmit']"
                    :params="params"
                    category="pods"
                    unit="byte"
                    :colors="['#853cff', '#30d878']"
                    :suffix="[$t('入流量'), $t('出流量')]">
                </Metric>
            </div>
            <bcs-tab class="workload-tab" :active.sync="activePanel" type="card" :label-height="40">
                <bcs-tab-panel name="pod" label="Pod" v-bkloading="{ isLoading: podLoading }">
                    <bk-table :data="pods">
                        <bk-table-column :label="$t('名称')" min-width="130" prop="metadata.name" sortable :resizable="false">
                            <template #default="{ row }">
                                <bk-button :disabled="rescheduleStatusMap[row.metadata.name]"
                                    class="bcs-button-ellipsis" text @click="gotoPodDetail(row)">{{ row.metadata.name }}</bk-button>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('镜像')" min-width="200" :resizable="false" :show-overflow-tooltip="false">
                            <template slot-scope="{ row }">
                                <span v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
                                    {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
                                </span>
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Status" width="120" :resizable="false">
                            <template slot-scope="{ row }">
                                <StatusIcon :status="handleGetExtData(row.metadata.uid, 'status')"></StatusIcon>
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Ready" width="100" :resizable="false">
                            <template slot-scope="{ row }">
                                {{handleGetExtData(row.metadata.uid, 'readyCnt')}}/{{handleGetExtData(row.metadata.uid, 'totalCnt')}}
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Restarts" width="100" :resizable="false">
                            <template slot-scope="{ row }">{{handleGetExtData(row.metadata.uid, 'restartCnt')}}</template>
                        </bk-table-column>
                        <bk-table-column label="Host IP" width="140" :resizable="false">
                            <template slot-scope="{ row }">{{row.status.hostIP || '--'}}</template>
                        </bk-table-column>
                        <bk-table-column label="Pod IP" width="140" :resizable="false">
                            <template slot-scope="{ row }">{{row.status.podIP || '--'}}</template>
                        </bk-table-column>
                        <bk-table-column label="Node" :resizable="false">
                            <template slot-scope="{ row }">{{row.spec.nodeName || '--'}}</template>
                        </bk-table-column>
                        <bk-table-column label="Age" :resizable="false">
                            <template #default="{ row }">
                                <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('操作')" width="140" :resizable="false">
                            <template #default="{ row }">
                                <bk-button text :disabled="rescheduleStatusMap[row.metadata.name]"
                                    @click="handleShowLog(row)">{{ $t('日志') }}</bk-button>
                                <bk-button class="ml10" :disabled="rescheduleStatusMap[row.metadata.name]"
                                    text @click="handleReschedule(row)">{{ $t('重新调度') }}</bk-button>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="event" :label="$t('事件')" v-if="category !== 'cronjobs'">
                    <bk-table :data="events"
                        v-bkloading="{ isLoading: eventLoading }"
                        :pagination="pagination"
                        @page-change="handlePageChange"
                        @page-limit-change="handlePageLimitChange">
                        <bk-table-column :label="$t('时间')" prop="eventTime" width="200"></bk-table-column>
                        <bk-table-column :label="$t('组件')" prop="component" width="100"></bk-table-column>
                        <bk-table-column :label="$t('对象')" prop="extraInfo.name" width="200"></bk-table-column>
                        <bk-table-column :label="$t('级别')" prop="level" width="100"></bk-table-column>
                        <bk-table-column :label="$t('事件内容')" prop="describe" min-width="300"></bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="label" :label="$t('标签')">
                    <bk-table :data="labels">
                        <bk-table-column label="Key" prop="key"></bk-table-column>
                        <bk-table-column label="Value" prop="value"></bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="annotations" :label="$t('注解')">
                    <bk-table :data="annotations">
                        <bk-table-column label="Key" prop="key"></bk-table-column>
                        <bk-table-column label="Value" prop="value"></bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
            </bcs-tab>
        </div>
        <bcs-sideslider quick-close :title="metadata.name" :is-show.sync="showYamlPanel" :width="800">
            <template #content>
                <Ace v-full-screen="{ tools: ['fullscreen', 'copy'], content: yaml }"
                    width="100%" height="100%" lang="yaml" read-only :value="yaml"></Ace>
            </template>
        </bcs-sideslider>
        <bcs-dialog class="log-dialog" v-model="logShow" width="80%" :show-footer="false" render-directive="if">
            <BcsLog
                :project-id="projectId"
                :cluster-id="clusterId"
                :namespace-id="curNamespace"
                :pod-id="curPodId"
                :default-container="defaultContainer"
                :global-loading="logLoading"
                :container-list="containerList">
            </BcsLog>
        </bcs-dialog>
    </div>
</template>
<script lang="ts">
    /* eslint-disable camelcase */
    import { defineComponent, computed, ref, onMounted, onBeforeUnmount, set } from '@vue/composition-api'
    import { bkOverflowTips } from 'bk-magic-vue'
    import StatusIcon from '../../common/status-icon'
    import Metric from '../../common/metric.vue'
    import useDetail from './use-detail'
    import detailBasicList from './detail-basic'
    import Ace from '@/components/ace-editor'
    import fullScreen from '@/directives/full-screen'
    import useInterval from '../../common/use-interval'
    import BcsLog from '@/components/bcs-log/index'
    import useLog from './use-log'

    export interface IDetail {
        manifest: any;
        manifestExt: any;
    }

    export interface IParams {
        pod_name_list: string[];
    }

    export default defineComponent({
        name: 'WorkloadDetail',
        components: {
            StatusIcon,
            Metric,
            Ace,
            BcsLog
        },
        directives: {
            bkOverflowTips,
            'full-screen': fullScreen
        },
        props: {
            namespace: {
                type: String,
                default: '',
                required: true
            },
            // workload类型
            category: {
                type: String,
                default: '',
                required: true
            },
            // kind类型
            kind: {
                type: String,
                default: '',
                required: true
            },
            // 名称
            name: {
                type: String,
                default: '',
                required: true
            },
            // 是否隐藏 更新 和 删除操作（兼容集群管理应用详情）
            hiddenOperate: {
                type: Boolean,
                default: false
            }
        },
        setup (props, ctx) {
            const { $store, $bkMessage, $i18n, $route } = ctx.root
            const {
                isLoading,
                detail,
                activePanel,
                labels,
                annotations,
                metadata,
                manifestExt,
                yaml,
                showYamlPanel,
                pagePerms,
                handleGetDetail,
                handleShowYamlPanel,
                handleUpdateResource,
                handleDeleteResource
            } = useDetail(ctx, {
                ...props,
                defaultActivePanel: 'pod',
                type: 'workloads'
            })
            const podLoading = ref(false)
            const workloadPods = ref<IDetail|null>(null)
            const basicInfoList = detailBasicList({
                category: props.category,
                detail
            })
            // pods数据
            const pods = computed(() => {
                return workloadPods.value?.manifest?.items || []
            })
            // 获取pod manifestExt数据
            const handleGetExtData = (uid, prop) => {
                return workloadPods.value?.manifestExt?.[uid]?.[prop]
            }
            // 指标参数
            const params = computed<IParams | null>(() => {
                const list = pods.value.map(item => item.metadata.name)
                return list.length
                    ? { pod_name_list: list, namespace: props.namespace }
                    : null
            })

            // 跳转pod详情
            const gotoPodDetail = (row) => {
                ctx.emit('pod-detail', row)
            }

            // 获取镜像tips
            const getImagesTips = (images) => {
                if (!images) {
                    return {
                        content: ''
                    }
                }
                return {
                    allowHTML: true,
                    maxWidth: 480,
                    content: images.join('<br />')
                }
            }

            const handleGetPodsData = async () => {
                // 获取工作负载下对应的pod数据
                const matchLabels = detail.value?.manifest?.spec?.selector?.matchLabels || {}
                const labelSelector = Object.keys(matchLabels).reduce((pre, key, index) => {
                    pre += `${index > 0 ? ',' : ''}${key}=${matchLabels[key]}`
                    return pre
                }, '')

                const data = await $store.dispatch('dashboard/listWorkloadPods', {
                    $namespaceId: props.namespace,
                    labelSelector: labelSelector,
                    ownerKind: props.kind,
                    ownerName: props.name,
                    format: "manifest"
                })
                return data
            }
            // 获取工作负载下的pods数据
            const handleGetWorkloadPods = async () => {
                podLoading.value = true
                workloadPods.value = await handleGetPodsData()
                podLoading.value = false
            }

            const projectId = computed(() => $route.params.projectId)
            const clusterId = computed(() => $store.state.curClusterId)
            // 重新调度
            const rescheduleStatusMap = ref({})
            const handleReschedule = async (row) => {
                set(rescheduleStatusMap.value, row.metadata.name, true)
                const result = await $store.dispatch('dashboard/reschedulePod', {
                    $namespaceId: props.namespace,
                    $podId: row.metadata.name
                })
                result && $bkMessage({
                    theme: 'success',
                    message: $i18n.t('调度成功')
                })
                rescheduleStatusMap.value[row.metadata.name] = false
            }
            // 事件列表
            const events = ref([])
            const eventLoading = ref(false)
            const categoryMap = {
                deployments: 'deployment',
                daemonsets: 'daemonset',
                statefulsets: 'statefulset',
                jobs: 'job'
            } // 兼容老接口 category 类型
            const pagination = ref({
                current: 1,
                count: 0,
                limit: 10
            })
            const handleGetEventList = async () => {
                // cronjobs 不支持事件查询
                if (props.category === 'cronjobs') return

                eventLoading.value = true
                const params = {
                    projectId: projectId.value,
                    instanceId: 0,
                    cluster_id: clusterId.value,
                    offset: (pagination.value.current - 1) * pagination.value.limit,
                    limit: pagination.value.limit,
                    name: props.name,
                    namespace: props.namespace,
                    category: categoryMap[props.category]
                }
                const { data } = await $store.dispatch('app/getEventList', params)
                    .catch(() => ({ data: { data: [], total: 0 } }))
                pagination.value.count = data.total
                events.value = data.data
                eventLoading.value = false
            }
            const handlePageChange = (page) => {
                pagination.value.current = page
                handleGetEventList()
            }
            const handlePageLimitChange = (limit) => {
                pagination.value.current = 1
                pagination.value.limit = limit
                handleGetEventList()
            }

            // 刷新Pod状态
            const handleRefreshPodsStatus = async () => {
                workloadPods.value = await handleGetPodsData()
            }
            const { start, stop } = useInterval(handleRefreshPodsStatus, 8000)
            onMounted(async () => {
                // 详情接口前置
                await handleGetDetail()
                await handleGetWorkloadPods()
                handleGetEventList()
                // 开启轮询
                start()
            })
            onBeforeUnmount(() => {
                stop()
            })

            return {
                isLoading,
                detail,
                metadata,
                manifestExt,
                basicInfoList,
                activePanel,
                params,
                pods,
                labels,
                annotations,
                podLoading,
                yaml,
                showYamlPanel,
                pagePerms,
                rescheduleStatusMap,
                projectId,
                clusterId,
                events,
                eventLoading,
                pagination,
                handlePageChange,
                handlePageLimitChange,
                handleShowYamlPanel,
                gotoPodDetail,
                handleGetExtData,
                getImagesTips,
                handleUpdateResource,
                handleDeleteResource,
                handleReschedule,
                ...useLog()
            }
        }
    })
</script>
<style lang="postcss" scoped>
@import './detail-info.css';
@import './pod-log.css';
.workload-detail {
    width: 100%;
    /deep/ .bk-sideslider .bk-sideslider-content {
        height: 100%;
    }
    &-info {
        @mixin detail-info 3;
    }
    &-body {
        background: #FAFBFD;
        padding: 0 24px;
        .workload-metric {
            display: flex;
            background: #fff;
            margin-top: 16px;
            height: 230px;
        }
        .workload-tab {
            margin-top: 16px;
        }
    }
}
</style>
