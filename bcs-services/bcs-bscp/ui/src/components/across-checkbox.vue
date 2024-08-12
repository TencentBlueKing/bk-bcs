<template>
  <span class="custom-checkbox-wrap">
    <input
      type="checkbox"
      :checked="checked"
      :disabled="disabled"
      @change="handleInputChange"
      class="custom-checkbox" />
  </span>
</template>

<script lang="ts" setup>
  const props = withDefaults(
    defineProps<{
      checked: boolean;
      disabled?: boolean;
      handleChange: Function;
    }>(),
    {
      checked: false,
      disabled: false,
      handleChange: () => {},
    },
  );

  const handleInputChange = () => {
    props.handleChange();
  };
</script>

<style lang="scss" scoped>
  .custom-checkbox-wrap {
    position: relative;
    display: inline-block;
    width: 16px;
    height: 16px;
    vertical-align: sub;
    .custom-checkbox {
      position: relative;
      appearance: none;
      width: 100%;
      height: 100%;
      border: 1px solid #979ba5;
      border-radius: 2px;
      cursor: pointer;
      transition: all 0.1s;
      &:checked {
        background-color: #3a84ff;
        border-color: #3a84ff;
        // color: red;
        &::after {
          content: '';
          position: absolute;
          top: 50%;
          left: 50%;
          width: 4px;
          height: 8px;
          border: 2px solid #fff;
          border-top: 0;
          border-left: 0;

          transform: translate(-50%, -60%) scaleY(1) rotate(45deg);
          transform-origin: center;
        }
      }
      &:disabled {
        border-color: #dcdee5;
        background-color: #fafbfd;
        cursor: not-allowed;
      }
    }
  }
</style>
