<script lang="ts" setup>
  import { ref, reactive, computed, watch } from 'vue'
  import cloneDeep from 'lodash';
  import { IVariableEditParams, IVariableCitedByConfigDetailItem } from '../../../../../../../../../types/variable';

  interface IErrorDetail {
    [key: string]: string[]
  }

  const props = withDefaults(defineProps<{
    list: IVariableEditParams[];
    citedList?: IVariableCitedByConfigDetailItem[]
    editable?: boolean;
    showCited?: boolean;
  }>(), {
    list: () => [],
    citedList: () => [],
    editable: true,
    showCited: false
  })

  const emits = defineEmits(['change'])

  let variables = reactive<IVariableEditParams[]>([])
  const errorDetails = ref<IErrorDetail>({})

  const cols = computed(() => {
    const tableCols = [
      { label: '变量名称', cls: 'name', prop: 'name' },
      { label: '类型', cls: 'type', prop: 'type' },
      { label: '变量值', cls: 'default_value', prop: 'default_val' },
      { label: '变量说明', cls: 'memo', prop: 'memo' }
    ]
    if (props.showCited) {
      tableCols.push({ label: '被引用', cls: 'cited', prop: 'cited' })
    }
    return tableCols
  })

  watch(() => props.list, val => {
    variables = cloneDeep(val).value()
  }, { immediate: true })

  const isCellEditable = (prop: string) => {
    return props.editable && ['type', 'default_val', 'memo'].includes(prop)
  }

  const isCellValRequired = (prop: string) => {
    return props.editable && ['type', 'default_val'].includes(prop)
  }

  const getCellVal = (variable: IVariableEditParams, prop: string) => {
    if (prop === 'cited') {
      return getCitedTpls(variable.name)
    }
    return variable[prop as keyof typeof variable]
  }

  const getCitedTpls = (name: string) => {
    const detail = props.citedList.find(item => item.variable_name === name)
    return detail?.references.map(item => item.name).join(',')
  }

  const deleteCellError = (name: string, key: string) => {
    change()
    if (errorDetails.value[name]?.includes(key)) {
      if (errorDetails.value[name].length === 0) {
        delete errorDetails.value[name]
      } else {
        errorDetails.value[name] = errorDetails.value[name].filter(item => item !== key)
      }
    }
  }

  const validate = () => {
    const errors: IErrorDetail = {}
    variables.forEach(variable => {
      ['type', 'default_val'].forEach(key => {
        if (variable[key as keyof typeof variable] === '') {
          if (errors[variable.name]) {
            errors[variable.name].push(key)
          } else {
            errors[variable.name] = [key]
          }
        }
      })
    })
    errorDetails.value = errors
    return Object.keys(errorDetails.value).length === 0
  }

  const change = () => {
    emits('change', variables)
  }

  defineExpose({
    validate
  })
</script>
<template>
  <div class="variables-table-wrapper">
    <table :class="['variables-table', { 'edit-mode': props.editable }]">
      <thead>
        <tr>
          <th v-for="col in cols" :key="col.prop" :class="['th-cell', col.cls]">
            <span :class="['label', { required: isCellValRequired(col.prop) }]">
              {{ col.label }}
            </span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(variable, index) in variables" :key="index">
          <td
            v-for="col in cols"
            :key="col.prop"
            :class="[
              'td-cell',
              col.cls,
              {
                'td-cell-edit': isCellEditable(col.prop),
                'has-error': errorDetails[variable.name]?.includes(col.prop)
              }
            ]">
            <template v-if="props.editable">
                <bk-select v-if="col.prop === 'type'" v-model="variable.type" :clearable="false" @change="deleteCellError(variable.name, col.prop)">
                  <bk-option id="string" label="string"></bk-option>
                  <bk-option id="number" label="number"></bk-option>
                </bk-select>
                <bk-input v-else-if="col.prop === 'default_val'" v-model="variable.default_val" @change="deleteCellError(variable.name, col.prop)" />
                <bk-input v-else-if="col.prop === 'memo'" v-model="variable.memo" @change="change" />
                <div v-else v-overflow-title class="cell">{{ getCellVal(variable, col.prop) }}</div>
            </template>
            <div v-else v-overflow-title class="cell">{{ getCellVal(variable, col.prop) }}</div>
          </td>
        </tr>
        <tr v-if="props.list.length === 0">
          <td :colspan="cols.length">
            <bk-exception class="empty-tips" type="empty" scene="part">暂无数据</bk-exception>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
<style lang="scss" scoped>
  .variables-table {
    width: 100%;
    border: 1px solid #dcdee5;
    table-layout: fixed;
    border-collapse: collapse;
    &.edit-mode {
      .td-cell {
        background: #f5f7fa;
        &.has-error {
          border: 1px double #ea3636;
        }
      }
      .td-cell-edit {
        padding: 0;
        :deep(.bk-input) {
          height: 42px;
          border: none;
          .bk-input--text {
            padding-left: 16px;
          }
        }
      }
    }
    .th-cell {
      padding: 0 16px;
      height: 42px;
      line-height: 20px;
      font-weight: normal;
      font-size: 12px;
      color: #313238;
      text-align: left;
      background: #fafbfd;
      border: 1px solid #dcdee5;
      .label {
        position: relative;
        &.required:after {
          position: absolute;
          top: 0;
          width: 14px;
          font-size: 14px;
          color: #ea3636;
          text-align: center;
          content: "*";
        }
      }
    }
    .td-cell {
      padding: 0 16px;
      height: 42px;
      line-height: 20px;
      font-size: 12px;
      text-align: left;
      color: #63656e;
      border: 1px solid #dcdee5;
    }
    .cell {
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .empty-tips {
      margin: 20px 0;
    }
  }
</style>
