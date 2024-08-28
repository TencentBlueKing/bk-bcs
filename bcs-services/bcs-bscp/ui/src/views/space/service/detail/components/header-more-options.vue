<template>
  <div class="more-options">
    <Ellipsis class="ellipsis" />
    <ul class="more-options-ul">
      <li class="more-options-li" @click="handleLinkTo">服务上线记录</li>
      <li class="more-options-li" v-if="props.approveStatus === APPROVE_STATUS.PendApproval" @click="handleUndo">
        撤销
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
  import { Ellipsis } from 'bkui-vue/lib/icon';
  import { APPROVE_STATUS } from '../../../../../constants/config';

  const props = withDefaults(
    defineProps<{
      approveStatus: string;
    }>(),
    {
      approveStatus: '',
    },
  );

  const handleLinkTo = () => {
    console.log('跳转服务上线记录'); // 等接口更新
  };
  const handleUndo = () => {
    console.log('撤销'); // 等接口更新
  };
</script>

<style lang="scss" scoped>
  .more-options {
    box-sizing: content-box;
    position: relative;
    margin: 0 -12px 0 0;
    width: 32px;
    height: 32px;
    cursor: pointer;
    &:hover {
      .more-options-ul {
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
  .more-options-ul {
    position: absolute;
    z-index: 1;
    right: 7px;
    top: 32px;
    display: none;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    box-shadow: 0 2px 6px 0 #0000001a;
  }
  .more-options-li {
    padding: 0 12px;
    min-width: 96px;
    line-height: 32px;
    font-size: 12px;
    color: #63656e;
    background-color: #fff;
    &:hover {
      background-color: #f5f7fa;
    }
  }
</style>
