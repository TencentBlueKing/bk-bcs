<template>
    <BaseLayout title="GameStatefulSets" kind="GameStatefulSet" type="crd" category="custom_objects" default-crd="gamestatefulsets.tkex.tencent.com"
        default-active-detail-type="yaml" :show-crd="false" :show-detail-tab="false">
        <template #default="{ curPageData, pageConf, handlePageChange, handlePageSizeChange, handleGetExtData, handleUpdateResource, handleDeleteResource,
                              handleSortChange, gotoDetail, renderCrdHeader, getJsonPathValue, additionalColumns, webAnnotations }">
            <bk-table
                :data="curPageData"
                :pagination="pageConf"
                @page-change="handlePageChange"
                @page-limit-change="handlePageSizeChange"
                @sort-change="handleSortChange">
                <bk-table-column :label="$t('名称')" prop="metadata.name" sortable>
                    <template #default="{ row }">
                        <bk-button class="bcs-button-ellipsis" text @click="gotoDetail(row)">{{ row.metadata.name }}</bk-button>
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
                    :prop="item.jsonPath"
                    :render-header="renderCrdHeader">
                    <template #default="{ row }">
                        <span>
                            {{ typeof getJsonPathValue(row, item.jsonPath) !== 'undefined'
                                ? getJsonPathValue(row, item.jsonPath) : '--' }}
                        </span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Age" :resizable="false" :show-overflow-tooltip="false">
                    <template #default="{ row }">
                        <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">{{ handleGetExtData(row.metadata.uid, 'age') }}</span>
                    </template>
                </bk-table-column>
                <bk-table-column :label="$t('操作')" :resizable="false" width="150">
                    <template #default="{ row }">
                        <bk-button text
                            @click="handleUpdateResource(row)">{{ $t('更新') }}</bk-button>
                        <bk-button class="ml10" text
                            v-authority="{
                                clickable: webAnnotations.perms.items[row.metadata.uid] ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.clickable : true,
                                content: webAnnotations.perms.items[row.metadata.uid] ? webAnnotations.perms.items[row.metadata.uid].deleteBtn.tip : '',
                                disablePerms: true
                            }"
                            @click="handleDeleteResource(row)">{{ $t('删除') }}</bk-button>
                    </template>
                </bk-table-column>
            </bk-table>
        </template>
    </BaseLayout>
</template>
<script>
    import { defineComponent } from '@vue/composition-api'
    import BaseLayout from '@/views/dashboard/common/base-layout'

    export default defineComponent({
        name: 'GameStatefulSets',
        components: { BaseLayout }
    })
</script>
