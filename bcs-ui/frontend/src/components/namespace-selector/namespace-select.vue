<template>
  <bcs-select
    class="bg-[#fff]"
    :value="value"
    :clearable="clearable"
    :searchable="searchable"
    :disabled="disabled"
    :loading="namespaceLoading || loading"
    :placeholder="$t('请选择命名空间')"
    @change="handleNamespaceChange">
    <bcs-option
      v-for="option in nsList"
      :key="option.name"
      :id="option.name"
      :name="option.name"
    ></bcs-option>
  </bcs-select>
</template>
<script lang="ts">
import { defineComponent, watch, toRefs, computed } from 'vue';
import { useSelectItemsNamespace } from '@/views/resource-view/namespace/use-namespace';
import $store from '@/store';

export default defineComponent({
  name: 'NamespaceSelect',
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    value: {
      type: String,
      default: '',
    },
    searchable: {
      type: Boolean,
      default: true,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    clearable: {
      type: Boolean,
      default: false,
    },
    // 必须存在namespace
    required: {
      type: Boolean,
      default: false,
    },
    // 数据源
    list: {
      type: Array,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  setup(props, ctx) {
    const { clusterId, value, required, list } = toRefs(props);
    const { namespaceLoading, namespaceList, getNamespaceData } = useSelectItemsNamespace();

    watch(clusterId,  () => {
      !list?.value && handleGetNsData();
    });

    const nsList = computed(() => list?.value || namespaceList.value);

    const handleNamespaceChange = (name) => {
      if (value.value === name) return;

      $store.commit('updateCurNamespace', name);
      ctx.emit('change', name);
    };

    const handleGetNsData = async () => {
      if (!clusterId.value) return;
      await getNamespaceData({
        clusterId: clusterId.value,
      });
      if (!namespaceList.value.find(item => item.name === value.value)) {
        handleNamespaceChange(required.value ? namespaceList.value[0]?.name : '');
      } else {
        $store.commit('updateCurNamespace', value.value);
      }
    };

    !list?.value && handleGetNsData();

    return {
      namespaceLoading,
      nsList,
      handleNamespaceChange,
    };
  },
});
</script>
