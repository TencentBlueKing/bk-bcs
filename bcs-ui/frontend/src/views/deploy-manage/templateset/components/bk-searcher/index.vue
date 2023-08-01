<template>
  <div class="bk-searcher-wrapper" ref="searchWrapper" v-clickoutside="hideFilterList">
    <div class="bk-searcher-mask" v-if="showMask"></div>
    <div class="bk-searcher" @click="foucusSearcher($event)">
      <ul class="search-params-wrapper" ref="searchParamsWrapper">
        <template v-if="fixedSearchParams && fixedSearchParams.length">
          <li v-for="(fsp, fspIndex) in fixedSearchParams" :key="fspIndex">
            <div class="selectable" @click.stop="fixedSearchParamsClickHandler($event, fsp, fspIndex)">
              <div class="name">{{fsp.text}}</div>
              <div class="value-container" v-if="fsp.value">
                <div class="value">{{fsp.value.text}}</div>
              </div>
            </div>
          </li>
        </template>
        <template v-if="searchParams && searchParams.length">
          <li v-for="(sp, spIndex) in searchParams" :key="spIndex">
            <div class="selectable" @click.stop="searchParamsClickHandler($event, sp, spIndex)">
              <div class="name">{{sp.text}}</div>
              <div class="value-container" v-if="sp.value">
                <div class="value">{{sp.value.text}}</div>
                <div
                  class="remove-search-params"
                  @click.stop="removeSearchParams($event, sp, spIndex)"><i class="bcs-icon bcs-icon-close"></i></div>
              </div>
            </div>
          </li>
        </template>
        <li ref="searchInputParent">
          <input
            type="text" class="input !h-[32px] !pl-[10px]" ref="searchInput" v-model="curInputValue"
            :style="{ maxWidth: `${maxInputWidth}px`, minWidth: `${minInputWidth}px` }"
            :maxlength="inputSearchKey || showFilterValue ? Infinity : 0"
            :placeholder="searchParams && searchParams.length
              ? '' : placeholder"
            @keyup="inputKeyup($event)"
            @keypress="preventKeyboardEvt($event)"
            @keydown="preventKeyboardEvt($event)">
        </li>
      </ul>
    </div>
    <div class="bk-searcher-dropdown-menu filter-list" v-show="showFilter">
      <div
        class="bk-searcher-dropdown-content"
        :class="showFilter ? 'is-show' : ''" :style="{ left: `${searcherDropdownLeft}px` }">
        <ul class="bk-searcher-dropdown-list" v-bkloading="{ isLoading: filterValueLoading }">
          <li v-for="(filter, filterIndex) in filterList" :key="filterIndex">
            <a href="javascript:void(0);" @click="selectFilter(filter, filterIndex)">{{filter.text}}</a>
          </li>
        </ul>
      </div>
    </div>

    <div class="bk-searcher-dropdown-menu filter-value-list" v-show="showFilterValue">
      <div
        class="bk-searcher-dropdown-content"
        ref="filterValueListNode"
        :class="showFilterValue ? 'is-show' : ''" :style="{ left: `${searcherDropdownLeft}px` }">
        <ul class="bk-searcher-dropdown-list" v-if="filterValueList && filterValueList.length">
          <li v-for="(fv, fvIndex) in filterValueList" :key="fvIndex">
            <a
              href="javascript:void(0);"
              :title="fv.text" :class="fvIndex === filterValueKeyboardIndex ? 'active' : ''"
              @click="selectFilterValue(fv)">{{fv.text}}</a>
          </li>
        </ul>
        <ul class="bk-searcher-dropdown-list" v-else>
          <li>
            <a href="javascript:void(0);">{{$t('generic.msg.empty.noData3')}}</a>
          </li>
        </ul>
      </div>
    </div>

  </div>
</template>

<script>
// eslint-disable-next-line vue/no-mutating-props
import clickoutside from './clickoutside';
import { getActualLeft, getStringLen, insertAfter } from '@/common/util';

export default {
  name: 'BkSearcher',
  directives: {
    clickoutside,
  },
  props: {
    // 固定的搜索参数
    fixedSearchParams: {
      type: Array,
      default: [],
    },
    // 过滤项下拉框列表
    filterList: {
      type: Array,
      default: [],
    },
    // 在文本框中输入任意关键字搜索，如果为空，那么说明不支持在文本框中输入任意关键字搜索
    inputSearchKey: {
      type: String,
      default: 'search',
    },
    // 输入框的最小宽度
    minInputWidth: {
      type: Number,
      default: 100,
    },
    // 输入框的最大宽度
    maxInputWidth: {
      type: Number,
      default: 200,
    },
    // 输入框的最大宽度
    mask: {
      type: Boolean,
      default: false,
    },
    placeholder: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      curInputValue: '',

      // 搜索参数
      searchParams: [],
      // 固定的搜索参数的 id
      fixedSearchParamsIds: [],
      // filterList 的缓存
      filterListCache: [],
      // 标识是由哪个 filter 弹出的 filter value 弹层
      filterValueKey: '',
      // 选中一个 filter 后，filter 的可选值集合
      filterValueList: [],
      // filterValueList 的缓存，搜索使用
      filterValueListCache: [],
      // filterValue 键盘上下的索引
      filterValueKeyboardIndex: -1,
      // 渲染 filter 的可选值集合的 loading
      filterValueLoading: false,
      // 是否显示 filter 的可选值集合
      showFilterValue: false,
      // 是否显示过滤项
      showFilter: false,
      // search-params-wrapper 里的 li 元素的 margin 值
      searchParamsItemMargin: 3,
      // 过滤项下拉框的左偏移
      searcherDropdownLeft: 0,
      showMask: false,
    };
  },
  computed: {
  },
  watch: {
    curInputValue(val) {
      const filterValueList = [];
      this.filterValueListCache.forEach((item) => {
        if (item.text.indexOf(val) >= 0) {
          filterValueList.push(JSON.parse(JSON.stringify(item)));
        }
      });
      this.filterValueList.splice(0, this.filterValueList.length, ...filterValueList);
      this.filterValueKeyboardIndex = -1;
    },
    showFilterValue(val) {
      if (val) {
        const { filterValueListNode } = this.$refs;
        filterValueListNode?.scrollTo(0, 0);
      }
    },
    mask(val) {
      this.showMask = val;
    },
  },
  mounted() {
    const fixedSearchParams = [];
    const fixedSearchParamsIds = [];
    this.fixedSearchParams.forEach((fsp) => {
      let selected = fsp.list.filter(val => val.isSelected)[0];
      if (!selected) {
        selected = fsp.list[0];
      }
      fsp.value = selected;
      fixedSearchParams.push(JSON.parse(JSON.stringify(fsp)));
      fixedSearchParamsIds.push(fsp.id);
    });
    // eslint-disable-next-line vue/no-mutating-props
    this.fixedSearchParams.splice(0, this.fixedSearchParams.length, ...fixedSearchParams);
    this.fixedSearchParamsIds.splice(0, this.fixedSearchParamsIds.length, ...fixedSearchParamsIds);

    const filterList = [];
    const searchParams = [];

    this.filterList.forEach((filter) => {
      if (filter.list) {
        const selected = filter.list.filter(val => val.isSelected)[0];
        if (selected) {
          filter.value = selected;
          this.filterValueKey = filter.id;
          searchParams.push(JSON.parse(JSON.stringify(filter)));
        } else {
          filterList.push(JSON.parse(JSON.stringify(filter)));
        }
      } else {
        filterList.push(JSON.parse(JSON.stringify(filter)));
      }
    });
    filterList.sort((a, b) => (a.id < b.id ? 1 : -1));
    this.searchParams.splice(0, this.searchParams.length, ...searchParams);

    // eslint-disable-next-line vue/no-mutating-props
    this.filterList.splice(0, this.filterList.length, ...filterList);
    this.filterListCache.splice(0, this.filterListCache.length, ...filterList);

    const { searchWrapper, searchParamsWrapper } = this.$refs;
    this.searcherDropdownLeft = getActualLeft(searchParamsWrapper) - getActualLeft(searchWrapper)
                + this.searchParamsItemMargin;

    // 绑定清除 searchParams 的方法，父组件调用方式如下：
    // this.$refs.bkSearcher.$emit('resetSearchParams')
    this.$on('resetSearchParams', (isEmitSearch) => {
      this.searchParams.splice(0, this.searchParams.length, ...[]);
      // 需要重新触发 search
      if (isEmitSearch) {
        this.$emit('search', this.fixedSearchParams);
      }
    });
  },
  methods: {
    /**
             * filter 选择事件回调
             *
             * @param {Object} filter 当前选择的 filter
             * @param {number} filterIndex 当前选择的 filter 的索引
             */
    selectFilter(filter) {
      this.filterValueLoading = true;
      // 当前 filter 有可用的值
      if (filter.list && !filter.dynamicData) {
        const searchParams = [];
        searchParams.splice(0, 0, ...this.searchParams);
        searchParams.push(filter);
        this.searchParams.splice(0, this.searchParams.length, ...searchParams);

        // 一个字符大约是 8 px，横向 padding 10 px，左右 padding 一共是 20 px
        this.searcherDropdownLeft += getStringLen(filter.text) * 8 + 10;
        setTimeout(() => {
          this.$refs.searchInput.focus();
          this.filterValueKey = filter.id;
          this.filterValueList.splice(0, this.filterValueList.length, ...filter.list);
          this.filterValueListCache.splice(0, this.filterValueListCache.length, ...filter.list);
          this.filterValueLoading = false;
          this.showFilter = false;
          this.showFilterValue = true;
        }, 300);
      } else {
        const searchParams = [];
        searchParams.splice(0, 0, ...this.searchParams);
        searchParams.push(filter);
        this.searchParams.splice(0, this.searchParams.length, ...searchParams);

        this.searcherDropdownLeft += getStringLen(filter.text) * 8 + 10;

        new Promise((resolve, reject) => {
          const fixedSearchParams = {};
          this.fixedSearchParams.forEach((fsp) => {
            fixedSearchParams[fsp.id] = fsp.value.id;
          });
          this.$emit('getFilterListData', filter, fixedSearchParams, resolve, reject);
        }).then((data) => {
          filter.list = data;
          // 标识 filter list 数据是根据请求异步获取的，不是一开始就设置好的
          // 用这个属性来控制，每次都从后端获取数据
          filter.dynamicData = true;
          filter = JSON.parse(JSON.stringify(filter));
          setTimeout(() => {
            this.$refs.searchInput.focus();
            this.filterValueKey = filter.id;
            this.filterValueList.splice(0, this.filterValueList.length, ...filter.list);
            this.filterValueListCache.splice(0, this.filterValueListCache.length, ...filter.list);
            this.filterValueLoading = false;
            this.showFilter = false;
            this.showFilterValue = true;
          }, 200);
        }, (err) => {
          console.error(err);
        });
      }
    },

    /**
             * filter value 选择事件回调
             *
             * @param {Object} fv 当前选择的 fv
             */
    selectFilterValue(fv) {
      const filterValueList = [];
      this.filterValueList.forEach((item) => {
        item.isSelected = item.id === fv.id;
        filterValueList.push(JSON.parse(JSON.stringify(item)));
      });
      this.filterValueList.splice(0, this.filterValueList.length, ...filterValueList);

      const selected = this.filterValueList.filter(val => val.isSelected)[0];

      // 说明选择的是固定参数的 fixedSearchParams
      if (this.fixedSearchParamsIds.indexOf(this.filterValueKey) > -1) {
        const fixedSearchParams = [];
        this.fixedSearchParams.forEach((fsp) => {
          if (fsp.id === this.filterValueKey) {
            fsp.value = selected;
          }
          fixedSearchParams.push(JSON.parse(JSON.stringify(fsp)));
        });
        // eslint-disable-next-line vue/no-mutating-props
        this.fixedSearchParams.splice(0, this.fixedSearchParams.length, ...fixedSearchParams);
      } else {
        // 选择的是 searchParams
        const searchParams = [];
        this.searchParams.forEach((sp) => {
          if (sp.id === this.filterValueKey) {
            sp.value = selected;
          }
          searchParams.push(JSON.parse(JSON.stringify(sp)));
        });
        this.searchParams.splice(0, this.searchParams.length, ...searchParams);

        // 更新 filterList，把选择的删除掉
        const filterList = [];
        this.filterList.forEach((item) => {
          if (item.id !== this.filterValueKey) {
            filterList.push(JSON.parse(JSON.stringify(item)));
          }
        });
        filterList.sort((a, b) => (a.id < b.id ? 1 : -1));
        // eslint-disable-next-line vue/no-mutating-props
        this.filterList.splice(0, this.filterList.length, ...filterList);
      }

      this.hideFilterList();
      this.$emit('search', this.fixedSearchParams.concat(this.searchParams));
    },

    /**
             * 固定的搜索参数 点击事件
             *
             * @param {Object} fsp 当前点击的固定参数对象
             * @param {number} fspIndex 当前点击的固定参数对象 索引
             */
    fixedSearchParamsClickHandler(e, fsp) {
      // 如果不存在，或者只有一个结果，那么不允许点击
      if (!fsp.list || fsp.list.length <= 1) {
        return;
      }
      this.hideFilterList();

      const target = e.currentTarget;

      const nameNode = target.querySelector('.name');
      // 当下拉框出现但是不选择，直接点击 .name 时，valueContainerNode 不存在
      const valueContainerNode = target.querySelector('.value-container');

      const { searchWrapper, searchParamsWrapper } = this.$refs;
      // this.searcherDropdownLeft = getActualLeft(target) - getActualLeft(searchWrapper)
      //     + this.searchParamsItemMargin + getActualLeft(valueContainerNode) - getActualLeft(nameNode)
      this.searcherDropdownLeft = getActualLeft(searchParamsWrapper) - getActualLeft(searchWrapper)
                    + this.searchParamsItemMargin
                    + getActualLeft(valueContainerNode || nameNode) - getActualLeft(nameNode);

      this.showFilter = false;
      this.showFilterValue = true;
      this.filterValueKey = fsp.id;
      this.filterValueList.splice(0, this.filterValueList.length, ...fsp.list);
      this.filterValueListCache.splice(0, this.filterValueListCache.length, ...fsp.list);

      insertAfter(this.$refs.searchInputParent, target.parentNode);
      this.$nextTick(() => {
        this.$refs.searchInput.focus();
        if (!valueContainerNode) {
          this.hideFilterList();
        }
      });
    },

    /**
             * 搜索参数 点击事件
             *
             * @param {Object} e 事件对象
             * @param {Object} sp 当前点击的固定参数对象
             * @param {number} spIndex 当前点击的固定参数对象 索引
             */
    searchParamsClickHandler(e, sp) {
      this.hideFilterList();

      const target = e.currentTarget;
      const nameNode = target.querySelector('.name');
      // 当下拉框出现但是不选择，直接点击 .name 时，valueContainerNode 不存在
      const valueContainerNode = target.querySelector('.value-container');

      const { searchWrapper } = this.$refs;
      this.searcherDropdownLeft = getActualLeft(target) - getActualLeft(searchWrapper)
                    + this.searchParamsItemMargin
                    + getActualLeft(valueContainerNode || nameNode) - getActualLeft(nameNode);

      this.showFilter = false;
      this.showFilterValue = true;
      this.filterValueKey = sp.id;
      this.filterValueList.splice(0, this.filterValueList.length, ...sp.list);
      this.filterValueListCache.splice(0, this.filterValueListCache.length, ...sp.list);

      insertAfter(this.$refs.searchInputParent, target.parentNode);
      this.$nextTick(() => {
        this.$refs.searchInput.focus();
        if (!valueContainerNode) {
          this.hideFilterList();
        }
      });
    },

    /**
             * 组件 click 事件
             */
    foucusSearcher() {
      this.$nextTick(() => {
        this.$refs.searchInput.focus();
      });

      if (this.showFilterValue) {
        return;
      }

      const { searchParamsWrapper, searchInput } = this.$refs;
      let searcherDropdownLeft = searchInput.offsetParent.offsetLeft + this.searchParamsItemMargin * 2;
      const offsetWidth = parseInt(searchParamsWrapper.offsetWidth, 10);
      if (searcherDropdownLeft > offsetWidth - this.minInputWidth * 2) {
        searcherDropdownLeft = offsetWidth - this.minInputWidth * 3 / 2 + this.searchParamsItemMargin * 2;
      }
      this.searcherDropdownLeft = searcherDropdownLeft;

      this.showFilterValue = false;
      this.showFilter = true;
    },

    /**
             * 删除当前点击的这个 param
             *
             * @param {Object} e 事件对象
             * @param {Object} sp 当前点击的固定参数对象
             * @param {number} spIndex 当前点击的固定参数对象 索引
             */
    removeSearchParams(e, sp) {
      const searchParams = [];
      this.searchParams.forEach((s) => {
        if (s.id !== sp.id) {
          searchParams.push(JSON.parse(JSON.stringify(s)));
        }
      });
      this.searchParams.splice(0, this.searchParams.length, ...searchParams);

      // 更新 filterList，把当前点击的删掉
      const filterList = [];
      this.filterList.forEach((item) => {
        filterList.push(JSON.parse(JSON.stringify(item)));
      });
      delete sp.value;
      filterList.push(JSON.parse(JSON.stringify(sp)));
      filterList.sort((a, b) => (a.id < b.id ? 1 : -1));
      // eslint-disable-next-line vue/no-mutating-props
      this.filterList.splice(0, this.filterList.length, ...filterList);
      this.hideFilterList();

      this.$emit('search', this.fixedSearchParams.concat(this.searchParams));
    },

    /**
             * 隐藏 filterList 和 filterValueList 的弹层
             */
    hideFilterList() {
      this.showFilter = false;
      this.showFilterValue = false;
      this.curInputValue = '';
      this.filterValueKeyboardIndex = -1;
      this.filterValueKey = '';
      this.filterValueList.splice(0, this.filterValueList.length, ...[]);
      this.filterValueListCache.splice(0, this.filterValueListCache.length, ...[]);
      this.$nextTick(() => {
        this.$refs.searchInput.blur();
      });
      // 关闭 filter 和 filter value 时，如果发现 searchParams 里有 不存在 value 的
      // 说明是还未选择 filter value 的情况，这时候要清除掉
      const searchParams = [];
      this.searchParams.forEach((sp) => {
        if (sp.value) {
          searchParams.push(JSON.parse(JSON.stringify(sp)));
        }
      });
      this.searchParams.splice(0, this.searchParams.length, ...searchParams);

      insertAfter(this.$refs.searchInputParent, this.$refs.searchParamsWrapper.lastChild);
    },

    /**
             * 输入框 keyup 事件
             *
             * @param {Object} e 事件对象
             */
    inputKeyup(e) {
      if (!this.showFilterValue) {
        return;
      }
      const { filterValueListNode } = this.$refs;
      // 最大高度 320，每个 item 高度 42
      switch (e.keyCode) {
        // down
        case 40:
          if (this.filterValueKeyboardIndex < this.filterValueList.length - 1) {
            // eslint-disable-next-line no-plusplus
            this.filterValueKeyboardIndex++;
            if (this.filterValueKeyboardIndex >= 8) {
              filterValueListNode.scrollTo(0, filterValueListNode.scrollTop + 45);
            }
          }
          break;
          // up
        case 38:
          if (this.filterValueKeyboardIndex > 0) {
            // eslint-disable-next-line no-plusplus
            this.filterValueKeyboardIndex--;
            if (this.filterValueKeyboardIndex < Math.ceil((filterValueListNode.scrollHeight - 320) / 42)) {
              filterValueListNode.scrollTo(0, filterValueListNode.scrollTop - 42);
            }
          }
          break;
          // enter
        case 13:
          // eslint-disable-next-line no-case-declarations
          const filterValueItem = this.filterValueList[this.filterValueKeyboardIndex];
          filterValueItem && this.selectFilterValue(filterValueItem);
          break;
        default:
      }
    },

    /**
             * 阻止 input 框一些按键的默认事件
             *
             * @param {Object} e 事件对象
             */
    preventKeyboardEvt(e) {
      switch (e.keyCode) {
        // down
        case 40:
          e.stopPropagation();
          e.preventDefault();
          break;
          // up
        case 38:
          e.stopPropagation();
          e.preventDefault();
          break;
          // left
        case 37:
          e.stopPropagation();
          e.preventDefault();
          break;
          // right
        case 39:
          e.stopPropagation();
          e.preventDefault();
          break;
        default:
      }
    },
  },
};
</script>

<style scoped>
    @import './index.css';
</style>
