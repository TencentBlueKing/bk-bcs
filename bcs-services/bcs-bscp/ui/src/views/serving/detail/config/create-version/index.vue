<script setup lang="ts">
    import { defineProps, ref } from 'vue'
    import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon'
    import VersionLayout from '../components/version-layout.vue'
    import ConfirmDialog from './confirm-dialog.vue'
    import ConfigDiff from '../config-diff.vue'

    const props = defineProps<{
        bkBizId: string,
        appId: number,
        appName: string
    }>()

    const showDiffPanel = ref(false)
    const isConfirmDialogShow = ref(false)
    
    const handleClose = () => {
        showDiffPanel.value = false
    }

</script>
<template>
    <section class="create-version">
        <bk-button theme="primary" @click="showDiffPanel = true">生成版本</bk-button>
        <VersionLayout v-if="showDiffPanel">
            <template #header>
                <section class="header-wrapper">
                    <span class="service-name" @click="handleClose">
                        <ArrowsLeft class="arrow-left" />
                        <span class="service-name">{{ props.appName }}</span>
                    </span>
                    <AngleRight class="arrow-right" />
                    生成版本
                </section>
            </template>
            <config-diff version-name="未命名版本"></config-diff>
            <template #footer>
                <section class="actions-wrapper">
                    <bk-button theme="primary" style="margin-right: 8px;" @click="isConfirmDialogShow = true">生成版本</bk-button>
                    <bk-button @click="handleClose">取消</bk-button>
                </section>
            </template>
        </VersionLayout>
        <ConfirmDialog
            v-model:show="isConfirmDialogShow"
            :bk-biz-id="props.bkBizId"
            :app-id="props.appId"
            @confirm="handleClose" />
    </section>
</template>
<style lang="scss" scoped>
    .header-wrapper {
        display: flex;
        align-items: center;
        padding: 0 24px;
        height: 100%;
        font-size: 12px;
        line-height: 1;
    }
    .service-name {
        display: flex;
        align-items: center;
        font-size: 12px;
        color: #3a84ff;
        cursor: pointer;
    }
    .arrow-left {
        font-size: 26px;
        color: #3884ff;
    }
    .arrow-right {
        font-size: 24px;
        color: #c4c6cc;
    }
    .actions-wrapper {
        display: flex;
        align-items: center;
        padding: 0 24px;
        height: 100%;
    }
</style>
