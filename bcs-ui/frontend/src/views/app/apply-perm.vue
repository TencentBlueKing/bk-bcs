<template>
  <bk-dialog
    :is-show.sync="dialogConf.isShow"
    :width="dialogConf.width"
    :quick-close="false"
    :title="dialogConf.title"
    @cancel="hide">
    <div class="permission-modal">
      <div class="permission-header">
        <span class="title-icon">
          <img :src="lockSvg" alt="permission-lock" class="lock-img" />
        </span>
        <h3>{{ $t('iam.title.perms') }}</h3>
      </div>
      <div v-bkloading="{ isLoading }">
        <bk-table :data="actionList">
          <bk-table-column :label="$t('iam.label.system')" prop="system" min-width="150">
            {{ siteName }}
          </bk-table-column>
          <bk-table-column :label="$t('iam.label.action')" prop="auth" min-width="220">
            <template #default="{ row }">
              {{ actionsMap[row.action_id] || '--' }}
            </template>
          </bk-table-column>
          <bk-table-column :label="$t('iam.label.resource')" prop="resource" min-width="220">
            <template #default="{ row }">
              {{ row.resource_name || '--' }}
            </template>
          </bk-table-column>
        </bk-table>
      </div>
    </div>
    <div class="permission-footer" slot="footer">
      <div class="button-group">
        <div
          v-bk-tooltips="{
            content: $t('iam.tips.emptyApplyUrl'),
            disabled: !!applyUrl
          }"
        >
          <bk-button theme="primary" :disabled="!applyUrl" @click="goApplyUrl">{{ $t('iam.button.apply') }}</bk-button>
        </div>
        <bk-button theme="default" @click="hide">{{ $t('generic.button.cancel') }}</bk-button>
      </div>
    </div>
  </bk-dialog>
</template>

<script>
import actionsMap from './actions-map';

import lockSvg from '@/images/lock-radius.svg';
import usePlatform from '@/composables/use-platform';
export default {
  name: 'ApplyPerm',
  data() {
    return {
      dialogConf: {
        isShow: false,
        width: 640,
      },
      applyUrl: '',
      actionList: [{}],
      lockSvg,
      actionsMap,
      isLoading: false,
      config: null,
    };
  },
  created() {
    const { config } = usePlatform();
    this.config = config;
  },
  destroyed() {
    this.applyUrl = '';
  },
  computed: {
    siteName() {
      return this.config.i18n?.name;
    }
  },
  methods: {
    hide() {
      this.isLoading = false;
      this.dialogConf.isShow = false;
      this.applyUrl = '';
      this.actionList = [{}];
    },
    async show(callbackData = {}) {
      this.dialogConf.isShow = true;
      let data = {};
      if (typeof callbackData === 'function') {
        this.isLoading = true;
        data = await callbackData();
        this.isLoading = false;
      } else {
        data = callbackData;
      }
      const { apply_url, action_list = [] } = data?.perms;

      this.applyUrl = apply_url;
      this.actionList = action_list;
    },
    goApplyUrl() {
      window.open(this.applyUrl);
      this.hide();
    },
  },
};
</script>

<style lang="postcss" scoped>
.permission-modal {
  .permission-header {
    text-align: center;
    .title-icon {
      display: inline-block;
    }
    .lock-img {
      width: 120px;
    }
    h3 {
      margin: 6px 0 24px;
      color: #63656e;
      font-size: 20px;
      font-weight: normal;
      line-height: 1;
    }
  }
}
.button-group {
    display: flex;
    justify-content: flex-end;
    .bk-button {
        margin-left: 7px;
    }
}
</style>
