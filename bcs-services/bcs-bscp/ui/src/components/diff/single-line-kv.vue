<template>
  <section class="single-line-kv-diff">
    <div ref="containerRef" class="config-diff-list">
      <div
        v-for="diffItem in diffConfigs"
        :class="['config-diff-item', { selected: props.selectedId === diffItem.id }]"
        :key="diffItem.id">
        <div :class="['diff-header', diffItem.diffType]">
          <span class="config-name">{{ diffItem.name }}</span>
        </div>
        <div class="diff-content">
          <div class="left-version-content">
            <div class="content-box">
              {{ diffItem.base.content }}
            </div>
          </div>
          <div class="right-version-content">
            <div class="content-box">
              {{ diffItem.current.content }}
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

  const props = defineProps<{
    diffConfigs: ISingleLineKVDIffItem[];
    selectedId?: number;
  }>();

  const containerRef = ref();

  watch(
    [() => props.diffConfigs, () => props.selectedId],
    () => {
      setScrollTop();
    },
    {
      flush: 'post',
    },
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
    padding-left: 15px;
    min-height: 52px;
    line-height: 20px;
    background: #ffffff;
    border: 1px solid #c4c6cc;
    border-radius: 2px;
    font-size: 12px;
  }
</style>
