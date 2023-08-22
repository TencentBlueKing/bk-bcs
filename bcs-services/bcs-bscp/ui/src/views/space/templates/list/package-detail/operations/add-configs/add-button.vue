<script lang="ts" setup>
  import { ref } from 'vue';
  import { Plus } from 'bkui-vue/lib/icon';
  import { useTemplateStore } from '../../../../../../../store/template'
  import AddFromExistingConfigs from './add-from-existing-configs/index.vue';
  import CreateNewConfig from './create-new-config/index.vue'
  import ImportConfigs from './import-configs.vue'

  const templateStore = useTemplateStore()

  const props = defineProps<{
    showAddExistingConfigOption?: boolean;
  }>()

  const emits = defineEmits(['added'])

  const buttonRef = ref()
  const silders = ref<Record<string, boolean>>({
    isAddOpen: false,
    isCreateOpen: false,
    isImportOpen: false
  })

  const handleOpenSlider = (slider: string) => {
    silders.value[slider] = true
    buttonRef.value.hide()
  }

  const handleAdded = () => {
    // 更新左侧套餐菜单栏及套餐下配置项数量
    templateStore.$patch(state => {
      state.needRefreshMenuFlag = true
    })
    emits('added')
  }

</script>
<template>
  <bk-popover
    ref="buttonRef"
    theme="light add-configs-button-popover"
    placement="bottom-end"
    trigger="click"
    width="122"
    :arrow="false">
    <bk-button
      theme="primary"
      class="create-config-btn">
      <Plus class="button-icon" />添加配置项
    </bk-button>
    <template #content>
      <div class="add-config-operations">
        <div v-if="props.showAddExistingConfigOption" class="operation-item" @click="handleOpenSlider('isAddOpen')">添加已有配置项</div>
        <div class="operation-item" @click="handleOpenSlider('isCreateOpen')">新建配置项</div>
        <div class="operation-item" @click="handleOpenSlider('isImportOpen')">导入配置项</div>
      </div>
    </template>
  </bk-popover>
  <AddFromExistingConfigs v-model:show="silders.isAddOpen" @added="handleAdded" />
  <CreateNewConfig v-model:show="silders.isCreateOpen" @added="handleAdded" />
  <ImportConfigs v-model:show="silders.isImportOpen" />
</template>
<style lang="scss" scoped>
  .create-config-btn {
    min-width: 122px;
  }
  .button-icon {
    font-size: 18px;
  }
</style>
<style lang="scss">
  .add-configs-button-popover.bk-popover.bk-pop2-content {
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
