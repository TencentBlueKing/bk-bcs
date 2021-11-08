<template>
    <div :class="systemCls" v-bkloading="{ isLoading, opacity: 1 }">
        <Navigation @create-project="handleCreateProject">
            <router-view :key="routerKey" v-if="!isLoading && !err" />
        </Navigation>
        <!-- 项目创建弹窗 -->
        <ProjectCreate v-model="showCreateDialog" :project-data="null"></ProjectCreate>
        <!-- 权限弹窗 -->
        <app-apply-perm ref="bkApplyPerm"></app-apply-perm>
        <!-- 登录弹窗 -->
        <BkPaaSLogin ref="login" :width="width" :height="height"></BkPaaSLogin>
    </div>
</template>
<script>
    import Navigation from '@/views/navigation.vue'
    import ProjectCreate from '@/views/project/project-create.vue'
    import BkPaaSLogin from '@blueking/paas-login'
    import { bus } from '@/common/bus'

    export default {
        name: 'app',
        components: { Navigation, ProjectCreate, BkPaaSLogin },
        data () {
            return {
                isLoading: false,
                showCreateDialog: false,
                err: null
            }
        },
        computed: {
            systemCls () {
                const platform = window.navigator.platform.toLowerCase()
                const cls = platform.indexOf('win') === 0 ? 'win' : 'mac'
                return this.$store.state.isEn ? `${cls} english` : cls
            },
            routerKey () {
                const { projectCode = '' } = this.$route.params
                return `${projectCode}-${this.$route.meta.isDashboard}`
            },
            curProjectCode () {
                return this.$store.state.curProjectCode
            },
            width () {
                return this.$INTERNAL ? 700 : 400
            },
            height () {
                return this.$INTERNAL ? 510 : 400
            }
        },
        created () {
            // 权限弹窗弹窗
            bus.$on('show-apply-perm-modal', (data) => {
                if (!data) return
                this.$refs.bkApplyPerm && this.$refs.bkApplyPerm.show(this.curProjectCode, data)
            })
            bus.$on('close-login-modal', () => {
                window.location.reload()
            })
            window.addEventListener('message', (event) => {
                if (event.data === 'closeLoginModal') {
                    window.location.reload()
                }
            })
            this.initBcsBaseData()
        },
        mounted () {
            document.title = this.$t('容器服务')
            window.$loginModal = this.$refs.login
        },
        methods: {
            // 初始化BCS基本数据
            async initBcsBaseData () {
                this.isLoading = true
                await Promise.all([
                    this.$store.dispatch('userInfo'),
                    this.$store.dispatch('getProjectList')
                ]).catch((err) => {
                    this.err = err
                })
                this.isLoading = false
            },
            handleCreateProject () {
                this.showCreateDialog = true
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
