<template>
    <bk-dialog
        :is-show.sync="dialogConf.isShow"
        :width="dialogConf.width"
        :quick-close="false"
        @cancel="hide"
        :title="dialogConf.title">
        <div class="permission-modal">
            <div class="permission-header">
                <span class="title-icon">
                    <img :src="lockSvg" alt="permission-lock" class="lock-img" />
                </span>
                <h3>{{ $t('该操作需要以下权限') }}</h3>
            </div>
            <div v-bkloading="{ isLoading }">
                <bk-table :data="actionList">
                    <bk-table-column :label="$t('系统')" prop="system" min-width="150">
                        <template>
                            {{ $t('容器管理平台') }}
                        </template>
                    </bk-table-column>
                    <bk-table-column :label="$t('需要申请的权限')" prop="auth" min-width="220">
                        <template slot-scope="{ row }">
                            {{ actionsMap[row.action_id] || '--' }}
                        </template>
                    </bk-table-column>
                    <bk-table-column :label="$t('关联的资源实例')" prop="resource" min-width="220">
                        <template slot-scope="{ row }">
                            {{ row.resource_name || '--' }}
                        </template>
                    </bk-table-column>
                </bk-table>
            </div>
        </div>
        <div class="permission-footer" slot="footer">
            <div class="button-group">
                <div v-bk-tooltips="{
                    content: $t('申请链接不存在'),
                    disabled: !!applyUrl
                }"
                >
                    <bk-button theme="primary" :disabled="!applyUrl" @click="goApplyUrl">{{ $t('去申请') }}</bk-button>
                </div>
                <bk-button theme="default" @click="hide">{{ $t('取消') }}</bk-button>
            </div>
        </div>
    </bk-dialog>
</template>

<script>
    /* eslint-disable camelcase */
    /* eslint-disable @typescript-eslint/camelcase */
    import lockSvg from '@/images/lock-radius.svg'
    import actionsMap from './actions-map'
    export default {
        name: 'apply-perm',
        data () {
            return {
                dialogConf: {
                    isShow: false,
                    width: 640
                },
                applyUrl: '',
                actionList: [{}],
                lockSvg,
                actionsMap,
                isLoading: false
            }
        },
        destroyed () {
            this.applyUrl = ''
        },
        methods: {
            hide () {
                this.isLoading = false
                this.dialogConf.isShow = false
                this.applyUrl = ''
                this.actionList = [{}]
            },
            show (data = {}) {
                const { apply_url, action_list = [] } = data?.perms

                this.applyUrl = apply_url
                this.actionList = action_list

                this.$nextTick(() => {
                    this.dialogConf.isShow = true
                })
            },
            goApplyUrl () {
                window.open(this.applyUrl)
                this.hide()
            }
        }
    }
</script>

<style lang="scss" scoped>
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
