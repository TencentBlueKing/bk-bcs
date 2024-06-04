<template>
  <bk-popover
    ref="buttonRef"
    theme="light create-config-button-popover"
    placement="bottom-end"
    trigger="click"
    width="122"
    :arrow="false"
    @after-show="isPopoverOpen = true"
    @after-hidden="isPopoverOpen = false">
    <div theme="primary" :class="['create-config-btn', { 'popover-open': isPopoverOpen }]">
      {{ isFileType ? t('新建配置文件') : t('新建配置项') }}
      <AngleDown class="angle-icon" />
    </div>
    <template #content>
      <div class="add-config-operations">
        <div
          v-cursor="{ active: !hasEditServicePerm }"
          :class="['operation-item', { 'bk-text-with-no-perm': !hasEditServicePerm }]"
          @click="handleManualCreateSlideOpen">
          {{ t('手动新增') }}
        </div>
        <div
          v-cursor="{ active: !hasEditServicePerm }"
          :class="['operation-item', { 'bk-text-with-no-perm': !hasEditServicePerm }]"
          @click="handleBatchImportDialogOpen">
          {{ t('批量导入') }}
        </div>
      </div>
    </template>
  </bk-popover>
  <ManualCreate
    v-model:show="isManualCreateSliderOpen"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @confirm="emits('created')" />
  <ManualCreateKv
    v-model:show="isManualCreateKvSliderOpen"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @confirm="emits('created')" />
  <BatchImportKv
    v-model:show="isBatchImportKvDialogOpen"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @confirm="emits('created')" />
  <BatchImportFile
    v-model:show="isBatchImportDialogOpen"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    @confirm="emits('created')" />
</template>
<script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import { useRoute } from 'vue-router';
  import { AngleDown } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import { useI18n } from 'vue-i18n';
  import useServiceStore from '../../../../../../../../store/service';
  import ManualCreate from './manual-create.vue';
  import ManualCreateKv from './manual-create-kv.vue';
  import BatchImportKv from './import-kv/index.vue';
  import BatchImportFile from './import-file/index.vue';

  const route = useRoute();
  const { t } = useI18n();

  const serviceStore = useServiceStore();
  const { permCheckLoading, hasEditServicePerm, isFileType } = storeToRefs(serviceStore);
  const { checkPermBeforeOperate } = serviceStore;

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const emits = defineEmits(['created', 'imported', 'uploaded']);

  const buttonRef = ref();
  const isPopoverOpen = ref(false);
  const isManualCreateSliderOpen = ref(false);
  const isManualCreateKvSliderOpen = ref(false);
  const isImportTemplatesDialogOpen = ref(false);
  const isBatchImportDialogOpen = ref(false);
  const isBatchImportKvDialogOpen = ref(false);

  onMounted(() => {
    if (route.query.pkg_id) {
      isImportTemplatesDialogOpen.value = true;
    }
  });

  const handleManualCreateSlideOpen = () => {
    buttonRef.value.hide();
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    if (isFileType.value) {
      isManualCreateSliderOpen.value = true;
    } else {
      isManualCreateKvSliderOpen.value = true;
    }
  };

  const handleBatchImportDialogOpen = () => {
    buttonRef.value.hide();
    if (permCheckLoading.value || !checkPermBeforeOperate('update')) {
      return;
    }
    if (isFileType.value) {
      isBatchImportDialogOpen.value = true;
    } else {
      isBatchImportKvDialogOpen.value = true;
    }
  };
</script>
<style lang="scss" scoped>
  .create-config-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 8px;
    min-width: 122px;
    height: 32px;
    line-height: 32px;
    color: #3a84ff;
    border: 1px solid #3a84ff;
    border-radius: 2px;
    cursor: pointer;
    &.popover-open {
      .angle-icon {
        transform: rotate(-180deg);
      }
    }
    .angle-icon {
      font-size: 20px;
      transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    }
  }
</style>
<style lang="scss">
  .create-config-button-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .add-config-operations {
      .operation-item {
        padding: 0 12px;
        min-width: 58px;
        height: 32px;
        line-height: 32px;
        color: #63656e;
        font-size: 12px;
        cursor: pointer;
        &:hover {
          background: #f5f7fa;
        }
      }
    }
  }
</style>
