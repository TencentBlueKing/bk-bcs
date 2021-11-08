<template>
    <div class="biz-container app-container" v-bkloading="{ isLoading, zIndex: 10 }">
        <template v-if="isUserBKService">
            <!-- isLoading为解决当前集群信息未设置时时序问题 -->
            <template v-if="!isLoading">
                <SideNav class="biz-side-bar"></SideNav>
                <!-- $route.path为解决应用模块动态组件没有刷新问题 -->
                <router-view :key="$route.path" />
                <!-- 终端 -->
                <SideTerminal></SideTerminal>
            </template>
        </template>
        <template v-else>
            <Unregistry></Unregistry>
        </template>
    </div>
</template>
<script lang="ts">
    import { computed, defineComponent, onBeforeMount, ref } from '@vue/composition-api'
    import SideNav from './side-nav.vue'
    import SideTerminal from '@/components/terminal/index.vue'
    import Unregistry from '@/views/unregistry.vue'

    export default defineComponent({
        name: 'home',
        components: {
            SideNav,
            SideTerminal,
            Unregistry
        },
        setup (props, ctx) {
            // 项目和集群的清空已经赋值操作有时序关系，请勿随意调整顺序
            const { $store, $route, $router, $bkMessage } = ctx.root
            const handleSetClusterStorageInfo = (curCluster?) => {
                if (curCluster) {
                    localStorage.setItem('bcs-cluster', curCluster.cluster_id)
                    sessionStorage.setItem('bcs-cluster', curCluster.cluster_id)
                    $store.commit('updateCurClusterId', curCluster.cluster_id)
                    $store.commit('cluster/forceUpdateCurCluster', curCluster)
                } else {
                    localStorage.removeItem('bcs-cluster')
                    sessionStorage.removeItem('bcs-cluster')
                    $store.commit('updateCurClusterId', '')
                    $store.commit('cluster/forceUpdateCurCluster', {})
                }
            }
            const projectList = computed(() => {
                return $store.state.sideMenu.onlineProjectList
            })
            const projectCode = $route.params.projectCode
            const curProject = projectList.value.find(item => item.project_code === projectCode)
            // 项目不存在时
            if (!curProject) {
                if (window.REGION === 'ieod') {
                    // 返回集群首页
                    const [firstProject = {}] = projectList.value
                    $router.push({
                        name: 'clusterMain',
                        params: {
                            projectId: firstProject.project_id,
                            projectCode: firstProject.project_code
                        }
                    })
                } else {
                    // 私有化版本返回项目管理页
                    $router.replace({ name: 'projectManage' })
                }
                return
            }

            if (localStorage.getItem('curProjectCode') !== projectCode) {
                // 切换不同项目时清空单集群信息
                handleSetClusterStorageInfo()
                const preProject = projectList.value.find(item => item.project_code === localStorage.getItem('curProjectCode'))
                if (curProject?.kind !== preProject?.kind) {
                    // 切换不同项目类型时重刷界面
                    window.location.reload()
                }
            }

            // 缓存当前项目信息
            localStorage.setItem('curProjectCode', projectCode)
            localStorage.setItem('curProjectId', curProject.project_id)
            $store.commit('updateProjectCode', projectCode)
            $store.commit('updateProjectId', curProject.project_id)
            $store.commit('updateCurProject', curProject)
            // 设置路由projectId和projectCode信息（旧模块很多地方用到），后续路由切换时也会在全局导航钩子上注入这个两个参数
            $route.params.projectId = curProject.project_id
            $route.params.projectCode = projectCode

            // 清空上一个项目的集群列表
            $store.commit('cluster/forceUpdateClusterList', [])

            // 设置当前视图类型（集群管理 or 资源视图）
            $store.commit('updateViewMode', $route.meta?.isDashboard ? 'dashboard' : 'cluster')

            // 项目未开启容器服务跳转未注册界面
            const isUserBKService = ref(true)
            if (curProject.kind === 0) {
                isUserBKService.value = false
                return
            }

            const isLoading = ref(false)
            onBeforeMount(async () => {
                // 获取项目的集群列表和菜单配置信息
                isLoading.value = true
                // 获取当前项目详情信息
                $store.dispatch('getProject', { projectId: curProject.project_id }).then((res) => {
                    $store.commit('updateCurProject', res.data)
                })
                await $store.dispatch('cluster/getClusterList', curProject.project_id).catch(err => {
                    $bkMessage({
                        theme: 'error',
                        message: err
                    })
                })

                // 设置单集群ID（1. 路由上有单集群参数但是为全部集群，刷新时还是全部集群 2. 路由上有单集群信息但是为单集群，刷新后为当前路由的单集群 3. 资源视图始终是单集群）
                let curClusterId = ''
                const pathClusterId = $route.params.clusterId
                const storageClusterId = localStorage.getItem('bcs-cluster') || ''
                if (pathClusterId && ($store.state.viewMode === 'dashboard' || storageClusterId)) { // 资源视图或者以前切换过单集群就以url上面的集群ID为主
                    curClusterId = pathClusterId
                } else {
                    curClusterId = storageClusterId
                }
                // 判断集群ID是否存在当前项目的集群列表中
                const stateClusterList = $store.state.cluster.clusterList || []
                const curCluster = stateClusterList?.find(cluster => cluster.cluster_id === curClusterId)
                if (curCluster) {
                    // 缓存单集群信息
                    handleSetClusterStorageInfo(curCluster)
                    $route.params.clusterId = curClusterId
                } else {
                    handleSetClusterStorageInfo()
                }

                if ($route.name !== 'clusterMain' && $route.params.clusterId && !curCluster) {
                    // path路径中存在集群ID，但是该集群ID不在集群列表中时跳转首页
                    $router.replace({
                        name: 'clusterMain'
                    })
                } else if ($route.name === 'clusterMain' && curCluster) {
                    // 集群ID存在，但是当前处于全部集群首页时需要跳回单集群概览页
                    $router.replace({
                        name: 'clusterOverview'
                    })
                }

                await $store.dispatch('getFeatureFlag').catch(err => {
                    $bkMessage({
                        theme: 'error',
                        message: err
                    })
                })
                isLoading.value = false
            })

            return {
                isLoading,
                isUserBKService
            }
        }
    })
</script>
