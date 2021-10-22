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
        <bcs-tab class="mt20" type="card" :label-height="40">
            <bcs-tab-panel name="data" label="Data">
                <bk-table :data="handleTransformObjToArr(data.data)">
                    <bk-table-column label="Key" prop="key"></bk-table-column>
                    <bk-table-column label="Value" prop="value"></bk-table-column>
                </bk-table>
            </bcs-tab-panel>
            <bcs-tab-panel name="lebel" :label="$t('标签')">
                <bk-table :data="handleTransformObjToArr(data.metadata.labels)">
                    <bk-table-column label="Key" prop="key"></bk-table-column>
                    <bk-table-column label="Value" prop="value"></bk-table-column>
                </bk-table>
            </bcs-tab-panel>
            <bcs-tab-panel name="annotation" :label="$t('注解')">
                <bk-table :data="handleTransformObjToArr(data.metadata.annotations)">
                    <bk-table-column label="Key" prop="key"></bk-table-column>
                    <bk-table-column label="Value" prop="value"></bk-table-column>
                </bk-table>
            </bcs-tab-panel>
        </bcs-tab>
    </div>
</template>
<script lang="ts">
    import { defineComponent } from '@vue/composition-api'

    export default defineComponent({
        name: 'SecretsDetail',
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
        },
        setup (props, ctx) {
            const handleTransformObjToArr = (obj) => {
                if (!obj) return []

                return Object.keys(obj).reduce<any[]>((data, key) => {
                    data.push({
                        key,
                        value: obj[key]
                    })
                    return data
                }, [])
            }

            return {
                handleTransformObjToArr
            }
        }
    })
</script>
<style lang="postcss" scoped>
@import './config-detail.css'
</style>
