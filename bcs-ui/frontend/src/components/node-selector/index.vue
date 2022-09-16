<template>
    <bk-dialog
        :is-show.sync="dialogConf.isShow"
        :width="dialogConf.width"
        :content="dialogConf.content"
        :has-header="dialogConf.hasHeader"
        :position="{ top: 60 }"
        :close-icon="dialogConf.closeIcon"
        :ext-cls="'biz-cluster-create-choose-dialog'"
        :quick-close="false"
        class="server-dialog"
        @confirm="chooseServer">
        <template slot="content">
            <div style="margin: -20px;" v-bkloading="{ isLoading: ccHostLoading }">
                <div class="biz-cluster-create-table-header">
                    <div class="left" style="height: 60px;">
                        {{$t('选择服务器')}}
                        <span class="remain-tip" v-if="remainCount">{{$t('已选择{remainCount}个节点', { remainCount: remainCount })}}</span>
                    </div>
                </div>
                <div style="min-height: 443px;">
                    <table class="bk-table has-table-hover biz-table biz-cluster-create-table" :style="{ borderBottomWidth: candidateHostList.length ? '1px' : 0 }">
                        <thead>
                            <tr>
                                <th style="width: 60px; text-align: right;">
                                    <label class="bk-form-checkbox mt5">
                                        <input type="checkbox" name="check-all-host" v-model="isCheckCurPageAll" @click="toggleCheckCurPage">
                                    </label>
                                </th>
                                <th style="width: 160px;">{{$t('主机名/IP')}}</th>
                                <th style="width: 220px;">{{$t('状态')}}</th>
                                <th style="width: 120px;">{{$t('容器数量')}}</th>
                                <th style="width: 200px;">{{$t('Pod数量')}}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <template v-if="curPageData.length">
                                <tr v-for="(host, index) in curPageData" @click.stop="rowClick" :key="index">
                                    <td style="width: 60px; text-align: right; ">
                                        <label class="bk-form-checkbox mt5">
                                            <input type="checkbox" name="check-host" v-model="host.isChecked" @click.stop="selectHost(candidateHostList)" :disabled="host.status !== 'RUNNING'">
                                        </label>
                                    </td>
                                    <td>
                                        {{host.inner_ip || '--'}}
                                    </td>
                                    <td>
                                        {{getHostStatus(host.status)}}
                                    </td>
                                    <td>
                                        {{ nodeMetric[host.inner_ip] ? nodeMetric[host.inner_ip].container_count : 0 }}
                                    </td>
                                    <td>
                                        {{ nodeMetric[host.inner_ip] ? nodeMetric[host.inner_ip].pod_count : 0 }}
                                    </td>
                                </tr>
                            </template>
                            <template v-if="!candidateHostList.length && !ccHostLoading">
                                <tr>
                                    <td colspan="7" style="top: 0;">
                                        <div class="bk-message-box no-data">
                                            <p class="message empty-message">{{$t('您在当前业务下没有主机资源，请联系业务运维')}}</p>
                                        </div>
                                    </td>
                                </tr>
                            </template>
                        </tbody>
                    </table>
                </div>
                <div class="biz-page-box" v-if="pageConf.show && candidateHostList.length">
                    <bk-pagination
                        :show-limit="false"
                        :current.sync="pageConf.curPage"
                        :count.sync="pageConf.count"
                        :limit="pageConf.pageSize"
                        @change="pageChange">
                    </bk-pagination>
                </div>
            </div>
        </template>
        <div slot="footer">
            <div class="bk-dialog-outer" style="overflow: hidden;">
                <div style="float: right;">
                    <bk-button type="primary" class="bk-dialog-btn bk-dialog-btn-confirm bk-btn-primary"
                        @click="chooseServer">
                        {{$t('确定')}}
                    </bk-button>
                    <bk-button type="button" @click="hiseChooseServer">
                        {{$t('取消')}}
                    </bk-button>
                </div>
            </div>
        </div>
    </bk-dialog>
</template>

<script>
    export default {
        props: {
            selected: {
                type: Array,
                default () {
                    return []
                }
            }
        },
        data () {
            return {
                curClusterId: '',
                dialogConf: {
                    isShow: false,
                    width: 920,
                    hasHeader: false,
                    closeIcon: false
                },
                ccHostLoading: false,
                clusterType: 'stag',
                // 弹层选择 master 节点，已经选择了多少个
                remainCount: 0,
                // 备选服务器集合
                candidateHostList: [],
                pageConf: {
                    count: 1,
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    show: true
                },

                TRUE: true,
                bkMessageInstance: null,
                // 已选服务器集合
                hostList: [],
                // 已选服务器集合的缓存，用于在弹框中选择，点击确定时才把 hostListCache 赋值给 hostList，同时清空 hostListCache
                // hostListCache: [],
                hostListCache: {},
                // 集群名称
                name: '',
                // nat
                // 当前页是否全选中
                isCheckCurPageAll: false,
                isChange: false,

                showStagTip: false,
                exceptionCode: null,
                curProject: [],
                nodeMetric: {}
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList || []
            },
            curPageData () {
                const { pageSize, curPage } = this.pageConf
                return this.candidateHostList.slice(pageSize * (curPage - 1), pageSize * curPage)
            }
        },
        watch: {
            curPageData: {
                handler (val) {
                    this.handleGetNodeOverview()
                }
            }
        },
        mounted () {
            this.curProject = Object.assign({}, this.onlineProjectList.filter(p => p.project_id === this.projectId)[0] || {})
        },
        methods: {
            async handleGetNodeOverview () {
                if (!this.curPageData.length) return
                const promiseList = []
                for (let i = 0; i < this.curPageData.length; i++) {
                    promiseList.push(
                        this.$store.dispatch('cluster/getNodeOverview', {
                            nodeIp: this.curPageData[i].inner_ip,
                            clusterId: this.curPageData[i].cluster_id,
                            projectId: this.projectId
                        }).then(data => {
                            this.$set(this.nodeMetric, this.curPageData[i].inner_ip, data.data)
                        })
                    )
                }
                await Promise.all(promiseList)
            },
            /**
             * 获取节点状态
             *
             * @param {string} status 节点状态
             */
            getHostStatus (status) {
                const statusMap = {
                    'INITIALIZATION': this.$t('初始化中'),
                    'RUNNING': this.$t('正常'),
                    'NOTREADY': this.$t('不正常'),
                    'REMOVABLE': this.$t('不可调度'),
                    'DELETING': this.$t('删除中'),
                    'ADD-FAILURE': this.$t('上架失败'),
                    'REMOVE-FAILURE': this.$t('下架失败'),
                    'UNKNOWN': this.$t('未知状态')
                }
                return statusMap[status] || this.$t('不正常')
            },

            /**
             * 选择服务器弹层搜索事件
             *
             * @param {Array} searchKeys 搜索字符数组
             */
            handleSearch (searchKeys) {
                this.fetchCCData({
                    offset: 0,
                    ipList: searchKeys
                })
            },

            /**
             * 弹层表格行选中
             *
             * @param {Object} e 事件对象
             */
            rowClick (e) {
                let target = e.target
                while (target.nodeName.toLowerCase() !== 'tr') {
                    target = target.parentNode
                }
                const checkboxNode = target.querySelector('input[type="checkbox"]')
                checkboxNode && checkboxNode.click()
            },

            /**
             * 选择服务器弹层确定按钮
             */
            chooseServer () {
                const list = Object.keys(this.hostListCache)
                const len = list.length
                if (!len) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择服务器')
                    })
                    return
                }

                const data = []
                list.forEach(key => {
                    data.push(this.hostListCache[key])
                })

                this.dialogConf.isShow = false
                this.hostList.splice(0, this.hostList.length, ...data)
                this.isCheckCurPageAll = false
                this.$emit('selected', data)
            },

            hiseChooseServer () {
                this.dialogConf.isShow = false
            },

            /**
             * 获取 cc 表格数据
             *
             * @param {Object} params ajax 查询参数
             */
            async fetchCCData (params = {}) {
                this.ccHostLoading = true
                try {
                    const res = await this.$store.dispatch('cluster/getK8sNodes', {
                        $clusterId: this.curClusterId
                    })

                    const count = res.length

                    this.pageConf.show = !!count
                    this.pageConf.count = count
                    this.pageConf.totalPage = Math.ceil(count / this.pageConf.pageSize)
                    if (this.pageConf.totalPage < this.pageConf.curPage) {
                        this.pageConf.curPage = 1
                    }

                    const list = res || []
                    list.forEach(item => {
                        if (this.hostListCache[item.inner_ip]) {
                            item.isChecked = true
                        }
                    })
                    this.candidateHostList.splice(0, this.candidateHostList.length, ...list)
                    this.initSelected(this.candidateHostList)
                    this.selectHost(this.candidateHostList)
                } catch (e) {
                    console.log(e)
                } finally {
                    this.ccHostLoading = false
                }
            },

            /**
             * 打开选择服务器弹层
             */
            async openDialog (clusterId) {
                this.curClusterId = clusterId
                this.remainCount = 0
                this.pageConf.curPage = 1
                this.dialogConf.isShow = true
                this.isCheckCurPageAll = false
                await this.fetchCCData()
            },

            initSelected (list) {
                list.forEach(item => {
                    item.isChecked = false
                    this.selected.forEach(selectItem => {
                        if (String(selectItem.inner_ip) === String(item.inner_ip)) {
                            item.isChecked = true
                        }
                    })
                })
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            pageChange (page) {
                this.pageConf.curPage = page
            },

            /**
             * 弹层表格全选
             */
            toggleCheckCurPage () {
                const isChecked = !this.isCheckCurPageAll
                this.candidateHostList.forEach(host => {
                    if (host.status === 'RUNNING') {
                        host.isChecked = isChecked
                    }
                })
                this.selectHost()
            },

            /**
             * 在选择服务器弹层中选择
             */
            selectHost (hosts = this.candidateHostList) {
                if (!hosts.length) {
                    return
                }

                setTimeout(() => {
                    const selectedHosts = hosts.filter(host => host.isChecked)

                    const canSelectedHosts = hosts.filter(host =>
                        host.status === 'RUNNING'
                    )

                    this.isCheckCurPageAll = selectedHosts.length === canSelectedHosts.length

                    // 清除 hostListCache
                    hosts.forEach(item => {
                        delete this.hostListCache[item.inner_ip]
                    })

                    // 重新根据选择的 host 设置到 hostListCache 中
                    selectedHosts.forEach(item => {
                        this.hostListCache[item.inner_ip] = item
                    })

                    this.remainCount = Object.keys(this.hostListCache).length
                }, 0)
            },

            /**
             * 已选服务器移除处理
             *
             * @param {Object} host 当前行的服务器
             * @param {number} index 当前行的服务器的索引
             */
            removeHost (host, index) {
                this.hostList.splice(index, 1)
                delete this.hostListCache[`${host.inner_ip}`]
            }
        }
    }
</script>

<style scoped lang="postcss">
    @import '../../css/variable.css';
    @import '../../css/mixins/clearfix.css';
    @import '../../css/mixins/ellipsis';
    .biz-cluster-create-table-header {
        @mixin clearfix;
        background-color: #fff;
        height: 60px;
        line-height: 59px;
        font-size: 16px;
        padding: 0 20px;
        border-bottom: none;
        border-top-left-radius: 2px;
        border-top-right-radius: 2px;
        .left {
            float: left;
            .tip {
                font-size: 12px;
                margin-left: 10px;
                color: #c3cdd7;
            }
            .remain-tip {
                font-size: 12px;
                margin-left: 10px;
                color: $dangerColor;
            }
        }
        .right {
            float: right;
        }

        .page-wrapper {
            height: 22px;
            display: inline-block;
            position: relative;
            top: -2px;
            line-height: 22px;
            ul {
                margin: 0;
                padding: 0;
                display: inline-block;
                overflow: hidden;
                height: 22px;
            }
            .page-item {
                min-width: 22px;
                height: 22px;
                line-height: 20px;
                text-align: center;
                display: inline-block;
                vertical-align: middle;
                font-size: 14px;
                float: left;
                margin-right: 0;
                border: 1px solid #c4c6cc;
                box-sizing: border-box;
                border-radius: 2px;
                overflow: hidden;
                i {
                    font-size: 12px;
                }
                &:first-child {
                    border-top-right-radius: 0;
                    border-bottom-right-radius: 0;
                }
                &:last-child {
                    border-top-left-radius: 0;
                    border-bottom-left-radius: 0;
                }
                &:hover {
                    border-color: $iconPrimaryColor;
                }
                &.disabled {
                    border-color: #c4c6cc !important;
                    .page-button {
                        cursor: not-allowed;
                        background-color: #fafafa;
                        &:hover {
                            color: #737987;
                        }
                    }
                }
                .page-button {
                    display: block;
                    color: #737987;
                    background-color: #fff;
                    &:hover {
                        color: $iconPrimaryColor;
                    }
                }
            }
        }
    }

    .biz-cluster-create-table {
        background-color: #fff;
        border: 1px solid #dde4eb;
        width: 800px;
        thead {
            background-color: #fafbfd;
            tr {
                th {
                    height: 40px;
                }
            }
        }
        tbody {
            tr {
                &:hover {
                    background-color: #fafbfd;
                }
                td {
                    height: 40px;
                    font-size: 12px;
                }
            }
        }
        .no-data {
            min-height: 399px;
            .empty-message {
                margin-top: 160px;
            }
        }
    }

    .biz-cluster-create-choose-dialog {
        .biz-cluster-create-table {
            border-left: none;
            border-right: none;
            border-bottom: none;
            width: 910px;
            thead {
                tr {
                    th {
                        padding-top: 0;
                        padding-bottom: 0;
                    }
                }
            }
            tbody {
                tr {
                    td {
                        padding-top: 0;
                        padding-bottom: 0;
                        position: relative;
                    }
                }
            }
            .name {
                @mixin ellipsis 120px
            }
            .inner-ip {
                @mixin ellipsis 200px
            }
            .idcunit {
                @mixin ellipsis 200px
            }
            .server-rack {
                @mixin ellipsis 130px
            }
            .device-class {
                @mixin ellipsis 80px
            }
        }

        .biz-cluster-create-table-header {
            border-left: none;
            border-right: none;
        }
        .biz-search-input {
            width: 320px;
        }
        .biz-page-box {
            padding: 10px 25px 10px 0;
            background-color: #fafbfd;
            border-top: 1px solid #dde4eb;
            margin-top: -1px;
        }

        .bk-dialog-footer.bk-d-footer {
            background-color: #fff;
        }
    }

    .server-tip {
        float: left;
        line-height: 17px;
        font-size: 12px;
        text-align: left;
        padding: 13px 0 0 20px;
        margin-left: 20px;

        li {
            list-style: circle;
        }
    }

    .biz-page-box {
        @mixin clearfix;
        padding: 30px 40px 35px 0;
        .bk-page {
            float: right;
        }
    }

    .bk-dialog-footer {
        overflow: hidden;
    }
</style>
