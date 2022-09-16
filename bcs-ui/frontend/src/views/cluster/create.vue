<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-cluster-create-title" @click="goIndex">
                <i class="bcs-icon bcs-icon-arrows-left back"></i>
                <span>{{$t('创建容器集群')}}</span>
            </div>
            <bk-guide></bk-guide>
        </div>
        <div class="biz-content-wrapper">
            <app-exception
                v-if="exceptionCode"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <div v-else class="biz-cluster-create-wrapper">
                <div class="biz-cluster-create-form-wrapper">
                    <div class="form-item" :class="isEn ? 'en' : ''" v-if="isK8sProject || isTkeProject" @mouseenter="tipsActive = 'engine'" @mouseleave="tipsActive = ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">{{$t('调度引擎')}}</label>
                        <div class="form-item-inner bk-button-group" style="line-height: 30px;">
                            <bk-button class="bk-button bk-default is-outline"
                                :class="coes === 'tke' ? 'active' : ''"
                                @click="coes = 'tke'">TKE</bk-button>
                        </div>
                    </div>

                    <div class="form-item" :class="isEn ? 'en' : ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">{{$t('集群类型')}}</label>
                        <div class="form-item-inner" style="line-height: 30px;">
                            <bk-radio-group v-model="clusterType" @change="toggleDev">
                                <bk-radio class="mr30" value="stag">{{$t('测试环境')}}</bk-radio>
                                <bk-radio class="cluster-prod" value="prod">{{$t('正式环境')}}</bk-radio>
                            </bk-radio-group>
                        </div>
                    </div>

                    <div class="form-item" :class="isEn ? 'en' : ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">{{$t('NAT检查')}}</label>
                        <div class="form-item-inner">
                            <bk-radio-group v-model="needNat">
                                <bk-radio class="mr30" :value="true" :disabled="isK8sProject">{{$t('是')}}</bk-radio>
                                <bk-radio :value="false" :disabled="isK8sProject">{{$t('否')}}</bk-radio>
                            </bk-radio-group>
                        </div>
                    </div>

                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''" v-if="isK8sProject || isTkeProject" @mouseenter="tipsActive = 'version'" @mouseleave="tipsActive = ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">
                            {{$t('版本')}}
                        </label>
                        <div class="form-item-inner dropdown">
                            <bk-selector
                                :selected.sync="versionKey"
                                :list="versionList"
                                :setting-key="'id'"
                                :display-key="'name'">
                            </bk-selector>
                        </div>
                    </div>

                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">{{$t('名称')}}</label>
                        <div class="form-item-inner">
                            <input maxlength="60" type="text" class="bk-form-input cluster-name" :placeholder="$t('请输入集群名称')"
                                :class="validate.name.illegal ? 'is-danger' : ''" v-model="name"
                            >
                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.name.illegal">
                                <p class="tip-text">{{$t('必填项，不超过60个字符')}}</p>
                            </div>
                        </div>
                    </div>
                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">{{$t('集群描述')}}</label>
                        <div class="form-item-inner">
                            <textarea maxlength="120"
                                v-model="description"
                                style="width: 320px;"
                                class="bk-form-textarea" :class="validate.description.illegal ? 'is-danger' : ''"
                                :placeholder="$t('请输入集群描述')">
                            </textarea>
                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.description.illegal">
                                <p class="tip-text">{{$t('必填项，不超过120个字符')}}</p>
                            </div>
                        </div>
                    </div>

                    <template v-if="isTkeProject">

                        <div class="form-item bk-form-item" :class="isEn ? 'en' : ''" @mouseenter="tipsActive = 'networdType'" @mouseleave="tipsActive = ''">
                            <label class="long">{{$t('网络类型')}}</label>
                            <div class="form-item-inner bk-button-group">
                                <bk-button class="bk-button bk-default is-outline"
                                    :class="networkKey === 'overlay' ? 'active' : ''"
                                    @click="networkKey = 'overlay'">overlay</bk-button>
                                <bk-button class="bk-button bk-default is-outline"
                                    :class="networkKey === 'underlay' ? 'active' : ''"
                                    @click="networkKey = 'underlay'">underlay</bk-button>
                            </div>
                        </div>
                    </template>

                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''" @mouseenter="tipsActive = 'area'" @mouseleave="tipsActive = ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">{{$t('所属地域')}}</label>
                        <div class="form-item-inner dropdown">
                            <bk-selector :placeholder="$t('请选择地域')"
                                :selected.sync="areaIndex"
                                :list="areaList"
                                :ext-cls="validate.area.illegal ? 'is-danger' : ''"
                                :searchable="true"
                                :setting-key="'areaId'"
                                :display-key="'showName'"
                                :search-key="'areaName'"
                                @item-selected="changeArea">
                            </bk-selector>
                        </div>
                    </div>

                    <template v-if="isTkeProject">

                        <div class="form-item bk-form-item" :class="isEn ? 'en' : ''" @mouseenter="tipsActive = 'VPC'" @mouseleave="tipsActive = ''">
                            <label class="long">{{$t('所属VPC')}}</label>
                            <div class="form-item-inner dropdown">
                                <bk-selector :placeholder="$t('请选择VPC')"
                                    :selected.sync="vpcIndex"
                                    :list="vpcList"
                                    :ext-cls="validate.vpc.illegal ? 'is-danger' : ''"
                                    :searchable="true"
                                    :setting-key="'vpcId'"
                                    :display-key="'vpcName'"
                                    :search-key="'vpcName'"
                                    @item-selected="changeVPC">
                                </bk-selector>
                            </div>
                        </div>

                        <div class="form-item bk-form-item" :class="isEn ? 'en' : ''" @mouseenter="tipsActive = 'netword'" @mouseleave="tipsActive = ''">
                            <label class="long">{{$t('容器网络')}}</label>
                            <div class="form-item-inner">
                                <div class="netword-area">
                                    <div class="netword-area-item" :class="isEn ? 'en' : ''">
                                        <label class="long">{{$t('IP数量')}}</label>
                                        <div class="netword-area-item-inner">
                                            <bk-selector v-if="clusterType === 'stag'"
                                                :selected.sync="ipNumberKey"
                                                :list="ipNumberTestList"
                                                :setting-key="'id'"
                                                :display-key="'name'"
                                                :ext-cls="validate.ipNumber.illegal ? 'is-danger' : ''">
                                            </bk-selector>
                                            <bk-selector v-if="clusterType === 'prod'"
                                                :selected.sync="ipNumberKey"
                                                :list="ipNumberProdList"
                                                :setting-key="'id'"
                                                :display-key="'name'"
                                                :ext-cls="validate.ipNumber.illegal ? 'is-danger' : ''">
                                            </bk-selector>
                                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.ipNumber.illegal">
                                                <p class="tip-text">{{validate.ipNumber.msg}}</p>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="netword-area-item" :class="isEn ? 'en' : ''">
                                        <label class="long">{{$t('Service数量上限/集群')}}</label>
                                        <div class="netword-area-item-inner">
                                            <bk-selector
                                                :selected.sync="serviceIpNumberKey"
                                                :list="serviceIpNumberList"
                                                :setting-key="'id'"
                                                :display-key="'name'"
                                                :ext-cls="validate.serviceIpNumber.illegal ? 'is-danger' : ''">
                                            </bk-selector>
                                        </div>
                                    </div>
                                    <div class="netword-area-item" :class="isEn ? 'en' : ''">
                                        <label class="long">{{$t('Pod数量上限/节点')}}</label>
                                        <div class="netword-area-item-inner">
                                            <bk-selector
                                                :selected.sync="podNumberPerNodeKey"
                                                :list="podNumberPerNodeList"
                                                :placeholder="$t('请选择Pod数量上限/节点')"
                                                :setting-key="'id'"
                                                :display-key="'name'"
                                                :ext-cls="validate.podNumberPerNode.illegal ? 'is-danger' : ''">
                                            </bk-selector>
                                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.podNumberPerNode.illegal">
                                                <p class="tip-text">{{validate.podNumberPerNode.msg}}</p>
                                            </div>
                                        </div>
                                    </div>
                                    <p class="computed-rules">{{$t('计算规则: (IP数量-Service的数量)/(Master数量+Node数量)')}}</p>
                                    <i18n v-if="isShowTips" path="当前容器网络配置下，集群最多 {count} 个节点(包含Master和Node)" class="computed-rules" tag="p">
                                        <strong place="count" style="color: #222222;">{{ maxClusterCount }}</strong>
                                    </i18n>
                                </div>
                            </div>
                        </div>
                    </template>

                    <div class="form-item" :class="isEn ? 'en' : ''" @mouseenter="tipsActive = 'master'" @mouseleave="tipsActive = ''">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''">{{$t('选择Master')}}</label>
                        <div class="form-item-inner">
                            <p style="font-size: 12px; padding-bottom: 6px;" v-if="disabledSelectMaster">{{$t('请先选择所属地域和所属VPC')}}</p>
                            <div class="host-opertation">
                                <bk-button
                                    class="server-select-button"
                                    type="default"
                                    :disabled="disabledSelectMaster"
                                    :class="[validate.host.illegal ? 'is-danger' : '', 'server-select-button', { disabled: disabledSelectMaster }]"
                                    @click="openDialog">
                                    <i class="bcs-icon bcs-icon-plus"></i>
                                    {{$t('选择服务器')}}
                                </bk-button>
                                <apply-host v-if="isTkeProject && $INTERNAL" class="ml10 apply-host" />
                            </div>
                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.host.illegal">
                                <p class="tip-text" style="width: 540px;">{{$t('请选择服务器')}}</p>
                            </div>
                        </div>
                    </div>
                    <div class="form-item" :class="isEn ? 'en' : ''" v-if="hostList.length">
                        <label :class="isK8sProject || isTkeProject ? 'long' : ''"></label>
                        <div class="form-item-inner">
                            <div class="biz-cluster-create-table-header" style="width: 540px;">
                                <div class="left">
                                    {{$t('已选服务器')}}
                                </div>
                            </div>
                            <table class="bk-table has-table-hover biz-table biz-cluster-create-table" style="width: 540px;">
                                <thead>
                                    <tr>
                                        <!-- <th style="text-align: left;padding-left: 30px;">
                                            {{$t('序号')}}
                                        </th> -->
                                        <th style="padding-left: 20px;">{{$t('IP地址')}}</th>
                                        <th>{{$t('机房')}}</th>
                                        <th>{{$t('机型')}}</th>
                                        <!-- <th>{{$t('机架')}}</th> -->
                                        <th style="padding-right: 20px;">{{$t('操作')}}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(host, index) in hostList" :key="index">
                                        <!-- <td style="text-align: left;padding-left: 30px;">
                                            {{index + 1}}
                                        </td> -->
                                        <td style="padding-left: 20px;">
                                            <div class="inner-ip">{{host.bk_host_innerip || '--'}}</div>
                                        </td>
                                        <td>{{host.idc_unit_name}}</td>
                                        <td>{{host.svr_device_class}}</td>
                                        <!-- <td>{{host.server_rack}}</td> -->
                                        <td style="padding-right: 20px;"><a href="javascript:void(0)" class="bk-text-button" @click="removeHost(host, index)">{{$t('移除')}}</a></td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''" v-if="!isTkeProject">
                        <label>{{$t('注意事项')}}</label>
                        <div class="form-item-inner" style="vertical-align: top;">
                            <div v-if="isK8sProject">
                                <bk-checkbox name="cluster-classify-checkbox" v-model="checkHostname">
                                    {{$t('服务器将按照系统规则修改主机名')}}
                                    <i class="bcs-icon bcs-icon-question-circle"
                                        style="vertical-align: middle; cursor: pointer;"
                                        v-bk-tooltips="{
                                            content: `<p>cluster id: BCS-K8S-40000, master ip: 127.0.0.1</p>
                                                    <p>${$t('修改后')}: ip-127-0-0-1-m-bcs-k8s-40000</p>`,
                                            placement: 'right'
                                        }"></i>
                                </bk-checkbox>
                            </div>
                            <div>
                                <bk-checkbox name="cluster-classify-checkbox" v-model="checkService">
                                    {{$t('服务器将安装容器服务相关组件')}}
                                </bk-checkbox>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="biz-cluster-create-form-tips">
                    <div :class="['tips-item', { 'active': tipsActive === 'engine' }]">
                        <h6 class="title">{{$t('调度引擎')}}</h6>
                        <div class="item-content">
                            <ul class="tips-list">
                                <li>TKE：{{$t('K8S容器编排引擎，集群为腾讯云构建，经BCS纳管，默认可以使用K8S Oteam版本或者腾讯云1.16版本。推荐自研上云使用。')}}</li>
                            </ul>
                        </div>
                    </div>
                    <div :class="['tips-item', { 'active': tipsActive === 'version' }]">
                        <h6 class="title">{{$t('版本')}}</h6>
                        <div class="item-content">
                            <template v-if="isTkeProject">{{$t('使用K8S Oteam版本或者腾讯云1.16版本。')}}</template>
                            <template v-else>{{$t('使用原生k8s版本，也可使用公司开源协同的k8s版本。')}}</template>
                        </div>
                    </div>
                    <template v-if="isTkeProject">
                        <div :class="['tips-item', { 'active': tipsActive === 'networdType' }]">
                            <h6 class="title">{{$t('网络类型')}}</h6>
                            <div class="item-content">
                                <ul class="tips-list">
                                    <li>Overlay：{{$t('创建容器后，会占用虚拟IP，集群内唯一。')}}</li>
                                    <li>Underlay：{{$t('创建容器后，会占用真实的IP；鉴于IP资源的限制，非必须情况下，推荐使用overlay类型。')}}</li>
                                </ul>
                            </div>
                        </div>
                    </template>
                    <div :class="['tips-item', { 'active': tipsActive === 'area' }]">
                        <h6 class="title">{{$t('所属地域')}}</h6>
                        <i18n v-if="isTkeProject" class="item-content" path="地域选择与{link}申请服务器的所属区域保持一致(如深圳、上海等)。">
                            <a :href="PROJECT_CONFIG.doc.yunti" style="color: #3A84FF;" target="_blank" place="link">{{$t('云梯')}}</a>
                        </i18n>
                        <div v-else class="item-content">
                            {{$t('业务服务部署所在的地域，如果没有希望选项，可以选择就近区域。')}}
                        </div>
                    </div>
                    <div :class="['tips-item', { 'active': tipsActive === 'VPC' }]" v-if="isTkeProject">
                        <h6 class="title">{{$t('所属VPC')}}</h6>
                        <p class="item-content">
                            {{$t('VPC选择与服务器所属VPC保持一致。')}}
                            <a :href="PROJECT_CONFIG.doc.VPCLink" style="color: #3A84FF;" target="_blank">{{$t('可用VPC列表')}}</a>
                        </p>
                    </div>
                    <div :class="['tips-item', { 'active': tipsActive === 'netword' }]" v-if="isTkeProject">
                        <h6 class="title">{{$t('容器网络')}}</h6>
                        <div class="item-content">
                            <p style="color: #FF9C01; padding-bottom: 8px;">{{$t('集群创建完成后不可更换，谨慎填写')}}</p>
                            <ul class="tips-list">
                                <li>{{$t('IP数量')}}：{{$t('集群内Pod、Service等资源所需要的网段的大小。')}}</li>
                                <li>{{$t('Service数量上限/集群')}}：{{$t('集群内需要service的数量。')}}</li>
                                <li>{{$t('Pod数量上限/节点')}}：{{$t('每个节点上pod数量上限；由于Master也占用IP资源，')}} {{$t('计算规则: (IP数量-Service的数量)/(Master数量+Node数量)')}}</li>
                            </ul>
                        </div>
                    </div>
                    <div :class="['tips-item', { 'active': tipsActive === 'master' }]">
                        <h6 class="title">{{$t('选择Master')}}</h6>
                        <div class="item-content">
                            <template v-if="isTkeProject">
                                <p>
                                    {{$t('TKE集群至少三个Master，具体规格参考: ')}}
                                    <a :href="PROJECT_CONFIG.doc.masterNodeGuide" style="color: #3A84FF;" target="_blank">
                                        {{$t('选择Master节点的规格。')}}
                                    </a>
                                </p>
                                <p class="mt5">{{$t('选择服务器：使用业务下已有的服务器')}}</p>
                                <p class="mt5">{{$t('申请服务器：新申请服务器，申请成功后，服务器转移到业务下')}}</p>
                            </template>
                            <template v-else>{{$t('测试环境允许单节点；正式环境必须至少三个节点。')}}</template>
                        </div>
                    </div>
                </div>
                <div class="biz-cluster-create-form-footer">
                    <bk-button type="primary" @click="createCluster">{{$t('确定')}}</bk-button>
                    <bk-button type="default" @click="goIndex">{{$t('取消')}}</bk-button>
                </div>
            </div>
        </div>

        <IpSelector v-model="dialogConf.isShow" :ip-list="hostList" @confirm="chooseServer"></IpSelector>

        <tip-dialog
            ref="clusterNoticeDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :title="$t('创建集群')"
            :width="680"
            :sub-title="$t('请确认以下配置：')"
            :check-list="clusterNoticeList"
            :confirm-btn-text="$t('确定，创建集群')"
            :cancel-btn-text="$t('我再想想')"
            :confirm-callback="saveCluster">
        </tip-dialog>
    </div>
</template>

<script>
    // import { bus } from '@/common/bus'
    import applyPerm from '@/mixins/apply-perm'
    import tipDialog from '@/components/tip-dialog'
    import ApplyHost from './apply-host.vue'
    import IpSelector from '@/components/ip-selector/selector-dialog.vue'

    export default {
        components: {
            tipDialog,
            ApplyHost,
            IpSelector
        },
        mixins: [applyPerm],
        beforeRouteLeave (to, from, next) {
            if (this.isChange) {
                const store = this.$store
                store.commit('updateAllowRouterChange', false)
                this.$bkInfo({
                    title: this.$t('确认'),
                    content: this.$t('确定要离开？数据未保存，离开后将会丢失'),
                    confirmFn () {
                        store.commit('updateAllowRouterChange', true)
                        next(true)
                    }
                })
                next(false)
            } else {
                next(true)
            }
        },
        data () {
            return {
                TRUE: true,
                areaId: '',
                areaName: '',
                areaList: [],
                areaIndex: -1,
                vpcId: null,
                vpcList: [],
                vpcIndex: null,
                clusterType: 'stag',
                ccSearchKeys: [],
                checkHostname: true,
                checkService: true,
                dialogConf: {
                    isShow: false,
                    width: 920,
                    hasHeader: false,
                    closeIcon: false
                },
                clusterClassify: 'private',
                clusterNoticeList: [],
                isShowGuide: false,
                pageConf: {
                    total: 1,
                    pageSize: 10,
                    curPage: 1,
                    allCount: 0,
                    show: true
                },
                validate: {
                    name: { illegal: false, msg: '' },
                    description: { illegal: false, msg: '' },
                    area: { illegal: false, msg: '' },
                    vpc: { illegal: false, msg: '' },
                    host: { illegal: false, msg: '' },
                    checkHostname: { illegal: false, msg: '' },
                    checkService: { illegal: false, msg: '' },
                    // IP数量
                    ipNumber: { illegal: false, msg: '' },
                    // Pod数量上限/节点
                    podNumberPerNode: { illegal: false, msg: '' },
                    // service数量上限/集群
                    serviceIpNumber: { illegal: false, msg: '' }
                },
                bkMessageInstance: null,
                // 已选服务器集合
                hostList: [],
                // 已选服务器集合的缓存，用于在弹框中选择，点击确定时才把 hostListCache 赋值给 hostList，同时清空 hostListCache
                // hostListCache: [],
                hostListCache: {},
                // 集群名称
                name: '',
                // nat
                needNat: true,
                // 集群描述
                description: '',
                // 备选服务器集合
                candidateHostList: [],
                // 当前页是否全选中
                isCheckCurPageAll: false,
                isChange: false,
                // 弹层选择 master 节点，已经选择了多少个
                remainCount: 0,
                ccHostLoading: false,
                exceptionCode: null,
                curProject: {},
                isK8sProject: false,
                isTkeProject: false,
                ccApplicationName: '',
                serviceIpNumberKey: '',
                serviceIpNumberList: [],
                ipNumberKey: '',
                ipNumberList: [],
                ipNumberProdList: [],
                ipNumberTestList: [],
                podNumberPerNodeKey: '',
                podNumberPerNodeList: [],
                versionKey: '',
                versionList: [],
                maxPodNumPerNode: 0,
                minPodNumPerNode: 0,
                maxServiceNum: 0,
                calcNodeNum: 0,
                networkKey: 'overlay',
                networkList: [{ id: 'overlay', name: 'overlay' }, { id: 'underlay', name: 'underlay' }],
                ccAppName: '',
                coes: 'mesos',
                tkeMoreConfig: false,
                tipsActive: ''
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
                return this.$store.state.sideMenu.onlineProjectList
            },
            isEn () {
                return this.$store.state.isEn
            },
            disabledSelectMaster () {
                if (this.isTkeProject) {
                    // vpcId可能为上个列表的选项，所以要判断vpcList
                    return !this.areaId || !this.vpcId || !this.vpcList.length
                }
                return false
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            isShowTips () {
                return !!this.ipNumberKey && !!this.serviceIpNumberKey && !!this.podNumberPerNodeKey
            },
            maxClusterCount () {
                // (IP数量-Service的数量) / Pod 数量  = Node数量
                if (this.ipNumberKey && this.serviceIpNumberKey && this.podNumberPerNodeKey) {
                    return Math.floor((this.ipNumberKey - this.serviceIpNumberKey) / this.podNumberPerNodeKey) || 0
                }
                return 0
            }
        },
        watch: {
            name (val) {
                const v = val.trim()
                if (v) {
                    this.isChange = true
                } else {
                    this.isChange = false
                }
                this.validate.name = v
                    ? {
                        illegal: false,
                        msg: ''
                    }
                    : {
                        illegal: true,
                        msg: this.$t('请输入集群名称')
                    }
            },
            description (val) {
                const v = val.trim()
                if (v) {
                    this.isChange = true
                } else {
                    this.isChange = false
                }
                this.validate.description = v
                    ? {
                        illegal: false,
                        msg: ''
                    }
                    : {
                        illegal: true,
                        msg: this.$t('请输入集群描述')
                    }
            },
            areaIndex (val) {
                this.isChange = true
                this.validate.area = val >= 0
                    ? {
                        illegal: false,
                        msg: ''
                    }
                    : {
                        illegal: true,
                        msg: this.$t('请选择地域')
                    }
            },
            vpcIndex (val) {
                this.isChange = true
                this.validate.vpc = val !== null
                    ? {
                        illegal: false,
                        msg: ''
                    }
                    : {
                        illegal: true,
                        msg: this.$t('请选择VPC')
                    }
            },
            hostList (val) {
                const isChange = val.length > 0
                if (isChange) {
                    this.isChange = true
                } else {
                    this.isChange = false
                }
                this.validate.host = val.length >= 0
                    ? {
                        illegal: false,
                        msg: ''
                    }
                    : {
                        illegal: true,
                        msg: this.$t('请选择服务器')
                    }
            },
            clusterType (val) {
                if (val !== 'stag') {
                    this.isChange = true
                    if (this.isTkeProject) {
                        this.ipNumberKey = this.ipNumberProdList[0].id
                    }
                } else {
                    this.isChange = false
                    if (this.isTkeProject) {
                        this.ipNumberKey = this.ipNumberTestList[0].id
                    }
                }
            },
            async ipNumberKey () {
                this.setPodNumberPerNodeList()
                this.setServiceIpNumberList()
                this.vpcList.splice(0, this.vpcList.length, ...[])
                await this.fetchVPC()
            },
            serviceIpNumberKey () {
                this.setPodNumberPerNodeList()
            },
            podNumberPerNodeKey (v) {
                const val = String(v)
                if (val) {
                    this.validate.podNumberPerNode.illegal = false
                    this.validate.podNumberPerNode.msg = ''
                }
            },
            networkKey (val, old) {
                val && val !== old && this.changeNetwork()
            },
            async coes (v) {
                this.clusterType = 'stag'
                this.needNat = true

                this.versionKey = ''
                this.versionList.splice(0, this.versionList.length, ...[])

                this.name = ''
                this.description = ''
                // this.hostSourceKey = 'biz_host_pool'
                this.networkKey = 'overlay'
                this.areaId = ''
                this.areaIndex = -1

                this.$nextTick(() => {
                    this.validate = Object.assign({}, {
                        name: { illegal: false, msg: '' },
                        description: { illegal: false, msg: '' },
                        area: { illegal: false, msg: '' },
                        vpc: { illegal: false, msg: '' },
                        host: { illegal: false, msg: '' },
                        checkHostname: { illegal: false, msg: '' },
                        checkService: { illegal: false, msg: '' },
                        // IP数量
                        ipNumber: { illegal: false, msg: '' },
                        // Pod数量上限/节点
                        podNumberPerNode: { illegal: false, msg: '' },
                        // service数量上限/集群
                        serviceIpNumber: { illegal: false, msg: '' }
                    })
                })

                // k8s
                this.isK8sProject = v === 'k8s'
                // tke
                this.isTkeProject = v === 'tke'

                if (v === 'tke') {
                    await this.getTKEConf()
                }

                if (v === 'k8s') {
                    await this.getK8SConf()
                }

                await this.getAreas()
                await this.getClusters()
            }
        },
        async created () {
            this.curProject = Object.assign({}, this.onlineProjectList.filter(p => p.project_id === this.projectId)[0] || {})

            if (this.curProject.kind === PROJECT_MESOS) {
                this.coes = 'mesos'
            } else {
                this.coes = 'tke'
            }

            // k8s
            this.isK8sProject = this.coes === 'k8s'
            // tke
            this.isTkeProject = this.coes === 'tke'

            if (this.isTkeProject) {
                this.getProject()
            }

            await this.getAreas()
            await this.getClusters()
            this.setPodNumberPerNodeList()
        },
        methods: {
            /**
             * 获取关联 CC 的数据
             */
            async getProject () {
                try {
                    const res = await this.$store.dispatch('getProject', { projectId: this.projectId })
                    const data = res.data || {}
                    this.ccAppName = data.cc_app_name
                } catch (e) {
                    console.log(e)
                }
            },

            hiseChooseServer () {
                this.dialogConf.isShow = false
            },

            /**
             * 设置 Pod数量上限/节点 下拉框的 list
             */
            setPodNumberPerNodeList () {
                const podNumberPerNodeList = []
                // const maxVal = (this.ipNumberKey - this.serviceIpNumberKey) / 4
                const maxVal = Math.min(128, this.maxPodNumPerNode)
                const repeatMap = {}
                for (let i = 1; i < maxVal + 1; i++) {
                    if (i >= 16 && (i & (i - 1)) === 0) {
                        const v = Math.min(i, this.maxPodNumPerNode)
                        if (!repeatMap[v] && v >= this.minPodNumPerNode) {
                            repeatMap[v] = 1
                            podNumberPerNodeList.push({
                                id: v,
                                name: v
                            })
                        }
                    }
                }
                this.podNumberPerNodeList.splice(0, this.podNumberPerNodeList.length, ...podNumberPerNodeList)
                this.podNumberPerNodeKey = ''
            },

            /**
             * 设置Service数量上限/集群 下拉框的 list
             */
            setServiceIpNumberList () {
                const serviceIpNumberList = []
                const maxVal = Math.min(this.ipNumberKey / 2, this.maxServiceNum)
                for (let i = 1; i < maxVal + 1; i++) {
                    if (i >= 16 && (i & (i - 1)) === 0) {
                        serviceIpNumberList.push({
                            id: i,
                            name: i
                        })
                    }
                }
                this.serviceIpNumberList.splice(0, this.serviceIpNumberList.length, ...serviceIpNumberList)
                this.serviceIpNumberKey = ''
            },

            /**
             * 获取所有的集群
             */
            async getClusters () {
                try {
                    await this.$store.dispatch('cluster/getClusterList', this.projectId)
                } catch (e) {
                    console.warn(e)
                }
            },

            /**
             * 获取 k8s version 配置
             */
            async getK8SConf () {
                try {
                    const res = await this.$store.dispatch('cluster/getK8SConf', { projectId: this.projectId })
                    const versionList = []
                    const version = res.data || []
                    version.forEach((item, index) => {
                        if (item.version_id !== '1.8.3') {
                            versionList.push({
                                id: item.version_id,
                                name: item.version_name
                            })
                        }
                    })
                    if (versionList.length) {
                        this.versionKey = versionList[0].id
                    }
                    this.versionList.splice(0, this.versionList.length, ...versionList)
                } catch (e) {
                    console.warn(e)
                }
            },

            /**
             * 获取 tke 配置
             */
            async getTKEConf () {
                try {
                    const res = await this.$store.dispatch('cluster/getTKEConf', { projectId: this.projectId })
                    this.maxPodNumPerNode = res.data.max_pod_num_per_node || 0
                    this.minPodNumPerNode = res.data.min_pod_num_per_node || 0
                    this.maxServiceNum = res.data.max_service_num || 0

                    const ipNumberProdList = []
                    const ipNumberProd = res.data.ip_number_for_prod_env || []
                    ipNumberProd.forEach((item, index) => {
                        if (index === 0 && this.clusterType === 'prod') {
                            this.ipNumberKey = item
                        }
                        ipNumberProdList.push({
                            id: item,
                            name: item
                        })
                    })
                    this.ipNumberProdList.splice(0, this.ipNumberProdList.length, ...ipNumberProdList)

                    const ipNumberTestList = []
                    const ipNumberText = res.data.ip_number_for_test_env || []
                    ipNumberText.forEach((item, index) => {
                        if (index === 0 && this.clusterType === 'stag') {
                            this.ipNumberKey = item
                        }
                        ipNumberTestList.push({
                            id: item,
                            name: item
                        })
                    })
                    this.ipNumberTestList.splice(0, this.ipNumberTestList.length, ...ipNumberTestList)

                    const versionList = []
                    const version = res.data.version_list || []
                    version.forEach((item, index) => {
                        if (item.version_id === '1.18.4') {
                            this.versionKey = item.version_id
                        }
                        versionList.push({
                            id: item.version_id,
                            name: item.version_name
                        })
                    })
                    this.versionList.splice(0, this.versionList.length, ...versionList)
                } catch (e) {
                    console.warn(e)
                }
            },

            /**
             * 获取所属地域
             */
            async getAreas () {
                try {
                    const res = await this.$store.dispatch('cluster/getAreaList', {
                        projectId: this.projectId,
                        data: {
                            coes: this.coes
                        }
                    })
                    const areaList = []
                    const list = res.data.results || []
                    list.forEach(item => {
                        areaList.push({
                            areaId: item.id,
                            areaName: item.name,
                            showName: item.chinese_name
                        })
                    })
                    this.areaList.splice(0, this.areaList.length, ...areaList)
                } catch (e) {
                    console.log(e)
                }
            },

            /**
             * 选择所属地域
             */
            async changeArea (index, data) {
                this.areaId = data.areaId
                this.areaName = data.areaName

                this.vpcList.splice(0, this.vpcList.length, ...[])
                if (this.isTkeProject) {
                    await this.fetchVPC()
                }
            },

            /**
             * 选择网络类型
             */
            async changeNetwork (index, data) {
                this.vpcList.splice(0, this.vpcList.length, ...[])
                await this.fetchVPC()
            },

            /**
             * 选择所属 VPC
             */
            changeVPC (index, data) {
                this.vpcId = data.vpcId

                const len = this.vpcList.length
                for (let i = len - 1; i >= 0; i--) {
                    if (String(this.vpcList[i].vpcId) === String(data.vpcId)) {
                        this.vpcIndex = data.vpcId
                        break
                    }
                }
            },

            /**
             * 切换测试环境和正式环境
             */
            toggleDev () {
                this.validate.host = {
                    illegal: false,
                    msg: ''
                }

                this.hostList.splice(0, this.hostList.length, ...[])
                this.hostListCache = Object.assign({}, {})
                this.isCheckCurPageAll = false
                this.pageConf.curPage = 1
            },

            async fetchVPC () {
                if (!this.areaName || !this.ipNumberKey || !this.networkKey) {
                    return
                }
                try {
                    const res = await this.$store.dispatch('cluster/getVPC', {
                        projectId: this.projectId,
                        data: {
                            region_name: this.areaName,
                            cidr_size: this.ipNumberKey,
                            network_type: this.networkKey,
                            coes: this.coes
                        }
                    })
                    const vpc = res.data || {}
                    const vpcList = []
                    Object.keys(vpc).forEach(key => {
                        vpcList.push({
                            vpcId: vpc[key],
                            vpcName: key
                        })
                    })
                    this.vpcList.splice(0, this.vpcList.length, ...vpcList)
                } catch (e) {
                    console.log(e)
                }
            },

            /**
             * 获取 cc 表格数据
             *
             * @param {Object} params ajax 查询参数
             */
            async fetchCCData (params = {}) {
                this.ccHostLoading = true
                try {
                    const args = {
                        projectId: this.projectId,
                        limit: this.pageConf.pageSize,
                        offset: params.offset,
                        ip_list: params.ipList || [],
                        coes: this.coes
                    }
                    if (this.isTkeProject) {
                        // args.host_source = this.hostSourceKey
                        args.node_role = 'master'
                        args.cluster_env = this.clusterType
                        args.region_name = this.areaName
                        args.network_type = this.networkKey
                    }

                    const res = await this.$store.dispatch('cluster/getCCHostList', args)

                    this.ccApplicationName = res.data.cc_application_name || ''

                    const count = res.data.count

                    this.pageConf.show = !!count
                    this.pageConf.total = Math.ceil(count / this.pageConf.pageSize)
                    if (this.pageConf.total < this.pageConf.curPage) {
                        this.pageConf.curPage = 1
                    }
                    this.pageConf.show = true
                    this.pageConf.allCount = count

                    const list = res.data.results || []
                    list.forEach(item => {
                        if (this.hostListCache[`${item.inner_ip}-${item.asset_id}`]) {
                            item.isChecked = true
                        } else {
                            item.isChecked = false
                        }
                    })

                    this.candidateHostList.splice(0, this.candidateHostList.length, ...list)
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
            async openDialog () {
                this.remainCount = 0
                this.pageConf.curPage = 1
                this.pageConf.allCount = 0
                this.dialogConf.isShow = true
                this.candidateHostList.splice(0, this.candidateHostList.length, ...[])
                this.isCheckCurPageAll = false
                // this.$refs.iPSearcher.clearSearchParams()

                // 会触发请求
                // await this.fetchCCData({
                //     offset: 0
                // })
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            pageChange (page) {
                this.pageConf.curPage = page
                this.fetchCCData({
                    offset: this.pageConf.pageSize * (page - 1),
                    ipList: this.ccSearchKeys || []
                })
            },

            /**
             * 弹层表格全选值修改
             */
            handleInputCheckCurPage (value) {
                this.isCheckCurPageAll = value
            },

            /**
             * 在选择服务器弹层中选择
             */
            selectHost (hosts = this.candidateHostList) {
                if (!hosts.length) {
                    return
                }

                this.$nextTick(() => {
                    const illegalLen = hosts.filter(host => host.is_used || String(host.agent) !== '1' || !host.is_valid).length
                    const selectedHosts = hosts.filter(host =>
                        host.isChecked === true && !host.is_used && String(host.agent) === '1' && host.is_valid
                    )

                    if (selectedHosts.length === hosts.length - illegalLen && hosts.length !== illegalLen) {
                        this.isCheckCurPageAll = true
                    } else {
                        this.isCheckCurPageAll = false
                    }

                    // 清除 hostListCache
                    hosts.forEach(item => {
                        delete this.hostListCache[`${item.inner_ip}-${item.asset_id}`]
                    })

                    // 重新根据选择的 host 设置到 hostListCache 中
                    selectedHosts.forEach(item => {
                        this.hostListCache[`${item.inner_ip}-${item.asset_id}`] = item
                    })

                    this.remainCount = Object.keys(this.hostListCache).length
                })
            },

            /**
             * 选择服务器弹层搜索事件
             *
             * @param {Array} searchKeys 搜索字符数组
             */
            handleSearch (searchKeys) {
                this.ccSearchKeys = searchKeys
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
            chooseServer (hostList = []) {
                const len = hostList.length
                if (!len) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择服务器')
                    })
                    return
                }

                if (len % 2 === 0) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请选择奇数个服务器')
                    })
                    return
                }

                if (this.isTkeProject && len < 3) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('最少选择三个服务器')
                    })
                    return
                }

                if (len > 7) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('最多选择七个服务器')
                    })
                    return
                }

                if (this.clusterType !== 'stag' && len < 3) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: this.$t('正式环境最少选择三个服务器，最多选择七个服务器')
                    })
                    return
                }

                this.dialogConf.isShow = false
                this.hostList.splice(0, this.hostList.length, ...hostList)
                this.isCheckCurPageAll = false
            },

            /**
             * 验证 form
             */
            formValidation () {
                let msg = ''
                if (!this.name.trim()) {
                    this.validate.name.illegal = true
                    this.validate.name.msg = this.$t('请输入集群名称')
                    msg = this.$t('请输入集群名称')
                } else if (!this.description.trim()) {
                    this.validate.description.illegal = true
                    this.validate.description.msg = this.$t('请输入集群描述')
                    msg = this.$t('请输入集群描述')
                } else if (this.areaIndex === -1) {
                    this.validate.area.illegal = true
                    this.validate.area.msg = this.$t('请选择所属地域')
                    msg = this.$t('请选择所属地域')
                } else if (this.isTkeProject && !this.vpcIndex) {
                    this.validate.vpc.illegal = true
                    this.validate.vpc.msg = this.$t('请选择所属VPC')
                    msg = this.$t('请选择所属VPC')
                } else if (this.isTkeProject && !this.podNumberPerNodeKey) {
                    this.validate.podNumberPerNode.illegal = true
                    this.validate.podNumberPerNode.msg = this.$t('请选择Pod数量上限/节点')
                    msg = this.$t('请选择Pod数量上限/节点')
                } else if (!this.hostList.length) {
                    this.validate.host.illegal = true
                    this.validate.host.msg = this.$t('请选择服务器')
                    msg = this.$t('请选择服务器')
                } else if (!this.checkHostname) {
                    this.validate.checkHostname.illegal = true
                    this.validate.checkHostname.msg = this.$t('请确认注意事项内容')
                    msg = this.$t('请确认注意事项内容')
                } else if (!this.checkService) {
                    this.validate.checkService.illegal = true
                    this.validate.checkService.msg = this.$t('请确认注意事项内容')
                    msg = this.$t('请确认注意事项内容')
                }

                if (msg) {
                    this.bkMessageInstance && this.bkMessageInstance.close()
                    this.bkMessageInstance = this.$bkMessage({
                        theme: 'error',
                        message: msg
                    })
                    return false
                }
                return true
            },

            /**
             * 已选服务器移除处理
             *
             * @param {Object} host 当前行的服务器
             */
            removeHost (host) {
                const index = this.hostList.findIndex(item => item.bk_host_innerip === host.bk_host_innerip && item.bk_cloud_id === host.bk_cloud_id)
                if (index > -1) {
                    this.hostList.splice(index, 1)
                }
            },

            /**
             * 确定按钮事件
             */
            async createCluster () {
                if (!this.formValidation()) {
                    return
                }

                if (this.isTkeProject) {
                    // (IP数量-Service的数量)/Pod数量上限每节点 - master数量 = x（向下取整）
                    this.calcNodeNum = Math.floor(
                        (this.ipNumberKey - this.serviceIpNumberKey) / this.podNumberPerNodeKey - this.hostList.length
                    )
                    this.clusterNoticeList.splice(0, this.clusterNoticeList.length, ...[
                        {
                            id: 1,
                            text: this.$t('集群仅允许创建{x}个节点，{y}个service，每个节点{z}个pod，创建后，不允许调整', {
                                x: this.calcNodeNum,
                                y: this.serviceIpNumberKey,
                                z: this.podNumberPerNodeKey
                            }),
                            isChecked: false
                        }
                    ])
                    this.$refs.clusterNoticeDialog.show()
                } else {
                    await this.saveCluster()
                }
            },

            async saveCluster () {
                const params = {
                    name: this.name,
                    area_id: this.areaId,
                    environment: this.clusterType,
                    cluster_type: this.clusterClassify,
                    master_ips: [],
                    need_nat: this.needNat,
                    description: this.description,
                    projectId: this.projectId,
                    version: this.versionKey,
                    vpc_id: this.vpcId,
                    coes: this.coes
                }
                if (this.isTkeProject) {
                    params.network_type = this.networkKey
                    // params.host_source = this.hostSourceKey
                    // params.kube_proxy_mode = this.kubeProxyMode
                }

                this.hostList.forEach(item => {
                    params.master_ips.push(item.bk_host_innerip)
                })

                if (this.isTkeProject) {
                    params.ip_number = this.ipNumberKey
                    params.pod_number_per_node = this.podNumberPerNodeKey
                    params.service_ip_number = this.serviceIpNumberKey
                }

                this.$refs.clusterNoticeDialog.hide()

                const h = this.$createElement
                this.$bkLoading({
                    title: h('span', this.$t('下发集群配置中，请稍候...'))
                })

                try {
                    await this.$store.dispatch('cluster/createCluster', params)
                    this.isChange = false
                    this.$bkMessage({
                        message: this.$t('下发集群配置完成'),
                        theme: 'success',
                        delay: 1000,
                        onClose: () => {
                            this.$bkLoading.hide()
                            this.goIndex()
                        }
                    })
                } catch (e) {
                    this.$bkLoading.hide()
                    if (e.code === 404) {
                        this.exceptionCode = {
                            code: '404',
                            msg: this.$t('当前访问的集群不存在')
                        }
                        this.isChange = false
                    } else if (e.code === 403) {
                        this.exceptionCode = {
                            code: '403',
                            msg: this.$t('Sorry，您的权限不足!')
                        }
                        this.isChange = false
                    }
                }
            },

            /**
             * 返回集群首页列表
             */
            goIndex () {
                if (this.curClusterId) {
                    this.$router.back()
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
             * 显示快速入门侧边栏
             */
            showGuide () {
                const guide = this.$refs.clusterGuide
                guide.show()
            },

            /**
             * 切换快速入门侧边栏状态
             *
             * @param {boolean} status 状态
             */
            toggleGuide (status) {
                this.isShowGuide = status
            }
        }
    }
</script>

<style scoped lang="postcss">
    @import './create.css';

    .server-tip {
        float: left;
        line-height: 17px;
        font-size: 12px;
        text-align: left;
        padding: 13px 0 13px 20px;
        margin-left: 20px;

        li {
            list-style: circle;
        }
    }

    .bk-dialog-footer .bk-dialog-outer button {
        margin-top: 20px;
    }
</style>
