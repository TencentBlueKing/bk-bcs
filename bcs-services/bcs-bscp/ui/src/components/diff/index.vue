<template>
  <Teleport :disabled="!isFullScreen" to="body">
    <section :class="['diff-comp-panel', { fullscreen: isFullScreen }]">
      <div class="top-area">
        <div class="left-panel">
          <slot name="leftHead"> </slot>
        </div>
        <div class="right-panel">
          <slot name="rightHead">
            <div class="panel-name">{{ props.panelName }}</div>
          </slot>
          <div v-if="props.diff.contentType === 'text'" class="action-btn">
            <template v-if="props.diff.is_secret && !props.diff.secret_hidden">
              <Unvisible v-if="isCipherShowSecret" class="view-icon" @click="isCipherShowSecret = false" />
              <Eye v-else class="view-icon" @click="isCipherShowSecret = true" />
            </template>
            <FilliscreenLine
              v-if="!isFullScreen"
              v-bk-tooltips="{
                content: $t('全屏'),
                placement: 'top',
                distance: 20,
              }"
              @click="handleOpenFullScreen" />
            <UnfullScreen
              v-else
              v-bk-tooltips="{
                content: $t('退出全屏'),
                placement: 'bottom',
                distance: 20,
              }"
              @click="handleCloseFullScreen" />
          </div>
        </div>
      </div>
      <bk-loading class="loading-wrapper" :loading="props.loading">
        <div v-if="!props.loading" class="detail-area">
          <File
            v-if="props.diff.contentType === 'file'"
            :downloadable="false"
            :current="props.diff.current.content as IFileConfigContentSummary"
            :base="props.diff.base.content as IFileConfigContentSummary"
            :id="props.id" />
          <Text
            v-else-if="props.diff.contentType === 'text'"
            :language="props.diff.current.language"
            :current="props.diff.current.content as string"
            :current-variables="props.diff.current.variables"
            :current-permission="currentPermission"
            :base="props.diff.base.content as string"
            :base-variables="props.diff.base.variables"
            :base-permission="basePermission"
            :is-secret="props.diff.is_secret"
            :secret-visible="!props.diff.secret_hidden"
            :is-cipher-show="isCipherShowSecret" />
          <template v-else-if="props.diff.contentType === 'singleLineKV'">
            <SingleLineKV
              v-if="props.diff.singleLineKVDiff"
              :diff-configs="props.diff.singleLineKVDiff"
              :selected-id="props.selectedKvConfigId" />
          </template>
        </div>
      </bk-loading>
    </section>
  </Teleport>
</template>
<script setup lang="ts">
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { FilliscreenLine, UnfullScreen, Unvisible, Eye } from 'bkui-vue/lib/icon';
  import BkMessage from 'bkui-vue/lib/message';
  import { IDiffDetail } from '../../../types/service';
  import { IFileConfigContentSummary } from '../../../types/config';
  import File from './file.vue';
  import Text from './text.vue';
  import SingleLineKV from './single-line-kv.vue';

  const { t } = useI18n();
  const props = defineProps<{
    panelName?: String;
    diff: IDiffDetail;
    id: number; // 服务ID或模板空间ID
    loading: boolean;
    selectedKvConfigId?: number; // 选中的kv类型配置id
  }>();

  const isFullScreen = ref(false);

  const isCipherShowSecret = ref(true);

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

  // 打开全屏
  const handleOpenFullScreen = () => {
    isFullScreen.value = true;
    window.addEventListener('keydown', handleEscClose, { once: true });
    BkMessage({
      theme: 'primary',
      message: t('按 Esc 即可退出全屏模式'),
    });
  };

  const handleCloseFullScreen = () => {
    isFullScreen.value = false;
    window.removeEventListener('keydown', handleEscClose);
  };

  // Esc按键事件处理
  const handleEscClose = (event: KeyboardEvent) => {
    if (event.code === 'Escape') {
      isFullScreen.value = false;
    }
  };
</script>
<style lang="scss" scoped>
  .diff-comp-panel {
    height: 100%;
    border-left: 1px solid #dcdee5;
    &.fullscreen {
      position: fixed;
      top: 0;
      left: -1px;
      width: 100vw;
      height: 100vh;
      z-index: 5000;
      .top-area {
        background: #313238;
      }
      .right-panel {
        border-color: #1d1d1d;
      }
      .fullscreen-btn {
        z-index: 1;
      }
    }
  }
  .top-area {
    display: flex;
    align-items: center;
    height: 49px;
    color: #313238;
    .left-panel,
    .right-panel {
      height: 100%;
      width: 50%;
      background-color: #313238;
      :deep(.bk-select) {
        .bk-input {
          border-color: #63656e;
        }
        .bk-input--text {
          color: #b6b6b6;
          background: #313238;
        }
      }
    }
    .right-panel {
      position: relative;
      border-left: 1px solid #1d1d1d;
      .action-btn {
        display: flex;
        gap: 8px;
        position: absolute;
        top: 16px;
        right: 14px;
        span {
          cursor: pointer;
          color: #979ba5;
          &:hover {
            color: #3a84ff;
          }
        }
      }
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
