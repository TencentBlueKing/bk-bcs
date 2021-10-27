<template>
    <app-auth ref="bkAuth"></app-auth>
</template>

<script>
    import { bus } from '@open/common/bus'
    import { getProjectById } from '@open/common/util'

    export default {
        data () {
            return {}
        },
        computed: {
            onlineProjectList () {
                return this.$store.state.sideMenu.onlineProjectList
            },
            curProjectCode () {
                return this.$store.state.curProjectCode
            },
            projectCode () {
                const route = this.$route
                // 从路由获取 projectCode
                if (route.params.projectCode) {
                    // this.setLocalStorage(route.params.projectCode)
                    return route.params.projectCode
                }

                // 从缓存获取projectId
                if (this.curProjectCode) {
                    for (const item of this.onlineProjectList) {
                        if (item.project_code === this.curProjectCode) {
                            return this.curProjectCode
                        }
                    }
                }

                // 直接显示第一个项目
                if (this.onlineProjectList.length) {
                    return this.onlineProjectList[0].project_code
                }

                return ''
            },
            projectId () {
                const route = this.$route
                // 从路由获取projectId
                if (route.params.projectId) {
                    // this.setLocalStorage(route.params.projectId)
                    return route.params.projectId
                }

                // 从缓存获取projectId
                if (localStorage.curProjectId) {
                    const projectId = localStorage.curProjectId
                    for (const item of this.onlineProjectList) {
                        if (item.project_id === projectId) {
                            return projectId
                        }
                    }
                }

                // 直接显示第一个项目
                if (this.onlineProjectList.length) {
                    const projectId = this.onlineProjectList[0].project_id
                    // this.setLocalStorage(projectId)
                    return projectId
                }

                return ''
            },
            parentRouteName () {
                const bcsRouteKeys = [
                    'containerServiceMain',
                    'clusterMain',
                    'clusterCreate',
                    'clusterOverview',
                    'clusterInfo',
                    'clusterNodeOverview',
                    'containerDetailForNode',
                    'deployments',
                    'deploymentsInstanceDetail',
                    'deploymentsInstanceDetail2',
                    'deploymentsContainerDetail',
                    'deploymentsContainerDetail2',
                    'deploymentsInstantiation',
                    'daemonset',
                    'daemonsetInstanceDetail',
                    'daemonsetInstanceDetail2',
                    'daemonsetContainerDetail',
                    'daemonsetContainerDetail2',
                    'daemonsetInstantiation',
                    'job',
                    'jobInstanceDetail',
                    'jobInstanceDetail2',
                    'jobContainerDetail',
                    'jobContainerDetail2',
                    'jobInstantiation',
                    'statefulset',
                    'statefulsetInstanceDetail',
                    'statefulsetInstanceDetail2',
                    'statefulsetContainerDetail',
                    'statefulsetContainerDetail2',
                    'statefulsetInstantiation',

                    'service',
                    'loadBalance',
                    'loadBalanceDetail',
                    'resourceMain',
                    'resourceConfigmap',
                    'resourceSecret',
                    'depotMain',
                    'imageLibrary',
                    'projectImage',
                    'clusterNode',
                    'nodeMain',
                    'myCollect',
                    'mcMain',
                    'operateAudit',
                    'eventQuery',
                    'configurationMain',
                    'namespace',
                    'templateset',
                    'configurationCreate',
                    'k8sTemplatesetApplication',
                    'k8sTemplatesetDeployment',
                    'k8sTemplatesetService',
                    'k8sTemplatesetConfigmap',
                    'k8sTemplatesetSecret',
                    'k8sTemplatesetIngress',
                    'k8sTemplatesetHPA',
                    'instantiation',
                    'metricManage'
                ]

                const routeName = this.$route.name

                let parentRouteName = ''

                if (bcsRouteKeys.includes(routeName)) {
                    parentRouteName = 'clusterMain'
                    document.title = this.$t('容器服务')
                }
                return parentRouteName
            },
            curViewType () {
                return this.$route.path.indexOf('dashboard') > -1 ? 'dashboard' : 'cluster'
            }
        },
        created () {
            // 点击导航模块名称时，会触发返回当前模块首页事件，由iframe内部进行返回首页的跳转
            window.addEventListener('order::backHome', () => {
                this.reloadPage(this.parentRouteName)
            })
        },
        mounted () {
            const self = this
            bus.$on('show-login-modal', data => {
                self.$refs.bkAuth && self.$refs.bkAuth.showLoginModal(data)
            })
            bus.$on('close-login-modal', () => {
                self.$refs.bkAuth && self.$refs.bkAuth.hideLoginModal()
            })
        },
        methods: {
            /**
             * 保存 projectId 和 projectCode 到本地存储中
             *
             * @param {string} projectId 项目 id
             */
            // setLocalStorage (projectId) {
            //     const project = getProjectById(projectId)
            //     const projectCode = project.project_code
            //     localStorage.setItem('curProjectId', projectId)
            //     localStorage.setItem('curProjectCode', projectCode)
            // },

            /**
             * 刷新当前页
             *
             * @param {string} routeName 当前路由名称
             */
            reloadPage (routeName) {
                const projectId = this.projectId
                const projectCode = this.projectCode || getProjectById(projectId).project_code
                const curRouteName = this.$route.name
                if (routeName === curRouteName) {
                    this.$emit('reloadCurPage')
                } else {
                    this.$router.push({
                        name: routeName,
                        params: {
                            projectId: projectId,
                            projectCode: projectCode,
                            needCheckPermission: true
                        }
                    })
                }
            }

            /**
             * 选中项目
             *
             * @param {string} projectCode 项目 code
             */
            // selectProject (projectCode) {
            //     const routeName = this.$route.name
            //     if (!routeName) {
            //         return false
            //     }
            //     // console.error('selectProject', projectCode)
            //     const projectId = getProjectByCode(projectCode).project_id

            //     this.setLocalStorage(projectId)

            //     // 切换项目统一路由返回集群列表页面
            //     const routerUrl = this.$router.resolve({ name: 'clusterMain' })
            //     // 通过path跳转，规避导航刚好处于cluster内页
            //     this.$router.push({
            //         path: routerUrl.href,
            //         params: {
            //             projectCode: projectCode,
            //             projectId: projectId,
            //             needCheckPermission: true
            //         },
            //         query: {}
            //     })
            // }
        }
    }
</script>
