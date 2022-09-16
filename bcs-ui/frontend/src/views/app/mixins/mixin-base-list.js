/**
 * @file 应用列表 list 页的 base mixin
 */

import moment from 'moment'

import { catchErrorHandler } from '@/common/util'
import bkSearcher from '@/components/bk-searcher'
import bkDiff from '@/components/bk-diff'
import applyPerm from '@/mixins/apply-perm'
import ace from '@/components/ace-editor'

import MonacoEditor from '@/components/monaco-editor/editor.vue'

export default {
    mixins: [applyPerm],
    components: {
        bkDiff,
        bkSearcher,
        ace,
        MonacoEditor
    },
    data () {
        return {
            bkSearcherFixedSearchParams: [
                // {
                //     id: 'cluster_type',
                //     text: this.$t('集群类型'),
                //     list: [{ id: 1, text: this.$t('正式集群') }, { id: 2, text: this.$t('测试集群'), isSelected: true }]
                // }
            ],
            bkSearcherMask: false,
            bkSearcherFilterList: [],
            showLoading: false,
            instanceLoading: false,
            search: '',
            showClearSearch: false,
            rollingUpdateDialogConf: {
                isShow: false,
                width: 1050,
                title: '',
                closeIcon: true,
                loading: false,
                oldContent: '',
                newContent: '',
                oldVer: '',
                verList: [],
                selectedVerId: '',
                instanceNum: '',
                newVersion: {
                    instance_num: '',
                    instance_num_key: '',
                    instance_num_var_flag: false
                },
                oldVariableList: [],
                newVariableList: [],
                isShowVariable: false,
                noDiffMsg: ''
            },
            rollingUpdateInNotPlatformDialogConf: {
                isShow: false,
                width: 1050,
                title: '',
                closeIcon: true,
                loading: false,
                content: '',
                // diff, edit
                showType: 'edit',
                toggleStr: this.$t('查看对比')
            },
            updateInNotPlatformDialogConf: {
                isShow: false,
                width: 1050,
                title: '',
                closeIcon: true,
                loading: false,
                content: ''
            },
            compareEditorConfig: {
                width: '100%',
                height: '100%',
                lang: 'json',
                readOnly: true,
                fullScreen: false,
                editor: null,
                value: ''
            },
            editorConfig: {
                width: '100%',
                height: '100%',
                lang: 'json',
                readOnly: false,
                fullScreen: false,
                editor: null,
                value: ''
            },
            editorValue: '',
            mouseInEditor: 'left',
            isUpdating: false,
            instanceTimer: null,
            // 当前操作的这个实例，供滚动升级、扩缩绒、重建、删除使用
            curInstance: {},
            // 缓存轮询 instanceList 的参数，作用是在各种操作的请求发出后，需要停掉 instanceTimer，重新开启一个轮询，
            // 这个时候需要在操作的回调里面调用 loopInstanceList 方法，需要 loopInstanceList 的参数
            loopInstanceListParams: [],
            instanceNumDialogConf: {
                isShow: false,
                width: 380,
                title: '',
                closeIcon: false
            },
            instanceNum: 0,
            reBuildDialogConf: {
                isShow: false,
                width: 840,
                title: '',
                closeIcon: false,
                loading: false,
                aceWidth: '100%',
                aceHeight: '100%',
                aceValue: '',
                aceLang: 'yaml',
                aceReadOnly: true,
                aceFullScreen: false,
                aceEditor: null
            },
            deleteDialogConf: {
                isShow: false,
                width: 400,
                title: '',
                content: '',
                closeIcon: false
            },
            forceDeleteDialogConf: {
                isShow: false,
                width: 400,
                title: '',
                content: '',
                closeIcon: false
            },
            bkMessageInstance: null,
            exceptionCode: null,

            // 页面显示的
            tmplMusterList: [],
            // 缓存 for search
            tmplMusterListTmp: [],
            // 视图，template 集群模板视图，namespace 命名空间视图
            viewMode: 'template',
            // namespace 视图的列表
            namespaceList: [],
            // for search
            namespaceListTmp: [],
            loopAppListParams: [],
            cancelLoop: false,
            searchParams: {
                // cluster_type: ''
            },
            searchParamsList: [],
            // 集群类型 1 正式，2 测试
            clusterType: '2',
            yamlEditorOptions: {
                readOnly: true,
                fontSize: 14
            },
            reRenderEditor: 0,
            reRenderJSONEditor: 0,
            monacoEditor: null,
            useEditorDiff: false,
            rollbackPreviousDialogConf: {
                isShow: false,
                width: 1050,
                title: '',
                closeIcon: true,
                loading: false,
                prevContent: '',
                curContent: ''
            },
            batchRebuildDialogConf: {
                isShow: false,
                list: [],
                tpl: null
            },
            clusterValue: '',
            clusterSearchSelectExpand: false,
            templateWebAnnotations: { perms: {} },
            namespaceWebAnnotations: { perms: {} },
            templateInstanceWebAnnotations: { perms: {} },
            namespaceInsWebAnnotations: { perms: {} }
        }
    },
    watch: {
        search (val) {
            if (val.trim()) {
                this.showClearSearch = true
            } else {
                this.showClearSearch = false
            }
            this.handleSearch()
        },
        searchParams () {
            this.cancelLoopAppList()
            this.cancelLoopInstanceList()
            // this.searchParamsListFromRoute
            // 为 undefined 说明是第一次进入
            // 为空数组说明只有默认的 fix 搜索条件进入 instance 然后返回的
            // 为有 item 的数组说明是带了搜索条件进入 instance 然后返回的
            // 为 ['fromSearch'] 时说明是在本页搜索导致的查询，这时不需要展开
            if (this.searchParamsListFromRoute && this.searchParamsListFromRoute[0] === 'fromSearch') {
                this.fetchData(false)
            } else {
                this.fetchData(this.searchParamsListFromRoute !== void 0)
            }
        },
        useEditorDiff () {
            this.reRenderEditor++
        },
        'rollbackPreviousDialogConf.isShow' (v) {
            const body = document.body
            if (body) {
                body.style.overflow = v ? 'hidden' : 'auto'
            }
        },
        showLoading (v) {
            this.bkSearcherMask = v
        },
        curClusterId (v) {
            this.clusterValue = v
        }
    },
    computed: {
        varList () {
            return this.$store.state.variable.varList
        },
        projectId () {
            return this.$route.params.projectId
        },
        projectCode () {
            return this.$route.params.projectCode
        },
        tplsetId () {
            return this.$route.params.tplsetId
        },
        namespaceId () {
            return this.$route.params.namespaceId
        },
        templateId () {
            return this.$route.params.templateId
        },
        isProdCluster () {
            return this.$route.params.isProdCluster
        },
        // 搜索条件中的 namespaceName
        namespaceNameInSearch () {
            return this.$route.params.namespaceNameInSearch
        },
        searchParamsListFromRoute () {
            return this.$route.params.searchParamsList || []
        },
        isEn () {
            return this.$store.state.isEn
        },
        curClusterId () {
            return this.$store.state.curClusterId
        },
        clusterList () {
            return this.$store.state.cluster.clusterList
        }
    },
    created () {
        let clusterId = ''
        const sessionStorageClusterId = sessionStorage['bcs-cluster']
        if (this.curClusterId) {
            clusterId = this.curClusterId
        } else if (sessionStorageClusterId && this.clusterList.length && this.clusterList.find(item => item.cluster_id === sessionStorageClusterId)) {
            // 应用进到pod页面，点击返回，回到记录的集群下,而不是选中第一个集群
            clusterId = sessionStorageClusterId
        } else if (this.clusterList.length) {
            clusterId = this.clusterList[0].cluster_id
        }
        this.clusterValue = clusterId

        if (!this.clusterValue) {
            this.showLoading = true
            this.cancelLoop = false
            setTimeout(() => {
                this.showLoading = false
            }, 1500)
            return
        }

        const appViewMode = localStorage.getItem('appViewMode')
        if (appViewMode) {
            this.viewMode = appViewMode
        } else {
            localStorage.setItem('appViewMode', this.viewMode)
        }

        if (appViewMode === 'namespace') {
            this.bkSearcherFilterList.splice(0, this.bkSearcherFilterList.length, ...[
                {
                    id: 'ns_id',
                    text: this.$t('命名空间名称')
                },
                {
                    id: 'app_id',
                    text: this.$t('应用名称')
                },
                {
                    id: 'app_status',
                    text: this.$t('状态'),
                    list: [
                        {
                            id: 1,
                            text: this.$t('正常')
                        },
                        {
                            id: 2,
                            text: this.$t('异常')
                        }
                    ]
                }
            ])
        }
        if (appViewMode === 'template') {
            this.bkSearcherFilterList.splice(0, this.bkSearcherFilterList.length, ...[
                {
                    id: 'muster_id',
                    text: this.$t('模板集名称')
                },
                {
                    id: 'app_id',
                    text: this.$t('应用名称')
                },
                {
                    id: 'ns_id',
                    text: this.$t('命名空间名称')
                }
            ])
        }

        if (!this.searchParamsListFromRoute || !this.searchParamsListFromRoute.length) {
            const searchParams = {}
            this.searchParams = JSON.parse(JSON.stringify(searchParams))
        } else {
            const searchParams = {}

            const bkSearcherFilterList = []
            bkSearcherFilterList.splice(0, 0, ...this.bkSearcherFilterList)

            this.searchParamsListFromRoute.forEach(item => {
                const obj = bkSearcherFilterList.filter(filter => filter.id === item.id)[0]
                if (obj) {
                    obj.list = item.list
                    obj.value = item.text
                }
                searchParams[item.id] = item.value.id
            })
            this.searchParams = JSON.parse(JSON.stringify(searchParams))
            this.bkSearcherFilterList.splice(0, this.bkSearcherFilterList.length, ...bkSearcherFilterList)
            this.searchParamsList.splice(0, this.searchParamsList.length, ...this.searchParamsListFromRoute)
        }

        this.initVarList()
    },
    beforeDestroy () {
        this.cancelLoopAppList()
        this.cancelLoopInstanceList()

        this.bkMessageInstance && this.bkMessageInstance.close()
        clearTimeout(this.instanceTimer)
        this.instanceTimer = null
    },
    methods: {
        clusterSearchSelectToggle (expand) {
            this.clusterSearchSelectExpand = expand
        },

        changeCluster (clusterId) {
            // 全部集群状态下，应用列表切换集群记录集群id
            sessionStorage['bcs-cluster'] = clusterId
            this.cancelLoopAppList()
            this.cancelLoopInstanceList()

            if (this.viewMode === 'namespace') {
                this.bkSearcherFilterList.splice(0, this.bkSearcherFilterList.length, ...[
                    {
                        id: 'ns_id',
                        text: this.$t('命名空间名称')
                    },
                    {
                        id: 'app_id',
                        text: this.$t('应用名称')
                    },
                    {
                        id: 'app_status',
                        text: this.$t('状态'),
                        list: [
                            {
                                id: 1,
                                text: this.$t('正常')
                            },
                            {
                                id: 2,
                                text: this.$t('异常')
                            }
                        ]
                    }
                ])
            }
            if (this.viewMode === 'template') {
                this.bkSearcherFilterList.splice(0, this.bkSearcherFilterList.length, ...[
                    {
                        id: 'muster_id',
                        text: this.$t('模板集名称')
                    },
                    {
                        id: 'app_id',
                        text: this.$t('应用名称')
                    },
                    {
                        id: 'ns_id',
                        text: this.$t('命名空间名称')
                    }
                ])
            }
            this.$refs.bkSearcher.$emit('resetSearchParams', true)

            // this.searchParamsListFromRoute
            // 为 undefined 说明是第一次进入
            // 为空数组说明只有默认的 fix 搜索条件进入 instance 然后返回的
            // 为有 item 的数组说明是带了搜索条件进入 instance 然后返回的
            // 为 ['fromSearch'] 时说明是在本页搜索导致的查询，这时不需要展开
            if (this.searchParamsListFromRoute && this.searchParamsListFromRoute[0] === 'fromSearch') {
                this.fetchData(false)
            } else {
                this.fetchData(this.searchParamsListFromRoute !== void 0)
            }
        },

        /**
         * 获取变量数据
         */
        async initVarList () {
            try {
                await this.$store.dispatch('variable/getBaseVarList', this.projectId)
            } catch (e) {
                catchErrorHandler(e, this)
            }
        },

        /**
         * 格式化时间
         *
         * @param {string} str 待格式的时间字符串
         *
         * @return {string} 格式后的时间字符串
         */
        formatDate (str) {
            // return moment(str).format('YYYY-MM-DD HH:mm:ss')
            return moment(str).format('MM-DD HH:mm')
        },

        /**
         * bk-search 搜索事件回调
         */
        bkSearch (data) {
            if (this.showLoading) {
                return
            }
            const p = []
            p.splice(0, 0, ...data)
            if (this.bkSearcherFixedSearchParams.length === 0) {
                // p.push({
                //     id: 'cluster_type',
                //     text: this.$t('集群类型'),
                //     list: [{ id: 1, text: this.$t('正式集群'), isSelected: true }],
                //     value: { id: 1, isSelected: true, text: this.$t('正式集群') }
                // })
            }
            const searchParamsList = []
            const searchParams = {}
            p.forEach(item => {
                searchParams[item.id] = item.value.id
                // if (item.id === 'cluster_type') {
                //     localStorage.setItem('clusterType', item.value.id)
                // } else {
                const len = item.list.length
                for (let i = 0; i < len; i++) {
                    if (item.list[i].id === item.value.id) {
                        item.list[i].isSelected = true
                        break
                    }
                }
                searchParamsList.push({
                    id: item.id,
                    list: item.list,
                    text: item.text,
                    value: item.value,
                    dynamicData: item.dynamicData
                })
                // }
            })

            this.searchParams = JSON.parse(JSON.stringify(searchParams))
            this.searchParamsList.splice(0, this.searchParamsList.length, ...searchParamsList)

            if (this.searchParamsListFromRoute && this.searchParamsListFromRoute.length) {
                // 搜搜时把 this.searchParamsListFromRoute 设置为特定的，为了让之前展开的收起来
                this.searchParamsListFromRoute.splice(0, this.searchParamsListFromRoute.length, ...['fromSearch'])
            }
        },

        /**
         * filter 选择事件回调，用于异步获取 filter 的 filter value 数据
         *
         * @param {Object} filter 当前选择的 filter
         * @param {Object} fixedSearchParams 固定的搜索参数，这里返回出来，有可能查询时需要
         * @param {Function} resolve promise resolve function
         * @param {Function} reject promise reject function
         */
        async bkSearcherGetFilterListData (filter, fixedSearchParams, resolve, reject) {
            const isNs = filter.id === 'ns_id'
            const isApp = filter.id === 'app_id'
            const isMuster = filter.id === 'muster_id'
            try {
                let url
                if (isNs) {
                    url = 'app/getAllNamespace4AppSearch'
                }
                if (isApp) {
                    url = 'app/getAllInstance4AppSearch'
                }
                if (isMuster) {
                    url = 'app/getAllMuster4AppSearch'
                }

                const fsp = Object.assign({}, fixedSearchParams)
                if (this.bkSearcherFixedSearchParams.length === 0) {
                    // TODO: 待接口支持 cluster_id 后，就可以去掉 cluster_type 条件
                    // fsp.cluster_type = 2
                }

                const params = Object.assign({ projectId: this.projectId, cluster_id: this.clusterValue }, fsp)

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch(url, params)

                const list = res.data || []
                list.forEach(item => {
                    if (isNs) {
                        item.id = item.ns_id || item.id
                        const clusterId = item.cluster_id || ''
                        item.text = item.name + (clusterId ? ` (${clusterId})` : '')
                    }
                    if (isApp) {
                        item.id = item.app_id
                        item.text = item.app_name
                    }
                    if (isMuster) {
                        item.id = item.muster_id
                        item.text = item.muster_name
                    }
                })
                resolve(list)
            } catch (e) {
                console.error(e)
            }
        },

        /**
         * 切换视图
         *
         * @param {string} str 视图
         */
        toggleView (str) {
            this.$refs.toggleDropdown.hide()
            this.viewMode = str
            localStorage.setItem('appViewMode', this.viewMode)

            if (this.viewMode === 'namespace') {
                this.bkSearcherFilterList.splice(0, this.bkSearcherFilterList.length, ...[
                    {
                        id: 'ns_id',
                        text: this.$t('命名空间名称')
                    },
                    {
                        id: 'app_id',
                        text: this.$t('应用名称')
                    },
                    {
                        id: 'app_status',
                        text: this.$t('状态'),
                        list: [
                            {
                                id: 1,
                                text: this.$t('正常')
                            },
                            {
                                id: 2,
                                text: this.$t('异常')
                            }
                        ]
                    }
                ])
            }
            if (this.viewMode === 'template') {
                this.bkSearcherFilterList.splice(0, this.bkSearcherFilterList.length, ...[
                    {
                        id: 'muster_id',
                        text: this.$t('模板集名称')
                    },
                    {
                        id: 'app_id',
                        text: this.$t('应用名称')
                    },
                    {
                        id: 'ns_id',
                        text: this.$t('命名空间名称')
                    }
                ])
            }

            this.cancelLoopAppList()
            this.cancelLoopInstanceList()

            this.$refs.bkSearcher.$emit('resetSearchParams', true)
        },

        /**
         * 获取列表数据
         *
         * @param {boolean} autoExpand 是否需要自动展开对应的模板集或者 namespace
         */
        async fetchData (autoExpand) {
            if (this.showLoading) {
                return
            }
            this.showLoading = true
            this.cancelLoop = false
            this.search = ''
            this.loopInstanceListParams.splice(0, this.loopInstanceListParams.length, ...[])
            this.loopAppListParams.splice(0, this.loopAppListParams.length, ...[])
            try {
                this.tmplMusterList.splice(0, this.tmplMusterList.length, ...[])
                this.tmplMusterListTmp.splice(0, this.tmplMusterListTmp.length, ...[])

                this.namespaceList.splice(0, this.namespaceList.length, ...[])
                this.namespaceListTmp.splice(0, this.namespaceListTmp.length, ...[])

                // 集群模板视图
                if (this.viewMode === 'template') {
                    const params = Object.assign({
                        projectId: this.projectId,
                        cluster_id: this.clusterValue
                    }, this.searchParams)

                    if (this.CATEGORY) {
                        params.category = this.CATEGORY
                    }

                    const res = await this.$store.dispatch('app/getMusters', params)
                    const list = res.data || []
                    this.templateWebAnnotations = res.web_annotations || { perms: {} }

                    list.forEach(item => {
                        this.tmplMusterList.push({
                            ...item,
                            templateList: []
                        })
                        this.tmplMusterListTmp.push({
                            ...item,
                            templateList: []
                        })
                    })

                    // 展开模板集的逻辑
                    if (this.tplsetId && autoExpand) {
                        setTimeout(async () => {
                            const len = this.tmplMusterList.length
                            for (let tmplMusterIndex = 0; tmplMusterIndex < len; tmplMusterIndex++) {
                                const tmplMuster = this.tmplMusterList[tmplMusterIndex]
                                if (String(tmplMuster.tmpl_muster_id) === String(this.tplsetId)) {
                                    const scrollToDom = document.querySelectorAll('.list-item-tplset')[tmplMusterIndex]
                                    if (scrollToDom) {
                                        window.scrollTo({
                                            top: scrollToDom.offsetTop,
                                            behavior: 'smooth'
                                        })
                                    }
                                    await this.toggleTmplMuster(tmplMuster, tmplMusterIndex)
                                    if (this.templateId) {
                                        const len1 = tmplMuster.templateList.length
                                        for (let index = 0; index < len1; index++) {
                                            const tpl = tmplMuster.templateList[index]
                                            if (String(tpl.tmpl_app_id) === String(this.templateId)) {
                                                await this.toggleTemplate(tmplMuster, tmplMusterIndex, tpl, index)
                                            }
                                        }
                                    }
                                    break
                                }
                            }
                        }, 4)
                    }
                } else {
                    // exist_app: 1: 有应用的命名空间 0: 有应用和没有应用的命名空间
                    const params = Object.assign({
                        projectId: this.projectId,
                        exist_app: 1,
                        cluster_id: this.clusterValue
                    }, this.searchParams)

                    if (this.CATEGORY) {
                        params.category = this.CATEGORY
                    }

                    const res = await this.$store.dispatch('app/getNamespaces', params)

                    const list = res.data || []
                    this.namespaceWebAnnotations = res.web_annotations || { perms: {} }

                    list.forEach(item => {
                        this.namespaceList.push({
                            ...item,
                            appList: []
                        })
                        this.namespaceListTmp.push({
                            ...item,
                            appList: []
                        })
                    })

                    // 展开 namespace 的逻辑
                    if (this.namespaceId && autoExpand) {
                        setTimeout(async () => {
                            const len = this.namespaceList.length
                            for (let i = 0; i < len; i++) {
                                if (String(this.namespaceList[i].id) === String(this.namespaceId)) {
                                    await this.toggleNamespace(this.namespaceList[i], i)
                                    const scrollToDom = document.querySelectorAll('.list-item-tplset')[i]
                                    scrollToDom && window.scrollTo(0, scrollToDom.offsetTop)
                                    break
                                }
                            }
                        }, 4)
                    }
                }
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                this.showLoading = false
            }
        },

        /**
         * 展开/收起当前模板集
         *
         * @param {Object} tmplMuster 当前模板集对象
         * @param {number} index 当前模板集对象的索引
         */
        async toggleTmplMuster (tmplMuster, index) {
            tmplMuster.isOpen = !tmplMuster.isOpen
            this.$set(this.tmplMusterList, index, tmplMuster)
            if (!tmplMuster.isOpen) {
                this.cancelLoopInstanceList()
                return
            }

            tmplMuster.templateLoading = true
            tmplMuster.templateList = []
            this.$set(this.tmplMusterList, index, tmplMuster)

            try {
                const params = Object.assign({
                    projectId: this.projectId,
                    tmplMusterId: tmplMuster.tmpl_muster_id,
                    cluster_id: this.clusterValue
                }, this.searchParams)

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getTemplateList', params)

                const list = res.data || []
                tmplMuster.templateList.splice(0, tmplMuster.templateList.length, ...[])
                list.forEach(item => {
                    tmplMuster.templateList.push({
                        ...item,
                        instanceList: []
                    })
                })
            } catch (e) {
                catchErrorHandler(e, this)
            } finally {
                tmplMuster.templateLoading = false
                this.$set(this.tmplMusterList, index, tmplMuster)
            }
        },

        /**
         * 设置 namespaceList 里的 namespace 的 isOpen 状态
         *
         * @param {Object} namespace 当前 namespace 对象
         * @param {number} index 当前 namespace 对象的索引
         */
        setNamespaceOpenStatus (namespace, index) {
            const namespaceList = []
            namespaceList.splice(0, 0, ...this.namespaceList)
            namespaceList.forEach(ns => {
                if (ns.id === namespace.id) {
                    ns.isOpen = !ns.isOpen
                } else {
                    ns.isOpen = false
                }
                clearTimeout(ns.timer)
                ns.timer = null
            })
            this.namespaceList.splice(0, this.namespaceList.length, ...namespaceList)
        },

        /**
         * 展开/收起当前 namespace
         *
         * @param {Object} namespace 当前 namespace 对象
         * @param {number} index 当前 namespace 对象的索引
         */
        async toggleNamespace (namespace, index) {
            this.setNamespaceOpenStatus(namespace, index)
            this.cancelLoop = !namespace.isOpen
            if (!namespace.isOpen) {
                clearTimeout(namespace.timer)
                namespace.timer = null
                return
            }

            this.loopAppListParams.splice(
                0,
                this.loopAppListParams.length,
                ...[namespace, index]
            )

            namespace.appLoading = true
            namespace.appList.splice(0, namespace.appList.length, ...[])
            namespace.isAllChecked = false
            if (namespace.prepareDeleteInstances && namespace.prepareDeleteInstances.length) {
                namespace.prepareDeleteInstances.splice(0, namespace.prepareDeleteInstances.length, ...[])
            }
            await this.fetchAppListInNamespaceViewMode(namespace, index)
        },

        /**
         * 应用列表命名空间视图展开命名空间获取命名空间下的应用列表
         *
         * @param {Object} namespace 当前 namespace 对象
         * @param {number} index 当前 namespace 对象的索引
         */
        async fetchAppListInNamespaceViewMode (namespace, index) {
            if (this.cancelLoop) {
                return
            }

            try {
                const params = Object.assign({
                    projectId: this.projectId,
                    namespaceId: namespace.id,
                    cluster_id: this.clusterValue
                }, this.searchParams)

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getAppListInNamespaceViewMode', params)
                this.namespaceInsWebAnnotations = {
                    perms: Object.assign(this.namespaceInsWebAnnotations.perms, res.web_annotations.perms || {})
                }

                const data = res.data || {}
                namespace.error_num = data.error_num || 0
                namespace.total_num = data.total_num || 0

                const instanceList = data.instance_list || []

                // 待批量删除的实例的 id 集合，用于之后实例的多选框删除
                if (!namespace.prepareDeleteInstances) {
                    namespace.prepareDeleteInstances = []
                } else {
                    if (namespace.prepareDeleteInstances.length === instanceList.length) {
                        namespace.isAllChecked = true
                    }
                }

                const list = []
                instanceList.forEach(item => {
                    if (namespace.prepareDeleteInstances.find(n => n.name === item.name)) {
                        item.isChecked = true
                    }

                    // k8s
                    if (this.CATEGORY) {
                        item.state = new this.State({
                            status: item.status,
                            backendStatus: item.backend_status,
                            instance: item.instance,
                            buildInstance: item.build_instance,
                            operType: item.oper_type,
                            operTypeFlag: item.oper_type_flag,
                            hpa: item.hpa,
                            category: this.CATEGORY,
                            islock: false
                        })
                    } else {
                        item.state = new this.State({
                            category: item.category,
                            backendStatus: item.backend_status,
                            applicationStatus: item.application_status,
                            deploymentStatus: item.deployment_status,
                            hpa: item.hpa,
                            islock: false
                        })
                    }
                    list.push(item)
                })

                namespace.appList.splice(0, namespace.appList.length, ...list)
                this.$set(this.namespaceList, index, namespace)
                if (this.cancelLoop) {
                    clearTimeout(namespace.timer)
                    namespace.timer = null
                } else {
                    namespace.timer = setTimeout(() => {
                        this.fetchAppListInNamespaceViewMode(namespace, index)
                    }, 5000)
                }
            } catch (e) {
                console.error(e)
            } finally {
                namespace.appLoading = false
            }
        },

        /**
         * 命名空间视图取消 loop
         */
        cancelLoopAppList () {
            // namespace 视图中是否有轮询的 namespace，如果有，就清除掉
            const namespaceList = []
            namespaceList.splice(0, 0, ...this.namespaceList)
            namespaceList.forEach(item => {
                if (item && item.timer) {
                    clearTimeout(item.timer)
                    item.timer = null
                }
            })
            this.namespaceList.splice(0, this.namespaceList.length, ...namespaceList)

            this.cancelLoop = true
        },

        /**
         * 清除 tmplMuster.templateList 里每个 tpl 的 isOpen 状态
         *
         * @param {Object} tmplMuster 当前模板集对象
         * @param {number} tmplMusterIndex 当前模板集对象的索引
         */
        clearShowInstanceStatus (tmplMuster, tmplMusterIndex) {
            const templateList = []
            templateList.splice(0, 0, ...tmplMuster.templateList)
            if (!templateList || !templateList.length) {
                return
            }
            templateList.forEach(tpl => {
                tpl.isOpen = false
                clearTimeout(tpl.timer)
                tpl.timer = null
            })
            tmplMuster.templateList.splice(0, tmplMuster.templateList.length, ...templateList)
            this.$set(this.tmplMusterList, tmplMusterIndex, tmplMuster)
        },

        /**
         * 切换显示隐藏模板下的实例
         *
         * @param {Object} tmplMuster 当前模板集对象
         * @param {number} tmplMusterIndex 当前模板集对象的索引
         * @param {Object} tpl 当前点击的模板对象
         * @param {number} index 当前点击的模板对象的索引
         */
        async toggleTemplate (tmplMuster, tmplMusterIndex, tpl, index) {
            const oldOpenStatus = tpl.isOpen
            this.clearShowInstanceStatus(tmplMuster, tmplMusterIndex)
            tpl.isOpen = !oldOpenStatus
            this.cancelLoop = !tpl.isOpen
            // 说明是收起
            if (oldOpenStatus) {
                clearTimeout(tpl.timer)
                tpl.timer = null
                return
            }

            // 把当前没有展开的模板集里的模板的 instance 实例列表收起
            this.tmplMusterList.forEach((tm, i) => {
                if (i !== tmplMusterIndex) {
                    this.clearShowInstanceStatus(tm, i)
                }
            })

            this.loopInstanceListParams.splice(
                0,
                this.loopInstanceListParams.length,
                ...[tmplMuster, tmplMusterIndex, tpl, index]
            )

            this.instanceLoading = true
            // 先清空 instanceList
            tpl.instanceList.splice(0, tpl.instanceList.length, ...[])
            tpl.isAllChecked = false
            if (tpl.prepareDeleteInstances && tpl.prepareDeleteInstances.length) {
                tpl.prepareDeleteInstances.splice(0, tpl.prepareDeleteInstances.length, ...[])
            }
            await this.fetchInstanceListInTemplateViewMode(tmplMuster, tmplMusterIndex, tpl, index)
        },

        /**
         * 轮询模板下的实例
         *
         * @param {Object} tmplMuster 当前模板集对象
         * @param {number} tmplMusterIndex 当前模板集对象的索引
         * @param {Object} tpl 当前点击的模板对象
         * @param {number} index 当前点击的模板对象的索引
         */
        async fetchInstanceListInTemplateViewMode (tmplMuster, tmplMusterIndex, tpl, index) {
            if (this.cancelLoop) {
                return
            }

            try {
                const params = Object.assign({
                    projectId: this.projectId,
                    tmplMusterId: tmplMuster.tmpl_muster_id,
                    templateId: tpl.tmpl_app_id,
                    category: tpl.category,
                    cluster_id: this.clusterValue
                }, this.searchParams)

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getInstanceList', params)
                this.templateInstanceWebAnnotations = {
                    perms: Object.assign(this.templateInstanceWebAnnotations.perms, res.web_annotations.perms || {})
                }

                const data = res.data || {}
                tpl.error_num = res.data.error_num || 0
                tpl.total_num = res.data.total_num || 0
                const instanceList = data.instance_list || []

                // 待批量删除的实例的 id 集合，用于之后实例的多选框删除
                if (!tpl.prepareDeleteInstances) {
                    tpl.prepareDeleteInstances = []
                } else {
                    if (tpl.prepareDeleteInstances.length === instanceList.length) {
                        tpl.isAllChecked = true
                    }
                }

                const list = []
                instanceList.forEach(item => {
                    item.templateId = tpl.tmpl_app_id
                    if (tpl.prepareDeleteInstances.indexOf(item.id) > -1) {
                        item.isChecked = true
                    }
                    // k8s
                    if (this.CATEGORY) {
                        item.state = new this.State({
                            status: item.status,
                            backendStatus: item.backend_status,
                            instance: item.instance,
                            buildInstance: item.build_instance,
                            operType: item.oper_type,
                            operTypeFlag: item.oper_type_flag,
                            hpa: item.hpa,
                            category: this.CATEGORY,
                            islock: false
                        })
                    } else {
                        item.state = new this.State({
                            category: item.category,
                            backendStatus: item.backend_status,
                            applicationStatus: item.application_status,
                            deploymentStatus: item.deployment_status,
                            hpa: item.hpa,
                            islock: false
                        })
                    }
                    list.push(item)
                })

                tpl.instanceList.splice(0, tpl.instanceList.length, ...list)

                if (this.cancelLoop) {
                    clearTimeout(tpl.timer)
                    tpl.timer = null
                } else {
                    tpl.timer = setTimeout(() => {
                        this.fetchInstanceListInTemplateViewMode(tmplMuster, tmplMusterIndex, tpl, index)
                    }, 5000)
                }
            } catch (e) {
                console.error(e)
            } finally {
                this.instanceLoading = false
            }
        },

        /**
         * 模板集视图取消 loop
         */
        cancelLoopInstanceList () {
            const tmplMusterList = []
            tmplMusterList.splice(0, 0, ...this.tmplMusterList)
            tmplMusterList.forEach(tmplMuster => {
                const templateList = []
                templateList.splice(0, 0, ...tmplMuster.templateList)
                if (templateList && templateList.length) {
                    templateList.forEach(tpl => {
                        if (tpl && tpl.timer) {
                            clearTimeout(tpl.timer)
                            tpl.timer = null
                        }
                    })
                    tmplMuster.templateList.splice(0, tmplMuster.templateList.length, ...templateList)
                }
            })

            this.tmplMusterList.splice(0, this.tmplMusterList.length, ...tmplMusterList)

            this.cancelLoop = true
        },

        /**
         * 显示更新弹层
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async showUpdate (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)

            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            if (!this.curInstance.from_platform && this.curInstance.id === 0) {
                this.updateInNotPlatformDialogConf.isShow = true
                this.updateInNotPlatformDialogConf.loading = true
                this.updateInNotPlatformDialogConf.title = this.$t('{instanceName}升级', {
                    instanceName: instance.name
                })

                try {
                    const params = {
                        projectId: this.projectId,
                        instanceId: this.curInstance.id,
                        name: this.curInstance.name,
                        namespace: this.curInstance.namespace,
                        category: this.curInstance.category,
                        cluster_id: this.curInstance.cluster_id
                    }
                    if (this.CATEGORY) {
                        params.category = this.CATEGORY
                    }
                    // 当前版本信息
                    const res = await this.$store.dispatch('app/getVersionInRollingUpdateInNotPlatform', params)
                    const content = JSON.stringify(res.data.json, null, 2)

                    this.$nextTick(() => {
                        setTimeout(() => {
                            this.compareEditorConfig.editor.gotoLine(0, 0, true)

                            this.editorConfig.editor.gotoLine(0, 0, true)
                            this.editorConfig.editor.focus()

                            // ace editor 同步滚动
                            const session = this.editorConfig.editor.getSession()
                            const compareSession = this.compareEditorConfig.editor.getSession()
                            session.on('changeScrollTop', scroll => {
                                compareSession.setScrollTop(parseInt(scroll, 10) || 0)
                            })
                            session.on('changeScrollLeft', scroll => {
                                compareSession.setScrollLeft(parseInt(scroll, 10) || 0)
                            })

                            compareSession.on('changeScrollTop', scroll => {
                                session.setScrollTop(parseInt(scroll, 10) || 0)
                            })
                            compareSession.on('changeScrollLeft', scroll => {
                                session.setScrollLeft(parseInt(scroll, 10) || 0)
                            })
                        }, 10)

                        this.editorValue = content
                        this.editorConfig.value = content
                        this.compareEditorConfig.value = content
                    })
                } catch (e) {
                    console.error(e)
                    this.$bkMessage({
                        theme: 'error',
                        message: e.message || e.data.msg || e.statusText
                    })
                } finally {
                    this.updateInNotPlatformDialogConf.loading = false
                }
            } else {
                this.updateDialogConf.isShow = true
                this.updateDialogConf.loading = true
                this.updateDialogConf.title = this.$t('{instanceName}升级', {
                    instanceName: instance.name
                })
                this.updateDialogConf.oldVer = this.curInstance.version
                this.updateDialogConf.verList.splice(0, this.updateDialogConf.verList.length, ...[])
                this.updateDialogConf.selectedVerId = -1
                this.updateDialogConf.instanceNum = this.curInstance.instance

                this.fetchVersionInfo(this.curInstance, this.updateDialogConf)
            }

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }
        },

        /**
         * 更新弹层更新按钮
         */
        updateInNotPlatformConfirm () {
            const me = this

            me.$bkInfo({
                title: this.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                confirmLoading: true,
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定更新【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        name: me.curInstance.name,
                        conf: me.editorValue,
                        cluster_id: me.curInstance.cluster_id
                    }
                    if (!me.curInstance.from_platform && me.curInstance.id === 0) {
                        params.namespace = me.curInstance.namespace
                        params.category = me.curInstance.category
                    }

                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/update4ApplicationInNotPlatform', params)
                        me.hideUpdateInNotPlatform()
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                }
            })
        },

        /**
         * 关闭更新弹层
         */
        hideUpdateInNotPlatform () {
            this.updateInNotPlatformDialogConf.isShow = false
            this.updateInNotPlatformDialogConf.loading = false

            this.curInstance = Object.assign({}, {})
            setTimeout(() => {
                this.editorValue = ''
                this.editorConfig.value = ''

                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
            }, 100)
        },

        /**
         * 更新弹层版本下拉框选择事件
         *
         * @param {number} id ver id
         * @param {Object} data ver 对象数据
         */
        async verSelectedInUpdate (id, data) {
            this.updateDialogConf.loading = true
            this.updateDialogConf.selectedVerId = id

            const lastNoDiffMsg = this.updateDialogConf.noDiffMsg
            this.updateDialogConf.noDiffMsg = ''
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.curInstance.id,
                    showVersionId: id,
                    cluster_id: this.curInstance.cluster_id
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const newRes = await this.$store.dispatch('app/getVersionInRollingUpdate', params)

                // 返回的内容和上次的一致，不会自动触发 diff 的 render
                if (newRes.data.yaml === this.updateDialogConf.newContent) {
                    this.updateDialogConf.noDiffMsg = lastNoDiffMsg
                }

                this.updateDialogConf.newContent = newRes.data.yaml
                this.updateDialogConf.instanceNum = newRes.data.instance_num
                this.updateDialogConf.newVersion.instance_num = newRes.data.instance_num
                this.updateDialogConf.newVersion.instance_num_key = newRes.data.instance_num_key
                this.updateDialogConf.newVersion.instance_num_var_flag = newRes.data.instance_num_var_flag

                this.updateDialogConf.newVariableList.splice(
                    0,
                    this.updateDialogConf.newVariableList.length,
                    ...(newRes.data.variable || [])
                )
                this.updateDialogConf.isShowVariable = false
            } catch (e) {
                console.error(e)
                this.$bkMessage({
                    theme: 'error',
                    message: e.message || e.data.msg || e.statusText
                })
            } finally {
                this.updateDialogConf.loading = false
            }
        },

        /**
         * 更新弹层切换编辑变量的元素
         */
        toggleVariableInUpdate () {
            this.updateDialogConf.isShowVariable = !this.updateDialogConf.isShowVariable
        },

        /**
         * 关闭更新弹层
         */
        hideUpdate () {
            this.updateDialogConf.isShow = false
            this.updateDialogConf.loading = false

            this.curInstance = Object.assign({}, {})
            setTimeout(() => {
                this.updateDialogConf.oldContent = ''
                this.updateDialogConf.newContent = ''
                this.updateDialogConf.oldVer = ''
                this.updateDialogConf.verList.splice(0, this.updateDialogConf.verList.length, ...[])
                this.updateDialogConf.selectedVerId = ''
                this.updateDialogConf.instanceNum = ''
                this.updateDialogConf.noDiffMsg = ''

                this.updateDialogConf.oldVariableList.splice(0, this.updateDialogConf.oldVariableList.length, ...[])
                this.updateDialogConf.newVariableList.splice(0, this.updateDialogConf.newVariableList.length, ...[])
                this.updateDialogConf.isShowVariable = false

                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
            }, 100)
        },

        /**
         * 更新弹层更新按钮
         */
        updateConfirm () {
            const me = this

            const variable = {}
            me.updateDialogConf.newVariableList.forEach(vari => {
                variable[vari.key] = vari.value
            })

            me.$bkInfo({
                title: this.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                confirmLoading: true,
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定更新【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        versionId: me.updateDialogConf.selectedVerId,
                        variable: variable,
                        cluster_id: me.curInstance.cluster_id
                    }

                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/update4Application', params)
                        me.hideUpdate()
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                }
            })
        },

        /**
         * ace 编辑器 annotation change 回调
         *
         * @param {Array} annotations annotations 数据
         */
        changeAnnotation (annotations) {
            const position = this.editorConfig.editor.getCursorPosition() || {}
            this.editorConfig.editor.gotoLine(annotations[0].row + 1, position.column || 0, true)
        },

        /**
         * 编辑器初始化之后的回调函数
         *
         * @param editor - 编辑器对象
         */
        editorInitAfter (editor) {
            this.editorConfig.editor = editor
        },

        editorChangeHandler (content) {
            this.editorValue = content
        },

        /**
         * 编辑器初始化之后的回调函数
         *
         * @param editor - 编辑器对象
         */
        editorInitAfterForCompare (editor) {
            editor.setOptions({
                readOnly: true,
                highlightActiveLine: false,
                highlightGutterLine: false
            })
            editor.renderer.hideCursor()
            editor.renderer.setMouseCursor('not-allowed')
            editor.renderer.setStyle('disabled')
            editor.renderer.$cursorLayer.element.style.opacity = 0
            this.compareEditorConfig.editor = editor
        },

        /**
         * ace editor 全屏
         *
         * @param {string} conf 当前全屏的是哪一个编辑器的标识
         */
        setFullScreen (conf) {
            const config = conf === 'compareEditor' ? this.compareEditorConfig : this.editorConfig
            config.fullScreen = true

            if (conf !== 'compareEditor') {
                this.$nextTick(() => {
                    config.editor.focus()
                })
            }
        },

        /**
         * 取消全屏
         */
        cancelFullScreen () {
            const conf = this.editorConfig.fullScreen ? this.editorConfig : this.compareEditorConfig
            conf.fullScreen = false
            this.$nextTick(() => {
                this.editorConfig.editor.focus()
            })
        },

        handleEditorMount (editorInstance, monacoEditor) {
            this.monacoEditor = monacoEditor
        },

        /**
         * 显示回滚上一版本弹层
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async showRollbackPrevious (instance, instanceIndex, instanceList) {
            // this.reRenderEditor++

            this.curInstance = Object.assign({}, instance)

            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            this.fetchRollbackPreviousInfo(this.curInstance)

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }
        },

        /**
         * 拉取回滚上一版本所需的信息
         *
         * @param {Object} instanceId 实例对象
         */
        async fetchRollbackPreviousInfo (instance) {
            if (this.rollbackPreviousDialogConf.loading) {
                return
            }

            this.rollbackPreviousDialogConf.loading = true

            try {
                const dropdownComps = Array.isArray(this.$refs.dropdown) ? this.$refs.dropdown : [this.$refs.dropdown]
                dropdownComps.forEach(item => {
                    item.hide()
                })

                const instanceId = instance.id

                const params = {
                    projectId: this.projectId,
                    instanceId: instanceId
                }

                if (instanceId === 0) {
                    params.name = instance.name
                    params.namespace = instance.namespace
                    params.category = instance.category
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const res = await this.$store.dispatch('app/getInstanceConfig4RollbackPrevious', params)
                const data = res.data || {}
                this.rollbackPreviousDialogConf.isShow = true

                setTimeout(() => {
                    this.rollbackPreviousDialogConf.title = this.$t('{instanceName}回滚上一版本', {
                        instanceName: instance.name
                    })

                    this.rollbackPreviousDialogConf.prevContent = data.last_config_yaml
                    this.rollbackPreviousDialogConf.curContent = data.current_config_yaml
                    this.useEditorDiff = true
                    this.reRenderEditor++
                })
            } catch (e) {
                console.error(e)
                // this.hideRollbackPrevious()

                this.$bkMessage({
                    theme: 'error',
                    message: e.message || e.data.msg || e.statusText
                })
            } finally {
                this.rollbackPreviousDialogConf.loading = false
            }
        },

        /**
         * 关闭回滚上一版本弹层
         */
        hideRollbackPrevious () {
            this.rollbackPreviousDialogConf.isShow = false
            this.rollbackPreviousDialogConf.loading = false

            this.curInstance = Object.assign({}, {})

            setTimeout(() => {
                this.rollbackPreviousDialogConf.prevContent = ''
                this.rollbackPreviousDialogConf.curContent = ''

                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }

                this.reRenderEditor++
                this.useEditorDiff = false
            }, 100)
        },

        /**
         * 关闭回滚上一版本弹层更新按钮
         */
        rollbackPreviousConfirm () {
            const me = this

            me.$bkInfo({
                title: me.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                confirmLoading: true,
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px',
                        width: '300px'
                    }
                }, this.$t('确定回滚上一版本【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id
                    }

                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/rollbackPrevious', params)
                        me.hideRollbackPrevious()
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                }
            })
        },

        /**
         * 显示滚动更新弹层
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async showRollingUpdate (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)

            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            if (!this.curInstance.from_platform && this.curInstance.id === 0) {
                this.rollingUpdateInNotPlatformDialogConf.isShow = true
                this.rollingUpdateInNotPlatformDialogConf.loading = true
                this.rollingUpdateInNotPlatformDialogConf.title = this.$t('{instanceName}滚动升级', {
                    instanceName: instance.name
                })

                try {
                    const params = {
                        projectId: this.projectId,
                        instanceId: this.curInstance.id,
                        name: this.curInstance.name,
                        namespace: this.curInstance.namespace,
                        category: this.curInstance.category,
                        cluster_id: this.curInstance.cluster_id
                    }
                    if (this.CATEGORY) {
                        params.category = this.CATEGORY
                    }
                    // 当前版本信息
                    const res = await this.$store.dispatch('app/getVersionInRollingUpdateInNotPlatform', params)
                    const content = JSON.stringify(res.data.json, null, 2)

                    this.$nextTick(() => {
                        setTimeout(() => {
                            this.compareEditorConfig.editor.gotoLine(0, 0, true)

                            this.editorConfig.editor.gotoLine(0, 0, true)
                            this.editorConfig.editor.focus()

                            // ace editor 同步滚动
                            const session = this.editorConfig.editor.getSession()
                            const compareSession = this.compareEditorConfig.editor.getSession()
                            session.on('changeScrollTop', scroll => {
                                compareSession.setScrollTop(parseInt(scroll, 10) || 0)
                            })
                            session.on('changeScrollLeft', scroll => {
                                compareSession.setScrollLeft(parseInt(scroll, 10) || 0)
                            })

                            compareSession.on('changeScrollTop', scroll => {
                                session.setScrollTop(parseInt(scroll, 10) || 0)
                            })
                            compareSession.on('changeScrollLeft', scroll => {
                                session.setScrollLeft(parseInt(scroll, 10) || 0)
                            })
                        }, 10)

                        this.editorValue = content
                        this.editorConfig.value = content
                        this.compareEditorConfig.value = content
                        this.reRenderJSONEditor++
                    })
                } catch (e) {
                    console.error(e)
                    this.$bkMessage({
                        theme: 'error',
                        message: e.message || e.data.msg || e.statusText
                    })
                } finally {
                    this.rollingUpdateInNotPlatformDialogConf.loading = false
                }
            } else {
                this.rollingUpdateDialogConf.isShow = true
                this.rollingUpdateDialogConf.loading = true
                this.rollingUpdateDialogConf.title = this.$t('{instanceName}滚动升级', {
                    instanceName: instance.name
                })
                this.rollingUpdateDialogConf.oldVer = this.curInstance.version
                this.rollingUpdateDialogConf.verList.splice(0, this.rollingUpdateDialogConf.verList.length, ...[])
                this.rollingUpdateDialogConf.selectedVerId = -1
                this.rollingUpdateDialogConf.instanceNum = this.curInstance.instance
                this.fetchVersionInfo(this.curInstance, this.rollingUpdateDialogConf)
            }

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }
        },

        /**
         * 非模板集滚动更新切换 diff 和 edit
         */
        toggleDiffEdit () {
            if (this.rollingUpdateInNotPlatformDialogConf.showType === 'edit') {
                if (this.compareEditorConfig.value === this.editorValue) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'primary',
                        delay: 2000,
                        message: this.$t('当前内容一致，没有差异')
                    })
                } else {
                    this.rollingUpdateInNotPlatformDialogConf.showType = 'diff'
                    this.rollingUpdateInNotPlatformDialogConf.toggleStr = this.$t('开始编辑')
                }
            } else {
                this.rollingUpdateInNotPlatformDialogConf.showType = 'edit'
                this.editorConfig.value = this.editorValue
                setTimeout(() => {
                    this.rollingUpdateInNotPlatformDialogConf.toggleStr = this.$t('查看对比')

                    this.compareEditorConfig.editor.gotoLine(0, 0, true)

                    this.editorConfig.editor.gotoLine(0, 0, true)
                    this.editorConfig.editor.focus()

                    // ace editor 同步滚动
                    const session = this.editorConfig.editor.getSession()
                    const compareSession = this.compareEditorConfig.editor.getSession()
                    session.on('changeScrollTop', scroll => {
                        compareSession.setScrollTop(parseInt(scroll, 10) || 0)
                    })
                    session.on('changeScrollLeft', scroll => {
                        compareSession.setScrollLeft(parseInt(scroll, 10) || 0)
                    })

                    compareSession.on('changeScrollTop', scroll => {
                        session.setScrollTop(parseInt(scroll, 10) || 0)
                    })
                    compareSession.on('changeScrollLeft', scroll => {
                        session.setScrollLeft(parseInt(scroll, 10) || 0)
                    })
                }, 10)
            }
        },

        /**
         * 关闭滚动更新弹层
         */
        hideRollingUpdateInNotPlatform () {
            this.rollingUpdateInNotPlatformDialogConf.isShow = false
            this.rollingUpdateInNotPlatformDialogConf.loading = false

            this.curInstance = Object.assign({}, {})
            setTimeout(() => {
                this.editorConfig.value = ''
                this.editorValue = ''
                this.rollingUpdateInNotPlatformDialogConf.showType = 'edit'
                this.rollingUpdateInNotPlatformDialogConf.toggleStr = this.$t('查看对比')

                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
                this.useEditorDiff = false
            }, 100)
        },

        /**
         * 滚动更新弹层更新按钮
         */
        rollingUpdateInNotPlatformConfirm () {
            const me = this

            me.$bkInfo({
                title: this.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                confirmLoading: true,
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定滚动升级【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        name: me.curInstance.name,
                        conf: me.editorValue,
                        cluster_id: me.curInstance.cluster_id
                    }
                    if (!me.curInstance.from_platform && me.curInstance.id === 0) {
                        params.namespace = me.curInstance.namespace
                        params.category = me.curInstance.category
                    }
                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/rollingUpdateInNotPlatform', params)
                        me.hideRollingUpdateInNotPlatform()
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                }
            })
        },

        /**
         * 查询版本信息
         *
         * @param {Object} instanceId 实例对象
         * @param {Object} dialogConfRef dialogConf 的引用
         */
        async fetchVersionInfo (instance, dialogConfRef) {
            try {
                const instanceId = instance.id

                const allVersionParams = {
                    projectId: this.projectId,
                    instanceId: instanceId,
                    cluster_id: instance.cluster_id
                }

                if (instanceId === 0) {
                    allVersionParams.name = instance.name
                    allVersionParams.namespace = instance.namespace
                    allVersionParams.category = instance.category
                }

                if (this.CATEGORY) {
                    allVersionParams.category = this.CATEGORY
                }

                // 查询所有版本信息
                const res = await this.$store.dispatch('app/getAllVersionInRollingUpdate', allVersionParams)
                const list = res.data || []
                if (list.length) {
                    dialogConfRef.verList.splice(0, dialogConfRef.verList.length, ...list)
                    dialogConfRef.selectedVerId = list[0].id

                    const oldVerParams = {
                        projectId: this.projectId,
                        instanceId: instanceId,
                        cluster_id: instance.cluster_id
                    }
                    if (this.CATEGORY) {
                        oldVerParams.category = this.CATEGORY
                    }
                    // 左侧当前版本信息
                    const oldRes = await this.$store.dispatch('app/getVersionInRollingUpdate', oldVerParams)

                    const newVerParams = {
                        projectId: this.projectId,
                        instanceId: instanceId,
                        showVersionId: dialogConfRef.selectedVerId,
                        cluster_id: instance.cluster_id
                    }
                    if (this.CATEGORY) {
                        newVerParams.category = this.CATEGORY
                    }
                    // 右侧要更新版本的信息
                    const newRes = await this.$store.dispatch('app/getVersionInRollingUpdate', newVerParams)
                    dialogConfRef.oldContent = oldRes.data.yaml
                    dialogConfRef.newContent = newRes.data.yaml
                    dialogConfRef.instanceNum = newRes.data.instance_num
                    dialogConfRef.newVersion.instance_num = newRes.data.instance_num
                    dialogConfRef.newVersion.instance_num_key = newRes.data.instance_num_key
                    dialogConfRef.newVersion.instance_num_var_flag = newRes.data.instance_num_var_flag

                    dialogConfRef.oldVariableList.splice(
                        0,
                        dialogConfRef.oldVariableList.length,
                        ...(oldRes.data.variable || [])
                    )
                    dialogConfRef.newVariableList.splice(
                        0,
                        dialogConfRef.newVariableList.length,
                        ...(newRes.data.variable || [])
                    )

                    this.useEditorDiff = true
                    this.reRenderEditor++
                }
            } catch (e) {
                console.error(e)
                this.$bkMessage({
                    theme: 'error',
                    message: e.message || e.data.msg || e.statusText
                })
            } finally {
                dialogConfRef.loading = false
            }
        },

        /**
         * diff 组件 change-count 事件监听
         *
         * @param {Number} count 变化的行数
         */
        getDiffChangeCount (count) {
            if (count === 0) {
                this.rollingUpdateDialogConf && (this.rollingUpdateDialogConf.noDiffMsg = this.$t('更新版本与当前版本无差异'))
                this.updateDialogConf && (this.updateDialogConf.noDiffMsg = this.$t('更新版本与当前版本无差异'))
                const diffWrapper = document.querySelector('.diff-wrapper')
                if (diffWrapper) {
                    this.$nextTick(() => {
                        const sideDiffNodes = diffWrapper.querySelectorAll('.d2h-file-side-diff')
                        if (sideDiffNodes) {
                            sideDiffNodes.forEach(node => {
                                const firstTr = node.querySelector('tr:nth-child(1)')
                                firstTr && firstTr.parentNode.removeChild(firstTr)
                            })
                        }
                    })
                }
            } else {
                this.rollingUpdateDialogConf && (this.rollingUpdateDialogConf.noDiffMsg = '')
                this.updateDialogConf && (this.updateDialogConf.noDiffMsg = '')
            }
        },

        instanceNumVarChange (val) {
            const varList = this.rollingUpdateDialogConf.newVariableList
            this.rollingUpdateDialogConf.newVersion.instance_num = val

            varList.forEach(item => {
                if (item.key === this.rollingUpdateDialogConf.newVersion.instance_num_key) {
                    this.rollingUpdateDialogConf.instanceNum = item.value
                }
            })
        },

        instanceNumChange (val) {
            this.rollingUpdateDialogConf.instanceNum = val
        },

        /**
         * 滚动更新弹层版本下拉框选择事件
         *
         * @param {number} id ver id
         * @param {Object} data ver 对象数据
         */
        async verSelected (id, data) {
            this.rollingUpdateDialogConf.loading = true
            this.rollingUpdateDialogConf.selectedVerId = id

            const lastNoDiffMsg = this.rollingUpdateDialogConf.noDiffMsg
            this.rollingUpdateDialogConf.noDiffMsg = ''
            try {
                const params = {
                    projectId: this.projectId,
                    instanceId: this.curInstance.id,
                    showVersionId: id,
                    cluster_id: this.curInstance.cluster_id
                }

                if (this.CATEGORY) {
                    params.category = this.CATEGORY
                }

                const newRes = await this.$store.dispatch('app/getVersionInRollingUpdate', params)

                // 返回的内容和上次的一致，不会自动触发 diff 的 render
                if (newRes.data.yaml === this.rollingUpdateDialogConf.newContent) {
                    this.rollingUpdateDialogConf.noDiffMsg = lastNoDiffMsg
                }

                this.rollingUpdateDialogConf.newContent = newRes.data.yaml
                this.rollingUpdateDialogConf.instanceNum = newRes.data.instance_num
                this.rollingUpdateDialogConf.newVersion.instance_num = newRes.data.instance_num
                this.rollingUpdateDialogConf.newVersion.instance_num_key = newRes.data.instance_num_key
                this.rollingUpdateDialogConf.newVersion.instance_num_var_flag = newRes.data.instance_num_var_flag

                this.rollingUpdateDialogConf.newVariableList.splice(
                    0,
                    this.rollingUpdateDialogConf.newVariableList.length,
                    ...(newRes.data.variable || [])
                )
                this.rollingUpdateDialogConf.isShowVariable = false
            } catch (e) {
                console.error(e)
                this.$bkMessage({
                    theme: 'error',
                    message: e.message || e.data.msg || e.statusText
                })
            } finally {
                this.reRenderEditor++
                this.rollingUpdateDialogConf.loading = false
            }
        },

        /**
         * 滚动升级弹层切换编辑变量的元素
         */
        toggleVariable () {
            this.rollingUpdateDialogConf.isShowVariable = !this.rollingUpdateDialogConf.isShowVariable
        },

        /**
         * 关闭滚动更新弹层
         */
        hideRollingUpdate () {
            this.rollingUpdateDialogConf.isShow = false
            this.rollingUpdateDialogConf.loading = false

            this.curInstance = Object.assign({}, {})
            setTimeout(() => {
                this.rollingUpdateDialogConf.oldContent = ''
                this.rollingUpdateDialogConf.newContent = ''
                this.rollingUpdateDialogConf.oldVer = ''
                this.rollingUpdateDialogConf.verList.splice(0, this.rollingUpdateDialogConf.verList.length, ...[])
                this.rollingUpdateDialogConf.selectedVerId = ''
                this.rollingUpdateDialogConf.instanceNum = ''
                this.rollingUpdateDialogConf.noDiffMsg = ''

                this.rollingUpdateDialogConf.oldVariableList.splice(
                    0,
                    this.rollingUpdateDialogConf.oldVariableList.length,
                    ...[]
                )
                this.rollingUpdateDialogConf.newVariableList.splice(
                    0,
                    this.rollingUpdateDialogConf.newVariableList.length,
                    ...[]
                )
                this.rollingUpdateDialogConf.isShowVariable = false

                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }

                this.useEditorDiff = false
            }, 100)
        },

        /**
         * 滚动更新弹层更新按钮
         */
        rollingUpdateConfirm () {
            const me = this

            if (!me.rollingUpdateDialogConf.instanceNum && me.CATEGORY !== 'daemonset') {
                me.bkMessageInstance && me.bkMessageInstance.close()
                me.bkMessageInstance = me.$bkMessage({
                    theme: 'error',
                    message: this.$t('请填写实例数')
                })
                return
            }

            const variable = {}
            me.rollingUpdateDialogConf.newVariableList.forEach(vari => {
                variable[vari.key] = vari.value
            })

            me.$bkInfo({
                title: me.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                confirmLoading: true,
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定滚动升级【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        version_id: me.rollingUpdateDialogConf.selectedVerId,
                        variable: variable,
                        cluster_id: me.curInstance.cluster_id
                    }

                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    if (me.CATEGORY !== 'daemonset') {
                        params.instance_num = me.rollingUpdateDialogConf.instanceNum
                    }

                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/rollingUpdate', params)
                        me.hideRollingUpdate()
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                }
            })
        },

        /**
         * 暂停滚动更新
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async pauseRollingUpdate (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }

            const me = this
            me.$bkInfo({
                title: me.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定暂停滚动升级【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        name: me.curInstance.name,
                        cluster_id: me.curInstance.cluster_id
                    }
                    if (!me.curInstance.from_platform && me.curInstance.id === 0) {
                        params.namespace = me.curInstance.namespace
                        params.category = me.curInstance.category
                    }

                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    // 用来触发组件的响应式，最终改变 me.curInstance.state
                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/pauseRollingUpdate', params)
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.cancelLoop = false
                            if (me.viewMode === 'namespace') {
                                me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                            } else {
                                me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                            }
                        }, 100)
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                },
                cancelFn () {
                    setTimeout(() => {
                        me.cancelLoop = false
                        if (me.viewMode === 'namespace') {
                            me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                        } else {
                            me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                        }
                    }, 100)
                }
            })
        },

        /**
         * 取消滚动更新
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async cancelRollingUpdate (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }

            const me = this
            me.$bkInfo({
                title: me.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定取消滚动升级【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        name: me.curInstance.name,
                        cluster_id: me.curInstance.cluster_id
                    }
                    if (!me.curInstance.from_platform && me.curInstance.id === 0) {
                        params.namespace = me.curInstance.namespace
                        params.category = me.curInstance.category
                    }

                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    // 用来触发组件的响应式，最终改变 me.curInstance.state
                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/cancelRollingUpdate', params)
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.cancelLoop = false
                            if (me.viewMode === 'namespace') {
                                me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                            } else {
                                me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                            }
                        }, 100)
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                },
                cancelFn () {
                    setTimeout(() => {
                        me.cancelLoop = false
                        if (me.viewMode === 'namespace') {
                            me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                        } else {
                            me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                        }
                    }, 100)
                }
            })
        },

        /**
         * 恢复滚动更新
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async resumeRollingUpdate (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }

            const me = this
            me.$bkInfo({
                title: me.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定恢复滚动升级【{instanceName}】？', { instanceName: me.curInstance.name })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        name: me.curInstance.name,
                        cluster_id: me.curInstance.cluster_id
                    }
                    if (!me.curInstance.from_platform && me.curInstance.id === 0) {
                        params.namespace = me.curInstance.namespace
                        params.category = me.curInstance.category
                    }

                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    // 用来触发组件的响应式，最终改变 me.curInstance.state
                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/resumeRollingUpdate', params)
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.cancelLoop = false
                            if (me.viewMode === 'namespace') {
                                me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                            } else {
                                me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                            }
                        }, 100)
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                },
                cancelFn () {
                    setTimeout(() => {
                        me.cancelLoop = false
                        if (me.viewMode === 'namespace') {
                            me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                        } else {
                            me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                        }
                    }, 100)
                }
            })
        },

        /**
         * 显示扩缩容弹层
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async showInstanceNum (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            this.instanceNumDialogConf.isShow = true
            this.instanceNumDialogConf.title = this.$t('{instanceName}扩缩容', {
                instanceName: instance.name
            })

            this.instanceNum = this.curInstance.instance

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }
        },

        /**
         * 关闭滚动更新弹层
         */
        hideInstanceNum () {
            this.instanceNumDialogConf.isShow = false
            this.curInstance = Object.assign({}, {})
            setTimeout(() => {
                this.instanceNum = 0
                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
            }, 100)
        },

        /**
         * 扩缩容弹层更新按钮
         */
        async instanceNumConfirm () {
            const me = this

            const originalNum = parseFloat(me.curInstance.instance)
            const instanceNum = parseFloat(me.instanceNum)
            if (originalNum === instanceNum) {
                me.bkMessageInstance = me.$bkMessage({
                    theme: 'primary',
                    message: me.$t('实例数量没有变化'),
                    delay: 1500
                })
                return
            }
            let msg = ''
            if (this.isEn) {
                if (instanceNum > originalNum) {
                    msg = `Confirm to scale up ${instanceNum} instances`
                } else {
                    msg = `Confirm to scale down ${instanceNum} instances`
                }
            } else {
                if (instanceNum > originalNum) {
                    msg = `确定扩容到 ${instanceNum} 个实例`
                } else {
                    msg = `确定缩容到 ${instanceNum} 个实例`
                }
            }
            me.$bkInfo({
                title: me.$t('确认操作'),
                confirmLoading: true,
                clsName: 'biz-confirm-dialog',
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, msg),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        name: me.curInstance.name,
                        instanceNum: instanceNum,
                        cluster_id: me.curInstance.cluster_id
                    }
                    if (!me.curInstance.from_platform && me.curInstance.id === 0) {
                        params.namespace = me.curInstance.namespace
                        params.category = me.curInstance.category
                    }
                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/scaleInstanceNum', params)
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        me.hideInstanceNum()
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                }
            })
        },

        /**
         * 显示重建弹层
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async showReBuild (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            this.reBuildDialogConf.isShow = true
            this.reBuildDialogConf.title = this.$t('确定重建【{instanceName}】？', { instanceName: instance.name })

            try {
                if (!this.curInstance.from_platform && this.curInstance.id === 0) {
                    const params = {
                        projectId: this.projectId,
                        instanceId: this.curInstance.id,
                        name: this.curInstance.name,
                        namespace: this.curInstance.namespace,
                        category: this.curInstance.category,
                        cluster_id: this.curInstance.cluster_id
                    }
                    this.reBuildDialogConf.loading = true
                    const res = await this.$store.dispatch('app/getVersionInRollingUpdateInNotPlatform', params)
                    const content = res.data.yaml
                    this.$nextTick(() => {
                        setTimeout(() => {
                            this.reBuildDialogConf.aceEditor.gotoLine(0, 0, true)
                        }, 10)

                        this.reBuildDialogConf.aceValue = content
                    })
                } else {
                    const params = {
                        projectId: this.projectId,
                        instanceId: this.curInstance.id,
                        cluster_id: this.curInstance.cluster_id
                    }
                    this.reBuildDialogConf.loading = true
                    const res = await this.$store.dispatch('app/getVersionInRollingUpdate', params)

                    const content = res.data.yaml
                    this.$nextTick(() => {
                        setTimeout(() => {
                            this.reBuildDialogConf.aceEditor.gotoLine(0, 0, true)
                        }, 10)

                        this.reBuildDialogConf.aceValue = content
                    })
                }
            } catch (e) {
                console.error(e)
                this.$bkMessage({
                    theme: 'error',
                    message: e.message || e.data.msg || e.statusText
                })
            } finally {
                this.reBuildDialogConf.loading = false
            }

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }
        },

        /**
         * 重建弹框中的编辑器初始化之后的回调函数
         *
         * @param editor - 编辑器对象
         */
        editorInitInReBuild (editor) {
            editor.setOptions({
                readOnly: true,
                highlightActiveLine: false,
                highlightGutterLine: false
            })
            editor.renderer.hideCursor()
            editor.renderer.setMouseCursor('not-allowed')
            editor.renderer.setStyle('ace-disabled')
            editor.renderer.$cursorLayer.element.style.opacity = 0
            this.reBuildDialogConf.aceEditor = editor
        },

        /**
         * 关闭重建弹层
         */
        hideReBuild () {
            this.reBuildDialogConf.isShow = false
            this.curInstance = Object.assign({}, {})
            this.reBuildDialogConf.aceValue = ''

            setTimeout(() => {
                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
            }, 100)
        },

        /**
         * 重建
         */
        async reBuildConfirm () {
            const params = {
                projectId: this.projectId,
                data: {
                    resource_list: [{
                        resource_kind: this.curInstance.category,
                        name: this.curInstance.name,
                        namespace: this.curInstance.namespace,
                        cluster_id: this.curInstance.cluster_id
                    }]
                }
            }
            const url = 'app/batchRebuild'

            this.isUpdating = true
            this.curInstance.state && this.curInstance.state.lock()
            try {
                await this.$store.dispatch(url, params)
            } catch (e) {
                console.log(e)
            } finally {
                this.isUpdating = false
                this.hideReBuild()
                setTimeout(() => {
                    this.curInstance.state && this.curInstance.state.unlock()
                }, 1000)
            }
        },

        /**
         * 显示删除弹层
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async showDelete (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            this.deleteDialogConf.isShow = true
            this.deleteDialogConf.title = this.$t('确认删除')
            this.deleteDialogConf.content = this.$t('确定要删除【{instanceName}】？', { instanceName: instance.name })

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }
        },

        /**
         * 关闭删除弹层
         */
        hideDelete () {
            this.deleteDialogConf.isShow = false
            this.curInstance = Object.assign({}, {})

            setTimeout(() => {
                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
            }, 100)
        },

        /**
         * 删除
         */
        async deleteConfirm () {
            const params = {
                projectId: this.projectId,
                instanceId: this.curInstance.id,
                name: this.curInstance.name,
                cluster_id: this.curInstance.cluster_id
            }
            if (!this.curInstance.from_platform && this.curInstance.id === 0) {
                params.namespace = this.curInstance.namespace
                params.category = this.curInstance.category
            }
            if (this.CATEGORY) {
                params.category = this.CATEGORY
            }

            this.isUpdating = true
            this.curInstance.state && this.curInstance.state.lock()
            try {
                await this.$store.dispatch('app/deleteInstance', params)
            } catch (e) {
                console.log(e)
            } finally {
                this.isUpdating = false
                this.hideDelete()
                setTimeout(() => {
                    this.curInstance.state && this.curInstance.state.unlock()
                }, 1000)
            }
        },

        /**
         * 显示强制删除弹层
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async showForceDelete (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            this.forceDeleteDialogConf.isShow = true
            this.forceDeleteDialogConf.title = this.$t('确认删除')
            this.forceDeleteDialogConf.content = this.$t('确定要强制删除【{instanceName}】？', {
                instanceName: instance.name
            })

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }
        },

        /**
         * 关闭强制删除弹层
         */
        hideForceDelete () {
            this.forceDeleteDialogConf.isShow = false
            this.curInstance = Object.assign({}, {})

            setTimeout(() => {
                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
            }, 100)
        },

        /**
         * 强制删除
         */
        async forceDeleteConfirm () {
            const params = {
                projectId: this.projectId,
                instanceId: this.curInstance.id,
                name: this.curInstance.name,
                cluster_id: this.curInstance.cluster_id
            }
            if (!this.curInstance.from_platform && this.curInstance.id === 0) {
                params.namespace = this.curInstance.namespace
                params.category = this.curInstance.category
            }
            if (this.CATEGORY) {
                params.category = this.CATEGORY
            }

            this.isUpdating = true
            this.curInstance.state && this.curInstance.state.lock()
            try {
                await this.$store.dispatch('app/forceDeleteInstance', params)
            } catch (e) {
                console.log(e)
            } finally {
                this.isUpdating = false
                this.hideForceDelete()
                setTimeout(() => {
                    this.curInstance.state && this.curInstance.state.unlock()
                }, 1000)
            }
        },

        /**
         * backend_status 为 BackendError 时，重试按钮操作即重新创建
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Array} instanceList 当前 tpl 下的实例集合
         */
        async reCreate (instance, instanceIndex, instanceList) {
            this.curInstance = Object.assign({}, instance)
            this.curInstance.instanceList = instanceList
            this.curInstance.instanceIndex = instanceIndex

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }

            const me = this
            me.$bkInfo({
                title: this.$t('确认操作'),
                clsName: 'biz-confirm-dialog',
                confirmLoading: true,
                content: me.$createElement('p', {
                    style: {
                        color: '#666',
                        fontSize: '14px',
                        marginLeft: '-7px'
                    }
                }, this.$t('确定重新创建【{instanceName}】？', {
                    instanceName: me.curInstance.name
                })),
                async confirmFn () {
                    const params = {
                        projectId: me.projectId,
                        instanceId: me.curInstance.id,
                        cluster_id: me.curInstance.cluster_id
                    }
                    if (me.CATEGORY) {
                        params.category = me.CATEGORY
                    }

                    // 用来触发组件的响应式，最终改变 me.curInstance.state
                    me.isUpdating = true
                    me.curInstance.state && me.curInstance.state.lock()
                    try {
                        await me.$store.dispatch('app/reCreate', params)
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.cancelLoop = false
                            if (me.viewMode === 'namespace') {
                                me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                            } else {
                                me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                            }
                        }, 100)
                        setTimeout(() => {
                            me.curInstance.state && me.curInstance.state.unlock()
                        }, 1000)
                    }
                },
                cancelFn () {
                    setTimeout(() => {
                        me.cancelLoop = false
                        if (me.viewMode === 'namespace') {
                            me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                        } else {
                            me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                        }
                    }, 100)
                }
            })
        },

        /**
         * 模板集视图，多选框全选中，for batch delete
         *
         * @param {Object} e 事件对象
         * @param {Object} tpl 当前点击的模板对象
         * @param {number} index 当前点击的模板对象的索引
         * @param {Array} 当前模板集下所有的模板数组
         * @param {boolean} checked 是否选中
         */
        checkAllInstance (tpl, index, tplList, checked) {
            const instanceList = []
            instanceList.splice(0, 0, ...tpl.instanceList)

            const prepareDeleteInstances = []
            prepareDeleteInstances.splice(0, 0, ...[])

            instanceList.forEach(item => {
                if (this.templateInstanceWebAnnotations.perms[item.iam_ns_id]
                    && this.templateInstanceWebAnnotations.perms[item.iam_ns_id].namespace_scoped_delete) {
                    item.isChecked = checked
                    checked && prepareDeleteInstances.push(item.id)
                }
            })

            tpl.isAllChecked = checked
            tpl.prepareDeleteInstances.splice(0, tpl.prepareDeleteInstances.length, ...prepareDeleteInstances)
            this.$set(tplList, index, tpl)
        },

        /**
         * 模板集视图，多选框选中，for batch delete
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Object} tpl 当前点击的模板对象
         * @param {number} index 当前点击的模板对象的索引
         * @param {Array} 当前模板集下所有的模板数组
         * @param {boolean} checked 是否选中
         */
        checkInstance (instance, instanceIndex, tpl, index, tplList, checked) {
            const id = instance.id
            const prepareDeleteInstances = []
            prepareDeleteInstances.splice(0, 0, ...tpl.prepareDeleteInstances)
            const prepareDeleteInstanceIndex = prepareDeleteInstances.indexOf(id)
            // 存在，说明这一次的点击是未选中
            if (prepareDeleteInstanceIndex > -1) {
                prepareDeleteInstances.splice(prepareDeleteInstanceIndex, 1)
            } else {
                prepareDeleteInstances.push(id)
            }

            const allLength = tpl.instanceList.length
            const invalidLength = tpl.instanceList.filter(inst => !this.getTemplateInsPerms(inst, 'namespace_scoped_delete')).length

            if (prepareDeleteInstances.length === allLength - invalidLength) {
                tpl.isAllChecked = true
            } else {
                tpl.isAllChecked = false
            }

            this.$set(tplList, index, tpl)

            tpl.prepareDeleteInstances.splice(0, tpl.prepareDeleteInstances.length, ...prepareDeleteInstances)
        },

        getTemplateInsPerms (instance, actionID) {
            return this.templateInstanceWebAnnotations.perms[instance.iam_ns_id]
                && this.templateInstanceWebAnnotations.perms[instance.iam_ns_id][actionID]
        },

        /**
         * 模板集视图，批量删除
         *
         * @param {Object} tpl 当前点击的模板对象
         * @param {number} index 当前点击的模板对象的索引
         */
        batchDelete (tpl, index) {
            if (!tpl.prepareDeleteInstances || !tpl.prepareDeleteInstances.length) {
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('还未选择{tmplAppName}下的实例', { tmplAppName: tpl.tmpl_app_name })
                })
                return
            }

            this.cancelLoopInstanceList()

            // const prepareDeleteInstances = []
            // prepareDeleteInstances.splice(0, 0, ...tpl.prepareDeleteInstances)
            // const names = []
            // tpl.instanceList.forEach(item => {
            //     if (prepareDeleteInstances.indexOf(item.id) > -1) {
            //         names.push(this.$createElement('li', {
            //             style: {
            //                 width: '300px',
            //                 margin: '0 auto'
            //             }
            //         }, item.name))
            //     }
            // })

            const me = this
            me.$bkInfo({
                title: this.$t('确认删除'),
                confirmLoading: true,
                content: this.$t('已选择 {len} 个实例，确定全部删除？', {
                    len: tpl.prepareDeleteInstances.length
                }),
                width: 360,
                async confirmFn () {
                    // 用来触发组件的响应式，最终改变 me.curInstance.state
                    me.isUpdating = true
                    tpl.instanceList.forEach(item => {
                        item.state.lock()
                    })
                    try {
                        await me.$store.dispatch('app/batchDeleteInstance', {
                            projectId: me.projectId,
                            inst_id_list: tpl.prepareDeleteInstances
                        })
                        tpl.instanceList.forEach(item => {
                            item.isChecked = false
                        })
                        tpl.isAllChecked = false
                        tpl.prepareDeleteInstances.splice(0, tpl.prepareDeleteInstances.length, ...[])
                        me.$bkMessage({
                            theme: 'success',
                            message: me.$t('删除任务已经下发成功，请注意状态变化')
                        })
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.cancelLoop = false
                            me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                        }, 100)
                        setTimeout(() => {
                            tpl.instanceList.forEach(item => {
                                item.state.unlock()
                            })
                        }, 1000)
                    }
                },
                cancelFn () {
                    setTimeout(() => {
                        me.cancelLoop = false
                        me.fetchInstanceListInTemplateViewMode(...me.loopInstanceListParams)
                    }, 100)
                }
            })
        },

        /**
         * 命名空间视图，多选框全选中，for batch delete
         *
         * @param {Object} e 事件对象
         * @param {Object} namespace 当前点击的命名对象
         * @param {number} namespaceIndex 当前点击的命名对象的索引
         * @param {Array} namespaceList 当前命名空间数组
         * @param {boolean} checked 是否选中
         */
        checkAllNamespace (namespace, namespaceIndex, namespaceList, checked) {
            const appList = []
            appList.splice(0, 0, ...namespace.appList)

            const prepareDeleteInstances = []
            prepareDeleteInstances.splice(0, 0, ...[])

            appList.forEach(item => {
                if (this.namespaceInsWebAnnotations.perms[item.iam_ns_id]
                    && this.namespaceInsWebAnnotations.perms[item.iam_ns_id].namespace_scoped_delete) {
                    item.isChecked = checked
                    checked && prepareDeleteInstances.push(item)
                }
            })

            namespace.isAllChecked = checked
            namespace.prepareDeleteInstances.splice(0, namespace.prepareDeleteInstances.length, ...prepareDeleteInstances)
            this.$set(namespaceList, namespaceIndex, namespace)
        },

        /**
         * 命名空间视图，多选框选中，for batch delete
         *
         * @param {Object} instance 当前实例
         * @param {number} instanceIndex 当前实例的索引
         * @param {Object} namespace 当前点击的命名对象
         * @param {number} namespaceIndex 当前点击的命名对象的索引
         * @param {Array} namespaceList 当前命名空间数组
         * @param {boolean} checked 是否选中
         */
        checkNamespace (instance, instanceIndex, namespace, namespaceIndex, namespaceList, checked) {
            const name = instance.name
            const prepareDeleteInstances = []
            prepareDeleteInstances.splice(0, 0, ...namespace.prepareDeleteInstances)

            const prepareDeleteInstanceIndex = prepareDeleteInstances.findIndex(n => n.name === name)
            // 存在，说明这一次的点击是未选中
            if (prepareDeleteInstanceIndex > -1) {
                prepareDeleteInstances.splice(prepareDeleteInstanceIndex, 1)
            } else {
                prepareDeleteInstances.push(instance)
            }

            const allLength = namespace.appList.length
            const invalidLength = namespace.appList.filter(inst => !this.getNamespaceInsPerms(inst, 'namespace_scoped_delete')).length

            if (prepareDeleteInstances.length === allLength - invalidLength) {
                namespace.isAllChecked = true
            } else {
                namespace.isAllChecked = false
            }

            this.$set(namespaceList, namespaceIndex, namespace)

            namespace.prepareDeleteInstances.splice(0, namespace.prepareDeleteInstances.length, ...prepareDeleteInstances)
        },

        getNamespaceInsPerms (instance, actionID) {
            return this.namespaceInsWebAnnotations.perms[instance.iam_ns_id]
                && this.namespaceInsWebAnnotations.perms[instance.iam_ns_id][actionID]
        },

        /**
         * 命名空间视图，批量删除
         *
         * @param {Object} namespace 当前点击的命名空间对象
         * @param {number} namespaceIndex 当前点击的命名空间对象的索引
         */
        batchDelete4Namespace (namespace, namespaceIndex) {
            if (!namespace.prepareDeleteInstances || !namespace.prepareDeleteInstances.length) {
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('还未选择{namespaceName}下的应用', { namespaceName: namespace.name })
                })
                return
            }
            this.cancelLoopAppList()

            const me = this
            me.$bkInfo({
                title: this.$t('确认删除'),
                confirmLoading: true,
                content: this.$t('已选择 {len} 个实例，确定全部删除？', {
                    len: namespace.prepareDeleteInstances.length
                }),
                width: 360,
                async confirmFn () {
                    // 用来触发组件的响应式，最终改变 me.curInstance.state
                    me.isUpdating = true
                    namespace.appList.forEach(item => {
                        item.state.lock()
                    })
                    try {
                        const resourceList = []
                        namespace.prepareDeleteInstances.forEach(n => {
                            if (n.id === 0) {
                                resourceList.push({
                                    resource_kind: n.category,
                                    name: n.name,
                                    namespace: namespace.name,
                                    cluster_id: namespace.cluster_id
                                })
                            }
                        })
                        await me.$store.dispatch('app/batchDeleteInstance', {
                            projectId: me.projectId,
                            inst_id_list: namespace.prepareDeleteInstances.filter(n => n.id !== 0).map(n => n.id),
                            resource_list: resourceList
                        })
                        namespace.appList.forEach(item => {
                            item.isChecked = false
                        })
                        namespace.isAllChecked = false
                        namespace.prepareDeleteInstances.splice(0, namespace.prepareDeleteInstances.length, ...[])
                        me.$bkMessage({
                            theme: 'success',
                            message: me.$t('删除任务已经下发成功，请注意状态变化')
                        })
                    } catch (e) {
                        me.bkMessageInstance = me.$bkMessage({
                            theme: 'error',
                            message: e.message || e.data.msg || e.statusText
                        })
                    } finally {
                        me.isUpdating = false
                        setTimeout(() => {
                            me.cancelLoop = false
                            me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                        }, 100)
                        setTimeout(() => {
                            namespace.appList.forEach(item => {
                                item.state.unlock()
                            })
                        }, 1000)
                    }
                },
                cancelFn () {
                    setTimeout(() => {
                        me.cancelLoop = false
                        me.fetchAppListInNamespaceViewMode(...me.loopAppListParams)
                    }, 100)
                }
            })
        },

        /**
         * 刷新
         */
        refresh () {
            if (this.showLoading) {
                return false
            }
            this.search = ''
            this.cancelLoopAppList()
            this.cancelLoopInstanceList()
            this.fetchData(false)
        },

        /**
         * 搜索事件
         */
        handleSearch () {
            const search = String(this.search || '').trim().toLowerCase()
            if (this.viewMode === 'template') {
                if (!search) {
                    this.tmplMusterListTmp.forEach(item => {
                        item.isOpen = false
                    })
                    this.tmplMusterList.splice(0, this.tmplMusterList.length, ...this.tmplMusterListTmp)
                    return
                }

                let inTmplList = false
                const results = this.tmplMusterListTmp.filter((tmplMuster, tmplMusterIndex) => {
                    inTmplList = false
                    const templateList = tmplMuster.templateList || []
                    templateList.forEach(tpl => {
                        const valid = (tpl.tmpl_app_name || '').toLowerCase().indexOf(search) > -1
                        if (valid) {
                            inTmplList = true
                        }
                    })
                    tmplMuster.isOpen = inTmplList
                    this.$set(this.tmplMusterList, tmplMusterIndex, tmplMuster)
                    return (tmplMuster.tmpl_muster_name || '').toLowerCase().indexOf(search) > -1 || inTmplList
                })
                this.tmplMusterList.splice(0, this.tmplMusterList.length, ...results)
            } else {
                if (!search) {
                    this.namespaceListTmp.forEach(item => {
                        item.isOpen = false
                    })
                    this.namespaceList.splice(0, this.namespaceList.length, ...this.namespaceListTmp)
                    return
                }

                let inTmplList = false
                const results = this.namespaceListTmp.filter((namespace, namespaceIndex) => {
                    inTmplList = false
                    const appList = namespace.appList || []
                    appList.forEach(app => {
                        const valid = (app.name || '').toLowerCase().indexOf(search) > -1
                        if (valid) {
                            inTmplList = true
                        }
                    })
                    namespace.isOpen = inTmplList
                    this.$set(this.namespaceList, namespaceIndex, namespace)
                    return (namespace.name || '').toLowerCase().indexOf(search) > -1 || inTmplList
                })
                this.namespaceList.splice(0, this.namespaceList.length, ...results)
            }
        },

        /**
         * 跳转到模板编辑页面
         *
         * @param {Object} tmplMuster 当前模板集对象
         * @param {Object} tpl 当前模板对象
         */
        goEditTemplate (tmplMuster, tpl) {
            const name = tpl.category === 'application'
                ? 'createTemplatesetApplication'
                : 'createTemplatesetDeployment'
            this.$router.push({
                name: name,
                params: {
                    projectId: this.projectId,
                    projectCode: this.projectCode,
                    templateId: tmplMuster.tmpl_muster_id
                }
            })
        },

        /**
         * 模板集视图，显示批量重建
         *
         * @param {Object} tpl 当前点击的模板对象
         * @param {number} index 当前点击的模板对象的索引
         */
        showBatchReBuild (tpl, index) {
            const appList = this.viewMode === 'namespace' ? tpl.appList : tpl.instanceList
            const checkedInstances = appList.filter(item => item.isChecked)
            if (!checkedInstances || !checkedInstances.length) {
                this.bkMessageInstance = this.$bkMessage({
                    theme: 'error',
                    message: this.$t('还未选择{tmplAppName}下的实例', { tmplAppName: tpl.tmpl_app_name })
                })
                return
            }

            if (this.viewMode === 'namespace') {
                this.cancelLoopAppList()
            } else {
                this.cancelLoopInstanceList()
            }

            const list = []
            checkedInstances.forEach(item => {
                list.push(item)
            })
            this.batchRebuildDialogConf.list.splice(0, this.batchRebuildDialogConf.list.length, ...list)
            this.batchRebuildDialogConf.isShow = true
            // 这里需要引用赋值，便于之后 设置 tpl.isAllChecked 为 false
            this.batchRebuildDialogConf.tpl = tpl
        },

        /**
         * 模板集视图，隐藏批量重建
         */
        hideBatchReBuild () {
            this.batchRebuildDialogConf.isShow = false
            setTimeout(() => {
                this.cancelLoop = false
                if (this.viewMode === 'namespace') {
                    this.fetchAppListInNamespaceViewMode(...this.loopAppListParams)
                } else {
                    this.fetchInstanceListInTemplateViewMode(...this.loopInstanceListParams)
                }
            }, 100)
            setTimeout(() => {
                this.batchRebuildDialogConf.list.splice(0, this.batchRebuildDialogConf.list.length, ...[])
                this.batchRebuildDialogConf.tpl = null
            }, 300)
        },

        /**
         * 批量重建确认
         */
        async batchRebuildConfirm () {
            const data = { resource_list: [] }
            this.batchRebuildDialogConf.list.forEach(item => {
                data.resource_list.push({
                    resource_kind: item.category,
                    name: item.name,
                    namespace: item.namespace,
                    cluster_id: item.cluster_id
                })
            })

            const params = {
                projectId: this.projectId,
                data
            }

            this.isUpdating = true
            this.curInstance.state && this.curInstance.state.lock()
            try {
                await this.$store.dispatch('app/batchRebuild', params)

                this.batchRebuildDialogConf.tpl[this.viewMode === 'namespace' ? 'appList' : 'instanceList'].forEach(item => {
                    item.isChecked = false
                })
                this.batchRebuildDialogConf.tpl.isAllChecked = false
                this.batchRebuildDialogConf.tpl.prepareDeleteInstances.splice(
                    0,
                    this.batchRebuildDialogConf.tpl.prepareDeleteInstances.length,
                    ...[]
                )

                this.bkMessageInstance = this.$bkMessage({
                    theme: 'success',
                    message: this.$t('任务下发成功')
                })
            } catch (e) {
                console.log(e)
            } finally {
                this.isUpdating = false
                this.hideBatchReBuild()
                setTimeout(() => {
                    this.curInstance.state && this.curInstance.state.unlock()
                }, 1000)
            }
        }
    }
}
