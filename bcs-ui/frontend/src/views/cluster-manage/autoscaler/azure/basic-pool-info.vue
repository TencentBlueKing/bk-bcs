<!-- eslint-disable max-len -->
<template>
  <bk-form class="node-pool-info" :model="nodePoolInfo" :rules="nodePoolInfoRules" ref="formRef">
    <bk-form-item
      :label="$t('cluster.ca.nodePool.label.nodeQuota')"
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
      :label="$t('cluster.ca.nodePool.create.enableAutoscale.title')"
      :desc="$t('cluster.ca.nodePool.create.enableAutoscale.tips')">
      <bk-checkbox v-model="nodePoolInfo.enableAutoscale" :disabled="isEdit"></bk-checkbox>
    </bk-form-item>
    <bk-form-item :label="$t('k8s.label')" property="labels" error-display-type="normal">
      <KeyValue
        class="max-w-[420px]"
        :min-item="0"
        :disable-delete-item="false"
        v-model="nodePoolInfo.nodeTemplate.labels">
      </KeyValue>
    </bk-form-item>
    <bk-form-item :label="$t('k8s.taint')" property="taints" error-display-type="normal">
      <span
        class="add-key-value-items" v-if="!nodePoolInfo.nodeTemplate.taints.length"
        @click="handleAddTaints">
        <i class="bk-icon icon-plus-circle-shape mr5"></i>
        {{$t('generic.button.add')}}
      </span>
      <Taints
        class="taints"
        :effect-options="getSchemaByProp('nodeTemplate.taints.effect').enum"
        :key-rules="[]"
        v-model="nodePoolInfo.nodeTemplate.taints"
        v-else>
      </Taints>
    </bk-form-item>
    <bk-form-item :label="$t('k8s.annotation')" property="annotations" error-display-type="normal">
      <KeyValue
        class="max-w-[420px]"
        :min-item="0"
        :disable-delete-item="false"
        v-model="nodePoolInfo.nodeTemplate.annotations">
      </KeyValue>
    </bk-form-item>
    <bk-form-item
      :label="$t('cluster.ca.nodePool.create.scalingMode.title')"
      :desc="$t('cluster.ca.nodePool.create.scalingMode.desc')">
      <bk-radio-group v-model="nodePoolInfo.autoScaling.scalingMode">
        <span class="inline-block">
          <bk-radio disabled value="Delete">{{$t('cluster.ca.nodePool.create.scalingMode.classic_scaling')}}</bk-radio>
        </span>
      </bk-radio-group>
    </bk-form-item>
    <!-- <bk-form-item
      :label="$t('cluster.ca.nodePool.create.cloudArea.title')"
      :desc="$t('cluster.ca.nodePool.create.cloudArea.desc')">
      <bcs-select
        :clearable="false"
        :loading="cloudLoading"
        searchable
        v-model="nodePoolInfo.bkCloudID"
        @change="handleCloudIDChange">
        <bcs-option
          v-for="item in cloudList"
          :key="item.bk_cloud_id"
          :id="item.bk_cloud_id"
          :name="item.bk_cloud_name">
        </bcs-option>
      </bcs-select>
    </bk-form-item> -->
  </bk-form>
</template>
<script lang="ts">
import { sortBy } from 'lodash';
import { defineComponent, onMounted, ref, toRefs } from 'vue';

import { nodemanCloudList } from '@/api/base';
import $i18n from '@/i18n/i18n-setup';
import Schema from '@/views/cluster-manage/autoscaler/resolve-schema';
import KeyValue from '@/views/cluster-manage/components/key-value.vue';
import Taints from '@/views/cluster-manage/components/new-taints.vue';

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
    const { defaultValues, schema, isEdit } = toRefs(props);
    const formRef = ref<any>(null);
    const nodePoolInfo = ref({
      // name: defaultValues.value.name || '', // 节点名称
      autoScaling: {
        maxSize: defaultValues.value.autoScaling?.maxSize, // 节点数量范围
        minSize: defaultValues.value.autoScaling?.minSize,
        scalingMode: 'Delete', // 扩缩容模式, 释放模式改为Delete
      },
      enableAutoscale: defaultValues.value.enableAutoscale, // 是否开启弹性伸缩
      nodeTemplate: {
        unSchedulable: 1, // 是否开启调度 0 代表开启调度，1 不可调度
        labels: defaultValues.value.nodeTemplate?.labels || {}, // 标签
        taints: defaultValues.value.nodeTemplate?.taints || [], // 污点
        annotations: defaultValues.value.nodeTemplate?.annotations || {}, // 注解
      },
      bkCloudID: isEdit.value ? defaultValues.value.area?.bkCloudID : 0,
      bkCloudName: defaultValues.value.area?.bkCloudName || '',
    });

    const nodePoolInfoRules = ref({
      labels: [
        {
          message: $i18n.t('generic.validate.labelValueEmpty'),
          trigger: 'custom',
          // eslint-disable-next-line max-len
          validator: () => Object.keys(nodePoolInfo.value.nodeTemplate.labels).every(key => !!nodePoolInfo.value.nodeTemplate.labels[key]),
        },
        {
          message: $i18n.t('generic.validate.labelKey'),
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
          message: $i18n.t('generic.validate.required1'),
          trigger: 'custom',
          validator: () => nodePoolInfo.value.nodeTemplate.taints.every(item => item.key && item.value && item.effect),
        },
        {
          message: $i18n.t('generic.validate.repeatKey'),
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
          message: $i18n.t('generic.validate.labelKey'),
          trigger: 'custom',
          validator: () => (nodePoolInfo.value.nodeTemplate.taints as any[])
            .every(item => /^[A-Za-z0-9._/-]+$/.test(item.key) && /^[A-Za-z0-9._/-]+$/.test(item.value)),
        },
      ],
      annotations: [
        {
          message: $i18n.t('generic.validate.annotationValueEmpty'),
          trigger: 'custom',
          validator: () => Object.keys(nodePoolInfo.value.nodeTemplate.annotations)
            .every(key => !!nodePoolInfo.value.nodeTemplate.annotations[key]),
        },
        {
          message: $i18n.t('generic.validate.labelKey'),
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
      //     message: $i18n.t('generic.validate.required'),
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

    // 云区域列表
    const cloudList = ref<any[]>([]);
    const cloudLoading = ref(false);
    const handleGetCloudList = async () => {
      cloudLoading.value = true;
      const data = await nodemanCloudList().catch(() => []);
      cloudList.value = sortBy(data, 'bk_cloud_name').sort((item) => {
        if (item.bk_cloud_id === 0) return -1;
        return 0;
      });
      cloudLoading.value = false;
    };
    const handleCloudIDChange = (id: string) => {
      const data = cloudList.value.find(item => item.bk_cloud_id === id);
      if (data) {
        nodePoolInfo.value.bkCloudName = data.bk_cloud_name;
      }
    };

    onMounted(() => {
      handleGetCloudList();
    });

    return {
      formRef,
      nodePoolInfo,
      nodePoolInfoRules,
      getSchemaByProp,
      handleAddTaints,
      validate,
      cloudLoading,
      cloudList,
      handleCloudIDChange,
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
