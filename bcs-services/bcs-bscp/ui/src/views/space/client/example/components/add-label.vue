<template>
  <!-- 标签 -->
  <div class="add-label-wrap">
    <span class="label-span">标签</span>
    <bk-popover
      content="与分组结合使用，实现服务实例的灰度发布场景，支持多个标签；若不需要灰度发布功能，此参数可不配置"
      placement="top-center">
      <info width="14" height="14" class="icon-info" />
    </bk-popover>
    <div class="add-label-button" @click="addItem">
      <plus width="12" height="12" class="add-label-plus" />
      添加
    </div>
  </div>
  <div class="label-content">
    <div class="label-item" v-for="(item, index) in labelArr" :key="index">
      <bk-input :id="'key' + index" v-model="item.key" />
      <span class="label-item-icon">=</span>
      <bk-input :id="'val' + index" v-model="item.value" />
      <div class="label-item-minus" @click="deleteItem(index)"></div>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Info, Plus } from 'bkui-vue/lib/icon';
  // import BkForm, { BkFormItem } from 'bkui-vue/lib/form';
  // import BkInput from 'bkui-vue/lib/input';
  import { debounce } from 'lodash';
  const emits = defineEmits(['send-label']);
  const labelArr = ref<{ key: string; value: string }[]>([]);
  // 添加项目
  const addItem = () => {
    const itemObj = {
      key: '',
      value: '',
    };
    labelArr.value.push(itemObj);
  };
  // 删除点击项
  const deleteItem = (index: number) => {
    labelArr.value.splice(index, 1);
  };
  // 数据传递
  const sendVal = debounce(
    () => {
      emits('send-label', labelArr);
      console.log(labelArr, '123123+++');
    },
    500,
    { leading: true },
  );
  // 数据变化后需要传递出去
  watch(labelArr.value, sendVal);
</script>

<style scoped lang="scss">
  .add-label-wrap {
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }
  .label-span {
    font-size: 12px;
    line-height: 20px;
    color: #63656e;
  }
  .icon-info {
    margin-left: 9px;
    color: #63656e;
    cursor: pointer;
  }
  .add-label-button {
    padding: 0 9px;
    margin-left: 9px;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    font-size: 12px;
    color: #3a84ff;
    cursor: pointer;
    border-left: 1px solid #dcdee5;
    user-select: none;
    &:active {
      opacity: 0.6;
    }
  }
  .add-label-plus {
    margin-right: 9px;
    line-height: 14px;
    text-align: center;
    border-radius: 50%;
    border: 1px solid #3a84ff;
  }
  .label-content {
    margin-top: 8px;
    width: 560px;
  }
  .label-item {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    .label-item-icon {
      margin: 0 4px;
      font-size: 12px;
    }
    & + .label-item {
      margin-top: 12px;
    }
  }
  .label-item-minus {
    position: relative;
    margin-left: 9px;
    flex-shrink: 0;
    width: 14px;
    height: 14px;
    border: 1px solid #979ba5;
    border-radius: 50%;
    cursor: pointer;
    &::after {
      content: '';
      position: absolute;
      left: 50%;
      top: 50%;
      transform: translate3d(-50%, -50%, 0);
      width: 8px;
      height: 0;
      border-top: 1px solid #979ba5;
    }
  }
</style>
