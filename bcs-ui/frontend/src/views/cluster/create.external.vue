<template>
    <div class="biz-content">
        <div class="biz-top-bar">
            <div class="biz-cluster-create-title" @click="goIndex">
                <i class="bcs-icon bcs-icon-arrows-left back"></i>
                <span>{{$t('创建容器集群')}}</span>
            </div>
            <div class="biz-actions">
                <bcs-popover :content="$t('快速入门')" placement="left" :transfer="true">
                    <a class="button" href="javascript:void(0)" @click.stop.prevent="showGuide">
                        <i class="bcs-icon bcs-icon-calendar"></i>
                    </a>
                </bcs-popover>
            </div>
        </div>
        <div class="biz-content-wrapper">
            <app-exception v-if="exceptionCode" :type="exceptionCode.code" :text="exceptionCode.msg"></app-exception>
            <div v-else class="biz-cluster-create-form-wrapper">
                <template v-if="clusterClassify !== 'public'">
                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''">
                        <label>{{$t('集群分类')}}：<span class="red">*</span></label>
                        <div class="form-item-inner">
                            <bk-radio-group v-model="clusterState">
                                <bk-radio value="bcs_new">{{$t('新建集群')}}</bk-radio>
                                <bk-radio value="existing">
                                    {{$t('导入集群')}}
                                    <i class="bcs-icon bcs-icon-question-circle" style="vertical-align: middle; cursor: pointer;" v-bk-tooltips="$t('导入已经存在的集群')"></i>
                                </bk-radio>
                            </bk-radio-group>
                        </div>
                    </div>
                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''">
                        <label>{{$t('名称：')}}<span class="red">*</span></label>
                        <div class="form-item-inner">
                            <input maxlength="64" type="text" class="bk-form-input cluster-name" :placeholder="$t('请输入集群名称')"
                                :class="validate.name.illegal ? 'is-danger' : ''" v-model="name"
                            >
                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.name.illegal">
                                <p class="tip-text">{{$t('必填项，不超过64个字符')}}</p>
                            </div>
                        </div>
                    </div>
                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''">
                        <label>{{$t('集群描述：')}}<span class="red">*</span></label>
                        <div class="form-item-inner">
                            <textarea maxlength="128" v-model="description" class="bk-form-textarea" :class="validate.description.illegal ? 'is-danger' : ''" :placeholder="$t('请输入集群描述')"></textarea>
                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.description.illegal">
                                <p class="tip-text">{{$t('必填项，不超过128个字符')}}</p>
                            </div>
                        </div>
                    </div>
                    <div class="form-item" :class="isEn ? 'en' : ''">
                        <label>{{$t('选择Master：')}}<span class="red">*</span></label>
                        <div class="form-item-inner">
                            <bk-button type="default" :class="validate.host.illegal ? 'is-danger' : ''" @click="openDialog">{{$t('选择服务器')}}</bk-button>
                            <div class="is-danger biz-cluster-create-form-tip" v-if="validate.host.illegal">
                                <p class="tip-text">{{$t('请选择服务器')}}</p>
                            </div>
                        </div>
                    </div>
                    <div class="form-item" :class="isEn ? 'en' : ''" v-if="hostList.length">
                        <label></label>
                        <div class="form-item-inner">
                            <div class="biz-cluster-create-table-header">
                                <div class="left">
                                    {{$t('已选服务器')}}
                                </div>
                            </div>
                            <table class="bk-table has-table-hover biz-table biz-cluster-create-table">
                                <thead>
                                    <tr>
                                        <th style="text-align: left;padding-left: 30px;">
                                            {{$t('序号')}}
                                        </th>
                                        <th>{{$t('IP地址')}}</th>
                                        <th>{{$t('操作')}}</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(host, index) in hostList" :key="index">
                                        <td style="text-align: left;padding-left: 30px;">
                                            {{index + 1}}
                                        </td>
                                        <td>{{host.bk_host_innerip}}</td>
                                        <td><a href="javascript:void(0)" class="bk-text-button" @click="removeHost(host, index)">{{$t('移除')}}</a></td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div class="form-item bk-form-item" :class="isEn ? 'en' : ''">
                        <label class="mt10">{{$t('注意事项：')}}<span class="red">*</span></label>
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
                    <div class="form-item" :class="isEn ? 'en' : ''">
                        <label></label>
                        <div class="form-item-inner">
                            <bk-button type="primary" @click="createCluster">{{$t('确定')}}</bk-button>
                            <bk-button type="default" @click="goIndex">{{$t('取消')}}</bk-button>
                        </div>
                    </div>
                </template>
            </div>
        </div>

        <IpSelector v-model="dialogConf.isShow" :ip-list="hostList" @confirm="chooseServer"></IpSelector>

        <tip-dialog
            ref="clusterNoticeDialog"
            icon="bcs-icon bcs-icon-exclamation-triangle"
            :title="$t('创建集群')"
            :sub-title="$t('此操作需要对你的主机进行如下操作，请知悉：')"
            :check-list="clusterNoticeList"
            :confirm-btn-text="$t('确定，创建集群')"
            :cancel-btn-text="$t('我再想想')"
            :confirm-callback="saveCluster">
        </tip-dialog>
    </div>
</template>

<script>
    import applyPerm from '@/mixins/apply-perm'
    import tipDialog from '@/components/tip-dialog'
    import IpSelector from '@/components/ip-selector/selector-dialog.vue'

    // import { bus } from '@/common/bus'

    export default {
        components: {
            tipDialog,
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
                clusterType: 'prod',
                ccSearchKeys: [],
                dialogConf: {
                    isShow: false,
                    width: 920,
                    hasHeader: false,
                    closeIcon: false
                },
                checkHostname: true,
                checkService: true,
                clusterClassify: 'private',
                clusterNoticeList: [
                    {
                        id: 1,
                        text: this.$t('按照规则修改主机名'),
                        isChecked: false
                    },
                    {
                        id: 2,
                        text: this.$t('安装容器服务相关的组件'),
                        isChecked: false
                    }
                ],
                isShowGuide: false,
                pageConf: {
                    totalPage: 1,
                    pageSize: 10,
                    curPage: 1,
                    show: true
                },
                validate: {
                    name: {
                        illegal: false,
                        msg: ''
                    },
                    description: {
                        illegal: false,
                        msg: ''
                    },
                    host: {
                        illegal: false,
                        msg: ''
                    },
                    checkHostname: {
                        illegal: false,
                        msg: ''
                    },
                    checkService: {
                        illegal: false,
                        msg: ''
                    }
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
                showStagTip: false,
                exceptionCode: null,
                curProject: {},
                isK8sProject: false,
                ccApplicationName: '',
                clusterState: 'bcs_new'
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
            }
        },
        watch: {
            name (val) {
                const v = val.trim()
                if (v) {
                    this.isChange = true
                    this.showStagTip = true
                } else {
                    this.isChange = false
                    this.showStagTip = false
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
            }
        },
        mounted () {
            this.curProject = Object.assign({}, this.onlineProjectList.filter(p => p.project_id === this.projectId)[0] || {})
            // k8s
            this.isK8sProject = this.curProject.kind === PROJECT_K8S
        },
        methods: {
            hiseChooseServer () {
                this.dialogConf.isShow = false
            },

            /**
             * 获取所有的集群
             */
            async getClusters () {
                try {
                    const res = await this.$store.dispatch('cluster/getClusterList', this.projectId)
                } catch (e) {
                    console.warn(e)
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

            /**
             * 获取 cc 表格数据
             *
             * @param {Object} params ajax 查询参数
             */
            async fetchCCData (params = {}) {
                this.ccHostLoading = true
                try {
                    const res = await this.$store.dispatch('cluster/getCCHostList', {
                        projectId: this.projectId,
                        limit: this.pageConf.pageSize,
                        offset: params.offset,
                        ip_list: params.ipList || []
                    })

                    this.ccApplicationName = res.data.cc_application_name || ''

                    const count = res.data.count

                    this.pageConf.show = !!count
                    this.pageConf.totalPage = Math.ceil(count / this.pageConf.pageSize)
                    if (this.pageConf.totalPage < this.pageConf.curPage) {
                        this.pageConf.curPage = 1
                    }

                    const list = res.data.results || []
                    list.forEach(item => {
                        if (this.hostListCache[`${item.inner_ip}-${item.asset_id}`]) {
                            item.isChecked = true
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
                this.dialogConf.isShow = true
                this.candidateHostList.splice(0, this.candidateHostList.length, ...[])
                this.isCheckCurPageAll = false
            },

            /**
             * 翻页回调
             *
             * @param {number} page 当前页
             */
            pageChange (page) {
                this.fetchCCData({
                    offset: this.pageConf.pageSize * (page - 1),
                    ipList: this.ccSearchKeys || []
                })
            },

            /**
             * 在选择服务器弹层中选择
             */
            selectHost (hosts = this.candidateHostList) {
                if (!hosts.length) {
                    return
                }
                setTimeout(() => {
                    const illegalLen = hosts.filter(host => host.is_used || String(host.agent) !== '1').length
                    const selectedHosts = hosts.filter(host =>
                        host.isChecked === true && !host.is_used && String(host.agent) === '1'
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
             * @param {number} index 当前行的服务器的索引
             */
            removeHost (host, index) {
                this.hostList.splice(index, 1)
                // delete this.hostListCache[`${host.inner_ip}-${host.asset_id}`]
            },

            /**
             * 确定按钮事件
             */
            createCluster () {
                if (!this.formValidation()) {
                    return
                }
                this.saveCluster()
                // this.clusterNoticeList.forEach(item => {
                //     item.isChecked = false
                // })
                // this.$refs.clusterNoticeDialog.show()
            },

            async saveCluster () {
                const params = {
                    name: this.name,
                    environment: this.clusterType,
                    cluster_type: this.clusterClassify,
                    master_ips: [],
                    need_nat: this.needNat,
                    description: this.description,
                    projectId: this.projectId,
                    cluster_state: this.clusterState
                }
                this.hostList.forEach(item => {
                    params.master_ips.push(item.bk_host_innerip)
                })

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
                    if (!e.code || e.code === 404) {
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
                this.$router.push({
                    name: 'clusterMain',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
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
