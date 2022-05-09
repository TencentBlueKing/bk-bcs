<template>
    <ipSelector
        ref="selectorRef"
        v-bkloading="{ isLoading }"
        :panels="panels"
        :height="height"
        :active.sync="active"
        :preview-data="previewData"
        :get-default-data="handleGetDefaultData"
        :get-search-table-data="handleGetSearchTableData"
        :static-table-config="staticTableConfig"
        :custom-input-table-config="staticTableConfig"
        :get-default-selections="getDefaultSelections"
        :get-row-disabled-status="getRowDisabledStatus"
        :get-row-tips-content="getRowTipsContent"
        :preview-operate-list="previewOperateList"
        :default-expand-level="1"
        :search-data-options="searchDataOptions"
        :tree-data-options="treeDataOptions"
        :preview-width="240"
        :left-panel-width="300"
        :preview-title="$t('已选节点')"
        :default-active-name="['nodes']"
        :enable-search-panel="false"
        :enable-tree-filter="true"
        :static-table-placeholder="$t('请输入IP，多IP可使用空格分隔')"
        :across-page="true"
        ip-key="bk_host_innerip"
        ellipsis-direction="ltr"
        :default-accurate="true"
        :default-selected-node="defaultSelectedNode"
        @check-change="handleCheckChange"
        @remove-node="handleRemoveNode"
        @menu-click="handleMenuClick"
        @search-selection-change="handleCheckChange">
        <template v-slot:collapse-title="{ item }">
            <i18n path="{count} 个IP节点">
                <span place="count" class="preview-count">{{ item.data.length }}</span>
            </i18n>
        </template>
    </ipSelector>
</template>
<script lang="ts">
    /* eslint-disable camelcase */
    import { defineComponent, reactive, toRefs, h, ref, watch } from '@vue/composition-api'
    import { ipSelector, AgentStatus } from './ip-selector'
    import './ip-selector.css'
    import { fetchBizTopo, fetchBizHosts, nodeAvailable } from '@/api/base'
    import { copyText } from '@/common/util'

    export interface ISelectorState {
        isLoading: boolean;
        panels: any[];
        active: string;
        previewData: any[];
        staticTableConfig: any[];
        previewOperateList: any[];
        searchDataOptions: any;
        treeDataOptions: any;
        curTreeNode: any;
    }
    export default defineComponent({
        name: 'ip-selector-bcs',
        components: {
            ipSelector
        },
        props: {
            // 回显IP列表
            ipList: {
                type: Array,
                default: () => ([])
            },
            height: {
                type: Number,
                default: 600
            }
        },
        setup (props, ctx) {
            const { $i18n } = ctx.root
            const statusMap = {
                0: 'terminated',
                1: 'running'
            }
            const textMap = {
                0: $i18n.t('异常'),
                1: $i18n.t('正常')
            }
            const renderIpAgentStatus = (row) => {
                return h(AgentStatus, {
                    props: {
                        type: 2,
                        data: [
                            {
                                status: statusMap[row.agent_alive],
                                display: textMap[row.agent_alive]
                            }
                        ]
                    }
                })
            }
            const state = reactive<ISelectorState>({
                isLoading: false,
                panels: [
                    {
                        name: 'static-topo',
                        label: $i18n.t('静态选择')
                    },
                    {
                        name: 'custom-input',
                        label: $i18n.t('自定义输入')
                    }
                ],
                active: 'static-topo',
                previewData: [],
                staticTableConfig: [
                    {
                        prop: 'bk_host_innerip',
                        label: $i18n.t('内网IP')
                    },
                    {
                        prop: 'agent_alive',
                        label: $i18n.t('Agent状态'),
                        render: renderIpAgentStatus
                    },
                    {
                        prop: 'idc_unit_name',
                        label: $i18n.t('机房')
                    },
                    {
                        prop: 'svr_device_class',
                        label: $i18n.t('机型')
                    }
                ],
                previewOperateList: [
                    {
                        id: 'removeAll',
                        label: $i18n.t('移除所有')
                    },
                    {
                        id: 'copyIp',
                        label: $i18n.t('复制IP')
                    }
                ],
                searchDataOptions: {},
                treeDataOptions: {
                    idKey: 'id',
                    nameKey: 'bk_inst_name',
                    childrenKey: 'child'
                },
                curTreeNode: null
            })
            const selectorRef = ref<any>(null)

            // 初始化回显列表
            const { ipList } = toRefs(props)
            watch(ipList, () => {
                const groups = state.previewData.find(item => item.id === 'nodes')
                if (groups) {
                    ipList.value.forEach(item => {
                        const index = groups.data.find(data => identityIp(data, item))
                        index === -1 && groups.data.push(item)
                    })
                } else {
                    state.previewData.push({
                        id: 'nodes',
                        data: [...ipList.value],
                        dataNameKey: 'bk_host_innerip'
                    })
                }
            }, { immediate: true })

            // 获取左侧Tree数据
            let treeData: any[] = []
            const defaultSelectedNode = ref()
            const handleSetTreeId = (nodes: any[] = []) => {
                nodes.forEach(node => {
                    node.id = `${node.bk_inst_id}-${node.bk_obj_id}`
                    if (node.child) {
                        handleSetTreeId(node.child)
                    }
                })
            }
            const handleGetDefaultData = async () => {
                if (!treeData.length) {
                    treeData = await fetchBizTopo().catch(() => [])
                    defaultSelectedNode.value = `${treeData[0]?.bk_inst_id}-${treeData[0]?.bk_obj_id}`
                    handleSetTreeId(treeData)
                }
                return treeData
            }
            const nodeAvailableMap = {}
            // 静态表格数据处理
            const handleGetStaticTableData = async (params) => {
                const { selections = [], current, limit, tableKeyword, accurate } = params
                const bizHostsParams: any = {
                    limit,
                    offset: (current - 1) * limit,
                    fuzzy: !accurate,
                    ip_list: tableKeyword.split(' ').filter(ip => !!ip)
                }
                const [node] = selections
                state.curTreeNode = node
                if (!node) return { total: 0, data: [] }

                if (node.bk_obj_id === 'set') {
                    bizHostsParams.set_id = node.bk_inst_id
                } else if (node.bk_obj_id === 'module') {
                    bizHostsParams.module_id = node.bk_inst_id
                }
                const data = await fetchBizHosts(bizHostsParams).catch(() => ({ results: [] }))
                const nodeAvailableData = await nodeAvailable({
                    innerIPs: data.results.map(item => {
                        return item.bk_host_innerip
                    })
                })
                Object.assign(nodeAvailableMap, nodeAvailableData)
                return {
                    total: data.count || 0,
                    data: data.results
                }
            }
            // 自定义输入表格数据处理
            const handleGetCustomInputTableData = async (params) => {
                const { accurate, ipList = [] } = params
                const bizHostsParams: any = {
                    desire_all_data: true,
                    fuzzy: !accurate,
                    ip_list: ipList
                }
                const data = await fetchBizHosts(bizHostsParams).catch(() => ({ results: [] }))
                return {
                    total: data.count || 0,
                    data: data.results
                }
            }
            const handleGetSearchTableData = async (params) => {
                if (state.active === 'static-topo') {
                    return handleGetStaticTableData(params)
                } else if (state.active === 'custom-input') {
                    return handleGetCustomInputTableData(params)
                }
            }
            // 判断两个IP节点是否相同
            const identityIp = (current, origin) => {
                return current.bk_cloud_id === origin.bk_cloud_id && current.bk_host_innerip === origin.bk_host_innerip
            }
            // 重新获取表格勾选状态
            const resetTableCheckedStatus = () => {
                selectorRef.value && selectorRef.value.handleGetDefaultSelections()
            }
            // 预览菜单点击事件
            const handleMenuClick = ({ menu }) => {
                if (menu.id === 'removeAll') {
                    state.previewData = []
                    resetTableCheckedStatus()
                    handleChange()
                } else if (menu.id === 'copyIp') {
                    const group = state.previewData.find(data => data.id === 'nodes')
                    const ipList = group?.data.map(item => item.bk_host_innerip) || []
                    copyText(ipList.join('\n'))
                    ctx.root.$bkMessage({
                        theme: 'success',
                        message: $i18n.t('成功复制IP {number} 个', { number: ipList.length })
                    })
                }
            }
            // 预览面板单个节点移除事件
            const handleRemoveNode = ({ child, item }) => {
                const group = state.previewData.find(data => data.id === item.id)
                const index = group?.data.findIndex(data => identityIp(data, child))

                index > -1 && group.data.splice(index, 1)
                resetTableCheckedStatus()
                handleChange()
            }
            // 本页选择
            const handleCurrentPageChecked = (data) => {
                const { selections = [], excludeData = [] } = data
                const index = state.previewData.findIndex((item: any) => item.id === 'nodes')
                if (index > -1) {
                    const { data } = state.previewData[index]
                    selections.forEach((select) => {
                        const index = data.findIndex(data => identityIp(data, select))

                        index === -1 && data.push(select)
                    })
                    excludeData.forEach((exclude) => {
                        const index = data.findIndex(data => identityIp(data, exclude))

                        index > -1 && data.splice(index, 1)
                    })
                } else {
                    state.previewData.push({
                        id: 'nodes',
                        data: [...selections],
                        dataNameKey: 'bk_host_innerip'
                    })
                }
            }
            // 静态选择跨页全选
            const handleStaticTopoAllChecked = async (data) => {
                const { excludeData = [], checkValue } = data
                if (checkValue === 1) {
                    excludeData.forEach((exclude) => {
                        const index = state.previewData.findIndex(data => identityIp(data, exclude))

                        index > -1 && state.previewData.splice(index, 1)
                    })
                } else if (checkValue === 2) {
                    state.isLoading = true
                    const params: any = {
                        desire_all_data: true
                    }
                    if (state.curTreeNode?.bk_obj_id === 'set') {
                        params.set_id = state.curTreeNode.bk_inst_id
                    } else if (state.curTreeNode?.bk_obj_id === 'module') {
                        params.module_id = state.curTreeNode.bk_inst_id
                    }
                    const data = await fetchBizHosts(params).catch(() => ({ results: [] }))
                    const ipList = data.results.filter(item => !getRowDisabledStatus(item))
                    state.previewData = [
                        {
                            id: 'nodes',
                            data: ipList,
                            dataNameKey: 'bk_host_innerip'
                        }
                    ]
                    state.isLoading = false
                } else if (checkValue === 0) {
                    state.previewData = []
                }
            }
            // 表格勾选事件
            const handleCheckChange = async (data) => {
                if (data?.checkType === 'current' || state.active === 'custom-input') {
                    handleCurrentPageChecked(data)
                } else if (data?.checkType === 'all') {
                    await handleStaticTopoAllChecked(data)
                }
                // 统一抛出change事件
                handleChange()
            }
            // 获取当前行的勾选状态
            const getDefaultSelections = (row) => {
                const group = state.previewData.find(data => data.id === 'nodes')
                return group?.data.some(data => identityIp(data, row))
            }
            // 表格表格当前行禁用状态
            const getRowDisabledStatus = (row) => {
                return !row.is_valid || nodeAvailableMap[row.bk_host_innerip]?.isExist
            }
            // 获取表格当前行tips内容
            const getRowTipsContent = (row) => {
                let tips: any = ''
                if (!row.is_valid) {
                    tips = $i18n.t('Docker机不允许使用')
                } else if (nodeAvailableMap[row.bk_host_innerip]?.isExist) {
                    const { clusterName = '', clusterID = '' } = nodeAvailableMap[row.bk_host_innerip]
                    tips = $i18n.t('IP已被 {name}{id} 占用', {
                        name: clusterName,
                        id: clusterID ? ` (${clusterID}) ` : ''
                    })
                }
                return tips
            }
            // 统一抛出change事件
            const handleChange = () => {
                const group = state.previewData.find(data => data.id === 'nodes')
                ctx.emit('change', group?.data || [])
            }
            // 获取IP节点数据
            const handleGetData = () => {
                const group = state.previewData.find(data => data.id === 'nodes')
                return group?.data || []
            }

            return {
                ...toRefs(state),
                selectorRef,
                handleGetDefaultData,
                handleGetSearchTableData,
                handleCheckChange,
                handleMenuClick,
                getDefaultSelections,
                handleRemoveNode,
                handleChange,
                getRowDisabledStatus,
                getRowTipsContent,
                handleGetData,
                defaultSelectedNode
            }
        }
    })
</script>
<style lang="postcss" scoped>
/deep/ .preview-count {
    color: #3a84ff;
    font-weight: 700;
    padding: 0 2px;
    font-size: 12px;
}
</style>
