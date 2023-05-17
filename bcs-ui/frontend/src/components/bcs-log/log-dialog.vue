<template>
  <bcs-dialog
    class="log-dialog"
    :value="value"
    width="80%"
    :show-footer="false"
    render-directive="if"
    @value-change="handleValueChange">
    <Log
      :project-id="projectID"
      :cluster-id="clusterId"
      :namespace-id="namespace"
      :pod-id="name"
      :default-container="defaultContainer"
      :global-loading="logLoading"
      :container-list="containerList" />
  </bcs-dialog>
</template>
<script lang="ts">
import { defineComponent, reactive, toRefs, watch } from 'vue';
import { useProject } from '@/composables/use-app';
import Log from './index';
import $store from '@/store';

export default defineComponent({
  name: 'LogDialog',
  components: { Log },
  model: {
    prop: 'value',
    event: 'change',
  },
  props: {
    value: {
      type: Boolean,
      default: false,
    },
    name: {
      type: String,
      default: '',
    },
    namespace: {
      type: String,
      default: '',
    },
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props, ctx) {
    const { value, name, namespace, clusterId } = toRefs(props);
    const { projectID } = useProject();
    // 获取日志容器组件容器列表数据
    const logState = reactive<{
      logLoading: boolean;
      defaultContainer: string;
      containerList: any[];
    }>({
      logLoading: false,
      defaultContainer: '',
      containerList: [],
    });
    watch(value, (show) => {
      if (!show) {
        logState.containerList = [];
        logState.defaultContainer = '';
      } else {
        handleShowLog();
      }
    });
    const handleGetContainer = async (podId: string, namespace: string, clusterId: string) => {
      const data = await $store.dispatch('log/podContainersList', {
        $podId: podId,
        $namespaceId: namespace,
        $clusterId: clusterId,
      });
      return data;
    };
    // 显示操作日志
    const handleShowLog = async () => {
      logState.containerList = await handleGetContainer(name.value, namespace.value, clusterId.value);
      logState.defaultContainer = logState.containerList[0]?.name;
    };
    const handleValueChange = (value) => {
      ctx.emit('change', value);
    };
    return {
      projectID,
      ...toRefs(logState),
      handleValueChange,
    };
  },
});
</script>
<style lang="postcss" scoped>
.log-dialog {
  /deep/ .bk-dialog-tool {
      display: none;
  }
  /deep/ .bk-dialog-close {
      display: none;
  }
  /deep/ .bk-dialog {
      top: 0;
      position: relative;
      height: calc(100% - 32px);
      float: right;
      margin: 16px;
      border-radius: 6px;
      transition-property: transform, opacity;
      transition: transform 200ms cubic-bezier(0.165, 0.84, 0.44, 1),opacity 100ms cubic-bezier(0.215, 0.61, 0.355, 1);
  }
  /deep/ .bk-dialog-body {
      padding: 0;
      height: 100%;
  }
  /deep/ .bk-dialog-wrapper .bk-dialog-content {
      border-radius: 6px;
      height: 100%;
      width: 100% !important;
  }
}

</style>
