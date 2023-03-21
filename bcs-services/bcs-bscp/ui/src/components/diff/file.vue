<script setup lang="ts">
    import { TextFill } from 'bkui-vue/lib/icon'

</script>
<template>
    <section class="file-diff">
        <div class="left-version-content">
            <div class="file-wrapper">
                <TextFill class="file-icon" />
                <div class="content">
                    <div class="name">host_model.bin</div>
                    <div class="time">2022-07-20 12:12:12</div>
                </div>
                <div class="size">4.24M</div>
            </div>
        </div>
        <div class="right-version-content">
            <div class="file-wrapper">
                <TextFill class="file-icon" />
                <div class="content">
                    <div class="name">host_model.bin</div>
                    <div class="time">2022-07-20 12:12:12</div>
                </div>
                <div class="size">5.6M</div>
            </div>
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
</style>
