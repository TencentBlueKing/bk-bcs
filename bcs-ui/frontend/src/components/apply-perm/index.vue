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
            <table class="permission-table table-header">
                <thead>
                    <tr>
                        <th width="20%">{{ $t('系统') }}</th>
                        <th width="30%">{{ $t('需要申请的权限') }}</th>
                        <th width="50%">{{ $t('关联的资源实例') }}</th>
                    </tr>
                </thead>
            </table>
            <div class="table-content">
                <table class="permission-table">
                    <tbody>
                        <template v-if="actionList.length > 0">
                            <tr v-for="(action, index) in actionList" :key="index">
                                <td width="20%">{{ $t('容器管理平台') }}</td>
                                <td width="30%">{{ actionsMap[action.action_id] || '--' }}</td>
                                <td width="50%">{{ action.resource_name || '--' }}</td>
                            </tr>
                        </template>
                        <tr v-else>
                            <td class="no-data" colspan="3">{{ $t('无数据') }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
        <div class="permission-footer" slot="footer">
            <div class="button-group">
                <bk-button theme="primary" :disabled="!applyUrl" @click="goApplyUrl">{{ $t('去申请') }}</bk-button>
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
                actionList: [],
                lockSvg,
                actionsMap
            }
        },
        destroyed () {
            this.applyUrl = ''
        },
        methods: {
            hide () {
                this.dialogConf.isShow = false
                this.applyUrl = ''
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
  .permission-table {
    width: 100%;
    color: #63656e;
    border-bottom: 1px solid #e7e8ed;
    border-collapse: collapse;
    table-layout: fixed;
    th,
    td {
      padding: 12px 18px;
      font-size: 12px;
      text-align: left;
      border-bottom: 1px solid #e7e8ed;
      word-break: break-all;
    }
    th {
      color: #313238;
      background: #f5f6fa;
    }
  }
  .table-content {
    max-height: 260px;
    border-bottom: 1px solid #e7e8ed;
    border-top: 0;
    overflow: auto;
    .permission-table {
      border-top: 0;
      border-bottom: 0;
      td:last-child {
        border-right: 0;
      }
      tr:last-child td {
        border-bottom: 0;
      }
      .resource-type-item {
        padding: 0;
        margin: 0;
      }
    }
    .no-data {
      // padding: 30px;
      text-align: center;
      color: #999;
    }
  }
}
.button-group {
  .bk-button {
    margin-left: 7px;
  }
}
</style>
