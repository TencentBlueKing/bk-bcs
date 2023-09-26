<script lang="ts" setup>
  import { ref } from 'vue'
  import { IDiffDetail } from '../../../../../../../../../types/service';
  import { IConfigDiffSelected } from '../../../../../../../../../types/config'
  import { IVariableEditParams } from '../../../../../../../../../types/variable'
  import Configs from './configs.vue'
  import Scripts from './scripts.vue'

  const props = defineProps<{
    baseVersionId: number;
    currentVersionId: number;
    unNamedVersionVariables?: IVariableEditParams[];
    selectedConfig?: IConfigDiffSelected;
  }>()

  const emits = defineEmits(['selected'])

  const selectedType = ref('config')

  const handleSelect = (data: IDiffDetail, type: string) => {
    selectedType.value = type
    emits('selected', data)
  }
</script>
<template>
  <div class="version-diff-side">
    <div class="config-list-apart">
      <Configs
        :base-version-id="props.baseVersionId"
        :current-version-id="props.currentVersionId"
        :un-named-version-variables="props.unNamedVersionVariables"
        :selected-config="props.selectedConfig"
        :actived="selectedType === 'config'"
        @selected="handleSelect($event, 'config')" />
    </div>
    <Scripts
      :base-version-id="props.baseVersionId"
      :current-version-id="props.currentVersionId"
      :actived="selectedType === 'script'"
      @selected="handleSelect($event, 'script')" />
  </div>
</template>
<style lang="scss" scoped>
  .version-diff-side {
    width: 264px;
    height: 100%;
    background: #fafbfd;
    .config-list-apart {
      height: calc(100% - 132px);
      background: #f0f1f5;
    }
    .scripts-menu {
      border-top: 1px solid #dcded5;
    }
  }
</style>
