<template>
  <div
    :class="[{ 'bcs-validate': isError }, type]"
    @focusin="handleFocus"
    @focusout="handleBlur">
    <slot :is-error="isError"></slot>
    <template v-if="isError">
      <span
        class="error-tip"
        v-if="errorDisplayType === 'tooltips'"
        v-bk-tooltips="errorMsg">
        <i class="bk-icon icon-exclamation-circle-shape"></i>
      </span>
      <div
        class="text-[#ea3636] text-[12px] leading-[18px]"
        v-else-if="errorDisplayType === 'normal'">
        {{ errorMsg }}
      </div>
    </template>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, PropType, ref, toRefs, watch } from 'vue';

import $i18n from '@/i18n/i18n-setup';

export interface IValidate {
  validator: Function | RegExp | string;
  message: string;
}
export default defineComponent({
  name: 'BCSValidate',
  props: {
    type: {
      type: String,
      default: 'input',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    message: {
      type: String,
      default: '',
    },
    rules: {
      type: Array as PropType<IValidate[]>,
      default: () => [],
    },
    trigger: {
      type: String,
      default: 'change',
    },
    value: {
      type: [String, Array, Object, Number],
      default: '',
    },
    meta: {
      type: [String, Array, Object, Number],
    },
    // 必填项
    required: {
      type: Boolean,
      default: false,
    },
    errorDisplayType: {
      type: String as PropType<'tooltips'|'normal'>,
      default: 'tooltips',
    },
  },
  emits: ['validate'],
  setup(props, ctx) {
    const { disabled, message, rules, value, meta, required } = toRefs(props);
    const focus = ref(false);
    function handleFocus() {
      focus.value = true;
    }
    function handleBlur() {
      focus.value = false;
      validate('blur');
    }

    async function validate(event?: string) {
      curErrMsg.value = '';
      if (required.value && !value.value && event === 'blur') {
        // 必填项校验
        curErrMsg.value = $i18n.t('generic.validate.required');
        return false;
      };
      if (!rules.value.length || !value.value) return true;

      const allPromise: Array<Promise<any>> = [];
      (rules.value as IValidate[]).forEach((item) => {
        // eslint-disable-next-line @typescript-eslint/no-misused-promises
        const promise = new Promise(async (resolve, reject) => {
          let result = false;
          if (typeof item.validator === 'function') {
            result = await item.validator(value.value, meta?.value);
          } else {
            result = new RegExp(item.validator).test(String(value.value));
          }
          if (result) {
            resolve(item);
          } else {
            reject(new Error(item.message));
          }
        });
        allPromise.push(promise);
      });

      return Promise.all(allPromise)
        .then(() => {
          curErrMsg.value = '';
          ctx.emit('validate', true);
          return true;
        })
        .catch((err) => {
          curErrMsg.value = err.message;
          ctx.emit('validate', false);
          return false;
        });
    }
    const curErrMsg = ref('');
    watch(value, () => {
      validate();
    }, { deep: true, immediate: true });

    const errorMsg = computed(() => curErrMsg.value || message.value);
    const isError = computed(() => !focus.value && !disabled.value && errorMsg.value);

    return {
      errorMsg,
      focus,
      isError,
      validate,
      handleFocus,
      handleBlur,
    };
  },
});
</script>
<style lang="postcss" scoped>
.bcs-validate {
position: relative;
.error-tip {
  font-size: 16px;
  position: absolute;
  right: 8px;
  top: 8px;
  line-height: 1;
  i {
    color: #ea3636 !important;
  }
}

&.input {
  >>> input {
    border-color: #ff5656 !important;
    color: #ff5656 !important;
  }
}
}
</style>
