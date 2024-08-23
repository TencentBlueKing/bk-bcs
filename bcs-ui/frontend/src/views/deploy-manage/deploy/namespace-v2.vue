<template>
  <div class="flex">
    <bcs-compose-form-item class="flex-1 flex items-center">
      <!-- 目前只支持集群模式 -->
      <bcs-select value="cluster" class="min-w-[88px]" :clearable="false">
        <bcs-option id="cluster" :name="$t('templateFile.label.byCluster')"></bcs-option>
      </bcs-select>
      <ClusterSelect class="!w-auto flex-1" v-model="clusterID" cluster-type="all" @change="handleClusterChange" />
    </bcs-compose-form-item>
    <NamespaceSelect
      :cluster-id="clusterID"
      class="flex-1 ml-[10px]"
      :clearable="true"
      v-model="ns" />
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import $store from '@/store';

type Emits = (e: 'change', clusterID: string, ns: string) => void;
const emits = defineEmits<Emits>();

const clusterID = ref($store.getters.curClusterId);
const ns = ref('');

function handleClusterChange() {
  ns.value = '';
}

watch(ns, () => {
  emits('change', clusterID.value, ns.value);
});
</script>
