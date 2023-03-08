<script setup lang="ts">
    import { defineProps, ref, watch } from 'vue'
    import { useStore } from 'vuex'
    import { PlayShape } from 'bkui-vue/lib/icon'
    import Diff from '../../../../../components/diff/index.vue'
    import { IConfigVersionItem } from '../../../../../types'

    const props = defineProps<{
        versionName?: string,
        configList: Array<IConfigVersionItem>
    }>()

    const selectedConfig = ref(0)

    watch(() => props.configList, () => {
        if (props.configList.length > 0) {
            // @ts-ignore
            selectedConfig.value = props.configList[0].id
        }
    }, { immediate: true })

</script>
<template>
    <section class="config-diff-panel">
        <aside class="config-list-side">
            <div class="title-area">
                <span class="title">配置项</span>
                <span>共 <span class="count">xx</span> 文件存在差异</span>
            </div>
            <ul class="configs-wrapper">
                <li
                    v-for="config in props.configList"
                    :key="config.id"
                    :class="{ active: selectedConfig === config.id }"
                    @click="selectedConfig = config.id">
                    <div class="name">{{ config.spec.name }}</div>
                    <div class="count-area">
                        <div class="num">16</div>
                    </div>
                    <PlayShape v-if="selectedConfig === config.id" class="arrow-icon" />
                </li>
            </ul>
        </aside>
        <div class="config-diff-detail">
            <diff :panelName="props.versionName" type="file">
                <template #leftHead>
                    <slot name="head"></slot>
                </template>
            </diff>
        </div>
    </section>
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
            height: 48px;
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
        height: calc(100% - 48px);
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
                .num {
                    background: #a3c5fd;
                    color: #ffffff;
                }
            }
            .name {
                width: calc(100% - 64px);
                line-height: 16px;
                font-size: 12px;
                white-space: nowrap;
                text-overflow: ellipsis;
                overflow: hidden;
            }
            .count-area {
                display: flex;
                align-items: center;
                height: 100%;
            }
            .num {
                min-width: 30px;
                padding: 0 8px;
                height: 16px;
                line-height: 16px;
                background: #f0f1f5;
                color: #979ba5;
                border-radius: 2px;
                text-align: center;
                font-size: 12px;
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
