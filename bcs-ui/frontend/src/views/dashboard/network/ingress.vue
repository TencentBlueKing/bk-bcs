<template>
    <BaseLayout title="Ingresses" kind="Ingress" category="ingresses" type="networks">
        <template #default="{ curPageData, pageConf, handlePageChange, handlePageSizeChange, handleGetExtData, handleShowDetail, handleSortChange,handleUpdateResource,handleDeleteResource, pagePerms }">
            <bk-table
                :data="curPageData"
                :pagination="pageConf"
                @page-change="handlePageChange"
                @page-limit-change="handlePageSizeChange"
                @sort-change="handleSortChange">
                <bk-table-column :label="$t('名称')" prop="metadata.name" sortable>
                    <template #default="{ row }">
                        <bk-button class="bcs-button-ellipsis" text @click="handleShowDetail(row)">{{ row.metadata.name }}</bk-button>
                    </template>
                </bk-table-column>
                <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" sortable></bk-table-column>
                <bk-table-column label="Class" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ row.spec.ingressClassName || '--' }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Hosts" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ handleGetExtData(row.metadata.uid, 'hosts').join(', ') || '*' }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Address" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ handleGetExtData(row.metadata.uid, 'addresses').join(', ') || '--' }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Ports" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ handleGetExtData(row.metadata.uid, 'default_ports') || '--' }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                    <template #default="{ row }">
                        <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">{{ handleGetExtData(row.metadata.uid, 'age') }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column :label="$t('操作')" :resizable="false" width="150">
                    <template #default="{ row }">
                        <bk-button text v-authority="{ clickable: pagePerms.update.clickable, content: pagePerms.update.tip }"
                            @click="handleUpdateResource(row)">{{ $t('更新') }}</bk-button>
                        <bk-button class="ml10" text v-authority="{ clickable: pagePerms.delete.clickable, content: pagePerms.delete.tip }"
                            @click="handleDeleteResource(row)">{{ $t('删除') }}</bk-button>
                    </template>
                </bk-table-column>
            </bk-table>
        </template>
        <template #detail="{ data, extData }">
            <IngressDetail :data="data" :ext-data="extData"></IngressDetail>
        </template>
    </BaseLayout>
</template>
<script>
    import { defineComponent } from '@vue/composition-api'
    import BaseLayout from '@open/views/dashboard/common/base-layout'
    import IngressDetail from './ingress-detail.vue'

    export default defineComponent({
        components: { BaseLayout, IngressDetail }
    })
</script>
