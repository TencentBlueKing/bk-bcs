<template>
  <!-- 标签 -->
  <div class="add-label-wrap">
    <span class="label-span">{{ $t('标签') }}</span>
    <info
      class="icon-info"
      v-bk-tooltips="{
        content: $t('与分组结合使用，实现服务实例的灰度发布场景，支持多个标签；若不需要灰度发布功能，此参数可不配置'),
        placement: 'top',
      }" />
    <div class="add-label-button" @click="addItem">
      <plus class="add-label-plus" />
      {{ $t('添加') }}
    </div>
  </div>
  <div class="label-content" v-if="labelArr.length">
    <div class="label-item" v-for="(item, index) in labelArr" :key="index">
      <bk-input
        :class="['bk-input-wrap', { 'is-error': showErrorKeyValidation[index] }]"
        :id="'key' + index"
        v-model.trim="item.key"
        @blur="validateKey(index)" />
      <span v-show="showErrorKeyValidation[index]" class="error-msg">
        {{ $t("仅支持字母，数字，'-'，'_'，'.' 及 '/' 且需以字母数字开头和结尾") }}
      </span>
      <span class="label-item-icon">=</span>
      <bk-input
        :class="['bk-input-wrap', { 'is-error': showErrorValueValidation[index] }]"
        :id="'val' + index"
        v-model.trim="item.value"
        @blur="validateValue(index)" />
      <span v-show="showErrorValueValidation[index]" class="error-msg is--value">
        {{ $t("需以字母、数字开头和结尾，可包含 '-'，'_'，'.' 和字母数字及负数") }}
      </span>
      <div class="label-item-minus" @click="deleteItem(index)"></div>
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Info, Plus } from 'bkui-vue/lib/icon';

  const emits = defineEmits(['send-label']);

  const labelArr = ref<{ key: string; value: string }[]>([]);
  const showErrorKeyValidation = ref<boolean[]>([]); // key的错误状态
  const showErrorValueValidation = ref<boolean[]>([]); // value的错误状态

  const keyValidateReg = new RegExp(
    '^[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?((\\.|\\/)[a-z0-9A-Z]([-_a-z0-9A-Z]*[a-z0-9A-Z])?)*$',
  );
  const valueValidateReg = new RegExp(/^(?:-?\d+(\.\d+)?|[A-Za-z0-9]([-A-Za-z0-9_.]*[A-Za-z0-9])?)$/);

  // 数据变化后需要传递出去
  watch(labelArr.value, () => {
    sendVal();
  });

  // 所有label验证状态
  const isAllValid = () => {
    let allValid = true;
    labelArr.value.forEach((item, index) => {
      // 批量检测时，展示先校验失败的错误信息
      keyValidateReg.test(item.key) ? validateValue(index) : validateKey(index);
    });
    allValid = !showErrorKeyValidation.value.includes(true) && !showErrorValueValidation.value.includes(true);
    return allValid;
  };
  // 验证key
  const validateKey = (index: number) => {
    showErrorKeyValidation.value[index] = !keyValidateReg.test(labelArr.value[index].key);
    if (showErrorValueValidation.value[index]) {
      showErrorValueValidation.value[index] = false;
    }
  };
  // 验证value
  const validateValue = (index: number) => {
    showErrorValueValidation.value[index] = !valueValidateReg.test(labelArr.value[index].value);
    if (showErrorKeyValidation.value[index]) {
      showErrorKeyValidation.value[index] = false;
    }
  };
  // 添加项目
  const addItem = () => {
    labelArr.value.push({
      key: '',
      value: '',
    });
    showErrorKeyValidation.value.push(false);
    showErrorValueValidation.value.push(false);
  };
  // 删除点击项
  const deleteItem = (index: number) => {
    labelArr.value.splice(index, 1);
    showErrorKeyValidation.value.splice(index, 1);
    showErrorValueValidation.value.splice(index, 1);
  };
  // 数据传递
  const sendVal = () => {
    // 处理数据格式用于展示
    const newArr = labelArr.value.map((item) => {
      // let { key, value } = item;
      // key与value的输入不符合时直接为空(同步临时目录输入)
      // if (!keyValidateReg.test(labelArr.value[index].key)) {
      //   key = '';
      // }
      // if (!valueValidateReg.test(labelArr.value[index].value)) {
      //   value = '';
      // }
      return `"${item.key}":"${item.value}"`;
    });
    emits('send-label', newArr);
  };
  defineExpose({
    isAllValid,
  });
</script>

<style scoped lang="scss">
  .add-label-wrap {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    height: 20px;
  }
  .label-span {
    font-size: 12px;
    line-height: 20px;
    color: #63656e;
  }
  .icon-info {
    margin-left: 9px;
    color: #979ba5;
    cursor: pointer;
  }
  .add-label-button {
    padding: 0 9px;
    margin-left: 9px;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    font-size: 12px;
    line-height: 20px;
    color: #3a84ff;
    cursor: pointer;
    border-left: 1px solid #dcdee5;
    user-select: none;
    &:active {
      opacity: 0.6;
    }
  }
  .add-label-plus {
    margin-right: 4px;
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
    position: relative;
    width: 100%;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    .label-item-icon {
      margin: 0 4px;
      font-size: 12px;
    }
    & + .label-item {
      margin-top: 18px;
    }
    .bk-input-wrap {
      flex: 1;
      &.is-error {
        border-color: #ea3636;
        &:focus-within {
          border-color: #3a84ff;
        }
      }
    }
    .error-msg {
      position: absolute;
      left: 0;
      bottom: -14px;
      font-size: 12px;
      line-height: 1;
      white-space: nowrap;
      color: #ea3636;
      animation: form-error-appear-animation 0.15s;
      &.is--value {
        left: 50%;
      }
    }
  }
  @keyframes form-error-appear-animation {
    0% {
      opacity: 0;
      transform: translateY(-30%);
    }
    100% {
      opacity: 1;
      transform: translateY(0);
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
    &:hover {
      border-color: #3a84ff;
      &::after {
        border-color: #3a84ff;
      }
    }
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
