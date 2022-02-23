<template>
    <div class="biz-content">
        <div class="biz-content-wrapper biz-cluster-info-wrapper">
            <div class="biz-cluster-info-inner">
                <div class="biz-cluster-tab-content" v-bkloading="{ isLoading: containerLoading, opacity: 1 }" style="min-height: 600px;">
                    <div class="biz-cluster-info-form-wrapper">
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
                                        {{clusterInfo.clusterName || '--'}}
                                        <a href="javascript:void(0);" class="bk-text-button ml10" @click="handleEditClusterName">
                                            <span class="bcs-icon bcs-icon-edit"></span>
                                        </a>
                                    </template>
                                    <template v-else>
                                        <div class="bk-form bk-name-form">
                                            <div class="bk-form-item">
                                                <div class="bk-form-inline-item">
                                                    <bcs-input style="width: 400px; margin-right: 15px;"
                                                        :maxlength="64"
                                                        :placeholder="$t('请输入集群名称，不超过64个字符')"
                                                        :value="clusterInfo.clusterName"
                                                        ref="clusterNameRef">
                                                    </bcs-input>
                                                </div>
                                                <div class="bk-form-inline-item">
                                                    <bk-button text @click="updateClusterName">{{$t('保存')}}</bk-button>
                                                    <bk-button text class="ml10" @click="isClusterNameEdit = false">{{$t('取消')}}</bk-button>
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
                                <div class="right">{{clusterInfo.engineType || '--'}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群ID')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.clusterID || '--'}}</div>
                            </div>
                            <div class="row" v-if="providerType === 'tke'">
                                <div class="left">
                                    <p>{{$t('TKE集群ID')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.systemID || '--'}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群版本')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.clusterBasicSettings ? clusterInfo.clusterBasicSettings.version : '--'}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('状态')}}</p>
                                </div>
                                <div class="right">{{clusterStatusMap[clusterInfo.status] || '--'}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('Master数量')}}</p>
                                </div>
                                <div class="right">
                                    <bk-button text :disabled="!masterNum" @click="handleShowMasterInfo">
                                        {{masterNum ? `${masterNum}${$t('个')}` : '--'}}
                                    </bk-button>
                                </div>
                            </div>
                            <!-- <div class="row">
                                <div class="left">
                                    <p>{{$t('节点数量')}}</p>
                                </div>
                                <div class="right">{{nodeCount}}</div>
                            </div> -->
                            <template v-if="$INTERNAL">
                                <!-- <div class="row">
                                    <div class="left">
                                        <p>{{$t('配置')}}</p>
                                    </div>
                                    <div class="right">{{configInfo}}</div>
                                </div> -->
                                <div class="row" v-if="providerType === 'tke'">
                                    <div class="left">
                                        <p>{{$t('网络类型')}}</p>
                                    </div>
                                    <div class="right">
                                        {{clusterInfo.networkType || '--'}}
                                    </div>
                                </div>
                                <div class="row">
                                    <div class="left">
                                        <p>{{$t('所属地域')}}</p>
                                    </div>
                                    <div class="right">
                                        {{clusterInfo.region || '--'}}
                                    </div>
                                </div>
                            </template>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('创建时间')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.createTime || '--'}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('更新时间')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.updateTime || '--'}}</div>
                            </div>
                            <div class="row">
                                <div class="left">
                                    <p>{{$t('集群描述')}}</p>
                                </div>
                                <div class="right">
                                    <template v-if="!isClusterDescEdit">
                                        {{clusterInfo.description || '--'}}
                                        <a href="javascript:void(0);" class="bk-text-button ml10" @click="handleEditClusterDesc">
                                            <span class="bcs-icon bcs-icon-edit"></span>
                                        </a>
                                    </template>
                                    <template v-else>
                                        <div class="bk-form bk-desc-form">
                                            <div class="bk-form-item">
                                                <div class="bk-form-inline-item">
                                                    <textarea maxlength="128"
                                                        :placeholder="$t('请输入集群描述，不超过128个字符')"
                                                        class="bk-form-textarea"
                                                        :value="clusterInfo.description"
                                                        ref="clusterDescRef">
                                                    </textarea>
                                                </div>
                                                <div class="bk-form-inline-item">
                                                    <bk-button text @click="updateClusterDesc">{{$t('保存')}}</bk-button>
                                                    <bk-button text class="ml10" @click="isClusterDescEdit = false">{{$t('取消')}}</bk-button>
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
                                    <bk-button text :disabled="!variableList.length" @click="handleSetVariable">
                                        {{variableList.length ? `${variableList.length}${$t('个')}` : '--'}}
                                    </bk-button>
                                </div>
                            </div>

                            <div class="row" v-if="providerType === 'tke'">
                                <div class="left">
                                    <p>VPC</p>
                                </div>
                                <div class="right">{{clusterInfo.vpcID || '--'}}</div>
                            </div>
                            <div class="row" v-if="providerType === 'tke'">
                                <div class="left">
                                    <p>{{$t('集群网络')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.networkSettings.clusterIPv4CIDR || '--'}}</div>
                            </div>
                            <!-- <div class="row" v-if="providerType === 'tke'">
                                <div class="left">
                                    <p>{{$t('Pod总量')}}</p>
                                </div>
                                <div class="right">{{}}</div>
                            </div> -->
                            <div class="row" v-if="providerType === 'tke'">
                                <div class="left">
                                    <p>{{$t('Service数量上限/集群')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.networkSettings.maxServiceNum || '--'}}</div>
                            </div>
                            <div class="row" v-if="providerType === 'tke'">
                                <div class="left">
                                    <p>{{$t('Pod数量上限/节点')}}</p>
                                </div>
                                <div class="right">{{clusterInfo.networkSettings.maxNodePodNum || '--'}}</div>
                            </div>
                            <div class="row" v-if="providerType === 'tke'">
                                <div class="left">
                                    <p>IPVS</p>
                                </div>
                                <div class="right">{{clusterInfo.clusterAdvanceSettings.IPVS}}</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <bk-dialog
            :is-show.sync="showMasterInfoDialog"
            :width="920"
            :has-header="false"
            :has-footer="false"
            :close-icon="true"
            ext-cls="biz-cluster-node-dialog"
            :quick-close="true"
            @confirm="showMasterInfoDialog = false"
            @cancel="showMasterInfoDialog = false">
            <template slot="content">
                <div style="margin: -20px;" v-bkloading="{ isLoading: masterInfoLoading, opacity: 1, zIndex: 1000 }">
                    <div class="biz-cluster-node-dialog-header">
                        <div class="left">
                            {{$t('Master信息')}}
                        </div>
                    </div>
                    <div style="min-height: 440px">
                        <bk-table :data="masterData">
                            <bk-table-column :label="$t('主机名称')" prop="host_name">
                                <template #default="{ row }">
                                    {{ row.host_name || '--' }}
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('内网IP')" prop="inner_ip"></bk-table-column>
                            <bk-table-column :label="$t('Agent状态')" prop="agent">
                                <template #default="{ row }">
                                    <StatusIcon :status="String(row.agent)" :status-color-map="statusColorMap">
                                        {{ taskStatusTextMap[String(row.agent)] }}
                                    </StatusIcon>
                                </template>
                            </bk-table-column>
                            <template v-if="$INTERNAL">
                                <bk-table-column :label="$t('机房')" prop="idc"></bk-table-column>
                                <bk-table-column :label="$t('机架')" prop="rack"></bk-table-column>
                                <bk-table-column :label="$t('机型')" prop="device_class"></bk-table-column>
                            </template>
                        </bk-table>
                    </div>
                </div>
            </template>
        </bk-dialog>

        <bk-sideslider
            :is-show.sync="showVariableInfo"
            :title="$t('设置变量')"
            :width="680"
            class="biz-cluster-set-variable-sideslider"
            :quick-close="false"
            @hidden="showVariableInfo = false">
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
                                    <div class="biz-key-value-item" v-for="variable in variableList" :key="variable.id">
                                        <bk-input style="width: 270px;" disabled :value="`${variable.name}(${variable.key})`"></bk-input>
                                        <span class="equals-sign">=</span>
                                        <bk-input class="right" style="width: 270px; margin-left: 35px;" :placeholder="$t('值')" v-model="variable.value"></bk-input>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="action-inner">
                            <bk-button type="primary" :loading="variableInfoLoading" @click="confirmSetVariable">
                                {{$t('保存')}}
                            </bk-button>
                            <bk-button type="button" class="ml10" :disabled="variableInfoLoading" @click="showVariableInfo = false">
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
    // import moment from 'moment'
    import StatusIcon from '@/views/dashboard/common/status-icon.tsx'
    export default {
        components: {
            StatusIcon
        },
        data () {
            return {
                masterInfoLoading: false,
                variableInfoLoading: false,
                containerLoading: false,
                showMasterInfoDialog: false,
                showVariableInfo: false,
                variableList: [],
                masterData: [],
                isClusterNameEdit: false,
                isClusterDescEdit: false,
                clusterInfo: {},
                providerType: 'k8s',
                clusterStatusMap: {
                    INITIALIZATION: this.$t('初始化中'),
                    RUNNING: this.$t('正常'),
                    DELETING: this.$t('删除中'),
                    'CREATE-FAILURE': this.$t('创建失败'),
                    'DELETE-FAILURE': this.$t('删除失败'),
                    DELETED: this.$t('已删除')
                },
                taskStatusTextMap: {
                    '1': this.$t('正常'),
                    '0': this.$t('异常')
                },
                statusColorMap: {
                    '1': 'green',
                    '0': 'red'
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
            clusterId () {
                return this.$route.params.clusterId
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            },
            curCluster () {
                return this.clusterList.find(item => item.cluster_id === this.clusterId) || {}
            },
            clusterPerm () {
                return this.$store.state.cluster.clusterPerm
            },
            masterNum () {
                return Object.keys(this.clusterInfo.master || {}).length
            },
            isSingleCluster () {
                const cluster = this.$store.state.cluster.curCluster
                return !!(cluster && Object.keys(cluster).length)
            }
        },
        async created () {
            this.fetchClusterInfo()
            this.fetchVariableInfo()
        },
        methods: {
            /**
             * 获取当前集群数据
             */
            async fetchClusterInfo () {
                this.containerLoading = true
                const res = await this.$store.dispatch('clustermanager/clusterDetail', {
                    $clusterId: this.clusterId
                }).catch(() => ({}))
                this.clusterInfo = res.data
                this.providerType = res.extra?.providerType
                this.containerLoading = false
            },

            /**
             * 获取变量信息
             */
            async fetchVariableInfo () {
                const res = await this.$store.dispatch('cluster/getClusterVariableInfo', {
                    projectId: this.projectId,
                    clusterId: this.clusterId
                }).catch(() => ({ data: [] }))
                this.variableList = res.data || []
            },

            handleEditClusterName () {
                this.isClusterNameEdit = true
                this.$nextTick(() => {
                    this.$refs.clusterNameRef.focus()
                })
            },

            handleEditClusterDesc () {
                this.isClusterDescEdit = true
                this.$nextTick(() => {
                    this.$refs.clusterDescRef.focus()
                })
            },

            async handleUpdateCluster (params) {
                this.containerLoading = true
                const result = await this.$store.dispatch('clustermanager/modifyCluster', {
                    $clusterId: this.clusterId,
                    ...params
                })
                if (result) {
                    await this.$store.dispatch('cluster/getClusterList', this.projectId)
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('修改成功')
                    })
                }
                this.containerLoading = false
                return result
            },
            // 更新当前集群信息
            handleUpdateCurCluster () {
                if (!this.isSingleCluster) return
                this.$store.commit('cluster/forceUpdateCurCluster', {
                    cluster_id: this.clusterInfo.clusterID,
                    name: this.clusterInfo.clusterName,
                    project_id: this.clusterInfo.projectID,
                    ...this.clusterInfo
                })
            },

            // 更新集群名称信息
            async updateClusterName () {
                const clusterName = this.$refs.clusterNameRef?.curValue
                if (!clusterName) return

                const result = await this.handleUpdateCluster({
                    clusterName
                })
                if (result) {
                    this.clusterInfo.clusterName = clusterName
                    this.isClusterNameEdit = false
                    this.handleUpdateCurCluster()
                }
            },
            // 更新集群描述信息
            async updateClusterDesc () {
                const description = this.$refs.clusterDescRef?.value
                if (!description) return

                const result = await this.handleUpdateCluster({
                    description
                })
                if (result) {
                    this.clusterInfo.description = description
                    this.isClusterDescEdit = false
                    this.handleUpdateCurCluster()
                }
            },

            // 设置集群变量
            handleSetVariable () {
                this.showVariableInfo = true
                this.fetchVariableInfo()
            },
            /**
             * 设置变量 sideslder 确认按钮
             */
            async confirmSetVariable () {
                this.variableInfoLoading = true
                const result = await this.$store.dispatch('cluster/updateClusterVariableInfo', {
                    projectId: this.projectId,
                    clusterId: this.curCluster.cluster_id,
                    cluster_vars: this.variableList
                }).catch(() => false)
                this.variableInfoLoading = false
                if (result) {
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('保存成功')
                    })
                    this.showVariableInfo = false
                }
            },

            /**
             * 显示 master 信息
             */
            async handleShowMasterInfo () {
                this.showMasterInfoDialog = true
                this.masterInfoLoading = true
                const res = await this.$store.dispatch('cluster/getClusterMasterInfo', {
                    projectId: this.projectId,
                    clusterId: this.curCluster.cluster_id
                }).catch(() => [])
                this.masterData = res.data
                this.masterInfoLoading = false
            },

            /**
             * 返回集群首页列表
             */
            goIndex () {
                this.$router.push({
                    name: 'clusterMain',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
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
                        clusterId: this.clusterId
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
                        clusterId: this.clusterId
                    }
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

    .biz-cluster-info-wrapper {
        padding: 0;
    }

</style>
