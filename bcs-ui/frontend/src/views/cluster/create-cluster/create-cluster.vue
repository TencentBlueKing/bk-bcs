<template>
    <section class="cluster">
        <div class="mode-wrapper mt15">
            <!-- 自建集群 -->
            <div class="mode-panel" @click="handleCreateCluster">
                <span class="mode-panel-icon"><i class="bcs-icon bcs-icon-sitemap"></i></span>
                <span class="mode-panel-title">{{ $t('自建集群') }}</span>
                <span class="mode-panel-desc">{{ $t('可自定义集群基本信息和集群版本') }}</span>
            </div>
            <!-- 导入集群 -->
            <div class="mode-panel disabled" v-bk-tooltips="$t('功能建设中')">
                <span class="mode-panel-icon"><i class="bcs-icon bcs-icon-upload"></i></span>
                <span class="mode-panel-title">{{ $t('导入集群') }}</span>
                <span class="mode-panel-desc">{{ $t('支持快速导入已存在的集群') }}</span>
            </div>
        </div>
        <div class="cluster-template-title">
            <span class="title">{{ $t('集群模板') }}</span>
            <!-- <bcs-button size="small" @click="handleCreateTemplate">{{ $t('新建集群模板') }}</bcs-button> -->
        </div>
        <bcs-table class="mt15"
            :data="curPageData"
            :pagination="pagination"
            v-bkloading="{ isLoading }"
            @page-change="pageChange"
            @page-limit-change="pageSizeChange">
            <bcs-table-column :label="$t('模板名称')">
                <template #default="{ row }">
                    <bcs-button text @click="handleShowDetail(row)">{{ row.name }}</bcs-button>
                </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('描述')" prop="description" min-width="300" show-overflow-tooltip>
                <template #default="{ row }">
                    {{ row.description || '--' }}
                </template>
            </bcs-table-column>
            <bcs-table-column :label="$t('创建者')" prop="creator"></bcs-table-column>
            <bcs-table-column :label="$t('更新者')" prop="updater"></bcs-table-column>
            <bcs-table-column :label="$t('更新时间')" prop="updateTime" min-width="200"></bcs-table-column>
            <!-- <bcs-table-column :label="$t('操作')">
                <template #default="{ row }">
                    <bcs-button text @click="handleEditTemplate(row)">{{ $t('编辑') }}</bcs-button>
                    <bcs-button class="ml15" text @click="handleDeleteTemplate(row)">{{ $t('删除') }}</bcs-button>
                </template>
            </bcs-table-column> -->
        </bcs-table>
        <bk-sideslider :is-show.sync="showDetail" quick-close :title="detailTitle" width="600">
            <template #content>
                <ace v-full-screen="{ tools: ['fullscreen', 'copy'], content: yaml }"
                    width="100%" height="100%" lang="yaml" read-only :value="yaml"></ace>
            </template>
        </bk-sideslider>
    </section>
</template>
<script lang="ts">
    import { computed, defineComponent, onMounted, ref } from '@vue/composition-api'
    import DashboardTopActions from '@/views/dashboard/common/dashboard-top-actions'
    import usePage from '@/views/dashboard/common/use-page'
    import * as ace from '@/components/ace-editor'
    import fullScreen from '@/directives/full-screen'
    import yamljs from 'js-yaml'

    export default defineComponent({
        name: 'CreateCluster',
        components: { DashboardTopActions, ace },
        directives: {
            'full-screen': fullScreen
        },
        setup (props, ctx) {
            const { $router, $store } = ctx.root

            // 自建集群
            const handleCreateCluster = () => {
                $router.push({ name: 'createFormCluster' })
            }

            // 获取表格数据
            const isLoading = ref(false)
            const tableData = ref([])
            const handleGetTableData = async () => {
                isLoading.value = true
                const data = await $store.dispatch('clustermanager/fetchCloudList')
                tableData.value = data
                pagination.value.count = data.length
                isLoading.value = false
            }
            // 前端分页
            const { pagination, curPageData, pageChange, pageSizeChange } = usePage(tableData)
            // 创建集群模板
            const handleCreateTemplate = () => {
                $router.push({ name: 'createClusterTemplate' })
            }
            // 编辑集群模板
            const handleEditTemplate = (row) => {}
            // 删除集群模板
            const handleDeleteTemplate = (row) => {}
            // 展示模板详情
            const showDetail = ref(false)
            const curCloud = ref<any>({})
            const detailTitle = computed(() => {
                return `${curCloud.value.name}${curCloud.value.description ? `( ${curCloud.value.description} )` : ''}`
            })
            const yaml = computed(() => {
                return yamljs.dump(curCloud.value)
            })
            const handleShowDetail = (row) => {
                showDetail.value = true
                curCloud.value = row
            }
            onMounted(() => {
                handleGetTableData()
            })
            return {
                isLoading,
                curPageData,
                pagination,
                showDetail,
                curCloud,
                detailTitle,
                yaml,
                handleCreateCluster,
                handleGetTableData,
                handleCreateTemplate,
                handleEditTemplate,
                handleDeleteTemplate,
                pageChange,
                pageSizeChange,
                handleShowDetail
            }
        }
    })
</script>
<style lang="postcss" scoped>
/deep/ .bk-sideslider-content {
    height: 100%;
}
.cluster {
    padding: 20px 24px;
    .title {
        font-size: 14px;
        font-weight: 700;
        text-align: left;
        color: #63656e;
        line-height: 22px;
    }
    .mode-wrapper {
        display: flex;
        align-items: center;
    }
    .mode-panel {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        margin-right: 24px;
        flex: 1;
        background: #fff;
        border-radius: 1px;
        box-shadow: 0px 2px 4px 0px rgba(25,25,41,0.05);
        height: 238px;
        cursor: pointer;
        &:hover {
            &:not(.disabled) {
                border: 1px solid #1768ef;
                .mode-panel-icon {
                    background: #e1ecff;
                }
                .mode-panel-title {
                    color: #3a84ff;
                }
            }
        }
        &:last-child {
            margin-right: 0;
        }
        &-icon {
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 40px;
            color: #979ba5;
            width: 80px;
            height: 80px;
            border-radius: 50%;
            background: #f5f7fa;
        }
        &-title {
            margin-top: 20px;
            font-size: 20px;
            font-weight: 400;
            color: #63656e;
            line-height: 28px;
        }
        &-desc {
            margin-top: 8px;
            font-size: 14px;
            font-weight: 400;
            text-align: center;
            color: #979ba5;
            line-height: 22px;
        }
        &.disabled {
            cursor: not-allowed;
        }
    }
    .cluster-template-title {
        display: flex;
        justify-content: space-between;
        margin-top: 40px;
    }
}
</style>
