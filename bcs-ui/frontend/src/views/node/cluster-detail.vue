<template>
    <!-- 集群详情 -->
    <div>
        <ContentHeader :title="curCluster.name"
            :desc="`(${curCluster.clusterID})`"
            :hide-back="isSingleCluster"
        ></ContentHeader>
        <div class="cluster-detail">
            <div class="cluster-detail-tab">
                <div v-for="item in tabItems"
                    :key="item.com"
                    :class="['item', { active: activeCom === item.com }]"
                    @click="handleChangeActive(item)"
                >
                    <span class="icon"><i :class="item.icon"></i></span>
                    {{item.title}}
                </div>
            </div>
            <div class="cluster-detail-content">
                <component :is="activeCom"
                    :node-menu="false"
                    :cluster-id="clusterId"
                    :hide-cluster-select="true"
                    :selected-fields="[
                        'container_count',
                        'pod_count',
                        'cpu_usage',
                        'memory_usage',
                        'disk_usage',
                        'diskio_usage'
                    ]"
                ></component>
            </div>
        </div>
    </div>
</template>
<script lang="ts">
    import { computed, defineComponent, ref, toRefs } from '@vue/composition-api'
    import ContentHeader from '@/views/content-header.vue'
    import node from './node.vue'
    import overview from '@/views/cluster/overview.vue'
    import info from '@/views/cluster/info.vue'
    import useDefaultClusterId from './use-default-clusterId'

    export default defineComponent({
        components: {
            info,
            node,
            overview,
            ContentHeader
        },
        props: {
            active: {
                type: String,
                default: 'overview'
            },
            clusterId: {
                type: String,
                default: '',
                required: true
            }
        },
        setup (props, ctx) {
            const { $store, $i18n, $router } = ctx.root
            const { active, clusterId } = toRefs(props)
            const activeCom = ref(active)
            const curCluster = computed(() => {
                return $store.state.cluster.clusterList
                    ?.find(item => item.clusterID === clusterId.value) || {}
            })
            const tabItems = ref([
                {
                    icon: 'bcs-icon bcs-icon-bar-chart',
                    title: $i18n.t('总览'),
                    com: 'overview'
                },
                {
                    icon: 'bcs-icon bcs-icon-list',
                    title: $i18n.t('节点管理'),
                    com: 'node'
                },
                {
                    icon: 'cc-icon icon-cc-machine',
                    title: $i18n.t('集群信息'),
                    com: 'info'
                }
            ])
            const handleChangeActive = (item) => {
                if (activeCom.value === item.com) return
                activeCom.value = item.com
                $router.replace({
                    name: 'clusterDetail',
                    query: {
                        active: item.com
                    }
                })
            }
            const { isSingleCluster } = useDefaultClusterId()
            return {
                isSingleCluster,
                curCluster,
                tabItems,
                activeCom,
                handleChangeActive
            }
        }
    })
</script>
<style lang="postcss" scoped>
.cluster-detail {
    margin: 20px;
    border: 1px solid #dfe0e5;
    &-tab {
        display: flex;
        height: 60px;
        line-height: 60px;
        border-bottom: 1px solid #dfe0e5;
        font-size: 14px;
        .item {
            display: flex;
           align-items: center;
           justify-content: center;
            min-width: 140px;
            cursor: pointer;
            &.active {
                color: #3a84ff;
                background-color: #fff;
                border-right: 1px solid #dfe0e5;
                border-left: 1px solid #dfe0e5;
                font-weight: 700;
                i {
                    font-weight: 700;
                }
            }
            &:first-child {
                border-left: none;
            }
            .icon {
                font-size: 16px;
                margin-right: 8px;
                width: 16px;
                height: 16px;
                display: flex;
                align-items: center;
            }
        }
    }
    &-content {
        background-color: #fff;
    }
}
</style>
