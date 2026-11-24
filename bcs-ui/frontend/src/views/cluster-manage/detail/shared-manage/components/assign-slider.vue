<template>
  <!-- 分配 / 变更归属侧边栏：表单阶段（命名空间列表 + 目标项目 + 配额）与结果阶段（逐个结果 + 重试失败项） -->
  <bcs-sideslider
    class="ns-assign-slider"
    :is-show="value"
    :width="680"
    quick-close
    :title="isTransfer
      ? $t('cluster.sharedManage.assign.slider.transferTitle')
      : $t('cluster.sharedManage.assign.slider.assignTitle')"
    :before-close="handleBeforeClose"
    @hidden="handleClose(false)">
    <template #content>
      <div class="px-[32px] pt-[24px] pb-[8px]" v-bkloading="{ isLoading: submitting, opacity: 1 }">
        <!-- 表单阶段 -->
        <template v-if="!results.length">
          <bcs-form :label-width="100">
            <bcs-form-item :label="$t('cluster.sharedManage.assign.slider.namespaces')">
              <div class="bg-[#fff] border border-[#F0F1F5] rounded-[2px] max-h-[168px] overflow-y-auto">
                <div
                  v-for="ns in items"
                  :key="ns.name"
                  class="flex items-center justify-between gap-[12px] text-[12px] leading-[20px] px-[10px] py-[5px]
                    border-b border-[#F0F1F5] last:border-b-0">
                  <span class="truncate text-[#313238]">{{ ns.name }}</span>
                  <span class="flex items-center shrink-0">
                    <bcs-tag v-if="!ns.projectCode" class="!h-[20px] !leading-[20px]">
                      {{ $t('cluster.sharedManage.assign.tag.unassigned') }}
                    </bcs-tag>
                    <span v-else class="text-[#979BA5]">
                      {{ $t('cluster.sharedManage.assign.slider.currentOwner', { name: projectName(ns.projectCode) }) }}
                    </span>
                  </span>
                </div>
              </div>
            </bcs-form-item>
            <!-- mixed 模式：勾选项里已有归属，提交后改为目标项目 -->
            <bcs-alert
              v-if="!isTransfer && mixedCount"
              type="warning"
              class="mt-[4px] mb-[16px]"
              :title="$t('cluster.sharedManage.assign.slider.mixedWarn', { num: mixedCount })">
            </bcs-alert>

            <bcs-form-item :label="$t('cluster.sharedManage.assign.slider.target')" required>
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
                {{ items.length > 1
                  ? $t('cluster.sharedManage.assign.slider.quotaTipsBatch')
                  : $t('cluster.sharedManage.assign.slider.quotaTips') }}
              </p>
            </bcs-form-item>
          </bcs-form>
          <bcs-alert
            type="info"
            class="mt-[8px]"
            :title="$t('cluster.sharedManage.assign.slider.info')">
          </bcs-alert>
        </template>

        <!-- 结果阶段：按命名空间逐个返回结果，部分失败不回滚 -->
        <template v-else>
          <div class="flex items-center gap-[10px] text-[14px] leading-[22px]">
            <i
              :class="[
                'w-[24px] h-[24px] rounded-full text-white not-italic flex items-center justify-center',
                'text-[14px] font-semibold shrink-0',
                failCount ? 'bg-[#FF9C01]' : 'bg-[#2DCB56]',
              ]">
              {{ failCount ? '!' : '✓' }}
            </i>
            <span class="text-[#313238]">
              {{ $t('cluster.sharedManage.assign.slider.resultTitle', { name: targetName, ok: okCount }) }}
              <span v-if="failCount" class="text-[#EA3636]">
                {{ $t('cluster.sharedManage.assign.slider.resultFail', { num: failCount }) }}
              </span>
            </span>
          </div>
          <div class="mt-[12px] border border-[#F0F1F5] rounded-[2px] max-h-[400px] overflow-y-auto">
            <div
              v-for="result in results"
              :key="result.name"
              class="flex items-center justify-between gap-[12px] text-[12px] leading-[20px] px-[10px] py-[6px]
                border-b border-[#F0F1F5] last:border-b-0">
              <span class="truncate text-[#313238]">{{ result.name }}</span>
              <bcs-tag v-if="result.ok" theme="success" class="shrink-0 !h-[22px] !leading-[22px]">
                {{ $t('cluster.sharedManage.assign.slider.success') }}
              </bcs-tag>
              <span v-else class="text-[#EA3636]">{{ result.msg }}</span>
            </div>
          </div>
        </template>
      </div>
    </template>
    <template #footer>
      <div class="flex items-center px-[24px]">
        <template v-if="!results.length">
          <bcs-button theme="primary" :loading="submitting" @click="handleSubmit">
            {{ $t('cluster.sharedManage.assign.slider.ok') }}
          </bcs-button>
          <bcs-button :disabled="submitting" @click="handleClose(false)">
            {{ $t('generic.button.cancel') }}
          </bcs-button>
        </template>
        <template v-else>
          <bcs-button v-if="failCount" theme="primary" :loading="submitting" @click="handleRetry">
            {{ $t('cluster.sharedManage.assign.slider.retry') }}
          </bcs-button>
          <bcs-button :theme="failCount ? 'default' : 'primary'" :disabled="submitting" @click="handleFinish">
            {{ $t('cluster.sharedManage.assign.slider.done') }}
          </bcs-button>
        </template>
      </div>
    </template>
  </bcs-sideslider>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import {
  allocateNamespace,
  INativeNamespace,
  IResourceQuota,
  parseCpu,
  parseMemoryGiB,
} from '@/api/modules/shared-manage';
import $bkMessage from '@/common/bkmagic';

interface IProjectOption {
  projectID: string
  projectCode: string
  name: string
}

// 分配 / 变更归属侧边栏
// assign：勾选命名空间批量分配（可含已有归属的，mixed 提示）；transfer：单个已归属命名空间变更归属
const props = defineProps({
  value: {
    type: Boolean,
    default: false,
  },
  clusterId: {
    type: String,
    default: '',
  },
  // 待分配的命名空间
  items: {
    type: Array as () => INativeNamespace[],
    default: () => [],
  },
  // 可见范围内的项目（不含集群所属项目）
  projects: {
    type: Array as () => IProjectOption[],
    default: () => [],
  },
  // 项目编码 -> 名称（回显当前归属用）
  projectMap: {
    type: Object as () => Record<string, string>,
    default: () => ({}),
  },
  mode: {
    type: String,
    default: 'assign',
  },
});

const emits = defineEmits(['input', 'refresh', 'go-settings']);

const isTransfer = computed(() => props.mode === 'transfer');
const projectName = (code: string) => props.projectMap[code] || code;

const target = ref('');
const cpu = ref(8);
const mem = ref(16);
const errs = ref({ target: '', quota: '' });
const submitting = ref(false);
const results = ref<{ name: string, ok: boolean, msg: string }[]>([]);
// 单个命名空间带入的原始配额（用户修改过才传 overwrite = true 覆盖已有配额）
let originCpu = 8;
let originMem = 16;

// 打开时初始化：单个且已有配额的命名空间自动带入现有值
watch(() => props.value, (value) => {
  if (!value) return;
  const quota = props.items.length === 1 ? props.items[0]?.quota : null;
  originCpu = quota ? (parseCpu(quota.cpuLimits) || 8) : 8;
  originMem = quota ? (parseMemoryGiB(quota.memoryLimits) || 16) : 16;
  cpu.value = originCpu;
  mem.value = originMem;
  target.value = '';
  errs.value = { target: '', quota: '' };
  results.value = [];
  submitting.value = false;
});

const mixedCount = computed(() => props.items.filter(ns => !!ns.projectCode).length);
const okCount = computed(() => results.value.filter(result => result.ok).length);
const failCount = computed(() => results.value.length - okCount.value);
const targetName = computed(() => projectName(target.value));

const handleTargetChange = (value: string) => {
  target.value = value;
  errs.value.target = '';
};

const buildQuota = (): IResourceQuota => ({
  cpuLimits: String(cpu.value),
  cpuRequests: String(cpu.value),
  memoryLimits: `${mem.value}Gi`,
  memoryRequests: `${mem.value}Gi`,
});

// 逐个提交（并行请求、逐个收集结果，部分失败不回滚已成功项）
const runAllocate = async (names: string[]) => Promise.all(names.map(async (name) => {
  const quota = buildQuota();
  // 批量统一配额不传 overwrite（默认 false，quota 仅对无配额的命名空间生效）；
  // 单个且用户修改了带入配额时传 true 覆盖已有配额
  const overwrite = props.items.length === 1 && !!props.items[0]?.quota
    && (cpu.value !== originCpu || mem.value !== originMem)
    ? true
    : undefined;
  const err = await allocateNamespace({
    $clusterId: props.clusterId,
    $namespace: name,
    targetProjectCode: target.value,
    quota,
    overwrite,
  }).then(() => '')
    .catch((error): string => (
      error?.message || window.i18n.t('cluster.sharedManage.assign.slider.failed')
    ));
  return { name, ok: !err, msg: err };
}));

const handleSubmit = () => {
  errs.value.target = target.value ? '' : window.i18n.t('cluster.sharedManage.assign.slider.errTarget');
  errs.value.quota = Number(cpu.value) >= 1 && Number(mem.value) >= 1
    ? ''
    : window.i18n.t('cluster.sharedManage.assign.slider.errQuota');
  if (errs.value.target || errs.value.quota) return;
  submitting.value = true;
  runAllocate(props.items.map(ns => ns.name)).then((list) => {
    submitting.value = false;
    results.value = list;
  });
};

// 只重试失败项（按 name 匹配，保留已成功结果）
const handleRetry = () => {
  const failedNames = results.value.filter(result => !result.ok).map(result => result.name);
  if (!failedNames.length) return;
  submitting.value = true;
  runAllocate(failedNames).then((list) => {
    submitting.value = false;
    results.value = results.value.map(result => (
      result.ok ? result : (list.find(item => item.name === result.name) || result)
    ));
  });
};

const handleFinish = () => {
  emits('input', false);
  if (okCount.value) {
    $bkMessage({
      theme: 'success',
      message: window.i18n.t('cluster.sharedManage.assign.msg.assignSuccess', {
        num: okCount.value,
        name: targetName.value,
      }),
    });
    emits('refresh');
  }
};

// 提交中禁止关闭（quick-close / esc）
const handleBeforeClose = () => !submitting.value;

const handleClose = (value: boolean) => {
  if (submitting.value) return;
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
