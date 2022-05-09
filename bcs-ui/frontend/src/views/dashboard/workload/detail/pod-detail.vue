<template>
    <div class="workload-detail">
        <div class="workload-detail-info" v-bkloading="{ isLoading }">
            <div class="workload-info-basic">
                <div class="basic-left">
                    <span class="name mr20">{{ metadata.name }}</span>
                    <div class="basic-wrapper">
                        <StatusIcon class="basic-item" :status="manifestExt.status"></StatusIcon>
                        <div class="basic-item">
                            <span class="label">Ready</span>
                            <span class="value">{{ manifestExt.readyCnt }} / {{ manifestExt.totalCnt }}</span>
                        </div>
                        <div class="basic-item">
                            <span class="label">Host IP</span>
                            <span class="value">{{ status.hostIP || '--' }}</span>
                        </div>
                        <div class="basic-item">
                            <span class="label">Pod IP</span>
                            <span class="value">{{ status.podIP || '--' }}</span>
                        </div>
                    </div>
                </div>
                <div class="btns">
                    <bk-button theme="primary" @click="handleShowYamlPanel">To YAML</bk-button>
                    <bk-button theme="primary"
                        v-authority="{ clickable: pagePerms.update.clickable, content: pagePerms.update.tip }"
                        @click="handleUpdateResource">{{$t('更新')}}</bk-button>
                    <bk-button theme="danger"
                        v-authority="{ clickable: pagePerms.delete.clickable, content: pagePerms.delete.tip }"
                        @click="handleDeleteResource">{{$t('删除')}}</bk-button>
                </div>
            </div>
            <div class="workload-main-info">
                <div class="info-item">
                    <span class="label">{{ $t('命名空间') }}</span>
                    <span class="value" v-bk-overflow-tips>{{ metadata.namespace }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('镜像') }}</span>
                    <span class="value" v-bk-overflow-tips="getImagesTips(manifestExt.images)">{{ manifestExt.images && manifestExt.images.join(', ') }}</span>
                </div>
                <div class="info-item">
                    <span class="label">{{ $t('节点') }}</span>
                    <span class="value" v-bk-overflow-tips>{{ spec.nodeName }}</span>
                </div>
                <div class="info-item">
                    <span class="label">UID</span>
                    <span class="value" v-bk-overflow-tips>{{ metadata.uid }}</span>
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
            <div class="workload-metric">
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
                <bcs-tab-panel name="container" :label="$t('容器')" v-bkloading="{ isLoading: containerLoading }">
                    <bk-table :data="container">
                        <bk-table-column :label="$t('容器名称')" prop="name">
                            <template #default="{ row }">
                                <bk-button class="bcs-button-ellipsis" text @click="gotoContainerDetail(row)">{{ row.name }}</bk-button>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('状态')" width="200" prop="status">
                            <template #default="{ row }">
                                <StatusIcon :status="row.status"></StatusIcon>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('镜像')" prop="image"></bk-table-column>
                        <bk-table-column :label="$t('操作')" width="200" :resizable="false" :show-overflow-tooltip="false">
                            <template #default="{ row }">
                                <bk-button text @click="handleShowTerminal(row)">WebConsole</bk-button>
                                <bk-popover placement="bottom" theme="light dropdown" :arrow="false" v-if="row.container_id && $INTERNAL && !isSharedCluster">
                                    <bk-button style="cursor: default;" text class="ml10">{{ $t('日志检索') }}</bk-button>
                                    <div slot="content">
                                        <ul>
                                            <a :href="logLinks[row.container_id] && logLinks[row.container_id].std_log_link"
                                                target="_blank" class="dropdown-item">
                                                {{ $t('标准输出检索') }}
                                            </a>
                                            <a :href="logLinks[row.container_id] && logLinks[row.container_id].file_log_link"
                                                target="_blank" class="dropdown-item">
                                                {{ $t('文件日志检索') }}
                                            </a>
                                        </ul>
                                    </div>
                                </bk-popover>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="conditions" :label="$t('状态（Conditions）')">
                    <bk-table :data="conditions">
                        <bk-table-column :label="$t('类别')" prop="type"></bk-table-column>
                        <bk-table-column :label="$t('状态')" prop="status">
                            <template #default="{ row }">
                                <StatusIcon :status="row.status"></StatusIcon>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('最后迁移时间')" prop="lastTransitionTime">
                            <template #default="{ row }">
                                {{ formatTime(row.lastTransitionTime, 'yyyy-MM-dd hh:mm:ss') }}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('原因')">
                            <template #default="{ row }">
                                {{ row.reason || '--' }}
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('消息')">
                            <template #default="{ row }">
                                {{ row.message || '--' }}
                            </template>
                        </bk-table-column>
                    </bk-table>
                </bcs-tab-panel>
                <bcs-tab-panel name="storage" :label="$t('存储')" v-bkloading="{ isLoading: storageLoading }">
                    <div class="storage storage-pvcs">
                        <div class="title">PersistentVolumeClaims</div>
                        <bk-table :data="storageTableData.pvcs">
                            <bk-table-column :label="$t('名称')" prop="metadata.name" sortable :resizable="false"></bk-table-column>
                            <bk-table-column label="Status">
                                <template #default="{ row }">
                                    <span>{{ row.status.phase || '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="Volume">
                                <template #default="{ row }">
                                    <span>{{ row.spec.volumeName || '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="Capacity">
                                <template #default="{ row }">
                                    <span>{{ row.status.capacity ? row.status.capacity.storage : '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="Access Modes">
                                <template #default="{ row }">
                                    <span>{{ handleGetExtData(row.metadata.uid, 'pvcs','accessModes').join(', ') }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="StorageClass">
                                <template #default="{ row }">
                                    <span>{{ row.spec.storageClassName || '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="VolumeMode">
                                <template #default="{ row }">
                                    <span>{{ row.spec.volumeMode || '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                                <template #default="{ row }">
                                    <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'pvcs','createTime') }">
                                        {{ handleGetExtData(row.metadata.uid, 'pvcs','age') }}
                                    </span>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                    <div class="storage storage-config">
                        <div class="title">ConfigMaps</div>
                        <bk-table :data="storageTableData.configmaps">
                            <bk-table-column :label="$t('名称')" prop="metadata.name" sortable :resizable="false"></bk-table-column>
                            <bk-table-column label="Data">
                                <template #default="{ row }">
                                    <span>{{ handleGetExtData(row.metadata.uid, 'configmaps','data').join(', ') || '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                                <template #default="{ row }">
                                    <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'configmaps','createTime') }">
                                        {{ handleGetExtData(row.metadata.uid, 'configmaps','age') }}
                                    </span>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
                    <div class="storage storage-secrets">
                        <div class="title">Secrets</div>
                        <bk-table :data="storageTableData.secrets">
                            <bk-table-column :label="$t('名称')" prop="metadata.name" sortable :resizable="false"></bk-table-column>
                            <bk-table-column label="Type">
                                <template #default="{ row }">
                                    <span>{{ row.type || '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="Data">
                                <template #default="{ row }">
                                    <span>{{ handleGetExtData(row.metadata.uid, 'secrets','data').join(', ') || '--' }}</span>
                                </template>
                            </bk-table-column>
                            <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                                <template #default="{ row }">
                                    <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'secrets','createTime') }">
                                        {{ handleGetExtData(row.metadata.uid, 'secrets','age') }}
                                    </span>
                                </template>
                            </bk-table-column>
                        </bk-table>
                    </div>
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
    </div>
</template>
<script lang="ts">
    /* eslint-disable camelcase */
    import { computed, defineComponent, onMounted, ref, toRefs } from '@vue/composition-api'
    import { bkOverflowTips } from 'bk-magic-vue'
    import StatusIcon from '../../common/status-icon'
    import Metric from '../../common/metric.vue'
    import useDetail from './use-detail'
    import { formatTime } from '@/common/util'
    import Ace from '@/components/ace-editor'
    import fullScreen from '@/directives/full-screen'

    export interface IDetail {
        manifest: any;
        manifest_ext: any;
    }

    export interface IStorage {
        pvcs: IDetail | null;
        configmaps: IDetail | null;
        secrets: IDetail | null;
    }

    export default defineComponent({
        name: 'PodDetail',
        components: {
            StatusIcon,
            Metric,
            Ace
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
            // pod 名称
            name: {
                type: String,
                default: '',
                required: true
            }
        },
        setup (props, ctx) {
            const { $store, $route, $INTERNAL } = ctx.root
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
                category: 'pods',
                defaultActivePanel: 'container',
                type: 'workloads'
            })
            const { name, namespace } = toRefs(props)
            const params = computed(() => {
                return {
                    namespace: namespace.value,
                    pod_name_list: [name.value]
                }
            })

            // 容器
            const container = ref<any[]>([])
            const containerLoading = ref(false)
            const logLinks = ref({})
            const handleGetContainer = async () => {
                containerLoading.value = true
                container.value = await $store.dispatch('dashboard/listContainers', {
                    $podId: name.value,
                    $namespaceId: namespace.value
                })
                if ($INTERNAL && container.value.length) {
                    logLinks.value = await $store.dispatch('dashboard/logLinks', {
                        container_ids: container.value.map(item => item.container_id).join(',')
                    })
                }
                containerLoading.value = false
            }
            // 状态
            const conditions = computed(() => {
                return detail.value?.manifest.status?.conditions || []
            })
            // status 数据
            const status = computed(() => detail.value?.manifest?.status || {})
            // spec 数据
            const spec = computed(() => detail.value?.manifest?.spec || {})

            // 存储
            const storage = ref<IStorage>({
                pvcs: null,
                configmaps: null,
                secrets: null
            })
            const storageTableData = computed(() => {
                return {
                    pvcs: storage.value.pvcs?.manifest.items || [],
                    configmaps: storage.value.configmaps?.manifest.items || [],
                    secrets: storage.value.secrets?.manifest.items || []
                }
            })
            // 获取存储数据
            const storageLoading = ref(false)
            const handleGetStorage = async () => {
                storageLoading.value = true
                const types = ['pvcs', 'configmaps', 'secrets']
                const promises = types.map(type => {
                    return $store.dispatch('dashboard/listStoragePods', {
                        $podId: name.value,
                        $type: type,
                        $namespaceId: namespace.value
                    })
                })
                const [pvcs = {}, configmaps = {}, secrets = {}] = await Promise.all(promises)
                storage.value = {
                    pvcs,
                    configmaps,
                    secrets
                }
                storageLoading.value = false
            }
            // 获取存储manifest_ext的字段
            const handleGetExtData = (uid, type, prop) => {
                return storage.value[type]?.manifest_ext?.[uid]?.[prop] || ''
            }

            // 跳转容器详情
            const gotoContainerDetail = (row) => {
                ctx.emit('container-detail', row)
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

            // 容器操作
            // 1. 跳转WebConsole
            const projectId = computed(() => $route.params.projectId)
            const clusterId = computed(() => $store.state.curClusterId)
            const terminalWins = new Map()
            const handleShowTerminal = (row) => {
                const url = `${window.DEVOPS_BCS_API_URL}/web_console/projects/${projectId.value}/clusters/${clusterId.value}/?namespace=${props.namespace}&pod_name=${props.name}&container_name=${row.name}`
                if (terminalWins.has(row.container_id)) {
                    const win = terminalWins.get(row.container_id)
                    if (!win.closed) {
                        terminalWins.get(row.container_id).focus()
                    } else {
                        const win = window.open(url, '_blank')
                        terminalWins.set(row.container_id, win)
                    }
                } else {
                    const win = window.open(url, '_blank')
                    terminalWins.set(row.container_id, win)
                }
            }
            // 2. 日志检索
            const isDropdownShow = ref(false)

            const isSharedCluster = computed(() => {
                return $store.getters['cluster/isSharedCluster']
            })

            onMounted(async () => {
                handleGetDetail()
                handleGetStorage()
                handleGetContainer()
            })

            return {
                params,
                container,
                conditions,
                storage,
                storageTableData,
                isLoading,
                detail,
                metadata,
                manifestExt,
                spec,
                status,
                activePanel,
                labels,
                annotations,
                storageLoading,
                containerLoading,
                yaml,
                showYamlPanel,
                pagePerms,
                isDropdownShow,
                logLinks,
                isSharedCluster,
                handleShowYamlPanel,
                handleGetStorage,
                handleGetContainer,
                gotoContainerDetail,
                handleGetExtData,
                formatTime,
                getImagesTips,
                handleUpdateResource,
                handleDeleteResource,
                handleShowTerminal
            }
        }
    })
</script>
<style lang="postcss" scoped>
@import './detail-info.css';
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
        .storage {
            margin-bottom: 24px;
            .title {
                font-size: 14px;
                color: #313238;
                margin-bottom: 8px;
            }
        }
    }
}
>>> .dropdown-item {
    display: block;
    height: 32px;
    line-height: 33px;
    padding: 0 16px;
    color: #63656e;
    font-size: 12px;
    text-decoration: none;
    white-space: nowrap;
    cursor: pointer;
    &:hover {
        background-color: #eaf3ff;
        color: #3a84ff;
    }
}
</style>
