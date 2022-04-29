<template>
    <section class="create-form-cluster">
        <bk-form :label-width="labelWidth" :model="basicInfo" :rules="basicDataRules" ref="basicFormRef">
            <bk-form-item :label="$t('集群名称')" property="clusterName" error-display-type="normal" required>
                <bk-input class="w640" v-model="basicInfo.clusterName"></bk-input>
            </bk-form-item>
            <bk-form-item :label="$t('云服务商')" property="provider" error-display-type="normal" required>
                <bcs-select :loading="templateLoading" class="w640" v-model="basicInfo.provider" :clearable="false">
                    <bcs-option v-for="item in templateList"
                        :key="item.cloudID"
                        :id="item.cloudID"
                        :name="item.name">
                    </bcs-option>
                </bcs-select>
            </bk-form-item>
            <bk-form-item :label="$t('版本')" property="clusterBasicSettings.version" error-display-type="normal" required>
                <bcs-select class="w640" v-model="basicInfo.clusterBasicSettings.version" searchable :clearable="false">
                    <bcs-option v-for="item in versionList" :key="item" :id="item" :name="item"></bcs-option>
                </bcs-select>
            </bk-form-item>
            <bk-form-item :label="$t('集群描述')">
                <bk-input v-model="basicInfo.description" type="textarea"></bk-input>
            </bk-form-item>
            <bk-form-item :label="$t('附加参数')" ref="extraInfoRef" v-show="expanded">
                <KeyValue
                    class="w700"
                    :show-footer="false"
                    :show-header="false"
                    :key-advice="keyAdvice"
                    ref="keyValueRef"
                ></KeyValue>
            </bk-form-item>
            <bk-form-item>
                <div class="action" @click="toggleSettings">
                    <i :class="['bk-icon', expanded ? 'icon-angle-double-up' : 'icon-angle-double-down']"></i>
                    <span>{{ expanded ? $t('收起更多设置') : $t('展开更多设置')}}</span>
                </div>
            </bk-form-item>
            <bk-form-item :label="$t('选择Master')" property="ipList" error-display-type="normal" required>
                <bk-button @click="handleShowIpSelector">
                    <i class="bcs-icon bcs-icon-plus" style="position: relative;top: -1px;"></i>
                    {{$t('选择服务器')}}
                </bk-button>
                <bk-table class="ip-list mt10" :data="basicInfo.ipList" v-if="basicInfo.ipList.length">
                    <bk-table-column type="index" :label="$t('序列')" width="60"></bk-table-column>
                    <bk-table-column :label="$t('内网IP')" prop="bk_host_innerip"></bk-table-column>
                    <bk-table-column :label="$t('机房')" prop="idc_name"></bk-table-column>
                    <bk-table-column :label="$t('机型')" prop="svr_device_class"></bk-table-column>
                    <bk-table-column :label="$t('操作')" width="100">
                        <template #default="{ row }">
                            <bcs-button text @click="handleRemoveServer(row)">{{$t('移除')}}</bcs-button>
                        </template>
                    </bk-table-column>
                </bk-table>
            </bk-form-item>
            <bk-form-item>
                <bk-button class="btn" theme="primary" @click="handleShowConfirmDialog">{{$t('确定')}}</bk-button>
                <bk-button class="btn" @click="handleCancel">{{$t('取消')}}</bk-button>
            </bk-form-item>
        </bk-form>
        <bcs-dialog v-model="confirmDialog"
            theme="primary"
            header-position="left"
            :title="$t('确定创建集群')"
            width="640">
            <div class="create-cluster-dialog">
                <div class="title">{{$t('请确认以下配置')}}:</div>
                <bcs-checkbox-group class="confirm-wrapper" v-model="createConfirm">
                    <bcs-checkbox value="1">{{$t('服务器将按照系统规则修改主机名')}}</bcs-checkbox>
                    <bcs-checkbox value="2" class="mt10">{{$t('服务器将安装容器服务相关组件')}}</bcs-checkbox>
                </bcs-checkbox-group>
            </div>
            <template #footer>
                <div>
                    <bcs-button :disabled="createConfirm.length < 2"
                        theme="primary"
                        :loading="loading"
                        @click="handleCreateCluster"
                    >{{$t('确定，创建集群')}}</bcs-button>
                    <bcs-button @click="confirmDialog = false">{{$t('我再想想')}}</bcs-button>
                </div>
            </template>
        </bcs-dialog>
        <IpSelector v-model="showIpSelector" @confirm="handleChooseServer"></IpSelector>
    </section>
</template>
<script lang="ts">
    import { defineComponent, onMounted, ref, computed, watch } from '@vue/composition-api'
    import IpSelector from '@/components/ip-selector/selector-dialog.vue'
    import useGoHome from '@/common/use-gohome'
    import KeyValue from '@/components/key-value.vue'
    import useFormLabel from '@/common/use-form-label'

    export default defineComponent({
        name: 'CreateFormCluster',
        components: {
            IpSelector,
            KeyValue
        },
        setup (props, ctx) {
            const { $i18n, $route, $bkMessage, $store } = ctx.root
            const { goHome } = useGoHome()
            const basicFormRef = ref<any>(null)
            const basicInfo = ref<{
                clusterName: string;
                description: string;
                provider: string;
                clusterBasicSettings: {
                    version: string;
                };
                ipList: any[];
            }>({
                    clusterName: '',
                    description: '',
                    provider: '',
                    clusterBasicSettings: {
                        version: ''
                    },
                    ipList: []
                })
            const templateList = ref<any[]>([])
            const versionList = computed(() => {
                const cloud = templateList.value.find(item => item.cloudID === basicInfo.value.provider)
                return cloud?.clusterManagement.availableVersion || []
            })
            const basicDataRules = ref({
                clusterName: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                provider: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                'clusterBasicSettings.version': [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                ipList: [
                    {
                        message: $i18n.t('仅支持奇数个服务器'),
                        validator (val) {
                            return val.length % 2 !== 0
                        },
                        trigger: 'blur'
                    }
                ]
            })
            const showIpSelector = ref(false)
            const confirmDialog = ref(false)
            const createConfirm = ref([])
            const loading = ref(false)
            const expanded = ref(false)
            const templateLoading = ref(false)
            const keyAdvice = ref([
                {
                    name: 'DOCKER_LIB',
                    desc: $i18n.t('Docker数据目录'),
                    default: ''
                },
                {
                    name: 'DOCKER_VERSION',
                    desc: $i18n.t('Docker版本'),
                    default: '19.03.9'
                },
                {
                    name: 'KUBELET_LIB',
                    desc: $i18n.t('kubelet数据目录'),
                    default: ''
                },
                {
                    name: 'K8S_VER',
                    desc: $i18n.t('集群版本'),
                    default: ''
                },
                {
                    name: 'K8S_SVC_CIDR',
                    desc: $i18n.t('集群Service网段'),
                    default: ''
                },
                {
                    name: 'K8S_POD_CIDR',
                    desc: $i18n.t('集群Pod网段'),
                    default: ''
                }
            ])
            
            const handleShowIpSelector = () => {
                showIpSelector.value = true
            }
            const handleChooseServer = (data) => {
                data.forEach(item => {
                    const index = basicInfo.value.ipList.findIndex(ipData => ipData.bk_cloud_id === item.bk_cloud_id
                        && ipData.bk_host_innerip === item.bk_host_innerip)
                    if (index === -1) {
                        basicInfo.value.ipList.push(item)
                    }
                })
                showIpSelector.value = false
            }
            const handleShowConfirmDialog = async () => {
                const result = await basicFormRef.value?.validate()
                if (!result) return
                confirmDialog.value = true
            }
            watch(confirmDialog, (value) => {
                if (!value) {
                    createConfirm.value = []
                }
            })
            const curProject = computed(() => {
                return $store.state.curProject
            })
            const user = computed(() => {
                return $store.state.user
            })
            const handleCreateCluster = async () => {
                loading.value = true
                confirmDialog.value = false
                loading.value = false
                const extraInfo = basicFormRef.value?.$refs?.extraInfoRef?.$refs?.keyValueRef?.labels || {}
                const params = {
                    environment: 'prod',
                    projectID: curProject.value.project_id,
                    businessID: String(curProject.value.cc_app_id),
                    engineType: 'k8s',
                    isExclusive: true,
                    clusterType: 'single',
                    creator: user.value.username,
                    manageType: 'INDEPENDENT_CLUSTER',
                    clusterName: basicInfo.value?.clusterName,
                    description: basicInfo.value?.description,
                    provider: basicInfo.value?.provider,
                    region: 'default',
                    vpcID: '',
                    clusterBasicSettings: basicInfo.value?.clusterBasicSettings,
                    networkType: 'overlay',
                    extraInfo: {
                        create_cluster: Object.keys(extraInfo).reduce((pre, key) => {
                            pre += `${key}=${extraInfo[key]};`
                            return pre
                        }, '')
                    },
                    networkSettings: {},
                    master: basicInfo.value.ipList.map((item: any) => item.bk_host_innerip)
                }
                const result = await $store.dispatch('clustermanager/createCluster', params)
                if (result) {
                    $bkMessage({
                        theme: 'success',
                        message: $i18n.t('创建成功')
                    })
                    goHome($route)
                }
            }
            const handleCancel = () => {
                goHome($route)
            }
            const toggleSettings = () => {
                expanded.value = !expanded.value
            }
            const handleGetTemplateList = async () => {
                templateLoading.value = true
                templateList.value = await $store.dispatch('clustermanager/fetchCloudList')
                basicInfo.value.provider = templateList.value[0]?.cloudID || ''
                templateLoading.value = false
            }
            const handleRemoveServer = async (row) => {
                const index = basicInfo.value.ipList.findIndex(item => item.bk_cloud_id === row.bk_cloud_id
                    && item.bk_host_innerip === row.bk_host_innerip)
                if (index > -1) {
                    basicInfo.value.ipList.splice(index, 1)
                }
            }
            const { labelWidth, initFormLabelWidth } = useFormLabel()
            onMounted(() => {
                handleGetTemplateList()
                initFormLabelWidth(basicFormRef.value)
            })
            return {
                labelWidth,
                keyAdvice,
                templateLoading,
                expanded,
                loading,
                templateList,
                versionList,
                basicFormRef,
                basicInfo,
                showIpSelector,
                basicDataRules,
                confirmDialog,
                createConfirm,
                handleChooseServer,
                handleShowIpSelector,
                handleShowConfirmDialog,
                handleCreateCluster,
                handleCancel,
                toggleSettings,
                handleGetTemplateList,
                handleRemoveServer
            }
        }
    })
</script>
<style lang="postcss" scoped>
.create-form-cluster {
    padding: 24px;
    /deep/ .bk-textarea-wrapper {
        width: 640px;
        height: 80px;
    }
    /deep/ .w640 {
        width: 640px;
        background: #fff;
    }
    /deep/ .w700 {
        width: 705px;
    }
    /deep/ .btn {
        width: 88px;
    }
    /deep/ .ip-list {
        max-width: 1000px;
    }
    /deep/ .action {
        i {
            font-size: 20px;
        }
        align-items: center;
        color: #3a84ff;
        cursor: pointer;
        display: flex;
        font-size: 14px;
    }
}
.create-cluster-dialog {
    padding: 0 4px 24px 4px;
    .title {
        font-size: 14px;
        font-weight: 700;
        margin-bottom: 14px
    }
    .confirm-wrapper {
        display: flex;
        flex-direction: column;
    }
}
</style>
