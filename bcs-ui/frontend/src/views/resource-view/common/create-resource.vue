<template>
  <bk-dialog
    :is-show="show"
    render-directive="if"
    :title="title"
    width="480"
    @cancel="cancel">
    <bk-form :label-width="110">
      <bk-form-item :label="$t('generic.label.editMode.text')" required>
        <bk-radio-group v-model="editMode">
          <bk-radio value="form" :disabled="disabledFormMode">
            {{ $t('generic.label.editMode.form') }}
          </bk-radio>
          <bk-radio value="yaml">YAML</bk-radio>
        </bk-radio-group>
      </bk-form-item>
      <bk-form-item :label="$t('generic.label.cluster')" required>
        <ClusterSelect class="!w-full" v-model="clusterID" cluster-type="all" />
      </bk-form-item>
      <bk-form-item :label="$t('k8s.namespace')" required>
        <NamespaceSelect :cluster-id="clusterID" v-model="namespace" />
      </bk-form-item>
    </bk-form>
    <template #footer>
      <div>
        <bk-button
          theme="primary"
          :disabled="!namespace
            || !clusterID
            || disabledCreateOfSharedCluster"
          @click="confirm">
          <div
            v-bk-tooltips="{
              content: $t('view.tips.sharedClusterDoNotSupportTheKind'),
              disabled: !disabledCreateOfSharedCluster
            }">
            {{ $t('dashboard.button.createResource') }}
          </div>
        </bk-button>
        <bk-button @click="cancel">{{ $t('generic.button.cancel') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<script setup lang="ts">
import { computed, onBeforeMount, PropType, ref } from 'vue';

import ClusterSelect from '@/components/cluster-selector/cluster-select.vue';
import NamespaceSelect from '@/components/namespace-selector/namespace-select.vue';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  type: {
    type: String,
    default: '',
    required: true,
  },
  category: {
    type: String,
    default: '',
  },
  kind: {
    type: String,
    default: '',
  },
  crd: {
    type: String,
    default: '',
  },
  formUpdate: {
    type: Boolean,
    default: false,
  },
  cancel: {
    type: Function,
  },
  // CRD资源的作用域
  scope: {
    type: String as PropType<'Namespaced'|'Cluster'>,
    default: '',
  },
  // CRD资源分两种，普通和定制，customized 用来区分普通和定制
  customized: {
    type: Boolean,
    default: false,
  },
  // CRD信息
  crdOptions: {
    type: Object as PropType<{
      group: string,
      version: string,
      resource: string,
      namespaced: boolean,
    }>,
    default: () => ({}),
  },
});

const emits = defineEmits(['cancel', 'confirm']);
const title = computed(() => `${$i18n.t('generic.button.create')} ${props.kind}`);

const { clusterList } = useClusterList();
const editMode = ref<'yaml'|'form'>('form');
const clusterID = ref('');
const namespace = ref('');
const disabledKindListOfSharedCluster = ref(['DaemonSet', 'PersistentVolume', 'StorageClass']);
const gameCRDList = ref(['GameDeployment', 'GameStatefulSet', 'HookTemplate', 'BscpConfig']);
const curCluster = computed(() => clusterList.value.find(item => item.clusterID === clusterID.value));
const disabledFormMode = computed(() => props.category === 'custom_objects'
  && !gameCRDList.value.includes(props.kind));

// 共享集群资源: 'DaemonSet', 'PersistentVolume', 'StorageClass' 和 scope为Cluster的不支持创建
const disabledCreateOfSharedCluster = computed(() => curCluster.value?.is_shared
  && (disabledKindListOfSharedCluster.value.includes(props.kind) || props.scope === 'Cluster'));

// 创建资源
const handleCreateResource = () => {
  $router.push({
    name: 'dashboardResourceUpdate',
    params: {
      defaultShowExample: (props.kind !== 'CustomObject') as any,
      namespace: namespace.value,
      clusterId: clusterID.value,
    },
    query: {
      type: props.type,
      category: props.category,
      kind: props.kind,
      crd: props.crd,
      formUpdate: props.formUpdate,
      scope: props.scope,
      customized: props.customized,
      version: props.crdOptions.version,
      group: props.crdOptions.group,
      resource: props.crdOptions.resource,
    },
  });
};
// 创建资源（表单模式）
const handleCreateFormResource = () => {
  $router.push({
    name: 'dashboardFormResourceUpdate',
    params: {
      namespace: namespace.value,
      clusterId: clusterID.value,
    },
    query: {
      type: props.type,
      category: props.category,
      kind: props.kind,
      crd: props.crd,
      formUpdate: props.formUpdate,
    },
  });
};

const cancel = () => {
  emits('cancel');
  props.cancel?.();
};
const confirm = () => {
  emits('confirm');
  if (editMode.value === 'form') {
    handleCreateFormResource();
  } else {
    handleCreateResource();
  }
};

onBeforeMount(() => {
  if (props.category === 'custom_objects' && !gameCRDList.value.includes(props.kind)) {
    editMode.value = 'yaml';
  } else {
    editMode.value = 'form';
  }
});
</script>
