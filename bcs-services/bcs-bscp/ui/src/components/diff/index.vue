<script setup lang="ts">
    import { ref, watch, onMounted } from 'vue'
    import { useRoute } from 'vue-router'
    import { IConfigDiffDetail } from '../../../types/config';
    import { getConfigContent } from '../../api/config';
    import { byteUnitConverse } from '../../utils';
    import File from './file.vue'
    import Text from './text.vue'

    const route = useRoute()
    const bkBizId = ref(String(route.params.spaceId))

    const props = defineProps<{
        panelName?: String,
        appId: number;
        config: IConfigDiffDetail,
    }>()

    const current = ref()
    const base = ref()
    const contentLoading = ref(true)

    watch(() => props.config, (val) => {
        fetchConfigContent()
    })

    onMounted(() => {
        fetchConfigContent()
    })

    const fetchConfigContent = async () => {
        contentLoading.value = true
        const { base: configBase, current: configCurrent } = props.config
        if (!configCurrent.signature) { // 被删除
            current.value = ''
            base.value = await getDetailData(configBase)
        } else if (!configBase.signature) { // 新增
            base.value = ''
            current.value = await getDetailData(configCurrent)
        } else if (configBase.signature !== configCurrent.signature) { // 修改
            base.value = await getDetailData(configBase)
            current.value = await getDetailData(configCurrent)
        } else { // 未变更
            const data = await getDetailData(configBase)
            base.value = data
            current.value = data
        }
        contentLoading.value = false
    }

    const getDetailData = async (config: { signature: string; byte_size: string, update_at: string }) => {
        if (props.config.file_type === 'binary') {
            const { id, name } = props.config
            const { signature, update_at } = config
            return { id, name, signature, update_at, size: byteUnitConverse(Number(config.byte_size)) }
        }
        const configContent = await getConfigContent(bkBizId.value, props.appId, config.signature)
        return String(configContent)
    }

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
                    <div class="panel-name">{{ panelName }}</div>
                </slot>
            </div>
        </div>
        <bk-loading class="loading-wrapper" :loading="contentLoading">
            <div v-if="!contentLoading" class="detail-area">
                <File
                    v-if="props.config.file_type === 'binary'"
                    :current="current"
                    :base="base" />
                <Text
                    v-else
                    :current="current"
                    :base="base" />
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
            border-left: 1px solid #1d1d1d;
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
