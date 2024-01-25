<template>
  <bk-select
    :model-value="props.id"
    :popover-options="{ theme: 'light bk-select-popover script-select-popover' }"
    :clearable="false"
    :disabled="props.disabled"
    :loading="props.loading"
    filterable
    :input-search="false"
    @change="emits('change', $event, props.type)">
    <bk-option
      v-for="script in props.list"
      :class="['script-option-item', { disabled: script.id && !script.versionId }]"
      :key="script.id"
      :value="script.id"
      :label="script.name">
      <div
        v-bk-tooltips="{ disabled: !script.id || script.versionId, content: t('该脚本未上线') }"
        class="option-wrapper"
        @click="handleScriptOptionClick(script.id, script.versionId, $event)">
        {{ script.name }}
      </div>
    </bk-option>
    <template #extension>
      <div class="selector-extension" @click="goToScriptList">
        <div class="selector-script">
          <i class="bk-bscp-icon icon-setting"></i>
          <span>{{t('脚本管理')}}</span>
        </div>
        <div class="refresh-area" @click.stop="emits('refresh')">
          <RightTurnLine class="refresh-icon" />
        </div>
      </div>
    </template>
  </bk-select>
</template>
<script lang="ts" setup>
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { storeToRefs } from 'pinia';
import { RightTurnLine } from 'bkui-vue/lib/icon';
import useGlobalStore from '../../../../../store/global';

const router = useRouter();
const { spaceId } = storeToRefs(useGlobalStore());
const { t } = useI18n();

const props = defineProps<{
  id: number;
  type: 'pre' | 'post';
  disabled: boolean;
  loading: boolean;
  list: { id: number; versionId: number; name: string; type: string }[]; // 脚本列表
}>();

const emits = defineEmits(['change', 'refresh']);

// 处理option中disabled状态的点击事件
const handleScriptOptionClick = (id: number, versionId: number, e: Event) => {
  if (id && !versionId) {
    e.stopPropagation();
  }
};

const goToScriptList = () => {
  const { href } = router.resolve({ name: 'script-list', params: { spaceId: spaceId.value } });
  window.open(href, '_blank');
};
</script>
<style lang="scss" scoped>
.selector-extension {
  position: relative;
  width: 100%;
  height: 100%;
  .selector-script {
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
  }

  .icon-setting {
    margin-right: 4px;
    color: #979ba5;
  }
  .refresh-area {
    display: flex;
    align-items: center;
    justify-content: center;
    position: absolute;
    top: 0;
    right: 0;
    height: 100%;
    width: 48px;
    font-size: 12px;
    cursor: pointer;
    &:hover .refresh-icon {
      color: #3a84ff;
    }
    &::before {
      display: block;
      content: '';
      position: absolute;
      left: 0;
      width: 1px;
      height: 16px;
      background-color: #dcdee5;
    }
    .refresh-icon {
      font-size: 16px;
      color: #979ba5;
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
      width: 100%;
    }
  }
}
</style>
