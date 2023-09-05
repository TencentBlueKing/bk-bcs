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
    <div class="config-list-apart">
      <Configs
        :base-version-id="props.baseVersionId"
        :current-version-id="props.currentVersionId"
        :current-config-id="selectedMenu"
        @selected="handleSelect" />
    </div>
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
    .config-list-apart {
      height: calc(100% - 132px);
      background: #f0f1f5;
    }
    .scripts-menu {
      border-top: 1px solid #dcded5;
    }
  }
</style>
