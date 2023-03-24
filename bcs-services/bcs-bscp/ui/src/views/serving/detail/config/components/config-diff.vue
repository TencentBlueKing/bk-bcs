<script setup lang="ts">
    import { ref, watch, withDefaults, onMounted } from 'vue'
    import { PlayShape } from 'bkui-vue/lib/icon'
    import Diff from '../../../../../components/diff/index.vue'
    import { storeToRefs } from 'pinia'
    import { useServingStore } from '../../../../../store/serving'
    import { IConfigItem, IConfigDetail, IConfigDiffDetail } from '../../../../../../types/config'
    import { getConfigItemDetail } from '../../../../../api/config'

    const { appData } = storeToRefs(useServingStore())

    const props = withDefaults(defineProps<{
        versionName?: string,
        currentVersion?: number,
        baseVersion: number|undefined,
        currentConfigList: Array<IConfigItem>, // 源配置列表
        baseConfigList: Array<IConfigItem>, // 对比目标配置列表
    }>(), {
        currentConfigList: () => [],
        baseConfigList: () => []
    })

    const selectedConfig = ref<IConfigDiffDetail>({
        id: 0,
        name: '',
        file_type: '',
        current: {
            signature: '',
            byte_size: '',
            update_at: ''
        },
        base: {
            signature: '',
            byte_size: '',
            update_at: ''
        }
    })
    const currentConfigDetailsLoading = ref(false)
    const baseConfigDetailsLoading = ref(false)
    const currentList = ref<IConfigDetail[]>([])
    const baseList = ref<IConfigDetail[]>([])
    const aggregatedList = ref<IConfigDiffDetail[]>([]) // 汇总的配置项列表
    const diffCount = ref(0)

    watch(() => props.baseConfigList, async(val) => {
        baseConfigDetailsLoading.value = true
        baseList.value = await getConfigDetails(val, props.baseVersion)
        calcDiff()
        selectedConfig.value = aggregatedList.value[0]
        baseConfigDetailsLoading.value = false
    })

    onMounted(async () => {
        currentConfigDetailsLoading.value = true
        currentList.value = await getConfigDetails(props.currentConfigList, <number>props.currentVersion)
        calcDiff()
        selectedConfig.value = aggregatedList.value[0]
        currentConfigDetailsLoading.value = false
    })

    const getConfigDetails = (list: IConfigItem[], version: number|undefined) => {
        const params: { release_id?: number } = {}
        if (version) {
            params.release_id = version
        }
        return Promise.all(list.map(item => getConfigItemDetail(item.id, <number>appData.value.id), params))
    }

    // 计算配置被修改、被删除、新增的差异
    const calcDiff = () => {
        diffCount.value = 0
        const list: IConfigDiffDetail[]= []
        currentList.value.forEach(currentItem => {
            const { config_item } = currentItem
            const baseItem = baseList.value.find(item => config_item.id === item.config_item.id)
            if (baseItem) {
                const diffConfig = {
                    id: config_item.id,
                    name: config_item.spec.name,
                    file_type: config_item.spec.file_type,
                    current: {
                        ...currentItem.content,
                        update_at: config_item.revision.update_at
                    },
                    base: {
                        ...baseItem.content,
                        update_at: baseItem.config_item.revision.update_at
                    }
                }
                if (currentItem.content.signature !== baseItem.content.signature) {
                    diffCount.value++
                }
                list.push(diffConfig)
            } else {
                diffCount.value++
                list.push({
                    id: config_item.id,
                    name: config_item.spec.name,
                    file_type: config_item.spec.file_type,
                    current: {
                        ...currentItem.content,
                        update_at: config_item.revision.update_at
                    },
                    base: {
                        signature: '',
                        byte_size: '',
                        update_at: ''
                    }
                })
            }
        })
        baseList.value.forEach(baseItem => {
            const { config_item: base_config_item } = baseItem
            const currentItem = currentList.value.find(item => base_config_item.id === item.config_item.id)
            if (!currentItem) {
                diffCount.value++
                list.push({
                    id: base_config_item.id,
                    name: base_config_item.spec.name,
                    file_type: base_config_item.spec.file_type,
                    current: {
                        signature: '',
                        byte_size: '',
                        update_at: ''
                    },
                    base: {
                        ...baseItem.content,
                        update_at: ''
                    }
                })
            }
        })
        aggregatedList.value = list
    }

</script>
<template>
    <bk-loading style="height: 100%;" :loading="currentConfigDetailsLoading || baseConfigDetailsLoading">
        <section class="config-diff-panel">
            <aside class="config-list-side">
                <div class="title-area">
                    <span class="title">配置项</span>
                    <span>共 <span class="count">{{ diffCount }}</span> 项配置有差异</span>
                </div>
                <ul class="configs-wrapper">
                    <li
                        v-for="config in aggregatedList"
                        :key="config.id"
                        :class="{ active: selectedConfig.id === config.id }"
                        @click="selectedConfig = config">
                        <div class="name">{{ config.name }}</div>
                        <PlayShape v-if="selectedConfig.id === config.id" class="arrow-icon" />
                    </li>
                </ul>
            </aside>
            <div class="config-diff-detail">
                <diff
                    :panelName="props.versionName"
                    :config="selectedConfig"
                    type="file">
                    <template #leftHead>
                        <slot name="baseHead"></slot>
                    </template>
                    <template #rightHead>
                        <slot name="currentHead"></slot>
                    </template>
                </diff>
            </div>
        </section>
    </bk-loading>
</template>
<style lang="scss" scoped>
    .config-diff-panel {
        display: flex;
        align-items: center;
        height: 100%;
        background: #ffffff;
        box-shadow: 0 2px 2px 0 rgba(0,0,0,0.15);
    }
    .config-list-side {
        width: 264px;
        height: 100%;
        .title-area {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 0 24px;
            height: 49px;
            color: #979ba5;
            font-size: 12px;
            border-bottom: 1px solid #dcdee5;
            .title {
                font-size: 14px;
                font-weight: 700;
                color: #63656e;
            }
            .count {
                color: #313238;
            }
        }
    }
    .configs-wrapper {
        height: calc(100% - 49px);
        overflow: auto;
        & > li {
            display: flex;
            align-items: center;
            justify-content: space-between;
            position: relative;
            padding: 0 24px;
            height: 41px;
            color: #313238;
            border-bottom: 1px solid #dcdee5;
            cursor: pointer;
            &:hover {
                background: #e1ecff;
                color: #3a84ff;
            }
            &.active {
                background: #e1ecff;
                color: #3a84ff;
            }
            .name {
                width: calc(100% - 24px);
                line-height: 16px;
                font-size: 12px;
                white-space: nowrap;
                text-overflow: ellipsis;
                overflow: hidden;
            }
            .arrow-icon {
                position: absolute;
                top: 50%;
                right: 5px;
                transform: translateY(-60%);
                font-size: 12px;
                color: #3a84ff;
            }
        }
    }
    .config-diff-detail {
        width: calc(100% - 264px);
        height: 100%;
    }
</style>
