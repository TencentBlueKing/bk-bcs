import Vue from 'vue'

const bkInfo = Vue.prototype.$bkInfo

const Info = function (opts = {}) {
    if (opts.content) {
        if (typeof opts.content === 'string') {
            opts.subTitle = opts.content
        }
        opts.subHeader = opts.content
    }

    if (opts.defaultInfo) {
        opts.extCls = 'default-info'
    }

    if (opts.clsName) {
        opts.extCls = opts.extCls + ' ' + opts.clsName
    }

    opts.closeIcon = true
    opts.confirmLoading = true

    if (!opts.width) {
        opts.width = 400
    }
    bkInfo(opts)
}

Vue.prototype.$bkInfo = Info

export default Info
