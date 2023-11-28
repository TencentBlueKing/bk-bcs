<template>
  <TopoSelect
    :show-empty="false"
    :value="value"
    :placeholder="placeholder"
    searchable
    :remote-method="remote"
    :clearable="false"
    ref="selectRef"
    :loading="topoLoading"
    @clear="handleClearScaleOutNode">
    <bcs-big-tree
      :data="topoData"
      :options="{
        idKey: 'instanceId',
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
          {{ data.instanceName }}
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

import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';

import { topologyHostCount } from '@/api/modules/cluster-manager';
import $store from '@/store';

interface ITopoData {
  count: number
  expanded: boolean
  instanceId: number
  instanceName: string
  objectId: string
  objectName: string
  child: ITopoData[]
}
export default defineComponent({
  name: 'TopoSelectTree',
  components: { TopoSelect, SelectExtension },
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
    const topoData = ref<Array<ITopoData>>([]);
    const addPathToTreeData = (data: ITopoData[], parent: string) => data
      .map((item) => {
        const path = parent ? `${parent} / ${item.instanceName}` : item.instanceName;
        return {
          ...item,
          child: addPathToTreeData(item.child, path),
          path,
        };
      });
    const handleGetTopoData = async () => {
      topoLoading.value = true;
      const data = await topologyHostCount({
        $biz: $store.state.curProject.businessID,
        $scope: 'biz',
        scopeList: [
          {
            scopeType: 'biz',
            scopeId: $store.state.curProject.businessID,
          },
        ],
      });
      topoData.value = addPathToTreeData(data, '');
      topoLoading.value = false;
    };

    const remote = (keyword: string) => {
      treeRef.value?.filter(keyword);
    };

    const handleChangeSelectedNode = (data: ITopoData) => {
      if (data.objectId === 'module') {
        ctx.emit('change', String(data.instanceId));
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
