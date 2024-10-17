<template>
  <div>
    <div class="bk-button-group" v-if="showHead">
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
            v-model="loginConfig.initLoginPassword"
            @blur="handlePasswordBlur">
          </bcs-input>
        </bk-form-item>
        <bk-form-item :label="$t('tke.label.confirmPassword')" class="!mt-[16px]" required>
          <bcs-input
            type="password"
            autocomplete="new-password"
            v-model="confirm"
            @blur="handleConfirmPasswordBlur">
          </bcs-input>
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
              :id="item[valueKey]"
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
          <bk-input type="textarea" v-model="loginConfig.keyPair.keySecret" @blur="handleKeySecretBlur"></bk-input>
        </bk-form-item>
      </template>
    </div>
  </div>
</template>
<script setup lang="ts">
import { onBeforeMount, PropType, ref, watch } from 'vue';

import { cloudKeyPairs } from '@/api/modules/cluster-manager';
import SelectExtension from '@/components/select-extension.vue';
import $store from '@/store';
import { IKeyItem } from '@/views/cluster-manage/types/types';;

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
  type: {
    type: String as PropType<'password'|'ssh'>,
    default: 'password',
  },
  initData: {
    type: Boolean,
    default: false,
  },
  showHead: {
    type: Boolean,
    default: true,
  },
  valueKey: {
    type: String,
    default: 'KeyID',
  },
});

const emits = defineEmits([
  'change',
  'type-change',
  'pass-change',
  'pass-blur',
  'confirm-pass-blur',
  'key-secret-blur',
]);
// 登录方式
const loginType = ref<'password'|'ssh'>(props.type);
watch(() => props.type, () => {
  loginType.value = props.type;
});

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
const keyPairs = ref<Array<IKeyItem>>($store.state.cloudMetadata.keyPairsList);
const cloudKeyPairsLoading = ref(false);
const handleGetCloudKeyPairs = async () => {
  if (!props.region || !props.cloudAccountID || !props.cloudID) return;
  cloudKeyPairsLoading.value = true;
  keyPairs.value = await cloudKeyPairs({
    $cloudId: props.cloudID,
    accountID: props.cloudAccountID,
    region: props.region,
  }).catch(() => []);
  $store.commit('cloudMetadata/updateKeyPairsList',  keyPairs.value);
  cloudKeyPairsLoading.value = false;
};

const handlePasswordBlur = (v) => {
  emits('pass-blur', v);
};
const handleConfirmPasswordBlur = (v) => {
  emits('confirm-pass-blur', v);
};
const handleKeySecretBlur = (v) => {
  emits('key-secret-blur', v);
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
);

onBeforeMount(() => {
  props.initData && handleGetCloudKeyPairs();
});
</script>
