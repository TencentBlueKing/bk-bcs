<template>
  <div class="bcs-sideslider-content">
    <div class="bcs-detail-item">
      <div class="bcs-detail-item-label">{{}}</div>
      <div class="bcs-detail-item-content"></div>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, onMounted, ref } from '@vue/composition-api';
import useFormLabel from '@/common/use-form-label';

export default defineComponent({
  name: 'ChartDetail',
  props: {
    chart: {
      type: Object,
      default: () => ({}),
    },
  },
  setup() {
    const formRef = ref(null);
    const { labelWidth, initFormLabelWidth } = useFormLabel();
    onMounted(() => {
      initFormLabelWidth(formRef.value);
    });

    return {
      labelWidth,
      formRef,
    };
  },
});
</script>
