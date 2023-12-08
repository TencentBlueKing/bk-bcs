<template>
  <div>
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
    <div class="bg-[#F5F7FA] mt-[16px] max-w-[500px] py-[16px] pr-[16px]">
      <template v-if="loginType === 'password'">
        <bk-form-item :label="$t('tke.label.setPassword')" required>
          <bcs-input
            type="password"
            autocomplete="new-password"
            v-model="loginConfig.initLoginPassword">
          </bcs-input>
        </bk-form-item>
        <bk-form-item :label="$t('tke.label.confirmPassword')" class="!mt-[16px]" required>
          <bcs-input type="password" autocomplete="new-password" v-model="confirm"></bcs-input>
        </bk-form-item>
      </template>
      <template v-else-if="loginType === 'ssh'">
        <bk-form-item
          :label="$t('tke.label.publicKey')"
          :label-width="100"
          :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.publicKey.desc')"
          required>
          <bcs-select
            class="bg-[#fff]"
            :clearable="false"
            :loading="cloudKeyPairsLoading"
            searchable
            v-model="loginConfig.keyPair.keyID">
            <bcs-option
              v-for="item in keyPairs"
              :key="item.KeyID"
              :id="item.KeyID"
              :name="`${item.KeyName}(${item.KeyID})`">
            </bcs-option>
            <template #extension>
              <SelectExtension
                :link-text="$t('cluster.ca.nodePool.create.loginType.ssh.button.create')"
                link="https://console.cloud.tencent.com/cvm/sshkey"
                @refresh="handleGetCloudKeyPairs" />
            </template>
          </bcs-select>
        </bk-form-item>
        <bk-form-item
          :label="$t('tke.label.secretKey')"
          :label-width="100"
          :desc="$t('cluster.ca.nodePool.create.loginType.ssh.label.privateKey.desc')"
          class="!mt-[16px]"
          required>
          <bk-input type="textarea" v-model="loginConfig.keyPair.keySecret"></bk-input>
        </bk-form-item>
      </template>
    </div>
  </div>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue';

import { IKeyItem } from '../tencent/types';

import { cloudKeyPairs } from '@/api/modules/cluster-manager';
import SelectExtension from '@/views/cluster-manage/add/common/select-extension.vue';

const props = defineProps({
  region: {
    type: String,
    default: '',
  },
  cloudAccountID: {
    type: String,
    default: '',
  },
  cloudID: {
    type: String,
    default: '',
  },
  value: {
    type: Object,
    default: () => ({}),
  },
  confirmPass: {
    type: String,
    default: '',
  },
});

const emits = defineEmits(['change', 'type-change', 'pass-change']);
// 登录方式
const loginType = ref<'password'|'ssh'>('password');
const loginConfig = ref({
  initLoginUsername: '',
  initLoginPassword: '',
  keyPair: {
    keyID: '',
    keySecret: '',
    keyPublic: '',
  },
});
const confirm = ref(props.confirmPass);
watch(() => props.confirmPass, () => {
  confirm.value = props.confirmPass;
});

watch(confirm, () => {
  emits('pass-change', confirm.value);
});

watch(loginType, () => {
  if (loginType.value === 'password') {
    loginConfig.value.keyPair = {
      keyID: '',
      keySecret: '',
      keyPublic: '',
    };
  } else if (loginType.value === 'ssh') {
    loginConfig.value.initLoginPassword = '';
    confirm.value = '';
  }
  emits('type-change', loginType.value);
});

watch(() => props.value, (newValue, oldValue) => {
  if (JSON.stringify(newValue) === JSON.stringify(oldValue)) return;

  loginConfig.value = Object.assign({
    initLoginUsername: '',
    initLoginPassword: '',
    keyPair: {
      keyID: '',
      keySecret: '',
      keyPublic: '',
    },
  }, props.value);
}, { deep: true, immediate: true });

watch(loginConfig, () => {
  emits('change', loginConfig.value);
}, { deep: true });

// 密钥
const keyPairs = ref<Array<IKeyItem>>([]);
const cloudKeyPairsLoading = ref(false);
const handleGetCloudKeyPairs = async () => {
  if (!props.region || !props.cloudAccountID || !props.cloudID) return;
  cloudKeyPairsLoading.value = true;
  keyPairs.value = await cloudKeyPairs({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
  }).catch(() => []);
  cloudKeyPairsLoading.value = false;
};

watch(
  [
    () => props.region,
    () => props.cloudAccountID,
    () => props.cloudID,
  ],
  () => {
    handleGetCloudKeyPairs();
  },
  { immediate: true },
);
</script>
