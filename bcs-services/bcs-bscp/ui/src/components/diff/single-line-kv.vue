<template>
  <section class="single-line-kv-diff">
    <div ref="containerRef" class="config-diff-list">
      <div
        v-for="diffItem in diffConfigList"
        :class="['config-diff-item', { selected: props.selectedId === diffItem.id }]"
        :key="diffItem.id">
        <div :class="['diff-header', diffItem.diffType]">
          <bk-overflow-title class="config-name" type="tips">
            {{ diffItem.name }}
          </bk-overflow-title>
        </div>
        <div class="diff-content">
          <div class="left-version-content">
            <div class="content-box">
              <div v-if="diffItem.is_secret" class="secret-content">
                <template v-if="!diffItem.secret_hidden">
                  <span>{{ diffItem.isCipherShowValue ? '********' : diffItem.base.content }}</span>
                  <div v-if="!diffItem.secret_hidden" class="actions">
                    <Unvisible
                      v-if="diffItem.isCipherShowValue"
                      class="view-icon"
                      @click="diffItem.isCipherShowValue = false" />
                    <Eye v-else class="view-icon" @click="diffItem.isCipherShowValue = true" />
                  </div>
                </template>
                <span v-else class="un-view-value">{{ $t('敏感数据不可见，无法查看实际内容') }}</span>
              </div>
              <span v-else>{{ diffItem.base.content }}</span>
            </div>
          </div>
          <div class="right-version-content">
            <div class="content-box">
              <div v-if="diffItem.is_secret" class="secret-content">
                <template v-if="!diffItem.secret_hidden">
                  <span>{{ diffItem.isCipherShowValue ? '********' : diffItem.current.content }}</span>
                  <div class="actions">
                    <Unvisible
                      v-if="diffItem.isCipherShowValue"
                      class="view-icon"
                      @click="diffItem.isCipherShowValue = false" />
                    <Eye v-else class="view-icon" @click="diffItem.isCipherShowValue = true" />
                  </div>
                </template>
                <span v-else class="un-view-value">{{ $t('敏感数据不可见，无法查看实际内容') }}</span>
              </div>
              <span v-else>{{ diffItem.current.content }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="vertical-split-line"></div>
  </section>
</template>

<script lang="ts" setup>
  import { ref, watch, onMounted } from 'vue';
  import { ISingleLineKVDIffItem } from '../../../types/service';
  import { Unvisible, Eye } from 'bkui-vue/lib/icon';

  const props = defineProps<{
    diffConfigs: ISingleLineKVDIffItem[];
    selectedId?: number;
  }>();

  const containerRef = ref();
  const diffConfigList = ref<ISingleLineKVDIffItem[]>(props.diffConfigs);

  watch(
    [() => props.diffConfigs, () => props.selectedId],
    () => {
      setScrollTop();
    },
    {
      flush: 'post',
    },
  );

  watch(
    () => props.diffConfigs,
    () => {
      diffConfigList.value = props.diffConfigs.map((item) => {
        return {
          ...item,
          isCipherShowValue: true,
        };
      });
    },
    { immediate: true },
  );

  onMounted(() => {
    setScrollTop();
  });

  const setScrollTop = () => {
    const selectedEl = containerRef.value.querySelector('.config-diff-item.selected');
    if (selectedEl) {
      containerRef.value.scrollTo(0, selectedEl.offsetTop);
    }
  };
</script>

<style scoped lang="scss">
  .single-line-kv-diff {
    position: relative;
    height: 100%;
    background: #f5f7fa;
    .vertical-split-line {
      position: absolute;
      left: calc(50% - 1px);
      top: 0;
      height: 100%;
      width: 1px;
      background: #dcdee5;
    }
  }
  .config-diff-list {
    height: 100%;
    overflow: auto;
  }
  .config-diff-item {
    margin-bottom: 12px;
    &.selected {
      .diff-content {
        background: #f0f1f5;
      }
      .content-box {
        border-color: #3a84ff;
      }
    }
  }
  .diff-header {
    padding: 2px 16px;
    background: #eaebf0;
    &.modify {
      background: #fff1db;
      .config-name {
        color: #fe9c00;
      }
    }
    &.add {
      background: #edf4ff;
      .config-name {
        color: #3a84ff;
      }
    }
    &.delete {
      background: #feebea;
      .config-name {
        color: #ea3536;
      }
    }
    .config-name {
      line-height: 20px;
      font-size: 12px;
      color: #63656e;
      max-width: 48%;
    }
  }
  .diff-content {
    display: flex;
    align-items: flex-start;
  }
  .left-version-content,
  .right-version-content {
    padding: 8px 16px 12px;
    width: 50%;
    height: 100%;
  }
  .content-box {
    display: flex;
    align-items: center;
    width: 435px;
    padding: 0 15px;
    min-height: 52px;
    line-height: 20px;
    background: #ffffff;
    border: 1px solid #c4c6cc;
    border-radius: 2px;
    font-size: 12px;
    .secret-content {
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
      .view-icon {
        cursor: pointer;
        font-size: 14px;
        color: #979ba5;
        &:hover {
          color: #3a84ff;
        }
      }
      .un-view-value {
        color: #c4c6cc;
      }
    }
  }
</style>
