<template>
  <TopoSelect
    :show-empty="false"
    :value="value"
    :placeholder="placeholder"
    searchable
    :remote-method="remote"
    :clearable="false"
    :loading="topoLoading"
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
    <template slot="extension">
      <SelectExtension
        :link-text="$t('tke.link.cmdb')"
        :link="PROJECT_CONFIG.cmdbhost"
        @refresh="handleGetTopoData" />
    </template>
  </TopoSelect>
</template>
<script lang="ts">
import TopoSelect from 'bk-magic-vue/lib/select';
import { defineComponent, onMounted, ref } from 'vue';

import { ccTopology } from '@/api/base';
import { useProject } from '@/composables/use-app';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';

export default defineComponent({
  name: 'TopoSelectTree',
  components: { TopoSelect, SelectExtension },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: [String, Number],
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
      .map((item) => {
        const path = parent ? `${parent} / ${item.bk_inst_name}` : item.bk_inst_name;
        return {
          ...item,
          child: addPathToTreeData(item.child, path),
          path,
        };
      });
    const { curProject } = useProject();
    const handleGetTopoData = async () => {
      topoLoading.value = true;
      const data = await ccTopology({
        $clusterId: props.clusterId || '-',
        bizID: curProject.value.businessID,
        filterInter: true,
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
      topoLoading,
      selectRef,
      treeRef,
      topoData,
      remote,
      handleChangeSelectedNode,
      handleClearScaleOutNode,
      handleGetTopoData,
    };
  },
});
</script>
