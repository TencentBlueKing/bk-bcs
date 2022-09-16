<template>
    <div class="detail p30">
        <!-- 基础信息 -->
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
                <label>Type</label>
                <span>{{ data.spec.type }}</span>
            </div>
            <div class="basic-info-item">
                <label>Cluster-ip</label>
                <span>{{ data.spec.clusterIP }}</span>
            </div>
            <div class="basic-info-item">
                <label>External-ip</label>
                <span>{{ extData.externalIP.join('/') || '--' }}</span>
            </div>
            <div class="basic-info-item">
                <label>Port(s)</label>
                <span>{{ extData.ports.join(' ') || '--' }}</span>
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
        <!-- 配置、标签、注解 -->
        <bcs-tab class="mt20" :label-height="40">
            <bcs-tab-panel name="config" :label="$t('配置')">
                <p class="detail-title">{{ $t('端口映射') }}（spec.ports）</p>
                <bk-table :data="data.spec.ports">
                    <bk-table-column label="Name" prop="name"></bk-table-column>
                    <bk-table-column label="Port" prop="port"></bk-table-column>
                    <bk-table-column label="Protocol" prop="protocol"></bk-table-column>
                    <bk-table-column label="TargetPort" prop="targetPort"></bk-table-column>
                    <bk-table-column label="NodePort" prop="nodePort">
                        <template #default="{ row }">
                            {{ row.nodePort || '--' }}
                        </template>
                    </bk-table-column>
                </bk-table>
            </bcs-tab-panel>
            <bcs-tab-panel name="label" :label="$t('标签')">
                <bk-table :data="handleTransformObjToArr(data.metadata.labels)">
                    <bk-table-column label="Key" prop="key"></bk-table-column>
                    <bk-table-column label="Value" prop="value"></bk-table-column>
                </bk-table>
            </bcs-tab-panel>
            <bcs-tab-panel name="annotations" :label="$t('注解')">
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
        name: 'ServiceDetail',
        props: {
            // 当前行数据
            data: {
                type: Object,
                default: () => ({})
            },
            // 当前行对应的manifest_ext数据
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
@import './network-detail.css'
</style>
