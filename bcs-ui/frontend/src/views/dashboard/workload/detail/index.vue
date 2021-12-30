<template>
    <div class="detail">
        <DetailTopNav :titles="titles" @change="handleNavChange"></DetailTopNav>
        <!-- <keep-alive>
        </keep-alive> -->
        <component
            :is="componentId"
            v-bind="componentProps"
            @pod-detail="handleGotoPodDetail"
            @container-detail="handleGotoContainerDetail">
        </component>
    </div>
</template>
<script lang="ts">
    import { defineComponent, ref, computed } from '@vue/composition-api'
    import DetailTopNav from './detail-top-nav.vue'
    import WorkloadDetail from './workload-detail.vue'
    import PodDetail from './pod-detail.vue'
    import ContainerDetail from './container-detail.vue'

    export type ComponentIdType = 'WorkloadDetail' | 'PodDetail' | 'ContainerDetail'
    export interface ITitle {
        name: string; // 展示名称
        id: string;// 组件ID
        params?: any; // 组件参数
    }

    export default defineComponent({
        components: {
            DetailTopNav,
            WorkloadDetail,
            PodDetail,
            ContainerDetail
        },
        props: {
            // 命名空间
            namespace: {
                type: String,
                default: ''
            },
            // workload类型
            category: {
                type: String,
                default: ''
            },
            // 名称
            name: {
                type: String,
                default: ''
            },
            // kind类型
            kind: {
                type: String,
                default: '',
                required: true
            }
        },
        setup (props, ctx) {
            const { $router, $store } = ctx.root
            // 区分首次进入pod详情还是其他workload详情
            const defaultComId = props.category === 'pods' ? 'PodDetail' : 'WorkloadDetail'
            // 子标题
            const subTitleMap = {
                deployments: 'Deploy',
                daemonsets: 'DS',
                statefulsets: 'STS',
                cronjobs: 'CJ',
                jobs: 'Job',
                pods: 'Pod',
                container: 'Container'
            }
            // 首字母大写
            const upperFirstLetter = (str: string) => {
                if (!str) return str

                return `${str.slice(0, 1).toUpperCase()}${str.slice(1)}`
            }
            // 顶部导航内容
            const titles = ref<ITitle[]>([
                {
                    name: upperFirstLetter(props.category),
                    id: ''
                },
                {
                    name: `${subTitleMap[props.category]}: ${props.name}`,
                    id: defaultComId,
                    params: {
                        ...props
                    }
                }
            ])
            const componentId = ref(defaultComId)

            const componentProps = computed(() => {
                return titles.value.find(item => item.id === componentId.value)?.params || {} // 详情组件所需的参数
            })

            const handleNavChange = (item: ITitle) => {
                const { id } = item
                const index = titles.value.findIndex(item => item.id === id)
                if (id === '') {
                    $router.push({ name: $store.getters.curNavName })
                } else {
                    componentId.value = id
                    if (index > -1) {
                        // 截取后面的导航
                        titles.value = titles.value.slice(0, index + 1)
                    } else {
                        titles.value.push(item)
                    }
                }
            }
            // 调转pod详情
            const handleGotoPodDetail = (row) => {
                handleNavChange({
                    name: `${subTitleMap.pods}: ${row.metadata.name}`,
                    id: 'PodDetail',
                    params: {
                        name: row.metadata.name,
                        namespace: row.metadata.namespace
                    }
                })
            }
            // 调转容器详情
            const handleGotoContainerDetail = (row) => {
                // 容器的父级Pod
                const { name } = titles.value.find(item => item.id === componentId.value)?.params || {}
                handleNavChange({
                    name: `${subTitleMap.container}: ${row.name}`,
                    id: 'ContainerDetail',
                    params: {
                        namespace: props.namespace,
                        pod: name,
                        name: row.name,
                        id: row.container_id
                    }
                })
            }

            return {
                componentId,
                componentProps,
                titles,
                handleNavChange,
                handleGotoPodDetail,
                handleGotoContainerDetail
            }
        }
    })
</script>
<style scoped>
.detail {
    flex: 1;
}
</style>
