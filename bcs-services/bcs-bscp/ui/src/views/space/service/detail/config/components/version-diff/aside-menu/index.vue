<template>
  <div class="version-diff-side">
    <div class="config-list-apart" :class="{ 'config-list-kv': !isFileType }">
      <Configs
        v-if="isFileType"
        :base-version-id="props.baseVersionId"
        :current-version-id="props.currentVersionId"
        :un-named-version-variables="props.unNamedVersionVariables"
        :selected-config="props.selectedConfig"
        :actived="selectedType === 'config'"
        :is-publish="props.isPublish"
        @selected="handleSelect($event, 'config')" />
      <ConfigsKv
        v-else
        :base-version-id="props.baseVersionId"
        :current-version-id="props.currentVersionId"
        :selected-id="props.selectedKvConfigId"
        :actived="selectedType === 'config'"
        :is-publish="props.isPublish"
        @selected="handleSelect($event, 'config')" />
    </div>
    <Scripts
      v-if="isFileType"
      :base-version-id="props.baseVersionId"
      :current-version-id="props.currentVersionId"
      :actived="selectedType === 'script'"
      @selected="handleSelect($event, 'script')" />
  </div>
</template>
<script lang="ts" setup>
  import { ref } from 'vue';
  import { IDiffDetail } from '../../../../../../../../../types/service';
  import { IConfigDiffSelected } from '../../../../../../../../../types/config';
  import { IVariableEditParams } from '../../../../../../../../../types/variable';
  import { storeToRefs } from 'pinia';
  import useServiceStore from '../../../../../../../../store/service';
  import Configs from './configs.vue';
  import ConfigsKv from './configs-kv.vue';
  import Scripts from './scripts.vue';

  const serviceStore = storeToRefs(useServiceStore());

  const { isFileType } = serviceStore;

  const props = defineProps<{
    baseVersionId: number;
    currentVersionId: number;
    unNamedVersionVariables?: IVariableEditParams[];
    selectedConfig?: IConfigDiffSelected;
    selectedKvConfigId?: number;
    isPublish: boolean;
  }>();

  const emits = defineEmits(['selected']);

  const selectedType = ref('config');

  const handleSelect = (data: IDiffDetail, type: string) => {
    selectedType.value = type;
    emits('selected', data);
  };
</script>
<style lang="scss" scoped>
  .version-diff-side {
    width: 264px;
    height: 100%;
    background: #fafbfd;
    .config-list-apart {
      height: calc(100% - 132px);
      background: #f0f1f5;
    }
    .config-list-kv {
      height: 100%;
    }
    .scripts-menu {
      border-top: 1px solid #dcded5;
    }
  }
</style>
