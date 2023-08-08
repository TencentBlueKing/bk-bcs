<script lang="ts" setup>
  import { ref } from 'vue'
  import { IDiffDetail } from '../../../../../../../../../types/service';
  import Configs from './configs.vue'
  import Scripts from './scripts.vue'

  const props = defineProps<{
    baseVersionId: number;
    currentVersionId: number;
    currentConfigId?: number;
  }>()

  const emits = defineEmits(['selected'])

  const selectedMenu = ref<string|number|undefined>(props.currentConfigId)

  const handleSelect = (id: string|number, data: IDiffDetail) => {
    selectedMenu.value = id
    emits('selected', data)
  }
</script>
<template>
  <div class="version-diff-side">
    <Configs
      class="config-list-menu"
      :base-version-id="props.baseVersionId"
      :current-version-id="props.currentVersionId"
      :current-config-id="selectedMenu"
      @selected="handleSelect" />
    <Scripts
      :base-version-id="props.baseVersionId"
      :current-version-id="props.currentVersionId"
      :value="selectedMenu"
      @selected="handleSelect" />
  </div>
</template>
<style lang="scss" scoped>
  .version-diff-side {
    width: 264px;
    height: 100%;
    background: #fafbfd;
    border-right: 1px solid #dcded5;
    .configs-menu {
      height: calc(100% - 132px);
      :deep(.list-wrapper) {
        height: calc(100% - 50px);
        overflow: auto;
      }
    }
    .scripts-menu {
      border-top: 1px solid #dcded5;
    }
  }
</style>
