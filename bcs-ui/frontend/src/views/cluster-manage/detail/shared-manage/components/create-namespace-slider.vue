<template>
  <!-- 创建命名空间侧边栏：归属项目（可见范围内）+ 前缀名称（实时预览/校验）+ 配额 -->
  <bcs-sideslider
    class="ns-create-slider"
    :is-show="value"
    :width="680"
    quick-close
    :title="$t('cluster.sharedManage.assign.create.title')"
    :before-close="handleBeforeClose"
    @hidden="handleClose(false)">
    <template #content>
      <div class="px-[32px] pt-[24px] pb-[8px]" v-bkloading="{ isLoading: loading, opacity: 1 }">
        <bcs-form :label-width="100">
          <bcs-form-item :label="$t('cluster.sharedManage.assign.create.project')" required>
            <bcs-select
              class="w-full"
              :value="target"
              searchable
              :clearable="false"
              :placeholder="$t('cluster.sharedManage.assign.slider.selectProject')"
              @change="handleTargetChange">
              <bcs-option
                v-for="project in projects"
                :key="project.projectID"
                :id="project.projectCode"
                :name="`${project.name}（${project.projectCode}）`" />
            </bcs-select>
            <p v-if="errs.target" class="text-[12px] text-[#EA3636]">{{ errs.target }}</p>
            <p class="text-[12px] text-[#979BA5]">
              {{ $t('cluster.sharedManage.assign.slider.onlyRange') }}
              <bcs-button text theme="primary" @click="handleGoSettings">
                <span class="text-[12px]">{{ $t('cluster.sharedManage.assign.slider.adjustRange') }}</span>
              </bcs-button>
            </p>
          </bcs-form-item>

          <bcs-form-item :label="$t('cluster.sharedManage.assign.create.name')" required>
            <bcs-input
              :value="suffix"
              :maxlength="64"
              :placeholder="$t('cluster.sharedManage.assign.create.suffixPlaceholder')"
              @input="handleSuffixInput">
              <template #prepend><span class="group-text">{{ nsPrefix }}</span></template>
            </bcs-input>
            <p v-if="errs.suffix" class="text-[12px] text-[#EA3636]">{{ errs.suffix }}</p>
          </bcs-form-item>

          <bcs-form-item :label="$t('cluster.sharedManage.assign.slider.quota')" required>
            <div class="flex items-center gap-[8px] text-[12px]">
              <span class="text-[#63656E]">CPU</span>
              <bcs-input v-model="cpu" class="w-[130px]" type="number" :min="1" :precision="0">
                <template #append>
                  <span class="group-text">{{ $t('cluster.sharedManage.assign.unit.cpu') }}</span>
                </template>
              </bcs-input>
              <span class="text-[#63656E]">{{ $t('cluster.sharedManage.assign.unit.mem') }}</span>
              <bcs-input v-model="mem" class="w-[130px]" type="number" :min="1" :precision="0">
                <template #append><span class="group-text">GiB</span></template>
              </bcs-input>
            </div>
            <p v-if="errs.quota" class="text-[12px] text-[#EA3636]">{{ errs.quota }}</p>
            <p class="text-[12px] text-[#979BA5]">
              {{ $t('cluster.sharedManage.assign.create.quotaTips') }}
            </p>
          </bcs-form-item>
        </bcs-form>
        <bcs-alert
          type="info"
          class="mt-[8px]"
          :title="$t('cluster.sharedManage.assign.create.info')">
        </bcs-alert>
      </div>
    </template>
    <template #footer>
      <div class="flex items-center px-[24px]">
        <bcs-button theme="primary" :loading="loading" @click="handleSubmit">
          {{ $t('cluster.sharedManage.assign.create.create') }}
        </bcs-button>
        <bcs-button :disabled="loading" @click="handleClose(false)">
          {{ $t('generic.button.cancel') }}
        </bcs-button>
      </div>
    </template>
  </bcs-sideslider>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import { createdNamespace } from '@/api/modules/project';
import { IResourceQuota } from '@/api/modules/shared-manage';
import $bkMessage from '@/common/bkmagic';
import { useProject } from '@/composables/use-app';

interface IProjectOption {
  projectID: string
  projectCode: string
  name: string
}

// 创建命名空间侧边栏（所属项目直接创建，不需要审批）
// 名称 = 平台配置前缀（默认 bcs）-所选项目英文名- + 用户输入后缀，与使用方申请规则一致
const props = defineProps({
  value: {
    type: Boolean,
    default: false,
  },
  clusterId: {
    type: String,
    default: '',
  },
  // 可见范围内的项目（不含集群所属项目）
  projects: {
    type: Array as () => IProjectOption[],
    default: () => [],
  },
  // 项目编码 -> 名称
  projectMap: {
    type: Object as () => Record<string, string>,
    default: () => ({}),
  },
  // 集群已有命名空间（重复校验）
  existingNames: {
    type: Array as () => string[],
    default: () => [],
  },
});

const emits = defineEmits(['input', 'refresh', 'go-settings']);

const { curProject } = useProject();

const target = ref('');
const suffix = ref('');
const cpu = ref(8);
const mem = ref(16);
const errs = ref({ target: '', suffix: '', quota: '' });
const loading = ref(false);

watch(() => props.value, (value) => {
  if (!value) return;
  target.value = '';
  suffix.value = '';
  cpu.value = 8;
  mem.value = 16;
  errs.value = { target: '', suffix: '', quota: '' };
  loading.value = false;
});

// 前缀由所选项目自动生成、不可修改；未选择项目时仅展示平台前缀
const nsPrefix = computed(() => `${window.BCS_NAMESPACE_PREFIX || 'bcs'}-${target.value}-`);
const fullName = computed(() => (target.value && suffix.value ? `${nsPrefix.value}${suffix.value}` : ''));
const projectName = computed(() => props.projectMap[target.value] || target.value);

const handleTargetChange = (value: string) => {
  target.value = value;
  errs.value.target = '';
};
const handleSuffixInput = (value: string) => {
  suffix.value = value.trim();
  errs.value.suffix = '';
};

const handleSubmit = async () => {
  const next = { target: '', suffix: '', quota: '' };
  if (!target.value) next.target = window.i18n.t('cluster.sharedManage.assign.create.errProject');
  if (!suffix.value) {
    next.suffix = window.i18n.t('cluster.sharedManage.assign.create.errSuffix');
  } else if (!/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(suffix.value)) {
    next.suffix = window.i18n.t('cluster.sharedManage.assign.create.errFormat');
  } else if (fullName.value.length > 63) {
    next.suffix = window.i18n.t('cluster.sharedManage.assign.create.errLength', { len: fullName.value.length });
  } else if (props.existingNames.includes(fullName.value)) {
    next.suffix = window.i18n.t('cluster.sharedManage.assign.create.errExist', { name: fullName.value });
  }
  if (Number(cpu.value) < 1 || Number(mem.value) < 1) {
    next.quota = window.i18n.t('cluster.sharedManage.assign.create.errQuota');
  }
  errs.value = next;
  if (next.target || next.suffix || next.quota) return;

  loading.value = true;
  const quota: IResourceQuota = {
    cpuLimits: String(cpu.value),
    cpuRequests: String(cpu.value),
    memoryLimits: `${mem.value}Gi`,
    memoryRequests: `${mem.value}Gi`,
  };
  // 复用使用方创建命名空间接口：当前项目（集群所属项目）作为操作方
  const result = await createdNamespace({
    $projectCode: curProject.value?.projectCode,
    $clusterId: props.clusterId,
    name: fullName.value,
    quota,
    labels: [],
    annotations: [],
  }).then(() => true)
    .catch(() => false);
  loading.value = false;
  if (result) {
    $bkMessage({
      theme: 'success',
      message: window.i18n.t('cluster.sharedManage.assign.create.successMsg', {
        project: projectName.value,
        name: fullName.value,
      }),
    });
    emits('input', false);
    emits('refresh');
  }
};

// 提交中禁止关闭（quick-close / esc）
const handleBeforeClose = () => !loading.value;

const handleClose = (value: boolean) => {
  if (loading.value) return;
  emits('input', value);
};

const handleGoSettings = () => {
  emits('input', false);
  emits('go-settings');
};
</script>

<style lang="postcss" scoped>
:deep(.bk-sideslider-content) {
  overflow: auto;
  height: calc(100vh - 100px);
}

:deep(.bk-sideslider-footer) {
  background-color: #fff !important;
  border-top-color: #dcdee5 !important;
}

:deep(.bk-form-control .group-box .group-text) {
  line-height: 30px;
}

:deep(.bk-form .bk-label) {
  font-size: 12px;
  color: #313238;
}
</style>
