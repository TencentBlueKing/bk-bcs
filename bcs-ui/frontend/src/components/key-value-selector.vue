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
        :value="(key in labelValue)"
        @change="handleLabelChange(key, labels[key])">
      </bk-checkbox>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, ref, watch } from 'vue';

type DataSourceItem = {
  label: string;
  value: string;
};
export default defineComponent({
  props: {
    datasource: {
      type: [Object, Array<DataSourceItem>],
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
    const labels = computed(() => {
      if (Array.isArray(props.datasource)) {
        return (props.datasource as DataSourceItem[]).reduce((pre, item) => {
          pre[item.label] = item.value;
          return pre;
        }, {});
      }
      return props.datasource;
    });
    const labelValue = ref({});

    const handleLabelChange = (key: string, value: string) => {
      if (key in labelValue.value) {
        delete labelValue.value[key];
      } else {
        labelValue.value[key] = value;
      }
      emit('change', labelValue.value);
    };

    watch(() => props.value, (value) => {
      labelValue.value = JSON.parse(JSON.stringify(value));
    }, { immediate: true });

    return {
      labels,
      labelValue,
      handleLabelChange,
    };
  },
});
</script>
