<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-cluster-node-title">
                <i class="bcs-icon bcs-icon-arrows-left back" @click="goIndex" v-if="!globalClusterId"></i>
                <template v-if="exceptionCode && exceptionCode.code !== 4005"><span>{{$t('返回')}}</span></template>
                <template v-else>
                    <template v-if="curClusterInPage.cluster_id">
                        <span @click="refreshCurRouter">{{curClusterInPage.name}}</span>
                        <span style="font-size: 12px; color: #c3cdd7;cursor:default;margin-left: 10px;">
                            （{{curClusterInPage.cluster_id}}）
                        </span>
                    </template>
                </template>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper biz-cluster-info-wrapper">
            <app-exception
                v-if="exceptionCode && !containerLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>

            <div v-if="!exceptionCode" class="biz-cluster-info-inner">
                <div class="biz-cluster-tab-header">
                    <div class="header-item" @click="goOverview">
                        <i class="bcs-icon bcs-icon-bar-chart"></i>{{$t('总览')}}
                    </div>
                    <div class="header-item" @click="goNode">
                        <i class="bcs-icon bcs-icon-list"></i>{{$t('节点管理')}}
                    </div>
                    <div class="header-item active">
                        <i class="cc-icon icon-cc-machine"></i>{{$t('集群信息')}}
                    </div>
                </div>
                <div class="biz-cluster-tab-content" v-bkloading="{ isLoading: containerLoading, opacity: 1 }" style="min-height: 600px;">
                    <div class="biz-cluster-info-form-wrapper" v-if="!containerLoading">
                        <div class="label">
                            {{$t('基本信息')}}
                        </div>
                        <div class="content">
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群名称')}}</p>
                                </div>
                                <div class="right">
                                    <template v-if="!isClusterNameEdit">
                                        {{curClusterName}}
                                        <a href="javascript:void(0);" class="bk-text-button ml10" @click="editClusterName">
                                            <span class="bcs-icon bcs-icon-edit"></span>
                                        </a>
                                    </template>
                                    <template v-else>
                                        <div class="bk-form bk-name-form">
                                            <div class="bk-form-item">
                                                <div class="bk-form-inline-item">
                                                    <bk-input style="width: 400px; margin-right: 15px;"
                                                        :maxlength="64"
                                                        :placeholder="$t('请输入集群名称，不超过64个字符')"
                                                        v-model="clusterEditName">
                                                    </bk-input>
                                                </div>
                                                <div class="bk-form-inline-item">
                                                    <a href="javascript:void(0);" class="bk-text-button" @click="updateClusterName">
                                                        {{$t('保存')}}
                                                    </a>
                                                    <a href="javascript:void(0);" class="bk-text-button" @click="cancelEditClusterName">
                                                        {{$t('取消')}}
                                                    </a>
                                                </div>
                                            </div>
                                        </div>
                                    </template>
                                </div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('调度引擎')}}</p>
                                </div>
                                <div class="right">{{coes}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群ID')}}</p>
                                </div>
                                <div class="right">{{curClusterId}}</div>
                            </div>
                            <div class="row" v-if="curClusterInPage.type === 'tke'">
                                <div class="left">
                                    <p>{{$t('TKE集群ID')}}</p>
                                </div>
                                <div class="right">{{extraClusterId}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群版本')}}</p>
                                </div>
                                <div class="right">{{version}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('状态')}}</p>
                                </div>
                                <div class="right">{{statusName}}</div>
                            </div>
                            <!-- <div class="row">
                                <div class="left">
                                    <p>版本</p>
                                </div>
                                <div class="right">{{ver}}</div>
                            </div> -->
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('Master数量')}}</p>
                                </div>
                                <div class="right">
                                    <a href="javascript:void(0);" class="bk-text-button" @click="showMasterInfo">
                                        {{masterCount}}
                                    </a>
                                </div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('节点数量')}}</p>
                                </div>
                                <div class="right">{{nodeCount}}</div>
                            </div>
                            <template v-if="$INTERNAL">
                                <div class="row">
                                    <div class="left">
                                        <p>{{$t('配置')}}</p>
                                    </div>
                                    <div class="right">{{configInfo}}</div>
                                </div>
                                <div class="row" v-if="curClusterInPage.type === 'tke'">
                                    <div class="left">
                                        <p>{{$t('网络类型')}}</p>
                                    </div>
                                    <div class="right">
                                        {{networkType}}
                                    </div>
                                </div>
                                <div class="row">
                                    <div class="left">
                                        <p>{{$t('所属地域')}}</p>
                                    </div>
                                    <div class="right">
                                        {{areaName}}
                                    </div>
                                </div>
                            </template>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('创建时间')}}</p>
                                </div>
                                <div class="right">{{createdTime}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('更新时间')}}</p>
                                </div>
                                <div class="right">{{updatedTime}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群描述')}}</p>
                                </div>
                                <div class="right">
                                    <template v-if="!isClusterDescEdit">
                                        {{description}}
                                        <a href="javascript:void(0);" class="bk-text-button ml10" @click="editClusterDesc">
                                            <span class="bcs-icon bcs-icon-edit"></span>
                                        </a>
                                    </template>
                                    <template v-else>
                                        <div class="bk-form bk-desc-form">
                                            <div class="bk-form-item">
                                                <div class="bk-form-inline-item">
                                                    <textarea maxlength="128" :placeholder="$t('请输入集群描述，不超过128个字符')" class="bk-form-textarea" v-model="clusterEditDesc"></textarea>
                                                </div>
                                                <div class="bk-form-inline-item">
                                                    <a href="javascript:void(0);" class="bk-text-button" @click="updateClusterDesc">
                                                        {{$t('保存')}}
                                                    </a>
                                                    <a href="javascript:void(0);" class="bk-text-button" @click="cancelEditClusterDesc">
                                                        {{$t('取消')}}
                                                    </a>
                                                </div>
                                            </div>
                                        </div>
                                    </template>
                                </div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群变量')}}</p>
                                </div>
                                <div class="right">
                                    <a href="javascript:void(0);" class="bk-text-button" @click="showSetVariable" v-if="variableCount !== '--'">
                                        {{variableCount}}
                                    </a>
                                    <span v-else>{{variableCount}}</span>
                                </div>
                            </div>

                            <div class="row" v-if="curClusterInPage.type === 'tke'">
                                <div class="left">
                                    <p>VPC</p>
                                </div>
                                <div class="right">{{vpcId}}</div>
                            </div>
                            <div class="row" v-if="curClusterInPage.type === 'tke'">
                                <div class="left">
                                    <p>{{$t('集群网络')}}</p>
                                </div>
                                <div class="right">{{clusterCidr}}</div>
                            </div>
                            <div class="row" v-if="curClusterInPage.type === 'tke'">
                                <div class="left">
                                    <p>{{$t('Pod总量')}}</p>
                                </div>
                                <div class="right">{{maxPodNum}}</div>
                            </div>
                            <div class="row" v-if="curClusterInPage.type === 'tke'">
                                <div class="left">
                                    <p>{{$t('Service数量上限/集群')}}</p>
                                </div>
                                <div class="right">{{maxServiceNum}}</div>
                            </div>
                            <div class="row" v-if="curClusterInPage.type === 'tke'">
                                <div class="left">
                                    <p>{{$t('Pod数量上限/节点')}}</p>
                                </div>
                                <div class="right">{{maxNodePodNum}}</div>
                            </div>
                            <div class="row" v-if="curClusterInPage.type === 'tke'">
                                <div class="left">
                                    <p>kube-proxy</p>
                                </div>
                                <div class="right">{{kubeProxy}}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <bk-dialog
            :position="{ top: 120 }"
            :is-show.sync="dialogConf.isShow"
            :width="dialogConf.width"
            :content="dialogConf.content"
            :has-header="dialogConf.hasHeader"
            :has-footer="dialogConf.hasFooter"
            :close-icon="true"
            :ext-cls="'biz-cluster-node-dialog'"
            :quick-close="true"
            @confirm="dialogConf.isShow = false"
            @cancel="dialogConf.isShow = false">
            <template slot="content">
                <div style="margin: -20px;" v-bkloading="{ isLoading: dialogConf.loading, opacity: 1 }">
                    <div class="biz-cluster-node-dialog-header">
                        <div class="left">
                            {{dialogConf.title}}
                        </div>
                        <!-- <div class="bk-dialog-tool" @click="closeDialog">
                            <i class="bk-dialog-close bcs-icon bcs-icon-close"></i>
                        </div> -->
                    </div>
                    <div style="min-height: 441px;" :style="{ borderBottomWidth: curPageData.length ? '1px' : 0 }">
                        <!-- <bk-table
                            :data="versionList"
                            :size="'medium'">
                            <bk-table-column :label="$t('版本号')" :show-overflow-tooltip="false" min-width="200">
                                <template slot-scope="props">
                                    <p>
                                        <span>{{props.row.name}}</span>
                                        <span v-if="props.row.show_version_id === curShowVersionId">{{$t('(当前)')}}</span>
                                    </p>

                                    <bcs-popover
                                        v-if="props.row.comment"
                                        :delay="300"
                                        :content="props.row.comment"
                                        placement="right">
                                        <span style="color: #3a84ff; font-size: 12px;">{{$t('版本说明')}}</span>
                                    </bcs-popover>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('更新时间')" :show-overflow-tooltip="true" min-width="150">
                                <template slot-scope="props">
                                    {{props.row.updated}}
                                </template>
                            </bk-table-column>
                        </bk-table> -->
                        <table class="bk-table has-table-hover biz-table biz-cluster-node-dialog-table">
                            <thead>
                                <tr>
                                    <th style="width: 160px; padding-left: 30px;">{{$t('主机名称')}}</th>
                                    <th style="width: 220px;">{{$t('内网IP')}}</th>
                                    <th style="width: 120px;">{{$t('Agent状态')}}</th>
                                    <template v-if="$INTERNAL">
                                        <th style="width: 170px;">{{$t('机房')}}</th>
                                        <th style="width: 150px;">{{$t('机架')}}</th>
                                        <th style="width: 100px;">{{$t('机型')}}</th>
                                    </template>
                                </tr>
                            </thead>
                            <tbody>
                                <template v-if="curPageData.length">
                                    <tr v-for="(host, index) in curPageData" :key="index">
                                        <td style="padding-left: 30px;">
                                            <bcs-popover placement="top">
                                                <div class="name biz-text-wrapper">{{host.host_name || '--'}}</div>
                                                <template slot="content">
                                                    <p style="text-align: left; white-space: normal;word-break: break-all;">{{host.host_name || '--'}}</p>
                                                </template>
                                            </bcs-popover>
                                        </td>
                                        <td>
                                            <bcs-popover placement="top" class="vm">
                                                <div class="inner-ip">{{host.inner_ip || '--'}}</div>
                                                <template slot="content">
                                                    <p style="text-align: left; white-space: normal;word-break: break-all;">{{host.inner_ip || '--'}}</p>
                                                </template>
                                            </bcs-popover>
                                        </td>
                                        <td>
                                            <span class="biz-success-text vm" style="vertical-align: super;" v-if="String(host.agent) === '1'">
                                                {{$t('正常')}}
                                            </span>
                                            <template v-else-if="String(host.agent) === '0'">
                                                <bcs-popover placement="top">
                                                    <span class="biz-warning-text f12 vm" style="vertical-align: super;">
                                                        {{$t('异常')}}
                                                    </span>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">
                                                            <template>
                                                                {{$t('Agent异常，请先')}}<a :href="PROJECT_CONFIG.doc.installAgent" target="_blank" style="color:#3a84ff">{{$t('安装')}}</a>
                                                            </template>
                                                        </p>
                                                    </template>
                                                </bcs-popover>
                                            </template>
                                            <span class="biz-danger-text f12" style="vertical-align: super;" v-else>
                                                {{$t('错误')}}
                                            </span>
                                        </td>
                                        <template v-if="$INTERNAL">
                                            <td>
                                                <bcs-popover placement="top">
                                                    <div class="idcunit vm">{{host.idc || '--'}}</div>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{host.idc || '--'}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>
                                                <bcs-popover placement="top">
                                                    <div class="server-rack vm">{{host.server_rack || '--'}}</div>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{host.server_rack || '--'}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                            <td>
                                                <bcs-popover placement="top">
                                                    <div class="device-class vm">{{host.device_class || '--'}}</div>
                                                    <template slot="content">
                                                        <p style="text-align: left; white-space: normal;word-break: break-all;">{{host.device_class || '--'}}</p>
                                                    </template>
                                                </bcs-popover>
                                            </td>
                                        </template>
                                    </tr>
                                </template>
                                <template v-if="!curPageData.length && !dialogConf.loading">
                                    <tr>
                                        <td colspan="6">
                                            <div class="bk-message-box no-data">
                                                <bcs-exception type="empty" scene="part"></bcs-exception>
                                            </div>
                                        </td>
                                    </tr>
                                </template>
                            </tbody>
                        </table>
                    </div>
                    <div class="biz-page-box" v-if="pageConf.show && curPageData.length && (curPageData.length >= pageConf.pageSize || pageConf.curPage !== 1)">
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
        </bk-dialog>

        <bk-sideslider
            :is-show.sync="setVariableConf.isShow"
            :title="setVariableConf.title"
            :width="setVariableConf.width"
            @hidden="hideSetVariable"
            class="biz-cluster-set-variable-sideslider"
            :quick-close="false">
            <div slot="content">
                <div class="wrapper" style="position: relative;">
                    <form class="bk-form bk-form-vertical set-label-form">
                        <div class="bk-form-item flex-item">
                            <div class="left">
                                <label class="bk-label label">{{$t('变量：')}}</label>
                            </div>
                        </div>
                        <div class="bk-form-item">
                            <div class="bk-form-content">
                                <div class="biz-key-value-wrapper mb10">
                                    <div class="biz-key-value-item" v-for="(variable, index) in variableList" :key="index">
                                        <bk-input style="width: 270px;" :disabled="true" v-model="variable.leftContent" v-bk-tooltips.top="variable.leftContent"></bk-input>
                                        <span class="equals-sign">=</span>
                                        <bk-input class="right" style="width: 270px; margin-left: 35px;" :placeholder="$t('值')" v-model="variable.value"></bk-input>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="action-inner">
                            <bk-button type="primary" :loading="setVariableConf.loading" @click="confirmSetVariable">
                                {{$t('保存')}}
                            </bk-button>
                            <bk-button type="button" :disabled="setVariableConf.loading" @click="hideSetVariable">
                                {{$t('取消')}}
                            </bk-button>
                        </div>
                    </form>
                </div>
            </div>
        </bk-sideslider>
    </div>
</template>

<script>
    import moment from 'moment'

    import { catchErrorHandler, formatBytes } from '@/common/util'

    export default {
        data () {
            return {
                isClusterNameEdit: false,
                isClusterDescEdit: false,
                containerLoading: true,
                curClusterInPage: {},
                dialogConf: {
                    isShow: false,
                    width: 920,
                    hasHeader: false,
                    hasFooter: false,
                    closeIcon: false,
                    title: '',
                    loading: false
                },
                pageConf: {
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    count: 1,
                    show: true
                },
                curPageData: [],
                masterList: [],
                winHeight: 0,
                curClusterName: '',
                curClusterId: '',
                clusterEditName: '',
                clusterEditDesc: '',
                status: '',
                statusName: '',
                ver: '',
                masterCount: '',
                nodeCount: '',
                networkType: '',
                areaName: '',
                createdTime: '',
                updatedTime: '',
                description: '',
                configInfo: '',
                variableCount: '',
                variableList: [],
                setVariableConf: {
                    isShow: false,
                    title: this.$t('设置变量'),
                    width: 680,
                    loading: false
                },
                bkMessageInstance: null,
                exceptionCode: null,
                extraClusterId: '',
                version: '',
                maxPodNum: 0,
                maxServiceNum: 0,
                maxNodePodNum: 0,
                vpcId: '--',
                clusterCidr: '',
                kubeProxy: '--'
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            clusterId () {
                return this.$route.params.clusterId
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            },
            curCluster () {
                const data = this.clusterList.find(item => item.cluster_id === this.clusterId) || {}
                this.curClusterInPage = Object.assign({}, data)
                return JSON.parse(JSON.stringify(data))
            },
            curProject () {
                return this.$store.state.curProject
            },
            isEn () {
                return this.$store.state.isEn
            },
            globalClusterId () {
                return this.$store.state.curClusterId
            },
            clusterPerm () {
                return this.$store.state.cluster.clusterPerm
            }
        },
        destroyed () {
            this.bkMessageInstance && this.bkMessageInstance.close()
        },
        async created () {
            this.fetchClusterInfo()
            // if (!this.curCluster || Object.keys(this.curCluster).length <= 0) {
            //     if (this.projectId && this.clusterId) {
            //         this.fetchData()
            //     }
            // } else {
            //     this.fetchClusterInfo()
            // }
            if (!this.clusterPerm[this.curCluster?.clusterID]?.policy?.view) {
                await this.$store.dispatch('getResourcePermissions', {
                    project_id: this.projectId,
                    policy_code: 'view',
                    // eslint-disable-next-line camelcase
                    resource_code: this.curCluster?.cluster_id,
                    resource_name: this.curCluster?.name,
                    resource_type: `cluster_${this.curCluster?.environment === 'prod' ? 'prod' : 'test'}`
                }).catch(err => {
                    this.containerLoading = false
                    this.exceptionCode = {
                        code: err.code,
                        msg: err.message
                    }
                })
            }
        },
        mounted () {
            this.winHeight = window.innerHeight
        },
        methods: {
            editClusterName () {
                this.clusterEditName = this.curClusterName
                this.isClusterNameEdit = true
            },
            editClusterDesc () {
                this.clusterEditDesc = this.description
                this.isClusterDescEdit = true
            },
            cancelEditClusterName () {
                this.clusterEditName = ''
                this.isClusterNameEdit = false
            },
            cancelEditClusterDesc () {
                this.clusterEditDesc = this.curClusterDesc
                this.isClusterDescEdit = false
            },

            /**
             * 获取当前集群数据
             */
            async fetchData () {
                this.containerLoading = true
                try {
                    await this.$store.dispatch('cluster/getCluster', {
                        projectId: this.projectId,
                        clusterId: this.clusterId
                    })

                    this.fetchClusterInfo()
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取当前集群数据
             */
            async fetchClusterInfo () {
                this.containerLoading = true
                try {
                    const res = await this.$store.dispatch('cluster/getClusterInfo', {
                        projectId: this.projectId,
                        clusterId: this.curCluster.cluster_id // 这里用 this.curCluster 来获取是为了使计算属性生效
                    })

                    const data = res.data || {}
                    this.curClusterName = data.cluster_name || '--'
                    this.curClusterId = data.cluster_id || '--'
                    this.status = data.status || '--'
                    this.statusName = this.isEn ? (data.status || '--') : (data.chinese_status_name || '--')
                    this.ver = data.ver || '--'
                    let masterCount = data.master_count || 0
                    if (masterCount) {
                        masterCount += this.isEn ? '' : '个'
                    } else {
                        masterCount = '--'
                    }
                    this.masterCount = masterCount

                    this.nodeCount = data.node_count || '--'
                    this.networkType = data.network_type || '--'
                    this.areaName = data.area_name || '--'
                    this.createdTime = data.created_at ? moment(data.created_at).format('YYYY-MM-DD HH:mm:ss') : '--'
                    this.updatedTime = data.updated_at ? moment(data.updated_at).format('YYYY-MM-DD HH:mm:ss') : '--'
                    this.description = data.description || '--'
                    this.coes = data.type === 'k8s' ? 'BCS-K8S' : data.type.toUpperCase()

                    this.fetchClusterOverview()
                    this.extraClusterId = data.extra_cluster_id || '--'
                    this.version = data.version || '--'
                    this.maxPodNum = data.max_pod_num || '--'
                    this.clusterCidr = data.cluster_cidr || '--'
                    this.maxServiceNum = data.max_service_num || '--'
                    this.maxNodePodNum = data.max_node_pod_num || '--'
                    this.vpcId = data.vpc_id || '--'
                    this.kubeProxy = data.kube_proxy || '--'

                    this.fetchVariableInfo()
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取集群使用率
             */
            async fetchClusterOverview () {
                try {
                    const res = await this.$store.dispatch('cluster/clusterOverview', {
                        projectId: this.projectId,
                        clusterId: this.curCluster.cluster_id // 这里用 this.curCluster 来获取是为了使计算属性生效
                    })
                    const cpu = res.data.cpu_usage || {}
                    const mem = res.data.memory_usage || {}
                    const cpuTotal = cpu.total || 0
                    const memTotal = formatBytes(mem.total_bytes) || 0

                    this.configInfo = this.isEn
                        ? `${cpuTotal}core ${memTotal}`
                        : `${cpuTotal}核 ${memTotal}`
                } catch (e) {
                    catchErrorHandler(e, this)
                }
            },

            /**
             * 获取变量信息
             */
            async fetchVariableInfo () {
                try {
                    const res = await this.$store.dispatch('cluster/getClusterVariableInfo', {
                        projectId: this.projectId,
                        clusterId: this.curCluster.cluster_id // 这里用 this.curCluster 来获取是为了使计算属性生效
                    })

                    let variableCount = res.count || 0
                    if (variableCount) {
                        variableCount += this.isEn ? '' : '个'
                    } else {
                        variableCount = '--'
                    }
                    this.variableCount = variableCount
                    const variableList = []
                    ;(res.data || []).forEach(item => {
                        item.leftContent = `${item.name}(${item.key})`
                        variableList.push(item)
                    })

                    this.variableList.splice(0, this.variableList.length, ...variableList)
                } catch (e) {
                    console.error(e)
                } finally {
                    this.containerLoading = false
                    setTimeout(() => {
                        this.setVariableConf.loading = false
                    }, 300)
                }
            },

            /**
             * 显示集群变量 sideslider
             */
            async showSetVariable () {
                this.setVariableConf.isShow = true
                await this.fetchVariableInfo()
            },

            /**
             * 设置变量 sideslder 确认按钮
             */
            async confirmSetVariable () {
                const variableList = []

                const len = this.variableList.length
                for (let i = 0; i < len; i++) {
                    const variable = this.variableList[i]
                    variableList.push({
                        id: variable.id,
                        key: variable.key,
                        name: variable.name,
                        value: variable.value
                    })
                }

                try {
                    this.setVariableConf.loading = true
                    await this.$store.dispatch('cluster/updateClusterVariableInfo', {
                        projectId: this.projectId,
                        clusterId: this.curCluster.cluster_id, // 这里用 this.curCluster 来获取是为了使计算属性生效
                        cluster_vars: variableList
                    })

                    this.hideSetVariable()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'success',
                        message: this.$t('保存成功')
                    })
                } catch (e) {
                    console.error(e)
                } finally {
                    setTimeout(() => {
                        this.setVariableConf.loading = false
                    }, 300)
                }
            },

            /**
             * 设置标签 sideslder 取消按钮
             */
            hideSetVariable () {
                this.setVariableConf.isShow = false
                this.variableList.splice(0, this.variableList.length, ...[])
            },

            /**
             * 显示 master 信息
             */
            async showMasterInfo () {
                try {
                    this.pageConf.curPage = 1
                    this.dialogConf.isShow = true
                    this.dialogConf.title = this.$t('Master信息')
                    this.dialogConf.loading = true
                    this.curPageData.splice(0, this.curPageData.length, ...[])

                    const res = await this.$store.dispatch('cluster/getClusterMasterInfo', {
                        projectId: this.projectId,
                        clusterId: this.curCluster.cluster_id // 这里用 this.curCluster 来获取是为了使计算属性生效
                    })
                    const list = res.data || []
                    this.masterList.splice(0, this.masterList.length, ...list)
                    this.initPageConf()
                    this.curPageData = this.getDataByPage(this.pageConf.curPage)
                } catch (e) {
                    console.log(e)
                } finally {
                    this.dialogConf.loading = false
                }
            },

            /**
             * 初始化弹层翻页条
             */
            initPageConf () {
                const total = this.masterList.length
                this.pageConf.totalPage = Math.ceil(total / this.pageConf.pageSize) || 1
                this.pageConf.count = total
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            pageChange (page) {
                this.pageConf.curPage = page
                const data = this.getDataByPage(page)
                this.curPageData.splice(0, this.curPageData.length, ...data)
            },

            /**
             * 获取当前这一页的数据
             *
             * @param {number} page 当前页
             *
             * @return {Array} 当前页数据
             */
            getDataByPage (page) {
                let startIndex = (page - 1) * this.pageConf.pageSize
                let endIndex = page * this.pageConf.pageSize
                if (startIndex < 0) {
                    startIndex = 0
                }
                if (endIndex > this.masterList.length) {
                    endIndex = this.masterList.length
                }
                const data = this.masterList.slice(startIndex, endIndex)
                return data
            },

            /**
             * 关闭弹窗
             */
            closeDialog () {
                this.dialogConf.isShow = false
            },

            /**
             * 刷新当前 router
             */
            refreshCurRouter () {
                typeof this.$parent.refreshRouterView === 'function' && this.$parent.refreshRouterView()
            },

            /**
             * 返回集群首页列表
             */
            goIndex () {
                const { params } = this.$route
                if (params.backTarget) {
                    this.$router.push({
                        name: params.backTarget,
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                } else {
                    this.$router.push({
                        name: 'clusterMain',
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                }
            },

            /**
             * 切换到节点管理
             */
            goOverview () {
                this.$router.push({
                    name: 'clusterOverview',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        clusterId: this.clusterId,
                        backTarget: this.$route.params.backTarget
                    }
                })
            },

            /**
             * 切换到节点管理
             */
            goNode () {
                this.$router.push({
                    name: 'clusterNode',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        clusterId: this.clusterId,
                        backTarget: this.$route.params.backTarget
                    }
                })
            },

            updateClusterName () {
                const projectId = this.projectId
                const clusterId = this.clusterId
                const data = {
                    name: this.clusterEditName
                }
                if (!data.name) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入集群名称')
                    })
                    return false
                }
                this.$store.dispatch(
                    'cluster/updateCluster',
                    { projectId: projectId, clusterId: clusterId, data }
                ).then(res => {
                    this.isClusterNameEdit = false
                    this.clusterEditName = ''
                    this.curClusterName = data.name
                    this.curClusterInPage.name = data.name
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('修改成功！')
                    })
                }).catch(res => {
                    this.$bkMessage({
                        theme: 'error',
                        message: res.message || res.data.msg || res.statusText
                    })
                })
                // 更新集群信息
                const curClusterList = this.clusterList.map(item => {
                    if (item.cluster_id === clusterId) {
                        item.name = this.clusterEditName
                    }
                    return item
                })
                const storeCluster = this.$store.state.cluster.curCluster || {}
                const newStoreCluster = curClusterList.find(item => item.cluster_id === storeCluster.cluster_id)
                if (newStoreCluster) {
                    this.$store.commit('cluster/forceUpdateCurCluster', newStoreCluster)
                }
                this.$store.commit('cluster/forceUpdateClusterList', curClusterList)
            },

            updateClusterDesc () {
                const projectId = this.projectId
                const clusterId = this.clusterId
                const data = {
                    description: this.clusterEditDesc
                }
                if (!data.description) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请输入集群描述')
                    })
                    return false
                }
                this.$store.dispatch(
                    'cluster/updateCluster',
                    { projectId: projectId, clusterId: clusterId, data }
                ).then(res => {
                    this.isClusterDescEdit = false
                    this.clusterEditDesc = ''
                    this.description = data.description
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('修改成功！')
                    })
                }).catch(res => {
                    this.$bkMessage({
                        theme: 'error',
                        message: res.message || res.data.msg || res.statusText
                    })
                })
            }
        }
    }
</script>

<style scoped lang="postcss">
    @import './info.css';

    .bk-name-form {
        line-height: 36px;

        .bk-form-input {
            width: 400px;
            margin-right: 15px;
            font-size: 12px;
        }
    }

    .bk-desc-form {
        line-height: 70px;

        .bk-form-textarea {
            width: 400px;
            margin-right: 15px;
            font-size: 12px;
        }
    }

</style>
