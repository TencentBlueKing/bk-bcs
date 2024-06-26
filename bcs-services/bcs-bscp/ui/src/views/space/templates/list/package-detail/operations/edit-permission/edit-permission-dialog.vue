<template>
  <bk-dialog
    :is-show="props.show"
    :title="$t('批量修改权限')"
    :theme="'primary'"
    quick-close
    ext-cls="batch-edit-perm-dialog"
    :width="640">
    <div class="selected-tag">
      {{ `${t('已选')} ` }} <span class="count">{{ props.configsLength }}</span> {{ `${t('个配置项')}` }}
    </div>
    <bk-form form-type="vertical" class="user-settings">
      <bk-form-item :label="t('文件权限')">
        <div class="perm-input">
          <bk-popover
            ext-cls="privilege-tips-wrap"
            theme="light"
            trigger="manual"
            placement="top"
            :is-show="showPrivilegeErrorTips">
            <bk-input
              v-model="privilegeInputVal"
              type="number"
              :placeholder="t('保持不变')"
              :disabled="loading"
              @blur="handlePrivilegeInputBlur" />
            <template #content>
              <div>{{ t('只能输入三位 0~7 数字') }}</div>
              <div class="privilege-tips-btn-area">
                <bk-button text theme="primary" @click="showPrivilegeErrorTips = false">{{ t('我知道了') }}</bk-button>
              </div>
            </template>
          </bk-popover>
          <bk-popover ext-cls="privilege-select-popover" theme="light" trigger="click" placement="bottom">
            <div class="perm-panel-trigger">
              <i class="bk-bscp-icon icon-configuration-line"></i>
            </div>
            <template #content>
              <div class="privilege-select-panel">
                <div v-for="(item, index) in PRIVILEGE_GROUPS" class="group-item" :key="index" :label="item">
                  <div class="header">{{ item }}</div>
                  <div class="checkbox-area">
                    <bk-checkbox-group
                      class="group-checkboxs"
                      :model-value="privilegeGroupsValue[index]"
                      @change="handleSelectPrivilege(index, $event)">
                      <bk-checkbox size="small" :label="4" :disabled="loading">
                        {{ t('读') }}
                      </bk-checkbox>
                      <bk-checkbox size="small" :label="2" :disabled="loading">{{ t('写') }}</bk-checkbox>
                      <bk-checkbox size="small" :label="1" :disabled="loading">{{ t('执行') }}</bk-checkbox>
                    </bk-checkbox-group>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </div>
      </bk-form-item>
      <bk-form-item :label="t('用户')">
        <bk-input v-model="localVal.user" :disabled="loading" :placeholder="t('保持不变')"></bk-input>
      </bk-form-item>
      <bk-form-item :label="t('用户组')">
        <bk-input v-model="localVal.user_group" :disabled="loading" :placeholder="t('保持不变')"></bk-input>
      </bk-form-item>
    </bk-form>
    <template v-if="currentPkg && currentPkg !== 'no_specified'">
      <p class="tips">{{ t('以下服务配置的未命名版本中引用此套餐的内容也将更新') }}</p>
      <div class="service-table">
        <bk-loading style="min-height: 100px" :loading="loading">
          <bk-table :data="citedList" :max-height="maxTableHeight">
            <bk-table-column :label="t('所在模板套餐')" prop="template_set_name"></bk-table-column>
            <bk-table-column :label="t('使用此套餐的服务')">
              <template #default="{ row }">
                <div v-if="row.app_id" class="app-info" @click="goToConfigPageImport(row.app_id)">
                  <div v-overflow-title class="name-text">{{ row.app_name }}</div>
                  <LinkToApp class="link-icon" :id="row.app_id" />
                </div>
              </template>
            </bk-table-column>
          </bk-table>
        </bk-loading>
      </div>
    </template>
    <template #footer>
      <bk-button
        theme="primary"
        style="margin-right: 8px"
        :loading="loading"
        :disabled="loading"
        @click="emits('confirm', { permission: localVal, appIds: citeByAppIds })">
        {{ t('保存') }}
      </bk-button>
      <bk-button @click="emits('update:show', false)">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IPackagesCitedByApps, ITemplateConfigItem } from '../../../../../../../../types/template';
  import { useRouter } from 'vue-router';
  import { storeToRefs } from 'pinia';
  import { getUnNamedVersionAppsBoundByPackages, getPackagesByTemplateIds } from '../../../../../../../api/template';
  import useGlobalStore from '../../../../../../../store/global';
  import useTemplateStore from '../../../../../../../store/template';
  import LinkToApp from '../../../components/link-to-app.vue';

  const { spaceId } = storeToRefs(useGlobalStore());
  const { currentTemplateSpace, currentPkg } = storeToRefs(useTemplateStore());
  const { t } = useI18n();
  const router = useRouter();

  const props = defineProps<{
    show: boolean;
    configsLength: number;
    loading: boolean;
    configs?: ITemplateConfigItem[];
  }>();

  const emits = defineEmits(['update:show', 'confirm']);

  const PRIVILEGE_GROUPS = [t('属主（own）'), t('属组（group）'), t('其他人（other）')];
  const PRIVILEGE_VALUE_MAP = {
    0: [],
    1: [1],
    2: [2],
    3: [1, 2],
    4: [4],
    5: [1, 4],
    6: [2, 4],
    7: [1, 2, 4],
  };
  const privilegeInputVal = ref('');
  const showPrivilegeErrorTips = ref(false);
  const localVal = ref({
    privilege: '',
    user: '',
    user_group: '',
  });
  const citedList = ref<IPackagesCitedByApps[]>([]);
  const tableLoading = ref(false);
  const pkgsIds = ref<number[]>([]);
  const citeByAppIds = ref<number[]>([]);

  watch(
    () => props.show,
    (val) => {
      if (val) {
        localVal.value = {
          privilege: '',
          user: '',
          user_group: '',
        };
        privilegeInputVal.value = '';
        if (currentPkg.value && currentPkg.value !== 'no_specified') {
          getCitedData();
        }
      }
    },
  );

  // 将权限数字拆分成三个分组配置
  const privilegeGroupsValue = computed(() => {
    const data: { [index: string]: number[] } = { 0: [], 1: [], 2: [] };
    if (typeof localVal.value.privilege === 'string' && localVal.value.privilege.length > 0) {
      const valArr = localVal.value.privilege.split('').map((i) => parseInt(i, 10));
      valArr.forEach((item, index) => {
        data[index as keyof typeof data] = PRIVILEGE_VALUE_MAP[item as keyof typeof PRIVILEGE_VALUE_MAP];
      });
    }
    return data;
  });

  const maxTableHeight = computed(() => {
    const windowHeight = window.innerHeight;
    return windowHeight * 0.6 - 200;
  });

  // 权限输入框失焦后，校验输入是否合法，如不合法回退到上次输入
  const handlePrivilegeInputBlur = () => {
    const val = String(privilegeInputVal.value);
    if (/^[0-7]{3}$/.test(val) || val === '') {
      localVal.value.privilege = val;
      showPrivilegeErrorTips.value = false;
    } else {
      privilegeInputVal.value = String(localVal.value.privilege);
      showPrivilegeErrorTips.value = true;
    }
  };

  // 选择文件权限
  const handleSelectPrivilege = (index: number, val: number[]) => {
    const groupsValue = { ...privilegeGroupsValue.value };
    groupsValue[index] = val;
    const digits: number[] = [];
    for (let i = 0; i < 3; i++) {
      let sum = 0;
      if (groupsValue[i].length > 0) {
        sum = groupsValue[i].reduce((acc, crt) => acc + crt, 0);
      }
      digits.push(sum);
    }

    // 选择其他权限 自动选择own的读取权限
    if (digits[0] < 4 && digits.some((item) => item > 0)) {
      digits[0] += 4;
    }
    const newVal = digits.every((item) => item === 0) ? '' : digits.join('');
    privilegeInputVal.value = newVal;
    localVal.value.privilege = newVal;
  };

  // 配置项被套餐引用数据
  const loadCiteByPkgsCountList = async () => {
    const ids = props.configs!.map((item) => item.id);
    const res = await getPackagesByTemplateIds(spaceId.value, currentTemplateSpace.value, ids);
    res.details.forEach((item) =>
      item.forEach((template) => {
        if (pkgsIds.value?.includes(template.template_set_id)) return;
        pkgsIds.value?.push(template.template_set_id);
      }),
    );
  };

  const getCitedData = async () => {
    tableLoading.value = true;
    const params = {
      start: 0,
      all: true,
    };
    if (currentPkg.value === 'all') {
      await loadCiteByPkgsCountList();
    }
    const template_set_ids: number[] = currentPkg.value === 'all' ? pkgsIds.value : [currentPkg.value as number];
    const res = await getUnNamedVersionAppsBoundByPackages(
      spaceId.value,
      currentTemplateSpace.value,
      template_set_ids,
      params,
    );
    citedList.value = res.details;
    citeByAppIds.value = citedList.value.map((Item) => Item.app_id);
    tableLoading.value = false;
  };

  const goToConfigPageImport = (id: number) => {
    const { href } = router.resolve({
      name: 'service-config',
      params: { appId: id },
      query: { pkg_id: currentTemplateSpace.value },
    });
    window.open(href, '_blank');
  };
</script>

<style scoped lang="scss">
  .user-settings {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }
  .perm-input {
    display: flex;
    align-items: center;
    width: 172px;
    :deep(.bk-input) {
      width: 140px;
      border-right: none;
      border-top-right-radius: 0;
      border-bottom-right-radius: 0;
      .bk-input--number-control {
        display: none;
      }
    }
    .perm-panel-trigger {
      width: 32px;
      height: 32px;
      text-align: center;
      background: #fafcfe;
      color: #3a84ff;
      border: 1px solid #3a84ff;
      cursor: pointer;
      &.disabled {
        color: #dcdee5;
        border-color: #dcdee5;
        cursor: not-allowed;
      }
    }
  }
  .privilege-select-panel {
    display: flex;
    align-items: top;
    border: 1px solid #dcdee5;
    .group-item {
      .header {
        padding: 0 16px;
        height: 42px;
        line-height: 42px;
        color: #313238;
        font-size: 12px;
        background: #fafbfd;
        border-bottom: 1px solid #dcdee5;
      }
      &:not(:last-of-type) {
        .header,
        .checkbox-area {
          border-right: 1px solid #dcdee5;
        }
      }
    }
    .checkbox-area {
      padding: 10px 16px 12px;
      background: #ffffff;
      &:not(:last-child) {
        border-right: 1px solid #dcdee5;
      }
    }
    .group-checkboxs {
      font-size: 12px;
      .bk-checkbox ~ .bk-checkbox {
        margin-left: 16px;
      }
      :deep(.bk-checkbox-label) {
        font-size: 12px;
      }
    }
  }
  .selected-tag {
    display: inline-block;
    height: 32px;
    background: #f0f1f5;
    line-height: 32px;
    border-radius: 16px;
    padding: 0 12px;
    margin: 8px 0px 16px;
    .count {
      color: #3a84ff;
    }
  }
</style>

<style lang="scss">
  .batch-operation-button-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .operation-item {
      padding: 0 12px;
      min-width: 58px;
      height: 32px;
      line-height: 32px;
      color: #63656e;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        background: #f5f7fa;
      }
    }
  }
  .app-info {
    display: flex;
    align-items: center;
    overflow: hidden;
    cursor: pointer;
    .name-text {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .link-icon {
      flex-shrink: 0;
      margin-left: 10px;
    }
  }
</style>
