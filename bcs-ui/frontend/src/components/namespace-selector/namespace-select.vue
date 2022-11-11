<template>
  <bcs-select
    class="w-[250px]"
    v-model="localValue"
    :clearable="false"
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
<script>
import { defineComponent, ref, onMounted } from '@vue/composition-api';
import { CUR_SELECT_NAMESPACE } from '@/common/constant';
import useDefaultClusterId from '@/views/node/use-default-clusterId';
import { useSelectItemsNamespace } from '@/views/dashboard/namespace/use-namespace';


export default defineComponent({
  name: 'NamespaceSelect',
  props: {
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
  },
  setup(props, ctx) {
    const localValue = ref(props.value);

    const { defaultClusterId } = useDefaultClusterId();
    const { namespaceLoading, namespaceList, getNamespaceData } = useSelectItemsNamespace();

    const handleNamespaceChange = (name) => {
      localValue.value = name;
      ctx.emit('change', name);
      localStorage.setItem(`${defaultClusterId.value}-${CUR_SELECT_NAMESPACE}`, name);
    };

    onMounted(() => {
      getNamespaceData({
        clusterId: defaultClusterId.value,
      });
    });

    return {
      localValue,
      namespaceLoading,
      namespaceList,
      handleNamespaceChange,
    };
  },
});
</script>
