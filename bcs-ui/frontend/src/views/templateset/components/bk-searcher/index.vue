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
            type="text" class="input" ref="searchInput" v-model="curInputValue"
            :style="{ maxWidth: `${maxInputWidth}px`, minWidth: `${minInputWidth}px` }"
            :maxlength="inputSearchKey || showFilterValue ? Infinity : 0"
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
            <a href="javascript:void(0);">{{$t('????????????')}}</a>
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
    // ?????????????????????
    fixedSearchParams: {
      type: Array,
      default: [],
    },
    // ????????????????????????
    filterList: {
      type: Array,
      default: [],
    },
    // ???????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
    inputSearchKey: {
      type: String,
      default: 'search',
    },
    // ????????????????????????
    minInputWidth: {
      type: Number,
      default: 100,
    },
    // ????????????????????????
    maxInputWidth: {
      type: Number,
      default: 200,
    },
    // ????????????????????????
    mask: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      curInputValue: '',

      // ????????????
      searchParams: [],
      // ???????????????????????? id
      fixedSearchParamsIds: [],
      // filterList ?????????
      filterListCache: [],
      // ?????????????????? filter ????????? filter value ??????
      filterValueKey: '',
      // ???????????? filter ??????filter ??????????????????
      filterValueList: [],
      // filterValueList ????????????????????????
      filterValueListCache: [],
      // filterValue ?????????????????????
      filterValueKeyboardIndex: -1,
      // ?????? filter ????????????????????? loading
      filterValueLoading: false,
      // ???????????? filter ??????????????????
      showFilterValue: false,
      // ?????????????????????
      showFilter: false,
      // search-params-wrapper ?????? li ????????? margin ???
      searchParamsItemMargin: 3,
      // ??????????????????????????????
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

    // ???????????? searchParams ??????????????????????????????????????????
    // this.$refs.bkSearcher.$emit('resetSearchParams')
    this.$on('resetSearchParams', (isEmitSearch) => {
      this.searchParams.splice(0, this.searchParams.length, ...[]);
      // ?????????????????? search
      if (isEmitSearch) {
        this.$emit('search', this.fixedSearchParams);
      }
    });
  },
  methods: {
    /**
             * filter ??????????????????
             *
             * @param {Object} filter ??????????????? filter
             * @param {number} filterIndex ??????????????? filter ?????????
             */
    selectFilter(filter) {
      this.filterValueLoading = true;
      // ?????? filter ???????????????
      if (filter.list && !filter.dynamicData) {
        const searchParams = [];
        searchParams.splice(0, 0, ...this.searchParams);
        searchParams.push(filter);
        this.searchParams.splice(0, this.searchParams.length, ...searchParams);

        // ????????????????????? 8 px????????? padding 10 px????????? padding ????????? 20 px
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
          // ?????? filter list ?????????????????????????????????????????????????????????????????????
          // ?????????????????????????????????????????????????????????
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
             * filter value ??????????????????
             *
             * @param {Object} fv ??????????????? fv
             */
    selectFilterValue(fv) {
      const filterValueList = [];
      this.filterValueList.forEach((item) => {
        item.isSelected = item.id === fv.id;
        filterValueList.push(JSON.parse(JSON.stringify(item)));
      });
      this.filterValueList.splice(0, this.filterValueList.length, ...filterValueList);

      const selected = this.filterValueList.filter(val => val.isSelected)[0];

      // ????????????????????????????????? fixedSearchParams
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
        // ???????????? searchParams
        const searchParams = [];
        this.searchParams.forEach((sp) => {
          if (sp.id === this.filterValueKey) {
            sp.value = selected;
          }
          searchParams.push(JSON.parse(JSON.stringify(sp)));
        });
        this.searchParams.splice(0, this.searchParams.length, ...searchParams);

        // ?????? filterList????????????????????????
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
             * ????????????????????? ????????????
             *
             * @param {Object} fsp ?????????????????????????????????
             * @param {number} fspIndex ????????????????????????????????? ??????
             */
    fixedSearchParamsClickHandler(e, fsp) {
      // ??????????????????????????????????????????????????????????????????
      if (!fsp.list || fsp.list.length <= 1) {
        return;
      }
      this.hideFilterList();

      const target = e.currentTarget;

      const nameNode = target.querySelector('.name');
      // ???????????????????????????????????????????????? .name ??????valueContainerNode ?????????
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
             * ???????????? ????????????
             *
             * @param {Object} e ????????????
             * @param {Object} sp ?????????????????????????????????
             * @param {number} spIndex ????????????????????????????????? ??????
             */
    searchParamsClickHandler(e, sp) {
      this.hideFilterList();

      const target = e.currentTarget;
      const nameNode = target.querySelector('.name');
      // ???????????????????????????????????????????????? .name ??????valueContainerNode ?????????
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
             * ?????? click ??????
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
             * ??????????????????????????? param
             *
             * @param {Object} e ????????????
             * @param {Object} sp ?????????????????????????????????
             * @param {number} spIndex ????????????????????????????????? ??????
             */
    removeSearchParams(e, sp) {
      const searchParams = [];
      this.searchParams.forEach((s) => {
        if (s.id !== sp.id) {
          searchParams.push(JSON.parse(JSON.stringify(s)));
        }
      });
      this.searchParams.splice(0, this.searchParams.length, ...searchParams);

      // ?????? filterList???????????????????????????
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
             * ?????? filterList ??? filterValueList ?????????
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
      // ?????? filter ??? filter value ?????????????????? searchParams ?????? ????????? value ???
      // ????????????????????? filter value ?????????????????????????????????
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
             * ????????? keyup ??????
             *
             * @param {Object} e ????????????
             */
    inputKeyup(e) {
      if (!this.showFilterValue) {
        return;
      }
      const { filterValueListNode } = this.$refs;
      // ???????????? 320????????? item ?????? 42
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
             * ?????? input ??????????????????????????????
             *
             * @param {Object} e ????????????
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
