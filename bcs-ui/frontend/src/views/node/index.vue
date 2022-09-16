<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-app-title">
                {{$t('节点')}}
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper biz-node-loading biz-node-content-wrapper" style="padding: 0;" v-bkloading="{ isLoading: pageLoading, opacity: 0.1 }">
            <app-exception
                v-if="exceptionCode && !pageLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <template v-if="!exceptionCode && !pageLoading">
                <div class="biz-panel-header biz-node-query">
                    <div class="left">
                        <bcs-button :disabled="!checkedNodeList.length" @click="showSetLabel">
                            <span>{{$t('设置标签')}}</span>
                        </bcs-button>
                        <bk-button @click="exportNode">
                            <span>{{$t('导出')}}</span>
                        </bk-button>
                        <bk-dropdown-menu :align="'left'" ref="copyIpDropdownMenu" class="copy-ip-dropdown">
                            <a href="javascript:void(0);" slot="dropdown-trigger" class="bk-text-button copy-ip-btn">
                                <span class="label">{{$t('复制IP')}}</span>
                                <i class="bcs-icon bcs-icon-angle-down dropdown-menu-angle-down"></i>
                            </a>
                            <ul class="bk-dropdown-list" slot="dropdown-content">
                                <li>
                                    <a href="javascript:void(0)" @click="copyIp('selected')" class="selected" :class="!checkedNodeList.length ? 'disabled' : ''">{{$t('复制所选IP')}}</a>
                                </li>
                                <li>
                                    <a href="javascript:void(0)" @click="copyIp('cur-page')" class="cur-page">{{$t('复制当前页IP')}}</a>
                                </li>
                                <li>
                                    <a href="javascript:void(0)" @click="copyIp('all')" class="all">{{$t('复制所有IP')}}</a>
                                </li>
                            </ul>
                        </bk-dropdown-menu>
                    </div>
                    <div class="right">
                        <bk-selector
                            class="cluster-selector"
                            v-if="!curClusterId"
                            :placeholder="$t('请选择')"
                            :searchable="true"
                            :setting-key="'cluster_id'"
                            :display-key="'name'"
                            :selected.sync="curSelectedClusterId"
                            :list="clusterList"
                            @item-selected="changeCluster">
                        </bk-selector>
                        <div class="biz-searcher-wrapper">
                            <node-searcher
                                :cluster-id="clusterId"
                                :project-id="projectId"
                                had-search-data
                                :search-labels-data="searchLabelsData"
                                ref="searcher"
                                @search="searchNodeList">
                            </node-searcher>
                        </div>
                        <span class="close-wrapper">
                            <template v-if="$refs.searcher && $refs.searcher.searchParams && $refs.searcher.searchParams.length">
                                <button class="bk-button bk-default is-outline is-icon" :title="$t('清除')" @click="clearSearchParams" style="border: 1px solid #c4c6cc;">
                                    <i class="bcs-icon bcs-icon-close"></i>
                                </button>
                            </template>
                            <template v-else>
                                <button class="bk-button bk-default is-outline is-icon">
                                </button>
                            </template>
                        </span>
                        <span class="refresh-wrapper">
                            <bcs-popover class="refresh" :content="$t('重置')" :transfer="true" :placement="'top-end'">
                                <button class="bk-button bk-default is-outline is-icon" @click="refresh">
                                    <i class="bcs-icon bcs-icon-refresh"></i>
                                </button>
                            </bcs-popover>
                        </span>
                    </div>
                </div>
                <div class="biz-node-list biz-table-wrapper">
                    <bk-table
                        v-bkloading="{ isLoading: showLoading, opacity: 0.9 }"
                        :key="renderTableIndex"
                        :data="curPageData"
                        :pagination="pageConf"
                        @page-change="handlePageChange"
                        @page-limit-change="handlePageLimitChange">
                        <bk-table-column :render-header="renderSelectionHeader" width="60">
                            <template slot-scope="{ row, $index }">
                                <bcs-checkbox v-model="row.isChecked" @change="checkNode(row, $index)"></bcs-checkbox>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('主机名/IP')" prop="name" :show-overflow-tooltip="true" width="200">
                            <template slot-scope="{ row }">
                                <a v-if="row.status === 'RUNNING'"
                                    href="javascript:void(0)"
                                    class="bk-text-button"
                                    @click="goNodeOverview(row)">
                                    {{row.inner_ip}}
                                </a>
                                <span v-else>{{ row.inner_ip }}</span>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('状态')" prop="cluster_name" width="200">
                            <template slot-scope="{ row }">
                                <loading-cell :style="{ left: 0 }"
                                    :ext-cls="['bk-spin-loading-mini', 'bk-spin-loading-danger']"
                                    v-if="['INITIALIZATION', 'DELETING'].includes(row.status)"
                                ></loading-cell>
                                <StatusIcon :status="row.status" :status-color-map="nodeStatusColorMap" v-else>
                                    {{ statusMap[row.status.toLowerCase()] }}
                                </StatusIcon>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('标签')" prop="source_type" :show-overflow-tooltip="false">
                            <template slot-scope="{ row, $index }">
                                <bcs-popover v-if="row.transformLabels.length" :delay="300" placement="left-end">
                                    <div class="label-list" :ref="`label_${pageConf.current}_${$index}`">
                                        <div v-for="(item, index) in row.transformLabels"
                                            class="label-item"
                                            :key="index">
                                            <span class="key">{{item.key}}</span> =
                                            <span class="value">{{item.value}}</span>
                                        </div>
                                        <span v-if="row.showExpand" class="ellipsis">...</span>
                                    </div>
                                    <template slot="content">
                                        <div class="label-tips">
                                            <div class="label-item" v-for="(taint, index) in row.transformLabels" :key="index">
                                                <span class="key">{{taint.key}}</span> =
                                                <span class="value">{{taint.value}}</span>
                                            </div>
                                        </div>
                                    </template>
                                </bcs-popover>
                                <template v-else>--</template>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('污点')" :show-overflow-tooltip="false">
                            <template slot-scope="{ row, $index }">
                                <bcs-popover v-if="row.transformTaints.length" :delay="300" placement="left">
                                    <div class="label-list" :ref="`taint_${pageConf.current}_${$index}`">
                                        <div v-for="(item, index) in row.transformTaints"
                                            class="label-item"
                                            :key="index">
                                            <span class="key">{{item.key}}</span>=
                                            <span class="value">{{item.displayValue}}</span>
                                        </div>
                                        <span v-if="row.showTaintExpand" class="ellipsis">...</span>
                                    </div>
                                    <template slot="content">
                                        <div class="label-tips">
                                            <div class="label-item" v-for="(taint, index) in row.transformTaints" :key="index">
                                                <span class="key">{{taint.key}}</span> =
                                                <span class="value">{{taint.displayValue}}</span>
                                            </div>
                                        </div>
                                    </template>
                                </bcs-popover>
                                <template v-else>--</template>
                            </template>
                        </bk-table-column>
                        <bk-table-column :label="$t('操作')" prop="permissions" width="240">
                            <template slot-scope="{ row }">
                                <bk-button text
                                    :disabled="row.status !== 'RUNNING'"
                                    @click.stop="showSetLabelInRow(row)">
                                    {{$t('设置标签')}}
                                </bk-button>
                                <bk-button text @click.stop="showTaintDialog(row)">{{$t('设置污点')}}</bk-button>
                                <bk-button text @click.stop="goClusterNode(row)">{{$t('更多操作')}}</bk-button>
                            </template>
                        </bk-table-column>
                    </bk-table>
                </div>
            </template>
        </div>

        <bk-sideslider
            :is-show.sync="setLabelConf.isShow"
            :title="setLabelConf.title"
            :width="setLabelConf.width"
            :quick-close="false"
            class="biz-cluster-set-label-sideslider"
            @hidden="hideSetLabel">
            <div slot="content">
                <div class="title-tip">{{$t('标签有助于整理你的资源（如 env:prod）')}}</div>
                <div class="wrapper" style="position: relative;">
                    <form class="bk-form bk-form-vertical set-label-form">
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('键')}}：</label>
                            </div>
                            <div class="right">
                                <label class="bk-label label">{{$t('值')}}：
                                    <template v-if="showMixinTip">
                                        <bcs-popover :delay="300" placement="top">
                                            <i class="bcs-icon bcs-icon-question-circle" style="vertical-align: middle;"></i>
                                            <div slot="content">
                                                <p class="app-biz-node-label-tip-content">{{$t('为什么会有混合值')}}：</p>
                                                <p class="app-biz-node-label-tip-content">{{$t('已选节点的标签中存在同一个键对应多个值')}}</p>
                                            </div>
                                        </bcs-popover>
                                    </template>
                                </label>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <div class="bk-form-content">
                                <div class="biz-key-value-wrapper mb10">
                                    <div class="biz-key-value-item" v-for="(label, index) in labelList" :key="index">
                                        <template v-if="label.key && label.fromData">
                                            <bk-input style="width: 280px;" disabled v-model="label.key" />
                                        </template>
                                        <template v-else>
                                            <bk-input style="width: 280px;" :placeholder="$t('键')" maxlength="30" v-model="label.key" />
                                        </template>
                                        <span class="equals-sign">=</span>
                                        <bk-input style="width: 280px; margin-left: 35px;" maxlength="30" :placeholder="$t('混合值')" v-if="label.isMixin" v-model="label.value" />
                                        <bk-input style="width: 280px; margin-left: 35px;" maxlength="30" :placeholder="$t('值')" v-else-if="!label.value" v-model="label.value" />
                                        <bk-input style="width: 280px; margin-left: 35px;" maxlength="30" :placeholder="$t('值')" v-else v-model="label.value" />

                                        <template v-if="labelList.length === 1">
                                            <button class="action-btn">
                                                <i class="bk-icon icon-plus-circle" @click.stop.prevent="addLabel"></i>
                                            </button>
                                        </template>
                                        <template v-else>
                                            <template v-if="index === labelList.length - 1">
                                                <button class="action-btn" @click.stop.prevent>
                                                    <i class="bk-icon icon-plus-circle mr5" @click.stop.prevent="addLabel"></i>
                                                    <i class="bk-icon icon-minus-circle" @click.stop.prevent="delLabel(label, index)"></i>
                                                </button>
                                            </template>
                                            <template v-else>
                                                <button class="action-btn">
                                                    <i class="bk-icon icon-plus-circle mr5" @click.stop.prevent="addLabel"></i>
                                                    <i class="bk-icon icon-minus-circle" @click.stop.prevent="delLabel(label, index)"></i>
                                                </button>
                                            </template>
                                        </template>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="action-inner">
                            <bk-button type="primary" :loading="setLabelConf.loading" @click="confirmSetLabel">
                                {{$t('保存')}}
                            </bk-button>
                            <bk-button :disalbed="setLabelConf.loading" @click="hideSetLabel">
                                {{$t('取消')}}
                            </bk-button>
                        </div>
                    </form>
                </div>
            </div>
        </bk-sideslider>

        <bk-sideslider
            :is-show.sync="taintDialog.isShow"
            :title="$t('设置污点')"
            :width="750"
            :quick-close="false">
            <div slot="content">
                <TaintContent
                    :cluster-id="curSelectedClusterId"
                    :nodes="taintDialog.nodes"
                    @cancel="handleHideTaintDialog" />
            </div>
        </bk-sideslider>
    </div>
</template>

<script>
    import axios from 'axios'
    import Clipboard from 'clipboard'
    import { catchErrorHandler } from '@/common/util'
    import LoadingCell from '../cluster/loading-cell'
    import nodeSearcher from '../cluster/searcher'
    import TaintContent from './taint.vue'
    import StatusIcon from '@/views/dashboard/common/status-icon.tsx'

    export default {
        components: {
            StatusIcon,
            LoadingCell,
            nodeSearcher,
            TaintContent
        },
        data () {
            return {
                showLoading: false,
                pageLoading: false,
                nodeList: [],
                renderTableIndex: 0,
                curPageData: [],
                // for search
                nodeListTmp: [],
                pageConf: {
                    count: 1,
                    limit: 10,
                    current: 1,
                    show: true
                },
                // 已选择的 nodeList
                checkedNodeList: [],
                // 单行 node
                curRowNode: {},
                setLabelConf: {
                    isShow: false,
                    title: this.$t('设置标签'),
                    width: 750,
                    loading: false
                },
                // 设置标签的参数
                labelList: [{ key: '', value: '' }],
                // 节点列表是否全选
                isCheckAllNode: false,
                // 是否显示混合值的提示
                showMixinTip: false,
                enableSetLabel: false,
                exceptionCode: null,
                timer: null,
                curSelectedClusterName: '',
                curSelectedClusterId: '',
                alreadySelectedNums: 0,
                searchParams: [],
                clipboardInstance: null,
                vueInstanceIsDestroy: false,
                taintDialog: {
                    isShow: false,
                    nodes: []
                },
                setInfos: [
                    {
                        key: 'showExpand',
                        label: 'label'
                    },
                    {
                        key: 'showTaintExpand',
                        label: 'taint'
                    }
                ],
                statusMap: {
                    initialization: this.$t('初始化中'),
                    running: this.$t('正常'),
                    deleting: this.$t('删除中'),
                    'add-failure': this.$t('上架失败'),
                    'remove-failure': this.$t('下架失败'),
                    removable: this.$t('不可调度'),
                    notready: this.$t('不正常'),
                    unknown: this.$t('未知状态')
                },
                nodeStatusColorMap: {
                    initialization: 'blue',
                    running: 'green',
                    deleting: 'blue',
                    'add-failure': 'red',
                    'remove-failure': 'red',
                    removable: '',
                    notready: 'red',
                    unknown: ''
                }
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            curProject () {
                return this.$store.state.curProject
            },
            labelDocUrl () {
                return this.curProject.kind === window.PROJECT_K8S ? this.PROJECT_CONFIG.doc.nodeLabelK8s : this.PROJECT_CONFIG.doc.nodeLabelMesos
            },
            isEn () {
                return this.$store.state.isEn
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            searchLabelsData () {
                const res = {}
                for (const node of this.nodeList) {
                    Object.keys(node.labels || {}).forEach(key => {
                        const value = node.labels[key]
                        if (Object.prototype.hasOwnProperty.call(res, key)) {
                            res[key].push(value)
                        } else {
                            res[key] = [value]
                        }
                    })
                }
                Object.keys(res).forEach(key => {
                    res[key] = [...new Set(res[key])]
                })
                return res
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            }
        },
        watch: {
            'checkedNodeList.length' (len) {
                this.enableSetLabel = !!len
                this.alreadySelectedNums = len
            }
        },
        beforeDestroy () {
            this.vueInstanceIsDestroy = true
            if (this.timer) {
                clearTimeout(this.timer)
                this.timer = null
            }

            this.clipboardInstance && this.clipboardInstance.destroy()
            if (this.clipboardInstance && this.clipboardInstance.off) {
                this.clipboardInstance.off('success')
            }
        },
        destroyed () {
            this.clipboardInstance && this.clipboardInstance.destroy()
            if (this.clipboardInstance && this.clipboardInstance.off) {
                this.clipboardInstance.off('success')
            }
        },
        async created () {
            this.vueInstanceIsDestroy = false
            this.pageConf.current = 1
            this.pageLoading = true
            await this.fetchData()
        },
        methods: {
            /**
             * 获取所有的集群
             */
            async getClusters () {
                try {
                    const list = this.clusterList
                    if (this.curClusterId) {
                        const match = list.find(item => {
                            return item.cluster_id === this.curClusterId
                        })
                        this.curSelectedClusterName = match ? match.name : this.$t('全部集群')
                        this.curSelectedClusterId = match ? match.cluster_id : 'all'
                    } else {
                        this.curSelectedClusterName = list.length ? list[0].name : this.$t('全部集群')
                        this.curSelectedClusterId = list.length ? list[0].cluster_id : 'all'
                    }
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 切换集群条件
             *
             * @param {Object} cluster 集群
             */
            async changeCluster (clusterId, cluster) {
                this.curSelectedClusterName = cluster.name
                this.curSelectedClusterId = clusterId
                this.pageConf.current = 1
                this.showLoading = true
                await this.fetchData(true)
                this.showLoading = false
            },

            /**
             * 分页大小更改
             *
             * @param {number} pageSize pageSize
             */
            handlePageLimitChange (pageSize) {
                this.renderTableIndex++
                this.pageConf.limit = pageSize
                this.pageConf.current = 1
                this.initPageConf()
                this.handlePageChange()
            },

            /**
             * 格式化日志
             *
             * @param {string} log 日志内容
             *
             * @return {strin} 格式化后的日志内容
             */
            formatLog (log) {
                // 换行
                log = log.replace(/##/ig, '<p class="html-tag"></p>').replace(/\|/ig, '<p class="html-tag"></p>')
                // 着色
                log = log.replace(/(Failed)/ig, '<span class="biz-danger-text">$1</span>')
                log = log.replace(/(OK)/ig, '<span class="biz-success-text">$1</span>')
                return log
            },

            /**
             * 查询节点列表数据
             *
             * @param {boolean} isPolling 是否是轮询
             */
            async fetchData (isPolling) {
                if (this.vueInstanceIsDestroy) return
                if (!isPolling) {
                    await this.getClusters()
                    this.showLoading = false
                }
                if (!this.clusterList.length) {
                    this.pageLoading = false
                    return
                }

                try {
                    const params = {
                        $clusterId: this.curSelectedClusterId
                    }
                    const res = await this.$store.dispatch('cluster/getK8sNodes', params)

                    const list = res

                    const nodeList = []
                    list.forEach(item => {
                        item.transformLabels = []
                        Object.entries(item.labels || {}).forEach(entries => {
                            item.transformLabels.push({
                                key: entries[0],
                                value: entries[1]
                            })
                        })
                        item.transformTaints = []
                        for (const taint of (item.taints || [])) {
                            item.transformTaints.push(Object.assign({}, taint, {
                                displayValue: taint.value && taint.effect ? taint.value + ' : ' + taint.effect : taint.value || taint.effect
                            }))
                        }

                        item.isExpandLabels = false
                        // 是否显示标签的展开按钮
                        item.showExpand = false
                        item.showTaintExpand = false
                        nodeList.push(item)
                    })

                    this.isCheckAllNode = false

                    this.nodeList.splice(0, this.nodeList.length, ...nodeList)
                    this.nodeListTmp.splice(0, this.nodeListTmp.length, ...nodeList)

                    if (this.curSelectedClusterId !== 'all') {
                        const newNodeList = []
                        newNodeList.splice(0, 0, ...this.nodeListTmp.filter(
                            node => node.cluster_id === this.curSelectedClusterId
                        ))

                        this.nodeList.splice(0, this.nodeList.length, ...newNodeList)
                    }

                    // this.initPageConf()
                    // this.curPageData = this.getDataByPage(this.pageConf.current)
                    if (this.$refs.searcher
                        && this.$refs.searcher.searchParams
                        && this.$refs.searcher.searchParams.length
                    ) {
                        this.searchNodeList(this.pageConf.current)
                    } else {
                        this.initPageConf()
                        this.curPageData = this.getDataByPage(this.pageConf.current, false)
                    }
                    setTimeout(() => {
                        this.curPageData.forEach((item, index) => {
                            this.setInfos.forEach(info => {
                                const el = this.$refs[`${info.label}_${this.pageConf.current}_${index}`]
                                item[info.key] = el && (el.offsetHeight < el.scrollHeight)
                            })
                        })
                    }, 0)

                    const checkNodeIdList = this.checkedNodeList.map(node => node.inner_ip)
                    this.nodeList.forEach(node => {
                        if (node.status === 'RUNNING') {
                            this.$set(node, 'isChecked', checkNodeIdList.indexOf(node.inner_ip) > -1)
                        }
                    })
                    // 当前页选中的
                    const selectedNodeList = this.curPageData.filter(node => node.isChecked === true)
                    // 当前页合法的
                    const validList = this.curPageData.filter(
                        node => node.status === 'RUNNING'
                    )
                    this.isCheckAllNode = selectedNodeList.length === validList.length

                    if (this.timer) {
                        clearTimeout(this.timer)
                        this.timer = null
                    }
                    this.timer = setTimeout(() => {
                        this.fetchData(true)
                    }, 30000)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.pageLoading = false
                    this.showLoading = false
                }
            },

            /**
             * 初始化翻页条
             */
            initPageConf () {
                const total = this.nodeList.length
                this.pageConf.count = total
                this.pageConf.show = true
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.limit)
            },

            /**
             * 获取当前这一页的数据
             *
             * @param {number} page 当前页
             *
             * @return {Array} 当前页数据
             */
            getDataByPage (page, clearCheck = true) {
                let startIndex = (page - 1) * this.pageConf.limit
                let endIndex = page * this.pageConf.limit
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.nodeList.length) {
                    endIndex = this.nodeList.length
                }
                if (clearCheck) this.checkedNodeList = []
                const data = this.nodeList.slice(startIndex, endIndex)
                return data
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            handlePageChange (page = 1) {
                this.pageConf.current = page
                const data = this.getDataByPage(page)
                this.curPageData.splice(0, this.curPageData.length, ...data)

                // 当前页选中的
                const selectedNodeList = this.curPageData.filter(node => node.isChecked === true)

                // 当前页合法的
                const validList = this.curPageData.filter(
                    node => node.status === 'RUNNING'
                )

                this.isCheckAllNode = selectedNodeList.length === validList.length

                setTimeout(() => {
                    this.curPageData.forEach((item, index) => {
                        this.setInfos.forEach(info => {
                            const el = this.$refs[`${info.label}_${this.pageConf.current}_${index}`]
                            item[info.key] = el && (el.offsetHeight < el.scrollHeight)
                        })
                    })
                }, 0)
            },

            /**
             * 清除 searcher 搜索条件
             */
            clearSearchParams () {
                this.$refs.searcher.clear()
                this.getSearchParams()
            },

            /**
             * 手动刷新表格数据
             */
            async refresh () {
                this.showLoading = true
                this.clearSearchParams()
                await this.fetchData()
            },

            /**
             * 延迟
             *
             * @param {Number} ms 毫秒数
             */
            timeout (ms) {
                return new Promise(resolve => setTimeout(resolve, ms))
            },

            /**
             * 获取 searcher 的参数
             *
             * @return {Object} 参数
             */
            getSearchParams () {
                const searchParams = (this.$refs.searcher && this.$refs.searcher.searchParams) || []
                const ipParams = searchParams.filter(item => item.id === 'ip').map(
                    item => item.valueArr.join(',')
                ).join(',')

                const labelsParams = searchParams.filter(item => item.id === 'labels')
                const labels = []
                labelsParams.forEach(label => {
                    label.valueArr.forEach(item => {
                        labels.push({
                            [`${label.key}`]: item
                        })
                    })
                })

                const statusListParams = searchParams.filter(item => item.id === 'status_list')
                const statusMap = {}
                statusListParams.forEach(statusItem => {
                    statusItem.valueArr.forEach(statusVal => {
                        statusMap[statusVal] = 1
                    })
                })

                return { ipParams, labels, statusList: Object.keys(statusMap) }
            },

            /**
             * nodeList 搜索
             */
            async searchNodeList (page = 1) {
                this.showLoading = true
                await this.timeout(1000)
                this.showLoading = false

                const newNodeList = []
                if (this.curSelectedClusterId === 'all') {
                    newNodeList.splice(0, 0, ...this.nodeListTmp)
                } else {
                    newNodeList.splice(0, 0, ...this.nodeListTmp.filter(
                        node => node.cluster_name === this.curSelectedClusterName
                    ))
                }

                const searchParams = this.getSearchParams()
                const ipParams = searchParams.ipParams || ''
                const ipList = ipParams
                    ? searchParams.ipParams.split(',')
                    : []

                const sLabels = searchParams.labels
                const len = sLabels.length

                const statusList = searchParams.statusList || []
                const statusListLen = statusList.length

                const results = []

                // 没有搜索条件，那么就是全部
                if (!ipList.length && !len && !statusListLen) {
                    results.splice(0, 0, ...newNodeList)
                } else {
                    const resultMap = {}
                    if (ipList.length) {
                        newNodeList.forEach(node => {
                            ipList.forEach(ip => {
                                if (String(node.inner_ip || '').toLowerCase() === ip) {
                                    resultMap[node.inner_ip] = node
                                }
                            })
                        })
                    } else {
                        newNodeList.forEach(node => {
                            resultMap[node.inner_ip] = node
                        })
                    }
                    Object.keys(resultMap).forEach(ip => {
                        const labels = Object.keys(resultMap[ip].labels).map(key => ({ [key]: resultMap[ip].labels[key] }))
                        for (let i = 0; i < len; i++) {
                            if (!labels.filter(
                                label => JSON.stringify(label) === JSON.stringify(sLabels[i])).length
                            ) {
                                delete resultMap[ip]
                                continue
                            }
                        }
                    })

                    if (statusListLen) {
                        Object.keys(resultMap).forEach(ip => {
                            if (statusList.indexOf(resultMap[ip].status) < 0) {
                                delete resultMap[ip]
                            }
                        })
                    }

                    Object.keys(resultMap).forEach(key => {
                        results.push(resultMap[key])
                    })
                }

                this.nodeList.splice(0, this.nodeList.length, ...results)

                this.pageConf.current = page

                this.initPageConf()
                this.curPageData = this.getDataByPage(this.pageConf.current)

                const checkNodeIdList = this.checkedNodeList.map(node => node.inner_ip)
                this.curPageData.forEach(item => {
                    if (item.permissions && item.permissions.edit && item.status === 'RUNNING') {
                        item.isChecked = checkNodeIdList.indexOf(item.inner_ip) > -1
                    }
                })

                // 当前页选中的
                const selectedNodeList = this.curPageData.filter(node => node.isChecked === true)

                // 当前页合法的
                const validList = this.curPageData.filter(
                    node => node.status === 'RUNNING'
                )

                this.isCheckAllNode = selectedNodeList.length === validList.length
            },

            /**
             * 节点列表行选中
             *
             * @param {Object} e 事件对象
             */
            nodeRowClick (e) {
                let target = e.target
                while (target.nodeName.toLowerCase() !== 'tr') {
                    target = target.parentNode
                }
                const checkboxNode = target.querySelector('input[type="checkbox"]')
                checkboxNode && checkboxNode.click()
            },

            /**
             * 节点列表全选
             */
            checkAllNode (value) {
                const isChecked = value
                this.curPageData.forEach(node => {
                    if (node.status === 'RUNNING') {
                        node.isChecked = isChecked
                    }
                })
                const checkedNodeList = []
                checkedNodeList.splice(0, 0, ...this.checkedNodeList)
                // 用于区分是否已经选择过
                const hasCheckedList = checkedNodeList.map(item => item.inner_ip)
                if (isChecked) {
                    const checkedList = this.curPageData.filter(
                        node => node.status === 'RUNNING' && !hasCheckedList.includes(node.inner_ip)
                    )
                    checkedNodeList.push(...checkedList)
                    this.checkedNodeList.splice(0, this.checkedNodeList.length, ...checkedNodeList)
                } else {
                    // 当前页所有合法的 node inner_ip 集合
                    const validIdList = this.curPageData.filter(
                        node => node.status === 'RUNNING'
                    ).map(node => node.inner_ip)

                    const newCheckedNodeList = []
                    this.checkedNodeList.forEach(checkedNode => {
                        if (validIdList.indexOf(checkedNode.inner_ip) < 0) {
                            newCheckedNodeList.push(JSON.parse(JSON.stringify(checkedNode)))
                        }
                    })
                    this.checkedNodeList.splice(0, this.checkedNodeList.length, ...newCheckedNodeList)
                }
            },

            isCheckAllDisabled () {
                return this.curPageData.every(i => !i.permissions.edit)
            },

            /**
             * 节点列表每一行的 checkbox 点击
             *
             * @param {Object} node 当前节点即当前行
             */
            checkNode (node) {
                this.$nextTick(() => {
                    // 当前页选中的
                    const selectedNodeList = this.curPageData.filter(node => node.isChecked === true)
                    console.log(this.curPageData, 'selectedNodeList')
                    // 当前页合法的
                    const validList = this.curPageData.filter(
                        node => node.status === 'RUNNING'
                    )
                    this.isCheckAllNode = selectedNodeList.length === validList.length

                    const checkedNodeList = []
                    if (node.isChecked) {
                        checkedNodeList.splice(0, checkedNodeList.length, ...this.checkedNodeList)
                        if (!this.checkedNodeList.filter(checkedNode => checkedNode.inner_ip === node.inner_ip).length) {
                            checkedNodeList.push(node)
                        }
                    } else {
                        this.checkedNodeList.forEach(checkedNode => {
                            if (checkedNode.inner_ip !== node.inner_ip) {
                                checkedNodeList.push(JSON.parse(JSON.stringify(checkedNode)))
                            }
                        })
                    }
                    this.checkedNodeList.splice(0, this.checkedNodeList.length, ...checkedNodeList)
                    console.log(this.checkedNodeList, 'checkedNodeList 单选')
                })
            },

            /**
             * 复制 IP
             *
             * @param {string} idx 复制的标识
             */
            copyIp (idx) {
                this.$refs.copyIpDropdownMenu && this.$refs.copyIpDropdownMenu.hide()
                let successMsg = ''
                // 复制所选 ip
                if (idx === 'selected') {
                    this.clipboardInstance = new Clipboard('.copy-ip-dropdown .selected', {
                        text: trigger => this.checkedNodeList.map(checkedNode => checkedNode.inner_ip).join('\n')
                    })
                    successMsg = this.$t('复制 {len} 个IP成功', { len: this.checkedNodeList.length })
                } else if (idx === 'cur-page') {
                    // 复制当前页 IP
                    this.clipboardInstance = new Clipboard('.copy-ip-dropdown .cur-page', {
                        text: trigger => this.curPageData.map(checkedNode => checkedNode.inner_ip).join('\n')
                    })
                    successMsg = this.$t('复制当前页IP成功')
                } else if (idx === 'all') {
                    // 复制所有 IP
                    this.clipboardInstance = new Clipboard('.copy-ip-dropdown .all', {
                        text: trigger => this.nodeList.map(checkedNode => checkedNode.inner_ip).join('\n')
                    })
                    successMsg = this.$t('复制所有IP成功')
                }
                this.clipboardInstance.on('success', e => {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'success',
                        message: successMsg
                    })
                })
                setTimeout(() => {
                    this.clipboardInstance.destroy()
                }, 500)
            },

            /**
             * 处理 mesosLabels
             */
            getMesosLabls (list = []) {
                const mixinValue = '*****-----$$$$$'
                const mesosLabls = {}
                let isFirst = true
                // 取labels合集
                list.forEach(item => {
                    const nodeMap = item || {}
                    const nodeKeys = Object.keys(nodeMap)
                    for (let i = 0, len = nodeKeys.length; i < len; i++) {
                        const nodeKey = nodeKeys[i]
                        const nodeLabels = nodeMap[nodeKey].reduce((res, label) => Object.assign(res, label), {})
                        if (isFirst) {
                            Object.assign(mesosLabls, nodeLabels)
                            isFirst = false
                            continue
                        }
                        const labelKeys = Object.keys(nodeLabels)
                        labelKeys.forEach(key => {
                            if (mesosLabls[key] === undefined || mesosLabls[key] !== nodeLabels[key]) {
                                mesosLabls[key] = mixinValue
                            }
                        })
                        // 获取非混合值keys且不属于当前节点的labels
                        const resKeys = Object.keys(mesosLabls).filter(key => mesosLabls[key] !== mixinValue && !nodeLabels[key])
                        // 剩余key确认为混合值
                        for (let j = 0, resLen = resKeys.length; j < resLen; j++) {
                            mesosLabls[resKeys[j]] = mixinValue
                        }
                    }
                })
                return mesosLabls
            },

            /**
             * 显示设置节点标签 sideslider
             */
            async showSetLabel () {
                if (!this.checkedNodeList.length) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择节点')
                    })
                    return
                }

                try {
                    this.setLabelConf.isShow = true
                    this.setLabelConf.loading = true

                    const params = {
                        $clusterId: this.curSelectedClusterId,
                        node_name_list: this.checkedNodeList.map(checkedNode => checkedNode.name)
                    }

                    const res = await this.$store.dispatch('cluster/fetchK8sNodeLabels', params)

                    const labelsList = [{}]

                    // 数据格式处理成和mesos相同
                    Object.keys(res).forEach(key => {
                        labelsList[0] = Object.assign(labelsList[0], {
                            [key]: Object.entries(res[key]).map(entries => ({ [entries[0]]: entries[1] }))
                        })
                    })

                    const labels = this.getMesosLabls(labelsList)
                    const list = Object.keys(labels)
                    const labelList = []
                    if (list.length) {
                        list.forEach((key, index) => {
                            const isMixin = labels[key] === '*****-----$$$$$'
                            if (isMixin) {
                                this.showMixinTip = true
                            }
                            const value = isMixin ? '' : labels[key]
                            labelList.push({
                                key,
                                fromData: 1,
                                value: value,
                                // 是否是混合值
                                isMixin: isMixin
                            })
                        })
                    }
                    labelList.push({ key: '', value: '' })
                    this.labelList.splice(0, this.labelList.length, ...labelList)
                } catch (e) {
                    console.error(e)
                } finally {
                    setTimeout(() => {
                        this.setLabelConf.loading = false
                    }, 300)
                }
            },

            /**
             * 单行里的显示设置节点标签 sideslider
             *
             * @param {Object} node 当前节点对象
             */
            async showSetLabelInRow (node) {
                try {
                    this.setLabelConf.isShow = true
                    this.setLabelConf.loading = true
                    const params = {
                        $clusterId: this.curSelectedClusterId,
                        node_name_list: [node.name]
                    }
                    const res = await this.$store.dispatch('cluster/fetchK8sNodeLabels', params)

                    const labels = res?.[node.inner_ip]
                    const list = Object.keys(labels)
                    const labelList = []
                    if (list.length) {
                        list.forEach((key, index) => {
                            const isMixin = labels[key] === '*****-----$$$$$'
                            if (isMixin) {
                                this.showMixinTip = true
                            }
                            const value = isMixin ? '' : labels[key]
                            labelList.push({
                                key,
                                fromData: 1,
                                value: value,
                                // 是否是混合值
                                isMixin: isMixin
                            })
                        })
                    }
                    labelList.push({ key: '', value: '' })
                    this.labelList.splice(0, this.labelList.length, ...labelList)

                    this.curRowNode = Object.assign({}, node)
                } catch (e) {
                    console.error(e)
                } finally {
                    setTimeout(() => {
                        this.setLabelConf.loading = false
                    }, 300)
                }
            },

            /**
             * sideslder 里添加 label 按钮
             */
            addLabel () {
                const labelList = []
                labelList.splice(0, labelList.length, ...this.labelList)
                labelList.push({ key: '', value: '' })
                this.labelList.splice(0, this.labelList.length, ...labelList)
            },

            /**
             * sideslder 里删除 label 按钮
             *
             * @param {Object} label 当前 label 对象
             * @param {number} index 当前 label 对象索引
             */
            delLabel (label, index) {
                const labelList = []
                labelList.splice(0, labelList.length, ...this.labelList)
                labelList.splice(index, 1)
                this.labelList.splice(0, this.labelList.length, ...labelList)
            },

            /**
             * 设置标签 sideslder 确认按钮
             */
            async confirmSetLabel () {
                const labelList = []
                labelList.splice(0, labelList.length, ...this.labelList)
                const len = labelList.length
                const labelInfo = {}
                for (let i = 0; i < len; i++) {
                    const key = labelList[i].key.trim()
                    const value = labelList[i].value.trim()
                    if (labelInfo[key]) {
                        this.$bkMessage({
                            theme: 'error',
                            message: this.$t('键值【{key}】重复，请重新填写', { key: key })
                        })
                        return
                    }

                    // if (!/^[a-z0-9A-Z][\w.-]{0,61}[a-z0-9A-Z]$/.test(key)) {
                    //     this.$bkMessage({
                    //         theme: 'error',
                    //         message: '键不符合规则，以字母数字开头结尾，只能包含".", "-"的大于1位不超过63位的字符串'
                    //     })
                    //     return
                    // }

                    // if (!/^[a-z0-9A-Z][\w.-]{0,61}[a-z0-9A-Z]$/.test(value)) {
                    //     this.$bkMessage({
                    //         theme: 'error',
                    //         message: '值不符合规则，以字母数字开头结尾，只能包含".", "-"的大于1位不超过63位的字符串'
                    //     })
                    //     return
                    // }

                    if (key) {
                        labelInfo[key] = labelList[i].isMixin && value === '' ? '*****-----$$$$$' : value
                    }
                }

                const resNodeList = []

                if (this.curRowNode && Object.keys(this.curRowNode).length) {
                    resNodeList.push(this.curRowNode)
                } else {
                    if (this.checkedNodeList.length) {
                        resNodeList.push(...this.checkedNodeList)
                    }
                }

                try {
                    this.setLabelConf.loading = true
                    const params = {
                        $clusterId: this.curSelectedClusterId,
                        node_label_list: resNodeList.map(node => {
                            const nodeLabelInfo = {}
                            Object.keys(labelInfo).forEach(key => {
                                if (labelInfo[key] === '*****-----$$$$$' && node.labels[key]) {
                                    nodeLabelInfo[key] = node.labels[key]
                                } else if (labelInfo[key] !== '*****-----$$$$$') {
                                    nodeLabelInfo[key] = labelInfo[key]
                                }
                            })
                            return {
                                node_name: node.name,
                                labels: nodeLabelInfo
                            }
                        })
                    }
                    await this.$store.dispatch('cluster/setK8sNodeLabels', params)

                    this.hideSetLabel()
                    this.checkedNodeList.splice(0, this.checkedNodeList.length, ...[])
                    setTimeout(() => {
                        this.curRowNode = null
                        this.fetchData(true)
                    }, 200)
                } catch (e) {
                    console.error(e)
                } finally {
                    this.setLabelConf.loading = false
                }
            },

            /**
             * 设置标签 sideslder 取消按钮
             */
            hideSetLabel () {
                this.curRowNode = null
                this.setLabelConf.isShow = false
                this.labelList.splice(0, this.labelList.length, ...[])
            },

            /**
             * 进入节点详情页面
             *
             * @param {Object} node 节点信息
             */
            async goNodeOverview (node) {
                this.$router.push({
                    name: 'clusterNodeOverview',
                    params: {
                        projectId: node.project_id,
                        projectCode: this.$route.params.projectCode,
                        nodeId: node.inner_ip,
                        clusterId: node.cluster_id,
                        backTarget: 'nodeMain'
                    }
                })
            },

            /**
             * 跳转到 clusterOverview
             *
             * @param {Object} node 当前节点对象
             */
            async goClusterOverview (node) {
                this.$router.push({
                    name: 'clusterOverview',
                    params: {
                        projectId: node.project_id,
                        projectCode: node.project_code,
                        clusterId: node.cluster_id,
                        backTarget: 'nodeMain'
                    }
                })
            },

            /**
             * 跳转到 clusterNode
             *
             * @param {Object} node 当前节点对象
             */
            async goClusterNode (node) {
                this.$router.push({
                    name: 'clusterNode',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        clusterId: node.cluster_id,
                        backTarget: 'nodeMain'
                    },
                    query: {
                        inner_ip: node.inner_ip
                    }
                })
            },

            /**
             * 节点导出
             */
            async exportNode () {
                // const link = document.createElement('a')
                // link.style.display = 'none'
                // link.href = `${DEVOPS_BCS_API_URL}/api/projects/${this.projectId}/nodes/export/?cluster_id=`
                //     + `${this.curSelectedClusterId === 'all' ? '' : this.curSelectedClusterId}`
                // document.body.appendChild(link)
                // link.click()

                const url = `${DEVOPS_BCS_API_URL}/api/projects/${this.projectId}/nodes/export/`

                const response = await axios({
                    url: url,
                    method: 'post',
                    responseType: 'blob', // 这句话很重要
                    data: {
                        cluster_id: this.curSelectedClusterId === 'all' ? '' : this.curSelectedClusterId
                    }
                })

                if (response.status !== 200) {
                    console.log('系统异常，请稍候再试')
                    return
                }

                const blob = new Blob([response.data], { type: response.headers['content-type'] })
                const a = window.document.createElement('a')
                const downUrl = window.URL.createObjectURL(blob)
                let filename = 'download.xls'
                const contentDisposition = response.headers['content-disposition']
                if (contentDisposition && contentDisposition.indexOf('filename=') !== -1) {
                    filename = contentDisposition.split('filename=')[1]
                    a.href = downUrl
                    a.download = filename || 'download.xls'
                    a.click()
                    window.URL.revokeObjectURL(downUrl)
                }
            },

            renderSelectionHeader () {
                if (this.curPageData.filter(node => node.status === 'RUNNING').length) {
                    return <bk-checkbox
                        v-if={this.curPageData.length}
                        name="check-all-node"
                        disabled={this.isCheckAllDisabled()}
                        v-model={this.isCheckAllNode}
                        onChange={this.checkAllNode} />
                }
                return <bk-checkbox v-if={this.curPageData.length} name="check-instance" disabled={true} />
            },

            rowSelectable (row, index) {
                return row.permissions.edit
            },

            /**
             * 单选
             * @param {array} selection 已经选中的行数
             * @param {object} row 当前选中的行
             */
            handlePageSelect (selection, row) {
                this.checkedNodeList = selection
                console.log(this.checkedNodeList, 'this.checkedNodeList')
            },

            /**
             * 全选
             */
            handlePageSelectAll (selection, row) {
                this.checkedNodeList = selection
                console.log(selection)
            },

            /**
             * 设置污点
             */
            showTaintDialog (row) {
                this.taintDialog.isShow = true
                this.taintDialog.nodes = [row]
            },

            /**
             * 关闭设置污点
             */
            handleHideTaintDialog (isRefetch) {
                this.taintDialog.isShow = false
                this.taintDialog.nodes = []
                isRefetch && this.fetchData(true)
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
