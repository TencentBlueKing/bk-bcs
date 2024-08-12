<template>
  <div style="margin-bottom: 16px">
    <div class="title">
      <div class="title-content" @click="expand = !expand">
        <DownShape :class="['fold-icon', { fold: !expand }]" />
        <div class="title-text">
          {{ isExsitTable ? t('已存在配置模板套餐') : t('新建配置模板套餐') }} <span>({{ tableData.length }})</span>
        </div>
      </div>
    </div>
    <div v-show="expand">
      <bk-table class="template-table" :border="['outer', 'row', 'col']" :row-class="getRowCls" :data="data">
        <bk-table-column :label="t('配置模板空间')" prop="template_space_name" :width="432">
          <template #default="{ row }">
            <div v-if="row.template_space_id" class="row-cell">
              <span :class="{ error: !row.template_space_exist }">{{ row.template_space_name }}</span>
              <Warn v-if="!row.template_space_exist" class="warn-icon" v-bk-tooltips="{ content: getErrorInfo(row) }" />
            </div>
          </template>
        </bk-table-column>
        <bk-table-column :label="t('模板套餐')" prop="template_set_name">
          <template #default="{ row }">
            <div v-if="row.template_space_id" class="row-cell">
              <span :class="{ error: !row.template_set_exist || row.template_set_is_empty }">
                {{ row.template_set_name }}
              </span>
              <Warn
                v-if="row.template_space_exist && (!row.template_set_exist || row.template_set_is_empty)"
                class="warn-icon"
                v-bk-tooltips="{ content: getErrorInfo(row) }" />
            </div>
          </template>
        </bk-table-column>
        <bk-table-column label="" align="center" :width="50">
          <template #default="{ row }">
            <i class="bk-bscp-icon icon-reduce delete-icon" @click="handleDeleteConfig(row)"></i>
          </template>
        </bk-table-column>
      </bk-table>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { DownShape, Warn } from 'bkui-vue/lib/icon';
  import { ImportTemplateConfigItem } from '../../../../../../../../../../types/template';
  import { cloneDeep } from 'lodash';

  const { t } = useI18n();

  const data = ref<ImportTemplateConfigItem[]>([]);

  const props = defineProps<{
    tableData: ImportTemplateConfigItem[];
    isExsitTable: boolean;
  }>();

  const emits = defineEmits(['change']);

  const expand = ref(true);

  watch(
    () => props.tableData,
    () => {
      data.value = cloneDeep(props.tableData);
    },
    { immediate: true, deep: true },
  );

  const handleDeleteConfig = (config: ImportTemplateConfigItem) => {
    emits('change', `${config.template_space_id} - ${config.template_set_id}`);
  };

  const getRowCls = (data: ImportTemplateConfigItem) => {
    if (!data.template_set_exist || !data.template_space_exist || data.template_set_is_empty) {
      return 'row-error';
    }
  };

  const getErrorInfo = (data: ImportTemplateConfigItem) => {
    if (!data.template_space_exist) {
      return t('模板空间不存在，无法导入，请先删除此模板');
    }
    if (!data.template_set_exist) {
      return t('模板套餐不存在，无法导入，请先删除此模板');
    }
    if (data.template_set_is_empty) {
      return t('模板套餐为空，无法导入，请先删除此模板');
    }
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
  .template-table {
    :deep(col) {
      min-width: auto !important;
    }
    :deep(.bk-table-body) {
      tr.row-error td {
        background: #fff3e1 !important;
      }
    }
  }
  .delete-icon {
    cursor: pointer;
    font-size: 14px;
    color: gray;
  }
  .warn-icon {
    color: #ff9c01;
    font-size: 14px;
    margin-left: 8px;
  }
  .row-cell {
    display: flex;
    align-items: center;
  }
</style>

<style lang="scss">
  .popover-wrap {
    padding: 0 !important;
  }
</style>
