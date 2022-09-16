<template>
    <div class="detail p30">
        <div class="detail-title">
            {{ $t('基础信息') }}
        </div>
        <div class="detail-content basic-info">
            <div class="basic-info-item">
                <label>{{ $t('命名空间') }}</label>
                <span>{{ data.metadata.namespace }}</span>
            </div>
            <div class="basic-info-item">
                <label>UID</label>
                <span class="bcs-ellipsis">{{ data.metadata.uid }}</span>
            </div>
            <div class="basic-info-item">
                <label>{{ $t('创建时间') }}</label>
                <span>{{ extData.createTime }}</span>
            </div>
            <div class="basic-info-item">
                <label>{{ $t('存在时间') }}</label>
                <span>{{ extData.age }}</span>
            </div>
        </div>
        <div class="mt20 mb10">
            Scerets
        </div>
        <div class="scerets-content">
            <bk-table :data="data.secrets">
                <bk-table-column label="name" prop="name"></bk-table-column>
            </bk-table>
        </div>
    </div>
</template>
<script lang="ts">
    import { defineComponent } from '@vue/composition-api'

    export default defineComponent({
        name: 'ConfigMapsDetail',
        props: {
            // 当前行数据
            data: {
                type: Object,
                default: () => ({})
            },
            extData: {
                type: Object,
                default: () => ({})
            }
        }
    })
</script>
<style lang="postcss" scoped>
.detail {
    font-size: 14px;
    /deep/ .bk-tab-label-item {
        background-color: #FAFBFD;
        border-bottom: 1px solid #dcdee5;
        line-height: 39px !important;
        height: 39px;
        &.active {
            border-bottom: none;
        }
    }
    /deep/ .bk-tab-label-wrapper {
        overflow: unset !important;
    }
    &-title {
        margin-bottom: 10px;
        color: #313238;
    }
    &-content {
        &.basic-info {
            border: 1px solid #dfe0e5;
            border-radius: 2px;
            .basic-info-item {
                display: flex;
                align-items: center;
                height: 32px;
                padding: 0 15px;
                &:nth-of-type(even) {
                    background: #F7F8FA;
                }
                label {
                    line-height: 32px;
                    border-right: 1px solid #dfe0e5;
                    width: 200px;
                }
                span {
                    padding: 0 15px;
                }
            }
        }
    }
}
</style>
