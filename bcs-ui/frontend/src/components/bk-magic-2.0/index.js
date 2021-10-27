import Vue from 'vue'
import bkButton from './bk-button.js'
import bkSwitcher from './bk-switcher.js'
import bkSideslider from './bk-sideslider'
import bkCollapse from './bk-collapse'
import bkCollapseItem from './bk-collapse-item'
import bkCheckbox from './bk-checkbox.js'
import bkTab from './bk-tab.js'
import bkTabpanel from './bk-tab-panel.js'
import bkDialog from './bk-dialog'
import bkInput from './bk-input.js'
import bkOption from './bk-option.js'
import bkSelect from './bk-select.js'
import bkDropdownMenu from './bk-dropdown-menu.js'
import bkPagination from './bk-pagination.js'
import bkTable from './bk-table.js'
import bkTableColumn from './bk-table-column.js'
import bkTableSettingContent from './bk-table-setting-content.js'
import bkTooltip from './bk-tooltip.js'
import bkRadio from './bk-radio.js'
import bkRadioButton from './bk-radio-button.js'
import bkRadioGroup from './bk-radio-group.js'
import bkTag from './bk-tag.js'
import bkDatePicker from './bk-date-picker.js'
import bkTree from './bk-tree.js'
import bkPaging from './bk-paging.js'
import bkInfo from './bk-info.js'
import bkException from './bk-exception.js'
import bkForm from './bk-form.js'
import bkFormItem from './bk-form-item.js'
import bkAlert from './bk-alert.js'
import bkPopover from './bk-popover.js'
import bkBigTree from './bk-big-tree.js'
import bkVirtualScroll from './bk-virtual-scroll.js'

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
    bkVirtualScroll
]

function install () {
    components.forEach(component => {
        Vue.component(component.name, component)
    })
}

export default {
    install
}
