<template>
  <div>
    <div
      class="flex items-center mb-[10px]"
      v-for="key in Object.keys(labels)" :key="key">
      <bk-input class="w-[230px]" :value="key" disabled></bk-input>
      <span class="mx-[15px] text-[#c3cdd7]">=</span>
      <bk-input class="w-[230px]" :value="labels[key]" disabled></bk-input>
      <bk-checkbox
        class="ml-[10px]"
        :value="(key in value)"
        @change="handleLabelChange(key, labels[key])">
      </bk-checkbox>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, ref, watch } from 'vue';

export default defineComponent({
  props: {
    labels: {
      type: Object,
      default: () => ({}),
      required: true,
    },
    value: {
      type: Object,
      default: () => ({}),
      required: true,
    },
  },
  emits: ['change'],
  setup(props, { emit }) {
    const labelValue = ref({});

    const handleLabelChange = (key, value) => {
      if (key in labelValue.value) {
        delete labelValue.value[key];
      } else {
        labelValue.value[key] = value;
      }
      emit('change', labelValue.value);
    };

    watch(() => props.value, (value) => {
      labelValue.value = JSON.parse(JSON.stringify(value));
    });

    return {
      handleLabelChange,
    };
  },
});
</script>
