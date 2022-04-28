<template>
    <div class="biz-container app-container" v-bkloading="{ isLoading, zIndex: 10 }">
        <!-- isLoading为解决当前集群和项目信息未设置时界面依赖时序问题 -->
        <template v-if="curProject && curProject.kind !== 0 && !isLoading">
            <SideNav class="biz-side-bar" :key="$i18n.locale"></SideNav>
            <div class="bcs-content">
                <ContentHeader
                    :title="$route.meta.title"
                    :hide-back="$route.meta.hideBack"
                    v-if="$route.meta.title"
                ></ContentHeader>
                <!-- $route.path为解决应用模块动态组件没有刷新问题 -->
                <router-view :key="`${$route.path}${$store.state.curClusterId}`" />
            </div>
            <!-- 终端 -->
            <SideTerminal></SideTerminal>
        </template>
        <template v-else-if="curProject && curProject.kind === 0">
            <Unregistry :cur-project="curProject"></Unregistry>
        </template>
        <bk-exception type="403" v-else-if="!projectList.length">
            <span>{{$t('无项目权限')}}</span>
            <div class="text-subtitle">{{$t('你没有相应项目的访问权限，请前往申请相关项目权限')}}</div>
            <a class="bk-text-button text-wrap" @click="handleGotoIAM">{{$t('去申请')}}</a>
        </bk-exception>
    </div>
</template>
<script lang="ts">
    /* eslint-disable camelcase */
    import { computed, defineComponent, onBeforeMount, ref } from '@vue/composition-api'
    import SideNav from './side-nav.vue'
    import SideTerminal from '@/components/terminal/index.vue'
    import Unregistry from '@/views/unregistry.vue'
    import ContentHeader from '@/views/content-header.vue'
    import { isProjectExit } from '@/api/base'

    export default defineComponent({
        name: 'home',
        components: {
            SideNav,
            SideTerminal,
            Unregistry,
            ContentHeader
        },
        setup (props, ctx) {
            // 项目和集群的清空已经赋值操作有时序关系，请勿随意调整顺序
            const { $store, $route, $router, $bkMessage } = ctx.root
            const handleGotoIAM = () => {
                window.open(window.BK_IAM_APP_URL)
            }
            const handleSetClusterStorageInfo = (curCluster?) => {
                if (curCluster) {
                    localStorage.setItem('bcs-cluster', curCluster.cluster_id)
                    sessionStorage.setItem('bcs-cluster', curCluster.cluster_id)
                    $store.commit('updateCurClusterId', curCluster.cluster_id)
                    $store.commit('cluster/forceUpdateCurCluster', curCluster)
                    $store.commit('cluster/forceUpdateClusterList', $store.state.cluster.allClusterList)
                    $route.params.clusterId = curCluster.cluster_id
                } else {
                    localStorage.removeItem('bcs-cluster')
                    sessionStorage.removeItem('bcs-cluster')
                    $store.commit('updateCurClusterId', '')
                    $store.commit('cluster/forceUpdateCurCluster', {})
                }
            }
            const projectList = computed(() => {
                return $store.state.sideMenu.onlineProjectList || []
            })
            const projectCode = $route.params.projectCode
            const localProjectCode = localStorage.getItem('curProjectCode')
            if (localProjectCode !== projectCode) {
                // 切换不同项目时清空单集群信息
                handleSetClusterStorageInfo()
            }
            // 获取当前项目（优先级：路由 > 本地缓存 > 取项目列表第一个）
            const curProject = ref<any>(null)
            if (projectCode) {
                curProject.value = projectList.value.find(item => item.project_code === projectCode)
            } else if (localProjectCode) {
                curProject.value = projectList.value.find(item => item.project_code === localProjectCode)
            } else if (projectList.value.length) {
                curProject.value = projectList.value[0]
            }
            // 校验项目
            const validateProject = async () => {
                // 1. 开启容器服务，且项目存在
                if (curProject.value) {
                    return curProject.value.kind !== 0
                }

                // 2. 校验项目不存在，但是ProjectCode存在的情况
                const _projectCode = projectCode || localProjectCode
                if (!_projectCode) return false

                const projectData = await isProjectExit({
                    project_code: _projectCode,
                    with_perms: false
                })
                if (projectData && projectData.length) {
                    // 项目存在，但是无权限
                    $router.push({
                        name: '403',
                        query: {
                            actionId: 'project_view',
                            resourceName: projectData[0]?.project_name,
                            permCtx: JSON.stringify({
                                project_id: projectData[0]?.project_id
                            }),
                            fromRoute: window.location.href
                        }
                    })
                } else if (projectList.value.length) {
                    // 项目不存在
                    const project = projectList.value.find(item => item.project_code === localProjectCode) || projectList.value[0]
                    const location = $router.resolve({
                        name: $route.name || 'entry',
                        params: {
                            projectCode: project.project_code
                        }
                    })
                    window.location.href = location.href
                }
                return false
            }
            // 设置项目信息
            const handleSetProjectStorageInfo = (curProject) => {
                if (!curProject) return
                // 缓存当前项目信息
                localStorage.setItem('curProjectCode', curProject.project_code)
                localStorage.setItem('curProjectId', curProject.project_id)
                $store.commit('updateProjectCode', curProject.project_code)
                $store.commit('updateProjectId', curProject.project_id)
                $store.commit('updateCurProject', curProject)
                // 设置路由projectId和projectCode信息（旧模块很多地方用到），后续路由切换时也会在全局导航钩子上注入这个两个参数
                $route.params.projectId = curProject.project_id
                $route.params.projectCode = curProject.project_code
            }
            handleSetProjectStorageInfo(curProject.value)

            // 清空上一个项目的集群列表
            $store.commit('cluster/forceUpdateClusterList', [])

            // 设置当前视图类型（集群管理 or 资源视图）
            $store.commit('updateViewMode', $route.meta?.isDashboard ? 'dashboard' : 'cluster')

            const isLoading = ref(false)
            onBeforeMount(async () => {
                isLoading.value = true
                const result = await validateProject()
                if (!result) {
                    isLoading.value = false
                    return
                }
                
                // 获取当前项目详情信息(异步获取)
                $store.dispatch('getProject', { projectId: curProject.value?.project_id }).then((res) => {
                    $store.commit('updateCurProject', res.data)
                })
                const res = await $store.dispatch('cluster/getClusterList', curProject.value?.project_id).catch(err => {
                    $bkMessage({
                        theme: 'error',
                        message: err
                    })
                    return { data: [] }
                })

                // 设置单集群ID
                // 1. 路由上有单集群参数但是为全部集群，刷新时还是全部集群
                // 2. 路由上有单集群信息但是为单集群，刷新后为当前路由的单集群
                // 3. 资源视图始终是单集群
                let curClusterId = ''
                const pathClusterId = $route.params.clusterId
                const storageClusterId = localStorage.getItem('bcs-cluster') || ''
                if (pathClusterId && ($store.state.viewMode === 'dashboard' || storageClusterId)) { // 资源视图或者以前切换过单集群就以url上面的集群ID为主
                    curClusterId = pathClusterId
                } else {
                    curClusterId = storageClusterId
                }
                // 判断集群ID是否存在当前项目的集群列表中
                const stateClusterList = res.data
                const curCluster = stateClusterList?.find(cluster => cluster.cluster_id === curClusterId)
                if (curCluster) {
                    // 缓存单集群信息
                    handleSetClusterStorageInfo(curCluster)
                } else {
                    handleSetClusterStorageInfo()
                }

                // 初始路由处理
                if ($route.name === 'entry') {
                    if (curProject.value?.kind === 2) {
                        // mesos需要二次刷新界面重新获取资源
                        const route = $router.resolve({
                            name: 'clusterMain'
                        })
                        window.location.href = route.href
                    } else {
                        // 路由中不带项目code时跳转到上一次缓存项目
                        $router.push({
                            name: 'clusterMain'
                        })
                    }
                } else if ($route.name !== 'clusterMain'
                    && pathClusterId
                    && !stateClusterList.some(cluster => cluster.cluster_id === pathClusterId)) {
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
                curProject,
                handleGotoIAM,
                projectList
            }
        }
    })
</script>
<style lang="postcss" scoped>
.bcs-content {
    width: 100%;
    flex: 1;
    max-width: calc(100vw - 261px);
}
</style>
