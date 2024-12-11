<template>
  <div class="flex flex-col flex-1 h-full">
    <DetailNav
      :list="navList"
      :cluster-id="clusterId"
      :active="componentId"
      class="flex-[0_0_auto]"
      @change="handleNavChange">
    </DetailNav>
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
import { isEqual } from 'lodash';
import { computed, defineComponent, onBeforeMount, ref } from 'vue';

import ContainerDetail from './container-detail.vue';
import DetailNav from './detail-nav.vue';
import PodDetail from './pod-detail.vue';
import WorkloadDetail from './workload-detail.vue';

import $router from '@/router';

export type ComponentIdType = 'WorkloadDetail' | 'PodDetail' | 'ContainerDetail';
export interface INavItem {
  name: string; // 展示名称
  kind?: string; // 类型
  id: ComponentIdType|'';// 组件ID
  params?: Record<string, any>; // 组件参数
  query?: Record<string, any>// 查询条件
}

export default defineComponent({
  name: 'DashboardDetail',
  components: {
    DetailNav,
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
    // 名称（注意，当 category 为 Pods时，这里的name就是Pod名称。下面的Pod名称是后来为了兼容进入workload 下Pod时没有记住当前Pod问题）
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
    clusterId: {
      type: String,
      default: '',
    },
    // pod名称
    pod: {
      type: String,
      default: '',
    },
    // 容器名称
    container: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    // 区分首次进入pod详情还是其他workload详情
    const defaultComId = props.category === 'pods' ? 'PodDetail' : 'WorkloadDetail';
    // 子标题
    const subTitleMap = {
      Deployment: 'Deploy',
      StatefulSet: 'DS',
      DaemonSet: 'STS',
      CronJob: 'CJ',
      Job: 'Job',
      Pod: 'Pod',
      GameDeployment: 'GameDeployment',
      GameStatefulSet: 'GameStatefulSet',
      Container: 'Container',
    };
    // 顶部导航内容
    const navList = computed<INavItem[]>(() => {
      const data: INavItem[] = [
        {
          name: `${props.kind}s`,
          id: '',
        },
      ];
      if (props.category === 'pods') {
        // Pods详情
        data.push({
          name: props.name,
          kind: subTitleMap[props.kind] || props.kind,
          id: 'PodDetail',
          params: {
            ...props,
          },
        });
      } else {
        // workload详情
        data.push(...[
          {
            name: props.name,
            kind: subTitleMap[props.kind] || props.kind,
            id: 'WorkloadDetail',
            params: {
              ...props,
            },
          },
          {
            name: props.pod,
            kind: subTitleMap.Pod,
            id: 'PodDetail',
            params: {
              name: props.pod,
              namespace: props.namespace,
              clusterId: props.clusterId,
              hiddenOperate: props.hiddenOperate,
            },
            query: {
              kind: props.kind,
              crd: props.crd,
              pod: props.pod,
            },
          },
        ] as INavItem[]);
      }
      // 容器详情
      data.push({
        name: props.container,
        kind: subTitleMap.Container,
        id: 'ContainerDetail',
        params: {
          namespace: props.namespace,
          pod: props.pod,
          name: props.container,
          clusterId: props.clusterId,
        },
        query: {
          kind: props.kind,
          crd: props.crd,
          pod: props.pod,
          container: props.container,
        },
      });
      return data;
    });
    const componentId = ref<ComponentIdType|''>('');
    // 详情组件所需的参数
    const componentProps = computed(() => navList.value.find(item => item.id === componentId.value)?.params || {});

    const handleNavChange = (item: INavItem) => {
      const { id, query } = item;
      if (id === '') {
        $router.back();
      } else {
        componentId.value = id;
        // 更新路由参数
        const newQuery = query ?? {
          crd: props.crd,
          kind: props.kind,
        };
        if (!isEqual($router.currentRoute.query, newQuery)) {
          $router.replace({
            query: newQuery,
          });
        }
      }
    };
    // 跳转pod详情
    const handleGotoPodDetail = async (row) => {
      await $router.replace({
        query: {
          kind: props.kind,
          crd: props.crd,
          pod: row.metadata.name,
        },
      });
      componentId.value = 'PodDetail';
    };
    // 跳转容器详情
    const handleGotoContainerDetail = async (row) => {
      await $router.replace({
        query: {
          kind: props.kind,
          crd: props.crd,
          pod: props.category === 'pods' ? props.name : props.pod,
          container: row.name,
        },
      });
      componentId.value = 'ContainerDetail';
    };

    onBeforeMount(() => {
      // 设置当前详情组件
      const { name, pod, container } = props;
      if (container && pod && name) {
        componentId.value = 'ContainerDetail';
      } else if (pod && name) {
        componentId.value = 'PodDetail';
      } else {
        componentId.value = defaultComId;
      }
    });

    return {
      componentId,
      componentProps,
      navList,
      handleNavChange,
      handleGotoPodDetail,
      handleGotoContainerDetail,
    };
  },
});
</script>
