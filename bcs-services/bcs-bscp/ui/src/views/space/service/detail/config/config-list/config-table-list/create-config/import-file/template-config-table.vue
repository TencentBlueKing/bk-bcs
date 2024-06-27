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
      <bk-table class="template-table" :border="['outer', 'row', 'col']" :data="data" :row-draggable="false">
        <bk-table-column :label="t('配置模板空间')" prop="template_space_name" :width="432"></bk-table-column>
        <bk-table-column :label="t('模板套餐')" prop="template_set_name"></bk-table-column>
        <bk-table-column label="" align="center" :width="50">
          <template #default="{ index }">
            <i class="bk-bscp-icon icon-reduce delete-icon" @click="handleDeleteConfig(index)"></i>
          </template>
        </bk-table-column>
      </bk-table>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { DownShape } from 'bkui-vue/lib/icon';
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
    { immediate: true },
  );

  const handleDeleteConfig = (index: number) => {
    data.value = data.value.filter((item, i) => i !== index);
    emits('change', data.value);
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
  }
  .delete-icon {
    cursor: pointer;
    font-size: 14px;
    color: gray;
  }
</style>

<style lang="scss">
  .popover-wrap {
    padding: 0 !important;
  }
</style>
