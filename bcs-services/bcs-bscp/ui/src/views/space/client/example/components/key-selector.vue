<template>
  <bk-select
    v-model="currentValue.key"
    ref="selectorRef"
    :class="['key-selector', { 'select-error': isError }]"
    :popover-options="{ theme: 'light bk-select-popover' }"
    :popover-min-width="360"
    :filterable="true"
    :input-search="false"
    :clearable="false"
    :loading="loading"
    :search-placeholder="$t('搜索名称/密钥/说明')"
    :no-data-text="$t('暂无可用密钥，可前往密钥管理新建/启用密钥，或将已有密钥关联至此服务')"
    :no-match-text="$t('搜索结果为空，可前往密钥管理新建/启用密钥，或将已有密钥关联至此服务')"
    @change="handleSelectChange">
    <template #trigger>
      <div class="selector-trigger">
        <bk-overflow-title v-if="currentValue.privacyCredential && currentValue.name" class="app-name" type="tips">
          {{ currentValue.name }}({{ currentValue.privacyCredential }})
        </bk-overflow-title>
        <span v-else class="no-app">{{ $t('请选择') }}</span>
        <AngleUpFill class="arrow-icon arrow-fill" />
      </div>
    </template>
    <bk-option
      v-for="item in credentialList"
      :key="item.id"
      :value="item.spec"
      :label="item.spec.name + item.spec.enc_credential + item.spec.memo">
      <div class="key-option-item">
        <div class="name-text">
          {{ item.spec.name }}({{ item.spec.privacyCredential }})&nbsp;
          <span class="name-text--desc">{{ item.spec.memo || '--' }}</span>
        </div>
      </div>
    </bk-option>
    <template #extension>
      <div class="selector-extensition">
        <div class="content" @click="linkTo">
          <i class="bk-bscp-icon icon-app-store app-icon"></i>
          {{ $t('密钥管理') }}
        </div>
        <div class="flush-data">
          <right-turn-line class="flush-data-icon" @click="flushData" />
        </div>
      </div>
    </template>
  </bk-select>
</template>

<script lang="ts" setup>
  import { ref, Ref, onMounted, inject } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { newICredentialItem } from '../../../../../../types/client';
  import { getCredentialList } from '../../../../../api/credentials';
  import { AngleUpFill, RightTurnLine } from 'bkui-vue/lib/icon';
  import { debounce } from 'lodash';

  const props = defineProps<{
    selectedKeyData: newICredentialItem['spec'] | null;
  }>();

  const emits = defineEmits(['current-key', 'selected-key-data']);

  const route = useRoute();
  const router = useRouter();

  const basicInfo = inject<{ serviceName: Ref<string>; serviceType: Ref<string> }>('basicInfo');
  const isError = ref(false);
  const loading = ref(true);
  const currentValue = ref({
    name: '',
    key: '',
    privacyCredential: '', // 脱敏密钥
  });
  const bizId = ref(String(route.params.spaceId));
  const credentialList = ref<newICredentialItem[]>([]);

  onMounted(async () => {
    await loadCredentialList();
    // 当前服务下其他示例已选择密钥时，载入选择的密钥
    if (props.selectedKeyData !== null) {
      handleSelectChange(props.selectedKeyData);
    }
  });

  // 表单校验失败检查密钥是否为空
  const validateCredential = () => {
    isError.value = !currentValue.value.privacyCredential;
    return !isError.value;
  };
  // 获取密钥列表
  const loadCredentialList = async () => {
    loading.value = true;
    try {
      const query = {
        start: 0,
        all: true,
      };
      const res = await getCredentialList(bizId.value, query);
      const filterCurServiceData = filterCurService(res.details);
      credentialList.value = dataMasking(filterCurServiceData);
    } catch (e) {
      console.error(e);
    } finally {
      loading.value = false;
    }
  };
  // 已启用且和当前服务有关联规则的密钥
  const filterCurService = (data: Array<any>) => {
    return data.filter((item) => {
      const splitStr = item.credential_scopes.map((str: string) => str.split('/')[0]);
      return splitStr.includes(basicInfo!.serviceName.value) && item.spec.enable;
    });
  };
  // 下拉列表操作
  const handleSelectChange = (val: newICredentialItem['spec']) => {
    const { name, enc_credential, privacyCredential } = val;
    currentValue.value.name = name;
    currentValue.value.key = enc_credential;
    currentValue.value.privacyCredential = privacyCredential;
    emits('current-key', enc_credential, privacyCredential);
    emits('selected-key-data', val);
    validateCredential();
  };
  // 密钥脱敏
  const dataMasking = (data: Array<newICredentialItem>) => {
    return data.map((item) => {
      const { enc_credential } = item.spec;
      const newItem = {
        ...item,
      };
      if (enc_credential.length > 6) {
        const newKey = `${enc_credential.substring(0, 3)}***${enc_credential.substring(enc_credential.length - 3)}`;
        newItem.spec.privacyCredential = newKey;
      } else {
        const newKey = `${enc_credential.substring(0, 1)}***${enc_credential.substring(enc_credential.length - 1)}`;
        newItem.spec.privacyCredential = newKey;
      }
      return newItem;
    });
  };
  const linkTo = () => {
    const routeData = router.resolve({ name: 'credentials-management' });
    window.open(routeData.href, '__blank');
  };
  const flushData = debounce(loadCredentialList, 300);
  defineExpose({
    validateCredential,
  });
</script>

<style scoped lang="scss">
  .key-selector {
    &.select-error .selector-trigger {
      border-color: #ea3636;
    }
    &.popover-show .selector-trigger {
      border-color: #3a84ff;
      box-shadow: 0 0 3px #a3c5fd;
      .arrow-icon {
        transform: rotate(-180deg);
      }
    }
    &.is-focus {
      .selector-trigger {
        outline: 0;
      }
    }
    .selector-trigger {
      padding: 0 10px 0;
      height: 32px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-radius: 2px;
      transition: all 0.3s;
      background: #ffffff;
      font-size: 14px;
      border: 1px solid #c4c6cc;
      .app-name {
        max-width: 220px;
        color: #313238;
      }
      .no-app {
        font-size: 12px;
        color: #c4c6cc;
      }
      .arrow-icon {
        margin-left: 13.5px;
        color: #979ba5;
        transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      }
    }
  }
  .key-option-item {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    width: 100%;
    .name-text {
      margin-right: 5px;
      flex: 1;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
  }
  .name-text--desc {
    font-size: 12px;
    line-height: 20px;
    color: #c4c6cc;
  }
  .selector-extensition {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    flex: 1;
    .content {
      flex: 1;
      height: 39px;
      line-height: 39px;
      text-align: center;
      background: #fafbfd;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
    .app-icon {
      font-size: 14px;
    }
    .flush-data {
      position: relative;
      display: flex;
      justify-content: center;
      align-items: center;
      width: 39px;
      height: 39px;
      flex-shrink: 0;
      text-align: center;
      cursor: pointer;
      &::after {
        content: '';
        position: absolute;
        left: 0;
        top: 50%;
        border-left: 1px solid #dcdee5;
        height: 16px;
        transform: translateY(-50%);
      }
      &:hover {
        color: #3a84ff;
      }
      &:active {
        opacity: 0.6;
      }
      &-icon {
        font-size: 14px;
      }
    }
  }
</style>
