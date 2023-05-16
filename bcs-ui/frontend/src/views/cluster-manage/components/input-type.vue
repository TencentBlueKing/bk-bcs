<template>
  <div class="input-type" v-bk-clickoutside="handleBlur">
    <bcs-select
      v-if="options.length"
      v-model="inputValue"
      ref="selectRef"
      @change="handleValueChange">
      <bcs-option
        v-for="item in options"
        :key="item"
        :id="item"
        :name="item">
      </bcs-option>
    </bcs-select>
    <template v-else>
      <bk-input
        v-if="type === 'string'"
        v-model="inputValue"
        ref="inputRef"
        @change="handleValueChange"
        @enter="handleEnter">
      </bk-input>
      <bk-input
        v-else-if="type === 'int'"
        type="number"
        v-model="inputValue"
        :min="range.min"
        :max="range.max"
        ref="inputRef"
        @change="handleValueChange"
        @enter="handleEnter">
      </bk-input>
      <bk-checkbox
        v-else-if="type === 'bool'"
        v-model="inputValue"
        @change="handleValueChange">
      </bk-checkbox>
      <bk-input
        v-else
        v-model="inputValue"
        ref="inputRef"
        @change="handleValueChange"
        @enter="handleEnter">
      </bk-input>
    </template>

  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref, getCurrentInstance } from 'vue';

export default defineComponent({
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    type: {
      type: String,
      default: '',
    },
    value: {
      type: [String, Number],
      default: '',
    },
    options: {
      type: Array,
      default: () => ([]),
    },
    range: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props, ctx) {
    const inputValue = ref<string|number>(props.value);
    const popup = ref();
    const handleValueChange = (val) => {
      ctx.emit('change', val);
    };
    const handleBlur = (mouseup) => {
      console.log(mouseup.target);
      ctx.emit('blur');
    };
    const handleEnter = () => {
      ctx.emit('enter');
    };
    const { proxy } = getCurrentInstance() || { proxy: null };
    const focus = () => {
      const $refs = proxy?.$refs || {};
      setTimeout(() => {
        $refs.inputRef && ($refs.inputRef as any).focus();
      }, 0);
    };

    onMounted(() => {
      const $refs = proxy?.$refs || {};
      const popper = ($refs.selectRef as any)?.$refs?.selectDropdown?.instance?.popper;
      popup.value = popper;
    });

    return {
      inputValue,
      handleValueChange,
      handleBlur,
      handleEnter,
      focus,
      popup,
    };
  },
});
</script>
<style lang="postcss" scoped>
.input-type {
    width: 100%;
    background: #fff;
}
</style>
