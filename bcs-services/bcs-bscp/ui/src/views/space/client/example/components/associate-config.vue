<template>
  <div class="associate-config">
    <!-- 标签 -->
    <div class="associate-config-wrap">
      <span class="label-span">{{ $t('启用配置文件筛选') }}</span>
      <info
        class="icon-info"
        v-bk-tooltips="{
          content: $t(
            '当客户端无需拉取配置服务中的全量配置文件时，可以启用此功能，指定相应的通配符，可仅拉取客户端所需的文件',
          ),
          placement: 'top',
        }" />
      <bk-switcher class="label-switch" :value="configSwitch" size="small" theme="primary" @change="handleSwitcher" />
    </div>
    <div class="associate-config-content" v-if="configSwitch">
      <div class="associate-info-wrap">
        <div class="associate-info-count">{{ $t('已设置的筛选规则', { count: ruleList.length }) }}</div>
        <bk-button class="associate-info-btn" theme="primary" size="small" text @click="openRuleConfig">
          <cog-shape class="btn-icon" />{{ $t('规则设置') }}
        </bk-button>
      </div>
      <associate-side-bar
        :show="sideBarShow"
        :id="-1"
        :perm-check-loading="permCheckLoading"
        :has-manage-perm="hasManagePerm"
        :example-rules="rules"
        :is-example-mode="true"
        @send-example-rules="updateRule"
        @close="handleClose" />
    </div>
  </div>
</template>
<script lang="ts" setup>
  import { ref, onMounted, nextTick } from 'vue';
  import { useRoute } from 'vue-router';
  import { Info, CogShape } from 'bkui-vue/lib/icon';
  import associateSideBar from '../../../credentials/associate-config-items/index.vue';
  import { permissionCheck } from '../../../../../api/index';
  import { ICredentialRule, IRuleUpdateParams } from '../../../../../../types/credential';

  const emits = defineEmits(['updateRules']);

  const route = useRoute();

  const spaceId = ref(Number(route.params.spaceId));
  const configSwitch = ref(false);
  const sideBarShow = ref(false);
  const permCheckLoading = ref(false);
  const hasManagePerm = ref(false);
  const ruleList = ref<{ app: string; scope: string; id: number }[]>([]);
  const rules = ref<ICredentialRule[]>([]);

  onMounted(() => {
    getPermData();
  });

  const getPermData = async () => {
    permCheckLoading.value = true;
    const res = await permissionCheck({
      resources: [
        {
          biz_id: spaceId.value,
          basic: {
            type: 'credential',
            action: 'manage',
          },
        },
      ],
    });
    hasManagePerm.value = res.is_allowed;
    permCheckLoading.value = false;
  };

  // 获取筛选的规则
  const updateRule = (data: IRuleUpdateParams) => {
    const { add_scope, alter_scope, del_id } = data;
    if (add_scope.length) {
      add_scope.forEach((item) => {
        ruleList.value.push({
          app: item.app,
          scope: item.scope,
          id: Math.floor(Math.random() * 10000) + 1, // 配置示例没有规则id配置与返回
        });
      });
    }
    if (alter_scope.length) {
      alter_scope.forEach((alter) => {
        const scopeToUpdate = ruleList.value.find((item) => item.id === alter.id);
        if (scopeToUpdate) {
          scopeToUpdate.scope = alter.scope;
        }
      });
    }
    if (del_id.length) {
      const filteredVal = ruleList.value.filter((item) => !del_id.includes(item.id));
      ruleList.value = filteredVal;
    }
    sendRules();
    sideBarShow.value = false;
  };

  const openRuleConfig = () => {
    rules.value = ruleList.value.map((item) => {
      return {
        id: item.id,
        spec: {
          scope: item.scope,
          app: item.app,
        },
        attachment: {
          biz_id: spaceId.value,
          credential_id: -1,
        },
        revision: {
          creator: '',
          reviser: '',
          create_at: '',
          update_at: '',
          expired_at: '',
        },
      };
    });
    sideBarShow.value = true;
  };

  const handleClose = () => {
    sideBarShow.value = false;
  };

  const handleSwitcher = (val: boolean) => {
    configSwitch.value = val;
    if (!val) {
      ruleList.value = [];
      sendRules();
    } else {
      // 打开文件筛选开关自动打开规则设置抽屉
      nextTick(() => {
        sideBarShow.value = true;
      });
    }
  };

  const sendRules = () => {
    const ruleString = ruleList.value.map((item) => item.scope);
    emits('updateRules', ruleString);
  };
</script>

<style scoped lang="scss">
  .associate-config-wrap {
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }
  .label-span {
    font-size: 12px;
    line-height: 20px;
    color: #63656e;
    &.em {
      cursor: pointer;
      color: #3a84ff;
    }
    &.required {
      padding-right: 10px;
      position: relative;
      &::after {
        content: '*';
        position: absolute;
        right: 0;
        top: 50%;
        transform: translateY(-50%);
        font-size: 12px;
        color: #ea3636;
      }
    }
  }
  .label-switch {
    margin: 0 16px 0 12px;
  }
  .associate-config-content {
    margin-top: 24px;
  }
  .icon-info {
    margin-left: 9px;
    font-size: 14px;
    color: #979ba5;
    cursor: pointer;
  }
  .popover-wrap {
    font-size: 14px;
  }
  .popover-block {
    line-height: 22px;
    color: #63656e;
    &-gap {
      margin-top: 20px;
    }
  }
  .popover-btn {
    display: block;
    margin: 8px 0 0 auto;
  }
  .associate-info-wrap {
    padding: 6px 12px;
    display: inline-flex;
    justify-content: flex-start;
    align-items: center;
    font-size: 12px;
    line-height: 20px;
    background-color: #f5f7fa;
  }
  .associate-info-count {
    position: relative;
    padding-right: 12px;
    &::after {
      content: '';
      position: absolute;
      right: 0;
      top: 0;
      width: 1px;
      height: 100%;
      background-color: #dcdee5;
    }
  }
  .associate-info-btn {
    margin-left: 12px;
    .btn-icon {
      margin-right: 4px;
      font-size: 16px;
    }
  }
</style>
