<template>
    <BaseLayout title="Jobs" kind="Job" category="jobs" type="workloads">
        <template #default="{ curPageData, pageConf, handlePageChange, handlePageSizeChange, handleGetExtData, gotoDetail, handleSortChange,handleUpdateResource,handleDeleteResource, pagePerms }">
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
                <bk-table-column :label="$t('命名空间')" prop="metadata.namespace" sortable></bk-table-column>
                <bk-table-column :label="$t('镜像')" width="450" :show-overflow-tooltip="false">
                    <template slot-scope="{ row }">
                        <span v-bk-tooltips.top="(handleGetExtData(row.metadata.uid, 'images') || []).join('<br />')">
                            {{ (handleGetExtData(row.metadata.uid, 'images') || []).join(', ') }}
                        </span>
                    </template>
                </bk-table-column>
                <bk-table-column label="Completions" :resizable="false">
                    <template slot-scope="{ row }">{{row.status.succeeded || 0}} / {{row.spec.completions}}</template>
                </bk-table-column>
                <bk-table-column label="Duration" :resizable="false">
                    <template slot-scope="{ row }">{{handleGetExtData(row.metadata.uid, 'duration') || '--'}}</template>
                </bk-table-column>
                <bk-table-column label="Age" :resizable="false">
                    <template #default="{ row }">
                        <span>{{handleGetExtData(row.metadata.uid, 'age')}}</span>
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
        components: { BaseLayout }
    })
</script>
