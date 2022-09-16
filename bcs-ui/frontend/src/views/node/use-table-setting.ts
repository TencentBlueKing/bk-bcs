import { ref } from '@vue/composition-api'
import VueI18n from 'vue-i18n'
import { CLUSTER_NODE_TABLE_COL } from '@/common/constant'

export interface IField {
    id: string;
    label: string | VueI18n.TranslateResult;
}
// 表格设置功能
export default function useTableSetting (fields: Array<IField>, selectedFields: Array<string>) {
    const stroageFields = JSON.parse(localStorage.getItem(CLUSTER_NODE_TABLE_COL) || '[]')
    const tableSetting = ref({
        size: 'medium',
        fields,
        selectedFields: fields.filter(item => {
            return selectedFields.includes(item.id) || stroageFields.includes(item.id)
        })
    })
    const isColumnRender = (id) => {
        return tableSetting.value.selectedFields.some(item => item.id === id)
    }
    const handleSettingChange = ({ size, fields }) => {
        tableSetting.value.size = size
        tableSetting.value.selectedFields = fields
        localStorage.setItem(CLUSTER_NODE_TABLE_COL, JSON.stringify(fields.map(item => item.id)))
    }
    return {
        tableSetting,
        isColumnRender,
        handleSettingChange
    }
}
