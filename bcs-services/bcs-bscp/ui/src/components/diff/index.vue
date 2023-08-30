<script setup lang="ts">
    import { IDiffDetail } from '../../../types/service';
    import { IFileConfigContentSummary } from '../../../types/config';
    import File from './file.vue'
    import Text from './text.vue'

    const props = defineProps<{
        panelName?: String;
        diff: IDiffDetail;
        loading: boolean
    }>()
</script>
<template>
    <section class="diff-comp-panel">
        <div class="top-area">
            <div class="left-panel">
                <slot name="leftHead">
                </slot>
            </div>
            <div class="right-panel">
                <slot name="rightHead">
                    <div class="panel-name">{{ props.panelName }}</div>
                </slot>
            </div>
        </div>
        <bk-loading class="loading-wrapper" :loading="props.loading">
            <div v-if="!props.loading" class="detail-area">
                <File
                    v-if="props.diff.contentType === 'file'"
                    :current="(props.diff.current.content as IFileConfigContentSummary)"
                    :base="(props.diff.base.content as IFileConfigContentSummary)" />
                <Text
                    v-else
                    :language="props.diff.current.language"
                    :current="(props.diff.current.content as string)"
                    :base="(props.diff.base.content as string)" />
            </div>
        </bk-loading>
    </section>
</template>
<style lang="scss" scoped>
    .diff-comp-panel {
        height: 100%;
        border-left: 1px solid #dcdee5;
    }
    .top-area {
        display: flex;
        align-items: center;
        height: 49px;
        color: #313238;
        // border-bottom: 1px solid #dcdee5;
        .left-panel,
        .right-panel {
            height: 100%;
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
    .loading-wrapper {
        height: calc(100% - 49px);
    }
    .detail-area {
        height: 100%;
    }
</style>
