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
(function (global, factory) {
  typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports, require('vue'), require('@babel/runtime/helpers/slicedToArray'), require('@babel/runtime/helpers/initializerDefineProperty'), require('@babel/runtime/helpers/classCallCheck'), require('@babel/runtime/helpers/createClass'), require('@babel/runtime/helpers/assertThisInitialized'), require('@babel/runtime/helpers/inherits'), require('@babel/runtime/helpers/possibleConstructorReturn'), require('@babel/runtime/helpers/getPrototypeOf'), require('@babel/runtime/helpers/applyDecoratedDescriptor'), require('@babel/runtime/helpers/initializerWarningHelper'), require('resize-detector'), require('@babel/runtime/helpers/toConsumableArray'), require('@babel/runtime/helpers/defineProperty'), require('@babel/runtime/helpers/asyncToGenerator'), require('@babel/runtime/regenerator'))
    : typeof define === 'function' && define.amd ? define(['exports', 'vue', '@babel/runtime/helpers/slicedToArray', '@babel/runtime/helpers/initializerDefineProperty', '@babel/runtime/helpers/classCallCheck', '@babel/runtime/helpers/createClass', '@babel/runtime/helpers/assertThisInitialized', '@babel/runtime/helpers/inherits', '@babel/runtime/helpers/possibleConstructorReturn', '@babel/runtime/helpers/getPrototypeOf', '@babel/runtime/helpers/applyDecoratedDescriptor', '@babel/runtime/helpers/initializerWarningHelper', 'resize-detector', '@babel/runtime/helpers/toConsumableArray', '@babel/runtime/helpers/defineProperty', '@babel/runtime/helpers/asyncToGenerator', '@babel/runtime/regenerator'], factory)
      : (global = typeof globalThis !== 'undefined' ? globalThis : global || self, factory(global['ip-selector'] = {}, global.Vue, global._slicedToArray, global._initializerDefineProperty, global._classCallCheck, global._createClass, global._assertThisInitialized, global._inherits, global._possibleConstructorReturn, global._getPrototypeOf, global._applyDecoratedDescriptor, null, global.resizeDetector, global._toConsumableArray$1, global._defineProperty$1, global._asyncToGenerator, global._regeneratorRuntime));
}(this, ((exports, Vue, _slicedToArray, _initializerDefineProperty, _classCallCheck, _createClass, _assertThisInitialized, _inherits, _possibleConstructorReturn, _getPrototypeOf, _applyDecoratedDescriptor, initializerWarningHelper, resizeDetector, _toConsumableArray$1, _defineProperty$1, _asyncToGenerator, _regeneratorRuntime) => {
  'use strict';

  function _interopDefaultLegacy(e) {
    return e && typeof e === 'object' && 'default' in e ? e : { default: e };
  }

  const Vue__default = /* #__PURE__*/_interopDefaultLegacy(Vue);
  const _slicedToArray__default = /* #__PURE__*/_interopDefaultLegacy(_slicedToArray);
  const _initializerDefineProperty__default = /* #__PURE__*/_interopDefaultLegacy(_initializerDefineProperty);
  const _classCallCheck__default = /* #__PURE__*/_interopDefaultLegacy(_classCallCheck);
  const _createClass__default = /* #__PURE__*/_interopDefaultLegacy(_createClass);
  const _assertThisInitialized__default = /* #__PURE__*/_interopDefaultLegacy(_assertThisInitialized);
  const _inherits__default = /* #__PURE__*/_interopDefaultLegacy(_inherits);
  const _possibleConstructorReturn__default = /* #__PURE__*/_interopDefaultLegacy(_possibleConstructorReturn);
  const _getPrototypeOf__default = /* #__PURE__*/_interopDefaultLegacy(_getPrototypeOf);
  const _applyDecoratedDescriptor__default = /* #__PURE__*/_interopDefaultLegacy(_applyDecoratedDescriptor);
  const _toConsumableArray__default = /* #__PURE__*/_interopDefaultLegacy(_toConsumableArray$1);
  const _defineProperty__default = /* #__PURE__*/_interopDefaultLegacy(_defineProperty$1);
  const _asyncToGenerator__default = /* #__PURE__*/_interopDefaultLegacy(_asyncToGenerator);
  const _regeneratorRuntime__default = /* #__PURE__*/_interopDefaultLegacy(_regeneratorRuntime);

  /**
    * vue-class-component v7.2.6
    * (c) 2015-present Evan You
    * @license MIT
    */

  function _typeof(obj) {
    if (typeof Symbol === 'function' && typeof Symbol.iterator === 'symbol') {
      _typeof = function (obj) {
        return typeof obj;
      };
    } else {
      _typeof = function (obj) {
        return obj && typeof Symbol === 'function' && obj.constructor === Symbol && obj !== Symbol.prototype ? 'symbol' : typeof obj;
      };
    }

    return _typeof(obj);
  }

  function _defineProperty(obj, key, value) {
    if (key in obj) {
      Object.defineProperty(obj, key, {
        value,
        enumerable: true,
        configurable: true,
        writable: true,
      });
    } else {
      obj[key] = value;
    }

    return obj;
  }

  function _toConsumableArray(arr) {
    return _arrayWithoutHoles(arr) || _iterableToArray(arr) || _nonIterableSpread();
  }

  function _arrayWithoutHoles(arr) {
    if (Array.isArray(arr)) {
      for (var i = 0, arr2 = new Array(arr.length); i < arr.length; i++) arr2[i] = arr[i];

      return arr2;
    }
  }

  function _iterableToArray(iter) {
    if (Symbol.iterator in Object(iter) || Object.prototype.toString.call(iter) === '[object Arguments]') return Array.from(iter);
  }

  function _nonIterableSpread() {
    throw new TypeError('Invalid attempt to spread non-iterable instance');
  }

  // The rational behind the verbose Reflect-feature check below is the fact that there are polyfills
  // which add an implementation for Reflect.defineMetadata but not for Reflect.getOwnMetadataKeys.
  // Without this check consumers will encounter hard to track down runtime errors.
  function reflectionIsSupported() {
    return typeof Reflect !== 'undefined' && Reflect.defineMetadata && Reflect.getOwnMetadataKeys;
  }
  function copyReflectionMetadata(to, from) {
    forwardMetadata(to, from);
    Object.getOwnPropertyNames(from.prototype).forEach((key) => {
      forwardMetadata(to.prototype, from.prototype, key);
    });
    Object.getOwnPropertyNames(from).forEach((key) => {
      forwardMetadata(to, from, key);
    });
  }

  function forwardMetadata(to, from, propertyKey) {
    const metaKeys = propertyKey ? Reflect.getOwnMetadataKeys(from, propertyKey) : Reflect.getOwnMetadataKeys(from);
    metaKeys.forEach((metaKey) => {
      const metadata = propertyKey ? Reflect.getOwnMetadata(metaKey, from, propertyKey) : Reflect.getOwnMetadata(metaKey, from);

      if (propertyKey) {
        Reflect.defineMetadata(metaKey, metadata, to, propertyKey);
      } else {
        Reflect.defineMetadata(metaKey, metadata, to);
      }
    });
  }

  const fakeArray = {
    __proto__: [],
  };
  const hasProto = fakeArray instanceof Array;
  function createDecorator(factory) {
    return function (target, key, index) {
      const Ctor = typeof target === 'function' ? target : target.constructor;

      if (!Ctor.__decorators__) {
        Ctor.__decorators__ = [];
      }

      if (typeof index !== 'number') {
        index = undefined;
      }

      Ctor.__decorators__.push(options => factory(options, key, index));
    };
  }
  function isPrimitive(value) {
    const type = _typeof(value);

    return value == null || type !== 'object' && type !== 'function';
  }

  function collectDataFromConstructor(vm, Component) {
    // override _init to prevent to init as Vue instance
    const originalInit = Component.prototype._init;

    Component.prototype._init = function () {
      const _this = this;

      // proxy to actual vm
      const keys = Object.getOwnPropertyNames(vm); // 2.2.0 compat (props are no longer exposed as self properties)

      if (vm.$options.props) {
        for (const key in vm.$options.props) {
          if (!vm.hasOwnProperty(key)) {
            keys.push(key);
          }
        }
      }

      keys.forEach((key) => {
        Object.defineProperty(_this, key, {
          get: function get() {
            return vm[key];
          },
          set: function set(value) {
            vm[key] = value;
          },
          configurable: true,
        });
      });
    }; // should be acquired class property values


    const data = new Component(); // restore original _init to avoid memory leak (#209)

    Component.prototype._init = originalInit; // create plain data object

    const plainData = {};
    Object.keys(data).forEach((key) => {
      if (data[key] !== undefined) {
        plainData[key] = data[key];
      }
    });

    return plainData;
  }

  const $internalHooks = ['data', 'beforeCreate', 'created', 'beforeMount', 'mounted', 'beforeDestroy', 'destroyed', 'beforeUpdate', 'updated', 'activated', 'deactivated', 'render', 'errorCaptured', 'serverPrefetch', // 2.6
  ];
  function componentFactory(Component) {
    const options = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};
    options.name = options.name || Component._componentTag || Component.name; // prototype props.

    const proto = Component.prototype;
    Object.getOwnPropertyNames(proto).forEach((key) => {
      if (key === 'constructor') {
        return;
      } // hooks


      if ($internalHooks.indexOf(key) > -1) {
        options[key] = proto[key];
        return;
      }

      const descriptor = Object.getOwnPropertyDescriptor(proto, key);

      if (descriptor.value !== void 0) {
        // methods
        if (typeof descriptor.value === 'function') {
          (options.methods || (options.methods = {}))[key] = descriptor.value;
        } else {
          // typescript decorated data
          (options.mixins || (options.mixins = [])).push({
            data: function data() {
              return _defineProperty({}, key, descriptor.value);
            },
          });
        }
      } else if (descriptor.get || descriptor.set) {
        // computed properties
        (options.computed || (options.computed = {}))[key] = {
          get: descriptor.get,
          set: descriptor.set,
        };
      }
    });
    (options.mixins || (options.mixins = [])).push({
      data: function data() {
        return collectDataFromConstructor(this, Component);
      },
    }); // decorate options

    const decorators = Component.__decorators__;

    if (decorators) {
      decorators.forEach(fn => fn(options));
      delete Component.__decorators__;
    } // find super


    const superProto = Object.getPrototypeOf(Component.prototype);
    const Super = superProto instanceof Vue__default.default ? superProto.constructor : Vue__default.default;
    const Extended = Super.extend(options);
    forwardStaticMembers(Extended, Component, Super);

    if (reflectionIsSupported()) {
      copyReflectionMetadata(Extended, Component);
    }

    return Extended;
  }
  const shouldIgnore = {
    prototype: true,
    arguments: true,
    callee: true,
    caller: true,
  };

  function forwardStaticMembers(Extended, Original, Super) {
    // We have to use getOwnPropertyNames since Babel registers methods as non-enumerable
    Object.getOwnPropertyNames(Original).forEach((key) => {
      // Skip the properties that should not be overwritten
      if (shouldIgnore[key]) {
        return;
      } // Some browsers does not allow reconfigure built-in properties


      const extendedDescriptor = Object.getOwnPropertyDescriptor(Extended, key);

      if (extendedDescriptor && !extendedDescriptor.configurable) {
        return;
      }

      const descriptor = Object.getOwnPropertyDescriptor(Original, key); // If the user agent does not support `__proto__` or its family (IE <= 10),
      // the sub class properties may be inherited properties from the super class in TypeScript.
      // We need to exclude such properties to prevent to overwrite
      // the component options object which stored on the extended constructor (See #192).
      // If the value is a referenced value (object or function),
      // we can check equality of them and exclude it if they have the same reference.
      // If it is a primitive value, it will be forwarded for safety.

      if (!hasProto) {
        // Only `cid` is explicitly exluded from property forwarding
        // because we cannot detect whether it is a inherited property or not
        // on the no `__proto__` environment even though the property is reserved.
        if (key === 'cid') {
          return;
        }

        const superDescriptor = Object.getOwnPropertyDescriptor(Super, key);

        if (!isPrimitive(descriptor.value) && superDescriptor && superDescriptor.value === descriptor.value) {
          return;
        }
      } // Warn if the users manually declare reserved properties

      Object.defineProperty(Extended, key, descriptor);
    });
  }

  function Component(options) {
    if (typeof options === 'function') {
      return componentFactory(options);
    }

    return function (Component) {
      return componentFactory(Component, options);
    };
  }

  Component.registerHooks = function registerHooks(keys) {
    $internalHooks.push.apply($internalHooks, _toConsumableArray(keys));
  };

  const __spreadArrays = (undefined && undefined.__spreadArrays) || function () {
    for (var s = 0, i = 0, il = arguments.length; i < il; i++) s += arguments[i].length;
    for (var r = Array(s), k = 0, i = 0; i < il; i++) for (let a = arguments[i], j = 0, jl = a.length; j < jl; j++, k++) r[k] = a[j];
    return r;
  };
  // Code copied from Vue/src/shared/util.js
  const hyphenateRE = /\B([A-Z])/g;
  const hyphenate = function (str) {
    return str.replace(hyphenateRE, '-$1').toLowerCase();
  };
  /**
   * decorator of an event-emitter function
   * @param  event The name of the event
   * @return MethodDecorator
   */
  function Emit(event) {
    return function (_target, propertyKey, descriptor) {
      const key = hyphenate(propertyKey);
      const original = descriptor.value;
      descriptor.value = function emitter() {
        const _this = this;
        const args = [];
        for (let _i = 0; _i < arguments.length; _i++) {
          args[_i] = arguments[_i];
        }
        const emit = function (returnValue) {
          const emitName = event || key;
          if (returnValue === undefined) {
            if (args.length === 0) {
              _this.$emit(emitName);
            } else if (args.length === 1) {
              _this.$emit(emitName, args[0]);
            } else {
              _this.$emit.apply(_this, __spreadArrays([emitName], args));
            }
          } else {
            args.unshift(returnValue);
            _this.$emit.apply(_this, __spreadArrays([emitName], args));
          }
        };
        const returnValue = original.apply(this, args);
        if (isPromise(returnValue)) {
          returnValue.then(emit);
        } else {
          emit(returnValue);
        }
        return returnValue;
      };
    };
  }
  function isPromise(obj) {
    return obj instanceof Promise || (obj && typeof obj.then === 'function');
  }

  /** @see {@link https://github.com/vuejs/vue-class-component/blob/master/src/reflect.ts} */
  const reflectMetadataIsSupported = typeof Reflect !== 'undefined' && typeof Reflect.getMetadata !== 'undefined';
  function applyMetadata(options, target, key) {
    if (reflectMetadataIsSupported) {
      if (!Array.isArray(options)
              && typeof options !== 'function'
              && !options.hasOwnProperty('type')
              && typeof options.type === 'undefined') {
        const type = Reflect.getMetadata('design:type', target, key);
        if (type !== Object) {
          options.type = type;
        }
      }
    }
  }

  /**
   * decorator of model
   * @param  event event name
   * @param options options
   * @return PropertyDecorator
   */
  function Model(event, options) {
    if (options === void 0) {
      options = {};
    }
    return function (target, key) {
      applyMetadata(options, target, key);
      createDecorator((componentOptions, k) => {
        (componentOptions.props || (componentOptions.props = {}))[k] = options;
        componentOptions.model = { prop: k, event: event || k };
      })(target, key);
    };
  }

  /**
   * decorator of a prop
   * @param  options the options for the prop
   * @return PropertyDecorator | void
   */
  function Prop(options) {
    if (options === void 0) {
      options = {};
    }
    return function (target, key) {
      applyMetadata(options, target, key);
      createDecorator((componentOptions, k) => {
        (componentOptions.props || (componentOptions.props = {}))[k] = options;
      })(target, key);
    };
  }

  /**
   * decorator of a ref prop
   * @param refKey the ref key defined in template
   */
  function Ref(refKey) {
    return createDecorator((options, key) => {
      options.computed = options.computed || {};
      options.computed[key] = {
        cache: false,
        get() {
          return this.$refs[refKey || key];
        },
      };
    });
  }

  /**
   * decorator of a watch function
   * @param  path the path or the expression to observe
   * @param  WatchOption
   * @return MethodDecorator
   */
  function Watch(path, options) {
    if (options === void 0) {
      options = {};
    }
    const _a = options.deep; const deep = _a === void 0 ? false : _a; const _b = options.immediate; const immediate = _b === void 0 ? false : _b;
    return createDecorator((componentOptions, handler) => {
      if (typeof componentOptions.watch !== 'object') {
        componentOptions.watch = Object.create(null);
      }
      const { watch } = componentOptions;
      if (typeof watch[path] === 'object' && !Array.isArray(watch[path])) {
        watch[path] = [watch[path]];
      } else if (typeof watch[path] === 'undefined') {
        watch[path] = [];
      }
      watch[path].push({ handler, deep, immediate });
    });
  }

  /* ! *****************************************************************************
  Copyright (c) Microsoft Corporation.

  Permission to use, copy, modify, and/or distribute this software for any
  purpose with or without fee is hereby granted.

  THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
  REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
  AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
  INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
  LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
  OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
  PERFORMANCE OF THIS SOFTWARE.
  ***************************************************************************** */

  function __spreadArray(to, from, pack) {
    if (pack || arguments.length === 2) for (var i = 0, l = from.length, ar; i < l; i++) {
      if (ar || !(i in from)) {
        if (!ar) ar = Array.prototype.slice.call(from, 0, i);
        ar[i] = from[i];
      }
    }
    return to.concat(ar || Array.prototype.slice.call(from));
  }

  /**
   * 判断属性props是否存在obj中
   * @param obj
   * @param props
   */

  const hasOwnProperty = function hasOwnProperty(obj, props) {
    if (Array.isArray(props)) {
      return props.every(str => Object.prototype.hasOwnProperty.call(obj, str));
    }

    return Object.prototype.hasOwnProperty.call(obj, props);
  };
  /**
   * 防抖装饰器
   * @param delay
   */

  const Debounce = function Debounce(delay) {
    if (delay === void 0) {
      delay = 200;
    }

    return function (target, key, descriptor) {
      const originFunction = descriptor.value;

      const getNewFunction = function getNewFunction() {
        let timer;

        const newFunction = function newFunction() {
          const _this = this;

          const args = [];

          for (let _i = 0; _i < arguments.length; _i++) {
            args[_i] = arguments[_i];
          }

          if (timer) window.clearTimeout(timer);
          timer = setTimeout(() => {
            originFunction.call.apply(originFunction, __spreadArray([_this], args, false));
          }, delay);
        };

        return newFunction;
      };

      descriptor.value = getNewFunction();
      return descriptor;
    };
  };
  /**
   * 关键字搜索
   * @param data
   * @param keyword
   * @param accurate 是否开启精确搜索
   */

  const defaultSearch = function defaultSearch(data, keyword, accurate) {
    if (accurate === void 0) {
      accurate = false;
    }

    if (!Array.isArray(data) || keyword.trim() === '') return data;
    return data.filter(item => Object.keys(item).some((key) => {
      if (typeof item[key] === 'string') {
        return !!accurate ? item[key] === keyword : item[key].indexOf(keyword.trim()) > -1;
      }

      return false;
    }));
  };

  const resize = {
    bind: function bind(el, binding) {
      resizeDetector.addListener(el, binding.value);
    },
    unbind: function unbind(el, binding) {
      resizeDetector.removeListener(el, binding.value);
    },
  };

  let _dec$d; let _dec2$d; let _dec3$d; let _dec4$b; let _dec5$b; let _class$d; let _class2$d; let _descriptor$d; let _descriptor2$d; let _descriptor3$b;

  function _createSuper$d(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$d(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$d() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const Menu = (_dec$d = Component({
    name: 'menu-list',
  }), _dec2$d = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec3$d = Prop({
    default: 'left',
    type: String,
  }), _dec4$b = Prop({
    default: '',
    type: String,
  }), _dec5$b = Emit('click'), _dec$d(_class$d = (_class2$d = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(Menu, _Vue);

    const _super = _createSuper$d(Menu);

    function Menu() {
      let _this;

      _classCallCheck__default.default(this, Menu);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'list', _descriptor$d, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'align', _descriptor2$d, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'theme', _descriptor3$b, _assertThisInitialized__default.default(_this));

      return _this;
    }

    _createClass__default.default(Menu, [{
      key: 'handleMenuClick',
      value: function handleMenuClick(item) {
        return item;
      },
    }]);

    return Menu;
  }(Vue__default.default)), (_descriptor$d = _applyDecoratedDescriptor__default.default(_class2$d.prototype, 'list', [_dec2$d], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$d = _applyDecoratedDescriptor__default.default(_class2$d.prototype, 'align', [_dec3$d], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$b = _applyDecoratedDescriptor__default.default(_class2$d.prototype, 'theme', [_dec4$b], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$d.prototype, 'handleMenuClick', [_dec5$b], Object.getOwnPropertyDescriptor(_class2$d.prototype, 'handleMenuClick'), _class2$d.prototype)), _class2$d)) || _class$d);

  function normalizeComponent(template, style, script, scopeId, isFunctionalTemplate, moduleIdentifier /* server only */, shadowMode, createInjector, createInjectorSSR, createInjectorShadow) {
    if (typeof shadowMode !== 'boolean') {
      createInjectorSSR = createInjector;
      createInjector = shadowMode;
      shadowMode = false;
    }
    // Vue.extend constructor export interop.
    const options = typeof script === 'function' ? script.options : script;
    // render functions
    if (template && template.render) {
      options.render = template.render;
      options.staticRenderFns = template.staticRenderFns;
      options._compiled = true;
      // functional template
      if (isFunctionalTemplate) {
        options.functional = true;
      }
    }
    // scopedId
    if (scopeId) {
      options._scopeId = scopeId;
    }
    let hook;
    if (moduleIdentifier) {
      // server build
      hook = function (context) {
        // 2.3 injection
        context =                  context // cached call
                      || (this.$vnode && this.$vnode.ssrContext) // stateful
                      || (this.parent && this.parent.$vnode && this.parent.$vnode.ssrContext); // functional
        // 2.2 with runInNewContext: true
        if (!context && typeof __VUE_SSR_CONTEXT__ !== 'undefined') {
          context = __VUE_SSR_CONTEXT__;
        }
        // inject component styles
        if (style) {
          style.call(this, createInjectorSSR(context));
        }
        // register component module identifier for async chunk inference
        if (context && context._registeredComponents) {
          context._registeredComponents.add(moduleIdentifier);
        }
      };
      // used by ssr in case component is cached and beforeCreate
      // never gets called
      options._ssrRegister = hook;
    } else if (style) {
      hook = shadowMode
        ? function (context) {
          style.call(this, createInjectorShadow(context, this.$root.$options.shadowRoot));
        }
        : function (context) {
          style.call(this, createInjector(context));
        };
    }
    if (hook) {
      if (options.functional) {
        // register for functional component in vue file
        const originalRender = options.render;
        options.render = function renderWithStyleInjection(h, context) {
          hook.call(context);
          return originalRender(h, context);
        };
      } else {
        // inject component registration as beforeCreate hook
        const existing = options.beforeCreate;
        options.beforeCreate = existing ? [].concat(existing, hook) : [hook];
      }
    }
    return script;
  }

  /* script */
  const __vue_script__$d = Menu;
  /* template */

  const __vue_render__$d = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('ul', {
      class: `menu ${_vm.theme}`,
    }, _vm._l(_vm.list, (item, index) => _c('li', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: !item.hidden,
        expression: '!item.hidden',
      }],
      key: index,
      staticClass: 'menu-item',
      style: {
        'text-align': _vm.align,
      },
      attrs: {
        disabled: item.disabled,
      },
      on: {
        click: function click($event) {
          !item.disabled && _vm.handleMenuClick(item);
        },
      },
    }, [_vm._t('default', [_vm._v(`\n      ${_vm._s(item.label)}\n    `)], {
      item,
    })], 2)), 0);
  };

  const __vue_staticRenderFns__$d = [];
  /* style */

  const __vue_inject_styles__$d = undefined;
  /* scoped */

  const __vue_scope_id__$d = 'data-v-135bacaa';
  /* module identifier */

  const __vue_module_identifier__$d = undefined;
  /* functional template */

  const __vue_is_functional_template__$d = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$d = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$d,
    staticRenderFns: __vue_staticRenderFns__$d,
  }, __vue_inject_styles__$d, __vue_script__$d, __vue_scope_id__$d, __vue_is_functional_template__$d, __vue_module_identifier__$d, false, undefined, undefined, undefined);

  let _dec$c; let _dec2$c; let _dec3$c; let _dec4$a; let _dec5$a; let _dec6$a; let _dec7$a; let _dec8$a; let _dec9$a; let _class$c; let _class2$c; let _descriptor$c; let _descriptor2$c; let _descriptor3$a; let _descriptor4$a; let _descriptor5$a; let _descriptor6$a;

  function _createSuper$c(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$c(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$c() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const SelectorTab = (_dec$c = Component({
    name: 'selector-tab',
    directives: {
      resize,
    },
  }), _dec2$c = Model('tab-change', {
    default: '',
    type: String,
  }), _dec3$c = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
    validator: function validator(v) {
      const item = v.find(item => !hasOwnProperty(item, ['name', 'label']));
      item && console.warn(item, '缺少必要属性');
      return !item;
    },
  }), _dec4$a = Prop({
    default: true,
    type: Boolean,
  }), _dec5$a = Ref('tabwrapper'), _dec6$a = Ref('tabcontent'), _dec7$a = Ref('tabItem'), _dec8$a = Emit('tab-change'), _dec9$a = Debounce(400), _dec$c(_class$c = (_class2$c = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(SelectorTab, _Vue);

    const _super = _createSuper$c(SelectorTab);

    function SelectorTab() {
      let _this;

      _classCallCheck__default.default(this, SelectorTab);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'active', _descriptor$c, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'panels', _descriptor2$c, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tabVisible', _descriptor3$a, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tabwrapper', _descriptor4$a, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tabcontent', _descriptor5$a, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tabItemRef', _descriptor6$a, _assertThisInitialized__default.default(_this));

      _this.hiddenPanels = [];
      _this.popoverInstance = null;
      _this.menuInstance = null;
      return _this;
    }

    _createClass__default.default(SelectorTab, [{
      key: 'showArrow',
      get: function get() {
        return this.hiddenPanels.length !== 0;
      },
    }, {
      key: 'beforeDestroy',
      value: function beforeDestroy() {
        if (this.menuInstance) {
          this.menuInstance.$off('click', this.handleMenuClick);
          this.menuInstance.$destroy();
        }
      },
    }, {
      key: 'handleMenuClick',
      value: function handleMenuClick(menu) {
        this.handleTabChange(menu);
        this.popoverInstance && this.popoverInstance.hide();
      },
    }, {
      key: 'handleShowList',
      value: function handleShowList(event) {
        if (!event.target) return;

        if (!this.menuInstance) {
          this.menuInstance = new __vue_component__$d().$mount();
          this.menuInstance.$props.list = this.panels;
          this.menuInstance.$off('click', this.handleMenuClick);
          this.menuInstance.$on('click', this.handleMenuClick);
        }

        if (!this.popoverInstance) {
          this.popoverInstance = this.$bkPopover(event.target, {
            content: this.menuInstance.$el,
            trigger: 'manual',
            arrow: false,
            theme: 'light ip-selector',
            maxWidth: 280,
            offset: '0, 0',
            sticky: true,
            duration: [275, 0],
            interactive: true,
            boundary: 'window',
            placement: 'bottom',
          });
        }

        this.popoverInstance.show();
      },
    }, {
      key: 'handleTabChange',
      value: function handleTabChange(panel) {
        return panel.name;
      },
    }, {
      key: 'handleResize',
      value: function handleResize() {
        const _this2 = this;

        if (!this.tabwrapper || !this.tabcontent) return;

        const _this$tabwrapper$getB = this.tabwrapper.getBoundingClientRect();
        const wrapperRight = _this$tabwrapper$getB.right;

        this.tabItemRef && this.tabItemRef.forEach((node) => {
          const _getBoundingClientRec = node.getBoundingClientRect();
          const nodeRight = _getBoundingClientRec.right;
          const nodeWidth = _getBoundingClientRec.width;

          const nameData = node.dataset.name;

          const index = _this2.hiddenPanels.findIndex(item => item.name === nameData); // 32: 折叠按钮宽度


          if (nodeRight + 32 > wrapperRight) {
            index === -1 && nameData && _this2.hiddenPanels.push({
              name: nameData,
              width: nodeWidth,
            });
          }
        });
        this.$nextTick(() => {
          const wrapperWidth = _this2.tabwrapper.clientWidth;
          let contentWidth = _this2.tabcontent.clientWidth; // 按顺序显示panel

          _this2.panels.forEach((item) => {
            const index = _this2.hiddenPanels.findIndex(data => data.name === item.name);

            if (index > -1 && contentWidth + _this2.hiddenPanels[index].width < wrapperWidth) {
              contentWidth += _this2.hiddenPanels[index].width;

              _this2.hiddenPanels.splice(index, 1);
            }
          });
        });
      },
    }]);

    return SelectorTab;
  }(Vue__default.default)), (_descriptor$c = _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'active', [_dec2$c], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$c = _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'panels', [_dec3$c], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$a = _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'tabVisible', [_dec4$a], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$a = _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'tabwrapper', [_dec5$a], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$a = _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'tabcontent', [_dec6$a], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$a = _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'tabItemRef', [_dec7$a], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'handleTabChange', [_dec8$a], Object.getOwnPropertyDescriptor(_class2$c.prototype, 'handleTabChange'), _class2$c.prototype), _applyDecoratedDescriptor__default.default(_class2$c.prototype, 'handleResize', [_dec9$a], Object.getOwnPropertyDescriptor(_class2$c.prototype, 'handleResize'), _class2$c.prototype)), _class2$c)) || _class$c);

  /* script */
  const __vue_script__$c = SelectorTab;
  /* template */

  const __vue_render__$c = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      staticClass: 'selector-tab',
    }, [_c('div', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: _vm.tabVisible && _vm.panels.length,
        expression: 'tabVisible && panels.length',
      }, {
        name: 'resize',
        rawName: 'v-resize',
        value: _vm.handleResize,
        expression: 'handleResize',
      }],
      ref: 'tabwrapper',
      staticClass: 'selector-tab-header',
    }, [_c('ul', {
      ref: 'tabcontent',
      staticClass: 'selector-tab-horizontal',
    }, [_vm._l(_vm.panels, item => _c('li', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: !item.hidden && !_vm.hiddenPanels.some(data => data.name === item.name),
        expression: '!item.hidden && !hiddenPanels.some(data => data.name === item.name)',
      }, {
        name: 'bk-tooltips',
        rawName: 'v-bk-tooltips.top',
        value: {
          disabled: !item.disabled || !item.tips,
          content: item.tips,
          delay: [300, 0],
        },
        expression: '{\n            disabled: !item.disabled || !item.tips,\n            content: item.tips,\n            delay: [300, 0]\n          }',
        modifiers: {
          top: true,
        },
      }],
      key: item.name,
      ref: 'tabItem',
      refInFor: true,
      class: ['tab-item', {
        active: _vm.active === item.name,
      }, {
        disabled: item.disabled,
      }],
      attrs: {
        'data-name': item.name,
      },
      on: {
        click: function click($event) {
          !item.disabled && _vm.handleTabChange(item);
        },
      },
    }, [_vm._t('label', [_vm._v(`\n          ${_vm._s(item.label)}\n        `)], null, {
      item,
    })], 2)), _c('li', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: _vm.showArrow,
        expression: 'showArrow',
      }],
      on: {
        click: _vm.handleShowList,
      },
    }, [_vm._m(0)])], 2)]), _c('div', {
      staticClass: 'selector-tab-content',
      style: {
        'border-top': _vm.tabVisible && _vm.panels.length ? 'none' : '',
      },
    }, [_vm._t('default')], 2)]);
  };

  const __vue_staticRenderFns__$c = [function () {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('span', {
      staticClass: 'selector-tab-all',
    }, [_c('i', {
      staticClass: 'bk-icon icon-angle-double-right',
    })]);
  }];
  /* style */

  const __vue_inject_styles__$c = undefined;
  /* scoped */

  const __vue_scope_id__$c = 'data-v-f1ee4408';
  /* module identifier */

  const __vue_module_identifier__$c = undefined;
  /* functional template */

  const __vue_is_functional_template__$c = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$c = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$c,
    staticRenderFns: __vue_staticRenderFns__$c,
  }, __vue_inject_styles__$c, __vue_script__$c, __vue_scope_id__$c, __vue_is_functional_template__$c, __vue_module_identifier__$c, false, undefined, undefined, undefined);

  let _dec$b; let _dec2$b; let _dec3$b; let _dec4$9; let _dec5$9; let _dec6$9; let _dec7$9; let _dec8$9; let _dec9$9; let _dec10$9; let _dec11$9; let _class$b; let _class2$b; let _descriptor$b; let _descriptor2$b; let _descriptor3$9; let _descriptor4$9; let _descriptor5$9; let _descriptor6$9;

  function _createSuper$b(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$b(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$b() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const SelectionColumn = ( // 表格自定义check列
    _dec$b = Component({
      name: 'selection-column',
    }), _dec2$b = Model('update-value', {
      default: 0,
      type: Number,
    }), _dec3$b = Prop({
      default: false,
      type: Boolean,
    }), _dec4$9 = Prop({
      default: false,
      type: Boolean,
    }), _dec5$9 = Prop({
      default: 'current',
      type: String,
    }), _dec6$9 = Prop({
      default: false,
      type: Boolean,
    }), _dec7$9 = Ref('popover'), _dec8$9 = Watch('defaultActive'), _dec9$9 = Emit('change'), _dec10$9 = Emit('change'), _dec11$9 = Emit('update-value'), _dec$b(_class$b = (_class2$b = /* #__PURE__*/(function (_Vue) {
      _inherits__default.default(SelectionColumn, _Vue);

      const _super = _createSuper$b(SelectionColumn);

      function SelectionColumn() {
        let _this;

        _classCallCheck__default.default(this, SelectionColumn);

        for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
          args[_key] = arguments[_key];
        }

        _this = _super.call.apply(_super, [this].concat(args));

        _initializerDefineProperty__default.default(_this, 'value', _descriptor$b, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'disabled', _descriptor2$b, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'loading', _descriptor3$9, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'defaultActive', _descriptor4$9, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'acrossPage', _descriptor5$9, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'popover', _descriptor6$9, _assertThisInitialized__default.default(_this));

        _this.menuInstance = new __vue_component__$d().$mount();
        _this.popoverInstance = null;
        _this.checkType = {
          active: 'current',
          list: [],
        };
        _this.isDropDownShow = false;
        return _this;
      }

      _createClass__default.default(SelectionColumn, [{
        key: 'handleDefaultActiveChange',
        value: // 取消勾选时重置checkType
      // @Watch('value')
      // private handleValueChange(v: CheckValue) {
      //   v === 0 && (this.checkType.active = 'current')
      // }
      function handleDefaultActiveChange() {
        this.checkType.active = this.defaultActive;
      },
      }, {
        key: 'created',
        value: function created() {
          this.checkType.list = [{
            id: 'current',
            label: this.$t('cluster.nodeList.button.selectPage'),
          }, {
            id: 'all',
            label: this.$t('cluster.nodeList.button.selectAcrossPage'),
          }];
        },
      }, {
        key: 'beforeDestroy',
        value: function beforeDestroy() {
          if (this.menuInstance) {
            this.menuInstance.$off('click', this.handleMenuClick);
            this.menuInstance.$destroy();
          }
        },
        /**
       * 全选操作
       * @param {String} type 全选类型：1. 本页权限 2. 跨页全选
       */

      }, {
        key: 'handleCheckAll',
        value: function handleCheckAll(type) {
          this.popover && this.popover.instance.hide();
          this.checkType.active = type;
          this.handleUpdateValue(2);
          return {
            value: 2,
            type,
          };
        },
        /**
       * 勾选事件
       */

      }, {
        key: 'handleCheckChange',
        value: function handleCheckChange(value) {
        // if (!value) {
        //   this.checkType.active = 'current'
        // }
          this.handleUpdateValue(value ? 2 : 0);
          return {
            value: value ? 2 : 0,
            type: this.checkType.active,
          };
        },
      }, {
        key: 'handleUpdateValue',
        value: function handleUpdateValue(v) {
          return v;
        },
      }, {
        key: 'handleMenuClick',
        value: function handleMenuClick(menu) {
          this.handleCheckAll(menu.id);
          this.popoverInstance && this.popoverInstance.hide();
        },
      }, {
        key: 'handleShowMenu',
        value: function handleShowMenu(event) {
          const _this2 = this;

          if (!event.target || this.disabled) return;
          this.menuInstance.$props.list = this.checkType.list;
          this.menuInstance.$props.align = 'center';
          this.menuInstance.$off('click', this.handleMenuClick);
          this.menuInstance.$on('click', this.handleMenuClick);

          if (!this.popoverInstance) {
            this.popoverInstance = this.$bkPopover(event.target, {
              content: this.menuInstance.$el,
              trigger: 'manual',
              arrow: false,
              theme: 'light ip-selector',
              maxWidth: 280,
              offset: '30, 0',
              sticky: true,
              duration: [275, 0],
              interactive: true,
              boundary: 'window',
              placement: 'bottom',
              onHidden: function onHidden() {
                _this2.isDropDownShow = false;
              },
              onShow: function onShow() {
                _this2.isDropDownShow = true;
              },
            });
          }

          this.popoverInstance.show();
        },
      }]);

      return SelectionColumn;
    }(Vue__default.default)), (_descriptor$b = _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'value', [_dec2$b], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor2$b = _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'disabled', [_dec3$b], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor3$9 = _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'loading', [_dec4$9], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor4$9 = _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'defaultActive', [_dec5$9], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor5$9 = _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'acrossPage', [_dec6$9], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor6$9 = _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'popover', [_dec7$9], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'handleDefaultActiveChange', [_dec8$9], Object.getOwnPropertyDescriptor(_class2$b.prototype, 'handleDefaultActiveChange'), _class2$b.prototype), _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'handleCheckAll', [_dec9$9], Object.getOwnPropertyDescriptor(_class2$b.prototype, 'handleCheckAll'), _class2$b.prototype), _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'handleCheckChange', [_dec10$9], Object.getOwnPropertyDescriptor(_class2$b.prototype, 'handleCheckChange'), _class2$b.prototype), _applyDecoratedDescriptor__default.default(_class2$b.prototype, 'handleUpdateValue', [_dec11$9], Object.getOwnPropertyDescriptor(_class2$b.prototype, 'handleUpdateValue'), _class2$b.prototype)), _class2$b)) || _class$b);

  /* script */
  const __vue_script__$b = SelectionColumn;
  /* template */

  const __vue_render__$b = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      staticClass: 'selection-header',
    }, [_c('bk-checkbox', {
      class: {
        'all-check': _vm.checkType.active === 'all',
        indeterminate: _vm.value === 1 && _vm.checkType.active === 'all',
      },
      attrs: {
        value: _vm.value === 2,
        indeterminate: _vm.value === 1,
        disabled: _vm.disabled,
      },
      on: {
        change: _vm.handleCheckChange,
      },
    }), _vm.acrossPage ? _c('i', {
      class: ['bk-icon selection-header-icon', {
        disabled: _vm.disabled,
      }, _vm.isDropDownShow ? 'icon-angle-up' : 'icon-angle-down'],
      on: {
        click: _vm.handleShowMenu,
      },
    }) : _vm._e()], 1);
  };

  const __vue_staticRenderFns__$b = [];
  /* style */

  const __vue_inject_styles__$b = undefined;
  /* scoped */

  const __vue_scope_id__$b = 'data-v-41f40c41';
  /* module identifier */

  const __vue_module_identifier__$b = undefined;
  /* functional template */

  const __vue_is_functional_template__$b = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$b = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$b,
    staticRenderFns: __vue_staticRenderFns__$b,
  }, __vue_inject_styles__$b, __vue_script__$b, __vue_scope_id__$b, __vue_is_functional_template__$b, __vue_module_identifier__$b, false, undefined, undefined, undefined);

  let _dec$a; let _dec2$a; let _dec3$a; let _dec4$8; let _dec5$8; let _dec6$8; let _dec7$8; let _dec8$8; let _dec9$8; let _dec10$8; let _dec11$8; let _dec12$8; let _dec13$8; let _dec14$7; let _dec15$7; let _dec16$7; let _class$a; let _class2$a; let _descriptor$a; let _descriptor2$a; let _descriptor3$8; let _descriptor4$8; let _descriptor5$8; let _descriptor6$8; let _descriptor7$8; let _descriptor8$7; let _descriptor9$6; let _descriptor10$6; let _descriptor11$6;

  function _createSuper$a(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$a(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$a() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const IpSelectorTable = (_dec$a = Component({
    name: 'ip-selector-table',
  }), _dec2$a = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec3$a = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec4$8 = Prop({
    default: function _default() {
      return {};
    },
    type: Object,
  }), _dec5$8 = Prop({
    type: Number,
  }), _dec6$8 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec7$8 = Prop({
    default: true,
    type: Boolean,
  }), _dec8$8 = Prop({
    default: '',
    type: String,
  }), _dec9$8 = Prop({
    default: 'rtl',
    type: String,
  }), _dec10$8 = Prop({
    default: false,
    type: Boolean,
  }), _dec11$8 = Prop({
    type: Function,
  }), _dec12$8 = Prop({
    type: Function,
  }), _dec13$8 = Watch('defaultSelections'), _dec14$7 = Emit('check-change'), _dec15$7 = Emit('page-change'), _dec16$7 = Emit('page-limit-change'), _dec$a(_class$a = (_class2$a = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(IpSelectorTable, _Vue);

    const _super = _createSuper$a(IpSelectorTable);

    function IpSelectorTable() {
      let _this;

      _classCallCheck__default.default(this, IpSelectorTable);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'data', _descriptor$a, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'config', _descriptor2$a, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'pagination', _descriptor3$8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'maxHeight', _descriptor4$8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultSelections', _descriptor5$8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'showSelectionColumn', _descriptor6$8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'emptyText', _descriptor7$8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'ellipsisDirection', _descriptor8$7, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'acrossPage', _descriptor9$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowDisabledStatus', _descriptor10$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowTipsContent', _descriptor11$6, _assertThisInitialized__default.default(_this));

      _this.selections = [];
      _this.excludeData = [];
      _this.checkValue = 0;
      _this.checkType = 'current';
      return _this;
    }

    _createClass__default.default(IpSelectorTable, [{
      key: 'created',
      value: function created() {
        this.selections = this.defaultSelections;
      },
    }, {
      key: 'handleDefaultSelectionsChange',
      value: function handleDefaultSelectionsChange() {
        const _this2 = this;

        this.selections = this.defaultSelections; // 重新计算当前页未被check的数据

        this.excludeData = this.data.reduce((pre, next) => {
          if (_this2.selections.indexOf(next) === -1) {
            pre.push(next);
          }

          return pre;
        }, []);
        this.updateCheckStatus();
      },
    }, {
      key: 'renderVnodeRow',
      value: function renderVnodeRow(item, row, column, $index) {
        if (!item.render) return item.label;
        this.$slots['vnodeRow'.concat($index)] = [item.render(row, column, $index)];
      },
    }, {
      key: 'renderHeader',
      value: function renderHeader(h) {
        const _this3 = this;

        return h(__vue_component__$b, {
          props: {
            value: this.checkValue,
            disabled: !this.data.length,
            defaultActive: this.checkType,
            acrossPage: this.acrossPage,
          },
          on: {
            'update-value': function updateValue(v) {
              _this3.checkValue = v;
            },
            change: this.handleSelectionChange,
          },
        });
      }, // 全选和取消全选操作
      // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'handleSelectionChange',
      value: function handleSelectionChange(_ref) {
        const _this4 = this;

        const { value } = _ref;
        const { type } = _ref;
        this.checkValue = value;
        this.checkType = type;
        this.excludeData = value === 0 ? _toConsumableArray__default.default(this.data) : _toConsumableArray__default.default(this.data.filter(item => !!_this4.getRowDisabledStatus(item)));
        this.selections = value === 2 ? _toConsumableArray__default.default(this.data.filter(item => !_this4.getRowDisabledStatus(item))) : [];
        this.handleCheckChange();
      },
    }, {
      key: 'handleRowCheckChange',
      value: function handleRowCheckChange(row, checked) {
        this.setRowSelection(row, checked);
        this.handleCheckChange();
      },
    }, {
      key: 'handleCheckChange',
      value: function handleCheckChange() {
        return {
          excludeData: this.excludeData,
          selections: this.selections,
          checkType: this.checkType,
          checkValue: this.checkValue,
        };
      },
    }, {
      key: 'handlePageChange',
      value: function handlePageChange(page) {
        this.checkType === 'current' && this.resetCheckedStatus();
        return page;
      },
    }, {
      key: 'getCheckedStatus',
      value: function getCheckedStatus(row) {
        if (this.checkType === 'current') {
          return this.selections.indexOf(row) > -1;
        }

        return this.excludeData.indexOf(row) === -1;
      }, // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'resetCheckedStatus',
      value: function resetCheckedStatus() {
        this.checkType = 'current';
        this.checkValue = 0;
        this.selections = [];
        this.excludeData = [];
      }, // 设置当前行选中状态
      // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'setRowSelection',
      value: function setRowSelection(row, checked) {
        const _this5 = this;

        if (checked) {
          this.selections.push(row);
        } else {
          const index = this.selections.indexOf(row);
          index > -1 && this.selections.splice(index, 1);
        }

        if (this.checkType === 'current') {
          // 重新计算当前页未被check的数据
          this.excludeData = this.data.reduce((pre, next) => {
            if (_this5.selections.indexOf(next) === -1) {
              pre.push(next);
            }

            return pre;
          }, []);
        } else {
          if (checked) {
            const _index = this.excludeData.indexOf(row);

            _index > -1 && this.excludeData.splice(_index, 1);
          } else {
            this.excludeData.push(row);
          }
        }

        this.updateCheckStatus();
      },
    }, {
      key: 'updateCheckStatus',
      value: function updateCheckStatus() {
        // 设置当前check状态
        if (!this.data.length) {
          this.checkValue = 0;
        } else if (this.excludeData.length === 0) {
          // 未选
          this.checkValue = 2;
        } else if ([this.pagination.count, this.data.length].includes(this.excludeData.length)) {
          // 取消全选
          this.checkValue = 0;
          this.checkType = 'current';
          this.selections = [];
        } else {
          // 半选
          this.checkValue = 1;
        }
      },
    }, {
      key: 'handleCellClass',
      value: function handleCellClass(_ref2) {
        const { columnIndex } = _ref2;

        if (this.showSelectionColumn && columnIndex === 0) {
          return 'selection-cell';
        }
      },
    }, {
      key: 'handleRowClass',
      value: function handleRowClass(_ref3) {
        const { row } = _ref3;

        if (row.disabled) {
          return 'row-disabled';
        }
      },
    }, {
      key: 'handlePageLimitChange',
      value: function handlePageLimitChange(limit) {
        return limit;
      },
    }]);

    return IpSelectorTable;
  }(Vue__default.default)), (_descriptor$a = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'data', [_dec2$a], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$a = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'config', [_dec3$a], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$8 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'pagination', [_dec4$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$8 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'maxHeight', [_dec5$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$8 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'defaultSelections', [_dec6$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$8 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'showSelectionColumn', [_dec7$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7$8 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'emptyText', [_dec8$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor8$7 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'ellipsisDirection', [_dec9$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor9$6 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'acrossPage', [_dec10$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor10$6 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'getRowDisabledStatus', [_dec11$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor11$6 = _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'getRowTipsContent', [_dec12$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'handleDefaultSelectionsChange', [_dec13$8], Object.getOwnPropertyDescriptor(_class2$a.prototype, 'handleDefaultSelectionsChange'), _class2$a.prototype), _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'handleCheckChange', [_dec14$7], Object.getOwnPropertyDescriptor(_class2$a.prototype, 'handleCheckChange'), _class2$a.prototype), _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'handlePageChange', [_dec15$7], Object.getOwnPropertyDescriptor(_class2$a.prototype, 'handlePageChange'), _class2$a.prototype), _applyDecoratedDescriptor__default.default(_class2$a.prototype, 'handlePageLimitChange', [_dec16$7], Object.getOwnPropertyDescriptor(_class2$a.prototype, 'handlePageLimitChange'), _class2$a.prototype)), _class2$a)) || _class$a);

  /* script */
  const __vue_script__$a = IpSelectorTable;
  /* template */

  const __vue_render__$a = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', [_c('bk-table', {
      key: _vm.maxHeight,
      staticClass: 'topo-table',
      attrs: {
        data: _vm.data,
        'outer-border': false,
        'header-border': false,
        'max-height': _vm.maxHeight,
        'empty-text': _vm.emptyText,
        'row-class-name': _vm.handleRowClass,
        'cell-class-name': _vm.handleCellClass,
        'header-cell-class-name': _vm.handleCellClass,
      },
    }, [_vm.showSelectionColumn ? _c('bk-table-column', {
      attrs: {
        'render-header': _vm.renderHeader,
        width: '50',
        resizable: false,
      },
      scopedSlots: _vm._u([{
        key: 'default',
        fn: function fn(ref) {
          const { row } = ref;
          return [_c('span', {
            directives: [{
              name: 'bk-tooltips',
              rawName: 'v-bk-tooltips',
              value: {
                placement: 'left',
                boundary: 'window',
                content: _vm.getRowTipsContent && _vm.getRowTipsContent(row),
                disabled: !_vm.getRowTipsContent || !_vm.getRowTipsContent(row),
              },
              expression: '{\n            placement: \'left\',\n            boundary: \'window\', \n            content: getRowTipsContent && getRowTipsContent(row),\n            disabled: !getRowTipsContent || !getRowTipsContent(row)\n          }',
            }],
          }, [_c('bk-checkbox', {
            attrs: {
              checked: _vm.getCheckedStatus(row),
              disabled: _vm.getRowDisabledStatus && _vm.getRowDisabledStatus(row),
            },
            on: {
              change: function change($event) {
                return _vm.handleRowCheckChange(row, $event);
              },
            },
          })], 1)];
        },
      }], null, false, 2260343160),
    }) : _vm._e(), _vm._l(_vm.config.filter(item => !item.hidden), item => _c('bk-table-column', {
      key: item.prop,
      attrs: {
        label: item.label,
        prop: item.prop,
        'min-width': item.minWidth,
        'show-overflow-tooltip': false,
      },
      scopedSlots: _vm._u([{
        key: 'default',
        fn: function fn(ref) {
          const { row } = ref;
          const { column } = ref;
          const { $index } = ref;
          return [item.render ? _c('div', [_vm._t(`vnodeRow${$index}`, [_vm._v(`\n            ${_vm._s(_vm.renderVnodeRow(item, row, column, $index))}\n          `)])], 2) : typeof row[item.prop] === 'number' ? _c('span', [_vm._v(`\n          ${_vm._s(row[item.prop])}\n        `)]) : _c('span', {
            directives: [{
              name: 'bk-overflow-tips',
              rawName: 'v-bk-overflow-tips',
            }],
            staticClass: 'column-string',
            style: {
              direction: _vm.ellipsisDirection,
            },
          }, [_vm._v(`\n          ${_vm._s(row[item.prop] || '--')}\n        `)])];
        },
      }], null, true),
    }))], 2), _c('bk-pagination', _vm._b({
      staticClass: 'mt10',
      on: {
        change: _vm.handlePageChange,
        'limit-change': _vm.handlePageLimitChange,
      },
    }, 'bk-pagination', _vm.pagination, false))], 1);
  };

  const __vue_staticRenderFns__$a = [];
  /* style */

  const __vue_inject_styles__$a = undefined;
  /* scoped */

  const __vue_scope_id__$a = 'data-v-910787bc';
  /* module identifier */

  const __vue_module_identifier__$a = undefined;
  /* functional template */

  const __vue_is_functional_template__$a = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$a = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$a,
    staticRenderFns: __vue_staticRenderFns__$a,
  }, __vue_inject_styles__$a, __vue_script__$a, __vue_scope_id__$a, __vue_is_functional_template__$a, __vue_module_identifier__$a, false, undefined, undefined, undefined);

  let _dec$9; let _dec2$9; let _dec3$9; let _dec4$7; let _dec5$7; let _dec6$7; let _dec7$7; let _dec8$7; let _dec9$7; let _dec10$7; let _dec11$7; let _dec12$7; let _dec13$7; let _dec14$6; let _dec15$6; let _dec16$6; let _dec17$5; let _dec18$3; let _dec19$2; let _dec20$2; let _dec21$2; let _class$9; let _class2$9; let _descriptor$9; let _descriptor2$9; let _descriptor3$7; let _descriptor4$7; let _descriptor5$7; let _descriptor6$7; let _descriptor7$7; let _descriptor8$6; let _descriptor9$5; let _descriptor10$5; let _descriptor11$5; let _descriptor12$5; let _descriptor13$4; let _descriptor14$4; let _descriptor15$3; let _descriptor16$2; let _descriptor17$2;

  function _createSuper$9(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$9(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$9() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const IpList = ( // IP列表
    _dec$9 = Component({
      name: 'ip-list',
      components: {
        IpSelectorTable: __vue_component__$a,
      },
    }), _dec2$9 = Prop({
      type: Function,
      required: true,
    }), _dec3$9 = Prop({
      type: Function,
    }), _dec4$7 = Prop({
      default: '',
      type: String,
    }), _dec5$7 = Prop({
      default: function _default() {
        return [];
      },
      type: Array,
    }), _dec6$7 = Prop({
      default: 20,
      type: Number,
    }), _dec7$7 = Prop({
      default: 0,
      type: Number,
    }), _dec8$7 = Prop({
      default: true,
      type: Boolean,
    }), _dec9$7 = Prop({
      default: false,
      type: Boolean,
    }), _dec10$7 = Prop({
      default: '',
      type: String,
    }), _dec11$7 = Prop({
      default: 'rtl',
      type: String,
    }), _dec12$7 = Prop({
      default: false,
      type: Boolean,
    }), _dec13$7 = Prop({
      type: Function,
    }), _dec14$6 = Prop({
      type: Function,
    }), _dec15$6 = Prop({
      default: false,
      type: Boolean,
    }), _dec16$6 = Prop({
      default: true,
      type: Boolean,
    }), _dec17$5 = Ref('ipListWrapper'), _dec18$3 = Ref('table'), _dec19$2 = Watch('slotHeight'), _dec20$2 = Debounce(300), _dec21$2 = Emit('check-change'), _dec$9(_class$9 = (_class2$9 = /* #__PURE__*/(function (_Vue) {
      _inherits__default.default(IpList, _Vue);

      const _super = _createSuper$9(IpList);

      function IpList() {
        let _this;

        _classCallCheck__default.default(this, IpList);

        for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
          args[_key] = arguments[_key];
        }

        _this = _super.call.apply(_super, [this].concat(args));

        _initializerDefineProperty__default.default(_this, 'getSearchTableData', _descriptor$9, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'getDefaultSelections', _descriptor2$9, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'ipListPlaceholder', _descriptor3$7, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'ipListTableConfig', _descriptor4$7, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'limit', _descriptor5$7, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'slotHeight', _descriptor6$7, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'showSelectionColumn', _descriptor7$7, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'disabledLoading', _descriptor8$6, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'emptyText', _descriptor9$5, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'ellipsisDirection', _descriptor10$5, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'acrossPage', _descriptor11$5, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'getRowDisabledStatus', _descriptor12$5, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'getRowTipsContent', _descriptor13$4, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'defaultAccurate', _descriptor14$4, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'showAccurate', _descriptor15$3, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'ipListWrapperRef', _descriptor16$2, _assertThisInitialized__default.default(_this));

        _initializerDefineProperty__default.default(_this, 'tableRef', _descriptor17$2, _assertThisInitialized__default.default(_this));

        _this.isLoading = false;
        _this.fullData = [];
        _this.frontendPagination = false;
        _this.tableData = [];
        _this.tableKeyword = '';
        _this.pagination = {
          current: 1,
          limit: _this.limit,
          count: 0,
          small: true,
          showLimit: false,
          showTotalCount: false,
          align: 'center',
          limitList: [20, 50, 100],
        };
        _this.maxHeight = 400;
        _this.defaultSelections = [];
        _this.accurate = false;
        return _this;
      }

      _createClass__default.default(IpList, [{
        key: 'handleSlotHeightChange',
        value: function handleSlotHeightChange() {
          this.computedTableLimit();
        },
      }, {
        key: 'created',
        value: function created() {
          this.accurate = this.defaultAccurate;
          this.handleGetDefaultData();
        },
      }, {
        key: 'mounted',
        value: function mounted() {
          this.computedTableLimit();
        },
      }, {
        key: 'computedTableLimit',
        value: function computedTableLimit() {
          const _this2 = this;

          // fix: 在弹窗时渲染IP选择器表格计算不准确问题
          setTimeout(() => {
            let _this2$pagination$lim;

            // 表格最大高度， 数字76: 去除 输入框 + margin + 分页组件 + margin 的高度
            _this2.maxHeight = _this2.ipListWrapperRef.clientHeight - _this2.slotHeight - 86;
            if (!_this2.maxHeight) return; // 表格分页条数，数字42: 去除表格header的高度

            const limit = Math.floor((_this2.maxHeight - 42) / 42);
            if (limit <= 0) return;

            if (!((_this2$pagination$lim = _this2.pagination.limitList) !== null && _this2$pagination$lim !== void 0 && _this2$pagination$lim.includes(limit))) {
              let _this2$pagination$lim2;

              (_this2$pagination$lim2 = _this2.pagination.limitList) === null || _this2$pagination$lim2 === void 0 ? void 0 : _this2$pagination$lim2.push(limit);
            }

            _this2.pagination.limit = limit;
          }, 0);
        }, // eslint-disable-next-line @typescript-eslint/member-ordering

      }, {
        key: 'handleGetDefaultData',
        value: function handleGetDefaultData() {
          const type = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : '';
          this.pagination.current = 1;
          this.pagination.count = 0;
          this.tableRef && this.tableRef.resetCheckedStatus();
          this.handleGetSearchData(type);
        }, // eslint-disable-next-line @typescript-eslint/member-ordering

      }, {
        key: 'handleGetDefaultSelections',
        value: function handleGetDefaultSelections() {
          const _this3 = this;

          // 获取默认勾选项
          this.defaultSelections = this.tableData.filter(row => _this3.getDefaultSelections && !!_this3.getDefaultSelections(row));
        }, // eslint-disable-next-line @typescript-eslint/member-ordering

      }, {
        key: 'selectionAllData',
        value: function selectionAllData() {
          const _this4 = this;

          this.$nextTick(() => {
            !!_this4.tableData.length && _this4.tableRef && _this4.tableRef.handleSelectionChange({
              value: 2,
              type: 'all',
            });
          });
        },
      }, {
        key: 'clearTableKeyWord',
        value: function clearTableKeyWord() {
          this.tableKeyword = '';
        },
      }, {
        key: 'handleAccurateChange',
        value: function handleAccurateChange() {
          this.handleGetSearchData('accurate-change');
        },
      }, {
        key: 'handleGetSearchData',
        value: (function () {
          const _handleGetSearchData = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee() {
            let type;
            let params;
            let _yield$this$getSearch;
            let total;
            let data;
            let _this$pagination;
            let limit;
            let current;
            const _args = arguments;

            return _regeneratorRuntime__default.default.wrap(function _callee$(_context) {
              while (1) {
                switch (_context.prev = _context.next) {
                  case 0:
                    type = _args.length > 0 && _args[0] !== undefined ? _args[0] : '';
                    _context.prev = 1;
                    this.isLoading = true;
                    params = {
                      current: this.pagination.current,
                      limit: this.pagination.limit,
                      tableKeyword: this.tableKeyword,
                      accurate: this.accurate,
                    };
                    _context.next = 6;
                    return this.getSearchTableData(params, type);

                  case 6:
                    _yield$this$getSearch = _context.sent;
                    total = _yield$this$getSearch.total;
                    data = _yield$this$getSearch.data;

                    if (data.length > this.pagination.limit) {
                      this.frontendPagination = true;
                      this.fullData = data; // 如果未分页，则前端自动分页

                      _this$pagination = this.pagination, limit = _this$pagination.limit, current = _this$pagination.current;
                      this.tableData = data.slice(limit * (current - 1), limit * current);
                    } else {
                      this.frontendPagination = false;
                      this.tableData = data || [];
                    }

                    this.pagination.count = total || 0;
                    this.handleGetDefaultSelections();
                    _context.next = 17;
                    break;

                  case 14:
                    _context.prev = 14;
                    _context.t0 = _context.catch(1);
                    console.log(_context.t0);

                  case 17:
                    _context.prev = 17;
                    this.isLoading = false;
                    return _context.finish(17);

                  case 20:
                  case 'end':
                    return _context.stop();
                }
              }
            }, _callee, this, [[1, 14, 17, 20]]);
          }));

          function handleGetSearchData() {
            return _handleGetSearchData.apply(this, arguments);
          }

          return handleGetSearchData;
        }()),
      }, {
        key: 'handlePageChange',
        value: function handlePageChange(page) {
          if (page === this.pagination.current) return;
          this.pagination.current = page;
          this.handleGetSearchData('page-change');
        },
      }, {
        key: 'handleLimitChange',
        value: function handleLimitChange(limit) {
          this.pagination.limit = limit;
          this.handleGetSearchData('limit-change');
        },
      }, {
        key: 'handleKeywordChange',
        value: function handleKeywordChange() {
          this.handleGetDefaultData('keyword-change');
        },
      }, {
        key: 'handleCheckChange',
        value: function handleCheckChange(data) {
          const { selections } = data;
          const { excludeData } = data;
          const { checkType } = data;
          const { checkValue } = data;
          let tmpSelections = selections;
          let tmpExcludeData = excludeData; // 前端分页

          if (this.frontendPagination && checkType === 'all') {
          // 跨页全选
            if (checkValue === 2) {
              tmpSelections = this.fullData.filter(item => (excludeData === null || excludeData === void 0 ? void 0 : excludeData.indexOf(item)) === -1);
            } else if (checkValue === 0) {
              tmpExcludeData = this.fullData.filter(item => selections.indexOf(item) === -1);
            }
          }

          return {
            selections: tmpSelections,
            excludeData: tmpExcludeData,
            checkType,
            checkValue,
          };
        },
      }]);

      return IpList;
    }(Vue__default.default)), (_descriptor$9 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'getSearchTableData', [_dec2$9], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor2$9 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'getDefaultSelections', [_dec3$9], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor3$7 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'ipListPlaceholder', [_dec4$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor4$7 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'ipListTableConfig', [_dec5$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor5$7 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'limit', [_dec6$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor6$7 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'slotHeight', [_dec7$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor7$7 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'showSelectionColumn', [_dec8$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor8$6 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'disabledLoading', [_dec9$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor9$5 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'emptyText', [_dec10$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor10$5 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'ellipsisDirection', [_dec11$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor11$5 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'acrossPage', [_dec12$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor12$5 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'getRowDisabledStatus', [_dec13$7], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor13$4 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'getRowTipsContent', [_dec14$6], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor14$4 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'defaultAccurate', [_dec15$6], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor15$3 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'showAccurate', [_dec16$6], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor16$2 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'ipListWrapperRef', [_dec17$5], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _descriptor17$2 = _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'tableRef', [_dec18$3], {
      configurable: true,
      enumerable: true,
      writable: true,
      initializer: null,
    }), _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'handleSlotHeightChange', [_dec19$2], Object.getOwnPropertyDescriptor(_class2$9.prototype, 'handleSlotHeightChange'), _class2$9.prototype), _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'handleKeywordChange', [_dec20$2], Object.getOwnPropertyDescriptor(_class2$9.prototype, 'handleKeywordChange'), _class2$9.prototype), _applyDecoratedDescriptor__default.default(_class2$9.prototype, 'handleCheckChange', [_dec21$2], Object.getOwnPropertyDescriptor(_class2$9.prototype, 'handleCheckChange'), _class2$9.prototype)), _class2$9)) || _class$9);

  /* script */
  const __vue_script__$9 = IpList;
  /* template */

  const __vue_render__$9 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      directives: [{
        name: 'bkloading',
        rawName: 'v-bkloading',
        value: {
          isLoading: _vm.isLoading && !_vm.disabledLoading,
        },
        expression: '{ isLoading: isLoading && !disabledLoading }',
      }],
      ref: 'ipListWrapper',
      staticClass: 'ip-list',
    }, [_c('div', {
      staticClass: 'ip-list-search',
    }, [_c('bk-input', {
      staticClass: 'search-input',
      attrs: {
        clearable: '',
        'right-icon': 'bk-icon icon-search',
        placeholder: _vm.ipListPlaceholder,
      },
      on: {
        change: _vm.handleKeywordChange,
      },
      model: {
        value: _vm.tableKeyword,
        callback: function callback($$v) {
          _vm.tableKeyword = $$v;
        },
        expression: 'tableKeyword',
      },
    }), _vm.showAccurate ? _c('bk-checkbox', {
      staticClass: 'ml10',
      on: {
        change: _vm.handleAccurateChange,
      },
      model: {
        value: _vm.accurate,
        callback: function callback($$v) {
          _vm.accurate = $$v;
        },
        expression: 'accurate',
      },
    }, [_vm._v(_vm._s(_vm.$t('generic.ipSelector.action.accurateSearch')))]) : _vm._e()], 1), _vm._t('tab'), _c('IpSelectorTable', {
      ref: 'table',
      staticClass: 'ip-list-table mt10',
      attrs: {
        data: _vm.tableData,
        config: _vm.ipListTableConfig,
        pagination: _vm.pagination,
        'max-height': _vm.maxHeight,
        'default-selections': _vm.defaultSelections,
        'show-selection-column': _vm.showSelectionColumn,
        'empty-text': _vm.emptyText,
        'ellipsis-direction': _vm.ellipsisDirection,
        'across-page': _vm.acrossPage,
        'get-row-disabled-status': _vm.getRowDisabledStatus,
        'get-row-tips-content': _vm.getRowTipsContent,
      },
      on: {
        'page-change': _vm.handlePageChange,
        'check-change': _vm.handleCheckChange,
        'page-limit-change': _vm.handleLimitChange,
      },
    })], 2);
  };

  const __vue_staticRenderFns__$9 = [];
  /* style */

  const __vue_inject_styles__$9 = undefined;
  /* scoped */

  const __vue_scope_id__$9 = 'data-v-2dbb6d0c';
  /* module identifier */

  const __vue_module_identifier__$9 = undefined;
  /* functional template */

  const __vue_is_functional_template__$9 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$9 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$9,
    staticRenderFns: __vue_staticRenderFns__$9,
  }, __vue_inject_styles__$9, __vue_script__$9, __vue_scope_id__$9, __vue_is_functional_template__$9, __vue_module_identifier__$9, false, undefined, undefined, undefined);

  let _dec$8; let _dec2$8; let _dec3$8; let _dec4$6; let _dec5$6; let _dec6$6; let _dec7$6; let _dec8$6; let _dec9$6; let _dec10$6; let _dec11$6; let _dec12$6; let _dec13$6; let _dec14$5; let _dec15$5; let _dec16$5; let _dec17$4; let _class$8; let _class2$8; let _descriptor$8; let _descriptor2$8; let _descriptor3$6; let _descriptor4$6; let _descriptor5$6; let _descriptor6$6; let _descriptor7$6; let _descriptor8$5; let _descriptor9$4; let _descriptor10$4; let _descriptor11$4; let _descriptor12$4; let _descriptor13$3; let _descriptor14$3; let _descriptor15$2;

  function ownKeys$1(object, enumerableOnly) {
    const keys = Object.keys(object); if (Object.getOwnPropertySymbols) {
      let symbols = Object.getOwnPropertySymbols(object); if (enumerableOnly) {
        symbols = symbols.filter(sym => Object.getOwnPropertyDescriptor(object, sym).enumerable);
      } keys.push.apply(keys, symbols);
    } return keys;
  }

  function _objectSpread$1(target) {
    for (let i = 1; i < arguments.length; i++) {
      var source = arguments[i] != null ? arguments[i] : {}; if (i % 2) {
        ownKeys$1(Object(source), true).forEach((key) => {
          _defineProperty__default.default(target, key, source[key]);
        });
      } else if (Object.getOwnPropertyDescriptors) {
        Object.defineProperties(target, Object.getOwnPropertyDescriptors(source));
      } else {
        ownKeys$1(Object(source)).forEach((key) => {
          Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key));
        });
      }
    } return target;
  }

  function _createSuper$8(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$8(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$8() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }

  const CustomInput = (_dec$8 = Component({
    name: 'custom-input',
    components: {
      IpSelectorTable: __vue_component__$a,
      IpListTable: __vue_component__$9,
    },
  }), _dec2$8 = Prop({
    type: Function,
    required: true,
  }), _dec3$8 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec4$6 = Prop({
    type: Function,
  }), _dec5$6 = Prop({
    default: 20,
    type: Number,
  }), _dec6$6 = Prop({
    default: 240,
    type: [Number, String],
  }), _dec7$6 = Prop({
    default: false,
    type: Boolean,
  }), _dec8$6 = Prop({
    default: 'ip',
    type: String,
  }), _dec9$6 = Prop({
    default: '',
    type: String,
  }), _dec10$6 = Prop({
    default: false,
    type: Boolean,
  }), _dec11$6 = Prop({
    type: Function,
  }), _dec12$6 = Prop({
    type: Function,
  }), _dec13$6 = Prop({
    default: false,
    type: Boolean,
  }), _dec14$5 = Prop({
    default: 'rtl',
    type: String,
  }), _dec15$5 = Prop({
    default: false,
    type: Boolean,
  }), _dec16$5 = Ref('table'), _dec17$4 = Emit('check-change'), _dec$8(_class$8 = (_class2$8 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(CustomInput, _Vue);

    const _super = _createSuper$8(CustomInput);

    function CustomInput() {
      let _this;

      _classCallCheck__default.default(this, CustomInput);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'getSearchTableData', _descriptor$8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'customInputTableConfig', _descriptor2$8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getDefaultSelections', _descriptor3$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'limit', _descriptor4$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'leftPanelWidth', _descriptor5$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'showTableTab', _descriptor6$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'ipKey', _descriptor7$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'customInputTablePlaceholder', _descriptor8$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'acrossPage', _descriptor9$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowDisabledStatus', _descriptor10$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowTipsContent', _descriptor11$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultAccurate', _descriptor12$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'ellipsisDirection', _descriptor13$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'showAccurate', _descriptor14$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tableRef', _descriptor15$2, _assertThisInitialized__default.default(_this));

      _this.errList = [];
      _this.temErrList = [];
      _this.goodList = [];
      _this.ipdata = '';
      _this.ipMatch = /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){3}$/;
      _this.ipTab = {
        active: 'inner',
        list: [],
      };
      _this.tabData = {
        inner: [],
        outer: [],
        other: [],
      };
      return _this;
    }

    _createClass__default.default(CustomInput, [{
      key: 'created',
      value: function created() {
        this.ipTab.list = [{
          id: 'inner',
          name: this.$t('generic.ipSelector.label.innerIp'),
        }, {
          id: 'outer',
          name: this.$t('generic.ipSelector.label.outerIp'),
        }, {
          id: 'other',
          name: this.$t('generic.ipSelector.label.otherIp'),
        }];
      },
    }, {
      key: 'handleTabClick',
      value: function handleTabClick(item) {
        this.ipTab.active = item.id;
        this.tableRef && this.tableRef.handleGetDefaultData();
      },
    }, {
      key: 'handleTableCheckChange',
      value: function handleTableCheckChange(data) {
        return data;
      },
    }, {
      key: 'handleDataChange',
      value: function handleDataChange() {
        this.goodList = [];
        this.errList = [];
      },
    }, {
      key: 'handleParseIp',
      value: function handleParseIp() {
        const _this2 = this;

        if (this.ipdata && this.ipdata.length) {
          this.ipTab.active = 'inner';
          const ipList = this.ipdata.split(/[\r\n]+/gm);
          const errList = new Set();
          const goodList = new Set();
          ipList.forEach((ips) => {
            const ip = ips.trim();

            if (ip.match(_this2.ipMatch)) {
              goodList.add(ip);
            } else {
              ip.length > 0 && errList.add(ip);
            }
          });

          if (errList.size > 0 && goodList.size === 0) {
            this.ipdata = Array.from(errList).join('\n');
            this.errList = Array.from(errList);
          } else {
            // 缓存当前错误IP
            this.temErrList = Array.from(errList);
          }

          this.goodList = Array.from(goodList);
          this.goodList.length && this.tableRef.handleGetDefaultData('input-change');
        }
      },
    }, {
      key: 'getTableData',
      value: (function () {
        const _getTableData = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee(params) {
          let type;
          let reqParams;
          let data;
          const _args = arguments;
          return _regeneratorRuntime__default.default.wrap(function _callee$(_context) {
            while (1) {
              switch (_context.prev = _context.next) {
                case 0:
                  type = _args.length > 1 && _args[1] !== undefined ? _args[1] : '';
                  _context.prev = 1;

                  if (this.goodList.length) {
                    _context.next = 4;
                    break;
                  }

                  return _context.abrupt('return', {
                    total: 0,
                    data: [],
                  });

                case 4:
                  reqParams = _objectSpread$1({
                    ipList: this.goodList,
                  }, params);

                  if (!(type === 'input-change')) {
                    _context.next = 8;
                    break;
                  }

                  _context.next = 8;
                  return this.handleParseDataChange(reqParams, type);

                case 8:
                  // eslint-disable-next-line
                  data = defaultSearch(this.tabData[this.ipTab.active], params.tableKeyword || '', !!params.accurate);
                  return _context.abrupt('return', {
                    total: data.length,
                    data,
                  });

                case 12:
                  _context.prev = 12;
                  _context.t0 = _context.catch(1);
                  console.log(_context.t0);
                  return _context.abrupt('return', {
                    total: 0,
                    data: [],
                  });

                case 16:
                case 'end':
                  return _context.stop();
              }
            }
          }, _callee, this, [[1, 12]]);
        }));

        function getTableData(_x) {
          return _getTableData.apply(this, arguments);
        }

        return getTableData;
      }()),
    }, {
      key: 'handleParseDataChange',
      value: (function () {
        const _handleParseDataChange = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee2(reqParams, type) {
          const _this3 = this;

          let res;
          return _regeneratorRuntime__default.default.wrap(function _callee2$(_context2) {
            while (1) {
              switch (_context2.prev = _context2.next) {
                case 0:
                  _context2.next = 2;
                  return this.getSearchTableData(reqParams, type);

                case 2:
                  res = _context2.sent;
                  // 分类数据
                  this.tabData = res.data.reduce((pre, next) => {
                    if (!!next.is_outerip) {
                      pre.outer.push(next);
                    } else if (!!next.is_external_ip) {
                      pre.other.push(next);
                    } else {
                      pre.inner.push(next);
                    }

                    return pre;
                  }, {
                    inner: [],
                    outer: [],
                    other: [],
                  });
                  this.goodList.forEach((ip) => {
                    // 对比返回值，找到全部错误IP
                    !res.data.some(item => item[_this3.ipKey] === ip) && _this3.temErrList.push(ip);
                  });
                  this.errList = _toConsumableArray__default.default(this.temErrList);
                  this.ipdata = this.errList.join('\n');
                  this.temErrList = [];
                  this.ipTab.active = ['inner', 'outer', 'other'].find(item => _this3.tabData[item] && _this3.tabData[item].length) || 'inner';
                  setTimeout(() => {
                    // 默认选择全部数据
                    // this.tableRef.selectionAllData()
                    _this3.$emit('check-change', {
                      selections: res.data,
                      excludeData: [],
                    });

                    _this3.tableRef && _this3.tableRef.handleGetDefaultData();
                  }, 0);

                case 10:
                case 'end':
                  return _context2.stop();
              }
            }
          }, _callee2, this);
        }));

        function handleParseDataChange(_x2, _x3) {
          return _handleParseDataChange.apply(this, arguments);
        }

        return handleParseDataChange;
      }()),
    }, {
      key: 'handleInputKeydown',
      value: function handleInputKeydown(e) {
        if (e.key === 'enter') {
          return true;
        }

        if (e.ctrlKey || e.shiftKey || e.metaKey) {
          return true;
        }

        if (!e.key.match(/[0-9.\s|,;]/) && !e.key.match(/(backspace|enter|ctrl|shift|tab)/mi)) {
          e.preventDefault();
        }
      }, // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'handleGetDefaultSelections',
      value: function handleGetDefaultSelections() {
        this.tableRef && this.tableRef.handleGetDefaultSelections();
      },
    }]);

    return CustomInput;
  }(Vue__default.default)), (_descriptor$8 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'getSearchTableData', [_dec2$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$8 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'customInputTableConfig', [_dec3$8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$6 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'getDefaultSelections', [_dec4$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$6 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'limit', [_dec5$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$6 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'leftPanelWidth', [_dec6$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$6 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'showTableTab', [_dec7$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7$6 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'ipKey', [_dec8$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor8$5 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'customInputTablePlaceholder', [_dec9$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor9$4 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'acrossPage', [_dec10$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor10$4 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'getRowDisabledStatus', [_dec11$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor11$4 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'getRowTipsContent', [_dec12$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor12$4 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'defaultAccurate', [_dec13$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor13$3 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'ellipsisDirection', [_dec14$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor14$3 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'showAccurate', [_dec15$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor15$2 = _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'tableRef', [_dec16$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$8.prototype, 'handleTableCheckChange', [_dec17$4], Object.getOwnPropertyDescriptor(_class2$8.prototype, 'handleTableCheckChange'), _class2$8.prototype)), _class2$8)) || _class$8);

  /* script */
  const __vue_script__$8 = CustomInput;
  /* template */

  const __vue_render__$8 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      staticClass: 'custom-input',
    }, [_c('div', {
      staticClass: 'custom-input-left',
      style: {
        width: isNaN(_vm.leftPanelWidth) ? _vm.leftPanelWidth : `${_vm.leftPanelWidth}px`,
      },
    }, [_c('bcs-input', {
      staticClass: 'ip-text',
      attrs: {
        placeholder: _vm.$t('generic.ipSelector.placeholder.searchIp2'),
        type: 'textarea',
      },
      on: {
        change: _vm.handleDataChange,
      },
      nativeOn: {
        keydown: function keydown($event) {
          return _vm.handleInputKeydown($event);
        },
      },
      model: {
        value: _vm.ipdata,
        callback: function callback($$v) {
          _vm.ipdata = $$v;
        },
        expression: 'ipdata',
      },
    }), _vm.errList.length ? _c('div', {
      staticClass: 'err-tips',
    }, [_vm._v(_vm._s(_vm.$t('generic.ipSelector.tips.ipError')))]) : _vm._e(), _c('bk-button', {
      staticClass: 'ip-parse',
      attrs: {
        theme: 'primary',
        outline: '',
      },
      on: {
        click: _vm.handleParseIp,
      },
    }, [_vm._v(`\n      ${_vm._s(_vm.$t('generic.ipSelector.button.parseIp'))}\n    `)])], 1), _c('div', {
      staticClass: 'custom-input-right ml20',
    }, [_c('IpListTable', {
      ref: 'table',
      attrs: {
        'get-search-table-data': _vm.getTableData,
        'ip-list-table-config': _vm.customInputTableConfig,
        'get-default-selections': _vm.getDefaultSelections,
        limit: _vm.limit,
        'slot-height': _vm.showTableTab ? 36 : 0,
        'ip-list-placeholder': _vm.customInputTablePlaceholder,
        'across-page': _vm.acrossPage,
        'get-row-disabled-status': _vm.getRowDisabledStatus,
        'get-row-tips-content': _vm.getRowTipsContent,
        'default-accurate': _vm.defaultAccurate,
        'ellipsis-direction': _vm.ellipsisDirection,
        'show-accurate': _vm.showAccurate,
      },
      on: {
        'check-change': _vm.handleTableCheckChange,
      },
      scopedSlots: _vm._u([{
        key: 'tab',
        fn: function fn() {
          return [_vm.showTableTab ? _c('ul', {
            staticClass: 'table-tab',
          }, _vm._l(_vm.ipTab.list, item => _c('li', {
            key: item.id,
            class: ['table-tab-item', {
              active: _vm.ipTab.active === item.id,
            }],
            on: {
              click: function click($event) {
                return _vm.handleTabClick(item);
              },
            },
          }, [_vm._v(`\n            ${_vm._s(item.name)}\n            `), _c('span', {
            staticClass: 'count',
          }, [_vm._v(_vm._s(`(${_vm.tabData[item.id] ? _vm.tabData[item.id].length : 0})`))])])), 0) : _vm._e()];
        },
        proxy: true,
      }]),
    })], 1)]);
  };

  const __vue_staticRenderFns__$8 = [];
  /* style */

  const __vue_inject_styles__$8 = undefined;
  /* scoped */

  const __vue_scope_id__$8 = 'data-v-9564f676';
  /* module identifier */

  const __vue_module_identifier__$8 = undefined;
  /* functional template */

  const __vue_is_functional_template__$8 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$8 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$8,
    staticRenderFns: __vue_staticRenderFns__$8,
  }, __vue_inject_styles__$8, __vue_script__$8, __vue_scope_id__$8, __vue_is_functional_template__$8, __vue_module_identifier__$8, false, undefined, undefined, undefined);

  let _dec$7; let _dec2$7; let _dec3$7; let _dec4$5; let _dec5$5; let _dec6$5; let _dec7$5; let _dec8$5; let _dec9$5; let _dec10$5; let _dec11$5; let _dec12$5; let _dec13$5; let _dec14$4; let _dec15$4; let _dec16$4; let _dec17$3; let _class$7; let _class2$7; let _descriptor$7; let _descriptor2$7; let _descriptor3$5; let _descriptor4$5; let _descriptor5$5; let _descriptor6$5; let _descriptor7$5; let _descriptor8$4; let _descriptor9$3; let _descriptor10$3; let _descriptor11$3; let _descriptor12$3; let _descriptor13$2; let _descriptor14$2;

  function _createSuper$7(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$7(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$7() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const StaticTopo$1 = (_dec$7 = Component({
    name: 'topo-tree',
  }), _dec2$7 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec3$7 = Prop({
    default: function _default() {
      return {};
    },
    type: Object,
  }), _dec4$5 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec5$5 = Prop({
    default: true,
    type: Boolean,
  }), _dec6$5 = Prop({
    default: 300,
    type: Number,
  }), _dec7$5 = Prop({
    default: false,
    type: Boolean,
  }), _dec8$5 = Prop({
    default: true,
    type: Boolean,
  }), _dec9$5 = Prop({
    type: Function,
  }), _dec10$5 = Prop({
    type: [Function, Boolean],
  }), _dec11$5 = Prop({
    default: 2,
    type: Number,
  }), _dec12$5 = Prop({
    default: '',
    type: [String, Number],
  }), _dec13$5 = Prop({
    default: '',
    type: String,
  }), _dec14$4 = Prop({
    default: false,
    type: Boolean,
  }), _dec15$4 = Ref('tree'), _dec16$4 = Watch('filter'), _dec17$3 = Emit('select-change'), _dec$7(_class$7 = (_class2$7 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(StaticTopo, _Vue);

    const _super = _createSuper$7(StaticTopo);

    function StaticTopo() {
      let _this;

      _classCallCheck__default.default(this, StaticTopo);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'defaultCheckedNodes', _descriptor$7, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'options', _descriptor2$7, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'nodes', _descriptor3$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'checkedable', _descriptor4$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'height', _descriptor5$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'expandOnClick', _descriptor6$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'showCount', _descriptor7$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'lazyMethod', _descriptor8$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'lazyDisabled', _descriptor9$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultExpandLevel', _descriptor10$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultSelectedNode', _descriptor11$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'filter', _descriptor12$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'enableTreeFilter', _descriptor13$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'treeRef', _descriptor14$2, _assertThisInitialized__default.default(_this));

      _this.defaultExpandedNodes = [];
      return _this;
    }

    _createClass__default.default(StaticTopo, [{
      key: 'nodeOptions',
      get: function get() {
        const nodeOptions = {
          idKey: 'id',
          nameKey: 'name',
          childrenKey: 'children',
        };
        return Object.assign(nodeOptions, this.options);
      },
    }, {
      key: 'created',
      value: function created() {
        this.defaultExpandedNodes = this.handleGetExpandNodeByDeep(this.defaultExpandLevel, this.nodes);
      },
    }, {
      key: 'mounted',
      value: function mounted() {
        if (!this.treeRef || !this.treeRef.getNodeById) return;
        const node = this.treeRef.getNodeById(this.defaultSelectedNode);
        node && this.handleSelectChange(node);
      },
    }, {
      key: 'getSelectedStatus',
      value: function getSelectedStatus(data) {
        const _this$nodeOptions$idK = this.nodeOptions.idKey;
        const idKey = _this$nodeOptions$idK === void 0 ? 'id' : _this$nodeOptions$idK;
        const id = data[idKey];
        return this.defaultCheckedNodes.includes(id);
      },
    }, {
      key: 'handleFilterTree',
      value: function handleFilterTree(filter) {
        this.enableTreeFilter && this.treeRef && this.treeRef.filter(filter);
      },
    }, {
      key: 'handleSelectChange',
      value: function handleSelectChange(treeNode) {
        return treeNode;
      },
    }, {
      key: 'handleSetChecked',
      value: function handleSetChecked(id) {
        if (this.treeRef) {
          this.treeRef.removeChecked();
          this.treeRef.setChecked(id, {
            emitEvent: false,
            beforeCheck: false,
            checked: true,
          });
        }
      },
    }, {
      key: 'addNode',
      value: function addNode(data, parentId) {
        this.treeRef && this.treeRef.addNode(data, parentId);
      },
    }, {
      key: 'handleGetExpandNodeByDeep',
      value: function handleGetExpandNodeByDeep() {
        const _this2 = this;

        const deep = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : 1;
        const treeData = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : [];
        const _this$nodeOptions = this.nodeOptions;
        const { idKey } = _this$nodeOptions;
        const { childrenKey } = _this$nodeOptions;
        return treeData.reduce((pre, node) => {
          (function (deep) {
            if (deep > 1 && Array.isArray(node[childrenKey]) && node[childrenKey].length > 0) {
              pre = pre.concat(_this2.handleGetExpandNodeByDeep(deep = deep - 1, node[childrenKey]));
            } else {
              pre = pre.concat(node[idKey]);
            }
          }(deep));

          return pre;
        }, []);
      },
    }]);

    return StaticTopo;
  }(Vue__default.default)), (_descriptor$7 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'defaultCheckedNodes', [_dec2$7], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$7 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'options', [_dec3$7], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$5 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'nodes', [_dec4$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$5 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'checkedable', [_dec5$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$5 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'height', [_dec6$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$5 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'expandOnClick', [_dec7$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7$5 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'showCount', [_dec8$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor8$4 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'lazyMethod', [_dec9$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor9$3 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'lazyDisabled', [_dec10$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor10$3 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'defaultExpandLevel', [_dec11$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor11$3 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'defaultSelectedNode', [_dec12$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor12$3 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'filter', [_dec13$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor13$2 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'enableTreeFilter', [_dec14$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor14$2 = _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'treeRef', [_dec15$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'handleFilterTree', [_dec16$4], Object.getOwnPropertyDescriptor(_class2$7.prototype, 'handleFilterTree'), _class2$7.prototype), _applyDecoratedDescriptor__default.default(_class2$7.prototype, 'handleSelectChange', [_dec17$3], Object.getOwnPropertyDescriptor(_class2$7.prototype, 'handleSelectChange'), _class2$7.prototype)), _class2$7)) || _class$7);

  /* script */
  const __vue_script__$7 = StaticTopo$1;
  /* template */

  const __vue_render__$7 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('bcs-big-tree', {
      ref: 'tree',
      attrs: {
        data: _vm.nodes,
        options: _vm.nodeOptions,
        height: _vm.height,
        selectable: '',
        'default-selected-node': _vm.defaultSelectedNode,
        'expand-on-click': _vm.expandOnClick,
        'default-checked-nodes': _vm.defaultCheckedNodes,
        'show-checkbox': '',
        'check-strictly': false,
        'default-expanded-nodes': _vm.defaultExpandedNodes,
        'lazy-method': _vm.lazyMethod,
        'lazy-disabled': _vm.lazyDisabled,
        padding: 20,
      },
      on: {
        'select-change': _vm.handleSelectChange,
      },
      scopedSlots: _vm._u([{
        key: 'default',
        fn: function fn(ref) {
          const { data } = ref;
          return [_c('div', {
            staticClass: 'node-label',
          }, [_c('span', {
            staticClass: 'label',
          }, [_vm._v(_vm._s(data[_vm.nodeOptions.nameKey]))]), _c('span', {
            directives: [{
              name: 'show',
              rawName: 'v-show',
              value: _vm.showCount,
              expression: 'showCount',
            }],
            class: ['num mr10', {
              selected: _vm.getSelectedStatus(data),
            }],
          }, [_vm._v(`\n        ${_vm._s(data[_vm.nodeOptions.childrenKey] ? data[_vm.nodeOptions.childrenKey].length : 0)}\n      `)])])];
        },
      }]),
    });
  };

  const __vue_staticRenderFns__$7 = [];
  /* style */

  const __vue_inject_styles__$7 = undefined;
  /* scoped */

  const __vue_scope_id__$7 = 'data-v-4115a8fa';
  /* module identifier */

  const __vue_module_identifier__$7 = undefined;
  /* functional template */

  const __vue_is_functional_template__$7 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$7 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$7,
    staticRenderFns: __vue_staticRenderFns__$7,
  }, __vue_inject_styles__$7, __vue_script__$7, __vue_scope_id__$7, __vue_is_functional_template__$7, __vue_module_identifier__$7, false, undefined, undefined, undefined);

  const commonjsGlobal = typeof globalThis !== 'undefined' ? globalThis : typeof window !== 'undefined' ? window : typeof global !== 'undefined' ? global : typeof self !== 'undefined' ? self : {};

  function getDefaultExportFromCjs(x) {
  	return x && x.__esModule && Object.prototype.hasOwnProperty.call(x, 'default') ? x.default : x;
  }

  const clickoutside$1 = { exports: {} };

  (function (module, exports) {
    (function (global, factory) {
      factory(exports) ;
    }(commonjsGlobal, (exports) => {
      const nodeList = [];
      const clickctx = '$clickoutsideCtx';
      let beginClick;
      document.addEventListener('mousedown', event => beginClick = event);
      document.addEventListener('mouseup', (event) => {
        nodeList.forEach((node) => {
          node[clickctx].clickoutsideHandler(event, beginClick);
        });
      });
      const bkClickoutside = {
        bind: function bind(el, binding, vnode) {
          const id = nodeList.push(el) - 1;
          const clickoutsideHandler = function clickoutsideHandler() {
            const mouseup = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
            const mousedown = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};
            if (!vnode.context
            || !mouseup.target || !mousedown.target || el.contains(mouseup.target)
            || el.contains(mousedown.target)
            || el === mouseup.target
            || vnode.context.popup
            && (vnode.context.popup.contains(mouseup.target)
            || vnode.context.popup.contains(mousedown.target)
            )) {
              return;
            }
            if (binding.expression
            && el[clickctx].callbackName
            && vnode.context[el[clickctx].callbackName]
            ) {
              vnode.context[el[clickctx].callbackName](mouseup, mousedown, el);
            } else {
              el[clickctx].bindingFn && el[clickctx].bindingFn(mouseup, mousedown, el);
            }
          };
          el[clickctx] = {
            id,
            clickoutsideHandler,
            callbackName: binding.expression,
            callbackFn: binding.value,
          };
        },
        update: function update(el, binding) {
          el[clickctx].callbackName = binding.expression;
          el[clickctx].callbackFn = binding.value;
        },
        unbind: function unbind(el) {
          for (let i = 0, len = nodeList.length; i < len; i++) {
            if (nodeList[i][clickctx].id === el[clickctx].id) {
              nodeList.splice(i, 1);
              break;
            }
          }
        },
      };
      bkClickoutside.install = function (Vue) {
        Vue.directive('bkClickoutside', bkClickoutside);
      };

      exports.default = bkClickoutside;

      Object.defineProperty(exports, '__esModule', { value: true });
    }));
  }(clickoutside$1, clickoutside$1.exports));

  const clickoutside = /* @__PURE__*/getDefaultExportFromCjs(clickoutside$1.exports);

  let _dec$6; let _dec2$6; let _dec3$6; let _dec4$4; let _dec5$4; let _dec6$4; let _dec7$4; let _dec8$4; let _dec9$4; let _dec10$4; let _dec11$4; let _dec12$4; let _dec13$4; let _dec14$3; let _dec15$3; let _dec16$3; let _class$6; let _class2$6; let _descriptor$6; let _descriptor2$6; let _descriptor3$4; let _descriptor4$4; let _descriptor5$4; let _descriptor6$4; let _descriptor7$4; let _descriptor8$3;

  function _createSuper$6(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$6(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$6() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const TopoSearch = (_dec$6 = Component({
    name: 'topo-search',
    directives: {
      clickoutside,
    },
  }), _dec2$6 = Model('change', {
    default: '',
    type: String,
  }), _dec3$6 = Prop({
    default: '',
    type: String,
  }), _dec4$4 = Prop({
    type: Function,
    required: true,
  }), _dec5$4 = Prop({
    default: 380,
    type: [Number, String],
  }), _dec6$4 = Prop({
    default: 300,
    type: Number,
  }), _dec7$4 = Prop({
    default: function _default() {
      return {};
    },
    type: Object,
  }), _dec8$4 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec9$4 = Prop({
    default: true,
    type: Boolean,
  }), _dec10$4 = Watch('defaultSelectionIds', {
    immediate: true,
  }), _dec11$4 = Emit('change'), _dec12$4 = Emit('hide'), _dec13$4 = Emit('show'), _dec14$3 = Emit('check-change'), _dec15$3 = Emit('check-change'), _dec16$3 = Debounce(300), _dec$6(_class$6 = (_class2$6 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(TopoSearch, _Vue);

    const _super = _createSuper$6(TopoSearch);

    function TopoSearch() {
      let _this;

      _classCallCheck__default.default(this, TopoSearch);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'value', _descriptor$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'placeholder', _descriptor2$6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'searchMethod', _descriptor3$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'resultWidth', _descriptor4$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'height', _descriptor5$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'options', _descriptor6$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultSelectionIds', _descriptor7$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'enableSearchPanel', _descriptor8$3, _assertThisInitialized__default.default(_this));

      _this.showPanel = false;
      _this.searchData = [];
      _this.selections = [];
      _this.isLoading = false;
      return _this;
    }

    _createClass__default.default(TopoSearch, [{
      key: 'dataOptions',
      get: function get() {
        const options = {
          idKey: 'id',
          nameKey: 'name',
          pathKey: 'node_path',
        };
        return Object.assign(options, this.options);
      },
    }, {
      key: 'handleDefaultSelectionsChange',
      value: function handleDefaultSelectionsChange() {
        const _this2 = this;

        this.selections = this.searchData.filter(item => _this2.defaultSelectionIds.includes(item.id));
      },
    }, {
      key: 'handleValueChange',
      value: function handleValueChange(v) {
        this.handleSearch(v);
        return v;
      },
    }, {
      key: 'handleClickoutside',
      value: function handleClickoutside() {
        this.showPanel = false;
      },
    }, {
      key: 'handleInputClick',
      value: function handleInputClick() {
        this.value !== '' && (this.showPanel = true);
      },
    }, {
      key: 'handleItemClick',
      value: function handleItemClick(item) {
        const _this3 = this;

        const index = this.selections.findIndex(select => select.id === item.id);

        if (index > -1) {
          this.selections.splice(index, 1);
        } else {
          this.selections.push(item);
        }

        return {
          selections: this.selections.map(select => select.data),
          excludeData: this.searchData.reduce((pre, next) => {
            if (_this3.selections.some(item => item.id === next.id)) return pre;
            pre.push(next.data);
            return pre;
          }, []),
        };
      },
    }, {
      key: 'handleCheckOrClearAll',
      value: function handleCheckOrClearAll() {
        this.selections = this.selections.length === this.searchData.length ? [] : _toConsumableArray__default.default(this.searchData);
        return {
          selections: this.selections.map(select => select.data),
          excludeData: !!this.selections.length ? [] : this.searchData.map(item => item.data),
        };
      },
    }, {
      key: 'getCheckStatus',
      value: function getCheckStatus(item) {
        return this.selections.some(select => select.id === item.id);
      }, // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'handleSearch',
      value: (function () {
        const _handleSearch = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee(keyword) {
          const _this4 = this;

          let data; let _this$dataOptions; let idKey; let nameKey; let pathKey;

          return _regeneratorRuntime__default.default.wrap(function _callee$(_context) {
            while (1) {
              switch (_context.prev = _context.next) {
                case 0:
                  this.showPanel = true;

                  if (!(!this.searchMethod || keyword === '')) {
                    _context.next = 5;
                    break;
                  }

                  this.searchData = [];
                  this.showPanel = false;
                  return _context.abrupt('return');

                case 5:
                  this.isLoading = true;
                  _context.next = 8;
                  return this.searchMethod(keyword).catch((err) => {
                    console.log(err);
                    return [];
                  });

                case 8:
                  data = _context.sent;
                  this.isLoading = false;
                  _this$dataOptions = this.dataOptions, idKey = _this$dataOptions.idKey, nameKey = _this$dataOptions.nameKey, pathKey = _this$dataOptions.pathKey;
                  this.searchData = Array.isArray(data) ? data.map((item, index) => {
                    const data = {
                      data: item,
                      id: item[idKey] || index,
                      label: item[nameKey],
                      path: pathKey ? item[pathKey] : '',
                    };
                    return data;
                  }) : [];
                  this.selections = this.searchData.filter(item => _this4.defaultSelectionIds.includes(item.id));

                case 13:
                case 'end':
                  return _context.stop();
              }
            }
          }, _callee, this);
        }));

        function handleSearch(_x) {
          return _handleSearch.apply(this, arguments);
        }

        return handleSearch;
      }()),
    }]);

    return TopoSearch;
  }(Vue__default.default)), (_descriptor$6 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'value', [_dec2$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$6 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'placeholder', [_dec3$6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$4 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'searchMethod', [_dec4$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$4 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'resultWidth', [_dec5$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$4 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'height', [_dec6$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$4 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'options', [_dec7$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7$4 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'defaultSelectionIds', [_dec8$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor8$3 = _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'enableSearchPanel', [_dec9$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'handleDefaultSelectionsChange', [_dec10$4], Object.getOwnPropertyDescriptor(_class2$6.prototype, 'handleDefaultSelectionsChange'), _class2$6.prototype), _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'handleValueChange', [_dec11$4], Object.getOwnPropertyDescriptor(_class2$6.prototype, 'handleValueChange'), _class2$6.prototype), _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'handleClickoutside', [_dec12$4], Object.getOwnPropertyDescriptor(_class2$6.prototype, 'handleClickoutside'), _class2$6.prototype), _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'handleInputClick', [_dec13$4], Object.getOwnPropertyDescriptor(_class2$6.prototype, 'handleInputClick'), _class2$6.prototype), _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'handleItemClick', [_dec14$3], Object.getOwnPropertyDescriptor(_class2$6.prototype, 'handleItemClick'), _class2$6.prototype), _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'handleCheckOrClearAll', [_dec15$3], Object.getOwnPropertyDescriptor(_class2$6.prototype, 'handleCheckOrClearAll'), _class2$6.prototype), _applyDecoratedDescriptor__default.default(_class2$6.prototype, 'handleSearch', [_dec16$3], Object.getOwnPropertyDescriptor(_class2$6.prototype, 'handleSearch'), _class2$6.prototype)), _class2$6)) || _class$6);

  /* script */
  const __vue_script__$6 = TopoSearch;
  /* template */

  const __vue_render__$6 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      staticClass: 'topo-search',
    }, [_c('bk-input', {
      attrs: {
        clearable: '',
        'right-icon': 'bk-icon icon-search',
        placeholder: _vm.placeholder,
        value: _vm.value,
      },
      on: {
        change: _vm.handleValueChange,
      },
      nativeOn: {
        click: function click($event) {
          return _vm.handleInputClick($event);
        },
      },
    }), _c('div', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: _vm.showPanel && _vm.enableSearchPanel,
        expression: 'showPanel && enableSearchPanel',
      }, {
        name: 'bk-clickoutside',
        rawName: 'v-bk-clickoutside',
        value: _vm.handleClickoutside,
        expression: 'handleClickoutside',
      }],
      staticClass: 'topo-search-result',
      style: {
        width: `${_vm.resultWidth}px`,
      },
    }, [_vm.searchData.length ? [_c('div', {
      staticClass: 'result-title',
    }, [_c('span', [_vm._v(_vm._s(_vm.$t('generic.ipSelector.title.searchResult')))]), _c('bk-button', {
      staticClass: 'select-all',
      attrs: {
        text: '',
      },
      on: {
        click: _vm.handleCheckOrClearAll,
      },
    }, [_vm._v(`\n          ${_vm._s(_vm.searchData.length === _vm.selections.length ? _vm.$t('generic.button.cancelSelectAll') : _vm.$t('generic.button.selectAll'))}\n        `)])], 1), _c('bk-virtual-scroll', {
      style: {
        height: `${_vm.height}px`,
      },
      attrs: {
        list: _vm.searchData,
        'item-height': 58,
      },
      scopedSlots: _vm._u([{
        key: 'default',
        fn: function fn(ref) {
          const { data } = ref;
          return [_c('div', {
            staticClass: 'result-panel-item',
            on: {
              click: function click($event) {
                return _vm.handleItemClick(data);
              },
            },
          }, [_c('div', {
            staticClass: 'item-left',
          }, [_c('span', {
            staticClass: 'item-left-name',
          }, [_vm._v(`\n                ${_vm._s(data.label)}\n              `)]), _c('span', {
            staticClass: 'item-left-path',
          }, [_vm._v(`\n                ${_vm._s(data.path)}\n              `)])]), _c('div', {
            staticClass: 'item-right',
          }, [_c('span', {
            class: ['checkbox', {
              'is-checked': _vm.getCheckStatus(data),
            }],
          })])])];
        },
      }], null, false, 2461268883),
    })] : _vm._e(), _c('div', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: !_vm.searchData.length,
        expression: '!searchData.length',
      }, {
        name: 'bkloading',
        rawName: 'v-bkloading',
        value: {
          isLoading: _vm.isLoading,
        },
        expression: '{ isLoading }',
      }],
      staticClass: 'result-empty',
    }, [_vm._v(`\n      ${_vm._s(_vm.$t('generic.msg.empty.noData1'))}\n    `)])], 2)], 1);
  };

  const __vue_staticRenderFns__$6 = [];
  /* style */

  const __vue_inject_styles__$6 = undefined;
  /* scoped */

  const __vue_scope_id__$6 = 'data-v-cfc7d892';
  /* module identifier */

  const __vue_module_identifier__$6 = undefined;
  /* functional template */

  const __vue_is_functional_template__$6 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$6 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$6,
    staticRenderFns: __vue_staticRenderFns__$6,
  }, __vue_inject_styles__$6, __vue_script__$6, __vue_scope_id__$6, __vue_is_functional_template__$6, __vue_module_identifier__$6, false, undefined, undefined, undefined);

  let _dec$5; let _dec2$5; let _dec3$5; let _dec4$3; let _dec5$3; let _dec6$3; let _dec7$3; let _dec8$3; let _dec9$3; let _dec10$3; let _dec11$3; let _dec12$3; let _dec13$3; let _dec14$2; let _dec15$2; let _dec16$2; let _dec17$2; let _dec18$2; let _dec19$1; let _dec20$1; let _dec21$1; let _dec22$1; let _dec23$1; let _dec24$1; let _dec25$1; let _dec26$1; let _dec27; let _dec28; let _dec29; let _dec30; let _dec31; let _dec32; let _dec33; let _dec34; let _class$5; let _class2$5; let _descriptor$5; let _descriptor2$5; let _descriptor3$3; let _descriptor4$3; let _descriptor5$3; let _descriptor6$3; let _descriptor7$3; let _descriptor8$2; let _descriptor9$2; let _descriptor10$2; let _descriptor11$2; let _descriptor12$2; let _descriptor13$1; let _descriptor14$1; let _descriptor15$1; let _descriptor16$1; let _descriptor17$1; let _descriptor18$1; let _descriptor19$1; let _descriptor20$1; let _descriptor21$1; let _descriptor22$1; let _descriptor23$1; let _descriptor24; let _descriptor25; let _descriptor26; let _descriptor27; let _descriptor28; let _descriptor29;

  function ownKeys(object, enumerableOnly) {
    const keys = Object.keys(object); if (Object.getOwnPropertySymbols) {
      let symbols = Object.getOwnPropertySymbols(object); if (enumerableOnly) {
        symbols = symbols.filter(sym => Object.getOwnPropertyDescriptor(object, sym).enumerable);
      } keys.push.apply(keys, symbols);
    } return keys;
  }

  function _objectSpread(target) {
    for (let i = 1; i < arguments.length; i++) {
      var source = arguments[i] != null ? arguments[i] : {}; if (i % 2) {
        ownKeys(Object(source), true).forEach((key) => {
          _defineProperty__default.default(target, key, source[key]);
        });
      } else if (Object.getOwnPropertyDescriptors) {
        Object.defineProperties(target, Object.getOwnPropertyDescriptors(source));
      } else {
        ownKeys(Object(source)).forEach((key) => {
          Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key));
        });
      }
    } return target;
  }

  function _createSuper$5(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$5(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$5() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const DynamicTopo = (_dec$5 = Component({
    name: 'dynamic-topo',
    components: {
      TopoTree: __vue_component__$7,
      TopoSearch: __vue_component__$6,
      IpListTable: __vue_component__$9,
    },
    directives: {
      resize,
    },
  }), _dec2$5 = Prop({
    type: Function,
    required: true,
  }), _dec3$5 = Prop({
    type: Function,
    required: true,
  }), _dec4$3 = Prop({
    type: Function,
  }), _dec5$3 = Prop({
    type: Function,
  }), _dec6$3 = Prop({
    type: Function,
  }), _dec7$3 = Prop({
    type: Function,
  }), _dec8$3 = Prop({
    type: [Function, Boolean],
  }), _dec9$3 = Prop({
    default: false,
    type: Boolean,
  }), _dec10$3 = Prop({
    default: 'auto',
    type: [Number, String],
  }), _dec11$3 = Prop({
    default: function _default() {
      return {};
    },
    type: Object,
  }), _dec12$3 = Prop({
    default: function _default() {
      return {};
    },
    type: Object,
  }), _dec13$3 = Prop({
    default: 20,
    type: Number,
  }), _dec14$2 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec15$2 = Prop({
    default: '',
    type: String,
  }), _dec16$2 = Prop({
    default: true,
    type: Boolean,
  }), _dec17$2 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec18$2 = Prop({
    default: false,
    type: Boolean,
  }), _dec19$1 = Prop({
    default: 2,
    type: Number,
  }), _dec20$1 = Prop({
    default: 240,
    type: [Number, String],
  }), _dec21$1 = Prop({
    default: '',
    type: [String, Number],
  }), _dec22$1 = Prop({
    default: 'rtl',
    type: String,
  }), _dec23$1 = Prop({
    default: true,
    type: Boolean,
  }), _dec24$1 = Prop({
    default: false,
    type: Boolean,
  }), _dec25$1 = Prop({
    default: false,
    type: Boolean,
  }), _dec26$1 = Prop({
    type: Function,
  }), _dec27 = Prop({
    type: Function,
  }), _dec28 = Ref('table'), _dec29 = Ref('leftWrapper'), _dec30 = Ref('tree'), _dec31 = Watch('defaultCheckedNodes'), _dec32 = Watch('selections'), _dec33 = Emit('search-selection-change'), _dec34 = Emit('check-change'), _dec$5(_class$5 = (_class2$5 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(DynamicTopo, _Vue);

    const _super = _createSuper$5(DynamicTopo);

    function DynamicTopo() {
      let _this;

      _classCallCheck__default.default(this, DynamicTopo);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'getDefaultData', _descriptor$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getSearchTableData', _descriptor2$5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getSearchTreeData', _descriptor3$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getDefaultSelections', _descriptor4$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getSearchResultSelections', _descriptor5$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'lazyMethod', _descriptor6$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'lazyDisabled', _descriptor7$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'expandOnClick', _descriptor8$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'resultWidth', _descriptor9$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'searchDataOptions', _descriptor10$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'treeDataOptions', _descriptor11$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'limit', _descriptor12$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'dynamicTableConfig', _descriptor13$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'dynamicTablePlaceholder', _descriptor14$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'transformToChildren', _descriptor15$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultCheckedNodes', _descriptor16$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'showCount', _descriptor17$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultExpandLevel', _descriptor18$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'leftPanelWidth', _descriptor19$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultSelectedNode', _descriptor20$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'ellipsisDirection', _descriptor21$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'enableSearchPanel', _descriptor22$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'enableTreeFilter', _descriptor23$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'acrossPage', _descriptor24, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowDisabledStatus', _descriptor25, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowTipsContent', _descriptor26, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tableRef', _descriptor27, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'leftWrapperRef', _descriptor28, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'treeRef', _descriptor29, _assertThisInitialized__default.default(_this));

      _this.isLoading = false;
      _this.treeKeyword = '';
      _this.treeHeight = 300;
      _this.emptyText = '';
      _this.nodes = [];
      _this.selections = [];
      _this.parentNode = null;
      _this.defaultSelectionIds = [];
      _this.searchPanelData = [];
      return _this;
    }

    _createClass__default.default(DynamicTopo, [{
      key: 'handleDefaultCheckedNodesChange',
      value: function handleDefaultCheckedNodesChange(data) {
        this.treeRef && this.treeRef.handleSetChecked(data);
      },
    }, {
      key: 'handleSelectionChange',
      value: function handleSelectionChange() {
        this.emptyText = !!this.selections.length ? this.$t('generic.msg.empty.noData2') : this.$t('generic.placeholder.select');
      },
    }, {
      key: 'created',
      value: function created() {
        this.emptyText = this.$t('generic.placeholder.select');
        this.handleGetDefaultData();
      },
    }, {
      key: 'mounted',
      value: function mounted() {
        this.treeHeight = this.leftWrapperRef.clientHeight - 32;
      },
    }, {
      key: 'handleGetDefaultData',
      value: (function () {
        const _handleGetDefaultData = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee() {
          let data;
          return _regeneratorRuntime__default.default.wrap(function _callee$(_context) {
            while (1) {
              switch (_context.prev = _context.next) {
                case 0:
                  _context.prev = 0;
                  this.isLoading = true;
                  _context.next = 4;
                  return this.getDefaultData();

                case 4:
                  data = _context.sent;
                  this.nodes = data || [];
                  _context.next = 11;
                  break;

                case 8:
                  _context.prev = 8;
                  _context.t0 = _context.catch(0);
                  console.log(_context.t0);

                case 11:
                  _context.prev = 11;
                  this.isLoading = false;
                  return _context.finish(11);

                case 14:
                case 'end':
                  return _context.stop();
              }
            }
          }, _callee, this, [[0, 8, 11, 14]]);
        }));

        function handleGetDefaultData() {
          return _handleGetDefaultData.apply(this, arguments);
        }

        return handleGetDefaultData;
      }()), // 搜索结果勾选事件

    }, {
      key: 'handleCheckChange',
      value: function handleCheckChange(data) {
        return data; // this.selections = selections
        // this.tableRef.handleGetDefaultData('selection-change')
      }, // 树select事件

    }, {
      key: 'handleSelectChange',
      value: function handleSelectChange(treeNode) {
        if (this.transformToChildren) {
          const _this$treeDataOptions = this.treeDataOptions.childrenKey;
          const childrenKey = _this$treeDataOptions === void 0 ? 'children' : _this$treeDataOptions;
          this.selections = treeNode.data[childrenKey] && treeNode.data[childrenKey].length ? treeNode.data[childrenKey] : treeNode.children.map(node => node.data);
          this.parentNode = treeNode;
        } else {
          this.selections = [treeNode.data];
          this.parentNode = treeNode.parent || null;
        }

        this.tableRef.handleGetDefaultData('selection-change');
        this.tableRef.clearTableKeyWord();
      },
    }, {
      key: 'handleSearchPanelShow',
      value: function handleSearchPanelShow() {
        const _this2 = this;

        if (this.getSearchResultSelections) {
          const _this$searchDataOptio = this.searchDataOptions.idKey;
          const idKey = _this$searchDataOptio === void 0 ? 'id' : _this$searchDataOptio;
          this.defaultSelectionIds = this.searchPanelData.reduce((pre, next) => {
            !!_this2.getSearchResultSelections(next) && pre.push(next[idKey]);
            return pre;
          }, []);
        }
      }, // 树搜索

    }, {
      key: 'searchTreeMethod',
      value: (function () {
        const _searchTreeMethod = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee2(treeKeyword) {
          return _regeneratorRuntime__default.default.wrap(function _callee2$(_context2) {
            while (1) {
              switch (_context2.prev = _context2.next) {
                case 0:
                  _context2.prev = 0;

                  if (!this.getSearchTreeData) {
                    _context2.next = 7;
                    break;
                  }

                  _context2.next = 4;
                  return this.getSearchTreeData({
                    treeKeyword,
                  });

                case 4:
                  this.searchPanelData = _context2.sent;
                  _context2.next = 8;
                  break;

                case 7:
                  this.searchPanelData = this.defaultTreeSearchMethod(this.nodes, '', treeKeyword);

                case 8:
                  this.handleSearchPanelShow();
                  return _context2.abrupt('return', this.searchPanelData);

                case 12:
                  _context2.prev = 12;
                  _context2.t0 = _context2.catch(0);
                  console.log(_context2.t0);
                  return _context2.abrupt('return', {
                    total: 0,
                    data: [],
                  });

                case 16:
                case 'end':
                  return _context2.stop();
              }
            }
          }, _callee2, this, [[0, 12]]);
        }));

        function searchTreeMethod(_x) {
          return _searchTreeMethod.apply(this, arguments);
        }

        return searchTreeMethod;
      }()), // 树默认搜索方法(结果是打平的数据)

    }, {
      key: 'defaultTreeSearchMethod',
      value: function defaultTreeSearchMethod(nodes, parent, treeKeyword) {
        const _this3 = this;

        const _this$treeDataOptions2 = this.treeDataOptions;
        const _this$treeDataOptions3 = _this$treeDataOptions2.nameKey;
        const nameKey = _this$treeDataOptions3 === void 0 ? 'name' : _this$treeDataOptions3;
        const _this$treeDataOptions4 = _this$treeDataOptions2.childrenKey;
        const childrenKey = _this$treeDataOptions4 === void 0 ? 'children' : _this$treeDataOptions4;
        const _this$searchDataOptio2 = this.searchDataOptions.pathKey;
        const pathKey = _this$searchDataOptio2 === void 0 ? 'node_path' : _this$searchDataOptio2;
        return nodes.reduce((pre, next) => {
          if (next[nameKey].includes(treeKeyword)) {
            pre.push(next);
          }

          if (next[childrenKey] && next[childrenKey].length) {
            pre.push.apply(pre, _toConsumableArray__default.default(_this3.defaultTreeSearchMethod(next[childrenKey], next[nameKey], treeKeyword)));
          }

          if (!next[pathKey]) {
            next[pathKey] = parent ? ''.concat(parent, ' / ').concat(next[nameKey]) : next[nameKey];
          }

          return pre;
        }, []);
      },
    }, {
      key: 'getTableData',
      value: (function () {
        const _getTableData = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee3(params, type) {
          let reqParams;
          return _regeneratorRuntime__default.default.wrap(function _callee3$(_context3) {
            while (1) {
              switch (_context3.prev = _context3.next) {
                case 0:
                  _context3.prev = 0;
                  reqParams = _objectSpread({
                    selections: this.selections,
                    parentNode: this.parentNode,
                  }, params);
                  _context3.next = 4;
                  return this.getSearchTableData(reqParams, type);

                case 4:
                  return _context3.abrupt('return', _context3.sent);

                case 7:
                  _context3.prev = 7;
                  _context3.t0 = _context3.catch(0);
                  console.log(_context3.t0);
                  return _context3.abrupt('return', {
                    total: 0,
                    data: [],
                  });

                case 11:
                case 'end':
                  return _context3.stop();
              }
            }
          }, _callee3, this, [[0, 7]]);
        }));

        function getTableData(_x2, _x3) {
          return _getTableData.apply(this, arguments);
        }

        return getTableData;
      }()),
    }, {
      key: 'handleTableCheckChange',
      value: function handleTableCheckChange(data) {
        return data;
      }, // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'handleGetDefaultSelections',
      value: function handleGetDefaultSelections() {
        this.tableRef && this.tableRef.handleGetDefaultSelections();
      },
    }]);

    return DynamicTopo;
  }(Vue__default.default)), (_descriptor$5 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'getDefaultData', [_dec2$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$5 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'getSearchTableData', [_dec3$5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$3 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'getSearchTreeData', [_dec4$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$3 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'getDefaultSelections', [_dec5$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$3 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'getSearchResultSelections', [_dec6$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$3 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'lazyMethod', [_dec7$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7$3 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'lazyDisabled', [_dec8$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor8$2 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'expandOnClick', [_dec9$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor9$2 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'resultWidth', [_dec10$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor10$2 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'searchDataOptions', [_dec11$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor11$2 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'treeDataOptions', [_dec12$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor12$2 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'limit', [_dec13$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor13$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'dynamicTableConfig', [_dec14$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor14$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'dynamicTablePlaceholder', [_dec15$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor15$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'transformToChildren', [_dec16$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor16$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'defaultCheckedNodes', [_dec17$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor17$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'showCount', [_dec18$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor18$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'defaultExpandLevel', [_dec19$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor19$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'leftPanelWidth', [_dec20$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor20$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'defaultSelectedNode', [_dec21$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor21$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'ellipsisDirection', [_dec22$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor22$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'enableSearchPanel', [_dec23$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor23$1 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'enableTreeFilter', [_dec24$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor24 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'acrossPage', [_dec25$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor25 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'getRowDisabledStatus', [_dec26$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor26 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'getRowTipsContent', [_dec27], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor27 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'tableRef', [_dec28], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor28 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'leftWrapperRef', [_dec29], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor29 = _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'treeRef', [_dec30], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'handleDefaultCheckedNodesChange', [_dec31], Object.getOwnPropertyDescriptor(_class2$5.prototype, 'handleDefaultCheckedNodesChange'), _class2$5.prototype), _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'handleSelectionChange', [_dec32], Object.getOwnPropertyDescriptor(_class2$5.prototype, 'handleSelectionChange'), _class2$5.prototype), _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'handleCheckChange', [_dec33], Object.getOwnPropertyDescriptor(_class2$5.prototype, 'handleCheckChange'), _class2$5.prototype), _applyDecoratedDescriptor__default.default(_class2$5.prototype, 'handleTableCheckChange', [_dec34], Object.getOwnPropertyDescriptor(_class2$5.prototype, 'handleTableCheckChange'), _class2$5.prototype)), _class2$5)) || _class$5);

  /* script */
  const __vue_script__$5 = DynamicTopo;
  /* template */

  const __vue_render__$5 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      directives: [{
        name: 'bkloading',
        rawName: 'v-bkloading',
        value: {
          isLoading: _vm.isLoading,
        },
        expression: '{ isLoading }',
      }],
      staticClass: 'dynamic-topo',
    }, [_c('div', {
      ref: 'leftWrapper',
      staticClass: 'dynamic-topo-left',
      style: {
        width: isNaN(_vm.leftPanelWidth) ? _vm.leftPanelWidth : `${_vm.leftPanelWidth}px`,
      },
    }, [_c('TopoSearch', {
      attrs: {
        'search-method': _vm.searchTreeMethod,
        placeholder: _vm.$t('generic.ipSelector.placeholder.searchTopo'),
        'result-width': _vm.resultWidth,
        options: _vm.searchDataOptions,
        'default-selection-ids': _vm.defaultSelectionIds,
        'enable-search-panel': _vm.enableSearchPanel,
      },
      on: {
        show: _vm.handleSearchPanelShow,
        'check-change': _vm.handleCheckChange,
      },
      model: {
        value: _vm.treeKeyword,
        callback: function callback($$v) {
          _vm.treeKeyword = $$v;
        },
        expression: 'treeKeyword',
      },
    }), _vm.nodes.length ? _c('TopoTree', {
      ref: 'tree',
      staticClass: 'topo-tree',
      attrs: {
        'default-checked-nodes': _vm.defaultCheckedNodes,
        options: _vm.treeDataOptions,
        nodes: _vm.nodes,
        height: 0,
        'show-count': _vm.showCount,
        'lazy-method': _vm.lazyMethod,
        'lazy-disabled': _vm.lazyDisabled,
        'default-expand-level': _vm.defaultExpandLevel,
        'expand-on-click': _vm.expandOnClick,
        'default-selected-node': _vm.defaultSelectedNode,
        filter: _vm.treeKeyword,
        'enable-tree-filter': _vm.enableTreeFilter,
      },
      on: {
        'select-change': _vm.handleSelectChange,
      },
    }) : _vm._e()], 1), _c('div', {
      staticClass: 'dynamic-topo-right ml10',
    }, [_c('IpListTable', {
      ref: 'table',
      attrs: {
        'get-search-table-data': _vm.getTableData,
        'ip-list-table-config': _vm.dynamicTableConfig,
        'ip-list-placeholder': _vm.dynamicTablePlaceholder,
        'get-default-selections': _vm.getDefaultSelections,
        'disabled-loading': _vm.isLoading,
        'empty-text': _vm.emptyText,
        'ellipsis-direction': _vm.ellipsisDirection,
        'across-page': _vm.acrossPage,
        'get-row-disabled-status': _vm.getRowDisabledStatus,
        'get-row-tips-content': _vm.getRowTipsContent,
      },
      on: {
        'check-change': _vm.handleTableCheckChange,
      },
    })], 1)]);
  };

  const __vue_staticRenderFns__$5 = [];
  /* style */

  const __vue_inject_styles__$5 = undefined;
  /* scoped */

  const __vue_scope_id__$5 = 'data-v-661929ef';
  /* module identifier */

  const __vue_module_identifier__$5 = undefined;
  /* functional template */

  const __vue_is_functional_template__$5 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$5 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$5,
    staticRenderFns: __vue_staticRenderFns__$5,
  }, __vue_inject_styles__$5, __vue_script__$5, __vue_scope_id__$5, __vue_is_functional_template__$5, __vue_module_identifier__$5, false, undefined, undefined, undefined);

  let _dec$4; let _dec2$4; let _dec3$4; let _dec4$2; let _dec5$2; let _dec6$2; let _dec7$2; let _dec8$2; let _dec9$2; let _dec10$2; let _dec11$2; let _dec12$2; let _dec13$2; let _dec14$1; let _dec15$1; let _dec16$1; let _dec17$1; let _dec18$1; let _dec19; let _dec20; let _dec21; let _dec22; let _dec23; let _dec24; let _dec25; let _dec26; let _class$4; let _class2$4; let _descriptor$4; let _descriptor2$4; let _descriptor3$2; let _descriptor4$2; let _descriptor5$2; let _descriptor6$2; let _descriptor7$2; let _descriptor8$1; let _descriptor9$1; let _descriptor10$1; let _descriptor11$1; let _descriptor12$1; let _descriptor13; let _descriptor14; let _descriptor15; let _descriptor16; let _descriptor17; let _descriptor18; let _descriptor19; let _descriptor20; let _descriptor21; let _descriptor22; let _descriptor23;

  function _createSuper$4(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$4(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$4() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const StaticTopo = (_dec$4 = Component({
    name: 'static-topo',
    components: {
      DynamicTopo: __vue_component__$5,
    },
  }), _dec2$4 = Prop({
    type: Function,
    required: true,
  }), _dec3$4 = Prop({
    type: Function,
    required: true,
  }), _dec4$2 = Prop({
    type: Function,
  }), _dec5$2 = Prop({
    type: Function,
  }), _dec6$2 = Prop({
    default: 'auto',
    type: [Number, String],
  }), _dec7$2 = Prop({
    default: function _default() {
      return {};
    },
    type: Object,
  }), _dec8$2 = Prop({
    default: function _default() {
      return {};
    },
    type: Object,
  }), _dec9$2 = Prop({
    default: 20,
    type: Number,
  }), _dec10$2 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec11$2 = Prop({
    default: '',
    type: String,
  }), _dec12$2 = Prop({
    type: Function,
  }), _dec13$2 = Prop({
    type: [Function, Boolean],
  }), _dec14$1 = Prop({
    default: 2,
    type: Number,
  }), _dec15$1 = Prop({
    default: false,
    type: Boolean,
  }), _dec16$1 = Prop({
    default: 240,
    type: [Number, String],
  }), _dec17$1 = Prop({
    default: '',
    type: [String, Number],
  }), _dec18$1 = Prop({
    default: 'rtl',
    type: String,
  }), _dec19 = Prop({
    default: true,
    type: Boolean,
  }), _dec20 = Prop({
    default: false,
    type: Boolean,
  }), _dec21 = Prop({
    default: false,
    type: Boolean,
  }), _dec22 = Prop({
    type: Function,
  }), _dec23 = Prop({
    type: Function,
  }), _dec24 = Ref('dynamicTopo'), _dec25 = Emit('check-change'), _dec26 = Emit('search-selection-change'), _dec$4(_class$4 = (_class2$4 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(StaticTopo, _Vue);

    const _super = _createSuper$4(StaticTopo);

    function StaticTopo() {
      let _this;

      _classCallCheck__default.default(this, StaticTopo);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'getDefaultData', _descriptor$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getSearchTableData', _descriptor2$4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getSearchTreeData', _descriptor3$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getDefaultSelections', _descriptor4$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'resultWidth', _descriptor5$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'searchDataOptions', _descriptor6$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'treeDataOptions', _descriptor7$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'limit', _descriptor8$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'staticTableConfig', _descriptor9$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'staticTablePlaceholder', _descriptor10$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'lazyMethod', _descriptor11$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'lazyDisabled', _descriptor12$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultExpandLevel', _descriptor13, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'expandOnClick', _descriptor14, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'leftPanelWidth', _descriptor15, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultSelectedNode', _descriptor16, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'ellipsisDirection', _descriptor17, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'enableSearchPanel', _descriptor18, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'enableTreeFilter', _descriptor19, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'acrossPage', _descriptor20, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowDisabledStatus', _descriptor21, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'getRowTipsContent', _descriptor22, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'dynamicTopoRef', _descriptor23, _assertThisInitialized__default.default(_this));

      return _this;
    }

    _createClass__default.default(StaticTopo, [{
      key: 'handleTableCheckChange',
      value: function handleTableCheckChange(data) {
        return data;
      },
    }, {
      key: 'handleSearchSelectionChange',
      value: function handleSearchSelectionChange(selections) {
        return selections;
      }, // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'handleGetDefaultSelections',
      value: function handleGetDefaultSelections() {
        this.dynamicTopoRef && this.dynamicTopoRef.handleGetDefaultSelections();
      },
    }]);

    return StaticTopo;
  }(Vue__default.default)), (_descriptor$4 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'getDefaultData', [_dec2$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$4 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'getSearchTableData', [_dec3$4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$2 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'getSearchTreeData', [_dec4$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$2 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'getDefaultSelections', [_dec5$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$2 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'resultWidth', [_dec6$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$2 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'searchDataOptions', [_dec7$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7$2 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'treeDataOptions', [_dec8$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor8$1 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'limit', [_dec9$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor9$1 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'staticTableConfig', [_dec10$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor10$1 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'staticTablePlaceholder', [_dec11$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor11$1 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'lazyMethod', [_dec12$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor12$1 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'lazyDisabled', [_dec13$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor13 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'defaultExpandLevel', [_dec14$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor14 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'expandOnClick', [_dec15$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor15 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'leftPanelWidth', [_dec16$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor16 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'defaultSelectedNode', [_dec17$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor17 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'ellipsisDirection', [_dec18$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor18 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'enableSearchPanel', [_dec19], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor19 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'enableTreeFilter', [_dec20], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor20 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'acrossPage', [_dec21], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor21 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'getRowDisabledStatus', [_dec22], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor22 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'getRowTipsContent', [_dec23], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor23 = _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'dynamicTopoRef', [_dec24], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'handleTableCheckChange', [_dec25], Object.getOwnPropertyDescriptor(_class2$4.prototype, 'handleTableCheckChange'), _class2$4.prototype), _applyDecoratedDescriptor__default.default(_class2$4.prototype, 'handleSearchSelectionChange', [_dec26], Object.getOwnPropertyDescriptor(_class2$4.prototype, 'handleSearchSelectionChange'), _class2$4.prototype)), _class2$4)) || _class$4);

  /* script */
  const __vue_script__$4 = StaticTopo;
  /* template */

  const __vue_render__$4 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('DynamicTopo', {
      ref: 'dynamicTopo',
      attrs: {
        'get-default-data': _vm.getDefaultData,
        'get-search-table-data': _vm.getSearchTableData,
        'get-search-tree-data': _vm.getSearchTreeData,
        'result-width': _vm.resultWidth,
        'search-data-options': _vm.searchDataOptions,
        'tree-data-options': _vm.treeDataOptions,
        limit: _vm.limit,
        'transform-to-children': false,
        'show-count': false,
        'dynamic-table-config': _vm.staticTableConfig,
        'get-default-selections': _vm.getDefaultSelections,
        'lazy-method': _vm.lazyMethod,
        'lazy-disabled': _vm.lazyDisabled,
        'default-expand-level': _vm.defaultExpandLevel,
        'expand-on-click': _vm.expandOnClick,
        'left-panel-width': _vm.leftPanelWidth,
        'default-selected-node': _vm.defaultSelectedNode,
        'ellipsis-direction': _vm.ellipsisDirection,
        'enable-tree-filter': _vm.enableTreeFilter,
        'enable-search-panel': _vm.enableSearchPanel,
        'dynamic-table-placeholder': _vm.staticTablePlaceholder,
        'across-page': _vm.acrossPage,
        'get-row-disabled-status': _vm.getRowDisabledStatus,
        'get-row-tips-content': _vm.getRowTipsContent,
      },
      on: {
        'check-change': _vm.handleTableCheckChange,
        'search-selection-change': _vm.handleSearchSelectionChange,
      },
    });
  };

  const __vue_staticRenderFns__$4 = [];
  /* style */

  const __vue_inject_styles__$4 = undefined;
  /* scoped */

  const __vue_scope_id__$4 = undefined;
  /* module identifier */

  const __vue_module_identifier__$4 = undefined;
  /* functional template */

  const __vue_is_functional_template__$4 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$4 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$4,
    staticRenderFns: __vue_staticRenderFns__$4,
  }, __vue_inject_styles__$4, __vue_script__$4, __vue_scope_id__$4, __vue_is_functional_template__$4, __vue_module_identifier__$4, false, undefined, undefined, undefined);

  let _dec$3; let _dec2$3; let _dec3$3; let _class$3; let _class2$3; let _descriptor$3; let _descriptor2$3;

  function _createSuper$3(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$3(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$3() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }

  const SelectorContent = (_dec$3 = Component({
    name: 'selector-content',
  }), _dec2$3 = Prop({
    default: '',
    type: String,
  }), _dec3$3 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec$3(_class$3 = (_class2$3 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(SelectorContent, _Vue);

    const _super = _createSuper$3(SelectorContent);

    function SelectorContent() {
      let _this;

      _classCallCheck__default.default(this, SelectorContent);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'active', _descriptor$3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'panels', _descriptor2$3, _assertThisInitialized__default.default(_this));

      _this.components = {
        'static-topo': __vue_component__$4,
        'custom-input': __vue_component__$8,
      };
      return _this;
    }

    _createClass__default.default(SelectorContent, [{
      key: 'currentComponent',
      get: function get() {
        const _this2 = this;

        const panel = this.panels.find(item => item.name === _this2.active);
        return panel !== null && panel !== void 0 && panel.component ? panel.component : this.components[this.active];
      },
    }, {
      key: 'include',
      get: function get() {
        return this.panels.reduce((pre, next) => {
          if (next.keepAlive) {
            pre.push(next.name);
          }

          return pre;
        }, []);
      },
    }, {
      key: 'handleGetDefaultSelections',
      value: function handleGetDefaultSelections() {
        try {
          this.$refs.layout.handleGetDefaultSelections();
        } catch (err) {
          console.log(err);
        }
      },
    }]);

    return SelectorContent;
  }(Vue__default.default)), (_descriptor$3 = _applyDecoratedDescriptor__default.default(_class2$3.prototype, 'active', [_dec2$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$3 = _applyDecoratedDescriptor__default.default(_class2$3.prototype, 'panels', [_dec3$3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  })), _class2$3)) || _class$3);

  /* script */
  const __vue_script__$3 = SelectorContent;
  /* template */

  const __vue_render__$3 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      staticClass: 'selector-content',
    }, [_c('keep-alive', {
      attrs: {
        include: _vm.include,
      },
    }, [_c(_vm.currentComponent, _vm._g(_vm._b({
      ref: 'layout',
      tag: 'component',
      staticClass: 'layout',
    }, 'component', _vm.$attrs, false), _vm.$listeners))], 1)], 1);
  };

  const __vue_staticRenderFns__$3 = [];
  /* style */

  const __vue_inject_styles__$3 = undefined;
  /* scoped */

  const __vue_scope_id__$3 = 'data-v-3092d424';
  /* module identifier */

  const __vue_module_identifier__$3 = undefined;
  /* functional template */

  const __vue_is_functional_template__$3 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$3 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$3,
    staticRenderFns: __vue_staticRenderFns__$3,
  }, __vue_inject_styles__$3, __vue_script__$3, __vue_scope_id__$3, __vue_is_functional_template__$3, __vue_module_identifier__$3, false, undefined, undefined, undefined);

  let _dec$2; let _dec2$2; let _dec3$2; let _dec4$1; let _dec5$1; let _dec6$1; let _dec7$1; let _dec8$1; let _dec9$1; let _dec10$1; let _dec11$1; let _dec12$1; let _dec13$1; let _class$2; let _class2$2; let _descriptor$2; let _descriptor2$2; let _descriptor3$1; let _descriptor4$1; let _descriptor5$1; let _descriptor6$1; let _descriptor7$1;

  function _createSuper$2(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$2(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$2() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }

  const SelectorPreview = (_dec$2 = Component({
    name: 'selector-preview',
    components: {
      Menu: __vue_component__$d,
    },
  }), _dec2$2 = Prop({
    default: 280,
    type: [Number, String],
  }), _dec3$2 = Prop({
    default: function _default() {
      return [100, 600];
    },
    type: Array,
  }), _dec4$1 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec5$1 = Prop({
    default: function _default() {
      return [];
    },
    type: [Array, Function],
  }), _dec6$1 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec7$1 = Prop({
    default: '',
    type: String,
  }), _dec8$1 = Ref('menuRef'), _dec9$1 = Watch('width'), _dec10$1 = Debounce(300), _dec11$1 = Emit('update:width'), _dec12$1 = Emit('menu-click'), _dec13$1 = Emit('remove-node'), _dec$2(_class$2 = (_class2$2 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(SelectorPreview, _Vue);

    const _super = _createSuper$2(SelectorPreview);

    function SelectorPreview() {
      let _this;

      _classCallCheck__default.default(this, SelectorPreview);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'width', _descriptor$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'range', _descriptor2$2, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'data', _descriptor3$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'operateList', _descriptor4$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultActiveName', _descriptor5$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'previewTitle', _descriptor6$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'menuRef', _descriptor7$1, _assertThisInitialized__default.default(_this));

      _this.preWidth = 280;
      _this.activeName = [];
      _this.hoverChild = null;
      _this.popoverInstance = null;
      _this.moreOperateList = [];
      _this.previewItem = null;
      return _this;
    }

    _createClass__default.default(SelectorPreview, [{
      key: 'isDataEmpty',
      get: function get() {
        return !this.data.length || this.data.every(item => !item.data.length);
      },
    }, {
      key: 'created',
      value: function created() {
        this.preWidth = this.width;
        this.activeName = this.defaultActiveName;
      },
    }, {
      key: 'handleChange',
      value: function handleChange(width) {
        this.preWidth = width;
      },
    }, {
      key: 'handleWidthChange',
      value: function handleWidthChange() {
        return this.preWidth;
      },
    }, {
      key: 'handleMenuItemClick',
      value: function handleMenuItemClick(menu, item) {
        return {
          menu,
          item,
        };
      },
    }, {
      key: 'removeNode',
      value: function removeNode(child, item) {
        const index = item.data.indexOf(child);
        this.hoverChild = index > -1 && item.data[index + 1] ? item.data[index + 1] : null;
        return {
          child,
          item,
        };
      },
    }, {
      key: 'handleMenuClick',
      value: function handleMenuClick(menu) {
        this.popoverInstance && this.popoverInstance.hide();
        this.handleMenuItemClick(menu, this.previewItem);
      },
    }, {
      key: 'handleShowMenu',
      value: (function () {
        const _handleShowMenu = _asyncToGenerator__default.default(/* #__PURE__*/_regeneratorRuntime__default.default.mark(function _callee(event, item) {
          const _this2 = this;

          let list;
          return _regeneratorRuntime__default.default.wrap(function _callee$(_context) {
            while (1) {
              switch (_context.prev = _context.next) {
                case 0:
                  if (event.target) {
                    _context.next = 2;
                    break;
                  }

                  return _context.abrupt('return');

                case 2:
                  if (!(typeof this.operateList === 'function')) {
                    _context.next = 8;
                    break;
                  }

                  _context.next = 5;
                  return this.operateList(item);

                case 5:
                  _context.t0 = _context.sent;
                  _context.next = 9;
                  break;

                case 8:
                  _context.t0 = this.operateList;

                case 9:
                  list = _context.t0;

                  if (!(!list || !list.length)) {
                    _context.next = 12;
                    break;
                  }

                  return _context.abrupt('return');

                case 12:
                  this.moreOperateList = list;
                  this.previewItem = item;
                  this.popoverInstance = this.$bkPopover(event.target, {
                    content: this.menuRef.$el,
                    trigger: 'manual',
                    arrow: false,
                    theme: 'light ip-selector',
                    maxWidth: 280,
                    offset: '0, 5',
                    sticky: true,
                    duration: [275, 0],
                    interactive: true,
                    boundary: 'window',
                    placement: 'bottom-end',
                    onHidden: function onHidden() {
                      _this2.popoverInstance && _this2.popoverInstance.destroy();
                      _this2.popoverInstance = null;
                    },
                  });
                  this.popoverInstance.show();

                case 16:
                case 'end':
                  return _context.stop();
              }
            }
          }, _callee, this);
        }));

        function handleShowMenu(_x, _x2) {
          return _handleShowMenu.apply(this, arguments);
        }

        return handleShowMenu;
      }()),
    }, {
      key: 'handleMouseDown',
      value: function handleMouseDown(e) {
        const _this3 = this;

        const node = e.target;
        const { parentNode } = node;
        if (!parentNode) return;
        const nodeRect = node.getBoundingClientRect();
        const rect = parentNode.getBoundingClientRect();

        document.onselectstart = function () {
          return false;
        };

        document.ondragstart = function () {
          return false;
        };

        const handleMouseMove = function handleMouseMove(event) {
          const _this3$range = _slicedToArray__default.default(_this3.range, 2);
          const min = _this3$range[0];
          const max = _this3$range[1];

          const newWidth = rect.right - event.clientX + nodeRect.width;

          if (newWidth < min) {
            _this3.preWidth = 0;
          } else {
            _this3.preWidth = Math.min(newWidth, max);
          }

          _this3.handleWidthChange();
        };

        const handleMouseUp = function handleMouseUp() {
          document.body.style.cursor = '';
          document.removeEventListener('mousemove', handleMouseMove);
          document.removeEventListener('mouseup', handleMouseUp);
          document.onselectstart = null;
          document.ondragstart = null;
        };

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
      },
    }]);

    return SelectorPreview;
  }(Vue__default.default)), (_descriptor$2 = _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'width', [_dec2$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$2 = _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'range', [_dec3$2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3$1 = _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'data', [_dec4$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4$1 = _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'operateList', [_dec5$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5$1 = _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'defaultActiveName', [_dec6$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6$1 = _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'previewTitle', [_dec7$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7$1 = _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'menuRef', [_dec8$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'handleChange', [_dec9$1], Object.getOwnPropertyDescriptor(_class2$2.prototype, 'handleChange'), _class2$2.prototype), _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'handleWidthChange', [_dec10$1, _dec11$1], Object.getOwnPropertyDescriptor(_class2$2.prototype, 'handleWidthChange'), _class2$2.prototype), _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'handleMenuItemClick', [_dec12$1], Object.getOwnPropertyDescriptor(_class2$2.prototype, 'handleMenuItemClick'), _class2$2.prototype), _applyDecoratedDescriptor__default.default(_class2$2.prototype, 'removeNode', [_dec13$1], Object.getOwnPropertyDescriptor(_class2$2.prototype, 'removeNode'), _class2$2.prototype)), _class2$2)) || _class$2);

  /* script */
  const __vue_script__$2 = SelectorPreview;
  /* template */

  const __vue_render__$2 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('div', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: isNaN(_vm.preWidth) || _vm.preWidth > 0,
        expression: 'isNaN(preWidth) || preWidth > 0',
      }],
      staticClass: 'selector-preview',
      style: {
        width: isNaN(_vm.preWidth) ? _vm.preWidth : `${_vm.preWidth}px`,
      },
    }, [_c('div', {
      staticClass: 'selector-preview-title',
    }, [_vm._t('title', [_vm._v(_vm._s(_vm.previewTitle))])], 2), _c('div', {
      staticClass: 'selector-preview-content',
    }, [!_vm.isDataEmpty ? [_c('bcs-collapse', {
      model: {
        value: _vm.activeName,
        callback: function callback($$v) {
          _vm.activeName = $$v;
        },
        expression: 'activeName',
      },
    }, _vm._l(_vm.data, item => _c('bcs-collapse-item', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: item.data && item.data.length,
        expression: 'item.data && item.data.length',
      }],
      key: item.id,
      attrs: {
        name: item.id,
        'hide-arrow': '',
      },
      scopedSlots: _vm._u([{
        key: 'default',
        fn: function fn() {
          return [_c('div', {
            staticClass: 'collapse-title',
          }, [_c('span', {
            staticClass: 'collapse-title-left',
          }, [_c('i', {
            class: ['bk-icon icon-angle-right', {
              expand: _vm.activeName.includes(item.id),
            }],
          }), _vm._t('collapse-title', [_vm._v(`\n                  ${_vm._s(item.name)}\n                `)], null, {
            item,
          })], 2), _c('span', {
            staticClass: 'collapse-title-right',
            on: {
              click: function click($event) {
                $event.stopPropagation();
                return _vm.handleShowMenu($event, item);
              },
            },
          }, [_c('i', {
            staticClass: 'bk-icon icon-more',
          })])])];
        },
        proxy: true,
      }, {
        key: 'content',
        fn: function fn() {
          return [_vm._t('collapse-content', [_c('ul', {
            staticClass: 'collapse-content',
          }, _vm._l(item.data, (child, index) => _c('li', {
            key: index,
            staticClass: 'collapse-content-item',
            on: {
              mouseenter: function mouseenter($event) {
                $event.stopPropagation();
                _vm.hoverChild = child;
              },
              mouseleave: function mouseleave($event) {
                $event.stopPropagation();
                _vm.hoverChild = null;
              },
            },
          }, [_c('span', {
            staticClass: 'left',
            attrs: {
              title: child[item.dataNameKey] || child.name || '--',
            },
          }, [_vm._v(`\n                    ${_vm._s(child[item.dataNameKey] || child.name || '--')}\n                  `)]), _c('span', {
            directives: [{
              name: 'show',
              rawName: 'v-show',
              value: _vm.hoverChild === child,
              expression: 'hoverChild === child',
            }],
            staticClass: 'right',
            on: {
              click: function click($event) {
                return _vm.removeNode(child, item);
              },
            },
          }, [_c('i', {
            staticClass: 'bk-icon icon-close-line',
          })])])), 0)], null, {
            item,
          })];
        },
        proxy: true,
      }], null, true),
    })), 1)] : [_c('bk-exception', {
      staticClass: 'empty',
      attrs: {
        type: 'empty',
        scene: 'part',
      },
    }, [_c('span', {
      staticClass: 'empty-text',
    }, [_vm._v(_vm._s(_vm.$t('generic.ipSelector.selected.emptyMsg')))])])]], 2), _c('div', {
      staticClass: 'drag',
      on: {
        mousedown: _vm.handleMouseDown,
      },
    }), _c('div', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: false,
        expression: 'false',
      }],
    }, [_c('Menu', {
      ref: 'menuRef',
      attrs: {
        theme: 'primary',
        list: _vm.moreOperateList,
      },
      on: {
        click: _vm.handleMenuClick,
      },
      scopedSlots: _vm._u([{
        key: 'default',
        fn: function fn(ref) {
          const { item } = ref;
          return _c('div', {
            staticClass: 'operate-item',
          }, [_c('span', {
            staticClass: 'operate-item-label',
          }, [_vm._v(_vm._s(item.label))])]);
        },
      }]),
    })], 1)]);
  };

  const __vue_staticRenderFns__$2 = [];
  /* style */

  const __vue_inject_styles__$2 = undefined;
  /* scoped */

  const __vue_scope_id__$2 = 'data-v-5af55c52';
  /* module identifier */

  const __vue_module_identifier__$2 = undefined;
  /* functional template */

  const __vue_is_functional_template__$2 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$2 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$2,
    staticRenderFns: __vue_staticRenderFns__$2,
  }, __vue_inject_styles__$2, __vue_script__$2, __vue_scope_id__$2, __vue_is_functional_template__$2, __vue_module_identifier__$2, false, undefined, undefined, undefined);

  let _dec$1; let _dec2$1; let _dec3$1; let _dec4; let _dec5; let _dec6; let _dec7; let _dec8; let _dec9; let _dec10; let _dec11; let _dec12; let _dec13; let _dec14; let _dec15; let _dec16; let _dec17; let _dec18; let _class$1; let _class2$1; let _descriptor$1; let _descriptor2$1; let _descriptor3; let _descriptor4; let _descriptor5; let _descriptor6; let _descriptor7; let _descriptor8; let _descriptor9; let _descriptor10; let _descriptor11; let _descriptor12;

  function _createSuper$1(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct$1(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct$1() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const IpSelector = (_dec$1 = Component({
    name: 'ip-selector',
    inheritAttrs: false,
    components: {
      SelectorTab: __vue_component__$c,
      SelectorContent: __vue_component__$3,
      SelectorPreview: __vue_component__$2,
    },
  }), _dec2$1 = Prop({
    default: '',
    type: String,
  }), _dec3$1 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
    required: true,
  }), _dec4 = Prop({
    default: true,
    type: Boolean,
  }), _dec5 = Prop({
    default: 280,
    type: [Number, String],
  }), _dec6 = Prop({
    default: function _default() {
      return [150, 600];
    },
    type: Array,
  }), _dec7 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec8 = Prop({
    default: function _default() {
      return [];
    },
    type: [Array, Function],
  }), _dec9 = Prop({
    default: '',
    type: [Number, String],
  }), _dec10 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec11 = Prop({
    default: '',
    type: String,
  }), _dec12 = Ref('tab'), _dec13 = Ref('preview'), _dec14 = Watch('active'), _dec15 = Emit('tab-change'), _dec16 = Emit('update:active'), _dec17 = Emit('menu-click'), _dec18 = Emit('remove-node'), _dec$1(_class$1 = (_class2$1 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(IpSelector, _Vue);

    const _super = _createSuper$1(IpSelector);

    function IpSelector() {
      let _this;

      _classCallCheck__default.default(this, IpSelector);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'active', _descriptor$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'panels', _descriptor2$1, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tabVisible', _descriptor3, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'previewWidth', _descriptor4, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'previewRange', _descriptor5, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'previewData', _descriptor6, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'previewOperateList', _descriptor7, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'height', _descriptor8, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'defaultActiveName', _descriptor9, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'previewTitle', _descriptor10, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'tabRef', _descriptor11, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'previewRef', _descriptor12, _assertThisInitialized__default.default(_this));

      _this.panelActive = '';
      _this.width = 280;
      _this.excludeEvents = ['tab-change', 'menu-click', 'remove-node'];
      return _this;
    }

    _createClass__default.default(IpSelector, [{
      key: 'contentEvents',
      get: // 不能丢到layout组件的事件
      function get() {
        const _this2 = this;

        return Object.keys(this.$listeners).reduce((pre, key) => {
          if (_this2.excludeEvents.includes(key)) return pre;

          pre[key] = function () {
            for (var _len2 = arguments.length, args = new Array(_len2), _key2 = 0; _key2 < _len2; _key2++) {
              args[_key2] = arguments[_key2];
            }

            _this2.$emit.apply(_this2, [key].concat(args));
          };

          return pre;
        }, {});
      },
    }, {
      key: 'handleActiveChange',
      value: function handleActiveChange() {
        this.panelActive = this.active;
      },
    }, {
      key: 'created',
      value: function created() {
        this.panelActive = this.active;
        this.width = this.previewWidth;

        if (!this.panelActive) {
          const _this$panels = _slicedToArray__default.default(this.panels, 1);
          const firstPanel = _this$panels[0];

          this.panelActive = firstPanel !== null && firstPanel !== void 0 && firstPanel.name ? firstPanel.name : '';
          this.$emit('update:active', this.panelActive);
        }
      }, // 展开预览面板

    }, {
      key: 'handleResetWidth',
      value: function handleResetWidth() {
        this.width = this.previewWidth;
      }, // tab切换

    }, {
      key: 'handleTabChange',
      value: function handleTabChange(active) {
        this.panelActive = active;
        return active;
      }, // 预览面板操作(移除IP、复制IP等操作)

    }, {
      key: 'handlePreviewMenuClick',
      value: function handlePreviewMenuClick(_ref) {
        const { menu } = _ref;
        const { item } = _ref;
        return {
          menu,
          item,
        };
      }, // 移除预览面板节点

    }, {
      key: 'handleRemoveNode',
      value: function handleRemoveNode(_ref2) {
        const { child } = _ref2;
        const { item } = _ref2;
        return {
          child,
          item,
        };
      }, // eslint-disable-next-line @typescript-eslint/member-ordering

    }, {
      key: 'handleGetDefaultSelections',
      value: function handleGetDefaultSelections() {
        try {
          this.$refs.content.handleGetDefaultSelections();
        } catch (err) {
          console.log(err);
        }
      },
    }]);

    return IpSelector;
  }(Vue__default.default)), (_descriptor$1 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'active', [_dec2$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2$1 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'panels', [_dec3$1], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor3 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'tabVisible', [_dec4], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor4 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'previewWidth', [_dec5], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor5 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'previewRange', [_dec6], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor6 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'previewData', [_dec7], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor7 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'previewOperateList', [_dec8], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor8 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'height', [_dec9], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor9 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'defaultActiveName', [_dec10], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor10 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'previewTitle', [_dec11], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor11 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'tabRef', [_dec12], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor12 = _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'previewRef', [_dec13], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'handleActiveChange', [_dec14], Object.getOwnPropertyDescriptor(_class2$1.prototype, 'handleActiveChange'), _class2$1.prototype), _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'handleTabChange', [_dec15, _dec16], Object.getOwnPropertyDescriptor(_class2$1.prototype, 'handleTabChange'), _class2$1.prototype), _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'handlePreviewMenuClick', [_dec17], Object.getOwnPropertyDescriptor(_class2$1.prototype, 'handlePreviewMenuClick'), _class2$1.prototype), _applyDecoratedDescriptor__default.default(_class2$1.prototype, 'handleRemoveNode', [_dec18], Object.getOwnPropertyDescriptor(_class2$1.prototype, 'handleRemoveNode'), _class2$1.prototype)), _class2$1)) || _class$1);

  /* script */
  const __vue_script__$1 = IpSelector;
  /* template */

  const __vue_render__$1 = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('section', {
      staticClass: 'ip-selector',
      style: {
        height: typeof _vm.height === 'number' ? `${_vm.height}px` : _vm.height,
      },
    }, [_c('SelectorTab', {
      ref: 'tab',
      staticClass: 'ip-selector-left',
      attrs: {
        panels: _vm.panels,
        active: _vm.panelActive,
        'tab-visible': _vm.tabVisible,
      },
      on: {
        'tab-change': _vm.handleTabChange,
      },
    }, [_c('SelectorContent', _vm._g(_vm._b({
      ref: 'content',
      attrs: {
        active: _vm.panelActive,
        panels: _vm.panels,
      },
    }, 'SelectorContent', _vm.$attrs, false), _vm.contentEvents))], 1), _vm.width === 0 ? _c('div', {
      staticClass: 'preview-toggle',
    }, [_c('div', {
      directives: [{
        name: 'bk-tooltips',
        rawName: 'v-bk-tooltips',
        value: {
          content: _vm.$t('generic.ipSelector.action.expandSelectedPanel'),
          showOnInit: true,
          placements: ['left'],
          delay: 300,
          boundary: 'window',
        },
        expression: '{\n           content: $t(\'点击展开\'),\n           showOnInit: true,\n           placements: [\'left\'],\n           delay: 300,\n           boundary: \'window\'\n         }',
      }],
      staticClass: 'open-preview',
      on: {
        click: function click($event) {
          $event.stopPropagation();
          return _vm.handleResetWidth($event);
        },
      },
    }, [_c('i', {
      staticClass: 'bk-icon icon-angle-left',
    })])]) : _c('SelectorPreview', {
      ref: 'preview',
      staticClass: 'ip-selector-right',
      attrs: {
        width: _vm.width,
        range: _vm.previewRange,
        data: _vm.previewData,
        'operate-list': _vm.previewOperateList,
        'default-active-name': _vm.defaultActiveName,
        previewTitle: _vm.previewTitle,
      },
      on: {
        'update:width': function updateWidth($event) {
          _vm.width = $event;
        },
        'menu-click': _vm.handlePreviewMenuClick,
        'remove-node': _vm.handleRemoveNode,
      },
      scopedSlots: _vm._u([{
        key: 'collapse-title',
        fn: function fn(ref) {
          const { item } = ref;
          return [_vm._t('collapse-title', [_vm._v(_vm._s(item.name))], null, {
            item,
          })];
        },
      }], null, true),
    })], 1);
  };

  const __vue_staticRenderFns__$1 = [];
  /* style */

  const __vue_inject_styles__$1 = undefined;
  /* scoped */

  const __vue_scope_id__$1 = 'data-v-d7dd15a8';
  /* module identifier */

  const __vue_module_identifier__$1 = undefined;
  /* functional template */

  const __vue_is_functional_template__$1 = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__$1 = /* #__PURE__*/normalizeComponent({
    render: __vue_render__$1,
    staticRenderFns: __vue_staticRenderFns__$1,
  }, __vue_inject_styles__$1, __vue_script__$1, __vue_scope_id__$1, __vue_is_functional_template__$1, __vue_module_identifier__$1, false, undefined, undefined, undefined);

  let _dec; let _dec2; let _dec3; let _class; let _class2; let _descriptor; let _descriptor2;

  function _createSuper(Derived) {
    const hasNativeReflectConstruct = _isNativeReflectConstruct(); return function _createSuperInternal() {
      const Super = _getPrototypeOf__default.default(Derived); let result; if (hasNativeReflectConstruct) {
        const NewTarget = _getPrototypeOf__default.default(this).constructor; result = Reflect.construct(Super, arguments, NewTarget);
      } else {
        result = Super.apply(this, arguments);
      } return _possibleConstructorReturn__default.default(this, result);
    };
  }

  function _isNativeReflectConstruct() {
    if (typeof Reflect === 'undefined' || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === 'function') return true; try {
      Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], () => {})); return true;
    } catch (e) {
      return false;
    }
  }
  const AgentStatus = (_dec = Component({
    name: 'agent-status',
  }), _dec2 = Prop({
    default: 0,
    type: Number,
  }), _dec3 = Prop({
    default: function _default() {
      return [];
    },
    type: Array,
  }), _dec(_class = (_class2 = /* #__PURE__*/(function (_Vue) {
    _inherits__default.default(AgentStatus, _Vue);

    const _super = _createSuper(AgentStatus);

    function AgentStatus() {
      let _this;

      _classCallCheck__default.default(this, AgentStatus);

      for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
        args[_key] = arguments[_key];
      }

      _this = _super.call.apply(_super, [this].concat(args));

      _initializerDefineProperty__default.default(_this, 'type', _descriptor, _assertThisInitialized__default.default(_this));

      _initializerDefineProperty__default.default(_this, 'data', _descriptor2, _assertThisInitialized__default.default(_this));

      return _this;
    }

    return AgentStatus;
  }(Vue__default.default)), (_descriptor = _applyDecoratedDescriptor__default.default(_class2.prototype, 'type', [_dec2], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  }), _descriptor2 = _applyDecoratedDescriptor__default.default(_class2.prototype, 'data', [_dec3], {
    configurable: true,
    enumerable: true,
    writable: true,
    initializer: null,
  })), _class2)) || _class);

  /* script */
  const __vue_script__ = AgentStatus;
  /* template */

  const __vue_render__ = function __vue_render__() {
    const _vm = this;

    const _h = _vm.$createElement;

    const _c = _vm._self._c || _h;

    return _c('ul', {
      staticClass: 'agent-status',
    }, _vm._l(_vm.data, (item, index) => _c('li', {
      key: index,
    }, [_vm.type === 0 ? _c('div', [_c('span', {
      directives: [{
        name: 'show',
        rawName: 'v-show',
        value: item.display,
        expression: 'item.display',
      }],
      class: ['status-font', `status-${String(item.status).toLocaleLowerCase()}`],
    }, [_vm._v(`\n        ${_vm._s(item.count)}\n      `)]), _c('span', [_vm._v(_vm._s(_vm._f('filterEmpty')(item.display)))]), index !== _vm.data.length - 1 ? _c('span', {
      staticClass: 'separator',
    }, [_vm._v(', ')]) : _vm._e()]) : _vm.type === 1 ? _c('div', [_c('span', {
      class: ['status-mark', `status-${String(item.status).toLocaleLowerCase()}`],
    }), _c('span', [_vm._v(_vm._s(_vm._f('filterEmpty')(item.display)))])]) : _vm.type === 2 ? _c('div', {
      staticClass: 'agent-status-2',
    }, [_c('span', {
      class: ['status-halo', `status-${String(item.status).toLocaleLowerCase()}-halo`],
    }), _c('span', [_vm._v(_vm._s(_vm._f('filterEmpty')(item.display)))])]) : _c('div', [_c('span', {
      class: ['status-count', !!item.errorCount ? 'status-terminated' : 'status-2'],
    }, [_vm._v(`\n        ${_vm._s(item.errorCount || 0)}\n      `)]), _c('span', [_vm._v(_vm._s(item.count || 0))])])])), 0);
  };

  const __vue_staticRenderFns__ = [];
  /* style */

  const __vue_inject_styles__ = undefined;
  /* scoped */

  const __vue_scope_id__ = 'data-v-a38a7d02';
  /* module identifier */

  const __vue_module_identifier__ = undefined;
  /* functional template */

  const __vue_is_functional_template__ = false;
  /* style inject */

  /* style inject SSR */

  /* style inject shadow dom */

  const __vue_component__ = /* #__PURE__*/normalizeComponent({
    render: __vue_render__,
    staticRenderFns: __vue_staticRenderFns__,
  }, __vue_inject_styles__, __vue_script__, __vue_scope_id__, __vue_is_functional_template__, __vue_module_identifier__, false, undefined, undefined, undefined);

  Vue__default.default.prototype.$t = Vue__default.default.prototype.$t || function (str) {
    return str;
  };

  const installable = __vue_component__$1;

  installable.install = function (Vue) {
    return Vue.component(__vue_component__$1.name, __vue_component__$1);
  };

  exports.AgentStatus = __vue_component__;
  exports.default = installable;
  exports.ipSelector = __vue_component__$1;

  Object.defineProperty(exports, '__esModule', { value: true });
})));
