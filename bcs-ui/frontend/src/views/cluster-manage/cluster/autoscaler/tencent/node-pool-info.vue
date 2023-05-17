<!-- eslint-disable max-len -->
<template>
  <bk-form class="node-pool-info" :model="nodePoolInfo" :rules="nodePoolInfoRules" ref="formRef">
    <bk-form-item :label="$t('节点池名称')" property="name" error-display-type="normal" required>
      <bk-input v-model="nodePoolInfo.name"></bk-input>
    </bk-form-item>
    <bk-form-item
      :label="$t('节点数量范围')"
      property="nodeNumRange"
      error-display-type="normal"
      :desc="$t('在设定的节点范围内自动调节，不会超出该设定范围')">
      <bk-input
        class="w74"
        v-model="nodePoolInfo.autoScaling.minSize"
        :min="getSchemaByProp('autoScaling.minSize').minimum"
        :max="getSchemaByProp('autoScaling.minSize').maximum"
        type="number">
      </bk-input>
      <span>~</span>
      <bk-input
        class="w74"
        v-model="nodePoolInfo.autoScaling.maxSize"
        :min="getSchemaByProp('autoScaling.maxSize').minimum"
        :max="getSchemaByProp('autoScaling.maxSize').maximum"
        type="number">
      </bk-input>
    </bk-form-item>
    <bk-form-item
      :label="$t('是否启用节点池')"
      :desc="$t('节点池启用后Autoscaler组件将会根据扩容算法使用该节点池资源，开启Autoscaler组件后必须要开启至少一个节点池')">
      <bk-checkbox v-model="nodePoolInfo.enableAutoscale" :disabled="isEdit"></bk-checkbox>
    </bk-form-item>
    <!-- <bk-form-item :label="$t('是否开启调度')">
            <bk-checkbox :true-value="0" :false-value="1"
                v-model="nodePoolInfo.nodeTemplate.unSchedulable"></bk-checkbox>
        </bk-form-item> -->
    <bk-form-item label="Labels" property="labels" error-display-type="normal">
      <KeyValue
        class="labels"
        :min-item="0"
        :disable-delete-item="false"
        v-model="nodePoolInfo.nodeTemplate.labels">
      </KeyValue>
    </bk-form-item>
    <bk-form-item label="Taints" property="taints" error-display-type="normal">
      <span
        class="add-key-value-items" v-if="!nodePoolInfo.nodeTemplate.taints.length"
        @click="handleAddTaints">
        <i class="bk-icon icon-plus-circle-shape mr5"></i>
        {{$t('添加')}}
      </span>
      <Taints
        class="taints"
        :effect-options="getSchemaByProp('nodeTemplate.taints.effect').enum"
        :key-rules="[]"
        v-model="nodePoolInfo.nodeTemplate.taints"
        v-else>
      </Taints>
    </bk-form-item>
    <bk-form-item
      :label="$t('实例创建策略')"
      :desc="$t('首选可用区（子网）优先：自动扩缩容会在您首选的可用区优先执行扩缩容，若首选可用区无法扩缩容，才会在其他可用区进行扩缩容<br/>多可用区（子网）打散 ：在节点池指定的多可用区（即指定多个子网）之间尽最大努力均匀分配CVM实例，只有配置了多个子网时该策略才能生效')">
      <bk-radio-group v-model="nodePoolInfo.autoScaling.multiZoneSubnetPolicy">
        <bk-radio value="PRIORITY">{{$t('首选可用区（子网）优先')}}</bk-radio>
        <bk-radio value="EQUALITY">{{$t('多可用区（子网）打散')}}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      :label="$t('重试策略')"
      :desc="$t('快速重试 ：立即重试，在较短时间内快速重试，连续失败超过一定次数（5次）后不再重试，<br/>间隔递增重试 ：间隔递增重试，随着连续失败次数的增加，重试间隔逐渐增大，重试间隔从秒级到1天不等，<br/>不重试：不进行重试，直到再次收到用户调用或者告警信息后才会重试')">
      <bk-radio-group v-model="nodePoolInfo.autoScaling.retryPolicy">
        <bk-radio value="IMMEDIATE_RETRY">{{$t('快速重试')}}</bk-radio>
        <bk-radio value="INCREMENTAL_INTERVALS">{{$t('间隔递增重试')}}</bk-radio>
        <bk-radio value="NO_RETRY">{{$t('不重试')}}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      :label="$t('扩缩容模式')"
      :desc="$t('释放模式：缩容时自动释放Cluster AutoScaler判断的空余节点， 扩容时自动创建新的CVM节点加入到伸缩组<br/>关机模式：扩容时优先对已关机的节点执行开机操作，节点数依旧不满足要求时再创建新的CVM节点')">
      <bk-radio-group v-model="nodePoolInfo.autoScaling.scalingMode">
        <bk-radio value="CLASSIC_SCALING">{{$t('释放模式')}}</bk-radio>
        <bk-radio value="WAKE_UP_STOPPED_SCALING">{{$t('关机模式')}}</bk-radio>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      :label="$t('云区域')"
      :desc="$t('Cluster Autoscaler组件在扩容节点时会使用节点管理API自动安装gse_agent，以保证监控、日志、标准运维的正常使用，这里的云区域是指节点管理安装gse_agent时指定的云区域，默认为“直连区域”')">
      <bcs-select
        :clearable="false"
        :loading="cloudLoading"
        v-model="nodePoolInfo.bkCloudID">
        <bcs-option
          v-for="item in cloudList"
          :key="item.bk_cloud_id"
          :id="item.bk_cloud_id"
          :name="item.bk_cloud_name">
        </bcs-option>
      </bcs-select>
    </bk-form-item>
    <bk-form-item :label="$t('模块')">
      <bcs-select
        :show-empty="false"
        allow-create
        v-model="nodePoolInfo.nodeTemplate.module.scaleOutModuleName"
        @clear="handleClearNode">
        <bcs-big-tree
          :data="topoData"
          :options="{
            idKey: 'bk_inst_id',
            childrenKey: 'child',
            nameKey: 'bk_inst_name'
          }"
          ref="treeRef"
          default-expand-all
          :default-checked-nodes="[nodePoolInfo.nodeTemplate.module.scaleOutModuleID]">
          <template #default="{ data, node }">
            <div @click="handleSelectNode(data, node)">
              <bcs-radio
                v-if="data.bk_obj_id === 'module'"
                :checked="nodePoolInfo.nodeTemplate.module.scaleOutModuleID === String(data.bk_inst_id)">
                {{ data.bk_inst_name }}
              </bcs-radio>
              <span v-else>{{ data.bk_inst_name }}</span>
            </div>
          </template>
        </bcs-big-tree>
      </bcs-select>
    </bk-form-item>
    <bk-form-item class="mt40" v-if="!isEdit">
      <bk-button theme="primary" @click="handleNext">{{ $t('下一步') }}</bk-button>
      <bk-button @click="handleCancel">{{ $t('取消') }}</bk-button>
    </bk-form-item>
  </bk-form>
</template>
<script lang="ts">
import { defineComponent, ref, toRefs, onMounted } from 'vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import Taints from '@/views/cluster-manage/components/new-taints.vue';
import $router from '@/router';
import $i18n from '@/i18n/i18n-setup';
import Schema from '@/views/cluster-manage/cluster/autoscaler/resolve-schema';
import { nodemanCloudList, ccTopology } from '@/api/base';

export default defineComponent({
  components: { KeyValue, Taints },
  props: {
    schema: {
      type: Object,
      default: () => ({}),
    },
    // 详情数据或者默认值
    defaultValues: {
      type: Object,
      default: () => ({}),
    },
    isEdit: {
      type: Boolean,
      default: false,
    },
    cluster: {
      type: Object,
      default: () => ({}),
    },
  },
  setup(props, ctx) {
    const { defaultValues, schema, cluster } = toRefs(props);
    const formRef = ref<any>(null);
    const nodePoolInfo = ref({
      name: defaultValues.value.name || '', // 节点名称
      autoScaling: {
        maxSize: defaultValues.value.autoScaling.maxSize, // 节点数量范围
        minSize: defaultValues.value.autoScaling.minSize,
        scalingMode: defaultValues.value.autoScaling.scalingMode, // 扩缩容模式
        multiZoneSubnetPolicy: defaultValues.value.autoScaling.multiZoneSubnetPolicy, // 实列创建策略
        retryPolicy: defaultValues.value.autoScaling.retryPolicy, // 重试策略
      },
      enableAutoscale: defaultValues.value.enableAutoscale, // 是否开启弹性伸缩
      nodeTemplate: {
        unSchedulable: defaultValues.value.nodeTemplate.unSchedulable || 0, // 是否开启调度 0 代表开启调度，1 不可调度
        labels: defaultValues.value.nodeTemplate.labels || {}, // 标签
        taints: defaultValues.value.nodeTemplate.taints || [], // 污点
        module: {
          scaleOutModuleID: defaultValues.value.nodeTemplate?.module?.scaleOutModuleID || '',
          scaleOutModuleName: defaultValues.value.nodeTemplate?.module?.scaleOutModuleName || '',
        },
      },
      bkCloudID: defaultValues.value.bkCloudID || 0,
    });

    const nodePoolInfoRules = ref({
      name: [
        {
          required: true,
          message: $i18n.t('必填项'),
          trigger: 'blur',
        },
      ],
      nodeNumRange: [
        {
          message: $i18n.t('请输入正确节点数量范围'),
          trigger: 'blur',
          validator: () => nodePoolInfo.value.autoScaling.minSize < nodePoolInfo.value.autoScaling.maxSize,
        },
      ],
      labels: [
        {
          message: $i18n.t('标签值不能为空'),
          trigger: 'custom',
          validator: () => Object.keys(nodePoolInfo.value.nodeTemplate.labels)
            .every(key => !!nodePoolInfo.value.nodeTemplate.labels[key]),
        },
      ],
      taints: [
        {
          message: $i18n.t('taints不能有空字段'),
          trigger: 'custom',
          validator: () => nodePoolInfo.value.nodeTemplate.taints.every(item => item.key && item.value && item.effect),
        },
        {
          message: $i18n.t('重复键'),
          trigger: 'custom',
          validator: () => {
            const data = nodePoolInfo.value.nodeTemplate.taints.reduce((pre, item) => {
              if (item.key) {
                pre.push(item.key);
              }
              return pre;
            }, []);
            const removeDuplicateData = new Set(data);
            return data.length === removeDuplicateData.size;
          },
        },
      ],
    });

    const handleAddTaints = () => {
      nodePoolInfo.value.nodeTemplate.taints.push({
        key: '',
        value: '',
        effect: 'PreferNoSchedule',
      });
    };

    const handleNext = async () => {
      const result = await formRef.value?.validate();
      if (!result) return;

      ctx.emit('next', nodePoolInfo.value);
    };
    const handleCancel = () => {
      $router.back();
    };

    const getSchemaByProp = props => Schema.getSchemaByProp(schema.value, props);
    // 云区域列表
    const cloudList = ref<any[]>([]);
    const cloudLoading = ref(false);
    const handleGetCloudList = async () => {
      cloudLoading.value = true;
      cloudList.value = await nodemanCloudList().catch(() => []);
      cloudLoading.value = false;
    };
    // 模块
    const treeRef = ref<any>(null);
    const topoLoading = ref(false);
    const topoData = ref<any[]>([]);
    const handleGetTopoData = async () => {
      topoLoading.value = true;
      const data = await ccTopology({
        $clusterId: cluster.value?.clusterID,
      });
      topoData.value = [data];
      topoLoading.value = false;
    };
    const getNodePath = (node) => {
      if (node.parent) {
        // eslint-disable-next-line camelcase
        return `${getNodePath(node.parent)} / ${node?.data?.bk_inst_name}`;
      }
      // eslint-disable-next-line camelcase
      return node?.data?.bk_inst_name;
    };
    const handleSelectNode = (data, node) => {
      const path = getNodePath(node);
      if (data.bk_obj_id === 'module') {
        nodePoolInfo.value.nodeTemplate.module.scaleOutModuleID = String(data.bk_inst_id);
        nodePoolInfo.value.nodeTemplate.module.scaleOutModuleName = path;
      }
    };
    const handleClearNode = () => {
      nodePoolInfo.value.nodeTemplate.module.scaleOutModuleID = '';
      nodePoolInfo.value.nodeTemplate.module.scaleOutModuleName = '';
    };

    onMounted(() => {
      handleGetCloudList();
      handleGetTopoData();
    });
    return {
      cloudList,
      cloudLoading,
      formRef,
      nodePoolInfo,
      nodePoolInfoRules,
      treeRef,
      topoLoading,
      topoData,
      handleNext,
      handleCancel,
      getSchemaByProp,
      handleAddTaints,
      handleSelectNode,
      handleClearNode,
    };
  },
});
</script>
<style lang="postcss" scoped>
.node-pool-info {
    >>> .w74 {
        width: 74px;
    }
    >>> .w160 {
        width: 160px;
    }
    >>> .labels,
    >>> .taints {
        .bk-form-control {
            width: 160px;
        }
    }
    >>> .bk-form-content {
        max-width: 600px;
    }
    >>> .add-key-value-items {
        font-size: 14px;
        color: #3a84ff;
        cursor: pointer;
        display: flex;
        align-items: center;
    }
}
</style>
