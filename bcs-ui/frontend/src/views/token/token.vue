<template>
    <div class="user-token">
        <div class="user-token-header">
            <span>
                <i class="bcs-icon bcs-icon-arrows-left back" @click="goBack"></i>
                <span class="title">{{$t('API密钥')}}</span>
            </span>
            <a class="bk-text-button help" :href="PROJECT_CONFIG.doc.token" target="_blank">
                {{ $t('使用说明') }}
            </a>
        </div>
        <bcs-alert type="info" class="mb15">
            <template #title>
                <div class="info-item">1. {{$t('API密钥适用于 BCS API 调用与 kubeconfig')}}</div>
                <div class="info-item">2.
                    <i18n path="API密钥与个人账户绑定，使用蓝鲸权限中心做权限控制，点击{0}可以查看与设置您的API密钥权限">
                        <a class="bk-text-button" :href="BK_IAM_APP_URL" target="_blank">
                            {{ $t('权限中心') }}
                        </a>
                    </i18n>
                </div>
                <div class="info-item">3. {{$t('为了您的应用安全，请妥善保存API密钥，请勿通过任何方式上传或者分享您的API密钥信息')}}</div>
            </template>
        </bcs-alert>
        <bk-table :data="data" v-bkloading="{ isLoading: loading }">
            <bk-table-column :label="$t('用户名')">
                <template #default>
                    <span>{{user.username}}</span>
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('API密钥')" min-width="300">
                <template #default="{ row }">
                    <div class="token-row">
                        <span>{{hiddenToken ? new Array(row.token.length).fill('*').join('') : row.token}}</span>
                        <i :class="['ml10 bcs-icon', `bcs-icon-${hiddenToken ? 'eye-slash' : 'eye'}`]"
                            @click="toggleHiddenToken"></i>
                        <i class="ml10 bcs-icon bcs-icon-copy" @click="handleCopyToken(row)"></i>
                    </div>
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('过期时间')" prop="expired_at">
                <template #default="{ row }">
                    <div>{{!row.expired_at ? $t('永久') : row.expired_at}}</div>
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('状态')">
                <template #default="{ row }">
                    <StatusIcon :status="String(row.status)"
                        :status-color-map="{
                            '1': 'green',
                            '0': 'gray'
                        }">
                        {{row.status === 1 ? $t('正常') : $t('已过期')}}
                    </StatusIcon>
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('操作')" width="150">
                <template #default="{ row }">
                    <bk-button
                        class="mr10"
                        theme="primary"
                        text
                        :disabled="!row.expired_at"
                        @click="handleRenewalToken(row)"
                    >{{$t('续期')}}</bk-button>
                    <bk-button theme="primary" text
                        @click="handleDeleteToken(row)">{{$t('删除')}}</bk-button>
                </template>
            </bk-table-column>
            <template #empty>
                <bcs-exception type="empty" scene="part">
                    <div>{{$t('您暂无当前操作权限，请新建API密钥后继续操作')}}</div>
                    <bcs-button class="create-token-btn" icon="plus" theme="primary"
                        @click="handleCreateToken">
                        {{$t('新建API密钥')}}
                    </bcs-button>
                </bcs-exception>
            </template>
        </bk-table>
        <!-- 使用案例 -->
        <div class="user-token-example" v-if="data.length">
            <div class="example-item">
                <div class="title">{{$t('Kubeconfig使用示例')}}:</div>
                <div class="code-wrapper">
                    {{kubeConfigExample}}
                </div>
            </div>
            <div class="example-item">
                <div class="title">{{$t('/root/.kube/demo_config内容示例如下')}}:</div>
                <div class="code-wrapper">
                    <ace :show-gutter="false" lang="yaml" :height="320"
                        v-full-screen="{ tools: ['copy'], content: demoConfigExample }"
                        read-only :value="demoConfigExample" width="100%">
                    </ace>
                </div>
            </div>
            <div class="example-item">
                <div class="title">{{$t('BCS API使用示例')}}:</div>
                <div class="code-wrapper">
                    <ace :show-gutter="false" :height="80"
                        v-full-screen="{ tools: ['copy'], content: bcsApiExample }"
                        read-only :value="bcsApiExample" width="100%">
                    </ace>
                </div>
            </div>
        </div>
        <!-- 创建 or 续期 token -->
        <bcs-dialog v-model="showTokenDialog"
            theme="primary"
            :mask-close="false"
            header-position="left"
            :title="operateType === 'create' ? $t('新建API密钥') : $t('续期API密钥')"
            width="640"
            :loading="updateLoading"
            @confirm="confirmUpdateTokenDialog"
            @cancel="cancelUpdateTokenDialog">
            <div class="create-token-dialog">
                <div class="title">{{$t('申请期限')}}</div>
                <div class="bk-button-group">
                    <bk-button v-for="item in timeList"
                        :key="item.id"
                        :class="['group-btn', { 'is-selected': item.id === active }]"
                        @click="handleSelectTime(item)"
                    >
                        {{item.name}}
                    </bk-button>
                    <bcs-input
                        type="number"
                        :min="1"
                        :precision="0"
                        :show-controls="false"
                        class="custom-input"
                        ref="customInputRef"
                        v-model="active"
                        v-if="isCustomTime">
                        <template slot="append">
                            <div class="custom-input-append">{{$t('天')}}</div>
                        </template>
                    </bcs-input>
                    <bk-button class="group-btn" v-else @click="handleCustomTime">{{$t('自定义')}}</bk-button>
                </div>
            </div>
        </bcs-dialog>
        <!-- 删除 token -->
        <bcs-dialog v-model="showDeleteDialog"
            theme="primary"
            header-position="left"
            :title="$t('删除该API密钥')"
            width="640">
            <div class="delete-token-dialog">
                <div class="title">{{$t('此操作无法撤回，请确认')}}:</div>
                <bcs-checkbox v-model="deleteConfirm">
                    {{$t('所有使用API密钥的 API 接口与 kubeconfig 将无法使用')}}
                </bcs-checkbox>
            </div>
            <template #footer>
                <div>
                    <bcs-button :disabled="!deleteConfirm"
                        theme="primary"
                        :loading="deleteLoading"
                        @click="confirmDeleteToken"
                    >{{$t('确定')}}</bcs-button>
                    <bcs-button @click="cancelDeleteDialog">{{$t('取消')}}</bcs-button>
                </div>
            </template>
        </bcs-dialog>
    </div>
</template>
<script lang="ts">
    import { defineComponent, ref, computed, onMounted } from '@vue/composition-api'
    import StatusIcon from '@/views/dashboard/common/status-icon'
    import { copyText } from '@/common/util'
    import * as ace from '@/components/ace-editor'
    import fullScreen from '@/directives/full-screen'
    import yamljs from 'js-yaml'
    import demoConfig from './demo-config'

    export default defineComponent({
        components: { StatusIcon, ace },
        directives: {
            'full-screen': fullScreen
        },
        setup (props, ctx) {
            const { $router, $i18n, $store, $bkMessage } = ctx.root
            const goBack = () => {
                $router.back()
            }
            // 用户信息
            const user = computed(() => {
                return $store.state.user
            })
            // 使用案例
            const kubeConfigExample = ref('kubectl --kubeconfig=/root/.kube/demo_config get node')
            const demoConfigExample = ref(yamljs.dump(demoConfig)
                .replace(new RegExp(/\$\{username\}/, 'g'), user.value.username)
                .replace(new RegExp(/\$\{token\}/, 'g'), '${' + $i18n.t('API密钥') + '}')
                .replace(new RegExp(/\$\{bcs_api_host\}/, 'g'), window.BCS_API_HOST))
            const bcsApiExample = ref('curl -X GET -H "Authorization: Bearer ${token}" -H "accept: application/json" "${bcs_api_host}/clusters/${cluster_id}/version"'
                .replace(new RegExp(/\$\{token\}/, 'g'), '${' + $i18n.t('API密钥') + '}')
                .replace(new RegExp(/\$\{bcs_api_host\}/, 'g'), window.BCS_API_HOST))
            
            const timeList = ref([
                {
                    id: -1,
                    name: $i18n.t('永久')
                },
                {
                    id: 30,
                    name: $i18n.t('{num}天', { num: 30 })
                },
                {
                    id: 3 * 30,
                    name: $i18n.t('{num}天', { num: 3 * 30 })
                },
                {
                    id: 6 * 30,
                    name: $i18n.t('{num}天', { num: 6 * 30 })
                },
                {
                    id: 365,
                    name: $i18n.t('{num}天', { num: 365 })
                }
            ])
            const active = ref(-1)
            const activeTimestamp = computed(() => {
                return active.value === -1 ? active.value : active.value * 24 * 60 * 60
            })
            
            // 自定义时间
            const isCustomTime = ref(false)
            const customInputRef = ref<any>(null)
            const handleSelectTime = (item) => {
                active.value = item.id
                isCustomTime.value = false
            }
            const handleCustomTime = () => {
                isCustomTime.value = true
                setTimeout(() => {
                    customInputRef.value.focus()
                    active.value = 1
                }, 0)
            }
            
            const hiddenToken = ref(true)
            // 隐藏Token
            const toggleHiddenToken = () => {
                hiddenToken.value = !hiddenToken.value
            }
            // 复制Token
            const handleCopyToken = (row) => {
                copyText(row.token)
                $bkMessage({
                    theme: 'success',
                    message: $i18n.t('复制成功')
                })
            }
            // token操作
            const operateType = ref<'create' | 'edit'>('create')
            const showTokenDialog = ref(false)
            const showDeleteDialog = ref(false)
            // 新建token事件
            const handleCreateToken = () => {
                showTokenDialog.value = true
                operateType.value = 'create'
            }
            // 取消更新token事件
            const cancelUpdateTokenDialog = () => {
                curEditRow.value = null
                active.value = -1
                isCustomTime.value = false
            }
            const curEditRow = ref<any>(null)
            // 续期Token事件
            const handleRenewalToken = (row) => {
                showTokenDialog.value = true
                operateType.value = 'edit'
                curEditRow.value = row
            }
            // 删除Token事件
            const handleDeleteToken = async (row) => {
                showDeleteDialog.value = true
                curEditRow.value = row
            }
            // 删除确认复选框
            const deleteConfirm = ref(false)
            // 取消删除事件
            const cancelDeleteDialog = () => {
                curEditRow.value = null
                deleteConfirm.value = false
                showDeleteDialog.value = false
            }
            // token列表
            const loading = ref(false)
            const data = ref([])
            const getTokenList = async () => {
                loading.value = true
                data.value = await $store.dispatch('token/getTokens', {
                    $username: user.value.username
                })
                loading.value = false
            }
            
            // 创建或者续期Token
            const updateLoading = ref(false)
            const confirmUpdateTokenDialog = async () => {
                updateLoading.value = true
                if (operateType.value === 'create') {
                    const result = await $store.dispatch('token/createToken', {
                        username: user.value.username,
                        expiration: activeTimestamp.value
                    })
                    result && $bkMessage({
                        theme: 'success',
                        message: $i18n.t('创建成功')
                    })
                } else if (operateType.value === 'edit' && curEditRow.value) {
                    const result = await $store.dispatch('token/updateToken', {
                        $token: curEditRow.value.token,
                        expiration: activeTimestamp.value
                    })
                    result && $bkMessage({
                        theme: 'success',
                        message: $i18n.t('续期成功')
                    })
                }
                updateLoading.value = false
                showTokenDialog.value = false
                cancelUpdateTokenDialog()
                getTokenList()
            }
            const deleteLoading = ref(false)
            const confirmDeleteToken = async () => {
                deleteLoading.value = true
                const result = await $store.dispatch('token/deleteToken', {
                    $token: curEditRow.value.token
                })
                deleteLoading.value = false
                showDeleteDialog.value = false
                cancelDeleteDialog()
                result && $bkMessage({
                    theme: 'success',
                    message: $i18n.t('删除成功')
                })
                getTokenList()
            }

            onMounted(() => {
                getTokenList()
            })
            return {
                deleteLoading,
                updateLoading,
                loading,
                data,
                deleteConfirm,
                showTokenDialog,
                showDeleteDialog,
                timeList,
                active,
                isCustomTime,
                customInputRef,
                user,
                hiddenToken,
                operateType,
                goBack,
                handleRenewalToken,
                handleDeleteToken,
                handleCreateToken,
                confirmUpdateTokenDialog,
                handleSelectTime,
                handleCustomTime,
                toggleHiddenToken,
                handleCopyToken,
                confirmDeleteToken,
                cancelDeleteDialog,
                cancelUpdateTokenDialog,
                kubeConfigExample,
                demoConfigExample,
                bcsApiExample,
                BK_IAM_APP_URL: window.BK_IAM_APP_URL
            }
        }
    })
</script>
<style lang="postcss" scoped>
.user-token {
    padding: 16px 84px;
}
.info-item {
    line-height: 20px;
}
.user-token-header {
    font-size: 16px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    .back {
        cursor: pointer;
        font-weight: 700;
        color: #3a84ff;
    }
    .title {
        margin-left: 8px;
    }
    .help {
        font-size: 14px;
    }
}
.token-row {
    display: flex;
    align-items: center;
    i {
        cursor: pointer;
    }
}
.create-token-btn {
    margin-top: 16px;
}
.create-token-dialog {
    padding: 0 4px 24px 4px;
    .title {
        font-size: 14px;
        margin-bottom: 14px;
    }
    .bk-button-group {
        display: flex;
    }
    .group-btn {
        min-width: 80px;
    }
    .custom-input {
        width: 80px;
        margin-left: -1px;
        >>> input {
            padding: 0 4px !important;
        }
        &-append {
            width: 28px;
            height: 32px;
            font-size: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
        }
    }
}
.delete-token-dialog {
    padding: 0 4px 24px 4px;
    .title {
        font-size: 14px;
        font-weight: 700;
        margin-bottom: 14px
    }
}
.user-token-example {
    margin-top: 24px;
    .example-item {
        margin-bottom: 24px;
        .title {
            margin-bottom: 12px;
            font-weight: 400;
            text-align: left;
            color: #313238;
            font-size: 14px;
        }
        .code-wrapper {
            font-size: 14px;
        }
    }
}
</style>
