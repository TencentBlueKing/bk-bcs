
<template>
  <div class="bk-input-box bk-selector">
    <input
      type="text"
      :disabled="disabled"
      :placeholder="placeholder"
      class="bk-form-input"
      autocomplete="off"
      v-bk-focus="autoFocus"
      :value="curValue"
      @focus="focusHandler"
      @blur="blurHandler"
      @input="userInput"
      @keyup="keyup"
      @paste="paste"
      :maxlength="maxlength" />
    <transition :name="listSlideName">
      <div class="bk-selector-list" v-show="isListPanelShow && resultList.length" :style="panelStyle">
        <ul class="selector-list-box">
          <li
            v-for="(item, index) of resultList"
            :key="index" class="bk-selector-list-item selected" @click="confirmSelect($event, index)">
            <div :class="['bk-selector-node', { 'bk-selector-selected': item.isSelected }]">
              <span :title="item[displayKey]" class="text">{{item[displayKey]}}</span>
            </div>
          </li>
        </ul>
      </div>
    </transition>
  </div>
</template>
<script>
import { debounce } from 'lodash';

import { getActualTop } from '@/common/util';
export default {
  name: 'ComBox',
  props: {
    type: {
      type: String,
      default: 'text', // text || number
    },
    isDecimals: {
      type: Boolean,
      default: false,
    },
    percentEnable: {
      type: Boolean,
      default: false,
    },
    isSelectMode: {
      type: Boolean,
      default: false,
    },
    value: {
      type: [Number, String],
      default: '',
    },
    placeholder: {
      type: String,
      default: '',
    },
    useChinese: {
      type: Boolean,
      default: true,
    },
    autoFocus: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: [String, Boolean],
      default: false,
    },
    regexp: {
      type: Object,
    },
    maxlength: {
      type: [Number, String],
    },
    min: {
      type: [Number, String],
      default: Number.NEGATIVE_INFINITY,
    },
    max: {
      type: [Number, String],
      default: Number.POSITIVE_INFINITY,
    },
    isLink: {
      type: Boolean,
      default: false,
    },
    steps: {
      type: Number,
      default: 1,
    },
    size: {
      type: String,
      default: 'large',
      validator(value) {
        return [
          'large',
          'small',
        ].indexOf(value) > -1;
      },
    },
    debounceTimer: {
      type: Number,
      default: 500,
    },
    searchKey: {
      type: String,
      default: 'key',
    },
    settingKey: {
      type: String,
      default: 'key',
    },
    displayKey: {
      type: String,
      default: 'key',
    },
    list: {
      type: Array,
      default() {
        return [];
      },
    },
    defaultList: {
      type: Array,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      inputMode: 'input', // search || input
      isSearch: false,
      isListPanelShow: false,
      searchPrefix: '{{',
      searchSuffix: '}}',
      // eslint-disable-next-line no-useless-escape
      searchReg: /\{\{([^\{\}]+)?\}\}/,
      chineseReg: /[\u4e00-\u9fa5]/g,
      keyWord: '',
      curSelectIndex: -1,
      resultList: this.defaultList || this.list,
      timer: 0,
      isMax: false,
      isMin: false,
      curValue: '',
      localValue: '',
      isFocus: false,
      maxNumber: this.max,
      minNumber: this.min,
      panelStyle: {},
      listSlideName: 'toggle-slide',
    };
  },
  watch: {
    min() {
      this.minNumber = Number(this.min);
    },
    max() {
      this.maxNumber = Number(this.max);
    },
    value: {
      immediate: true,
      handler(value) {
        if (value !== this.localValue || value !== this.curValue) {
          this.changeCurValue(this.isLink);
        }
      },
    },
    defaultList: {
      immediate: true,
      handler() {
        this.changeCurValue(this.isLink);
      },
    },
    list: {
      immediate: true,
      handler() {
        this.changeCurValue(this.isLink);
      },
    },
    keyWord(val) {
      const { searchKey } = this;
      let sourceList = [];
      let targetList = [];
      let keyWord = val;
      if (this.inputMode === 'input' && this.defaultList) {
        sourceList = this.defaultList;
        targetList = this.defaultList;
      } else if (this.inputMode === 'search') {
        sourceList = this.list;
        targetList = this.list;
        keyWord = this.getKeyWord(val);
      }
      if (keyWord) {
        const key = keyWord.toLowerCase();
        targetList = sourceList.filter(item => item[searchKey] && String(item[searchKey]).toLowerCase()
          .indexOf(key) > -1);
      }
      targetList.forEach((item) => {
        item.isSelected = false;
      });
      this.curSelectIndex = -1;
      this.resultList = JSON.parse(JSON.stringify(targetList));
    },
  },
  mounted() {
    this.initInputLayout();
    this.numberInput = debounce((event) => {
      const { value } = event.target;
      this.numberInputHandler(value, event.target);
    }, this.debounceTimer);
  },
  methods: {
    changeCurValue(isTrigger) {
      const value = `${this.value}`;
      // 如果是选择模式，从列表项匹配
      if (this.isSelectMode) {
        const selectItem = this.getItemByKey(this.value);
        if (selectItem) {
          if (selectItem.type === 'variable') {
            this.curValue = `{{${selectItem[this.displayKey]}}}`;
          } else {
            this.curValue = selectItem[this.displayKey];
          }
          // 用户可以配置自动触发，用于实现多个联动
          if (isTrigger) {
            this.$emit('item-selected', value, selectItem, isTrigger);
          }
        } else {
          this.curValue = value;
        }
        return;
      }
      if (value === '' || value.startsWith(this.searchPrefix)) {
        this.curValue = value;
        return;
      }
      // let newVal = parseInt(value)
      // if (this.type === 'decimals') {
      //     newVal = Number(value)
      // }
      this.curValue = value;
    },
    paste(event) {
      this.$emit('paste', event);
    },
    keyup(event) {
      const { code } = event;
      switch (code) {
        case 'ArrowDown':
          if (this.type === 'number' && this.inputMode === 'input') {
            this.minus();
          } else {
            this.selectNextItem();
          }
          break;
        case 'ArrowUp':
          if (this.type === 'number' && this.inputMode === 'input') {
            this.add();
          } else {
            this.selectPrevItem();
          }
          break;
        case 'Enter':
          if (this.inputMode === 'search' || this.defaultList) {
            this.confirmSelect(event);
          } else {
            this.$emit('blur', event);
          }
          break;
      }
    },
    selectNextItem() {
      this.curSelectIndex = this.curSelectIndex + 1;
      if (this.curSelectIndex >= this.resultList.length) {
        this.curSelectIndex = this.resultList.length - 1;
      }
      this.selectItemByIndex(this.curSelectIndex);
    },
    selectPrevItem() {
      this.curSelectIndex = this.curSelectIndex - 1;
      if (this.curSelectIndex < 0) {
        this.curSelectIndex = 0;
      }
      this.selectItemByIndex(this.curSelectIndex);
    },
    selectItemByIndex(index) {
      this.resultList.forEach((item, i) => {
        if (i === index) {
          item.isSelected = true;
        } else {
          item.isSelected = false;
        }
      });
      this.resultList = JSON.parse(JSON.stringify(this.resultList));
      this.setScrollTop();
    },
    getItemByKey(key) {
      let selectItem = null;
      for (const item of this.defaultList) {
        if (String(item[this.settingKey]) === String(key)) {
          selectItem = item;
          selectItem.type = 'normal';
        }
      }
      for (const item of this.list) {
        if (`{{${item[this.settingKey]}}}` === key) {
          selectItem = item;
          selectItem.type = 'variable';
        }
      }
      return selectItem;
    },
    selectItemByKey(key) {
      if (!key) {
        this.curSelectIndex = -1;
      }
      this.resultList.forEach((item, index) => {
        if (String(item[this.settingKey]) === String(key)) {
          item.isSelected = true;
          this.curSelectIndex = index;
        } else {
          item.isSelected = false;
        }
      });
      this.resultList = JSON.parse(JSON.stringify(this.resultList));
      this.setScrollTop();
    },
    setScrollTop() {
      const MAX_SHOW_NUM = 3;
      const LIST_ITEM_HEIGHT = 42;
      if (this.curSelectIndex > MAX_SHOW_NUM) {
        const offset = this.curSelectIndex - MAX_SHOW_NUM;
        const scrollTop = offset * LIST_ITEM_HEIGHT;
        this.$el.querySelector('.selector-list-box').scrollTop = scrollTop;
      }
    },
    confirmSelect(event, index) {
      if (index !== undefined) {
        this.curSelectIndex = index;
      }
      if (this.resultList[this.curSelectIndex]) {
        const selectItem = this.resultList[this.curSelectIndex];
        const val = selectItem[this.settingKey];
        const text = selectItem[this.displayKey];
        if (val) {
          let inputText = '';
          let inputVal = '';
          let preVal = '';
          let newVal = '';
          if (this.inputMode === 'search') {
            inputText = `${this.searchPrefix}${text}${this.searchSuffix}`;
            inputVal = `${this.searchPrefix}${val}${this.searchSuffix}`;
            preVal = this.value;
            newVal = inputVal;
            if (preVal && this.searchReg.exec(preVal)) {
              newVal = preVal.replace(this.searchReg, inputVal);
            }
          } else {
            inputText = `${text}`;
            inputVal = `${val}`;
            preVal = this.value;
            newVal = inputVal;
          }
          event.target.value = inputText;
          this.curValue = inputText;
          this.localValue = val;
          this.$emit('update:value', newVal);
          this.$emit('input', newVal);
          this.$emit('item-selected', newVal, selectItem, false);
          this.hideListPanel();
          this.inputMode = 'input';
        }
      }
    },
    getPower(val) {
      const valueString = val.toString();
      const dotPosition = valueString.indexOf('.');
      let power = 0;
      if (dotPosition > -1) {
        power = valueString.length - dotPosition - 1;
      }
      return Math.pow(10, power);
    },
    checkMinMax(val) {
      if (val <= this.minNumber) {
        val = this.minNumber;
        this.isMin = true;
      } else {
        this.isMin = false;
      }
      if (val >= this.maxNumber) {
        val = this.maxNumber;
        this.isMax = true;
      } else {
        this.isMax = false;
      }
      return val;
    },
    textInput(val, event) {
      if (val.startsWith(this.searchPrefix)) {
        this.inputMode = 'search';
        this.keyWord = val;
        this.showListPanel(event);
      } else {
        this.inputMode = 'input';
        this.keyWord = '';
        if (this.defaultList) {
          this.keyWord = val;
          this.showListPanel(event);
        } else {
          this.hideListPanel();
        }
      }
      this.curValue = val;
      this.$emit('change', val);
      this.$emit('input', val);
      // 如果不是选择模式，直接修改值
      if (!this.isSelectMode) {
        this.$emit('update:value', val);
      }
    },
    userInput(event) {
      let val = event.target.value;
      if (!this.useChinese) {
        val = val.replace(this.chineseReg, '');
      }
      if (this.type === 'number' && val && this.searchReg.exec(val)) {
        const match = this.searchReg.exec(val);
        val = match[0];
        event.target.value = val;
      }
      if (this.type === 'number') {
        if (val === '{') {
          this.curValue = val;
          this.hideListPanel();
          return;
        }
        // 支持负数
        if (val === '-') {
          this.curValue = val;
          this.hideListPanel();
          return;
        }
        if (val.startsWith(this.searchPrefix)) {
          this.inputMode = 'search';
          this.keyWord = val;
          this.showListPanel(event);
          this.curValue = val;
          this.$emit('change', val);
          this.$emit('input', val);
          return;
        }
        this.inputMode = 'input';
        this.numberInput(event);
      } else {
        this.textInput(val, event);
      }
    },
    numberInputHandler(value, target) {
      if (value === '') {
        this.$emit('update:value', value);
        this.$emit('change', value);
        this.curValue = value;
        target && (target.value = value);
        return;
      }
      if (value !== '' && value.indexOf('.') === (value.length - 1)) {
        return;
      }
      if (value !== '' && value.indexOf('.') > -1 && Number(value) === 0) {
        return;
      }
      // 支持输入百分比
      let hasPercentCode = false;
      if (this.percentEnable) {
        const valueStr = value.split('');
        const lastChar = valueStr[valueStr.length - 1];
        if (lastChar === '%') {
          value = value.substr(0, (valueStr.length - 1));
          hasPercentCode = true;
        }
      }
      let newVal = parseInt(value);
      if (this.isDecimals) {
        newVal = Number(value);
      }
      if (!isNaN(newVal)) {
        if (hasPercentCode) {
          newVal += '%';
        }
        this.setNumberValue(newVal, target);
      } else {
        target.value = this.curValue;
      }
    },
    setNumberValue(val, target) {
      val = this.checkMinMax(val);
      this.$emit('update:value', val);
      this.$emit('change', val);
      this.$emit('input', val);
      this.curValue = val;
      target && (target.value = val);
    },
    add() {
      if (this.disabled) return;
      const value = this.value || 0;
      if (typeof value !== 'number') return this.curValue;
      const power = this.getPower(value);
      const newVal = (power * value + power * this.steps) / power;
      if (newVal > this.max) return;
      this.setNumberValue(newVal);
    },
    minus() {
      if (this.disabled) return;
      const value = this.value || 0;
      if (typeof value !== 'number') return this.curValue;
      const power = this.getPower(value);
      const newVal = parseInt(power * value - power * this.steps) / power;
      if (newVal < this.min) return;
      this.setNumberValue(newVal);
    },
    getKeyWord(val) {
      val = `${val}`;
      let keyWord = '';
      if (val.startsWith(this.searchPrefix)) {
        const startIndex = this.searchPrefix.length;
        let endIndex = val.length;
        if (val.endsWith(this.searchSuffix)) {
          endIndex = val.length - this.searchSuffix.length;
        }
        keyWord = val.substring(startIndex, endIndex);
      } else if (this.searchReg.exec(val)) {
        const match = this.searchReg.exec(val);
        if (match?.[1]) {
          keyWord = match[1];
        }
      } else {
        keyWord = '';
      }
      return keyWord;
    },
    initInputLayout() {
      // const element = this.$el
    },
    showListPanel(event) {
      if (this.disabled) {
        return;
      }
      this.initSelectorPosition(event.currentTarget);
      clearTimeout(this.timer);
      this.isListPanelShow = true;
    },
    hideListPanel() {
      this.timer = setTimeout(() => {
        this.isListPanelShow = false;
      }, 200);
    },
    focusHandler(event) {
      if (!this.isSelectMode) {
        event.target.select();
      }
      const val = `${this.value}`;
      const key = this.getKeyWord(val);
      if (key) {
        this.inputMode = 'search';
        const list = JSON.parse(JSON.stringify(this.list));
        list.forEach((item) => {
          item.isSelected = false;
        });
        this.resultList = list;
        this.selectItemByKey(key);
        this.setScrollTop();
        this.showListPanel(event);
      } else {
        this.inputMode = 'input';
        // 如果有下拉列表，在输入模式显示下拉列表
        if (this.defaultList && this.defaultList) {
          const list = JSON.parse(JSON.stringify(this.defaultList));
          list.forEach((item) => {
            item.isSelected = false;
          });
          this.resultList = list;
          this.selectItemByKey(val);
          this.setScrollTop();
          this.showListPanel(event);
        } else {
          this.resultList = [];
        }
      }
      this.$emit('focus', event);
    },
    blurHandler(event) {
      if (this.type === 'number') {
        this.curValue = this.value;
      } else if (this.isSelectMode) {
        const selectItem = this.getItemByKey(this.value);
        if (selectItem) {
          if (selectItem.type === 'variable') {
            this.curValue = `{{${selectItem[this.displayKey]}}}`;
          } else {
            this.curValue = selectItem[this.displayKey];
          }
        } else {
          this.curValue = '';
        }
      }
      this.hideListPanel();
      this.timer = setTimeout(() => {
        this.$emit('blur', event);
      }, 200);
    },
    initSelectorPosition(currentTarget) {
      if (currentTarget) {
        const distanceTop = getActualTop(currentTarget);
        const winHeight = document.body.clientHeight;
        let ySet = {};
        let listHeight = this.list.length * 42;
        if (listHeight > 160) {
          listHeight = 160;
        }
        const scrollTop = document.documentElement.scrollTop || document.body.scrollTop;
        if ((distanceTop + listHeight + 42 - scrollTop) < winHeight) {
          ySet = {
            top: '34px',
            bottom: 'auto',
          };
          this.listSlideName = 'toggle-slide';
        } else {
          ySet = {
            top: 'auto',
            bottom: '34px',
          };
          this.listSlideName = 'toggle-slide2';
        }
        this.panelStyle = { ...ySet, minWidth: '235px' };
      }
    },
  },
};
</script>
<style scoped>
    @import './index.css';
</style>
