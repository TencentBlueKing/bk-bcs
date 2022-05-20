<template>
    <div class="tencent-cloud">
        <div class="tencent-cloud-operate">
            <bk-button theme="primary" @click="handleShowCreateDialog">{{$t('新建凭证')}}</bk-button>
            <bk-input class="w400"
                :placeholder="$t('搜索名称、SecretID、集群ID')"
                clearable
                v-model="searchValue">
            </bk-input>
        </div>
        <bcs-table class="mt20"
            :data="curPageData"
            :pagination="pageConf"
            v-bkloading="{ isLoading: loading }"
            @page-change="pageChange"
            @page-limit-change="pageSizeChange">
            <bcs-table-column :label="$t('名称')">
                <template #default="{ row }">
                    {{ row.account.accountName }}
                </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('描述')">
                <template #default="{ row }">
                    {{ row.account.desc }}
                </template>
            </bcs-table-column>
            <bcs-table-column label="SecretID">
                <template #default="{ row }">
                    {{ row.account.account.secretID }}
                </template>
            </bcs-table-column>
            <bcs-table-column label="SecretKey">
                <template #default="{ row }">
                    {{ row.account.account.secretKey }}
                </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('关联集群')">
                <template #default="{ row }">
                    {{ row.clusters.join(',') || '--' }}
                </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('创建时间')">
                <template #default="{ row }">
                    {{ row.account.updateTime }}
                </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('操作')" width="100">
                <template #default="{ row }">
                    <bk-button text @click="handleDeleteAccount(row)">{{$t('删除')}}</bk-button>
                </template>
            </bcs-table-column>
        </bcs-table>
        <bcs-dialog v-model="showDialog"
            theme="primary"
            :mask-close="false"
            header-position="left"
            width="600"
            :title="$t('新建凭证')">
            <bk-form :label-width="100" :model="account" :rules="formRules" ref="formRef">
                <bk-form-item :label="$t('名称')" property="accountName" error-display-type="normal" required>
                    <bk-input v-model="account.accountName"></bk-input>
                </bk-form-item>
                <bk-form-item :label="$t('描述')">
                    <bk-input type="textarea" v-model="account.desc"></bk-input>
                </bk-form-item>
                <bk-form-item label="SecretID" property="account.secretID" error-display-type="normal" required>
                    <bk-input v-model="account.account.secretID"></bk-input>
                </bk-form-item>
                <bk-form-item label="SecretKey" property="account.secretKey" error-display-type="normal" required>
                    <bk-input v-model="account.account.secretKey"></bk-input>
                </bk-form-item>
            </bk-form>
            <template #footer>
                <div>
                    <bk-button theme="primary"
                        :loading="createLoading"
                        @click="handleCreateAccount">
                        {{ $t('确定') }}
                    </bk-button>
                    <bk-button @click="() => showDialog = false">{{ $t('取消') }}</bk-button>
                </div>
            </template>
        </bcs-dialog>
    </div>
</template>
<script lang="ts">
    import { defineComponent, onMounted, ref, computed } from '@vue/composition-api'
    import $store from '@/store'
    import $i18n from '@/i18n/i18n-setup'
    import usePage from '@/views/dashboard/common/use-page'
    import useTableSearch from '@/views/dashboard/common/use-search'

    export default defineComponent({
        setup (props, ctx) {
            const { $bkMessage, $bkInfo } = ctx.root
            const curProject = computed(() => {
                return $store.state.curProject
            })
            const user = computed(() => {
                return $store.state.user
            })

            const loading = ref(false)
            const createLoading = ref(false)
            const showDialog = ref(false)
            const data = ref([])
            const formRef = ref<any>()
            const account = ref({
                "accountName": "",
                "desc": "",
                "account": {
                    "secretID": "",
                    "secretKey": ""
                },
                "enable": true,
                "creator": user.value.username,
                "projectID": curProject.value.project_id
            })
            const formRules = ref({
                accountName: [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                'account.secretID': [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ],
                'account.secretKey': [
                    {
                        required: true,
                        message: $i18n.t('必填项'),
                        trigger: 'blur'
                    }
                ]
            })
            const keys = ref(['account.accountName', 'account.account.secretID', 'clusters']) // 模糊搜索字段
            const { tableDataMatchSearch, searchValue } = useTableSearch(data, keys)
            const { pageChange, pageSizeChange, curPageData, pageConf } = usePage(tableDataMatchSearch)

            const handleGetCloud = async () => {
                loading.value = true
                data.value = await $store.dispatch('clustermanager/cloudAccounts', {
                    $cloudId: 'tencentCloud',
                    projectID: curProject.value.project_id
                })
                loading.value = false
            }
            const handleDeleteAccount = (row) => {
                $bkInfo({
                    type: 'warning',
                    clsName: 'custom-info-confirm',
                    subTitle: row.account.accountID,
                    title: $i18n.t('删除凭证'),
                    defaultInfo: true,
                    confirmFn: async (vm) => {
                        loading.value = true
                        const result = await $store.dispatch('clustermanager/deleteCloudAccounts', {
                            $cloudId: 'tencentCloud',
                            $accountID: row.account.accountID
                        })
                        if (result) {
                            $bkMessage({
                                theme: 'success',
                                message: $i18n.t('删除成功')
                            })
                            await handleGetCloud()
                        }
                        loading.value = false
                    }
                })
            }
            const handleCreateAccount = async () => {
                const valid = await formRef.value?.validate()
                if (!valid) return
                
                createLoading.value = true
                const result = await $store.dispatch('clustermanager/createCloudAccounts', {
                    $cloudId: 'tencentCloud',
                    ...account.value
                })
                createLoading.value = false
                if (!result) return
                
                showDialog.value = false
                $bkMessage({
                    theme: 'success',
                    message: $i18n.t('创建成功')
                })
                handleGetCloud()
            }
            const handleShowCreateDialog = () => {
                showDialog.value = true
            }
            onMounted(() => {
                handleGetCloud()
            })
            return {
                curPageData,
                searchValue,
                showDialog,
                loading,
                createLoading,
                data,
                formRules,
                account,
                formRef,
                pageConf,
                pageChange,
                pageSizeChange,
                handleCreateAccount,
                handleDeleteAccount,
                handleShowCreateDialog
            }
        }
    })
</script>
<style lang="postcss" scoped>
.tencent-cloud {
    padding: 20px;
    .w400 {
        width: 400px;
    }
    &-operate {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }
}
 /deep/ .bk-form-content .form-error-tip {
    text-align: left;
}
</style>
