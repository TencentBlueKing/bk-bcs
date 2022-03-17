<template>
    <div class="biz-content">
        <biz-header ref="commonHeader"
            @exception="exceptionHandler"
            @saveIngressSuccess="saveIngressSuccess"
            @switchVersion="initResource"
            @exmportToYaml="exportToYaml">
        </biz-header>
        <div class="biz-content-wrapper biz-confignation-wrapper" v-bkloading="{ isLoading: isTemplateSaving }">
            <app-exception
                v-if="exceptionCode && !isDataLoading"
                :type="exceptionCode.code"
                :text="exceptionCode.msg">
            </app-exception>
            <div class="biz-tab-box" v-else v-show="!isDataLoading" style="overflow:hidden;">
                <biz-tabs @tab-change="tabResource" ref="commonTab"></biz-tabs>
                <bk-dialog
                    :is-show.sync="dialogConf.isShow"
                    :width="dialogConf.width"
                    :content="dialogConf.content"
                    :has-header="dialogConf.hasHeader"
                    :close-icon="dialogConf.closeIcon"
                    :quick-close="false"
                    @confirm="addLocalIngress"
                    @cancel="dialogConf.isShow = false">
                    <template slot="content">
                        <div style="margin: -20px;">
                            <div class="ingress-type-header">
                                {{$t('请选择类型')}}
                            </div>
                            <div class="bk-form" style="padding: 30px 0 10px 0;">
                                <div class="bk-form-item">
                                    <label class="bk-label" style="width: 140px;">{{$t('类型')}}：</label>
                                    <div class="bk-form-content" style="margin-left: 140px">
                                        <label class="bk-form-radio">
                                            <input type="radio" name="type" value="STGW" disabled="disabled">
                                            <i class="bk-radio-text">STGW</i>
                                        </label>
                                        <label class="bk-form-radio">
                                            <input type="radio" name="type" value="TGW" disabled="disabled">
                                            <i class="bk-radio-text">TGW</i>
                                        </label>
                                        <label class="bk-form-radio">
                                            <input type="radio" name="type" value="K8S" checked="checked">
                                            <i class="bk-radio-text">{{$t('K8S原生')}}</i>
                                        </label>
                                    </div>
                                </div>
                            </div>
                            <p class="pb30 f12 biz-danger-text">{{$t('提示：调度引擎为BCS-K8S的集群，需在左侧菜单“网络” => “LoadBalancer”中新建LoadBalancer，Ingress规则才能生效（通常单个集群只需要创建一个LoadBalancer）')}}</p>
                        </div>
                    </template>
                </bk-dialog>

                <div class="biz-tab-content" v-bkloading="{ isLoading: isTabChanging }">
                    <bk-alert type="info" class="mb20">
                        <div slot="title">
                            <div>
                                {{$t('Ingress是管理外部访问集群内服务的对象，可配置访问的URL、基于名称的虚拟主机等。 Ingress controller负责实现Ingress，BCS使用的是nginx-controller')}}，
                                <a class="bk-text-button" :href="PROJECT_CONFIG.doc.k8sIngress" target="_blank">{{$t('详情查看文档')}}</a>
                            </div>
                            <div class="mt5">
                                {{$t('提示：调度引擎为BCS-K8S的集群，需在左侧菜单“网络” => “LoadBalancer”中新建LoadBalancer，Ingress规则才能生效（通常单个集群只需要创建一个LoadBalancer）')}}
                            </div>
                        </div>
                    </bk-alert>
                    <template v-if="!ingresss.length">
                        <div class="biz-guide-box mt0">
                            <bk-button type="primary" @click.stop.prevent="addLocalIngress">
                                <i class="bcs-icon bcs-icon-plus"></i>
                                <span style="margin-left: 0;">{{$t('添加')}}Ingress</span>
                            </bk-button>
                        </div>
                    </template>
                    <template v-else>
                        <div class="biz-configuration-topbar">
                            <div class="biz-list-operation">
                                <div class="item" v-for="(ingress, index) in ingresss" :key="ingress.id">
                                    <bk-button :class="['bk-button', { 'bk-primary': curIngress.id === ingress.id }]" @click.stop="setCurIngress(ingress, index)">
                                        {{(ingress && ingress.config.metadata.name) || $t('未命名')}}
                                        <span class="biz-update-dot" v-show="ingress.isEdited"></span>
                                    </bk-button>
                                    <span class="bcs-icon bcs-icon-close" @click.stop="removeIngress(ingress, index)"></span>
                                </div>

                                <bcs-popover ref="applicationTooltip" :content="$t('添加Ingress')" placement="top">
                                    <bk-button class="bk-button bk-default is-outline is-icon" @click.stop="addLocalIngress">
                                        <i class="bcs-icon bcs-icon-plus"></i>
                                    </bk-button>
                                </bcs-popover>
                            </div>
                        </div>
                        <div class="biz-configuration-content" style="position: relative;">
                            <k8s-ingress :ingress-data="curIngress" :key="curIngress" :version="curVersion"></k8s-ingress>
                        </div>
                    </template>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import header from './header.vue'
    import tabs from './tabs.vue'
    import ingressParams from '@/json/k8s-ingress.json'

    export default {
        components: {
            'biz-header': header,
            'biz-tabs': tabs
        },
        data () {
            return {
                isTabChanging: false,
                curIngress: ingressParams,
                isDataLoading: true,
                exceptionCode: null,
                compareTimer: 0,
                setTimer: 0,
                dialogConf: {
                    isShow: false,
                    width: 500,
                    hasHeader: false,
                    closeIcon: false
                },
                isIngressTypeShow: true
            }
        },
        computed: {
            varList () {
                return this.$store.state.variable.varList
            },
            versionList () {
                const list = this.$store.state.k8sTemplate.versionList
                return list
            },
            isTemplateSaving () {
                return this.$store.state.k8sTemplate.isTemplateSaving
            },
            curTemplate () {
                return this.$store.state.k8sTemplate.curTemplate
            },
            deployments () {
                return this.$store.state.k8sTemplate.deployments
            },
            services () {
                return this.$store.state.k8sTemplate.services
            },
            configmaps () {
                return this.$store.state.k8sTemplate.configmaps
            },
            secrets () {
                return this.$store.state.k8sTemplate.secrets
            },
            daemonsets () {
                return this.$store.state.k8sTemplate.daemonsets
            },
            jobs () {
                return this.$store.state.k8sTemplate.jobs
            },
            statefulsets () {
                return this.$store.state.k8sTemplate.statefulsets
            },
            ingresss () {
                return this.$store.state.k8sTemplate.ingresss
            },
            curVersion () {
                return this.$store.state.k8sTemplate.curVersion
            },
            templateId () {
                return this.$route.params.templateId
            },
            projectId () {
                return this.$route.params.projectId
            }
        },
        async beforeRouteLeave (to, form, next) {
            // 修改模板集信息
            // await this.$refs.commonHeader.saveTemplate()
            clearInterval(this.compareTimer)
            clearTimeout(this.setTimer)
            next(true)
        },
        mounted () {
            this.isDataLoading = true
            this.$refs.commonHeader.initTemplate((data) => {
                this.initResource(data)
                this.isDataLoading = false
            })
        },
        methods: {
            async initServices (version) {
                const projectId = this.projectId
                await this.$store.dispatch('k8sTemplate/getServicesByVersion', { projectId, version })
            },
            exceptionHandler (exceptionCode) {
                this.isDataLoading = false
                this.exceptionCode = exceptionCode
            },
            async initResource (data) {
                const version = data.latest_version_id || data.version

                if (version) {
                    await this.initServices(version)
                } else {
                    this.isLoadingServices = false
                }
                if (data.ingresss && data.ingresss.length) {
                    this.setCurIngress(data.ingresss[0], 0)
                } else if (data.ingress && data.ingress.length) {
                    this.setCurIngress(data.ingress[0], 0)
                }
            },
            exportToYaml (data) {
                this.$router.push({
                    name: 'K8sYamlTemplateset',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode,
                        templateId: 0
                    },
                    query: {
                        action: 'export'
                    }
                })
            },
            saveIngressSuccess (params) {
                this.ingresss.forEach(item => {
                    if (params.responseData.id === item.id || params.preId === item.id) {
                        item.cache = JSON.parse(JSON.stringify(item))
                    }
                })
                if (params.responseData.id === this.curIngress.id || params.preId === this.curIngress.config.id) {
                    this.updateLocalData(params.resource)
                }
            },
            updateLocalData (data) {
                if (data.id) {
                    this.curIngress.config.id = data.id
                    this.curIngressId = data.id
                }
                if (data.version) {
                    this.$store.commit('k8sTemplate/updateCurVersion', data.version)
                }

                this.$store.commit('k8sTemplate/updateIngresss', this.ingresss)
                setTimeout(() => {
                    this.ingresss.forEach(item => {
                        if (item.id === data.id) {
                            this.setCurIngress(item)
                        }
                    })
                }, 500)
            },
            async tabResource (type, target) {
                this.isTabChanging = true
                await this.$refs.commonHeader.saveTemplate()
                await this.$refs.commonHeader.autoSaveResource(type)
                this.$refs.commonTab.goResource(target)
            },
            addLocalIngress () {
                const ingress = JSON.parse(JSON.stringify(ingressParams))
                const index = this.ingresss.length
                const now = +new Date()
                const ingressName = 'ingress-' + (index + 1)

                ingress.id = 'local_' + now
                ingress.isEdited = true

                ingress.config.metadata.name = ingressName
                this.ingresss.push(ingress)

                this.setCurIngress(ingress, index)
                this.dialogConf.isShow = false
            },
            setCurIngress (ingress, index) {
                this.curIngress = ingress
                this.curIngressId = ingress.id

                clearInterval(this.compareTimer)
                clearTimeout(this.setTimer)
                this.setTimer = setTimeout(() => {
                    if (!this.curIngress.cache) {
                        this.curIngress.cache = JSON.parse(JSON.stringify(ingress))
                    }
                    this.watchChange()
                }, 500)
            },
            watchChange () {
                this.compareTimer = setInterval(() => {
                    const appCopy = JSON.parse(JSON.stringify(this.curIngress))
                    const cacheCopy = JSON.parse(JSON.stringify(this.curIngress.cache))
                    // 删除无用属性
                    delete appCopy.isEdited
                    delete appCopy.cache
                    delete appCopy.id

                    delete cacheCopy.isEdited
                    delete cacheCopy.cache
                    delete cacheCopy.id

                    const appStr = JSON.stringify(appCopy)
                    const cacheStr = JSON.stringify(cacheCopy)

                    if (String(this.curIngress.id).indexOf('local_') > -1) {
                        this.curIngress.isEdited = true
                    } else if (appStr !== cacheStr) {
                        this.curIngress.isEdited = true
                    } else {
                        this.curIngress.isEdited = false
                    }
                }, 1000)
            },
            removeLocalIngress (ingress, index) {
                // 是否删除当前项
                if (this.curIngress.id === ingress.id) {
                    if (index === 0 && this.ingresss[index + 1]) {
                        this.setCurIngress(this.ingresss[index + 1])
                    } else if (this.ingresss[0]) {
                        this.setCurIngress(this.ingresss[0])
                    }
                }
                this.ingresss.splice(index, 1)
            },
            showIngressTypeBox () {
                this.dialogConf.isShow = true
            },
            hideIngressTypeBox () {
                this.isIngressTypeShow = false
            },
            removeIngress (ingress, index) {
                const self = this
                const projectId = this.projectId
                const version = this.curVersion
                const ingressId = ingress.id
                this.$bkInfo({
                    title: this.$t('确认删除'),
                    content: this.$createElement('p', { style: { 'text-align': 'left' } }, `${this.$t('删除Ingress')}：${ingress.config.metadata.name || this.$t('未命名')}`),
                    confirmFn () {
                        if (ingressId.indexOf && ingressId.indexOf('local_') > -1) {
                            self.removeLocalIngress(ingress, index)
                        } else {
                            self.$store.dispatch('k8sTemplate/removeIngress', { ingressId, version, projectId }).then(res => {
                                const data = res.data
                                self.removeLocalIngress(ingress, index)

                                if (data.version) {
                                    self.$store.commit('k8sTemplate/updateCurVersion', data.version)
                                    self.$store.commit('k8sTemplate/updateBindVersion', true)
                                }
                            }, res => {
                                const message = res.message
                                self.$bkMessage({
                                    theme: 'error',
                                    message: message
                                })
                            })
                        }
                    }
                })
            }
        }
    }
</script>

<style scoped>
    @import './ingress.css';
</style>
