<template>
    <div v-show="value" v-bk-clickoutside="handleHideClusterSelector" class="biz-cluster-selector">
        <div class="biz-cluster-search">
            <bk-input
                ref="searchInput"
                :placeholder="$t('请输入关键字')"
                behavior="simplicity"
                v-model="searchValue"
                clearable
                show-word-limit
                right-icon="bk-icon icon-search">
            </bk-input>
        </div>
        <div class="biz-cluster-list">
            <ul v-if="filterClusterList.length">
                <li
                    v-for="(cluster, index) in filterClusterList"
                    :key="index"
                    :class="{ 'active': activeClusterId === cluster.cluster_id }"
                    @click="handleToggleCluster(cluster)">
                    {{ cluster.name }}
                    <p style="color: #979ba5;">{{ cluster.cluster_id }}</p>
                </li>
            </ul>
            <div v-else class="cluster-nodata">{{ $t('暂无数据') }}</div>
        </div>
        <div class="biz-cluster-action" v-if="curViewType === 'cluster'">
            <span class="action-item" @click="gotCreateCluster">
                <i class="bcs-icon bcs-icon-plus-circle"></i>
                {{ $t('新增集群') }}
            </span>
            <span class="line">|</span>
            <span class="action-item" @click="handleToggleCluster({
                name: $t('全部集群'),
                cluster_id: ''
            })">
                <i class="bcs-icon bcs-icon-quanbujiqun"></i>
                {{ $t('全部集群') }}
            </span>
        </div>
    </div>
</template>

<script>
    import { isEmpty } from '@open/common/util'
    const BCS_CLUSTER = 'bcs-cluster'

    export default {
        name: 'cluster-selector',
        model: {
            event: 'display-change',
            prop: 'value'
        },
        props: {
            value: {
                type: Boolean,
                default: false
            }
        },
        data () {
            return {
                searchValue: '',
                createPermission: false,
                activeClusterId: ''
            }
        },
        computed: {
            projectId () {
                return this.$route.params.projectId
            },
            projectCode () {
                return this.$route.params.projectCode
            },
            curClusterList () {
                return this.$store.state.cluster.clusterList || []
            },
            filterClusterList () {
                return isEmpty(this.searchValue) ? this.curClusterList : this.curClusterList.filter(item => item.name.includes(this.searchValue))
            },
            curViewType () {
                return this.$route.path.indexOf('dashboard') > -1 ? 'dashboard' : 'cluster'
            }
        },
        watch: {
            value (show) {
                if (show) {
                    this.$nextTick(() => {
                        this.$refs.searchInput.focus()
                    })
                } else {
                    this.searchValue = ''
                }
            }
        },
        created () {
            this.activeClusterId = this.$store.state.curClusterId
        },
        methods: {
            async getClusterCreatePermission () {
                this.createPermission = await this.$store.dispatch('getMultiResourcePermissions', {
                    project_id: this.projectId,
                    operator: 'or',
                    resource_list: [
                        {
                            policy_code: 'create',
                            resource_type: 'cluster_test'
                        },
                        {
                            policy_code: 'create',
                            resource_type: 'cluster_prod'
                        }
                    ]
                }).then(() => true).catch(() => false)
            },
            /**
             * 点击除自身元素外，关闭集群选择弹窗
             */
            handleHideClusterSelector () {
                this.$emit('display-change', false)
            },

            /**
             * 集群切换
             * @param {Object} cluster 集群信息
             */
            handleToggleCluster (cluster) {
                this.activeClusterId = cluster.cluster_id
                this.handleSaveClusterInfo(cluster)
                this.handleHideClusterSelector()
                this.$store.dispatch('getFeatureFlag')
                // 抛出选中的集群信息
                this.$emit('change', cluster)
            },

            /**
             * 保存cluster信息
             */
            handleSaveClusterInfo (cluster) {
                localStorage.setItem(BCS_CLUSTER, cluster.cluster_id)
                sessionStorage.setItem(BCS_CLUSTER, cluster.cluster_id)
                this.$store.commit('cluster/forceUpdateCurCluster', cluster.cluster_id ? cluster : {})
                this.$store.commit('updateCurClusterId', cluster.cluster_id)
            },

            /**
             * 新建集群
             */
            async gotCreateCluster () {
                await this.getClusterCreatePermission()

                if (!this.createPermission) return

                this.handleHideClusterSelector()
                this.$router.push({
                    name: 'clusterCreate',
                    params: {
                        projectId: this.projectId,
                        projectCode: this.projectCode
                    }
                })
            }
        }
    }
</script>

<style scoped>
    @import './index.css';
</style>
