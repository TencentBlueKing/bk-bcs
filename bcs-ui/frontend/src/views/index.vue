<template>
    <div class="biz-container">
        <div class="biz-side-bar">
            <side-nav></side-nav>
            <side-terminal></side-terminal>
        </div>
        <router-view :key="refreshRouterViewTimer || $route.path"></router-view>
    </div>
</template>

<script>
    import SideNav from './side-nav'
    import SideTerminal from '@open/components/terminal'
    import { getProjectByCode } from '@open/common/util'

    export default {
        components: {
            SideNav,
            SideTerminal
        },
        data () {
            return {
                refreshRouterViewTimer: null
            }
        },
        computed: {
            menuList () {
                return this.$store.state.sideMenu.menuList
            },
            k8sMenuList () {
                return this.$store.state.sideMenu.k8sMenuList
            },
            projectCode () {
                return this.$route.params.projectCode
            }
        },
        watch: {
            // 'projectCode' () {
            //     if (this.projectCode) {
            //         this.$store.commit('updateCurProject', this.projectCode)
            //     }
            // },
            '$route' (to, from) {
                if (to.params.needCheckPermission) {
                    this.checkPermission(to.name)
                }
            }
        },
        created () {
            this.checkPermission(this.$route.name)
        },
        methods: {
            /**
             * 检测当前项目的权限
             *
             * @param {string} routeName routeName，对于 menuList 中 menu.pathName
             */
            async checkPermission (routeName) {
                const currentProject = getProjectByCode(window.$currentProjectId)
                let menuList = this.menuList
                const kind = currentProject.kind
                // k8s, tke
                if (kind === 1 || kind === 3) {
                    menuList = this.k8sMenuList
                }

                const len = menuList.length
                let continueLoop = true
                for (let i = len - 1; i >= 0; i--) {
                    if (!continueLoop) {
                        break
                    }
                    const menu = menuList[i]
                    if ((menu.pathName || []).indexOf(routeName) > -1) {
                        continueLoop = false
                        break
                    }
                    if (menu.children) {
                        const childrenLen = menu.children.length
                        for (let j = childrenLen - 1; j >= 0; j--) {
                            if ((menu.children[j].pathName || []).indexOf(routeName) > -1) {
                                continueLoop = false
                                break
                            }
                        }
                    }
                }
            },
            /**
             * 通过改变 router-view 里的 key 来实现强制刷新 router-view
             */
            refreshRouterView () {
                this.refreshRouterViewTimer = +new Date()
            }
        }
    }
</script>
