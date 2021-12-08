<template>
    <section class="create-form-cluster">
        <FormGroup :title="$t('基本信息')">
            <bk-form :label-width="100" :model="basicInfo" :rules="basicDataRules" ref="basicForm">
                <bk-form-item :label="$t('集群名称')" property="clusterName" error-display-type="normal" required>
                    <bk-input v-model="basicInfo.clusterName"></bk-input>
                </bk-form-item>
                <bk-form-item :label="$t('集群环境')" required>
                    <bk-radio-group v-model="basicInfo.environment">
                        <bk-radio value="debug">
                            {{ $t('测试环境') }}
                        </bk-radio>
                        <bk-radio value="prod">
                            {{ $t('正式环境') }}
                        </bk-radio>
                    </bk-radio-group>
                </bk-form-item>
                <bk-form-item :label="$t('模板名称')" property="provider" error-display-type="normal" required>
                    <div class="template-name">
                        <bcs-select v-model="basicInfo.provider" :clearable="false">
                            <bcs-option v-for="item in templateList"
                                :key="item.cloudID"
                                :id="item.cloudID"
                                :name="item.name">
                            </bcs-option>
                        </bcs-select>
                        <!-- <bk-button text class="ml10">{{ $t('新增集群模板') }}</bk-button> -->
                    </div>
                </bk-form-item>
                <bk-form-item :label="$t('描述')">
                    <bk-input v-model="basicInfo.description" type="textarea"></bk-input>
                </bk-form-item>
            </bk-form>
        </FormGroup>
        <FormGroup :title="$t('集群选项')" class="mt15">
            <!-- <template #title>
                <div class="bk-button-group">
                    <bk-button @click.native.stop="handleChangeMode('form')" :class="createMode === 'form' ? 'is-selected' : ''" size="small">{{ $t('表单结构') }}</bk-button>
                    <bk-button @click.native.stop="handleChangeMode('yaml')" :class="createMode === 'yaml' ? 'is-selected' : ''" size="small">{{ $t('Yaml格式') }}</bk-button>
                </div>
            </template> -->
            <template #default>
                <FormMode v-if="createMode === 'form'"
                    :version-list="versionList"
                    :cloud-id="basicInfo.provider"
                    ref="formMode"></FormMode>
                <YamlMode v-else></YamlMode>
            </template>
        </FormGroup>
        <FormGroup :title="$t('选择Master')" :desc="$t('仅支持数量为3,5和7个')" class="mt15">
            <template #title>
                <bk-button text v-if="ipList.length" @click.native.stop="handleOpenSelector">
                    <i class="bcs-icon bcs-icon-plus" style="position: relative;top: -1px;"></i>
                    {{ $t('选择服务器') }}
                </bk-button>
            </template>
            <div :class="['choose-server-btn', { 'error-btn': ipErrorTips }]" @click.stop="handleOpenSelector" v-if="!ipList.length">
                <i class="bcs-icon bcs-icon-plus mr5" style="position: relative;top: 1px;"></i>
                <span>{{ $t('选择服务器') }}</span>
            </div>
            <div class="choose-server-list" v-else>
                <bk-table :data="ipList">
                    <bk-table-column type="index" :label="$t('序列')" width="60"></bk-table-column>
                    <bk-table-column :label="$t('内网IP')" prop="bk_host_innerip"></bk-table-column>
                    <bk-table-column :label="$t('机房')" prop="idc_name"></bk-table-column>
                    <bk-table-column :label="$t('机型')" prop="svr_device_class"></bk-table-column>
                    <bk-table-column :label="$t('操作')" width="100">
                        <template #default="{ row }">
                            <bk-button text @click="handleDeleteIp(row)">{{ $t('移除') }}</bk-button>
                        </template>
                    </bk-table-column>
                </bk-table>
            </div>
            <p class="error-tips" v-if="ipErrorTips">{{ ipErrorTips }}</p>
        </FormGroup>
        <div class="footer">
            <bk-button class="btn" theme="primary" :loading="creating" @click="handleCreateCluster">{{$t('创建')}}</bk-button>
            <bk-button class="btn ml15" @click="handleCancel">{{$t('取消')}}</bk-button>
        </div>
        <IpSelector v-model="showIpSelector" @confirm="handleChooseServer"></IpSelector>
    </section>
</template>
<script lang="ts">
    import { computed, defineComponent, onMounted, ref } from '@vue/composition-api'
    import IpSelector from '@/components/ip-selector/selector-dialog.vue'
    import FormGroup from './form-group.vue'
    import FormMode from './form-mode.vue'
    import YamlMode from './yaml-mode.vue'
    import { TranslateResult } from 'vue-i18n'

    export default defineComponent({
        name: 'CreateFormCluster',
        components: {
            FormGroup,
            FormMode,
            YamlMode,
            IpSelector
        },
        setup (props, ctx) {
            const { $store, $i18n, $bkMessage, $router } = ctx.root
            const createMode = ref<'form' | 'yaml'>('form')
            const basicInfo = ref({
                clusterName: '', // 集群名称
                environment: 'debug', // 集群环境
                provider: '', // 云模板ID
                description: '' // 描述
            })
            // 更改集群选项模式
            const handleChangeMode = (mode) => {
                createMode.value = mode
            }
            const templateList = ref<any[]>([])
            const templateLoading = ref(false)
            const handleGetTemplateList = async () => {
                templateLoading.value = true
                templateList.value = await $store.dispatch('clustermanager/fetchCloudList')
                templateLoading.value = false
            }
            // 版本列表
            const versionList = computed(() => {
                const cloud = templateList.value.find(item => item.cloudID === basicInfo.value.provider)
                return cloud?.clusterManagement.availableVersion || []
            })

            const showIpSelector = ref(false)
            // 打开IP选择器
            const handleOpenSelector = () => {
                showIpSelector.value = true
            }
            // 选择IP节点
            const ipList = ref<any[]>([])
            const ipErrorTips = ref<TranslateResult>('')
            const validateServer = (data) => {
                if (!data.length) {
                    ipErrorTips.value = $i18n.t('必填项')
                    return false
                }
                if (![3, 5, 7].includes(data.length)) {
                    ipErrorTips.value = $i18n.t('仅支持数量为3,5和7个')
                    return false
                }
                return true
            }
            const handleChooseServer = (data = []) => {
                const validate = validateServer(data)
                if (!validate) {
                    showIpSelector.value = false
                    return
                }
                ipErrorTips.value = ''
                ipList.value = data
                showIpSelector.value = false
            }
            const handleDeleteIp = (row) => {
                const index = ipList.value.findIndex(item => item.bk_host_innerip === row.bk_host_innerip)
                index > -1 && ipList.value.splice(index, 1)
            }
            const basicForm = ref<any>(null)
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
                ]
            })
            const formMode = ref<any>(null)
            const creating = ref(false)
            // 创建集群
            const handleCreateCluster = async () => {
                await Promise.all([basicForm.value.validate(), formMode.value.validate()])

                const clusterData = formMode.value?.formData
                const validate = validateServer(ipList.value)
                if (!validate) return

                creating.value = true
                const result = await $store.dispatch('clustermanager/createCluster', {
                    ...basicInfo.value,
                    ...clusterData,
                    master: ipList.value.map(item => item.bk_host_innerip)
                })
                if (result) {
                    $bkMessage({
                        theme: 'success',
                        message: $i18n.t('创建成功')
                    })
                    $router.push({ name: 'clusterMain' })
                }
                creating.value = false
            }
            const handleCancel = async () => {
                $router.push({ name: 'clusterCreate' })
            }
            onMounted(() => {
                handleGetTemplateList()
            })
            return {
                creating,
                ipErrorTips,
                basicInfo,
                createMode,
                showIpSelector,
                ipList,
                templateList,
                versionList,
                basicForm,
                formMode,
                basicDataRules,
                handleChangeMode,
                handleOpenSelector,
                handleChooseServer,
                handleDeleteIp,
                handleCreateCluster,
                handleCancel
            }
        }
    })
</script>
<style lang="postcss" scoped>
.create-form-cluster {
    padding: 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    /deep/ .bk-input-text {
        width: 400px;
    }
    /deep/ .bk-select {
        width: 400px;
    }
    /deep/ .bk-textarea-wrapper {
        width: 400px;
    }
    /deep/ .form-group {
        width: 80%;
        max-width: 1000px;
    }
    .template-name {
        display: flex;
    }
    .cluster-config {
        display: flex;
        align-items: center;
        justify-content: space-between;
        width: 100%;
        .title {
            font-size: 14px;
            font-weight: 700;
            line-height: 22px;
        }
    }
    .choose-server-btn {
        border: 1px dashed #c4c6cc;
        height: 40px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #3A84FF;
        width: 630px;
        margin-left: 24px;
        &:hover {
            border-color: #3A84FF;
            cursor: pointer;
        }
    }
    .choose-server-list {
        padding-left: 24px;
    }
    .footer {
        width: 80%;
        max-width: 1000px;
        .btn {
            margin-top: 20px;
            width: 88px;
        }
    }
    .error-btn {
        border: 1px solid #ea3636;
        color: #ea3636;
    }
    .error-tips {
        margin-top: 5px;
        color: #ea3636;
        padding: 0 24px;
    }
}
</style>
