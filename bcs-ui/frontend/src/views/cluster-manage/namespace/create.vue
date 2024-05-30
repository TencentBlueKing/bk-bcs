<template>
  <LayoutContent :title="$tc('dashboard.ns.create.title')" :cluster-id="clusterId">
    <div class="p-[20px] h-full overflow-auto">
      <div class="border border-solid border-[#dcdee5] p-[20px] bg-[#ffffff]">
        <bk-form
          ref="namespaceForm"
          v-model="formData"
          :rules="rules"
          form-type="vertical">
          <!-- 共享集群 -->
          <template v-if="isSharedCluster">
            <bk-form-item
              :label="$t('generic.label.name')"
              required
              property="name"
              error-display-type="normal"
              :desc="$t('dashboard.ns.validate.sharedClusterNs', { prefix: nsPrefix })">
              <bk-input v-model="formData.name" class="w-[620px]" maxlength="30">
                <div slot="prepend">
                  <div class="group-text">{{ nsPrefix }}</div>
                </div>
              </bk-input>
            </bk-form-item>
          </template>
          <!-- 普通集群 -->
          <template v-else>
            <bk-form-item
              :label="$t('generic.label.name')"
              :required="true"
              error-display-type="normal"
              property="name">
              <bk-input v-model="formData.name" class="w-[620px]"></bk-input>
            </bk-form-item>
            <bk-form-item
              :label="$t('k8s.label')"
              error-display-type="normal"
              property="labels">
              <bk-form
                class="flex mb-[10px] items-center"
                v-for="(label, index) in formData.labels"
                :key="index"
                ref="labelsForm">
                <bk-form-item>
                  <bk-input v-model="label.key" placeholder="Key" class="w-[300px]" @blur="validate"></bk-input>
                </bk-form-item>
                <span class="px-[5px]">=</span>
                <bk-form-item>
                  <bk-input
                    :placeholder="$t('generic.label.value')"
                    v-model="label.value"
                    class="w-[300px]"
                    @blur="validate">
                  </bk-input>
                </bk-form-item>
                <i class="bk-icon icon-minus-line ml-[5px] cursor-pointer" @click="handleRemoveLabel(index)" />
              </bk-form>
              <span
                class="text-[14px] text-[#3a84ff] cursor-pointer flex items-center h-[32px]"
                @click="handleAddLabel">
                <i class="bk-icon icon-plus-circle-shape mr5"></i>
                {{$t('generic.button.add')}}
              </span>
            </bk-form-item>
            <bk-form-item
              :label="$t('k8s.annotation')"
              error-display-type="normal"
              property="annotations">
              <bk-form
                class="flex mb-[10px] items-center"
                v-for="(annotation, index) in formData.annotations"
                :key="index"
                ref="annotationsForm">
                <bk-form-item>
                  <bk-input v-model="annotation.key" placeholder="Key" class="w-[300px]" @blur="validate"></bk-input>
                </bk-form-item>
                <span class="px-[5px]">=</span>
                <bk-form-item>
                  <bk-input
                    v-model="annotation.value"
                    placeholder="Value"
                    class="w-[300px]"
                    @blur="validate">
                  </bk-input>
                </bk-form-item>
                <i class="bk-icon icon-minus-line ml-[5px] cursor-pointer" @click="handleRemoveAnnotation(index)"></i>
              </bk-form>
              <span
                class="text-[14px] text-[#3a84ff] cursor-pointer flex items-center h-[32px]"
                @click="handleAddAnnotation">
                <i class="bk-icon icon-plus-circle-shape mr5"></i>
                {{$t('generic.button.add')}}
              </span>
            </bk-form-item>
          </template>
          <bk-form-item
            :label="$t('dashboard.ns.create.quota')"
            :required="isSharedCluster"
            :desc="isSharedCluster ? {
              content: $t('dashboard.ns.create.sharedClusterQuotaTips'),
              width: 360,
            } : ''"
            error-display-type="normal"
            property="quota">
            <div class="flex">
              <div class="flex mr-[20px]">
                <span class="mr-[15px] text-[14px]">CPU</span>
                <bcs-input
                  v-model="formData.quota.cpuRequests"
                  class="w-[250px]"
                  type="number"
                  :min="1"
                  :max="512000"
                  :precision="0">
                  <div class="group-text" slot="append">{{ $t('units.suffix.cores') }}</div>
                </bcs-input>
              </div>
              <div class="flex">
                <span class="mr-[15px] text-[14px]">MEM</span>
                <bcs-input
                  v-model="formData.quota.memoryRequests"
                  class="w-[250px]"
                  type="number"
                  :min="1"
                  :max="1024000"
                  :precision="0">
                  <div class="group-text" slot="append">GiB</div>
                </bcs-input>
              </div>
            </div>
          </bk-form-item>
        </bk-form>
      </div>
    </div>
    <div>
      <bcs-button
        theme="primary"
        class="w-[88px] ml-[20px]"
        @click="handleCreated"
        :loading="isLoading">{{ $t('generic.button.create') }}</bcs-button>
      <bcs-button
        class="w-[88px]"
        :disabled="isLoading"
        @click="handleCancel">{{ $t('generic.button.cancel') }}</bcs-button>
    </div>
  </LayoutContent>
</template>

<script lang='ts'>
import { computed, defineComponent, ref } from 'vue';

import { useNamespace } from './use-namespace';

import $bkMessage from '@/common/bkmagic';
import { KEY_REGEXP } from '@/common/constant';
import $bkInfo from '@/components/bk-magic-2.0/bk-info';
import LayoutContent from '@/components/layout/Content.vue';
import { useCluster, useProject } from '@/composables/use-app';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router';

export default defineComponent({
  name: 'CreateNamespace',
  components: {
    LayoutContent,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const formData = ref<{
      name: string
      quota: {
        cpuLimits: string
        cpuRequests: string
        memoryLimits: string
        memoryRequests: string
      }
      labels: any[]
      annotations: any[]
    }>({
      name: '',
      quota: {
        cpuLimits: '',
        cpuRequests: '',
        memoryLimits: '',
        memoryRequests: '',
      },
      labels: [],
      annotations: [],
    });

    const { clusterMap } = useCluster();
    const isSharedCluster = computed(() => !!clusterMap.value[props.clusterId]?.is_shared);
    const { handleCreatedNamespace } = useNamespace();

    const rules = {
      name: [
        {
          validator() {
            return /^[a-z0-9]([-a-z0-9]*[a-z0-9]){0,64}?$/g.test(formData.value.name);
          },
          message: $i18n.t('dashboard.ns.validate.name'),
          trigger: 'blur',
        },
      ],
      quota: isSharedCluster.value ? [
        {
          validator() {
            return Number(formData.value.quota.cpuRequests) >= 1 && Number(formData.value.quota.memoryRequests) >= 1;
          },
          message: $i18n.t('dashboard.ns.validate.setMinMaxMemCpu'),
          trigger: 'blur',
        },
      ] : [],
      labels: [
        {
          validator() {
            // eslint-disable-next-line no-eval
            const regx = new RegExp(KEY_REGEXP);
            return formData.value.labels.every(item => item.key && regx.test(item.key) && regx.test(item.value));
          },
          message: $i18n.t('generic.validate.labelKey1'),
          trigger: 'blur',
        },
      ],
      annotations: [
        {
          validator() {
            // eslint-disable-next-line no-eval
            const regx = new RegExp(KEY_REGEXP);
            return formData.value.annotations.every(item => item.key && regx.test(item.key));
          },
          message: $i18n.t('generic.validate.labelKey1'),
          trigger: 'blur',
        },
      ],

    };

    const { projectCode } = useProject();
    const namespaceForm = ref();
    const isLoading = ref(false);

    const handleCancel = () => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('generic.msg.info.exitEdit.text'),
        subTitle: $i18n.t('generic.msg.info.exitEdit.subTitle'),
        defaultInfo: true,
        confirmFn: () => {
          $router.back();
        },
      });
    };

    const handleAddLabel = () => {
      const label = { key: '', value: '' };
      formData.value.labels.push(label);
    };

    const handleRemoveLabel = (index) => {
      formData.value.labels.splice(index, 1);
    };

    const handleAddAnnotation = () => {
      const label = { key: '', value: '' };
      formData.value.annotations.push(label);
    };

    const handleRemoveAnnotation = (index) => {
      formData.value.annotations.splice(index, 1);
    };

    const validate = () => {
      namespaceForm.value.validate();
    };
    const nsPrefix = computed(() => `${window.BCS_NAMESPACE_PREFIX || 'bcs'}-${projectCode.value}-`);
    const handleCreated = () => {
      namespaceForm.value.validate().then(async () => {
        if (!props.clusterId) return;
        let { name } = formData.value;
        if (isSharedCluster.value) {
          name = `${nsPrefix.value}${name}`;
        }
        isLoading.value = true;
        let quota: Record<string, string> | null = null;
        if (formData.value.quota.cpuRequests || formData.value.quota.memoryRequests) {
          quota = {
            cpuLimits: String(formData.value.quota.cpuRequests),
            cpuRequests: String(formData.value.quota.cpuRequests),
            memoryLimits: `${formData.value.quota.memoryRequests}Gi`,
            memoryRequests: `${formData.value.quota.memoryRequests}Gi`,
          };
        }
        const result = await handleCreatedNamespace({
          $clusterId: props.clusterId,
          ...formData.value,
          name,
          quota,
        });
        isLoading.value = false;
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('generic.msg.success.create'),
          });
          $router.back();
        };
      })
        .catch(() => false);
    };

    return {
      nsPrefix,
      rules,
      formData,
      isLoading,
      isSharedCluster,
      namespaceForm,
      projectCode,
      handleCancel,
      handleAddLabel,
      handleRemoveLabel,
      handleAddAnnotation,
      handleRemoveAnnotation,
      handleCreated,
      validate,
    };
  },
});
</script>
