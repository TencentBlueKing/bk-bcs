<script setup lang="ts">
    import { defineProps } from 'vue'
    import File from './file.vue'
    import Text from './text.vue'

    const props = defineProps<{
        type: String,
        panelName?: String
    }>()

</script>
<template>
    <section class="diff-comp-panel">
        <div class="top-area">
            <div class="left-panel">
                <slot name="leftHead">
                    <div class="panel-name">{{ panelName }}</div>
                </slot>
            </div>
            <div class="right-panel">
                <slot name="rightHead">
                    <div class="config-select-area">
                        <span>对比版本：</span>
                        <bk-select>
                        </bk-select>
                    </div>
                </slot>
            </div>
        </div>
        <div class="detail-area">
            <File v-if="props.type === 'file'" />
            <Text v-else />
        </div>
    </section>
</template>
<style lang="scss" scoped>
    .diff-comp-panel {
        height: 100%;
        border: 1px solid #dcdee5;
    }
    .top-area {
        display: flex;
        align-items: center;
        color: #313238;
        border-bottom: 1px solid #dcdee5;
        .left-panel,
        .right-panel {
            width: 50%;
        }
        .right-panel {
            border-left: 1px solid #dcdee5;
        }
        .panel-name {
            padding: 16px;
            font-size: 12px;
            line-height: 1;
            white-space: nowrap;
            text-overflow: ellipsis;
            overflow: hidden;
        }
        .config-select-area {
            display: flex;
            align-items: center;
            padding: 8px 16px;
            font-size: 12px;
        }
    }
    .detail-area {
        height: calc(100% - 49px);
    }
</style>
