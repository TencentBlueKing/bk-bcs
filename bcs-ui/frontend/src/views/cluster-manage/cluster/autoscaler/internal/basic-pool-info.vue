<!-- eslint-disable max-len -->
<template>
  <bk-form class="node-pool-info" :model="nodePoolInfo" :rules="nodePoolInfoRules" ref="formRef">
    <!-- <bk-form-item :label="$t('节点规格名称')" property="name" error-display-type="normal" required>
      <bk-input
        v-model="nodePoolInfo.name"
        :placeholder="$t('名称不超过255个字符，仅支持中文、英文、数字、下划线，分隔符(-)及小数点')">
      </bk-input>
    </bk-form-item> -->
    <bk-form-item
      :label="$t('节点配额')"
      property="maxSize"
      error-display-type="normal">
      <bk-input
        class="w74"
        int
        v-model="nodePoolInfo.autoScaling.maxSize"
        :min="getSchemaByProp('autoScaling.maxSize').minimum"
        :max="getSchemaByProp('autoScaling.maxSize').maximum"
        type="number">
      </bk-input>
    </bk-form-item>
    <bk-form-item
      :label="$t('是否启用节点规格')"
      :desc="$t('节点规格启用后Autoscaler组件将会根据扩容算法使用该节点规格资源，开启Autoscaler组件后必须要开启至少一个节点规格')">
      <bk-checkbox v-model="nodePoolInfo.enableAutoscale" :disabled="isEdit"></bk-checkbox>
    </bk-form-item>
    <bk-form-item :label="$t('标签')" property="labels" error-display-type="normal">
      <KeyValue
        class="max-w-[420px]"
        :min-item="0"
        :disable-delete-item="false"
        v-model="nodePoolInfo.nodeTemplate.labels">
      </KeyValue>
    </bk-form-item>
    <bk-form-item :label="$t('污点')" property="taints" error-display-type="normal">
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
    <bk-form-item :label="$t('注解')" property="annotations" error-display-type="normal">
      <KeyValue
        class="max-w-[420px]"
        :min-item="0"
        :disable-delete-item="false"
        v-model="nodePoolInfo.nodeTemplate.annotations">
      </KeyValue>
    </bk-form-item>
    <bk-form-item
      :label="$t('实例创建策略')"
      :desc="$t('首选可用区（子网）优先：自动扩缩容会在您首选的可用区优先执行扩缩容，若首选可用区无法扩缩容，才会在其他可用区进行扩缩容<br/>多可用区（子网）打散 ：在节点规格指定的多可用区（即指定多个子网）之间尽最大努力均匀分配CVM实例，只有配置了多个子网时该策略才能生效')">
      <bk-radio-group v-model="nodePoolInfo.autoScaling.multiZoneSubnetPolicy">
        <span class="inline-block" v-bk-tooltips="$t('自研上云环境暂不支持修改')">
          <bk-radio value="PRIORITY" disabled>{{$t('首选可用区（子网）优先')}}</bk-radio>
          <bk-radio value="EQUALITY" disabled>{{$t('多可用区（子网）打散')}}</bk-radio>
        </span>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      :label="$t('重试策略')"
      :desc="$t('快速重试 ：立即重试，在较短时间内快速重试，连续失败超过一定次数（5次）后不再重试，<br/>间隔递增重试 ：间隔递增重试，随着连续失败次数的增加，重试间隔逐渐增大，重试间隔从秒级到1天不等，<br/>不重试：不进行重试，直到再次收到用户调用或者告警信息后才会重试')">
      <bk-radio-group v-model="nodePoolInfo.autoScaling.retryPolicy">
        <span class="inline-block" v-bk-tooltips="$t('自研上云环境暂不支持修改')">
          <bk-radio value="IMMEDIATE_RETRY" disabled>{{$t('快速重试')}}</bk-radio>
          <bk-radio value="INCREMENTAL_INTERVALS" disabled>{{$t('间隔递增重试')}}</bk-radio>
          <bk-radio value="NO_RETRY" disabled>{{$t('不重试')}}</bk-radio>
        </span>
      </bk-radio-group>
    </bk-form-item>
    <bk-form-item
      :label="$t('扩缩容模式')"
      :desc="$t('释放模式：缩容时自动释放Cluster AutoScaler判断的空余节点， 扩容时自动创建新的CVM节点加入到伸缩组<br/>关机模式：扩容时优先对已关机的节点执行开机操作，节点数依旧不满足要求时再创建新的CVM节点')">
      <bk-radio-group v-model="nodePoolInfo.autoScaling.scalingMode">
        <span class="inline-block" v-bk-tooltips="$t('自研上云环境暂不支持修改')">
          <bk-radio value="CLASSIC_SCALING" disabled>{{$t('释放模式')}}</bk-radio>
          <bk-radio value="WAKE_UP_STOPPED_SCALING" disabled>{{$t('关机模式')}}</bk-radio>
        </span>
      </bk-radio-group>
    </bk-form-item>
    <!-- <bk-form-item
      :label="$t('扩容后转移模块')"
      :desc="$t('扩容节点后节点转移到关联业务的CMDB模块')"
      error-display-type="normal"
      required
      property="nodeTemplate.module.scaleOutModuleID">
      <TopoSelectTree
        v-model="nodePoolInfo.nodeTemplate.module.scaleOutModuleID"
        :placeholder="$t('请选择业务 CMDB topo 模块')"
        :cluster-id="cluster.clusterID"
        @node-data-change="handleScaleOutDataChange" />
    </bk-form-item> -->
    <!-- <bk-form-item
      :label="$t('缩容后转移模块')"
      :desc="$t('缩容节点后节点转移到关联业务的CMDB模块，此选项仅适用于自有资源池场景，平台提供的资源池场景无需选择')">
      <TopoSelectTree
        v-model="nodePoolInfo.nodeTemplate.module.scaleInModuleID"
        :placeholder="$t('请选择业务 CMDB topo 模块')"
        :cluster-id="cluster.clusterID"
        @node-data-change="handleScaleInDataChange" />
    </bk-form-item> -->
  </bk-form>
</template>
<script lang="ts">
import { defineComponent, ref, toRefs } from 'vue';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import Taints from '@/views/cluster-manage/components/new-taints.vue';
import $i18n from '@/i18n/i18n-setup';
import Schema from '@/views/cluster-manage/cluster/autoscaler/resolve-schema';

export default defineComponent({
  name: 'BasciPoolInfo',
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
  setup(props) {
    const { defaultValues, schema } = toRefs(props);
    const formRef = ref<any>(null);
    const nodePoolInfo = ref({
      // name: defaultValues.value.name || '', // 节点名称
      autoScaling: {
        maxSize: defaultValues.value.autoScaling?.maxSize, // 节点数量范围
        minSize: defaultValues.value.autoScaling?.minSize,
        scalingMode: defaultValues.value.autoScaling?.scalingMode, // 扩缩容模式
        multiZoneSubnetPolicy: defaultValues.value.autoScaling?.multiZoneSubnetPolicy, // 实列创建策略
        retryPolicy: defaultValues.value.autoScaling?.retryPolicy, // 重试策略
      },
      enableAutoscale: defaultValues.value.enableAutoscale, // 是否开启弹性伸缩
      nodeTemplate: {
        unSchedulable: defaultValues.value.nodeTemplate?.unSchedulable || 0, // 是否开启调度 0 代表开启调度，1 不可调度
        labels: defaultValues.value.nodeTemplate?.labels || {}, // 标签
        taints: defaultValues.value.nodeTemplate?.taints || [], // 污点
        annotations: defaultValues.value.nodeTemplate?.annotations || {}, // 注解
        // module: {
        //   scaleOutModuleID: defaultValues.value.nodeTemplate?.module?.scaleOutModuleID || '',
        //   scaleOutModuleName: defaultValues.value.nodeTemplate?.module?.scaleOutModuleName || '',
        //   scaleInModuleID: defaultValues.value.nodeTemplate?.module?.scaleInModuleID || '',
        //   scaleInModuleName: defaultValues.value.nodeTemplate?.module?.scaleInModuleName || '',
        // },
      },
      // bkCloudID: defaultValues.value.bkCloudID || 0,
    });

    const nodePoolInfoRules = ref({
      labels: [
        {
          message: $i18n.t('标签值不能为空'),
          trigger: 'custom',
          // eslint-disable-next-line max-len
          validator: () => Object.keys(nodePoolInfo.value.nodeTemplate.labels).every(key => !!nodePoolInfo.value.nodeTemplate.labels[key]),
        },
        {
          message: $i18n.t('仅支持字母，数字和字符(-_./)'),
          trigger: 'custom',
          validator: () => {
            const keys = Object.keys(nodePoolInfo.value.nodeTemplate.labels);
            const values = keys.map(key => nodePoolInfo.value.nodeTemplate.labels[key]);
            return keys.every(v => /^[A-Za-z0-9._/-]+$/.test(v)) && values.every(v => /^[A-Za-z0-9._/-]+$/.test(v));
          },
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
        {
          message: $i18n.t('仅支持字母，数字和字符(-_./)'),
          trigger: 'custom',
          validator: () => (nodePoolInfo.value.nodeTemplate.taints as any[])
            .every(item => /^[A-Za-z0-9._/-]+$/.test(item.key) && /^[A-Za-z0-9._/-]+$/.test(item.value)),
        },
      ],
      annotations: [
        {
          message: $i18n.t('注解值不能为空'),
          trigger: 'custom',
          validator: () => Object.keys(nodePoolInfo.value.nodeTemplate.annotations)
            .every(key => !!nodePoolInfo.value.nodeTemplate.annotations[key]),
        },
        {
          message: $i18n.t('仅支持字母，数字和字符(-_./)'),
          trigger: 'custom',
          validator: () => {
            const keys = Object.keys(nodePoolInfo.value.nodeTemplate.annotations);
            const values = keys.map(key => nodePoolInfo.value.nodeTemplate.annotations[key]);
            return keys.every(v => /^[A-Za-z0-9._/-]+$/.test(v)) && values.every(v => /^[A-Za-z0-9._/-]+$/.test(v));
          },
        },
      ],
      // 'nodeTemplate.module.scaleOutModuleID': [
      //   {
      //     required: true,
      //     message: $i18n.t('必填项'),
      //     trigger: 'blur',
      //   },
      // ],
    });

    const validate = async () => {
      const result = await formRef.value?.validate();
      return result;
    };
    const handleAddTaints = () => {
      nodePoolInfo.value.nodeTemplate.taints.push({
        key: '',
        value: '',
        effect: 'PreferNoSchedule',
      });
    };

    const getSchemaByProp = props => Schema.getSchemaByProp(schema.value, props);


    return {
      formRef,
      nodePoolInfo,
      nodePoolInfoRules,
      getSchemaByProp,
      handleAddTaints,
      validate,
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
