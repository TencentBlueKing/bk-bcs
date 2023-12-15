<template>
  <LayoutContent :title="$tc('新建命名空间')">
    <div class="p-[20px] h-full overflow-auto">
      <div class="border border-solid border-[#dcdee5] p-[20px] bg-[#ffffff]">
        <bk-form
          ref="namespaceForm"
          v-model="formData"
          :rules="rules"
          form-type="vertical">
          <bk-form-item
            v-if="isSharedCluster"
            :label="$t('名称')"
            :required="true"
            property="name"
            error-display-type="normal"
            :desc="$t('命名规则：ieg-项目英文名称-自定义名称')">
            <bk-input v-model="formData.name" class="w-[620px]" maxlength="30">
              <div slot="prepend">
                <div class="group-text">{{ 'ieg-' + projectCode + '-' }}</div>
              </div>
            </bk-input>
          </bk-form-item>
          <bk-form-item
            v-else
            :label="$t('名称')"
            :required="true"
            error-display-type="normal"
            property="name">
            <bk-input v-model="formData.name" class="w-[620px]"></bk-input>
          </bk-form-item>
          <bk-form-item
            v-if="!isSharedCluster"
            :label="$t('标签')"
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
                <bk-input :placeholder="$t('值')" v-model="label.value" class="w-[300px]" @blur="validate"></bk-input>
              </bk-form-item>
              <i class="bk-icon icon-minus-line ml-[5px] cursor-pointer" @click="handleRemoveLabel(index)" />
            </bk-form>
            <i class="bk-icon icon-plus-line ml-[5px] cursor-pointer" @click="handleAddLabel" />
          </bk-form-item>
          <bk-form-item
            v-if="!isSharedCluster"
            :label="$t('注解')"
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
                <bk-input v-model="annotation.value" placeholder="Value" class="w-[300px]" @blur="validate"></bk-input>
              </bk-form-item>
              <i class="bk-icon icon-minus-line ml-[5px] cursor-pointer" @click="handleRemoveAnnotation(index)"></i>
            </bk-form>
            <i class="bk-icon icon-plus-line ml-[5px] cursor-pointer" @click="handleAddAnnotation"></i>
          </bk-form-item>
          <bk-form-item
            :label="$t('配额设置')"
            :required="isSharedCluster"
            error-display-type="normal"
            property="quota"
            v-bind="quotaFormItemConfig">
            <div class="flex">
              <div class="flex mr-[20px]">
                <span class="mr-[15px] text-[14px]">CPU</span>
                <bk-input
                  v-model="formData.quota.cpuRequests"
                  class="w-[250px]"
                  type="number"
                  :min="1"
                  int
                  :max="512000"
                  :precision="0">
                  <div class="group-text" slot="append">{{ $t('核') }}</div>
                </bk-input>
              </div>
              <div class="flex">
                <span class="mr-[15px] text-[14px]">MEM</span>
                <bk-input
                  v-model="formData.quota.memoryRequests"
                  class="w-[250px]"
                  type="number"
                  :min="1"
                  int
                  :max="1024000"
                  :precision="0">
                  <div class="group-text" slot="append">G</div>
                </bk-input>
              </div>
            </div>
          </bk-form-item>
          <div id="quota-tip" v-if="isSharedCluster">
            <p>{{ $t('1.创建命名会进入审批流程，如需加急审批请主动联系审批人') }}</p>
            <p>{{ $t('2.为了避免产生过多资源碎片，CPU/内存资源比不应大于1/4') }}</p>
          </div>
        </bk-form>
      </div>
    </div>
    <div>
      <bcs-button
        theme="primary"
        class="w-[88px] mr-[10px] ml-[20px]"
        @click="handleCreated"
        :loading="isLoading">{{ $t('创建') }}</bcs-button>
      <bcs-button class="w-[88px]" :loading="isLoading" @click="handleCancel">{{ $t('取消') }}</bcs-button>
    </div>
  </LayoutContent>
</template>

<script>
import { defineComponent, computed, ref } from '@vue/composition-api';
import LayoutContent from '@/components/layout/Content.vue';
import { useNamespace } from './use-namespace';
import { useCluster } from '@/common/use-app';
import { KEY_REGEXP } from '@/common/constant';

export default defineComponent({
  name: 'CreateNamespace',
  components: {
    LayoutContent,
  },
  setup(props, ctx) {
    const { $bkInfo, $store, $i18n, $router, $route, $bkMessage } = ctx.root;
    const quotaTipsCof = ref({
      allowHtml: true,
      width: 380,
      content: '#quota-tip',
      placement: 'top',
    });
    const formData = ref({
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

    const quotaFormItemConfig = computed(() => {
      const config = {
        required: isSharedCluster.value,

      };
      if (isSharedCluster.value) {
        config.desc = quotaTipsCof.value;
      };
      return config;
    });

    const clusterId = computed(() => $route.params.clusterId);

    const { isSharedCluster } = useCluster();
    const { handleCreatedNamespace } = useNamespace();

    const rules = {
      name: [
        {
          validator() {
            return /^[a-z0-9]([-a-z0-9]*[a-z0-9]){0,64}?$/g.test(formData.value.name);
          },
          message: $i18n.t('命名空间名称只能包含小写字母、数字以及连字符(-)，连字符（-）后面必须接英文或者数字'),
          trigger: 'blur',
        },
      ],
      quota: isSharedCluster.value ? [
        {
          validator() {
            return Boolean(formData.value.quota.cpuRequests) >= 1 && Boolean(formData.value.quota.memoryRequests) >= 1;
          },
          message: $i18n.t('共享集群需设置MEM、CPU配额，且两者最小值不小于0'),
          trigger: 'blur',
        },
      ] : [],
      labels: [
        {
          validator() {
            // eslint-disable-next-line no-eval
            const regx = new RegExp(KEY_REGEXP);
            return formData.value.labels.every(item => regx.test(item.key) && regx.test(item.value));
          },
          message: $i18n.t('仅支持字母，数字，\'-\'，\'_\'，\'.\' 及 \'/\' 且需以字母数字开头和结尾'),
          trigger: 'blur',
        },
      ],
      annotations: [
        {
          validator() {
            // eslint-disable-next-line no-eval
            const regx = new RegExp(KEY_REGEXP);
            return formData.value.annotations.every(item => regx.test(item.key) && regx.test(item.value));
          },
          message: $i18n.t('仅支持字母，数字，\'-\'，\'_\'，\'.\' 及 \'/\' 且需以字母数字开头和结尾'),
          trigger: 'blur',
        },
      ],
    };

    const projectCode = computed(() => $route.params.projectCode);
    const namespaceForm = ref();
    const isLoading = ref(false);

    const handleCancel = () => {
      $bkInfo({
        type: 'warning',
        clsName: 'custom-info-confirm',
        title: $i18n.t('确认退出当前编辑状态'),
        subTitle: $i18n.t('退出后，你修改的内容将丢失'),
        defaultInfo: true,
        confirmFn: () => {
          $router.push({ name: $store.getters.curNavName });
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
    const handleCreated = () => {
      namespaceForm.value.validate().then(async () => {
        let { name } = formData.value;
        if (isSharedCluster.value) {
          name = `ieg-${projectCode.value}-${name}`;
        }
        isLoading.value = true;
        let quota = null;
        if (formData.value.quota.cpuRequests || formData.value.quota.memoryRequests) {
          quota = Object.assign({}, {
            cpuLimits: String(formData.value.quota.cpuRequests),
            cpuRequests: String(formData.value.quota.cpuRequests),
            memoryLimits: `${formData.value.quota.memoryRequests}Gi`,
            memoryRequests: `${formData.value.quota.memoryRequests}Gi`,
          });
        }
        const result = await handleCreatedNamespace({
          $clusterId: clusterId.value,
          ...formData.value,
          name,
          quota,
        });
        isLoading.value = false;
        if (result) {
          $bkMessage({
            theme: 'success',
            message: $i18n.t('创建成功'),
          });
          $router.push({
            name: $store.getters.curNavName,
          });
        };
      });
    };

    return {
      rules,
      quotaFormItemConfig,
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
