<template>
    <div>
        <bk-table
            :data="data"
            @row-mouse-enter="handleMouseEnter"
            @row-mouse-leave="handleMouseLeave">
            <bk-table-column :label="$t('步骤')" prop="taskName" width="160">
                <template #default="{ row }">
                    <span class="task-name">
                        <span class="bcs-ellipsis">
                            {{row.taskName}}
                        </span>
                        <i
                            class="bcs-icon bcs-icon-fenxiang"
                            v-if="row.params && row.params.taskUrl"
                            @click="handleGotoSops(row)"></i>
                    </span>
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('状态')" prop="status" width="120">
                <template #default="{ row }">
                    <LoadingIcon v-if="row.status === 'RUNNING'">
                        <span class="bcs-ellipsis">{{ $t('运行中') }}</span>
                    </LoadingIcon>
                    <StatusIcon :status="row.status" :status-color-map="taskStatusColorMap" v-else>
                        {{ taskStatusTextMap[row.status.toLowerCase()] }}
                    </StatusIcon>
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('执行时间')" width="220">
                <template #default="{ row }">
                    {{row.start}}
                </template>
            </bk-table-column>
            <bk-table-column :label="$t('总耗时')">
                <template #default="{ row }">
                    {{timeDelta(row.start, row.end) || '--'}}
                </template>
            </bk-table-column>
            <bk-table-column min-width="120" :label="$t('内容')" prop="message">
                <template #default="{ row }">
                    <bcs-button text
                        v-if="row.message"
                        @click="handleShowValuesDetail(row)"
                    >{{$t('查看')}}
                    </bcs-button>
                    <span v-else>--</span>
                    <span
                        class="copy-icon ml5"
                        v-if="activeTask === row.taskName"
                        v-bk-tooltips="$t('复制内容')"
                        @click="handleCopyValues(row)">
                        <i class="bcs-icon bcs-icon-copy"></i>
                    </span>
                </template>
            </bk-table-column>
        </bk-table>
        <bcs-dialog
            header-position="left"
            :show-footer="false"
            width="860"
            v-model="showValuesDialog">
            <template #header>
                <div class="bcs-flex">
                    <span>{{ curTaskRow.taskName }}</span>
                    <span
                        class="copy-icon ml10"
                        v-bk-tooltips="$t('复制 Values')"
                        @click="handleCopyValues(curTaskRow)">
                        <i class="bcs-icon bcs-icon-copy"></i>
                    </span>
                </div>
            </template>
            <div class="task-message">{{curTaskRow.message}}</div>
        </bcs-dialog>
    </div>
</template>
<script lang="ts">
    import { defineComponent, ref } from '@vue/composition-api'
    import { taskStatusTextMap, taskStatusColorMap } from '@/common/constant'
    import { timeDelta, copyText } from '@/common/util'
    import StatusIcon from '@/views/dashboard/common/status-icon'
    import LoadingIcon from '@/components/loading-icon.vue'

    export default defineComponent({
        name: 'TaskList',
        components: {
            StatusIcon,
            LoadingIcon
        },
        props: {
            data: {
                type: Array,
                default: () => []
            }
        },
        setup (props, ctx) {
            const { $bkMessage, $i18n } = ctx.root

            const curTaskRow = ref({})
            const showValuesDialog = ref(false)
            const activeTask = ref('')
            // 跳转标准运维
            const handleGotoSops = (row) => {
                window.open(row.params.taskUrl)
            }
            const handleShowValuesDetail = (row) => {
                curTaskRow.value = row
                showValuesDialog.value = true
            }
            const handleCopyValues = (row) => {
                copyText(row.message)
                $bkMessage({
                    theme: 'success',
                    message: $i18n.t('复制成功')
                })
            }
            const handleMouseEnter = (index, event, row) => {
                activeTask.value = row.taskName
            }
            const handleMouseLeave = () => {
                activeTask.value = ''
            }
            return {
                taskStatusColorMap,
                taskStatusTextMap,
                timeDelta,
                curTaskRow,
                activeTask,
                showValuesDialog,
                handleGotoSops,
                handleCopyValues,
                handleShowValuesDetail,
                handleMouseLeave,
                handleMouseEnter
            }
        }
    })
</script>
<style lang="postcss" scoped>
>>> .task-name {
    display: flex;
    align-items: center;
    justify-content: space-between;
    i {
        color: #3a84ff;
        cursor: pointer;
    }
}
>>> .copy-icon {
    color: #3a84ff;
    cursor: pointer;
    font-size: 14px;
}
.task-message {
    height: 500px;
    overflow-y: scroll;
    padding-right: 20px;
}
</style>
