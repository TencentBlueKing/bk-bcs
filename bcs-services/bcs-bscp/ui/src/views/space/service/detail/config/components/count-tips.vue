<template>
  <bk-popover placement="bottom-start" theme="light">
    <div :class="['tips-wrap', { full: current === max }]">
      <Info class="tips-icon" />
      {{ `${current} / ${max}` }}
    </div>
    <template #content>
      <div>
        {{
          $t(tips, {
            n: max,
            m: current,
            p: max - current,
          })
        }}
      </div>
    </template>
  </bk-popover>
</template>

<script lang="ts" setup>
  import { Info } from 'bkui-vue/lib/icon';
  import { computed } from 'vue';

  const props = defineProps<{
    max: number;
    current: number;
    isTemp: boolean;
    isFileType: boolean;
  }>();

  const tips = computed(() => {
    if (props.isTemp) {
      return '为确保最佳用户体验，此服务的模板文件数量限制为 {n} 个。当前已有 {m} 个，还可添加 {p} 个';
    }
    if (props.isFileType) {
      return '为确保最佳用户体验，此服务的配置文件数量限制为 {n} 个。当前已有 {m} 个，还可添加 {p} 个';
    }
    return '为确保最佳用户体验，此服务的配置项数量限制为 {n} 个。当前已有 {m} 个，还可添加 {p} 个';
  });
</script>

<style scoped lang="scss">
  .tips-wrap {
    display: flex;
    align-items: center;
    height: 32px;
    background: #f0f5ff;
    border-radius: 2px;
    padding: 5px 8px 7px;
    font-size: 12px;
    color: #63656e;
    cursor: pointer;
    .tips-icon {
      font-size: 14px;
      color: #699df4;
      margin-right: 5px;
    }
    &.full {
      background: #fff3e1;
      .tips-icon {
        color: #ff9c01;
      }
    }
  }
</style>
