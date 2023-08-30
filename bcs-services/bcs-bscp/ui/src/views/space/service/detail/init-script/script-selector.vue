<script lang="ts" setup>
  import { useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { RightTurnLine } from 'bkui-vue/lib/icon';
  import { useGlobalStore } from '../../../../../store/global'

  const router = useRouter()
  const { spaceId } = storeToRefs(useGlobalStore())

  const props = defineProps<{
    id: number;
    type: 'pre'|'post';
    disabled: boolean;
    loading: boolean;
    list: { id: number; versionId: number; name: string; type: string; }[]; // 脚本列表
  }>()

  const emits = defineEmits(['change', 'refresh'])

  // 处理option中disabled状态的点击事件
  const handleScriptOptionClick = (id: number, versionId: number, e: Event) => {
    if (id && !versionId) {
      e.stopPropagation()
    }
  }

  const goToScriptList = () => {
    const { href } = router.resolve({ name: 'script-list', params: { spaceId: spaceId.value } })
    window.open(href, '_blank')
  }

</script>
<template>
  <bk-select
    :model-value="props.id"
    :popover-options="{ theme: 'light bk-select-popover script-select-popover' }"
    :clearable="false"
    :disabled="props.disabled"
    :loading="props.loading"
    @change="emits('change', $event, props.type)">
    <bk-option
      v-for="script in props.list"
      :class="['script-option-item', { disabled: script.id && !script.versionId }]"
      :key="script.id"
      :value="script.id"
      :label="script.name">
      <div
        v-bk-tooltips="{ disabled: !script.id || script.versionId, content: '该脚本未上线' }"
        class="option-wrapper"
        @click="handleScriptOptionClick(script.id, script.versionId, $event)">
        {{ script.name }}
      </div>
    </bk-option>
    <template #extension>
      <div class="selector-extension" @click="goToScriptList">
        <i class="bk-bscp-icon icon-setting"></i>
        <span>脚本管理</span>
        <div class="refresh-area">
          <RightTurnLine class="refresh-icon" @click.stop="emits('refresh')" />
        </div>
      </div>
    </template>
  </bk-select>
</template>
<style lang="scss" scoped>
  .selector-extension {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    color: #63656e;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
      .icon-setting {
        color: #3a84ff;
      }
    }
    .icon-setting {
      margin-right: 4px;
      color: #979ba5;
    }
    .refresh-area {
      position: absolute;
      top: 12px;
      right: 0;
      height: 16px;
      width: 48px;
      font-size: 12px;
      border-left: 1px solid #dcdee5;
      text-align: center;
      vertical-align: middle;
      .refresh-icon {
        font-size: 16px;
        color: #979ba5;
        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
</style>
<style lang="scss">
  .script-select-popover.bk-popover.bk-select-popover .bk-select-content-wrapper {
    .bk-select-option.script-option-item {
      padding: 0;
      &.disabled {
        cursor: not-allowed;
        .option-wrapper {
          color: #dcdee5 !important;
        }
      }
      .option-wrapper {
        padding: 0 12px;
      }
    }
  }
</style>
