<template>
  <bcs-select
    class="bg-[#fff]"
    :value="value"
    :clearable="clearable"
    :searchable="searchable"
    :disabled="disabled"
    :loading="namespaceLoading"
    :placeholder="$t('请选择命名空间')"
    @change="handleNamespaceChange">
    <bcs-option
      v-for="option in namespaceList"
      :key="option.name"
      :id="option.name"
      :name="option.name"
    ></bcs-option>
  </bcs-select>
</template>
<script lang="ts">
import { defineComponent, onMounted, watch, toRefs } from '@vue/composition-api';
import useDefaultClusterId from '@/views/node/use-default-clusterId';
import { useSelectItemsNamespace } from '@/views/dashboard/namespace/use-namespace';
import { CUR_SELECT_NAMESPACE } from '@/common/constant';


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
  },
  setup(props, ctx) {
    const { clusterId } = toRefs(props);
    const { defaultClusterId } = useDefaultClusterId();
    const { namespaceLoading, namespaceList, getNamespaceData } = useSelectItemsNamespace();

    watch(clusterId,  () => {
      getNamespaceData({
        clusterId: clusterId.value || defaultClusterId.value,
      });
    });

    const handleNamespaceChange = (name) => {
      ctx.emit('change', name);
      sessionStorage.setItem(CUR_SELECT_NAMESPACE, name);
    };

    onMounted(() => {
      getNamespaceData({
        clusterId: clusterId.value || defaultClusterId.value,
      });
    });

    return {
      namespaceLoading,
      namespaceList,
      handleNamespaceChange,
    };
  },
});
</script>
