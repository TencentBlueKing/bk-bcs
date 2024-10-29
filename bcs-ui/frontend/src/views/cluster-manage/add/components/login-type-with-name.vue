<template>
  <bk-form :model="loginConfig" :rules="nodePoolConfigRules" ref="loginRef">
    <div class="bk-button-group">
      <bk-button
        :class="[{ 'is-selected': loginType === 'password' }]"
        @click="loginType = 'password'">
        {{ $t('tke.label.loginType.password') }}
      </bk-button>
      <bk-button
        :class="[{ 'is-selected': loginType === 'ssh' }]"
        @click="loginType = 'ssh'">
        {{ $t('tke.label.loginType.ssh') }}
      </bk-button>
    </div>
    <div class="bg-[#F5F7FA] mt-[16px] py-[16px] pr-[16px]">
      <template v-if="loginType === 'password'">
        <bk-form-item
          :label-width="100"
          :label="$t('azureCloud.label.loginUser')"
          property="loginConfig.launchTemplate.initLoginUsername"
          error-display-type="normal">
          <bcs-input v-model="loginConfig.initLoginUsername"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('azureCA.password.set')"
          property="loginConfig.launchTemplate.initLoginPassword"
          error-display-type="normal"
          :label-width="100">
          <bcs-input
            type="password"
            v-model="loginConfig.initLoginPassword" />
        </bk-form-item>
        <bk-form-item
          :label="$t('azureCA.password.confirm')"
          property="loginConfig.launchTemplate.confirmPassword"
          error-display-type="normal"
          :label-width="100">
          <bcs-input
            type="password"
            v-model="loginConfig.confirmPassword" />
        </bk-form-item>
      </template>
      <template v-else-if="loginType === 'ssh'">
        <bk-form-item
          :label-width="100"
          :label="$t('azureCloud.label.loginUser')"
          property="loginConfig.launchTemplate.initLoginUsername"
          error-display-type="normal">
          <bcs-input v-model="loginConfig.initLoginUsername"></bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.text')"
          :label-width="100"
          :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.desc')"
          property="loginConfig.launchTemplate.keyPair.keyPublic"
          error-display-type="normal">
          <bcs-input
            type="textarea"
            :rows="4"
            v-model="loginConfig.keyPair.keyPublic">
          </bcs-input>
        </bk-form-item>
        <bk-form-item
          :label="$t('cluster.ca.nodePool.create.loginType.ssh.label.privateKey.text')"
          :label-width="100"
          :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.privateKey.desc')"
          property="loginConfig.launchTemplate.keyPair.keySecret"
          error-display-type="normal">
          <bcs-input
            type="textarea"
            :rows="4"
            :placeholder="$t('generic.placeholder.input')"
            v-model="loginConfig.keyPair.keySecret" />
        </bk-form-item>
      </template>
    </div>
  </bk-form>
</template>
<script setup lang="ts">
import { PropType, ref, watch } from 'vue';

import $i18n from '@/i18n/i18n-setup';

const props = defineProps({
  value: {
    type: Object,
    default: () => ({}),
  },
  type: {
    type: String as PropType<'password'|'ssh'>,
    default: 'password',
  },
});

const emits = defineEmits([
  'change',
]);
// 登录方式
const loginType = ref<'password'|'ssh'>(props.type);
watch(() => props.type, () => {
  loginType.value = props.type;
});

const loginConfig = ref({
  initLoginUsername: '',
  initLoginPassword: '',
  confirmPassword: '',
  keyPair: {
    keySecret: '',
    keyPublic: '',
  },
});

const nodePoolConfigRules = ref({
  // 密码校验
  'loginConfig.launchTemplate.initLoginPassword': [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => loginConfig.value.initLoginPassword.length > 0,
    },
    {
      message: $i18n.t('cluster.ca.nodePool.create.validate.password'),
      trigger: 'blur',
      validator: () => /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[^]{8,30}$/.test(loginConfig.value.initLoginPassword),
    },
  ],
  'loginConfig.launchTemplate.confirmPassword': [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => loginConfig.value.confirmPassword.length > 0,
    },
    {
      message: $i18n.t('cluster.ca.nodePool.create.validate.passwordNotSame'),
      trigger: 'blur',
      validator: () => loginConfig.value.confirmPassword === loginConfig.value.initLoginPassword,
    },
  ],
  'loginConfig.launchTemplate.initLoginUsername': [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => loginConfig.value.initLoginUsername.length > 0,
    },
    {
      message: $i18n.t('generic.validate.notRootName'),
      trigger: 'blur',
      validator: () => loginConfig.value.initLoginUsername !== 'root',
    },
  ],
  // 密钥校验
  'loginConfig.launchTemplate.keyPair.keyPublic': [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => loginConfig.value.keyPair.keyPublic.length > 0,
    },
  ],
  'loginConfig.launchTemplate.keyPair.keySecret': [
    {
      message: $i18n.t('generic.validate.required'),
      trigger: 'blur',
      validator: () => loginConfig.value.keyPair.keySecret.length > 0,
    },
  ],
});

const loginRef = ref();
function validate() {
  return loginRef.value?.validate();
};

function clearError() {
  loginRef.value?.clearError();
}

watch(loginType, () => {
  if (loginType.value === 'password') {
    loginConfig.value.keyPair = {
      keySecret: '',
      keyPublic: '',
    };
  } else if (loginType.value === 'ssh') {
    loginConfig.value.initLoginPassword = '';
    loginConfig.value.confirmPassword = '';
  }
  loginRef.value?.clearError();
});

watch(() => props.value, (newValue, oldValue) => {
  if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;

  loginConfig.value = Object.assign({
    initLoginUsername: '',
    initLoginPassword: '',
    confirmPassword: '',
    keyPair: {
      initLoginUsername: '',
      keySecret: '',
      keyPublic: '',
    },
  }, props.value);
}, { deep: true, immediate: true });

watch(loginConfig, () => {
  emits('change', loginConfig.value);
}, { deep: true });


defineExpose({
  validate,
  clearError,
});
</script>
