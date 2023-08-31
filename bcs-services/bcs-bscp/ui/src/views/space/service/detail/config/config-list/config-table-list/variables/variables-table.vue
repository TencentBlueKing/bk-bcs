<script lang="ts" setup>
  import { onMounted, ref, computed } from 'vue'
  import { IVariableEditParams } from '../../../../../../../../../types/variable';

  const props = withDefaults(defineProps<{
    list: IVariableEditParams[];
    editable: boolean;
    showCited: boolean;
  }>(), {
    editable: true,
    showCited: false
  })

  const variables = ref<IVariableEditParams[]>([])

  const cols = computed(() => {
    const tableCols = [
      { label: '变量说明', cls: 'name' },
      { label: '类型', cls: 'type' },
      { label: '变量值', cls: 'value' },
      { label: '变量说明', cls: 'memo' }
    ]
    if (props.showCited) {
      tableCols.push({ label: '被引用', cls: 'cited' })
    }
    return tableCols
  })

  onMounted(() => {
    if (props.showCited) {
      getCitedData()
    }
  })

  const getCitedData = () => {}

  const getCitedTpls = (name: string) => {}

</script>
<template>
  <div class="variables-table-wrapper">
    <table class="variables-table">
      <thead>
        <tr>
          <th v-for="(col, index) in cols" :key="index" :class="col.cls">{{ col.label }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="variable in variables" :key="variable.name">
          <td>{{ variable.name }}</td>
          <td>{{ variable.type }}</td>
          <td>{{ variable.default_val }}</td>
          <td>{{ variable.memo }}</td>
          <td v-if="props.showCited">{{ getCitedTpls(variable.name) }}</td>
        </tr>
        <tr v-if="props.list.length === 0">
          <td :colspan="cols.length">
            <bk-exception class="empty-tips" type="empty" scene="part">暂无变量数据</bk-exception>
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
    th {
      padding: 0 16px;
      height: 42px;
      line-height: 20px;
      font-weight: normal;
      font-size: 12px;
      color: #313238;
      background: #fafbfd;
      border: 1px solid #dcdee5;
    }
    td {
      padding: 0 16px;
      height: 42px;
      line-height: 20px;
      font-size: 12px;
      border: 1px solid #dcdee5;
    }
    .empty-tips {
      margin: 20px 0;
    }
  }
</style>
