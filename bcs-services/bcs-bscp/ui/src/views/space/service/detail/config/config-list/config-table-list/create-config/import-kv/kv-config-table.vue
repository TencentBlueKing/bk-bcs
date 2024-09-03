<template>
  <div style="margin-bottom: 16px">
    <div class="title">
      <div class="title-content" @click="emits('changeExpand')">
        <DownShape :class="['fold-icon', { fold: !expand }]" />
        <div class="title-text">
          {{ isExsitTable ? $t('已存在配置文件') : $t('新建配置文件') }} <span>({{ tableData.length }})</span>
        </div>
      </div>
    </div>
    <bk-table
      v-show="expand"
      :data="data"
      :border="['outer', 'row', 'col']"
      class="kv-config-table"
      :cell-class="getCellCls"
      show-overflow-tooltip>
      <bk-table-column :label="$t('配置项名称')" prop="key" width="320" property="key"></bk-table-column>
      <bk-table-column :label="$t('数据类型')" prop="kv_type" width="200" property="type"></bk-table-column>
      <bk-table-column :label="$t('配置项值预览')" prop="value" width="280">
        <template #default="{ row }">
          <div v-if="row.key" :class="{ hidden: isSecretHidden(row) }" type="tips">
            {{ isSecretHidden(row) ? $t('敏感数据不可见，无法查看实际内容') : row.value }}
          </div>
        </template>
      </bk-table-column>
      <bk-table-column :label="$t('配置项描述')" prop="memo" property="memo">
        <template #default="{ row }">
          <div v-if="row.key" class="memo">
            <div v-if="memoEditKey !== row.key" class="memo-display" @click="handleOpenMemoEdit(row.key)">
              <div class="memo-text" type="tips">
                {{ row.memo || '--' }}
              </div>
            </div>
            <bk-input
              v-else
              ref="memoInputRef"
              class="memo-input"
              type="textarea"
              :model-value="row.memo"
              :autosize="{ maxRows: 4 }"
              :resize="false"
              @blur="handleMemoEditBlur(row, $event)" />
          </div>
        </template>
      </bk-table-column>
      <bk-table-column label="" width="50">
        <template #default="{ row }">
          <i class="bk-bscp-icon icon-reduce delete-icon" @click="handleDelete(row)"></i>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<script lang="ts" setup>
  import { ref, nextTick, watch, onMounted, computed } from 'vue';
  import { IConfigKvItem } from '../../../../../../../../../../types/config';
  import { DownShape } from 'bkui-vue/lib/icon';
  import { cloneDeep, isEqual } from 'lodash';

  const props = defineProps<{
    tableData: IConfigKvItem[];
    isExsitTable: boolean;
    expand: boolean;
  }>();

  const emits = defineEmits(['changeExpand', 'change']);

  const memoEditKey = ref('');
  const memoInputRef = ref();
  const data = ref<IConfigKvItem[]>([]);
  const initData = ref<IConfigKvItem[]>([]);

  onMounted(() => {
    data.value = cloneDeep(props.tableData);
    initData.value = cloneDeep(props.tableData);
  });

  watch(
    () => props.tableData,
    () => {
      data.value = cloneDeep(props.tableData);
      initData.value = cloneDeep(props.tableData);
    },
    { deep: true },
  );

  watch(
    () => data.value,
    () => {
      if (isEqual(data.value, initData.value)) {
        return;
      }
      emits('change', data.value);
    },
    { deep: true },
  );

  const isSecretHidden = computed(() => (config: IConfigKvItem) => {
    return config.kv_type === 'secret' && config.secret_hidden;
  });

  const handleDelete = (item: IConfigKvItem) => {
    const index = data.value.findIndex((kv) => kv.key === item.key);
    if (index > -1) {
      data.value.splice(index, 1);
    }
  };

  // 添加自定义单元格class
  const getCellCls = ({ property }: { property: string }) => {
    if (property === 'memo') return 'memo-cell';
    return ['key', 'type', 'value'].includes(property) ? 'disabled-cell' : '';
  };

  const handleOpenMemoEdit = (key: string) => {
    memoEditKey.value = key;
    nextTick(() => {
      memoInputRef.value?.focus();
    });
  };

  const handleMemoEditBlur = (kv: IConfigKvItem, e: FocusEvent) => {
    memoEditKey.value = '';
    const val = (e.target as HTMLInputElement).value.trim();
    kv.memo = val;
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
  .kv-config-table {
    :deep(col) {
      min-width: 50px !important;
    }
    :deep(.bk-table-body-content) {
      .disabled-cell {
        background-color: #f5f7fa;
      }
      .memo-cell {
        .cell {
          padding: 0 !important;
        }
        .memo {
          position: relative;
          padding: 0 16px;
        }
        .memo-input {
          width: 100%;
          height: 41px;
          position: absolute;
          top: 0;
          left: 0;
        }
      }
    }
    .delete-icon {
      cursor: pointer;
      font-size: 14px;
      color: #c4c6cc;
      &:hover {
        color: #3a84ff;
      }
    }
  }

  .hidden {
    color: #c4c6cc;
  }
</style>
