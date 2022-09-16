<template>
    <div>
        <ContentHeader :title="$t('日志采集规则')" :desc="cluster.clusterName"></ContentHeader>
        <div class="biz-content-wrapper">
            <div class="content-header mb20">
                <bcs-button icon="plus" theme="primary" @click="handleCreateLog">{{$t('新建规则')}}</bcs-button>
                <div class="right">
                    <bcs-select class="mw248"
                        :placeholder="$t('命名空间')"
                        searchable
                        :loading="nsLoading"
                        v-model="searchData.namespace">
                        <bcs-option v-for="item in namespaceList"
                            :key="item.name"
                            :id="item.name"
                            :name="item.name">
                        </bcs-option>
                    </bcs-select>
                    <bcs-select class="ml10 mw248"
                        :placeholder="$t('应用类型')"
                        v-model="searchData.kind">
                        <bcs-option v-for="name in kinds"
                            :key="name" :id="name"
                            :name="name"></bcs-option>
                    </bcs-select>
                    <bcs-input class="ml10 mw300"
                        :placeholder="$t('应用名')"
                        right-icon="bk-icon icon-search"
                        v-model="searchData.name"></bcs-input>
                </div>
            </div>
            <bcs-table :data="curPageData"
                :pagination="pagination"
                v-bkloading="{ isLoading: loading }"
                size="medium"
                @page-change="pageChange"
                @page-limit-change="pageSizeChange">
                <bcs-table-column :label="$t('名称')">
                    <template #default="{ row }">
                        <span v-bk-tooltips="{ disabled: !row.deleted, content: $t('采集规则不存在') }">
                            <bcs-button text
                                :disabled="row.deleted"
                                @click="handleShowDetail(row)"
                            >
                                <span class="bcs-ellipsis">{{row.config_name}}</span>
                            </bcs-button>
                        </span>
                    </template>
                </bcs-table-column>
                <bcs-table-column :label="$t('集群')">
                    <template #default>
                        {{cluster.clusterName}}
                    </template>
                </bcs-table-column>
                <bcs-table-column :label="$t('命名空间')" prop="namespace"></bcs-table-column>
                <bcs-table-column :label="$t('日志源')" width="100">
                    <template #default="{ row }">
                        {{logSourceTypeMap[row.log_source_type]}}
                    </template>
                </bcs-table-column>
                <bcs-table-column :label="$t('选择器')" show-overflow-tooltip min-width="300">
                    <template #default="{ row }">
                        <template v-if="!row.deleted">
                            <!-- 选择容器 -->
                            <template v-if="row.log_source_type === 'selected_containers'">
                                <p>{{$t('类型')}}：{{row.workload.kind || '--'}}</p>
                                <p>{{$t('名称')}}：{{row.workload.name || '--'}}</p>
                            </template>
                            <!-- 选择标签 -->
                            <template v-else-if="row.log_source_type === 'selected_labels'">
                                <div class="row-label" v-if="Object.keys(row.selector.match_labels || {}).length">
                                    <div class="mb5">{{$t('匹配标签')}}：</div>
                                    <div class="row-label-content">
                                        <span v-for="key in Object.keys(row.selector.match_labels || {})"
                                            :key="key"
                                            class="tag mr5 mb5">
                                            {{`${key} : ${row.selector.match_labels[key]}`}}
                                        </span>
                                    </div>
                                </div>
                                <div class="row-label" v-if="(row.selector.match_expressions || []).length">
                                    <div class="mb5">{{$t('匹配表达式')}}：</div>
                                    <div class="row-label-content">
                                        <span v-for="item in row.selector.match_expressions"
                                            :key="item.key"
                                            class="tag mr5 mb5">
                                            {{item.key || '--'}}
                                            <span>{{item.operator || '--'}}</span>
                                            {{item.values || '--'}}
                                        </span>
                                    </div>
                                </div>
                            </template>
                            <!-- 所有容器 -->
                            <template v-else-if="row.log_source_type === 'all_containers'">
                                --
                            </template>
                        </template>
                        <template v-else>--</template>
                    </template>
                </bcs-table-column>
                <bcs-table-column :label="$t('日志信息')" show-overflow-tooltip min-width="300">
                    <template #default="{ row }">
                        <template v-if="!row.deleted">
                            <template v-if="row.log_source_type === 'selected_containers'">
                                <div v-for="(item, index) in row.workload.container_confs"
                                    :key="index"
                                    class="bcs-ellipsis">
                                    {{`${item.name}: ${(item.log_paths || []).join(';') || '--'}`}}
                                </div>
                            </template>
                            <template v-else-if="row.log_source_type === 'selected_labels'">
                                {{row.selector.log_paths.join(';') || '--'}}
                            </template>
                            <template v-else-if="row.log_source_type === 'all_containers'">
                                --
                            </template>
                        </template>
                        <template v-else>--</template>
                    </template>
                </bcs-table-column>
                <bcs-table-column :label="$t('更新人')" prop="updator" width="120"></bcs-table-column>
                <bcs-table-column :label="$t('更新时间')" prop="updated"></bcs-table-column>
                <bcs-table-column :label="$t('操作')" width="120">
                    <template #default="{ row }">
                        <span v-bk-tooltips="{ disabled: !row.deleted, content: $t('采集规则不存在') }">
                            <bcs-button text
                                :disabled="row.deleted"
                                @click="handleUpdateLog(row)"
                            >
                                {{$t('更新')}}
                            </bcs-button>
                        </span>
                        <bcs-button text class="ml10" @click="handleDeleteLog(row)">{{$t('删除')}}</bcs-button>
                    </template>
                </bcs-table-column>
            </bcs-table>
        </div>
        <!-- 详情 -->
        <bcs-sideslider :is-show.sync="showDetail"
            :quick-close="true"
            :width="800"
            :title="currentRow ? currentRow.config_name : ''">
            <template #content>
                <LogDetail
                    :log-source-type-map="logSourceTypeMap"
                    :data="currentRow"
                    :cluster="cluster">
                </LogDetail>
            </template>
        </bcs-sideslider>
        <!-- 编辑 -->
        <bcs-sideslider :is-show.sync="showEdit"
            :width="800"
            :title="currentRow ? $t('编辑规则') : $t('新建规则')">
            <template #content>
                <LogListEdit class="pd30"
                    :namespace-list="namespaceList"
                    :kinds="kinds"
                    :id="currentRow ? currentRow.config_id : null"
                    :cluster-id="clusterId"
                    @cancel="showEdit = false"
                    @confirm="handleEditConfirm">
                </LogListEdit>
            </template>
        </bcs-sideslider>
    </div>
</template>
<script lang="ts">
    import { computed, defineComponent, onMounted, ref } from '@vue/composition-api'
    import ContentHeader from '@/views/content-header.vue'
    import $store from '@/store'
    import usePage from '@/views/dashboard/common/use-page'
    import $i18n from '@/i18n/i18n-setup'
    import LogListEdit from './log-list-edit.vue'
    import LogDetail from './log-detail.vue'

    export default defineComponent({
        components: { ContentHeader, LogListEdit, LogDetail },
        props: {
            clusterId: {
                type: String,
                default: ''
            }
        },
        setup (props, ctx) {
            const { $bkInfo, $bkMessage } = ctx.root
            const cluster = computed(() => {
                return ($store.state as any).cluster.clusterList
                    .find(item => item.clusterID === props.clusterId) || {}
            })
            const showDetail = ref(false)
            const showEdit = ref(false)
            const logSourceTypeMap = {
                'selected_containers': $i18n.t('指定容器'),
                'selected_labels': $i18n.t('指定标签'),
                'all_containers': $i18n.t('所有容器')
            }
            const kinds = ref(['Deployment', 'DaemonSet', 'Job', 'StatefulSet', 'GameStatefulSet'])
            
            // 日志规则列表
            const logList = ref<any[]>([])
            const loading = ref(false)
            const handleGetLogList = async () => {
                loading.value = true
                logList.value = await $store.dispatch('crdcontroller/logCollectList', {
                    $clusterId: props.clusterId
                })
                loading.value = false
            }
            const searchData = ref({
                namespace: '',
                kind: '',
                name: ''
            })
            const filterLogList = computed(() => {
                // 搜索
                return logList.value.filter(item => {
                    return (!searchData.value.namespace || item.namespace === searchData.value.namespace)
                        && (!searchData.value.name || item.config_name.includes(searchData.value.name))
                        && (!searchData.value.kind || item.workload?.kind === searchData.value.kind)
                })
            })
            const {
                pageChange,
                pageSizeChange,
                curPageData,
                pagination
            } = usePage(filterLogList)

            // 命名空间
            const namespaceList = ref([])
            const curProjectId = computed(() => {
                return $store.state.curProjectId
            })
            const nsLoading = ref(false)
            const handleGetNameSpaceList = async () => {
                nsLoading.value = true
                const { data = [] } = await $store.dispatch('crdcontroller/getNameSpaceListByCluster', {
                    projectId: curProjectId.value,
                    clusterId: props.clusterId
                })
                namespaceList.value = data
                nsLoading.value = false
            }

            // 操作
            const currentRow = ref(null)
            const handleCreateLog = () => {
                currentRow.value = null
                showEdit.value = true
            }
            const handleShowDetail = (row) => {
                currentRow.value = row
                showDetail.value = true
            }
            const handleUpdateLog = (row) => {
                currentRow.value = row
                showEdit.value = true
            }
            const handleDeleteLog = (row) => {
                $bkInfo({
                    type: 'warning',
                    clsName: 'custom-info-confirm',
                    title: $i18n.t('确认删除规则'),
                    subTitle: row.config_name,
                    defaultInfo: true,
                    confirmFn: async () => {
                        const result = await $store.dispatch('crdcontroller/deleteLogCollect', {
                            $configId: row.config_id,
                            $clusterId: props.clusterId
                        })
                        if (result) {
                            $bkMessage({
                                theme: 'success',
                                message: $i18n.t('删除成功')
                            })
                            handleGetLogList()
                        }
                    }
                })
            }
            const handleEditConfirm = () => {
                showEdit.value = false
                handleGetLogList()
            }

            onMounted(() => {
                handleGetLogList()
                handleGetNameSpaceList()
            })
            return {
                currentRow,
                searchData,
                cluster,
                showEdit,
                showDetail,
                namespaceList,
                kinds,
                loading,
                nsLoading,
                pageChange,
                pageSizeChange,
                curPageData,
                pagination,
                handleShowDetail,
                logSourceTypeMap,
                handleUpdateLog,
                handleDeleteLog,
                handleCreateLog,
                handleEditConfirm
            }
        }
    })
</script>
<style lang="postcss" scoped>
.content-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    .right {
        display: flex;
        background: #fff;
    }
    .mw248 {
        min-width: 248px;
    }
    .mw300 {
        min-width: 300px;
    }
}
.pd30 {
    padding: 30px;
}
.row-label {
    overflow: hidden;
    margin: 5px 0;
    &-content {
        display: flex;
        flex-wrap: wrap;
        .tag {
            background-color: rgba(151,155,165,.1);
            border-color: rgba(220,222,229,.6);
            color: #63656e;
            display: inline-block;
            font-size: 12px;
            padding: 0 10px;
            min-height: 22px;
            cursor: default;
            box-sizing: border-box;
            line-height: 22px;
        }
    }
}
</style>
