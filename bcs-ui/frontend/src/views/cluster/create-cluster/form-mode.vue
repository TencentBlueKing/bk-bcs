<template>
    <bk-form :label-width="100" :model="formData" :rules="rules" ref="formMode">
        <bk-form-item :label="$t('版本')" property="version" error-display-type="normal" required>
            <bcs-select v-model="formData.clusterBasicSettings.version" :clearable="false">
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
            <bcs-select v-model="formData.region" :loading="regionLoading" :clearable="false">
                <bcs-option v-for="item in regionList" :key="item.region" :id="item.region" :name="item.regionName"></bcs-option>
            </bcs-select>
        </bk-form-item>
        <bk-form-item :label="$t('所属VPC')" property="vpcID" error-display-type="normal" required>
            <bcs-select v-model="formData.vpcID" :loading="vpcLoading" :clearable="false">
                <bcs-option v-for="item in vpcList"
                    :key="item.vpcID"
                    :id="item.vpcID"
                    :name="`${item.vpcName} (${item.vpcID})`"></bcs-option>
            </bcs-select>
        </bk-form-item>
        <bk-form-item :label="$t('容器网络')" property="network" error-display-type="normal" required>
            <div class="container-network">
                <div class="container-network-item mr32">
                    <div>{{ $t('IP数量') }}</div>
                    <bcs-select class="w240" v-model="formData.networkSettings.clusterIPv4CIDR" :clearable="false">
                        <bcs-option v-for="item in duplicateVpcCidrList" :key="item.cidr" :id="item.cidr" :name="item.IPNumber"></bcs-option>
                    </bcs-select>
                </div>
                <div class="container-network-item">
                    <div>{{ $t('Service数量上限/集群') }}</div>
                    <bcs-select class="w240" v-model="formData.networkSettings.maxServiceNum" :clearable="false">
                        <bcs-option v-for="item in serviceIpNumList" :key="item" :id="item" :name="item"></bcs-option>
                    </bcs-select>
                </div>
                <div class="container-network-item">
                    <div>{{ $t('Pod数量上限/节点') }}</div>
                    <bcs-select class="w240" v-model="formData.networkSettings.maxNodePodNum" :clearable="false">
                        <bcs-option v-for="item in nodePodNumList" :key="item" :id="item" :name="item"></bcs-option>
                    </bcs-select>
                </div>
                <div class="network-tips">{{ $t('计算规则: (IP数量-Service的数量)/(Master数量+Node数量)') }}</div>
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
                    clusterIPv4CIDR: '',
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
            const { cloudId } = toRefs(props)
            watch(cloudId, () => {
                // 模板ID切换时重置表单数据
                formData.value.clusterBasicSettings.version = ''
                formData.value.region = ''
                formData.value.vpcID = ''
                getRegionList()
                getVpcList()
            })
            watch(() => [formData.value.region, formData.value.networkType], () => {
                // 区域和网络类型变更时重置vpcId
                formData.value.vpcID = ''
                getVpcList()
            })
            // vpc列表
            const vpcList = ref([])
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
            // service ip选择列表
            const serviceIpNumList = computed(() => {
                const ipNumber = vpcCidrList.value.find(item =>
                    item.cidr === formData.value.networkSettings.clusterIPv4CIDR)?.IPNumber
                if (!ipNumber) return []

                const minExponential = Math.log2(32)
                const maxExponential = Math.log2(ipNumber / 2)

                return getIpNumRange(minExponential, maxExponential)
            })
            // pod数量列表
            const nodePodNumList = computed(() => {
                const ipNumber = vpcCidrList.value.find(item =>
                    item.cidr === formData.value.networkSettings.clusterIPv4CIDR)?.IPNumber
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
                    clusterIPv4CIDR: '',
                    maxNodePodNum: '',
                    maxServiceNum: ''
                })
                getVpccidrList()
            })
            const vpcCidrList = ref<any[]>([])
            const duplicateVpcCidrList = computed(() => {
                const IPNumbers: number[] = []
                return vpcCidrList.value.filter(item => {
                    if (!IPNumbers.includes(item.IPNumber)) {
                        IPNumbers.push(item.IPNumber)
                        return true
                    }
                    return false
                })
            })
            const getVpccidrList = async () => {
                if (!formData.value.vpcID) return

                vpcCidrList.value = await $store.dispatch('clustermanager/fetchVpccidrList', {
                    $vpcID: formData.value.vpcID
                })
            }
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
            return {
                rules,
                formMode,
                formData,
                vpcList,
                vpcLoading,
                regionList,
                duplicateVpcCidrList,
                regionLoading,
                vpcCidrList,
                serviceIpNumList,
                nodePodNumList,
                validate,
                getData
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
    width: 544px;
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
    margin-top: -15px;
}
</style>
