<template>
  <div :class="['package-config-table', { 'table-open': props.open }]">
    <div class="head-area" @click="emits('toggleOpen', props.pkg.id)">
      <RightShape class="triangle-icon" />
      <div class="title">{{ props.pkg.name }}</div>
    </div>
    <div v-show="props.open" class="config-table-wrapper">
      <table class="config-table">
        <thead>
          <tr>
            <th class="th-cell name">
              <div class="name-info">
                <bk-checkbox
                  :model-value="isAllSelected"
                  :indeterminate="isIndeterminate"
                  @change="handleAllSelectionChange" />
                <div class="name-text">{{ t('配置文件名称') }}</div>
              </div>
            </th>
            <th class="th-cell path">{{ t('配置文件路径') }}</th>
            <th class="th-cell memo">{{ t('配置文件描述') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="config in props.configList" :key="config.id">
            <td class="td-cell name">
              <div class="cell name-info">
                <bk-checkbox
                  :model-value="isConfigSelected(config.id)"
                  @change="handleConfigSelectionChange($event, config)" />
                <div class="name-text">{{ config.spec.name }}</div>
              </div>
            </td>
            <td class="td-cell name">
              <div class="cell">
                {{ config.spec.path }}
              </div>
            </td>
            <td class="td-cell name">
              <div class="cell">
                {{ config.spec.memo || '--' }}
              </div>
            </td>
          </tr>
          <tr v-if="props.configList.length === 0">
            <td class="td-cell" :colspan="3">
              <bk-exception class="empty-tips" type="empty" scene="part">{{ t('暂无配置文件') }}</bk-exception>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { RightShape } from 'bkui-vue/lib/icon';
  import { ITemplateConfigItem } from '../../../../../../../../../types/template';

  const { t } = useI18n();
  const props = defineProps<{
    pkg: { id: number | string; name: string };
    open: boolean;
    configList: ITemplateConfigItem[];
    selectedConfigs: { id: number; name: string }[];
  }>();

  const emits = defineEmits(['toggleOpen', 'update:selectedConfigs', 'change']);

  const isAllSelected = computed(() => {
    const res = props.configList.length > 0 && props.configList.every((item) => isConfigSelected(item.id));
    return res;
  });

  const isIndeterminate = computed(() => {
    const res = props.configList.length > 0 && props.selectedConfigs.length > 0 && !isAllSelected.value;
    return res;
  });

  const isConfigSelected = (id: number) => props.selectedConfigs.findIndex((item) => item.id === id) > -1;

  const handleAllSelectionChange = (checked: boolean) => {
    const configs = props.selectedConfigs.slice();
    if (checked) {
      props.configList.forEach((config) => {
        if (!configs.find((item) => item.id === config.id)) {
          const { id, spec } = config;
          configs.push({ id, name: spec.name });
        }
      });
    } else {
      props.configList.forEach((config) => {
        const index = configs.findIndex((item) => item.id === config.id);
        if (index > -1) {
          configs.splice(index, 1);
        }
      });
    }
    emits('update:selectedConfigs', configs);
    emits('change');
  };

  const handleConfigSelectionChange = (checked: boolean, config: ITemplateConfigItem) => {
    const configs = props.selectedConfigs.slice();
    if (checked) {
      if (!configs.find((item) => item.id === config.id)) {
        const { id, spec } = config;
        configs.push({ id, name: spec.name });
      }
    } else {
      const index = configs.findIndex((item) => item.id === config.id);
      if (index > -1) {
        configs.splice(index, 1);
      }
    }
    emits('update:selectedConfigs', configs);
    emits('change');
  };
</script>
<style lang="scss" scoped>
  .package-config-table.table-open {
    .triangle-icon {
      transform: rotate(90deg);
    }
  }
  .head-area {
    display: flex;
    align-items: center;
    padding: 0 8px;
    height: 28px;
    background: #eaebf0;
    cursor: pointer;
    .triangle-icon {
      margin-right: 8px;
      font-size: 12px;
      color: #979ba5;
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
    .title {
      font-size: 12px;
      font-weight: 700;
      color: #63656e;
    }
  }
  .config-table-wrapper {
    position: relative;
    max-height: 60vh;
    overflow: auto;
  }
  .config-table {
    width: 100%;
    border: 1px solid #dcdee5;
    border-top: none;
    table-layout: fixed;
    border-collapse: collapse;
    thead {
      position: sticky;
      top: 0;
      z-index: 1;
    }
    .th-cell {
      padding: 0 16px;
      height: 42px;
      color: #313238;
      font-size: 12px;
      font-weight: normal;
      text-align: left;
      background: #fafbfd;
      border-bottom: 1px solid #dcdee5;
    }
    .td-cell {
      padding: 0 16px;
      text-align: left;
      border-bottom: 1px solid #dcdee5;
    }
    .cell {
      height: 42px;
      line-height: 42px;
      color: #63656e;
      font-size: 12px;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .name-info {
      display: flex;
      align-items: center;
      height: 100%;
      .name-text {
        margin-left: 8px;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
      }
    }
    .empty-tips {
      margin-bottom: 20px;
      font-size: 12px;
      color: #63656e;
    }
  }
</style>
