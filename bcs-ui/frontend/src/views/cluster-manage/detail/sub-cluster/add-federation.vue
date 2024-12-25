<template>
  <div class="p-[24px] text-[12px] overflow-auto">
    <bk-form class="bg-[#fff] py-[20px]" ref="formRef">
      <bk-form-item :label="$t('cluster.create.label.hostResource')">
        <bk-radio-group v-model="nodeSource">
          <bk-radio value="create" disabled>
            <span
              v-bk-tooltips="{
                content: $t('cluster.create.federation.createSubCluster')
              }">
              {{ $t('manualNode.title.source.existingServer') }}
            </span>
          </bk-radio>
          <bk-radio value="existing">
            {{ $t('cluster.create.label.existingClusters') }}
          </bk-radio>
        </bk-radio-group>
      </bk-form-item>
      <bk-form-item
        :label="$t('generic.label.cluster')"
        :rules="rules.hostClusterId"
        property="hostClusterId"
        error-display-type="normal"
        required>
        <div class="flex items-center">
          <bcs-select
            :loading="isLoading"
            class="w-[500px]"
            v-model="hostClusterId"
            searchable
            :clearable="false">
            <bcs-option
              v-for="item in clusterData"
              :key="item.clusterID"
              :id="item.clusterID"
              :name="item.clusterName"
              :disabled="item.disabled"
              v-bk-tooltips="{
                content: item.disabled ?
                  $t('cluster.create.federation.tips.disabledSub', [unavailableClustersMap[item.clusterID]]) : '',
                delay: [300, 0],
                interactive: false,
              }"
              v-authority="{
                clickable: item.clickable,
                actionId: 'cluster_manage',
                resourceName: item.clusterName,
                disablePerms: true,
                permCtx: {
                  project_id: projectID,
                  cluster_id: item.clusterID
                }
              }" />
          </bcs-select>
          <bk-button icon="icon-refresh" @click="handleGetClusterList" class="ml10" :disabled="isLoading"></bk-button>
        </div>
      </bk-form-item>
    </bk-form>
    <div class="mt25">
      <!-- 从节点池添加节点 -->
      <bk-button
        class="mw88"
        theme="primary"
        :loading="saving"
        @click="handleConfirm">
        {{$t('generic.button.confirm')}}
      </bk-button>
      <bk-button class="mw88 ml10" @click="handleCancel">{{$t('generic.button.cancel')}}</bk-button>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from 'vue';

import { isAdding } from './use-federation';

import { addFederalCluster, getFederalCluster } from '@/api/modules/cluster-manager';
import $bkMessage from '@/common/bkmagic';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import { ICluster, useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import $store from '@/store/index';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';

export default defineComponent({
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const curCluster = computed<ICluster>(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === props.clusterId) || {});
    const provider = computed(() => curCluster.value.provider);

    const nodeSource = ref<'create'|'existing'>('existing');
    const saving = ref(false);
    const formRef = ref<any>(null);

    const rules = {
      hostClusterId: [
        {
          message: $i18n.t('generic.validate.required'),
          trigger: 'blur',
          validator: () => !!hostClusterId.value,
        },
      ],
    };

    const user = computed(() => $store.state.user);
    const subCluster = computed(() => clusterData.value?.find(item => item.clusterID === hostClusterId.value));
    const handleConfirm = async () => {
      const result = await formRef.value?.validate().catch(() => false);
      if (!result) return;
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('cluster.create.federation.addClusterTitle'),
        subTitle: `${subCluster.value?.clusterName}(${subCluster.value?.clusterID})`,
        defaultInfo: true,
        confirmLoading: true,
        confirmFn: async () => {
          const data = await addFederalCluster({
            $fedClusterId: props.clusterId,
            $subClusterId: hostClusterId.value,
            creator: user.value.username,
          }).catch(() => false);
          if (data) {
            $bkMessage({
              theme: 'success',
              message: $i18n.t('generic.msg.success.deliveryTask'),
            });
            $router.push({
              name: 'clusterMain',
              query: {
                active: 'subCluster',
                clusterId: props.clusterId,
              },
            });
            // // 异步添加中，接口可能未返回数据，需将轮询标识设置为true
            isAdding.value = true;
          }
        },
      });
    };

    const handleCancel = () => {
      $router.back();
    };

    // 集群列表
    const {
      clusterList,
      federationClusters,
      getClusterList,
    } = useClusterList();
    const isLoading = ref(false);
    const { projectID } = useProject();
    const hostClusterId = ref<string>('');
    const clusterWebAnnotations = computed(() => $store.state.cluster.clusterWebAnnotations.perms);
    // 展示除（共享集群、sub-cluster、fed-cluster）的集群
    const clusterData = computed(() => clusterList.value
      ?.filter(cur => !cur.is_shared && !cur.labels?.['federation.bkbcs.tencent.com/is-fed-cluster'])
      ?.map(item => ({
        clusterID: item.clusterID,
        clusterName: item.clusterName,
        clickable: isClickable(item),
        disabled: hasDisallowKey(item),
      })));
    const unavailableClusters = computed(() => clusterData.value.filter(item => item.disabled));
    async function handleGetClusterList() {
      isLoading.value = true;
      await getClusterList();
      isLoading.value = false;
    };

    // 没有集群管理权限&集群不是host集群（被占用）：禁用
    function isClickable(cluster: ICluster) {
      return clusterWebAnnotations.value[cluster.clusterID]?.cluster_manage;
    };

    // 是否准许加入联邦
    const disallowKeyList = [
      // 'federation.bkbcs.tencent.com/is-fed-cluster',
      'federation.bkbcs.tencent.com/is-sub-cluster',
    ];
    function hasDisallowKey(cluster: ICluster) {
      return disallowKeyList.some(item => cluster?.labels?.[item]);
    }

    // 获取成员集群对应的联邦集群ID
    const unavailableClustersMap = ref({});
    watch(federationClusters, () => {
      if (unavailableClusters.value.length === 0) return;
      federationClusters.value.forEach(async (cluster) => {
        const result = await getFederalCluster({
          $federationClusterId: cluster.clusterID,
        }).catch(() => ({}));
        if (result.sub_clusters) {
          result.sub_clusters.forEach((item) => {
            // 这样赋值页面才会更新
            unavailableClustersMap.value = {
              ...unavailableClustersMap.value,
              [item.sub_cluster_id]: cluster.clusterID,
            };
          });
        }
      });
    });

    onMounted(() => {
      handleGetClusterList();
    });

    return {
      provider,
      curCluster,
      nodeSource,
      saving,
      isLoading,
      clusterData,
      projectID,
      hostClusterId,
      rules,
      formRef,
      unavailableClustersMap,
      handleConfirm,
      handleCancel,
      handleGetClusterList,
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
