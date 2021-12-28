/* eslint-disable camelcase */
import { defineComponent, computed, ref, watch, onMounted, toRefs } from '@vue/composition-api'
import DashboardTopActions from './dashboard-top-actions'
// import useCluster from './use-cluster'
import useInterval from './use-interval'
import useNamespace from './use-namespace'
import usePage from './use-page'
import useSearch from './use-search'
import useSubscribe, { ISubscribeData, ISubscribeParams } from './use-subscribe'
import useTableData from './use-table-data'
import { sort } from '@/common/util'
import yamljs from 'js-yaml'
import * as ace from '@/components/ace-editor'
import './base-layout.css'
import fullScreen from '@/directives/full-screen'
import { CUR_SELECT_NAMESPACE, CUR_SELECT_CRD } from '@/common/constant'

export default defineComponent({
    name: 'BaseLayout',
    components: {
        ace
    },
    directives: {
        'full-screen': fullScreen
    },
    props: {
        title: {
            type: String,
            default: '',
            required: true
        },
        // 父分类（crd类型的需要特殊处理），eg: workloads、networks（注意复数）
        type: {
            type: String,
            default: '',
            required: true
        },
        // 子分类，eg: deployments、ingresses
        category: {
            type: String,
            default: '',
            required: true
        },
        // 轮询时类型（type为crd时，kind仅作为资源详情展示的title用），eg: Deployment、Ingress（注意首字母大写）
        kind: {
            type: String,
            default: '',
            required: true
        },
        // 是否显示命名空间（不展示的话不会发送获取命名空间列表的请求）
        showNameSpace: {
            type: Boolean,
            default: true
        },
        // 是否显示创建资源按钮
        showCreate: {
            type: Boolean,
            default: true
        },
        // 默认CRD值
        defaultCrd: {
            type: String,
            default: ''
        },
        // 是否显示crd下拉菜单
        showCrd: {
            type: Boolean,
            default: false
        },
        // 是否显示总览和yaml切换的tab
        showDetailTab: {
            type: Boolean,
            default: true
        },
        // 默认展示详情标签
        defaultActiveDetailType: {
            type: String,
            default: 'overview'
        }
    },
    setup (props, ctx) {
        const { $router, $i18n, $bkInfo, $store, $bkMessage } = ctx.root
        const { type, category, kind, showNameSpace, showCrd, defaultActiveDetailType, defaultCrd } = toRefs(props)

        // crd
        const currentCrd = ref(defaultCrd.value || sessionStorage.getItem(CUR_SELECT_CRD) || '')
        const crdLoading = ref(false)
        // crd 数据
        const crdData = ref<ISubscribeData|null>(null)
        // crd 列表
        const crdList = computed(() => {
            return crdData.value?.manifest?.items || []
        })
        const currentCrdExt = computed(() => {
            const item = crdList.value.find(item => item.metadata.name === currentCrd.value)
            return crdData.value?.manifest_ext?.[item?.metadata?.uid] || {}
        })
        // 未选择crd时提示
        const crdTips = computed(() => {
            return type.value === 'crd' && !currentCrd.value ? $i18n.t('请选择CRD') : ''
        })
        // 自定义资源的kind类型是根据选择的crd确定的
        const crdKind = computed(() => {
            return currentCrdExt.value.kind
        })
        // 自定义CRD（GameStatefulSets、GameDeployments、CustomObjects）
        const customCrd = computed(() => {
            return (type.value === 'crd' && kind.value !== 'CustomResourceDefinition')
        })
        const handleGetCrdData = async () => {
            crdLoading.value = true
            const res = await fetchCRDData()
            crdData.value = res.data
            crdLoading.value = false
            // 校验初始化的crd值是否正确
            const crd = crdData.value?.manifest?.items?.find(item => item.metadata.name === currentCrd.value)
            if (!crd) {
                currentCrd.value = ''
                sessionStorage.removeItem(CUR_SELECT_CRD)
            }
        }
        const handleCrdChange = async (value) => {
            sessionStorage.setItem(CUR_SELECT_CRD, value)
            handleGetTableData()
        }
        const renderCrdHeader = (h, { column }) => {
            const additionalData = additionalColumns.value.find(item => item.name === column.label)
            return h('span', {
                directives: [
                    {
                        name: 'bk-tooltips',
                        value: {
                            content: additionalData?.description || column.label,
                            placement: 'top',
                            boundary: 'window'
                        }
                    }
                ]
            }, [column.label])
        }
        const getJsonPathValue = (row, path: string) => {
            const keys = path.split('.').filter(str => !!str)
            return keys.reduce((data, key) => {
                if (typeof data === 'object') {
                    return data?.[key]
                }
                return data
            }, row)
        }

        // 命名空间
        const namespaceDisabled = computed(() => {
            const { scope } = currentCrdExt.value
            return type.value === 'crd' && scope && scope !== 'Namespaced'
        })
        // 获取命名空间
        const { namespaceLoading, namespaceValue, namespaceList, getNamespaceData } = useNamespace(ctx)

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
        const {
            isLoading,
            data,
            webAnnotations,
            handleFetchList,
            fetchCRDData,
            handleFetchCustomResourceList
        } = useTableData(ctx)

        // 获取表格数据
        const handleGetTableData = async (subscribe = true) => {
            // 获取表格数据
            if (type.value === 'crd') {
                // crd scope 为Namespaced时需要传递命名空间（gamedeployments、gamestatefulsets两个特殊的资源）
                const customResourceNamespace = currentCrdExt.value?.scope === 'Namespaced'
                    || ['gamedeployments.tkex.tencent.com', 'gamestatefulsets.tkex.tencent.com'].includes(defaultCrd.value)
                    ? namespaceValue.value
                    : ''
                // crd 界面无需传当前crd参数
                const crd = customCrd.value ? currentCrd.value : ''
                await handleFetchCustomResourceList(crd, category.value, customResourceNamespace)
            } else {
                await handleFetchList(type.value, category.value, namespaceValue.value)
            }

            // 重新订阅（获取表格数据之后，resourceVersion可能会变更）
            subscribe && handleStartSubscribe()
        }

        const pagePerms = computed(() => { // 界面权限
            return {
                create: webAnnotations.value.perms?.page?.create_btn || {},
                delete: webAnnotations.value.perms?.page?.delete_btn || {},
                update: webAnnotations.value.perms?.page?.update_btn || {}
            }
        })
        const additionalColumns = computed(() => { // 动态表格字段
            return webAnnotations.value.additional_columns || []
        })
        const tableData = computed(() => {
            const items = JSON.parse(JSON.stringify(data.value.manifest.items || []))
            const { prop, order } = sortData.value
            return prop ? sort(items, prop, order) : items
        })
        const resourceVersion = computed(() => {
            return data.value.manifest?.metadata?.resourceVersion || ''
        })

        // 模糊搜索功能
        const keys = ref(['metadata.name']) // 模糊搜索字段
        const { tableDataMatchSearch, searchValue } = useSearch(tableData, keys)

        const handleNamespaceSelected = (value) => {
            sessionStorage.setItem(CUR_SELECT_NAMESPACE, value)
            handleGetTableData()
        }

        // 分页
        const { pagination, curPageData, pageConf, pageChange, pageSizeChange } = usePage(tableDataMatchSearch)
        // 搜索时重置分页
        watch([searchValue, namespaceValue, currentCrd], () => {
            pageConf.current = 1
        })

        // 订阅事件
        const { initParams, handleSubscribe } = useSubscribe(data, ctx)
        const { start, stop } = useInterval(handleSubscribe, 5000)
        const subscribeKind = computed(() => {
            // 自定义资源（非CustomResourceDefinition类型的crd）的kind是根据选择的crd动态获取的，不能取props的kind值
            return kind.value === 'CustomObject' ? crdKind.value : kind.value
        })

        // GameDeployment、GameStatefulSet apiVersion前端固定
        const apiVersion = computed(() => {
            return ['GameDeployment', 'GameStatefulSet'].includes(kind.value) ? 'tkex.tencent.com/v1alpha1' : currentCrdExt.value.api_version
        })
        const handleStartSubscribe = () => {
            // 自定义的CRD订阅时必须传apiVersion
            // eslint-disable-next-line @typescript-eslint/camelcase
            if (!subscribeKind.value || !resourceVersion.value || (customCrd.value && !apiVersion.value)) return

            stop()
            const params: ISubscribeParams = {
                kind: subscribeKind.value,
                resource_version: resourceVersion.value,
                api_version: apiVersion.value,
                namespace: namespaceValue.value

            }
            if (customCrd.value) {
                params.crd_name = currentCrd.value
            }
            initParams(params)
            start()
        }

        // 获取额外字段方法
        const handleGetExtData = (uid: string, ext?: string) => {
            const extData = data.value.manifest_ext[uid] || {}
            return ext ? extData[ext] : extData
        }

        // 跳转详情界面
        const gotoDetail = (row) => {
            $router.push({
                name: 'dashboardWorkloadDetail',
                params: {
                    category: category.value,
                    name: row.metadata.name,
                    namespace: row.metadata.namespace
                },
                query: {
                    kind: subscribeKind.value
                }
            })
        }

        // 详情侧栏
        const showDetailPanel = ref(false)
        // 当前详情行数据
        const curDetailRow = ref<any>({
            data: {},
            extData: {}
        })
        // 侧栏展示类型
        const detailType = ref({
            active: defaultActiveDetailType.value,
            list: [
                {
                    id: 'overview',
                    name: window.i18n.t('总览')
                },
                {
                    id: 'yaml',
                    name: 'YAML'
                }
            ]
        })
        // 显示侧栏详情
        const handleShowDetail = (row) => {
            curDetailRow.value.data = row
            curDetailRow.value.extData = handleGetExtData(row.metadata.uid)
            showDetailPanel.value = true
        }
        // 切换详情类型
        const handleChangeDetailType = (type) => {
            detailType.value.active = type
        }
        // 重置详情类型
        watch(showDetailPanel, () => {
            handleChangeDetailType(defaultActiveDetailType.value)
        })
        // yaml内容
        const yaml = computed(() => {
            return yamljs.dump(curDetailRow.value.data || {})
        })
        // 创建资源
        const handleCreateResource = () => {
            $router.push({
                name: 'dashboardResourceUpdate',
                params: {
                    defaultShowExample: (type.value !== 'crd') as any
                },
                query: {
                    type: type.value,
                    category: category.value,
                    kind: kind.value,
                    crd: currentCrd.value
                }
            })
        }
        // 更新资源
        const handleUpdateResource = (row) => {
            const { name, namespace } = row.metadata || {}
            $router.push({
                name: 'dashboardResourceUpdate',
                params: {
                    namespace,
                    name
                },
                query: {
                    type: type.value,
                    category: category.value,
                    kind: type.value === 'crd' ? kind.value : row.kind,
                    crd: currentCrd.value
                }
            })
        }
        // 删除资源
        const handleDeleteResource = (row) => {
            const { name, namespace } = row.metadata || {}
            $bkInfo({
                type: 'warning',
                clsName: 'custom-info-confirm',
                title: $i18n.t('确认删除当前资源'),
                subTitle: $i18n.t('确认删除资源 {kind}: {name}', { kind: row.kind, name }),
                defaultInfo: true,
                confirmFn: async (vm) => {
                    let result = false
                    if (type.value === 'crd') {
                        result = await $store.dispatch('dashboard/customResourceDelete', {
                            data: { namespace },
                            $crd: currentCrd.value,
                            $category: category.value,
                            $name: name
                        })
                    } else {
                        result = await $store.dispatch('dashboard/resourceDelete', {
                            $namespaceId: namespace,
                            $type: type.value,
                            $category: category.value,
                            $name: name
                        })
                    }
                    result && $bkMessage({
                        theme: 'success',
                        message: $i18n.t('删除成功')
                    })
                    handleGetTableData()
                }
            })
        }

        onMounted(async () => {
            isLoading.value = true
            const list: Promise<any>[] = []
            // 获取命名空间下拉列表
            if (showNameSpace.value) {
                list.push(getNamespaceData())
            }

            // 获取CRD下拉列表
            if (showCrd.value) {
                list.push(handleGetCrdData())
            }
            await Promise.all(list) // 等待初始数据加载完毕

            await handleGetTableData(false)// 关闭默认触发订阅的逻辑，等待CRD类型的列表初始化完后开始订阅
            isLoading.value = false

            // 所有资源就绪后开始订阅
            handleStartSubscribe()
        })

        return {
            namespaceValue,
            namespaceLoading,
            namespaceDisabled,
            showDetailPanel,
            curDetailRow,
            yaml,
            detailType,
            isLoading,
            pagePerms,
            pageConf: pagination,
            nameValue: searchValue,
            data,
            curPageData,
            namespaceList,
            currentCrd,
            crdLoading,
            crdList,
            currentCrdExt,
            additionalColumns,
            crdTips,
            getJsonPathValue,
            renderCrdHeader,
            stop,
            handlePageChange: pageChange,
            handlePageSizeChange: pageSizeChange,
            handleGetExtData,
            handleSortChange,
            gotoDetail,
            handleShowDetail,
            handleChangeDetailType,
            handleUpdateResource,
            handleDeleteResource,
            handleCreateResource,
            handleCrdChange,
            handleNamespaceSelected
        }
    },
    render () {
        return (
            <div class="biz-content base-layout">
                <div class="biz-top-bar">
                    <div class="dashboard-top-title">
                        {this.title}
                    </div>
                    <DashboardTopActions />
                </div>
                <div class="biz-content-wrapper" v-bkloading={{ isLoading: this.isLoading, zIndex: 10 }}>
                    <div class="base-layout-operate mb20">
                        {
                            this.showCreate ? (
                                <bk-button v-authority={{
                                    clickable: this.pagePerms.create?.clickable,
                                    content: this.pagePerms.create?.tip || this.crdTips || this.$t('无权限')
                                }}
                                class="resource-create"
                                icon="plus"
                                theme="primary"
                                onClick={this.handleCreateResource}>
                                    { this.$t('创建') }
                                </bk-button>
                            ) : <div></div>
                        }

                        <div class="search-wapper">
                            {
                                this.showCrd
                                    ? (
                                        <div class="select-wrapper">
                                            <span class="select-prefix">CRD</span>
                                            <bcs-select loading={this.crdLoading}
                                                class="dashboard-select"
                                                v-model={this.currentCrd}
                                                searchable
                                                clearable={false}
                                                placeholder={this.$t('选择CRD')}
                                                onChange={this.handleCrdChange}>
                                                {
                                                    this.crdList.map(option => (
                                                        <bcs-option
                                                            key={option.metadata.name}
                                                            id={option.metadata.name}
                                                            name={option.metadata.name}>
                                                        </bcs-option>
                                                    ))
                                                }
                                            </bcs-select>
                                        </div>
                                    )
                                    : null
                            }
                            {/** Scope类型不为Namespace时，隐藏命名空间方式会跳动，暂时用假的select替换 */}
                            {
                                this.showNameSpace
                                    ? (
                                        <div class="select-wrapper">
                                            <span class="select-prefix">{this.$t('命名空间')}</span>
                                            {
                                                this.namespaceDisabled
                                                    ? <bcs-select class="dashboard-select" placeholder={this.$t('请选择命名空间')} disabled></bcs-select>
                                                    : <bcs-select
                                                        v-bk-tooltips={{ disabled: !this.namespaceDisabled, content: this.crdTips }}
                                                        loading={this.namespaceLoading}
                                                        class="dashboard-select"
                                                        v-model={this.namespaceValue}
                                                        onSelected={this.handleNamespaceSelected}
                                                        searchable
                                                        clearable={false}
                                                        disabled={this.namespaceDisabled}
                                                        placeholder={this.$t('请选择命名空间')}>
                                                        {
                                                            this.namespaceList.map(option => (
                                                                <bcs-option
                                                                    key={option.metadata.name}
                                                                    id={option.metadata.name}
                                                                    name={option.metadata.name}>
                                                                </bcs-option>
                                                            ))
                                                        }
                                                    </bcs-select>
                                            }
                                        </div>
                                    )
                                    : null
                            }
                            <bk-input
                                class="search-input"
                                clearable
                                v-model={this.nameValue}
                                right-icon="bk-icon icon-search"
                                placeholder={this.$t('输入名称搜索')}>
                            </bk-input>
                        </div>
                    </div>
                    {
                        this.$scopedSlots.default && this.$scopedSlots.default({
                            isLoading: this.isLoading,
                            pageConf: this.pageConf,
                            data: this.data,
                            curPageData: this.curPageData,
                            handlePageChange: this.handlePageChange,
                            handlePageSizeChange: this.handlePageSizeChange,
                            handleGetExtData: this.handleGetExtData,
                            handleSortChange: this.handleSortChange,
                            gotoDetail: this.gotoDetail,
                            handleShowDetail: this.handleShowDetail,
                            handleUpdateResource: this.handleUpdateResource,
                            handleDeleteResource: this.handleDeleteResource,
                            getJsonPathValue: this.getJsonPathValue,
                            renderCrdHeader: this.renderCrdHeader,
                            pagePerms: this.pagePerms,
                            additionalColumns: this.additionalColumns,
                            namespaceDisabled: this.namespaceDisabled
                        })
                    }
                </div>
                <bcs-sideslider
                    quick-close
                    isShow={this.showDetailPanel}
                    width={800}
                    {
                    ...{
                        on: {
                            'update:isShow': (show: boolean) => {
                                this.showDetailPanel = show
                            }
                        },
                        scopedSlots: {
                            header: () => (
                                <div class="detail-header">
                                    <span>{this.curDetailRow.data?.metadata?.name}</span>
                                    {
                                        this.showDetailTab
                                            ? (<div class="bk-button-group">
                                                {
                                                    this.detailType.list.map(item => (
                                                        <bk-button class={{ 'is-selected': this.detailType.active === item.id }}
                                                            onClick={() => {
                                                                this.handleChangeDetailType(item.id)
                                                            }}>
                                                            {item.name}
                                                        </bk-button>
                                                    ))
                                                }
                                            </div>)
                                            : null
                                    }
                                </div>
                            ),
                            content: () =>
                                this.detailType.active === 'overview'
                                    ? (this.$scopedSlots.detail && this.$scopedSlots.detail({
                                        ...this.curDetailRow
                                    }))
                                    : <ace v-full-screen={{ tools: ['fullscreen', 'copy'], content: this.yaml }}
                                        width="100%" height="100%" lang="yaml"
                                        readOnly={true} value={this.yaml}></ace>
                        }
                    }
                    }></bcs-sideslider>
            </div>
        )
    }
})
