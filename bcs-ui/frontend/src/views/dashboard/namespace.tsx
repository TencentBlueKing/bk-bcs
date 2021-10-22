import { defineComponent, computed, ref, watch, onMounted } from '@vue/composition-api'
import DashboardTopActions from './common/dashboard-top-actions'
// import useCluster from './common/use-cluster'
import useInterval from './common/use-interval'
import useNamespace from './common/use-namespace'
import usePage from './common/use-page'
import useSearch from './common/use-search'
import useSubscribe from './common/use-subscribe'
import { sort } from '@/common/util'
import './namespace.css'

export default defineComponent({
    name: 'Namespace',
    components: {
        DashboardTopActions
    },
    setup (props, ctx) {
        const keys = ref(['metadata.name'])

        // 初始化集群列表信息
        // useCluster(ctx)
        // 获取命名空间
        const { namespaceLoading, namespaceData, getNamespaceData } = useNamespace(ctx)

        // 排序
        const sortData = ref({
            prop: '',
            order: ''
        })
        const handleSortChange = (data) => {
            sortData.value = {
                prop: data.prop,
                order: data.order
            }
        }

        // 表格数据
        const tableData = computed(() => {
            const items = JSON.parse(JSON.stringify(namespaceData.value.manifest?.items || []))
            const { prop, order } = sortData.value
            return prop ? sort(items, prop, order) : items
        })
        // resourceVersion
        const resourceVersion = computed(() => {
            return namespaceData.value.manifest?.metadata?.resourceVersion || ''
        })
        // 搜索功能
        const { tableDataMatchSearch, searchValue } = useSearch(tableData, keys)

        // 分页
        const { pagination, curPageData, pageConf, pageChange, pageSizeChange } = usePage(tableDataMatchSearch)
        // 搜索时重置分页
        watch(searchValue, () => {
            pageConf.current = 1
        })

        // 处理额外字段
        const handleExtCol = (row: any, key: string) => {
            const ext = namespaceData.value.manifest_ext[row.metadata?.uid] || {}
            return ext[key] || '--'
        }

        // 订阅事件
        const { initParams, handleSubscribe } = useSubscribe(namespaceData, ctx)
        const { start, stop } = useInterval(handleSubscribe, 5000)

        watch(resourceVersion, (newVersion, oldVersion) => {
            if (newVersion && newVersion !== oldVersion) {
                stop()
                initParams('Namespace', resourceVersion.value)
                resourceVersion.value && start()
            }
        })

        onMounted(() => {
            getNamespaceData()
        })

        return {
            namespaceLoading,
            pagination,
            searchValue,
            curPageData,
            pageChange,
            pageSizeChange,
            handleExtCol,
            handleSortChange
        }
    },
    render () {
        return (
            <div class="biz-content">
                <div class="biz-top-bar">
                    <div class="dashboard-top-title">
                        {this.$t('命名空间')}
                    </div>
                    <DashboardTopActions />
                </div>
                <div class="biz-content-wrapper" v-bkloading={{ isLoading: this.namespaceLoading }}>
                    <bcs-input class="mb20 search-input"
                        right-icon="bk-icon icon-search"
                        placeholder={this.$t('搜索名称')}
                        v-model={this.searchValue}>
                    </bcs-input>
                    <bcs-table data={this.curPageData}
                        pagination={this.pagination}
                        on-page-change={this.pageChange}
                        on-page-limit-change={this.pageSizeChange}
                        on-sort-change={this.handleSortChange}>
                        <bcs-table-column label={this.$t('名称')} sortable prop="metadata.name"
                            scopedSlots={{
                                default: ({ row }: { row: any }) => row.metadata?.name || '--'
                            }}>
                        </bcs-table-column>
                        <bcs-table-column label={this.$t('状态')}
                            scopedSlots={{
                                default: ({ row }: { row: any }) => row.status?.phase || '--'
                            }}>
                        </bcs-table-column>
                        <bcs-table-column label={this.$t('Age')}
                            scopedSlots={{
                                default: ({ row }: { row: any }) => this.handleExtCol(row, 'age')
                            }}>
                        </bcs-table-column>
                    </bcs-table>
                </div>
            </div>
        )
    }
})
