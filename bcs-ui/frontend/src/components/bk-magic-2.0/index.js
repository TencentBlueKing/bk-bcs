/*
* Tencent is pleased to support the open source community by making
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
*
* Copyright (C) 2021 THL A29 Limited, a Tencent company.  All rights reserved.
*
* 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) is licensed under the MIT License.
*
* License for 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition):
*
* ---------------------------------------------------
* Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
* documentation files (the "Software"), to deal in the Software without restriction, including without limitation
* the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and
* to permit persons to whom the Software is furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all copies or substantial portions of
* the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO
* THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF
* CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
* IN THE SOFTWARE.
*/
import Vue from 'vue';
import bkButton from './bk-button.js';
import bkSwitcher from './bk-switcher.js';
import bkSideslider from './bk-sideslider';
import bkCollapse from './bk-collapse';
import bkCollapseItem from './bk-collapse-item';
import bkCheckbox from './bk-checkbox.js';
import bkTab from './bk-tab.js';
import bkTabpanel from './bk-tab-panel.js';
import bkDialog from './bk-dialog';
import bkInput from './bk-input.js';
import bkOption from './bk-option.js';
import bkSelect from './bk-select.js';
import bkDropdownMenu from './bk-dropdown-menu.js';
import bkPagination from './bk-pagination.js';
import bkTable from './bk-table.js';
import bkTableColumn from './bk-table-column.js';
import bkTableSettingContent from './bk-table-setting-content.js';
import bkTooltip from './bk-tooltip.js';
import bkRadio from './bk-radio.js';
import bkRadioButton from './bk-radio-button.js';
import bkRadioGroup from './bk-radio-group.js';
import bkTag from './bk-tag.js';
import bkDatePicker from './bk-date-picker.js';
import bkTree from './bk-tree.js';
import bkPaging from './bk-paging.js';
import bkInfo from './bk-info.js';
import bkException from './bk-exception.js';
import bkForm from './bk-form.js';
import bkFormItem from './bk-form-item.js';
import bkAlert from './bk-alert.js';
import bkPopover from './bk-popover.js';
import bkBigTree from './bk-big-tree.js';
import bkVirtualScroll from './bk-virtual-scroll.js';

const components = [
  bkButton,
  bkOption,
  bkSelect,
  bkDropdownMenu,
  bkSwitcher,
  bkSideslider,
  bkCollapse,
  bkCollapseItem,
  bkCheckbox,
  bkInput,
  bkPagination,
  bkTable,
  bkTableColumn,
  bkTableSettingContent,
  bkTab,
  bkTabpanel,
  bkDialog,
  bkTooltip,
  bkRadio,
  bkRadioButton,
  bkRadioGroup,
  bkTag,
  bkDatePicker,
  bkTree,
  bkPaging,
  bkInfo,
  bkException,
  bkForm,
  bkFormItem,
  bkAlert,
  bkPopover,
  bkBigTree,
  bkVirtualScroll,
];

function install() {
  components.forEach((component) => {
    Vue.component(component.name, component);
  });
}

export default {
  install,
};
