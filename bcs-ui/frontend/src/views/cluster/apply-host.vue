<template>
    <div class="apply-host-wrapper" v-if="$INTERNAL">
        <div class="apply-host-btn">
            <span v-if="!hasAuth"
                class="bk-default bk-button-normal bk-button is-disabled"
                v-bk-tooltips="authTips">
                {{$t('申请服务器')}}
            </span>
            <span v-else-if="applyHostButton.disabled"
                class="bk-default bk-button-normal bk-button is-disabled"
                v-bk-tooltips="applyHostButton.tips">
                {{$t('申请服务器')}}
            </span>
            <bk-button v-else
                :theme="theme"
                :disabled="applyHostButton.disabled"
                @click="handleOpenApplyHost">
                {{$t('申请服务器')}}
            </bk-button>
        </div>
        <bcs-dialog
            :position="{ top: 80 }"
            v-model="applyDialogShow"
            :close-icon="false"
            :width="900"
            :title="$t('申请服务器')"
            render-directive="if"
            header-position="left"
            ext-cls="apply-host-dialog">
            <bk-alert type="info" class="mb20" v-if="applyHostButton.tips">
                <template #title><div v-html="applyHostButton.tips"></div></template>
            </bk-alert>
            <bk-form ext-cls="apply-form"
                ref="applyForm"
                :label-width="100"
                :model="formdata"
                :rules="rules">
                <bk-form-item property="region" :label="$t('所属地域')" :required="true" :desc="defaultInfo.areaDesc">
                    <bk-selector :placeholder="$t('请选择地域')"
                        :selected.sync="formdata.region"
                        :list="areaList"
                        :searchable="true"
                        :setting-key="'areaName'"
                        :display-key="'showName'"
                        :search-key="'areaName'"
                        :is-loading="isAreaLoading"
                        :disabled="defaultInfo.disabled">
                    </bk-selector>
                </bk-form-item>
                <bk-form-item property="networkKey" :label="$t('网络类型')" :desc="defaultInfo.netWorkDesc" :required="true">
                    <div class="bk-button-group">
                        <bcs-button
                            :disabled="defaultInfo.networkKey && defaultInfo.networkKey !== 'overlay'"
                            :class="{ 'active': formdata.networkKey === 'overlay', 'network-btn': true, 'network-zIndex': defaultInfo.networkKey === 'overlay' }"
                            @click="formdata.networkKey = 'overlay'">overlay</bcs-button>
                        <bcs-button
                            :disabled="defaultInfo.networkKey && defaultInfo.networkKey !== 'underlay'"
                            :class="{ 'active': formdata.networkKey === 'underlay', 'network-btn': true, 'network-zIndex': defaultInfo.networkKey === 'underlay' }"
                            @click="formdata.networkKey = 'underlay'">underlay</bcs-button>
                    </div>
                </bk-form-item>
                <bk-form-item property="zone_id" :label="$t('园区')" :required="true">
                    <bk-selector :placeholder="$t('请选择园区')"
                        :selected.sync="formdata.zone_id"
                        :list="zoneList"
                        :searchable="true"
                        setting-key="value"
                        display-key="label"
                        search-key="label"
                    >
                    </bk-selector>
                </bk-form-item>
                <bk-form-item property="vpc_name" :label="$t('所属VPC')" :required="true" :desc="defaultInfo.vpcDesc">
                    <bk-selector :placeholder="$t('请选择VPC')"
                        :selected.sync="formdata.vpc_name"
                        :list="vpcList"
                        :searchable="true"
                        :setting-key="'vpcId'"
                        :display-key="'vpcName'"
                        :search-key="'vpcName'"
                        :disabled="defaultInfo.disabled">
                    </bk-selector>
                </bk-form-item>
                <bk-form-item ext-cls="has-append-item" :label="$t('数据盘')" property="disk_size">
                    <div class="disk-inner">
                        <bcs-select class="w200" v-model="formdata.disk_type" :clearable="false">
                            <bcs-option v-for="item in diskTypeList"
                                :key="item.value"
                                :id="item.value"
                                :name="item.label"
                            >
                            </bcs-option>
                        </bcs-select>
                        <bk-input v-model="formdata.disk_size" type="number" :min="50" :placeholder="$t('请输入50的倍数的数值')">
                            <div class="group-text" slot="append">GB</div>
                        </bk-input>
                    </div>
                </bk-form-item>
                <bk-form-item :label="$t('需求数量')">
                    <bk-number-input
                        :value.sync="formdata.replicas"
                        :min="1"
                        :max="50"
                        :ex-style="{ 'width': '325px' }"
                        :placeholder="$t('请输入')">
                    </bk-number-input>
                </bk-form-item>
                <bk-form-item class="custom-item" :label="$t('机型')">
                    <div class="form-item-inner">
                        <label class="inner-label">CPU</label>
                        <div class="inner-content">
                            <bk-selector :selected.sync="hostData.cpu"
                                :list="cpuList"
                                :searchable="true"
                                :setting-key="'id'"
                                :display-key="'name'"
                                :search-key="'name'">
                            </bk-selector>
                        </div>
                    </div>
                    <div :class="['form-item-inner', !isEn && 'ml40']" :style="{ width: isEn ? '310px' : '286px' }">
                        <label class="inner-label">{{$t('内存')}}</label>
                        <div class="inner-content">
                            <bk-selector :selected.sync="hostData.mem"
                                :list="memList"
                                :searchable="true"
                                :setting-key="'id'"
                                :display-key="'name'"
                                :search-key="'name'">
                            </bk-selector>
                        </div>
                    </div>
                    <bk-button theme="primary" @click.stop="hanldeReloadHosts">{{$t('查询')}}</bk-button>
                </bk-form-item>
                <bk-form-item ref="hostItem"
                    style="flex: 0 0 100%;"
                    label=""
                    v-bkloading="{ isLoading: isHostLoading }"
                    :required="true"
                    property="cvm_type"
                    class="host-item">
                    <bk-radio-group v-model="formdata.cvm_type">
                        <bk-table :data="hostTableData" :max-height="320" :height="320" style="overflow-y: hidden;">
                            <bk-table-column label="" width="40" :resizable="false">
                                <template slot-scope="{ row }">
                                    <bk-radio name="host" :value="row.specifications" @change="handleRadioChange(row)"></bk-radio>
                                </template>
                            </bk-table-column>
                            <bk-table-column :label="$t('机型')" prop="model" :show-overflow-tooltip="{ interactive: false }"></bk-table-column>
                            <bk-table-column :label="$t('规格')" prop="specifications" :show-overflow-tooltip="{ interactive: false }"></bk-table-column>
                            <bk-table-column label="CPU" prop="cpu" width="80"></bk-table-column>
                            <bk-table-column :label="$t('内存')" prop="mem" width="80"></bk-table-column>
                            <bk-table-column :label="$t('备注')" prop="description" :show-overflow-tooltip="{ interactive: false }"></bk-table-column>
                        </bk-table>
                    </bk-radio-group>
                    <span class="checked-host-tips" style="height: 30px;">
                        {{formdata.cvm_type ? this.$t('已选择') + '：' + getHostInfoString : ' '}}
                    </span>
                </bk-form-item>
            </bk-form>
            <template slot="footer">
                <i18n v-show="isShowFooterTips" class="tips" target path="数据盘大小/CPU核数 > 50，申请的服务器有运管人工审批环节；如需加急，请联系{name}协助推动">
                    <a href="wxwork://message/?username=dommyzhang" style="color: #3A84FF;" place="name">dommyzhang</a>
                </i18n>
                <bk-button theme="primary" :loading="isSubmitLoading" @click.stop="handleSubmitApply">{{$t('确定')}}</bk-button>
                <bk-button theme="default" :disabled="isSubmitLoading" @click.stop="handleApplyHostClose">{{$t('取消')}}</bk-button>
            </template>
        </bcs-dialog>
    </div>
</template>

<script>
    export default {
        props: {
            theme: {
                type: String,
                default: 'default'
            },
            isBackfill: {
                type: Boolean,
                default: false
            },
            clusterId: {
                type: String,
                default: ''
            }
        },
        data () {
            return {
                timer: null,
                applyHostButton: {
                    disabled: true,
                    tips: ''
                },
                isSubmitLoading: false,
                isAreaLoading: false,
                isHostLoading: false,
                applyDialogShow: false,
                isFirstLoadData: true,
                areaList: [],
                vpcList: [],
                zoneList: [],
                diskTypeList: [],
                formdata: {
                    region: '',
                    disk_size: 50,
                    replicas: 1,
                    cvm_type: '',
                    vpc_name: '',
                    zone_id: '',
                    disk_type: '',
                    networkKey: 'overlay'
                },
                rules: {
                    region: [{
                        required: true,
                        trigger: 'blur',
                        message: this.$t('请选择地域')
                    }],
                    disk_size: [{
                        validator: (value) => value >= 50 && value % 50 === 0,
                        trigger: 'blur',
                        message: this.$t('请输入50的倍数的数值')
                    }],
                    cvm_type: [{
                        required: true,
                        trigger: 'change',
                        message: this.$t('请选择机型')
                    }]
                },
                hostData: {
                    cpu: 0,
                    mem: 0
                },
                checkedHostInfo: {},
                hostTableData: [],
                defaultInfo: {
                    areaDesc: '',
                    vpcDesc: '',
                    netWorkDesc: '',
                    disabled: false
                },
                clusterInfo: {},
                hasAuth: false,
                maintainers: [],
                cpuList: [{
                    id: 0,
                    name: this.$t('全部')
                }, {
                    id: 4,
                    name: 4 + this.$t('核')
                }, {
                    id: 8,
                    name: 8 + this.$t('核')
                }, {
                    id: 12,
                    name: 12 + this.$t('核')
                }, {
                    id: 16,
                    name: 16 + this.$t('核')
                }, {
                    id: 20,
                    name: 20 + this.$t('核')
                }, {
                    id: 24,
                    name: 24 + this.$t('核')
                }, {
                    id: 28,
                    name: 28 + this.$t('核')
                }, {
                    id: 32,
                    name: 32 + this.$t('核')
                }, {
                    id: 84,
                    name: 84 + this.$t('核')
                }],
                memList: [{
                    id: 0,
                    name: this.$t('全部')
                }, {
                    id: 8,
                    name: '8GB'
                }, {
                    id: 16,
                    name: '16GB'
                }, {
                    id: 24,
                    name: '24GB'
                }, {
                    id: 32,
                    name: '32GB'
                }, {
                    id: 36,
                    name: '36GB'
                }, {
                    id: 48,
                    name: '48GB'
                }, {
                    id: 56,
                    name: '56GB'
                }, {
                    id: 60,
                    name: '60GB'
                }, {
                    id: 64,
                    name: '64GB'
                }, {
                    id: 80,
                    name: '80GB'
                }, {
                    id: 128,
                    name: '128GB'
                }, {
                    id: 160,
                    name: '160GB'
                }, {
                    id: 320,
                    name: '320GB'
                }]
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            getHostInfoString () {
                if (!this.formdata.cvm_type) return ''
                return `${this.checkedHostInfo.specifications} （${this.checkedHostInfo.model}，${this.checkedHostInfo.cpu + this.checkedHostInfo.mem}）`
            },
            isEn () {
                return this.$store.state.isEn
            },
            isShowFooterTips () {
                if (!this.formdata.cvm_type) return false
                if (!this.checkedHostInfo.cpu) return false
                return this.formdata.disk_size / parseInt(this.checkedHostInfo.cpu) > 50
            },
            authTips () {
                const users = this.maintainers.join(', ')
                return {
                    content: this.$t('您不是当前项目绑定业务的运维人员，如需申请机器，请联系业务运维人员 {maintainers}', { maintainers: users }),
                    width: 240
                }
            },
            userInfo () {
                return this.$store.state.user
            }
        },
        watch: {
            'formdata.networkKey' (val, old) {
                val && val !== old && this.changeNetwork()
            },
            'formdata.cvm_type' () {
                this.$refs.applyForm && this.$refs.applyForm.$refs.hostItem && this.$refs.applyForm.$refs.hostItem.clearError()
            },
            'formdata.region': {
                immediate: true,
                async handler (value, old) {
                    if (value !== old) {
                        this.formdata.vpc_name = ''
                        this.vpcList = []
                        await this.fetchVPC()
                        await this.fetchZone()
                    }
                }
            },
            clusterId: {
                immediate: true,
                async handler (value, old) {
                    if (value && value !== old) {
                        await this.fetchClusterInfo()
                    }
                }
            }
        },
        async created () {
            await this.getBizMaintainers()

            if (!this.hasAuth) return
            if (this.isBackfill) {
                this.defaultInfo = {
                    areaDesc: this.$t('和集群所属区域一致'),
                    vpcDesc: this.$t('和集群所属vpc一致'),
                    netWorkDesc: this.$t('和集群网络类型一致'),
                    disabled: true
                }
            }
            this.getApplyHostStatus()
            this.fetchDiskType()
        },
        beforeDestroy () {
            clearTimeout(this.timer) && (this.timer = null)
        },
        methods: {
            /**
             * 获取申请权限
             */
            async getBizMaintainers () {
                const res = await this.$store.dispatch('cluster/getBizMaintainers')
                this.maintainers = res.maintainers
                this.hasAuth = this.maintainers.includes(this.userInfo.username)
            },
            /**
             * 获取当前集群数据
             */
            async fetchClusterInfo () {
                if (!this.clusterId) return

                try {
                    const res = await this.$store.dispatch('clustermanager/clusterDetail', {
                        $clusterId: this.clusterId
                    })
                    this.clusterInfo = res.data || {}
                    if (this.clusterInfo.networkType && this.isBackfill) {
                        this.formdata.networkKey = this.clusterInfo.networkType
                        this.defaultInfo.networkKey = this.formdata.networkKey
                    }
                } catch (e) {
                    console.error(e)
                }
            },
            /**
             * 获取所属地域
             */
            async getAreas () {
                try {
                    const list = await this.$store.dispatch('clustermanager/fetchCloudRegion', {
                        $cloudId: 'tencentCloud'
                    })
                    this.areaList = list.map(item => ({
                        areaId: item.cloudID,
                        areaName: item.region,
                        showName: item.regionName
                    }))
                    if (this.clusterInfo.region && this.isBackfill) {
                        const area = this.areaList.find(item => item.areaName === this.clusterInfo.region)
                        if (area) {
                            this.formdata.region = area.areaName
                        }
                    } else if (this.areaList.length) {
                        // 默认选中第一个
                        this.formdata.region = this.areaList[0].areaName
                    }
                } catch (e) {
                    this.areaList = []
                    console.error(e)
                } finally {
                    this.isAreaLoading = false
                }
            },

            /**
             * 选择网络类型
             */
            async changeNetwork (index, data) {
                this.vpcList = []
                await this.fetchVPC()
            },

            /**
             * 获取园区列表
             */
            async fetchZone () {
                if (!this.formdata.region) return

                try {
                    const data = await this.$store.dispatch('cluster/getZoneList', {
                        projectId: this.projectId,
                        region: this.formdata.region
                    })
                    this.zoneList = data.data
                    if (this.clusterInfo.zone_id && this.isBackfill) {
                        const zone = this.zoneList.find(item => item.value === this.clusterInfo.zone_id)
                        if (zone) {
                            this.formdata.zone_id = zone.value
                        }
                    } else if (this.zoneList.length) {
                        this.formdata.zone_id = this.zoneList[0].value
                    }
                } catch (e) {
                    console.error(e)
                }
            },

            /**
             * 获取数据盘类型列表
             */
            async fetchDiskType () {
                try {
                    const data = await this.$store.dispatch('cluster/getDiskTypeList', {
                        projectId: this.projectId
                    })
                    this.diskTypeList = data.data
                    if (this.clusterInfo.disk_type && this.isBackfill) {
                        const diskType = this.diskTypeList.find(item => item.value === this.clusterInfo.disk_type)
                        if (diskType) {
                            this.formdata.disk_type = diskType.value
                        }
                    } else if (this.diskTypeList.length) {
                        this.formdata.disk_type = this.diskTypeList[1].value
                    }
                } catch (e) {
                    console.error(e)
                }
            },

            async fetchVPC () {
                if (!this.formdata.region) {
                    return
                }
                try {
                    const data = await this.$store.dispatch('clustermanager/fetchCloudVpc', {
                        cloudID: 'tencentCloud',
                        region: this.formdata.region,
                        networkType: this.formdata.networkKey
                    })
                    const vpcList = data.map(item => ({
                        vpcId: item.vpcID,
                        vpcName: item.vpcName
                    }))
                    this.vpcList.splice(0, this.vpcList.length, ...vpcList)
                    if (this.clusterInfo.vpcID && this.isBackfill) {
                        const vpc = this.vpcList.find(item => item.vpcId === this.clusterInfo.vpcID)
                        if (vpc) {
                            this.formdata.vpc_name = vpc.vpcId
                        } else {
                            // 回填不上则直接显示当前vpc id
                            this.vpcList.unshift({
                                vpcId: this.clusterInfo.vpcID,
                                vpcName: this.clusterInfo.vpcID
                            })
                            this.formdata.vpc_name = this.clusterInfo.vpcID
                        }
                    } else if (this.vpcList.length) {
                        // 默认选中第一个
                        this.formdata.vpc_name = this.vpcList[0].vpcId
                    }
                    this.isFirstLoadData && this.hanldeReloadHosts()
                } catch (e) {
                    console.error(e)
                }
            },
            /**
             * 获取主机列表
             */
            async getHosts () {
                if (this.isFirstLoadData) {
                    this.isFirstLoadData = false
                }
                this.$refs.applyForm && this.$refs.applyForm.$refs.hostItem && this.$refs.applyForm.$refs.hostItem.clearError()
                try {
                    const res = await this.$store.dispatch('cluster/getSCRHosts', {
                        projectId: this.projectId,
                        region: this.formdata.region,
                        vpc_name: this.formdata.vpc_name,
                        cpu_core_num: this.hostData.cpu,
                        mem_size: this.hostData.mem
                    })
                    const list = res.data || []
                    this.hostTableData = list.map(item => {
                        const getRow = item.describe.split('; ')
                        return {
                            model: getRow[0],
                            specifications: item.value,
                            cpu: getRow[1].replace('CPU:', ''),
                            mem: getRow[2].replace('MEM:', ''),
                            description: getRow[3].replace('备注:', '')
                        }
                    })
                } catch (e) {
                    this.hostTableData = []
                    console.error(e)
                } finally {
                    this.isHostLoading = false
                }
            },
            /**
             * 申请服务器
             */
            async handleSubmitApply () {
                const validate = await this.$refs.applyForm.validate()
                if (!validate) return
                try {
                    this.isSubmitLoading = true
                    await this.$store.dispatch('cluster/applySCRHost', Object.assign({}, this.formdata, {
                        projectId: this.projectId
                    }))
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('主机申请提交成功')
                    })
                    this.handleApplyHostClose()
                } catch (e) {
                    console.error(e)
                } finally {
                    this.isSubmitLoading = false
                }
            },
            hanldeReloadHosts () {
                this.isHostLoading = true
                this.formdata.cvm_type = ''
                this.getHosts()
            },
            handleRadioChange (row) {
                this.checkedHostInfo = row
            },

            /**
             * 打开申请服务器 dialog
             */
            handleOpenApplyHost () {
                // reset
                this.formdata = {
                    ...this.formdata,
                    region: '',
                    vpc_name: '',
                    disk_size: 50,
                    replicas: 1,
                    cvm_type: ''
                }

                this.hostData = {
                    cpu: 0,
                    mem: 0
                }

                this.applyDialogShow = true
                this.isAreaLoading = true
                this.getAreas()
            },

            /**
             * 关闭申请服务器 dialog
             */
            handleApplyHostClose () {
                this.applyDialogShow = false
                clearTimeout(this.timer) && (this.timer = null)
                this.getApplyHostStatus()
            },

            /**
             * 查看主机申请状态
             */
            async getApplyHostStatus () {
                try {
                    const res = await this.$store.dispatch('cluster/checkApplyHostStatus', { projectId: this.projectId })
                    const data = res.data || {}
                    const status = data.status
                    if (status && status !== 'FINISHED') {
                        const tipsContentMap = {
                            RUNNING: this.$t('主机申请中')
                        }
                        this.applyHostButton.tips = `${tipsContentMap[status] || this.$t('项目下存在主机申请失败的单据，请联系申请者【{name}】或', { name: data.operator })}
                        <a href="${data.scr_url}" target="_blank" style="color: #3a84ff;">${this.$t('查看详情')}</a>`
                    } else {
                        this.applyHostButton.tips = ''
                    }
                    this.applyHostButton.disabled = status === 'RUNNING'
                    if (this.timer) {
                        clearTimeout(this.timer)
                        this.timer = null
                    }
                    if (status === 'RUNNING') {
                        this.timer = setTimeout(() => {
                            this.getApplyHostStatus()
                        }, 15000)
                    }
                } catch (e) {
                    console.error(e)
                }
            }
        }
    }
</script>

<style lang="postcss" scoped>
.apply-host-wrapper {
}
.apply-host-dialog {
    .bk-dialog-footer {
        .tips {
            display: inline-block;
            vertical-align: middle;
            max-width: 640px;
            font-size: 12px;
            margin-right: 8px;
            text-align: left;
        }
    }
}
.apply-form {
    display: flex;
    flex-wrap: wrap;
    /deep/ .bk-form-item {
        flex: 0 0 50%;
        margin-top: 0;
        margin-bottom: 20px;
        &.custom-item {
            flex: 0 0 100%;
            .bk-form-content {
                display: flex;
            }
        }
        &.is-error {
            .bk-selector-wrapper > input {
                border-color: #ff5656 !important;
                color: #ff5656;
            }
            .bk-selector-list input {
                border-color: #dde4eb !important;
                color: #63656e !important;
            }
            &.host-item {
                .tooltips-icon {
                    left: 16px;
                    right: unset !important;
                    top: 14px;
                }
            }
        }
        &.has-append-item {
            .tooltips-icon {
                right: 74px !important;
            }
        }
        .bk-button-group {
            width: 100%;
            display: block;
            .network-btn {
                width: 50%;
            }
            .network-zIndex {
                z-index: 1;
            }
        }
        .disk-inner {
            display: flex;
            .w200 {
                width: 200px;
            }
        }
        .form-item-inner {
            width: 326px;
            display: flex;
            margin-right: 20px;
        }
        .inner-label {
            height: 32px;
            line-height: 32px;
            margin-right: 10px;
        }
        .inner-content {
            flex: 1;
        }
        .checked-host-tips {
            display: block;
            margin-top: 10px;
        }
    }
}
</style>
