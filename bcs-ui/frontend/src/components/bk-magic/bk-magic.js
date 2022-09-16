(function (global, factory) {
	typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports, require('vue')) :
	typeof define === 'function' && define.amd ? define(['exports', 'vue'], factory) :
	(factory((global.bkMagic = {}),global.Vue));
}(this, (function (exports,Vue) { 'use strict';

Vue = Vue && Vue.hasOwnProperty('default') ? Vue['default'] : Vue;

var bkButton = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('button', { staticClass: "bk-button", class: ['bk-' + _vm.themeType, 'bk-button-' + _vm.size, { 'is-disabled': _vm.disabled, 'is-loading': _vm.loading }], attrs: { "title": _vm.title, "type": _vm.buttonType, "disabled": _vm.disabled }, on: { "click": _vm.handleClick } }, [_vm.icon ? _c('i', { staticClass: "bk-icon", class: ['icon-' + _vm.icon] }) : _vm._e(), _vm._v(" "), _c('span', [_vm._t("default")], 2)]);
    }, staticRenderFns: [],
    name: 'bk-button',
    props: {
        type: {
            type: String,
            default: 'default',
            validator: function validator(value) {
                var types = value.split(' ');
                var buttons = ['button', 'submit', 'reset'];
                var thenme = ['default', 'info', 'primary', 'warning', 'success', 'danger'];
                var valid = true;

                types.forEach(function (type) {
                    if (buttons.indexOf(type) === -1 && thenme.indexOf(type) === -1) {
                        valid = false;
                    }
                });
                return valid;
            }
        },
        size: {
            type: String,
            default: 'normal',
            validator: function validator(value) {
                return ['mini', 'small', 'normal', 'large'].indexOf(value) > -1;
            }
        },
        title: {
            type: String,
            default: ''
        },
        icon: String,
        disabled: Boolean,
        loading: Boolean
    },
    computed: {
        buttonType: function buttonType() {
            var types = this.type.split(' ');
            return types.find(function (type) {
                return type === 'submit' || type === 'button' || type === 'reset';
            });
        },
        themeType: function themeType() {
            var types = this.type.split(' ');
            return types.find(function (type) {
                return type !== 'submit' && type !== 'button' && type !== 'reset';
            });
        }
    },
    methods: {
        handleClick: function handleClick(e) {
            var _this = this;

            if (!this.disabled && !this.loading) {
                this.$emit('click', e);
            }
            this.$nextTick(function () {
                _this.$el.blur();
            });
        }
    }
};

bkButton.install = function (Vue$$1) {
  Vue$$1.component(bkButton.name, bkButton);
};

var nodeList = [];
var clickctx = '$clickoutsideCtx';
var beginClick = void 0;

document.addEventListener('mousedown', function (event) {
    return beginClick = event;
});

document.addEventListener('mouseup', function (event) {
    nodeList.forEach(function (node) {
        node[clickctx].clickoutsideHandler(event, beginClick);
    });
});

var clickoutside = {
    bind: function bind(el, binding, vnode) {
        var id = nodeList.push(el) - 1;
        var clickoutsideHandler = function clickoutsideHandler() {
            var mouseup = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
            var mousedown = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};

            if (!vnode.context || !mouseup.target || !mousedown.target || el.contains(mouseup.target) || el.contains(mousedown.target) || el === mouseup.target || vnode.context.popup && (vnode.context.popup.contains(mouseup.target) || vnode.context.popup.contains(mousedown.target))) return;

            if (binding.expression && el[clickctx].callbackName && vnode.context[el[clickctx].callbackName]) {
                    vnode.context[el[clickctx].callbackName]();
                } else {
                el[clickctx].bindingFn && el[clickctx].bindingFn();
            }
        };

        el[clickctx] = {
            id: id,
            clickoutsideHandler: clickoutsideHandler,
            callbackName: binding.expression,
            callbackFn: binding.value
        };
    },
    update: function update(el, binding) {
        el[clickctx].callbackName = binding.expression;
        el[clickctx].callbackFn = binding.value;
    },
    unbind: function unbind(el) {
        for (var i = 0, len = nodeList.length; i < len; i++) {
            if (nodeList[i][clickctx].id === el[clickctx].id) {
                nodeList.splice(i, 1);
                break;
            }
        }
    }
};

var bkDropdownMenu = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { directives: [{ name: "clickoutside", rawName: "v-clickoutside", value: _vm.handleClickoutside, expression: "handleClickoutside" }], staticClass: "bk-dropdown-menu", class: { disabled: _vm.disabled }, attrs: { "id": "bk-dropdown-menu1" }, on: { "click": _vm.handleClick, "mouseover": _vm.handleMouseover, "mouseout": _vm.handleMouseout } }, [_c('div', { staticClass: "bk-dropdown-trigger" }, [_vm._t("dropdown-trigger")], 2), _vm._v(" "), _c('div', { class: ['bk-dropdown-content', { 'is-show': _vm.isShow, 'right-align': _vm.align === 'right', 'center-align': _vm.align === 'center', 'left-align': _vm.align === 'left' }], style: _vm.menuStyle }, [_vm._t("dropdown-content")], 2)]);
    }, staticRenderFns: [],
    name: 'bk-dropdown-menu',
    directives: {
        clickoutside: clickoutside
    },
    props: {
        trigger: {
            type: String,
            default: 'mouseover',
            validator: function validator(event) {
                return ['click', 'mouseover'].includes(event);
            }
        },
        align: {
            type: String,
            default: 'left'
        },
        disabled: {
            type: Boolean,
            default: false
        }
    },
    data: function data() {
        return {
            menuStyle: null,
            timer: 0,
            isShow: false
        };
    },

    methods: {
        handleClick: function handleClick() {
            if (this.disabled || this.trigger !== 'click') return;
            this.isShow ? this.hide() : this.show();
        },
        handleMouseover: function handleMouseover() {
            if (this.trigger === 'mouseover' && !this.disabled) {
                this.show();
            }
        },
        handleMouseout: function handleMouseout() {
            if (this.trigger === 'mouseover' && !this.disabled) {
                this.hide();
            }
        },
        handleClickoutside: function handleClickoutside() {
            if (this.isShow) {
                this.hide();
            }
        },
        show: function show() {
            clearTimeout(this.timer);
            if (this.isShow) return;
            var OFFSET = 3;
            var container = this.$el;
            var trigger = container.querySelector('.bk-dropdown-trigger');
            var menuList = container.querySelector('.bk-dropdown-content');
            var triggerHeight = trigger.clientHeight;
            var menuHeight = menuList.clientHeight;
            var docHeight = window.innerHeight ? window.innerHeight : document.body.clientHeight;
            var scrollTop = document.body.scrollTop;
            var triggerBtnOffTop = trigger.offsetTop;
            var parent = trigger.offsetParent;
            while (parent) {
                triggerBtnOffTop += parent.offsetTop;
                parent = parent.offsetParent;
            }
            var menuOffsetTop = triggerHeight + OFFSET;
            if (scrollTop + docHeight - (triggerBtnOffTop + triggerHeight) > menuHeight + OFFSET) {
                this.menuStyle = {
                    top: menuOffsetTop + 'px'
                };
            } else {
                this.menuStyle = {
                    bottom: menuOffsetTop + 'px'
                };
            }
            this.isShow = true;
            this.$emit('show');
        },
        hide: function hide() {
            var _this = this;

            this.timer = setTimeout(function () {
                _this.isShow = false;
                _this.$emit('hide');
            }, 200);
        }
    }
};

bkDropdownMenu.install = function (Vue$$1) {
    Vue$$1.component(bkDropdownMenu.name, bkDropdownMenu);
};

var _typeof = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) {
  return typeof obj;
} : function (obj) {
  return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj;
};











var classCallCheck = function (instance, Constructor) {
  if (!(instance instanceof Constructor)) {
    throw new TypeError("Cannot call a class as a function");
  }
};

var createClass = function () {
  function defineProperties(target, props) {
    for (var i = 0; i < props.length; i++) {
      var descriptor = props[i];
      descriptor.enumerable = descriptor.enumerable || false;
      descriptor.configurable = true;
      if ("value" in descriptor) descriptor.writable = true;
      Object.defineProperty(target, descriptor.key, descriptor);
    }
  }

  return function (Constructor, protoProps, staticProps) {
    if (protoProps) defineProperties(Constructor.prototype, protoProps);
    if (staticProps) defineProperties(Constructor, staticProps);
    return Constructor;
  };
}();







var _extends = Object.assign || function (target) {
  for (var i = 1; i < arguments.length; i++) {
    var source = arguments[i];

    for (var key in source) {
      if (Object.prototype.hasOwnProperty.call(source, key)) {
        target[key] = source[key];
      }
    }
  }

  return target;
};





















var slicedToArray = function () {
  function sliceIterator(arr, i) {
    var _arr = [];
    var _n = true;
    var _d = false;
    var _e = undefined;

    try {
      for (var _i = arr[Symbol.iterator](), _s; !(_n = (_s = _i.next()).done); _n = true) {
        _arr.push(_s.value);

        if (i && _arr.length === i) break;
      }
    } catch (err) {
      _d = true;
      _e = err;
    } finally {
      try {
        if (!_n && _i["return"]) _i["return"]();
      } finally {
        if (_d) throw _e;
      }
    }

    return _arr;
  }

  return function (arr, i) {
    if (Array.isArray(arr)) {
      return arr;
    } else if (Symbol.iterator in Object(arr)) {
      return sliceIterator(arr, i);
    } else {
      throw new TypeError("Invalid attempt to destructure non-iterable instance");
    }
  };
}();













var toConsumableArray = function (arr) {
  if (Array.isArray(arr)) {
    for (var i = 0, arr2 = Array(arr.length); i < arr.length; i++) arr2[i] = arr[i];

    return arr2;
  } else {
    return Array.from(arr);
  }
};

function isVNode(node) {
    return (typeof node === 'undefined' ? 'undefined' : _typeof(node)) === 'object' && node.hasOwnProperty('componentOptions');
}





function getActualTop(node) {
    var actualTop = node.offsetTop;
    var current = node.offsetParent;

    while (current !== null) {
        actualTop += current.offsetTop;
        current = current.offsetParent;
    }

    return actualTop;
}

function getActualLeft(node) {
    var actualLeft = node.offsetLeft;
    var current = node.offsetParent;

    while (current !== null) {
        actualLeft += current.offsetLeft;
        current = current.offsetParent;
    }

    return actualLeft;
}

function addClass(node, className) {
    var classNames = className.split(' ');
    if (node.nodeType === 1) {
        if (!node.className && classNames.length === 1) {
            node.className = className;
        } else {
            var setClass = ' ' + node.className + ' ';
            classNames.forEach(function (cl) {
                if (setClass.indexOf(' ' + cl + ' ') < 0) {
                    setClass += cl + ' ';
                }
            });
            var rtrim = /^\s+|\s+$/;
            node.className = setClass.replace(rtrim, '');
        }
    }
}

function removeClass(node, className) {
    var classNames = className.split(' ');
    if (node.nodeType === 1) {
        var setClass = ' ' + node.className + ' ';
        classNames.forEach(function (cl) {
            setClass = setClass.replace(' ' + cl + ' ', ' ');
        });
        var rtrim = /^\s+|\s+$/;
        node.className = setClass.replace(rtrim, '');
    }
}

function camelize(str) {
    return str.replace(/-(\w)/g, function (strMatch, p1) {
        return p1.toUpperCase();
    });
}

function getStyle(elem, prop) {
    if (!elem || !prop) {
        return false;
    }

    var value = elem.style[camelize(prop)];

    if (!value) {
        var css = '';
        if (document.defaultView && document.defaultView.getComputedStyle) {
            css = document.defaultView.getComputedStyle(elem, null);
            value = css ? css.getPropertyValue(prop) : null;
        }
    }

    return String(value);
}

var monthLong = {
    '01': 'January',
    '02': 'February',
    '03': 'March',
    '04': 'April',
    '05': 'May',
    '06': 'June',
    '07': 'July',
    '08': 'August',
    '09': 'September',
    '10': 'October',
    '11': 'November',
    '12': 'December'
};

var monthShort = {
    '01': 'Jan',
    '02': 'Feb',
    '03': 'Mar',
    '04': 'Apr',
    '05': 'May',
    '06': 'Jun',
    '07': 'Jul',
    '08': 'Aug',
    '09': 'Sep',
    '10': 'Oct',
    '11': 'Nov',
    '12': 'Dec'
};

function formatMonth(month) {
    var locale = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : 'en-US';
    var isShort = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : false;

    if (locale === 'en-US') {
        return isShort ? monthShort[month] : monthLong[month];
    }
    return month;
}

function debounce(func, wait, immediate) {
    var timeout = void 0;
    var result = void 0;
    var debounced = function debounced() {
        var context = this;
        var args = arguments;

        if (timeout) {
            clearTimeout(timeout);
        }
        if (immediate) {
            var callNow = !timeout;
            timeout = setTimeout(function () {
                timeout = null;
            }, wait);
            if (callNow) {
                result = func.apply(context, args);
            }
        } else {
            timeout = setTimeout(function () {
                func.apply(context, args);
            }, wait);
        }
        return result;
    };

    debounced.cancel = function () {
        clearTimeout(timeout);
        timeout = null;
    };

    return debounced;
}

var OVERFLOW_PROPERTYS = ['overflow', 'overflow-x', 'overflow-y'];

var SCROLL_TYPES = ['scroll', 'auto'];

var MAX = 4;

var ROOT = document.body;

var VERTICAL = ['top', 'bottom'];

var HORIZONTAL = ['left', 'right'];

var DEFAULT_PLACEMENT_QUEUE = ['top', 'right', 'bottom', 'left'];

function checkScrollable(element) {
    var css = window.getComputedStyle(element, null);
    return OVERFLOW_PROPERTYS.some(function (property) {
        return ~SCROLL_TYPES.indexOf(css[property]);
    });
}

function getScrollContainer(el) {
    if (!el) {
        return ROOT;
    }

    var parent = el.parentNode;
    while (parent && parent !== ROOT) {
        if (checkScrollable(parent)) {
            return parent;
        }
        parent = parent.parentNode;
    }
    return ROOT;
}

function getBestPlacement(queue) {
    return queue.sort(function (a, b) {
        return b.weight - a.weight;
    })[0];
}

function getBoxMargin(el, parent) {
    if (!el) {
        return;
    }
    var eBox = el.getBoundingClientRect();
    var pBox = parent.getBoundingClientRect();

    var vw = pBox.width,
        vh = pBox.height;
    var width = eBox.width,
        height = eBox.height;


    var top = eBox.top - pBox.top;
    var left = eBox.left - pBox.left;
    var right = eBox.right - pBox.left;
    var bottom = eBox.bottom - pBox.top;

    var midX = left + width / 2;
    var midY = top + height / 2;

    var vertex = {
        tl: { x: left, y: top },
        tr: { x: right, y: top },
        br: { x: right, y: bottom },
        bl: { x: left, y: bottom }
    };

    return {
        width: width,
        height: height,
        margin: {
            top: {
                placement: 'top',
                size: top,
                start: vertex.tl,
                mid: { x: midX, y: top },
                end: vertex.tr
            },
            bottom: {
                placement: 'bottom',
                size: vh - bottom,
                start: vertex.bl,
                mid: { x: midX, y: bottom },
                end: vertex.br
            },
            left: {
                placement: 'left',
                size: left,
                start: vertex.tl,
                mid: { x: left, y: midY },
                end: vertex.bl
            },
            right: {
                placement: 'right',
                size: vw - right,
                start: vertex.tr,
                mid: { x: right, y: midY },
                end: vertex.br
            }
        }
    };
}

function computePlacementInfo(ref, container, target, limitQueue, offset) {
    if (!ref || !target) {
        return;
    }
    var placementQueue = limitQueue && limitQueue.length ? limitQueue : DEFAULT_PLACEMENT_QUEUE;

    var _getBoxMargin = getBoxMargin(ref, container),
        ew = _getBoxMargin.width,
        eh = _getBoxMargin.height,
        margin = _getBoxMargin.margin;

    var _target$getBoundingCl = target.getBoundingClientRect(),
        tw = _target$getBoundingCl.width,
        th = _target$getBoundingCl.height;

    var dw = (tw - ew) / 2;
    var dh = (th - eh) / 2;

    var queueLen = placementQueue.length;
    var processedQueue = Object.keys(margin).map(function (key) {
        var placementItem = margin[key];

        var index = placementQueue.indexOf(placementItem.placement);
        placementItem.weight = index > -1 ? MAX - index : MAX - queueLen;

        var verSingleBiasCheck = ~VERTICAL.indexOf(placementItem.placement) && placementItem.size > th + offset;

        var verFullBiasCheck = verSingleBiasCheck && margin.left.size > dw && margin.right.size > dw;

        var horSingleBiasCheck = HORIZONTAL.indexOf(placementItem.placement) > -1 && placementItem.size > tw + offset;

        var horFullBiasCheck = horSingleBiasCheck && margin.top.size > dh && margin.bottom.size > dh;

        placementItem.dVer = margin.top.size - margin.bottom.size;

        placementItem.dHor = margin.left.size - margin.right.size;
        placementItem.mod = 'edge';

        if (verFullBiasCheck || horFullBiasCheck) {
            placementItem.mod = 'mid';
            placementItem.weight += 3 + placementItem.weight / MAX;
            return placementItem;
        }
        if (verSingleBiasCheck || horSingleBiasCheck) {
            placementItem.weight += 2 + placementItem.weight / MAX;
        }
        return placementItem;
    });
    return Object.assign({ ew: ew, eh: eh, tw: tw, th: th, dw: dw, dh: dh }, getBestPlacement(processedQueue));
}

function computeCoordinateBaseMid(placementInfo, offset) {
    var placement = placementInfo.placement,
        mid = placementInfo.mid,
        tw = placementInfo.tw,
        th = placementInfo.th;

    switch (placement) {
        case 'top':
            return {
                placement: 'top-mid',
                x: mid.x - tw / 2,
                y: mid.y - th - offset
            };
        case 'bottom':
            return {
                placement: 'bottom-mid',
                x: mid.x - tw / 2,
                y: mid.y + offset
            };
        case 'left':
            return {
                placement: 'left-mid',
                x: mid.x - tw - offset,
                y: mid.y - th / 2
            };
        case 'right':
            return {
                placement: 'right-mid',
                x: mid.x + offset,
                y: mid.y - th / 2
            };
        default:
    }
}

function computeArrowPos(placement, offset, size) {
    var start = offset + 'px';
    var end = offset - size * 2 + 'px';
    var posMap = {
        'top-start': { top: '100%', left: start },
        'top-mid': { top: '100%', left: '50%' },
        'top-end': { top: '100%', right: end },

        'bottom-start': { top: '0', left: start },
        'bottom-mid': { top: '0', left: '50%' },
        'bottom-end': { top: '0', right: end },

        'left-start': { top: start, left: '100%' },
        'left-mid': { top: '50%', left: '100%' },
        'left-end': { bottom: end, left: '100%' },

        'right-start': { top: start, left: '0' },
        'right-mid': { top: '50%', left: '0' },
        'right-end': { bottom: end, left: '0' }
    };
    return posMap[placement];
}

function computeCoordinateBaseEdge(placementInfo, offset) {
    var placement = placementInfo.placement,
        start = placementInfo.start,
        end = placementInfo.end,
        dHor = placementInfo.dHor,
        dVer = placementInfo.dVer,
        tw = placementInfo.tw,
        th = placementInfo.th,
        ew = placementInfo.ew,
        eh = placementInfo.eh;

    var nearRight = dHor > 0;
    var nearBottom = dVer > 0;
    switch (placement) {
        case 'top':
            return {
                placement: nearRight ? 'top-end' : 'top-start',
                x: nearRight ? end.x - tw : start.x,
                y: start.y - th - offset,
                arrowsOffset: ew / 2
            };
        case 'bottom':
            return {
                placement: nearRight ? 'bottom-end' : 'bottom-start',
                x: nearRight ? end.x - tw : start.x,
                y: end.y + offset,
                arrowsOffset: ew / 2
            };
        case 'left':
            return {
                placement: nearBottom ? 'left-end' : 'left-start',
                x: start.x - tw - offset,
                y: nearBottom ? end.y - th : start.y,
                arrowsOffset: eh / 2
            };
        case 'right':
            return {
                placement: nearBottom ? 'right-end' : 'right-start',
                x: end.x + offset,
                y: nearBottom ? end.y - th : start.y,
                arrowsOffset: eh / 2
            };
        default:
    }
}

var requestAnimationFrame$1 = window.requestAnimationFrame || window.webkitRequestAnimationFrame || window.mozRequestAnimationFrame || window.oRequestAnimationFrame || window.msRequestAnimationFrame || function (callback) {
    window.setTimeout(callback, 1000 / 60);
};
var cancelAnimationFrame$1 = window.cancelAnimationFrame || window.webkitCancelAnimationFrame || window.mozCancelAnimationFrame || window.oCancelAnimationFrame || window.msCancelAnimationFrame || function (id) {
    window.clearTimeout(id);
};

var bkSideslider = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('transition', { attrs: { "name": "slide" } }, [_c('article', { directives: [{ name: "show", rawName: "v-show", value: _vm.isShow, expression: "isShow" }], staticClass: "bk-sideslider", on: { "click": function click($event) {
                    if ($event.target !== $event.currentTarget) {
                        return null;
                    }return _vm.handleQuickClose($event);
                } } }, [_c('section', { staticClass: "bk-sideslider-wrapper", class: [{ left: _vm.direction === 'left', right: _vm.direction === 'right' }], style: { width: _vm.width + 'px' } }, [_c('div', { staticClass: "bk-sideslider-header" }, [_c('div', { staticClass: "bk-sideslider-closer", style: { float: _vm.calcDirection }, on: { "click": _vm.handleClose } }, [_c('i', { staticClass: "bk-icon", class: 'icon-angle-' + _vm.direction })]), _vm._v(" "), _c('div', { staticClass: "bk-sideslider-title", style: { padding: this.calcDirection === 'left' ? '0 0 0 50px' : '0 50px 0 0', 'text-align': this.calcDirection } }, [_vm._v(" " + _vm._s(_vm.title || '标题') + " ")])]), _vm._v(" "), _c('div', { staticClass: "bk-sideslider-content" }, [_vm._t("content")], 2)])])]);
    }, staticRenderFns: [],
    name: 'bk-sideslider',
    props: {
        isShow: {
            type: Boolean,
            default: false
        },
        title: {
            type: String,
            default: ''
        },
        quickClose: {
            type: Boolean,
            default: false
        },
        width: {
            default: 400
        },
        direction: {
            type: String,
            default: 'right',
            validator: function validator(value) {
                return ['left', 'right'].indexOf(value) > -1;
            }
        }
    },
    watch: {
        isShow: function isShow(val) {
            var _this = this;

            var root = document.documentElement;
            if (val) {
                addClass(root, 'bk-sideslider-show');
                if (this.isScrollY()) {
                    addClass(root, 'has-sideslider-padding');
                }
                setTimeout(function () {
                    _this.$emit('shown');
                }, 200);
            } else {
                removeClass(root, 'bk-sideslider-show has-sideslider-padding');
                setTimeout(function () {
                    _this.$emit('hidden');
                }, 200);
            }
        }
    },
    computed: {
        calcDirection: function calcDirection() {
            return this.direction === 'left' ? 'right' : 'left';
        }
    },
    methods: {
        isScrollY: function isScrollY() {
            return document.documentElement.offsetHeight > document.documentElement.clientHeight;
        },
        show: function show() {
            var root = document.documentElement;
            addClass(root, 'bk-sideslider-show');
            this.isShow = true;
        },
        hide: function hide() {
            var root = document.querySelector('html');
            removeClass(root, 'bk-sideslider-show');
            this.isShow = false;
        },
        handleClose: function handleClose() {
            this.$emit('update:isShow', false);
        },
        handleQuickClose: function handleQuickClose() {
            if (this.quickClose) {
                this.handleClose();
            }
        }
    },
    destroyed: function destroyed() {
        var root = document.querySelector('html');
        removeClass(root, 'bk-sideslider-show');
    }
};

bkSideslider.install = function (Vue$$1) {
    Vue$$1.component(bkSideslider.name, bkSideslider);
};

var Switcher = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { class: _vm.classObject }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.enabled, expression: "enabled" }], attrs: { "type": "checkbox", "disabled": _vm.disabled }, domProps: { "checked": Array.isArray(_vm.enabled) ? _vm._i(_vm.enabled, null) > -1 : _vm.enabled }, on: { "change": function change($event) {
                    var $$a = _vm.enabled,
                        $$el = $event.target,
                        $$c = $$el.checked ? true : false;if (Array.isArray($$a)) {
                        var $$v = null,
                            $$i = _vm._i($$a, $$v);if ($$el.checked) {
                            $$i < 0 && (_vm.enabled = $$a.concat([$$v]));
                        } else {
                            $$i > -1 && (_vm.enabled = $$a.slice(0, $$i).concat($$a.slice($$i + 1)));
                        }
                    } else {
                        _vm.enabled = $$c;
                    }
                } } }), _vm._v(" "), _c('label', { directives: [{ name: "show", rawName: "v-show", value: _vm.showText, expression: "showText" }], staticClass: "switcher-label" }, [_c('span', { staticClass: "switcher-text on-text" }, [_vm._v(_vm._s(_vm.onText))]), _vm._v(" "), _c('span', { staticClass: "switcher-text off-text" }, [_vm._v(_vm._s(_vm.offText))])])]);
    }, staticRenderFns: [],
    name: 'bk-switcher',
    props: {
        disabled: {
            type: Boolean,
            default: false
        },
        showText: {
            type: Boolean,
            default: true
        },
        selected: {
            type: Boolean,
            default: false
        },
        onText: {
            type: String,
            default: 'ON'
        },
        offText: {
            type: String,
            default: 'OFF'
        },
        isOutline: {
            type: Boolean,
            default: false
        },
        isSquare: {
            type: Boolean,
            default: false
        },
        size: {
            type: String,
            default: 'normal',
            validator: function validator(value) {
                return ['normal', 'small'].indexOf(value) > -1;
            }
        }
    },
    data: function data() {
        return {
            label: this.selected ? this.onText : this.offText,
            enabled: !!this.selected
        };
    },

    watch: {
        enabled: function enabled(val) {
            this.label = this.enabled ? this.onText : this.offText;
            this.$emit('change', val);
        },
        selected: function selected(val) {
            this.enabled = !!val;
        }
    },
    computed: {
        classObject: function classObject() {
            var enabled = this.enabled,
                disabled = this.disabled,
                size = this.size,
                showText = this.showText,
                isOutline = this.isOutline,
                isSquare = this.isSquare;

            var style = {
                'bk-switcher': true,
                'bk-switcher-outline': isOutline,
                'bk-switcher-square': isSquare,
                'show-label': true,
                'is-disabled': disabled,
                'is-checked': enabled,
                'is-unchecked': !enabled
            };
            if (size) {
                var sizeStr = 'bk-switcher-' + size;
                style[sizeStr] = true;
            }
            return style;
        }
    }
};

Switcher.install = function (Vue$$1) {
    Vue$$1.component(Switcher.name, Switcher);
};

var Render = {
    name: 'render',
    functional: true,
    props: {
        node: Object,
        displayKey: String,
        tpl: Function
    },
    render: function render(h, ct) {
        var parentClass = 'bk-selector-node';
        var textClass = 'text';
        if (ct.props.tpl) {
            return ct.props.tpl(ct.props.node, ct);
        }
        return h(
            'div',
            { 'class': parentClass },
            [h('span', {
                domProps: {
                    'innerHTML': ct.props.node[ct.props.displayKey]
                },
                'class': textClass })]
        );
    }
};

var TagInpute = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-tag-selector", on: { "click": function click($event) {
                    _vm.foucusInputer($event);
                } } }, [_c('div', { class: ['bk-tag-input', { 'active': _vm.isEdit, 'disabled': _vm.disabled }] }, [_c('ul', { ref: "tagList", staticClass: "tag-list" }, [_vm._l(_vm.localTagList, function (tag, index) {
            return _c('li', { key: index, staticClass: "key-node", on: { "click": function click($event) {
                        $event.stopPropagation();_vm.selectTag($event, tag);
                    } } }, [_c('span', { staticClass: "tag" }, [_vm._v(_vm._s(tag[_vm.displayKey]))]), _vm._v(" "), !_vm.disabled && _vm.hasDeleteIcon ? _c('a', { staticClass: "remove-key", attrs: { "href": "javascript:void(0)" }, on: { "click": function click($event) {
                        $event.stopPropagation();_vm.removeTag($event, tag, index);
                    } } }, [_c('i', { staticClass: "bk-icon icon-close" })]) : _vm._e()]);
        }), _vm._v(" "), _c('li', { ref: "staffInput", attrs: { "id": "staffInput" } }, [!_vm.disabled ? _c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.curInputValue, expression: "curInputValue" }], ref: "input", staticClass: "input", attrs: { "type": "text" }, domProps: { "value": _vm.curInputValue }, on: { "input": [function ($event) {
                    if ($event.target.composing) {
                        return;
                    }_vm.curInputValue = $event.target.value;
                }, _vm.input], "focus": _vm.focusInput, "paste": _vm.paste, "blur": _vm.blurHandler, "keydown": _vm.keyupHandler } }) : _vm._e()])], 2), _vm._v(" "), _c('p', { directives: [{ name: "show", rawName: "v-show", value: !_vm.isEdit && !_vm.localTagList.length && !_vm.curInputValue.length, expression: "!isEdit && !localTagList.length && !curInputValue.length" }], staticClass: "placeholder" }, [_vm._v(_vm._s(_vm.placeholder))])]), _vm._v(" "), _c('transition', { attrs: { "name": "optionList" } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.showList && _vm.renderList.length, expression: "showList && renderList.length" }], staticClass: "bk-selector-list" }, [_c('ul', { ref: "selectorList", staticClass: "outside-ul", style: { 'max-height': _vm.contentMaxHeight + 'px' } }, _vm._l(_vm.renderList, function (data, index) {
            return _c('li', { key: index, staticClass: "bk-selector-list-item", class: _vm.activeClass(index), on: { "click": function click($event) {
                        _vm.setValTab(data, 'select');
                    } } }, [_c('Render', { attrs: { "node": data, "displayKey": _vm.displayKey, "tpl": _vm.tpl } })], 1);
        }))])])], 1);
    }, staticRenderFns: [],
    name: 'bk-tag-input',
    components: { Render: Render },
    props: {
        placeholder: {
            type: String,
            default: '请输入并按Enter结束'
        },
        value: {
            type: Array,
            default: function _default() {
                return [];
            }
        },

        disabled: {
            type: Boolean,
            default: false
        },
        hasDeleteIcon: {
            type: Boolean,
            default: false
        },
        separator: {
            type: String,
            default: ''
        },
        maxData: {
            type: Number,
            default: -1
        },
        maxResult: {
            type: Number,
            default: 5
        },
        isBlurTrigger: {
            type: Boolean,
            default: true
        },
        saveKey: {
            type: String,
            default: 'id'
        },
        displayKey: {
            type: String,
            default: 'name'
        },
        searchKey: {
            type: String,
            default: 'name'
        },
        list: {
            type: Array,
            default: []
        },
        contentMaxHeight: {
            type: Number,
            default: 300
        },
        allowCreate: {
            type: Boolean,
            default: false
        },
        tpl: Function,
        pasteFn: Function
    },
    data: function data() {
        return {
            curInputValue: '',
            cacheVal: '',
            timer: 0,
            focusList: this.allowCreate ? -1 : 0,
            isEdit: false,
            showList: false,
            isCanRemoveTag: false,
            tagList: [],
            localTagList: [],
            renderList: [],
            initData: []
        };
    },
    created: function created() {
        this.getData();
    },

    watch: {
        curInputValue: function curInputValue(newVal, oldVal) {
            var _this = this;

            if (newVal !== '' && this.renderList.length) {
                this.showList = true;
            } else {
                setTimeout(function () {
                    _this.showList = false;
                }, 100);
            }
        },
        showList: function showList(val) {
            var _this2 = this;

            if (val) {
                this.$nextTick(function () {
                    _this2.$refs.selectorList.scrollTop = 0;
                });
            }
        },
        list: function list(val) {
            if (val) {
                this.getData();
            }
        }
    },
    methods: {
        getCharLength: function getCharLength(str) {
            var len = str.length;
            var bitLen = 0;

            for (var i = 0; i < len; i++) {
                if ((str.charCodeAt(i) & 0xff00) !== 0) {
                    bitLen++;
                }
                bitLen++;
            }

            return bitLen;
        },
        filterData: function filterData(val) {
            var _this3 = this;

            this.renderList = [].concat(toConsumableArray(this.initData.filter(function (item) {
                return item[_this3.searchKey].indexOf(val) > -1;
            })));
            if (this.renderList.length > this.maxResult) {
                this.renderList = [].concat(toConsumableArray(this.renderList.slice(0, this.maxResult)));
            }
        },
        activeClass: function activeClass(i) {
            return {
                'bk-selector-selected': i === this.focusList
            };
        },
        updateData: function updateData() {
            var _localTagList,
                _this4 = this;

            var newList = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : [];

            (_localTagList = this.localTagList).splice.apply(_localTagList, [0, this.localTagList.length].concat([]));
            this.curInputValue = '';
            this.initData = [].concat(toConsumableArray(this.list));
            if (newList.length) {
                newList.map(function (tag) {
                    _this4.initData.filter(function (val) {
                        if (tag === val[_this4.saveKey]) {
                            _this4.localTagList.push(val);
                            _this4.tagList.push(val[_this4.saveKey]);
                        } else if (_this4.allowCreate && !tag.includes(tag)) {
                            var temp = {};

                            temp[_this4.saveKey] = tag;
                            temp[_this4.displayKey] = tag;
                            _this4.localTagList.push(temp);
                            _this4.tagList.push(tag);
                        }
                    });
                });
                newList.forEach(function (tag) {
                    _this4.initData = _this4.initData.filter(function (val) {
                        return !tag.includes(val[_this4.saveKey]);
                    });
                });
            }
        },
        getSiteInfo: function getSiteInfo() {
            var res = {
                index: 0,
                temp: []
            };
            var nodes = this.$refs.tagList.childNodes;

            for (var i = 0; i < nodes.length; i++) {
                var node = nodes[i];

                if (!(node.nodeType === 3 && !/\S/.test(node.nodeValue))) {
                    res.temp.push(node);
                }
            }

            Object.keys(res.temp).forEach(function (key) {
                if (res.temp[key].id === 'staffInput') res.index = key;
            });

            return res;
        },
        getData: function getData() {
            var _this5 = this;

            this.initData = [].concat(toConsumableArray(this.list));

            if (this.value.length) {
                this.value.map(function (tag) {
                    _this5.initData.filter(function (val) {
                        if (tag === val[_this5.saveKey]) {
                            _this5.localTagList.push(val);
                            _this5.tagList.push(val[_this5.saveKey]);
                        } else if (_this5.allowCreate && !tag.includes(tag)) {
                            var temp = {};

                            temp[_this5.saveKey] = tag;
                            temp[_this5.displayKey] = tag;
                            _this5.localTagList.push(temp);
                            _this5.tagList.push(tag);
                        }
                    });
                });

                this.value.forEach(function (tag) {
                    _this5.initData = _this5.initData.filter(function (val) {
                        return !tag.includes(val[_this5.saveKey]);
                    });
                });
            }
        },
        selectTag: function selectTag(event, tag) {
            if (this.disabled) return;

            var domLen = event.target.parentNode.offsetWidth;
            var resSite = (event.target.parentNode.offsetWidth - 20) / 2;

            if (event.offsetX > resSite) {
                this.insertAfter(this.$refs.staffInput, event.target.parentNode);
            } else {
                this.$refs.tagList.insertBefore(this.$refs.staffInput, event.target.parentNode);
            }

            this.foucusInputer(event);
            this.$refs.input.focus();
            this.$refs.input.style.width = 12 + 'px';
        },
        input: function input(event) {
            if (this.maxData === -1 || this.maxData > this.tagList.length) {
                var value = event.target.value;

                var charLen = this.getCharLength(value);

                this.cacheVal = value;
                if (charLen) {
                    this.isCanRemoveTag = false;
                    this.filterData(value);
                    this.$refs.input.style.width = charLen * 12 + 'px';
                } else {
                    this.isCanRemoveTag = true;
                }
            } else {
                this.blurHandler();
                this.curInputValue = '';
                this.showList = false;
            }

            this.isEdit = true;

            this.focusList = this.allowCreate ? -1 : 0;
        },
        focusInput: function focusInput() {
            this.isCanRemoveTag = true;
        },
        paste: function paste(event) {
            var _this6 = this;

            event.preventDefault();

            var value = event.clipboardData.getData('text');
            var valArr = this.pasteFn ? this.pasteFn(value) : this.defaultPasteFn(value);
            var tags = [];

            valArr.map(function (val) {
                return tags.push(val[_this6.saveKey]);
            });

            if (tags.length) {
                var nodes = this.$refs.tagList.childNodes;
                var result = this.getSiteInfo(nodes);
                var localTags = [];
                var localInitDara = [];

                this.initData.map(function (data) {
                    localInitDara.push(data[_this6.saveKey]);
                });
                tags = tags.filter(function (tag) {
                    return tag && tag.trim() && !_this6.tagList.includes(tag) && localInitDara.includes(tag);
                });

                if (this.maxData !== -1) {
                    if (this.tagList.length < this.maxData) {
                        var differ = this.maxData - this.tagList.length;
                        if (tags.length > differ) {
                            tags = [].concat(toConsumableArray(tags.slice(0, differ)));
                        }
                    } else {
                        tags = [];
                    }
                }

                tags.map(function (tag) {
                    var temp = {};
                    temp[_this6.saveKey] = tag;
                    temp[_this6.displayKey] = tag;
                    localTags.push(temp);
                });

                if (tags.length) {
                    this.tagList = [].concat(toConsumableArray(this.tagList.slice(0, result.index)), toConsumableArray(tags), toConsumableArray(this.tagList.slice(result.index, this.tagList.length)));
                    this.localTagList = [].concat(toConsumableArray(this.localTagList.slice(0, result.index)), localTags, toConsumableArray(this.localTagList.slice(result.index, this.localTagList.length)));

                    var site = nodes[parseInt(result.index) + 1];
                    this.insertAfter(this.$refs.staffInput, site);
                    this.$refs.input.focus();
                    this.$refs.input.style.width = 12 + 'px';
                    tags.map(function (tag) {
                        _this6.initData = _this6.initData.filter(function (val) {
                            return !tag.includes(val[_this6.saveKey]);
                        });
                    });

                    this.handlerChange('select');
                }
            }
        },
        defaultPasteFn: function defaultPasteFn(val) {
            var _this7 = this;

            var target = [];
            var textArr = val.split(';');

            textArr.map(function (item) {
                if (item.match(/^[a-zA-Z][a-zA-Z_]+/g)) {
                    var finalItem = item.match(/^[a-zA-Z][a-zA-Z_]+/g).join('');
                    var temp = {};
                    temp[_this7.saveKey] = finalItem;
                    temp[_this7.displayKey] = finalItem;
                    target.push(temp);
                }
            });
            return target;
        },
        updateScrollTop: function updateScrollTop() {
            var _this8 = this;

            var panelObj = this.$el.querySelector('.bk-selector-list .outside-ul');
            var panelInfo = {
                height: panelObj.clientHeight,
                yAxios: panelObj.getBoundingClientRect().y
            };

            this.$nextTick(function () {
                var activeObj = _this8.$el.querySelector('.bk-selector-list .bk-selector-selected');
                var activeInfo = {
                    height: activeObj.clientHeight,
                    yAxios: activeObj.getBoundingClientRect().y
                };

                if (activeInfo.yAxios < panelInfo.yAxios) {
                    var currentScTop = panelObj.scrollTop;
                    panelObj.scrollTop = currentScTop - (panelInfo.yAxios - activeInfo.yAxios);
                }

                var distanceToBottom = activeInfo.yAxios + activeInfo.height - panelInfo.yAxios;

                if (distanceToBottom > panelInfo.height) {
                    var _currentScTop = panelObj.scrollTop;
                    panelObj.scrollTop = _currentScTop + distanceToBottom - panelInfo.height;
                }
            });
        },
        keyupHandler: function keyupHandler(event) {
            var target = void 0;
            var val = event.target.value;
            var valLen = this.getCharLength(val);
            var result = this.getSiteInfo();
            var nodes = this.$refs.tagList.childNodes;

            switch (event.code) {
                case 'ArrowUp':
                    event.preventDefault();
                    this.focusList--;
                    this.focusList = this.focusList < 0 ? -1 : this.focusList;
                    if (this.focusList === -1) {
                        this.focusList = this.renderList.length - 1;
                    }
                    this.updateScrollTop();
                    break;
                case 'ArrowDown':
                    event.preventDefault();
                    this.focusList++;
                    this.focusList = this.focusList > this.renderList.length - 1 ? this.renderList.length : this.focusList;
                    if (this.focusList === this.renderList.length) {
                        this.focusList = 0;
                    }
                    this.updateScrollTop();
                    break;
                case 'ArrowLeft':
                    this.isEdit = true;
                    if (!valLen) {
                        if (parseInt(result.index) > 1) {
                            var leftsite = nodes[parseInt(result.index) - 2];
                            this.insertAfter(this.$refs.staffInput, leftsite);
                            this.$refs.input.value = '';
                            this.$refs.input.style.width = 12 + 'px';
                        } else {
                            var _nodes = this.$refs.tagList.childNodes;
                            this.$refs.tagList.insertBefore(this.$refs.staffInput, _nodes[0]);
                        }
                        this.$refs.input.focus();
                    }
                    break;
                case 'ArrowRight':
                    this.isEdit = true;
                    if (!valLen) {
                        var rightsite = nodes[parseInt(result.index) + 1];
                        this.insertAfter(this.$refs.staffInput, rightsite);
                        this.$refs.input.focus();
                    }
                    break;
                case 'Enter':
                case 'NumpadEnter':
                    if (!this.allowCreate && this.showList || this.allowCreate && this.focusList >= 0 && this.showList) {
                        this.setValTab(this.renderList[this.focusList], 'select');
                        this.showList = false;
                    } else if (this.allowCreate) {
                        var tag = this.curInputValue;
                        this.setValTab(tag, 'custom');
                    }
                    this.cacheVal = '';
                    break;
                case 'Backspace':
                    if (parseInt(result.index) !== 0 && !this.curInputValue.length) {
                        target = this.localTagList[result.index - 1];
                        this.backspaceHandler(result.index, target);
                    }
                    break;
                default:
                    break;
            }
        },
        setValTab: function setValTab(item, type) {
            var _this9 = this;

            var nodes = this.$refs.tagList.childNodes;
            var result = this.getSiteInfo(nodes);
            var isSelected = false;
            var tags = [];
            var newVal = void 0;

            if (type === 'custom') {
                if (this.separator) {
                    var localTags = [];

                    tags = item.split(this.separator);
                    tags = tags.filter(function (tag) {
                        return tag && tag.trim() && !_this9.tagList.includes(tag);
                    });
                    tags = [].concat(toConsumableArray(new Set(tags)));
                    tags.map(function (tag) {
                        var temp = {};
                        temp[_this9.saveKey] = tag;
                        temp[_this9.displayKey] = tag;
                        localTags.push(temp);
                    });

                    if (tags.length) {
                        this.tagList = [].concat(toConsumableArray(this.tagList.slice(0, result.index)), toConsumableArray(tags), toConsumableArray(this.tagList.slice(result.index, this.tagList.length)));
                        this.localTagList = [].concat(toConsumableArray(this.localTagList.slice(0, result.index)), localTags, toConsumableArray(this.localTagList.slice(result.index, this.localTagList.length)));
                        isSelected = true;
                    }
                } else {
                    if ((typeof item === 'undefined' ? 'undefined' : _typeof(item)) === 'object') {
                        newVal = item[this.saveKey];
                        if (newVal && !this.tagList.includes(newVal)) {
                            newVal = newVal.replace(/\s+/g, '');

                            if (newVal.length) {
                                this.tagList = [].concat(toConsumableArray(this.tagList.slice(0, result.index)), [newVal], toConsumableArray(this.tagList.slice(result.index, this.tagList.length)));
                                this.localTagList = [].concat(toConsumableArray(this.localTagList.slice(0, result.index)), [item], toConsumableArray(this.localTagList.slice(result.index, this.localTagList.length)));
                                isSelected = true;
                            }
                        }
                    } else {
                        var localItem = {};
                        newVal = item.trim();
                        localItem[this.saveKey] = newVal;
                        localItem[this.displayKey] = newVal;

                        if (newVal.length && !this.tagList.includes(newVal)) {
                            this.tagList = [].concat(toConsumableArray(this.tagList.slice(0, result.index)), [newVal], toConsumableArray(this.tagList.slice(result.index, this.tagList.length)));
                            this.localTagList = [].concat(toConsumableArray(this.localTagList.slice(0, result.index)), [localItem], toConsumableArray(this.localTagList.slice(result.index, this.localTagList.length)));
                            isSelected = true;
                        }
                    }
                }
            } else {
                newVal = item[this.saveKey];
                this.tagList = [].concat(toConsumableArray(this.tagList.slice(0, result.index)), [newVal], toConsumableArray(this.tagList.slice(result.index, this.tagList.length)));
                this.localTagList = [].concat(toConsumableArray(this.localTagList.slice(0, result.index)), [item], toConsumableArray(this.localTagList.slice(result.index, this.localTagList.length)));
                isSelected = true;
            }

            if (isSelected) {
                var site = nodes[parseInt(result.index) + 1];
                this.insertAfter(this.$refs.staffInput, site);
                this.$refs.input.focus();
                this.$refs.input.style.width = 12 + 'px';
                if (this.allowCreate && this.separator) {
                    tags.map(function (tag) {
                        _this9.initData = _this9.initData.filter(function (val) {
                            return !tag.includes(val[_this9.saveKey]);
                        });
                    });
                } else {
                    this.initData = this.initData.filter(function (val) {
                        return !newVal.includes(val[_this9.saveKey]);
                    });
                }
            }

            this.handlerChange('select');
            this.clearInput();
        },
        backspaceHandler: function backspaceHandler(index, target) {
            var _this10 = this;

            if (!this.curInputValue) {
                if (this.isCanRemoveTag) {
                    this.tagList = [].concat(toConsumableArray(this.tagList.slice(0, index - 1)), toConsumableArray(this.tagList.slice(index, this.tagList.length)));
                    this.localTagList = [].concat(toConsumableArray(this.localTagList.slice(0, index - 1)), toConsumableArray(this.localTagList.slice(index, this.localTagList.length)));

                    var nodes = this.$refs.tagList.childNodes;
                    var result = this.getSiteInfo(nodes);
                    var key = parseInt(result.index) === 1 ? 1 : parseInt(result.index) - 2;
                    var site = nodes[key];

                    this.insertAfter(this.$refs.staffInput, site);
                    this.$refs.input.focus();
                    var isExistInit = this.list.some(function (item) {
                        return item === target[_this10.saveKey];
                    });
                    if (this.allowCreate && isExistInit || !this.allowCreate) {
                        this.initData.push(target);
                    }

                    this.$refs.input.style.width = 12 + 'px';
                    this.handlerChange('remove');
                }
                this.isCanRemoveTag = true;
            }
        },
        removeTag: function removeTag(event, data, index) {
            var _this11 = this;

            this.tagList.splice(index, 1);
            this.localTagList.splice(index, 1);

            var isExistInit = this.list.some(function (item) {
                return item === data[_this11.saveKey];
            });

            if (this.allowCreate && isExistInit || !this.allowCreate) {
                this.initData.push(data);
            }

            this.$refs.input.style.width = 12 + 'px';
            this.resetInput();
            this.handlerChange('remove');
        },
        handlerChange: function handlerChange(type) {
            this.$emit('input', this.tagList);
            this.$emit('change', this.tagList);
            this.$emit(type);
            this.$emit('update:tags', this.tagList);
        },
        clearInput: function clearInput() {
            this.curInputValue = '';
        },
        blurHandler: function blurHandler() {
            var _this12 = this;

            this.resetInput();
            this.timer = setTimeout(function () {
                _this12.clearInput();
                _this12.isEdit = false;
            }, 300);
        },
        foucusInputer: function foucusInputer(event) {
            var _this13 = this;

            if (this.disabled) return;

            if (event.target.className === 'bk-tag-input active' || event.target.className === 'tag-list') {
                setTimeout(function () {
                    _this13.curInputValue = _this13.cacheVal;
                }, 100);
            } else {
                this.cacheVal = '';
            }

            clearTimeout(this.timer);
            this.isEdit = true;
            this.$nextTick(function () {
                _this13.$el.querySelector('.input').focus();
            });
        },
        insertAfter: function insertAfter(newElement, targetElement) {
            var parent = targetElement.parentNode;

            if (parent.lastChild === targetElement) {
                parent.appendChild(newElement);
            } else {
                parent.insertBefore(newElement, targetElement.nextSibling);
            }
        },
        resetInput: function resetInput() {
            var nodes = this.$refs.tagList.childNodes;
            var result = this.getSiteInfo(nodes);

            if (result.index !== result.temp.length) {
                this.clearInput();
                var site = nodes[nodes.length - 1];

                this.insertAfter(this.$refs.staffInput, site);
            }
        }
    }
};

TagInpute.install = function (Vue$$1) {
    Vue$$1.component(TagInpute.name, TagInpute);
};

var defaultLang = {
    lang: 'zh-CN',
    datePicker: {
        test: '我们{vari}hello {ccc}!@#$%^&&*({})',

        selectDate: '选择日期',

        topBarFormatView: '{yyyy}年{mm}月',
        weekdays: {
            sun: '日',
            mon: '一',
            tue: '二',
            wed: '三',
            thu: '四',
            fri: '五',
            sat: '六'
        },
        today: '今天',

        test1: '{mm} {yyyy}',
        test3: '{mmmm} {yyyy}',
        test2: '{yyyy}年{mm}月'
    },
    dateRange: {
        selectDate: '选择日期',
        datePicker: {
            topBarFormatView: '{yyyy}年{mm}月',
            weekdays: {
                sun: '日',
                mon: '一',
                tue: '二',
                wed: '三',
                thu: '四',
                fri: '五',
                sat: '六'
            },
            today: '今天'
        },
        yestoday: '昨天',
        lastweek: '最近一周',
        lastmonth: '最近一个月',
        last3months: '最近三个月',
        ok: '确定',
        clear: '清空'
    },
    dialog: {
        title: '这是标题',
        content: '这是内容',
        ok: '确定',
        cancel: '取消'
    },
    selector: {
        pleaseselect: '请选择',
        emptyText: '暂无数据',
        searchEmptyText: '无匹配数据'
    },
    infobox: {
        title: '这是标题',
        ok: '确定',
        cancel: '取消',
        pleasewait: '请稍等',
        success: '操作成功',
        continue: '继续',
        failure: '操作失败',
        closeafter3s: '此窗口3s后关闭',
        riskoperation: '此操作存在风险'
    },
    message: {
        close: '关闭'
    },
    sideslider: {
        title: '标题'
    },
    steps: {
        step1: '步骤1',
        step2: '步骤2',
        step3: '步骤3'
    },
    uploadFile: {
        drag: '拖拽到此处上传或',
        click: '点击上传',
        uploadDone: '上传完毕'
    }
};

var canUseSymbol = typeof Symbol === 'function' && Symbol.for;

var REACT_ELEMENT_TYPE = canUseSymbol ? Symbol.for('react.element') : 0xeac7;

function isReactElement(value) {
    return value.$$typeof === REACT_ELEMENT_TYPE;
}

function isNonNullObject(value) {
    return !!value && (typeof value === 'undefined' ? 'undefined' : _typeof(value)) === 'object';
}

function isSpecial(value) {
    var stringValue = Object.prototype.toString.call(value);

    return stringValue === '[object RegExp]' || stringValue === '[object Date]' || isReactElement(value);
}

function defaultIsMergeableObject (value) {
    return isNonNullObject(value) && !isSpecial(value);
}

function emptyTarget(val) {
    return Array.isArray(val) ? [] : {};
}

function cloneUnlessOtherwiseSpecified(value, options) {
    return options.clone !== false && options.isMergeableObject(value) ? deepmerge(emptyTarget(value), value, options) : value;
}

function defaultArrayMerge(target, source, options) {
    return target.concat(source).map(function (element) {
        return cloneUnlessOtherwiseSpecified(element, options);
    });
}

function mergeObject(target, source, options) {
    var destination = {};
    if (options.isMergeableObject(target)) {
        Object.keys(target).forEach(function (key) {
            destination[key] = cloneUnlessOtherwiseSpecified(target[key], options);
        });
    }
    Object.keys(source).forEach(function (key) {
        if (!options.isMergeableObject(source[key]) || !target[key]) {
            destination[key] = cloneUnlessOtherwiseSpecified(source[key], options);
        } else {
            destination[key] = deepmerge(target[key], source[key], options);
        }
    });

    return destination;
}

function deepmerge(target, source, options) {
    options = options || {};
    options.arrayMerge = options.arrayMerge || defaultArrayMerge;
    options.isMergeableObject = options.isMergeableObject || defaultIsMergeableObject;

    var sourceIsArray = Array.isArray(source);
    var targetIsArray = Array.isArray(target);
    var sourceAndTargetTypesMatch = sourceIsArray === targetIsArray;

    if (!sourceAndTargetTypesMatch) {
        return cloneUnlessOtherwiseSpecified(source, options);
    } else if (sourceIsArray) {
        return options.arrayMerge(target, source, options);
    }
    return mergeObject(target, source, options);
}

deepmerge.all = function deepmergeAll(array, options) {
    if (!Array.isArray(array)) {
        throw new Error('first argument should be an array');
    }

    return array.reduce(function (prev, next) {
        return deepmerge(prev, next, options);
    }, {});
};

var curLang = defaultLang;

var merged = false;

var i18nHandler = function i18nHandler() {
    var i18n = Object.getPrototypeOf(this || Vue).$t;

    if (typeof i18n === 'function') {
        if (!merged && !!Vue.locale) {
            merged = true;
            Vue.locale(Vue.config.lang, deepmerge(curLang, Vue.locale(Vue.config.lang) || {}, { clone: true }));
        }
        return i18n.apply(this, arguments);
    }
};

var escape = function escape(str) {
    return String(str).replace(/([.*+?^=!:${}()|[\]\/\\])/g, '\\$1');
};

var t = function t(path, data) {
    var value = i18nHandler.apply(this, arguments);
    if (value !== null && typeof value !== 'undefined') {
        return value;
    }

    var arr = path.split('.');
    var current = curLang;
    var len = arr.length;

    for (var i = 0; i < len; i++) {
        value = current[arr[i]];
        if (i === len - 1) {
            if (data && typeof value === 'string') {
                return value.replace(/\{(?=\w+)/g, '').replace(/(\w+)\}/g, '$1').replace(new RegExp(Object.keys(data).map(escape).join('|'), 'g'), function ($0) {
                    return data[$0];
                });
            }
            return value;
        }
        if (!value) {
            return '';
        }
        current = value;
    }
    return '';
};

var use = function use(l) {
    if (l) {
        curLang = deepmerge(curLang, l);
    }
};

var i18n = function i18n(fn) {
    i18nHandler = fn || i18nHandler;
};

var locale = { use: use, t: t, i18n: i18n };

var locale$1 = {
    methods: {
        t: function t$$1() {
            for (var _len = arguments.length, args = Array(_len), _key = 0; _key < _len; _key++) {
                args[_key] = arguments[_key];
            }

            return t.apply(this, args);
        }
    }
};

var bkDialog = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('transition', { attrs: { "name": "displacement-fade-show" }, on: { "before-enter": _vm.handleBeforeEnter, "after-enter": _vm.handleAfterEnter, "before-leave": _vm.handleBeforeLeave, "after-leave": _vm.handleAfterLeave } }, [_c('section', { directives: [{ name: "show", rawName: "v-show", value: _vm.isShow, expression: "isShow" }], staticClass: "bk-dialog", class: _vm.extCls }, [_c('div', { staticClass: "bk-dialog-wrapper" }, [_c('div', { staticClass: "bk-dialog-position", on: { "click": function click($event) {
                    if ($event.target !== $event.currentTarget) {
                        return null;
                    }return _vm.handleQuickClose($event);
                } } }, [_c('div', { ref: "dialogContent", staticClass: "bk-dialog-style", style: { width: typeof _vm.width === 'String' ? _vm.width : _vm.width + 'px' } }, [_vm.hasHeader || _vm.closeIcon || _vm.draggable ? _c('div', { staticClass: "bk-dialog-tool clearfix", class: { draggable: _vm.draggable }, on: { "mousedown": function mousedown($event) {
                    if (!('button' in $event) && _vm._k($event.keyCode, "left", 37, $event.key, ["Left", "ArrowLeft"])) {
                        return null;
                    }if ('button' in $event && $event.button !== 0) {
                        return null;
                    }_vm.handlerDragStart($event);
                } } }, [_vm._t("tools"), _vm._v(" "), _vm.closeIcon ? _c('i', { staticClass: "bk-dialog-close bk-icon icon-close", on: { "click": function click($event) {
                    $event.stopPropagation();return _vm.handleCancel($event);
                } } }) : _vm._e()], 2) : _vm._e(), _vm._v(" "), _vm.hasHeader ? _c('div', { staticClass: "bk-dialog-header" }, [_vm._t("header", [_c('h3', { staticClass: "bk-dialog-title" }, [_vm._v(_vm._s(_vm.title || _vm.defaultTitle))])])], 2) : _vm._e(), _vm._v(" "), _vm.defaultContent !== false ? _c('div', { staticClass: "bk-dialog-body", style: { padding: _vm.calcPadding } }, [_vm._t("content", [_vm._v(_vm._s(_vm.defaultContent))])], 2) : _vm._e(), _vm._v(" "), _vm.hasFooter ? _c('div', { staticClass: "bk-dialog-footer bk-d-footer", style: { 'margin-top': _vm.content === false ? '36px' : '' } }, [_vm._t("footer", [_c('div', { staticClass: "bk-dialog-outer" }, [_c('button', { staticClass: "bk-dialog-btn bk-dialog-btn-confirm", class: 'bk-btn-' + _vm.theme, attrs: { "type": "button", "name": "confirm" }, on: { "click": _vm.handleConfirm } }, [_vm._v(" " + _vm._s(_vm.confirm ? _vm.confirm : _vm.t('dialog.ok')) + " ")]), _vm._v(" "), _c('button', { staticClass: "bk-dialog-btn bk-dialog-btn-cancel", attrs: { "type": "button", "name": "cancel" }, on: { "click": _vm.handleCancel } }, [_vm._v(" " + _vm._s(_vm.cancel ? _vm.cancel : _vm.t('dialog.cancel')) + " ")])])])], 2) : _vm._e()])])])])]);
    }, staticRenderFns: [],
    name: 'bk-dialog',
    mixins: [locale$1],
    props: {
        isShow: {
            type: Boolean,
            default: false
        },
        width: {
            type: [Number, String],
            default: 400
        },
        title: {
            type: String,
            default: ''
        },
        content: {
            type: String,
            default: ''
        },
        hasHeader: {
            type: Boolean,
            default: true
        },
        draggable: {
            type: Boolean,
            default: false
        },
        extCls: {
            type: String,
            default: ''
        },
        padding: {
            type: [Number, String],
            default: 20
        },
        closeIcon: {
            type: Boolean,
            default: true
        },
        theme: {
            type: String,
            default: 'primary',
            validator: function validator(value) {
                return ['info', 'primary', 'warning', 'success', 'danger'].indexOf(value) > -1;
            }
        },
        confirm: {
            type: String,
            default: ''
        },
        cancel: {
            type: String,
            default: ''
        },
        quickClose: {
            type: Boolean,
            default: true
        },
        hasFooter: {
            type: Boolean,
            default: true
        }
    },
    data: function data() {
        return {
            defaultTitle: this.t('dialog.title'),
            defaultContent: this.t('dialog.content'),
            dragState: {}
        };
    },

    computed: {
        calcPadding: function calcPadding() {
            var type = _typeof(this.padding).toLowerCase();

            return type === 'string' ? this.padding : this.padding + 'px';
        }
    },
    watch: {
        isShow: function isShow(val) {
            var _this = this;

            if (val) {
                addClass(document.body, 'bk-dialog-shown');
            } else {
                setTimeout(function () {
                    removeClass(document.body, 'bk-dialog-shown');
                    if (_this.draggable) {
                        _this.resetDragPostion();
                    }
                }, 200);
            }
        }
    },
    created: function created() {
        if (this.title) {
            this.defaultTitle = this.title;
        }
        if (this.content) {
            this.defaultContent = this.content;
        }
    },

    methods: {
        close: function close() {
            this.$emit('update:isShow', false);
        },
        handleConfirm: function handleConfirm() {
            this.$emit('confirm', this.close);
        },
        handleCancel: function handleCancel() {
            this.$emit('cancel', this.close);
        },
        handleQuickClose: function handleQuickClose() {
            if (this.quickClose) {
                this.close();
            }
        },
        handleBeforeEnter: function handleBeforeEnter() {
            this.$emit('before-transition-enter');
        },
        handleAfterEnter: function handleAfterEnter() {
            this.$emit('after-transition-enter');
        },
        handleBeforeLeave: function handleBeforeLeave() {
            this.$emit('before-transition-leave');
        },
        handleAfterLeave: function handleAfterLeave() {
            this.$emit('after-transition-leave');
        },
        handlerDragStart: function handlerDragStart(event) {
            var _this2 = this;

            if (!this.draggable) return false;
            var $dialogContent = this.$refs.dialogContent;
            var computedStyle = window.getComputedStyle($dialogContent);
            document.onselectstart = function () {
                return false;
            };
            document.ondragstart = function () {
                return false;
            };
            document.body.style.cursor = 'move';
            this.dragState = {
                startX: event.clientX,
                startY: event.clientY,
                contentRect: $dialogContent.getBoundingClientRect(),
                dialogRect: this.$el.getBoundingClientRect(),
                startPosLeft: parseInt(computedStyle.left, 10) || 0,
                startPosTop: parseInt(computedStyle.top, 10) || 0,
                dragging: true,
                animationId: null
            };

            var handleMousemove = function handleMousemove(event) {
                _this2.dragState.animationId = requestAnimationFrame$1(function () {
                    var dragState = _this2.dragState;
                    var contentRect = dragState.contentRect;
                    var dialogRect = dragState.dialogRect;
                    var deltaX = event.clientX - dragState.startX;
                    var deltaY = event.clientY - dragState.startY;
                    deltaX = Math.floor(Math.max(-1 * contentRect.x, Math.min(deltaX, dialogRect.width - contentRect.x - contentRect.width)));
                    deltaY = Math.floor(Math.max(-1 * contentRect.top, Math.min(deltaY, dialogRect.height - contentRect.y - contentRect.height)));
                    $dialogContent.style.left = dragState.startPosLeft + deltaX + 'px';
                    $dialogContent.style.top = dragState.startPosTop + deltaY + 'px';
                });
            };

            var handleMouseup = function handleMouseup(event) {
                event.stopPropagation();
                event.preventDefault();
                cancelAnimationFrame$1(_this2.dragState.animationId);
                _this2.dragState = {};
                document.onselectstart = null;
                document.ondragstart = null;
                document.body.style.cursor = 'default';
                document.removeEventListener('mousemove', handleMousemove);
                document.removeEventListener('mouseup', handleMouseup);
            };

            document.addEventListener('mousemove', handleMousemove);
            document.addEventListener('mouseup', handleMouseup);
        },
        resetDragPostion: function resetDragPostion() {
            var $dialogContent = this.$refs.dialogContent;
            $dialogContent.style.left = 0;
            $dialogContent.style.top = 0;
        }
    }
};

bkDialog.install = function (Vue$$1) {
    Vue$$1.component(bkDialog.name, bkDialog);
};

var bkIconButton = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('button', { staticClass: "bk-icon-button", class: ['bk-' + _vm.themeType, 'bk-button-' + _vm.size, { 'is-disabled': _vm.disabled, 'is-loading': _vm.loading }], attrs: { "title": _vm.title, "disabled": _vm.disabled, "type": _vm.buttonType }, on: { "click": _vm.handleClick } }, [_c('i', { staticClass: "bk-icon", class: ['icon-' + (_vm.icon || 'bk')] }), _vm._v(" "), !_vm.hideText ? _c('i', { staticClass: "bk-text", class: { 'is-disabled': _vm.disabled } }, [_vm._t("default")], 2) : _vm._e()]);
    }, staticRenderFns: [],
    name: 'bk-icon-button',
    props: {
        type: {
            type: String,
            default: 'default',
            validator: function validator(value) {
                var types = value.split(' ');
                var buttons = ['button', 'submit', 'reset'];
                var thenme = ['default', 'info', 'primary', 'warning', 'success', 'danger'];
                var valid = true;

                types.forEach(function (type) {
                    if (buttons.indexOf(type) === -1 && thenme.indexOf(type) === -1) {
                        valid = false;
                    }
                });
                return valid;
            }
        },
        size: {
            type: String,
            default: 'normal',
            validator: function validator(value) {
                return ['mini', 'small', 'normal', 'large'].indexOf(value) > -1;
            }
        },
        title: {
            type: String,
            default: ''
        },
        icon: String,
        disabled: Boolean,
        loading: Boolean,
        hideText: false
    },
    computed: {
        buttonType: function buttonType() {
            var types = this.type.split(' ');
            return types.find(function (type) {
                return type === 'submit' || type === 'button' || type === 'reset';
            });
        },
        themeType: function themeType() {
            var types = this.type.split(' ');
            return types.find(function (type) {
                return type !== 'submit' && type !== 'button' && type !== 'reset';
            });
        }
    },
    methods: {
        handleClick: function handleClick(e) {
            if (!this.disabled && !this.loading) {
                this.$emit('click', e);
            }
        }
    }
};

bkIconButton.install = function (Vue$$1) {
  Vue$$1.component(bkIconButton.name, bkIconButton);
};

var ViewModel = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('transition', { attrs: { "name": "fade" } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.isShow, expression: "isShow" }], staticClass: "bk-loading", style: {
                position: _vm.type === 'directive' ? 'absolute' : 'fixed',
                backgroundColor: "rgba(255, 255, 255, " + _vm.opacity + ")"
            } }, [_c('div', { staticClass: "bk-loading-wrapper" }, [_c('div', { staticClass: "bk-loading1" }, [_c('div', { staticClass: "point point1" }), _vm._v(" "), _c('div', { staticClass: "point point2" }), _vm._v(" "), _c('div', { staticClass: "point point3" }), _vm._v(" "), _c('div', { staticClass: "point point4" })]), _vm._v(" "), _c('div', { staticClass: "bk-loading-title" }, [_vm._t("default", [_vm._v(_vm._s(_vm.title))])], 2)])])]);
    }, staticRenderFns: [],
    name: 'bk-loading',
    data: function data() {
        return {
            opacity: -1,
            isShow: false,
            hide: false,
            title: '',
            type: 'full'
        };
    },

    watch: {
        hide: function hide(newVal) {
            if (newVal) {
                this.isShow = false;
                this.$el.addEventListener('transitionend', this.destroyEl);
            }
        }
    },
    methods: {
        destroyEl: function destroyEl() {
            this.$el.removeEventListener('transitionend', this.destroyEl);
            this.$destroy();
            this.$el.parentNode.removeChild(this.$el);
        }
    },
    mounted: function mounted() {
        this.hide = false;
    }
};

var LoadingConstructor = Vue.extend(ViewModel);
var instance = void 0;

var Loading = function Loading() {
    var options = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};

    if (typeof options === 'string') {
        options = {
            title: options
        };
    }

    options.opacity = options.opacity || 0.9;

    instance = new LoadingConstructor({
        data: options
    });

    if (isVNode(instance.title)) {
        instance.$slots.default = [instance.title];
        instance.title = null;
    } else {
        delete instance.$slots.default;
    }

    instance.viewmodel = instance.$mount();
    document.body.appendChild(instance.viewmodel.$el);
    instance.$dom = instance.viewmodel.$el;
    instance.viewmodel.isShow = true;

    return instance.viewmodel;
};

Loading.hide = function () {
    instance.viewmodel.hide = true;
};

Vue.prototype.$bkLoading = Loading;

var Model = Vue.extend(ViewModel);

function toggle(el, binding) {
    if (!el.$vm) {
        el.$vm = el.viewmodel.$mount();
        el.appendChild(el.$vm.$el);
    }

    if (binding.value.isLoading) {
        Vue.nextTick(function () {
            el.$vm.isShow = true;
        });
    } else {
        el.$vm.isShow = false;
    }

    var title = binding.value.title;

    if (title) {
        el.$vm.title = title;
    }
}

var install = function install(Vue$$1) {
    Vue$$1.directive('bkloading', {
        inserted: function inserted(el, binding) {
            var value = binding.value;

            var position = getComputedStyle(el).position;
            var options = {};

            if (!position || position !== 'relative' || position !== 'absolute') {
                el.style.position = 'relative';
            }

            for (var key in value) {
                if (key !== 'isLoading') {
                    options[key] = value[key];
                }
            }

            options.type = 'directive';
            options.opacity = options.opacity || 0.9;

            var loading = new Model({
                data: options
            });

            el.viewmodel = loading;
            toggle(el, binding);
        },
        update: function update(el, binding) {
            toggle(el, binding);
        }
    });
};

Vue.use(install);

var bkLoading = {
    Loading: Loading,
    directive: install
};

var bkSteps = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-steps", class: 'bk-steps-' + _vm.theme }, _vm._l(_vm.defaultSteps, function (step, index) {
            return _c('div', { staticClass: "bk-step", class: { done: _vm.curStep > index + 1, current: _vm.curStep === index + 1 } }, [_c('span', { staticClass: "bk-step-indicator", class: 'bk-step-' + (_vm.iconType(step) ? 'icon' : 'number'), style: { cursor: _vm.controllable ? 'pointer' : '' }, on: { "click": function click($event) {
                        _vm.jumpTo(index + 1);
                    } } }, [_vm.iconType(step) ? _c('i', { staticClass: "bk-icon", class: 'icon-' + _vm.isIcon(step) }) : _c('span', [_vm._v(_vm._s(_vm.isIcon(step)))])]), _vm._v(" "), step.title ? _c('span', { staticClass: "bk-step-title" }, [_vm._v(" " + _vm._s(step.title) + " ")]) : _vm._e()]);
        }));
    }, staticRenderFns: [],
    name: 'bk-steps',
    mixins: [locale$1],
    props: {
        steps: {
            type: Array,
            default: function _default() {
                return [];
            }
        },
        curStep: {
            type: Number,
            default: 1
        },
        controllable: {
            type: Boolean,
            default: false
        },
        theme: {
            type: String,
            default: 'primary'
        }
    },
    data: function data() {
        return {
            defaultSteps: [{
                title: this.t('steps.step1'),
                icon: 1
            }, {
                title: this.t('steps.step2'),
                icon: 2
            }, {
                title: this.t('steps.step3'),
                icon: 3
            }]
        };
    },
    created: function created() {
        if (this.steps && this.steps.length) {
            var _defaultSteps;

            var defaultSteps = [];
            this.steps.forEach(function (step) {
                if (typeof step === 'string') {
                    defaultSteps.push(step);
                } else {
                    defaultSteps.push({
                        title: step.title,
                        icon: step.icon
                    });
                }
            });
            (_defaultSteps = this.defaultSteps).splice.apply(_defaultSteps, [0, this.defaultSteps.length].concat(defaultSteps));
        }
    },

    methods: {
        iconType: function iconType(step) {
            var icon = step.icon;

            if (icon) {
                return typeof icon === 'string';
            } else {
                return typeof step === 'string';
            }
        },
        isIcon: function isIcon(step) {
            return step.icon ? step.icon : step;
        },
        jumpTo: function jumpTo(index) {
            if (this.controllable && index !== this.curStep) {
                this.$emit('update:curStep', index);
                this.$emit('step-changed', index);
            }
        }
    }
};

bkSteps.install = function (Vue$$1) {
    Vue$$1.component(bkSteps.name, bkSteps);
};

var bkBadge = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-badge-wrapper", style: { 'vertical-align': _vm.$slots.default ? 'middle' : '', 'cursor': _vm.icon ? 'pointer' : '' } }, [_vm._t("default"), _vm._v(" "), _c('transition', { attrs: { "name": "fade-center" } }, [_c('span', { directives: [{ name: "show", rawName: "v-show", value: _vm.visible, expression: "visible" }], staticClass: "bk-badge", class: [_vm.theme && !_vm.hexTheme ? 'bk-' + _vm.theme : '', _vm.$slots.default ? _vm.position : '', { pinned: _vm.$slots.default, dot: _vm.dot }], style: {
                color: _vm.hexTheme ? '#fff' : '',
                backgroundColor: _vm.hexTheme ? _vm.theme : ''
            }, on: { "mouseenter": _vm.handleHover, "mouseleave": _vm.handleLeave } }, [_vm.icon && !_vm.dot ? _c('i', { staticClass: "bk-icon", class: 'icon-' + _vm.icon }) : _vm._e(), _vm._v(" "), !_vm.icon && !_vm.dot ? _c('span', [_vm._v(" " + _vm._s(_vm.text) + " ")]) : _vm._e()])])], 2);
    }, staticRenderFns: [],
    name: 'bk-badge',
    props: {
        theme: {
            type: String,
            default: '',
            validator: function validator(value) {
                return ['', 'primary', 'info', 'warning', 'danger', 'success'].indexOf(value) > -1 || value.indexOf('#') === 0;
            }
        },
        val: {
            type: [Number, String],
            default: 1
        },
        icon: {
            type: String,
            default: ''
        },
        max: {
            type: Number,
            default: -1
        },
        dot: {
            type: Boolean,
            default: false
        },
        visible: {
            type: Boolean,
            default: true
        },
        position: {
            type: String,
            default: 'top-right'
        }
    },
    computed: {
        text: function text() {
            var _type = _typeof(this.val);
            var _max = this.max;
            var _value = this.val;
            var _icon = this.icon;

            if (_icon) {
                return _icon;
            } else {
                if (_type === 'number' && _max > -1 && _value > _max) {
                    return _max + '+';
                } else {
                    return _value;
                }
            }
        },
        hexTheme: function hexTheme() {
            return (/^#[0-9a-fA-F]{3,6}$/.test(this.theme)
            );
        }
    },
    methods: {
        handleHover: function handleHover() {
            this.$emit('hover');
        },
        handleLeave: function handleLeave() {
            this.$emit('leave');
        }
    }
};

bkBadge.install = function (Vue$$1) {
    Vue$$1.component(bkBadge.name, bkBadge);
};

var Message = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('transition', { attrs: { "name": "displacement-fade-show" } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.isShow, expression: "isShow" }], staticClass: "bk-message", class: _vm.isClass ? _vm.theme : '', style: { 'background-color': _vm.isClass ? '' : _vm.theme } }, [_c('div', { staticClass: "bk-message-icon-wrapper" }, [_c('div', { staticClass: "bk-message-icon" }, [_vm.icon ? _c('i', { staticClass: "bk-icon", class: 'icon-' + _vm.calcIcon }) : _vm._e()])]), _vm._v(" "), _c('div', { staticClass: "bk-message-content" }, [_vm._t("default", [_vm._v(_vm._s(_vm.message))])], 2), _vm._v(" "), _vm.hasCloseIcon ? _c('div', { staticClass: "bk-message-close", on: { "click": _vm.close } }, [_c('i', { staticClass: "bk-icon icon-close", attrs: { "title": _vm.t('message.close') } })]) : _vm._e()])]);
    }, staticRenderFns: [],
    name: 'bk-message',
    mixins: [locale$1],
    data: function data() {
        return {
            theme: 'primary',
            icon: 'check-1',
            message: 'hahaha',
            delay: 3000,
            hasCloseIcon: false,
            onClose: function onClose() {},
            onShow: function onShow() {},
            isShow: false,
            countdownId: 0,
            visible: false
        };
    },

    computed: {
        isClass: function isClass() {
            return true;
        },
        calcIcon: function calcIcon() {
            var theme = this.theme;
            var icon = void 0;

            if (!this.icon) return;

            switch (theme) {
                case 'error':
                    icon = 'close';
                    break;
                case 'warning':
                    icon = 'exclamation';
                    break;
                case 'success':
                    icon = 'check-1';
                    break;
                case 'primary':
                    icon = 'dialogue-shape';
                    break;
            }

            return icon;
        }
    },
    watch: {
        visible: function visible(val) {
            if (!val) {
                this.isShow = false;
                this.$el.addEventListener('transitionend', this._destroyEl);
            }
        }
    },
    mounted: function mounted() {
        this.visible = true;

        if (this.delay) {
            this.startCountDown();
        }
    },

    methods: {
        _destroyEl: function _destroyEl() {
            this.$el.removeEventListener('transitionend', this._destroyEl);
            this.$destroy();
            this.$el.parentNode.removeChild(this.$el);
        },
        close: function close() {
            this.onClose && this.onClose(this);
            this.visible = false;
        },
        handleShow: function handleShow() {
            this.onShow && this.onShow(this);
            this.close();
        },
        startCountDown: function startCountDown() {
            var _this = this;

            this.countdownId = setTimeout(function () {
                clearTimeout(_this.countdownId);
                _this.close();
            }, this.delay);
        }
    }
};

var MessageConstructor = Vue.extend(Message);
var instance$1 = void 0;
var instancesArr = [];
var count = 0;
var zIndex = new Date().getFullYear();

var Msg = function Msg() {
    var options = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};

    var id = 'bkMessage' + count++;
    var usrClose = options.onClose;
    var type = (typeof options === 'undefined' ? 'undefined' : _typeof(options)).toLowerCase();

    if (type === 'string' || type === 'number') {
        options = {
            message: options
        };
    }

    options.onClose = function () {
        Msg.close(id, usrClose);
    };

    instance$1 = new MessageConstructor({
        data: options
    });

    if (isVNode(instance$1.message)) {
        instance$1.$slots.default = [instance$1.message];
        instance$1.message = null;
    }

    instance$1.id = id;
    instance$1.viewmodel = instance$1.$mount();
    document.body.appendChild(instance$1.viewmodel.$el);
    instance$1.$dom = instance$1.viewmodel.$el;
    instance$1.$dom.style.zIndex = zIndex++;
    instance$1.viewmodel.isShow = true;
    instancesArr.push(instance$1);
    return instance$1.viewmodel;
};

Msg.close = function (id, usrClose) {
    var len = instancesArr.length;
    for (var index = 0; index < len; index++) {
        if (id === instancesArr[index].id) {
            usrClose && usrClose(instancesArr[index]);
        }
        instancesArr.splice(index, 1);
        break;
    }
};

Vue.prototype.$bkMessage = Msg;

var supportsPassive = false;
document.addEventListener('passive-check', function () {}, {
    get passive() {
        supportsPassive = {
            passive: true
        };
    }
});

var ViewModel$1 = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('transition', { attrs: { "name": "v-tooltips-fade" } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.visible, expression: "visible" }], staticClass: "v-tooltips-container", class: _vm.boxClass, style: _vm.boxStyle, on: { "mouseenter": _vm.showTooltips, "mouseleave": function mouseleave($event) {
                    _vm.hiddenTooltips(true);
                } } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.placement, expression: "placement" }], staticClass: "v-tooltips-arrows", class: _vm.placement, style: _vm.arrowBox }), _vm._v(" "), _vm.title ? _c('span', { staticClass: "v-tooltips-title" }, [_vm._v(_vm._s(_vm.title))]) : _vm._e(), _vm._v(" "), _vm.content ? _c('p', { staticClass: "v-tooltips-content", style: _vm.contentHeight }, [_vm._v(" " + _vm._s(_vm.content) + " ")]) : _vm._e(), _vm._v(" "), _vm.customComponent ? _c(_vm.customComponent, _vm._g(_vm._b({ tag: "component", on: { "hidden-tooltips": _vm.hiddenTooltips, "update-tooltips": _vm.updateTooltips } }, 'component', _vm.customProps, false), _vm.customListeners)) : _vm._e()], 1)]);
    }, staticRenderFns: [],
    name: 'bk-tooltips',
    props: {
        title: {
            type: String,
            default: ''
        },

        content: {
            type: String,
            default: ''
        },

        customProps: {
            type: Object,
            default: function _default() {
                return {};
            }
        },

        customComponent: {
            type: [String, Function, Object],
            default: ''
        },

        customListeners: Object,

        target: null,

        container: null,

        placements: {
            type: Array,
            default: function _default() {
                return ['top', 'right', 'bottom', 'left'];
            }
        },

        duration: {
            type: Number,
            default: 300
        },

        arrowsSize: {
            type: Number,
            default: 8
        },

        width: {
            type: [String, Number],
            default: 'auto'
        },

        height: {
            type: [String, Number],
            default: 'auto'
        },

        zIndex: {
            type: Number,
            default: 999
        },

        theme: {
            type: String,
            default: 'dark'
        },

        customClass: {
            type: String,
            default: ''
        },

        onShow: {
            type: Function,
            default: function _default() {}
        },

        onClose: {
            type: Function,
            default: function _default() {}
        }
    },
    data: function data() {
        return {
            placement: '',
            visible: false,
            arrowsPos: {},
            containerNode: null,
            targetParentNode: null,
            visibleTimer: null
        };
    },

    computed: {
        arrowBox: function arrowBox() {
            return Object.assign({ borderWidth: this.arrowsSize + 'px' }, this.arrowsPos);
        },
        boxStyle: function boxStyle() {
            var width = this.width;
            return {
                width: typeof width === 'string' ? width : width + 'px',
                zIndex: this.zIndex
            };
        },
        boxClass: function boxClass() {
            var customClass = this.customClass,
                theme = this.theme;

            return [customClass, theme];
        },
        contentHeight: function contentHeight() {
            var height = this.height;
            return {
                height: typeof height === 'string' ? height : height + 'px'
            };
        }
    },
    watch: {
        visible: function visible(val) {
            if (val) {
                this.onShow && this.onShow(this);
            } else {
                this.onClose && this.onClose(this);
            }
        }
    },
    methods: {
        showTooltips: function showTooltips() {
            clearTimeout(this.visibleTimer);
            this.visible = true;
        },
        hiddenTooltips: function hiddenTooltips(immediate) {
            if (this.duration <= -1) {
                return;
            }
            if (immediate) {
                this.visible = false;
            } else {
                this.setVisible(false);
            }
        },
        updateTooltips: function updateTooltips() {
            this.setContainerNode();
            this.showTooltips();
            this.$nextTick(this.setPosition);
        },
        setContainerNode: function setContainerNode() {
            var $el = this.$el,
                target = this.target,
                container = this.container,
                targetParentNode = this.targetParentNode,
                oldNode = this.containerNode;

            if (!target || target.parentNode === targetParentNode) {
                return;
            }

            this.targetParentNode = target.parentNode;

            var newNode = container || getScrollContainer(target);
            if (newNode === oldNode) {
                return;
            }

            if ($el.parentNode !== newNode) {
                newNode.appendChild($el);
            }

            var position = window.getComputedStyle(newNode, null).position;
            if (!position || position === 'static') {
                newNode.style.position = 'relative';
            }
            if (oldNode) {
                oldNode.removeEventListener('scroll', this.scrollHandler, supportsPassive);
            }

            if (checkScrollable(newNode)) {
                newNode.addEventListener('scroll', this.scrollHandler, supportsPassive);
            }
            this.containerNode = newNode;
        },
        setPosition: function setPosition() {
            var $el = this.$el,
                target = this.target,
                containerNode = this.containerNode,
                placements = this.placements,
                arrowsSize = this.arrowsSize;

            if (!$el || !target || !containerNode) {
                return;
            }
            var placementInfo = computePlacementInfo(target, containerNode, $el, placements, arrowsSize);
            var coordinate = placementInfo.mod === 'mid' ? computeCoordinateBaseMid(placementInfo, arrowsSize) : computeCoordinateBaseEdge(placementInfo, arrowsSize);

            this.setArrowsPos(coordinate);
            this.placement = coordinate.placement;

            var x = Math.round(coordinate.x + containerNode.scrollLeft);
            var y = Math.round(coordinate.y + containerNode.scrollTop);
            this.$el.style.transform = 'translate3d(' + x + 'px, ' + y + 'px, 0)';
        },
        setArrowsPos: function setArrowsPos(_ref) {
            var placement = _ref.placement,
                arrowsOffset = _ref.arrowsOffset;

            this.arrowsPos = computeArrowPos(placement, arrowsOffset, this.arrowsSize);
        },
        setVisible: function setVisible(v) {
            var _this = this;

            clearTimeout(this.visibleTimer);
            this.visibleTimer = setTimeout(function () {
                _this.visible = v;
                _this.visibleTimer = null;
            }, this.duration);
        },

        scrollHandler: debounce(function () {
            this.setPosition();
        }, 200, true),

        clearScrollEvent: function clearScrollEvent() {
            if (this.containerNode) {
                this.containerNode.removeEventListener('scroll', this.scrollHandler, supportsPassive);
            }
        },
        removeParentNode: function removeParentNode() {
            if (this.$el.parentNode) {
                this.$el.parentNode.removeChild(this.$el);
            }
        },
        destroy: function destroy() {
            this.clearScrollEvent();
            this.removeParentNode();
            this.$destroy();
        }
    }
};

var TooltipsConstructor = Vue.extend(ViewModel$1);

var props = ViewModel$1.props;
var defaultOptions = {};
Object.keys(props).forEach(function (key) {
    var prop = props[key];
    var dv = prop.default;
    if (prop && prop.default != null) {
        defaultOptions[key] = typeof dv === 'function' ? dv() : dv;
    }
});

var tooltipsInstance = null;

function tooltips(options) {
    options = options || {};

    if (tooltipsInstance && tooltipsInstance.$el.parentNode) {
        Object.assign(tooltipsInstance, defaultOptions, options);
        if (tooltipsInstance.target) {
            tooltipsInstance.updateTooltips();
        } else {
            tooltipsInstance.hiddenTooltips();
        }
        return tooltipsInstance;
    }

    tooltipsInstance = new TooltipsConstructor({
        propsData: options
    }).$mount();

    tooltipsInstance.updateTooltips();
    return tooltipsInstance;
}

function clearEvent(el) {
    if (el._tooltipsHandler) {
        el.removeEventListener('click', el._tooltipsHandler);
        el.removeEventListener('mouseenter', el._tooltipsHandler);
    }
    if (el._tooltipsMouseleaveHandler) {
        el.removeEventListener('mouseleave', el._tooltipsMouseleaveHandler);
    }
    delete el._tooltipsHandler;
    delete el._tooltipsMouseleaveHandler;
    delete el._tooltipsOptions;
    delete el._tooltipsInstance;
}

var directive$1 = {
    install: function install(Vue$$1, options) {
        options = options || {};

        var allPlacements = ['top', 'right', 'bottom', 'left'];

        Vue$$1.directive('bktooltips', {
            bind: function bind(el, binding) {
                clearEvent(el);

                var _binding$modifiers = binding.modifiers,
                    click = _binding$modifiers.click,
                    light = _binding$modifiers.light;

                var limitPlacementQueue = allPlacements.filter(function (placement) {
                    return binding.modifiers[placement];
                });

                el._tooltipsOptions = binding.value;

                el._tooltipsHandler = function tooltipsHandler() {
                    if (this._tooltipsOptions == null) {
                        return;
                    }
                    var options = this._tooltipsOptions;
                    var placements = limitPlacementQueue.length ? limitPlacementQueue : allPlacements;
                    var mix = {
                        placements: placements,
                        theme: light ? 'light' : 'dark'
                    };

                    var tipOptions = (typeof options === 'undefined' ? 'undefined' : _typeof(options)) === 'object' ? Object.assign(mix, options, { target: this }) : Object.assign(mix, { content: String(options), target: this });

                    this._tooltipsInstance = tooltips(tipOptions);
                };
                el._tooltipsMouseleaveHandler = function tooltipsMouseleaveHandler() {
                    if (this._tooltipsInstance) {
                        this._tooltipsInstance.hiddenTooltips();
                    }
                };

                if (click) {
                    el.addEventListener('click', el._tooltipsHandler);
                } else {
                    el.addEventListener('mouseenter', el._tooltipsHandler);
                }
                el.addEventListener('mouseleave', el._tooltipsMouseleaveHandler);
            },
            update: function update(el, _ref) {
                var value = _ref.value,
                    oldValue = _ref.oldValue;

                if (value === oldValue) {
                    return;
                }
                el._tooltipsOptions = value;
            },
            unbind: function unbind(el) {
                var instance = el._tooltipsInstance;
                if (instance && instance.destroy) {
                    instance.destroy();
                }
                clearEvent(el);
            }
        });
    }
};

var bkTooltips = { tooltips: tooltips, directive: directive$1 };

var bkDropdown = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { directives: [{ name: "clickoutside", rawName: "v-clickoutside", value: _vm.close, expression: "close" }], staticClass: "bk-selector", class: [_vm.extCls, { 'open': _vm.open }], on: { "click": _vm.openFn } }, [_c('div', { staticClass: "bk-selector-wrapper" }, [_c('input', { staticClass: "bk-selector-input", class: { placeholder: _vm.selectedText === _vm.defaultPlaceholder, active: _vm.open }, attrs: { "readonly": "readonly", "placeholder": _vm.defaultPlaceholder, "disabled": _vm.disabled }, domProps: { "value": _vm.selectedText }, on: { "mouseover": _vm.showClearFn, "mouseleave": function mouseleave($event) {
                    _vm.showClear = false;
                } } }), _vm._v(" "), _c('i', { directives: [{ name: "show", rawName: "v-show", value: !_vm.isLoading && !_vm.showClear, expression: "!isLoading && !showClear" }], class: ['bk-icon icon-angle-down bk-selector-icon', { 'disabled': _vm.disabled }] }), _vm._v(" "), _c('i', { directives: [{ name: "show", rawName: "v-show", value: !_vm.isLoading && _vm.showClear, expression: "!isLoading && showClear" }], class: ['bk-icon icon-close bk-selector-icon clear-icon'], on: { "mouseover": _vm.showClearFn, "mouseleave": function mouseleave($event) {
                    _vm.showClear = false;
                }, "click": function click($event) {
                    _vm.clearSelected($event);
                } } }), _vm._v(" "), _c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.isLoading, expression: "isLoading" }], staticClass: "bk-spin-loading bk-spin-loading-mini bk-spin-loading-primary selector-loading-icon" }, [_c('div', { staticClass: "rotate rotate1" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate2" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate3" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate4" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate5" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate6" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate7" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate8" })])]), _vm._v(" "), _c('transition', { attrs: { "name": _vm.listSlideName } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: !_vm.isLoading && _vm.open, expression: "!isLoading && open" }], staticClass: "bk-selector-list", style: _vm.panelStyle }, [_vm.searchable ? _c('div', { staticClass: "bk-selector-search-item", on: { "click": function click($event) {
                    $event.stopPropagation();
                } } }, [_c('i', { staticClass: "bk-icon icon-search" }), _vm._v(" "), _c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.condition, expression: "condition" }], ref: "searchNode", attrs: { "type": "text" }, domProps: { "value": _vm.condition }, on: { "input": [function ($event) {
                    if ($event.target.composing) {
                        return;
                    }_vm.condition = $event.target.value;
                }, _vm.inputFn] } })]) : _vm._e(), _vm._v(" "), _c('ul', { style: { 'max-height': _vm.contentMaxHeight + 'px' } }, [_vm._l(_vm.localList, function (item, index) {
            return _vm.localList.length !== 0 ? _c('li', { class: ['bk-selector-list-item', item.children && item.children.length ? 'bk-selector-group-list-item' : ''] }, [item.children && item.children.length ? [_c('div', { staticClass: "bk-selector-group-name" }, [_vm._v(_vm._s(item[_vm.displayKey]))]), _vm._v(" "), _c('ul', { staticClass: "bk-selector-group-list" }, _vm._l(item.children, function (child, index) {
                return _c('li', { staticClass: "bk-selector-list-item" }, [_c('div', { staticClass: "bk-selector-node bk-selector-sub-node", class: { 'bk-selector-selected': !_vm.multiSelect && child[_vm.settingKey] === _vm.selected } }, [_c('div', { staticClass: "text", on: { "click": function click($event) {
                            $event.stopPropagation();_vm.selectItem(child, $event);
                        } } }, [_vm.multiSelect ? _c('label', { staticClass: "bk-form-checkbox bk-checkbox-small mr0 bk-selector-multi-label" }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.localSelected, expression: "localSelected" }], attrs: { "type": "checkbox", "name": 'multiSelect' + +new Date() }, domProps: { "value": child[_vm.settingKey], "checked": Array.isArray(_vm.localSelected) ? _vm._i(_vm.localSelected, child[_vm.settingKey]) > -1 : _vm.localSelected }, on: { "change": function change($event) {
                            var $$a = _vm.localSelected,
                                $$el = $event.target,
                                $$c = $$el.checked ? true : false;if (Array.isArray($$a)) {
                                var $$v = child[_vm.settingKey],
                                    $$i = _vm._i($$a, $$v);if ($$el.checked) {
                                    $$i < 0 && (_vm.localSelected = $$a.concat([$$v]));
                                } else {
                                    $$i > -1 && (_vm.localSelected = $$a.slice(0, $$i).concat($$a.slice($$i + 1)));
                                }
                            } else {
                                _vm.localSelected = $$c;
                            }
                        } } }), _vm._v(" " + _vm._s(child[_vm.displayKey]) + " ")]) : [_vm._v(" " + _vm._s(child[_vm.displayKey]) + " ")]], 2), _vm._v(" "), _vm.tools !== false ? _c('div', { staticClass: "bk-selector-tools" }, [_vm.tools.edit !== false ? _c('i', { staticClass: "bk-icon icon-edit2 bk-selector-list-icon", on: { "click": function click($event) {
                            $event.stopPropagation();_vm.editFn(index);
                        } } }) : _vm._e(), _vm._v(" "), _vm.tools.del !== false ? _c('i', { staticClass: "bk-icon icon-close bk-selector-list-icon", on: { "click": function click($event) {
                            $event.stopPropagation();_vm.delFn(index);
                        } } }) : _vm._e()]) : _vm._e()])]);
            }))] : [_c('div', { staticClass: "bk-selector-node", class: { 'bk-selector-selected': !_vm.multiSelect && item[_vm.settingKey] === _vm.selected, 'is-disabled': item.isDisabled } }, [_c('div', { staticClass: "text", attrs: { "title": item[_vm.displayKey] }, on: { "click": function click($event) {
                        $event.stopPropagation();_vm.selectItem(item, $event);
                    } } }, [_vm.multiSelect ? _c('label', { staticClass: "bk-form-checkbox bk-checkbox-small mr0 bk-selector-multi-label" }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.localSelected, expression: "localSelected" }], attrs: { "type": "checkbox", "name": 'multiSelect' + +new Date() }, domProps: { "value": item[_vm.settingKey], "checked": Array.isArray(_vm.localSelected) ? _vm._i(_vm.localSelected, item[_vm.settingKey]) > -1 : _vm.localSelected }, on: { "change": function change($event) {
                        var $$a = _vm.localSelected,
                            $$el = $event.target,
                            $$c = $$el.checked ? true : false;if (Array.isArray($$a)) {
                            var $$v = item[_vm.settingKey],
                                $$i = _vm._i($$a, $$v);if ($$el.checked) {
                                $$i < 0 && (_vm.localSelected = $$a.concat([$$v]));
                            } else {
                                $$i > -1 && (_vm.localSelected = $$a.slice(0, $$i).concat($$a.slice($$i + 1)));
                            }
                        } else {
                            _vm.localSelected = $$c;
                        }
                    } } }), _vm._v(" " + _vm._s(item[_vm.displayKey]) + " ")]) : [_vm._v(" " + _vm._s(item[_vm.displayKey]) + " ")]], 2), _vm._v(" "), _vm.tools !== false ? _c('div', { staticClass: "bk-selector-tools" }, [_vm.tools.edit !== false ? _c('i', { staticClass: "bk-icon icon-edit2 bk-selector-list-icon", on: { "click": function click($event) {
                        $event.stopPropagation();_vm.editFn(index);
                    } } }) : _vm._e(), _vm._v(" "), _vm.tools.del !== false ? _c('i', { staticClass: "bk-icon icon-close bk-selector-list-icon", on: { "click": function click($event) {
                        $event.stopPropagation();_vm.delFn(index);
                    } } }) : _vm._e()]) : _vm._e()])]], 2) : _vm._e();
        }), _vm._v(" "), !_vm.isLoading && _vm.localList.length === 0 ? _c('li', { staticClass: "bk-selector-list-item" }, [_c('div', { staticClass: "text no-search-result" }, [_vm._v(" " + _vm._s(_vm.list.length ? _vm.defaultSearchEmptyText : _vm.defaultEmptyText) + " ")])]) : _vm._e()], 2), _vm._v(" "), _vm._t("default")], 2)])], 1);
    }, staticRenderFns: [],
    name: 'bk-dropdown',
    mixins: [locale$1],
    directives: {
        clickoutside: clickoutside
    },
    props: {
        extCls: {
            type: String
        },
        isLoading: {
            type: Boolean,
            default: false
        },
        hasCreateItem: {
            type: Boolean,
            default: false
        },
        tools: {
            type: [Object, Boolean],
            default: false
        },
        list: {
            type: Array,
            required: true
        },
        filterList: {
            type: Array,
            default: function _default() {
                return [];
            }
        },
        selected: {
            type: [Number, Array, String],
            required: true
        },
        placeholder: {
            type: [String, Boolean],
            default: ''
        },

        isLink: {
            type: [String, Boolean],
            default: false
        },
        displayKey: {
            type: String,
            default: 'name'
        },
        disabled: {
            type: [String, Boolean, Number],
            default: false
        },
        multiSelect: {
            type: Boolean,
            default: false
        },
        searchable: {
            type: Boolean,
            default: false
        },
        searchKey: {
            type: String,
            default: 'name'
        },
        allowClear: {
            type: Boolean,
            default: false
        },
        settingKey: {
            type: String,
            default: 'id'
        },
        initPreventTrigger: {
            type: Boolean,
            default: false
        },
        emptyText: {
            type: String,
            default: ''
        },
        searchEmptyText: {
            type: String,
            default: ''
        },
        contentMaxHeight: {
            type: Number,
            default: 300
        }
    },
    data: function data() {
        return {
            open: false,
            selectedList: this.calcSelected(this.selected),
            condition: '',

            localSelected: this.selected,

            showClear: false,
            panelStyle: {},
            listSlideName: 'toggle-slide',
            defaultPlaceholder: this.t('selector.pleaseselect'),
            defaultEmptyText: this.t('selector.emptyText'),
            defaultSearchEmptyText: this.t('selector.searchEmptyText')
        };
    },

    computed: {
        localList: function localList() {
            var _this = this;

            if (!this.multiSelect) {
                this.list.forEach(function (item) {
                    if (_this.filterList.includes(item[_this.settingKey])) {
                        item.isDisabled = true;
                    } else {
                        item.isDisabled = false;
                    }
                });
            }
            if (this.searchable && this.condition) {
                var arr = [];
                var key = this.searchKey;

                var len = this.list.length;
                for (var i = 0; i < len; i++) {
                    var item = this.list[i];
                    if (item.children) {
                        var results = [];
                        var childLen = item.children.length;
                        for (var j = 0; j < childLen; j++) {
                            var child = item.children[j];
                            if (child[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                results.push(child);
                            }
                        }
                        if (results.length) {
                            var cloneItem = Object.assign({}, item);
                            cloneItem.children = results;
                            arr.push(cloneItem);
                        }
                    } else {
                        if (item[key].toLowerCase().includes(this.condition.toLowerCase())) {
                            arr.push(item);
                        }
                    }
                }

                return arr;
            }
            return this.list;
        },
        currentItem: function currentItem() {
            return this.list[this.localSelected];
        },
        selectedText: function selectedText() {
            var _this2 = this;

            var text = this.defaultPlaceholder;
            var textArr = [];
            if (Array.isArray(this.selectedList) && this.selectedList.length) {
                this.selectedList.forEach(function (item) {
                    textArr.push(item[_this2.displayKey]);
                });
            } else if (this.selectedList) {
                this.selectedList[this.displayKey] && textArr.push(this.selectedList[this.displayKey]);
            }
            return textArr.length ? textArr.join(',') : this.defaultPlaceholder;
        }
    },
    watch: {
        selected: function selected(newVal) {
            if (this.list.length) {
                this.selectedList = this.calcSelected(this.selected, this.isLink);
            }

            this.localSelected = this.selected;
        },
        list: function list(newVal) {
            if (this.selected) {
                this.selectedList = this.calcSelected(this.selected, this.isLink);
            } else {
                this.selectedList = [];
            }
        },
        localSelected: function localSelected(val) {
            if (this.list.length) {
                this.selectedList = this.calcSelected(this.localSelected, this.isLink);
            }
        },
        open: function open(newVal) {
            var searchNode = this.$refs.searchNode;
            if (searchNode) {
                if (newVal) {
                    this.$nextTick(function () {
                        searchNode.focus();
                    });
                }
            }
            this.$emit('visible-toggle', newVal);
        }
    },
    created: function created() {
        if (this.placeholder) {
            this.defaultPlaceholder = this.placeholder;
        }
        if (this.emptyText) {
            this.defaultEmptyText = this.emptyText;
        }
        if (this.searchEmptyText) {
            this.defaultSearchEmptyText = this.searchEmptyText;
        }
    },
    mounted: function mounted() {
        this.popup = this.$el;
        if (this.isLink) {
            if (this.list.length && this.selected) {
                this.calcSelected(this.selected, this.isLink);
            }
        }
    },

    methods: {
        getItem: function getItem(key) {
            var _this3 = this;

            var data = null;

            this.list.forEach(function (item) {
                if (!item.children) {
                    if (String(item[_this3.settingKey]) === String(key)) {
                        data = item;
                    }
                } else {
                    var children = item.children;
                    children.forEach(function (child) {
                        if (String(child[_this3.settingKey]) === String(key)) {
                            data = child;
                        }
                    });
                }
            });
            return data;
        },
        calcSelected: function calcSelected(selected, isTrigger) {
            var data = null;

            if (Array.isArray(selected)) {
                data = [];
                var len = selected.length;
                for (var i = 0; i < len; i++) {
                    var item = this.getItem(selected[i]);
                    if (item) {
                        data.push(item);
                    }
                }

                if (data.length && isTrigger && !this.initPreventTrigger) {
                    this.$emit('item-selected', selected, data, isTrigger);
                }
            } else if (selected !== undefined) {
                var _item = this.getItem(selected);
                if (_item) {
                    data = _item;
                }
                if (data && isTrigger && !this.initPreventTrigger) {
                    this.$emit('item-selected', selected, data, isTrigger);
                }
            }
            return data;
        },
        close: function close() {
            this.open = false;
        },
        initSelectorPosition: function initSelectorPosition(currentTarget) {
            if (currentTarget) {
                var distanceLeft = getActualLeft(currentTarget);
                var distanceTop = getActualTop(currentTarget);
                var winWidth = document.body.clientWidth;
                var winHeight = document.body.clientHeight;
                var ySet = {};
                var listHeight = this.list.length * 42;
                if (listHeight > 160) {
                    listHeight = 160;
                }
                var scrollTop = document.documentElement.scrollTop || document.body.scrollTop;

                if (distanceTop + listHeight + 42 - scrollTop < winHeight) {
                    ySet = {
                        top: '40px',
                        bottom: 'auto'
                    };

                    this.listSlideName = 'toggle-slide';
                } else {
                    ySet = {
                        top: 'auto',
                        bottom: '40px'
                    };

                    this.listSlideName = 'toggle-slide2';
                }

                this.panelStyle = _extends({}, ySet);
            }
        },
        openFn: function openFn(event) {
            if (this.disabled) {
                return;
            }

            if (!this.disabled) {
                if (!this.open && event) {
                    this.initSelectorPosition(event.currentTarget);
                }
                this.open = !this.open;
            }
        },
        calcList: function calcList() {
            if (this.searchable) {
                var arr = [];
                var key = this.searchKey;

                var len = this.list.length;
                for (var i = 0; i < len; i++) {
                    var item = this.list[i];
                    if (item.children) {
                        var results = [];
                        var childLen = item.children.length;
                        for (var j = 0; j < childLen; j++) {
                            var child = item.children[j];
                            if (child[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                results.push(child);
                            }
                        }
                        if (results.length) {
                            var cloneItem = Object.assign({}, item);
                            cloneItem.children = results;
                            arr.push(cloneItem);
                        }
                    } else {
                        if (item[key].toLowerCase().includes(this.condition.toLowerCase())) {
                            arr.push(item);
                        }
                    }
                }

                this.localList = arr;
            } else {
                this.localList = this.list;
            }
        },
        showClearFn: function showClearFn() {
            if (this.allowClear && !this.multiSelect && this.localSelected !== -1 && this.localSelected !== '') {
                this.showClear = true;
            }
        },
        clearSelected: function clearSelected(e) {
            this.$emit('clear', this.localSelected);
            this.localSelected = -1;
            this.showClear = false;
            e.stopPropagation();
            this.$emit('update:selected', '');
        },
        selectItem: function selectItem(data, event) {
            var _this4 = this;

            if (data.isDisabled) return;
            setTimeout(function () {
                _this4.toggleSelect(data, event);
            }, 10);
        },
        toggleSelect: function toggleSelect(data, event) {
            var $selected = void 0;
            var $selectedList = void 0;
            var settingKey = this.settingKey;
            var isMultiSelect = this.multiSelect;
            var list = this.localList;
            var index = data && data[settingKey] !== undefined ? data[settingKey] : undefined;

            if (isMultiSelect && event.target.tagName.toLowerCase() === 'label') {
                return;
            }
            if (index !== undefined) {
                if (!isMultiSelect) {
                    $selected = index;
                } else {
                    $selected = this.localSelected;
                }

                this.$emit('update:selected', $selected);
                $selectedList = this.calcSelected($selected);
            } else {
                this.$emit('update:selected', -1);
            }

            this.$emit('item-selected', $selected, $selectedList);

            if (!isMultiSelect) {
                this.openFn();
            }
        },
        editFn: function editFn(index) {
            this.$emit('edit', index);
            this.openFn();
        },
        delFn: function delFn(index) {
            this.$emit('del', index);
            this.openFn();
        },
        createFn: function createFn(e) {
            this.$emit('create');
            this.openFn();
            e.stopPropagation();
        },
        inputFn: function inputFn() {
            this.$emit('typing', this.condition);
        }
    }
};

bkDropdown.install = function (Vue$$1) {
    Vue$$1.component(bkDropdown.name, bkDropdown);
};

var InfoBox = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('transition', { attrs: { "name": "displacement-fade-show" } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.isShow, expression: "isShow" }], staticClass: "bk-dialog", class: _vm.clsName }, [_c('div', { staticClass: "bk-dialog-wrapper" }, [_c('div', { staticClass: "bk-dialog-position", on: { "click": function click($event) {
                    if ($event.target !== $event.currentTarget) {
                        return null;
                    }return _vm.handleQuickClose($event);
                } } }, [_c('div', { staticClass: "bk-dialog-style" }, [_vm.closeIcon ? _c('div', { staticClass: "bk-dialog-tool", on: { "click": _vm.handleCancel } }, [_c('i', { staticClass: "bk-dialog-close bk-icon icon-close" })]) : _vm._e(), _vm._v(" "), _vm.type === 'default' ? _c('div', { staticClass: "bk-dialog-header" }, [_c('h3', { staticClass: "bk-dialog-title" }, [_vm._v(" " + _vm._s(_vm.title) + " ")])]) : _vm._e(), _vm._v(" "), _c('div', { staticClass: "bk-dialog-body", class: [{ 'bk-dialog-default-status': _vm.type !== 'default' }, 'bk-dialog-' + _vm.type, _vm.type === 'default' && _vm.content === false ? 'p0' : ''] }, [_vm.type === 'default' && _vm.content !== false ? _vm._t("content", [_vm._v(" " + _vm._s(_vm.content) + " ")]) : _vm._e(), _vm._v(" "), _vm.type !== 'default' ? _c('div', { staticClass: "bk-dialog-row" }, [_vm.type === 'loading' ? _c('img', { staticClass: "bk-dialog-mark bk-dialog-loading", attrs: { "src": "../../bk-magic-ui/src/images/default_loading.png", "alt": "loading" } }) : _c('p', [_c('i', { staticClass: "bk-icon bk-dialog-mark", class: ['bk-dialog-' + _vm.type, 'icon-' + _vm.calcIcon] })]), _vm._v(" "), _vm.statusOpts.title !== false ? _vm._t("statusTitle", [_c('h3', { staticClass: "bk-dialog-title bk-dialog-row" }, [_vm._v(" " + _vm._s(_vm.statusOpts.title ? _vm.statusOpts.title : _vm.calcStatusOpts.title) + " ")])]) : _vm._e(), _vm._v(" "), _vm.type !== 'warning' && _vm.statusOpts.subtitle !== false ? _vm._t("statusSubtitle", [_c('h5', { staticClass: "bk-dialog-subtitle bk-dialog-row" }, [_vm._v(" " + _vm._s(_vm.statusOpts.subtitle ? _vm.statusOpts.subtitle : _vm.calcStatusOpts.subtitle) + " ")])]) : _vm._e()], 2) : _vm._e()], 2), _vm._v(" "), _vm.type === 'default' || _vm.type === 'warning' ? _c('div', { staticClass: "bk-dialog-footer", staticStyle: { "font-size": "0" } }, [_c('button', { staticClass: "bk-dialog-btn bk-dialog-btn-confirm", class: 'bk-btn-' + _vm.theme, attrs: { "type": "button", "name": "confirm" }, on: { "click": _vm.handleConfirm } }, [_vm._v(_vm._s(_vm.confirm))]), _vm._v(" "), _c('button', { staticClass: "bk-dialog-btn bk-dialog-btn-cancel", attrs: { "type": "button", "name": "cancel" }, on: { "click": _vm.handleCancel } }, [_vm._v(_vm._s(_vm.cancel))])]) : _vm._e()])])])])]);
    }, staticRenderFns: [],
    name: 'bk-info-box',
    mixins: [locale$1],
    data: function data() {
        return {
            isShow: false,
            clsName: '',
            type: 'default',
            title: this.t('infobox.title'),
            content: false,
            icon: '',
            statusOpts: {},
            closeIcon: true,
            theme: 'primary',
            confirm: this.t('infobox.ok'),
            cancel: this.t('infobox.cancel'),
            quickClose: false,
            delay: false,
            confirmFn: function confirmFn() {},
            cancelFn: function cancelFn() {},
            shown: function shown() {},
            hidden: function hidden() {},

            hide: true,
            delayId: -1
        };
    },

    computed: {
        calcIcon: function calcIcon() {
            var _icon = '';

            if (this.icon) return this.icon;

            switch (this.type) {
                case 'success':
                    _icon = 'check-1';
                    break;
                case 'error':
                    _icon = 'close';
                    break;
                case 'warning':
                    _icon = 'exclamation';
                    break;
            }

            return _icon;
        },
        calcStatusOpts: function calcStatusOpts() {
            var opts = {};

            switch (this.type) {
                case 'loading':
                    opts.title = 'loading';
                    opts.subtitle = this.t('infobox.pleasewait');
                    break;
                case 'success':
                    opts.title = this.t('infobox.success');
                    opts.subtitle = this.t('infobox.continue') + '>>';
                    break;
                case 'error':
                    opts.title = this.t('infobox.failure');
                    opts.subtitle = this.t('infobox.closeafter3s');
                    break;
                case 'warning':
                    opts.title = this.t('infobox.riskoperation');
                    break;
            }

            return opts;
        }
    },
    watch: {
        hide: function hide(val) {
            if (val) {
                this.isShow = false;
                this.$el.addEventListener('transitionend', this.destroyEl);
            }
        },
        isShow: function isShow(val) {
            if (val) {
                this.shown && this.shown();
            } else {
                this.hidden && this.hidden();
            }
        }
    },
    mounted: function mounted() {
        this.hide = false;
        if (this.delay) {
            this.startCountDown();
        }
    },
    beforeDestroy: function beforeDestroy() {
        clearTimeout(this.delayId);
    },

    methods: {
        destroyEl: function destroyEl() {
            this.$el.removeEventListener('transitionend', this.destroyEl);
            this.$destroy();
            this.$el.parentNode.removeChild(this.$el);
        },
        close: function close() {
            this.hide = true;
        },
        handleConfirm: function handleConfirm() {
            this.confirmFn && this.confirmFn(this.close);
            this.close();
        },
        handleCancel: function handleCancel() {
            this.cancelFn && this.cancelFn(this.close);
            this.close();
        },
        handleQuickClose: function handleQuickClose() {
            if (this.quickClose) {
                this.close();
            }
        },
        startCountDown: function startCountDown() {
            var _this = this;

            this.delayId = setTimeout(function () {
                _this.close();
            }, this.delay);
        }
    }
};

var InfoBoxConstructor = Vue.extend(InfoBox);
var instance$2 = void 0;
var instancesArr$1 = [];
var count$1 = 0;
var zIndex$1 = new Date().getFullYear();

var Info = function Info() {
    var options = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};

    var id = 'bkInfoBox' + count$1++;

    if (typeof options === 'string') {
        options = {
            title: options
        };
    }

    instance$2 = new InfoBoxConstructor({
        data: options
    });

    if (isVNode(instance$2.content)) {
        instance$2.$slots.content = [instance$2.content];
        instance$2.content = null;
    } else {
        delete instance$2.$slots.content;
    }

    if (isVNode(instance$2.statusOpts.title)) {
        instance$2.$slots.statusTitle = [instance$2.statusOpts.statusTitle];
        instance$2.statusOpts.statusTitle = null;
    } else {
        delete instance$2.$slots.statusTitle;
    }

    if (isVNode(instance$2.statusOpts.subtitle)) {
        instance$2.$slots.statusSubtitle = [instance$2.statusOpts.subtitle];
        instance$2.statusOpts.subtitle = null;
    } else {
        delete instance$2.$slots.statusSubtitle;
    }

    instance$2.id = id;
    instance$2.viewmodel = instance$2.$mount();
    document.body.appendChild(instance$2.viewmodel.$el);
    instance$2.$dom = instance$2.viewmodel.$el;
    instance$2.$dom.style.zIndex = zIndex$1++;
    instance$2.viewmodel.isShow = true;
    instancesArr$1.push(instance$2);
    return instance$2.viewmodel;
};

Info.hide = function () {
    var id = instance$2.id;

    var len = instancesArr$1.length;
    for (var index = 0; index < len; index++) {
        if (id === instancesArr$1[index].id) {
            instance$2.viewmodel.hide = true;
        }
        instancesArr$1.splice(index, 1);
        break;
    }
};
Vue.prototype.$bkInfo = Info;

/**!
 * @fileOverview Kickass library to create and place poppers near their reference elements.
 * @version 1.14.5
 * @license
 * Copyright (c) 2016 Federico Zivolo and contributors
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */
var isBrowser = typeof window !== 'undefined' && typeof document !== 'undefined';

var longerTimeoutBrowsers = ['Edge', 'Trident', 'Firefox'];
var timeoutDuration = 0;
for (var i = 0; i < longerTimeoutBrowsers.length; i += 1) {
  if (isBrowser && navigator.userAgent.indexOf(longerTimeoutBrowsers[i]) >= 0) {
    timeoutDuration = 1;
    break;
  }
}

function microtaskDebounce(fn) {
  var called = false;
  return function () {
    if (called) {
      return;
    }
    called = true;
    window.Promise.resolve().then(function () {
      called = false;
      fn();
    });
  };
}

function taskDebounce(fn) {
  var scheduled = false;
  return function () {
    if (!scheduled) {
      scheduled = true;
      setTimeout(function () {
        scheduled = false;
        fn();
      }, timeoutDuration);
    }
  };
}

var supportsMicroTasks = isBrowser && window.Promise;

/**
* Create a debounced version of a method, that's asynchronously deferred
* but called in the minimum time possible.
*
* @method
* @memberof Popper.Utils
* @argument {Function} fn
* @returns {Function}
*/
var debounce$1 = supportsMicroTasks ? microtaskDebounce : taskDebounce;

/**
 * Check if the given variable is a function
 * @method
 * @memberof Popper.Utils
 * @argument {Any} functionToCheck - variable to check
 * @returns {Boolean} answer to: is a function?
 */
function isFunction(functionToCheck) {
  var getType = {};
  return functionToCheck && getType.toString.call(functionToCheck) === '[object Function]';
}

/**
 * Get CSS computed property of the given element
 * @method
 * @memberof Popper.Utils
 * @argument {Eement} element
 * @argument {String} property
 */
function getStyleComputedProperty(element, property) {
  if (element.nodeType !== 1) {
    return [];
  }
  // NOTE: 1 DOM access here
  var window = element.ownerDocument.defaultView;
  var css = window.getComputedStyle(element, null);
  return property ? css[property] : css;
}

/**
 * Returns the parentNode or the host of the element
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element
 * @returns {Element} parent
 */
function getParentNode(element) {
  if (element.nodeName === 'HTML') {
    return element;
  }
  return element.parentNode || element.host;
}

/**
 * Returns the scrolling parent of the given element
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element
 * @returns {Element} scroll parent
 */
function getScrollParent(element) {
  // Return body, `getScroll` will take care to get the correct `scrollTop` from it
  if (!element) {
    return document.body;
  }

  switch (element.nodeName) {
    case 'HTML':
    case 'BODY':
      return element.ownerDocument.body;
    case '#document':
      return element.body;
  }

  // Firefox want us to check `-x` and `-y` variations as well

  var _getStyleComputedProp = getStyleComputedProperty(element),
      overflow = _getStyleComputedProp.overflow,
      overflowX = _getStyleComputedProp.overflowX,
      overflowY = _getStyleComputedProp.overflowY;

  if (/(auto|scroll|overlay)/.test(overflow + overflowY + overflowX)) {
    return element;
  }

  return getScrollParent(getParentNode(element));
}

var isIE11 = isBrowser && !!(window.MSInputMethodContext && document.documentMode);
var isIE10 = isBrowser && /MSIE 10/.test(navigator.userAgent);

/**
 * Determines if the browser is Internet Explorer
 * @method
 * @memberof Popper.Utils
 * @param {Number} version to check
 * @returns {Boolean} isIE
 */
function isIE(version) {
  if (version === 11) {
    return isIE11;
  }
  if (version === 10) {
    return isIE10;
  }
  return isIE11 || isIE10;
}

/**
 * Returns the offset parent of the given element
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element
 * @returns {Element} offset parent
 */
function getOffsetParent(element) {
  if (!element) {
    return document.documentElement;
  }

  var noOffsetParent = isIE(10) ? document.body : null;

  // NOTE: 1 DOM access here
  var offsetParent = element.offsetParent || null;
  // Skip hidden elements which don't have an offsetParent
  while (offsetParent === noOffsetParent && element.nextElementSibling) {
    offsetParent = (element = element.nextElementSibling).offsetParent;
  }

  var nodeName = offsetParent && offsetParent.nodeName;

  if (!nodeName || nodeName === 'BODY' || nodeName === 'HTML') {
    return element ? element.ownerDocument.documentElement : document.documentElement;
  }

  // .offsetParent will return the closest TH, TD or TABLE in case
  // no offsetParent is present, I hate this job...
  if (['TH', 'TD', 'TABLE'].indexOf(offsetParent.nodeName) !== -1 && getStyleComputedProperty(offsetParent, 'position') === 'static') {
    return getOffsetParent(offsetParent);
  }

  return offsetParent;
}

function isOffsetContainer(element) {
  var nodeName = element.nodeName;

  if (nodeName === 'BODY') {
    return false;
  }
  return nodeName === 'HTML' || getOffsetParent(element.firstElementChild) === element;
}

/**
 * Finds the root node (document, shadowDOM root) of the given element
 * @method
 * @memberof Popper.Utils
 * @argument {Element} node
 * @returns {Element} root node
 */
function getRoot(node) {
  if (node.parentNode !== null) {
    return getRoot(node.parentNode);
  }

  return node;
}

/**
 * Finds the offset parent common to the two provided nodes
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element1
 * @argument {Element} element2
 * @returns {Element} common offset parent
 */
function findCommonOffsetParent(element1, element2) {
  // This check is needed to avoid errors in case one of the elements isn't defined for any reason
  if (!element1 || !element1.nodeType || !element2 || !element2.nodeType) {
    return document.documentElement;
  }

  // Here we make sure to give as "start" the element that comes first in the DOM
  var order = element1.compareDocumentPosition(element2) & Node.DOCUMENT_POSITION_FOLLOWING;
  var start = order ? element1 : element2;
  var end = order ? element2 : element1;

  // Get common ancestor container
  var range = document.createRange();
  range.setStart(start, 0);
  range.setEnd(end, 0);
  var commonAncestorContainer = range.commonAncestorContainer;

  // Both nodes are inside #document

  if (element1 !== commonAncestorContainer && element2 !== commonAncestorContainer || start.contains(end)) {
    if (isOffsetContainer(commonAncestorContainer)) {
      return commonAncestorContainer;
    }

    return getOffsetParent(commonAncestorContainer);
  }

  // one of the nodes is inside shadowDOM, find which one
  var element1root = getRoot(element1);
  if (element1root.host) {
    return findCommonOffsetParent(element1root.host, element2);
  } else {
    return findCommonOffsetParent(element1, getRoot(element2).host);
  }
}

/**
 * Gets the scroll value of the given element in the given side (top and left)
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element
 * @argument {String} side `top` or `left`
 * @returns {number} amount of scrolled pixels
 */
function getScroll(element) {
  var side = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : 'top';

  var upperSide = side === 'top' ? 'scrollTop' : 'scrollLeft';
  var nodeName = element.nodeName;

  if (nodeName === 'BODY' || nodeName === 'HTML') {
    var html = element.ownerDocument.documentElement;
    var scrollingElement = element.ownerDocument.scrollingElement || html;
    return scrollingElement[upperSide];
  }

  return element[upperSide];
}

/*
 * Sum or subtract the element scroll values (left and top) from a given rect object
 * @method
 * @memberof Popper.Utils
 * @param {Object} rect - Rect object you want to change
 * @param {HTMLElement} element - The element from the function reads the scroll values
 * @param {Boolean} subtract - set to true if you want to subtract the scroll values
 * @return {Object} rect - The modifier rect object
 */
function includeScroll(rect, element) {
  var subtract = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : false;

  var scrollTop = getScroll(element, 'top');
  var scrollLeft = getScroll(element, 'left');
  var modifier = subtract ? -1 : 1;
  rect.top += scrollTop * modifier;
  rect.bottom += scrollTop * modifier;
  rect.left += scrollLeft * modifier;
  rect.right += scrollLeft * modifier;
  return rect;
}

/*
 * Helper to detect borders of a given element
 * @method
 * @memberof Popper.Utils
 * @param {CSSStyleDeclaration} styles
 * Result of `getStyleComputedProperty` on the given element
 * @param {String} axis - `x` or `y`
 * @return {number} borders - The borders size of the given axis
 */

function getBordersSize(styles, axis) {
  var sideA = axis === 'x' ? 'Left' : 'Top';
  var sideB = sideA === 'Left' ? 'Right' : 'Bottom';

  return parseFloat(styles['border' + sideA + 'Width'], 10) + parseFloat(styles['border' + sideB + 'Width'], 10);
}

function getSize(axis, body, html, computedStyle) {
  return Math.max(body['offset' + axis], body['scroll' + axis], html['client' + axis], html['offset' + axis], html['scroll' + axis], isIE(10) ? parseInt(html['offset' + axis]) + parseInt(computedStyle['margin' + (axis === 'Height' ? 'Top' : 'Left')]) + parseInt(computedStyle['margin' + (axis === 'Height' ? 'Bottom' : 'Right')]) : 0);
}

function getWindowSizes(document) {
  var body = document.body;
  var html = document.documentElement;
  var computedStyle = isIE(10) && getComputedStyle(html);

  return {
    height: getSize('Height', body, html, computedStyle),
    width: getSize('Width', body, html, computedStyle)
  };
}

var classCallCheck$1 = function (instance, Constructor) {
  if (!(instance instanceof Constructor)) {
    throw new TypeError("Cannot call a class as a function");
  }
};

var createClass$1 = function () {
  function defineProperties(target, props) {
    for (var i = 0; i < props.length; i++) {
      var descriptor = props[i];
      descriptor.enumerable = descriptor.enumerable || false;
      descriptor.configurable = true;
      if ("value" in descriptor) descriptor.writable = true;
      Object.defineProperty(target, descriptor.key, descriptor);
    }
  }

  return function (Constructor, protoProps, staticProps) {
    if (protoProps) defineProperties(Constructor.prototype, protoProps);
    if (staticProps) defineProperties(Constructor, staticProps);
    return Constructor;
  };
}();





var defineProperty$1 = function (obj, key, value) {
  if (key in obj) {
    Object.defineProperty(obj, key, {
      value: value,
      enumerable: true,
      configurable: true,
      writable: true
    });
  } else {
    obj[key] = value;
  }

  return obj;
};

var _extends$1 = Object.assign || function (target) {
  for (var i = 1; i < arguments.length; i++) {
    var source = arguments[i];

    for (var key in source) {
      if (Object.prototype.hasOwnProperty.call(source, key)) {
        target[key] = source[key];
      }
    }
  }

  return target;
};

/**
 * Given element offsets, generate an output similar to getBoundingClientRect
 * @method
 * @memberof Popper.Utils
 * @argument {Object} offsets
 * @returns {Object} ClientRect like output
 */
function getClientRect(offsets) {
  return _extends$1({}, offsets, {
    right: offsets.left + offsets.width,
    bottom: offsets.top + offsets.height
  });
}

/**
 * Get bounding client rect of given element
 * @method
 * @memberof Popper.Utils
 * @param {HTMLElement} element
 * @return {Object} client rect
 */
function getBoundingClientRect(element) {
  var rect = {};

  // IE10 10 FIX: Please, don't ask, the element isn't
  // considered in DOM in some circumstances...
  // This isn't reproducible in IE10 compatibility mode of IE11
  try {
    if (isIE(10)) {
      rect = element.getBoundingClientRect();
      var scrollTop = getScroll(element, 'top');
      var scrollLeft = getScroll(element, 'left');
      rect.top += scrollTop;
      rect.left += scrollLeft;
      rect.bottom += scrollTop;
      rect.right += scrollLeft;
    } else {
      rect = element.getBoundingClientRect();
    }
  } catch (e) {}

  var result = {
    left: rect.left,
    top: rect.top,
    width: rect.right - rect.left,
    height: rect.bottom - rect.top
  };

  // subtract scrollbar size from sizes
  var sizes = element.nodeName === 'HTML' ? getWindowSizes(element.ownerDocument) : {};
  var width = sizes.width || element.clientWidth || result.right - result.left;
  var height = sizes.height || element.clientHeight || result.bottom - result.top;

  var horizScrollbar = element.offsetWidth - width;
  var vertScrollbar = element.offsetHeight - height;

  // if an hypothetical scrollbar is detected, we must be sure it's not a `border`
  // we make this check conditional for performance reasons
  if (horizScrollbar || vertScrollbar) {
    var styles = getStyleComputedProperty(element);
    horizScrollbar -= getBordersSize(styles, 'x');
    vertScrollbar -= getBordersSize(styles, 'y');

    result.width -= horizScrollbar;
    result.height -= vertScrollbar;
  }

  return getClientRect(result);
}

function getOffsetRectRelativeToArbitraryNode(children, parent) {
  var fixedPosition = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : false;

  var isIE10 = isIE(10);
  var isHTML = parent.nodeName === 'HTML';
  var childrenRect = getBoundingClientRect(children);
  var parentRect = getBoundingClientRect(parent);
  var scrollParent = getScrollParent(children);

  var styles = getStyleComputedProperty(parent);
  var borderTopWidth = parseFloat(styles.borderTopWidth, 10);
  var borderLeftWidth = parseFloat(styles.borderLeftWidth, 10);

  // In cases where the parent is fixed, we must ignore negative scroll in offset calc
  if (fixedPosition && isHTML) {
    parentRect.top = Math.max(parentRect.top, 0);
    parentRect.left = Math.max(parentRect.left, 0);
  }
  var offsets = getClientRect({
    top: childrenRect.top - parentRect.top - borderTopWidth,
    left: childrenRect.left - parentRect.left - borderLeftWidth,
    width: childrenRect.width,
    height: childrenRect.height
  });
  offsets.marginTop = 0;
  offsets.marginLeft = 0;

  // Subtract margins of documentElement in case it's being used as parent
  // we do this only on HTML because it's the only element that behaves
  // differently when margins are applied to it. The margins are included in
  // the box of the documentElement, in the other cases not.
  if (!isIE10 && isHTML) {
    var marginTop = parseFloat(styles.marginTop, 10);
    var marginLeft = parseFloat(styles.marginLeft, 10);

    offsets.top -= borderTopWidth - marginTop;
    offsets.bottom -= borderTopWidth - marginTop;
    offsets.left -= borderLeftWidth - marginLeft;
    offsets.right -= borderLeftWidth - marginLeft;

    // Attach marginTop and marginLeft because in some circumstances we may need them
    offsets.marginTop = marginTop;
    offsets.marginLeft = marginLeft;
  }

  if (isIE10 && !fixedPosition ? parent.contains(scrollParent) : parent === scrollParent && scrollParent.nodeName !== 'BODY') {
    offsets = includeScroll(offsets, parent);
  }

  return offsets;
}

function getViewportOffsetRectRelativeToArtbitraryNode(element) {
  var excludeScroll = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : false;

  var html = element.ownerDocument.documentElement;
  var relativeOffset = getOffsetRectRelativeToArbitraryNode(element, html);
  var width = Math.max(html.clientWidth, window.innerWidth || 0);
  var height = Math.max(html.clientHeight, window.innerHeight || 0);

  var scrollTop = !excludeScroll ? getScroll(html) : 0;
  var scrollLeft = !excludeScroll ? getScroll(html, 'left') : 0;

  var offset = {
    top: scrollTop - relativeOffset.top + relativeOffset.marginTop,
    left: scrollLeft - relativeOffset.left + relativeOffset.marginLeft,
    width: width,
    height: height
  };

  return getClientRect(offset);
}

/**
 * Check if the given element is fixed or is inside a fixed parent
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element
 * @argument {Element} customContainer
 * @returns {Boolean} answer to "isFixed?"
 */
function isFixed(element) {
  var nodeName = element.nodeName;
  if (nodeName === 'BODY' || nodeName === 'HTML') {
    return false;
  }
  if (getStyleComputedProperty(element, 'position') === 'fixed') {
    return true;
  }
  return isFixed(getParentNode(element));
}

/**
 * Finds the first parent of an element that has a transformed property defined
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element
 * @returns {Element} first transformed parent or documentElement
 */

function getFixedPositionOffsetParent(element) {
  // This check is needed to avoid errors in case one of the elements isn't defined for any reason
  if (!element || !element.parentElement || isIE()) {
    return document.documentElement;
  }
  var el = element.parentElement;
  while (el && getStyleComputedProperty(el, 'transform') === 'none') {
    el = el.parentElement;
  }
  return el || document.documentElement;
}

/**
 * Computed the boundaries limits and return them
 * @method
 * @memberof Popper.Utils
 * @param {HTMLElement} popper
 * @param {HTMLElement} reference
 * @param {number} padding
 * @param {HTMLElement} boundariesElement - Element used to define the boundaries
 * @param {Boolean} fixedPosition - Is in fixed position mode
 * @returns {Object} Coordinates of the boundaries
 */
function getBoundaries(popper, reference, padding, boundariesElement) {
  var fixedPosition = arguments.length > 4 && arguments[4] !== undefined ? arguments[4] : false;

  // NOTE: 1 DOM access here

  var boundaries = { top: 0, left: 0 };
  var offsetParent = fixedPosition ? getFixedPositionOffsetParent(popper) : findCommonOffsetParent(popper, reference);

  // Handle viewport case
  if (boundariesElement === 'viewport') {
    boundaries = getViewportOffsetRectRelativeToArtbitraryNode(offsetParent, fixedPosition);
  } else {
    // Handle other cases based on DOM element used as boundaries
    var boundariesNode = void 0;
    if (boundariesElement === 'scrollParent') {
      boundariesNode = getScrollParent(getParentNode(reference));
      if (boundariesNode.nodeName === 'BODY') {
        boundariesNode = popper.ownerDocument.documentElement;
      }
    } else if (boundariesElement === 'window') {
      boundariesNode = popper.ownerDocument.documentElement;
    } else {
      boundariesNode = boundariesElement;
    }

    var offsets = getOffsetRectRelativeToArbitraryNode(boundariesNode, offsetParent, fixedPosition);

    // In case of HTML, we need a different computation
    if (boundariesNode.nodeName === 'HTML' && !isFixed(offsetParent)) {
      var _getWindowSizes = getWindowSizes(popper.ownerDocument),
          height = _getWindowSizes.height,
          width = _getWindowSizes.width;

      boundaries.top += offsets.top - offsets.marginTop;
      boundaries.bottom = height + offsets.top;
      boundaries.left += offsets.left - offsets.marginLeft;
      boundaries.right = width + offsets.left;
    } else {
      // for all the other DOM elements, this one is good
      boundaries = offsets;
    }
  }

  // Add paddings
  padding = padding || 0;
  var isPaddingNumber = typeof padding === 'number';
  boundaries.left += isPaddingNumber ? padding : padding.left || 0;
  boundaries.top += isPaddingNumber ? padding : padding.top || 0;
  boundaries.right -= isPaddingNumber ? padding : padding.right || 0;
  boundaries.bottom -= isPaddingNumber ? padding : padding.bottom || 0;

  return boundaries;
}

function getArea(_ref) {
  var width = _ref.width,
      height = _ref.height;

  return width * height;
}

/**
 * Utility used to transform the `auto` placement to the placement with more
 * available space.
 * @method
 * @memberof Popper.Utils
 * @argument {Object} data - The data object generated by update method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function computeAutoPlacement(placement, refRect, popper, reference, boundariesElement) {
  var padding = arguments.length > 5 && arguments[5] !== undefined ? arguments[5] : 0;

  if (placement.indexOf('auto') === -1) {
    return placement;
  }

  var boundaries = getBoundaries(popper, reference, padding, boundariesElement);

  var rects = {
    top: {
      width: boundaries.width,
      height: refRect.top - boundaries.top
    },
    right: {
      width: boundaries.right - refRect.right,
      height: boundaries.height
    },
    bottom: {
      width: boundaries.width,
      height: boundaries.bottom - refRect.bottom
    },
    left: {
      width: refRect.left - boundaries.left,
      height: boundaries.height
    }
  };

  var sortedAreas = Object.keys(rects).map(function (key) {
    return _extends$1({
      key: key
    }, rects[key], {
      area: getArea(rects[key])
    });
  }).sort(function (a, b) {
    return b.area - a.area;
  });

  var filteredAreas = sortedAreas.filter(function (_ref2) {
    var width = _ref2.width,
        height = _ref2.height;
    return width >= popper.clientWidth && height >= popper.clientHeight;
  });

  var computedPlacement = filteredAreas.length > 0 ? filteredAreas[0].key : sortedAreas[0].key;

  var variation = placement.split('-')[1];

  return computedPlacement + (variation ? '-' + variation : '');
}

/**
 * Get offsets to the reference element
 * @method
 * @memberof Popper.Utils
 * @param {Object} state
 * @param {Element} popper - the popper element
 * @param {Element} reference - the reference element (the popper will be relative to this)
 * @param {Element} fixedPosition - is in fixed position mode
 * @returns {Object} An object containing the offsets which will be applied to the popper
 */
function getReferenceOffsets(state, popper, reference) {
  var fixedPosition = arguments.length > 3 && arguments[3] !== undefined ? arguments[3] : null;

  var commonOffsetParent = fixedPosition ? getFixedPositionOffsetParent(popper) : findCommonOffsetParent(popper, reference);
  return getOffsetRectRelativeToArbitraryNode(reference, commonOffsetParent, fixedPosition);
}

/**
 * Get the outer sizes of the given element (offset size + margins)
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element
 * @returns {Object} object containing width and height properties
 */
function getOuterSizes(element) {
  var window = element.ownerDocument.defaultView;
  var styles = window.getComputedStyle(element);
  var x = parseFloat(styles.marginTop) + parseFloat(styles.marginBottom);
  var y = parseFloat(styles.marginLeft) + parseFloat(styles.marginRight);
  var result = {
    width: element.offsetWidth + y,
    height: element.offsetHeight + x
  };
  return result;
}

/**
 * Get the opposite placement of the given one
 * @method
 * @memberof Popper.Utils
 * @argument {String} placement
 * @returns {String} flipped placement
 */
function getOppositePlacement(placement) {
  var hash = { left: 'right', right: 'left', bottom: 'top', top: 'bottom' };
  return placement.replace(/left|right|bottom|top/g, function (matched) {
    return hash[matched];
  });
}

/**
 * Get offsets to the popper
 * @method
 * @memberof Popper.Utils
 * @param {Object} position - CSS position the Popper will get applied
 * @param {HTMLElement} popper - the popper element
 * @param {Object} referenceOffsets - the reference offsets (the popper will be relative to this)
 * @param {String} placement - one of the valid placement options
 * @returns {Object} popperOffsets - An object containing the offsets which will be applied to the popper
 */
function getPopperOffsets(popper, referenceOffsets, placement) {
  placement = placement.split('-')[0];

  // Get popper node sizes
  var popperRect = getOuterSizes(popper);

  // Add position, width and height to our offsets object
  var popperOffsets = {
    width: popperRect.width,
    height: popperRect.height
  };

  // depending by the popper placement we have to compute its offsets slightly differently
  var isHoriz = ['right', 'left'].indexOf(placement) !== -1;
  var mainSide = isHoriz ? 'top' : 'left';
  var secondarySide = isHoriz ? 'left' : 'top';
  var measurement = isHoriz ? 'height' : 'width';
  var secondaryMeasurement = !isHoriz ? 'height' : 'width';

  popperOffsets[mainSide] = referenceOffsets[mainSide] + referenceOffsets[measurement] / 2 - popperRect[measurement] / 2;
  if (placement === secondarySide) {
    popperOffsets[secondarySide] = referenceOffsets[secondarySide] - popperRect[secondaryMeasurement];
  } else {
    popperOffsets[secondarySide] = referenceOffsets[getOppositePlacement(secondarySide)];
  }

  return popperOffsets;
}

/**
 * Mimics the `find` method of Array
 * @method
 * @memberof Popper.Utils
 * @argument {Array} arr
 * @argument prop
 * @argument value
 * @returns index or -1
 */
function find(arr, check) {
  // use native find if supported
  if (Array.prototype.find) {
    return arr.find(check);
  }

  // use `filter` to obtain the same behavior of `find`
  return arr.filter(check)[0];
}

/**
 * Return the index of the matching object
 * @method
 * @memberof Popper.Utils
 * @argument {Array} arr
 * @argument prop
 * @argument value
 * @returns index or -1
 */
function findIndex(arr, prop, value) {
  // use native findIndex if supported
  if (Array.prototype.findIndex) {
    return arr.findIndex(function (cur) {
      return cur[prop] === value;
    });
  }

  // use `find` + `indexOf` if `findIndex` isn't supported
  var match = find(arr, function (obj) {
    return obj[prop] === value;
  });
  return arr.indexOf(match);
}

/**
 * Loop trough the list of modifiers and run them in order,
 * each of them will then edit the data object.
 * @method
 * @memberof Popper.Utils
 * @param {dataObject} data
 * @param {Array} modifiers
 * @param {String} ends - Optional modifier name used as stopper
 * @returns {dataObject}
 */
function runModifiers(modifiers, data, ends) {
  var modifiersToRun = ends === undefined ? modifiers : modifiers.slice(0, findIndex(modifiers, 'name', ends));

  modifiersToRun.forEach(function (modifier) {
    if (modifier['function']) {
      // eslint-disable-line dot-notation
      console.warn('`modifier.function` is deprecated, use `modifier.fn`!');
    }
    var fn = modifier['function'] || modifier.fn; // eslint-disable-line dot-notation
    if (modifier.enabled && isFunction(fn)) {
      // Add properties to offsets to make them a complete clientRect object
      // we do this before each modifier to make sure the previous one doesn't
      // mess with these values
      data.offsets.popper = getClientRect(data.offsets.popper);
      data.offsets.reference = getClientRect(data.offsets.reference);

      data = fn(data, modifier);
    }
  });

  return data;
}

/**
 * Updates the position of the popper, computing the new offsets and applying
 * the new style.<br />
 * Prefer `scheduleUpdate` over `update` because of performance reasons.
 * @method
 * @memberof Popper
 */
function update() {
  // if popper is destroyed, don't perform any further update
  if (this.state.isDestroyed) {
    return;
  }

  var data = {
    instance: this,
    styles: {},
    arrowStyles: {},
    attributes: {},
    flipped: false,
    offsets: {}
  };

  // compute reference element offsets
  data.offsets.reference = getReferenceOffsets(this.state, this.popper, this.reference, this.options.positionFixed);

  // compute auto placement, store placement inside the data object,
  // modifiers will be able to edit `placement` if needed
  // and refer to originalPlacement to know the original value
  data.placement = computeAutoPlacement(this.options.placement, data.offsets.reference, this.popper, this.reference, this.options.modifiers.flip.boundariesElement, this.options.modifiers.flip.padding);

  // store the computed placement inside `originalPlacement`
  data.originalPlacement = data.placement;

  data.positionFixed = this.options.positionFixed;

  // compute the popper offsets
  data.offsets.popper = getPopperOffsets(this.popper, data.offsets.reference, data.placement);

  data.offsets.popper.position = this.options.positionFixed ? 'fixed' : 'absolute';

  // run the modifiers
  data = runModifiers(this.modifiers, data);

  // the first `update` will call `onCreate` callback
  // the other ones will call `onUpdate` callback
  if (!this.state.isCreated) {
    this.state.isCreated = true;
    this.options.onCreate(data);
  } else {
    this.options.onUpdate(data);
  }
}

/**
 * Helper used to know if the given modifier is enabled.
 * @method
 * @memberof Popper.Utils
 * @returns {Boolean}
 */
function isModifierEnabled(modifiers, modifierName) {
  return modifiers.some(function (_ref) {
    var name = _ref.name,
        enabled = _ref.enabled;
    return enabled && name === modifierName;
  });
}

/**
 * Get the prefixed supported property name
 * @method
 * @memberof Popper.Utils
 * @argument {String} property (camelCase)
 * @returns {String} prefixed property (camelCase or PascalCase, depending on the vendor prefix)
 */
function getSupportedPropertyName(property) {
  var prefixes = [false, 'ms', 'Webkit', 'Moz', 'O'];
  var upperProp = property.charAt(0).toUpperCase() + property.slice(1);

  for (var i = 0; i < prefixes.length; i++) {
    var prefix = prefixes[i];
    var toCheck = prefix ? '' + prefix + upperProp : property;
    if (typeof document.body.style[toCheck] !== 'undefined') {
      return toCheck;
    }
  }
  return null;
}

/**
 * Destroys the popper.
 * @method
 * @memberof Popper
 */
function destroy() {
  this.state.isDestroyed = true;

  // touch DOM only if `applyStyle` modifier is enabled
  if (isModifierEnabled(this.modifiers, 'applyStyle')) {
    this.popper.removeAttribute('x-placement');
    this.popper.style.position = '';
    this.popper.style.top = '';
    this.popper.style.left = '';
    this.popper.style.right = '';
    this.popper.style.bottom = '';
    this.popper.style.willChange = '';
    this.popper.style[getSupportedPropertyName('transform')] = '';
  }

  this.disableEventListeners();

  // remove the popper if user explicity asked for the deletion on destroy
  // do not use `remove` because IE11 doesn't support it
  if (this.options.removeOnDestroy) {
    this.popper.parentNode.removeChild(this.popper);
  }
  return this;
}

/**
 * Get the window associated with the element
 * @argument {Element} element
 * @returns {Window}
 */
function getWindow(element) {
  var ownerDocument = element.ownerDocument;
  return ownerDocument ? ownerDocument.defaultView : window;
}

function attachToScrollParents(scrollParent, event, callback, scrollParents) {
  var isBody = scrollParent.nodeName === 'BODY';
  var target = isBody ? scrollParent.ownerDocument.defaultView : scrollParent;
  target.addEventListener(event, callback, { passive: true });

  if (!isBody) {
    attachToScrollParents(getScrollParent(target.parentNode), event, callback, scrollParents);
  }
  scrollParents.push(target);
}

/**
 * Setup needed event listeners used to update the popper position
 * @method
 * @memberof Popper.Utils
 * @private
 */
function setupEventListeners(reference, options, state, updateBound) {
  // Resize event listener on window
  state.updateBound = updateBound;
  getWindow(reference).addEventListener('resize', state.updateBound, { passive: true });

  // Scroll event listener on scroll parents
  var scrollElement = getScrollParent(reference);
  attachToScrollParents(scrollElement, 'scroll', state.updateBound, state.scrollParents);
  state.scrollElement = scrollElement;
  state.eventsEnabled = true;

  return state;
}

/**
 * It will add resize/scroll events and start recalculating
 * position of the popper element when they are triggered.
 * @method
 * @memberof Popper
 */
function enableEventListeners() {
  if (!this.state.eventsEnabled) {
    this.state = setupEventListeners(this.reference, this.options, this.state, this.scheduleUpdate);
  }
}

/**
 * Remove event listeners used to update the popper position
 * @method
 * @memberof Popper.Utils
 * @private
 */
function removeEventListeners(reference, state) {
  // Remove resize event listener on window
  getWindow(reference).removeEventListener('resize', state.updateBound);

  // Remove scroll event listener on scroll parents
  state.scrollParents.forEach(function (target) {
    target.removeEventListener('scroll', state.updateBound);
  });

  // Reset state
  state.updateBound = null;
  state.scrollParents = [];
  state.scrollElement = null;
  state.eventsEnabled = false;
  return state;
}

/**
 * It will remove resize/scroll events and won't recalculate popper position
 * when they are triggered. It also won't trigger `onUpdate` callback anymore,
 * unless you call `update` method manually.
 * @method
 * @memberof Popper
 */
function disableEventListeners() {
  if (this.state.eventsEnabled) {
    cancelAnimationFrame(this.scheduleUpdate);
    this.state = removeEventListeners(this.reference, this.state);
  }
}

/**
 * Tells if a given input is a number
 * @method
 * @memberof Popper.Utils
 * @param {*} input to check
 * @return {Boolean}
 */
function isNumeric(n) {
  return n !== '' && !isNaN(parseFloat(n)) && isFinite(n);
}

/**
 * Set the style to the given popper
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element - Element to apply the style to
 * @argument {Object} styles
 * Object with a list of properties and values which will be applied to the element
 */
function setStyles(element, styles) {
  Object.keys(styles).forEach(function (prop) {
    var unit = '';
    // add unit if the value is numeric and is one of the following
    if (['width', 'height', 'top', 'right', 'bottom', 'left'].indexOf(prop) !== -1 && isNumeric(styles[prop])) {
      unit = 'px';
    }
    element.style[prop] = styles[prop] + unit;
  });
}

/**
 * Set the attributes to the given popper
 * @method
 * @memberof Popper.Utils
 * @argument {Element} element - Element to apply the attributes to
 * @argument {Object} styles
 * Object with a list of properties and values which will be applied to the element
 */
function setAttributes(element, attributes) {
  Object.keys(attributes).forEach(function (prop) {
    var value = attributes[prop];
    if (value !== false) {
      element.setAttribute(prop, attributes[prop]);
    } else {
      element.removeAttribute(prop);
    }
  });
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by `update` method
 * @argument {Object} data.styles - List of style properties - values to apply to popper element
 * @argument {Object} data.attributes - List of attribute properties - values to apply to popper element
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The same data object
 */
function applyStyle(data) {
  // any property present in `data.styles` will be applied to the popper,
  // in this way we can make the 3rd party modifiers add custom styles to it
  // Be aware, modifiers could override the properties defined in the previous
  // lines of this modifier!
  setStyles(data.instance.popper, data.styles);

  // any property present in `data.attributes` will be applied to the popper,
  // they will be set as HTML attributes of the element
  setAttributes(data.instance.popper, data.attributes);

  // if arrowElement is defined and arrowStyles has some properties
  if (data.arrowElement && Object.keys(data.arrowStyles).length) {
    setStyles(data.arrowElement, data.arrowStyles);
  }

  return data;
}

/**
 * Set the x-placement attribute before everything else because it could be used
 * to add margins to the popper margins needs to be calculated to get the
 * correct popper offsets.
 * @method
 * @memberof Popper.modifiers
 * @param {HTMLElement} reference - The reference element used to position the popper
 * @param {HTMLElement} popper - The HTML element used as popper
 * @param {Object} options - Popper.js options
 */
function applyStyleOnLoad(reference, popper, options, modifierOptions, state) {
  // compute reference element offsets
  var referenceOffsets = getReferenceOffsets(state, popper, reference, options.positionFixed);

  // compute auto placement, store placement inside the data object,
  // modifiers will be able to edit `placement` if needed
  // and refer to originalPlacement to know the original value
  var placement = computeAutoPlacement(options.placement, referenceOffsets, popper, reference, options.modifiers.flip.boundariesElement, options.modifiers.flip.padding);

  popper.setAttribute('x-placement', placement);

  // Apply `position` to popper before anything else because
  // without the position applied we can't guarantee correct computations
  setStyles(popper, { position: options.positionFixed ? 'fixed' : 'absolute' });

  return options;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by `update` method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function computeStyle(data, options) {
  var x = options.x,
      y = options.y;
  var popper = data.offsets.popper;

  // Remove this legacy support in Popper.js v2

  var legacyGpuAccelerationOption = find(data.instance.modifiers, function (modifier) {
    return modifier.name === 'applyStyle';
  }).gpuAcceleration;
  if (legacyGpuAccelerationOption !== undefined) {
    console.warn('WARNING: `gpuAcceleration` option moved to `computeStyle` modifier and will not be supported in future versions of Popper.js!');
  }
  var gpuAcceleration = legacyGpuAccelerationOption !== undefined ? legacyGpuAccelerationOption : options.gpuAcceleration;

  var offsetParent = getOffsetParent(data.instance.popper);
  var offsetParentRect = getBoundingClientRect(offsetParent);

  // Styles
  var styles = {
    position: popper.position
  };

  // Avoid blurry text by using full pixel integers.
  // For pixel-perfect positioning, top/bottom prefers rounded
  // values, while left/right prefers floored values.
  var offsets = {
    left: Math.floor(popper.left),
    top: Math.round(popper.top),
    bottom: Math.round(popper.bottom),
    right: Math.floor(popper.right)
  };

  var sideA = x === 'bottom' ? 'top' : 'bottom';
  var sideB = y === 'right' ? 'left' : 'right';

  // if gpuAcceleration is set to `true` and transform is supported,
  //  we use `translate3d` to apply the position to the popper we
  // automatically use the supported prefixed version if needed
  var prefixedProperty = getSupportedPropertyName('transform');

  // now, let's make a step back and look at this code closely (wtf?)
  // If the content of the popper grows once it's been positioned, it
  // may happen that the popper gets misplaced because of the new content
  // overflowing its reference element
  // To avoid this problem, we provide two options (x and y), which allow
  // the consumer to define the offset origin.
  // If we position a popper on top of a reference element, we can set
  // `x` to `top` to make the popper grow towards its top instead of
  // its bottom.
  var left = void 0,
      top = void 0;
  if (sideA === 'bottom') {
    // when offsetParent is <html> the positioning is relative to the bottom of the screen (excluding the scrollbar)
    // and not the bottom of the html element
    if (offsetParent.nodeName === 'HTML') {
      top = -offsetParent.clientHeight + offsets.bottom;
    } else {
      top = -offsetParentRect.height + offsets.bottom;
    }
  } else {
    top = offsets.top;
  }
  if (sideB === 'right') {
    if (offsetParent.nodeName === 'HTML') {
      left = -offsetParent.clientWidth + offsets.right;
    } else {
      left = -offsetParentRect.width + offsets.right;
    }
  } else {
    left = offsets.left;
  }
  if (gpuAcceleration && prefixedProperty) {
    styles[prefixedProperty] = 'translate3d(' + left + 'px, ' + top + 'px, 0)';
    styles[sideA] = 0;
    styles[sideB] = 0;
    styles.willChange = 'transform';
  } else {
    // othwerise, we use the standard `top`, `left`, `bottom` and `right` properties
    var invertTop = sideA === 'bottom' ? -1 : 1;
    var invertLeft = sideB === 'right' ? -1 : 1;
    styles[sideA] = top * invertTop;
    styles[sideB] = left * invertLeft;
    styles.willChange = sideA + ', ' + sideB;
  }

  // Attributes
  var attributes = {
    'x-placement': data.placement
  };

  // Update `data` attributes, styles and arrowStyles
  data.attributes = _extends$1({}, attributes, data.attributes);
  data.styles = _extends$1({}, styles, data.styles);
  data.arrowStyles = _extends$1({}, data.offsets.arrow, data.arrowStyles);

  return data;
}

/**
 * Helper used to know if the given modifier depends from another one.<br />
 * It checks if the needed modifier is listed and enabled.
 * @method
 * @memberof Popper.Utils
 * @param {Array} modifiers - list of modifiers
 * @param {String} requestingName - name of requesting modifier
 * @param {String} requestedName - name of requested modifier
 * @returns {Boolean}
 */
function isModifierRequired(modifiers, requestingName, requestedName) {
  var requesting = find(modifiers, function (_ref) {
    var name = _ref.name;
    return name === requestingName;
  });

  var isRequired = !!requesting && modifiers.some(function (modifier) {
    return modifier.name === requestedName && modifier.enabled && modifier.order < requesting.order;
  });

  if (!isRequired) {
    var _requesting = '`' + requestingName + '`';
    var requested = '`' + requestedName + '`';
    console.warn(requested + ' modifier is required by ' + _requesting + ' modifier in order to work, be sure to include it before ' + _requesting + '!');
  }
  return isRequired;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by update method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function arrow(data, options) {
  var _data$offsets$arrow;

  // arrow depends on keepTogether in order to work
  if (!isModifierRequired(data.instance.modifiers, 'arrow', 'keepTogether')) {
    return data;
  }

  var arrowElement = options.element;

  // if arrowElement is a string, suppose it's a CSS selector
  if (typeof arrowElement === 'string') {
    arrowElement = data.instance.popper.querySelector(arrowElement);

    // if arrowElement is not found, don't run the modifier
    if (!arrowElement) {
      return data;
    }
  } else {
    // if the arrowElement isn't a query selector we must check that the
    // provided DOM node is child of its popper node
    if (!data.instance.popper.contains(arrowElement)) {
      console.warn('WARNING: `arrow.element` must be child of its popper element!');
      return data;
    }
  }

  var placement = data.placement.split('-')[0];
  var _data$offsets = data.offsets,
      popper = _data$offsets.popper,
      reference = _data$offsets.reference;

  var isVertical = ['left', 'right'].indexOf(placement) !== -1;

  var len = isVertical ? 'height' : 'width';
  var sideCapitalized = isVertical ? 'Top' : 'Left';
  var side = sideCapitalized.toLowerCase();
  var altSide = isVertical ? 'left' : 'top';
  var opSide = isVertical ? 'bottom' : 'right';
  var arrowElementSize = getOuterSizes(arrowElement)[len];

  //
  // extends keepTogether behavior making sure the popper and its
  // reference have enough pixels in conjunction
  //

  // top/left side
  if (reference[opSide] - arrowElementSize < popper[side]) {
    data.offsets.popper[side] -= popper[side] - (reference[opSide] - arrowElementSize);
  }
  // bottom/right side
  if (reference[side] + arrowElementSize > popper[opSide]) {
    data.offsets.popper[side] += reference[side] + arrowElementSize - popper[opSide];
  }
  data.offsets.popper = getClientRect(data.offsets.popper);

  // compute center of the popper
  var center = reference[side] + reference[len] / 2 - arrowElementSize / 2;

  // Compute the sideValue using the updated popper offsets
  // take popper margin in account because we don't have this info available
  var css = getStyleComputedProperty(data.instance.popper);
  var popperMarginSide = parseFloat(css['margin' + sideCapitalized], 10);
  var popperBorderSide = parseFloat(css['border' + sideCapitalized + 'Width'], 10);
  var sideValue = center - data.offsets.popper[side] - popperMarginSide - popperBorderSide;

  // prevent arrowElement from being placed not contiguously to its popper
  sideValue = Math.max(Math.min(popper[len] - arrowElementSize, sideValue), 0);

  data.arrowElement = arrowElement;
  data.offsets.arrow = (_data$offsets$arrow = {}, defineProperty$1(_data$offsets$arrow, side, Math.round(sideValue)), defineProperty$1(_data$offsets$arrow, altSide, ''), _data$offsets$arrow);

  return data;
}

/**
 * Get the opposite placement variation of the given one
 * @method
 * @memberof Popper.Utils
 * @argument {String} placement variation
 * @returns {String} flipped placement variation
 */
function getOppositeVariation(variation) {
  if (variation === 'end') {
    return 'start';
  } else if (variation === 'start') {
    return 'end';
  }
  return variation;
}

/**
 * List of accepted placements to use as values of the `placement` option.<br />
 * Valid placements are:
 * - `auto`
 * - `top`
 * - `right`
 * - `bottom`
 * - `left`
 *
 * Each placement can have a variation from this list:
 * - `-start`
 * - `-end`
 *
 * Variations are interpreted easily if you think of them as the left to right
 * written languages. Horizontally (`top` and `bottom`), `start` is left and `end`
 * is right.<br />
 * Vertically (`left` and `right`), `start` is top and `end` is bottom.
 *
 * Some valid examples are:
 * - `top-end` (on top of reference, right aligned)
 * - `right-start` (on right of reference, top aligned)
 * - `bottom` (on bottom, centered)
 * - `auto-end` (on the side with more space available, alignment depends by placement)
 *
 * @static
 * @type {Array}
 * @enum {String}
 * @readonly
 * @method placements
 * @memberof Popper
 */
var placements = ['auto-start', 'auto', 'auto-end', 'top-start', 'top', 'top-end', 'right-start', 'right', 'right-end', 'bottom-end', 'bottom', 'bottom-start', 'left-end', 'left', 'left-start'];

// Get rid of `auto` `auto-start` and `auto-end`
var validPlacements = placements.slice(3);

/**
 * Given an initial placement, returns all the subsequent placements
 * clockwise (or counter-clockwise).
 *
 * @method
 * @memberof Popper.Utils
 * @argument {String} placement - A valid placement (it accepts variations)
 * @argument {Boolean} counter - Set to true to walk the placements counterclockwise
 * @returns {Array} placements including their variations
 */
function clockwise(placement) {
  var counter = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : false;

  var index = validPlacements.indexOf(placement);
  var arr = validPlacements.slice(index + 1).concat(validPlacements.slice(0, index));
  return counter ? arr.reverse() : arr;
}

var BEHAVIORS = {
  FLIP: 'flip',
  CLOCKWISE: 'clockwise',
  COUNTERCLOCKWISE: 'counterclockwise'
};

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by update method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function flip(data, options) {
  // if `inner` modifier is enabled, we can't use the `flip` modifier
  if (isModifierEnabled(data.instance.modifiers, 'inner')) {
    return data;
  }

  if (data.flipped && data.placement === data.originalPlacement) {
    // seems like flip is trying to loop, probably there's not enough space on any of the flippable sides
    return data;
  }

  var boundaries = getBoundaries(data.instance.popper, data.instance.reference, options.padding, options.boundariesElement, data.positionFixed);

  var placement = data.placement.split('-')[0];
  var placementOpposite = getOppositePlacement(placement);
  var variation = data.placement.split('-')[1] || '';

  var flipOrder = [];

  switch (options.behavior) {
    case BEHAVIORS.FLIP:
      flipOrder = [placement, placementOpposite];
      break;
    case BEHAVIORS.CLOCKWISE:
      flipOrder = clockwise(placement);
      break;
    case BEHAVIORS.COUNTERCLOCKWISE:
      flipOrder = clockwise(placement, true);
      break;
    default:
      flipOrder = options.behavior;
  }

  flipOrder.forEach(function (step, index) {
    if (placement !== step || flipOrder.length === index + 1) {
      return data;
    }

    placement = data.placement.split('-')[0];
    placementOpposite = getOppositePlacement(placement);

    var popperOffsets = data.offsets.popper;
    var refOffsets = data.offsets.reference;

    // using floor because the reference offsets may contain decimals we are not going to consider here
    var floor = Math.floor;
    var overlapsRef = placement === 'left' && floor(popperOffsets.right) > floor(refOffsets.left) || placement === 'right' && floor(popperOffsets.left) < floor(refOffsets.right) || placement === 'top' && floor(popperOffsets.bottom) > floor(refOffsets.top) || placement === 'bottom' && floor(popperOffsets.top) < floor(refOffsets.bottom);

    var overflowsLeft = floor(popperOffsets.left) < floor(boundaries.left);
    var overflowsRight = floor(popperOffsets.right) > floor(boundaries.right);
    var overflowsTop = floor(popperOffsets.top) < floor(boundaries.top);
    var overflowsBottom = floor(popperOffsets.bottom) > floor(boundaries.bottom);

    var overflowsBoundaries = placement === 'left' && overflowsLeft || placement === 'right' && overflowsRight || placement === 'top' && overflowsTop || placement === 'bottom' && overflowsBottom;

    // flip the variation if required
    var isVertical = ['top', 'bottom'].indexOf(placement) !== -1;
    var flippedVariation = !!options.flipVariations && (isVertical && variation === 'start' && overflowsLeft || isVertical && variation === 'end' && overflowsRight || !isVertical && variation === 'start' && overflowsTop || !isVertical && variation === 'end' && overflowsBottom);

    if (overlapsRef || overflowsBoundaries || flippedVariation) {
      // this boolean to detect any flip loop
      data.flipped = true;

      if (overlapsRef || overflowsBoundaries) {
        placement = flipOrder[index + 1];
      }

      if (flippedVariation) {
        variation = getOppositeVariation(variation);
      }

      data.placement = placement + (variation ? '-' + variation : '');

      // this object contains `position`, we want to preserve it along with
      // any additional property we may add in the future
      data.offsets.popper = _extends$1({}, data.offsets.popper, getPopperOffsets(data.instance.popper, data.offsets.reference, data.placement));

      data = runModifiers(data.instance.modifiers, data, 'flip');
    }
  });
  return data;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by update method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function keepTogether(data) {
  var _data$offsets = data.offsets,
      popper = _data$offsets.popper,
      reference = _data$offsets.reference;

  var placement = data.placement.split('-')[0];
  var floor = Math.floor;
  var isVertical = ['top', 'bottom'].indexOf(placement) !== -1;
  var side = isVertical ? 'right' : 'bottom';
  var opSide = isVertical ? 'left' : 'top';
  var measurement = isVertical ? 'width' : 'height';

  if (popper[side] < floor(reference[opSide])) {
    data.offsets.popper[opSide] = floor(reference[opSide]) - popper[measurement];
  }
  if (popper[opSide] > floor(reference[side])) {
    data.offsets.popper[opSide] = floor(reference[side]);
  }

  return data;
}

/**
 * Converts a string containing value + unit into a px value number
 * @function
 * @memberof {modifiers~offset}
 * @private
 * @argument {String} str - Value + unit string
 * @argument {String} measurement - `height` or `width`
 * @argument {Object} popperOffsets
 * @argument {Object} referenceOffsets
 * @returns {Number|String}
 * Value in pixels, or original string if no values were extracted
 */
function toValue(str, measurement, popperOffsets, referenceOffsets) {
  // separate value from unit
  var split = str.match(/((?:\-|\+)?\d*\.?\d*)(.*)/);
  var value = +split[1];
  var unit = split[2];

  // If it's not a number it's an operator, I guess
  if (!value) {
    return str;
  }

  if (unit.indexOf('%') === 0) {
    var element = void 0;
    switch (unit) {
      case '%p':
        element = popperOffsets;
        break;
      case '%':
      case '%r':
      default:
        element = referenceOffsets;
    }

    var rect = getClientRect(element);
    return rect[measurement] / 100 * value;
  } else if (unit === 'vh' || unit === 'vw') {
    // if is a vh or vw, we calculate the size based on the viewport
    var size = void 0;
    if (unit === 'vh') {
      size = Math.max(document.documentElement.clientHeight, window.innerHeight || 0);
    } else {
      size = Math.max(document.documentElement.clientWidth, window.innerWidth || 0);
    }
    return size / 100 * value;
  } else {
    // if is an explicit pixel unit, we get rid of the unit and keep the value
    // if is an implicit unit, it's px, and we return just the value
    return value;
  }
}

/**
 * Parse an `offset` string to extrapolate `x` and `y` numeric offsets.
 * @function
 * @memberof {modifiers~offset}
 * @private
 * @argument {String} offset
 * @argument {Object} popperOffsets
 * @argument {Object} referenceOffsets
 * @argument {String} basePlacement
 * @returns {Array} a two cells array with x and y offsets in numbers
 */
function parseOffset(offset, popperOffsets, referenceOffsets, basePlacement) {
  var offsets = [0, 0];

  // Use height if placement is left or right and index is 0 otherwise use width
  // in this way the first offset will use an axis and the second one
  // will use the other one
  var useHeight = ['right', 'left'].indexOf(basePlacement) !== -1;

  // Split the offset string to obtain a list of values and operands
  // The regex addresses values with the plus or minus sign in front (+10, -20, etc)
  var fragments = offset.split(/(\+|\-)/).map(function (frag) {
    return frag.trim();
  });

  // Detect if the offset string contains a pair of values or a single one
  // they could be separated by comma or space
  var divider = fragments.indexOf(find(fragments, function (frag) {
    return frag.search(/,|\s/) !== -1;
  }));

  if (fragments[divider] && fragments[divider].indexOf(',') === -1) {
    console.warn('Offsets separated by white space(s) are deprecated, use a comma (,) instead.');
  }

  // If divider is found, we divide the list of values and operands to divide
  // them by ofset X and Y.
  var splitRegex = /\s*,\s*|\s+/;
  var ops = divider !== -1 ? [fragments.slice(0, divider).concat([fragments[divider].split(splitRegex)[0]]), [fragments[divider].split(splitRegex)[1]].concat(fragments.slice(divider + 1))] : [fragments];

  // Convert the values with units to absolute pixels to allow our computations
  ops = ops.map(function (op, index) {
    // Most of the units rely on the orientation of the popper
    var measurement = (index === 1 ? !useHeight : useHeight) ? 'height' : 'width';
    var mergeWithPrevious = false;
    return op
    // This aggregates any `+` or `-` sign that aren't considered operators
    // e.g.: 10 + +5 => [10, +, +5]
    .reduce(function (a, b) {
      if (a[a.length - 1] === '' && ['+', '-'].indexOf(b) !== -1) {
        a[a.length - 1] = b;
        mergeWithPrevious = true;
        return a;
      } else if (mergeWithPrevious) {
        a[a.length - 1] += b;
        mergeWithPrevious = false;
        return a;
      } else {
        return a.concat(b);
      }
    }, [])
    // Here we convert the string values into number values (in px)
    .map(function (str) {
      return toValue(str, measurement, popperOffsets, referenceOffsets);
    });
  });

  // Loop trough the offsets arrays and execute the operations
  ops.forEach(function (op, index) {
    op.forEach(function (frag, index2) {
      if (isNumeric(frag)) {
        offsets[index] += frag * (op[index2 - 1] === '-' ? -1 : 1);
      }
    });
  });
  return offsets;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by update method
 * @argument {Object} options - Modifiers configuration and options
 * @argument {Number|String} options.offset=0
 * The offset value as described in the modifier description
 * @returns {Object} The data object, properly modified
 */
function offset(data, _ref) {
  var offset = _ref.offset;
  var placement = data.placement,
      _data$offsets = data.offsets,
      popper = _data$offsets.popper,
      reference = _data$offsets.reference;

  var basePlacement = placement.split('-')[0];

  var offsets = void 0;
  if (isNumeric(+offset)) {
    offsets = [+offset, 0];
  } else {
    offsets = parseOffset(offset, popper, reference, basePlacement);
  }

  if (basePlacement === 'left') {
    popper.top += offsets[0];
    popper.left -= offsets[1];
  } else if (basePlacement === 'right') {
    popper.top += offsets[0];
    popper.left += offsets[1];
  } else if (basePlacement === 'top') {
    popper.left += offsets[0];
    popper.top -= offsets[1];
  } else if (basePlacement === 'bottom') {
    popper.left += offsets[0];
    popper.top += offsets[1];
  }

  data.popper = popper;
  return data;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by `update` method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function preventOverflow(data, options) {
  var boundariesElement = options.boundariesElement || getOffsetParent(data.instance.popper);

  // If offsetParent is the reference element, we really want to
  // go one step up and use the next offsetParent as reference to
  // avoid to make this modifier completely useless and look like broken
  if (data.instance.reference === boundariesElement) {
    boundariesElement = getOffsetParent(boundariesElement);
  }

  // NOTE: DOM access here
  // resets the popper's position so that the document size can be calculated excluding
  // the size of the popper element itself
  var transformProp = getSupportedPropertyName('transform');
  var popperStyles = data.instance.popper.style; // assignment to help minification
  var top = popperStyles.top,
      left = popperStyles.left,
      transform = popperStyles[transformProp];

  popperStyles.top = '';
  popperStyles.left = '';
  popperStyles[transformProp] = '';

  var boundaries = getBoundaries(data.instance.popper, data.instance.reference, options.padding, boundariesElement, data.positionFixed);

  // NOTE: DOM access here
  // restores the original style properties after the offsets have been computed
  popperStyles.top = top;
  popperStyles.left = left;
  popperStyles[transformProp] = transform;

  options.boundaries = boundaries;

  var order = options.priority;
  var popper = data.offsets.popper;

  var check = {
    primary: function primary(placement) {
      var value = popper[placement];
      if (popper[placement] < boundaries[placement] && !options.escapeWithReference) {
        value = Math.max(popper[placement], boundaries[placement]);
      }
      return defineProperty$1({}, placement, value);
    },
    secondary: function secondary(placement) {
      var mainSide = placement === 'right' ? 'left' : 'top';
      var value = popper[mainSide];
      if (popper[placement] > boundaries[placement] && !options.escapeWithReference) {
        value = Math.min(popper[mainSide], boundaries[placement] - (placement === 'right' ? popper.width : popper.height));
      }
      return defineProperty$1({}, mainSide, value);
    }
  };

  order.forEach(function (placement) {
    var side = ['left', 'top'].indexOf(placement) !== -1 ? 'primary' : 'secondary';
    popper = _extends$1({}, popper, check[side](placement));
  });

  data.offsets.popper = popper;

  return data;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by `update` method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function shift(data) {
  var placement = data.placement;
  var basePlacement = placement.split('-')[0];
  var shiftvariation = placement.split('-')[1];

  // if shift shiftvariation is specified, run the modifier
  if (shiftvariation) {
    var _data$offsets = data.offsets,
        reference = _data$offsets.reference,
        popper = _data$offsets.popper;

    var isVertical = ['bottom', 'top'].indexOf(basePlacement) !== -1;
    var side = isVertical ? 'left' : 'top';
    var measurement = isVertical ? 'width' : 'height';

    var shiftOffsets = {
      start: defineProperty$1({}, side, reference[side]),
      end: defineProperty$1({}, side, reference[side] + reference[measurement] - popper[measurement])
    };

    data.offsets.popper = _extends$1({}, popper, shiftOffsets[shiftvariation]);
  }

  return data;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by update method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function hide(data) {
  if (!isModifierRequired(data.instance.modifiers, 'hide', 'preventOverflow')) {
    return data;
  }

  var refRect = data.offsets.reference;
  var bound = find(data.instance.modifiers, function (modifier) {
    return modifier.name === 'preventOverflow';
  }).boundaries;

  if (refRect.bottom < bound.top || refRect.left > bound.right || refRect.top > bound.bottom || refRect.right < bound.left) {
    // Avoid unnecessary DOM access if visibility hasn't changed
    if (data.hide === true) {
      return data;
    }

    data.hide = true;
    data.attributes['x-out-of-boundaries'] = '';
  } else {
    // Avoid unnecessary DOM access if visibility hasn't changed
    if (data.hide === false) {
      return data;
    }

    data.hide = false;
    data.attributes['x-out-of-boundaries'] = false;
  }

  return data;
}

/**
 * @function
 * @memberof Modifiers
 * @argument {Object} data - The data object generated by `update` method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {Object} The data object, properly modified
 */
function inner(data) {
  var placement = data.placement;
  var basePlacement = placement.split('-')[0];
  var _data$offsets = data.offsets,
      popper = _data$offsets.popper,
      reference = _data$offsets.reference;

  var isHoriz = ['left', 'right'].indexOf(basePlacement) !== -1;

  var subtractLength = ['top', 'left'].indexOf(basePlacement) === -1;

  popper[isHoriz ? 'left' : 'top'] = reference[basePlacement] - (subtractLength ? popper[isHoriz ? 'width' : 'height'] : 0);

  data.placement = getOppositePlacement(placement);
  data.offsets.popper = getClientRect(popper);

  return data;
}

/**
 * Modifier function, each modifier can have a function of this type assigned
 * to its `fn` property.<br />
 * These functions will be called on each update, this means that you must
 * make sure they are performant enough to avoid performance bottlenecks.
 *
 * @function ModifierFn
 * @argument {dataObject} data - The data object generated by `update` method
 * @argument {Object} options - Modifiers configuration and options
 * @returns {dataObject} The data object, properly modified
 */

/**
 * Modifiers are plugins used to alter the behavior of your poppers.<br />
 * Popper.js uses a set of 9 modifiers to provide all the basic functionalities
 * needed by the library.
 *
 * Usually you don't want to override the `order`, `fn` and `onLoad` props.
 * All the other properties are configurations that could be tweaked.
 * @namespace modifiers
 */
var modifiers = {
  /**
   * Modifier used to shift the popper on the start or end of its reference
   * element.<br />
   * It will read the variation of the `placement` property.<br />
   * It can be one either `-end` or `-start`.
   * @memberof modifiers
   * @inner
   */
  shift: {
    /** @prop {number} order=100 - Index used to define the order of execution */
    order: 100,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: shift
  },

  /**
   * The `offset` modifier can shift your popper on both its axis.
   *
   * It accepts the following units:
   * - `px` or unit-less, interpreted as pixels
   * - `%` or `%r`, percentage relative to the length of the reference element
   * - `%p`, percentage relative to the length of the popper element
   * - `vw`, CSS viewport width unit
   * - `vh`, CSS viewport height unit
   *
   * For length is intended the main axis relative to the placement of the popper.<br />
   * This means that if the placement is `top` or `bottom`, the length will be the
   * `width`. In case of `left` or `right`, it will be the `height`.
   *
   * You can provide a single value (as `Number` or `String`), or a pair of values
   * as `String` divided by a comma or one (or more) white spaces.<br />
   * The latter is a deprecated method because it leads to confusion and will be
   * removed in v2.<br />
   * Additionally, it accepts additions and subtractions between different units.
   * Note that multiplications and divisions aren't supported.
   *
   * Valid examples are:
   * ```
   * 10
   * '10%'
   * '10, 10'
   * '10%, 10'
   * '10 + 10%'
   * '10 - 5vh + 3%'
   * '-10px + 5vh, 5px - 6%'
   * ```
   * > **NB**: If you desire to apply offsets to your poppers in a way that may make them overlap
   * > with their reference element, unfortunately, you will have to disable the `flip` modifier.
   * > You can read more on this at this [issue](https://github.com/FezVrasta/popper.js/issues/373).
   *
   * @memberof modifiers
   * @inner
   */
  offset: {
    /** @prop {number} order=200 - Index used to define the order of execution */
    order: 200,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: offset,
    /** @prop {Number|String} offset=0
     * The offset value as described in the modifier description
     */
    offset: 0
  },

  /**
   * Modifier used to prevent the popper from being positioned outside the boundary.
   *
   * A scenario exists where the reference itself is not within the boundaries.<br />
   * We can say it has "escaped the boundaries" — or just "escaped".<br />
   * In this case we need to decide whether the popper should either:
   *
   * - detach from the reference and remain "trapped" in the boundaries, or
   * - if it should ignore the boundary and "escape with its reference"
   *
   * When `escapeWithReference` is set to`true` and reference is completely
   * outside its boundaries, the popper will overflow (or completely leave)
   * the boundaries in order to remain attached to the edge of the reference.
   *
   * @memberof modifiers
   * @inner
   */
  preventOverflow: {
    /** @prop {number} order=300 - Index used to define the order of execution */
    order: 300,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: preventOverflow,
    /**
     * @prop {Array} [priority=['left','right','top','bottom']]
     * Popper will try to prevent overflow following these priorities by default,
     * then, it could overflow on the left and on top of the `boundariesElement`
     */
    priority: ['left', 'right', 'top', 'bottom'],
    /**
     * @prop {number} padding=5
     * Amount of pixel used to define a minimum distance between the boundaries
     * and the popper. This makes sure the popper always has a little padding
     * between the edges of its container
     */
    padding: 5,
    /**
     * @prop {String|HTMLElement} boundariesElement='scrollParent'
     * Boundaries used by the modifier. Can be `scrollParent`, `window`,
     * `viewport` or any DOM element.
     */
    boundariesElement: 'scrollParent'
  },

  /**
   * Modifier used to make sure the reference and its popper stay near each other
   * without leaving any gap between the two. Especially useful when the arrow is
   * enabled and you want to ensure that it points to its reference element.
   * It cares only about the first axis. You can still have poppers with margin
   * between the popper and its reference element.
   * @memberof modifiers
   * @inner
   */
  keepTogether: {
    /** @prop {number} order=400 - Index used to define the order of execution */
    order: 400,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: keepTogether
  },

  /**
   * This modifier is used to move the `arrowElement` of the popper to make
   * sure it is positioned between the reference element and its popper element.
   * It will read the outer size of the `arrowElement` node to detect how many
   * pixels of conjunction are needed.
   *
   * It has no effect if no `arrowElement` is provided.
   * @memberof modifiers
   * @inner
   */
  arrow: {
    /** @prop {number} order=500 - Index used to define the order of execution */
    order: 500,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: arrow,
    /** @prop {String|HTMLElement} element='[x-arrow]' - Selector or node used as arrow */
    element: '[x-arrow]'
  },

  /**
   * Modifier used to flip the popper's placement when it starts to overlap its
   * reference element.
   *
   * Requires the `preventOverflow` modifier before it in order to work.
   *
   * **NOTE:** this modifier will interrupt the current update cycle and will
   * restart it if it detects the need to flip the placement.
   * @memberof modifiers
   * @inner
   */
  flip: {
    /** @prop {number} order=600 - Index used to define the order of execution */
    order: 600,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: flip,
    /**
     * @prop {String|Array} behavior='flip'
     * The behavior used to change the popper's placement. It can be one of
     * `flip`, `clockwise`, `counterclockwise` or an array with a list of valid
     * placements (with optional variations)
     */
    behavior: 'flip',
    /**
     * @prop {number} padding=5
     * The popper will flip if it hits the edges of the `boundariesElement`
     */
    padding: 5,
    /**
     * @prop {String|HTMLElement} boundariesElement='viewport'
     * The element which will define the boundaries of the popper position.
     * The popper will never be placed outside of the defined boundaries
     * (except if `keepTogether` is enabled)
     */
    boundariesElement: 'viewport'
  },

  /**
   * Modifier used to make the popper flow toward the inner of the reference element.
   * By default, when this modifier is disabled, the popper will be placed outside
   * the reference element.
   * @memberof modifiers
   * @inner
   */
  inner: {
    /** @prop {number} order=700 - Index used to define the order of execution */
    order: 700,
    /** @prop {Boolean} enabled=false - Whether the modifier is enabled or not */
    enabled: false,
    /** @prop {ModifierFn} */
    fn: inner
  },

  /**
   * Modifier used to hide the popper when its reference element is outside of the
   * popper boundaries. It will set a `x-out-of-boundaries` attribute which can
   * be used to hide with a CSS selector the popper when its reference is
   * out of boundaries.
   *
   * Requires the `preventOverflow` modifier before it in order to work.
   * @memberof modifiers
   * @inner
   */
  hide: {
    /** @prop {number} order=800 - Index used to define the order of execution */
    order: 800,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: hide
  },

  /**
   * Computes the style that will be applied to the popper element to gets
   * properly positioned.
   *
   * Note that this modifier will not touch the DOM, it just prepares the styles
   * so that `applyStyle` modifier can apply it. This separation is useful
   * in case you need to replace `applyStyle` with a custom implementation.
   *
   * This modifier has `850` as `order` value to maintain backward compatibility
   * with previous versions of Popper.js. Expect the modifiers ordering method
   * to change in future major versions of the library.
   *
   * @memberof modifiers
   * @inner
   */
  computeStyle: {
    /** @prop {number} order=850 - Index used to define the order of execution */
    order: 850,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: computeStyle,
    /**
     * @prop {Boolean} gpuAcceleration=true
     * If true, it uses the CSS 3D transformation to position the popper.
     * Otherwise, it will use the `top` and `left` properties
     */
    gpuAcceleration: true,
    /**
     * @prop {string} [x='bottom']
     * Where to anchor the X axis (`bottom` or `top`). AKA X offset origin.
     * Change this if your popper should grow in a direction different from `bottom`
     */
    x: 'bottom',
    /**
     * @prop {string} [x='left']
     * Where to anchor the Y axis (`left` or `right`). AKA Y offset origin.
     * Change this if your popper should grow in a direction different from `right`
     */
    y: 'right'
  },

  /**
   * Applies the computed styles to the popper element.
   *
   * All the DOM manipulations are limited to this modifier. This is useful in case
   * you want to integrate Popper.js inside a framework or view library and you
   * want to delegate all the DOM manipulations to it.
   *
   * Note that if you disable this modifier, you must make sure the popper element
   * has its position set to `absolute` before Popper.js can do its work!
   *
   * Just disable this modifier and define your own to achieve the desired effect.
   *
   * @memberof modifiers
   * @inner
   */
  applyStyle: {
    /** @prop {number} order=900 - Index used to define the order of execution */
    order: 900,
    /** @prop {Boolean} enabled=true - Whether the modifier is enabled or not */
    enabled: true,
    /** @prop {ModifierFn} */
    fn: applyStyle,
    /** @prop {Function} */
    onLoad: applyStyleOnLoad,
    /**
     * @deprecated since version 1.10.0, the property moved to `computeStyle` modifier
     * @prop {Boolean} gpuAcceleration=true
     * If true, it uses the CSS 3D transformation to position the popper.
     * Otherwise, it will use the `top` and `left` properties
     */
    gpuAcceleration: undefined
  }
};

/**
 * The `dataObject` is an object containing all the information used by Popper.js.
 * This object is passed to modifiers and to the `onCreate` and `onUpdate` callbacks.
 * @name dataObject
 * @property {Object} data.instance The Popper.js instance
 * @property {String} data.placement Placement applied to popper
 * @property {String} data.originalPlacement Placement originally defined on init
 * @property {Boolean} data.flipped True if popper has been flipped by flip modifier
 * @property {Boolean} data.hide True if the reference element is out of boundaries, useful to know when to hide the popper
 * @property {HTMLElement} data.arrowElement Node used as arrow by arrow modifier
 * @property {Object} data.styles Any CSS property defined here will be applied to the popper. It expects the JavaScript nomenclature (eg. `marginBottom`)
 * @property {Object} data.arrowStyles Any CSS property defined here will be applied to the popper arrow. It expects the JavaScript nomenclature (eg. `marginBottom`)
 * @property {Object} data.boundaries Offsets of the popper boundaries
 * @property {Object} data.offsets The measurements of popper, reference and arrow elements
 * @property {Object} data.offsets.popper `top`, `left`, `width`, `height` values
 * @property {Object} data.offsets.reference `top`, `left`, `width`, `height` values
 * @property {Object} data.offsets.arrow] `top` and `left` offsets, only one of them will be different from 0
 */

/**
 * Default options provided to Popper.js constructor.<br />
 * These can be overridden using the `options` argument of Popper.js.<br />
 * To override an option, simply pass an object with the same
 * structure of the `options` object, as the 3rd argument. For example:
 * ```
 * new Popper(ref, pop, {
 *   modifiers: {
 *     preventOverflow: { enabled: false }
 *   }
 * })
 * ```
 * @type {Object}
 * @static
 * @memberof Popper
 */
var Defaults = {
  /**
   * Popper's placement.
   * @prop {Popper.placements} placement='bottom'
   */
  placement: 'bottom',

  /**
   * Set this to true if you want popper to position it self in 'fixed' mode
   * @prop {Boolean} positionFixed=false
   */
  positionFixed: false,

  /**
   * Whether events (resize, scroll) are initially enabled.
   * @prop {Boolean} eventsEnabled=true
   */
  eventsEnabled: true,

  /**
   * Set to true if you want to automatically remove the popper when
   * you call the `destroy` method.
   * @prop {Boolean} removeOnDestroy=false
   */
  removeOnDestroy: false,

  /**
   * Callback called when the popper is created.<br />
   * By default, it is set to no-op.<br />
   * Access Popper.js instance with `data.instance`.
   * @prop {onCreate}
   */
  onCreate: function onCreate() {},

  /**
   * Callback called when the popper is updated. This callback is not called
   * on the initialization/creation of the popper, but only on subsequent
   * updates.<br />
   * By default, it is set to no-op.<br />
   * Access Popper.js instance with `data.instance`.
   * @prop {onUpdate}
   */
  onUpdate: function onUpdate() {},

  /**
   * List of modifiers used to modify the offsets before they are applied to the popper.
   * They provide most of the functionalities of Popper.js.
   * @prop {modifiers}
   */
  modifiers: modifiers
};

/**
 * @callback onCreate
 * @param {dataObject} data
 */

/**
 * @callback onUpdate
 * @param {dataObject} data
 */

// Utils
// Methods
var Popper = function () {
  /**
   * Creates a new Popper.js instance.
   * @class Popper
   * @param {HTMLElement|referenceObject} reference - The reference element used to position the popper
   * @param {HTMLElement} popper - The HTML element used as the popper
   * @param {Object} options - Your custom options to override the ones defined in [Defaults](#defaults)
   * @return {Object} instance - The generated Popper.js instance
   */
  function Popper(reference, popper) {
    var _this = this;

    var options = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : {};
    classCallCheck$1(this, Popper);

    this.scheduleUpdate = function () {
      return requestAnimationFrame(_this.update);
    };

    // make update() debounced, so that it only runs at most once-per-tick
    this.update = debounce$1(this.update.bind(this));

    // with {} we create a new object with the options inside it
    this.options = _extends$1({}, Popper.Defaults, options);

    // init state
    this.state = {
      isDestroyed: false,
      isCreated: false,
      scrollParents: []
    };

    // get reference and popper elements (allow jQuery wrappers)
    this.reference = reference && reference.jquery ? reference[0] : reference;
    this.popper = popper && popper.jquery ? popper[0] : popper;

    // Deep merge modifiers options
    this.options.modifiers = {};
    Object.keys(_extends$1({}, Popper.Defaults.modifiers, options.modifiers)).forEach(function (name) {
      _this.options.modifiers[name] = _extends$1({}, Popper.Defaults.modifiers[name] || {}, options.modifiers ? options.modifiers[name] : {});
    });

    // Refactoring modifiers' list (Object => Array)
    this.modifiers = Object.keys(this.options.modifiers).map(function (name) {
      return _extends$1({
        name: name
      }, _this.options.modifiers[name]);
    })
    // sort the modifiers by order
    .sort(function (a, b) {
      return a.order - b.order;
    });

    // modifiers have the ability to execute arbitrary code when Popper.js get inited
    // such code is executed in the same order of its modifier
    // they could add new properties to their options configuration
    // BE AWARE: don't add options to `options.modifiers.name` but to `modifierOptions`!
    this.modifiers.forEach(function (modifierOptions) {
      if (modifierOptions.enabled && isFunction(modifierOptions.onLoad)) {
        modifierOptions.onLoad(_this.reference, _this.popper, _this.options, modifierOptions, _this.state);
      }
    });

    // fire the first update to position the popper in the right place
    this.update();

    var eventsEnabled = this.options.eventsEnabled;
    if (eventsEnabled) {
      // setup event listeners, they will take care of update the position in specific situations
      this.enableEventListeners();
    }

    this.state.eventsEnabled = eventsEnabled;
  }

  // We can't use class properties because they don't get listed in the
  // class prototype and break stuff like Sinon stubs


  createClass$1(Popper, [{
    key: 'update',
    value: function update$$1() {
      return update.call(this);
    }
  }, {
    key: 'destroy',
    value: function destroy$$1() {
      return destroy.call(this);
    }
  }, {
    key: 'enableEventListeners',
    value: function enableEventListeners$$1() {
      return enableEventListeners.call(this);
    }
  }, {
    key: 'disableEventListeners',
    value: function disableEventListeners$$1() {
      return disableEventListeners.call(this);
    }

    /**
     * Schedules an update. It will run on the next UI update available.
     * @method scheduleUpdate
     * @memberof Popper
     */


    /**
     * Collection of utilities useful when writing custom modifiers.
     * Starting from version 1.7, this method is available only if you
     * include `popper-utils.js` before `popper.js`.
     *
     * **DEPRECATION**: This way to access PopperUtils is deprecated
     * and will be removed in v2! Use the PopperUtils module directly instead.
     * Due to the high instability of the methods contained in Utils, we can't
     * guarantee them to follow semver. Use them at your own risk!
     * @static
     * @private
     * @type {Object}
     * @deprecated since version 1.8
     * @member Utils
     * @memberof Popper
     */

  }]);
  return Popper;
}();

/**
 * The `referenceObject` is an object that provides an interface compatible with Popper.js
 * and lets you use it as replacement of a real DOM node.<br />
 * You can use this method to position a popper relatively to a set of coordinates
 * in case you don't have a DOM node to use as reference.
 *
 * ```
 * new Popper(referenceObject, popperNode);
 * ```
 *
 * NB: This feature isn't supported in Internet Explorer 10.
 * @name referenceObject
 * @property {Function} data.getBoundingClientRect
 * A function that returns a set of coordinates compatible with the native `getBoundingClientRect` method.
 * @property {number} data.clientWidth
 * An ES6 getter that will return the width of the virtual reference element.
 * @property {number} data.clientHeight
 * An ES6 getter that will return the height of the virtual reference element.
 */


Popper.Utils = (typeof window !== 'undefined' ? window : global).PopperUtils;
Popper.placements = placements;
Popper.Defaults = Defaults;

var Popper$2 = {
    props: {
        placement: {
            type: String,
            default: 'bottom'
        },
        boundariesPadding: {
            type: Number,
            default: 5
        },
        reference: Object,
        popper: Object,
        offset: {
            default: 0
        },
        value: {
            type: Boolean,
            default: false
        },
        transition: String,
        options: {
            type: Object,
            default: function _default() {
                return {
                    gpuAcceleration: false,
                    boundariesElement: 'body'
                };
            }
        }
    },
    data: function data() {
        return {
            visible: this.value
        };
    },

    watch: {
        value: {
            immediate: true,
            handler: function handler(val) {
                this.visible = val;
                this.$emit('input', val);
            }
        },
        visible: function visible(val) {
            if (val) {
                this.updatePopper();
                this.$emit('on-show', this);
            } else {
                this.destroyPopper();
                this.$emit('on-hide', this);
            }
            this.$emit('input', val);
        }
    },
    methods: {
        createPopper: function createPopper() {
            var _this = this;

            if (!/^(top|bottom|left|right)(-start|-end)?$/g.test(this.placement)) {
                return;
            }

            var options = this.options;
            var popper = this.popper || this.$refs.popper;
            var reference = this.reference || this.$refs.reference;

            if (!popper || !reference) {
                return;
            }

            if (this.popperJS && this.popperJS.hasOwnProperty('destroy')) {
                this.popperJS.destroy();
            }

            options.placement = this.placement;
            options.offset = this.offset;

            this.popperJS = new Popper(reference, popper, Object.assign({}, options, {
                onCreate: function onCreate(popper) {
                    _this.resetTransformOrigin(popper.instance.popper);
                    _this.$nextTick(_this.updatePopper);
                    _this.$emit('created', _this);
                }
            }));
        },
        updatePopper: function updatePopper() {
            this.popperJS ? this.popperJS.update() : this.createPopper();
        },
        doDestroy: function doDestroy() {
            if (this.visible) {
                return;
            }
            this.popperJS.destroy();
            this.popperJS = null;
        },
        destroyPopper: function destroyPopper() {
            if (this.popperJS) {
                this.resetTransformOrigin(this.popperJS.popper);
            }
        },
        resetTransformOrigin: function resetTransformOrigin(popperNode) {
            var placementMap = { top: 'bottom', bottom: 'top', left: 'right', right: 'left' };
            var placement = popperNode.getAttribute('x-placement').split('-')[0];
            var origin = placementMap[placement];
            popperNode.style.transformOrigin = ['top', 'bottom'].indexOf(placement) > -1 ? 'center ' + origin : origin + ' center';
        }
    },
    beforeDestroy: function beforeDestroy() {
        if (this.popperJS) {
            this.popperJS.destroy();
        }
    }
};

function getTarget(node) {
    if (node === void 0) {
        node = document.body;
    }
    if (node === true) {
        return document.body;
    }
    return node instanceof window.Node ? node : document.querySelector(node);
}

var directive$2 = {
    inserted: function inserted(el, _ref, vnode) {
        var value = _ref.value;

        if (el.dataset.transfer !== 'true') {
            return false;
        }
        el.className = el.className ? el.className + ' v-transfer-dom' : 'v-transfer-dom';
        var parentNode = el.parentNode;
        if (!parentNode) {
            return;
        }
        var home = document.createComment('');
        var hasMovedOut = false;

        if (value !== false) {
            parentNode.replaceChild(home, el);
            getTarget(value).appendChild(el);
            hasMovedOut = true;
        }
        if (!el.__transferDomData) {
            el.__transferDomData = {
                parentNode: parentNode,
                home: home,
                target: getTarget(value),
                hasMovedOut: hasMovedOut
            };
        }
    },
    componentUpdated: function componentUpdated(el, _ref2) {
        var value = _ref2.value;

        if (el.dataset.transfer !== 'true') {
            return false;
        }

        var ref$1 = el.__transferDomData;
        if (!ref$1) {
            return;
        }

        var parentNode = ref$1.parentNode;
        var home = ref$1.home;
        var hasMovedOut = ref$1.hasMovedOut;

        if (!hasMovedOut && value) {
            parentNode.replaceChild(home, el);

            getTarget(value).appendChild(el);
            el.__transferDomData = Object.assign({}, el.__transferDomData, { hasMovedOut: true, target: getTarget(value) });
        } else if (hasMovedOut && value === false) {
            parentNode.replaceChild(el, home);
            el.__transferDomData = Object.assign({}, el.__transferDomData, { hasMovedOut: false, target: getTarget(value) });
        } else if (value) {
            getTarget(value).appendChild(el);
        }
    },
    unbind: function unbind(el) {
        if (el.dataset.transfer !== 'true') {
            return false;
        }
        el.className = el.className.replace('v-transfer-dom', '');
        var ref$1 = el.__transferDomData;
        if (!ref$1) return;
        if (el.__transferDomData.hasMovedOut === true) {
            el.__transferDomData.parentNode && el.__transferDomData.parentNode.appendChild(el);
        }
        el.__transferDomData = null;
    }
};

var oneOf = function oneOf(value, validList) {
    for (var i = 0; i < validList.length; i++) {
        if (value === validList[i]) {
            return true;
        }
    }
    return false;
};

var bkTooltip = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-tooltip", on: { "mouseenter": _vm.handleShowPopper, "mouseleave": _vm.handleClosePopper } }, [_c('div', { ref: "reference", staticClass: "bk-tooltip-rel" }, [_vm._t("default")], 2), _vm._v(" "), _c('transition', { attrs: { "name": "fade" } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: !_vm.disabled && (_vm.visible || _vm.always), expression: "!disabled && (visible || always)" }, { name: "transfer-dom", rawName: "v-transfer-dom" }], ref: "popper", staticClass: "bk-tooltip-popper", attrs: { "data-transfer": _vm.transfer }, on: { "mouseenter": _vm.handleShowPopper, "mouseleave": _vm.handleClosePopper } }, [_c('div', { staticClass: "bk-tooltip-content" }, [_c('div', { staticClass: "bk-tooltip-arrow" }), _vm._v(" "), _c('div', { staticClass: "bk-tooltip-inner", style: { width: _vm.width + 'px' } }, [_vm._t("content", [_vm._v(_vm._s(_vm.content))])], 2)])])])], 1);
    }, staticRenderFns: [],
    name: 'bk-tooltip',
    directives: { TransferDom: directive$2 },
    mixins: [Popper$2],
    props: {
        placement: {
            validator: function validator(value) {
                return oneOf(value, ['top', 'top-start', 'top-end', 'bottom', 'bottom-start', 'bottom-end', 'left', 'left-start', 'left-end', 'right', 'right-start', 'right-end']);
            },

            default: 'bottom'
        },
        content: {
            type: [String, Number],
            default: ''
        },
        delay: {
            type: Number,
            default: 100
        },
        width: {
            type: [String, Number],
            default: 'auto'
        },
        disabled: {
            type: Boolean,
            default: false
        },
        controlled: {
            type: Boolean,
            default: false
        },
        always: {
            type: Boolean,
            default: false
        },
        transfer: {
            type: Boolean,
            default: false
        }
    },
    data: function data() {
        return {};
    },

    methods: {
        handleShowPopper: function handleShowPopper() {
            var _this = this;

            if (this.timeout) {
                clearTimeout(this.timeout);
            }
            this.timeout = setTimeout(function () {
                _this.visible = true;
            }, this.delay);
        },
        handleClosePopper: function handleClosePopper() {
            var _this2 = this;

            if (this.timeout) {
                clearTimeout(this.timeout);
                if (!this.controlled) {
                    this.timeout = setTimeout(function () {
                        _this2.visible = false;
                    }, 100);
                }
            }
        }
    },
    mounted: function mounted() {
        if (this.always) {
            this.updatePopper();
        }
    }
};

bkTooltip.install = function (Vue$$1) {
  Vue$$1.component(bkTooltip.name, bkTooltip);
};

/**
 * @category Common Helpers
 * @summary Is the given argument an instance of Date?
 *
 * @description
 * Is the given argument an instance of Date?
 *
 * @param {*} argument - the argument to check
 * @returns {Boolean} the given argument is an instance of Date
 *
 * @example
 * // Is 'mayonnaise' a Date?
 * var result = isDate('mayonnaise')
 * //=> false
 */
function isDate (argument) {
  return argument instanceof Date
}

var is_date = isDate;

var MILLISECONDS_IN_HOUR = 3600000;
var MILLISECONDS_IN_MINUTE = 60000;
var DEFAULT_ADDITIONAL_DIGITS = 2;

var parseTokenDateTimeDelimeter = /[T ]/;
var parseTokenPlainTime = /:/;

// year tokens
var parseTokenYY = /^(\d{2})$/;
var parseTokensYYY = [
  /^([+-]\d{2})$/, // 0 additional digits
  /^([+-]\d{3})$/, // 1 additional digit
  /^([+-]\d{4})$/ // 2 additional digits
];

var parseTokenYYYY = /^(\d{4})/;
var parseTokensYYYYY = [
  /^([+-]\d{4})/, // 0 additional digits
  /^([+-]\d{5})/, // 1 additional digit
  /^([+-]\d{6})/ // 2 additional digits
];

// date tokens
var parseTokenMM = /^-(\d{2})$/;
var parseTokenDDD = /^-?(\d{3})$/;
var parseTokenMMDD = /^-?(\d{2})-?(\d{2})$/;
var parseTokenWww = /^-?W(\d{2})$/;
var parseTokenWwwD = /^-?W(\d{2})-?(\d{1})$/;

// time tokens
var parseTokenHH = /^(\d{2}([.,]\d*)?)$/;
var parseTokenHHMM = /^(\d{2}):?(\d{2}([.,]\d*)?)$/;
var parseTokenHHMMSS = /^(\d{2}):?(\d{2}):?(\d{2}([.,]\d*)?)$/;

// timezone tokens
var parseTokenTimezone = /([Z+-].*)$/;
var parseTokenTimezoneZ = /^(Z)$/;
var parseTokenTimezoneHH = /^([+-])(\d{2})$/;
var parseTokenTimezoneHHMM = /^([+-])(\d{2}):?(\d{2})$/;

/**
 * @category Common Helpers
 * @summary Convert the given argument to an instance of Date.
 *
 * @description
 * Convert the given argument to an instance of Date.
 *
 * If the argument is an instance of Date, the function returns its clone.
 *
 * If the argument is a number, it is treated as a timestamp.
 *
 * If an argument is a string, the function tries to parse it.
 * Function accepts complete ISO 8601 formats as well as partial implementations.
 * ISO 8601: http://en.wikipedia.org/wiki/ISO_8601
 *
 * If all above fails, the function passes the given argument to Date constructor.
 *
 * @param {Date|String|Number} argument - the value to convert
 * @param {Object} [options] - the object with options
 * @param {0 | 1 | 2} [options.additionalDigits=2] - the additional number of digits in the extended year format
 * @returns {Date} the parsed date in the local time zone
 *
 * @example
 * // Convert string '2014-02-11T11:30:30' to date:
 * var result = parse('2014-02-11T11:30:30')
 * //=> Tue Feb 11 2014 11:30:30
 *
 * @example
 * // Parse string '+02014101',
 * // if the additional number of digits in the extended year format is 1:
 * var result = parse('+02014101', {additionalDigits: 1})
 * //=> Fri Apr 11 2014 00:00:00
 */
function parse (argument, dirtyOptions) {
  if (is_date(argument)) {
    // Prevent the date to lose the milliseconds when passed to new Date() in IE10
    return new Date(argument.getTime())
  } else if (typeof argument !== 'string') {
    return new Date(argument)
  }

  var options = dirtyOptions || {};
  var additionalDigits = options.additionalDigits;
  if (additionalDigits == null) {
    additionalDigits = DEFAULT_ADDITIONAL_DIGITS;
  } else {
    additionalDigits = Number(additionalDigits);
  }

  var dateStrings = splitDateString(argument);

  var parseYearResult = parseYear(dateStrings.date, additionalDigits);
  var year = parseYearResult.year;
  var restDateString = parseYearResult.restDateString;

  var date = parseDate(restDateString, year);

  if (date) {
    var timestamp = date.getTime();
    var time = 0;
    var offset;

    if (dateStrings.time) {
      time = parseTime(dateStrings.time);
    }

    if (dateStrings.timezone) {
      offset = parseTimezone(dateStrings.timezone);
    } else {
      // get offset accurate to hour in timezones that change offset
      offset = new Date(timestamp + time).getTimezoneOffset();
      offset = new Date(timestamp + time + offset * MILLISECONDS_IN_MINUTE).getTimezoneOffset();
    }

    return new Date(timestamp + time + offset * MILLISECONDS_IN_MINUTE)
  } else {
    return new Date(argument)
  }
}

function splitDateString (dateString) {
  var dateStrings = {};
  var array = dateString.split(parseTokenDateTimeDelimeter);
  var timeString;

  if (parseTokenPlainTime.test(array[0])) {
    dateStrings.date = null;
    timeString = array[0];
  } else {
    dateStrings.date = array[0];
    timeString = array[1];
  }

  if (timeString) {
    var token = parseTokenTimezone.exec(timeString);
    if (token) {
      dateStrings.time = timeString.replace(token[1], '');
      dateStrings.timezone = token[1];
    } else {
      dateStrings.time = timeString;
    }
  }

  return dateStrings
}

function parseYear (dateString, additionalDigits) {
  var parseTokenYYY = parseTokensYYY[additionalDigits];
  var parseTokenYYYYY = parseTokensYYYYY[additionalDigits];

  var token;

  // YYYY or ±YYYYY
  token = parseTokenYYYY.exec(dateString) || parseTokenYYYYY.exec(dateString);
  if (token) {
    var yearString = token[1];
    return {
      year: parseInt(yearString, 10),
      restDateString: dateString.slice(yearString.length)
    }
  }

  // YY or ±YYY
  token = parseTokenYY.exec(dateString) || parseTokenYYY.exec(dateString);
  if (token) {
    var centuryString = token[1];
    return {
      year: parseInt(centuryString, 10) * 100,
      restDateString: dateString.slice(centuryString.length)
    }
  }

  // Invalid ISO-formatted year
  return {
    year: null
  }
}

function parseDate (dateString, year) {
  // Invalid ISO-formatted year
  if (year === null) {
    return null
  }

  var token;
  var date;
  var month;
  var week;

  // YYYY
  if (dateString.length === 0) {
    date = new Date(0);
    date.setUTCFullYear(year);
    return date
  }

  // YYYY-MM
  token = parseTokenMM.exec(dateString);
  if (token) {
    date = new Date(0);
    month = parseInt(token[1], 10) - 1;
    date.setUTCFullYear(year, month);
    return date
  }

  // YYYY-DDD or YYYYDDD
  token = parseTokenDDD.exec(dateString);
  if (token) {
    date = new Date(0);
    var dayOfYear = parseInt(token[1], 10);
    date.setUTCFullYear(year, 0, dayOfYear);
    return date
  }

  // YYYY-MM-DD or YYYYMMDD
  token = parseTokenMMDD.exec(dateString);
  if (token) {
    date = new Date(0);
    month = parseInt(token[1], 10) - 1;
    var day = parseInt(token[2], 10);
    date.setUTCFullYear(year, month, day);
    return date
  }

  // YYYY-Www or YYYYWww
  token = parseTokenWww.exec(dateString);
  if (token) {
    week = parseInt(token[1], 10) - 1;
    return dayOfISOYear(year, week)
  }

  // YYYY-Www-D or YYYYWwwD
  token = parseTokenWwwD.exec(dateString);
  if (token) {
    week = parseInt(token[1], 10) - 1;
    var dayOfWeek = parseInt(token[2], 10) - 1;
    return dayOfISOYear(year, week, dayOfWeek)
  }

  // Invalid ISO-formatted date
  return null
}

function parseTime (timeString) {
  var token;
  var hours;
  var minutes;

  // hh
  token = parseTokenHH.exec(timeString);
  if (token) {
    hours = parseFloat(token[1].replace(',', '.'));
    return (hours % 24) * MILLISECONDS_IN_HOUR
  }

  // hh:mm or hhmm
  token = parseTokenHHMM.exec(timeString);
  if (token) {
    hours = parseInt(token[1], 10);
    minutes = parseFloat(token[2].replace(',', '.'));
    return (hours % 24) * MILLISECONDS_IN_HOUR +
      minutes * MILLISECONDS_IN_MINUTE
  }

  // hh:mm:ss or hhmmss
  token = parseTokenHHMMSS.exec(timeString);
  if (token) {
    hours = parseInt(token[1], 10);
    minutes = parseInt(token[2], 10);
    var seconds = parseFloat(token[3].replace(',', '.'));
    return (hours % 24) * MILLISECONDS_IN_HOUR +
      minutes * MILLISECONDS_IN_MINUTE +
      seconds * 1000
  }

  // Invalid ISO-formatted time
  return null
}

function parseTimezone (timezoneString) {
  var token;
  var absoluteOffset;

  // Z
  token = parseTokenTimezoneZ.exec(timezoneString);
  if (token) {
    return 0
  }

  // ±hh
  token = parseTokenTimezoneHH.exec(timezoneString);
  if (token) {
    absoluteOffset = parseInt(token[2], 10) * 60;
    return (token[1] === '+') ? -absoluteOffset : absoluteOffset
  }

  // ±hh:mm or ±hhmm
  token = parseTokenTimezoneHHMM.exec(timezoneString);
  if (token) {
    absoluteOffset = parseInt(token[2], 10) * 60 + parseInt(token[3], 10);
    return (token[1] === '+') ? -absoluteOffset : absoluteOffset
  }

  return 0
}

function dayOfISOYear (isoYear, week, day) {
  week = week || 0;
  day = day || 0;
  var date = new Date(0);
  date.setUTCFullYear(isoYear, 0, 4);
  var fourthOfJanuaryDay = date.getUTCDay() || 7;
  var diff = week * 7 + day + 1 - fourthOfJanuaryDay;
  date.setUTCDate(date.getUTCDate() + diff);
  return date
}

var parse_1 = parse;

/**
 * @category Year Helpers
 * @summary Return the start of a year for the given date.
 *
 * @description
 * Return the start of a year for the given date.
 * The result will be in the local timezone.
 *
 * @param {Date|String|Number} date - the original date
 * @returns {Date} the start of a year
 *
 * @example
 * // The start of a year for 2 September 2014 11:55:00:
 * var result = startOfYear(new Date(2014, 8, 2, 11, 55, 00))
 * //=> Wed Jan 01 2014 00:00:00
 */
function startOfYear (dirtyDate) {
  var cleanDate = parse_1(dirtyDate);
  var date = new Date(0);
  date.setFullYear(cleanDate.getFullYear(), 0, 1);
  date.setHours(0, 0, 0, 0);
  return date
}

var start_of_year = startOfYear;

/**
 * @category Day Helpers
 * @summary Return the start of a day for the given date.
 *
 * @description
 * Return the start of a day for the given date.
 * The result will be in the local timezone.
 *
 * @param {Date|String|Number} date - the original date
 * @returns {Date} the start of a day
 *
 * @example
 * // The start of a day for 2 September 2014 11:55:00:
 * var result = startOfDay(new Date(2014, 8, 2, 11, 55, 0))
 * //=> Tue Sep 02 2014 00:00:00
 */
function startOfDay (dirtyDate) {
  var date = parse_1(dirtyDate);
  date.setHours(0, 0, 0, 0);
  return date
}

var start_of_day = startOfDay;

var MILLISECONDS_IN_MINUTE$1 = 60000;
var MILLISECONDS_IN_DAY = 86400000;

/**
 * @category Day Helpers
 * @summary Get the number of calendar days between the given dates.
 *
 * @description
 * Get the number of calendar days between the given dates.
 *
 * @param {Date|String|Number} dateLeft - the later date
 * @param {Date|String|Number} dateRight - the earlier date
 * @returns {Number} the number of calendar days
 *
 * @example
 * // How many calendar days are between
 * // 2 July 2011 23:00:00 and 2 July 2012 00:00:00?
 * var result = differenceInCalendarDays(
 *   new Date(2012, 6, 2, 0, 0),
 *   new Date(2011, 6, 2, 23, 0)
 * )
 * //=> 366
 */
function differenceInCalendarDays (dirtyDateLeft, dirtyDateRight) {
  var startOfDayLeft = start_of_day(dirtyDateLeft);
  var startOfDayRight = start_of_day(dirtyDateRight);

  var timestampLeft = startOfDayLeft.getTime() -
    startOfDayLeft.getTimezoneOffset() * MILLISECONDS_IN_MINUTE$1;
  var timestampRight = startOfDayRight.getTime() -
    startOfDayRight.getTimezoneOffset() * MILLISECONDS_IN_MINUTE$1;

  // Round the number of days to the nearest integer
  // because the number of milliseconds in a day is not constant
  // (e.g. it's different in the day of the daylight saving time clock shift)
  return Math.round((timestampLeft - timestampRight) / MILLISECONDS_IN_DAY)
}

var difference_in_calendar_days = differenceInCalendarDays;

/**
 * @category Day Helpers
 * @summary Get the day of the year of the given date.
 *
 * @description
 * Get the day of the year of the given date.
 *
 * @param {Date|String|Number} date - the given date
 * @returns {Number} the day of year
 *
 * @example
 * // Which day of the year is 2 July 2014?
 * var result = getDayOfYear(new Date(2014, 6, 2))
 * //=> 183
 */
function getDayOfYear (dirtyDate) {
  var date = parse_1(dirtyDate);
  var diff = difference_in_calendar_days(date, start_of_year(date));
  var dayOfYear = diff + 1;
  return dayOfYear
}

var get_day_of_year = getDayOfYear;

/**
 * @category Week Helpers
 * @summary Return the start of a week for the given date.
 *
 * @description
 * Return the start of a week for the given date.
 * The result will be in the local timezone.
 *
 * @param {Date|String|Number} date - the original date
 * @param {Object} [options] - the object with options
 * @param {Number} [options.weekStartsOn=0] - the index of the first day of the week (0 - Sunday)
 * @returns {Date} the start of a week
 *
 * @example
 * // The start of a week for 2 September 2014 11:55:00:
 * var result = startOfWeek(new Date(2014, 8, 2, 11, 55, 0))
 * //=> Sun Aug 31 2014 00:00:00
 *
 * @example
 * // If the week starts on Monday, the start of the week for 2 September 2014 11:55:00:
 * var result = startOfWeek(new Date(2014, 8, 2, 11, 55, 0), {weekStartsOn: 1})
 * //=> Mon Sep 01 2014 00:00:00
 */
function startOfWeek (dirtyDate, dirtyOptions) {
  var weekStartsOn = dirtyOptions ? (Number(dirtyOptions.weekStartsOn) || 0) : 0;

  var date = parse_1(dirtyDate);
  var day = date.getDay();
  var diff = (day < weekStartsOn ? 7 : 0) + day - weekStartsOn;

  date.setDate(date.getDate() - diff);
  date.setHours(0, 0, 0, 0);
  return date
}

var start_of_week = startOfWeek;

/**
 * @category ISO Week Helpers
 * @summary Return the start of an ISO week for the given date.
 *
 * @description
 * Return the start of an ISO week for the given date.
 * The result will be in the local timezone.
 *
 * ISO week-numbering year: http://en.wikipedia.org/wiki/ISO_week_date
 *
 * @param {Date|String|Number} date - the original date
 * @returns {Date} the start of an ISO week
 *
 * @example
 * // The start of an ISO week for 2 September 2014 11:55:00:
 * var result = startOfISOWeek(new Date(2014, 8, 2, 11, 55, 0))
 * //=> Mon Sep 01 2014 00:00:00
 */
function startOfISOWeek (dirtyDate) {
  return start_of_week(dirtyDate, {weekStartsOn: 1})
}

var start_of_iso_week = startOfISOWeek;

/**
 * @category ISO Week-Numbering Year Helpers
 * @summary Get the ISO week-numbering year of the given date.
 *
 * @description
 * Get the ISO week-numbering year of the given date,
 * which always starts 3 days before the year's first Thursday.
 *
 * ISO week-numbering year: http://en.wikipedia.org/wiki/ISO_week_date
 *
 * @param {Date|String|Number} date - the given date
 * @returns {Number} the ISO week-numbering year
 *
 * @example
 * // Which ISO-week numbering year is 2 January 2005?
 * var result = getISOYear(new Date(2005, 0, 2))
 * //=> 2004
 */
function getISOYear (dirtyDate) {
  var date = parse_1(dirtyDate);
  var year = date.getFullYear();

  var fourthOfJanuaryOfNextYear = new Date(0);
  fourthOfJanuaryOfNextYear.setFullYear(year + 1, 0, 4);
  fourthOfJanuaryOfNextYear.setHours(0, 0, 0, 0);
  var startOfNextYear = start_of_iso_week(fourthOfJanuaryOfNextYear);

  var fourthOfJanuaryOfThisYear = new Date(0);
  fourthOfJanuaryOfThisYear.setFullYear(year, 0, 4);
  fourthOfJanuaryOfThisYear.setHours(0, 0, 0, 0);
  var startOfThisYear = start_of_iso_week(fourthOfJanuaryOfThisYear);

  if (date.getTime() >= startOfNextYear.getTime()) {
    return year + 1
  } else if (date.getTime() >= startOfThisYear.getTime()) {
    return year
  } else {
    return year - 1
  }
}

var get_iso_year = getISOYear;

/**
 * @category ISO Week-Numbering Year Helpers
 * @summary Return the start of an ISO week-numbering year for the given date.
 *
 * @description
 * Return the start of an ISO week-numbering year,
 * which always starts 3 days before the year's first Thursday.
 * The result will be in the local timezone.
 *
 * ISO week-numbering year: http://en.wikipedia.org/wiki/ISO_week_date
 *
 * @param {Date|String|Number} date - the original date
 * @returns {Date} the start of an ISO year
 *
 * @example
 * // The start of an ISO week-numbering year for 2 July 2005:
 * var result = startOfISOYear(new Date(2005, 6, 2))
 * //=> Mon Jan 03 2005 00:00:00
 */
function startOfISOYear (dirtyDate) {
  var year = get_iso_year(dirtyDate);
  var fourthOfJanuary = new Date(0);
  fourthOfJanuary.setFullYear(year, 0, 4);
  fourthOfJanuary.setHours(0, 0, 0, 0);
  var date = start_of_iso_week(fourthOfJanuary);
  return date
}

var start_of_iso_year = startOfISOYear;

var MILLISECONDS_IN_WEEK = 604800000;

/**
 * @category ISO Week Helpers
 * @summary Get the ISO week of the given date.
 *
 * @description
 * Get the ISO week of the given date.
 *
 * ISO week-numbering year: http://en.wikipedia.org/wiki/ISO_week_date
 *
 * @param {Date|String|Number} date - the given date
 * @returns {Number} the ISO week
 *
 * @example
 * // Which week of the ISO-week numbering year is 2 January 2005?
 * var result = getISOWeek(new Date(2005, 0, 2))
 * //=> 53
 */
function getISOWeek (dirtyDate) {
  var date = parse_1(dirtyDate);
  var diff = start_of_iso_week(date).getTime() - start_of_iso_year(date).getTime();

  // Round the number of days to the nearest integer
  // because the number of milliseconds in a week is not constant
  // (e.g. it's different in the week of the daylight saving time clock shift)
  return Math.round(diff / MILLISECONDS_IN_WEEK) + 1
}

var get_iso_week = getISOWeek;

/**
 * @category Common Helpers
 * @summary Is the given date valid?
 *
 * @description
 * Returns false if argument is Invalid Date and true otherwise.
 * Invalid Date is a Date, whose time value is NaN.
 *
 * Time value of Date: http://es5.github.io
 *
 * @param {Date} date - the date to check
 * @returns {Boolean} the date is valid
 * @throws {TypeError} argument must be an instance of Date
 *
 * @example
 * // For the valid date:
 * var result = isValid(new Date(2014, 1, 31))
 * //=> true
 *
 * @example
 * // For the invalid date:
 * var result = isValid(new Date(''))
 * //=> false
 */
function isValid (dirtyDate) {
  if (is_date(dirtyDate)) {
    return !isNaN(dirtyDate)
  } else {
    throw new TypeError(toString.call(dirtyDate) + ' is not an instance of Date')
  }
}

var is_valid = isValid;

function buildDistanceInWordsLocale () {
  var distanceInWordsLocale = {
    lessThanXSeconds: {
      one: 'less than a second',
      other: 'less than {{count}} seconds'
    },

    xSeconds: {
      one: '1 second',
      other: '{{count}} seconds'
    },

    halfAMinute: 'half a minute',

    lessThanXMinutes: {
      one: 'less than a minute',
      other: 'less than {{count}} minutes'
    },

    xMinutes: {
      one: '1 minute',
      other: '{{count}} minutes'
    },

    aboutXHours: {
      one: 'about 1 hour',
      other: 'about {{count}} hours'
    },

    xHours: {
      one: '1 hour',
      other: '{{count}} hours'
    },

    xDays: {
      one: '1 day',
      other: '{{count}} days'
    },

    aboutXMonths: {
      one: 'about 1 month',
      other: 'about {{count}} months'
    },

    xMonths: {
      one: '1 month',
      other: '{{count}} months'
    },

    aboutXYears: {
      one: 'about 1 year',
      other: 'about {{count}} years'
    },

    xYears: {
      one: '1 year',
      other: '{{count}} years'
    },

    overXYears: {
      one: 'over 1 year',
      other: 'over {{count}} years'
    },

    almostXYears: {
      one: 'almost 1 year',
      other: 'almost {{count}} years'
    }
  };

  function localize (token, count, options) {
    options = options || {};

    var result;
    if (typeof distanceInWordsLocale[token] === 'string') {
      result = distanceInWordsLocale[token];
    } else if (count === 1) {
      result = distanceInWordsLocale[token].one;
    } else {
      result = distanceInWordsLocale[token].other.replace('{{count}}', count);
    }

    if (options.addSuffix) {
      if (options.comparison > 0) {
        return 'in ' + result
      } else {
        return result + ' ago'
      }
    }

    return result
  }

  return {
    localize: localize
  }
}

var build_distance_in_words_locale = buildDistanceInWordsLocale;

var commonFormatterKeys = [
  'M', 'MM', 'Q', 'D', 'DD', 'DDD', 'DDDD', 'd',
  'E', 'W', 'WW', 'YY', 'YYYY', 'GG', 'GGGG',
  'H', 'HH', 'h', 'hh', 'm', 'mm',
  's', 'ss', 'S', 'SS', 'SSS',
  'Z', 'ZZ', 'X', 'x'
];

function buildFormattingTokensRegExp (formatters) {
  var formatterKeys = [];
  for (var key in formatters) {
    if (formatters.hasOwnProperty(key)) {
      formatterKeys.push(key);
    }
  }

  var formattingTokens = commonFormatterKeys
    .concat(formatterKeys)
    .sort()
    .reverse();
  var formattingTokensRegExp = new RegExp(
    '(\\[[^\\[]*\\])|(\\\\)?' + '(' + formattingTokens.join('|') + '|.)', 'g'
  );

  return formattingTokensRegExp
}

var build_formatting_tokens_reg_exp = buildFormattingTokensRegExp;

function buildFormatLocale () {
  // Note: in English, the names of days of the week and months are capitalized.
  // If you are making a new locale based on this one, check if the same is true for the language you're working on.
  // Generally, formatted dates should look like they are in the middle of a sentence,
  // e.g. in Spanish language the weekdays and months should be in the lowercase.
  var months3char = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  var monthsFull = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
  var weekdays2char = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];
  var weekdays3char = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  var weekdaysFull = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  var meridiemUppercase = ['AM', 'PM'];
  var meridiemLowercase = ['am', 'pm'];
  var meridiemFull = ['a.m.', 'p.m.'];

  var formatters = {
    // Month: Jan, Feb, ..., Dec
    'MMM': function (date) {
      return months3char[date.getMonth()]
    },

    // Month: January, February, ..., December
    'MMMM': function (date) {
      return monthsFull[date.getMonth()]
    },

    // Day of week: Su, Mo, ..., Sa
    'dd': function (date) {
      return weekdays2char[date.getDay()]
    },

    // Day of week: Sun, Mon, ..., Sat
    'ddd': function (date) {
      return weekdays3char[date.getDay()]
    },

    // Day of week: Sunday, Monday, ..., Saturday
    'dddd': function (date) {
      return weekdaysFull[date.getDay()]
    },

    // AM, PM
    'A': function (date) {
      return (date.getHours() / 12) >= 1 ? meridiemUppercase[1] : meridiemUppercase[0]
    },

    // am, pm
    'a': function (date) {
      return (date.getHours() / 12) >= 1 ? meridiemLowercase[1] : meridiemLowercase[0]
    },

    // a.m., p.m.
    'aa': function (date) {
      return (date.getHours() / 12) >= 1 ? meridiemFull[1] : meridiemFull[0]
    }
  };

  // Generate ordinal version of formatters: M -> Mo, D -> Do, etc.
  var ordinalFormatters = ['M', 'D', 'DDD', 'd', 'Q', 'W'];
  ordinalFormatters.forEach(function (formatterToken) {
    formatters[formatterToken + 'o'] = function (date, formatters) {
      return ordinal(formatters[formatterToken](date))
    };
  });

  return {
    formatters: formatters,
    formattingTokensRegExp: build_formatting_tokens_reg_exp(formatters)
  }
}

function ordinal (number) {
  var rem100 = number % 100;
  if (rem100 > 20 || rem100 < 10) {
    switch (rem100 % 10) {
      case 1:
        return number + 'st'
      case 2:
        return number + 'nd'
      case 3:
        return number + 'rd'
    }
  }
  return number + 'th'
}

var build_format_locale = buildFormatLocale;

/**
 * @category Locales
 * @summary English locale.
 */
var en = {
  distanceInWords: build_distance_in_words_locale(),
  format: build_format_locale()
};

/**
 * @category Common Helpers
 * @summary Format the date.
 *
 * @description
 * Return the formatted date string in the given format.
 *
 * Accepted tokens:
 * | Unit                    | Token | Result examples                  |
 * |-------------------------|-------|----------------------------------|
 * | Month                   | M     | 1, 2, ..., 12                    |
 * |                         | Mo    | 1st, 2nd, ..., 12th              |
 * |                         | MM    | 01, 02, ..., 12                  |
 * |                         | MMM   | Jan, Feb, ..., Dec               |
 * |                         | MMMM  | January, February, ..., December |
 * | Quarter                 | Q     | 1, 2, 3, 4                       |
 * |                         | Qo    | 1st, 2nd, 3rd, 4th               |
 * | Day of month            | D     | 1, 2, ..., 31                    |
 * |                         | Do    | 1st, 2nd, ..., 31st              |
 * |                         | DD    | 01, 02, ..., 31                  |
 * | Day of year             | DDD   | 1, 2, ..., 366                   |
 * |                         | DDDo  | 1st, 2nd, ..., 366th             |
 * |                         | DDDD  | 001, 002, ..., 366               |
 * | Day of week             | d     | 0, 1, ..., 6                     |
 * |                         | do    | 0th, 1st, ..., 6th               |
 * |                         | dd    | Su, Mo, ..., Sa                  |
 * |                         | ddd   | Sun, Mon, ..., Sat               |
 * |                         | dddd  | Sunday, Monday, ..., Saturday    |
 * | Day of ISO week         | E     | 1, 2, ..., 7                     |
 * | ISO week                | W     | 1, 2, ..., 53                    |
 * |                         | Wo    | 1st, 2nd, ..., 53rd              |
 * |                         | WW    | 01, 02, ..., 53                  |
 * | Year                    | YY    | 00, 01, ..., 99                  |
 * |                         | YYYY  | 1900, 1901, ..., 2099            |
 * | ISO week-numbering year | GG    | 00, 01, ..., 99                  |
 * |                         | GGGG  | 1900, 1901, ..., 2099            |
 * | AM/PM                   | A     | AM, PM                           |
 * |                         | a     | am, pm                           |
 * |                         | aa    | a.m., p.m.                       |
 * | Hour                    | H     | 0, 1, ... 23                     |
 * |                         | HH    | 00, 01, ... 23                   |
 * |                         | h     | 1, 2, ..., 12                    |
 * |                         | hh    | 01, 02, ..., 12                  |
 * | Minute                  | m     | 0, 1, ..., 59                    |
 * |                         | mm    | 00, 01, ..., 59                  |
 * | Second                  | s     | 0, 1, ..., 59                    |
 * |                         | ss    | 00, 01, ..., 59                  |
 * | 1/10 of second          | S     | 0, 1, ..., 9                     |
 * | 1/100 of second         | SS    | 00, 01, ..., 99                  |
 * | Millisecond             | SSS   | 000, 001, ..., 999               |
 * | Timezone                | Z     | -01:00, +00:00, ... +12:00       |
 * |                         | ZZ    | -0100, +0000, ..., +1200         |
 * | Seconds timestamp       | X     | 512969520                        |
 * | Milliseconds timestamp  | x     | 512969520900                     |
 *
 * The characters wrapped in square brackets are escaped.
 *
 * The result may vary by locale.
 *
 * @param {Date|String|Number} date - the original date
 * @param {String} [format='YYYY-MM-DDTHH:mm:ss.SSSZ'] - the string of tokens
 * @param {Object} [options] - the object with options
 * @param {Object} [options.locale=enLocale] - the locale object
 * @returns {String} the formatted date string
 *
 * @example
 * // Represent 11 February 2014 in middle-endian format:
 * var result = format(
 *   new Date(2014, 1, 11),
 *   'MM/DD/YYYY'
 * )
 * //=> '02/11/2014'
 *
 * @example
 * // Represent 2 July 2014 in Esperanto:
 * var eoLocale = require('date-fns/locale/eo')
 * var result = format(
 *   new Date(2014, 6, 2),
 *   'Do [de] MMMM YYYY',
 *   {locale: eoLocale}
 * )
 * //=> '2-a de julio 2014'
 */
function format (dirtyDate, dirtyFormatStr, dirtyOptions) {
  var formatStr = dirtyFormatStr ? String(dirtyFormatStr) : 'YYYY-MM-DDTHH:mm:ss.SSSZ';
  var options = dirtyOptions || {};

  var locale = options.locale;
  var localeFormatters = en.format.formatters;
  var formattingTokensRegExp = en.format.formattingTokensRegExp;
  if (locale && locale.format && locale.format.formatters) {
    localeFormatters = locale.format.formatters;

    if (locale.format.formattingTokensRegExp) {
      formattingTokensRegExp = locale.format.formattingTokensRegExp;
    }
  }

  var date = parse_1(dirtyDate);

  if (!is_valid(date)) {
    return 'Invalid Date'
  }

  var formatFn = buildFormatFn(formatStr, localeFormatters, formattingTokensRegExp);

  return formatFn(date)
}

var formatters = {
  // Month: 1, 2, ..., 12
  'M': function (date) {
    return date.getMonth() + 1
  },

  // Month: 01, 02, ..., 12
  'MM': function (date) {
    return addLeadingZeros(date.getMonth() + 1, 2)
  },

  // Quarter: 1, 2, 3, 4
  'Q': function (date) {
    return Math.ceil((date.getMonth() + 1) / 3)
  },

  // Day of month: 1, 2, ..., 31
  'D': function (date) {
    return date.getDate()
  },

  // Day of month: 01, 02, ..., 31
  'DD': function (date) {
    return addLeadingZeros(date.getDate(), 2)
  },

  // Day of year: 1, 2, ..., 366
  'DDD': function (date) {
    return get_day_of_year(date)
  },

  // Day of year: 001, 002, ..., 366
  'DDDD': function (date) {
    return addLeadingZeros(get_day_of_year(date), 3)
  },

  // Day of week: 0, 1, ..., 6
  'd': function (date) {
    return date.getDay()
  },

  // Day of ISO week: 1, 2, ..., 7
  'E': function (date) {
    return date.getDay() || 7
  },

  // ISO week: 1, 2, ..., 53
  'W': function (date) {
    return get_iso_week(date)
  },

  // ISO week: 01, 02, ..., 53
  'WW': function (date) {
    return addLeadingZeros(get_iso_week(date), 2)
  },

  // Year: 00, 01, ..., 99
  'YY': function (date) {
    return addLeadingZeros(date.getFullYear(), 4).substr(2)
  },

  // Year: 1900, 1901, ..., 2099
  'YYYY': function (date) {
    return addLeadingZeros(date.getFullYear(), 4)
  },

  // ISO week-numbering year: 00, 01, ..., 99
  'GG': function (date) {
    return String(get_iso_year(date)).substr(2)
  },

  // ISO week-numbering year: 1900, 1901, ..., 2099
  'GGGG': function (date) {
    return get_iso_year(date)
  },

  // Hour: 0, 1, ... 23
  'H': function (date) {
    return date.getHours()
  },

  // Hour: 00, 01, ..., 23
  'HH': function (date) {
    return addLeadingZeros(date.getHours(), 2)
  },

  // Hour: 1, 2, ..., 12
  'h': function (date) {
    var hours = date.getHours();
    if (hours === 0) {
      return 12
    } else if (hours > 12) {
      return hours % 12
    } else {
      return hours
    }
  },

  // Hour: 01, 02, ..., 12
  'hh': function (date) {
    return addLeadingZeros(formatters['h'](date), 2)
  },

  // Minute: 0, 1, ..., 59
  'm': function (date) {
    return date.getMinutes()
  },

  // Minute: 00, 01, ..., 59
  'mm': function (date) {
    return addLeadingZeros(date.getMinutes(), 2)
  },

  // Second: 0, 1, ..., 59
  's': function (date) {
    return date.getSeconds()
  },

  // Second: 00, 01, ..., 59
  'ss': function (date) {
    return addLeadingZeros(date.getSeconds(), 2)
  },

  // 1/10 of second: 0, 1, ..., 9
  'S': function (date) {
    return Math.floor(date.getMilliseconds() / 100)
  },

  // 1/100 of second: 00, 01, ..., 99
  'SS': function (date) {
    return addLeadingZeros(Math.floor(date.getMilliseconds() / 10), 2)
  },

  // Millisecond: 000, 001, ..., 999
  'SSS': function (date) {
    return addLeadingZeros(date.getMilliseconds(), 3)
  },

  // Timezone: -01:00, +00:00, ... +12:00
  'Z': function (date) {
    return formatTimezone(date.getTimezoneOffset(), ':')
  },

  // Timezone: -0100, +0000, ... +1200
  'ZZ': function (date) {
    return formatTimezone(date.getTimezoneOffset())
  },

  // Seconds timestamp: 512969520
  'X': function (date) {
    return Math.floor(date.getTime() / 1000)
  },

  // Milliseconds timestamp: 512969520900
  'x': function (date) {
    return date.getTime()
  }
};

function buildFormatFn (formatStr, localeFormatters, formattingTokensRegExp) {
  var array = formatStr.match(formattingTokensRegExp);
  var length = array.length;

  var i;
  var formatter;
  for (i = 0; i < length; i++) {
    formatter = localeFormatters[array[i]] || formatters[array[i]];
    if (formatter) {
      array[i] = formatter;
    } else {
      array[i] = removeFormattingTokens(array[i]);
    }
  }

  return function (date) {
    var output = '';
    for (var i = 0; i < length; i++) {
      if (array[i] instanceof Function) {
        output += array[i](date, formatters);
      } else {
        output += array[i];
      }
    }
    return output
  }
}

function removeFormattingTokens (input) {
  if (input.match(/\[[\s\S]/)) {
    return input.replace(/^\[|]$/g, '')
  }
  return input.replace(/\\/g, '')
}

function formatTimezone (offset, delimeter) {
  delimeter = delimeter || '';
  var sign = offset > 0 ? '-' : '+';
  var absOffset = Math.abs(offset);
  var hours = Math.floor(absOffset / 60);
  var minutes = absOffset % 60;
  return sign + addLeadingZeros(hours, 2) + delimeter + addLeadingZeros(minutes, 2)
}

function addLeadingZeros (number, targetLength) {
  var output = Math.abs(number).toString();
  while (output.length < targetLength) {
    output = '0' + output;
  }
  return output
}

var format_1 = format;

var oneOf$1 = function oneOf(value, validList) {
    for (var i = 0; i < validList.length; i++) {
        if (value === validList[i]) {
            return true;
        }
    }
    return false;
};

var BkDate = function () {
    function BkDate(weekdays) {
        classCallCheck(this, BkDate);

        this.weekdays = weekdays;

        var dater = new Date();

        this.currentDay = {
            year: dater.getFullYear(),
            month: dater.getMonth() + 1,
            day: dater.getDate()
        };

        this.currentTime = {
            hour: dater.getHours(),
            minute: dater.getMinutes() + 1,
            second: dater.getSeconds()
        };

        this.year = this.currentDay.year;
        this.month = this.currentDay.month;
        this.day = this.currentDay.day;
    }

    createClass(BkDate, [{
        key: 'setDate',
        value: function setDate(date) {
            var dateItems = date.split('-');
            if (dateItems[0]) {
                this.year = parseInt(dateItems[0]);
            }
            if (dateItems[1]) {
                this.month = parseInt(dateItems[1]);
            }
            if (dateItems[2]) {
                this.day = parseInt(dateItems[2]);
            }
        }
    }, {
        key: 'formatDateString',
        value: function formatDateString(value) {
            return parseInt(value) < 10 ? '0' + value : value;
        }
    }, {
        key: 'getFormatDate',
        value: function getFormatDate() {
            return this.year + '-' + this.formatDateString(this.month) + '-' + this.formatDateString(this.day);
        }
    }, {
        key: 'getCurrentMouthDays',
        value: function getCurrentMouthDays() {
            return new Date(this.year, this.month, 0).getDate();
        }
    }, {
        key: 'getLastMouthDays',
        value: function getLastMouthDays() {
            return new Date(this.year, this.month - 1, 0).getDate();
        }
    }, {
        key: 'getCurrentMonthBeginWeek',
        value: function getCurrentMonthBeginWeek() {
            return new Date(this.year, this.month - 1, 1).getDay();
        }
    }]);
    return BkDate;
}();

var bkDatePicker = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { directives: [{ name: "clickoutside", rawName: "v-clickoutside", value: _vm.close, expression: "close" }], staticClass: "bk-date-picker", class: _vm.disabled ? 'disabled' : '', on: { "click": _vm.openDater } }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.selectedValue, expression: "selectedValue" }], attrs: { "type": "text", "name": "date-select", "readonly": "readonly", "disabled": _vm.disabled, "placeholder": _vm.t('datePicker.selectDate') }, domProps: { "value": _vm.selectedValue }, on: { "input": function input($event) {
                    if ($event.target.composing) {
                        return;
                    }_vm.selectedValue = $event.target.value;
                } } }), _vm._v(" "), _c('transition', { attrs: { "name": _vm.transitionName } }, [_vm.showDatePanel ? _c('div', { staticClass: "date-dropdown-panel", style: _vm.panelStyle }, [_c('div', { staticClass: "date-top-bar" }, [_c('span', { staticClass: "year-switch-icon pre-year fl", on: { "click": function click($event) {
                    _vm.switchToYear('last');
                } } }), _vm._v(" "), _c('span', { staticClass: "month-switch-icon pre-month fl", on: { "click": function click($event) {
                    _vm.switchToMonth('last');
                } } }), _vm._v(" "), _c('span', { staticClass: "current-date" }, [_vm._v(_vm._s(_vm.topBarFormatView))]), _vm._v(" "), _c('span', { staticClass: "year-switch-icon next-year fr", on: { "click": function click($event) {
                    _vm.switchToYear('next');
                } } }), _vm._v(" "), _c('span', { staticClass: "month-switch-icon next-month fr", on: { "click": function click($event) {
                    _vm.switchToMonth('next');
                } } })]), _vm._v(" "), _c('div', { staticClass: "date-select-panel" }, [_c('dl', [_c('dt', _vm._l(_vm.BkDate.weekdays, function (day) {
            return _c('span', { staticClass: "date-item-view", domProps: { "innerHTML": _vm._s(day) } });
        })), _vm._v(" "), _vm._l(_vm.lastMonthList, function (lastMonthItem) {
            return _c('dd', [_c('span', { staticClass: "date-item-view date-disable-item" }, [_vm._v(_vm._s(lastMonthItem))])]);
        }), _vm._v(" "), _vm._l(_vm.BkDate.getCurrentMouthDays(), function (currentMonthItem) {
            return _c('dd', [_c('span', { class: { 'date-table-item': _vm.isAvailableDate(currentMonthItem), 'date-item-view date-disable-item': !_vm.isAvailableDate(currentMonthItem), 'selected': _vm.shouldBeSelected(currentMonthItem), 'today': _vm.shouldShowToday(currentMonthItem) === _vm.t('datePicker.today') }, on: { "click": function click($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.selectDay(currentMonthItem);
                    } } }, [_vm._v(_vm._s(_vm.shouldShowToday(currentMonthItem)))])]);
        }), _vm._v(" "), _vm._l(_vm.nextMonthList, function (nextMonthItem) {
            return _c('dd', [_c('span', { staticClass: "date-item-view date-disable-item" }, [_vm._v(_vm._s(nextMonthItem))])]);
        })], 2)]), _vm._v(" "), _vm.timer ? _c('div', { staticClass: "time-set-panel" }, _vm._l(_vm.currentTime, function (timeItem, index) {
            return _c('div', { staticClass: "time-item" }, [_c('input', { attrs: { "type": "number", "name": "" }, domProps: { "value": timeItem }, on: { "blur": function blur($event) {
                        _vm.blurTime($event, index);
                    }, "input": function input($event) {
                        _vm.inputTime($event, index);
                    } } })]);
        })) : _vm._e()]) : _vm._e()])], 1);
    }, staticRenderFns: [],
    name: 'bk-date-picker',
    mixins: [locale$1],
    directives: {
        clickoutside: clickoutside
    },
    props: {
        autoClose: {
            type: Boolean,
            default: true
        },
        disabled: {
            type: Boolean,
            default: false
        },
        timer: {
            type: Boolean,
            default: false
        },
        initDate: {
            type: String,
            default: ''
        },
        startDate: {
            type: String,
            default: ''
        },
        endDate: {
            type: String,
            default: ''
        },
        position: {
            validator: function validator(value) {
                return oneOf$1(value, ['top', 'bottom']);
            },

            default: 'bottom'
        }
    },
    data: function data() {
        var transitionName = 'toggle-slide';
        var panelStyle = {};
        var positionArr = this.position.split('-');
        if (positionArr.indexOf('top') > -1) {
            panelStyle.bottom = '38px';
            transitionName = 'toggle-slide2';
        } else {
            panelStyle.top = '38px';
        }

        var weekdays = [this.t('datePicker.weekdays.sun'), this.t('datePicker.weekdays.mon'), this.t('datePicker.weekdays.tue'), this.t('datePicker.weekdays.wed'), this.t('datePicker.weekdays.thu'), this.t('datePicker.weekdays.fri'), this.t('datePicker.weekdays.sat')];

        var bkDate = new BkDate(weekdays);

        return {
            panelStyle: panelStyle,
            transitionName: transitionName,
            BkDate: bkDate,
            selectedValue: this.initDate || '',
            currentDate: new Date(),
            showDatePanel: false,
            isSetTimer: false,
            firstTime: true
        };
    },

    computed: {
        topBarFormatView: function topBarFormatView() {

            var month = this.BkDate.month >= 10 ? this.BkDate.month : '0' + this.BkDate.month;

            return this.t('datePicker.topBarFormatView', {
                mmmm: formatMonth(month, this.t('lang')),
                mm: formatMonth(month, this.t('lang'), true),
                yyyy: this.BkDate.year
            });
        },
        lastMonthList: function lastMonthList() {
            var lastMonthVisibleNum = this.BkDate.getCurrentMonthBeginWeek();
            var lastMonthDays = this.BkDate.getLastMouthDays();
            var lastMonthVisibleList = [];
            for (var i = lastMonthVisibleNum - 1; i >= 0; i--) {
                lastMonthVisibleList.push(lastMonthDays - i);
            }
            return lastMonthVisibleList;
        },
        nextMonthList: function nextMonthList() {
            var lastMonthVisibleNum = this.BkDate.getCurrentMonthBeginWeek();
            var currentMonthDays = this.BkDate.getCurrentMouthDays();
            var nextMonthVisibleList = 42 - lastMonthVisibleNum - currentMonthDays;
            return nextMonthVisibleList;
        },
        currentTime: function currentTime() {
            var time = [];
            if (this.firstTime) {
                time = [this.formatValue(this.BkDate.currentTime.hour), this.formatValue(this.BkDate.currentTime.minute), this.formatValue(this.BkDate.currentTime.second)];
                this.firstTime = false;
            } else {
                time = [this.formatTime(this.BkDate.currentTime.hour), this.formatTime(this.BkDate.currentTime.minute), this.formatTime(this.BkDate.currentTime.second)];
            }
            return time;
        }
    },
    watch: {
        initDate: function initDate() {
            this.BkDate.setDate(this.initDate);
            if (this.selectedValue !== this.initDate) {
                this.$emit('change', this.selectedValue, this.initDate);
            }
            this.showDate();
            this.$emit('date-selected', this.selectedValue);

            if (this.autoClose && !this.isSetTimer) {
                this.close();
            }
        }
    },
    created: function created() {
        this.BkDate.setDate(this.initDate);
    },

    methods: {
        selectDay: function selectDay(value) {
            if (!this.isAvailableDate(value)) {
                return;
            }

            var newSelectedDate = this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(value);

            if (this.timer) {
                newSelectedDate += ' ' + this.formatValue(this.BkDate.currentTime.hour) + ':' + this.formatValue(this.BkDate.currentTime.minute) + ':' + this.formatValue(this.BkDate.currentTime.second);
            }

            if (this.selectedValue !== newSelectedDate) {
                this.$emit('change', newSelectedDate, this.selectedValue);
            }

            this.BkDate.setDate(newSelectedDate);
            this.showDate();

            this.$emit('date-selected', this.selectedValue);

            if (this.autoClose) {
                this.close();
            }
        },
        showDate: function showDate() {
            var selectedDate = this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(this.BkDate.day);

            var selectedTime = void 0;
            if (this.timer) {
                selectedTime = ' ' + this.formatTime(this.BkDate.currentTime.hour) + ':' + this.formatTime(this.BkDate.currentTime.minute) + ':' + this.formatTime(this.BkDate.currentTime.second);
            } else {
                selectedTime = '';
            }
            this.selectedValue = selectedDate + '' + selectedTime;
        },
        formatValue: function formatValue(value) {
            return parseInt(value) < 10 ? '0' + value : value;
        },
        formatTime: function formatTime(value) {
            return value;
        },
        shouldBeSelected: function shouldBeSelected(value) {
            return this.BkDate.day === value;
        },
        shouldShowToday: function shouldShowToday(value) {
            var currentSelectedDate = {
                year: this.BkDate.year,
                month: this.BkDate.month,
                day: value
            };
            var current = {
                year: this.currentDate.getFullYear(),
                month: this.currentDate.getMonth() + 1,
                day: this.currentDate.getDate()
            };
            var isToday = JSON.stringify(currentSelectedDate) === JSON.stringify(current);
            if (isToday) {
                return this.t('datePicker.today');
            }
            return value;
        },
        switchToMonth: function switchToMonth(type) {
            var toMonthDate = {};
            var year = this.BkDate.year;
            var month = this.BkDate.month;
            switch (type) {
                case 'last':
                    toMonthDate.year = month - 1 > 0 ? year : year - 1;
                    toMonthDate.month = month - 1 > 0 ? month - 1 : 12;
                    break;
                case 'next':
                    toMonthDate.year = month + 1 > 12 ? year + 1 : year;
                    toMonthDate.month = month + 1 > 12 ? 1 : month + 1;
                    break;
                default:
                    break;
            }

            this.BkDate.setDate(toMonthDate.year + '-' + toMonthDate.month + '-' + this.BkDate.day);
        },
        switchToYear: function switchToYear(type) {
            var toYearDate = {};
            var year = this.BkDate.year;
            switch (type) {
                case 'last':
                    toYearDate.year = year - 1 > 0 ? year - 1 : 0;
                    break;
                case 'next':
                    toYearDate.year = year + 1;
                    break;
                default:
                    break;
            }

            this.BkDate.setDate(toYearDate.year + '-' + this.BkDate.month + '-' + this.BkDate.day);
        },
        setTime: function setTime(type, index) {
            var option = ['hour', 'minute', 'second'][index];
            var defaultTime = _extends({}, this.BkDate.currentTime);
            defaultTime.hour = Number(defaultTime.hour);
            defaultTime.minute = Number(defaultTime.minute);
            defaultTime.second = Number(defaultTime.second);
            switch (option) {
                case 'hour':
                    if (type === 'up') {
                        defaultTime.hour = defaultTime.hour + 1 < 24 ? defaultTime.hour + 1 > 10 ? defaultTime.hour + 1 : '0' + (defaultTime.hour + 1) : '00';
                    }
                    if (type === 'down') {
                        defaultTime.hour = defaultTime.hour - 1 >= 0 ? defaultTime.hour - 1 > 10 ? defaultTime.hour - 1 : '0' + (defaultTime.hour - 1) : 23;
                    }
                    break;
                case 'minute':
                    if (type === 'up') {
                        defaultTime.minute = defaultTime.minute + 1 < 60 ? defaultTime.minute + 1 > 10 ? defaultTime.minute + 1 : '0' + (defaultTime.minute + 1) : '00';
                    }
                    if (type === 'down') {
                        defaultTime.minute = defaultTime.minute - 1 >= 0 ? defaultTime.minute - 1 > 10 ? defaultTime.minute - 1 : '0' + (defaultTime.minute - 1) : 59;
                    }
                    break;
                case 'second':
                    if (type === 'up') {
                        defaultTime.second = defaultTime.second + 1 < 60 ? defaultTime.second + 1 > 10 ? defaultTime.second + 1 : '0' + (defaultTime.second + 1) : '00';
                    }
                    if (type === 'down') {
                        defaultTime.second = defaultTime.second - 1 >= 0 ? defaultTime.second - 1 > 10 ? defaultTime.second - 1 : '0' + (defaultTime.second - 1) : 59;
                    }
                    break;
                default:
            }

            this.timeCommon(defaultTime);
        },
        inputTime: function inputTime(event, index) {
            var timeVal = event.target.value;
            var option = ['hour', 'minute', 'second'][index];
            var defaultTime = _extends({}, this.BkDate.currentTime);
            switch (option) {
                case 'hour':
                    var hourRes = /^(2[0-3]|[0-1]?\d)$/;
                    defaultTime.hour = hourRes.test(timeVal) ? timeVal : '';
                    break;
                case 'minute':
                    var minuteRes = /^[0-5]?[0-9]$/;
                    defaultTime.minute = minuteRes.test(timeVal) ? timeVal : '';
                    break;
                case 'second':
                    var secondRes = /^[0-5]?[0-9]$/;
                    defaultTime.second = secondRes.test(timeVal) ? timeVal : '';
                    break;
                default:
            }
            this.timeCommon(defaultTime);
        },
        blurTime: function blurTime(event, index) {
            var timeVal = event.target.value;
            var option = ['hour', 'minute', 'second'][index];
            var defaultTime = _extends({}, this.BkDate.currentTime);
            var timeInfo = (timeVal === '' ? '00' : Number(timeVal) < 10 ? '0' + Number(timeVal) : timeVal).slice(0, 2);
            switch (option) {
                case 'hour':
                    defaultTime.hour = timeInfo;
                    break;
                case 'minute':
                    defaultTime.minute = timeInfo;
                    break;
                case 'second':
                    defaultTime.second = timeInfo;
                    break;
                default:
            }

            this.timeCommon(defaultTime);
        },
        timeCommon: function timeCommon(defaultTime) {
            this.BkDate.currentTime = _extends({}, defaultTime);
            this.showDate();
            this.isSetTimer = true;
            if (this.selectedValue !== this.initDate) {
                this.$emit('change', this.initDate, this.selectedValue);
            }
            this.$emit('date-selected', this.selectedValue);
        },
        isAvailableDate: function isAvailableDate(day) {
            var cmpTime = new Date(this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(day)).getTime();
            var startTime = void 0,
                endTime = void 0;
            var checkStartTime = true;
            var checkEndTime = true;
            if (this.startDate) {
                startTime = new Date(this.startDate).getTime();
                checkStartTime = cmpTime >= startTime;
            }
            if (this.endDate) {
                endTime = new Date(this.endDate).getTime();
                checkEndTime = cmpTime <= endTime;
            }
            return checkStartTime && checkEndTime;
        },
        openDater: function openDater() {
            if (this.disabled) return;
            this.showDatePanel = true;
        },
        close: function close() {
            this.showDatePanel = false;
            this.isSetTimer = false;
        }
    }
};

bkDatePicker.install = function (Vue$$1) {
  Vue$$1.component(bkDatePicker.name, bkDatePicker);
};

var datepicker = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "date-picker" }, [_c('div', { staticClass: "date-top-bar" }, [_vm.bkDate.preYearDisabled ? _c('span', { staticClass: "year-switch-icon pre-year fl is-disabled" }) : _c('span', { staticClass: "year-switch-icon pre-year fl", on: { "click": function click($event) {
                    _vm.switchToYear('last');
                } } }), _vm._v(" "), _vm.bkDate.preMonthDisabled ? _c('span', { staticClass: "month-switch-icon pre-month fl is-disabled" }) : _c('span', { staticClass: "month-switch-icon pre-month fl", on: { "click": function click($event) {
                    _vm.switchToMonth('last');
                } } }), _vm._v(" "), _c('span', { staticClass: "current-date" }, [_vm._v(_vm._s(_vm.topBarFormatView.text))]), _vm._v(" "), _vm.bkDate.nextYearDisabled ? _c('span', { staticClass: "year-switch-icon next-year fr is-disabled" }) : _c('span', { staticClass: "year-switch-icon next-year fr", on: { "click": function click($event) {
                    _vm.switchToYear('next');
                } } }), _vm._v(" "), _vm.bkDate.nextMonthDisabled ? _c('span', { staticClass: "month-switch-icon next-month fr is-disabled" }) : _c('span', { staticClass: "month-switch-icon next-month fr", on: { "click": function click($event) {
                    _vm.switchToMonth('next');
                } } })]), _vm._v(" "), _c('div', { staticClass: "date-select-panel" }, [_c('dl', [_c('dt', _vm._l(_vm.BkDate.weekdays, function (day) {
            return _c('span', { staticClass: "date-item-view", domProps: { "innerHTML": _vm._s(day) } });
        })), _vm._v(" "), _vm._l(_vm.lastMonthList, function (lastMonthItem) {
            return _c('dd', [_c('span', { staticClass: "date-item-view date-disable-item" }, [_vm._v(_vm._s(lastMonthItem))])]);
        }), _vm._v(" "), _vm._l(_vm.BkDate.getCurrentMouthDays(), function (currentMonthItem) {
            return _c('dd', [_c('span', { class: { 'date-table-item': _vm.isAvailableDate(currentMonthItem) && !_vm.isDisabledDate(currentMonthItem), 'date-item-view date-disable-item': !_vm.isAvailableDate(currentMonthItem) || _vm.isDisabledDate(currentMonthItem), 'date-range-view': _vm.isInRange(currentMonthItem), 'today': _vm.shouldShowToday(currentMonthItem) === _vm.t('dateRange.datePicker.today'), 'selected': _vm.shouldBeSelected(currentMonthItem) }, on: { "click": function click($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.selectDay(currentMonthItem);
                    } } }, [_vm._v(_vm._s(_vm.shouldShowToday(currentMonthItem)))])]);
        }), _vm._v(" "), _vm._l(_vm.nextMonthList, function (nextMonthItem) {
            return _c('dd', [_c('span', { staticClass: "date-item-view date-disable-item" }, [_vm._v(_vm._s(nextMonthItem))])]);
        })], 2)]), _vm._v(" "), _vm.timer ? _c('div', { staticClass: "time-set-panel" }, [_vm.BkDate.setTimer ? _vm._l(_vm.currentTime, function (timeItem, index) {
            return _c('div', { staticClass: "time-item" }, [_c('input', { attrs: { "readonly": "readonly", "type": "number", "name": "" }, domProps: { "value": timeItem } }), _vm._v(" "), _c('span', { staticClass: "time-option fr" }, [_c('i', { staticClass: "up", on: { "click": function click($event) {
                        $event.preventDefault();$event.stopPropagation();_vm.setTime('up', index);
                    } } }), _vm._v(" "), _c('i', { staticClass: "down", on: { "click": function click($event) {
                        $event.preventDefault();$event.stopPropagation();_vm.setTime('down', index);
                    } } })])]);
        }) : _vm._l(_vm.currentTime, function (timeItem, index) {
            return _c('div', { staticClass: "time-item" }, [_c('input', { attrs: { "disabled": "disabled", "type": "number", "name": "" }, domProps: { "value": timeItem } }), _vm._v(" "), _vm._m(0, true)]);
        })], 2) : _vm._e()]);
    }, staticRenderFns: [function () {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('span', { staticClass: "time-option fr" }, [_c('i', { staticClass: "up no-hover" }), _vm._v(" "), _c('i', { staticClass: "down no-hover" })]);
    }],
    mixins: [locale$1],
    props: {
        initDate: {
            type: String,
            default: ''
        },
        startDate: {
            type: String,
            default: ''
        },
        endDate: {
            type: String,
            default: ''
        },
        selectedRange: {
            type: Array,
            default: function _default() {
                return [];
            }
        },

        selectedRangeTmp: {
            type: Array,
            default: function _default() {
                return [];
            }
        },
        timer: {
            type: Boolean,
            default: false
        },
        bkDate: {
            type: Object,
            default: function _default() {
                return {};
            }
        },
        type: {
            type: String,
            default: ''
        },
        dateLimit: {
            type: String,
            default: ''
        }
    },
    directives: {
        clickoutside: clickoutside
    },
    data: function data() {
        return {
            BkDate: this.bkDate,
            selectedValue: this.initDate,
            currentDate: new Date()
        };
    },

    computed: {
        topBarFormatView: function topBarFormatView() {
            var month = this.formatValue(this.BkDate.month);
            return {
                text: this.t('dateRange.datePicker.topBarFormatView', {
                    mmmm: formatMonth(month, this.t('lang')),
                    mm: formatMonth(month, this.t('lang'), true),
                    yyyy: this.BkDate.year
                }),
                value: this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-01'
            };
        },
        lastMonthList: function lastMonthList() {
            var lastMonthVisibleNum = this.BkDate.getCurrentMonthBeginWeek();
            var lastMonthDays = this.BkDate.getLastMouthDays();
            var lastMonthVisibleList = [];
            for (var i = lastMonthVisibleNum - 1; i >= 0; i--) {
                lastMonthVisibleList.push(lastMonthDays - i);
            }
            return lastMonthVisibleList;
        },
        nextMonthList: function nextMonthList() {
            var lastMonthVisibleNum = this.BkDate.getCurrentMonthBeginWeek();
            var currentMonthDays = this.BkDate.getCurrentMouthDays();
            var nextMonthVisibleList = 42 - lastMonthVisibleNum - currentMonthDays;
            return nextMonthVisibleList;
        },
        currentTime: function currentTime() {
            var time = [this.formatValue(this.BkDate.currentTime.hour), this.formatValue(this.BkDate.currentTime.minute), this.formatValue(this.BkDate.currentTime.second)];
            return time;
        }
    },
    watch: {
        bkDate: function bkDate(value) {
            this.BkDate = value;
        }
    },
    methods: {
        selectDay: function selectDay(value) {
            if (!this.isAvailableDate(value) || this.isDisabledDate(value)) return;

            var newSelectedDate = this.BkDate.year + '-' + this.BkDate.month + '-' + value;
            this.BkDate.setDate(newSelectedDate);

            var selectedDate = this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(this.BkDate.day);

            var selectedTime = '';
            if (this.timer) {
                selectedTime = ' ' + this.formatValue(this.BkDate.currentTime.hour) + ':' + this.formatValue(this.BkDate.currentTime.minute) + ':' + this.formatValue(this.BkDate.currentTime.second);
            } else {
                selectedTime = '';
            }

            this.selectedValue = selectedDate + '' + selectedTime;
            this.$emit('date-selected', this.selectedValue);
        },
        formatValue: function formatValue(value) {
            return parseInt(value) < 10 ? 0 + '' + value : value;
        },
        shouldBeSelected: function shouldBeSelected(value) {
            var selectedDate = this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(value);

            var triggerDate = selectedDate + '';
            return this.selectedRangeTmp.indexOf(triggerDate) >= 0;
        },
        shouldShowToday: function shouldShowToday(value) {
            var currentSelectedDate = {
                year: this.BkDate.year,
                month: this.BkDate.month,
                day: value
            };
            var current = {
                year: this.currentDate.getFullYear(),
                month: this.currentDate.getMonth() + 1,
                day: this.currentDate.getDate()
            };
            var isToday = JSON.stringify(currentSelectedDate) === JSON.stringify(current);
            if (isToday) {
                return this.t('dateRange.datePicker.today');
            }
            return value;
        },
        switchToMonth: function switchToMonth(type) {
            var toMonthDate = {};
            var year = this.BkDate.year;
            var month = this.BkDate.month;
            switch (type) {
                case 'last':
                    toMonthDate.year = month - 1 > 0 ? year : year - 1;
                    toMonthDate.month = month - 1 > 0 ? month - 1 : 12;
                    break;
                case 'next':
                    toMonthDate.year = month + 1 > 12 ? year + 1 : year;
                    toMonthDate.month = month + 1 > 12 ? 1 : month + 1;
                    break;
                default:
                    break;
            }

            this.BkDate.setDate(toMonthDate.year + '-' + toMonthDate.month + '-' + this.BkDate.day);
            this.$emit('date-quick-switch', {
                type: type,

                value: toMonthDate.year + '-' + this.formatValue(toMonthDate.month) + '-01'
            });
        },
        switchToYear: function switchToYear(type) {
            var toYearDate = {};
            var year = this.BkDate.year;
            switch (type) {
                case 'last':
                    toYearDate.year = year - 1 > 0 ? year - 1 : 0;
                    break;
                case 'next':
                    toYearDate.year = year + 1;
                    break;
                default:
                    break;
            }

            this.BkDate.setDate(toYearDate.year + '-' + this.BkDate.month + '-' + this.BkDate.day);
            this.$emit('date-quick-switch', {
                type: type,
                value: toYearDate.year + '-' + this.formatValue(this.BkDate.month) + '-01'
            });
        },
        isDisabledDate: function isDisabledDate(day) {
            var cmpTime = new Date(this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(day)).getTime();
            var timeStamp = void 0;
            if (this.dateLimit) {
                timeStamp = new Date(this.dateLimit).getTime();
                if (this.type === 'start') {
                    return cmpTime < timeStamp;
                }
                if (this.type === 'end') {
                    return cmpTime > timeStamp;
                }
            } else {
                return false;
            }
        },
        isAvailableDate: function isAvailableDate(day) {
            var cmpTime = new Date(this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(day)).getTime();
            var startTime = void 0,
                endTime = void 0;
            var checkStartTime = true;
            var checkEndTime = true;
            if (this.startDate) {
                startTime = new Date(this.startDate).getTime();
                checkStartTime = cmpTime >= startTime;
            }
            if (this.endDate) {
                endTime = new Date(this.endDate).getTime();
                checkEndTime = cmpTime <= endTime;
            }
            return checkStartTime && checkEndTime;
        },
        isInRange: function isInRange(day) {
            if (!this.selectedRange[0]) return false;

            var dayTime = new Date(this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(day)).getTime();
            var startDateTime = new Date(this.selectedRange[0]).getTime();
            var endDateTime = new Date(this.selectedRange[1]).getTime();
            return dayTime - startDateTime > 0 && dayTime - endDateTime < 0;
        },
        setTime: function setTime(type, index) {
            var option = ['hour', 'minute', 'second'][index];
            var defaultTime = _extends({}, this.BkDate.currentTime);
            switch (option) {
                case 'hour':
                    if (type === 'up') {
                        defaultTime.hour = defaultTime.hour + 1 < 24 ? defaultTime.hour + 1 : 0;
                    }
                    if (type === 'down') {
                        defaultTime.hour = defaultTime.hour - 1 >= 0 ? defaultTime.hour - 1 : 23;
                    }
                    break;
                case 'minute':
                    if (type === 'up') {
                        defaultTime.minute = defaultTime.minute + 1 < 60 ? defaultTime.minute + 1 : 0;
                    }
                    if (type === 'down') {
                        defaultTime.minute = defaultTime.minute - 1 >= 0 ? defaultTime.minute - 1 : 59;
                    }
                    break;
                case 'second':
                    if (type === 'up') {
                        defaultTime.second = defaultTime.second + 1 < 60 ? defaultTime.second + 1 : 0;
                    }
                    if (type === 'down') {
                        defaultTime.second = defaultTime.second - 1 >= 0 ? defaultTime.second - 1 : 59;
                    }
                    break;
                default:
            }

            this.BkDate.currentTime = _extends({}, defaultTime);

            var selectedDate = this.BkDate.year + '-' + this.formatValue(this.BkDate.month) + '-' + this.formatValue(this.BkDate.day);
            var selectedTime = '';
            if (this.timer) {
                selectedTime = ' ' + this.formatValue(this.BkDate.currentTime.hour) + ':' + this.formatValue(this.BkDate.currentTime.minute) + ':' + this.formatValue(this.BkDate.currentTime.second);
            } else {
                selectedTime = '';
            }
            this.selectedValue = selectedDate + '' + selectedTime;

            this.$emit('date-selected', this.selectedValue, this.BkDate.index);
        }
    }
};

/**
 * @category Day Helpers
 * @summary Add the specified number of days to the given date.
 *
 * @description
 * Add the specified number of days to the given date.
 *
 * @param {Date|String|Number} date - the date to be changed
 * @param {Number} amount - the amount of days to be added
 * @returns {Date} the new date with the days added
 *
 * @example
 * // Add 10 days to 1 September 2014:
 * var result = addDays(new Date(2014, 8, 1), 10)
 * //=> Thu Sep 11 2014 00:00:00
 */
function addDays (dirtyDate, dirtyAmount) {
  var date = parse_1(dirtyDate);
  var amount = Number(dirtyAmount);
  date.setDate(date.getDate() + amount);
  return date
}

var add_days = addDays;

/**
 * @category Day Helpers
 * @summary Subtract the specified number of days from the given date.
 *
 * @description
 * Subtract the specified number of days from the given date.
 *
 * @param {Date|String|Number} date - the date to be changed
 * @param {Number} amount - the amount of days to be subtracted
 * @returns {Date} the new date with the days subtracted
 *
 * @example
 * // Subtract 10 days from 1 September 2014:
 * var result = subDays(new Date(2014, 8, 1), 10)
 * //=> Fri Aug 22 2014 00:00:00
 */
function subDays (dirtyDate, dirtyAmount) {
  var amount = Number(dirtyAmount);
  return add_days(dirtyDate, -amount)
}

var sub_days = subDays;

/**
 * @category Month Helpers
 * @summary Get the number of days in a month of the given date.
 *
 * @description
 * Get the number of days in a month of the given date.
 *
 * @param {Date|String|Number} date - the given date
 * @returns {Number} the number of days in a month
 *
 * @example
 * // How many days are in February 2000?
 * var result = getDaysInMonth(new Date(2000, 1))
 * //=> 29
 */
function getDaysInMonth (dirtyDate) {
  var date = parse_1(dirtyDate);
  var year = date.getFullYear();
  var monthIndex = date.getMonth();
  var lastDayOfMonth = new Date(0);
  lastDayOfMonth.setFullYear(year, monthIndex + 1, 0);
  lastDayOfMonth.setHours(0, 0, 0, 0);
  return lastDayOfMonth.getDate()
}

var get_days_in_month = getDaysInMonth;

/**
 * @category Month Helpers
 * @summary Add the specified number of months to the given date.
 *
 * @description
 * Add the specified number of months to the given date.
 *
 * @param {Date|String|Number} date - the date to be changed
 * @param {Number} amount - the amount of months to be added
 * @returns {Date} the new date with the months added
 *
 * @example
 * // Add 5 months to 1 September 2014:
 * var result = addMonths(new Date(2014, 8, 1), 5)
 * //=> Sun Feb 01 2015 00:00:00
 */
function addMonths (dirtyDate, dirtyAmount) {
  var date = parse_1(dirtyDate);
  var amount = Number(dirtyAmount);
  var desiredMonth = date.getMonth() + amount;
  var dateWithDesiredMonth = new Date(0);
  dateWithDesiredMonth.setFullYear(date.getFullYear(), desiredMonth, 1);
  dateWithDesiredMonth.setHours(0, 0, 0, 0);
  var daysInMonth = get_days_in_month(dateWithDesiredMonth);
  // Set the last day of the new month
  // if the original date was the last day of the longer month
  date.setMonth(desiredMonth, Math.min(daysInMonth, date.getDate()));
  return date
}

var add_months = addMonths;

/**
 * @category Month Helpers
 * @summary Subtract the specified number of months from the given date.
 *
 * @description
 * Subtract the specified number of months from the given date.
 *
 * @param {Date|String|Number} date - the date to be changed
 * @param {Number} amount - the amount of months to be subtracted
 * @returns {Date} the new date with the months subtracted
 *
 * @example
 * // Subtract 5 months from 1 February 2015:
 * var result = subMonths(new Date(2015, 1, 1), 5)
 * //=> Mon Sep 01 2014 00:00:00
 */
function subMonths (dirtyDate, dirtyAmount) {
  var amount = Number(dirtyAmount);
  return add_months(dirtyDate, -amount)
}

var sub_months = subMonths;

/**
 * @category Hour Helpers
 * @summary Get the hours of the given date.
 *
 * @description
 * Get the hours of the given date.
 *
 * @param {Date|String|Number} date - the given date
 * @returns {Number} the hours
 *
 * @example
 * // Get the hours of 29 February 2012 11:45:00:
 * var result = getHours(new Date(2012, 1, 29, 11, 45))
 * //=> 11
 */
function getHours (dirtyDate) {
  var date = parse_1(dirtyDate);
  var hours = date.getHours();
  return hours
}

var get_hours = getHours;

/**
 * @category Minute Helpers
 * @summary Get the minutes of the given date.
 *
 * @description
 * Get the minutes of the given date.
 *
 * @param {Date|String|Number} date - the given date
 * @returns {Number} the minutes
 *
 * @example
 * // Get the minutes of 29 February 2012 11:45:05:
 * var result = getMinutes(new Date(2012, 1, 29, 11, 45, 5))
 * //=> 45
 */
function getMinutes (dirtyDate) {
  var date = parse_1(dirtyDate);
  var minutes = date.getMinutes();
  return minutes
}

var get_minutes = getMinutes;

/**
 * @category Second Helpers
 * @summary Get the seconds of the given date.
 *
 * @description
 * Get the seconds of the given date.
 *
 * @param {Date|String|Number} date - the given date
 * @returns {Number} the seconds
 *
 * @example
 * // Get the seconds of 29 February 2012 11:45:05.123:
 * var result = getSeconds(new Date(2012, 1, 29, 11, 45, 5, 123))
 * //=> 5
 */
function getSeconds (dirtyDate) {
  var date = parse_1(dirtyDate);
  var seconds = date.getSeconds();
  return seconds
}

var get_seconds = getSeconds;

/**
 * @category Second Helpers
 * @summary Set the seconds to the given date.
 *
 * @description
 * Set the seconds to the given date.
 *
 * @param {Date|String|Number} date - the date to be changed
 * @param {Number} seconds - the seconds of the new date
 * @returns {Date} the new date with the seconds setted
 *
 * @example
 * // Set 45 seconds to 1 September 2014 11:30:40:
 * var result = setSeconds(new Date(2014, 8, 1, 11, 30, 40), 45)
 * //=> Mon Sep 01 2014 11:30:45
 */
function setSeconds (dirtyDate, dirtySeconds) {
  var date = parse_1(dirtyDate);
  var seconds = Number(dirtySeconds);
  date.setSeconds(seconds);
  return date
}

var set_seconds = setSeconds;

/**
 * @category Minute Helpers
 * @summary Set the minutes to the given date.
 *
 * @description
 * Set the minutes to the given date.
 *
 * @param {Date|String|Number} date - the date to be changed
 * @param {Number} minutes - the minutes of the new date
 * @returns {Date} the new date with the minutes setted
 *
 * @example
 * // Set 45 minutes to 1 September 2014 11:30:40:
 * var result = setMinutes(new Date(2014, 8, 1, 11, 30, 40), 45)
 * //=> Mon Sep 01 2014 11:45:40
 */
function setMinutes (dirtyDate, dirtyMinutes) {
  var date = parse_1(dirtyDate);
  var minutes = Number(dirtyMinutes);
  date.setMinutes(minutes);
  return date
}

var set_minutes = setMinutes;

/**
 * @category Hour Helpers
 * @summary Set the hours to the given date.
 *
 * @description
 * Set the hours to the given date.
 *
 * @param {Date|String|Number} date - the date to be changed
 * @param {Number} hours - the hours of the new date
 * @returns {Date} the new date with the hours setted
 *
 * @example
 * // Set 4 hours to 1 September 2014 11:30:00:
 * var result = setHours(new Date(2014, 8, 1, 11, 30), 4)
 * //=> Mon Sep 01 2014 04:30:00
 */
function setHours (dirtyDate, dirtyHours) {
  var date = parse_1(dirtyDate);
  var hours = Number(dirtyHours);
  date.setHours(hours);
  return date
}

var set_hours = setHours;

/**
 * @category Common Helpers
 * @summary Is the first date after the second one?
 *
 * @description
 * Is the first date after the second one?
 *
 * @param {Date|String|Number} date - the date that should be after the other one to return true
 * @param {Date|String|Number} dateToCompare - the date to compare with
 * @returns {Boolean} the first date is after the second date
 *
 * @example
 * // Is 10 July 1989 after 11 February 1987?
 * var result = isAfter(new Date(1989, 6, 10), new Date(1987, 1, 11))
 * //=> true
 */
function isAfter (dirtyDate, dirtyDateToCompare) {
  var date = parse_1(dirtyDate);
  var dateToCompare = parse_1(dirtyDateToCompare);
  return date.getTime() > dateToCompare.getTime()
}

var is_after = isAfter;

/**
 * @category Common Helpers
 * @summary Is the first date before the second one?
 *
 * @description
 * Is the first date before the second one?
 *
 * @param {Date|String|Number} date - the date that should be before the other one to return true
 * @param {Date|String|Number} dateToCompare - the date to compare with
 * @returns {Boolean} the first date is before the second date
 *
 * @example
 * // Is 10 July 1989 before 11 February 1987?
 * var result = isBefore(new Date(1989, 6, 10), new Date(1987, 1, 11))
 * //=> false
 */
function isBefore (dirtyDate, dirtyDateToCompare) {
  var date = parse_1(dirtyDate);
  var dateToCompare = parse_1(dirtyDateToCompare);
  return date.getTime() < dateToCompare.getTime()
}

var is_before = isBefore;

/**
 * @category Year Helpers
 * @summary Are the given dates in the same year?
 *
 * @description
 * Are the given dates in the same year?
 *
 * @param {Date|String|Number} dateLeft - the first date to check
 * @param {Date|String|Number} dateRight - the second date to check
 * @returns {Boolean} the dates are in the same year
 *
 * @example
 * // Are 2 September 2014 and 25 September 2014 in the same year?
 * var result = isSameYear(
 *   new Date(2014, 8, 2),
 *   new Date(2014, 8, 25)
 * )
 * //=> true
 */
function isSameYear (dirtyDateLeft, dirtyDateRight) {
  var dateLeft = parse_1(dirtyDateLeft);
  var dateRight = parse_1(dirtyDateRight);
  return dateLeft.getFullYear() === dateRight.getFullYear()
}

var is_same_year = isSameYear;

/**
 * @category Month Helpers
 * @summary Are the given dates in the same month?
 *
 * @description
 * Are the given dates in the same month?
 *
 * @param {Date|String|Number} dateLeft - the first date to check
 * @param {Date|String|Number} dateRight - the second date to check
 * @returns {Boolean} the dates are in the same month
 *
 * @example
 * // Are 2 September 2014 and 25 September 2014 in the same month?
 * var result = isSameMonth(
 *   new Date(2014, 8, 2),
 *   new Date(2014, 8, 25)
 * )
 * //=> true
 */
function isSameMonth (dirtyDateLeft, dirtyDateRight) {
  var dateLeft = parse_1(dirtyDateLeft);
  var dateRight = parse_1(dirtyDateRight);
  return dateLeft.getFullYear() === dateRight.getFullYear() &&
    dateLeft.getMonth() === dateRight.getMonth()
}

var is_same_month = isSameMonth;

/**
 * @category Day Helpers
 * @summary Are the given dates in the same day?
 *
 * @description
 * Are the given dates in the same day?
 *
 * @param {Date|String|Number} dateLeft - the first date to check
 * @param {Date|String|Number} dateRight - the second date to check
 * @returns {Boolean} the dates are in the same day
 *
 * @example
 * // Are 4 September 06:00:00 and 4 September 18:00:00 in the same day?
 * var result = isSameDay(
 *   new Date(2014, 8, 4, 6, 0),
 *   new Date(2014, 8, 4, 18, 0)
 * )
 * //=> true
 */
function isSameDay (dirtyDateLeft, dirtyDateRight) {
  var dateLeftStartOfDay = start_of_day(dirtyDateLeft);
  var dateRightStartOfDay = start_of_day(dirtyDateRight);

  return dateLeftStartOfDay.getTime() === dateRightStartOfDay.getTime()
}

var is_same_day = isSameDay;

/**
 * @category Hour Helpers
 * @summary Return the start of an hour for the given date.
 *
 * @description
 * Return the start of an hour for the given date.
 * The result will be in the local timezone.
 *
 * @param {Date|String|Number} date - the original date
 * @returns {Date} the start of an hour
 *
 * @example
 * // The start of an hour for 2 September 2014 11:55:00:
 * var result = startOfHour(new Date(2014, 8, 2, 11, 55))
 * //=> Tue Sep 02 2014 11:00:00
 */
function startOfHour (dirtyDate) {
  var date = parse_1(dirtyDate);
  date.setMinutes(0, 0, 0);
  return date
}

var start_of_hour = startOfHour;

/**
 * @category Hour Helpers
 * @summary Are the given dates in the same hour?
 *
 * @description
 * Are the given dates in the same hour?
 *
 * @param {Date|String|Number} dateLeft - the first date to check
 * @param {Date|String|Number} dateRight - the second date to check
 * @returns {Boolean} the dates are in the same hour
 *
 * @example
 * // Are 4 September 2014 06:00:00 and 4 September 06:30:00 in the same hour?
 * var result = isSameHour(
 *   new Date(2014, 8, 4, 6, 0),
 *   new Date(2014, 8, 4, 6, 30)
 * )
 * //=> true
 */
function isSameHour (dirtyDateLeft, dirtyDateRight) {
  var dateLeftStartOfHour = start_of_hour(dirtyDateLeft);
  var dateRightStartOfHour = start_of_hour(dirtyDateRight);

  return dateLeftStartOfHour.getTime() === dateRightStartOfHour.getTime()
}

var is_same_hour = isSameHour;

/**
 * @category Minute Helpers
 * @summary Return the start of a minute for the given date.
 *
 * @description
 * Return the start of a minute for the given date.
 * The result will be in the local timezone.
 *
 * @param {Date|String|Number} date - the original date
 * @returns {Date} the start of a minute
 *
 * @example
 * // The start of a minute for 1 December 2014 22:15:45.400:
 * var result = startOfMinute(new Date(2014, 11, 1, 22, 15, 45, 400))
 * //=> Mon Dec 01 2014 22:15:00
 */
function startOfMinute (dirtyDate) {
  var date = parse_1(dirtyDate);
  date.setSeconds(0, 0);
  return date
}

var start_of_minute = startOfMinute;

/**
 * @category Minute Helpers
 * @summary Are the given dates in the same minute?
 *
 * @description
 * Are the given dates in the same minute?
 *
 * @param {Date|String|Number} dateLeft - the first date to check
 * @param {Date|String|Number} dateRight - the second date to check
 * @returns {Boolean} the dates are in the same minute
 *
 * @example
 * // Are 4 September 2014 06:30:00 and 4 September 2014 06:30:15
 * // in the same minute?
 * var result = isSameMinute(
 *   new Date(2014, 8, 4, 6, 30),
 *   new Date(2014, 8, 4, 6, 30, 15)
 * )
 * //=> true
 */
function isSameMinute (dirtyDateLeft, dirtyDateRight) {
  var dateLeftStartOfMinute = start_of_minute(dirtyDateLeft);
  var dateRightStartOfMinute = start_of_minute(dirtyDateRight);

  return dateLeftStartOfMinute.getTime() === dateRightStartOfMinute.getTime()
}

var is_same_minute = isSameMinute;

/**
 * @category Second Helpers
 * @summary Return the start of a second for the given date.
 *
 * @description
 * Return the start of a second for the given date.
 * The result will be in the local timezone.
 *
 * @param {Date|String|Number} date - the original date
 * @returns {Date} the start of a second
 *
 * @example
 * // The start of a second for 1 December 2014 22:15:45.400:
 * var result = startOfSecond(new Date(2014, 11, 1, 22, 15, 45, 400))
 * //=> Mon Dec 01 2014 22:15:45.000
 */
function startOfSecond (dirtyDate) {
  var date = parse_1(dirtyDate);
  date.setMilliseconds(0);
  return date
}

var start_of_second = startOfSecond;

/**
 * @category Second Helpers
 * @summary Are the given dates in the same second?
 *
 * @description
 * Are the given dates in the same second?
 *
 * @param {Date|String|Number} dateLeft - the first date to check
 * @param {Date|String|Number} dateRight - the second date to check
 * @returns {Boolean} the dates are in the same second
 *
 * @example
 * // Are 4 September 2014 06:30:15.000 and 4 September 2014 06:30.15.500
 * // in the same second?
 * var result = isSameSecond(
 *   new Date(2014, 8, 4, 6, 30, 15),
 *   new Date(2014, 8, 4, 6, 30, 15, 500)
 * )
 * //=> true
 */
function isSameSecond (dirtyDateLeft, dirtyDateRight) {
  var dateLeftStartOfSecond = start_of_second(dirtyDateLeft);
  var dateRightStartOfSecond = start_of_second(dirtyDateRight);

  return dateLeftStartOfSecond.getTime() === dateRightStartOfSecond.getTime()
}

var is_same_second = isSameSecond;

/**
 * @category Month Helpers
 * @summary Get the number of calendar months between the given dates.
 *
 * @description
 * Get the number of calendar months between the given dates.
 *
 * @param {Date|String|Number} dateLeft - the later date
 * @param {Date|String|Number} dateRight - the earlier date
 * @returns {Number} the number of calendar months
 *
 * @example
 * // How many calendar months are between 31 January 2014 and 1 September 2014?
 * var result = differenceInCalendarMonths(
 *   new Date(2014, 8, 1),
 *   new Date(2014, 0, 31)
 * )
 * //=> 8
 */
function differenceInCalendarMonths (dirtyDateLeft, dirtyDateRight) {
  var dateLeft = parse_1(dirtyDateLeft);
  var dateRight = parse_1(dirtyDateRight);

  var yearDiff = dateLeft.getFullYear() - dateRight.getFullYear();
  var monthDiff = dateLeft.getMonth() - dateRight.getMonth();

  return yearDiff * 12 + monthDiff
}

var difference_in_calendar_months = differenceInCalendarMonths;

/**
 * @category Common Helpers
 * @summary Compare the two dates and return -1, 0 or 1.
 *
 * @description
 * Compare the two dates and return 1 if the first date is after the second,
 * -1 if the first date is before the second or 0 if dates are equal.
 *
 * @param {Date|String|Number} dateLeft - the first date to compare
 * @param {Date|String|Number} dateRight - the second date to compare
 * @returns {Number} the result of the comparison
 *
 * @example
 * // Compare 11 February 1987 and 10 July 1989:
 * var result = compareAsc(
 *   new Date(1987, 1, 11),
 *   new Date(1989, 6, 10)
 * )
 * //=> -1
 *
 * @example
 * // Sort the array of dates:
 * var result = [
 *   new Date(1995, 6, 2),
 *   new Date(1987, 1, 11),
 *   new Date(1989, 6, 10)
 * ].sort(compareAsc)
 * //=> [
 * //   Wed Feb 11 1987 00:00:00,
 * //   Mon Jul 10 1989 00:00:00,
 * //   Sun Jul 02 1995 00:00:00
 * // ]
 */
function compareAsc (dirtyDateLeft, dirtyDateRight) {
  var dateLeft = parse_1(dirtyDateLeft);
  var timeLeft = dateLeft.getTime();
  var dateRight = parse_1(dirtyDateRight);
  var timeRight = dateRight.getTime();

  if (timeLeft < timeRight) {
    return -1
  } else if (timeLeft > timeRight) {
    return 1
  } else {
    return 0
  }
}

var compare_asc = compareAsc;

/**
 * @category Month Helpers
 * @summary Get the number of full months between the given dates.
 *
 * @description
 * Get the number of full months between the given dates.
 *
 * @param {Date|String|Number} dateLeft - the later date
 * @param {Date|String|Number} dateRight - the earlier date
 * @returns {Number} the number of full months
 *
 * @example
 * // How many full months are between 31 January 2014 and 1 September 2014?
 * var result = differenceInMonths(
 *   new Date(2014, 8, 1),
 *   new Date(2014, 0, 31)
 * )
 * //=> 7
 */
function differenceInMonths (dirtyDateLeft, dirtyDateRight) {
  var dateLeft = parse_1(dirtyDateLeft);
  var dateRight = parse_1(dirtyDateRight);

  var sign = compare_asc(dateLeft, dateRight);
  var difference = Math.abs(difference_in_calendar_months(dateLeft, dateRight));
  dateLeft.setMonth(dateLeft.getMonth() - sign * difference);

  // Math.abs(diff in full months - diff in calendar months) === 1 if last calendar month is not full
  // If so, result must be decreased by 1 in absolute value
  var isLastMonthNotFull = compare_asc(dateLeft, dateRight) === -sign;
  return sign * (difference - isLastMonthNotFull)
}

var difference_in_months = differenceInMonths;

var oneOf$2 = function oneOf(value, validList) {
    for (var i = 0; i < validList.length; i++) {
        if (value === validList[i]) {
            return true;
        }
    }
    return false;
};

var BkDate$1 = function () {
    function BkDate(flag, weekdays, time) {
        classCallCheck(this, BkDate);

        this.weekdays = weekdays;

        var dater = time ? new Date(time) : new Date();

        this.currentDay = {
            year: dater.getFullYear(),
            month: dater.getMonth() + 1,
            day: dater.getDate()
        };

        this.currentTime = {
            hour: dater.getHours(),
            minute: dater.getMinutes(),
            second: dater.getSeconds()
        };

        this.year = this.currentDay.year;
        this.month = this.currentDay.month;
        this.day = this.currentDay.day;

        this.setTimer = false;

        this.index = flag === 'start' ? 0 : 1;

        this.preYearDisabled = false;

        this.preMonthDisabled = false;

        this.nextYearDisabled = false;

        this.nextMonthDisabled = false;
    }

    createClass(BkDate, [{
        key: 'setDate',
        value: function setDate(date) {
            var dateItems = date.split('-');
            if (dateItems[0]) {
                this.year = parseInt(dateItems[0]);
            }
            if (dateItems[1]) {
                this.month = parseInt(dateItems[1]);
            }
            if (dateItems[2]) {
                this.day = parseInt(dateItems[2]);
            }
        }
    }, {
        key: 'formatDateString',
        value: function formatDateString(value) {
            return parseInt(value) < 10 ? 0 + '' + value : value;
        }
    }, {
        key: 'getFormatToday',
        value: function getFormatToday() {
            return this.currentDay.year + '-' + this.formatDateString(this.currentDay.month) + '-' + this.formatDateString(this.currentDay.day);
        }
    }, {
        key: 'getFormatDate',
        value: function getFormatDate() {
            return this.year + '-' + this.formatDateString(this.month) + '-' + this.formatDateString(this.day);
        }
    }, {
        key: 'getCurrentMouthDays',
        value: function getCurrentMouthDays() {
            return new Date(this.year, this.month, 0).getDate();
        }
    }, {
        key: 'getLastMouthDays',
        value: function getLastMouthDays() {
            return new Date(this.year, this.month - 1, 0).getDate();
        }
    }, {
        key: 'getCurrentMonthBeginWeek',
        value: function getCurrentMonthBeginWeek() {
            return new Date(this.year, this.month - 1, 1).getDay();
        }
    }]);
    return BkDate;
}();

var bkDateRange = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { directives: [{ name: "clickoutside", rawName: "v-clickoutside", value: _vm.close, expression: "close" }], staticClass: "bk-date-range", class: _vm.disabled ? 'disabled' : '', style: _vm.bkDateWidthObj, on: { "click": _vm.openDater } }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.selectedDateView, expression: "selectedDateView" }], attrs: { "type": "text", "name": "date-select", "readonly": "true", "disabled": _vm.disabled, "placeholder": _vm.defaultPlaceholder }, domProps: { "value": _vm.selectedDateView }, on: { "input": function input($event) {
                    if ($event.target.composing) {
                        return;
                    }_vm.selectedDateView = $event.target.value;
                } } }), _vm._v(" "), _c('transition', { attrs: { "name": _vm.transitionName } }, [_vm.showDatePanel ? _c('div', { class: ['date-dropdown-panel', 'daterange-dropdown-panel', { 'has-sidebar': _vm.quickSelect }], style: _vm.panelStyle }, [_c('date-picker', { ref: "startDater", staticClass: "start-date date-select-container fl", attrs: { "selected-range": _vm.selectedDateRange, "selected-range-tmp": _vm.selectedDateRangeTmp, "initDate": _vm.initStartDate, "bkDate": _vm.bkDateStart, "type": 'start', "dateLimit": _vm.startDateMin, "timer": _vm.timer }, on: { "date-quick-switch": _vm.dateQuickSwitch, "date-selected": _vm.triggerSelect } }), _vm._v(" "), _c('date-picker', { ref: "endDater", staticClass: "end-date date-select-container fl", attrs: { "selected-range": _vm.selectedDateRange, "selected-range-tmp": _vm.selectedDateRangeTmp, "initDate": _vm.initEndDate, "bkDate": _vm.bkDateEnd, "dateLimit": _vm.endDateMax, "type": 'end', "timer": _vm.timer }, on: { "date-quick-switch": _vm.dateQuickSwitch, "date-selected": _vm.triggerSelect } }), _vm._v(" "), _vm.quickSelect ? _c('div', { staticClass: "range-config fl" }, _vm._l(_vm.defaultRanges, function (range) {
            return _c('a', { class: { 'active': _vm.shouldBeMatched(range) }, attrs: { "href": "javascript:;" }, on: { "click": function click($event) {
                        $event.stopPropagation();_vm.changeRanges(range);
                    } } }, [_vm._v(_vm._s(range.text))]);
        })) : _vm._e(), _vm._v(" "), _vm.quickSelect ? _c('div', { staticClass: "range-action fl" }, [_c('a', { attrs: { "href": "javascript:;" }, on: { "click": function click($event) {
                    $event.stopPropagation();_vm.showDatePanel = false;
                } } }, [_vm._v(_vm._s(_vm.t('dateRange.ok')))]), _vm._v(" "), _c('a', { attrs: { "href": "javascript:;" }, on: { "click": function click($event) {
                    $event.stopPropagation();return _vm.clear($event);
                } } }, [_vm._v(_vm._s(_vm.t('dateRange.clear')))])]) : _vm._e()], 1) : _vm._e()])], 1);
    }, staticRenderFns: [],
    name: 'bk-date-range',
    mixins: [locale$1],
    components: {
        'date-picker': datepicker
    },
    props: {
        placeholder: {
            type: String,
            default: ''
        },
        disabled: {
            type: Boolean,
            default: false
        },
        align: {
            type: String,
            default: 'left'
        },
        quickSelect: {
            type: Boolean,
            default: true
        },
        rangeSeparator: {
            type: String,
            default: '-'
        },
        initDate: {
            type: String,
            default: ''
        },
        startDate: {
            type: String,
            default: ''
        },
        endDate: {
            type: String,
            default: ''
        },
        ranges: {
            type: Object,
            default: function _default() {}
        },
        timer: {
            type: Boolean,
            default: false
        },
        position: {
            validator: function validator(value) {
                return oneOf$2(value, ['top', 'bottom', 'left', 'right', 'top-left', 'top-right', 'bottom-left', 'bottom-right']);
            },

            default: 'bottom-right'
        },
        startDateMin: {
            type: String,
            default: ''
        },
        endDateMax: {
            type: String,
            default: ''
        }
    },
    data: function data() {
        var sdr = [format_1(new Date(), 'YYYY-MM-DD HH:mm:ss')];
        var sdrt = [format_1(sdr[0], 'YYYY-MM-DD')];

        var weekdays = [this.t('dateRange.datePicker.weekdays.sun'), this.t('dateRange.datePicker.weekdays.mon'), this.t('dateRange.datePicker.weekdays.tue'), this.t('dateRange.datePicker.weekdays.wed'), this.t('dateRange.datePicker.weekdays.thu'), this.t('dateRange.datePicker.weekdays.fri'), this.t('dateRange.datePicker.weekdays.sat')];
        var bkDateStart = this.startDate ? new BkDate$1('start', weekdays, this.startDate) : new BkDate$1('start', weekdays);
        var bkDateEnd = this.endDate ? new BkDate$1('end', weekdays, this.endDate) : new BkDate$1('end', weekdays);

        var initStartDate = this.startDate ? format_1(this.startDate, 'YYYY-MM-DD') : format_1(sub_months(new Date(), 1), 'YYYY-MM-DD');
        var initStartDateCopy = this.startDate ? format_1(this.startDate, 'YYYY-MM') : format_1(sub_months(new Date(), 1), 'YYYY-MM');
        var initEndDate = this.endDate ? format_1(this.endDate, 'YYYY-MM-DD') : format_1(new Date(), 'YYYY-MM-DD');
        var initEndDateCopy = this.endDate ? format_1(this.endDate, 'YYYY-MM') : format_1(new Date(), 'YYYY-MM');
        if (initStartDateCopy === initEndDateCopy) {
            initEndDate = format_1(add_months(initEndDate, 1), 'YYYY-MM-DD');
            bkDateStart.nextMonthDisabled = true;
            bkDateStart.nextYearDisabled = true;
            bkDateEnd.preMonthDisabled = true;
            bkDateEnd.preYearDisabled = true;
        } else {
            if (difference_in_months(initEndDate, initStartDate) > 12) {
                bkDateEnd.preMonthDisabled = false;
                bkDateEnd.preYearDisabled = false;
                bkDateStart.nextMonthDisabled = false;
                bkDateStart.nextYearDisabled = false;
            } else {
                bkDateEnd.preYearDisabled = true;
                bkDateStart.nextYearDisabled = true;
                if (difference_in_months(initEndDate, initStartDate) > 1) {
                    bkDateEnd.preMonthDisabled = false;
                    bkDateStart.nextMonthDisabled = false;
                } else {
                    bkDateEnd.preMonthDisabled = true;
                    bkDateStart.nextMonthDisabled = true;
                }
            }
        }

        bkDateStart.setDate(initStartDate);
        bkDateEnd.setDate(initEndDate);

        var transitionName = 'toggle-slide';
        var panelStyle = {};
        var positionArr = this.position.split('-');
        if (positionArr.indexOf('top') > -1) {
            panelStyle.bottom = '38px';
            transitionName = 'toggle-slide2';
        } else {
            panelStyle.top = '38px';
        }

        if (positionArr.indexOf('left') > -1) {
            panelStyle.right = 0;
        } else {
            panelStyle.left = 0;
        }

        return {
            panelStyle: panelStyle,
            transitionName: transitionName,
            bkDateWidthObj: this.timer ? { width: '350px' } : {},
            showDatePanel: false,
            selectedDateView: '',
            selectedRange: '',
            initStartDate: initStartDate,
            initEndDate: initEndDate,
            bkDateStart: bkDateStart,
            bkDateEnd: bkDateEnd,
            selectedDateRange: sdr,
            selectedDateRangeTmp: sdrt,
            defaultPlaceholder: this.t('dateRange.selectDate'),
            defaultRanges: [{
                text: this.t('dateRange.yestoday'),

                value: [sub_days(new Date(), 1), new Date()]
            }, {
                text: this.t('dateRange.lastweek'),

                value: [sub_days(new Date(), 7), new Date()]
            }, {
                text: this.t('dateRange.lastmonth'),

                value: [sub_months(new Date(), 1), new Date()]
            }, {
                text: this.t('dateRange.last3months'),

                value: [sub_months(new Date(), 3), new Date()]
            }]
        };
    },

    directives: {
        clickoutside: clickoutside
    },
    created: function created() {
        var _this = this;

        if (this.ranges && Object.keys(this.ranges).length) {
            var _defaultRanges;

            var defaultRanges = [];
            Object.keys(this.ranges).forEach(function (range) {
                defaultRanges.push({
                    text: range,
                    value: _this.ranges[range]
                });
            });
            (_defaultRanges = this.defaultRanges).splice.apply(_defaultRanges, [0, this.defaultRanges.length].concat(defaultRanges));
        }

        if (this.placeholder) {
            this.defaultPlaceholder = this.placeholder;
        }

        var hour = '';
        var minute = '';
        var second = '';

        if (this.startDate) {
            hour = get_hours(new Date(this.startDate));
            minute = get_minutes(new Date(this.startDate));
            second = get_seconds(new Date(this.startDate));
            this.selectedDateRange.unshift(this.timer ? format_1(set_hours(set_minutes(set_seconds(this.startDate, second), minute), hour), 'YYYY-MM-DD HH:mm:ss') : this.startDate);
            this.selectedDateRangeTmp.unshift(this.timer ? format_1(set_hours(set_minutes(set_seconds(this.startDate, second), minute), hour), 'YYYY-MM-DD') : this.startDate);
        }
        if (this.endDate) {
            hour = get_hours(new Date(this.endDate));
            minute = get_minutes(new Date(this.endDate));
            second = get_seconds(new Date(this.endDate));
            this.selectedDateRange.pop();
            this.selectedDateRange.push(this.timer ? format_1(set_hours(set_minutes(set_seconds(this.endDate, second), minute), hour), 'YYYY-MM-DD HH:mm:ss') : this.endDate);
            this.selectedDateRangeTmp.pop();
            this.selectedDateRangeTmp.push(this.timer ? format_1(set_hours(set_minutes(set_seconds(this.endDate, second), minute), hour), 'YYYY-MM-DD') : this.endDate);
        }
    },
    updated: function updated() {},

    watch: {
        showDatePanel: function showDatePanel(val) {
            if (!val) {
                this.$emit('close', this.selectedDateView);
            } else {
                this.$emit('show', this.selectedDateView);
            }
        },
        selectedDateRange: function selectedDateRange() {
            var seed = this.timer ? this.selectedDateRange : this.selectedDateRangeTmp;
            if (seed.length === 2) {
                var newSelectedDate = seed.join(' ' + this.rangeSeparator + ' ');

                if (this.selectedDateView !== newSelectedDate) {
                    this.$emit('change', this.selectedDateView, newSelectedDate);
                }
                this.selectedDateView = newSelectedDate;
                this.bkDateStart.setTimer = true;
                this.bkDateEnd.setTimer = true;
            } else {
                this.bkDateStart.setTimer = false;
                this.bkDateEnd.setTimer = false;
            }
        },
        selectedDateView: function selectedDateView() {
            var endDateTmp = '';
            var formatDateStart = format_1(this.selectedDateRange[0], 'YYYY-MM');
            var formatDateEnd = format_1(this.selectedDateRange[1], 'YYYY-MM');

            this.initStartDate = this.selectedDateView.split(' ' + this.rangeSeparator + ' ')[0];

            if (is_same_year(formatDateStart, formatDateEnd) && is_same_month(formatDateStart, formatDateEnd)) {
                this.initEndDate = format_1(add_months(this.selectedDateRange[0], 1), this.timer ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD');
                endDateTmp = format_1(this.selectedDateRange[1], this.timer ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD') || '';
            } else {
                this.initEndDate = this.selectedDateView.split(' ' + this.rangeSeparator + ' ')[1];
                endDateTmp = this.initEndDate || '';
            }
            this.$emit('update:startDate', this.initStartDate);
            this.$emit('update:endDate', endDateTmp);
        },
        startDate: function startDate(value) {
            if (!value || !this.endDate) {
                this.clear();
            } else {
                this.initDateData(value, this.endDate);
                this.initStartDate = value;
                this.handlerDate('start');
            }
        },
        endDate: function endDate(value) {
            if (!value || !this.startDate) {
                this.clear();
            } else {
                this.initDateData(this.startDate, value);
                this.initEndDate = value;
                this.handlerDate('end');
            }
        }
    },
    methods: {
        handlerDate: function handlerDate(type) {
            var sdr = [format_1(new Date(), 'YYYY-MM-DD HH:mm:ss')];
            var sdrt = [format_1(sdr[0], 'YYYY-MM-DD')];
            var weekdays = [this.t('dateRange.datePicker.weekdays.sun'), this.t('dateRange.datePicker.weekdays.mon'), this.t('dateRange.datePicker.weekdays.tue'), this.t('dateRange.datePicker.weekdays.wed'), this.t('dateRange.datePicker.weekdays.thu'), this.t('dateRange.datePicker.weekdays.fri'), this.t('dateRange.datePicker.weekdays.sat')];
            var bkDateStart = this.startDate ? new BkDate$1('start', weekdays, this.startDate) : new BkDate$1('start', weekdays);
            var bkDateEnd = this.endDate ? new BkDate$1('end', weekdays, this.endDate) : new BkDate$1('end', weekdays);
            var initStartDate = this.startDate ? format_1(this.startDate, 'YYYY-MM-DD') : format_1(sub_months(new Date(), 1), 'YYYY-MM-DD');
            var initStartDateCopy = this.startDate ? format_1(this.startDate, 'YYYY-MM') : format_1(sub_months(new Date(), 1), 'YYYY-MM');
            var initEndDate = this.endDate ? format_1(this.endDate, 'YYYY-MM-DD') : format_1(new Date(), 'YYYY-MM-DD');
            var initEndDateCopy = this.endDate ? format_1(this.endDate, 'YYYY-MM') : format_1(new Date(), 'YYYY-MM');
            if (initStartDateCopy === initEndDateCopy) {
                initEndDate = format_1(add_months(initEndDate, 1), 'YYYY-MM-DD');
                bkDateStart.nextMonthDisabled = true;
                bkDateStart.nextYearDisabled = true;
                bkDateEnd.preMonthDisabled = true;
                bkDateEnd.preYearDisabled = true;
            } else {
                if (difference_in_months(initEndDate, initStartDate) > 12) {
                    bkDateEnd.preMonthDisabled = false;
                    bkDateEnd.preYearDisabled = false;
                    bkDateStart.nextMonthDisabled = false;
                    bkDateStart.nextYearDisabled = false;
                } else {
                    bkDateEnd.preYearDisabled = true;
                    bkDateStart.nextYearDisabled = true;
                    if (difference_in_months(initEndDate, initStartDate) > 1) {
                        bkDateEnd.preMonthDisabled = false;
                        bkDateStart.nextMonthDisabled = false;
                    } else {
                        bkDateEnd.preMonthDisabled = true;
                        bkDateStart.nextMonthDisabled = true;
                    }
                }
            }
            bkDateStart.setDate(initStartDate);
            bkDateEnd.setDate(initEndDate);
            if (type === 'start') {
                this.bkDateStart = bkDateStart;
            }
            if (type === 'end') {
                this.bkDateEnd = bkDateEnd;
            }
        },
        initDateData: function initDateData(startDate, endDate) {
            var _this2 = this;

            if (this.ranges && Object.keys(this.ranges).length) {
                var _defaultRanges2;

                var defaultRanges = [];
                Object.keys(this.ranges).forEach(function (range) {
                    defaultRanges.push({
                        text: range,
                        value: _this2.ranges[range]
                    });
                });
                (_defaultRanges2 = this.defaultRanges).splice.apply(_defaultRanges2, [0, this.defaultRanges.length].concat(defaultRanges));
            }

            if (this.placeholder) {
                this.defaultPlaceholder = this.placeholder;
            }
            var hour = '';
            var minute = '';
            var second = '';
            hour = get_hours(new Date(startDate));
            minute = get_minutes(new Date(startDate));
            second = get_seconds(new Date(startDate));
            this.selectedDateRange.shift();
            this.selectedDateRangeTmp.shift();
            this.selectedDateRange.unshift(this.timer ? format_1(set_hours(set_minutes(set_seconds(startDate, second), minute), hour), 'YYYY-MM-DD HH:mm:ss') : startDate);
            this.selectedDateRangeTmp.unshift(this.timer ? format_1(set_hours(set_minutes(set_seconds(startDate, second), minute), hour), 'YYYY-MM-DD') : startDate);
            hour = get_hours(new Date(endDate));
            minute = get_minutes(new Date(endDate));
            second = get_seconds(new Date(endDate));
            this.selectedDateRange.pop();
            this.selectedDateRangeTmp.pop();
            this.selectedDateRange.push(this.timer ? format_1(set_hours(set_minutes(set_seconds(endDate, second), minute), hour), 'YYYY-MM-DD HH:mm:ss') : endDate);
            this.selectedDateRangeTmp.push(this.timer ? format_1(set_hours(set_minutes(set_seconds(endDate, second), minute), hour), 'YYYY-MM-DD') : endDate);
        },
        dateQuickSwitch: function dateQuickSwitch(date) {
            var startDateInfo = this.$refs.startDater;
            var startTopDate = startDateInfo.topBarFormatView.value;

            var endDateInfo = this.$refs.endDater;
            var endTopDate = endDateInfo.topBarFormatView.value;

            if (startTopDate === endTopDate) {
                switch (date.type) {
                    case 'next':
                        endDateInfo.BkDate.setDate(format_1(add_months(endTopDate, 1), 'YYYY-MM'));
                        break;
                    case 'last':
                        startDateInfo.BkDate.setDate(format_1(add_months(startTopDate, -1), 'YYYY-MM'));
                        break;
                    default:
                        break;
                }
            }

            if (difference_in_months(endTopDate, startTopDate) > 12) {
                this.bkDateEnd.preMonthDisabled = false;
                this.bkDateEnd.preYearDisabled = false;
                this.bkDateStart.nextMonthDisabled = false;
                this.bkDateStart.nextYearDisabled = false;
            } else {
                this.bkDateEnd.preYearDisabled = true;
                this.bkDateStart.nextYearDisabled = true;
                if (difference_in_months(endTopDate, startTopDate) > 1) {
                    this.bkDateEnd.preMonthDisabled = false;
                    this.bkDateStart.nextMonthDisabled = false;
                } else {
                    this.bkDateEnd.preMonthDisabled = true;
                    this.bkDateStart.nextMonthDisabled = true;
                }
            }
        },
        changeRanges: function changeRanges(range) {
            var rangeStartDate = format_1(range.value[0], 'YYYY-MM-DD HH:mm:ss');
            var rangeEndDate = format_1(range.value[1], 'YYYY-MM-DD HH:mm:ss');
            var rangeStartDateTmp = format_1(range.value[0], 'YYYY-MM-DD');
            var rangeEndDateTmp = format_1(range.value[1], 'YYYY-MM-DD');

            if (rangeStartDateTmp === this.selectedDateRangeTmp[0] && rangeEndDateTmp === this.selectedDateRangeTmp[1]) {
                return;
            }
            this.selectedDateRange = [rangeStartDate, rangeEndDate];
            this.selectedDateRangeTmp = [rangeStartDateTmp, rangeEndDateTmp];
            this.initStartDate = this.timer ? rangeStartDate : rangeStartDateTmp;
            this.initEndDate = this.timer ? rangeEndDate : rangeEndDateTmp;
            this.selectedRange = range.text;

            var formatDateStart = format_1(rangeStartDate, 'YYYY-MM');
            var formatDateEnd = format_1(rangeEndDate, 'YYYY-MM');

            if (is_same_year(formatDateStart, formatDateEnd) && is_same_month(formatDateStart, formatDateEnd)) {
                this.initEndDate = format_1(add_months(this.selectedDateRange[0], 1), this.timer ? 'YYYY-MM-DD HH:mm:ss' : 'YYYY-MM-DD');
            }

            this.bkDateStart.setDate(this.initStartDate);
            this.bkDateEnd.setDate(this.initEndDate);
        },
        triggerSelect: function triggerSelect(date, bkDateIndex) {
            if (bkDateIndex !== undefined) {
                var hour = get_hours(date);
                var minute = get_minutes(date);
                var second = get_seconds(date);
                this.selectedDateRange.splice(bkDateIndex, 1, format_1(set_hours(set_minutes(set_seconds(this.selectedDateRange[bkDateIndex], second), minute), hour), 'YYYY-MM-DD HH:mm:ss'));
            } else {
                var selectedLen = this.selectedDateRange.length;

                date = format_1(date, 'YYYY-MM-DD HH:mm:ss');
                var dateTmp = format_1(date, 'YYYY-MM-DD');

                switch (selectedLen) {
                    case 0:
                        this.selectedDateRange.push(date);
                        this.selectedDateRangeTmp.push(dateTmp);
                        break;
                    case 1:
                        if (is_same_year(date, this.selectedDateRange[0]) && is_same_month(date, this.selectedDateRange[0]) && is_same_day(date, this.selectedDateRange[0]) && is_same_hour(date, this.selectedDateRange[0]) && is_same_minute(date, this.selectedDateRange[0]) && is_same_second(date, this.selectedDateRange[0]) || is_after(date, this.selectedDateRange[0])) {
                            this.selectedDateRange.push(date);
                            this.selectedDateRangeTmp.push(dateTmp);
                        }

                        if (is_before(date, this.selectedDateRange[0])) {
                            this.selectedDateRange = [date];
                            this.selectedDateRangeTmp = [dateTmp];
                        }
                        break;
                    case 2:
                        this.selectedDateRange = [date];
                        this.selectedDateRangeTmp = [dateTmp];
                        break;
                    default:
                }
            }
        },
        shouldBeMatched: function shouldBeMatched(range) {
            var isMatched = this.selectedDateRangeTmp[0] === format_1(range.value[0], 'YYYY-MM-DD') && this.selectedDateRangeTmp[1] === format_1(range.value[1], 'YYYY-MM-DD');
            return isMatched;
        },
        openDater: function openDater() {
            if (this.disabled) {
                return;
            }
            this.showDatePanel = true;
        },
        close: function close() {
            if (this.selectedDateView && this.selectedDateRange.length === 0) {
                this.selectedDateRange = this.selectedDateView.split(' ' + this.rangeSeparator + ' ');
            }
            this.showDatePanel = false;
        },
        clear: function clear() {
            this.$emit('change', this.selectedDateView, '');

            this.selectedDateView = '';
            this.selectedDateRange = [];
            this.selectedDateRangeTmp = [];

            var date = format_1(new Date(), 'YYYY-MM-DD HH:mm:ss');
            var dateTmp = format_1(date, 'YYYY-MM-DD');
            this.selectedDateRange.push(date);
            this.selectedDateRangeTmp.push(dateTmp);

            this.initStartDate = this.startDate || format_1(sub_months(new Date(), 1), 'YYYY-MM-DD');
            this.initEndDate = this.endDate || format_1(new Date(), 'YYYY-MM-DD');
            this.bkDateStart.setDate(this.initStartDate);
            this.bkDateEnd.setDate(this.initEndDate);

            this.bkDateStart.currentTime = this.bkDateEnd.currentTime = {
                hour: get_hours(new Date()),
                minute: get_minutes(new Date()),
                second: get_seconds(new Date())
            };

            this.showDatePanel = false;
        }
    }
};

bkDateRange.install = function (Vue$$1) {
  Vue$$1.component(bkDateRange.name, bkDateRange);
};

var bkSelector = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { directives: [{ name: "clickoutside", rawName: "v-clickoutside", value: _vm.close, expression: "close" }], staticClass: "bk-selector", class: [_vm.extCls, { 'open': _vm.open }], on: { "click": _vm.openFn, "keyup": function keyup($event) {
                    if (!('button' in $event) && _vm._k($event.keyCode, "enter", 13, $event.key, "Enter")) {
                        return null;
                    }return _vm.updateSelect($event);
                } } }, [_c('div', { staticClass: "bk-selector-wrapper" }, [_c('input', { staticClass: "bk-selector-input", class: { placeholder: _vm.selectedText === _vm.defaultPlaceholder, active: _vm.open }, attrs: { "readonly": "readonly", "placeholder": _vm.defaultPlaceholder, "disabled": _vm.disabled }, domProps: { "value": _vm.selectedText }, on: { "mouseover": _vm.showClearFn, "mouseleave": function mouseleave($event) {
                    _vm.showClear = false;
                } } }), _vm._v(" "), _c('i', { directives: [{ name: "show", rawName: "v-show", value: !_vm.isLoading && !_vm.showClear, expression: "!isLoading && !showClear" }], class: ['bk-icon icon-angle-down bk-selector-icon', { 'disabled': _vm.disabled }] }), _vm._v(" "), _c('i', { directives: [{ name: "show", rawName: "v-show", value: !_vm.isLoading && _vm.showClear, expression: "!isLoading && showClear" }], class: ['bk-icon icon-close bk-selector-icon clear-icon'], on: { "mouseover": _vm.showClearFn, "mouseleave": function mouseleave($event) {
                    _vm.showClear = false;
                }, "click": function click($event) {
                    _vm.clearSelected($event);
                } } }), _vm._v(" "), _c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.isLoading, expression: "isLoading" }], staticClass: "bk-spin-loading bk-spin-loading-mini bk-spin-loading-primary selector-loading-icon" }, [_c('div', { staticClass: "rotate rotate1" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate2" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate3" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate4" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate5" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate6" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate7" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate8" })])]), _vm._v(" "), _c('transition', { attrs: { "name": _vm.listSlideName } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: !_vm.isLoading && _vm.open, expression: "!isLoading && open" }], staticClass: "bk-selector-list", style: _vm.panelStyle }, [_vm.searchable ? _c('div', { staticClass: "bk-selector-search-item", on: { "click": function click($event) {
                    $event.stopPropagation();
                } } }, [_c('i', { staticClass: "bk-icon icon-search" }), _vm._v(" "), _c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.condition, expression: "condition" }], ref: "searchNode", attrs: { "type": "text", "placeholder": _vm.searchPlaceholder }, domProps: { "value": _vm.condition }, on: { "input": [function ($event) {
                    if ($event.target.composing) {
                        return;
                    }_vm.condition = $event.target.value;
                }, _vm.inputFn] } })]) : _vm._e(), _vm._v(" "), _c('ul', { staticClass: "outside-ul", style: { 'max-height': _vm.contentMaxHeight + 'px' } }, [_vm._l(_vm.localList, function (item, i) {
            return _vm.localList.length !== 0 ? _c('li', { key: i, class: ['bk-selector-list-item', item.children && item.children.length ? 'bk-selector-group-list-item' : '', { 'active': _vm.activeIndex === i }] }, [item.children && item.children.length && _vm.hasChildren ? [_c('div', { staticClass: "bk-selector-group-name" }, [_vm._v(_vm._s(item[_vm.displayKey]))]), _vm._v(" "), _c('ul', { staticClass: "bk-selector-group-list" }, _vm._l(item.children, function (child, index) {
                return _c('li', { staticClass: "bk-selector-list-item", class: [{ 'active': _vm.itemIndex === index && _vm.groupIndex === i && _vm.hasChildren }] }, [_c('div', { staticClass: "bk-selector-node bk-selector-sub-node", class: { 'bk-selector-selected': !_vm.multiSelect && child[_vm.settingKey] === _vm.selected } }, [_c('div', { staticClass: "text", on: { "click": function click($event) {
                            $event.stopPropagation();_vm.selectItem(child, $event);
                        } } }, [_vm.multiSelect ? _c('label', { staticClass: "bk-form-checkbox bk-checkbox-small mr0 bk-selector-multi-label" }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.localSelected, expression: "localSelected" }], attrs: { "type": "checkbox", "name": 'multiSelect' + +new Date() }, domProps: { "value": child[_vm.settingKey], "checked": Array.isArray(_vm.localSelected) ? _vm._i(_vm.localSelected, child[_vm.settingKey]) > -1 : _vm.localSelected }, on: { "change": function change($event) {
                            var $$a = _vm.localSelected,
                                $$el = $event.target,
                                $$c = $$el.checked ? true : false;if (Array.isArray($$a)) {
                                var $$v = child[_vm.settingKey],
                                    $$i = _vm._i($$a, $$v);if ($$el.checked) {
                                    $$i < 0 && (_vm.localSelected = $$a.concat([$$v]));
                                } else {
                                    $$i > -1 && (_vm.localSelected = $$a.slice(0, $$i).concat($$a.slice($$i + 1)));
                                }
                            } else {
                                _vm.localSelected = $$c;
                            }
                        } } }), _vm._v(" " + _vm._s(child[_vm.displayKey]) + " ")]) : [_vm._v(" " + _vm._s(child[_vm.displayKey]) + " ")]], 2), _vm._v(" "), _vm.tools !== false ? _c('div', { staticClass: "bk-selector-tools" }, [_vm.tools.edit !== false ? _c('i', { staticClass: "bk-icon icon-edit2 bk-selector-list-icon", on: { "click": function click($event) {
                            $event.stopPropagation();_vm.editFn(index);
                        } } }) : _vm._e(), _vm._v(" "), _vm.tools.del !== false ? _c('i', { staticClass: "bk-icon icon-close bk-selector-list-icon", on: { "click": function click($event) {
                            $event.stopPropagation();_vm.delFn(index);
                        } } }) : _vm._e()]) : _vm._e()])]);
            }))] : [_c('div', { staticClass: "bk-selector-node", class: { 'bk-selector-selected': !_vm.multiSelect && item[_vm.settingKey] === _vm.selected, 'is-disabled': item.isDisabled } }, [_c('div', { staticClass: "text", attrs: { "title": item[_vm.displayKey] }, on: { "click": function click($event) {
                        $event.stopPropagation();_vm.selectItem(item, $event);
                    } } }, [_vm.multiSelect ? _c('label', { staticClass: "bk-form-checkbox bk-checkbox-small mr0 bk-selector-multi-label" }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.localSelected, expression: "localSelected" }], attrs: { "type": "checkbox", "name": 'multiSelect' + +new Date() }, domProps: { "value": item[_vm.settingKey], "checked": Array.isArray(_vm.localSelected) ? _vm._i(_vm.localSelected, item[_vm.settingKey]) > -1 : _vm.localSelected }, on: { "change": function change($event) {
                        var $$a = _vm.localSelected,
                            $$el = $event.target,
                            $$c = $$el.checked ? true : false;if (Array.isArray($$a)) {
                            var $$v = item[_vm.settingKey],
                                $$i = _vm._i($$a, $$v);if ($$el.checked) {
                                $$i < 0 && (_vm.localSelected = $$a.concat([$$v]));
                            } else {
                                $$i > -1 && (_vm.localSelected = $$a.slice(0, $$i).concat($$a.slice($$i + 1)));
                            }
                        } else {
                            _vm.localSelected = $$c;
                        }
                    } } }), _vm._v(" " + _vm._s(item[_vm.displayKey]) + " ")]) : [_vm._v(" " + _vm._s(item[_vm.displayKey]) + " ")]], 2), _vm._v(" "), _vm.tools !== false ? _c('div', { staticClass: "bk-selector-tools" }, [_vm.tools.edit !== false ? _c('i', { staticClass: "bk-icon icon-edit2 bk-selector-list-icon", on: { "click": function click($event) {
                        $event.stopPropagation();_vm.editFn(_vm.index);
                    } } }) : _vm._e(), _vm._v(" "), _vm.tools.del !== false ? _c('i', { staticClass: "bk-icon icon-close bk-selector-list-icon", on: { "click": function click($event) {
                        $event.stopPropagation();_vm.delFn(_vm.index);
                    } } }) : _vm._e()]) : _vm._e()])]], 2) : _vm._e();
        }), _vm._v(" "), !_vm.isLoading && _vm.localList.length === 0 ? _c('li', { staticClass: "bk-selector-list-item" }, [_c('div', { staticClass: "text no-search-result" }, [_vm._v(" " + _vm._s(_vm.list.length ? _vm.defaultSearchEmptyText : _vm.defaultEmptyText) + " ")])]) : _vm._e()], 2), _vm._v(" "), _vm._t("default")], 2)])], 1);
    }, staticRenderFns: [],
    name: 'bk-selector',
    mixins: [locale$1],
    directives: {
        clickoutside: clickoutside
    },
    props: {
        extCls: {
            type: String
        },
        searchPlaceholder: {
            type: String
        },
        isLoading: {
            type: Boolean,
            default: false
        },
        hasCreateItem: {
            type: Boolean,
            default: false
        },
        hasChildren: {
            type: [Boolean, String],
            default: false
        },
        tools: {
            type: [Object, Boolean],
            default: false
        },
        list: {
            type: Array,
            required: true
        },
        filterList: {
            type: Array,
            default: function _default() {
                return [];
            }
        },
        selected: {
            type: [Number, Array, String],
            required: true
        },
        placeholder: {
            type: [String, Boolean],
            default: ''
        },

        isLink: {
            type: [String, Boolean],
            default: false
        },
        displayKey: {
            type: String,
            default: 'name'
        },
        disabled: {
            type: [String, Boolean, Number],
            default: false
        },
        multiSelect: {
            type: Boolean,
            default: false
        },
        searchable: {
            type: Boolean,
            default: false
        },
        searchKey: {
            type: String,
            default: 'name'
        },
        allowClear: {
            type: Boolean,
            default: false
        },
        settingKey: {
            type: String,
            default: 'id'
        },
        initPreventTrigger: {
            type: Boolean,
            default: false
        },
        emptyText: {
            type: String,
            default: ''
        },
        searchEmptyText: {
            type: String,
            default: ''
        },
        contentMaxHeight: {
            type: Number,
            default: 300
        }
    },
    data: function data() {
        return {
            open: false,
            selectedList: this.calcSelected(this.selected),
            condition: '',

            localSelected: this.selected,

            showClear: false,
            activeIndex: -1,
            groupIndex: 0,
            itemIndex: -1,
            isKeydown: false,
            listInterval: null,
            panelStyle: {},
            listSlideName: 'toggle-slide',
            defaultPlaceholder: this.t('selector.pleaseselect'),
            defaultEmptyText: this.t('selector.emptyText'),
            defaultSearchEmptyText: this.t('selector.searchEmptyText')
        };
    },

    computed: {
        localList: function localList() {
            var _this = this;

            if (!this.multiSelect) {
                this.list.forEach(function (item) {
                    if (_this.filterList.includes(item[_this.settingKey])) {
                        item.isDisabled = true;
                    } else {
                        item.isDisabled = false;
                    }
                });
            }
            if (this.searchable && this.condition) {
                var arr = [];
                var key = this.searchKey;
                var len = this.list.length;
                for (var i = 0; i < len; i++) {
                    var item = this.list[i];
                    if (item.children) {
                        var results = [];
                        var childLen = item.children.length;
                        for (var j = 0; j < childLen; j++) {
                            var child = item.children[j];
                            if (child[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                results.push(child);
                            }
                        }
                        if (results.length) {
                            var cloneItem = Object.assign({}, item);
                            cloneItem.children = results;
                            arr.push(cloneItem);
                        }
                    } else {
                        if (item[key].toLowerCase().includes(this.condition.toLowerCase())) {
                            arr.push(item);
                        }
                    }
                }

                return arr;
            }
            return this.list;
        },
        currentItem: function currentItem() {
            return this.list[this.localSelected];
        },
        selectedText: function selectedText() {
            var _this2 = this;

            var text = this.defaultPlaceholder;
            var textArr = [];
            if (Array.isArray(this.selectedList) && this.selectedList.length) {
                this.selectedList.forEach(function (item) {
                    textArr.push(item[_this2.displayKey]);
                });
            } else if (this.selectedList) {
                this.selectedList[this.displayKey] && textArr.push(this.selectedList[this.displayKey]);
            }

            return textArr.length ? textArr.join(',') : this.defaultPlaceholder;
        }
    },
    watch: {
        selected: function selected(newVal) {
            if (this.list.length) {
                this.selectedList = this.calcSelected(this.selected, this.isLink);
            }
            this.localSelected = this.selected;
        },
        list: function list(newVal) {
            if (this.selected) {
                this.selectedList = this.calcSelected(this.selected, this.isLink);
            } else {
                this.selectedList = [];
            }
        },
        localSelected: function localSelected(val) {
            if (this.list.length) {
                this.selectedList = this.calcSelected(this.localSelected, this.isLink);
            }
        },
        open: function open(newVal) {
            var _this3 = this;

            var searchNode = this.$refs.searchNode;
            if (searchNode) {
                if (newVal) {
                    this.$nextTick(function () {
                        searchNode.focus();
                    });
                }
            }
            this.$emit('visible-toggle', newVal);

            if (newVal) {
                window.onkeydown = function (e) {
                    switch (e.keyCode) {
                        case 38:
                            _this3.listUp();
                            break;

                        case 40:
                            _this3.listDown();
                            break;

                        default:
                            break;
                    }
                };

                document.onkeydown = function (e) {
                    e = e || event;
                    if (e.keyCode === 38) {
                        return false;
                    }
                    if (e.keyCode === 40) {
                        return false;
                    }
                };
            } else {
                window.onkeydown = null;
                document.onkeydown = null;
            }
        },
        placeholder: function placeholder(value) {
            this.defaultPlaceholder = value || this.t('selector.pleaseselect');
        }
    },
    created: function created() {
        if (this.placeholder) {
            this.defaultPlaceholder = this.placeholder;
        }
        if (this.emptyText) {
            this.defaultEmptyText = this.emptyText;
        }
        if (this.searchEmptyText) {
            this.defaultSearchEmptyText = this.searchEmptyText;
        }
    },
    mounted: function mounted() {
        this.popup = this.$el;
        if (this.isLink) {
            if (this.list.length && this.selected) {
                this.calcSelected(this.selected, this.isLink);
            }
        }
    },

    methods: {
        listUp: function listUp() {
            var maxIndex = 0;
            if (this.hasChildren) {
                var arr = [];
                this.localList.forEach(function (list, index) {
                    list.children.forEach(function (l, i) {
                        arr.push(l);
                    });
                });

                if (this.groupIndex > 0) {
                    this.itemIndex--;
                    if (this.itemIndex < 0) {
                        this.itemIndex = this.localList[this.groupIndex - 1].children.length - 1;
                        this.groupIndex--;
                    }
                } else {
                    this.groupIndex = this.localList.length - 1;
                    this.itemIndex = this.localList[this.groupIndex].children.length - 1;
                }
            } else {
                maxIndex = this.localList.length - 1;
                if (this.activeIndex > 0) {
                    this.activeIndex--;
                } else {
                    this.activeIndex = maxIndex;
                }
            }

            this.updateScrollTop();
        },
        listDown: function listDown() {
            var maxIndex = 0;
            if (this.hasChildren) {
                var arr = [];
                this.localList.forEach(function (list, index) {
                    list.children.forEach(function (l, i) {
                        arr.push(l);
                    });
                });

                if (this.groupIndex < this.localList.length) {
                    this.itemIndex++;
                    if (this.itemIndex > this.localList[this.groupIndex].children.length - 1) {
                        this.groupIndex++;
                        this.itemIndex = 0;
                    }
                    if (this.groupIndex === this.localList.length) {
                        this.groupIndex = 0;
                        this.itemIndex = 0;
                    }
                } else {
                    this.groupIndex = 0;
                    this.itemIndex = 0;
                }
            } else {
                maxIndex = this.localList.length - 1;
                if (this.activeIndex < maxIndex) {
                    this.activeIndex++;
                } else {
                    this.activeIndex = 0;
                }
            }

            this.updateScrollTop();
        },
        updateScrollTop: function updateScrollTop(type) {
            var _this4 = this;

            var panelObj = this.$el.querySelector('.bk-selector-list .outside-ul');
            var panelInfo = {
                height: panelObj.clientHeight,
                yAxios: panelObj.getBoundingClientRect().y
            };
            this.$nextTick(function () {
                var activeObj = _this4.$el.querySelector('.bk-selector-list .active');
                var activeInfo = {
                    height: activeObj.clientHeight,
                    yAxios: activeObj.getBoundingClientRect().y
                };

                if (activeInfo.yAxios < panelInfo.yAxios) {
                    var currentScTop = panelObj.scrollTop;
                    panelObj.scrollTop = currentScTop - (panelInfo.yAxios - activeInfo.yAxios);
                }

                var distanceToBottom = activeInfo.yAxios + activeInfo.height - panelInfo.yAxios;

                if (distanceToBottom > panelInfo.height) {
                    var _currentScTop = panelObj.scrollTop;
                    panelObj.scrollTop = _currentScTop + distanceToBottom - panelInfo.height;
                }
            });
        },
        updateSelect: function updateSelect() {
            var _this5 = this;

            var activeItem = void 0;
            if (!this.hasChildren) {
                activeItem = this.localList[this.activeIndex];
            } else {
                activeItem = this.localList[this.groupIndex].children[this.itemIndex];
            }
            if (this.multiSelect) {
                var isAdded = this.selectedList.some(function (item) {
                    return item[_this5.settingKey] === activeItem[_this5.settingKey];
                });
                if (isAdded) {
                    this.selectedList = this.selectedList.filter(function (item) {
                        return item[_this5.settingKey] !== activeItem[_this5.settingKey];
                    });
                    this.localSelected = this.localSelected.filter(function (item) {
                        return item !== activeItem[_this5.settingKey];
                    });
                    this.$emit('update:selected', this.localSelected);
                } else {
                    this.selectedList.push(activeItem);
                    this.localSelected.push(activeItem[this.settingKey]);
                    this.localSelected.sort();
                }
                if (Array.isArray(this.localSelected)) {
                    var _data = [];
                    var _iteratorNormalCompletion = true;
                    var _didIteratorError = false;
                    var _iteratorError = undefined;

                    try {
                        for (var _iterator = this.localSelected[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                            var key = _step.value;

                            var params = this.getItem(key);
                            if (params) {
                                _data.push(params);
                            }
                        }
                    } catch (err) {
                        _didIteratorError = true;
                        _iteratorError = err;
                    } finally {
                        try {
                            if (!_iteratorNormalCompletion && _iterator.return) {
                                _iterator.return();
                            }
                        } finally {
                            if (_didIteratorError) {
                                throw _iteratorError;
                            }
                        }
                    }

                    this.$emit('item-selected', this.localSelected, _data);
                }
                return;
            }
            this.$el.querySelector('.bk-selector-list .active').click();
            this.selectedList = this.localList[this.activeIndex];
            this.$emit('update:selected', this.selectedList[this.settingKey]);
            var data = {};
            this.list.forEach(function (item) {
                if (item[_this5.settingKey] === _this5.selectedList[_this5.settingKey]) {
                    data = item;
                }
            });

            this.$emit('item-selected', this.selectedList[this.settingKey], data);
        },
        getItem: function getItem(key) {
            var _this6 = this;

            var data = null;

            this.list.forEach(function (item) {
                if (!item.children) {
                    if (String(item[_this6.settingKey]) === String(key)) {
                        data = item;
                    }
                } else {
                    var children = item.children;
                    children.forEach(function (child) {
                        if (String(child[_this6.settingKey]) === String(key)) {
                            data = child;
                        }
                    });
                }
            });
            return data;
        },
        calcSelected: function calcSelected(selected, isTrigger) {
            var data = null;

            if (Array.isArray(selected)) {
                data = [];
                var len = selected.length;
                for (var i = 0; i < len; i++) {
                    var item = this.getItem(selected[i]);
                    if (item) {
                        data.push(item);
                    }
                }

                if (data.length && isTrigger && !this.initPreventTrigger) {
                    this.$emit('item-selected', selected, data, isTrigger);
                }
            } else if (selected !== undefined) {
                var _item = this.getItem(selected);
                if (_item) {
                    data = _item;
                }
                if (data && isTrigger && !this.initPreventTrigger) {
                    this.$emit('item-selected', selected, data, isTrigger);
                }
            }
            return data;
        },
        close: function close() {
            this.open = false;
        },
        initSelectorPosition: function initSelectorPosition(currentTarget) {
            if (currentTarget) {
                var distanceLeft = getActualLeft(currentTarget);
                var distanceTop = getActualTop(currentTarget);
                var winWidth = document.body.clientWidth;
                var winHeight = document.body.clientHeight;
                var ySet = {};
                var listHeight = this.list.length * 42;
                if (listHeight > 160) {
                    listHeight = 160;
                }
                var scrollTop = document.documentElement.scrollTop || document.body.scrollTop;

                if (distanceTop + listHeight + 42 - scrollTop < winHeight) {
                    ySet = {
                        top: '40px',
                        bottom: 'auto'
                    };

                    this.listSlideName = 'toggle-slide';
                } else {
                    ySet = {
                        top: 'auto',
                        bottom: '40px'
                    };

                    this.listSlideName = 'toggle-slide2';
                }

                this.panelStyle = _extends({}, ySet);
            }
        },
        openFn: function openFn(event) {
            if (this.disabled) {
                return;
            }

            if (!this.disabled) {
                if (!this.open && event) {
                    this.initSelectorPosition(event.currentTarget);
                }
                this.open = !this.open;
                this.$emit('visible-toggle', this.open);
            }
        },
        calcList: function calcList() {
            if (this.searchable) {
                var arr = [];
                var key = this.searchKey;

                var len = this.list.length;
                for (var i = 0; i < len; i++) {
                    var item = this.list[i];
                    if (item.children) {
                        var results = [];
                        var childLen = item.children.length;
                        for (var j = 0; j < childLen; j++) {
                            var child = item.children[j];
                            if (child[key].toLowerCase().includes(this.condition.toLowerCase())) {
                                results.push(child);
                            }
                        }
                        if (results.length) {
                            var cloneItem = Object.assign({}, item);
                            cloneItem.children = results;
                            arr.push(cloneItem);
                        }
                    } else {
                        if (item[key].toLowerCase().includes(this.condition.toLowerCase())) {
                            arr.push(item);
                        }
                    }
                }

                this.localList = arr;
            } else {
                this.localList = this.list;
            }
        },
        showClearFn: function showClearFn() {
            if (this.allowClear && this.localSelected !== -1 && this.localSelected !== '') {
                this.showClear = true;
            }
        },
        clearSelected: function clearSelected(e) {
            this.$emit('clear', this.localSelected);
            this.localSelected = -1;
            this.showClear = false;
            e.stopPropagation();
            if (this.multiSelect) {
                this.$emit('update:selected', []);
            } else {
                this.$emit('update:selected', '');
            }
        },
        selectItem: function selectItem(data, event) {
            var _this7 = this;

            if (data.isDisabled) {
                return;
            }
            setTimeout(function () {
                _this7.toggleSelect(data, event);
            }, 10);
        },
        toggleSelect: function toggleSelect(data, event) {
            var $selected = void 0;
            var $selectedList = void 0;
            var settingKey = this.settingKey;
            var isMultiSelect = this.multiSelect;
            var list = this.localList;
            var index = data && data[settingKey] !== undefined ? data[settingKey] : undefined;

            if (isMultiSelect && event.target.tagName.toLowerCase() === 'label') {
                return;
            }
            if (index !== undefined) {
                if (!isMultiSelect) {
                    $selected = index;
                } else {
                    $selected = this.localSelected;
                }

                this.$emit('update:selected', $selected);
                $selectedList = this.calcSelected($selected);
            } else {
                this.$emit('update:selected', -1);
            }

            this.$emit('item-selected', $selected, $selectedList);

            if (!isMultiSelect) {
                this.openFn();
            }
        },
        editFn: function editFn(index) {
            this.$emit('edit', index);
            this.openFn();
        },
        delFn: function delFn(index) {
            this.$emit('del', index);
            this.openFn();
        },
        createFn: function createFn(e) {
            this.$emit('create');
            this.openFn();
            e.stopPropagation();
        },
        inputFn: function inputFn() {
            this.$emit('typing', this.condition);
        }
    }
};

bkSelector.install = function (Vue$$1) {
    Vue$$1.component(bkSelector.name, bkSelector);
};

var Render$1 = {
    name: 'RenderCell',
    functional: true,
    props: {
        render: Function
    },
    render: function render(h, ctx) {
        return ctx.props.render(h);
    }
};

var bkTab = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { ref: "bkTab2", class: ['bk-tab2', { 'bk-tab2-small': _vm.size === 'small' }] }, [_vm.needScroll ? [_c('div', { staticClass: "bk-tab2-control left", class: _vm.leftScrollDisabled ? 'disabled' : '', on: { "click": function click($event) {
                    _vm.scroll('left');
                } } }, [_c('i', { staticClass: "bk-tab2-icon bk-icon icon-angle-left" })]), _vm._v(" "), _c('div', { staticClass: "bk-tab2-control right", class: _vm.rightScrollDisabled ? 'disabled' : '', on: { "click": function click($event) {
                    _vm.scroll('right');
                } } }, [_c('i', { staticClass: "bk-tab2-icon bk-icon icon-angle-right" })])] : _vm._e(), _vm._v(" "), _c('div', { ref: "bkTab2Header", class: ['bk-tab2-head', { 'is-fill': _vm.type === 'fill' }, { 'scroll': _vm.needScroll }] }, [_c('div', { ref: "bkTab2Nav", staticClass: "bk-tab2-nav", style: _vm.bkTab2NavStyleObj }, _vm._l(_vm.navList, function (item, index) {
            return item.show ? _c('div', { key: index, staticClass: "tab2-nav-item", class: { 'active': _vm.calcActiveName === item.name }, attrs: { "title": item.title }, on: { "click": function click($event) {
                        _vm.toggleTab(index, $event);
                    } } }, [item.label ? _c('Render', { attrs: { "render": item.label } }) : _vm._e(), _vm._v(" "), item.tag !== undefined ? _c('div', { staticClass: "bk-panel-label" }, [_c('span', { staticClass: "bk-panel-tag" }, [_vm._v(_vm._s(item.tag))])]) : _vm._e(), _vm._v(" "), _c('span', { staticClass: "bk-panel-title" }, [_vm._v(_vm._s(item.title))])], 1) : _vm._e();
        })), _vm._v(" "), _vm.hasSetting && !_vm.needScroll ? _c('div', { staticClass: "bk-tab2-action" }, [_c('div', { staticClass: "action-wrapper" }, [_vm._t("setting")], 2)]) : _vm._e()]), _vm._v(" "), _c('div', { staticClass: "bk-tab2-content" }, [_vm._t("default")], 2)], 2);
    }, staticRenderFns: [],
    name: 'bk-tab',
    props: {
        activeName: {
            type: String,
            default: ''
        },
        type: {
            type: String,
            default: ''
        },
        size: {
            type: String,
            default: ''
        }
    },
    components: {
        Render: Render$1
    },
    data: function data() {
        return {
            navList: [],
            calcActiveName: this.activeName,
            hasSetting: false,
            needScroll: false,
            bkTab2NavStyleObj: {},
            translateX: 0,
            bkTab2HeaderWidth: 0,
            bkTab2NavWidth: 0,
            scrollDistance: 180,
            rightScrollDisabled: false,
            leftScrollDisabled: true
        };
    },

    watch: {
        activeName: function activeName(val) {
            this.calcActiveName = val;
        },
        calcActiveName: function calcActiveName() {
            var _this = this;

            var index = 0;
            this.navList.forEach(function (nav, i) {
                if (nav.name === _this.calcActiveName) {
                    index = i;
                }
            });
            this.$emit('tab-changed', this.calcActiveName, index);
        },
        translateX: function translateX() {
            var node = this.$refs.bkTab2Nav;
            if (!node) {
                return;
            }
            node.style.transform = node.style.webkitTransform = node.style.MozTransform = node.style.msTransform = node.style.OTransform = 'translateX(' + this.translateX + 'px)';
        }
    },
    mounted: function mounted() {
        var _this2 = this;

        if (!this.activeName) {
            this.calcActiveName = this.navList[0].name;
        }
        this.hasSetting = !!this.$slots.setting;
        this.$nextTick(function () {
            var bkTab2Width = parseInt(getStyle(_this2.$refs.bkTab2, 'width'), 10) || 0;
            var bkTab2NavWidth = parseInt(getStyle(_this2.$refs.bkTab2Nav, 'width'), 10) || 0;
            if (bkTab2Width < bkTab2NavWidth) {
                _this2.needScroll = true;
                _this2.bkTab2NavStyleObj = {
                    transform: 'translateX(' + _this2.translateX + 'px)'
                };
            } else {
                _this2.bkTab2NavStyleObj = {
                    width: '100%'
                };
            }
            _this2.bkTab2HeaderWidth = parseInt(getStyle(_this2.$refs.bkTab2Header, 'width'), 10) || 0;
            _this2.bkTab2NavWidth = parseInt(getStyle(_this2.$refs.bkTab2Nav, 'width'), 10) || 0;
        });
    },

    methods: {
        toggleTab: function toggleTab(index) {
            this.calcActiveName = this.navList[index].name;
            this.$emit('update:activeName', this.calcActiveName);
        },
        addNavItem: function addNavItem(item) {
            this.navList.push(item);
            return this.navList.length - 1;
        },
        scroll: function scroll(direction) {
            this.rightScrollDisabled = false;
            this.leftScrollDisabled = false;
            var leftDistance = 0;
            if (direction === 'right') {
                leftDistance = this.bkTab2NavWidth - (this.bkTab2HeaderWidth - 60 + Math.abs(this.translateX));
                if (leftDistance <= this.scrollDistance) {
                    this.translateX -= leftDistance;
                    this.rightScrollDisabled = true;
                } else {
                    this.translateX -= this.scrollDistance;
                }
            } else {
                if (Math.abs(this.translateX) <= this.scrollDistance) {
                    this.leftScrollDisabled = true;
                }
                leftDistance = Math.abs(this.translateX);
                if (leftDistance >= this.scrollDistance) {
                    this.translateX += this.scrollDistance;
                } else {
                    this.translateX += leftDistance;
                }
            }
        },
        updateList: function updateList(index, data) {
            var list = [];
            this.navList.forEach(function (nav) {
                list.push(Object.assign({}, nav));
            });

            var item = list[index];
            Object.keys(item).forEach(function (key) {
                item[key] = data[key];
            });
            this.navList = [].concat(list);

            if (this.calcActiveName === data.name && !data.show) {
                this.calcActiveName = this.navList[0].name;
            }
        }
    }
};

var bkTabpanel = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('section', { class: { 'active': _vm.isActive, 'bk-tab2-pane': !_vm.show || !_vm.isActive } }, [_vm._t("default")], 2);
    }, staticRenderFns: [],
    name: 'bk-tabpanel',
    props: {
        name: {
            type: String,
            default: ''
        },
        title: {
            type: String,
            default: ''
        },
        show: {
            type: Boolean,
            default: true
        },
        tag: {}
    },
    data: function data() {
        return {
            index: -1
        };
    },

    computed: {
        isActive: function isActive() {
            return this.$parent.calcActiveName === this.name;
        },
        panelParams: function panelParams() {
            return {
                title: this.title,
                name: this.name,
                show: this.show,
                tag: this.tag
            };
        }
    },
    watch: {
        panelParams: {
            deep: true,
            handler: function handler(val) {
                this.$parent.updateList(this.index, val);
            }
        }
    },
    mounted: function mounted() {
        var panelParams = JSON.parse(JSON.stringify(this.panelParams));
        var pannelSlots = this.$slots;
        var parentSlots = this.$parent.$slots;
        for (var slot in pannelSlots) {
            if (slot === 'tag') {
                (function () {
                    var vnode = pannelSlots['tag'];
                    panelParams.vnode = vnode;
                    panelParams.label = function (h) {
                        return h('div', {
                            class: ['bk-panel-label']
                        }, vnode);
                    };
                })();
            }
        }
        this.index = this.$parent.addNavItem(panelParams);
    }
};

bkTab.install = function (Vue$$1) {
    Vue$$1.component(bkTab.name, bkTab);
};

bkTabpanel.install = function (Vue$$1) {
    Vue$$1.component(bkTabpanel.name, bkTabpanel);
};

var Tab = {
    bkTab: bkTab,
    bkTabpanel: bkTabpanel
};

var bkPagination = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-page-count" }, [_c('div', { staticClass: "bk-total-page" }, [_vm._v(_vm._s('共计' + _vm.totalPage + '页'))]), _vm._v(" "), _c('bk-selector', { attrs: { "placeholder": '页数', "selected": _vm.paginationIndex, "list": _vm.paginationListTmp, "setting-key": 'id', "display-key": 'count' }, on: { "update:selected": function updateSelected($event) {
                    _vm.paginationIndex = $event;
                } } })], 1);
    }, staticRenderFns: [],
    name: 'bk-pagination',
    components: {
        bkSelector: bkSelector
    },
    props: {
        paginationCount: {
            type: Number,
            default: 10,
            validator: function validator(value) {
                return value >= 0;
            }
        },
        totalPage: {
            type: Number,
            default: 5,
            validator: function validator(value) {
                return value >= 0;
            }
        },
        paginationList: {
            type: Array,
            default: function _default() {
                return [10, 20, 50, 100];
            }
        }
    },
    data: function data() {
        return {
            paginationIndex: this.paginationCount,
            paginationListTmp: []
        };
    },
    created: function created() {
        this.initData();
    },

    watch: {
        paginationCount: function paginationCount(value) {
            if (this.paginationList.includes(value)) {
                this.paginationIndex = value;
            } else {
                this.paginationIndex = this.paginationList[0];
            }
        },
        paginationIndex: function paginationIndex(value) {
            this.$emit('update:paginationCount', value);
        }
    },
    methods: {
        initData: function initData() {
            this.paginationListTmp = this.paginationList.map(function (page) {
                return {
                    id: page,
                    count: page
                };
            });
        }
    }
};

var bkPaging = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _vm.totalPage > 0 ? _c('div', { class: ['bk-page', { 'bk-page-compact': _vm.type === 'compact', 'bk-page-small': _vm.size === 'small' }] }, [_vm.paginationAble && _vm.location === 'left' ? [_c('bk-pagination', { attrs: { "total-page": _vm.totalPage, "pagination-count": _vm.paginationCountTmp, "pagination-list": _vm.paginationList }, on: { "update:paginationCount": function updatePaginationCount($event) {
                    _vm.paginationCountTmp = $event;
                } } })] : _vm._e(), _vm._v(" "), _c('ul', [_c('li', { staticClass: "page-item", class: { disabled: _vm.curPage === 1 }, on: { "click": _vm.prevPage } }, [_vm._m(0)]), _vm._v(" "), _c('li', { directives: [{ name: "show", rawName: "v-show", value: _vm.renderList[0] > 1, expression: "renderList[0] > 1" }], staticClass: "page-item", on: { "click": function click($event) {
                    _vm.jumpToPage(1);
                } } }, [_c('a', { staticClass: "page-button", attrs: { "href": "javascript:void(0);" } }, [_vm._v("1")])]), _vm._v(" "), _c('li', { directives: [{ name: "show", rawName: "v-show", value: _vm.renderList[0] > 2 && _vm.curPage > 3, expression: "renderList[0] > 2 && curPage > 3" }], class: ['page-item', { 'page-omit': _vm.type !== 'compact' }], on: { "click": _vm.prevGroup } }, [_c('span', { staticClass: "page-button" }, [_vm._v("...")])]), _vm._v(" "), _vm._l(_vm.renderList, function (item) {
            return _c('li', { staticClass: "page-item", class: { 'cur-page': item === _vm.curPage }, on: { "click": function click($event) {
                        _vm.jumpToPage(item);
                    } } }, [_c('a', { staticClass: "page-button", attrs: { "href": "javascript:void(0);" } }, [_vm._v(_vm._s(item))])]);
        }), _vm._v(" "), _c('li', { directives: [{ name: "show", rawName: "v-show", value: _vm.renderList[_vm.renderList.length - 1] < _vm.calcTotalPage - 1, expression: "renderList[renderList.length - 1] < calcTotalPage - 1" }], class: ['page-item', { 'page-omit': _vm.type !== 'compact' }], on: { "click": _vm.nextGroup } }, [_c('span', { staticClass: "page-button" }, [_vm._v("...")])]), _vm._v(" "), _c('li', { directives: [{ name: "show", rawName: "v-show", value: _vm.renderList[_vm.renderList.length - 1] !== _vm.calcTotalPage, expression: "renderList[renderList.length - 1] !== calcTotalPage" }], staticClass: "page-item", class: { 'cur-page': _vm.curPage === _vm.calcTotalPage }, on: { "click": function click($event) {
                    _vm.jumpToPage(_vm.calcTotalPage);
                } } }, [_c('a', { staticClass: "page-button", attrs: { "href": "javascript:void(0);" } }, [_vm._v(_vm._s(_vm.calcTotalPage))])]), _vm._v(" "), _c('li', { staticClass: "page-item", class: { disabled: _vm.curPage === _vm.calcTotalPage }, on: { "click": _vm.nextPage } }, [_vm._m(1)])], 2), _vm._v(" "), _vm.paginationAble && _vm.location === 'right' ? [_c('bk-pagination', { attrs: { "total-page": _vm.totalPage, "pagination-count": _vm.paginationCountTmp, "pagination-list": _vm.paginationList }, on: { "update:paginationCount": function updatePaginationCount($event) {
                    _vm.paginationCountTmp = $event;
                } } })] : _vm._e()], 2) : _vm._e();
    }, staticRenderFns: [function () {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('a', { staticClass: "page-button", attrs: { "href": "javascript:void(0);" } }, [_c('i', { staticClass: "bk-icon icon-angle-left" })]);
    }, function () {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('a', { staticClass: "page-button", attrs: { "href": "javascript:void(0);" } }, [_c('i', { staticClass: "bk-icon icon-angle-right" })]);
    }],
    name: 'bk-paging',
    components: {
        bkPagination: bkPagination
    },
    props: {
        type: {
            type: String,
            default: 'default',
            validator: function validator(value) {
                return ['default', 'compact'].indexOf(value) > -1;
            }
        },
        size: {
            type: String,
            default: 'default',
            validator: function validator(value) {
                return ['default', 'small'].indexOf(value) > -1;
            }
        },
        curPage: {
            type: Number,
            default: 1,
            required: true
        },
        totalPage: {
            type: Number,
            default: 5,
            validator: function validator(value) {
                return value >= 0;
            }
        },
        paginationCount: {
            type: Number,
            default: 10,
            validator: function validator(value) {
                return value >= 0;
            }
        },
        paginationAble: {
            type: Boolean,
            default: false
        },
        location: {
            type: String,
            default: 'left',
            validator: function validator(value) {
                return ['left', 'right'].indexOf(value) > -1;
            }
        },
        paginationList: {
            type: Array,
            default: function _default() {
                return [10, 20, 50, 100];
            }
        }
    },
    data: function data() {
        return {
            pageSize: 5,
            renderList: [],
            curGroup: 1,
            paginationCountTmp: this.paginationCount
        };
    },

    computed: {
        calcTotalPage: function calcTotalPage() {
            if (this.totalPage >= this.curPage) {
                return this.totalPage;
            }

            this.$emit('update:curPage', this.totalPage);
            return this.curPage;
        }
    },
    created: function created() {
        this.calcPageList(this.curPage);
    },

    watch: {
        curPage: function curPage(newVal) {
            this.calcPageList(newVal);
        },
        totalPage: function totalPage(newVal) {
            this.calcPageList(this.curPage);
        },
        paginationCountTmp: function paginationCountTmp(newVal) {
            this.$emit('pagination-change', newVal);
        },
        paginationCount: function paginationCount(newVal) {
            this.paginationCountTmp = newVal;
        }
    },
    methods: {
        _array: function _array(size) {
            return Array.apply(null, {
                length: size
            });
        },
        calcPageList: function calcPageList(curPage) {
            var total = this.calcTotalPage;
            var pageSize = this.pageSize;
            var size = pageSize > total ? total : pageSize;

            if (curPage >= size - 1) {
                if (total - curPage > Math.floor(size / 2)) {
                    this.renderList = this._array(size).map(function (v, i) {
                        return i + curPage - Math.ceil(size / 2) + 1;
                    });
                } else {
                    this.renderList = this._array(size).map(function (v, i) {
                        return total - i;
                    }).reverse();
                }
            } else {
                this.renderList = this._array(size).map(function (v, i) {
                    return i + 1;
                });
            }
        },
        prevGroup: function prevGroup() {
            var pageSize = this.pageSize;
            var middlePage = this.renderList[Math.ceil(this.renderList.length / 2)];

            if (middlePage - pageSize < 1) {
                this.calcPageList(1);
            } else {
                this.calcPageList(middlePage - pageSize);
            }
            this.jumpToPage(this.renderList[Math.floor(this.renderList.length / 2)]);
        },
        nextGroup: function nextGroup() {
            var pageSize = this.pageSize;
            var totalPage = this.calcTotalPage;
            var middlePage = this.renderList[Math.ceil(this.renderList.length / 2)];

            if (middlePage + pageSize > totalPage) {
                this.calcPageList(totalPage);
            } else {
                this.calcPageList(middlePage + pageSize);
            }
            this.jumpToPage(this.renderList[Math.floor(this.renderList.length / 2)]);
        },
        prevPage: function prevPage() {
            if (this.curPage !== 1) {
                this.jumpToPage(this.curPage - 1);
            }
        },
        nextPage: function nextPage() {
            if (this.curPage !== this.calcTotalPage) {
                this.jumpToPage(this.curPage + 1);
            }
        },
        jumpToPage: function jumpToPage(page) {
            this.$emit('update:curPage', page);
            this.$emit('page-change', page);
        }
    }
};

bkPaging.install = function (Vue$$1) {
    Vue$$1.component(bkPaging.name, bkPaging);
};

var bkTransfer = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { ref: "transfer", staticClass: "bk-transfer" }, [_c('div', { staticClass: "source-list" }, [_vm.slot['left-header'] ? _c('div', { staticClass: "slot-header" }, [_c('div', { staticClass: "slot-content" }, [_vm._t("left-header")], 2)]) : _c('div', { staticClass: "header" }, [_vm._v(" " + _vm._s(_vm.title[0] ? _vm.title[0] : '左侧列表') + "（共" + _vm._s(_vm.dataList.length) + "条） "), _vm.dataList.length === 0 ? _c('span', { staticClass: "disabled" }, [_vm._v("全部添加")]) : _c('span', { on: { "click": _vm.allToRight } }, [_vm._v("全部添加")])]), _vm._v(" "), _vm.dataList.length ? [_c('ul', { staticClass: "content" }, _vm._l(_vm.dataList, function (item, index) {
            return _c('li', { key: index, on: { "click": function click($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.leftClick(index);
                    }, "mouseover": function mouseover($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.leftMouseover(index);
                    }, "mouseleave": function mouseleave($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.leftMouseleave(index);
                    } } }, [_c('span', [_vm._v(_vm._s(item[_vm.displayCode]))]), _vm._v(" "), _c('span', { staticClass: "icon-wrapper", class: [index === _vm.leftHoverIndex ? 'hover' : ''] }, [_c('i', { staticClass: "bk-icon icon-arrows-right" })])]);
        }))] : [_vm.slot['left-empty-content'] ? _c('div', [_vm._t("left-empty-content")], 2) : _c('div', { staticClass: "empty" }, [_vm._v(" " + _vm._s(_vm.emptyContent[0] ? _vm.emptyContent[0] : '无数据') + " ")])]], 2), _vm._v(" "), _c('div', { staticClass: "transfer" }), _vm._v(" "), _c('div', { staticClass: "target-list" }, [_vm.slot['right-header'] ? _c('div', { staticClass: "slot-header" }, [_c('div', { staticClass: "slot-content" }, [_vm._t("right-header")], 2)]) : _c('div', { staticClass: "header" }, [_vm._v(" " + _vm._s(_vm.title[1] ? _vm.title[1] : '右侧列表') + "（共" + _vm._s(_vm.hasSelectedList.length) + "条） "), _vm.hasSelectedList.length === 0 ? _c('span', { staticClass: "disabled" }, [_vm._v("全部移除")]) : _c('span', { on: { "click": _vm.allToLeft } }, [_vm._v("全部移除")])]), _vm._v(" "), _vm.hasSelectedList.length ? [_c('ul', { staticClass: "content" }, _vm._l(_vm.hasSelectedList, function (item, index) {
            return _c('li', { key: index, on: { "click": function click($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.rightClick(index);
                    }, "mouseover": function mouseover($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.rightMouseover(index);
                    }, "mouseleave": function mouseleave($event) {
                        $event.stopPropagation();$event.preventDefault();_vm.rightMouseleave(index);
                    } } }, [_c('span', [_vm._v(_vm._s(item[_vm.displayCode]))]), _vm._v(" "), _c('span', { staticClass: "icon-wrapper", class: [index === _vm.rightHoverIndex ? 'hover' : ''] }, [_c('i', { staticClass: "bk-icon icon-close" })])]);
        }))] : [_vm.slot['right-empty-content'] ? _c('div', [_vm._t("right-empty-content")], 2) : _c('div', { staticClass: "empty" }, [_vm._v(" " + _vm._s(_vm.emptyContent[1] ? _vm.emptyContent[1] : '未选择任何项') + " ")])]], 2)]);
    }, staticRenderFns: [],
    name: 'bk-transfer',
    props: {
        title: {
            type: Array,
            default: function _default() {
                return ['左侧列表', '右侧列表'];
            }
        },
        emptyContent: {
            type: Array,
            default: function _default() {
                return ['无数据', '未选择任何项'];
            }
        },
        displayKey: {
            type: String,
            default: 'value'
        },
        settingKey: {
            type: String,
            default: 'id'
        },
        sortKey: {
            type: String,
            default: ''
        },
        sourceList: {
            type: Array,
            default: function _default() {
                return [];
            }
        },
        targetList: {
            type: Array,
            default: function _default() {
                return [];
            }
        },
        hasHeader: {
            type: Boolean,
            default: false
        },
        sortable: {
            type: Boolean,
            default: false
        }
    },
    data: function data() {
        return {
            dataList: [],
            hasSelectedList: [],
            sortList: [],
            leftHoverIndex: -1,
            rightHoverIndex: -1,
            slot: {},
            displayCode: this.displayKey,
            settingCode: this.settingKey,
            sortCode: this.sortKey,
            isSortFlag: this.sortable
        };
    },

    computed: {
        typeFlag: function typeFlag() {
            if (!this.sourceList || !this.sourceList.length) {
                return 'empty';
            } else {
                var str = this.sourceList.toString();
                if (str.indexOf('[object Object]') !== -1) {
                    return true;
                } else {
                    return false;
                }
            }
        }
    },
    watch: {
        sourceList: {
            handler: function handler(value) {
                if (!value || !value.length) {
                    return;
                }
                this.initData();
                this.initSort();
            },
            deep: true
        },
        targetList: {
            handler: function handler(value) {
                this.initData();
                this.initSort();
            },
            deep: true
        },
        displayKey: function displayKey(value) {
            this.displayCode = value;
            this.initData();
        },
        settingKey: function settingKey(value) {
            this.settingCode = value;
            this.initData();
        },
        sortKey: function sortKey(value) {
            this.sortCode = value;
            this.initSort();
        },
        sortable: function sortable(value) {
            this.isSortFlag = value;
            this.initSort();
        }
    },
    created: function created() {
        if (this.typeFlag !== 'empty') {
            if (!this.typeFlag) {
                this.generalInit();
            } else {
                this.init();
            }
            this.initSort();
        }
        this.slot = Object.assign({}, this.$slots);
    },

    methods: {
        initData: function initData() {
            if (this.typeFlag !== 'empty') {
                if (!this.typeFlag) {
                    this.generalInit();
                } else {
                    this.init();
                }
            }
        },
        generalInit: function generalInit() {
            if (!this.targetList.length || this.targetList.length > this.sourceList.length) {
                var _hasSelectedList;

                var list = [];
                for (var i = 0; i < this.sourceList.length; i++) {
                    list.push({
                        index: i,
                        value: this.sourceList[i]
                    });
                }
                this.dataList = [].concat(list);
                (_hasSelectedList = this.hasSelectedList).splice.apply(_hasSelectedList, [0, this.hasSelectedList.length].concat([]));
                this.$emit('change', this.dataList, [], []);
            } else {
                var _list = [];
                var valueList = [];
                for (var _i = 0; _i < this.sourceList.length; _i++) {
                    _list.push({
                        index: _i,
                        value: this.sourceList[_i]
                    });
                }
                for (var j = 0; j < _list.length; j++) {
                    for (var k = 0; k < this.targetList.length; k++) {
                        if (_list[j].value === this.targetList[k]) {
                            valueList.push(_list[j]);
                        }
                    }
                }
                this.hasSelectedList = [].concat(valueList);
                var result = _list.filter(function (item1) {
                    return valueList.every(function (item2) {
                        return item2['index'] !== item1['index'];
                    });
                });
                this.dataList = [].concat(toConsumableArray(result));
                this.$emit('change', this.dataList, [].concat(toConsumableArray(this.generalListHandler(this.hasSelectedList))), []);
            }
        },
        init: function init() {
            var _this = this;

            if (!this.targetList.length || this.targetList.length > this.sourceList.length) {
                this.dataList = [].concat(toConsumableArray(this.sourceList));
                this.hasSelectedList = [];
                this.$emit('change', this.dataList, [], []);
            } else {
                var result = this.sourceList.filter(function (item1) {
                    return _this.targetList.every(function (item2) {
                        return item2 !== item1[_this.settingCode];
                    });
                });
                var hasTempList = [];
                this.sourceList.forEach(function (item1) {
                    _this.targetList.forEach(function (item2) {
                        if (item1[_this.settingCode] === item2) {
                            hasTempList.push(item1);
                        }
                    });
                });
                this.hasSelectedList = [].concat(hasTempList);
                this.dataList = [].concat(toConsumableArray(result));
                var list = [].concat(toConsumableArray(this.sourceListHandler(this.hasSelectedList)));
                this.$emit('change', this.dataList, this.hasSelectedList, list);
            }
        },
        generalListHandler: function generalListHandler(list) {
            var templateList = [];
            if (!list.length) {
                return [];
            } else {
                var dataList = [].concat(toConsumableArray(list));
                dataList.forEach(function (item) {
                    templateList.push(item.value);
                });
                return templateList;
            }
        },
        sourceListHandler: function sourceListHandler(list) {
            var _this2 = this;

            var templateList = [];
            if (!list.length) {
                return [];
            } else {
                var dataList = [].concat(toConsumableArray(list));
                dataList.forEach(function (item) {
                    for (var key in item) {
                        if (key === _this2.settingCode) {
                            templateList.push(item[key]);
                        }
                    }
                });
                return templateList;
            }
        },
        initSort: function initSort() {
            var _this3 = this;

            var templateList = [];
            if (!this.typeFlag) {
                if (this.isSortFlag) {
                    this.sortCode = 'index';
                } else {
                    this.sortCode = '';
                }
                for (var k = 0; k < this.sourceList.length; k++) {
                    templateList.push({
                        index: k,
                        value: this.sourceList[k]
                    });
                }
            } else {
                if (!this.isSortFlag) {
                    this.sortCode = '';
                }
                templateList = [].concat(toConsumableArray(this.sourceList));
            }
            if (this.sortCode) {
                var arr = [];
                templateList.forEach(function (item) {
                    arr.push(item[_this3.sortCode]);
                });
                this.sortList = [].concat(arr);
                if (this.sortList.length === this.sourceList.length) {
                    var list = [].concat(toConsumableArray(this.dataList));
                    this.dataList = [].concat(toConsumableArray(this.sortDataList(list, this.sortCode, this.sortList)));
                }
            }
        },
        sortDataList: function sortDataList(list, key, sortList) {
            var arr = sortList;
            return list.sort(function (a, b) {
                return arr.indexOf(a[key]) - arr.indexOf(b[key]) >= 0;
            });
        },
        allToRight: function allToRight() {
            this.leftHoverIndex = -1;
            var dataList = this.dataList;
            var hasSelectedList = this.hasSelectedList;
            while (dataList.length) {
                var transferItem = dataList.shift();
                hasSelectedList.push(transferItem);
                if (this.sortList.length === this.sourceList.length) {
                    this.hasSelectedList = [].concat(toConsumableArray(this.sortDataList(hasSelectedList, this.sortCode, this.sortList)));
                } else {
                    this.hasSelectedList = [].concat(toConsumableArray(hasSelectedList));
                }

            }
            if (!this.typeFlag) {
                this.$emit('change', this.dataList, [].concat(toConsumableArray(this.generalListHandler(this.hasSelectedList))), []);
            } else {
                var list = [].concat(toConsumableArray(this.sourceListHandler(this.hasSelectedList)));
                this.$emit('change', this.dataList, this.hasSelectedList, list);
            }
        },
        allToLeft: function allToLeft() {
            this.rightHoverIndex = -1;
            var hasSelectedList = this.hasSelectedList;
            var dataList = this.dataList;
            while (hasSelectedList.length) {
                var transferItem = hasSelectedList.shift();
                dataList.push(transferItem);
                if (this.sortList.length === this.sourceList.length) {
                    this.dataList = [].concat(toConsumableArray(this.sortDataList(dataList, this.sortCode, this.sortList)));
                } else {
                    this.dataList = [].concat(toConsumableArray(dataList));
                }

            }
            if (!this.typeFlag) {
                this.$emit('change', this.dataList, [].concat(toConsumableArray(this.generalListHandler(this.hasSelectedList))), []);
            } else {
                var list = [].concat(toConsumableArray(this.sourceListHandler(this.hasSelectedList)));
                this.$emit('change', this.dataList, this.hasSelectedList, list);
            }
        },
        leftClick: function leftClick(index) {
            this.leftHoverIndex = -1;
            var transferItem = this.dataList.splice(index, 1)[0];
            var hasSelectedList = this.hasSelectedList;
            hasSelectedList.push(transferItem);
            if (this.sortList.length === this.sourceList.length) {
                this.hasSelectedList = [].concat(toConsumableArray(this.sortDataList(hasSelectedList, this.sortCode, this.sortList)));
            } else {
                this.hasSelectedList = [].concat(toConsumableArray(hasSelectedList));
            }
            if (!this.typeFlag) {
                this.$emit('change', this.dataList, [].concat(toConsumableArray(this.generalListHandler(this.hasSelectedList))), []);
            } else {
                var list = [].concat(toConsumableArray(this.sourceListHandler(this.hasSelectedList)));
                this.$emit('change', this.dataList, this.hasSelectedList, list);
            }
        },
        rightClick: function rightClick(index) {
            this.rightHoverIndex = -1;
            var transferItem = this.hasSelectedList.splice(index, 1)[0];
            var dataList = this.dataList;
            dataList.push(transferItem);
            if (this.sortList.length === this.sourceList.length) {
                this.dataList = [].concat(toConsumableArray(this.sortDataList(dataList, this.sortCode, this.sortList)));
            } else {
                this.dataList = [].concat(toConsumableArray(dataList));
            }
            if (!this.typeFlag) {
                this.$emit('change', this.dataList, [].concat(toConsumableArray(this.generalListHandler(this.hasSelectedList))), []);
            } else {
                var list = [].concat(toConsumableArray(this.sourceListHandler(this.hasSelectedList)));
                this.$emit('change', this.dataList, this.hasSelectedList, list);
            }
        },
        leftMouseover: function leftMouseover(index) {
            this.leftHoverIndex = index;
        },
        leftMouseleave: function leftMouseleave(index) {
            this.leftHoverIndex = -1;
        },
        rightMouseover: function rightMouseover(index) {
            this.rightHoverIndex = index;
        },
        rightMouseleave: function rightMouseleave(index) {
            this.rightHoverIndex = -1;
        }
    }
};

bkTransfer.install = function (Vue$$1) {
    Vue$$1.component(bkTransfer.name, bkTransfer);
};

var Render$2 = {
    name: 'render',
    functional: true,
    props: {
        node: Object,
        tpl: Function
    },
    render: function render(h, ct) {
        var titleClass = ct.props.node.selected ? 'node-title node-selected' : 'node-title';
        if (ct.props.tpl) {
            return ct.props.tpl(ct.props.node, ct);
        }
        return h('span', {
            domProps: {
                'innerHTML': ct.props.node.name
            },
            attrs: { title: ct.props.node.title },
            'class': titleClass,
            style: 'user-select: none',
            on: {
                'click': function click() {
                    return ct.parent.nodeSelected(ct.props.node);
                }
            }
        });
    }
};

var Transition = {
    'before-enter': function beforeEnter(el) {
        if (!el.dataset) el.dataset = {};

        el.dataset.oldPaddingTop = el.style.paddingTop;
        el.dataset.oldPaddingBottom = el.style.paddingBottom;

        el.style.height = '0';
        el.style.paddingTop = 0;
        el.style.paddingBottom = 0;
    },
    'enter': function enter(el) {
        el.dataset.oldOverflow = el.style.overflow;
        if (el.scrollHeight !== 0) {
            el.style.height = el.scrollHeight + 'px';
            el.style.paddingTop = el.dataset.oldPaddingTop;
            el.style.paddingBottom = el.dataset.oldPaddingBottom;
        } else {
            el.style.height = '';
            el.style.paddingTop = el.dataset.oldPaddingTop;
            el.style.paddingBottom = el.dataset.oldPaddingBottom;
        }

        el.style.overflow = 'hidden';
    },
    'after-enter': function afterEnter(el) {
        el.style.height = '';
        el.style.overflow = el.dataset.oldOverflow;
    },
    'before-leave': function beforeLeave(el) {
        if (!el.dataset) el.dataset = {};
        el.dataset.oldPaddingTop = el.style.paddingTop;
        el.dataset.oldPaddingBottom = el.style.paddingBottom;
        el.dataset.oldOverflow = el.style.overflow;

        el.style.height = el.scrollHeight + 'px';
        el.style.overflow = 'hidden';
    },
    'leave': function leave(el) {
        if (el.scrollHeight !== 0) {
            el.style.height = 0;
            el.style.paddingTop = 0;
            el.style.paddingBottom = 0;
        }
    },
    'after-leave': function afterLeave(el) {
        el.style.height = '';
        el.style.overflow = el.dataset.oldOverflow;
        el.style.paddingTop = el.dataset.oldPaddingTop;
        el.style.paddingBottom = el.dataset.oldPaddingBottom;
    }
};
var CollapseTransition = {
    name: 'CollapseTransition',
    functional: true,
    render: function render(h, _ref) {
        var children = _ref.children;

        var data = {
            on: Transition
        };
        return h('transition', data, children);
    }
};

var bkTree = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('ul', { staticClass: "bk-tree", class: { 'bk-has-border-tree': _vm.isBorder } }, [_vm._l(_vm.data, function (item, index) {
            return (item.hasOwnProperty('visible') ? item.visible : true) ? _c('li', { key: item[_vm.nodeKey] ? item[_vm.nodeKey] : item.name, class: { 'leaf': _vm.isLeaf(item), 'tree-first-node': !_vm.parent && index === 0, 'tree-only-node': !_vm.parent && _vm.data.length === 1, 'tree-second-node': !_vm.parent && index === 1, 'single': !_vm.multiple }, on: { "drop": function drop($event) {
                        _vm.drop(item, $event);
                    }, "dragover": function dragover($event) {
                        _vm.dragover($event);
                    } } }, [_c('div', { class: ['tree-drag-node', !_vm.multiple ? 'tree-singe' : ''], attrs: { "draggable": _vm.draggable }, on: { "dragstart": function dragstart($event) {
                        _vm.drag(item, $event);
                    } } }, [!item.parent || item.children && item.children.length || item.async ? _c('span', { class: ['bk-icon', 'tree-expanded-icon', item.expanded ? 'icon-down-shape' : 'icon-right-shape'], on: { "click": function click($event) {
                        _vm.expandNode(item);
                    } } }) : _vm._e(), _vm._v(" "), _vm.multiple && !item.nocheck ? _c('label', { class: [item.halfcheck ? 'bk-form-half-checked' : 'bk-form-checkbox', 'bk-checkbox-small', 'mr5'] }, [_vm.multiple ? _c('input', { directives: [{ name: "model", rawName: "v-model", value: item.checked, expression: "item.checked" }], attrs: { "type": "checkbox", "disabled": item.disabled }, domProps: { "checked": Array.isArray(item.checked) ? _vm._i(item.checked, null) > -1 : item.checked }, on: { "change": [function ($event) {
                        var $$a = item.checked,
                            $$el = $event.target,
                            $$c = $$el.checked ? true : false;if (Array.isArray($$a)) {
                            var $$v = null,
                                $$i = _vm._i($$a, $$v);if ($$el.checked) {
                                $$i < 0 && _vm.$set(item, "checked", $$a.concat([$$v]));
                            } else {
                                $$i > -1 && _vm.$set(item, "checked", $$a.slice(0, $$i).concat($$a.slice($$i + 1)));
                            }
                        } else {
                            _vm.$set(item, "checked", $$c);
                        }
                    }, function ($event) {
                        $event.stopPropagation();_vm.changeCheckStatus(item, $event);
                    }] } }) : _vm._e()]) : _vm._e(), _vm._v(" "), _c('div', { staticClass: "tree-node", on: { "click": function click($event) {
                        _vm.triggerExpand(item);
                    } } }, [item.icon || item.openedIcon || item.closedIcon ? _c('span', { staticClass: "node-icon bk-icon", class: _vm.setNodeIcon(item) }) : _vm._e(), _vm._v(" "), item.loading && item.expanded ? _c('div', { staticClass: "bk-spin-loading bk-spin-loading-mini bk-spin-loading-primary loading" }, [_c('div', { staticClass: "rotate rotate1" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate2" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate3" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate4" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate5" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate6" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate7" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate8" })]) : _vm._e(), _vm._v(" "), _c('Render', { attrs: { "node": item, "tpl": _vm.tpl } })], 1)]), _vm._v(" "), _c('collapse-transition', [!_vm.isLeaf(item) ? _c('bk-tree', { directives: [{ name: "show", rawName: "v-show", value: item.expanded, expression: "item.expanded" }], attrs: { "dragAfterExpanded": _vm.dragAfterExpanded, "draggable": _vm.draggable, "tpl": _vm.tpl, "data": item.children, "halfcheck": _vm.halfcheck, "parent": item, "isDeleteRoot": _vm.isDeleteRoot, "multiple": _vm.multiple }, on: { "dropTreeChecked": _vm.nodeCheckStatusChange, "async-load-nodes": _vm.asyncLoadNodes, "on-expanded": _vm.onExpanded, "on-click": _vm.onClick, "on-check": _vm.onCheck, "on-drag-node": _vm.onDragNode } }) : _vm._e()], 1)], 1) : _vm._e();
        }), _vm._v(" "), _vm.isEmpty && _vm.searchFlag ? _c('p', { staticClass: "search-no-data" }, [_vm._v(_vm._s(_vm.emptyText))]) : _vm._e()], 2);
    }, staticRenderFns: [],
    name: 'bk-tree',
    props: {
        data: {
            type: Array,
            default: function _default() {
                return [];
            }
        },
        parent: {
            type: Object,
            default: function _default() {
                return null;
            }
        },
        multiple: {
            type: Boolean,
            default: false
        },
        nodeKey: {
            type: String,
            default: 'id'
        },
        draggable: {
            type: Boolean,
            default: false
        },
        hasBorder: {
            type: Boolean,
            default: false
        },
        dragAfterExpanded: {
            type: Boolean,
            default: true
        },
        isDeleteRoot: {
            type: Boolean,
            default: false
        },
        emptyText: {
            type: String,
            default: '暂无数据'
        },
        tpl: Function
    },
    components: { Render: Render$2, CollapseTransition: CollapseTransition },
    data: function data() {
        return {
            halfcheck: true,
            isBorder: this.hasBorder,
            bkTreeDrag: {},
            visibleStatus: [],
            isEmpty: false,
            searchFlag: false
        };
    },

    watch: {
        data: function data() {
            this.initTreeData();
        },
        hasBorder: function hasBorder(value) {
            this.isBorder = !!value;
        }
    },
    mounted: function mounted() {
        var _this = this;

        this.$on('childChecked', function (node, checked) {
            if (node.children && node.children.length) {
                var _iteratorNormalCompletion = true;
                var _didIteratorError = false;
                var _iteratorError = undefined;

                try {
                    for (var _iterator = node.children[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                        var child = _step.value;

                        if (!child.disabled) {
                            _this.$set(child, 'checked', checked);
                        }
                        _this.$emit('on-check', child, checked);
                    }
                } catch (err) {
                    _didIteratorError = true;
                    _iteratorError = err;
                } finally {
                    try {
                        if (!_iteratorNormalCompletion && _iterator.return) {
                            _iterator.return();
                        }
                    } finally {
                        if (_didIteratorError) {
                            throw _iteratorError;
                        }
                    }
                }
            }
        });

        this.$on('parentChecked', function (node, checked) {
            if (!node.disabled) {
                _this.$set(node, 'checked', checked);
            }
            if (!node.parent) return false;
            var someBortherNodeChecked = node.parent.children.some(function (node) {
                return node.checked;
            });
            var allBortherNodeChecked = node.parent.children.every(function (node) {
                return node.checked;
            });
            if (_this.halfcheck) {
                allBortherNodeChecked ? _this.$set(node.parent, 'halfcheck', false) : someBortherNodeChecked ? _this.$set(node.parent, 'halfcheck', true) : _this.$set(node.parent, 'halfcheck', false);
                if (!checked && someBortherNodeChecked) {
                    _this.$set(node.parent, 'halfcheck', true);
                    return false;
                }
                _this.$emit('parentChecked', node.parent, checked);
            } else {
                if (checked && allBortherNodeChecked) _this.$emit('parentChecked', node.parent, checked);
                if (!checked) _this.$emit('parentChecked', node.parent, checked);
            }
        });

        this.$on('on-check', function (node, checked) {
            _this.$emit('parentChecked', node, checked);
            _this.$emit('childChecked', node, checked);
            _this.$emit('dropTreeChecked', node, checked);
        });

        this.$on('toggleshow', function (node, isShow) {
            _this.$set(node, 'visible', isShow);
            _this.visibleStatus.push(node.visible);
            if (_this.visibleStatus.every(function (item) {
                return !item;
            })) {
                _this.isEmpty = true;
                return;
            }
            if (isShow && node.parent) {
                _this.searchFlag = false;
                _this.$emit('toggleshow', node.parent, isShow);
            }
        });
        this.$on('cancelSelected', function (root) {
            var _iteratorNormalCompletion2 = true;
            var _didIteratorError2 = false;
            var _iteratorError2 = undefined;

            try {
                for (var _iterator2 = root.$children[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
                    var child = _step2.value;
                    var _iteratorNormalCompletion3 = true;
                    var _didIteratorError3 = false;
                    var _iteratorError3 = undefined;

                    try {
                        for (var _iterator3 = child.data[Symbol.iterator](), _step3; !(_iteratorNormalCompletion3 = (_step3 = _iterator3.next()).done); _iteratorNormalCompletion3 = true) {
                            var node = _step3.value;

                            child.$set(node, 'selected', false);
                        }
                    } catch (err) {
                        _didIteratorError3 = true;
                        _iteratorError3 = err;
                    } finally {
                        try {
                            if (!_iteratorNormalCompletion3 && _iterator3.return) {
                                _iterator3.return();
                            }
                        } finally {
                            if (_didIteratorError3) {
                                throw _iteratorError3;
                            }
                        }
                    }

                    if (child.$children) child.$emit('cancelSelected', child);
                }
            } catch (err) {
                _didIteratorError2 = true;
                _iteratorError2 = err;
            } finally {
                try {
                    if (!_iteratorNormalCompletion2 && _iterator2.return) {
                        _iterator2.return();
                    }
                } finally {
                    if (_didIteratorError2) {
                        throw _iteratorError2;
                    }
                }
            }
        });
        this.initTreeData();
    },
    destroyed: function destroyed() {
        this.$delete(window, 'bkTreeDrag');
    },

    methods: {
        gid: function gid() {
            return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
                var r = Math.random() * 16 | 0;
                var v = c === 'x' ? r : r & 0x3 | 0x8;
                return v.toString(16);
            });
        },
        setDragNode: function setDragNode(id, node) {
            window['bkTreeDrag'] = {};
            window['bkTreeDrag'][id] = node;
        },
        getDragNode: function getDragNode(id) {
            return window['bkTreeDrag'][id];
        },
        hasInGenerations: function hasInGenerations(root, node) {
            if (root.hasOwnProperty('children') && root.children) {
                var _iteratorNormalCompletion4 = true;
                var _didIteratorError4 = false;
                var _iteratorError4 = undefined;

                try {
                    for (var _iterator4 = root.children[Symbol.iterator](), _step4; !(_iteratorNormalCompletion4 = (_step4 = _iterator4.next()).done); _iteratorNormalCompletion4 = true) {
                        var rn = _step4.value;

                        if (rn === node) return true;
                        if (rn.children) return this.hasInGenerations(rn, node);
                    }
                } catch (err) {
                    _didIteratorError4 = true;
                    _iteratorError4 = err;
                } finally {
                    try {
                        if (!_iteratorNormalCompletion4 && _iterator4.return) {
                            _iterator4.return();
                        }
                    } finally {
                        if (_didIteratorError4) {
                            throw _iteratorError4;
                        }
                    }
                }

                return false;
            }
        },
        setNodeIcon: function setNodeIcon(node) {
            if (node.children && node.children.length) {
                if (node.expanded) {
                    return node.openedIcon;
                } else {
                    return node.closedIcon;
                }
            } else {
                return node.icon;
            }
        },
        drop: function drop(node, ev) {
            ev.preventDefault();
            ev.stopPropagation();
            var gid = ev.dataTransfer.getData('gid');
            var drag = this.getDragNode(gid);

            if (drag.parent === node || drag.parent === null || drag === node) return false;

            if (this.hasInGenerations(drag, node)) return false;
            var dragHost = drag.parent.children;
            if (node.children && node.children.indexOf(drag) === -1) {
                node.children.push(drag);
                dragHost.splice(dragHost.indexOf(drag), 1);
            } else {
                this.$set(node, 'openedIcon', 'icon-folder-open');
                this.$set(node, 'closedIcon', 'icon-folder');
                this.$set(node, 'children', [drag]);
                dragHost.splice(dragHost.indexOf(drag), 1);
            }
            this.$set(node, 'expanded', this.dragAfterExpanded);
            this.$emit('on-drag-node', { dragNode: drag, targetNode: node });
        },
        drag: function drag(node, ev) {
            var gid = this.gid();
            this.setDragNode(gid, node);
            ev.dataTransfer.setData('gid', gid);
        },
        dragover: function dragover(ev) {
            ev.preventDefault();
            ev.stopPropagation();
        },
        initTreeData: function initTreeData() {
            var _iteratorNormalCompletion5 = true;
            var _didIteratorError5 = false;
            var _iteratorError5 = undefined;

            try {
                for (var _iterator5 = this.data[Symbol.iterator](), _step5; !(_iteratorNormalCompletion5 = (_step5 = _iterator5.next()).done); _iteratorNormalCompletion5 = true) {
                    var node = _step5.value;

                    this.$set(node, 'parent', this.parent);
                    if (node.children && node.children.length) {
                        if (node.hasOwnProperty('disabled')) {
                            this.$delete(node, 'disabled');
                        }
                        if (node.hasOwnProperty('icon')) {
                            this.$delete(node, 'icon');
                        }
                    } else {
                        if (node.hasOwnProperty('openedIcon')) {
                            this.$delete(node, 'openedIcon');
                        }
                        if (node.hasOwnProperty('closedIcon')) {
                            this.$delete(node, 'closedIcon');
                        }
                    }
                    if (this.multiple) {
                        if (node.hasOwnProperty('selected')) {
                            this.$delete(node, 'selected');
                        }
                    } else {
                        if (node.hasOwnProperty('checked')) {
                            this.$delete(node, 'checked');
                        }
                    }
                }
            } catch (err) {
                _didIteratorError5 = true;
                _iteratorError5 = err;
            } finally {
                try {
                    if (!_iteratorNormalCompletion5 && _iterator5.return) {
                        _iterator5.return();
                    }
                } finally {
                    if (_didIteratorError5) {
                        throw _iteratorError5;
                    }
                }
            }
        },
        expandNode: function expandNode(node) {
            this.$set(node, 'expanded', !node.expanded);
            if (node.async && !node.children) {
                this.$emit('async-load-nodes', node);
            }
            if (node.children && node.children.length) {
                this.$emit('on-expanded', node, node.expanded);
            }
        },
        onExpanded: function onExpanded(node) {
            if (node.children && node.children.length) {
                this.$emit('on-expanded', node, node.expanded);
            }
        },
        triggerExpand: function triggerExpand(item) {
            if (!item.parent || item.children && item.children.length || item.async) {
                this.expandNode(item);
            }
        },
        asyncLoadNodes: function asyncLoadNodes(node) {
            if (node.async && !node.children) {
                this.$emit('async-load-nodes', node);
            }
        },
        isLeaf: function isLeaf(node) {
            return !(node.children && node.children.length) && node.parent && !node.async;
        },
        addNode: function addNode(parent, newNode) {
            var addnode = {};
            this.$set(parent, 'expanded', true);
            if (typeof newNode === 'undefined') {
                throw new ReferenceError('newNode is required but undefined');
            }
            if ((typeof newNode === 'undefined' ? 'undefined' : _typeof(newNode)) === 'object' && !newNode.hasOwnProperty('name')) {
                throw new ReferenceError('the name property is missed');
            }
            if ((typeof newNode === 'undefined' ? 'undefined' : _typeof(newNode)) === 'object' && !newNode.hasOwnProperty(this.nodeKey)) {
                throw new ReferenceError('the nodeKey property is missed');
            }
            if ((typeof newNode === 'undefined' ? 'undefined' : _typeof(newNode)) === 'object' && newNode.hasOwnProperty('name') && newNode.hasOwnProperty(this.nodeKey)) {
                addnode = Object.assign({}, newNode);
            }
            if (this.isLeaf(parent)) {
                this.$set(parent, 'children', []);
                parent.children.push(addnode);
            } else {
                parent.children.push(addnode);
            }
            this.$emit('addNode', { parentNode: parent, newNode: newNode });
        },
        addNodes: function addNodes(parent, newChildren) {
            var _iteratorNormalCompletion6 = true;
            var _didIteratorError6 = false;
            var _iteratorError6 = undefined;

            try {
                for (var _iterator6 = newChildren[Symbol.iterator](), _step6; !(_iteratorNormalCompletion6 = (_step6 = _iterator6.next()).done); _iteratorNormalCompletion6 = true) {
                    var n = _step6.value;

                    this.addNode(parent, n);
                }
            } catch (err) {
                _didIteratorError6 = true;
                _iteratorError6 = err;
            } finally {
                try {
                    if (!_iteratorNormalCompletion6 && _iterator6.return) {
                        _iterator6.return();
                    }
                } finally {
                    if (_didIteratorError6) {
                        throw _iteratorError6;
                    }
                }
            }
        },
        onClick: function onClick(node) {
            this.$emit('on-click', node);
        },
        onCheck: function onCheck(node, checked) {
            this.$emit('on-check', node, checked);
        },
        nodeCheckStatusChange: function nodeCheckStatusChange(node, checked) {
            this.$emit('dropTreeChecked', node, checked);
        },
        onDragNode: function onDragNode(event) {
            this.$emit('on-drag-node', event);
        },
        delNode: function delNode(parent, node) {
            if (parent === null || typeof parent === 'undefined') {
                if (this.isDeleteRoot) {
                    this.data.splice(0, 1);
                } else {
                    throw new ReferenceError('the root element can\'t deleted!');
                }
            } else {
                parent.children.splice(parent.children.indexOf(node), 1);
            }
            this.$emit('delNode', { parentNode: parent, delNode: node });
        },
        changeCheckStatus: function changeCheckStatus(node, $event) {
            this.$emit('on-check', node, $event.target.checked);
        },
        nodeSelected: function nodeSelected(node) {
            var getRoot = function getRoot(el) {
                if (el.$parent.$el.nodeName === 'UL') {
                    el = el.$parent;
                    return getRoot(el);
                }return el;
            };
            var root = getRoot(this);
            if (!this.multiple) {
                var _iteratorNormalCompletion7 = true;
                var _didIteratorError7 = false;
                var _iteratorError7 = undefined;

                try {
                    for (var _iterator7 = (root.data || [])[Symbol.iterator](), _step7; !(_iteratorNormalCompletion7 = (_step7 = _iterator7.next()).done); _iteratorNormalCompletion7 = true) {
                        var rn = _step7.value;

                        this.$set(rn, 'selected', false);
                        this.$emit('cancelSelected', root);
                    }
                } catch (err) {
                    _didIteratorError7 = true;
                    _iteratorError7 = err;
                } finally {
                    try {
                        if (!_iteratorNormalCompletion7 && _iterator7.return) {
                            _iterator7.return();
                        }
                    } finally {
                        if (_didIteratorError7) {
                            throw _iteratorError7;
                        }
                    }
                }
            }

            this.$set(node, 'selected', !node.selected);
            this.$emit('on-click', node);
        },
        nodeDataHandler: function nodeDataHandler(opt, data, keyParton) {
            data = data || this.data;
            var res = [];
            var keyValue = keyParton;
            var _iteratorNormalCompletion8 = true;
            var _didIteratorError8 = false;
            var _iteratorError8 = undefined;

            try {
                for (var _iterator8 = data[Symbol.iterator](), _step8; !(_iteratorNormalCompletion8 = (_step8 = _iterator8.next()).done); _iteratorNormalCompletion8 = true) {
                    var node = _step8.value;
                    var _iteratorNormalCompletion9 = true;
                    var _didIteratorError9 = false;
                    var _iteratorError9 = undefined;

                    try {
                        for (var _iterator9 = Object.entries(opt)[Symbol.iterator](), _step9; !(_iteratorNormalCompletion9 = (_step9 = _iterator9.next()).done); _iteratorNormalCompletion9 = true) {
                            var _ref = _step9.value;

                            var _ref2 = slicedToArray(_ref, 2);

                            var key = _ref2[0];
                            var value = _ref2[1];

                            if (node[key] === value) {
                                if (!keyValue.length || !keyValue) {
                                    var n = Object.assign({}, node);
                                    delete n['parent'];
                                    if (!(n.children && n.children.length)) {
                                        res.push(n);
                                    }
                                } else {
                                    var _n = {};
                                    if (Object.prototype.toString.call(keyValue) === '[object Array]') {
                                        for (var i = 0; i < keyValue.length; i++) {
                                            if (node.hasOwnProperty(keyValue[i])) {
                                                _n[keyValue[i]] = node[keyValue[i]];
                                            }
                                        }
                                    }
                                    if (Object.prototype.toString.call(keyValue) === '[object String]') {
                                        _n[keyValue] = node[keyValue];
                                    }
                                    if (!(node.children && node.children.length)) {
                                        res.push(_n);
                                    }
                                }
                            }
                        }
                    } catch (err) {
                        _didIteratorError9 = true;
                        _iteratorError9 = err;
                    } finally {
                        try {
                            if (!_iteratorNormalCompletion9 && _iterator9.return) {
                                _iterator9.return();
                            }
                        } finally {
                            if (_didIteratorError9) {
                                throw _iteratorError9;
                            }
                        }
                    }

                    if (node.children && node.children.length) {
                        res = res.concat(this.nodeDataHandler(opt, node.children, keyValue));
                    }
                }
            } catch (err) {
                _didIteratorError8 = true;
                _iteratorError8 = err;
            } finally {
                try {
                    if (!_iteratorNormalCompletion8 && _iterator8.return) {
                        _iterator8.return();
                    }
                } finally {
                    if (_didIteratorError8) {
                        throw _iteratorError8;
                    }
                }
            }

            return res;
        },
        getNode: function getNode(keyParton) {
            if (!this.multiple) {
                return this.nodeDataHandler({ selected: true }, this.data, keyParton);
            } else {
                return this.nodeDataHandler({ checked: true }, this.data, keyParton);
            }
        },
        searchNode: function searchNode(filter, data) {
            this.searchFlag = true;
            data = data || this.data;
            var _iteratorNormalCompletion10 = true;
            var _didIteratorError10 = false;
            var _iteratorError10 = undefined;

            try {
                for (var _iterator10 = data[Symbol.iterator](), _step10; !(_iteratorNormalCompletion10 = (_step10 = _iterator10.next()).done); _iteratorNormalCompletion10 = true) {
                    var node = _step10.value;

                    var searched = filter ? typeof filter === 'function' ? filter(node) : node['name'].indexOf(filter) > -1 : false;
                    this.$set(node, 'searched', searched);
                    this.$set(node, 'visible', false);
                    this.$emit('toggleshow', node, filter ? searched : true);
                    if (node.children && node.children.length) {
                        var _visibleStatus;

                        if (searched) this.$set(node, 'expanded', true);
                        (_visibleStatus = this.visibleStatus).splice.apply(_visibleStatus, [0, this.visibleStatus.length].concat([]));
                        this.searchNode(filter, node.children);
                    }
                }
            } catch (err) {
                _didIteratorError10 = true;
                _iteratorError10 = err;
            } finally {
                try {
                    if (!_iteratorNormalCompletion10 && _iterator10.return) {
                        _iterator10.return();
                    }
                } finally {
                    if (_didIteratorError10) {
                        throw _iteratorError10;
                    }
                }
            }
        }
    }
};

bkTree.install = function (Vue$$1) {
    Vue$$1.component(bkTree.name, bkTree);
};

var bkCollapse = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-collapse" }, [_vm._t("default")], 2);
    }, staticRenderFns: [],
    name: 'bkCollapse',
    props: {
        accordion: {
            type: Boolean,
            default: false
        },
        value: {
            type: [Array, String]
        }
    },
    watch: {
        value: function value(val) {
            this.currentValue = val;
        },
        currentValue: function currentValue() {
            this.setActive();
        }
    },
    data: function data() {
        return {
            currentValue: this.value
        };
    },


    methods: {
        setActive: function setActive() {
            var activeKey = this.getActiveKey();
            this.$children.forEach(function (child, index) {
                var name = child.name || index.toString();
                child.isActive = activeKey.indexOf(name) > -1;
                child.index = index;
            });
        },
        getActiveKey: function getActiveKey() {
            var activeKey = this.currentValue || [];
            var accordion = this.accordion;
            if (!Array.isArray(activeKey)) {
                activeKey = [activeKey];
            }
            if (accordion && activeKey.length > 1) {
                activeKey = [activeKey[0]];
            }
            for (var i = 0; i < activeKey.length; i++) {
                activeKey[i] = activeKey[i].toString();
            }
            return activeKey;
        },
        toggle: function toggle(data) {
            var name = data.name.toString();
            var newActiveKey = [];
            if (this.accordion) {
                if (!data.isActive) {
                    newActiveKey.push(name);
                }
            } else {
                var activeKey = this.getActiveKey();
                var nameIndex = activeKey.indexOf(name);
                if (data.isActive) {
                    if (nameIndex > -1) {
                        activeKey.splice(nameIndex, 1);
                    }
                } else {
                    if (nameIndex < 0) {
                        activeKey.push(name);
                    }
                }
                newActiveKey = activeKey;
            }
            this.currentValue = newActiveKey;
            this.$emit('input', newActiveKey);
            this.$emit('item-click', newActiveKey);
        }
    },
    mounted: function mounted() {
        this.setActive();
    }
};

bkCollapse.install = function (Vue$$1) {
    Vue$$1.component(bkCollapse.name, bkCollapse);
};

var Transition$1 = {
    beforeEnter: function beforeEnter(el) {
        addClass(el, 'collapse-transition');
        if (!el.dataset) {
            el.dataset = {};
        }

        el.dataset.oldPaddingTop = el.style.paddingTop;
        el.dataset.oldPaddingBottom = el.style.paddingBottom;

        el.style.height = '0';
        el.style.paddingTop = 0;
        el.style.paddingBottom = 0;
    },
    enter: function enter(el) {
        el.dataset.oldOverflow = el.style.overflow;
        if (el.scrollHeight !== 0) {
            el.style.height = el.scrollHeight + 'px';
            el.style.paddingTop = el.dataset.oldPaddingTop;
            el.style.paddingBottom = el.dataset.oldPaddingBottom;
        } else {
            el.style.height = '';
            el.style.paddingTop = el.dataset.oldPaddingTop;
            el.style.paddingBottom = el.dataset.oldPaddingBottom;
        }

        el.style.overflow = 'hidden';
    },
    afterEnter: function afterEnter(el) {
        removeClass(el, 'collapse-transition');
        el.style.height = '';
        el.style.overflow = el.dataset.oldOverflow;
    },
    beforeLeave: function beforeLeave(el) {
        if (!el.dataset) el.dataset = {};
        el.dataset.oldPaddingTop = el.style.paddingTop;
        el.dataset.oldPaddingBottom = el.style.paddingBottom;
        el.dataset.oldOverflow = el.style.overflow;

        el.style.height = el.scrollHeight + 'px';
        el.style.overflow = 'hidden';
    },
    leave: function leave(el) {
        if (el.scrollHeight !== 0) {
            addClass(el, 'collapse-transition');
            el.style.height = 0;
            el.style.paddingTop = 0;
            el.style.paddingBottom = 0;
        }
    },
    afterLeave: function afterLeave(el) {
        removeClass(el, 'collapse-transition');
        el.style.height = '';
        el.style.overflow = el.dataset.oldOverflow;
        el.style.paddingTop = el.dataset.oldPaddingTop;
        el.style.paddingBottom = el.dataset.oldPaddingBottom;
    }
};

var CollapseTransition$1 = {
    name: 'CollapseTransition',
    functional: true,
    render: function render(h, _ref) {
        var children = _ref.children;

        var data = {
            on: Transition$1
        };

        return h('transition', data, children);
    }
};

var bkCollapseItem = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-collapse-item", class: { 'bk-collapse-item-active': this.isActive } }, [_c('div', { staticClass: "bk-collapse-item-header", on: { "click": _vm.toggle } }, [_c('span', { staticClass: "fr", class: { 'collapse-expand': this.isActive } }, [!_vm.hideArrow ? _c('i', { staticClass: "bk-icon icon-angle-right" }) : _vm._e()]), _vm._v(" "), _vm._t("icon"), _vm._v(" "), _vm._t("default")], 2), _vm._v(" "), _c('collapse-transition', [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.isActive, expression: "isActive" }], staticClass: "bk-collapse-item-content" }, [_c('div', { staticClass: "bk-collapse-item-detail" }, [_vm._t("content")], 2)])])], 1);
    }, staticRenderFns: [], _scopeId: 'data-v-2d32dbf7',
    name: 'bkCollapseItem',
    components: { CollapseTransition: CollapseTransition$1 },
    props: {
        name: {
            type: String
        },
        hideArrow: {
            type: Boolean,
            default: false
        }
    },
    data: function data() {
        return {
            index: 0,
            isActive: false
        };
    },


    methods: {
        toggle: function toggle() {
            this.$parent.toggle({
                name: this.name || this.index,
                isActive: this.isActive
            });
        }
    }
};

bkCollapseItem.install = function (Vue$$1) {
    Vue$$1.component(bkCollapseItem.name, bkCollapseItem);
};

var bkRound = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-circle" }, [_c('svg', { attrs: { "width": _vm.radius, "height": _vm.radius, "viewBox": "0 0 100 100", "version": "1.1" } }, [_c('circle', { staticClass: "progress-background", attrs: { "cx": "50", "cy": "50", "r": "50", "fill": "transparent", "stroke-width": _vm.config.strokeWidth, "stroke": _vm.config.bgColor } }), _vm._v(" "), _c('circle', { staticClass: "progress-bar", class: 'circle' + _vm.config.index, attrs: { "cx": "50", "cy": "50", "r": "50", "fill": "transparent", "stroke-width": _vm.config.strokeWidth, "stroke": _vm.config.activeColor, "stroke-dasharray": _vm.dashArray, "stroke-dashoffset": _vm.dashOffset } })]), _vm._v(" "), _vm.numShow ? _c('div', { staticClass: "num", style: _vm.numStyle }, [_vm._v(" " + _vm._s(Math.round(_vm.percentFixed * 100)) + "% ")]) : _vm._e(), _vm._v(" "), _vm.title ? _c('div', { staticClass: "title", style: _vm.titleStyle }, [_vm._v(" " + _vm._s(_vm.title) + " ")]) : _vm._e()]);
    }, staticRenderFns: [],
    name: 'bk-round',
    props: {
        config: {
            type: Object,
            default: function _default() {
                return {
                    strokeWidth: 5,
                    bgColor: 'gray',
                    activeColor: 'green',
                    index: 0
                };
            }
        },
        percent: {
            type: Number,
            default: 0
        },
        title: {
            type: String
        },
        titleStyle: {
            type: Object,
            default: function _default() {
                return {
                    fontSize: '16px'
                };
            }
        },
        numShow: {
            type: Boolean,
            default: true
        },
        numStyle: {
            type: Object,
            default: function _default() {
                return {
                    fontSize: '16px'
                };
            }
        },
        radius: {
            type: String,
            default: '100px'
        }
    },
    computed: {
        dashOffset: function dashOffset() {
            return this.percentFixed > 1 ? false : (1 - this.percentFixed) * this.dashArray;
        },
        percentFixed: function percentFixed() {
            return Number(this.percent.toFixed(2));
        }
    },
    data: function data() {
        return {
            dashArray: Math.PI * 100
        };
    }
};

bkRound.install = function (Vue$$1) {
    Vue$$1.component(bkRound.name, bkRound);
};

var img = new Image();img.src = 'data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPCEtLSBHZW5lcmF0b3I6IEFkb2JlIElsbHVzdHJhdG9yIDIyLjEuMCwgU1ZHIEV4cG9ydCBQbHVnLUluIC4gU1ZHIFZlcnNpb246IDYuMDAgQnVpbGQgMCkgIC0tPgo8c3ZnIHZlcnNpb249IjEuMSIgaWQ9IuWbvuWxgl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiBzdHlsZT0iZW5hYmxlLWJhY2tncm91bmQ6bmV3IDAgMCAxMDI0IDEwMjQ7IiB4bWw6c3BhY2U9InByZXNlcnZlIj4KPHN0eWxlIHR5cGU9InRleHQvY3NzIj4KCS5zdDB7ZmlsbDojNzM3OTg3O30KPC9zdHlsZT4KPGc+Cgk8cmVjdCB4PSI1MTIiIHk9IjIyNCIgY2xhc3M9InN0MCIgd2lkdGg9Ijk2IiBoZWlnaHQ9IjY0Ii8+Cgk8cmVjdCB4PSI0MTYiIHk9IjI4OCIgY2xhc3M9InN0MCIgd2lkdGg9Ijk2IiBoZWlnaHQ9IjY0Ii8+Cgk8cmVjdCB4PSI1MTIiIHk9IjM1MiIgY2xhc3M9InN0MCIgd2lkdGg9Ijk2IiBoZWlnaHQ9IjY0Ii8+Cgk8cGF0aCBjbGFzcz0ic3QwIiBkPSJNNDE2LDY0MGgxOTJWNDgwaC05NnYtNjRoLTk2VjY0MHogTTQ2NCw1MjhoOTZ2NjRoLTk2VjUyOHoiLz4KCTxwYXRoIGNsYXNzPSJzdDAiIGQ9Ik05Niw5NnY4MzJoODMyVjk2SDk2eiBNODY0LDg2NEgxNjBWMTYwaDI1NnY2NGg5NnYtNjRoMzUyVjg2NHoiLz4KPC9nPgo8L3N2Zz4K';var uploadZip = img.src;

var img$1 = new Image();img$1.src = 'data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPCEtLSBHZW5lcmF0b3I6IEFkb2JlIElsbHVzdHJhdG9yIDIyLjEuMCwgU1ZHIEV4cG9ydCBQbHVnLUluIC4gU1ZHIFZlcnNpb246IDYuMDAgQnVpbGQgMCkgIC0tPgo8c3ZnIHZlcnNpb249IjEuMSIgaWQ9IuWbvuWxgl8xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiBzdHlsZT0iZW5hYmxlLWJhY2tncm91bmQ6bmV3IDAgMCAxMDI0IDEwMjQ7IiB4bWw6c3BhY2U9InByZXNlcnZlIj4KPHN0eWxlIHR5cGU9InRleHQvY3NzIj4KCS5zdDB7ZmlsbDojNzM3OTg3O30KPC9zdHlsZT4KPGc+Cgk8cGF0aCBjbGFzcz0ic3QwIiBkPSJNNzA0LDY0TDcwNCw2NEgxMjh2ODk2aDc2OFYyNTZsMCwwTDcwNCw2NHogTTcwNCwxNTQuNUw4MDUuNSwyNTZINzA0VjE1NC41eiBNODMyLDg5NkgxOTJWMTI4aDQ0OHYxOTJoMTkyCgkJVjg5NnoiLz4KCTxyZWN0IHg9IjI4OCIgeT0iMzIwIiBjbGFzcz0ic3QwIiB3aWR0aD0iMjU2IiBoZWlnaHQ9IjY0Ii8+Cgk8cmVjdCB4PSIyODgiIHk9IjQ0OCIgY2xhc3M9InN0MCIgd2lkdGg9IjQ0OCIgaGVpZ2h0PSI2NCIvPgoJPHJlY3QgeD0iMjg4IiB5PSI1NzYiIGNsYXNzPSJzdDAiIHdpZHRoPSI0NDgiIGhlaWdodD0iNjQiLz4KCTxyZWN0IHg9IjI4OCIgeT0iNzA0IiBjbGFzcz0ic3QwIiB3aWR0aD0iNDQ4IiBoZWlnaHQ9IjY0Ii8+CjwvZz4KPC9zdmc+Cg==';var uploadFile = img$1.src;

var img$2 = new Image();img$2.src = 'data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4KPCEtLSBHZW5lcmF0b3I6IEFkb2JlIElsbHVzdHJhdG9yIDIyLjEuMCwgU1ZHIEV4cG9ydCBQbHVnLUluIC4gU1ZHIFZlcnNpb246IDYuMDAgQnVpbGQgMCkgIC0tPgo8c3ZnIHZlcnNpb249IjEuMSIgaWQ9IuWbvuWxgl8yIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIKCSB2aWV3Qm94PSIwIDAgMTAyNCAxMDI0IiBzdHlsZT0iZW5hYmxlLWJhY2tncm91bmQ6bmV3IDAgMCAxMDI0IDEwMjQ7IiB4bWw6c3BhY2U9InByZXNlcnZlIj4KPHN0eWxlIHR5cGU9InRleHQvY3NzIj4KCS5zdDB7ZmlsbDojYzNjZGQ3O30KPC9zdHlsZT4KPGcgaWQ9IuS4iuS8oCI+Cgk8cGF0aCBjbGFzcz0ic3QwIiBkPSJNODk3LDQyMi45Yy0yOC40LTI4LjQtNjMuMy00Ny44LTEwMS4zLTU3Yy0xMC4xLTU3LjktMzcuNi0xMTEuMi04MC0xNTMuNkM2NjEuMywxNTgsNTg4LjksMTI4LDUxMiwxMjgKCQlzLTE0OS4zLDMwLTIwMy42LDg0LjRjLTQyLjQsNDIuNC02OS45LDk1LjctODAsMTUzLjZjLTM4LjEsOS4yLTczLDI4LjYtMTAxLjMsNTdjLTQwLjYsNDAuNi02Myw5NC42LTYzLDE1MnYxMAoJCWMwLDU3LjQsMjIuNCwxMTEuNCw2MywxNTJzOTQuNiw2MywxNTIsNjNoNDF2LTY0aC00MWMtODMuMywwLTE1MS02Ny43LTE1MS0xNTF2LTEwYzAtODMuMyw2Ny43LTE1MSwxNTEtMTUxaDkuMgoJCWMtMC4xLTIuNi0wLjItNS4zLTAuMi03LjlsMCwwbDAsMGMwLTEuNywwLTMuMywwLjEtNC45YzAtMC40LDAtMC44LDAtMS4zYzAtMS42LDAuMS0zLjIsMC4yLTQuN2MwLTAuNCwwLTAuNywwLjEtMS4xCgkJYzAuMS0xLjIsMC4xLTIuNSwwLjItMy43YzAtMC42LDAuMS0xLjIsMC4xLTEuOGMwLjEtMS40LDAuMi0yLjcsMC4zLTQuMWMwLjEtMC44LDAuMi0xLjYsMC4zLTIuNWMwLjEtMC42LDAuMS0xLjIsMC4yLTEuOQoJCWMxLjItMTAuMywzLjEtMjAuNCw1LjYtMzAuMmwwLDBjOS44LTM4LjQsMjkuOC03My42LDU4LjYtMTAyLjNDMzk1LjksMjE1LjMsNDUyLjIsMTkyLDUxMiwxOTJzMTE2LjEsMjMuMywxNTguNCw2NS42CgkJYzI4LjcsMjguNyw0OC43LDYzLjksNTguNiwxMDIuM2wwLDBjMi41LDkuOCw0LjQsMTkuOSw1LjYsMzAuMmMwLjEsMC42LDAuMSwxLjIsMC4yLDEuOWMwLjEsMC44LDAuMiwxLjYsMC4zLDIuNQoJCWMwLjEsMS40LDAuMiwyLjcsMC4zLDQuMWMwLDAuNiwwLjEsMS4yLDAuMSwxLjhjMC4xLDEuMiwwLjIsMi41LDAuMiwzLjdjMCwwLjQsMCwwLjcsMC4xLDEuMWMwLjEsMS42LDAuMSwzLjIsMC4yLDQuNwoJCWMwLDAuNCwwLDAuOCwwLDEuM2MwLDEuNiwwLjEsMy4zLDAuMSw0LjlsMCwwbDAsMGMwLDIuNy0wLjEsNS4zLTAuMiw3LjloOS4yYzgzLjMsMCwxNTEsNjcuNywxNTEsMTUxdjEwYzAsODMuMy02Ny43LDE1MS0xNTEsMTUxCgkJaC00MXY2NGg0MWM1Ny40LDAsMTExLjQtMjIuNCwxNTItNjNzNjMtOTQuNiw2My0xNTJ2LTEwQzk2MCw1MTcuNSw5MzcuNiw0NjMuNSw4OTcsNDIyLjl6Ii8+Cgk8cG9seWdvbiBjbGFzcz0ic3QwIiBwb2ludHM9IjM3Ni4yLDYwMi45IDQyMS41LDY0OC4xIDQ4MCw1ODkuNiA0ODAsODk2IDU0NCw4OTYgNTQ0LDU4OS42IDYwMi41LDY0OC4xIDY0Ny44LDYwMi45IDUxMiw0NjcuMSAJIi8+CjwvZz4KPC9zdmc+Cgo=';var uploadIcon = img$2.src;

var bkUpload = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-upload" }, [_c('div', { staticClass: "file-wrapper", class: { 'isdrag': _vm.isdrag } }, [_c('img', { staticClass: "upload-icon", attrs: { "src": _vm.uploadIcon } }), _vm._v(" "), _c('span', { staticClass: "drop-upload" }, [_vm._v(_vm._s(_vm.dragText))]), _vm._v(" "), _c('span', { staticClass: "click-upload" }, [_vm._v(_vm._s(_vm.clickText))]), _vm._v(" "), _c('input', { ref: "uploadel", attrs: { "accept": _vm.accept, "multiple": _vm.multiple, "type": "file" }, on: { "change": _vm.selectFile } })]), _vm._v(" "), _vm.tip ? _c('p', { staticClass: "tip" }, [_vm._v(_vm._s(_vm.tip))]) : _vm._e(), _vm._v(" "), _vm.fileList.length ? _c('div', { staticClass: "all-file" }, _vm._l(_vm.fileList, function (file, index) {
            return _c('div', { key: index }, [_c('div', { staticClass: "file-item" }, [_c('div', { staticClass: "file-icon" }, [_c('img', { attrs: { "src": _vm.getIcon(file) } })]), _vm._v(" "), !file.done ? _c('i', { staticClass: "bk-icon icon-close close-upload", on: { "click": function click($event) {
                        _vm.deleteFile(index, file);
                    } } }) : _vm._e(), _vm._v(" "), _c('div', { staticClass: "file-info" }, [_c('div', { staticClass: "file-name" }, [_c('span', [_vm._v(_vm._s(file.name))])]), _vm._v(" "), _c('div', { staticClass: "file-message" }, [_c('span', { directives: [{ name: "show", rawName: "v-show", value: !file.done && file.status === 'running', expression: "!file.done && file.status === 'running'" }], staticClass: "upload-speed" }, [_vm._v(_vm._s(_vm.speed) + _vm._s(_vm.unit))]), _vm._v(" "), _c('span', { directives: [{ name: "show", rawName: "v-show", value: !file.done, expression: "!file.done" }], staticClass: "file-size" }, [_vm._v(_vm._s(_vm.filesize(file.size)))]), _vm._v(" "), _c('span', { directives: [{ name: "show", rawName: "v-show", value: file.done, expression: "file.done" }], staticClass: "file-size done" }, [_vm._v(_vm._s(_vm.t('uploadFile.uploadDone')))])]), _vm._v(" "), _c('div', { staticClass: "progress-bar-wrapper" }, [_c('div', { staticClass: "progress-bar", class: { 'fail': file.errorMsg }, style: { width: file.progress } })])])]), _vm._v(" "), file.errorMsg ? _c('p', { staticClass: "error-msg" }, [_vm._v(_vm._s(file.errorMsg))]) : _vm._e()]);
        })) : _vm._e()]);
    }, staticRenderFns: [],
    name: 'bk-upload',
    mixins: [locale$1],
    props: {
        name: {
            type: String,
            default: 'upload_file'
        },
        multiple: {
            type: Boolean,
            default: true
        },
        accept: {
            type: String,
            default: '*'
        },
        delayTime: {
            type: Number,
            default: 0
        },
        url: {
            required: true,
            type: String
        },
        size: {
            type: [Number, Object],
            default: function _default() {
                return {
                    maxFileSize: 5,
                    maxImgSize: 1
                };
            }
        },
        handleResCode: {
            type: Function,
            default: function _default(res) {
                if (res.code === 0) {
                    return true;
                } else {
                    return false;
                }
            }
        },
        header: [Array, Object],
        tip: {
            type: String,
            default: ''
        },
        validateName: {
            type: RegExp
        },
        withCredentials: {
            type: Boolean,
            default: false
        }
    },
    data: function data() {
        return {
            dragText: this.t('uploadFile.drag'),
            clickText: this.t('uploadFile.click'),
            showDialog: true,
            fileList: [],
            width: 0,
            barEl: null,
            fileIndex: null,
            speed: 0,
            total: 0,
            unit: 'kb/s',
            isdrag: false,
            progress: 0,
            uploadIcon: uploadIcon
        };
    },

    watch: {
        'fileIndex': function fileIndex(val) {
            if (val !== null && val < this.fileList.length) {
                this.uploadFile(this.fileList[val], val);
            }
        }
    },
    methods: {
        filesize: function filesize(val) {
            var size = val / 1000;
            if (size < 1) {
                return val.toFixed(3) + ' KB';
            } else {
                var index = size.toString().indexOf('.');
                return size.toString().slice(0, index + 2) + ' MB';
            }
        },
        selectFile: function selectFile(e) {
            var _this = this;

            var files = Array.from(e.target.files);
            if (!files.length) return;
            files.forEach(function (file, i) {
                var fileObj = {
                    name: file.name,
                    originSize: file.size,
                    size: file.size / 1000,
                    maxFileSize: null,
                    maxImgSize: null,
                    type: file.type,
                    fileHeader: '',
                    origin: file,
                    base64: '',
                    status: '',
                    done: false,
                    responseData: '',
                    speed: null,
                    errorMsg: '',
                    progress: ''
                };
                var index = fileObj.type.indexOf('/');
                var type = fileObj.type.slice(0, index);
                var safariImageType = fileObj.type.indexOf('application/x-photoshop') > -1;
                fileObj.fileHeader = type;
                if (typeof _this.size === 'number') {
                    fileObj.maxFileSize = _this.size;
                    fileObj.maxImgSize = _this.size;
                } else {
                    fileObj.maxFileSize = _this.size.maxFileSize;
                    fileObj.maxImgSize = _this.size.maxImgSize;
                }
                if (type === 'image' || safariImageType) {
                    _this.handleImage(fileObj, file);
                }
                if ((type !== 'image' || !safariImageType) && fileObj.size > fileObj.maxFileSize * 1000) {
                    fileObj.errorMsg = fileObj.name + '\u6587\u4EF6\u4E0D\u80FD\u8D85\u8FC7' + fileObj.maxFileSize + 'MB';
                }
                if (_this.validateName) {
                    if (!_this.validateName.test(fileObj.name)) {
                        fileObj.errorMsg = '文件名不合法';
                    }
                }
                _this.fileList.push(fileObj);
            });
            var len = this.fileList.length;
            var fileIndex = this.fileIndex;
            if (len - 1 === fileIndex) {
                this.uploadFile(this.fileList[fileIndex], fileIndex);
            } else {
                this.fileIndex = 0;
            }
            e.target.value = '';
        },
        hideFileList: function hideFileList() {
            var _this2 = this;

            if (this.delayTime) {
                setTimeout(function () {
                    _this2.fileList = [];
                }, this.delayTime);
            }
        },
        uploadFile: function uploadFile$$1(fileObj) {
            var _this3 = this;

            if (fileObj.errorMsg) {
                this.fileIndex += 1;
                fileObj.progress = 100 + '%';
                return;
            }
            var formData = new FormData();
            var xhr = new XMLHttpRequest();
            formData.append(this.name, fileObj.origin);
            this.isdrag = false;
            fileObj.xhr = xhr;
            xhr.onreadystatechange = function () {
                if (xhr.readyState === 4) {
                    if (xhr.status === 200) {
                        try {
                            var response = JSON.parse(xhr.responseText);
                            if (_this3.handleResCode(response)) {
                                fileObj.done = true;
                                fileObj.responseData = response;
                                _this3.$emit('on-success', fileObj, _this3.fileList);
                            } else {
                                fileObj.errorMsg = response.message;
                                _this3.$emit('on-error', fileObj, _this3.fileList);
                            }
                        } catch (error) {
                            fileObj.progress = 100 + '%';
                            fileObj.errorMsg = error.message;
                        }
                    }
                    _this3.fileIndex += 1;
                    _this3.unit = 'kb/s';
                    _this3.total = 0;
                    fileObj.status = 'done';
                    if (_this3.fileIndex === _this3.fileList.length) {
                        _this3.$emit('on-done', _this3.fileList);
                        _this3.hideFileList();
                    }
                }
            };
            var uploadProgress = function uploadProgress(e) {
                if (e.lengthComputable) {
                    var percentComplete = Math.round(e.loaded * 100 / e.total);
                    var kb = Math.round(e.loaded / 1000);
                    fileObj.progress = percentComplete + '%';
                    _this3.speed = kb - _this3.total;
                    _this3.total = kb;
                    _this3.unit = 'kb/s';
                    if (_this3.speed > 1000) {
                        _this3.speed = Math.round(_this3.speed / 1000);
                        _this3.unit = 'mb/s';
                    }
                    _this3.$emit('on-progress', e, fileObj, _this3.fileList);
                }
                fileObj.status = 'running';
            };
            xhr.upload.addEventListener('progress', uploadProgress, false);
            xhr.withCredentials = this.withCredentials;
            xhr.open('POST', this.url, true);
            if (this.header) {
                if (Array.isArray(this.header)) {
                    this.header.forEach(function (head) {
                        var headerKey = _this3.header.name;
                        var headerVal = _this3.header.value;
                        xhr.setRequestHeader(headerKey, headerVal);
                    });
                } else {
                    var headerKey = this.header.name;
                    var headerVal = this.header.value;
                    xhr.setRequestHeader(headerKey, headerVal);
                }
            }
            xhr.send(formData);
        },
        handleImage: function handleImage(fileObj, file) {
            var _this4 = this;

            var isJPGPNG = /image\/(jpg|png|jpeg)$/.test(fileObj.type);
            if (!isJPGPNG) {
                fileObj.errorMsg = '只允许上传JPG|PNG|JPEG格式的图片';
                return false;
            }
            if (fileObj.size > fileObj.maxImgSize * 1000) {
                fileObj.errorMsg = '\u56FE\u7247\u5927\u5C0F\u4E0D\u80FD\u8D85\u8FC7' + fileObj.maxImgSize + 'MB';
                return false;
            }
            var reader = new FileReader();
            reader.onload = function (e) {
                _this4.smallImage(reader.result, fileObj);
            };
            reader.readAsDataURL(file);
            return true;
        },
        smallImage: function smallImage(result, fileObj) {
            var img = new Image();
            var canvas = document.createElement('canvas');
            var context = canvas.getContext('2d');
            img.onload = function () {
                var originWidth = img.width;
                var originHeight = img.height;
                var maxWidth = 42;
                var maxHeight = 42;
                var targetWidth = originWidth;
                var targetHeight = originHeight;
                if (originWidth > maxWidth || originHeight > maxHeight) {
                    if (originWidth / originHeight > maxWidth / maxHeight) {
                        targetWidth = maxWidth;
                        targetHeight = Math.round(maxWidth * (originHeight / originWidth));
                    } else {
                        targetWidth = maxWidth;
                        targetHeight = maxHeight;
                    }
                }
                canvas.width = targetWidth;
                canvas.height = targetHeight;
                context.clearRect(0, 0, targetWidth, targetHeight);
                context.drawImage(img, 0, 0, targetWidth, targetHeight);
                fileObj['base64'] = canvas.toDataURL();
            };
            img.src = result;
        },
        getIcon: function getIcon(file) {
            if (file.base64) {
                return file.base64;
            }
            var isZip = false;
            var zipType = ['zip', 'rar', 'tar', 'gz'];
            for (var i = 0; i < zipType.length; i++) {
                if (file.type.indexOf(zipType[i]) > -1) {
                    isZip = true;
                    break;
                }
            }
            if (isZip) {
                return uploadZip;
            } else {
                return uploadFile;
            }
        },
        deleteFile: function deleteFile(index, file) {
            if (file.xhr) {
                file.xhr.abort();
            }
            this.fileList.splice(index, 1);
            var len = this.fileList.length;
            if (!len) {
                this.fileIndex = null;
            }
            if (index === 0 && len) {
                this.fileIndex = 0;
                this.uploadFile(this.fileList[0]);
            }
        }
    },
    mounted: function mounted() {
        var _this5 = this;

        var uploadEl = this.$refs.uploadel;
        uploadEl.addEventListener('dragenter', function (e) {
            _this5.isdrag = true;
        });
        uploadEl.addEventListener('dragleave', function (e) {
            _this5.isdrag = false;
        });
        uploadEl.addEventListener('dragend', function (e) {
            _this5.isdrag = false;
        });
    }
};

bkUpload.install = function (Vue$$1) {
    Vue$$1.component(bkUpload.name, bkUpload);
};

var bkTimeline = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-timeline" }, [_c('ul', _vm._l(_vm.list, function (item, index) {
            return _c('li', { key: index, staticClass: "bk-timeline-dot", class: _vm.makeClass(item) }, [item.tag !== '' ? _c('div', { staticClass: "bk-timeline-time", domProps: { "innerHTML": _vm._s(item.tag) }, on: { "click": function click($event) {
                        _vm.toggle(item);
                    } } }) : _vm._e(), _vm._v(" "), _c('div', { staticClass: "bk-timeline-content" }, [_vm.isNode(item) ? [_vm._t('nodeContent' + index, [_vm._v(_vm._s(_vm.nodeContent(item, index)))])] : [_c('div', { attrs: { "title": _vm.computedTitle(item.content) }, domProps: { "innerHTML": _vm._s(item.content) } })]], 2)]);
        }))]);
    }, staticRenderFns: [],
    name: 'bk-timeline',
    props: {
        list: {
            type: Array,
            required: true
        },
        titleAble: {
            type: Boolean,
            default: false
        }
    },
    data: function data() {
        return {
            colorReg: /default|primary|warning|success|danger/
        };
    },

    methods: {
        toggle: function toggle(item) {
            this.$emit('select', item);
        },
        makeClass: function makeClass(item) {
            if (!item.type || !this.colorReg.test(item.type)) {
                return 'primary';
            }
            return item.type;
        },
        isNode: function isNode(data) {
            if (isVNode(data.content)) {
                return true;
            } else {
                return false;
            }
        },
        nodeContent: function nodeContent(data, index) {
            this.$slots['nodeContent' + index] = [data.content];
        },
        computedTitle: function computedTitle(str) {
            return this.titleAble ? str.replace(/<[^>]+>/g, '') : '';
        }
    }
};

bkTimeline.install = function (Vue$$1) {
    Vue$$1.component(bkTimeline.name, bkTimeline);
};

var bkProcess = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-process" }, [_c('ul', { style: { paddingBottom: _vm.paddingBottom + 'px' } }, _vm._l(_vm.dataList, function (item, index) {
            return _c('li', { key: index, class: { success: _vm.curProcess >= index + 1, current: item.isLoading && index === _vm.curProcess - 1 }, style: { cursor: _vm.controllables ? 'pointer' : '' }, on: { "click": function click($event) {
                        _vm.toggle(item, index);
                    } } }, [_vm._v(" " + _vm._s(item[_vm.displayKey]) + " "), item.isLoading && index === _vm.curProcess - 1 ? _c('div', { staticClass: "bk-spin-loading bk-spin-loading-mini bk-spin-loading-white" }, [_c('div', { staticClass: "rotate rotate1" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate2" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate3" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate4" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate5" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate6" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate7" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate8" })]) : _c('i', { staticClass: "bk-icon icon-check-1" }), _vm._v(" "), _c('dl', { directives: [{ name: "show", rawName: "v-show", value: item.steps && item.steps.length && _vm.showFlag, expression: "item.steps && item.steps.length && showFlag" }], ref: "stepsDom", refInFor: true, staticClass: "bk-process-step" }, _vm._l(item.steps, function (step, stepIndex) {
                return _c('dd', { key: stepIndex }, [_vm._v(" " + _vm._s(step[_vm.displayKey]) + " "), step.isLoading && index === _vm.curProcess - 1 ? _c('div', { staticClass: "bk-spin-loading bk-spin-loading-mini bk-spin-loading-primary steps-loading" }, [_c('div', { staticClass: "rotate rotate1" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate2" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate3" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate4" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate5" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate6" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate7" }), _vm._v(" "), _c('div', { staticClass: "rotate rotate8" })]) : _c('i', { staticClass: "bk-icon icon-check-1" })]);
            }))]);
        })), _vm._v(" "), _vm.toggleFlag ? _c('a', { staticClass: "bk-process-toggle", attrs: { "href": "javascript:;" }, on: { "click": _vm.toggleProcess } }, [_c('i', { staticClass: "bk-icon", class: _vm.showFlag ? 'icon-angle-up' : 'icon-angle-down' })]) : _vm._e()]);
    }, staticRenderFns: [],
    name: 'bk-process',
    props: {
        list: {
            type: Array,
            required: true
        },
        controllable: {
            type: Boolean,
            default: false
        },
        showSteps: {
            type: Boolean,
            default: false
        },
        curProcess: {
            type: Number,
            default: 0
        },
        displayKey: {
            type: String,
            required: true
        }
    },
    watch: {
        list: {
            handler: function handler(value) {
                this.initToggleFlag(value);
                this.dataList = [].concat(toConsumableArray(value));
                this.calculateMaxBottom(value);
            },
            deep: true
        },
        curProcess: function curProcess(newValue, oldValue) {
            if (newValue > this.list.length + 1) {
                return;
            }
            this.setParentProcessLoad(this.list);
        }
    },
    data: function data() {
        return {
            toggleFlag: false,
            showFlag: this.showSteps,
            dataList: this.list,
            controllables: this.controllable,
            paddingBottom: 0,
            maxBottom: 0,
            stepsClientHeight: 32
        };
    },
    created: function created() {
        this.setParentProcessLoad(this.list);
    },
    mounted: function mounted() {
        this.initToggleFlag(this.list);
        this.calculateMaxBottom(this.list);
        if (this.showFlag) {
            this.paddingBottom = this.maxBottom;
        } else {
            this.paddingBottom = 0;
        }
    },

    methods: {
        initToggleFlag: function initToggleFlag(list) {
            if (!list.length) {
                this.toggleFlag = false;
            } else {
                for (var i = 0; i < list.length; i++) {
                    if (list[i].steps && list[i].steps.length) {
                        this.toggleFlag = true;
                        break;
                    }
                }
            }
        },
        setParentProcessLoad: function setParentProcessLoad(list) {
            var _list;

            var dataList = [].concat(toConsumableArray(list));
            var curProcess = this.curProcess - 1 || 0;
            if (!dataList.length) {
                return;
            }
            if (curProcess === dataList.length) {
                this.$set(dataList[curProcess - 1], 'isLoading', false);
            } else {
                for (var i = 0; i < dataList.length; i++) {
                    var loadFlag = false;
                    if (dataList[curProcess].steps && dataList[curProcess].steps.length) {
                        var steps = dataList[curProcess].steps;
                        for (var j = 0; j < steps.length; j++) {
                            if (steps[j]['isLoading']) {
                                loadFlag = true;
                            }
                        }
                        if (loadFlag) {
                            if (curProcess > 0) {
                                this.$set(dataList[curProcess - 1], 'isLoading', false);
                            }
                            this.$set(dataList[curProcess], 'isLoading', true);
                        }
                    }
                }
            }
            (_list = this.list).splice.apply(_list, [0, this.list.length].concat(toConsumableArray(dataList)));
        },
        toggleProcess: function toggleProcess() {
            this.showFlag = !this.showFlag;
            if (this.showFlag) {
                this.paddingBottom = this.maxBottom;
            } else {
                this.paddingBottom = 0;
            }
        },
        calculateMaxBottom: function calculateMaxBottom(list) {
            var processList = [].concat(toConsumableArray(list));
            var stepsLengthList = [];
            if (!processList.length) {
                this.maxBottom = 0;
                return;
            }
            processList.forEach(function (item) {
                if (item.steps) {
                    stepsLengthList.push(item.steps.length);
                }
            });
            this.maxBottom = Math.max.apply(null, stepsLengthList) * this.stepsClientHeight;
        },
        toggle: function toggle(item, index) {
            if (!this.controllables) {
                return;
            }
            this.$emit('update:curProcess', index + 1);
            this.$emit('process-changed', index + 1, item);
        }
    }
};

bkProcess.install = function (Vue$$1) {
    Vue$$1.component(bkProcess.name, bkProcess);
};

var bkCombox = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { directives: [{ name: "clickoutside", rawName: "v-clickoutside", value: _vm.close, expression: "close" }], staticClass: "bk-combobox", class: [_vm.extCls, { 'open': _vm.open }] }, [_c('div', { staticClass: "bk-combobox-wrapper" }, [_c('input', { directives: [{ name: "model", rawName: "v-model", value: _vm.localValue, expression: "localValue" }], staticClass: "bk-combobox-input", class: { 'active': _vm.open }, attrs: { "placeholder": _vm.placeholder }, domProps: { "value": _vm.localValue }, on: { "input": [function ($event) {
                    if ($event.target.composing) {
                        return;
                    }_vm.localValue = $event.target.value;
                }, _vm.onInput] } }), _vm._v(" "), _vm.localValue.length > 0 ? _c('div', { staticClass: "bk-combobox-icon-clear", on: { "click": function click($event) {
                    _vm.localValue = '';
                } } }, [_vm._m(0)]) : _vm._e(), _vm._v(" "), _c('div', { staticClass: "bk-combobox-icon-box", on: { "click": _vm.openFn } }, [_c('i', { staticClass: "bk-icon icon-angle-down bk-combobox-icon" })])]), _vm._v(" "), _c('transition', { attrs: { "name": "toggle-slide" } }, [_c('div', { directives: [{ name: "show", rawName: "v-show", value: _vm.open, expression: "open" }], staticClass: "bk-combobox-list" }, [_vm.showList.length > 0 ? _c('ul', _vm._l(_vm.showList, function (item, index) {
            return _c('li', { staticClass: "bk-combobox-item", class: { 'bk-combobox-item-target': index === 0 && _vm.localValue.length > 0 }, on: { "click": function click($event) {
                        $event.stopPropagation();_vm.selectItem(item);
                    } } }, [_c('div', { staticClass: "text" }, [_vm._v(" " + _vm._s(item) + " ")])]);
        })) : _c('ul', [_c('li', { staticClass: "bk-combobox-item", attrs: { "disabled": "disabled" } }, [_c('div', { staticClass: "text" }, [_vm._v(" 无匹配数据 ")])])])])])], 1);
    }, staticRenderFns: [function () {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('span', { attrs: { "title": "clear" } }, [_c('i', { staticClass: "bk-icon icon-close" })]);
    }],
    name: 'bk-combobox',
    props: {
        placeholder: {
            type: String,
            default: ''
        },
        list: {
            type: Array
        },
        value: {
            type: String,
            required: true
        },
        extCls: {
            type: String
        }
    },
    computed: {
        showList: function showList() {
            if (this.localValue === '') {
                return this.list;
            } else {
                var newList = [];
                var _iteratorNormalCompletion = true;
                var _didIteratorError = false;
                var _iteratorError = undefined;

                try {
                    for (var _iterator = this.list[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
                        var item = _step.value;

                        if (item.indexOf(this.localValue) !== -1) {
                            newList.push(item);
                        }
                    }
                } catch (err) {
                    _didIteratorError = true;
                    _iteratorError = err;
                } finally {
                    try {
                        if (!_iteratorNormalCompletion && _iterator.return) {
                            _iterator.return();
                        }
                    } finally {
                        if (_didIteratorError) {
                            throw _iteratorError;
                        }
                    }
                }

                return newList;
            }
        }
    },
    data: function data() {
        return {
            open: false,
            localValue: this.value
        };
    },

    watch: {
        localValue: function localValue() {
            this.$emit('update:value', this.localValue);
        },
        value: function value(newVal) {
            this.localValue = newVal;
        }
    },
    directives: {
        clickoutside: clickoutside
    },
    methods: {
        selectItem: function selectItem(item) {
            this.localValue = item;
            this.close();

            this.$emit('update:value', this.localValue);
            this.$emit('item-selected', item);
        },
        openFn: function openFn() {
            if (!this.disabled) {
                this.open = !this.open;
                this.$emit('visible-toggle', this.open);
            }
        },
        close: function close() {
            this.open = false;
            this.$emit('visible-toggle', this.open);
        },
        onInput: function onInput() {
            this.open = true;
            this.$emit('update:value', this.localValue);
            this.$emit('input', this.localValue);
        }
    }
};

bkCombox.install = function (Vue$$1) {
    Vue$$1.component(bkCombox.name, bkCombox);
};

bkPagination.install = function (Vue$$1) {
    Vue$$1.component(bkPagination.name, bkPagination);
};

var img$3 = new Image();img$3.src = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAATQAAACvCAYAAABzVyKrAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyZpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNi1jMDY3IDc5LjE1Nzc0NywgMjAxNS8wMy8zMC0yMzo0MDo0MiAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIDIwMTUgKFdpbmRvd3MpIiB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOjMwOUZCOThENjZCMTExRTc5RkM5QUE4MjdBRUMyRDkyIiB4bXBNTTpEb2N1bWVudElEPSJ4bXAuZGlkOjMwOUZCOThFNjZCMTExRTc5RkM5QUE4MjdBRUMyRDkyIj4gPHhtcE1NOkRlcml2ZWRGcm9tIHN0UmVmOmluc3RhbmNlSUQ9InhtcC5paWQ6MzA5RkI5OEI2NkIxMTFFNzlGQzlBQTgyN0FFQzJEOTIiIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6MzA5RkI5OEM2NkIxMTFFNzlGQzlBQTgyN0FFQzJEOTIiLz4gPC9yZGY6RGVzY3JpcHRpb24+IDwvcmRmOlJERj4gPC94OnhtcG1ldGE+IDw/eHBhY2tldCBlbmQ9InIiPz7MAuyRAAAbzUlEQVR42uyd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';var Building = img$3.src;

var img$4 = new Image();img$4.src = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAATQAAACcCAYAAADxuVeoAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyZpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNi1jMDY3IDc5LjE1Nzc0NywgMjAxNS8wMy8zMC0yMzo0MDo0MiAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIDIwMTUgKFdpbmRvd3MpIiB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOjMwNzg4MkU0NjZEMjExRTdCNUNERTU3NzJFRUNCMTdBIiB4bXBNTTpEb2N1bWVudElEPSJ4bXAuZGlkOjMwNzg4MkU1NjZEMjExRTdCNUNERTU3NzJFRUNCMTdBIj4gPHhtcE1NOkRlcml2ZWRGcm9tIHN0UmVmOmluc3RhbmNlSUQ9InhtcC5paWQ6MzA3ODgyRTI2NkQyMTFFN0I1Q0RFNTc3MkVFQ0IxN0EiIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6MzA3ODgyRTM2NkQyMTFFN0I1Q0RFNTc3MkVFQ0IxN0EiLz4gPC9yZGY6RGVzY3JpcHRpb24+IDwvcmRmOlJERj4gPC94OnhtcG1ldGE+IDw/eHBhY2tldCBlbmQ9InIiPz6E8KPvAAAjWUlEQVR42uxd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';var notFound = img$4.src;

var img$5 = new Image();img$5.src = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAATQAAACYCAYAAABqKBW+AAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyZpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNi1jMDY3IDc5LjE1Nzc0NywgMjAxNS8wMy8zMC0yMzo0MDo0MiAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIDIwMTUgKFdpbmRvd3MpIiB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOkQ1QTNEM0FGNjZENDExRTc4QTYwOEQ2QTE1RUUyRTg3IiB4bXBNTTpEb2N1bWVudElEPSJ4bXAuZGlkOkQ1QTNEM0IwNjZENDExRTc4QTYwOEQ2QTE1RUUyRTg3Ij4gPHhtcE1NOkRlcml2ZWRGcm9tIHN0UmVmOmluc3RhbmNlSUQ9InhtcC5paWQ6RDVBM0QzQUQ2NkQ0MTFFNzhBNjA4RDZBMTVFRTJFODciIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6RDVBM0QzQUU2NkQ0MTFFNzhBNjA4RDZBMTVFRTJFODciLz4gPC9yZGY6RGVzY3JpcHRpb24+IDwvcmRmOlJERj4gPC94OnhtcG1ldGE+IDw/eHBhY2tldCBlbmQ9InIiPz5gqeqEAAAicElEQVR42uydCZgV1ZXHqx8NiCKboraArC6AIIvsawTBjUVQR00iGZeeVmcyxkyWmckkZo8zSSbGibbopxjH5YuCfIgIIqASBQQFdWhRBEQUFGTft+45v0vdzu3H21/V6/ea8/++01Vdr151ddWtf51z7lmKqqqqPIVCET7Ky8v7FBUV3SbP3DBZtvOfvfUir4pMLisrW6ZXKTsUKaEpFKETWVMhsAflWbshwW48iE+K3C7EtkevmhKaQpGPZNZSFgtEutavX9+74IILvPPOO89r0aKF+Xzr1q3exx9/7FVUVHhHjhxh07uy34hbbrllq149JTSFotbw6KOPdhRSaldZWVkUiUR2y/JT2fyUyPDmzZt7V1xxhXfqqafG/O7OnTu9l156yduxYwe/viIyWjS1Sr2q6aFYL4FCkZUG1kgWdwqB/aMQWFu7Xdar9zn55JO9sWPHeo0aNYp7nKZNm3pjxozxpk6d6u3bt2+kbJok8phe4fQQ0UugUGRMZh1l8ZbIf0FmEFarVq281q1be6effrp30kknmf2GDRuWkMwsTjnlFK9///721+/oFVaTU6HIFZm1lsUikdZoV4MHDzZEVlRUVGO/Q4cOeQ0aNEj5uGh2jz/+uHfw4EF+7Shm51q92qqhKRRhYwpkVlJS4k2YMMFr06bNcWQG0iEz80BGIh7H9NFVL3N6UB+aQpG6VtZQCKdMtKhvyK8XY0aOHj3aa9iwYaB/B5+bj3HyNw/IcqFoagf0DqiGplAEgkceeaSVLN4SMvsDZIY2dvHFF1f7yYKEo+ndIvKyyEYhtjv1LqRw7dSHplAk18xksVikR7Nmzbx+/foZE7O4OBwDh7i0jRs3Gn/atm3bvM2bN9uP/lM0tR/oHVFCUygyxuTJk+8ScvlvyOzqq68O3MRMhjVr1njz5s2zoSCXCKkt0LuiJqdCkRGESK5niWaWazIDHTt2NOatj2/rHVFCUygyf0gikZ4sCcuoLZAu5WOg3pH40FlOhSK5hmaW6bhnysvLk+4jpmONfe3vseDMfDbXO6IamkKRDd7hx+eff15rJ7Blyxa7uk5vh2poCkXGKCoqelq0s/5Llizxzj777JT8aIm0rXT3RTNcunSpPZcX9I6ohqZQZAwhlIdk8S6VMKZNm+atW7fOO3z4cA3CIVVp//79tgRQUH/X++KLL7xZs2Z5n332GZu2yrbf6h1J8PLRsA2FIjnKy8sJrJ0p0qOWTmFrJBIZW1pa+qbeDSW02ngAAjE3Uvk72R5DkfK1JjHzdhFSnzqLnOKbgSx2yrN0QEinRWVlZf0Azd1P5bjTi4uL77311ls36l1IDPWhKRQpQl4ch2Rxny8KJTRFEg0An2ZDeXD211WNNExtVaHQSYHcoJnIr0QqRJh/J3XlaocMRojMktW9Ivtk/VORR0XOrcPX5ByRJSKUqcbpfq38v6frUFFkZaKrDy10jaW1T2CdYuz2APfAO+aXOaYyFxe7M2WUjLlbNJcHE/2dfNBs0tDQGov8UOS7ItGlKohgXeEdq6mP/LUQtVWFmpx1FTiHiRvq1LJly6pBgwYV0STjk08+8RYvXszU/x3sVK9ePa93795ely5dTDkaGmYsX77cW7VqFQ/8A0IWLeTB/mUdsAa+JfILkeoKhrZkNUGrmzZtihw9erSXbEa+D6HL//6mQ3DvyHU4qsNKoRpa7WgsaCK/btq0adXEiROL3OqllISZMWOGqVBKN6CzzjrruGN89NFH3oIFC2zKzTh5mGcUqIZGofz7RUyG9ZlnnmlInBI5w4cPN63dgJAZpGbIjbirr776KjrdaJuv7c6D4OT/Xq0jTaEaWu6wQWT50KFDe0aXYj7jjDO8IUOGGBMzFpkBEpIJ4Fy4cOEUX0MpNOAn+40I1SqKGjdubJqAUD3isceONTRyE74hOX5HqGxx4MABQ3qQG7Jr1y6aWU70BRJd72hv84TgtuiQU0JThAc6YT/5wgvxs1WSaVddu3Y1UmCwfrK7RRpB2t27d/d69uzp0WyX6Heah1BfDJKLB8zvDh06GAG7d++uJjeIbv/+/bSNu8WXKiG49xyCe12u7T4dgkpoCkXGLgyR74nc5fl+sk6dOhmtzCUum+Sdbjke/I+dO3c2gilK13HIzfe/FR05cuQi2Q1hwuGQ73+b5xPcUvW/KaEpFKmC1KDnvGP+MtPtCDLDP0a/yRp2+IYNGRFaDeYsKjK9L5EePXoY/xuan/W/bdmypYGQ3nDZFfm5yE4huAVWgxNy+1BvmRKaQhEPEyAzJjnQnjApKyoqjFCdgioVzGgyIcCECPuxLSjgf+P4SN++fU2yuOt/27lzZ1PZbbwv+N8+82r6377QW6iEplBYjOIHs5Zt27aNJhNToQKxYCIk3Z6V6QASbd++vRGwZ8+eavOU5f79+1EPv+UL/reVDsG9JgS3R2+pElpeI1ZogabbBHJd6/umndGQoskEZ74lEpaU2cl1OWt8eJi/NkQE/5s9JyFf/G8XymYE/99h+Z8WW+1NZImMkyN6p5XQFCcG8Js1btGixXH+MoAzP5pMwuhnmQ5OO+00I8y+UmL7yy+/rNYot2zZUl+2DZHdkJ/CyUJwr3p/879V6C1XQstHtJaB2luW+FOIQq/ytY1LRAgJ2CQyx6+yoEhibqaqdUEk+QT8eSUlJUb69Olj/H/WZEaL2759+6my2xhf0Eg3en+bPYXgtKyPElqtgvACAj7vdq7BMhmoJEn/WKSNsy+J4nfKoJ2pwyUuRqRDaPkOfHvt2rUzAvbu3VvD/7Zv3z5mM77pCwRX4Zinr8pY2aVDQgktl/iZyPd5M7dq1apyx44dkd27d5OWY1JzxHSqOuOMM4poTCHmEdHu02XQ3i4D9WEdMjUh14VKIn2DnrV0AaHglwurU3kyYEaff/75RgDdzB3/G9kcXWQzQs/MI3JN3vL+NsGwWMbNYR0pOdJUTqRcTn9SAEfOe/IA1r/ssssqzznnHBKivdmzZ5v4qO7du1cOGDAg4lchNUniNMcQEJQ5RAbnIh02Na4p4RpTIbOxY8eG8jfef/99YwY2atTIa9KkifHJIZBobQP/G2Eo1v/Gum1754PZ0lkybv5OR4tqaKG8cEXmdO7c+XIhs3psIIZp5MiRhrgGDhxY4ykhXefgwYOHV6xY8SfvWGkbRU1cGqa5ST4nZAaYHUVw4PPCYeYSYoPk6FtpX0K5BKRKCApCd3Nyb9Ha0OAIUxHNnxSJk3WYKKGFhbdFxqxcudJDbNgGJs3QoUNjfoHUHUQREyPDJLRdu2K7o7AsCAdBIBBeSlZzg+BqaxaVXFXi8BDImPPzCrOwgBKa4oQzNwk068TLoGXLlqH8DY7dtGlTExiLayAe+Iw2cwjAsW/JjSVEk2v4refAXB0tSmiK/IcJ1yCYNixzDzJD0MiYHEDjQWtjPZH/FzOVeDcE4H9zCS5s/xsTB/v2mWIfn2vcmhKaog6Ym5iCEJAlEWYLMyU+6zNDiBfD8W5NTggO31oiWP8bTnuOxbkEcV6qnSmhKeqGucmECgHIXps2bWLuQw4nWgrmIkArcrUktKZMwbGs9gZwxltyY2knEmIBzY5zsueF/w2i5LyC8r8poSmhKQoL1P1vAaFATtGg0YtvclUDrQqSQwB+LUtuLLPxc/FdUq8QYB3yluCS+d+CPC+Oh3YKd3rHgm0VSmiKPIcJ18B/Fs/Ew4xL5OdCq3L9XGhGbpwZmlOm4FgIkxWcA+RqyQ3NLJPzQtDkkp0Xddn8zl3vlZWVfalDRQlNkf8YlcjchJAovAh5WCKJ1tiigVaFWD8XsWWWSLL1v/F9xPrfsjkvjsP/B1nG0t4cc/NlHSZKaKFBSwUFg/LycoKTB/BwJ0p3ws9lCcmaoa4ZSBHGeLCzmgidoKz/zZqB2frfos/LnhPLVP1v8ZLsHULT+DMlNEUBYJhIAzpXESeW8mArLvaaN29uBEBolkgQp8HycYjlf3MnGLIpFsl5uf63VM6L/zvW/44mR/s971iT6IU6VJTQFPmPQNKdbECuDcqN9nNF5UTWAH4uYr0QgJ/LJbhs/G/ueVn/mxv/xnlZ7S4apDz5/jnt+q6EpjiRCC0a+MwQ8iKtn8sSSap+LiqkWP+bJTgc+UH439zzijfzqeamEpqigFBeXo7TrCsPNA1PooF2gvMcMoFUMoXr52Im1frfLMGl6n9jxpFj2TizoM4rHnRCQAlNUVgw2QFMBsRKH7LFEM3gKi6u1pKQbP1crv8Nxz3EZk3UZP43u2/Q5+UC/56fjE4H93d1qCihKfIfCcM1/AfaAJLZvn27EYBvyo0zy6ZgIyRk+3IC18+VzP8W1nk52hlt8Sp1qCihKfLb3MQRZcptxwuojVfuB2Am4uNCgI0zg0QwCbNJGLf+N8xgG1rhxpklCqQN6rw03UkJTVFY6CZyFg5ya/pFA8e5Nb1w0icCRIO4fi63YGOmwJFvtS1AKpIb/xbGeUGYzHAqoSmhKQoHlyYyN4GbMG79XNaZT6hFPLh+LojB+rkskaQT7xYNQjiaNWtmJNvzgsg7dOhw3H5U0fUDcleJublBh4oSmiL/MTKRuRmNaD8X+Z0ukaTr53LjzIL0v6VzXmhrScxNDddQQlPkO8rLy1GRyBDIOP6MdCXE+rmYEbXaTyp+LsSPwj8uziwb/1s65xUvZEPDNZTQFIWFwTz7aDXZ5FFauAUbCQGxfi7r60rVz+U2THHjzIIoJBl9XmhysWqlYWoSeyfAdn1Nh4oSmqJAzM142hlmWjZaUrSfC7+WG2eWyM/lNkyxx3LN02wKNkafVyxQ+8w3U5dok2ElNEXqZh8RoJQMuUHkQpHGuT6HeIS2fv16EyoRVGMSvks1C1vRAu3InalM5OeK1TDFjTMLumGKhmsooSnSJzM88TNFetTaYCkuNmEZsQDR4MiPbkwSVGNg6+eiwkcmDVPwvVn/W9ANi3VCQAlNkb5mZsiMkAh6hDLTGFTKTjKsWbPGmzt3riGzWLOL+LJipR5FNwYOqjFJvIYpVntLtWFK9Hll0rAYrdTXBKlrtERHqxKaIjn+wZLZhAkTsorJygQbNmxIaG7i34Jc021MElRj4EQNU5Bk/jd7Xpk0LHa0s/llZWVHdagqoSmS40Z+oJnlmszchzYeoUEk3bp1S7sxievnCrthihtnlsl5oREnKbet5qYSmiJF0GEp5YDWIEEaE9oLfqd4JactarMxSSrnZf1v6Z4Xwb3nnHNOTO1OJwSU0BTpwzjLcuUzi2VuptsdPVFjklQaA8drTBJGw5RkDYvjBe5Cvn683Cdibq7WYaqEpshzJDM3U0VQjUnchilhNCyOdV6x+o66ZK/amRKaogCA9uI3zA283HYmjUnc88pVwxQ0sHjmrvrPlNAUBQTMPTQUKkzESsqGdCC8XDUmiYcwG6bEm+Xkf6e8EPyqhKaEpiggczNRMUdbGNFtDGwLIwbdmCSbhilBnJcLTF9/tvSdsrKybTpalNAUBUJo8eqfudVpYzUGtgnjQTYGdhumZNqwOIjz0mKOSmiKAoKtIAEBxOuOnqjcdqzGJG6cWZANUyA0l+DSbZiSyXk5EwJaLkgJTZHvsBUkCG2IF+TaqVOnlBsDQzKun8s2JrFxZtkUbLRdzN2GKWGeF6EdNl5OsKC8vHyvLFeJTBP5o5ige3QEKaEp8tDcTDS7Gd0Y2BZGTLcxSXRjYHxnQTRMyfa8CMSNZW47s5sWp4j09uVmIbgxQmof6ChSQlPkCZLlb0bDxobZmK0gGgMH0TAlm/NKVp124MCBXvfu3c0sK4nuixcvpqJHR/lohpBaT9XUlNAUeQAedmK8ML/QUjIaWHEaA1siSdXP5TZMsb6ubPJZUz0vW9EjFe0Vk5z18ePHe9OnT4fUOsnmO0Xu1dGkhKaoZdgZPCYDgghxAPEak2TaGDishik2/g2tK1b8Gr42NEm0Rht865Jl3759vVmzZvHrNUpoSmiKAjA3MdPQSoJsTBIdZ5Zuw5SgGxYnI/t414ZJFB8X6EhSQlPUMtyGufEe2nXr1hkNyy3YmG1jEuvnchuTZNoYOKjzSsXcjIYzI9xYR5MSmiIBMMtWrlzpffzxx8b0SVS4MAg8/fTTetHjYP78+UYsiWF+EsbStWtXvThKaIpUMG3atGozS5E/sLOcyIcffqgXRAlNkQogs9roKaBIDGZKMdP9sA29IEpoilRQWz0FFInBi6V9+/bG//f8889Xl+8+EVFeXt6nqKjotqqqqmGybOdP7qwXeVVkcllZ2TIlNIVBbfUUUKQG7k2/fv28OXPmnIhE1lQI7EFZvcHOUDsz1ef6cqvs96Qsbw8r6FgJrYBQGz0FFOkh6EKYBUJmLWWxQAisKxMkF1xwgXfeeedVx+mR98pEVkVFRdGRI0e+IZu6yXdGCKltVUI7wU0bRX4j6K7s+YaHH36449GjR9vJalEkEtldWVn5qaw/JdKV7IsrrrjiuJLlZJsgzAC/9NJLmOQXyeZnhNRGC6lVKqEpFIpcamC8SUnl+meRtna7m91BvN/YsWMT1pfDBzxmzBhv6tSpxA2OlE2TRB4L8lwjersUCkUCMmsvi6Uiv4fMICyyITCtSRezJcqHDRuWUrFMgp7xBfv4jpqcCoUiV2SGQ/B1kdbNmjXzBgwYYPqTRmdeELaSjjuEAOQ33niD9DV8aR3E7FyrGppCoQgbj0JmaGQTJ0702rZtGzONLF3fLulpTs5roKkVSmgKhSKWdnaJLC4lFGXUqFGBT3Y4BQDGyd+6VOQkJTSFQhEWvsmPiy66KKsGN/HgaHq3eMf6MmwUUrtTCU2hUISB4fxo165dKAfH5OzSpYuJWfOLiFJx83+E1LKqIaeTAgqFIhZMhDCTAWGAiQHEYs2aNd68efMIBfm+kNqssrKy11RDUygUQcHUpkpUSThIdOzY0bv44ovtr3epyalQKILEan7ksnoI6VI+BqrJqVAoggSNEbpT3412gLEgpqFZinlY4/dEsPvGgjPz2Vw1NIVCESQeweyE0ChcmQvQC9XHOtXQFApFYBBNao1oXH+orKz83uzZs73LL7/8uJaG0dpWIu0rGSg1tHTpUrNeVFT0gmpoCoUiaPwrpicNceg3+vrrrxttLVFP1UyIjOY2VOHwG9BslW2/VQ1Noch/FAsZ3Dd//vx9a9eu/UHQpXNC0NKOipY2TlbvFU3t2xUVFcUiYf7JrZFIZGxpaekXqqEpAsPq1au9jz76SC9EsOgmsmThwoV3CJn9i6zP9JO/8930PCLyXVmlhhkVN1Z5fkhHUBATk5pqfywuLu4uZPZmVm8MHWeKaDIjwBFs3LjRGz58eEH/P8RRYSpRMbV3795ez549c/r3p0yZcvbRo0d/In/35l69ehUTTPrJJ59QaeJy+fiDhx566B4xse4X0jiU58SGavZdX/IWqqEp4mLVqlXeq6++WtBk9vLLL5v/A7/P+++/n7sHKxLxhKweP3DgwNrDhw+XrlixopgyO23atPGuu+46U4ZH0Bh/kWgoaydPnvwj0dhKdNQpoSkCxLnnnmvy6wqd1CAwHM1oQxadO3cO7e/h3CbsYMmSJdVkKttuErJqSBT81VdfXV1mhyKHlKq+8sorTZFE2a+V7P9z+Wi9aMXM8N0s0kZHo5qcigBA9VFLZu6yUMxPS2b0yrTA1OzTp0/WGh9almhdZumGMYgJaRpB205HVKjAvLzwwgtN6elYQFtDmN3D2S6EWL+kpOQq+egqf5e1Tz755Nu7d+/+SIixQo7NNCDBWl8JGe6T7+7V0aqEpkgCSrsUKqlBNLNmzTKhABYQGf6zbBAdBV+vXj3vtttuq/6dUtQUQCTanQoVlKjG7EwF7IscPXq0RgFF+V86CJl1sBqgixdffNEsy8rKinTEKqEp6iCpoSXNnDnTjTg39et79OgR2N9A80IgsOjS05dddllWx4Yko+/BiBEjvG3btnm7du3y9u7d6xEThoaIFgoBKpTQFHWQ1HjQITN6QFoMHjzYmHxBYtKkSTn7n6gSi09ToYSmOIFIDc0FMtu+fXv1OQ8dOjTQSYAbb7xRB4MSmqKukFos4NeJ91kuQZyZJTNw/vnnBz6j2aRJk7y6J0899ZQSbQxo2IYiKQjbsFoZIKwDrS0fyMxqiqeddlr175zrypUr6/Q9waeGKFRDU9QhMgM46enIzcyfnRBYuHChcZzT5KO2QbgHBEu2Ag5+0KJFCxPW0bVr15RnQxVKaIo6TmYWzDpCam7IxqJFiwypZRuykY2Jh3+PmLjoyq9UrUCoN0ZpHoJtFUpoihCAb+y1114rGDKzIISC6HsIhDxUQI0tSK1fv35Zm3iZaGaWzAiuJYSkVatW5jOCfhcvXmw+g4QnTJhwXNiGIn2orqs4DrHIDD9VPpOZBaEOpBURgW+xfPly780338z5uRD9b8kMwmrfvr0hXYR1ttFViVCTDz74QAdevmlo5eXlHI/God8Q6R6JROrJ2/5DkWcbNWp0/6RJkzRVI89BtY1YZFZQg7q42AS5zp07tzqX87333jMaULaaWrrXEqCZ0YE8GmzjfObMmWP2DTpmTjW07MiMSgF/ZVVksEgTUblPETLrJeu/Pnjw4PInnnjifL3khYNCJDMLyGvUqFGmPZoF/qpcwk5QWDMzFkh5ArnsrqQaWmqaGVUCehOv07dvX6PyY6Js2rTJ+DHkhp27d+/emVOmTLnr0KFDzYTs6sv+61q2bPnmxIkTD+utyA8QmW7zBp22YoX5to5EvJEjR5oJA2YYu3Xrltfnmg40/ixck/PvIbNTTz3VlElhGt2ChF3eUDNmzPA2b97c6cCBAzOj3mKbhBD/raysbIrejvxAoROZC16qQ4YMMZJryMvazGQyAYDPLBb8OvomjCMd5Fugb10zOc3rgqoGLpm5Po1LLrnEOEfPPvtsE39DNLd/EzFVHxNSu0dvh6IugXEOqJFG4nw02Gbrp9l9FfmhoXVx/QGxwGzODTfccNx2atcvWLAAM+fHQmpzRVN7Q2+LIt+QiYnXpUsXM8HCLCa10pgcsM8ImhlhGzt37jRZDuyrqEVCE/JpJ4tfiIwQMZXu3FIq6Zg3xPgsW7aMmADqlSuhKfIOmZh4TEwQQmJj0ZjNjAZkxj7pxqBpLmeAhCZkRtG5xbgJok3LTMBsmhAaq4P0lhQWCFolMn/Pnj3G3XDWWWfFDFFwwaQDDzgJ5TjDzzzzTA//a9DfyQeQAUC8GalPWCN25hP/Gi/zTFOfNI8zWA3tXu4JM5nUnIpXYjhVUOXTWqZ6SwoDRMG//fbbpvEIhQ5drQR/0KBBg2Jq7ASb8h23OoZ1V2CSUWM/GoRbEBy7Y8eOGttpNMLfyXb8hQ0IixlWxFa+nThxog6iPCK0Ufz42te+5pJRxti8ebNdXau3JP9x+PDhGulFLqiiCgERJjF27FijSdnvPP/889XJ2dHAp/Tcc8+ZkJ9evXpVa2XUOXN7A7j49NNPzSwi6U5uff8woCZe3SY0oyMHkQrDm97O9Aie11uS/4CsXDJDQ8LUhKysSQWxTZ8+3cQj4oqAlNyZPsxSPoPoICYb+/bWW2+Z6H4mkYhh3L17d/V3SGsiDIjvbNiwwYwdjsl3rrrqqlD/ZzXx6jahzRcZSzUGKoNmUimAAW/bfjFwBbyGf6e3JP9BuI31+xCqQ81++3IjZ5I0I6thQVbRgAAxuaxJCjmh8UFQVmN3tHYDgmOvvfba6rHG2Jk9e7apZhG2dqao44Qmg/mHMviGrl+/vtkTTzwRxHlQ72VcWVnZVr0l+Q/MyK9//euGgKId8wMHDjRdj5jkiTZJmSnECR5dowxNjeNR7mfdunU1mn9Aevjk8NW6znOc6tdff73R4NINSs0GTIJkOvkVFLQ5SsCEVlpa+kF5eXl/79jkwKUimTjSsD/WiUwV+b2Q2Ta9HYWDRFo5wdP4zyAbNCkewObNm5sQhXhuCo5HmhImJLOmLBs3bmzIM15IAyZorsgMDZFuSzRjCWJ21U4OyLhP+7u225Qst2R5Dj/xz+GnQV6rsI4bpsnJyZLpO14fbUU88OCn+/DjW8NPlm/ApwfRQtK1HS5CeMzNN98MuX6cJencY8k1KPIJ67gpW4/62CkUyWHDSWLN7GaoEGSknUUR21+zIR20ZV9jvsdqVUGQWdDHVUJTKAIGfj7MX7cZSx5gbjakQxNjzPwgyCes46aLougW84qs3lBM2/HaJWnVVuv7P5Gn+VjeyIcyPG6VfasrCn6MHKeppYP169ebsJZu3bqta9GiBRntlZmQDoRja8WtWbPGe+WVV2zozD3pmolhHVcJrXYHKlX8KI3UI84uK0Sukpv6uRKaElomhMYM67PPPmsS2iORyB2lpaUPpks6Ke6eMvmEdVw1OWtfMzNkRozV6NGjjdMWYd1PzYHoXpR9G+oVK2wQ+2Zj7TJFJj60N954w5CZ4P3KysqHsyUzx9d1HPGkYiaGdVzV0Gqf0P5JFn+0zTCik7MJQSDtx89F/LYM5PtVQytMEILyzDPPmNnOAQMG5KzvJ7msfkYN4U4DZSy8k+FYrWEegiDMwrCOqxpa7cAk+CVrhuHuqyhMuI1WCASGaMLGO++8Y8hMyKJKTM1bgiIzfF1Itg78sI6bCbQvZ/Kbda4sfiMy0t/0isgPZVCtdnYz2dSpNMMQ9NSrWtggc4HgVrqzQzSYgZT4DqOvJiWHyFUVVMrx77j11lufDIrMLOy6r1Hdk078WFjHVQ0tHDKjSxWjaYJIE19Yf8v/LBNoQ5g6ACrM8gCTBkVVWirSxqskki15lpSU0P7xGiGzh4ImM5d80tWowjquElp4+JVIM2KQbrrpJiN+A9tm/mfVVgE/4pW5AbYZhncsjENRB8DDOn78eJOjSpltijhmC2Yy3QojYmKuGjdu3CDRbDKqRJMK6WRCPmEdV03OcGHMTLfuG+t//vOfWR3l7EecWX/MD/IY8Zm5uXpuMwx/X0UdARkE1113nSl22bNndt4EyjLRZ4A8zWuuueaIkNmfZPO/i2TUoNudhcQ5T+NlJA3ENBPDOq5qaLk1P4+LIXKAKfAus5iYHlSMsGCdbf4M53v+voo6BMxOJgrcCr2Edrz44oumAGUysA+z4PiaKGUupuvyRYsW0az7riDILEvU0KjCOm5Q0LCNxIOCSiATKPVs63o569PkDTPR2TdZYO27IldmE1hLXFsmjWgUuQPFJ9GyqMxBpDzAD8YMONVDXFA0Eoc/mpnlNtHKflZZWUlWSWXIY9soWJYHrDUR0LFtmFFRrq+/amiJgbq/wy1S6K/v8D+rhk9U/fy36jLno2X+tr6ZkJmPPfxI5KNT5AfwleJLEzJDs/qlyD4Ii9g1yAsfGVi9erXZ5pMZ95cOap1KS0sfCJvMckCUqVg1Smi5hgysVbKg7ts0ZzPr/f3Povc/JHKfrPZxNrN+X6Z5nD5e5ke8hrWK/ECUr3SO3PMfQVL169d/Soisiniyv/zlL6Yab0lJCWEeVWKu/q/sc57s+x8ie/QqZmlyPvjgg3oVUkMN9TyE/ROBcBGq+jYiG8E2rKXAoSI/zEy3cbDggAjdYaobETRo0KD/oUOHHpDVnsSrjR8/frmYnHfMnTt3cQGN6do+blLoLGdhgAdjiMhCeWAaxWpYq8gb7PfvVY2uKkJmENeASCTyu5NPPrmqZcuWd0+dOlVjEkPQ0KJnBYqiWDbh99PYV6FQKDLhlJR5Rn1oCoWizuD/BRgAudC6zWvii84AAAAASUVORK5CYII=';var permissions = img$5.src;

var img$6 = new Image();img$6.src = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAATQAAAC1CAYAAADRGgNWAAAAGXRFWHRTb2Z0d2FyZQBBZG9iZSBJbWFnZVJlYWR5ccllPAAAAyZpVFh0WE1MOmNvbS5hZG9iZS54bXAAAAAAADw/eHBhY2tldCBiZWdpbj0i77u/IiBpZD0iVzVNME1wQ2VoaUh6cmVTek5UY3prYzlkIj8+IDx4OnhtcG1ldGEgeG1sbnM6eD0iYWRvYmU6bnM6bWV0YS8iIHg6eG1wdGs9IkFkb2JlIFhNUCBDb3JlIDUuNi1jMDY3IDc5LjE1Nzc0NywgMjAxNS8wMy8zMC0yMzo0MDo0MiAgICAgICAgIj4gPHJkZjpSREYgeG1sbnM6cmRmPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5LzAyLzIyLXJkZi1zeW50YXgtbnMjIj4gPHJkZjpEZXNjcmlwdGlvbiByZGY6YWJvdXQ9IiIgeG1sbnM6eG1wPSJodHRwOi8vbnMuYWRvYmUuY29tL3hhcC8xLjAvIiB4bWxuczp4bXBNTT0iaHR0cDovL25zLmFkb2JlLmNvbS94YXAvMS4wL21tLyIgeG1sbnM6c3RSZWY9Imh0dHA6Ly9ucy5hZG9iZS5jb20veGFwLzEuMC9zVHlwZS9SZXNvdXJjZVJlZiMiIHhtcDpDcmVhdG9yVG9vbD0iQWRvYmUgUGhvdG9zaG9wIENDIDIwMTUgKFdpbmRvd3MpIiB4bXBNTTpJbnN0YW5jZUlEPSJ4bXAuaWlkOkVDMTA0MTVCQjI1NTExRTc5RTVFOTYzQzZBNkI5NTJEIiB4bXBNTTpEb2N1bWVudElEPSJ4bXAuZGlkOkVDMTA0MTVDQjI1NTExRTc5RTVFOTYzQzZBNkI5NTJEIj4gPHhtcE1NOkRlcml2ZWRGcm9tIHN0UmVmOmluc3RhbmNlSUQ9InhtcC5paWQ6RUMxMDQxNTlCMjU1MTFFNzlFNUU5NjNDNkE2Qjk1MkQiIHN0UmVmOmRvY3VtZW50SUQ9InhtcC5kaWQ6RUMxMDQxNUFCMjU1MTFFNzlFNUU5NjNDNkE2Qjk1MkQiLz4gPC9yZGY6RGVzY3JpcHRpb24+IDwvcmRmOlJERj4gPC94OnhtcG1ldGE+IDw/eHBhY2tldCBlbmQ9InIiPz69EWjoAAAtaUlEQVR42uyd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';var maintain = img$6.src;

var bkException = { render: function render() {
        var _vm = this;var _h = _vm.$createElement;var _c = _vm._self._c || _h;return _c('div', { staticClass: "bk-exception" }, [_c('img', { attrs: { "src": _vm.images[_vm.type], "alt": _vm.type } }), _vm._v(" "), _c('div', { staticClass: "exception-text" }, [_vm._t("default", [_vm._v(_vm._s(_vm.tipText[_vm.type]))])], 2)]);
    }, staticRenderFns: [],
    name: 'bk-exception',
    props: {
        type: {
            type: [String, Number],
            default: 404,
            validator: function validator(value) {
                return [404, 403, 500, '404', '403', '500', 'building'].indexOf(value) > -1;
            }
        }
    },
    data: function data() {
        return {
            images: {
                403: permissions,
                404: notFound,
                500: maintain,
                building: Building
            },
            tipText: {
                404: '页面找不到了！',
                403: 'Sorry，您的权限不足',
                500: '服务维护中，请稍后...',
                building: '功能正在建设中···'
            }
        };
    }
};

bkException.install = function (Vue$$1) {
    Vue$$1.component(bkException.name, bkException);
};

var enUS = {
    lang: 'en-US',
    datePicker: {
        selectDate: 'Select Date',

        topBarFormatView: '{mmmm} {yyyy}',
        weekdays: {
            sun: 'Sun',
            mon: 'Mon',
            tue: 'Tue',
            wed: 'Wed',
            thu: 'Thu',
            fri: 'Fri',
            sat: 'Sat'
        },
        today: 'Today'
    },
    dateRange: {
        selectDate: 'Select Date',
        datePicker: {
            topBarFormatView: '{mmmm} {yyyy}',
            weekdays: {
                sun: 'Sun',
                mon: 'Mon',
                tue: 'Tue',
                wed: 'Wed',
                thu: 'Thu',
                fri: 'Fri',
                sat: 'Sat'
            },
            today: 'Today'
        },
        yestoday: 'Yestoday',
        lastweek: 'Last week',
        lastmonth: 'Last month',
        last3months: 'Last 3 months',
        ok: 'OK',
        clear: 'CLEAR'
    },
    dialog: {
        title: 'This is the title',
        content: 'This is the content',
        ok: 'Confirm',
        cancel: 'Cancel'
    },
    selector: {
        pleaseselect: 'Please select',
        emptyText: 'No data',
        searchEmptyText: 'Unmatched data'
    },
    infobox: {
        title: 'This is the title',
        ok: 'Confirm',
        cancel: 'Cancel',
        pleasewait: 'Please wait',
        success: 'Success',
        continue: 'Continue',
        failure: 'Failure',
        closeafter3s: 'Close this window after 3s',
        riskoperation: 'This operation is at risk'
    },
    message: {
        close: 'CLOSE'
    },
    sideslider: {
        title: 'title'
    },
    steps: {
        step1: 'Step1',
        step2: 'Step2',
        step3: 'Step3'
    },
    uploadFile: {
        drag: 'Try dragging an file here or',
        click: 'click to upload',
        uploadDone: 'Upload finished'
    }
};

var langPkg = {
    enUS: enUS,
    zhCN: defaultLang
};

var bkTab$1 = Tab.bkTab;
var bkTabpanel$1 = Tab.bkTabpanel;


var components = {
    bkBadge: bkBadge,
    bkButton: bkButton,
    bkCollapse: bkCollapse,
    bkCollapseItem: bkCollapseItem,
    bkCombox: bkCombox,
    bkDatePicker: bkDatePicker,
    bkDateRange: bkDateRange,
    bkDialog: bkDialog,
    bkDropdown: bkDropdown,
    bkDropdownMenu: bkDropdownMenu,
    bkIconButton: bkIconButton,
    bkPaging: bkPaging,
    bkProcess: bkProcess,
    bkRound: bkRound,
    bkSelector: bkSelector,
    bkSideslider: bkSideslider,
    bkSteps: bkSteps,
    bkSwitcher: Switcher,
    bkTab: bkTab$1,
    bkTabpanel: bkTabpanel$1,
    bkTagInput: TagInpute,

    bkTimeline: bkTimeline,
    bkTooltip: bkTooltip,
    bkTransfer: bkTransfer,
    bkTree: bkTree,
    bkUpload: bkUpload,
    bkException: bkException,
    bkPagination: bkPagination
};

function install$1(Vue$$1) {
    var opts = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};

    locale.use(opts.locale);
    locale.i18n(opts.i18n);

    Object.keys(components).forEach(function (key) {
        Vue$$1.component(components[key].name, components[key]);
    });
}

Vue.use(bkLoading.directive);
Vue.prototype.$bkLoading = bkLoading.Loading;

Vue.prototype.$bkMessage = Msg;

Vue.prototype.$bkInfo = Info;

Vue.use(bkTooltips.directive);

Vue.prototype.$tooltips = bkTooltips.tooltips;

if (typeof window !== 'undefined' && window.Vue) {
    install$1(window.Vue);
}

var bkMagic = _extends({
    bkException: bkException
}, components, {
    bkLoading: bkLoading,
    bkMessage: Msg,
    bkTooltips: bkTooltips,
    bkInfoBox: Info,
    locale: locale,
    langPkg: langPkg,
    localeMixin: locale$1,
    install: install$1
});

exports.install = install$1;
exports.bkBadge = bkBadge;
exports.bkButton = bkButton;
exports.bkCollapse = bkCollapse;
exports.bkCollapseItem = bkCollapseItem;
exports.bkCombox = bkCombox;
exports.bkDatePicker = bkDatePicker;
exports.bkDateRange = bkDateRange;
exports.bkDialog = bkDialog;
exports.bkDropdown = bkDropdown;
exports.bkDropdownMenu = bkDropdownMenu;
exports.bkIconButton = bkIconButton;
exports.bkPaging = bkPaging;
exports.bkProcess = bkProcess;
exports.bkRound = bkRound;
exports.bkSelector = bkSelector;
exports.bkSideslider = bkSideslider;
exports.bkSteps = bkSteps;
exports.bkSwitcher = Switcher;
exports.bkTab = bkTab$1;
exports.bkTabpanel = bkTabpanel$1;
exports.bkTagInput = TagInpute;
exports.bkTimeline = bkTimeline;
exports.bkTooltip = bkTooltip;
exports.bkTransfer = bkTransfer;
exports.bkTree = bkTree;
exports.bkUpload = bkUpload;
exports.bkLoading = bkLoading;
exports.bkMessage = Msg;
exports.bkTooltips = bkTooltips;
exports.bkInfoBox = Info;
exports.locale = locale;
exports.bkException = bkException;
exports.langPkg = langPkg;
exports.localeMixin = locale$1;
exports.bkPagination = bkPagination;
exports.default = bkMagic;

Object.defineProperty(exports, '__esModule', { value: true });

})));
//# sourceMappingURL=bk-magic-vue.js.map
