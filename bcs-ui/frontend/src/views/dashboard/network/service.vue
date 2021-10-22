<template>
    <BaseLayout title="Services" kind="Service" category="services" type="networks">
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
                <bk-table-column label="Type" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ row.spec.type || '--' }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Cluster-ip" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ row.spec.clusterIP || '--' }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column label="External-ip" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ handleGetExtData(row.metadata.uid, 'externalIP').join(', ') || '--' }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Ports" :resizable="false">
                    <template #default="{ row }">
                        <span>{{ handleGetExtData(row.metadata.uid, 'ports').join(', ') || '*' }}</span>
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
            <ServiceDetail :data="data" :ext-data="extData"></ServiceDetail>
        </template>
    </BaseLayout>
</template>
<script>
    import { defineComponent } from '@vue/composition-api'
    import BaseLayout from '@open/views/dashboard/common/base-layout'
    import ServiceDetail from './service-detail.vue'

    export default defineComponent({
        components: { BaseLayout, ServiceDetail }
    })
</script>
