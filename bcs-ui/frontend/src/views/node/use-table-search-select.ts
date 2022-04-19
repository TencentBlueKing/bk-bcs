import { ref, computed, Ref } from '@vue/composition-api'
import { TranslateResult } from 'vue-i18n'

export interface ISearchSelectData {
    name: TranslateResult;
    id: string;
    multiable?: boolean;
    children?: ISearchSelectData[];
    conditions?: any[];
}
export interface ITableSearchConfig {
    searchSelectDataSource: Ref<ISearchSelectData[]>;
    filteredValue: Ref<Record<string, string[]>>; // 表格列需要配置 prop 和 column-key属性（表格组件BUG）
}

// searchSelect组件 和 table filter搜索联动
export default function useTableSearchSelect (
    {
        searchSelectDataSource,
        filteredValue
    }: ITableSearchConfig
) {
    // 搜索项有值后就不展示了
    const searchSelectData = computed(() => {
        const ids = searchSelectValue.value.map(item => item.id)
        return searchSelectDataSource.value.filter(item => !ids.includes(item.id))
    })
    const searchSelectValue = ref<any[]>([])
    // 刷新表格
    const tableKey = ref(new Date().getTime())
    const handleFilterChange = (filtersData) => {
        Object.keys(filtersData).forEach(prop => {
            const data = searchSelectDataSource.value.find(data => data.id === prop)
            const index = searchSelectValue.value.findIndex(item => item.id === prop)
            const values = data?.children?.filter(v => filtersData[prop].includes(v.id))
            searchSelectValue.value.splice(index, 1)
            if (values?.length) {
                searchSelectValue.value.push({
                    id: data?.id,
                    name: data?.name,
                    values
                })
            }
        })
    }
    const handleSearchSelectChange = (list) => {
        // 重置表格筛选项
        Object.keys(filteredValue.value).forEach(prop => {
            filteredValue.value[prop] = []
        })
        list.forEach(item => {
            if (filteredValue.value[item.id]) {
                filteredValue.value[item.id] = item.values.map(v => v.id)
            }
        })
        tableKey.value = new Date().getTime()
    }
    const handleClearSearchSelect = () => {
        handleSearchSelectChange(searchSelectValue.value)
    }
    const handleResetSearchSelect = () => {
        searchSelectValue.value = []
        Object.keys(filteredValue.value).forEach(prop => {
            filteredValue.value[prop] = []
        })
        tableKey.value = new Date().getTime()
    }

    return {
        tableKey,
        searchSelectData,
        searchSelectValue,
        handleFilterChange,
        handleSearchSelectChange,
        handleClearSearchSelect,
        handleResetSearchSelect
    }
}
