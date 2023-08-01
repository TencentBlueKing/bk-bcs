<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="bk-expression">
    <div class="biz-keys-list mb10" v-if="list.length">
      <div class="biz-key-item" v-for="(keyItem, index) in list" :key="index">
        <template v-if="varList.length">
          <bkbcs-input
            type="text"
            :placeholder="keyPlaceholder || $t('generic.label.key')"
            :style="{ width: `${keyInputWidth}px` }"
            :value.sync="keyItem.key"
            :list="varList"
            @input="valueChange"
            @blur="handleBlur">
          </bkbcs-input>
        </template>
        <template v-else>
          <input
            type="text"
            class="bk-form-input"
            :placeholder="keyPlaceholder || $t('generic.label.key')"
            :style="{ width: `${keyInputWidth}px` }"
            v-model="keyItem.key"
            @input="valueChange"
            @blur="handleBlur"
          />
        </template>

        <div class="operator">
          <bk-selector
            style="width: 132px;"
            :placeholder="$t('generic.placeholder.select')"
            :setting-key="'id'"
            :display-key="'name'"
            :selected.sync="keyItem.operator"
            :list="operatorList"
            @item-selected="valueChange">
          </bk-selector>
        </div>

        <template v-if="['In', 'NotIn'].includes(keyItem.operator)">
          <template v-if="varList.length">
            <bkbcs-input
              type="text"
              :placeholder="valuePlaceholder || $t('generic.label.value')"
              :style="{ width: `${valueInputWidth}px` }"
              :value.sync="keyItem.values"
              :list="varList"
              @input="valueChange"
              @blur="handleBlur"
            >
            </bkbcs-input>
          </template>
          <template v-else>
            <input
              type="text"
              class="bk-form-input"
              :placeholder="valuePlaceholder || $t('generic.label.value')"
              :style="{ width: `${valueInputWidth}px` }"
              v-model="keyItem.values"
              @input="valueChange"
              @blur="handleBlur"
            />
          </template>
        </template>
        <template v-else>
          <span class="holder" :style="{ width: `${valueInputWidth}px` }"></span>
        </template>

        <bk-button class="action-btn" @click.stop.prevent="addKey">
          <i class="bcs-icon bcs-icon-plus"></i>
        </bk-button>
        <bk-button class="action-btn" @click.stop.prevent="removeKey(keyItem, index)">
          <i class="bcs-icon bcs-icon-minus"></i>
        </bk-button>
      </div>
    </div>
    <div v-else class="expression-action">
      <bk-button class="bk-button bk-button-small" @click.stop.prevent="addKey">
        <span class="bcs-icon bcs-icon-plus f13 vm"></span>
        <span class="text ml0">{{$t('plugin.tools.addPattern')}}</span>
      </bk-button>
    </div>
    <slot>
      <p :class="['biz-tip']">{{tip}}</p>
    </slot>
  </div>
</template>

<script>
export default {
  props: {
    keyList: {
      type: Array,
      default: [],
    },
    tip: {
      type: String,
      default: '',
    },
    varList: {
      type: Array,
      default() {
        return [];
      },
    },
    keyPlaceholder: {
      type: String,
      default: '',
    },
    valuePlaceholder: {
      type: String,
      default: '',
    },
    dataKey: {
      type: String,
      default: '',
    },
    keyInputWidth: {
      type: Number,
      default: 205,
    },
    valueInputWidth: {
      type: Number,
      default: 205,
    },
    useKeyTrim: {
      type: Boolean,
      default: true,
    },
    useValueTrim: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      list: this.keyList,
      operatorList: [
        {
          id: 'In',
          name: 'In',
        },
        {
          id: 'NotIn',
          name: 'NotIn',
        },
        {
          id: 'Exists',
          name: 'Exists',
        },
        {
          id: 'DoesNotExist',
          name: 'DoesNotExist',
        },
      ],
    };
  },
  watch: {
    'keyList'() {
      if (this.keyList?.length) {
        this.list = this.keyList;
      }
    },
  },
  methods: {
    addKey() {
      const params = {
        key: this.dataKey || '',
        operator: 'In',
        values: '',
      };
      this.list.push(params);
      const obj = this.getKeyObject(true);
      const list = this.getKeyList();
      this.$emit('change', list, obj);
    },
    removeKey(item, index) {
      this.list.splice(index, 1);
      const obj = this.getKeyObject(true);
      const list = this.getKeyList();
      this.$emit('change', list, obj);
    },
    valueChange() {
      this.$nextTick(() => {
        const obj = this.getKeyObject(true);
        const list = this.getKeyList();
        this.$emit('change', list, obj);
      });
    },
    handleBlur() {
      this.$nextTick(() => {
        const obj = this.getKeyObject(true);
        const list = this.list.map((item) => {
          const obj = { ...item };
          if (this.useKeyTrim) {
            obj.key = item.key.trim();
          }
          if (this.useValueTrim) {
            obj.values = item.values.trim();
          }
          return obj;
        });
        this.$emit('change', list, obj);
      });
    },
    formatData() {
      // 去掉空值
      if (this.list.length) {
        const results = [];
        const keyObj = {};
        const { length } = this.list;
        this.list.forEach((item) => {
          if (item.key || item.values) {
            if (!keyObj[item.key]) {
              results.push(item);
              keyObj[item.key] = true;
            }
          }
        });
        const patchLength = results.length - length;
        if (patchLength > 0) {
          for (let i = 0; i < patchLength; i++) {
            results.push({
              key: '',
              operator: 'In',
              value: '',
            });
          }
        }
        this.list.splice(0, this.list.length, ...results);
        this.$emit('change', this.list);
      }
    },
    getKeyList(isAll) {
      const list = this.list.map((item) => {
        const obj = { ...item };
        if (this.useKeyTrim) {
          obj.key = item.key.trim();
        }
        if (this.useValueTrim) {
          obj.values = item.values.trim();
        }
        if (['Exists', 'DoesNotExist'].includes(item.operator)) {
          delete obj.values;
        }
        return obj;
      });
      if (isAll) {
        return list;
      }
      return list.filter(item => item.key);
    },
    getKeyObject(isAll) {
      const results = this.getKeyList(isAll);
      if (results.length === 0) {
        return {};
      }
      const obj = {};
      results.forEach((item) => {
        if (isAll) {
          obj[item.key] = item.values;
        } else if (item.key && item.values) {
          obj[item.key] = item.values;
        }
      });
      return obj;
    },
  },
};
</script>

<style scoped lang="postcss">
    @import '@/css/variable.css';

    .biz-keys-list .action-btn {
        width: auto;
        padding: 0;
        margin-left: 5px;
        &.disabled {
            cursor: default;
            color: #ddd !important;
            border-color: #ddd !important;
            .bcs-icon {
                color: #ddd !important;
                border-color: #ddd !important;
            }
        }
        &:hover {
            color: $primaryColor;
            border-color: $primaryColor;
            .bcs-icon {
                color: $primaryColor;
                border-color: $primaryColor;
            }
        }
    }
    .holder {
        height: 36px;
        line-height: 1;
        color: #666;
        background-color: #fafafa;
        display: inline-block;
        border-radius: 2px;
        border: 1px solid #eee;
        vertical-align: middle;
    }
    .expression-action {
        padding: 25px;
        text-align: center;
        background: #FFF;
        border: 1px solid #DCDEE5;
        border-radius: 2px;
    }
</style>
