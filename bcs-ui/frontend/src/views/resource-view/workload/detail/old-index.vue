<template>
  <div class="flex flex-col flex-1 h-full">
    <DetailTopNav
      :list="titles"
      :cluster-id="clusterId"
      :active="componentId"
      class="flex-[0_0_auto]"
      @change="handleNavChange">
    </DetailTopNav>
    <component
      :is="componentId"
      v-bind="componentProps"
      class="flex-1 overflow-auto pb-[16px]"
      @pod-detail="handleGotoPodDetail"
      @container-detail="handleGotoContainerDetail">
    </component>
  </div>
</template>
<script lang="ts">
// 旧版资源视图详情（模板集再使用，后续会删掉）
import { computed, defineComponent, ref } from 'vue';

import ContainerDetail from './container-detail.vue';
import DetailTopNav from './detail-nav.vue';
import PodDetail from './pod-detail.vue';
import WorkloadDetail from './workload-detail.vue';

import $router from '@/router';

export type ComponentIdType = 'WorkloadDetail' | 'PodDetail' | 'ContainerDetail';
export interface ITitle {
  name: string; // 展示名称
  id: string;// 组件ID
  params?: any; // 组件参数
}

export default defineComponent({
  name: 'DashboardDetail',
  components: {
    DetailTopNav,
    WorkloadDetail,
    PodDetail,
    ContainerDetail,
  },
  props: {
    // 命名空间
    namespace: {
      type: String,
      default: '',
    },
    // workload类型
    category: {
      type: String,
      default: '',
    },
    // 名称
    name: {
      type: String,
      default: '',
    },
    // crd名称
    crd: {
      type: String,
      default: '',
    },
    // kind类型
    kind: {
      type: String,
      default: '',
      required: true,
    },
    // 是否隐藏 更新 和 删除操作（兼容集群管理应用详情）
    hiddenOperate: {
      type: Boolean,
      default: false,
    },
    // 区分是否从 集群管理node-pods 跳转过来的
    from: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
    },
    nodeId: {
      type: String,
      default: '',
    },
    nodeName: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    // 区分首次进入pod详情还是其他workload详情
    const defaultComId = props.category === 'pods' ? 'PodDetail' : 'WorkloadDetail';
    // 子标题
    const subTitleMap = {
      deployments: 'Deploy',
      daemonsets: 'DS',
      statefulsets: 'STS',
      cronjobs: 'CJ',
      jobs: 'Job',
      pods: 'Pod',
      container: 'Container',
    };
    const cobjKindMap = {
      GameDeployment: 'GameDeployments',
      GameStatefulSet: 'GameStatefulSets',
    };
    const cobjSubTitleMap = {
      GameDeployment: 'GameDeployment',
      GameStatefulSet: 'GameStatefulSet',
    };
    // 首字母大写
    const upperFirstLetter = (str: string) => {
      if (!str) return str;

      return `${str.slice(0, 1).toUpperCase()}${str.slice(1)}`;
    };
    // 顶部导航内容
    const titles = ref<ITitle[]>([
      {
        name: upperFirstLetter(cobjKindMap[props.kind] || props.category),
        id: '',
      },
      {
        name: `${cobjSubTitleMap[props.kind] || subTitleMap[props.category]}: ${props.name}`,
        id: defaultComId,
        params: {
          ...props,
        },
      },
    ]);
    const componentId = ref(defaultComId);
    // 详情组件所需的参数
    const componentProps = computed(() => titles.value.find(item => item.id === componentId.value)?.params || {});

    const handleNavChange = (item: ITitle) => {
      const { id } = item;
      const index = titles.value.findIndex(item => item.id === id);
      if (id === '') {
        $router.back();
      } else {
        componentId.value = id;
        if (index > -1) {
          // 截取后面的导航
          titles.value = titles.value.slice(0, index + 1);
        } else {
          titles.value.push(item);
        }
      }
    };
    // 跳转pod详情
    const handleGotoPodDetail = (row) => {
      handleNavChange({
        name: `${subTitleMap.pods}: ${row.metadata.name}`,
        id: 'PodDetail',
        params: {
          name: row.metadata.name,
          namespace: row.metadata.namespace,
          clusterId: props.clusterId,
          hiddenOperate: props.hiddenOperate,
        },
      });
    };
    // 调转容器详情
    const handleGotoContainerDetail = (row) => {
      // 容器的父级Pod
      const { name } = titles.value.find(item => item.id === componentId.value)?.params || {};
      handleNavChange({
        name: `${subTitleMap.container}: ${row.name}`,
        id: 'ContainerDetail',
        params: {
          namespace: props.namespace,
          pod: name,
          name: row.name,
          id: row.containerID,
          clusterId: props.clusterId,
        },
      });
    };

    return {
      componentId,
      componentProps,
      titles,
      handleNavChange,
      handleGotoPodDetail,
      handleGotoContainerDetail,
    };
  },
});
</script>
