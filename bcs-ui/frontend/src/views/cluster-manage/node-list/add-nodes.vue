<template>
  <div class="bcs-content-wrapper p-[24px] text-[12px]">
    <bk-form class="bg-[#fff] py-[20px]">
      <bk-form-item :label="$t('manualNode.title.source.text')">
        <bk-radio-group v-model="nodeSource">
          <bk-radio
            value="ip"
            :disabled="isKubeConfigImportCluster"
            v-if="['tencentCloud','tencentPublicCloud','bluekingCloud'].includes(provider)">
            <span
              v-bk-tooltips="{
                disabled: !isKubeConfigImportCluster,
                content: $t('cluster.nodeList.tips.disableImportClusterAddNode')
              }">
              {{ $t('manualNode.title.source.existingServer') }}
            </span>
          </bk-radio>
          <bk-radio
            value="nodePool"
            v-if="curCluster && curCluster.autoScale && provider !== 'bluekingCloud'">
            {{ $t('manualNode.title.source.addFromNodePool') }}
          </bk-radio>
        </bk-radio-group>
      </bk-form-item>
    </bk-form>
    <AddNodesFromCmdb :cluster-id="clusterId" v-if="nodeSource === 'ip'" />
    <AddNodesFromNodepool :cluster-id="clusterId" :node-pool="nodePool" v-else-if="nodeSource === 'nodePool'" />
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onBeforeMount, PropType, ref } from 'vue';

import AddNodesFromCmdb from './add-nodes-from-cmdb.vue';
import AddNodesFromNodepool from './add-nodes-from-nodepool.vue';

import { ICluster } from '@/composables/use-app';
import $store from '@/store/index';

export default defineComponent({
  components: { AddNodesFromNodepool, AddNodesFromCmdb },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
    source: {
      type: String as PropType<'nodePool' | 'ip'>,
      default: '',
    },
    nodePool: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const curCluster = computed<ICluster>(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId) || {});
    const isKubeConfigImportCluster = computed(() => curCluster.value.clusterCategory === 'importer'
      && curCluster.value.importCategory === 'kubeConfig');
    const provider = computed(() => curCluster.value.provider);

    const nodeSource = ref<'nodePool'|'ip'>('ip');
    const setDefaultNodeSource = () => {
      switch (provider.value) {
        case 'bluekingCloud':
          nodeSource.value = 'ip';// bluekingCloud 只能添加IP节点
          break;
        case 'gcpCloud':
          nodeSource.value = 'nodePool';// gcpCloud 只能通过节点池添加节点
          break;
        default:
          nodeSource.value = props.source || 'nodePool';// 默认推荐节点池方式添加
      }
    };

    onBeforeMount(() => {
      setDefaultNodeSource();
    });

    return {
      isKubeConfigImportCluster,
      provider,
      curCluster,
      nodeSource,
    };
  },
});
</script>
<style lang="postcss" scoped>
.choose-node-template {
  padding: 24px;
  >>> .choose-node {
    .form-group-content {
      padding-top: 0;
    }
  }
}
</style>
