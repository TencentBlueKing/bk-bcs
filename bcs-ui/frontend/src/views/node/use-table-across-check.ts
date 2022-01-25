import { ref, Ref } from '@vue/composition-api'
import { CreateElement } from 'vue'
import AcrossCheck, { CheckType } from '@/components/across-check.vue'

export interface IAcrossCheckConfig {
    tableData: Ref<any[]>;
    curPageData: Ref<any[]>;
    rowKey?: string;
}
// 表格跨页全选功能
export default function useTableAcrossCheck ({ tableData, curPageData, rowKey = 'inner_ip' }: IAcrossCheckConfig) {
    // 0 未选，1 当前页半选， 2 跨页半选，3 当前页全选，4 跨页全选
    const selectType = ref(CheckType.Uncheck)
    const selections = ref<any[]>([])
    const renderSelection = (h: CreateElement) => {
        return h(AcrossCheck, {
            props: {
                value: selectType.value,
                disabled: !tableData.value.length
            },
            on: {
                change: handleSelectTypeChange
            }
        })
    }
    // 表头全选事件
    const handleSelectTypeChange = (value) => {
        switch (value) {
            case CheckType.Uncheck:
                handleClearSelection()
                break
            case CheckType.Checked:
                handleSelectCurrentPage()
                break
            case CheckType.AcrossChecked:
                handleSelectionAll()
                break
        }
    }
    // 当前页全选
    const handleSelectCurrentPage = () => {
        selectType.value = CheckType.Checked
        selections.value = [...curPageData.value]
    }
    // 跨页全选
    const handleSelectionAll = () => {
        selectType.value = CheckType.AcrossChecked
        selections.value = [...tableData.value]
    }
    // 清空全选
    const handleClearSelection = () => {
        selectType.value = CheckType.Uncheck
        selections.value = []
    }
    // 表格行勾选后重置状态
    const handleSetSelectType = () => {
        if (selections.value.length === 0) {
            selectType.value = CheckType.Uncheck
        } else if (selections.value.length < curPageData.value.length
            && [CheckType.Checked, CheckType.Uncheck].includes(selectType.value)) {
            // 从当前页全选 -> 当前页半选
            selectType.value = CheckType.HalfChecked
        } else if (selections.value.length === curPageData.value.length
            && selectType.value !== CheckType.HalfAcrossChecked) {
            selectType.value = CheckType.Checked
        } else if (selections.value.length < tableData.value.length
            && selectType.value === CheckType.AcrossChecked) {
            // 从跨页全选 -> 跨页半选
            selectType.value = CheckType.HalfAcrossChecked
        } else if (selections.value.length === tableData.value.length) {
            selectType.value = CheckType.AcrossChecked
        }
    }
    // 重新设置状态
    const handleResetCheckStatus = () => {
        selectType.value = CheckType.Uncheck
        selections.value = []
    }
    // 当前行选中事件
    const handleRowCheckChange = (value, row) => {
        const index = selections.value.findIndex(item => item[rowKey] === row[rowKey])
        if (value && index === -1) {
            selections.value.push(row)
        } else if (!value && index > -1) {
            selections.value.splice(index, 1)
        }
        handleSetSelectType()
    }

    return {
        selectType,
        selections,
        renderSelection,
        handleResetCheckStatus,
        handleRowCheckChange,
        handleSelectionAll,
        handleClearSelection
    }
}
