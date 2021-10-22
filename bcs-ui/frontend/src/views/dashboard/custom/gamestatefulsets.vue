<template>
    <BaseLayout title="GameStatefulSets" kind="GameStatefulSet" type="crd" category="custom_objects" default-crd="gamestatefulsets.tkex.tencent.com"
        default-active-detail-type="yaml" :show-crd="false" :show-detail-tab="false">
        <template #default="{ curPageData, pageConf, handlePageChange, handlePageSizeChange, handleGetExtData, handleUpdateResource, handleDeleteResource,
                              handleSortChange, handleShowDetail, renderCrdHeader, getJsonPathValue, additionalColumns, pagePerms }">
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
                <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" min-width="100" sortable>
                    <template #default="{ row }">
                        {{ row.metadata.namespace || '--' }}
                    </template>
                </bk-table-column>
                <bk-table-column
                    v-for="item in additionalColumns"
                    :key="item.name"
                    :label="item.name"
                    :prop="item.JSONPath"
                    :render-header="renderCrdHeader">
                    <template #default="{ row }">
                        <span>{{ getJsonPathValue(row, item.JSONPath) || '--' }}</span>
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
    </BaseLayout>
</template>
<script>
    import { defineComponent } from '@vue/composition-api'
    import BaseLayout from '@open/views/dashboard/common/base-layout'

    export default defineComponent({
        name: 'GameStatefulSets',
        components: { BaseLayout }
    })
</script>
