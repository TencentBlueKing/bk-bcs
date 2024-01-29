<template>
  <div
    :class="[
      'step-tab-label flex items-center',
      {
        'label-active': !disabled && (active || innerActive),
        disabled
      }
    ]"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave">
    <span class="step-number">{{ stepNum }}</span>
    <span>{{ title }}</span>
    <i
      class="bk-icon icon-exclamation-circle-shape text-[#EA3636] text-[14px] ml-[8px]"
      v-bk-tooltips="$t('tke.validate.formErr')"
      v-if="isError">
    </i>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, watch } from 'vue';

export default defineComponent({
  name: 'StepTabLabel',
  props: {
    active: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    stepNum: {
      type: Number,
    },
    title: {
      type: String,
      default: '',
    },
    isError: {
      type: Boolean,
      default: false,
    },
  },
  setup(props) {
    watch(() => props.active, () => {
      innerActive.value = props.active;
    });
    const innerActive = ref(props.active);
    const handleMouseEnter = () => {
      innerActive.value = true;
    };
    const handleMouseLeave = () => {
      innerActive.value = false;
    };

    return {
      innerActive,
      handleMouseEnter,
      handleMouseLeave,
    };
  },
});
</script>
<style lang="postcss" scoped>
.step-tab-label {
  .step-number {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border: 1px solid #979ba5;
    border-radius: 50%;
    text-align: center;
    background-color: transparent;
    margin-right: 4px;
    transform: scale(0.8);
  }
  &.label-active .step-number {
    border-color: #3a84ff;
    color: #3a84ff;
  }
  &.disabled {
    cursor: not-allowed;
    color: #C4C6CC;
    .step-number {
      border-color: #C4C6CC;
      color: #C4C6CC;
      background-color: transparent;
    }
  }
}
</style>
