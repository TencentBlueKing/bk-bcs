<template>
    <bk-form :label-width="100" :model="formData" :rules="rules" ref="formMode">
        <bk-form-item :label="$t('版本')" property="version" error-display-type="normal" required>
            <bcs-select v-model="formData.clusterBasicSettings.version" searchable :clearable="false">
                <bcs-option v-for="item in versionList" :key="item" :id="item" :name="item"></bcs-option>
            </bcs-select>
        </bk-form-item>
        <bk-form-item :label="$t('网络类型')" required>
            <bcs-select v-model="formData.networkType" :clearable="false">
                <bcs-option id="underlay" name="underlay"></bcs-option>
                <bcs-option id="overlay" name="overlay"></bcs-option>
            </bcs-select>
        </bk-form-item>
        <bk-form-item :label="$t('所属区域')" property="region" error-display-type="normal" required>
            <bcs-select v-model="formData.region" :loading="regionLoading" searchable :clearable="false">
                <bcs-option v-for="item in regionList" :key="item.region" :id="item.region" :name="item.regionName"></bcs-option>
            </bcs-select>
        </bk-form-item>
        <bk-form-item :label="$t('所属VPC')" property="vpcID" error-display-type="normal" required>
            <bcs-select v-model="formData.vpcID" :loading="vpcLoading" searchable :clearable="false">
                <!-- VPC可用容器网络IP数量最低限制 -->
                <bcs-option v-for="item in vpcList"
                    :key="item.vpcID"
                    :id="item.vpcID"
                    :name="item.vpcName"
                    :disabled="environment === 'prod'
                        ? item.availableIPNum < 4096
                        : item.availableIPNum < 2048
                    "
                    v-bk-tooltips="{
                        content: $t('可用容器网络IP数量不足'),
                        disabled: environment === 'prod'
                            ? item.availableIPNum >= 4096
                            : item.availableIPNum >= 2048
                    }">
                    <div class="vpc-option">
                        <span>
                            {{item.vpcName}}
                            <span class="vpc-id">{{`(${item.vpcID})`}}</span>
                        </span>
                        <span class="vpc-ip">
                            {{item.availableIPNum}}
                        </span>
                    </div>
                </bcs-option>
            </bcs-select>
            <span v-if="curVpc">
                {{$t('可用容器网络IP {num} 个', { num: curVpc.availableIPNum })}}
            </span>
        </bk-form-item>
        <bk-form-item :label="$t('容器网络')" property="network" error-display-type="normal" required>
            <div class="container-network">
                <div class="container-network-item mr32">
                    <div>{{ $t('IP数量') }}</div>
                    <bcs-select class="w240" v-model="formData.networkSettings.cidrStep" :clearable="false">
                        <bcs-option v-for="ip in cidrStepList"
                            :key="ip"
                            :id="ip"
                            :name="ip"
                        ></bcs-option>
                        <template #extension>
                            <a :href="PROJECT_CONFIG.doc.contact"
                                class="bk-text-button"
                            >{{ $t('不满足需求，请联系蓝鲸容器助手') }}</a>
                        </template>
                    </bcs-select>
                </div>
                <div class="container-network-item">
                    <div>{{ $t('集群内Service数量上限') }}</div>
                    <bcs-select class="w240" v-model="formData.networkSettings.maxServiceNum" :clearable="false">
                        <bcs-option v-for="item in serviceIpNumList"
                            :key="item"
                            :id="item"
                            :name="item"></bcs-option>
                        <template #extension>
                            <a :href="PROJECT_CONFIG.doc.contact"
                                class="bk-text-button"
                            >{{ $t('不满足需求，请联系蓝鲸容器助手') }}</a>
                        </template>
                    </bcs-select>
                </div>
                <div class="container-network-item">
                    <div>{{ $t('单节点Pod数量上限') }}</div>
                    <bcs-select class="w240" v-model="formData.networkSettings.maxNodePodNum" :clearable="false">
                        <bcs-option v-for="item in nodePodNumList" :key="item" :id="item" :name="item"></bcs-option>
                    </bcs-select>
                </div>
                <i18n class="network-tips"
                    path="容器网络资源有限，请合理分配，当前容器网络配置下，集群最多可以添加 {count} 个节点">
                    <span place="count" class="count">{{ maxNodeCount }}</span>
                </i18n>
                <i18n class="network-tips"
                    path="当容器网络资源超额使用时，会触发容器网络自动扩容，扩容后集群最多可以添加 {count} 个节点">
                    <span place="count" class="count">{{ maxCapacityCount }}</span>
                </i18n>
                <div class="network-tips">{{$t('集群可添加节点数（包含Master节点与Node节点）= (IP数量 - Service的数量) / 单节点Pod数量上限')}}</div>
            </div>
        </bk-form-item>
    </bk-form>
</template>
<script lang="ts">
    import { computed, defineComponent, ref, toRefs, watch, set } from '@vue/composition-api'

    export default defineComponent({
        name: 'FormMode',
        props: {
            versionList: {
                type: Array,
                default: () => []
            },
            cloudId: {
                type: String,
                default: ''
            },
            cidrStepList: {
                type: Array,
                default: () => ([])
            },
            environment: {
                type: String,
                default: ''
            }
        },
        setup (props, ctx) {
            const { $store, $i18n } = ctx.root
            const formData = ref({
                clusterBasicSettings: {
                    version: ''
                },
                networkType: 'overlay',
                region: '',
                vpcID: '',
                networkSettings: {
                    cidrStep: '',
                    maxNodePodNum: '',
                    maxServiceNum: ''
                }
            })
            watch(formData, (newdata, oldData) => {
                ctx.emit('change', newdata, oldData)
            }, { deep: true })
            // 区域列表
            const regionList = ref([])
            const regionLoading = ref(false)
            const getRegionList = async () => {
                regionLoading.value = true
                regionList.value = await $store.dispatch('clustermanager/fetchCloudRegion', {
                    $cloudId: cloudId.value
                })
                regionLoading.value = false
            }
            const { cloudId, environment } = toRefs(props)
            watch(cloudId, () => {
                // 模板ID切换时重置表单数据
                formData.value.clusterBasicSettings.version = ''
                formData.value.region = ''
                formData.value.vpcID = ''
                getRegionList()
                getVpcList()
            })
            watch(environment, () => {
                formData.value.networkSettings.cidrStep = ''
                formData.value.vpcID = ''
            })
            watch(() => [formData.value.region, formData.value.networkType], () => {
                // 区域和网络类型变更时重置vpcId
                formData.value.vpcID = ''
                getVpcList()
            })
            // vpc列表
            const vpcList = ref<any[]>([])
            const vpcLoading = ref(false)
            const getVpcList = async () => {
                vpcLoading.value = true
                vpcList.value = await $store.dispatch('clustermanager/fetchCloudVpc', {
                    cloudID: cloudId.value,
                    region: formData.value.region,
                    networkType: formData.value.networkType
                })
                vpcLoading.value = false
            }
            const getIpNumRange = (minExponential, maxExponential) => {
                const list: number[] = []
                for (let i = minExponential; i <= maxExponential; i++) {
                    list.push(Math.pow(2, i))
                }
                return list
            }
            // 当前选择VPC
            const curVpc = computed(() => {
                return vpcList.value.find(item => item.vpcID === formData.value.vpcID)
            })
            // service ip选择列表
            const serviceIpNumList = computed(() => {
                const ipNumber = Number(formData.value.networkSettings.cidrStep)
                if (!ipNumber) return []

                const minExponential = Math.log2(128)
                const maxExponential = Math.log2(ipNumber / 2)

                return getIpNumRange(minExponential, maxExponential)
            })
            // pod数量列表
            const nodePodNumList = computed(() => {
                const ipNumber = Number(formData.value.networkSettings.cidrStep)
                const serviceNumber = Number(formData.value.networkSettings.maxServiceNum)
                if (!ipNumber || !serviceNumber) return []

                const minExponential = Math.log2(16)
                const maxExponential = Math.log2(Math.min(ipNumber - serviceNumber, 128))
                return getIpNumRange(minExponential, maxExponential)
            })
            // 网络配置信息
            watch(() => formData.value.vpcID, () => {
                // 重置网络配置
                set(formData.value, 'networkSettings', {
                    cidrStep: '',
                    maxNodePodNum: '',
                    maxServiceNum: ''
                })
            })
            // 表单校验
            const rules = ref({
                version: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                region: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                vpcID: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                network: [
                    {
                        validator: function (val) {
                            const network = formData.value.networkSettings
                            return Object.keys(network).every(key => !!network[key])
                        },
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ]
            })
            const formMode = ref<any>(null)
            const validate = async () => {
                const result = await formMode.value?.validate()
                return result
            }
            const getData = () => {
                return formData.value
            }
            const maxNodeCount = computed(() => {
                const { cidrStep, maxServiceNum, maxNodePodNum } = formData.value.networkSettings
                if (cidrStep && maxServiceNum && maxNodePodNum) {
                    return Math.floor((Number(cidrStep) - Number(maxServiceNum)) / Number(maxNodePodNum)) || 0
                }
                return 0
            })
            const maxCapacityCount = computed(() => {
                const { cidrStep, maxServiceNum, maxNodePodNum } = formData.value.networkSettings
                if (cidrStep && maxServiceNum && maxNodePodNum) {
                    return Math.floor((Number(cidrStep) * 5 - Number(maxServiceNum)) / Number(maxNodePodNum)) || 0
                }
                return 0
            })
            return {
                rules,
                formMode,
                formData,
                vpcList,
                vpcLoading,
                regionList,
                regionLoading,
                serviceIpNumList,
                nodePodNumList,
                validate,
                getData,
                maxNodeCount,
                maxCapacityCount,
                curVpc
            }
        }
    })
</script>
<style lang="postcss" scoped>
/deep/ .container-network {
    display: flex;
    flex-wrap: wrap;
    background-color: #F5F6FA;
    padding: 8px 16px;
    max-width: 600px;
    .container-network-item {
        margin-bottom: 16px;
        .w240 {
            width: 240px;
            background-color: #fff;
            border-color: #c4c6cc !important;
        }
    }
    .mr32 {
        margin-right: 32px;
    }
}
.network-tips {
    color: #979ba5;
    font-size: 12px;
    line-height: 1;
    margin-bottom: 10px;
    .count {
        color: #222;
    }
}
.vpc-option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    .vpc-id {
        color: #979BA5;
    }
    .vpc-ip {
        color: #979BA5;
        background: #F0F1F5;
        display: flex;
        align-items: center;
        justify-content: center;
        height: 16px;
        padding: 0 4px;
    }
}
</style>
