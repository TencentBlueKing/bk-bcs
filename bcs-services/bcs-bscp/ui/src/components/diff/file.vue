<script setup lang="ts">
    import { useRoute } from 'vue-router'
    import { TextFill } from 'bkui-vue/lib/icon'
    import { IFileConfigContentSummary } from '../../../types/config';
    import { getConfigContent } from '../../api/config'
    import { fileDownload } from '../../utils/file'

    const route = useRoute()
    const bkBizId = String(route.params.spaceId)
    const appId = Number(route.params.appId)

    const props = defineProps<{
        current: IFileConfigContentSummary,
        base: IFileConfigContentSummary
    }>()

      // 下载已上传文件
    const handleDownloadFile = async (config: IFileConfigContentSummary) => {
        const { signature, name } = config
        const res = await getConfigContent(bkBizId, appId, signature)
        fileDownload(res, `${name}.bin`)
    }

</script>
<template>
    <section class="file-diff">
        <div class="left-version-content">
            <div v-if="props.base" class="file-wrapper" @click="handleDownloadFile(props.base)">
                <TextFill class="file-icon" />
                <div class="content">
                    <div class="name">{{ props.base.name }}</div>
                    <div class="time">{{ props.base.update_at }}</div>
                </div>
                <div class="size">{{ props.base.size }}</div>
            </div>
            <bk-exception v-else class="exception-tips" scene="part" type="empty">该版本下文件不存在</bk-exception>
        </div>
        <div class="right-version-content">
            <div v-if="props.current" class="file-wrapper" @click="handleDownloadFile(props.current)">
                <TextFill class="file-icon" />
                <div class="content">
                    <div class="name">{{ props.current.name }}</div>
                    <div class="time">{{ props.current.update_at }}</div>
                </div>
                <div class="size">{{ props.current.size }}</div>
            </div>
            <bk-exception v-else class="exception-tips" scene="part" theme="empty">文件已被删除</bk-exception>
        </div>
    </section>
</template>
<style lang="scss" scoped>
    .file-diff {
        display: flex;
        align-items: center;
        height: 100%;
        background: #fafbfd;
    }
    .left-version-content,
    .right-version-content {
        padding: 24px;
        width: 50%;
        height: 100%;
    }
    .right-version-content {
        border-left: 1px solid #dcdee5;
    }
    .file-wrapper {
        padding: 21px 16px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        background: #ffffff;
        font-size: 12px;
        border: 1px solid #c4c6cc;
        border-radius: 2px;
        cursor: pointer;
        &:hover {
            border-color: #3a84ff;
        }
    }
    .file-icon {
        margin-right: 17px;
        font-size: 28px;
        color: #63656e;
    }
    .content {
        flex: 1;
        .name {
            color: #63656e;
            line-height: 20px;
        }
        .time {
            margin-top: 2px;
            color: #979ba5;
            line-height: 16px;
        }
    }
    .size {
        color: #63656e;
        font-weight: 700;
    }
    .exception-tips {
        margin-top: 100px;
        font-size: 12px;
    }
</style>
