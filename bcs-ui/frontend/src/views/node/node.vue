<template>
    <div class="cluster-node">
        <bcs-alert
            type="info"
            class="cluster-node-tip"
        >
            <div slot="title">
                {{$t('集群就绪后，您可以创建命名空间、推送项目镜像到仓库，然后通过服务配置模板集部署服务 ')}}
                <i18n path="当前集群已添加节点数（含Master） {nodes}，还可添加节点数 {remainNodes}"
                    v-if="remainNodesCount > 0"
                >
                    <span place="nodes" class="num">{{nodesCount}}</span>
                    <span place="remainNodes" class="num">{{remainNodesCount}}</span>
                </i18n>
            </div>
        </bcs-alert>
        <!-- 操作栏 -->
        <div class="cluster-node-operate">
            <div class="left">
                <template v-if="!nodeMenu">
                    <bcs-button theme="primary"
                        icon="plus"
                        class="add-node mr10"
                        v-authority="{
                            clickable: webAnnotations.perms[localClusterId]
                                && webAnnotations.perms[localClusterId].cluster_manage,
                            actionId: 'cluster_manage',
                            resourceName: curSelectedCluster.clusterName,
                            disablePerms: true,
                            permCtx: {
                                project_id: curProject.project_id,
                                cluster_id: localClusterId
                            }
                        }"
                        @click="handleAddNode"
                    >{{$t('添加节点')}}</bcs-button>
                    <template v-if="$INTERNAL && curSelectedCluster.providerType === 'tke'">
                        <apply-host class="mr10"
                            theme="primary"
                            :cluster-id="localClusterId"
                            :is-backfill="true" />
                    </template>
                </template>
                <bcs-dropdown-menu :disabled="!selections.length"
                    class="mr10"
                    v-authority="{
                        clickable: webAnnotations.perms[localClusterId]
                            && webAnnotations.perms[localClusterId].cluster_manage,
                        actionId: 'cluster_manage',
                        resourceName: curSelectedCluster.clusterName,
                        disablePerms: true,
                        permCtx: {
                            project_id: curProject.project_id,
                            cluster_id: localClusterId
                        }
                    }">
                    <div class="dropdown-trigger-btn" slot="dropdown-trigger">
                        <span>{{$t('批量')}}</span>
                        <i class="bk-icon icon-angle-down"></i>
                    </div>
                    <ul class="bk-dropdown-list" slot="dropdown-content">
                        <li @click="handleBatchEnableNodes">{{$t('允许调度')}}</li>
                        <li @click="handleBatchStopNodes">{{$t('停止调度')}}</li>
                        <li @click="handleBatchReAddNodes">{{$t('重新添加')}}</li>
                        <div style="width: 100px; height:32px;" v-bk-tooltips="{ content: $t('注：IP状态为停止调度才能做POD迁移操作'), disabled: !podDisabled, placement: 'top' }">
                            <li :disabled="podDisabled" @click="handleBatchPodScheduler">{{$t('Pod迁移')}}</li>
                        </div>
                        <li @click="handleBatchSetLabels">{{$t('设置标签')}}</li>
                        <li @click="handleBatchDeleteNodes">{{$t('删除')}}</li>
                        <!-- <li>{{$t('导出')}}</li> -->
                    </ul>
                </bcs-dropdown-menu>
                <bcs-dropdown-menu ref="copyDropdownRef">
                    <div class="dropdown-trigger-btn" slot="dropdown-trigger">
                        <span>{{$t('复制')}}</span>
                        <i class="bk-icon icon-angle-down"></i>
                    </div>
                    <ul class="bk-dropdown-list" slot="dropdown-content">
                        <li :disabled="item.disabled"
                            v-for="item in ipCopyList" :key="item.id"
                            @click="handleCopy(item)"
                        >
                            {{item.name}}
                        </li>
                    </ul>
                </bcs-dropdown-menu>
            </div>
            <div class="right">
                <ClusterSelect class="mr10"
                    v-model="localClusterId"
                    @change="handleClusterChange"
                    v-if="!hideClusterSelect && !isSingleCluster"
                ></ClusterSelect>
                <bcs-search-select
                    clearable
                    class="search-select"
                    :data="searchSelectData"
                    :show-condition="false"
                    :show-popover-tag-change="false"
                    :placeholder="$t('搜索IP，标签、状态')"
                    v-model="searchSelectValue"
                    @change="searchSelectChange"
                    @clear="handleClearSearchSelect">
                </bcs-search-select>
            </div>
        </div>
        <!-- 节点列表 -->
        <div :class="{ 'cluster-node-wrapper': nodeMenu }">
            <bcs-table class="mt20"
                :outer-border="false"
                :size="tableSetting.size"
                :data="curPageData"
                ref="tableRef"
                :key="tableKey"
                v-bkloading="{ isLoading: tableLoading }"
                @filter-change="handleFilterChange"
            >
                <template #prepend>
                    <transition name="fade">
                        <div class="selection-tips" v-if="selectType !== CheckType.Uncheck">
                            <i18n path="已选 {num} 条">
                                <span place="num" class="tips-num">{{selections.length}}</span>
                            </i18n>
                            <bk-button
                                ext-cls="tips-btn"
                                text
                                v-if="selectType === CheckType.AcrossChecked"
                                @click="handleClearSelection">
                                {{ $t('取消选择所有数据') }}
                            </bk-button>
                            <bk-button
                                ext-cls="tips-btn"
                                text
                                v-else
                                @click="handleSelectionAll">
                                <i18n path="选择所有 {num} 条">
                                    <span place="num" class="tips-num">{{pagination.count}}</span>
                                </i18n>
                            </bk-button>
                        </div>
                    </transition>
                </template>
                <bcs-table-column
                    :render-header="renderSelection"
                    width="70"
                    :resizable="false"
                >
                    <template #default="{ row }">
                        <bcs-checkbox
                            :checked="selections.some(item => item.inner_ip === row.inner_ip)"
                            @change="(value) => handleRowCheckChange(value, row)"
                        ></bcs-checkbox>
                    </template>
                </bcs-table-column>
                <bcs-table-column :label="$t('内网IP')" prop="inner_ip" width="120" sortable>
                    <template #default="{ row }">
                        <bcs-button
                            :disabled="['INITIALIZATION', 'DELETING'].includes(row.status)"
                            text
                            v-authority="{
                                clickable: webAnnotations.perms[localClusterId]
                                    && webAnnotations.perms[localClusterId].cluster_view,
                                actionId: 'cluster_view',
                                resourceName: curSelectedCluster.clusterName,
                                disablePerms: true,
                                permCtx: {
                                    project_id: curProject.project_id,
                                    cluster_id: localClusterId
                                }
                            }"
                            @click="handleGoOverview(row)"
                        >
                            {{ row.inner_ip }}
                        </bcs-button>
                    </template>
                </bcs-table-column>
                <bcs-table-column
                    :label="$t('状态')"
                    :filters="filtersDataSource.status"
                    :filtered-value="filteredValue.status"
                    width="120"
                    column-key="status"
                    prop="status">
                    <template #default="{ row }">
                        <LoadingIcon
                            v-if="['INITIALIZATION', 'DELETING'].includes(row.status)"
                        >
                            {{ nodeStatusMap[row.status.toLowerCase()] }}
                        </LoadingIcon>
                        <StatusIcon :status="row.status"
                            :status-color-map="nodeStatusColorMap"
                            v-else
                        >
                            {{ nodeStatusMap[row.status.toLowerCase()] }}
                        </StatusIcon>
                    </template>
                </bcs-table-column>
                <bcs-table-column
                    :label="$t('所属集群')"
                    prop="cluster_name"
                    key="cluster_name"
                    v-if="isColumnRender('cluster_name')"
                ></bcs-table-column>
                <bcs-table-column
                    :label="$t('容器数量')"
                    width="100"
                    align="right"
                    prop="container_count"
                    key="container_count"
                    v-if="isColumnRender('container_count')">
                    <template #default="{ row }">
                        {{ nodeMetric[row.inner_ip] ? nodeMetric[row.inner_ip].container_count : '--'}}
                    </template>
                </bcs-table-column>
                <bcs-table-column
                    :label="$t('Pod数量')"
                    width="100"
                    align="right"
                    prop="pod_count"
                    key="pod_count"
                    v-if="isColumnRender('pod_count')">
                    <template #default="{ row }">
                        {{ nodeMetric[row.inner_ip] ? nodeMetric[row.inner_ip].pod_count : '--'}}
                    </template>
                </bcs-table-column>
                <bcs-table-column min-width="200" :label="$t('标签')" key="source_type" v-if="isColumnRender('source_type')">
                    <template #default="{ row }">
                        <span v-if="!row.labels || !Object.keys(row.labels).length">--</span>
                        <bcs-popover v-else :delay="300" placement="left" class="popover">
                            <div class="row-label">
                                <span class="label" v-for="key in Object.keys(row.labels)" :key="key">
                                    {{ `${key}=${row.labels[key]}` }}
                                </span>
                            </div>
                            <template slot="content">
                                <div class="labels-tips">
                                    <div v-for="key in Object.keys(row.labels)" :key="key">
                                        <span>{{ `${key}=${row.labels[key]}` }}</span>
                                    </div>
                                </div>
                            </template>
                        </bcs-popover>
                    </template>
                </bcs-table-column>
                <bcs-table-column min-width="200" :label="$t('污点')" key="taint" v-if="isColumnRender('taint')">
                    <template #default="{ row }">
                        <span v-if="!row.taints || !row.taints.length">--</span>
                        <bcs-popover v-else :delay="300" placement="left" class="popover">
                            <div class="row-label">
                                <span class="label" v-for="(taint, index) in row.taints" :key="index">
                                    {{
                                        `${taint.key}=${taint.value && taint.effect
                                            ? taint.value + ' : ' + taint.effect
                                            : taint.value || taint.effect}`
                                    }}
                                </span>
                            </div>
                            <template slot="content">
                                <div class="labels-tips">
                                    <div class="label" v-for="(taint, index) in row.taints" :key="index">
                                        {{
                                            `${taint.key}=${taint.value && taint.effect
                                                ? taint.value + ' : ' + taint.effect
                                                : taint.value || taint.effect}`
                                        }}
                                    </div>
                                </div>
                            </template>
                        </bcs-popover>
                    </template>
                </bcs-table-column>
                <bcs-table-column
                    :label="$t('CPU')"
                    :sort-method="(pre, next) => sortMethod(pre, next, 'cpu_usage')"
                    key="cpu_usage"
                    sortable
                    align="center"
                    v-if="isColumnRender('cpu_usage')"
                >
                    <template #default="{ row }">
                        <LoadingCell v-if="!nodeMetric[row.inner_ip]"></LoadingCell>
                        <RingCell :percent="nodeMetric[row.inner_ip].cpu_usage"
                            fill-color="#3ede78"
                            v-else
                        ></RingCell>
                    </template>
                </bcs-table-column>
                <bcs-table-column
                    :label="$t('内存')"
                    :sort-method="(pre, next) => sortMethod(pre, next, 'memory_usage')"
                    key="memory_usage"
                    sortable
                    align="center"
                    v-if="isColumnRender('memory_usage')"
                >
                    <template #default="{ row }">
                        <LoadingCell v-if="!nodeMetric[row.inner_ip]"></LoadingCell>
                        <RingCell :percent="nodeMetric[row.inner_ip].memory_usage"
                            fill-color="#3a84ff"
                            v-else
                        ></RingCell>
                    </template>
                </bcs-table-column>
                <bcs-table-column
                    :label="$t('磁盘')"
                    :sort-method="(pre, next) => sortMethod(pre, next, 'disk_usage')"
                    key="disk_usage"
                    sortable
                    align="center"
                    v-if="isColumnRender('disk_usage')"
                >
                    <template #default="{ row }">
                        <LoadingCell v-if="!nodeMetric[row.inner_ip]"></LoadingCell>
                        <RingCell :percent="nodeMetric[row.inner_ip].disk_usage"
                            fill-color="#853cff"
                            v-else
                        ></RingCell>
                    </template>
                </bcs-table-column>
                <bcs-table-column
                    :label="$t('磁盘IO')"
                    :sort-method="(pre, next) => sortMethod(pre, next, 'diskio_usage')"
                    key="diskio_usage"
                    sortable
                    align="center"
                    v-if="isColumnRender('diskio_usage')"
                >
                    <template #default="{ row }">
                        <LoadingCell v-if="!nodeMetric[row.inner_ip]"></LoadingCell>
                        <RingCell :percent="nodeMetric[row.inner_ip].diskio_usage"
                            fill-color="#853cff"
                            v-else
                        ></RingCell>
                    </template>
                </bcs-table-column>
                <bcs-table-column :label="$t('操作')" width="260">
                    <template #default="{ row }">
                        <div class="node-operate-wrapper"
                            v-authority="{
                                clickable: webAnnotations.perms[localClusterId]
                                    && webAnnotations.perms[localClusterId].cluster_manage,
                                actionId: 'cluster_manage',
                                resourceName: curSelectedCluster.clusterName,
                                disablePerms: true,
                                permCtx: {
                                    project_id: curProject.project_id,
                                    cluster_id: localClusterId
                                }
                            }">
                            <template v-if="row.status === 'RUNNING'">
                                <bk-button text class="mr10" @click="handleSetLabel(row)">{{$t('设置标签')}}</bk-button>
                                <bk-button text class="mr10" @click="handleSetTaint(row)">{{$t('设置污点')}}</bk-button>
                            </template>
                            <bk-button text @click="handleStopNode(row)" v-if="row.status === 'RUNNING'">
                                {{ $t('停止调度') }}
                            </bk-button>
                            <bk-button text
                                v-if="['INITIALIZATION', 'DELETING', 'REMOVE-FAILURE', 'ADD-FAILURE'].includes(row.status)"
                                @click="handleShowLog(row)"
                            >
                                {{$t('查看日志')}}
                            </bk-button>
                            <template v-if="row.status === 'REMOVABLE'">
                                <bk-button text @click="handleEnableNode(row)">
                                    {{ $t('允许调度') }}
                                </bk-button>
                                <bk-button text class="ml10" @click="handleSchedulerNode(row)">
                                    {{ $t('pod迁移') }}
                                </bk-button>
                            </template>
                            <bk-button text class="ml10"
                                v-if="['REMOVE-FAILURE', 'ADD-FAILURE', 'REMOVABLE', 'NOTREADY'].includes(row.status)"
                                @click="handleDeleteNode(row)"
                            >
                                {{ $t('删除') }}
                            </bk-button>
                            <bk-button text class="ml10"
                                v-if="['REMOVE-FAILURE', 'ADD-FAILURE'].includes(row.status)"
                                @click="handleRetry(row)"
                            >{{ $t('重试') }}</bk-button>
                        </div>
                    </template>
                </bcs-table-column>
                <bcs-table-column type="setting">
                    <bcs-table-setting-content
                        :fields="tableSetting.fields"
                        :selected="tableSetting.selectedFields"
                        :max="tableSetting.max"
                        :size="tableSetting.size"
                        @setting-change="handleSettingChange"
                    >
                    </bcs-table-setting-content>
                </bcs-table-column>
            </bcs-table>
            <bcs-pagination
                class="pagination"
                :limit="pagination.limit"
                :count="pagination.count"
                :current="pagination.current"
                align="right"
                show-total-count
                show-selection-count
                :selection-count="selections.length"
                @change="pageChange"
                @limit-change="pageSizeChange">
            </bcs-pagination>
        </div>
        <!-- 设置标签 -->
        <bcs-sideslider
            :is-show.sync="setLabelConf.isShow"
            :width="750"
            :quick-close="false"
        >
            <template #header>
                <span>{{setLabelConf.title}}</span>
                <span class="sideslider-tips">{{$t('标签有助于整理你的资源（如 env:prod）')}}</span>
            </template>
            <template #content>
                <KeyValue class="key-value-content"
                    :model-value="setLabelConf.data"
                    :loading="setLabelConf.btnLoading"
                    :key-desc="setLabelConf.keyDesc"
                    v-bkloading="{ isLoading: setLabelConf.loading }"
                    @cancel="handleLabelEditCancel"
                    @confirm="handleLabelEditConfirm"
                ></KeyValue>
            </template>
        </bcs-sideslider>
        <!-- 设置污点 -->
        <bcs-sideslider
            :is-show.sync="taintConfig.isShow"
            :title="$t('设置污点')"
            :width="750"
            :quick-close="false"
        >
            <template #content>
                <TaintContent :cluster-id="localClusterId"
                    :nodes="taintConfig.nodes"
                    @confirm="handleConfirmTaintDialog"
                    @cancel="handleHideTaintDialog"
                ></TaintContent>
            </template>
        </bcs-sideslider>
        <!-- 查看日志 -->
        <bk-sideslider
            :is-show.sync="logSideDialogConf.isShow"
            :title="logSideDialogConf.title"
            :width="640"
            @hidden="closeLog"
            :quick-close="true">
            <div slot="content">
                <div class="log-wrapper" v-bkloading="{ isLoading: logSideDialogConf.loading }">
                    <bk-table :data="logSideDialogConf.taskData">
                        <bk-table-column :label="$t('步骤')" prop="taskName" width="160"></bk-table-column>
                        <bk-table-column :label="$t('状态')" prop="status" width="120">
                            <template #default="{ row }">
                                <div class="log-wrapper-status" v-if="row.status === 'RUNNING'">
                                    <loading-cell :style="{ left: 0, margin: 0 }"
                                        :ext-cls="['bk-spin-loading-mini', 'bk-spin-loading-danger']"></loading-cell>
                                    <span class="ml5">{{ $t('运行中') }}</span>
                                </div>
                                <StatusIcon :status="row.status" :status-color-map="taskStatusColorMap" v-else>
                                    {{ taskStatusTextMap[row.status.toLowerCase()] }}
                                </StatusIcon>
                            </template>
                        </bk-table-column>
                        <bk-table-column min-width="120" :label="$t('内容')" prop="message"></bk-table-column>
                    </bk-table>
                </div>
            </div>
        </bk-sideslider>
        <!-- 确认删除 -->
        <tip-dialog
            ref="removeNodeDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :show-close="false"
            :sub-title="$t('此操作无法撤回，请确认：')"
            :tips="$t('注意: 节点状态以集群中的状态为准；点击【删除】后，节点状态可能会仍然处于不可调度')"
            :check-list="deleteNodeNoticeList"
            :confirm-btn-text="$t('确定')"
            :confirming-btn-text="$t('删除中...')"
            :canceling-btn-text="$t('取消')"
            :confirm-callback="confirmDelNode"
            :cancel-callback="cancelDelNode">
        </tip-dialog>
        <!-- IP选择器 -->
        <IpSelector v-model="showIpSelector" @confirm="chooseServer"></IpSelector>
    </div>
</template>
<script lang="ts">
    import { defineComponent, ref, PropType, onMounted, watch, set, computed } from '@vue/composition-api'
    import StatusIcon from '@/views/dashboard/common/status-icon'
    import ClusterSelect from '@/components/cluster-selector/cluster-select.vue'
    import LoadingIcon from '@/components/loading-icon.vue'
    import { nodeStatusColorMap, nodeStatusMap, taskStatusTextMap, taskStatusColorMap } from '@/common/constant'
    import useNode from './use-node'
    import useTableSetting from './use-table-setting'
    import usePage from '@/views/dashboard/common/use-page'
    import useTableSearchSelect, { ISearchSelectData } from './use-table-search-select'
    import useTableAcrossCheck from './use-table-across-check'
    import { CheckType } from '@/components/across-check.vue'
    import RingCell from '@/views/cluster/ring-cell.vue'
    import LoadingCell from '@/views/cluster/loading-cell.vue'
    import { copyText } from '@/common/util'
    import useInterval from '@/views/dashboard/common/use-interval'
    import KeyValue, { IData } from '@/components/key-value.vue'
    import TaintContent from './taint.vue'
    import tipDialog from '@/components/tip-dialog/index.vue'
    import ApplyHost from '@/views/cluster/apply-host.vue'
    import { TranslateResult } from 'vue-i18n'
    import IpSelector from '@/components/ip-selector/selector-dialog.vue'
    import useDefaultClusterId from './use-default-clusterId'

    export default defineComponent({
        name: 'node',
        components: {
            StatusIcon,
            LoadingIcon,
            ClusterSelect,
            RingCell,
            LoadingCell,
            KeyValue,
            TaintContent,
            tipDialog,
            ApplyHost,
            IpSelector
        },
        props: {
            selectedFields: {
                type: Array as PropType<Array<string>>,
                default: ['source_type', 'taint']
            },
            clusterId: {
                type: String,
                default: ''
            },
            nodeMenu: {
                type: Boolean,
                default: true
            },
            hideClusterSelect: {
                type: Boolean,
                default: false
            }
        },
        setup (props, ctx) {
            const { $i18n, $router, $bkMessage, $store, $bkInfo } = ctx.root
            const webAnnotations = computed(() => {
                return $store.state.cluster.clusterWebAnnotations
            })
            const curProject = computed(() => {
                return $store.state.curProject
            })
            // 表格设置字段配置
            const fields = [
                {
                    id: 'cluster_name',
                    label: $i18n.t('所属集群')
                },
                {
                    id: 'container_count',
                    label: $i18n.t('容器数量')
                },
                {
                    id: 'pod_count',
                    label: $i18n.t('Pod数量')
                },
                {
                    id: 'source_type',
                    label: $i18n.t('标签')
                },
                {
                    id: 'taint',
                    label: $i18n.t('污点')
                },
                {
                    id: 'cpu_usage',
                    label: 'CPU'
                },
                {
                    id: 'memory_usage',
                    label: $i18n.t('内存')
                },
                {
                    id: 'disk_usage',
                    label: $i18n.t('磁盘')
                },
                {
                    id: 'diskio_usage',
                    label: $i18n.t('磁盘IO')
                }
            ]
            // 表格表头搜索项配置
            const filtersDataSource = ref({
                status: Object.keys(nodeStatusMap).map(key => ({
                    text: nodeStatusMap[key],
                    value: key
                }))
            })
            // 表格搜索项选中值
            const filteredValue = ref<Record<string, string[]>>({
                status: []
            })
            // searchSelect数据源配置
            const searchSelectDataSource = computed<ISearchSelectData[]>(() => {
                return [
                    {
                        name: $i18n.t('IP地址'),
                        id: 'inner_ip',
                        placeholder: $i18n.t('多IP用换行符分割')
                    },
                    {
                        name: $i18n.t('状态'),
                        id: 'status',
                        multiable: true,
                        children: Object.keys(nodeStatusMap).map(key => ({
                            id: key,
                            name: nodeStatusMap[key]
                        }))
                    },
                    {
                        name: $i18n.t('标签'),
                        id: 'labels',
                        multiable: true,
                        children: labels.value.map(label => ({
                            id: label,
                            name: label
                        }))
                    }
                ]
            })
            // 表格搜索联动
            const {
                tableKey,
                searchSelectData,
                searchSelectValue,
                handleFilterChange,
                handleSearchSelectChange,
                handleClearSearchSelect
                // handleResetSearchSelect
            } = useTableSearchSelect({
                searchSelectDataSource,
                filteredValue
            })
            const searchSelectChange = (list) => {
                handleResetCheckStatus()
                handleSearchSelectChange(list)
            }

            watch(searchSelectValue, () => {
                handleResetPage()
            })

            const {
                tableSetting,
                handleSettingChange,
                isColumnRender
            } = useTableSetting(fields, props.selectedFields)

            const sortMethod = (pre, next, prop) => {
                const preNumber = parseFloat(nodeMetric.value[pre.inner_ip]?.[prop] || 0)
                const nextNumber = parseFloat(nodeMetric.value[next.inner_ip]?.[prop] || 0)
                if (preNumber > nextNumber) {
                    return -1
                } else if (preNumber < nextNumber) {
                    return 1
                }
                return 0
            }
            
            const {
                getNodeList,
                getTaskData,
                toggleNodeDispatch,
                schedulerNode,
                deleteNode,
                addNode,
                getNodeOverview,
                batchToggleNodeDispatch
            } = useNode()
            
            const tableLoading = ref(false)
            // 初始化当前集群ID
            const { defaultClusterId, clusterList, isSingleCluster } = useDefaultClusterId()
            const localClusterId = ref(props.clusterId || defaultClusterId.value || '')
            const curSelectedCluster = computed(() => {
                return clusterList.value.find(item => item.clusterID === localClusterId.value) || {}
            })
           
            // 全量表格数据
            const tableData = ref<any[]>([])
            
            const parseSearchSelectValue = computed(() => {
                const searchValues: { id: string; value: Set<any> }[] = []
                searchSelectValue.value.forEach(item => {
                    let tmp: string[] = []
                    if (item.id === 'inner_ip') {
                        item.values.forEach(v => {
                            tmp.push(...v.id.replace(/\s+/g, "").split('|'))
                        })
                    } else {
                        tmp = item.values.map(v => v.id)
                    }
                    searchValues.push({
                        id: item.id,
                        value: new Set(tmp)
                    })
                })
                return searchValues
            })
            // 过滤后的表格数据
            const filterTableData = computed(() => {
                if (!searchSelectValue.value.length) return tableData.value
                
                return tableData.value.filter(row => {
                    return parseSearchSelectValue.value.some(item => {
                        if (!row[item.id]) return false
                        if (item.id === 'labels') {
                            return Object.keys(row[item.id]).some(key => {
                                return item.value.has(`${key}=${row[item.id][key]}`)
                            })
                        }
                        return item.value.has(row[item.id].toLowerCase())
                    })
                })
            })
            // 分页后的表格数据
            const {
                curPageData,
                pagination,
                pageChange,
                pageSizeChange,
                handleResetPage,
                pageConf
            } = usePage(filterTableData)

            // 搜索标签
            const labels = computed(() => {
                const data: string[] = []
                tableData.value.forEach(item => {
                    Object.keys(item.labels || {}).forEach(key => {
                        const label = `${key}=${item.labels[key]}`
                        const index = data.indexOf(label)
                        index === -1 && data.push(label)
                    })
                })
                return data
            })
            const {
                selectType,
                selections,
                handleResetCheckStatus,
                renderSelection,
                handleRowCheckChange,
                handleSelectionAll,
                handleClearSelection
            } = useTableAcrossCheck({ tableData: filterTableData, curPageData })

            const handleGoOverview = (row) => {
                $router.push({
                    name: 'clusterNodeOverview',
                    params: {
                        nodeId: row.inner_ip,
                        clusterId: row.cluster_id
                    }
                })
            }

            const copyDropdownRef = ref<any>(null)
            const ipCopyList = computed(() => {
                return [
                    {
                        id: 'checked',
                        name: $i18n.t('勾选IP'),
                        disabled: !selections.value.length
                    },
                    {
                        id: 'currentPage',
                        name: $i18n.t('当前页IP'),
                        disabled: !curPageData.value.length
                    },
                    {
                        id: 'allPage',
                        name: $i18n.t('所有IP'),
                        disabled: !filterTableData.value.length
                    }
                ]
            })
            // IP复制
            const handleCopy = (item) => {
                if (item.disabled) return

                let ipData: string[] = []
                switch (item.id) {
                    case 'checked':
                        ipData = selections.value.map(data => data.inner_ip)
                        break
                    case 'currentPage':
                        ipData = curPageData.value.map(data => data.inner_ip)
                        break
                    case 'allPage':
                        ipData = filterTableData.value.map(data => data.inner_ip)
                        break
                }
                copyText(ipData.join('\n'))
                $bkMessage({
                    theme: 'success',
                    message: $i18n.t('成功复制 {num} 个IP', { num: ipData.length })
                })
                copyDropdownRef.value && copyDropdownRef.value.hide()
            }

            // 设置污点
            const taintConfig = ref<{
                isShow: boolean;
                nodes: any[];
            }>({
                    isShow: false,
                    nodes: []
                })
            const handleSetTaint = (row) => {
                taintConfig.value.isShow = true
                taintConfig.value.nodes = [row]
            }
            const handleConfirmTaintDialog = () => {
                handleGetNodeData()
                handleHideTaintDialog()
            }
            const handleHideTaintDialog = () => {
                taintConfig.value.isShow = false
                taintConfig.value.nodes = []
            }

            // 设置标签（批量设置标签的交互有点奇怪，后续优化）
            const setLabelConf = ref<{
                isShow: boolean;
                loading: boolean;
                btnLoading: boolean;
                keyDesc: any;
                rows: any[];
                data: IData[];
                title: string;
            }>({
                    isShow: false,
                    loading: false,
                    btnLoading: false,
                    keyDesc: '',
                    rows: [],
                    data: [],
                    title: ''
                })
            const handleSetLabel = async (row) => {
                setLabelConf.value.isShow = true
                const rows = Array.isArray(row) ? row : [row]
                setLabelConf.value.loading = true
                const data = await $store.dispatch('cluster/fetchK8sNodeLabels', {
                    $clusterId: localClusterId.value,
                    node_name_list: rows.map(item => item.name)
                })
                setLabelConf.value.loading = false
                // 批量设置时暂时只展示相同Key的项
                const labelArr = rows.reduce<any[]>((pre, row) => {
                    const label = data[row.inner_ip]
                    Object.keys(label).forEach(key => {
                        const index = pre.findIndex(item => item.key === key)
                        if (index > -1) {
                            pre[index].value = ''
                            pre[index].repeat += 1
                            pre[index].placeholder = $i18n.t('不变')
                        } else {
                            pre.push({
                                key,
                                value: label[key],
                                repeat: 1
                            })
                        }
                    })
                    return pre
                }, []).filter(item => item.repeat === rows.length)
                
                set(setLabelConf, 'value', Object.assign(setLabelConf.value, {
                    data: labelArr,
                    rows,
                    title: rows.length > 1 ? $i18n.t('批量设置标签') : $i18n.t('设置标签'),
                    keyDesc: rows.length > 1 ? $i18n.t('批量设置只展示相同Key的标签') : ''
                }))
            }
            const handleLabelEditCancel = () => {
                set(setLabelConf, 'value', Object.assign(setLabelConf.value, {
                    isShow: false,
                    keyDesc: '',
                    rows: [],
                    title: '',
                    data: {}
                }))
            }
            const mergeLaels = (_originLabels, _newLabels) => {
                const originLabels = JSON.parse(JSON.stringify(_originLabels))
                const newLabels = JSON.parse(JSON.stringify(_newLabels))
                // 批量编辑
                if (setLabelConf.value.rows.length > 1) {
                    const oldLabels = setLabelConf.value.data
                    oldLabels.forEach(item => {
                        if (!newLabels.hasOwnProperty(item.key)) {
                            // 删除去除的key
                            delete originLabels[item.key]
                        } else if (!newLabels[item.key]) {
                            // 未修改的Key保持不变
                            delete newLabels[item.key]
                        }
                    })
                    return Object.assign({}, originLabels, newLabels)
                } else {
                    return newLabels
                }
            }
            const handleLabelEditConfirm = async (labels) => {
                setLabelConf.value.btnLoading = true
             
                const result = await $store.dispatch('cluster/setK8sNodeLabels', {
                    // eslint-disable-next-line camelcase
                    $clusterId: localClusterId.value,
                    node_label_list: setLabelConf.value.rows.map(item => {
                        return {
                            node_name: item.name,
                            labels: mergeLaels(item.labels, labels)
                        }
                    })
                }).then(() => true).catch(() => false)
                setLabelConf.value.btnLoading = false
                if (result) {
                    handleLabelEditCancel()
                    handleResetCheckStatus()
                    handleGetNodeData()
                }
            }

            // 弹窗二次确认
            const bkComfirmInfo = ({
                title, subTitle, callback
                }: {
                title: TranslateResult;
                subTitle: TranslateResult;
                callback: Function;
            }) => {
                $bkInfo({
                    type: 'warning',
                    clsName: 'custom-info-confirm',
                    subTitle,
                    title,
                    defaultInfo: true,
                    confirmFn: async (vm) => {
                        await callback()
                    }
                })
            }

            // 停止调度
            const handleStopNode = (row) => {
                bkComfirmInfo({
                    title: $i18n.t('确认对节点 {ip} 停止调度', { ip: row.inner_ip }),
                    subTitle: $i18n.t('如果有使用Ingress及LoadBalancer类型的Service，节点停止调度后，Service Controller会剔除LB到nodePort的映射'),
                    callback: async () => {
                        const result = await toggleNodeDispatch({
                            clusterId: row.cluster_id,
                            nodeName: [row.name],
                            status: 'REMOVABLE'
                        })
                        result && handleGetNodeData()
                    }
                })
            }
            // 允许调度
            const handleEnableNode = (row) => {
                bkComfirmInfo({
                    title: $i18n.t('确认允许调度'),
                    subTitle: $i18n.t('确认对节点 {ip} 允许调度', { ip: row.inner_ip }),
                    callback: async () => {
                        const result = await toggleNodeDispatch({
                            clusterId: row.cluster_id,
                            nodeName: [row.name],
                            status: 'RUNNING'
                        })
                        result && handleGetNodeData()
                    }
                })
            }
            // Pod迁移
            const handleSchedulerNode = (row) => {
                bkComfirmInfo({
                    title: $i18n.t('确认Pod迁移'),
                    subTitle: $i18n.t('确认要对节点 {ip} 上的Pod进行迁移', { ip: row.inner_ip }),
                    callback: async () => {
                        await schedulerNode({
                            clusterId: row.cluster_id,
                            nodeIps: [row.inner_ip]
                        })
                        // result && handleGetNodeData()
                    }
                })
            }
            // 节点删除
            const removeNodeDialog = ref<any>(null)
            const deleteNodeNoticeList = ref([
                {
                    id: 1,
                    text: $i18n.t('当前节点上正在运行的容器会被调度到其它可用节点'),
                    isChecked: false
                },
                {
                    id: 2,
                    text: $i18n.t('清理容器服务系统组件'),
                    isChecked: false
                },
                {
                    id: 3,
                    text: $i18n.t('节点删除后服务器如不再使用请尽快回收，避免产生不必要的成本'),
                    isChecked: false
                }
            ])
            const curDeleteRows = ref<any[]>([])
            const handleDeleteNode = async (row) => {
                curDeleteRows.value = [row]
                removeNodeDialog.value.title = $i18n.t(`确认要删除节点【{innerIp}】？`, {
                    innerIp: row.inner_ip
                })
                removeNodeDialog.value.show()
            }
            const cancelDelNode = () => {
                curDeleteRows.value = []
            }
            const delNode = async (clusterId: string, nodeIps: string[]) => {
                const result = await deleteNode({
                    clusterId,
                    nodeIps
                })
                result && handleGetNodeData()
                handleResetPage()
                handleResetCheckStatus()
            }
            const confirmDelNode = async () => {
                removeNodeDialog.value.isConfirming = true
                await delNode(localClusterId.value, curDeleteRows.value.map(item => item.inner_ip))
                removeNodeDialog.value.isConfirming = false
            }
            const addClusterNode = async (clusterId: string, nodeIps: string[]) => {
                const result = await addNode({
                    clusterId,
                    nodeIps
                })
                result && handleGetNodeData()
            }
            // 节点重试
            const handleRetry = (row) => {
                if (row.status === 'REMOVE-FAILURE') {
                    // 删除重试
                    bkComfirmInfo({
                        title: $i18n.t('确认删除节点', { ip: row.inner_ip }),
                        subTitle: row.inner_ip,
                        callback: async () => {
                            await delNode(row.cluster_id, [row.inner_ip])
                        }
                    })
                } else if (row.status === 'ADD-FAILURE') {
                    // 添加重试
                    bkComfirmInfo({
                        title: $i18n.t('确认添加节点', { ip: row.inner_ip }),
                        subTitle: row.inner_ip,
                        callback: async () => {
                            await addClusterNode(row.cluster_id, [row.inner_ip])
                        }
                    })
                }
            }
            // 批量允许调度
            const handleBatchEnableNodes = () => {
                if (!selections.value.length) return

                bkComfirmInfo({
                    title: $i18n.t('请确认是否批量允许调度'),
                    subTitle: $i18n.t('请确认是否允许 {ip} 等 {num} 个IP调度', {
                        ip: selections.value[0].inner_ip,
                        num: selections.value.length
                    }),
                    callback: async () => {
                        const result = await batchToggleNodeDispatch({
                            clusterId: localClusterId.value,
                            nodeNameList: selections.value.map(item => item.name),
                            status: 'RUNNING'
                        })
                        result && handleGetNodeData()
                    }
                })
            }
            // 批量停止调度
            const handleBatchStopNodes = () => {
                if (!selections.value.length) return

                bkComfirmInfo({
                    title: $i18n.t('请确认是否批量停止调度'),
                    subTitle: $i18n.t('请确认是否停止 {ip} 等 {num} 个IP调度', {
                        ip: selections.value[0].inner_ip,
                        num: selections.value.length
                    }),
                    callback: async () => {
                        const result = await batchToggleNodeDispatch({
                            clusterId: localClusterId.value,
                            nodeNameList: selections.value.map(item => item.name),
                            status: 'REMOVABLE'
                        })
                        result && handleGetNodeData()
                    }
                })
            }
            // 重新添加节点
            const handleBatchReAddNodes = () => {
                if (!selections.value.length) return

                bkComfirmInfo({
                    title: $i18n.t('确认重新添加节点'),
                    subTitle: $i18n.t('请确认是否对 {ip} 等 {num} 个IP进行操作系统初始化和安装容器服务相关组件操作', {
                        num: selections.value.length,
                        ip: selections.value[0].inner_ip }),
                    callback: async () => {
                        await addClusterNode(localClusterId.value, selections.value.map(item => item.inner_ip))
                    }
                })
            }
            // 批量设置标签
            const handleBatchSetLabels = () => {
                if (!selections.value.length) return

                handleSetLabel(selections.value)
            }
            // 批量删除节点
            const handleBatchDeleteNodes = () => {
                bkComfirmInfo({
                    title: $i18n.t('确认删除节点'),
                    subTitle: $i18n.t('确认是否删除 {ip} 等 {num} 个节点', {
                        num: selections.value.length,
                        ip: selections.value[0].inner_ip
                    }),
                    callback: async () => {
                        await delNode(localClusterId.value, selections.value.map(item => item.inner_ip))
                    }
                })
            }
            // 批量Pod迁移
            const handleBatchPodScheduler = () => {
                if (!selections.value.length) return

                if (selections.value.length > 10) {
                    $bkMessage({
                        theme: 'warning',
                        message: $i18n.t('最多只能批量迁移10个节点')
                    })
                    return
                }
                bkComfirmInfo({
                    title: $i18n.t('确认Pod迁移'),
                    subTitle: $i18n.t('确认要对 {ip} 等 {num} 个节点上的Pod进行迁移', {
                        num: selections.value.length,
                        ip: selections.value[0].inner_ip
                    }),
                    callback: async () => {
                        await schedulerNode({
                            clusterId: localClusterId.value,
                            nodeIps: selections.value.map(item => item.inner_ip)
                        })
                        // result && handleGetNodeData()
                    }
                })
            }
            // 添加节点
            const showIpSelector = ref(false)
            const handleAddNode = () => {
                showIpSelector.value = true
            }
            const chooseServer = (data) => {
                if (!data.length) return
                bkComfirmInfo({
                    title: $i18n.t('确认添加节点'),
                    subTitle: $i18n.t('请确认是否对 {ip} 等 {num} 个IP进行操作系统初始化和安装容器服务相关组件操作', {
                        ip: data[0].bk_host_innerip,
                        num: data.length
                    }),
                    callback: async () => {
                        await addClusterNode(localClusterId.value, data.map(item => item.bk_host_innerip))
                        showIpSelector.value = false
                    }
                })
            }
            // 查看日志
            const logSideDialogConf = ref({
                isShow: false,
                title: '',
                taskData: [],
                row: null,
                loading: false
            })
            const handleShowLog = async (row) => {
                logSideDialogConf.value.isShow = true
                logSideDialogConf.value.title = row.inner_ip
                logSideDialogConf.value.row = row
                logSideDialogConf.value.loading = true
                await getTaskTableData(row)
                logSideDialogConf.value.loading = false
            }
            const getTaskTableData = async (row) => {
                const { taskData, latestTask } = await getTaskData({
                    clusterId: row.cluster_id,
                    nodeIP: row.inner_ip
                })
                logSideDialogConf.value.taskData = taskData || []
                if (['RUNNING', 'INITIALZING'].includes(latestTask?.status)) {
                    logIntervalStart()
                } else {
                    logIntervalStop()
                }
            }
            const { stop: logIntervalStop, start: logIntervalStart } = useInterval(async () => {
                const row = logSideDialogConf.value.row as any
                if (!row) {
                    logIntervalStop()
                    return
                }
                const { taskData, latestTask } = await getTaskData({
                    clusterId: row.cluster_id,
                    nodeIP: row.inner_ip
                })
                logSideDialogConf.value.taskData = taskData || []
                if (!['RUNNING', 'INITIALZING'].includes(latestTask?.status)) {
                    logIntervalStop()
                }
            }, 5000)
            const closeLog = () => {
                logSideDialogConf.value.row = null
                logIntervalStop()
            }
            
            // 获取节点指标
            const nodeMetric = ref({})
            const handleGetNodeOverview = async () => {
                const data = curPageData.value.filter(item => !nodeMetric.value[item.inner_ip])
                const promiseList: Promise<any>[] = []
                for (let i = 0; i < data.length; i++) {
                    (function (item) {
                        promiseList.push(
                            getNodeOverview({
                                nodeIP: item.inner_ip,
                                clusterId: item.cluster_id
                            }).then(data => {
                                set(nodeMetric.value, item.inner_ip, data)
                            })
                        )
                    })(data[i])
                }
                await Promise.all(promiseList)
            }
            watch(curPageData, async () => {
                await handleGetNodeOverview()
            })
            // 切换集群
            const handleGetNodeData = async () => {
                tableLoading.value = true
                tableData.value = await getNodeList(localClusterId.value)
                tableLoading.value = false
            }

            const handleClusterChange = async () => {
                stop()
                await handleGetNodeData()
                handleResetPage()
                handleResetCheckStatus()
                if (tableData.value.length) {
                    start()
                }
            }
            const podDisabled = computed(() => {
                return !selections.value.every(select => select.status === 'REMOVABLE')
            })

            watch(pageConf, () => {
                // 非跨页全选在分页变更时重置selections
                if (![
                    CheckType.AcrossChecked,
                    CheckType.HalfAcrossChecked
                ].includes(selectType.value)) {
                    handleResetCheckStatus()
                }
            })

            const { stop, start } = useInterval(async () => {
                tableData.value = await getNodeList(localClusterId.value)
            }, 5000)

            const nodesCount = computed(() => {
                return tableData.value.length + Object.keys(curSelectedCluster.value?.master || {}).length
            })
            const getCidrIpNum = (cidr) => {
                const mask = Number(cidr.split('/')[1] || 0)
                if (mask <= 0) {
                    return 0
                }
                return Math.pow(2, 32 - mask)
            }
            const remainNodesCount = computed(() => {
                const { cidrStep, maxNodePodNum, maxServiceNum, clusterIPv4CIDR, multiClusterCIDR = [] } = curSelectedCluster.value?.networkSettings || {}
                let totalCidrStep = 0
                if (multiClusterCIDR.length < 3) {
                    totalCidrStep = (5 - multiClusterCIDR.length) * cidrStep + multiClusterCIDR.reduce((pre, cidr) => {
                        pre += getCidrIpNum(cidr)
                        return pre
                    }, 0)
                } else {
                    totalCidrStep = [clusterIPv4CIDR, ...multiClusterCIDR].reduce((pre, cidr) => {
                        pre += getCidrIpNum(cidr)
                        return pre
                    }, 0)
                }
                return Math.floor((totalCidrStep - maxServiceNum - maxNodePodNum * nodesCount.value) / maxNodePodNum)
            })
            
            onMounted(async () => {
                await handleGetNodeData()
                if (tableData.value.length) {
                    start()
                }
            })
            return {
                nodesCount,
                remainNodesCount,
                isSingleCluster,
                curSelectedCluster,
                taskStatusTextMap,
                taskStatusColorMap,
                logSideDialogConf,
                showIpSelector,
                removeNodeDialog,
                deleteNodeNoticeList,
                searchSelectData,
                searchSelectValue,
                tableKey,
                filtersDataSource,
                filteredValue,
                selectType,
                selections,
                pagination,
                curPageData,
                nodeStatusColorMap,
                nodeStatusMap,
                tableSetting,
                taintConfig,
                setLabelConf,
                tableLoading,
                localClusterId,
                CheckType,
                nodeMetric,
                copyDropdownRef,
                ipCopyList,
                renderSelection,
                pageChange,
                pageSizeChange,
                isColumnRender,
                handleSettingChange,
                handleGetNodeData,
                handleSelectionAll,
                handleClearSelection,
                handleRowCheckChange,
                handleFilterChange,
                searchSelectChange,
                handleClearSearchSelect,
                handleGoOverview,
                handleCopy,
                handleSetLabel,
                sortMethod,
                handleLabelEditCancel,
                handleLabelEditConfirm,
                handleConfirmTaintDialog,
                handleHideTaintDialog,
                handleSetTaint,
                handleEnableNode,
                handleStopNode,
                handleDeleteNode,
                handleSchedulerNode,
                confirmDelNode,
                cancelDelNode,
                handleRetry,
                handleBatchEnableNodes,
                handleBatchStopNodes,
                handleBatchReAddNodes,
                handleBatchSetLabels,
                handleBatchDeleteNodes,
                handleAddNode,
                chooseServer,
                handleClusterChange,
                handleShowLog,
                closeLog,
                handleBatchPodScheduler,
                podDisabled,
                webAnnotations,
                curProject
            }
        }
    })
</script>
<style lang="postcss" scoped>
.cluster-node-wrapper {
    margin: 0 20px;
    border: 1px solid #dfe0e5;
    border-top: none;
}
.cluster-node-tip {
    margin: 20px;
    .num {
        font-weight: 700;
    }
}
.cluster-node-operate {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
    .left {
        display: flex;
        align-items: center;
        .add-node {
            min-width: 120px;
        }
    }
    .right {
        display: flex;
        .search-select {
            width: 400px;
        }
    }
}
/deep/ .dropdown-trigger-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid #c4c6cc;
    border-radius: 2px;
    padding: 0 15px;
    height: 32px;
    min-width: 68px;
    font-size: 14px;
    .icon-angle-down {
        font-size: 22px;
    }
}
/deep/ .bk-dropdown-list {
    min-width: 100px;
    max-height: unset;
    li {
        display: block;
        height: 32px;
        line-height: 33px;
        padding: 0 16px;
        color: #63656e;
        font-size: 14px;
        cursor: pointer;
        &:hover {
            background-color: #eaf3ff;
            color: #3a84ff;
        }
        &[disabled] {
            pointer-events: none;
            color: #c3cdd7;
            cursor: not-allowed;
        }
    }
}
/deep/ .bk-table-column-setting {
    border-top: 1px solid #dfe0e5;
    .bk-tooltip-ref {
        display: flex;
        align-items: center;
        justify-content: center;
    }
}
.tips-enter-active {
  transition: opacity .5s;
}
.tips-enter,
.tips-leave-to {
  opacity: 0;
}
.selection-tips {
    height: 30px;
    background: #ebecf0;
    display: flex;
    align-items: center;
    justify-content: center;
    .tips-num {
        font-weight: bold;
    }
    .tips-btn {
        font-size: 12px;
        margin-left: 5px;
    }
}
.pagination {
    padding: 14px 16px;
    height: 60px;
    background: #fff;
    >>> .bk-page-total-count {
        color: #63656e;
    }
}
.row-label {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    height: 22px;
    .label {
        display: inline-block;
        align-self: center;
        background: #f0f1f5;
        border-radius: 2px;
        line-height: 22px;
        padding: 0 8px;
        margin-right: 6px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap
    }
}
.sideslider-tips {
    color: #c3cdd7;
    font-size: 12px;
    font-weight: normal;
    margin-left: 10px;
}
.key-value-content {
    padding: 30px;
}
.log-wrapper {
    padding: 20px;
}
.labels-tips {
    max-height: 260px;
    overflow: auto;
}
.popover {
    width: 100%;
    /deep/ .bk-tooltip-ref {
        display: block;
    }
}
</style>
