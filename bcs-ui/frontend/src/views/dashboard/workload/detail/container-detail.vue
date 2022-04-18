<template>
    <div class="workload-detail">
        <div class="workload-detail-info" v-bkloading="{ isLoading }">
            <div class="workload-info-basic">
                <div class="basic-left">
                    <span class="name mr20">{{ detail && detail.container_name }}</span>
                    <div class="basic-wrapper">
                        <div class="basic-item">
                            <span class="label">{{ $t('主机IP') }}</span>
                            <span class="value">{{ detail && detail.host_ip || '--' }}</span>
                        </div>
                        <div class="basic-item">
                            <span class="label">{{ $t('容器IP') }}</span>
                            <span class="value">{{ detail && detail.container_ip || '--' }}</span>
                        </div>
                    </div>
                </div>
            </div>
            <div class="workload-main-info">
                <div class="info-item">
                    <span class="label">{{ $t('主机名称') }}</span>
                    <span class="value" v-bk-overflow-tips>{{ detail && detail.host_name }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('容器ID') }}</span>
                    <span class="value" v-bk-overflow-tips>{{ detail && detail.container_id }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('镜像') }}</span>
                    <span class="value" v-bk-overflow-tips>{{ detail && detail.image }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('网络模式') }}</span>
                    <span class="value">{{ detail && detail.network_mode }}</span>
                </div>
            </div>
        </div>
        <div class="workload-detail-body">
            <div class="workload-metric">
                <Metric :title="$t('CPU使用率')" metric="cpu_usage" :params="params" category="containers" colors="#30d878"></Metric>
                <Metric :title="$t('内存使用率')" metric="memory_usage" :params="params" unit="byte" category="containers" colors="#3a84ff"></Metric>
                <Metric :title="$t('磁盘IO总量')"
                    :metric="['disk_read', 'disk_write']"
                    :params="params"
                    category="containers"
                    unit="byte"
                    :colors="['#853cff', '#30d878']"
                    :suffix="[$t('读'), $t('写')]">
                </Metric>
            </div>
            <bcs-tab class="workload-tab" :active.sync="activePanel" type="card" :label-height="40">
                <bcs-tab-panel name="ports" :label="$t('端口映射')">
                    <bk-table :data="ports">
                        <bk-table-column label="Name" prop="name">
                            <template #default="{ row }">
                                {{ row.name || '--' }}
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Host Port" prop="hostPort">
                            <template #default="{ row }">
                                {{ row.hostPort || '--' }}
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Container Port" prop="containerPort"></bk-table-column>
                        <bk-table-column label="Protocol" prop="protocol"></bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="command" :label="$t('命令')">
                    <bk-table :data="command">
                        <bk-table-column label="Command" prop="command">
                            <template #default="{ row }">
                                {{ row.command || '--' }}
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Args" prop="args">
                            <template #default="{ row }">
                                {{ row.args || '--' }}
                            </template>
                        </bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="volumes" :label="$t('挂载卷')">
                    <bk-table :data="volumes">
                        <bk-table-column label="Host Path" prop="host_path"></bk-table-column>
                        <bk-table-column label="Mount Path" prop="mount_path"></bk-table-column>
                        <bk-table-column label="ReadOnly" prop="readonly">
                            <template #default="{ row }">
                                {{ row.readonly }}
                            </template>
                        </bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="env" :label="$t('环境变量')">
                    <bk-table :data="envs" v-bkloading="{ isLoading: envsTableLoading }">
                        <bk-table-column label="Key" prop="name"></bk-table-column>
                        <bk-table-column label="Value" prop="value"></bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="label" :label="$t('标签')">
                    <bk-table :data="labels">
                        <bk-table-column label="Key" prop="key"></bk-table-column>
                        <bk-table-column label="Value" prop="val"></bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="resources" :label="$t('资源限制')">
                    <bk-table :data="resources">
                        <bk-table-column label="Cpu">
                            <template #default="{ row }">
                                {{ `requests: ${row.requests ? row.requests.cpu : '--'} | limits: ${row.limits ? row.limits.cpu : '--'}` }}
                            </template>
                        </bk-table-column>
                        <bk-table-column label="Memory">
                            <template #default="{ row }">
                                {{ `requests: ${row.requests ? row.requests.memory : '--'} | limits: ${row.limits ? row.limits.memory : '--'}` }}
                            </template>
                        </bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
            </bcs-tab>
        </div>
    </div>
</template>
<script lang="ts">
    /* eslint-disable camelcase */
    import { defineComponent, toRefs, computed, onMounted, ref } from '@vue/composition-api'
    import { bkOverflowTips } from 'bk-magic-vue'
    import StatusIcon from '../../common/status-icon'
    import Metric from '../../common/metric.vue'

    export interface IContainerDetail {
        command: object;
        labels: any[];
        ports: any[];
        resources: object;
        volumes: any[];
    }

    export default defineComponent({
        name: 'ContainerDetail',
        components: {
            StatusIcon,
            Metric
        },
        directives: {
            bkOverflowTips
        },
        props: {
            namespace: {
                type: String,
                default: '',
                required: true
            },
            // pod名
            pod: {
                type: String,
                default: '',
                required: true
            },
            // 容器名
            name: {
                type: String,
                default: '',
                required: true
            },
            // 容器ID
            id: {
                type: String,
                default: '',
                required: true
            }
        },
        setup (props, ctx) {
            const { name, namespace, pod } = toRefs(props)
            const { $store } = ctx.root

            // 详情loading
            const isLoading = ref(false)
            // 环境变量表格loading
            const envsTableLoading = ref(false)
            // 详情数据
            const detail = ref<IContainerDetail|null>(null)
            const activePanel = ref('ports')

            // 图表指标参数
            const params = computed(() => {
                return {
                    // container_ids: [id.value],
                    namespace: namespace.value,
                    pod_name: pod.value,
                    container_name: name.value,
                    $podId: pod.value
                }
            })

            // 端口映射
            const ports = computed(() => {
                return detail.value?.ports || []
            })
            // 命令
            const command = computed(() => {
                return [detail.value?.command || {}]
            })
            // 挂载卷
            const volumes = computed(() => {
                return detail.value?.volumes || []
            })
            // 标签数据
            const labels = computed(() => {
                return detail.value?.labels || []
            })
            // 资源限额
            const resources = computed(() => {
                return [detail.value?.resources || { requests: {}, limits: {} }]
            })
            // 环境变量
            const envs = ref([])

            // 容器详情
            const handleGetDetail = async () => {
                isLoading.value = true
                detail.value = await $store.dispatch('dashboard/retrieveContainerDetail', {
                    $namespaceId: namespace.value,
                    $category: 'pods',
                    $name: pod.value,
                    $containerName: name.value
                })
                isLoading.value = false
                return detail.value
            }

            // 容器环境变量
            const handleGetContainerEnv = async () => {
                envsTableLoading.value = true
                envs.value = await $store.dispatch('dashboard/fetchContainerEnvInfo', {
                    $namespaceId: namespace.value,
                    $podId: pod.value,
                    $containerName: name.value
                })
                envsTableLoading.value = false
                return envs.value
            }

            onMounted(() => {
                handleGetDetail()
                handleGetContainerEnv()
            })

            return {
                params,
                isLoading,
                detail,
                activePanel,
                ports,
                command,
                resources,
                volumes,
                labels,
                envs,
                handleGetDetail,
                handleGetContainerEnv
            }
        }
    })
</script>
<style lang="postcss" scoped>
@import './detail-info.css';
.workload-detail {
    width: 100%;
    &-info {
        @mixin detail-info 4;
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
