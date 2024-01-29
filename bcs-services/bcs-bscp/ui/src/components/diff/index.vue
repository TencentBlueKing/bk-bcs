<template>
  <section class="diff-comp-panel">
    <div class="top-area">
      <div class="left-panel">
        <slot name="leftHead"> </slot>
      </div>
      <div class="right-panel">
        <slot name="rightHead">
          <div class="panel-name">{{ props.panelName }}</div>
        </slot>
      </div>
    </div>
    <bk-loading class="loading-wrapper" :loading="props.loading">
      <div v-if="!props.loading" class="detail-area">
        <File
          v-if="props.diff.contentType === 'file'"
          :downloadable="false"
          :current="props.diff.current.content as IFileConfigContentSummary"
          :base="props.diff.base.content as IFileConfigContentSummary"
          :id="props.id"
        />
        <Text
          v-else-if="props.diff.contentType === 'text'"
          :language="props.diff.current.language"
          :current="(props.diff.current.content as string)"
          :current-variables="props.diff.current.variables"
          :current-permission="currentPermission"
          :base="(props.diff.base.content as string)"
          :base-variables="props.diff.base.variables"
          :base-permission="basePermission"
        />
        <Kv
          v-else
          :current="(props.diff.current.content as string)"
          :base="(props.diff.base.content as string)"
        />
      </div>
    </bk-loading>
  </section>
</template>
<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { IDiffDetail } from '../../../types/service';
import { IFileConfigContentSummary } from '../../../types/config';
import File from './file.vue';
import Text from './text.vue';
import Kv from './kv.vue';

const { t } = useI18n();
const props = defineProps<{
  panelName?: String;
  diff: IDiffDetail;
  id: number; // 服务ID或模板空间ID
  loading: boolean;
}>();

const currentPermission = computed(() => {
  if (!props.diff.base.permission) return;
  return `${t('权限')}:${props.diff.current.permission?.privilege}
  ${t('用户')}:${props.diff.current.permission?.user}
  ${t('用户组')}:${props.diff.current.permission?.user_group}`;
});
const basePermission = computed(() => {
  if (!props.diff.base.permission) return;
  return `${t('权限')}:${props.diff.base.permission?.privilege}
  ${t('用户')}:${props.diff.base.permission?.user}
  ${t('用户组')}:${props.diff.base.permission?.user_group}`;
});
</script>
<style lang="scss" scoped>
.diff-comp-panel {
  height: 100%;
  border-left: 1px solid #dcdee5;
}
.top-area {
  display: flex;
  align-items: center;
  height: 49px;
  color: #313238;
  // border-bottom: 1px solid #dcdee5;
  .left-panel,
  .right-panel {
    height: 100%;
    width: 50%;
  }
  .right-panel {
    border-left: 1px solid #dcdee5;
  }
  .panel-name {
    padding: 16px;
    font-size: 12px;
    line-height: 1;
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
  }
  .config-select-area {
    display: flex;
    align-items: center;
    padding: 8px 16px;
    font-size: 12px;
  }
}
.loading-wrapper {
  height: calc(100% - 49px);
}
.detail-area {
  height: 100%;
}
</style>
