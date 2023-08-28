<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="bk-keyer">
    <div class="biz-keys-list mb10">
      <div class="biz-key-item" v-for="(keyItem, index) in list" :key="index">
        <template v-if="varList.length">
          <bkbcs-input
            type="text"
            :placeholder="keyPlaceholder || $t('generic.label.key')"
            :style="{ width: `${keyInputWidth}px` }"
            :value.sync="keyItem.key"
            :list="varList"
            :disabled="!!keyItem.disabled"
            @input="valueChange"
            @blur="handleBlur"
            @paste="pasteKey(keyItem, $event)">
          </bkbcs-input>
        </template>
        <template v-else>
          <input
            type="text"
            class="bk-form-input"
            :placeholder="keyPlaceholder || $t('generic.label.key')"
            :style="{ width: `${keyInputWidth}px` }"
            v-model="keyItem.key"
            :disabled="!!keyItem.disabled"
            @paste="pasteKey(keyItem, $event)"
            @input="valueChange"
            @blur="handleBlur"
          />
        </template>

        <span class="operator">=</span>

        <template v-if="varList.length">
          <bkbcs-input
            :type="valueType"
            :placeholder="valuePlaceholder || $t('generic.label.value')"
            :style="{ width: `${valueInputWidth}px` }"
            :value.sync="keyItem.value"
            :list="varList"
            :disabled="keyItem.disabled"
            @input="valueChange"
            @blur="handleBlur"
          >
          </bkbcs-input>
        </template>
        <template v-else>
          <input
            :type="valueType"
            class="bk-form-input"
            :placeholder="valuePlaceholder || $t('generic.label.value')"
            :style="{ width: `${valueInputWidth}px` }"
            :disabled="keyItem.disabled"
            v-model="keyItem.value"
            @input="valueChange"
            @blur="handleBlur"
          />
        </template>

        <bk-button class="action-btn" @click.stop.prevent="addKey" style="min-width: 20px;">
          <i class="bcs-icon bcs-icon-plus"></i>
        </bk-button>
        <bk-button
          class="action-btn"
          v-if="list.length > 1"
          :disabled="keyItem.disabled"
          @click.stop.prevent="removeKey(keyItem, index)"
          style="min-width: 20px; background-color: #FFFFFF;">
          <i class="bcs-icon bcs-icon-minus"></i>
        </bk-button>
        <bk-checkbox
          v-if="isLinkToSelector"
          v-model="keyItem.isSelector"
          style="margin-left: 20px;"
          @change="valueChange">
          {{addToSelectorStr || $t('generic.keyer.actions.add')}}
        </bk-checkbox>
        <div v-if="keyItem.linkMessage" class="biz-tip mt5 f12" style="line-height: 1;">{{keyItem.linkMessage}}</div>
      </div>
    </div>
    <slot>
      <p
        style="line-height: 1;"
        :class="['biz-tip', { 'is-danger': isTipChange }]">{{tip ? tip : $t('generic.keyer.msg.info')}}</p>
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
    isTipChange: {
      type: Boolean,
      default: false,
    },
    isLinkToSelector: {
      type: Boolean,
      default: false,
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
    addToSelectorStr: {
      type: String,
      default: '',
    },
    dataKey: {
      type: String,
      default: '',
    },
    keyInputWidth: {
      type: Number,
      default: 240,
    },
    valueInputWidth: {
      type: Number,
      default: 240,
    },
    useKeyTrim: {
      type: Boolean,
      default: true,
    },
    useValueTrim: {
      type: Boolean,
      default: true,
    },
    valueType: {
      type: String,
      default: 'text',
    },
  },
  data() {
    return {
      list: this.keyList,
    };
  },
  watch: {
    'keyList'() {
      if (this.keyList?.length) {
        this.list = this.keyList;
      } else {
        this.list = [{
          key: '',
          value: '',
        }];
      }
    },
  },
  methods: {
    addKey() {
      const params = {
        key: this.dataKey || '',
        value: '',
      };
      if (this.isLinkToSelector) {
        params.isSelector = false;
      }
      this.list.push(params);
      const obj = this.getKeyObject(true);
      this.$emit('change', this.list, obj);
    },
    removeKey(item, index) {
      this.list.splice(index, 1);
      const obj = this.getKeyObject(true);
      this.$emit('change', this.list, obj);
    },
    valueChange() {
      this.$nextTick(() => {
        const obj = this.getKeyObject(true);
        this.$emit('change', this.list, obj);
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
            obj.value = item.value.trim();
          }
          return obj;
        });
        this.$emit('change', list, obj);
      });
    },
    pasteKey(item, event) {
      const cache = item.key;
      const clipboard = event.clipboardData;
      const text = clipboard.getData('Text');

      if (text && text.indexOf('=') > -1) {
        this.paste(event);
        item.key = cache;
        setTimeout(() => {
          item.key = cache;
        }, 0);
      }
    },
    paste(event) {
      const clipboard = event.clipboardData;
      const text = clipboard.getData('Text');
      const items = text.split('\n');
      items.forEach((item) => {
        if (item.indexOf('=') > -1) {
          const arr = item.split('=');
          this.list.push({
            key: this.useKeyTrim ? arr[0].trim() : arr[0],
            value: this.useValueTrim ? arr[1].trim() : arr[1],
          });
        }
      });
      setTimeout(() => {
        this.formatData();
      }, 10);

      return false;
    },
    formatData() {
      // 去掉空值
      if (this.list.length) {
        const results = [];
        const keyObj = {};
        const { length } = this.list;
        this.list.forEach((item) => {
          if (item.key || item.value) {
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
          obj.value = item.value.trim();
        }
        return obj;
      });
      if (isAll) {
        return list;
      }
      return list.filter(item => item.key && item.value);
    },
    getKeyObject(isAll) {
      const results = this.getKeyList(isAll);
      if (results.length === 0) {
        return {};
      }
      const obj = {};
      results.forEach((item) => {
        if (isAll) {
          obj[item.key] = item.value;
        } else if (item.key && item.value) {
          obj[item.key] = item.value;
        }
      });
      return obj;
    },
  },
};
</script>

<style scoped lang="postcss">
    @import '@/css/variable.css';
    input[type="number"] {
        &::-webkit-outer-spin-button,&::-webkit-inner-spin-button {
            appearance: none;
        }
        -moz-appearance: textfield;
    }

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
    .is-danger {
        color: $dangerColor;
    }
</style>
