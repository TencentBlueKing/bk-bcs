<template>
  <div class="more-actions">
    <Ellipsis class="ellipsis" />
    <ul class="more-actions-ul">
      <li class="more-actions-li" @click="handleUndo">{{ $t('撤销') }}</li>
    </ul>
  </div>
</template>

<script setup lang="ts">
  import { Ellipsis } from 'bkui-vue/lib/icon';
  const emits = defineEmits(['handleUndo']);

  const handleUndo = () => {
    emits('handleUndo', 'undo');
  };
</script>

<style lang="scss" scoped>
  .more-actions {
    box-sizing: content-box;
    position: absolute;
    right: 0;
    top: -5px;
    display: inline-block;
    vertical-align: middle;
    width: 32px;
    height: 32px;
    cursor: pointer;
    &:hover {
      .more-actions-ul {
        display: block;
      }
      .ellipsis {
        color: #3a84ff;
      }
      &::after {
        background-color: #dcdee5;
      }
    }
    &::after {
      content: '';
      position: absolute;
      left: 50%;
      top: 50%;
      transform: translate(-50%, -50%);
      width: 20px;
      height: 20px;
      border-radius: 50%;
    }
    .ellipsis {
      position: absolute;
      z-index: 1;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%) rotate(90deg);
      font-size: 16px;
      font-weight: 700;
      color: #9a9fa9;
    }
  }
  .more-actions-ul {
    position: absolute;
    right: 7px;
    top: 32px;
    display: none;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    box-shadow: 0 2px 6px 0 #0000001a;
  }
  .more-actions-li {
    position: relative;
    z-index: 2;
    padding: 4px 12px;
    min-width: 62px;
    line-height: 32px;
    font-size: 12px;
    color: #63656e;
    background-color: #fff;
    &:hover {
      background-color: #f5f7fa;
    }
  }
</style>
