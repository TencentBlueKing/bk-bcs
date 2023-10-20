<template>
  <div class="title">
    <div class="title-content" @click="expand = !expand">
      <DownShape :class="['fold-icon', { fold: !expand }]" />
      <div class="title-text">
        {{ headText }} <span>({{ tableData.length }})</span>
      </div>
    </div>
  </div>
  <table class="table" v-if="expand">
    <thead>
      <tr>
        <th>配置项名称</th>
        <th>配置项路径</th>
        <th>配置项格式</th>
        <th>配置项描述</th>
        <th>文件权限</th>
        <th>用户</th>
        <th>用户组</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(item, index) in tableData" :key="index">
        <td>{{ item.name }}</td>
        <td>{{ item.path }}</td>
        <td>{{ item.file_type }}</td>
        <td>{{ item.memo }}</td>
        <td>{{ item.privilege }}</td>
        <td>{{ item.user }}</td>
        <td>{{ item.user_group }}</td>
      </tr>
    </tbody>
  </table>
</template>

<script lang="ts" setup>
import { ref, h, watch, computed } from 'vue';
import { DownShape, EditLine } from 'bkui-vue/lib/icon';
import { IConfigImport } from '../../../../../../../../../types/config';
const expand = ref(true);
const data = ref<IConfigImport[]>();
const inputStr = ref('');
const props = withDefaults(
  defineProps<{
    tableData: IConfigImport[];
    headText: string;
  }>(),
  {}
);

const emits = defineEmits(['update:tableData']);

watch(
  () => props.tableData,
  () => {
    data.value = props.tableData;
  }
);

const copyData = computed(() => {});

const editInfo = (row: IConfigImport) => {};
// 自定义渲染表格头部
const renderLabel = (label: string) => {
  return () =>
    h('div', { class: 'bk-table-column-label' }, [
      h('span', { class: 'bk-table-column-label-text', style: 'margin-right:10px' }, label),
      h(EditLine, { fill: '#3A84FF' }),
    ]);
};
</script>

<style scoped lang="scss">
.title {
  height: 28px;
  background: #eaebf0;
  border-radius: 2px 2px 0 0;
  .title-content {
    display: flex;
    align-items: center;
    height: 100%;
    margin-left: 10px;
    cursor: pointer;
    .fold-icon {
      margin-right: 8px;
      font-size: 14px;
      color: #979ba5;
      transition: transform 0.2s ease-in-out;
      &.fold {
        transform: rotate(-90deg);
      }
    }
    .title-text {
      font-weight: 700;
      font-size: 12px;
      color: #63656e;
      span {
        font-size: 12px;
        color: #979ba5;
      }
    }
  }
}
.table {
  width: 100%;
  border-collapse: collapse;
  th,
  td {
    border: 1px solid #dcdee5;
    height: 40px;
    padding: 0 16px;
    text-align: left;
  }
  th {
    font-size: 12px;
    color: #313238;
  }
}
</style>
