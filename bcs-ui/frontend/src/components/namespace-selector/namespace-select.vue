<template>
  <bcs-select
    class="bg-[#fff]"
    :value="value"
    :clearable="clearable"
    :searchable="searchable"
    :disabled="disabled"
    :loading="namespaceLoading || loading"
    :placeholder="$t('dashboard.ns.validate.emptyNs')"
    @change="handleNamespaceChange">
    <bcs-option
      v-for="option in nsList"
      :key="option.name"
      :id="option.name"
      :name="option.name"
    ></bcs-option>
    <template #extension>
      <SelectExtension
        :link-text="$t('dashboard.ns.create.title')"
        @link="handleGotoNs"
        @refresh="handleGetNsData" />
    </template>
  </bcs-select>
</template>
<script lang="ts">
import { computed, defineComponent, PropType, toRefs, watch } from 'vue';

import $router from '@/router';
import $store from '@/store';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';
import { useSelectItemsNamespace } from '@/views/cluster-manage/namespace/use-namespace';

export default defineComponent({
  name: 'NamespaceSelect',
  components: { SelectExtension },
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
      type: Array as PropType<any[]>,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    updateStorage: {
      type: Boolean,
      default: true,
    },
  },
  setup(props, ctx) {
    const { clusterId, value, required, list, updateStorage } = toRefs(props);
    const { namespaceLoading, namespaceList, getNamespaceData } = useSelectItemsNamespace();

    watch(clusterId,  () => {
      !list?.value && handleGetNsData();
    });

    const nsList = computed(() => list?.value || namespaceList.value);

    const handleNamespaceChange = (name) => {
      if (value.value === name) return;

      updateStorage.value && $store.commit('updateCurNamespace', name);
      ctx.emit('change', name);
    };

    // 创建命名空间
    const handleGotoNs = () => {
      const { href } = $router.resolve({
        name: 'createNamespace',
        params: {
          clusterId: props.clusterId,
        },
      });
      window.open(href);
    };

    // 获取命名空间数据
    const handleGetNsData = async () => {
      if (!clusterId.value) return;
      await getNamespaceData({
        clusterId: clusterId.value,
      }, required.value);
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
      handleGotoNs,
      handleGetNsData,
    };
  },
});
</script>
