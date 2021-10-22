<template>
    <div id="app" class="biz-app" :class="systemCls">
        <app-header ref="appHeader" @reloadCurPage="reloadCurPage"></app-header>
        <!-- 这里的v-if不要去掉，为了解决项目初始化问题，todo：后续优化!!! -->
        <div style="height: 100%;" v-if="isLoading">
            <div class="bk-loading" v-bkloading="{ isLoading }"></div>
        </div>
        <template v-else>
            <div class="app-container" :style="{ minHeight: minHeight + 'px' }" v-if="isUserBKService">
                <router-view :key="routerKey" />
            </div>
            <div v-else>
                <bcs-unregistry :cc-list="ccList"
                    :default-kind="kind"
                    @kind-change="handleKindChange"
                    @cc-change="handleCmdbChange"
                    @update-project="updateProject">
                </bcs-unregistry>
            </div>
        </template>
        <app-apply-perm ref="bkApplyPerm"></app-apply-perm>
    </div>
</template>
<script>
    /* eslint-disable camelcase */
    import { bus } from '@open/common/bus'
    import { getProjectByCode } from '@open/common/util'
    import Img403 from '@/images/403.png'
    import BcsUnregistry from '@open/components/bcs-unregistry/unregistry.vue'

    export default {
        name: 'app',
        components: {
            BcsUnregistry
        },
        data () {
            return {
                routerKey: +new Date(),
                systemCls: 'mac',
                minHeight: 768,
                isUserBKService: true,
                curProject: null,
                isIEGProject: true,
                height: 0,
                ccKey: '',
                ccList: [],
                kind: 1, // 业务编排类型
                // 前一次选中的编排类型，用于选中 tke 请求失败后，单选框恢复到上一个状态
                prevKind: 1,
                enableBtn: false, // 提交按钮是否可用
                projectId: '',
                projectCode: '',
                isLoading: true
            }
        },
        computed: {
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            }
        },
        watch: {
            '$route' (to, from) {
                this.checkProject()
                this.initProjectId()
                if (window.$syncUrl) {
                    const path = this.$route.fullPath.replace(new RegExp(`^${SITE_URL}`), '')
                    window.$syncUrl(path)
                }
            },
            ccKey (val) {
                this.enableBtn = val !== null && val !== undefined
            },
            kind (v, old) {
                this.prevKind = old
                this.fetchCCList()
            }
        },
        async created () {
            const platform = window.navigator.platform.toLowerCase()
            if (platform.indexOf('win') === 0) {
                this.systemCls = 'win'
            }

            if (this.$store.state.isEn) {
                this.systemCls += ' english'
            }

            window.addEventListener('change::$currentProjectId', async e => {
                await this.initBcsBaseData(e.detail.currentProjectId)
            })
        },
        mounted () {
            document.title = this.$t('容器服务')

            this.initContainerSize()
            window.onresize = () => {
                this.initContainerSize()
                this.height = window.innerHeight
            }

            this.height = window.innerHeight

            const self = this
            bus.$on('show-apply-perm-modal', data => {
                const projectCode = self.$route.params.projectCode
                self.$refs.bkApplyPerm && self.$refs.bkApplyPerm.show(projectCode, data)
            })
            bus.$on('show-error-message', data => {
                self.$bkMessage({
                    theme: 'error',
                    message: data
                })
            })
            bus.$on('show-apply-perm', data => {
                const projectCode = self.$route.params.projectCode
                const content = ''
                    + '<div class="biz-top-bar">'
                    + '<div class="biz-back-btn" onclick="history.back()">'
                    + '<i class="bcs-icon bcs-icon-arrows-left back"></i>'
                    + '<span></span>'
                    + '</div>'
                    + '</div>'
                    + '<div class="bk-exception bk-exception-center">'
                    + `<img src="${Img403}"/>`
                    + '<h2 class="exception-text">'
                    + `<p class="f14">${self.$t('Sorry，您的权限不足，请去')}`
                    + `<a class="bk-text-button" href="${data.apply_url}&project_code=${projectCode}" target="_blank">${self.$t('申请')}</a>`
                    + '</p>'
                    + '</h2>'
                    + '</div>'

                document.querySelector('.biz-content').innerHTML = content
            })
            bus.$on('close-apply-perm-modal', data => {
                self.$refs.bkApplyPerm && self.$refs.bkApplyPerm.hide()
            })
        },
        methods: {
            // 设置单集群信息
            handleSetClusterInfo (clusterId = '') {
                // 判断集群ID是否存在当前项目的集群列表中
                const stateClusterList = this.$store.state.cluster.clusterList || []
                const curCluster = stateClusterList?.find(cluster => cluster.cluster_id === clusterId)
                if (!curCluster) {
                    clusterId = ''
                }
                localStorage.setItem('bcs-cluster', clusterId)
                sessionStorage.setItem('bcs-cluster', clusterId)
                this.$store.commit('updateCurClusterId', clusterId)
                this.$store.commit('cluster/forceUpdateCurCluster', curCluster || {})

                if (this.$route.params.clusterId && !curCluster) {
                    // url路径中存在集群ID，但是该集群ID不在集群列表中时跳转首页
                    this.$router.replace({
                        name: 'clusterMain',
                        params: {
                            needCheckPermission: true
                        }
                    })
                } else if (this.$route.name === 'clusterMain' && clusterId) {
                    // 集群ID存在，但是当前处于全部集群首页时需要跳回集群概览页
                    this.$router.replace({
                        name: 'clusterOverview',
                        params: {
                            projectId: this.$store.state.curProjectId,
                            projectCode: this.$store.state.curProjectCode,
                            clusterId,
                            needCheckPermission: true
                        }
                    })
                }
            },
            // 设置项目信息
            handleSetProjectInfo (projectCode, projectId) {
                localStorage.setItem('curProjectCode', projectCode)
                localStorage.setItem('curProjectId', projectId)
                this.$store.commit('updateProjectCode', projectCode)
                this.$store.commit('updateProjectId', projectId)
            },
            // 初始化BCS基本数据（有先后顺序，请勿乱动）
            async initBcsBaseData (projectCode) {
                this.isLoading = true
                this.initProjectId(projectCode)
                // 清空集群列表
                this.$store.commit('cluster/forceUpdateClusterList', [])
                // 切换不同项目时清空单集群信息
                if (localStorage.getItem('curProjectCode') !== projectCode) {
                    localStorage.removeItem('bcs-cluster')
                    sessionStorage.removeItem('bcs-cluster')
                    this.$store.commit('updateCurClusterId', '')
                    localStorage.setItem('curProjectCode', projectCode)
                    window.location.href = `${window.location.origin}${SITE_URL}/${projectCode}`
                }
                const projectList = await this.$store.dispatch('getProjectList').catch(() => ([]))
                // 检查是否开启容器服务
                await this.checkProject()
                if (!this.isUserBKService) {
                    this.isLoading = false
                    return
                }

                const curBcsProject = projectList.find(item => item.project_code === projectCode)
                if (curBcsProject?.project_id) {
                    await this.$store.dispatch('cluster/getClusterList', curBcsProject.project_id)
                }

                // 设置项目存储信息
                this.handleSetProjectInfo(projectCode, curBcsProject?.project_id || '')

                // 设置当前集群ID
                let curClusterId = ''
                const pathClusterId = this.$route.params.clusterId
                const storageClusterId = localStorage.getItem('bcs-cluster') || ''
                if (pathClusterId && (this.$route.path.indexOf('dashboard') > -1 || storageClusterId)) {
                    // 资源视图或者以前切换过单集群就以url上面的集群ID为主
                    curClusterId = pathClusterId
                } else {
                    curClusterId = storageClusterId
                }

                this.handleSetClusterInfo(curClusterId)
                // 获取当前视图类型
                this.$store.commit('updateViewMode', this.$route.path.indexOf('dashboard') > -1 ? 'dashboard' : 'cluster')
                // 获取菜单配置信息
                await this.$store.dispatch('getFeatureFlag')
                // 更新菜单
                this.$store.commit('updateCurProject', projectCode)
                this.isLoading = false
            },
            /**
             * 初始化容器最小高度
             */
            initContainerSize () {
                const WIN_MIN_HEIGHT = 768
                const APP_FOOTER_HEIGHT = 210
                const winHeight = window.innerHeight
                if (winHeight <= WIN_MIN_HEIGHT) {
                    this.minHeight = WIN_MIN_HEIGHT
                } else {
                    this.minHeight = winHeight - APP_FOOTER_HEIGHT
                }
            },

            /**
             * 检测项目是否是 IEG 项目，是否使用了蓝鲸服务
             */
            async checkProject () {
                const projectCode = window.$currentProjectId
                if (projectCode && this.onlineProjectList.length) {
                    for (const project of this.onlineProjectList) {
                        if (project.project_code === projectCode) {
                            this.curProject = project
                            this.isUserBKService = project.kind !== 0
                            if (!this.isUserBKService) {
                                this.checkUser()
                                await this.fetchCCList()
                                if (project.cc_app_id !== 0) {
                                    this.ccKey = project.cc_app_id
                                }
                            }
                        }
                    }
                }
            },

            async checkUser () {
                try {
                    const res = await this.$store.dispatch('getUserBgInfo')
                    if (!res.data.is_ieg) {
                        this.$bkInfo({
                            clsName: 'not-ieg-user-infobox',
                            type: 'default',
                            quickClose: false,
                            title: this.$t('非IEG用户请使用对应BG的容器服务平台')
                        })
                        return
                    }
                } catch (e) {
                    console.log(e)
                }
            },

            /**
             * 初始化时，将通过 projectCode 值获取 projectId 并存储在路由中
             */
            initProjectId (projectCode = window.$currentProjectId || this.$route.params.projectCode) {
                if (window.$currentProjectId) {
                    this.projectCode = window.$currentProjectId
                    this.$route.params.projectCode = this.projectCode
                    const project = getProjectByCode(this.projectCode)
                    const projectId = project.project_id
                    if (projectId) {
                        this.$route.params.projectId = projectId
                        this.projectId = projectId
                    }
                }
            },

            /**
             * 获取关联 CC 的数据
             */
            async fetchCCList () {
                try {
                    const res = await this.$store.dispatch('getCCList', {
                        project_kind: this.kind,
                        project_id: this.curProject.project_id
                    })
                    this.ccList = [...(res.data || [])]
                } catch (e) {
                    this.kind = this.prevKind
                }
            },

            /**
             * 启用容器服务 更新项目
             */
            async updateProject () {
                try {
                    this.isLoading = true
                    await this.$store.dispatch('editProject', Object.assign({}, this.curProject, {
                        // deploy_type 值固定，就是原来页面上的：部署类型：容器部署
                        deploy_type: [2],
                        // kind 业务编排类型
                        kind: parseInt(this.kind, 10),
                        // use_bk 值固定，就是原来页面上的：使用蓝鲸部署服务
                        use_bk: true,
                        cc_app_id: this.ccKey
                    }))

                    // await this.$store.dispatch('getProjectList')

                    this.$nextTick(() => {
                        window.location.reload()
                        // 这里不需要设置 isLoading 为 false，页面刷新后，isLoading 的值会重置为 true
                        // 如果设置了后，页面会闪烁一下
                        // this.isLoading = false
                    })
                } catch (e) {
                    console.error(e)
                    this.isLoading = false
                }
            },

            /**
             * 改变 routerKey，刷新 router
             */
            reloadCurPage () {
                this.routerKey = +new Date()
            },
            handleKindChange (kind) {
                this.kind = kind
            },
            handleCmdbChange (ccKey) {
                this.ccKey = ccKey
            }
        }
    }
</script>
<style lang="postcss">
    @import '@/css/reset.css';
    @import '@/css/app.css';
    @import '@/css/animation.css';

    .app-container {
        min-width: 1280px;
        min-height: 768px;
        position: relative;
        display: flex;
        background: #fafbfd;
        min-height: 100% !important;
        padding-top: 0;
    }
    .app-content {
        flex: 1;
        background: #fafbfd;
    }
    .biz-guide-box {
        .desc {
            width: auto;
            margin: 0 auto 25px;
            position: relative;
            top: 12px;
        }
        .biz-app-form {
            .form-item {
                .form-item-inner {
                    width: 340px;
                    .bk-form-radio {
                        width: 115px;
                    }
                }
            }
        }
    }
    .biz-list-operation {
        .item {
            float: none;
        }
    }

    .not-ieg-user-infobox {
        .bk-dialog-style {
            width: 500px;
        }
    }
</style>
