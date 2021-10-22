<template>
    <div>
        <div class="biz-side-title cluster-selector">
            <!-- 全部集群 -->
            <template v-if="!curClusterInfo.cluster_id">
                <img src="@/images/bcs2.svg" class="all-icon">
                <span class="cluster-name-all">{{$t('全部集群')}}</span>
            </template>
            <!-- 单集群 -->
            <template v-else-if="curClusterInfo.cluster_id && curClusterInfo.name">
                <span class="icon">{{ curClusterInfo.name[0] }}</span>
                <span>
                    <span class="cluster-name" :title="curClusterInfo.name">{{ curClusterInfo.name }}</span>
                    <br>
                    <span class="cluster-id">{{ curClusterInfo.cluster_id }}</span>
                </span>
            </template>
            <!-- 异常情况 -->
            <template v-else>
                <img src="@/images/bcs2.svg" class="all-icon">
                <span class="cluster-name-all">{{$t('容器服务')}}</span>
            </template>
            <i class="biz-conf-btn bcs-icon bcs-icon-qiehuan f12" @click.stop="showClusterSelector"></i>
            <img v-if="featureCluster" class="dot" src="@/images/new.svg" />
            <cluster-selector v-model="isShowClusterSelector" @change="handleChangeCluster" />
        </div>
        <div class="resouce-toggle" v-if="curClusterInfo.cluster_id || curViewType === 'dashboard'">
            <span v-for="item in viewList"
                :key="item.id"
                :class="['tab bcs-ellipsis', { active: curViewType === item.id }]"
                @click="handleChangeView(item)">
                {{item.name}}
            </span>
        </div>
        <div class="side-nav">
            <bk-menu :list="menuList" :menu-change-handler="menuSelected"></bk-menu>
            <p class="biz-copyright">Copyright © 2012-{{curYear}} Tencent BlueKing. All Rights Reserved</p>
        </div>

        <bk-dialog
            :width="500"
            :title="projectConfDialog.title"
            :quick-close="false"
            :is-show.sync="projectConfDialog.isShow"
            :has-footer="!isHasCluster"
            @cancel="projectConfDialog.isShow = false">
            <template slot="content">
                <form class="bk-form mb30">
                    <div class="bk-form-item">
                        <label class="bk-label" style="width:160px;">{{$t('英文缩写')}}：</label>
                        <div class="bk-form-content" style="margin-left:160px;">
                            <span style="line-height: 34px;">{{englishName}}</span>
                        </div>
                    </div>
                    <div class="bk-form-item is-required">
                        <label class="bk-label" style="width:160px;">{{$t('编排类型')}}：</label>
                        <div class="bk-form-content" style="margin-left:160px;">
                            <bk-radio :checked="kind === 1" disabled>K8S</bk-radio>
                        </div>
                    </div>

                    <div class="bk-form-item is-required">
                        <label class="bk-label" style="width:160px;">{{$t('关联CMDB业务')}}：</label>
                        <div class="bk-form-content" style="margin-left:160px;">
                            <div style="display: inline-block;" class="mr5">
                                <template v-if="ccList.length && !isHasCluster">
                                    <bk-selector
                                        style="width: 250px;"
                                        :placeholder="$t('请选择')"
                                        :searchable="true"
                                        :setting-key="'id'"
                                        :display-key="'name'"
                                        :selected.sync="ccKey"
                                        :list="ccList"
                                        :disabled="!canFormEdit">
                                    </bk-selector>
                                </template>
                                <template v-else>
                                    <bkbcs-input disabled v-model="curProject.cc_app_name" style="width: 250px;"></bkbcs-input>
                                </template>
                            </div>
                            <bcs-popover placement="top" :content="$t('关联业务后，您可以从对应的业务下选择机器，搭建容器集群')">
                                <span style="font-size: 13px;cursor: pointer;">
                                    <i class="bcs-icon bcs-icon-info-circle"></i>
                                </span>
                            </bcs-popover>
                            <template v-if="!canEdit && !isHasCluster">
                                <p class="desc mt15" style="text-align: left;">{{$t('当前账号没有管理员权限，不可编辑，')}}<br />{{$t('请')}}<a :href="bkAppHost" target="_blank" class="bk-text-button">{{$t('点击申请权限')}}</a></p>
                            </template>
                            <template v-else-if="!ccList.length && !isHasCluster">
                                <p class="desc mt15" style="text-align: left;">{{$t('当前账号在蓝鲸配置平台无业务，请联系运维在蓝鲸配置平台关联业务，')}}<a :href="bkCCHost" target="_blank" class="bk-text-button">{{$t('点击查看业务和运维信息')}}</a></p>
                            </template>
                        </div>
                    </div>
                    <div class="bk-form-item" v-if="isHasCluster">
                        <label class="bk-label" style="width:150px;"></label>
                        <div class="bk-form-content" style="margin-left: 150px; width: 260px;">
                            {{$t('该项目下已有集群信息，如需更改编排类型和绑定业务信息，请先删除已有集群')}}
                        </div>
                    </div>
                </form>
            </template>
            <div slot="footer">
                <div class="biz-footer">
                    <template v-if="!canEdit">
                        <bcs-popover :content="$t('没有管理员权限')" placement="top">
                            <bk-button type="primary" :disabled="true">{{$t('保存')}}</bk-button>
                        </bcs-popover>
                    </template>
                    <template v-else-if="!ccList.length">
                        <bcs-popover :content="$t('请选择要关联的CMDB业务')" placement="top">
                            <bk-button type="primary" :disabled="true">{{$t('保存')}}</bk-button>
                        </bcs-popover>
                    </template>
                    <template v-else>
                        <bk-button type="primary" :disabled="!canEdit || !ccList.length" @click="updateProject" :loading="isLoading">{{$t('保存')}}</bk-button>
                    </template>
                    <bk-button type="default" @click="projectConfDialog.isShow = false">{{$t('取消')}}</bk-button>
                </div>
            </div>
        </bk-dialog>
    </div>
</template>

<script>
    import bkMenu from '@open/components/menu'
    import clusterSelector from '@open/components/cluster-selector'
    import { bus } from '@open/common/bus'
    import { catchErrorHandler } from '@open/common/util'

    export default {
        name: 'side-nav',
        components: {
            bkMenu,
            clusterSelector
        },
        data () {
            return {
                isLoading: false,
                isShowClusterSelector: false,
                projectIdTimer: null,
                projectConfDialog: {
                    isShow: false,
                    title: ''
                },
                bkCCHost: window.BK_CC_HOST + '/#/business',
                ccKey: '',
                canEdit: false,
                ccList: [],
                englishName: '',
                kind: -1, // 业务编排类型
                selectorMenuData: {
                    isChild: false,
                    item: {
                        icon: 'bcs-icon-jq-colony',
                        name: this.$t('概览'),
                        pathName: ['clusterOverview', 'clusterNode', 'clusterNodeOverview', 'clusterInfo'],
                        roleId: 'overview:menu'
                    },
                    itemIndex: 0
                },
                viewList: [
                    {
                        id: 'cluster',
                        name: this.$t('集群管理')
                    },
                    {
                        id: 'dashboard',
                        name: this.$t('资源视图')
                    }
                ],
                featureCluster: !localStorage.getItem('FEATURE_CLUSTER')
            }
        },
        computed: {
            curClusterInfo () {
                return this.$store.state.cluster.curCluster || {}
            },
            bkAppHost () {
                // bkApplyHost: window.BK_IAM_APP_URL,
                if (window.BK_IAM_APP_URL) {
                    return window.BK_IAM_APP_URL
                }
                return `${window.DEVOPS_HOST}/console/perm/apply-join-project?project_code=${this.projectCode}`
            },
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            },
            curClusterId () {
                return this.$store.state.curClusterId
            },
            curProject () {
                const project = this.$store.state.curProject
                if (!project.cc_app_name) {
                    project.cc_app_name = ''
                }
                return project
            },
            clusterList () {
                return this.$store.state.cluster.clusterList
            },
            menuList () {
                if (this.$route.meta.isDashboard) {
                    return this.$store.state.sideMenu.dashboardMenuList
                }
                if (this.curProject && this.curClusterInfo.cluster_id) {
                    return this.$store.state.sideMenu.clusterk8sMenuList
                }
                return this.$store.state.sideMenu.k8sMenuList
            },
            curYear () {
                return (new Date()).getFullYear()
            },
            isHasCluster () {
                return this.clusterList.length > 0
            },
            isHasCC () {
                return this.ccList.length > 0
            },
            canFormEdit () {
                // 有权限并且没有集群
                return this.canEdit && !this.isHasCluster
            },
            curViewType () {
                return this.$route.path.indexOf('dashboard') > -1 ? 'dashboard' : 'cluster'
            }
        },
        watch: {
            '$route' (to, from) {
                if (!['imageDetail'].includes(to.name)) {
                    this.$store.dispatch('updateMenuListSelected', {
                        isDashboard: this.$route.meta.isDashboard,
                        pathName: to.name,
                        kind: to.query.kind,
                        idx: 'bcs'
                    })
                }
            },
            'projectId' () {
                this.$store.dispatch('updateMenuListSelected', {
                    isDashboard: this.$route.meta.isDashboard,
                    pathName: this.$route.name,
                    kind: this.$route.query.kind,
                    idx: 'bcs'
                })
            },
            curClusterId () {
                this.$store.dispatch('updateMenuListSelected', {
                    isDashboard: this.$route.meta.isDashboard,
                    pathName: this.$route.name,
                    kind: this.$route.query.kind,
                    idx: 'bcs'
                })
            }
        },
        async created () {
            this.$store.dispatch('updateMenuListSelected', {
                isDashboard: this.$route.meta.isDashboard,
                pathName: this.$route.name,
                idx: 'bcs',
                kind: this.$route.query.kind,
                projectType: (this.curProject && (this.curProject.kind === PROJECT_K8S || this.curProject.kind === PROJECT_TKE)) ? 'k8s' : ''
            })
            if (window.bus) {
                window.bus.$on('showProjectConfDialog', () => {
                    this.showProjectConfDialog()
                })
            }
            bus.$off('cluster-change')
            bus.$on('cluster-change', (data) => {
                this.menuSelected(data)
            })

            await this.getProject()
        },
        methods: {
            /**
             * 显示项目配置窗口
             */
            async showProjectConfDialog () {
                await this.fetchCCList()
                this.ccKey = this.curProject.cc_app_id
                this.englishName = this.curProject.english_name
                this.projectConfDialog.isShow = true
                this.projectConfDialog.title = `${this.$t('项目')}【${this.curProject.project_name}】`
            },

            /**
             * 左侧导航 menu 选择事件
             *
             * @param {Object} data menu 数据
             */
            menuSelected (data) {
                const curSelected = data.child || data.item
                const projectCode = this.projectCode
                if (!curSelected.pathName) {
                    return false
                }
                if (curSelected.externalLink) {
                    const url = `${DEVOPS_HOST}${curSelected.externalLink}${projectCode}/?project_id=${this.projectId}`
                    window.top.location.href = url
                } else {
                    this.$router.push({
                        name: curSelected.pathName[0],
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode,
                            clusterId: this.curClusterInfo.cluster_id,
                            needCheckPermission: true
                        }
                    })
                }
                return this.$store.state.allowRouterChange
            },

            /**
             * 获取关联 CC 的数据
             */
            async fetchCCList () {
                try {
                    const res = await this.$store.dispatch('getCCList', {
                        project_kind: this.kind,
                        project_id: this.projectId
                    })
                    this.ccList = [...(res.data || [])]

                    if (this.curProject.cc_app_id) {
                        const curCCItem = this.ccList.find(item => {
                            return String(item.id) === String(this.curProject.cc_app_id)
                        })

                        // 判断当前类型cc列表是否含有当前项目业务，如果没有则加入
                        if (!curCCItem && String(this.kind) === String(this.curProject.kind)) {
                            this.ccList.unshift({
                                id: this.curProject.cc_app_id,
                                name: this.curProject.cc_app_name
                            })
                        }
                    }
                } catch (e) {
                }
            },

            /**
             * 获取当前项目数据
             */
            async getProject () {
                try {
                    const res = await this.$store.dispatch('getProject', { projectId: this.projectId })
                    this.curProject.cc_app_id = res.data.cc_app_id
                    this.curProject.cc_app_name = res.data.cc_app_name
                    this.curProject.kind = res.data.kind
                    this.kind = res.data.kind
                    this.canEdit = res.data.can_edit
                } catch (e) {
                }
            },

            /**
             * 更新项目信息
             */
            async updateProject () {
                if (this.isLoading) return

                if (!this.ccKey) {
                    this.$bkMessage({
                        theme: 'error',
                        message: this.$t('请关联CMDB业务')
                    })
                    return false
                }

                try {
                    this.isLoading = true
                    await this.$store.dispatch('editProject', Object.assign({}, this.curProject, {
                        // deploy_type 值固定，就是原来页面上的：部署类型：容器部署
                        deploy_type: [2],
                        // kind 业务编排类型：1 Kubernetes
                        kind: parseInt(this.kind, 10),
                        // use_bk 值固定，就是原来页面上的：使用蓝鲸部署服务
                        use_bk: true,
                        cc_app_id: this.ccKey
                    }))

                    this.projectConfDialog.isShow = false
                    this.$bkMessage({
                        theme: 'success',
                        message: this.$t('更新成功')
                    })

                    setTimeout(() => {
                        // 当前窗口不是顶层窗口，存在 iframe
                        if (window.self !== window.top) {
                            // 父窗口就是顶层窗口，只有一层 iframe 嵌套
                            if (window.top === window.parent) {
                                // 导航项目切换时，iframe依然保存上个项目的路由信息
                                const matchs = this.$route.path.match(/\/bcs\/(\w+)\//)
                                let url = this.$route.fullPath
                                if (matchs && matchs.length > 1) {
                                    const projectCode = matchs[1]
                                    url = url.replace(projectCode, this.projectCode)
                                }
                                window.$syncUrl(url.replace(new RegExp(`^${SITE_URL}`), ''), true)
                            } else { // 父窗口不是顶层窗口，存在多层 iframe 嵌套
                                window.location.reload()
                            }
                        }
                    }, 200)
                } catch (e) {
                    catchErrorHandler(e, this)
                } finally {
                    this.isLoading = false
                }
            },

            /**
             * 显示集群切换选择器弹窗
             */
            showClusterSelector () {
                this.isShowClusterSelector = true
            },

            /**
             * 集群切换
             */
            handleChangeCluster (cluster) {
                localStorage.setItem('FEATURE_CLUSTER', true)
                if (!cluster.cluster_id) {
                    sessionStorage.removeItem('bcs-selected-menu-data')
                    if (this.$route.meta.isDashboard) {
                        this.$router.push({
                            name: 'dashboard',
                            params: {
                                projectId: this.projectId,
                                projectCode: this.projectCode
                            }
                        })
                    } else {
                        this.$router.push({
                            name: 'clusterMain',
                            params: {
                                needCheckPermission: true
                            }
                        })
                    }
                } else {
                    const data = this.$route.meta.isDashboard
                        ? {
                            isChild: true,
                            item: {
                                name: this.$t('命名空间'),
                                isSaveData: true,
                                icon: "bcs-icon-namespace",
                                roleId: "workload:menu",
                                pathName: ["dashboardNamespace"],
                                isSelected: true,
                                isOpen: false
                            },
                            itemIndex: 0
                        }
                        : {
                            isChild: false,
                            item: {
                                icon: 'bcs-icon-jq-colony',
                                name: this.$t('概览'),
                                pathName: ['clusterOverview', 'clusterNode', 'clusterInfo'],
                                roleId: 'overview:menu'
                            },
                            itemIndex: 0
                        }
                    bus.$emit('cluster-change', data)
                }
            },

            handleChangeView (item) {
                if (item.id === this.curViewType) return
                this.$store.commit('updateViewMode', item.id)
                item.id === 'dashboard' ? this.goDashboard() : this.goCluster()
            },

            goDashboard () {
                // 从 metric 管理 router 跳转过来时，url 有 cluster_id 的 query
                const newQuery = JSON.parse(JSON.stringify(this.$route.query))
                delete newQuery.cluster_id
                this.$router.replace({ query: newQuery })

                setTimeout(() => {
                    const routerUrl = this.$router.resolve({
                        name: 'dashboard',
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                    window.$syncUrl(routerUrl.href.replace(new RegExp(`^${SITE_URL}`), ''), true)
                    sessionStorage.removeItem('bcs-selected-menu-data')
                }, 0)
            },

            goCluster () {
                setTimeout(() => {
                    const routerUrl = this.$router.resolve({
                        name: 'clusterMain',
                        params: {
                            projectId: this.projectId,
                            projectCode: this.projectCode
                        }
                    })
                    window.$syncUrl(routerUrl.href.replace(new RegExp(`^${SITE_URL}`), ''), true)
                    sessionStorage.removeItem('bcs-selected-menu-data')
                }, 0)
            }
        }
    }
</script>

<style scoped lang="postcss">
    .biz-side-title {
        position: relative;
    }
    .cluster-selector {
        background: #fafbfd;
    }
    .resouce-toggle {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 10px 0;
        .tab {
            display: flex;
            align-items: center;
            justify-content: center;
            background: #f7f8f9;
            border: 1px solid #dde4eb;
            margin-left: -1px;
            font-size: 12px;
            height: 24px;
            padding: 0 26px;
            cursor: pointer;
            white-space: nowrap;
            &.active {
                background: #fff;
                color: #3a84ff;
            }
            &.disabled {
                cursor: not-allowed;
            }
            &:first-child {
                border-radius: 3px 0 0 3px;
            }
            &:last-child {
                border-radius: 0 3px 3px 0;
            }
        }
    }
    .biz-conf-btn {
        position: absolute;
        right: 10px;
        top: 16px;
        font-size: 12px;
        cursor: pointer;
        width: 30px;
        height: 30px;
        text-align: center;
        line-height: 30px;
        z-index: 100;
    }
    .biz-footer {
        text-align: right;
        padding: 0 20px;
    }
    .cluster-name {
        max-width: 150px;
        overflow: hidden;
        text-overflow: ellipsis;
        display: inline-block;
        white-space: nowrap;
        margin-top: 2px;
    }
    .cluster-name-all {
        font-size: 16px;
    }
    .dot {
        position: absolute;
        display: inline-block;
        width: 16px;
        height: 16px;
        top: 16px;
        right: 4px;
        z-index: 1;
        padding: 2px;
    }
</style>
