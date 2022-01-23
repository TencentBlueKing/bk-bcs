<template>
    <div
        :class="['bk-dropdown-menu biz-terminal active', { 'active': isActive }]"
        v-if="curProject && clusterList.length && !isSharedCluster"
        @mouseover="handlerMouseover"
        @mouseout="handlerMouseout">
        <div class="bk-dropdown-trigger">
            <div class="biz-terminal-trigger">
                <img src="@/images/terminal.svg" class="icon">
                <span class="text">WebConsole</span>
                <a href="https://bk.tencent.com/docs/document/6.0/144/6541" target="_blank" class="terminal-helper bk-text-icon">
                    <i class="bcs-icon bcs-icon-helper"></i>
                </a>
            </div>
            <transition name="fade">
                <div :class="['bk-dropdown-content is-show']" style="bottom: 44px; right: 0; position: absolute;" v-show="isShow">
                    <div class="search-box">
                        <bkbcs-input
                            v-model="keyword"
                            :placeholder="$t('请输入集群名或ID')"
                            @focus="isFocus = true"
                            @blur="isFocus = false">
                        </bkbcs-input>
                    </div>
                    <ul class="bk-dropdown-list" v-if="searchList.length">
                        <li>
                            <a href="javascript:;" v-for="(cluster, index) in searchList" :key="index" @click="goWebConsole(cluster)" style="padding: 0 10px;" :title="cluster.name">{{cluster.name}}</a>
                        </li>
                    </ul>
                    <div v-else class="no-result">
                        {{$t('无数据')}}
                    </div>
                </div>
            </transition>
        </div>
    </div>
</template>

<script>
    import { mapGetters } from 'vuex'
    export default {
        data () {
            return {
                PROJECT_K8S: window.PROJECT_K8S,
                PROJECT_TKE: window.PROJECT_TKE,
                terminalWins: null,
                isActive: false,
                isShow: false,
                showTimer: 0,
                activeTimer: 0,
                keyword: '',
                isFocus: false
            }
        },
        computed: {
            searchList () {
                const clusterList = [...this.$store.state.cluster.clusterList]
                const list = clusterList.filter(item => {
                    const name = item.name.toLowerCase()
                    const clusterId = item.cluster_id.toLowerCase()
                    const keyword = this.keyword.toLowerCase()
                    return name.indexOf(keyword) > -1 || clusterId.indexOf(keyword) > -1
                })

                return list
            },
            clusterList () {
                return [...this.$store.state.cluster.clusterList]
            },
            clusterPerm () {
                return this.$store.state.cluster.clusterPerm
            },
            curProject () {
                return this.$store.state.curProject
            },
            projectId () {
                return this.$route.params.projectId
            },
            routeName () {
                return this.$route.name
            },
            ...mapGetters('cluster', ['isSharedCluster'])
        },
        methods: {
            handlerMouseover () {
                clearTimeout(this.showTimer)
                clearTimeout(this.activeTimer)
                this.isActive = true
                if (!this.isFocus) {
                    this.keyword = ''
                }
                this.showTimer = setTimeout(() => {
                    this.isShow = true
                }, 200)
            },
            handlerMouseout () {
                clearTimeout(this.showTimer)
                clearTimeout(this.activeTimer)
                if (this.isFocus) {
                    return false
                }
                this.showTimer = setTimeout(() => {
                    this.isShow = false
                }, 100)
                this.activeTimer = setTimeout(() => {
                    this.isActive = false
                }, 400)
            },
            async goWebConsole (cluster) {
                if (!this.clusterPerm[cluster.clusterID]?.policy?.use) {
                    const type = `cluster_${cluster.environment === 'prod' ? 'prod' : 'test'}`
                    const params = {
                        project_id: this.projectId,
                        policy_code: 'use',
                        resource_code: cluster.cluster_id,
                        resource_name: cluster.name,
                        resource_type: type
                    }
                    await this.$store.dispatch('getResourcePermissions', params)
                }

                const clusterId = cluster.cluster_id
                const url = `${DEVOPS_BCS_API_URL}/web_console/projects/${this.projectId}/mgr/#cluster=${clusterId}`

                this.keyword = ''
                this.isShow = false
                if (this.terminalWins) {
                    if (!this.terminalWins.closed) {
                        this.terminalWins.postMessage({
                            clusterId: clusterId,
                            clusterName: cluster.name
                        }, DEVOPS_BCS_HOST)
                        this.terminalWins.focus()
                    } else {
                        this.terminalWins = window.open(url, '')
                    }
                } else {
                    this.terminalWins = window.open(url, '')
                }
            }
        }
    }
</script>

<style scoped lang="postcss">
    @import "@/css/mixins/ellipsis.css";

    .biz-terminal {
        position: fixed;
        right: 10px;
        bottom: 10px;
        z-index: 1900;
        &.active {
            .biz-terminal-trigger {
                width: 200px;
                border-radius: 2px;
                .text {
                    display: inline-block;
                }
                .terminal-helper {
                    display: inline-block;
                }
            }
        }
        .biz-terminal-trigger {
            height: 42px;
            text-align: left;
            line-height: 40px;
            background: #fff;
            overflow: hidden;
            border-radius: 50%;
            transition: all ease 0.3s;
            box-shadow: 0 0 12px rgba(0, 0, 0, 0.1);
            cursor: pointer;
            z-index: 10;
            padding: 0 10px;

            .icon {
                width: 16px;
                vertical-align: middle;
            }
            .terminal-helper {
                display: none;
                font-size: 14px;
            }
            .text {
                color: #333;
                font-size: 14px;
                vertical-align: middle;
                display: inline-block;
                overflow: hidden;
                display: none;

            }
        }

        .terminal-helper {
            vertical-align: middle;
            &:hover {
                color: red;
                .bcs-icon {
                    color: red !important;
                }
            }
        }

        .bk-dropdown-list {
            > li {
                width: 200px;
                a {
                    display: block;
                    vertical-align: middle;
                    width: 200px;
                    @mixin ellipsis 200px;
                }
            }
        }

        .no-result {
            line-height: 82px;
            text-align: center;
            font-size: 14px;
        }

        .search-box {
            padding: 10px;
            border-bottom: 1px solid #eee;
        }
    }
</style>
