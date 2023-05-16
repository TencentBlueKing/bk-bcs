<template>
  <div class="taint-wrapper" v-bkloading="{ isLoading, opacity: 1 }">
    <template v-if="values.length">
      <div class="labels">
        <span>{{$t('键')}}：</span>
        <span>{{$t('值')}}：</span>
        <span>{{$t('影响')}}：</span>
      </div>
      <BcsTaints
        class="taints"
        :effect-options="effectList"
        :min-items="0"
        ref="taintRef"
        v-model="values">
      </BcsTaints>
    </template>
    <span
      class="add-btn mb15"
      v-else
      @click="handleAddTaint">
      <i class="bk-icon icon-plus-circle-shape mr5"></i>
      {{$t('添加')}}
    </span>
    <div>
      <bk-button
        theme="primary"
        class="min-w-[88px]"
        :loading="isSubmitting"
        @click="handleSubmit">{{$t('确定')}}</bk-button>
      <bk-button
        theme="default"
        class="min-w-[88px]"
        @click="handleCancel">{{$t('取消')}}</bk-button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref, toRefs, PropType, watch } from 'vue';
import useNode from '../node-list/use-node';
import BcsTaints from './new-taints.vue';

interface IValueItem {
  key: string;
  value: string;
  effect: string;
}

export default defineComponent({
  name: 'NodeTaint',
  components: { BcsTaints },
  props: {
    clusterId: {
      type: String,
      required: true,
    },
    nodes: {
      type: Array as PropType<any[]>,
      default: () => [],
    },
    // todo内置污点不可删除
    disabledDeleteKeys: {
      type: Array,
      default: () => [
        'node.kubernetes.io/not-ready',
        'node.kubernetes.io/unreachable',
        'node.kubernetes.io/memory-pressure',
        'node.kubernetes.io/disk-pressure',
        'node.kubernetes.io/pid-pressure',
        'node.kubernetes.io/network-unavailable',
        'node.kubernetes.io/unschedulable',
        'node.cloudprovider.kubernetes.io/uninitialized',
      ],
    },
  },
  setup(props, ctx) {
    const { nodes, clusterId } = toRefs(props);
    const isLoading = ref<boolean>(false);
    const isSubmitting = ref<boolean>(false);
    const effectList = ref(['PreferNoSchedule', 'NoExecute', 'NoSchedule']);
    const values = ref<IValueItem[]>([]);
    const taintRef = ref<any>(null);

    watch(values, () => {
      ctx.emit('data-change', values);
    }, { deep: true });
    const { setNodeTaints } = useNode();
    // 提交数据
    const handleSubmit = async () => {
      const validate = taintRef.value?.validate();
      if (!validate && values.value.length) return;

      isSubmitting.value = true;
      // data是单个节点设置污点的结果，多个节点需要另外处理
      const data: IValueItem[] = [];
      for (const item of values.value) {
        // 只提交填了key的行
        item.key && data.push(item);
      }
      const result =  await setNodeTaints({
        clusterID: clusterId.value,
        nodes: nodes.value.map(node => ({
          nodeName: node.nodeName,
          taints: data,
        })),
      });
      isSubmitting.value = false;
      result && ctx.emit('confirm');
    };
    // 关闭弹窗
    const handleCancel = (refetch = false) => {
      ctx.emit('cancel', refetch);
    };
    const handleAddTaint = () => {
      values.value.push({ key: '', value: '', effect: 'PreferNoSchedule' });
    };
    onMounted(() => {
      values.value = nodes.value.reduce((pre, current) => {
        if (current.taints) {
          pre.push(...current.taints);
        }
        return pre;
      }, []);
    });
    return {
      taintRef,
      isLoading,
      isSubmitting,
      values,
      effectList,
      handleSubmit,
      handleCancel,
      handleAddTaint,
    };
  },
});
</script>

<style lang="postcss" scoped>
@define-mixin flex-layout {
  display: flex;
  align-items: center;
}
.add-btn {
  cursor: pointer;
  background: #fff;
  border: 1px dashed #c4c6cc;
  border-radius: 2px;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  font-size: 14px;
  &:hover {
      border-color: #3a84ff;
      color: #3a84ff;
  }
}
.taint-wrapper {
  padding: 20px;
  .labels {
      @mixin flex-layout;
      font-size: 14px;
      margin-bottom: 20px;
      > span {
          flex: 0 0 calc(100% / 3 - 20px);
      }
  }
  >>> .taints {
      .key {
          width: 190px;
      }
      .value {
          width: 200px;
      }
  }
}
</style>
