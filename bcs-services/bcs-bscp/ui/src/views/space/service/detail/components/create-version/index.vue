<template>
  <bk-button
    v-if="versionData.id === 0"
    v-cursor="{ active: !props.hasPerm }"
    theme="primary"
    :class="['trigger-button', { 'bk-button-with-no-perm': !props.hasPerm }]"
    :disabled="!props.hasPerm || allExistConfigCount === 0 || props.permCheckLoading || conflictFileCount > 0"
    @click="handleBtnClick">
    {{ t('生成版本') }}
  </bk-button>
  <CreateVersionSlider
    v-model:show="isVersionSliderShow"
    ref="createSliderRef"
    :bk-biz-id="props.bkBizId"
    :app-id="props.appId"
    :is-diff-slider-show="isDiffSliderShow"
    @created="handleCreated" />
</template>
<script setup lang="ts">
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import Message from 'bkui-vue/lib/message';
  import useGlobalStore from '../../../../../../store/global';
  import useConfigStore from '../../../../../../store/config';
  import { IConfigVersion } from '../../../../../../../types/config';
  import CreateVersionSlider from './create-version-slider.vue';

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    permCheckLoading: boolean;
    hasPerm: boolean;
  }>();

  const emits = defineEmits(['confirm']);
  const { t } = useI18n();

  const { permissionQuery, showApplyPermDialog } = storeToRefs(useGlobalStore());
  const { allExistConfigCount, versionData, conflictFileCount } = storeToRefs(useConfigStore());
  const isVersionSliderShow = ref(false);
  const isDiffSliderShow = ref(false);
  const createSliderRef = ref();

  const permissionQueryResource = computed(() => [
    {
      biz_id: props.bkBizId,
      basic: {
        type: 'app',
        action: 'generate_release',
        resource_id: props.appId,
      },
    },
  ]);

  const handleBtnClick = () => {
    if (props.hasPerm) {
      isVersionSliderShow.value = true;
    } else {
      permissionQuery.value = { resources: permissionQueryResource.value };
      showApplyPermDialog.value = true;
    }
  };

  const handleCreated = (versionData: IConfigVersion, isPublish: boolean) => {
    isDiffSliderShow.value = false;
    isVersionSliderShow.value = false;
    emits('confirm', versionData, isPublish);
    Message({ theme: 'success', message: t('新版本已生成') });
  };
</script>
<style lang="scss" scoped>
  .trigger-button {
    margin-left: 8px;
  }
  .create-version-btn {
    margin-right: 8px;
  }
</style>
