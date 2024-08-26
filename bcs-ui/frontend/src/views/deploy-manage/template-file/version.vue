<template>
  <bcs-dialog :value="value" width="480" @cancel="cancel">
    <template #header>
      <div class="flex items-center mt-[-14px]">
        <span class="text-[#313238] text-[20px] leading-[28px]">{{ title }}</span>
        <bcs-divider direction="vertical"></bcs-divider>
        <span class="text-[14px] leading-[20px]">{{ subTitle }}</span>
      </div>
    </template>
    <bk-form :key="formKey" :model="formData" :rules="rules" form-type="vertical" ref="versionFormRef">
      <!-- 版本类型 -->
      <!-- <bk-form-item :label="$t('templateFile.label.versionType')" required>
        <bk-radio-group class="flex flex-col gap-[10px]" v-model="versionType">
          <bk-radio value="major">{{ $t('templateFile.label.semver.major') }}</bk-radio>
          <bk-radio value="minor" class="!ml-[0px]">{{ $t('templateFile.label.semver.minor') }}</bk-radio>
          <bk-radio value="patch" class="!ml-[0px]">{{ $t('templateFile.label.semver.patch') }}</bk-radio>
        </bk-radio-group>
      </bk-form-item> -->
      <!-- 版本号 -->
      <bk-form-item :label="$t('templateFile.label.semverNo')">
        <bcs-checkbox
          class="flex items-center px-[8px] absolute top-[-30px] left-[60px] h-[28px] bg-[#F5F7FA]"
          v-model="isPreRelease">
          {{ $t('templateFile.label.semver.preRelease') }}
        </bcs-checkbox>
        <div class="text-[#979BA5] leading-[20px] mb-[6px] pt-[6px]">
          (<span v-for="item, index in semverList" :key="item.id">
            {{ item.name}}
            <template v-if="index < (semverList.length - 1)">
              {{ `${item.id === 'patch' ? '-' : '.'}` }}
            </template>
          </span>)
        </div>
        <div class="flex items-center gap-[5px]">
          <span v-for="item, index in semverList" :key="item.id">
            <bcs-input
              type="number"
              class="w-[48px]"
              :show-controls="false"
              :min="0"
              :precision="0"
              v-model="semverData[item.id]"
              v-if="item.id !== 'preRelease'" />
            <template v-else>
              <bcs-input
                class="w-[120px]"
                v-model="semverData.preReleaseTag" />
              <span class="px-[2px]">.</span>
              <bcs-input
                type="number"
                class="w-[48px]"
                :show-controls="false"
                :min="1"
                :precision="0"
                v-model="semverData[item.id]" />
            </template>
            <span class="ml-[2px]" v-if="index < (semverList.length - 1)">
              {{ `${item.id === 'patch' ? '-' : '.'}` }}
            </span>
          </span>
        </div>
      </bk-form-item>
      <!-- 版本日志 -->
      <bk-form-item
        :label="$t('templateFile.label.semverDesc')"
        error-display-type="normal"
        property="versionDescription"
        required>
        <bcs-input type="textarea" :maxlength="256" v-model="formData.versionDescription"></bcs-input>
      </bk-form-item>
    </bk-form>
    <template #footer>
      <bcs-button :loading="loading" theme="primary" @click="confirm">{{ $t('generic.button.create') }}</bcs-button>
      <bcs-button @click="cancel">{{ $t('generic.button.cancel') }}</bcs-button>
    </template>
  </bcs-dialog>
</template>
<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import $i18n from '@/i18n/i18n-setup';

const props = defineProps({
  value: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '',
  },
  subTitle: {
    type: String,
    default: '',
  },
  // 当前版本
  version: {
    type: String,
    default: '',
  },
  loading: {
    type: Boolean,
    default: false,
  },
  // 是否自动更新当前版本号
  autoUpdate: {
    type: Boolean,
    default: false,
  },
});

const emits = defineEmits(['cancel', 'confirm']);

const formKey = ref(0);
const versionFormRef = ref();
const versionType = ref('major');
const formData = ref({
  versionDescription: '',
  version: '',
});
const rules = ref({
  versionDescription: [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      required: true,
    },
  ],
});
const semverData = ref({
  major: 1,
  minor: 0,
  patch: 0,
  preReleaseTag: 'alpha',
  preRelease: 1,
});
const isPreRelease = ref(false);
const semverList = computed(() => {
  const data = [
    {
      id: 'major',
      name: $i18n.t('templateFile.label.semver.major'),
    },
    {
      id: 'minor',
      name: $i18n.t('templateFile.label.semver.minor'),
    },
    {
      id: 'patch',
      name: $i18n.t('templateFile.label.semver.patch'),
    },
  ];

  if (isPreRelease.value) {
    data.push({
      id: 'preRelease',
      name: $i18n.t('templateFile.label.semver.preRelease'),
    });
  }

  return data;
});

function parseSemverVersion(version: string) {
  const regex = /^(\d+)\.(\d+)\.(\d+)(?:-([A-Za-z]+)\.?([0-9A-Za-z-.]*))?$/;
  const match = version.match(regex);

  if (!match) return {
    major: 1,
    minor: 0,
    patch: 0,
    preReleaseTag: 'alpha',
    preRelease: 1,
  };

  return {
    major: parseInt(match[1], 10),
    minor: parseInt(match[2], 10),
    patch: parseInt(match[3], 10),
    preReleaseTag: match[4] || 'alpha',
    preRelease: parseInt(match[5], 10) || 1,
  };
}

function cancel() {
  emits('cancel');
}

async function confirm() {
  const result = await versionFormRef.value?.validate().catch(() => false);
  if (!result) return;

  const { major, minor, patch, preRelease, preReleaseTag } = semverData.value;
  formData.value.version = isPreRelease.value
    ? `${major}.${minor}.${patch}-${preReleaseTag}.${preRelease}`
    : `${major}.${minor}.${patch}`;
  emits('confirm', formData.value);
}

watch(() => props.value, () => {
  if (!props.value) {
    // 重置校验状态
    formKey.value = new Date().getTime();
    return;
  };
  formData.value.version = props.version;
  formData.value.versionDescription = '';
  semverData.value = parseSemverVersion(props.version);
  if (props.autoUpdate && versionType.value) {
    semverData.value[versionType.value] += 1;
  }
}, { immediate: true });

watch(versionType, () => {
  if (!versionType.value || !props.autoUpdate) return;

  const data = parseSemverVersion(props.version);

  data[versionType.value] += 1;
  semverData.value = data;
});
</script>
