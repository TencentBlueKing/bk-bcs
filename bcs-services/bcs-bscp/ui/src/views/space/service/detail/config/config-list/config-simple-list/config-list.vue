<template>
  <section class="current-config-list">
    <bk-loading :loading="loading">
      <div v-if="configList.length > 0" class="config-list-wrapper">
        <div
          v-for="config in configList"
          :class="['config-item', { disabled: config.file_state === 'DELETE' }]"
          :key="config.id"
          @click="handleEditConfigOpen(config)"
        >
          <div class="config-name">{{ config.spec.name }}</div>
          <div class="config-type">{{ getConfigTypeName(config.spec.file_type) }}</div>
        </div>
      </div>
      <bk-exception v-else scene="part" type="empty" :description="t('暂无数据')"></bk-exception>
    </bk-loading>
    <EditConfig v-model:show="editDialogShow" :bk-biz-id="props.bkBizId" :app-id="props.appId" :config-id="configId" />
  </section>
</template>
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import useConfigStore from '../../../../../../../store/config';
import { ICommonQuery } from '../../../../../../../../types/index';
import { IConfigItem } from '../../../../../../../../types/config';
import { getConfigList, getReleasedConfigList } from '../../../../../../../api/config';
import { getConfigTypeName } from '../../../../../../../utils/config';
import EditConfig from '../config-table-list/edit-config.vue';

const store = useConfigStore();
const { versionData } = storeToRefs(store);
const { t } = useI18n();

const props = defineProps<{
  bkBizId: string;
  appId: number;
}>();

const loading = ref(false);
const configList = ref<Array<IConfigItem>>([]);
const configId = ref(0);
const editDialogShow = ref(false);

watch(
  () => versionData.value.id,
  () => {
    getListData();
  },
);

onMounted(() => {
  getListData();
});

const getListData = async () => {
  // 拉取到版本列表之前不加在列表数据
  if (typeof versionData.value.id !== 'number') {
    return;
  }

  loading.value = true;
  try {
    const params: ICommonQuery = {
      start: 0,
      all: true,
    };

    let res;
    if (versionData.value.id === 0) {
      res = await getConfigList(props.bkBizId, props.appId, params);
    } else {
      res = await getReleasedConfigList(props.bkBizId, props.appId, versionData.value.id, params);
    }
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
};

const handleEditConfigOpen = (config: IConfigItem) => {
  if (config.file_state === 'DELETE') {
    return;
  }
  editDialogShow.value = true;
  configId.value = config.id;
};
</script>
<style lang="scss" scoped>
.current-config-list {
  padding: 24px;
  height: 100%;
  background: #fafbfd;
  overflow: auto;
}
.config-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
  background: #ffffff;
  box-shadow: 0 1px 1px 0 rgba(0, 0, 0, 0.06);
  border-radius: 2px;
  cursor: pointer;
  &.disabled {
    cursor: not-allowed;
    .config-type,
    .config-name {
      color: #dcdee5;
    }
  }
  &:not(.disabled):hover {
    background: #e1ecff;
  }
  .config-name {
    padding: 0 16px;
    width: 242px;
    height: 40px;
    line-height: 40px;
    color: #313238;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
  }
  .config-type {
    color: #979ba5;
  }
}
</style>
