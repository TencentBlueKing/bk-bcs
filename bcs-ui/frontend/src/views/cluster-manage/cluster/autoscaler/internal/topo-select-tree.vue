<template>
  <TopoSelect
    :show-empty="false"
    :value="value"
    :placeholder="placeholder"
    searchable
    :remote-method="remote"
    :clearable="false"
    ref="selectRef"
    @clear="handleClearScaleOutNode">
    <bcs-big-tree
      :data="topoData"
      :options="{
        idKey: 'bk_inst_id',
        childrenKey: 'child',
        nameKey: 'path'
      }"
      ref="treeRef"
      default-expand-all
      :default-checked-nodes="[value]"
      :default-selected-node="value"
      selectable>
      <template #default="{ data }">
        <div @click="handleChangeSelectedNode(data)">
          {{ data.bk_inst_name }}
        </div>
      </template>
    </bcs-big-tree>
  </TopoSelect>
</template>
<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue';
import TopoSelect from 'bk-magic-vue/lib/select';
import { ccTopology } from '@/api/base';

export default defineComponent({
  name: 'TopoSelectTree',
  components: { TopoSelect },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: String,
      default: '',
    },
    placeholder: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const selectRef = ref<any>(null);
    const treeRef = ref<any>(null);
    const topoLoading = ref(false);
    const topoData = ref<any[]>([]);
    const addPathToTreeData = (data: any[], parent: string) => data
      .filter(item => !(item.bk_obj_id === 'set' && item.bk_obj_name === 'set'))
      .map((item) => {
        const path = parent ? `${parent} / ${item.bk_inst_name}` : item.bk_inst_name;
        return {
          ...item,
          child: addPathToTreeData(item.child, path),
          path,
        };
      });
    const handleGetTopoData = async () => {
      topoLoading.value = true;
      const data = await ccTopology({
        $clusterId: props.clusterId,
      });
      topoData.value = addPathToTreeData([data], '');
      topoLoading.value = false;
    };

    const remote = (keyword: string) => {
      treeRef.value?.filter(keyword);
    };

    const handleChangeSelectedNode = (data) => {
      if (data.bk_obj_id === 'module') {
        ctx.emit('change', String(data.bk_inst_id));
        ctx.emit('node-data-change', {
          ...data,
        });
        selectRef.value?.close();
      }
    };
    const handleClearScaleOutNode = () => {
      ctx.emit('change', '');
      ctx.emit('node-data-change', null);
    };

    onMounted(() => {
      handleGetTopoData();
    });

    return {
      selectRef,
      treeRef,
      topoData,
      remote,
      handleChangeSelectedNode,
      handleClearScaleOutNode,
    };
  },
});
</script>
