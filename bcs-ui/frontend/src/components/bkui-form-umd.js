(function (global, factory) {
  typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports, require('vue'), require('bk-magic-vue/dist/bk-magic-vue.min.css'), require('bk-magic-vue')) :
  typeof define === 'function' && define.amd ? define(['exports', 'vue', 'bk-magic-vue/dist/bk-magic-vue.min.css', 'bk-magic-vue'], factory) :
  (global = typeof globalThis !== 'undefined' ? globalThis : global || self, factory(global["bkui-form"] = {}, global.Vue, null, global.bkMagic));
})(this, (function (exports, Vue, bkMagicVue_min_css, bkMagic) { 'use strict';

  function _interopDefaultLegacy (e) { return e && typeof e === 'object' && 'default' in e ? e : { 'default': e }; }

  function _interopNamespace(e) {
    if (e && e.__esModule) return e;
    var n = Object.create(null);
    if (e) {
      Object.keys(e).forEach(function (k) {
        if (k !== 'default') {
          var d = Object.getOwnPropertyDescriptor(e, k);
          Object.defineProperty(n, k, d.get ? d : {
            enumerable: true,
            get: function () { return e[k]; }
          });
        }
      });
    }
    n["default"] = e;
    return Object.freeze(n);
  }

  var Vue__default = /*#__PURE__*/_interopDefaultLegacy(Vue);
  var bkMagic__namespace = /*#__PURE__*/_interopNamespace(bkMagic);

  function ownKeys(object, enumerableOnly) {
    var keys = Object.keys(object);

    if (Object.getOwnPropertySymbols) {
      var symbols = Object.getOwnPropertySymbols(object);
      enumerableOnly && (symbols = symbols.filter(function (sym) {
        return Object.getOwnPropertyDescriptor(object, sym).enumerable;
      })), keys.push.apply(keys, symbols);
    }

    return keys;
  }

  function _objectSpread2(target) {
    for (var i = 1; i < arguments.length; i++) {
      var source = null != arguments[i] ? arguments[i] : {};
      i % 2 ? ownKeys(Object(source), !0).forEach(function (key) {
        _defineProperty(target, key, source[key]);
      }) : Object.getOwnPropertyDescriptors ? Object.defineProperties(target, Object.getOwnPropertyDescriptors(source)) : ownKeys(Object(source)).forEach(function (key) {
        Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key));
      });
    }

    return target;
  }

  function _typeof(obj) {
    "@babel/helpers - typeof";

    return _typeof = "function" == typeof Symbol && "symbol" == typeof Symbol.iterator ? function (obj) {
      return typeof obj;
    } : function (obj) {
      return obj && "function" == typeof Symbol && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj;
    }, _typeof(obj);
  }

  function asyncGeneratorStep(gen, resolve, reject, _next, _throw, key, arg) {
    try {
      var info = gen[key](arg);
      var value = info.value;
    } catch (error) {
      reject(error);
      return;
    }

    if (info.done) {
      resolve(value);
    } else {
      Promise.resolve(value).then(_next, _throw);
    }
  }

  function _asyncToGenerator(fn) {
    return function () {
      var self = this,
          args = arguments;
      return new Promise(function (resolve, reject) {
        var gen = fn.apply(self, args);

        function _next(value) {
          asyncGeneratorStep(gen, resolve, reject, _next, _throw, "next", value);
        }

        function _throw(err) {
          asyncGeneratorStep(gen, resolve, reject, _next, _throw, "throw", err);
        }

        _next(undefined);
      });
    };
  }

  function _classCallCheck(instance, Constructor) {
    if (!(instance instanceof Constructor)) {
      throw new TypeError("Cannot call a class as a function");
    }
  }

  function _defineProperties(target, props) {
    for (var i = 0; i < props.length; i++) {
      var descriptor = props[i];
      descriptor.enumerable = descriptor.enumerable || false;
      descriptor.configurable = true;
      if ("value" in descriptor) descriptor.writable = true;
      Object.defineProperty(target, descriptor.key, descriptor);
    }
  }

  function _createClass(Constructor, protoProps, staticProps) {
    if (protoProps) _defineProperties(Constructor.prototype, protoProps);
    if (staticProps) _defineProperties(Constructor, staticProps);
    Object.defineProperty(Constructor, "prototype", {
      writable: false
    });
    return Constructor;
  }

  function _defineProperty(obj, key, value) {
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
  }

  function _objectWithoutPropertiesLoose(source, excluded) {
    if (source == null) return {};
    var target = {};
    var sourceKeys = Object.keys(source);
    var key, i;

    for (i = 0; i < sourceKeys.length; i++) {
      key = sourceKeys[i];
      if (excluded.indexOf(key) >= 0) continue;
      target[key] = source[key];
    }

    return target;
  }

  function _objectWithoutProperties(source, excluded) {
    if (source == null) return {};

    var target = _objectWithoutPropertiesLoose(source, excluded);

    var key, i;

    if (Object.getOwnPropertySymbols) {
      var sourceSymbolKeys = Object.getOwnPropertySymbols(source);

      for (i = 0; i < sourceSymbolKeys.length; i++) {
        key = sourceSymbolKeys[i];
        if (excluded.indexOf(key) >= 0) continue;
        if (!Object.prototype.propertyIsEnumerable.call(source, key)) continue;
        target[key] = source[key];
      }
    }

    return target;
  }

  function _slicedToArray(arr, i) {
    return _arrayWithHoles(arr) || _iterableToArrayLimit(arr, i) || _unsupportedIterableToArray(arr, i) || _nonIterableRest();
  }

  function _toConsumableArray(arr) {
    return _arrayWithoutHoles(arr) || _iterableToArray(arr) || _unsupportedIterableToArray(arr) || _nonIterableSpread();
  }

  function _arrayWithoutHoles(arr) {
    if (Array.isArray(arr)) return _arrayLikeToArray(arr);
  }

  function _arrayWithHoles(arr) {
    if (Array.isArray(arr)) return arr;
  }

  function _iterableToArray(iter) {
    if (typeof Symbol !== "undefined" && iter[Symbol.iterator] != null || iter["@@iterator"] != null) return Array.from(iter);
  }

  function _iterableToArrayLimit(arr, i) {
    var _i = arr == null ? null : typeof Symbol !== "undefined" && arr[Symbol.iterator] || arr["@@iterator"];

    if (_i == null) return;
    var _arr = [];
    var _n = true;
    var _d = false;

    var _s, _e;

    try {
      for (_i = _i.call(arr); !(_n = (_s = _i.next()).done); _n = true) {
        _arr.push(_s.value);

        if (i && _arr.length === i) break;
      }
    } catch (err) {
      _d = true;
      _e = err;
    } finally {
      try {
        if (!_n && _i["return"] != null) _i["return"]();
      } finally {
        if (_d) throw _e;
      }
    }

    return _arr;
  }

  function _unsupportedIterableToArray(o, minLen) {
    if (!o) return;
    if (typeof o === "string") return _arrayLikeToArray(o, minLen);
    var n = Object.prototype.toString.call(o).slice(8, -1);
    if (n === "Object" && o.constructor) n = o.constructor.name;
    if (n === "Map" || n === "Set") return Array.from(o);
    if (n === "Arguments" || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n)) return _arrayLikeToArray(o, minLen);
  }

  function _arrayLikeToArray(arr, len) {
    if (len == null || len > arr.length) len = arr.length;

    for (var i = 0, arr2 = new Array(len); i < len; i++) arr2[i] = arr[i];

    return arr2;
  }

  function _nonIterableSpread() {
    throw new TypeError("Invalid attempt to spread non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
  }

  function _nonIterableRest() {
    throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
  }

  function _createForOfIteratorHelper(o, allowArrayLike) {
    var it = typeof Symbol !== "undefined" && o[Symbol.iterator] || o["@@iterator"];

    if (!it) {
      if (Array.isArray(o) || (it = _unsupportedIterableToArray(o)) || allowArrayLike && o && typeof o.length === "number") {
        if (it) o = it;
        var i = 0;

        var F = function () {};

        return {
          s: F,
          n: function () {
            if (i >= o.length) return {
              done: true
            };
            return {
              done: false,
              value: o[i++]
            };
          },
          e: function (e) {
            throw e;
          },
          f: F
        };
      }

      throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
    }

    var normalCompletion = true,
        didErr = false,
        err;
    return {
      s: function () {
        it = it.call(o);
      },
      n: function () {
        var step = it.next();
        normalCompletion = step.done;
        return step;
      },
      e: function (e) {
        didErr = true;
        err = e;
      },
      f: function () {
        try {
          if (!normalCompletion && it.return != null) it.return();
        } finally {
          if (didErr) throw err;
        }
      }
    };
  }

  // eslint-disable-next-line no-new-func
  var isRegExp = function isRegExp(regExpStr) {
    var _Function;

    return ((_Function = new Function("return ".concat(regExpStr, ";"))()) === null || _Function === void 0 ? void 0 : _Function.constructor) === RegExp;
  };
  var isObj = function isObj(val) {
    return Object.prototype.toString.call(val) === '[object Object]';
  };
  (function globalSelf() {
    try {
      if (typeof self !== 'undefined') {
        return self;
      }
    } catch (e) {}

    try {
      if (typeof window !== 'undefined') {
        return window;
      }
    } catch (e) {}

    try {
      if (typeof global !== 'undefined') {
        return global;
      }
    } catch (e) {} // eslint-disable-next-line no-new-func


    return Function('return this')();
  })();
  var hasOwnProperty = function hasOwnProperty(obj, key) {
    return Object.prototype.hasOwnProperty.call(obj, key);
  };
  var valueType = function valueType(value) {
    if (Array.isArray(value)) {
      return 'array';
    }

    if (typeof value === 'string') {
      return 'string';
    }

    if (typeof value === 'boolean') {
      return 'boolean';
    }

    if (!isNaN(value)) {
      return 'number';
    }

    if (value === null) {
      return 'null';
    }

    if (_typeof(value) === 'object') {
      return 'object';
    }

    return _typeof(value);
  };
  var merge = function merge(target, source) {
    if (isObj(source)) {
      return Object.keys(source).reduce(function (pre, key) {
        var _target;

        pre[key] = merge(((_target = target) === null || _target === void 0 ? void 0 : _target[key]) || {}, source[key]);
        return pre;
      }, JSON.parse(JSON.stringify(target)));
    }

    if (Array.isArray(source)) {
      target = Array.isArray(target) ? target : [];
      return source.map(function (item, index) {
        if (target[index]) {
          return merge(target[index], item);
        }

        return item;
      });
    }

    return source;
  };
  function intersection(arr1, arr2) {
    return arr1.filter(function (item) {
      return arr2.includes(item);
    });
  } // 最大公约数

  function gcd(a, b) {
    if (b === 0) return a;
    return gcd(b, a % b);
  } // 最小公倍数

  function scm(a, b) {
    return a * b / gcd(a, b);
  } // 获取type对应的初始化值

  function initializationValue(type) {
    switch (type) {
      case 'any':
        return undefined;

      case 'array':
        return [];

      case 'boolean':
        return false;

      case 'integer':
        return 0;

      case 'null':
        return null;

      case 'number':
        return 0;

      case 'object':
        return {};

      case 'string':
        return '';
    }
  }
  function mergeDeep(target) {
    for (var _len = arguments.length, sources = new Array(_len > 1 ? _len - 1 : 0), _key = 1; _key < _len; _key++) {
      sources[_key - 1] = arguments[_key];
    }

    if (!sources.length) return target;
    var source = sources.shift();

    if (isObj(target) && isObj(source)) {
      for (var key in source) {
        if (isObj(source[key])) {
          if (!target[key]) Object.assign(target, _defineProperty({}, key, {}));
          mergeDeep(target[key], source[key]);
        } else {
          Object.assign(target, _defineProperty({}, key, source[key]));
        }
      }
    }

    return mergeDeep.apply(void 0, [target].concat(sources));
  }

  function isArguments(object) {
    return Object.prototype.toString.call(object) === '[object Arguments]';
  }

  function deepEquals(a, b) {
    var ca = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : [];
    var cb = arguments.length > 3 && arguments[3] !== undefined ? arguments[3] : [];

    // Partially extracted from node-deeper and adapted to exclude comparison
    // checks for functions.
    // https://github.com/othiym23/node-deeper
    if (a === b) {
      return true;
    }

    if (typeof a === 'function' || typeof b === 'function') {
      // Assume all functions are equivalent
      // see https://github.com/mozilla-services/react-jsonschema-form/issues/255
      return true;
    }

    if (_typeof(a) !== 'object' || _typeof(b) !== 'object') {
      return false;
    }

    if (a === null || b === null) {
      return false;
    }

    if (a instanceof Date && b instanceof Date) {
      return a.getTime() === b.getTime();
    }

    if (a instanceof RegExp && b instanceof RegExp) {
      return a.source === b.source && a.global === b.global && a.multiline === b.multiline && a.lastIndex === b.lastIndex && a.ignoreCase === b.ignoreCase;
    }

    if (isArguments(a) || isArguments(b)) {
      if (!(isArguments(a) && isArguments(b))) {
        return false;
      }

      var slice = Array.prototype.slice;
      return deepEquals(slice.call(a), slice.call(b), ca, cb);
    }

    if (a.constructor !== b.constructor) {
      return false;
    }

    var ka = Object.keys(a);
    var kb = Object.keys(b); // don't bother with stack acrobatics if there's nothing there

    if (ka.length === 0 && kb.length === 0) {
      return true;
    }

    if (ka.length !== kb.length) {
      return false;
    }

    var cal = ca.length; // eslint-disable-next-line no-plusplus

    while (cal--) {
      if (ca[cal] === a) {
        return cb[cal] === b;
      }
    }

    ca.push(a);
    cb.push(b);
    ka.sort();
    kb.sort(); // eslint-disable-next-line no-plusplus

    for (var j = ka.length - 1; j >= 0; j--) {
      if (ka[j] !== kb[j]) {
        return false;
      }
    }

    var key; // eslint-disable-next-line no-plusplus

    for (var k = ka.length - 1; k >= 0; k--) {
      key = ka[k];

      if (!deepEquals(a[key], b[key], ca, cb)) {
        return false;
      }
    }

    ca.pop();
    cb.pop();
    return true;
  }
  function orderProperties(properties, order) {
    if (!Array.isArray(order)) {
      return properties;
    }

    var arrayToHash = function arrayToHash(arr) {
      return arr.reduce(function (prev, curr) {
        prev[curr] = true;
        return prev;
      }, {});
    };

    var errorPropList = function errorPropList(arr) {
      return arr.length > 1 ? "properties '".concat(arr.join('\', \''), "'") : "property '".concat(arr[0], "'");
    };

    var propertyHash = arrayToHash(properties);
    var orderFiltered = order.filter(function (prop) {
      return prop === '*' || propertyHash[prop];
    });
    var orderHash = arrayToHash(orderFiltered);
    var rest = properties.filter(function (prop) {
      return !orderHash[prop];
    });
    var restIndex = orderFiltered.indexOf('*');

    if (restIndex === -1) {
      if (rest.length) {
        throw new Error("uiSchema order list does not contain ".concat(errorPropList(rest)));
      }

      return orderFiltered;
    }

    if (restIndex !== orderFiltered.lastIndexOf('*')) {
      throw new Error('uiSchema order list contains more than one wildcard item');
    }

    var complete = _toConsumableArray(orderFiltered);

    complete.splice.apply(complete, [restIndex, 1].concat(_toConsumableArray(rest)));
    return complete;
  }

  /**
   * Registry注册Form组件全局相关内容
   */
  var Registry = /*#__PURE__*/function () {
    function Registry() {
      _classCallCheck(this, Registry);

      this.widgets = new Map();
      this.components = new Map();
      this.fields = new Map();
    }

    _createClass(Registry, [{
      key: "addComponentsMap",
      value: function addComponentsMap() {
        var coms = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
        this.addMap('components', coms);
      }
    }, {
      key: "getComponent",
      value: function getComponent(key) {
        return this.components.get(key);
      }
    }, {
      key: "addFieldsMap",
      value: function addFieldsMap() {
        var map = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
        this.addMap('fields', map);
      }
    }, {
      key: "getField",
      value: function getField(key) {
        return this.fields.get(key);
      }
    }, {
      key: "addBaseWidgets",
      value: function addBaseWidgets() {
        var widgets = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
        this.addMap('widgets', widgets);
      } // 获取基础控件

    }, {
      key: "getBaseWidget",
      value: function getBaseWidget(name) {
        if (this.widgets.has(name)) {
          return this.widgets.get(name);
        }

        if (name && name.indexOf(Registry.namespace) !== 0) {
          return "".concat(Registry.namespace, "-").concat(name);
        }

        return name;
      }
    }, {
      key: "addMap",
      value: function addMap(props, map) {
        var _this = this;

        if (!props || !map) return;
        Object.entries(map).forEach(function (_ref) {
          var _ref2 = _slicedToArray(_ref, 2),
              key = _ref2[0],
              value = _ref2[1];

          if (_this[props].has(key)) {
            console.warn('repeat key', key);
            return;
          }

          _this[props].set(key, value);
        });
      }
    }]);

    return Registry;
  }();
  Registry.namespace = 'bk';
  Registry.context = {};
  var registry = new Registry();

  var _excluded$5 = ["allOf"],
      _excluded2$1 = ["name"],
      _excluded3 = ["name"];

  var Schema = /*#__PURE__*/function () {
    function Schema(rootSchema) {
      _classCallCheck(this, Schema);

      Schema.rootSchema = rootSchema;
    }

    _createClass(Schema, null, [{
      key: "resolveAllOf",
      value: function resolveAllOf(schema) {
        var _schema$allOf,
            _this = this;

        var allOfSchema = _objectSpread2(_objectSpread2({}, schema), {}, {
          allOf: (_schema$allOf = schema.allOf) === null || _schema$allOf === void 0 ? void 0 : _schema$allOf.map(function (item) {
            return _this.resolveSchema(item);
          })
        });

        try {
          // const { allOf, ...reset } = allOfSchema;
          return this.mergeSchemaAllOf();
        } catch (e) {
          console.warn(e);

          allOfSchema.allOf;
              var reset = _objectWithoutProperties(allOfSchema, _excluded$5);

          return reset;
        }
      } // todo

    }, {
      key: "mergeSchemaAllOf",
      value: function mergeSchemaAllOf() {
        var _this2 = this;

        for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
          args[_key] = arguments[_key];
        }

        if (args.length < 2) return args[0];
        var preVal = {};
        var copyArgs = [].concat(args);

        var _loop = function _loop() {
          var obj1 = isObj(copyArgs[0]) ? copyArgs[0] : {};
          var obj2 = isObj(copyArgs[1]) ? copyArgs[1] : {};
          preVal = Object.assign({}, obj1);
          Object.keys(obj2).reduce(function (acc, key) {
            var left = obj1[key];
            var right = obj2[key]; // 左右一边为object

            if (isObj(left) || isObj(right)) {
              // 两边同时为object
              if (isObj(left) && isObj(right)) {
                acc[key] = _this2.mergeSchemaAllOf(left, right);
              } else {
                // 其中一边为 object
                var _ref = isObj(left) ? [left, right] : [right, left],
                    _ref2 = _slicedToArray(_ref, 2),
                    objTypeData = _ref2[0],
                    baseTypeData = _ref2[1];

                if (key === 'additionalProperties') {
                  acc[key] = baseTypeData === true ? objTypeData : false; // default false
                } else {
                  acc[key] = objTypeData;
                }
              } // 一边为array

            } else if (Array.isArray(left) || Array.isArray(right)) {
              // 同为数组取交集
              if (Array.isArray(left) && Array.isArray(right)) {
                // 数组里面嵌套对象不支持 因为我不知道该怎么合并
                if (isObj(left[0]) || isObj(right[0])) {
                  throw new Error('暂不支持如上数组对象元素合并');
                } // 交集


                var intersectionArray = intersection(_toConsumableArray(left), _toConsumableArray(right)); // 没有交集

                if (intersectionArray.length <= 0) {
                  throw new Error('无法合并如上数据');
                }

                if (intersectionArray.length === 0 && key === 'type') {
                  // 自己取出值
                  acc[key] = intersectionArray[0];
                } else {
                  acc[key] = intersectionArray;
                }
              } else {
                // 其中一边为 Array
                // 查找包含关系
                var _ref3 = Array.isArray(left) ? [left, right] : [right, left],
                    _ref4 = _slicedToArray(_ref3, 2),
                    arrayTypeData = _ref4[0],
                    _baseTypeData = _ref4[1]; // 空值直接合并另一边


                if (_baseTypeData === undefined) {
                  acc[key] = arrayTypeData;
                } else {
                  if (!arrayTypeData.includes(_baseTypeData)) {
                    throw new Error('无法合并如下数据');
                  }

                  acc[key] = _baseTypeData;
                }
              }
            } else if (left !== undefined && right !== undefined) {
              // 两边都不是 undefined - 基础数据类型 string number boolean...
              if (key === 'maxLength' || key === 'maximum' || key === 'maxItems' || key === 'exclusiveMaximum' || key === 'maxProperties') {
                acc[key] = Math.min(left, right);
              } else if (key === 'minLength' || key === 'minimum' || key === 'minItems' || key === 'exclusiveMinimum' || key === 'minProperties') {
                acc[key] = Math.max(left, right);
              } else if (key === 'multipleOf') {
                // 获取最小公倍数
                acc[key] = scm(left, right);
              } else {
                // if (left !== right) {
                //     throw new Error('无法合并如下数据');
                // }
                acc[key] = left;
              }
            } else {
              // 一边为undefined
              acc[key] = left === undefined ? right : left;
            }

            return acc;
          }, preVal); // 先进先出

          copyArgs.splice(0, 2, preVal);
        };

        while (copyArgs.length >= 2) {
          _loop();
        }

        return preVal;
      }
    }, {
      key: "resolveRef",
      value: function resolveRef() {}
    }, {
      key: "resolveDependencies",
      value: function resolveDependencies() {}
    }, {
      key: "resolveAdditionalProperties",
      value: function resolveAdditionalProperties() {}
    }, {
      key: "resolveSchema",
      value: function resolveSchema(schema) {
        if (!isObj(schema)) return {};

        if (hasOwnProperty(schema, 'allOf')) {
          schema = this.resolveAllOf(schema);
        }

        if (hasOwnProperty(schema, '$ref')) ;

        return schema;
      } // 获取Schema字段默认值

    }, {
      key: "getSchemaDefaultValue",
      value: function getSchemaDefaultValue(_schema) {
        var _this3 = this;

        var schema = isObj(_schema) ? _schema : {};

        switch (this.getSchemaType(schema)) {
          case 'null':
            return null;

          case 'object':
            return Object.keys(schema.properties || {}).reduce(function (pre, key) {
              var _schema$properties;

              var defaultValue = _this3.getSchemaDefaultValue((_schema$properties = schema.properties) === null || _schema$properties === void 0 ? void 0 : _schema$properties[key]);

              pre[key] = defaultValue;
              return pre;
            }, {});

          case 'array':
            // todo
            return Array.isArray(schema.items) ? schema.items.map(function (item) {
              return _this3.getSchemaDefaultValue(item);
            }) : [];
        }

        return schema.default || initializationValue(schema.type || 'any');
      }
    }, {
      key: "getSchemaType",
      value: function getSchemaType(schema) {
        var type = schema.type;

        if (!type && schema.const) {
          return valueType(schema.const);
        }

        if (!type && schema.enum) {
          return 'string';
        }

        if (!type && schema.items) {
          return 'array';
        }

        return type;
      }
    }, {
      key: "getSchemaField",
      value: function getSchemaField(schema) {
        // 自定义Field组件
        var field = schema['ui:field'];

        if (field) {
          return field;
        } // default field


        var defaultField = registry.getField(this.getSchemaType(schema));

        if (defaultField) {
          return {
            name: defaultField
          };
        }

        return {
          name: null
        };
      }
    }, {
      key: "isMultiSelect",
      value: function isMultiSelect(schema) {
        if (!schema.uniqueItems || !schema.items) {
          return false;
        }

        return Array.isArray(schema.items.enum);
      }
    }, {
      key: "isTupleArray",
      value: function isTupleArray(schema) {
        var _schema$items;

        return Array.isArray(schema.items) && ((_schema$items = schema.items) === null || _schema$items === void 0 ? void 0 : _schema$items.length) > 0 && schema.items.every(function (item) {
          return isObj(item);
        });
      } // 是否是自定义数组类型控件（如：label）

    }, {
      key: "isCustomArrayWidget",
      value: function isCustomArrayWidget(schema) {
        var _schema$uiComponent;

        var com = (_schema$uiComponent = schema['ui:component']) === null || _schema$uiComponent === void 0 ? void 0 : _schema$uiComponent.name;
        return schema.type === 'array' && com;
      }
    }, {
      key: "getGroupWrap",
      value: function getGroupWrap(schema) {
        var _ref5 = schema['ui:group'] || {},
            name = _ref5.name,
            groupVnode = _objectWithoutProperties(_ref5, _excluded2$1);

        return _objectSpread2(_objectSpread2({}, groupVnode), {}, {
          name: registry.getComponent(name) || name || registry.getComponent('group')
        });
      }
    }, {
      key: "getUiComponent",
      value: function getUiComponent(schema) {
        var _ref6 = schema['ui:component'] || {},
            name = _ref6.name,
            vnodeData = _objectWithoutProperties(_ref6, _excluded3);

        return _objectSpread2({
          name: registry.getComponent(name) || name
        }, vnodeData);
      }
    }, {
      key: "getDefaultWidget",
      value: function getDefaultWidget(schema) {
        var type = this.getSchemaType(schema); // 默认转换策略

        var defaultComponent = null;

        if (type === 'string' && Array.isArray(schema.enum)) {
          // string类型的枚举数据默认用select组件
          defaultComponent = registry.getComponent('select');
        }

        if (type === 'array' && Schema.isMultiSelect(schema)) {
          // array类型多选默认用checkbox组件
          defaultComponent = registry.getComponent('checkbox');
        }

        if (defaultComponent) {
          return defaultComponent;
        } // 找不到对应组件就使用默认type对应的widget


        var typeComponentMap = {
          string: 'input',
          number: 'input',
          integer: 'input',
          boolean: 'switcher',
          null: ''
        };
        var name = typeComponentMap[type];
        var defaultWidget = registry.getComponent(name) || registry.getBaseWidget(name);

        if (defaultWidget) {
          return defaultWidget;
        }

        console.warn("\u672A\u6CE8\u518C\u7C7B\u578B".concat(type, "\u5BF9\u5E94\u7684\u9ED8\u8BA4\u8868\u5355\u9879"));
        return null;
      }
    }, {
      key: "isRequired",
      value: function isRequired(schema, name) {
        return Array.isArray(schema.required) && schema.required.includes(name);
      }
    }, {
      key: "getUiOptions",
      value: function getUiOptions(schema) {
        var options = _objectSpread2({
          showTitle: true,
          label: schema.title,
          desc: schema.description,
          minLength: schema.minLength,
          maxLength: schema.maxLength
        }, schema['ui:props'] || {});

        return _objectSpread2(_objectSpread2({}, options), {}, {
          // 0.1 兼容formItem设置 labelWidth 0 不生效问题
          labelWidth: options.showTitle ? options.labelWidth || 150 : 0.1
        });
      } // 当前属性是否被依赖

    }, {
      key: "getDependencies",
      value: function getDependencies(schema, name) {
        return Object.entries(schema.dependencies || {}).find(function (data) {
          return Array.isArray(data[1]) && data[1].includes(name);
        });
      }
    }, {
      key: "orderProps",
      value: function orderProps() {}
    }, {
      key: "resolveDefaultDatasource",
      value: function resolveDefaultDatasource(schema) {
        var _schema$uiComponent2, _schema$uiComponent2$, _schema$items2;

        if ((_schema$uiComponent2 = schema['ui:component']) !== null && _schema$uiComponent2 !== void 0 && (_schema$uiComponent2$ = _schema$uiComponent2.props) !== null && _schema$uiComponent2$ !== void 0 && _schema$uiComponent2$.datasource) {
          return schema['ui:component'].props.datasource;
        }

        var data = [];

        if (Array.isArray(schema.enum)) {
          data = schema.enum;
        } else if (Array.isArray((_schema$items2 = schema.items) === null || _schema$items2 === void 0 ? void 0 : _schema$items2.enum)) {
          data = schema.items.enum || [];
        }

        return data.map(function (value) {
          return {
            value: value,
            label: value
          };
        });
      }
    }]);

    return Schema;
  }();

  Schema.rootSchema = void 0;

  // 布局管理器
  var Layout = /*#__PURE__*/function () {
    function Layout() {
      var layout = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : [];

      _classCallCheck(this, Layout);

      this.layout = void 0;
      this.layout = this.parseLayout({
        prop: '',
        group: layout
      });
    }

    _createClass(Layout, [{
      key: "transformValueToPixel",
      value: function transformValueToPixel(value) {
        if (typeof value === 'number') {
          return "".concat(value, "px");
        }

        return value;
      }
    }, {
      key: "parseLayout",
      value: function parseLayout(config) {
        var _this = this;

        var dim = this.getLayoutDimension(config.group);
        var gridTemplate = this.parseGridTemplate(dim);
        var group = (config.group || []).map(function (current) {
          if (!Array.isArray(current)) {
            console.error("layout ".concat(JSON.stringify(current), " error, must be a array"));
            return [];
          }

          var parseCurrent = current.map(function (item) {
            if (typeof item === 'string' || typeof item === 'number') {
              return {
                prop: item,
                item: {
                  gridArea: item,
                  overflow: 'hidden'
                }
              };
            }

            if (_typeof(item) === 'object' && Reflect.has(item, 'group')) {
              return _this.parseLayout(item);
            }

            return item;
          });
          return parseCurrent;
        });
        var gridTemplateAreas = this.parseGridTemplateAreas(dim, group);
        var layoutConfig = {
          prop: config.prop || '',
          item: _objectSpread2({
            gridArea: config.prop,
            overflow: 'hidden'
          }, config.item),
          container: _objectSpread2(_objectSpread2({
            display: 'grid',
            gridTemplateAreas: gridTemplateAreas,
            gridGap: '24px'
          }, gridTemplate), config.container),
          group: group
        };
        return layoutConfig;
      }
    }, {
      key: "getLayoutDimension",
      value: function getLayoutDimension() {
        var layout = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : [];
        return layout.reduce(function (pre, current) {
          if (!Array.isArray(current)) {
            console.error("layout ".concat(JSON.stringify(current), " error, must be a array"));
          }

          if (current) {
            pre.columns = Math.max(pre.columns, current.length || 0);
          }

          pre.rows += 1;
          return pre;
        }, {
          columns: 0,
          rows: 0
        });
      }
    }, {
      key: "parseGridTemplate",
      value: function parseGridTemplate(dim) {
        return {
          gridTemplateColumns: this.repeatTemplate(dim.columns, false),
          gridTemplateRows: this.repeatTemplate(dim.rows)
        };
      }
    }, {
      key: "parseGridTemplateAreas",
      value: function parseGridTemplateAreas(gridTemplate, group) {
        var _this2 = this;

        return group.reduce(function (area, rows) {
          var newRows = _this2.autoFillColumns(rows, gridTemplate.columns);

          area += "\"".concat(newRows.join(' '), "\"\n");
          return area;
        }, '');
      } // 当前列数不够时自动填充

    }, {
      key: "autoFillColumns",
      value: function autoFillColumns(rows, maxColumns) {
        if (rows.length === maxColumns) {
          // 等于最大列时不扩展rows
          return rows.map(function (row) {
            return row.prop;
          });
        } // 均分补全的列


        var fillLen = Math.floor((maxColumns - rows.length) / rows.length);
        var newRows = rows.reduce(function (pre, row) {
          pre.push(row.prop);
          var fillData = new Array(fillLen).fill(row.prop);
          return pre.concat(fillData);
        }, []); // todo: 补全不够均分的列

        if (newRows.length < maxColumns) {
          var remainData = new Array(maxColumns - newRows.length).fill('.');
          return newRows.concat(remainData);
        }

        return newRows;
      }
    }, {
      key: "repeatTemplate",
      value: function repeatTemplate(len) {
        var auto = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : true;
        return new Array(len + 1).join(auto ? 'auto ' : '1fr ').trim();
      }
    }], [{
      key: "findLayoutByProp",
      value: function findLayoutByProp(prop, layout) {
        var layoutConfig;
        layout.find(function (row) {
          layoutConfig = row.find(function (col) {
            return col.prop === prop;
          });
          return layoutConfig;
        });
        return layoutConfig;
      }
    }]);

    return Layout;
  }();

  function createProxy(schema, context) {
    return new Proxy(schema, {
      get: function get(target, key, receiver) {
        if (typeof target[key] === 'function') {
          return target[key].apply(context);
        }

        return Reflect.get(target, key, receiver);
      }
    });
  }

  var Path = /*#__PURE__*/function () {
    function Path() {
      _classCallCheck(this, Path);
    }

    _createClass(Path, null, [{
      key: "getCurPath",
      value: function getCurPath(parent, current) {
        return parent === '' ? current : [parent, current].join(this.separator);
      }
    }, {
      key: "getPathVal",
      value: function getPathVal(obj, path) {
        var leftDeviation = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : 0;
        if (!path) return obj;
        var pathArr = path.split(this.separator);

        for (var i = 0; i < pathArr.length - leftDeviation; i += 1) {
          if (obj === undefined) return undefined;
          obj = pathArr[i] === '' ? obj : obj[pathArr[i]];
        }

        return obj;
      }
    }, {
      key: "setPathValue",
      value: function setPathValue() {
        var obj = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
        var path = arguments.length > 1 ? arguments[1] : undefined;
        var value = arguments.length > 2 ? arguments[2] : undefined;
        var newValue = JSON.parse(JSON.stringify(obj));
        var pathArr = path.split(this.separator);
        var lastProp = pathArr.pop() || '';
        var target = newValue;
        pathArr.forEach(function (prop) {
          if (!target[prop]) target[prop] = {};
          target = target[prop];
        });
        target[lastProp] = value;
        return newValue;
      } // 获取当前路径的字路径（相对于父路径）

    }, {
      key: "getSubPath",
      value: function getSubPath(parent, current) {
        return current.replace("".concat(parent, "."), '');
      } // 获取路径最后一个属性

    }, {
      key: "getPathLastProp",
      value: function getPathLastProp(path) {
        return path.split(this.separator).pop();
      } // 获取父路径

    }, {
      key: "getParentPath",
      value: function getParentPath(path) {
        var arrPath = path.split(this.separator);

        if (arrPath.length <= 1) {
          return ''; // root path
        }

        return arrPath.slice(0, arrPath.length - 1).join(this.separator);
      }
    }]);

    return Path;
  }();

  Path.separator = '.';

  var _excluded$4 = ["params", "responseType", "method", "headers", "responseParse"];
  var methodsWithoutData = ['GET', 'HEAD', 'OPTIONS', 'DELETE'];
  var defaultConfig = {
    responseType: 'json',
    method: 'GET',
    params: {},
    headers: {},
    cache: 'no-cache'
  };
  var request = /*#__PURE__*/(function () {
    var _ref = _asyncToGenerator( /*#__PURE__*/regeneratorRuntime.mark(function _callee(url) {
      var config,
          _mergeDeep,
          params,
          _mergeDeep$responseTy,
          responseType,
          _mergeDeep$method,
          method,
          _mergeDeep$headers,
          headers,
          responseParse,
          reset,
          body,
          requestURL,
          response,
          res,
          _responseParse$dataKe,
          dataKey,
          labelKey,
          valueKey,
          data,
          _args = arguments;

      return regeneratorRuntime.wrap(function _callee$(_context) {
        while (1) {
          switch (_context.prev = _context.next) {
            case 0:
              config = _args.length > 1 && _args[1] !== undefined ? _args[1] : {};
              _mergeDeep = mergeDeep(defaultConfig, config), params = _mergeDeep.params, _mergeDeep$responseTy = _mergeDeep.responseType, responseType = _mergeDeep$responseTy === void 0 ? 'json' : _mergeDeep$responseTy, _mergeDeep$method = _mergeDeep.method, method = _mergeDeep$method === void 0 ? 'GET' : _mergeDeep$method, _mergeDeep$headers = _mergeDeep.headers, headers = _mergeDeep$headers === void 0 ? {} : _mergeDeep$headers, responseParse = _mergeDeep.responseParse, reset = _objectWithoutProperties(_mergeDeep, _excluded$4);
              requestURL = url; // 处理参数

              if (methodsWithoutData.includes(method.toUpperCase())) {
                requestURL = "".concat(requestURL, "?").concat(isObj(params) ? new URLSearchParams(params) : params || '');
              } else {
                if (isObj(params)) {
                  // 处理JSON类型的请求
                  headers['Content-Type'] = 'application/json;charset=utf-8';
                  body = JSON.stringify(params);
                } else {
                  body = params;
                }
              }

              _context.prev = 4;
              _context.next = 7;
              return fetch(requestURL, _objectSpread2(_objectSpread2({}, reset), {}, {
                method: method.toLocaleUpperCase(),
                headers: headers,
                body: body
              }));

            case 7:
              response = _context.sent;

              if (!response.ok) {
                _context.next = 21;
                break;
              }

              _context.next = 11;
              return response[responseType]();

            case 11:
              res = _context.sent;

              if (!responseParse) {
                _context.next = 20;
                break;
              }

              if (!(typeof responseParse === 'function')) {
                _context.next = 15;
                break;
              }

              return _context.abrupt("return", Promise.resolve(responseParse(res)));

            case 15:
              if (!isObj(responseParse)) {
                _context.next = 19;
                break;
              }

              _responseParse$dataKe = responseParse.dataKey, dataKey = _responseParse$dataKe === void 0 ? 'data' : _responseParse$dataKe, labelKey = responseParse.labelKey, valueKey = responseParse.valueKey;
              data = ((res === null || res === void 0 ? void 0 : res[dataKey]) || []).map(function (item) {
                return _objectSpread2(_objectSpread2({}, item), {}, {
                  label: item === null || item === void 0 ? void 0 : item[labelKey],
                  value: item === null || item === void 0 ? void 0 : item[valueKey]
                });
              });
              return _context.abrupt("return", Promise.resolve(data));

            case 19:
              return _context.abrupt("return", Promise.resolve(res));

            case 20:
              return _context.abrupt("return", Promise.resolve(res));

            case 21:
              _context.t0 = response.status;
              _context.next = _context.t0 === 400 ? 24 : _context.t0 === 401 ? 25 : _context.t0 === 403 ? 26 : _context.t0 === 500 ? 26 : 27;
              break;

            case 24:
              return _context.abrupt("break", 27);

            case 25:
              return _context.abrupt("break", 27);

            case 26:
              return _context.abrupt("break", 27);

            case 27:
              _context.next = 33;
              break;

            case 29:
              _context.prev = 29;
              _context.t1 = _context["catch"](4);
              console.error('Request Failed', _context.t1);
              Promise.reject(_context.t1);

            case 33:
            case "end":
              return _context.stop();
          }
        }
      }, _callee, null, [[4, 29]]);
    }));

    return function (_x) {
      return _ref.apply(this, arguments);
    };
  })();

  var props$1 = {
    // 表单值
    value: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    // 全量Schema
    rules: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    schema: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    validator: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    width: {
      type: [String, Number],
      default: '100%'
    },
    // 表单布局
    layout: {
      type: Array,
      default: function _default() {
        return [];
      }
    },
    // 表单类型
    formType: {
      type: String,
      default: 'horizontal'
    },
    // 表单全局上下文
    context: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    // http请求适配器
    httpAdapter: {
      type: Object,
      default: function _default() {
        return {
          request: request,
          responseParse: function responseParse(res) {
            return res;
          } // fetch请求返回数据通用解析函数
          // responseParse: {
          //   dataKey: 'data',
          //   valueKey: 'value',
          //   labelKey: 'label',
          // }, // 也可以是对像形式

        };
      }
    }
  };

  var props = {
    // 当前项shema
    schema: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    // 当前路径（唯一标识）
    path: {
      type: String,
      default: ''
    },
    // 是否必须字段
    required: {
      type: Boolean,
      default: false
    },
    // 全量数据（只读）
    rootData: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    // 当前值
    value: {
      type: [String, Number, Array, Object, Boolean]
    },
    // 布局配置
    layout: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    // 当前全局变量上下文
    context: {
      type: Object,
      default: function _default() {
        return {};
      }
    },
    // 当前项是否可移除
    removeable: {
      type: Boolean,
      default: false
    },
    // http请求适配器
    httpAdapter: {
      type: Object,
      default: function _default() {
        return {
          request: request,
          responseParse: function responseParse(res) {
            return res;
          } // fetch请求返回数据通用解析函数
          // responseParse: {
          //   dataKey: 'data',
          //   valueKey: 'value',
          //   labelKey: 'label',
          // }, // 也可以是对像形式

        };
      }
    }
  };

  var SchemaField = Vue__default["default"].extend({
    name: 'SchemaField',
    functional: true,
    props: props,
    render: function render(h, ctx) {
      var _ctx$props = ctx.props,
          schema = _ctx$props.schema,
          rootData = _ctx$props.rootData,
          path = _ctx$props.path;
      var resolveSchema = Schema.resolveSchema(schema);
      if (!Object.keys(schema).length) return h();

      var _Schema$getSchemaFiel = Schema.getSchemaField(resolveSchema),
          name = _Schema$getSchemaFiel.name,
          fieldProps = _Schema$getSchemaFiel.props;

      return name ? h(name, _objectSpread2(_objectSpread2({}, ctx.data), {}, {
        props: _objectSpread2(_objectSpread2(_objectSpread2({}, fieldProps), ctx.props), {}, {
          value: Path.getPathVal(rootData, path),
          schema: createProxy(resolveSchema, ctx)
        })
      })) : h();
    }
  });

  var _excluded$3 = ["name"];
  var ObjectField = Vue__default["default"].extend({
    name: 'ObjectField',
    functional: true,
    props: props,
    render: function render(h, ctx) {
      var _ctx$props = ctx.props,
          schema = _ctx$props.schema,
          path = _ctx$props.path,
          layout = _ctx$props.layout,
          rootData = _ctx$props.rootData;
      var properties = orderProperties(Object.keys(schema.properties), schema['ui:order']);
      var vNodeList = properties.map(function (name) {
        var curPath = Path.getCurPath(path, name);
        var lastProp = curPath.split('.').pop();
        var layoutConfig = Layout.findLayoutByProp(lastProp, layout.group || []) || {};
        return h(SchemaField, _objectSpread2(_objectSpread2({}, ctx.data), {}, {
          key: curPath,
          props: _objectSpread2(_objectSpread2({}, ctx.props), {}, {
            schema: schema.properties[name],
            required: Schema.isRequired(schema, name),
            path: curPath,
            layout: layoutConfig,
            removeable: false // todo: 不往下传递可删除属性

          })
        }));
      });

      var _Schema$getGroupWrap = Schema.getGroupWrap(schema),
          name = _Schema$getGroupWrap.name,
          vnodeData = _objectWithoutProperties(_Schema$getGroupWrap, _excluded$3); // todo: wrap组件不要透传ctx.data，不然会引起组件无限递归问题


      return h(name, mergeDeep({
        props: _objectSpread2(_objectSpread2({}, ctx.props), {}, {
          value: Path.getPathVal(rootData, path),
          path: path
        }),
        style: _objectSpread2({}, ctx.data.style || {}),
        on: _objectSpread2({}, ctx.data.on || {})
      }, vnodeData), _toConsumableArray(vNodeList));
    }
  });

  /* eslint-disable @typescript-eslint/no-unused-vars */

  var getContext = function getContext(instance) {
    var context = instance.context,
        loadDataSource = instance.loadDataSource,
        validate = instance.validate,
        schema = instance.schema,
        rootData = instance.rootData,
        widgetNode = instance.widgetNode;
    return {
      $self: instance,
      $context: context,
      $schema: schema,
      $rules: schema.rules,
      $loadDataSource: loadDataSource,
      $validate: validate,
      $rootData: rootData,
      $widgetNode: widgetNode
    };
  };
  var executeExpression = function executeExpression(expression, instance) {
    var $dep = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : [];

    var _getContext = getContext(instance),
        $self = _getContext.$self,
        $context = _getContext.$context,
        $schema = _getContext.$schema,
        $rules = _getContext.$rules,
        $loadDataSource = _getContext.$loadDataSource,
        $validate = _getContext.$validate,
        $rootData = _getContext.$rootData,
        $widgetNode = _getContext.$widgetNode;

    var context = {
      $self: $self,
      $context: $context,
      $schema: $schema,
      $rules: $rules,
      $loadDataSource: $loadDataSource,
      $validate: $validate,
      $rootData: $rootData,
      $widgetNode: $widgetNode,
      $dep: $dep
    };

    if (typeof expression === 'string') {
      if (!/^{{.+}}$/.test(expression.trim())) return expression;
      var expStr = expression.trim().replace(/(^{{)|(}}$)/g, '').trim();
      var innerFuncs = ['$loadDataSource', '$validate'];

      var func = function func(_ref) {
        _ref.$self;
            _ref.$context;
            _ref.$schema;
            _ref.$rules;
            _ref.$loadDataSource;
            _ref.$validate;
            _ref.$rootData;
            _ref.$widgetNode;
            _ref.$dep;
        // eslint-disable-next-line no-eval
        return innerFuncs.includes(expStr) ? eval("".concat(expStr, "()")) : eval(expStr);
      };

      return func(context);
    }

    if (isObj(expression)) {
      Object.keys(expression).forEach(function (key) {
        expression[key] = executeExpression(expression[key], instance);
      });
      return expression;
    }
  }; // 解析字符串模板（无上下文时使用，只能解析context内容）
  var isExpression = function isExpression(expression) {
    return typeof expression === 'string' && /{{.*}}/.test(expression);
  };

  var WidgetNode = /*#__PURE__*/function () {
    // todo
    function WidgetNode(config) {
      _classCallCheck(this, WidgetNode);

      this.id = void 0;
      this.instance = void 0;
      this.parent = void 0;
      this.index = void 0;
      this.children = void 0;
      var id = config.id,
          instance = config.instance,
          parent = config.parent,
          index = config.index,
          _config$children = config.children,
          children = _config$children === void 0 ? [] : _config$children;
      this.id = id;
      this.index = index;
      this.instance = instance;
      this.parent = parent;
      this.children = children;
    }
    /**
     * 获取 parents
     */


    _createClass(WidgetNode, [{
      key: "parents",
      get: function get() {
        if (!this.parent) {
          return [];
        }

        return [].concat(_toConsumableArray(this.parent.parents), [this.parent]);
      } // 第一个子节点

    }, {
      key: "firstChild",
      get: function get() {
        return this.children[0] || null;
      } // 最后一个子节点

    }, {
      key: "lastChild",
      get: function get() {
        return this.children[this.children.length - 1] || null;
      } // 指定属性下的同胞节点

    }, {
      key: "getSibling",
      value: function getSibling(lastProp) {
        var _this$parent;

        var id = this.id.replace(Path.getPathLastProp(this.id) || '', lastProp);
        return (_this$parent = this.parent) === null || _this$parent === void 0 ? void 0 : _this$parent.children.find(function (node) {
          return node.id === id;
        });
      }
      /**
       * 是否是叶子节点
       */

    }, {
      key: "isLeaf",
      get: function get() {
        return !this.children.length;
      }
    }, {
      key: "appendChild",
      value: function appendChild(node) {
        var _this$children;

        var nodes = Array.isArray(node) ? node : [node];
        var offset = node.index !== undefined ? node.index : this.children.length;

        (_this$children = this.children).splice.apply(_this$children, [offset, 0].concat(_toConsumableArray(nodes)));

        this.children.slice(offset).forEach(function (node, index) {
          node.index = offset + index;
        });
        return nodes;
      }
    }, {
      key: "removeChild",
      value: function removeChild(node) {
        var _this = this;

        var nodes = Array.isArray(node) ? node : [node];
        var removedChildIndex = [];
        nodes.forEach(function (node) {
          var index = node.index;
          removedChildIndex.push(index);

          _this.children.splice(index, 1);
        });
        var minIndex = Math.min.apply(Math, removedChildIndex);
        this.children.slice(minIndex).forEach(function (node, index) {
          node.index = minIndex + index;
        });
        return nodes;
      }
    }]);

    return WidgetNode;
  }();
  var WidgetTree = /*#__PURE__*/function () {
    function WidgetTree() {
      _classCallCheck(this, WidgetTree);

      this.widgetMap = {};
    }

    _createClass(WidgetTree, [{
      key: "addWidgetNode",
      value: function addWidgetNode(path, instance, index) {
        if (path === '') {
          // 根节点
          var node = new WidgetNode({
            id: '',
            index: index,
            parent: null,
            instance: instance,
            children: []
          });
          this.widgetMap[path] = node;
        } else {
          // 普通节点
          var parentId = Path.getParentPath(path);
          var parentNode = this.widgetMap[parentId];

          if (!parentNode) {
            console.warn('Unexpected parent id, add widget failed');
            return;
          }

          var _node = new WidgetNode({
            id: path,
            index: index,
            parent: parentNode,
            instance: instance,
            children: []
          });

          parentNode.appendChild(_node);
          this.widgetMap[path] = _node;
        }
      }
    }, {
      key: "removeWidgetNode",
      value: function removeWidgetNode(path, instance) {
        var node = this.widgetMap[path];

        if (node !== null && node !== void 0 && node.parent) {
          var children = node.parent.children;
          var index = children.findIndex(function (item) {
            return item.instance === instance;
          });

          if (index > -1) {
            children.splice(index, 1);
            children.slice(index).forEach(function (node, i) {
              node.index = index + i;
            });
          }
        }

        delete this.widgetMap[path];
      }
    }]);

    return WidgetTree;
  }();
  var widgetTree = new WidgetTree();

  var reactionsMap = {};

  var subscribe = function subscribe(path, typeName, fn) {
    if (!reactionsMap[path]) {
      reactionsMap[path] = {
        lifetime: {},
        effect: {},
        fns: []
      };
    }

    if (typeName === 'valChange') {
      reactionsMap[path].fns.push(fn);
    } else {
      var _typeName$split = typeName.split('/'),
          _typeName$split2 = _slicedToArray(_typeName$split, 2),
          type = _typeName$split2[0],
          name = _typeName$split2[1];

      if (!reactionsMap[path][type][name]) {
        reactionsMap[path][type][name] = [fn];
      } else {
        reactionsMap[path][type][name].push(fn);
      }
    }
  };
  /**
   * 解析单个reaction
   *
   * @param crtInsPath 当前表单组件的path
   * @param targetPath 需要操作执行操作表单的path
   * @param reaction 传入的reacion配置
   * @returns viod
   */


  var resolveReaction = function resolveReaction(crtInsPath, targetPath, reaction) {
    return function () {
      var crtInstance = widgetTree.widgetMap[crtInsPath].instance; // 当前组件实例，用来条件表达式判断

      var operateInstance = widgetTree.widgetMap[targetPath].instance; // 需要执行操作的组件实例，可能为其他组件也可能为当前组件

      var fullfill = true;
      var deps = [];

      if (reaction.source) {
        var sources = Array.isArray(reaction.source) ? reaction.source : [reaction.source];
        sources.forEach(function (item) {
          var instance = widgetTree.widgetMap[parsePath(item, crtInsPath)].instance;
          deps.push(instance);
        });
      }

      if (typeof reaction.if === 'string') {
        fullfill = executeExpression(reaction.if, crtInstance, deps);
      }

      var operations = fullfill ? reaction.then : reaction.else;
      executeOperations(operations, operateInstance, deps);
    };
  };

  var executeOperations = function executeOperations(operations, instance, deps) {
    if (operations) {
      if (operations.state) {
        Object.keys(operations.state).forEach(function (key) {
          var val = operations.state[key];

          if (typeof val === 'string' && /^{{.+}}$/.test(val.trim())) {
            val = executeExpression(val, instance, deps);
          }

          instance.setState(key, val);
        });
      }

      if (Array.isArray(operations.actions)) {
        operations.actions.forEach(function (item) {
          executeExpression(item, instance, deps);
        });
      }
    }
  };

  var parsePath = function parsePath(path, instance) {
    return isExpression(path) ? executeExpression(path, instance) : path;
  };

  var reactionRegister = function reactionRegister(path) {
    var reactions = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : [];
    var instance = widgetTree.widgetMap[path].instance;

    if (reactions && Array.isArray(reactions)) {
      reactions.forEach(function (reaction) {
        // source 和 target 互斥，如果schema中同时定义，则取source，如果都没有定义则默认source为当前表单
        // source可能会用数组同时定义多个
        var subscribePaths = reaction.source ? Array.isArray(reaction.source) ? reaction.source : [reaction.source] : [path];
        var targePath = parsePath(typeof reaction.target === 'string' ? reaction.target : path, instance);
        subscribePaths.forEach(function (p) {
          var sourcePathItem = parsePath(p, instance);

          if (typeof reaction.lifetime === 'string') {
            subscribe(sourcePathItem, "lifetime/".concat(reaction.lifetime), resolveReaction(path, targePath, reaction));
          }

          if (typeof reaction.effect === 'string') {
            subscribe(sourcePathItem, "effect/".concat(reaction.effect), resolveReaction(path, targePath, reaction));
          }

          if (!reaction.lifetime && !reaction.effect) {
            subscribe(sourcePathItem, 'valChange', resolveReaction(path, targePath, reaction));
          }
        });
      });
    }
  };
  var reactionDispatch = function reactionDispatch(path, typeName) {
    var _typeName$split3 = typeName.split('/'),
        _typeName$split4 = _slicedToArray(_typeName$split3, 2),
        _typeName$split4$ = _typeName$split4[0],
        type = _typeName$split4$ === void 0 ? '' : _typeName$split4$,
        _typeName$split4$2 = _typeName$split4[1],
        name = _typeName$split4$2 === void 0 ? '' : _typeName$split4$2;

    var fns = [];

    if (reactionsMap[path]) {
      var _reactionsMap$path, _reactionsMap$path$ty;

      if (type === 'valChange') {
        fns = reactionsMap[path].fns;
      } else if ((_reactionsMap$path = reactionsMap[path]) !== null && _reactionsMap$path !== void 0 && (_reactionsMap$path$ty = _reactionsMap$path[type]) !== null && _reactionsMap$path$ty !== void 0 && _reactionsMap$path$ty[name]) {
        fns = reactionsMap[path][type][name];
      }

      fns.forEach(function (fn) {
        fn();
      });
    }
  };

  var commonjsGlobal = typeof globalThis !== 'undefined' ? globalThis : typeof window !== 'undefined' ? window : typeof global !== 'undefined' ? global : typeof self !== 'undefined' ? self : {};

  function getDefaultExportFromCjs (x) {
  	return x && x.__esModule && Object.prototype.hasOwnProperty.call(x, 'default') ? x['default'] : x;
  }

  var ajv$1 = {exports: {}};

  var core$2 = {};

  var validate$1 = {};

  var boolSchema = {};

  var errors = {};

  var codegen = {};

  var code$1 = {};

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.regexpCode = exports.getEsmExportName = exports.getProperty = exports.safeStringify = exports.stringify = exports.strConcat = exports.addCodeArg = exports.str = exports._ = exports.nil = exports._Code = exports.Name = exports.IDENTIFIER = exports._CodeOrName = void 0;
  class _CodeOrName {
  }
  exports._CodeOrName = _CodeOrName;
  exports.IDENTIFIER = /^[a-z$_][a-z$_0-9]*$/i;
  class Name extends _CodeOrName {
      constructor(s) {
          super();
          if (!exports.IDENTIFIER.test(s))
              throw new Error("CodeGen: name must be a valid identifier");
          this.str = s;
      }
      toString() {
          return this.str;
      }
      emptyStr() {
          return false;
      }
      get names() {
          return { [this.str]: 1 };
      }
  }
  exports.Name = Name;
  class _Code extends _CodeOrName {
      constructor(code) {
          super();
          this._items = typeof code === "string" ? [code] : code;
      }
      toString() {
          return this.str;
      }
      emptyStr() {
          if (this._items.length > 1)
              return false;
          const item = this._items[0];
          return item === "" || item === '""';
      }
      get str() {
          var _a;
          return ((_a = this._str) !== null && _a !== void 0 ? _a : (this._str = this._items.reduce((s, c) => `${s}${c}`, "")));
      }
      get names() {
          var _a;
          return ((_a = this._names) !== null && _a !== void 0 ? _a : (this._names = this._items.reduce((names, c) => {
              if (c instanceof Name)
                  names[c.str] = (names[c.str] || 0) + 1;
              return names;
          }, {})));
      }
  }
  exports._Code = _Code;
  exports.nil = new _Code("");
  function _(strs, ...args) {
      const code = [strs[0]];
      let i = 0;
      while (i < args.length) {
          addCodeArg(code, args[i]);
          code.push(strs[++i]);
      }
      return new _Code(code);
  }
  exports._ = _;
  const plus = new _Code("+");
  function str(strs, ...args) {
      const expr = [safeStringify(strs[0])];
      let i = 0;
      while (i < args.length) {
          expr.push(plus);
          addCodeArg(expr, args[i]);
          expr.push(plus, safeStringify(strs[++i]));
      }
      optimize(expr);
      return new _Code(expr);
  }
  exports.str = str;
  function addCodeArg(code, arg) {
      if (arg instanceof _Code)
          code.push(...arg._items);
      else if (arg instanceof Name)
          code.push(arg);
      else
          code.push(interpolate(arg));
  }
  exports.addCodeArg = addCodeArg;
  function optimize(expr) {
      let i = 1;
      while (i < expr.length - 1) {
          if (expr[i] === plus) {
              const res = mergeExprItems(expr[i - 1], expr[i + 1]);
              if (res !== undefined) {
                  expr.splice(i - 1, 3, res);
                  continue;
              }
              expr[i++] = "+";
          }
          i++;
      }
  }
  function mergeExprItems(a, b) {
      if (b === '""')
          return a;
      if (a === '""')
          return b;
      if (typeof a == "string") {
          if (b instanceof Name || a[a.length - 1] !== '"')
              return;
          if (typeof b != "string")
              return `${a.slice(0, -1)}${b}"`;
          if (b[0] === '"')
              return a.slice(0, -1) + b.slice(1);
          return;
      }
      if (typeof b == "string" && b[0] === '"' && !(a instanceof Name))
          return `"${a}${b.slice(1)}`;
      return;
  }
  function strConcat(c1, c2) {
      return c2.emptyStr() ? c1 : c1.emptyStr() ? c2 : str `${c1}${c2}`;
  }
  exports.strConcat = strConcat;
  // TODO do not allow arrays here
  function interpolate(x) {
      return typeof x == "number" || typeof x == "boolean" || x === null
          ? x
          : safeStringify(Array.isArray(x) ? x.join(",") : x);
  }
  function stringify(x) {
      return new _Code(safeStringify(x));
  }
  exports.stringify = stringify;
  function safeStringify(x) {
      return JSON.stringify(x)
          .replace(/\u2028/g, "\\u2028")
          .replace(/\u2029/g, "\\u2029");
  }
  exports.safeStringify = safeStringify;
  function getProperty(key) {
      return typeof key == "string" && exports.IDENTIFIER.test(key) ? new _Code(`.${key}`) : _ `[${key}]`;
  }
  exports.getProperty = getProperty;
  //Does best effort to format the name properly
  function getEsmExportName(key) {
      if (typeof key == "string" && exports.IDENTIFIER.test(key)) {
          return new _Code(`${key}`);
      }
      throw new Error(`CodeGen: invalid export name: ${key}, use explicit $id name mapping`);
  }
  exports.getEsmExportName = getEsmExportName;
  function regexpCode(rx) {
      return new _Code(rx.toString());
  }
  exports.regexpCode = regexpCode;

  }(code$1));

  var scope = {};

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.ValueScope = exports.ValueScopeName = exports.Scope = exports.varKinds = exports.UsedValueState = void 0;
  const code_1 = code$1;
  class ValueError extends Error {
      constructor(name) {
          super(`CodeGen: "code" for ${name} not defined`);
          this.value = name.value;
      }
  }
  var UsedValueState;
  (function (UsedValueState) {
      UsedValueState[UsedValueState["Started"] = 0] = "Started";
      UsedValueState[UsedValueState["Completed"] = 1] = "Completed";
  })(UsedValueState = exports.UsedValueState || (exports.UsedValueState = {}));
  exports.varKinds = {
      const: new code_1.Name("const"),
      let: new code_1.Name("let"),
      var: new code_1.Name("var"),
  };
  class Scope {
      constructor({ prefixes, parent } = {}) {
          this._names = {};
          this._prefixes = prefixes;
          this._parent = parent;
      }
      toName(nameOrPrefix) {
          return nameOrPrefix instanceof code_1.Name ? nameOrPrefix : this.name(nameOrPrefix);
      }
      name(prefix) {
          return new code_1.Name(this._newName(prefix));
      }
      _newName(prefix) {
          const ng = this._names[prefix] || this._nameGroup(prefix);
          return `${prefix}${ng.index++}`;
      }
      _nameGroup(prefix) {
          var _a, _b;
          if (((_b = (_a = this._parent) === null || _a === void 0 ? void 0 : _a._prefixes) === null || _b === void 0 ? void 0 : _b.has(prefix)) || (this._prefixes && !this._prefixes.has(prefix))) {
              throw new Error(`CodeGen: prefix "${prefix}" is not allowed in this scope`);
          }
          return (this._names[prefix] = { prefix, index: 0 });
      }
  }
  exports.Scope = Scope;
  class ValueScopeName extends code_1.Name {
      constructor(prefix, nameStr) {
          super(nameStr);
          this.prefix = prefix;
      }
      setValue(value, { property, itemIndex }) {
          this.value = value;
          this.scopePath = (0, code_1._) `.${new code_1.Name(property)}[${itemIndex}]`;
      }
  }
  exports.ValueScopeName = ValueScopeName;
  const line = (0, code_1._) `\n`;
  class ValueScope extends Scope {
      constructor(opts) {
          super(opts);
          this._values = {};
          this._scope = opts.scope;
          this.opts = { ...opts, _n: opts.lines ? line : code_1.nil };
      }
      get() {
          return this._scope;
      }
      name(prefix) {
          return new ValueScopeName(prefix, this._newName(prefix));
      }
      value(nameOrPrefix, value) {
          var _a;
          if (value.ref === undefined)
              throw new Error("CodeGen: ref must be passed in value");
          const name = this.toName(nameOrPrefix);
          const { prefix } = name;
          const valueKey = (_a = value.key) !== null && _a !== void 0 ? _a : value.ref;
          let vs = this._values[prefix];
          if (vs) {
              const _name = vs.get(valueKey);
              if (_name)
                  return _name;
          }
          else {
              vs = this._values[prefix] = new Map();
          }
          vs.set(valueKey, name);
          const s = this._scope[prefix] || (this._scope[prefix] = []);
          const itemIndex = s.length;
          s[itemIndex] = value.ref;
          name.setValue(value, { property: prefix, itemIndex });
          return name;
      }
      getValue(prefix, keyOrRef) {
          const vs = this._values[prefix];
          if (!vs)
              return;
          return vs.get(keyOrRef);
      }
      scopeRefs(scopeName, values = this._values) {
          return this._reduceValues(values, (name) => {
              if (name.scopePath === undefined)
                  throw new Error(`CodeGen: name "${name}" has no value`);
              return (0, code_1._) `${scopeName}${name.scopePath}`;
          });
      }
      scopeCode(values = this._values, usedValues, getCode) {
          return this._reduceValues(values, (name) => {
              if (name.value === undefined)
                  throw new Error(`CodeGen: name "${name}" has no value`);
              return name.value.code;
          }, usedValues, getCode);
      }
      _reduceValues(values, valueCode, usedValues = {}, getCode) {
          let code = code_1.nil;
          for (const prefix in values) {
              const vs = values[prefix];
              if (!vs)
                  continue;
              const nameSet = (usedValues[prefix] = usedValues[prefix] || new Map());
              vs.forEach((name) => {
                  if (nameSet.has(name))
                      return;
                  nameSet.set(name, UsedValueState.Started);
                  let c = valueCode(name);
                  if (c) {
                      const def = this.opts.es5 ? exports.varKinds.var : exports.varKinds.const;
                      code = (0, code_1._) `${code}${def} ${name} = ${c};${this.opts._n}`;
                  }
                  else if ((c = getCode === null || getCode === void 0 ? void 0 : getCode(name))) {
                      code = (0, code_1._) `${code}${c}${this.opts._n}`;
                  }
                  else {
                      throw new ValueError(name);
                  }
                  nameSet.set(name, UsedValueState.Completed);
              });
          }
          return code;
      }
  }
  exports.ValueScope = ValueScope;

  }(scope));

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.or = exports.and = exports.not = exports.CodeGen = exports.operators = exports.varKinds = exports.ValueScopeName = exports.ValueScope = exports.Scope = exports.Name = exports.regexpCode = exports.stringify = exports.getProperty = exports.nil = exports.strConcat = exports.str = exports._ = void 0;
  const code_1 = code$1;
  const scope_1 = scope;
  var code_2 = code$1;
  Object.defineProperty(exports, "_", { enumerable: true, get: function () { return code_2._; } });
  Object.defineProperty(exports, "str", { enumerable: true, get: function () { return code_2.str; } });
  Object.defineProperty(exports, "strConcat", { enumerable: true, get: function () { return code_2.strConcat; } });
  Object.defineProperty(exports, "nil", { enumerable: true, get: function () { return code_2.nil; } });
  Object.defineProperty(exports, "getProperty", { enumerable: true, get: function () { return code_2.getProperty; } });
  Object.defineProperty(exports, "stringify", { enumerable: true, get: function () { return code_2.stringify; } });
  Object.defineProperty(exports, "regexpCode", { enumerable: true, get: function () { return code_2.regexpCode; } });
  Object.defineProperty(exports, "Name", { enumerable: true, get: function () { return code_2.Name; } });
  var scope_2 = scope;
  Object.defineProperty(exports, "Scope", { enumerable: true, get: function () { return scope_2.Scope; } });
  Object.defineProperty(exports, "ValueScope", { enumerable: true, get: function () { return scope_2.ValueScope; } });
  Object.defineProperty(exports, "ValueScopeName", { enumerable: true, get: function () { return scope_2.ValueScopeName; } });
  Object.defineProperty(exports, "varKinds", { enumerable: true, get: function () { return scope_2.varKinds; } });
  exports.operators = {
      GT: new code_1._Code(">"),
      GTE: new code_1._Code(">="),
      LT: new code_1._Code("<"),
      LTE: new code_1._Code("<="),
      EQ: new code_1._Code("==="),
      NEQ: new code_1._Code("!=="),
      NOT: new code_1._Code("!"),
      OR: new code_1._Code("||"),
      AND: new code_1._Code("&&"),
      ADD: new code_1._Code("+"),
  };
  class Node {
      optimizeNodes() {
          return this;
      }
      optimizeNames(_names, _constants) {
          return this;
      }
  }
  class Def extends Node {
      constructor(varKind, name, rhs) {
          super();
          this.varKind = varKind;
          this.name = name;
          this.rhs = rhs;
      }
      render({ es5, _n }) {
          const varKind = es5 ? scope_1.varKinds.var : this.varKind;
          const rhs = this.rhs === undefined ? "" : ` = ${this.rhs}`;
          return `${varKind} ${this.name}${rhs};` + _n;
      }
      optimizeNames(names, constants) {
          if (!names[this.name.str])
              return;
          if (this.rhs)
              this.rhs = optimizeExpr(this.rhs, names, constants);
          return this;
      }
      get names() {
          return this.rhs instanceof code_1._CodeOrName ? this.rhs.names : {};
      }
  }
  class Assign extends Node {
      constructor(lhs, rhs, sideEffects) {
          super();
          this.lhs = lhs;
          this.rhs = rhs;
          this.sideEffects = sideEffects;
      }
      render({ _n }) {
          return `${this.lhs} = ${this.rhs};` + _n;
      }
      optimizeNames(names, constants) {
          if (this.lhs instanceof code_1.Name && !names[this.lhs.str] && !this.sideEffects)
              return;
          this.rhs = optimizeExpr(this.rhs, names, constants);
          return this;
      }
      get names() {
          const names = this.lhs instanceof code_1.Name ? {} : { ...this.lhs.names };
          return addExprNames(names, this.rhs);
      }
  }
  class AssignOp extends Assign {
      constructor(lhs, op, rhs, sideEffects) {
          super(lhs, rhs, sideEffects);
          this.op = op;
      }
      render({ _n }) {
          return `${this.lhs} ${this.op}= ${this.rhs};` + _n;
      }
  }
  class Label extends Node {
      constructor(label) {
          super();
          this.label = label;
          this.names = {};
      }
      render({ _n }) {
          return `${this.label}:` + _n;
      }
  }
  class Break extends Node {
      constructor(label) {
          super();
          this.label = label;
          this.names = {};
      }
      render({ _n }) {
          const label = this.label ? ` ${this.label}` : "";
          return `break${label};` + _n;
      }
  }
  class Throw extends Node {
      constructor(error) {
          super();
          this.error = error;
      }
      render({ _n }) {
          return `throw ${this.error};` + _n;
      }
      get names() {
          return this.error.names;
      }
  }
  class AnyCode extends Node {
      constructor(code) {
          super();
          this.code = code;
      }
      render({ _n }) {
          return `${this.code};` + _n;
      }
      optimizeNodes() {
          return `${this.code}` ? this : undefined;
      }
      optimizeNames(names, constants) {
          this.code = optimizeExpr(this.code, names, constants);
          return this;
      }
      get names() {
          return this.code instanceof code_1._CodeOrName ? this.code.names : {};
      }
  }
  class ParentNode extends Node {
      constructor(nodes = []) {
          super();
          this.nodes = nodes;
      }
      render(opts) {
          return this.nodes.reduce((code, n) => code + n.render(opts), "");
      }
      optimizeNodes() {
          const { nodes } = this;
          let i = nodes.length;
          while (i--) {
              const n = nodes[i].optimizeNodes();
              if (Array.isArray(n))
                  nodes.splice(i, 1, ...n);
              else if (n)
                  nodes[i] = n;
              else
                  nodes.splice(i, 1);
          }
          return nodes.length > 0 ? this : undefined;
      }
      optimizeNames(names, constants) {
          const { nodes } = this;
          let i = nodes.length;
          while (i--) {
              // iterating backwards improves 1-pass optimization
              const n = nodes[i];
              if (n.optimizeNames(names, constants))
                  continue;
              subtractNames(names, n.names);
              nodes.splice(i, 1);
          }
          return nodes.length > 0 ? this : undefined;
      }
      get names() {
          return this.nodes.reduce((names, n) => addNames(names, n.names), {});
      }
  }
  class BlockNode extends ParentNode {
      render(opts) {
          return "{" + opts._n + super.render(opts) + "}" + opts._n;
      }
  }
  class Root extends ParentNode {
  }
  class Else extends BlockNode {
  }
  Else.kind = "else";
  class If extends BlockNode {
      constructor(condition, nodes) {
          super(nodes);
          this.condition = condition;
      }
      render(opts) {
          let code = `if(${this.condition})` + super.render(opts);
          if (this.else)
              code += "else " + this.else.render(opts);
          return code;
      }
      optimizeNodes() {
          super.optimizeNodes();
          const cond = this.condition;
          if (cond === true)
              return this.nodes; // else is ignored here
          let e = this.else;
          if (e) {
              const ns = e.optimizeNodes();
              e = this.else = Array.isArray(ns) ? new Else(ns) : ns;
          }
          if (e) {
              if (cond === false)
                  return e instanceof If ? e : e.nodes;
              if (this.nodes.length)
                  return this;
              return new If(not(cond), e instanceof If ? [e] : e.nodes);
          }
          if (cond === false || !this.nodes.length)
              return undefined;
          return this;
      }
      optimizeNames(names, constants) {
          var _a;
          this.else = (_a = this.else) === null || _a === void 0 ? void 0 : _a.optimizeNames(names, constants);
          if (!(super.optimizeNames(names, constants) || this.else))
              return;
          this.condition = optimizeExpr(this.condition, names, constants);
          return this;
      }
      get names() {
          const names = super.names;
          addExprNames(names, this.condition);
          if (this.else)
              addNames(names, this.else.names);
          return names;
      }
  }
  If.kind = "if";
  class For extends BlockNode {
  }
  For.kind = "for";
  class ForLoop extends For {
      constructor(iteration) {
          super();
          this.iteration = iteration;
      }
      render(opts) {
          return `for(${this.iteration})` + super.render(opts);
      }
      optimizeNames(names, constants) {
          if (!super.optimizeNames(names, constants))
              return;
          this.iteration = optimizeExpr(this.iteration, names, constants);
          return this;
      }
      get names() {
          return addNames(super.names, this.iteration.names);
      }
  }
  class ForRange extends For {
      constructor(varKind, name, from, to) {
          super();
          this.varKind = varKind;
          this.name = name;
          this.from = from;
          this.to = to;
      }
      render(opts) {
          const varKind = opts.es5 ? scope_1.varKinds.var : this.varKind;
          const { name, from, to } = this;
          return `for(${varKind} ${name}=${from}; ${name}<${to}; ${name}++)` + super.render(opts);
      }
      get names() {
          const names = addExprNames(super.names, this.from);
          return addExprNames(names, this.to);
      }
  }
  class ForIter extends For {
      constructor(loop, varKind, name, iterable) {
          super();
          this.loop = loop;
          this.varKind = varKind;
          this.name = name;
          this.iterable = iterable;
      }
      render(opts) {
          return `for(${this.varKind} ${this.name} ${this.loop} ${this.iterable})` + super.render(opts);
      }
      optimizeNames(names, constants) {
          if (!super.optimizeNames(names, constants))
              return;
          this.iterable = optimizeExpr(this.iterable, names, constants);
          return this;
      }
      get names() {
          return addNames(super.names, this.iterable.names);
      }
  }
  class Func extends BlockNode {
      constructor(name, args, async) {
          super();
          this.name = name;
          this.args = args;
          this.async = async;
      }
      render(opts) {
          const _async = this.async ? "async " : "";
          return `${_async}function ${this.name}(${this.args})` + super.render(opts);
      }
  }
  Func.kind = "func";
  class Return extends ParentNode {
      render(opts) {
          return "return " + super.render(opts);
      }
  }
  Return.kind = "return";
  class Try extends BlockNode {
      render(opts) {
          let code = "try" + super.render(opts);
          if (this.catch)
              code += this.catch.render(opts);
          if (this.finally)
              code += this.finally.render(opts);
          return code;
      }
      optimizeNodes() {
          var _a, _b;
          super.optimizeNodes();
          (_a = this.catch) === null || _a === void 0 ? void 0 : _a.optimizeNodes();
          (_b = this.finally) === null || _b === void 0 ? void 0 : _b.optimizeNodes();
          return this;
      }
      optimizeNames(names, constants) {
          var _a, _b;
          super.optimizeNames(names, constants);
          (_a = this.catch) === null || _a === void 0 ? void 0 : _a.optimizeNames(names, constants);
          (_b = this.finally) === null || _b === void 0 ? void 0 : _b.optimizeNames(names, constants);
          return this;
      }
      get names() {
          const names = super.names;
          if (this.catch)
              addNames(names, this.catch.names);
          if (this.finally)
              addNames(names, this.finally.names);
          return names;
      }
  }
  class Catch extends BlockNode {
      constructor(error) {
          super();
          this.error = error;
      }
      render(opts) {
          return `catch(${this.error})` + super.render(opts);
      }
  }
  Catch.kind = "catch";
  class Finally extends BlockNode {
      render(opts) {
          return "finally" + super.render(opts);
      }
  }
  Finally.kind = "finally";
  class CodeGen {
      constructor(extScope, opts = {}) {
          this._values = {};
          this._blockStarts = [];
          this._constants = {};
          this.opts = { ...opts, _n: opts.lines ? "\n" : "" };
          this._extScope = extScope;
          this._scope = new scope_1.Scope({ parent: extScope });
          this._nodes = [new Root()];
      }
      toString() {
          return this._root.render(this.opts);
      }
      // returns unique name in the internal scope
      name(prefix) {
          return this._scope.name(prefix);
      }
      // reserves unique name in the external scope
      scopeName(prefix) {
          return this._extScope.name(prefix);
      }
      // reserves unique name in the external scope and assigns value to it
      scopeValue(prefixOrName, value) {
          const name = this._extScope.value(prefixOrName, value);
          const vs = this._values[name.prefix] || (this._values[name.prefix] = new Set());
          vs.add(name);
          return name;
      }
      getScopeValue(prefix, keyOrRef) {
          return this._extScope.getValue(prefix, keyOrRef);
      }
      // return code that assigns values in the external scope to the names that are used internally
      // (same names that were returned by gen.scopeName or gen.scopeValue)
      scopeRefs(scopeName) {
          return this._extScope.scopeRefs(scopeName, this._values);
      }
      scopeCode() {
          return this._extScope.scopeCode(this._values);
      }
      _def(varKind, nameOrPrefix, rhs, constant) {
          const name = this._scope.toName(nameOrPrefix);
          if (rhs !== undefined && constant)
              this._constants[name.str] = rhs;
          this._leafNode(new Def(varKind, name, rhs));
          return name;
      }
      // `const` declaration (`var` in es5 mode)
      const(nameOrPrefix, rhs, _constant) {
          return this._def(scope_1.varKinds.const, nameOrPrefix, rhs, _constant);
      }
      // `let` declaration with optional assignment (`var` in es5 mode)
      let(nameOrPrefix, rhs, _constant) {
          return this._def(scope_1.varKinds.let, nameOrPrefix, rhs, _constant);
      }
      // `var` declaration with optional assignment
      var(nameOrPrefix, rhs, _constant) {
          return this._def(scope_1.varKinds.var, nameOrPrefix, rhs, _constant);
      }
      // assignment code
      assign(lhs, rhs, sideEffects) {
          return this._leafNode(new Assign(lhs, rhs, sideEffects));
      }
      // `+=` code
      add(lhs, rhs) {
          return this._leafNode(new AssignOp(lhs, exports.operators.ADD, rhs));
      }
      // appends passed SafeExpr to code or executes Block
      code(c) {
          if (typeof c == "function")
              c();
          else if (c !== code_1.nil)
              this._leafNode(new AnyCode(c));
          return this;
      }
      // returns code for object literal for the passed argument list of key-value pairs
      object(...keyValues) {
          const code = ["{"];
          for (const [key, value] of keyValues) {
              if (code.length > 1)
                  code.push(",");
              code.push(key);
              if (key !== value || this.opts.es5) {
                  code.push(":");
                  (0, code_1.addCodeArg)(code, value);
              }
          }
          code.push("}");
          return new code_1._Code(code);
      }
      // `if` clause (or statement if `thenBody` and, optionally, `elseBody` are passed)
      if(condition, thenBody, elseBody) {
          this._blockNode(new If(condition));
          if (thenBody && elseBody) {
              this.code(thenBody).else().code(elseBody).endIf();
          }
          else if (thenBody) {
              this.code(thenBody).endIf();
          }
          else if (elseBody) {
              throw new Error('CodeGen: "else" body without "then" body');
          }
          return this;
      }
      // `else if` clause - invalid without `if` or after `else` clauses
      elseIf(condition) {
          return this._elseNode(new If(condition));
      }
      // `else` clause - only valid after `if` or `else if` clauses
      else() {
          return this._elseNode(new Else());
      }
      // end `if` statement (needed if gen.if was used only with condition)
      endIf() {
          return this._endBlockNode(If, Else);
      }
      _for(node, forBody) {
          this._blockNode(node);
          if (forBody)
              this.code(forBody).endFor();
          return this;
      }
      // a generic `for` clause (or statement if `forBody` is passed)
      for(iteration, forBody) {
          return this._for(new ForLoop(iteration), forBody);
      }
      // `for` statement for a range of values
      forRange(nameOrPrefix, from, to, forBody, varKind = this.opts.es5 ? scope_1.varKinds.var : scope_1.varKinds.let) {
          const name = this._scope.toName(nameOrPrefix);
          return this._for(new ForRange(varKind, name, from, to), () => forBody(name));
      }
      // `for-of` statement (in es5 mode replace with a normal for loop)
      forOf(nameOrPrefix, iterable, forBody, varKind = scope_1.varKinds.const) {
          const name = this._scope.toName(nameOrPrefix);
          if (this.opts.es5) {
              const arr = iterable instanceof code_1.Name ? iterable : this.var("_arr", iterable);
              return this.forRange("_i", 0, (0, code_1._) `${arr}.length`, (i) => {
                  this.var(name, (0, code_1._) `${arr}[${i}]`);
                  forBody(name);
              });
          }
          return this._for(new ForIter("of", varKind, name, iterable), () => forBody(name));
      }
      // `for-in` statement.
      // With option `ownProperties` replaced with a `for-of` loop for object keys
      forIn(nameOrPrefix, obj, forBody, varKind = this.opts.es5 ? scope_1.varKinds.var : scope_1.varKinds.const) {
          if (this.opts.ownProperties) {
              return this.forOf(nameOrPrefix, (0, code_1._) `Object.keys(${obj})`, forBody);
          }
          const name = this._scope.toName(nameOrPrefix);
          return this._for(new ForIter("in", varKind, name, obj), () => forBody(name));
      }
      // end `for` loop
      endFor() {
          return this._endBlockNode(For);
      }
      // `label` statement
      label(label) {
          return this._leafNode(new Label(label));
      }
      // `break` statement
      break(label) {
          return this._leafNode(new Break(label));
      }
      // `return` statement
      return(value) {
          const node = new Return();
          this._blockNode(node);
          this.code(value);
          if (node.nodes.length !== 1)
              throw new Error('CodeGen: "return" should have one node');
          return this._endBlockNode(Return);
      }
      // `try` statement
      try(tryBody, catchCode, finallyCode) {
          if (!catchCode && !finallyCode)
              throw new Error('CodeGen: "try" without "catch" and "finally"');
          const node = new Try();
          this._blockNode(node);
          this.code(tryBody);
          if (catchCode) {
              const error = this.name("e");
              this._currNode = node.catch = new Catch(error);
              catchCode(error);
          }
          if (finallyCode) {
              this._currNode = node.finally = new Finally();
              this.code(finallyCode);
          }
          return this._endBlockNode(Catch, Finally);
      }
      // `throw` statement
      throw(error) {
          return this._leafNode(new Throw(error));
      }
      // start self-balancing block
      block(body, nodeCount) {
          this._blockStarts.push(this._nodes.length);
          if (body)
              this.code(body).endBlock(nodeCount);
          return this;
      }
      // end the current self-balancing block
      endBlock(nodeCount) {
          const len = this._blockStarts.pop();
          if (len === undefined)
              throw new Error("CodeGen: not in self-balancing block");
          const toClose = this._nodes.length - len;
          if (toClose < 0 || (nodeCount !== undefined && toClose !== nodeCount)) {
              throw new Error(`CodeGen: wrong number of nodes: ${toClose} vs ${nodeCount} expected`);
          }
          this._nodes.length = len;
          return this;
      }
      // `function` heading (or definition if funcBody is passed)
      func(name, args = code_1.nil, async, funcBody) {
          this._blockNode(new Func(name, args, async));
          if (funcBody)
              this.code(funcBody).endFunc();
          return this;
      }
      // end function definition
      endFunc() {
          return this._endBlockNode(Func);
      }
      optimize(n = 1) {
          while (n-- > 0) {
              this._root.optimizeNodes();
              this._root.optimizeNames(this._root.names, this._constants);
          }
      }
      _leafNode(node) {
          this._currNode.nodes.push(node);
          return this;
      }
      _blockNode(node) {
          this._currNode.nodes.push(node);
          this._nodes.push(node);
      }
      _endBlockNode(N1, N2) {
          const n = this._currNode;
          if (n instanceof N1 || (N2 && n instanceof N2)) {
              this._nodes.pop();
              return this;
          }
          throw new Error(`CodeGen: not in block "${N2 ? `${N1.kind}/${N2.kind}` : N1.kind}"`);
      }
      _elseNode(node) {
          const n = this._currNode;
          if (!(n instanceof If)) {
              throw new Error('CodeGen: "else" without "if"');
          }
          this._currNode = n.else = node;
          return this;
      }
      get _root() {
          return this._nodes[0];
      }
      get _currNode() {
          const ns = this._nodes;
          return ns[ns.length - 1];
      }
      set _currNode(node) {
          const ns = this._nodes;
          ns[ns.length - 1] = node;
      }
  }
  exports.CodeGen = CodeGen;
  function addNames(names, from) {
      for (const n in from)
          names[n] = (names[n] || 0) + (from[n] || 0);
      return names;
  }
  function addExprNames(names, from) {
      return from instanceof code_1._CodeOrName ? addNames(names, from.names) : names;
  }
  function optimizeExpr(expr, names, constants) {
      if (expr instanceof code_1.Name)
          return replaceName(expr);
      if (!canOptimize(expr))
          return expr;
      return new code_1._Code(expr._items.reduce((items, c) => {
          if (c instanceof code_1.Name)
              c = replaceName(c);
          if (c instanceof code_1._Code)
              items.push(...c._items);
          else
              items.push(c);
          return items;
      }, []));
      function replaceName(n) {
          const c = constants[n.str];
          if (c === undefined || names[n.str] !== 1)
              return n;
          delete names[n.str];
          return c;
      }
      function canOptimize(e) {
          return (e instanceof code_1._Code &&
              e._items.some((c) => c instanceof code_1.Name && names[c.str] === 1 && constants[c.str] !== undefined));
      }
  }
  function subtractNames(names, from) {
      for (const n in from)
          names[n] = (names[n] || 0) - (from[n] || 0);
  }
  function not(x) {
      return typeof x == "boolean" || typeof x == "number" || x === null ? !x : (0, code_1._) `!${par(x)}`;
  }
  exports.not = not;
  const andCode = mappend(exports.operators.AND);
  // boolean AND (&&) expression with the passed arguments
  function and(...args) {
      return args.reduce(andCode);
  }
  exports.and = and;
  const orCode = mappend(exports.operators.OR);
  // boolean OR (||) expression with the passed arguments
  function or(...args) {
      return args.reduce(orCode);
  }
  exports.or = or;
  function mappend(op) {
      return (x, y) => (x === code_1.nil ? y : y === code_1.nil ? x : (0, code_1._) `${par(x)} ${op} ${par(y)}`);
  }
  function par(x) {
      return x instanceof code_1.Name ? x : (0, code_1._) `(${x})`;
  }

  }(codegen));

  var util = {};

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.checkStrictMode = exports.getErrorPath = exports.Type = exports.useFunc = exports.setEvaluated = exports.evaluatedPropsToName = exports.mergeEvaluated = exports.eachItem = exports.unescapeJsonPointer = exports.escapeJsonPointer = exports.escapeFragment = exports.unescapeFragment = exports.schemaRefOrVal = exports.schemaHasRulesButRef = exports.schemaHasRules = exports.checkUnknownRules = exports.alwaysValidSchema = exports.toHash = void 0;
  const codegen_1 = codegen;
  const code_1 = code$1;
  // TODO refactor to use Set
  function toHash(arr) {
      const hash = {};
      for (const item of arr)
          hash[item] = true;
      return hash;
  }
  exports.toHash = toHash;
  function alwaysValidSchema(it, schema) {
      if (typeof schema == "boolean")
          return schema;
      if (Object.keys(schema).length === 0)
          return true;
      checkUnknownRules(it, schema);
      return !schemaHasRules(schema, it.self.RULES.all);
  }
  exports.alwaysValidSchema = alwaysValidSchema;
  function checkUnknownRules(it, schema = it.schema) {
      const { opts, self } = it;
      if (!opts.strictSchema)
          return;
      if (typeof schema === "boolean")
          return;
      const rules = self.RULES.keywords;
      for (const key in schema) {
          if (!rules[key])
              checkStrictMode(it, `unknown keyword: "${key}"`);
      }
  }
  exports.checkUnknownRules = checkUnknownRules;
  function schemaHasRules(schema, rules) {
      if (typeof schema == "boolean")
          return !schema;
      for (const key in schema)
          if (rules[key])
              return true;
      return false;
  }
  exports.schemaHasRules = schemaHasRules;
  function schemaHasRulesButRef(schema, RULES) {
      if (typeof schema == "boolean")
          return !schema;
      for (const key in schema)
          if (key !== "$ref" && RULES.all[key])
              return true;
      return false;
  }
  exports.schemaHasRulesButRef = schemaHasRulesButRef;
  function schemaRefOrVal({ topSchemaRef, schemaPath }, schema, keyword, $data) {
      if (!$data) {
          if (typeof schema == "number" || typeof schema == "boolean")
              return schema;
          if (typeof schema == "string")
              return (0, codegen_1._) `${schema}`;
      }
      return (0, codegen_1._) `${topSchemaRef}${schemaPath}${(0, codegen_1.getProperty)(keyword)}`;
  }
  exports.schemaRefOrVal = schemaRefOrVal;
  function unescapeFragment(str) {
      return unescapeJsonPointer(decodeURIComponent(str));
  }
  exports.unescapeFragment = unescapeFragment;
  function escapeFragment(str) {
      return encodeURIComponent(escapeJsonPointer(str));
  }
  exports.escapeFragment = escapeFragment;
  function escapeJsonPointer(str) {
      if (typeof str == "number")
          return `${str}`;
      return str.replace(/~/g, "~0").replace(/\//g, "~1");
  }
  exports.escapeJsonPointer = escapeJsonPointer;
  function unescapeJsonPointer(str) {
      return str.replace(/~1/g, "/").replace(/~0/g, "~");
  }
  exports.unescapeJsonPointer = unescapeJsonPointer;
  function eachItem(xs, f) {
      if (Array.isArray(xs)) {
          for (const x of xs)
              f(x);
      }
      else {
          f(xs);
      }
  }
  exports.eachItem = eachItem;
  function makeMergeEvaluated({ mergeNames, mergeToName, mergeValues, resultToName, }) {
      return (gen, from, to, toName) => {
          const res = to === undefined
              ? from
              : to instanceof codegen_1.Name
                  ? (from instanceof codegen_1.Name ? mergeNames(gen, from, to) : mergeToName(gen, from, to), to)
                  : from instanceof codegen_1.Name
                      ? (mergeToName(gen, to, from), from)
                      : mergeValues(from, to);
          return toName === codegen_1.Name && !(res instanceof codegen_1.Name) ? resultToName(gen, res) : res;
      };
  }
  exports.mergeEvaluated = {
      props: makeMergeEvaluated({
          mergeNames: (gen, from, to) => gen.if((0, codegen_1._) `${to} !== true && ${from} !== undefined`, () => {
              gen.if((0, codegen_1._) `${from} === true`, () => gen.assign(to, true), () => gen.assign(to, (0, codegen_1._) `${to} || {}`).code((0, codegen_1._) `Object.assign(${to}, ${from})`));
          }),
          mergeToName: (gen, from, to) => gen.if((0, codegen_1._) `${to} !== true`, () => {
              if (from === true) {
                  gen.assign(to, true);
              }
              else {
                  gen.assign(to, (0, codegen_1._) `${to} || {}`);
                  setEvaluated(gen, to, from);
              }
          }),
          mergeValues: (from, to) => (from === true ? true : { ...from, ...to }),
          resultToName: evaluatedPropsToName,
      }),
      items: makeMergeEvaluated({
          mergeNames: (gen, from, to) => gen.if((0, codegen_1._) `${to} !== true && ${from} !== undefined`, () => gen.assign(to, (0, codegen_1._) `${from} === true ? true : ${to} > ${from} ? ${to} : ${from}`)),
          mergeToName: (gen, from, to) => gen.if((0, codegen_1._) `${to} !== true`, () => gen.assign(to, from === true ? true : (0, codegen_1._) `${to} > ${from} ? ${to} : ${from}`)),
          mergeValues: (from, to) => (from === true ? true : Math.max(from, to)),
          resultToName: (gen, items) => gen.var("items", items),
      }),
  };
  function evaluatedPropsToName(gen, ps) {
      if (ps === true)
          return gen.var("props", true);
      const props = gen.var("props", (0, codegen_1._) `{}`);
      if (ps !== undefined)
          setEvaluated(gen, props, ps);
      return props;
  }
  exports.evaluatedPropsToName = evaluatedPropsToName;
  function setEvaluated(gen, props, ps) {
      Object.keys(ps).forEach((p) => gen.assign((0, codegen_1._) `${props}${(0, codegen_1.getProperty)(p)}`, true));
  }
  exports.setEvaluated = setEvaluated;
  const snippets = {};
  function useFunc(gen, f) {
      return gen.scopeValue("func", {
          ref: f,
          code: snippets[f.code] || (snippets[f.code] = new code_1._Code(f.code)),
      });
  }
  exports.useFunc = useFunc;
  var Type;
  (function (Type) {
      Type[Type["Num"] = 0] = "Num";
      Type[Type["Str"] = 1] = "Str";
  })(Type = exports.Type || (exports.Type = {}));
  function getErrorPath(dataProp, dataPropType, jsPropertySyntax) {
      // let path
      if (dataProp instanceof codegen_1.Name) {
          const isNumber = dataPropType === Type.Num;
          return jsPropertySyntax
              ? isNumber
                  ? (0, codegen_1._) `"[" + ${dataProp} + "]"`
                  : (0, codegen_1._) `"['" + ${dataProp} + "']"`
              : isNumber
                  ? (0, codegen_1._) `"/" + ${dataProp}`
                  : (0, codegen_1._) `"/" + ${dataProp}.replace(/~/g, "~0").replace(/\\//g, "~1")`; // TODO maybe use global escapePointer
      }
      return jsPropertySyntax ? (0, codegen_1.getProperty)(dataProp).toString() : "/" + escapeJsonPointer(dataProp);
  }
  exports.getErrorPath = getErrorPath;
  function checkStrictMode(it, msg, mode = it.opts.strictSchema) {
      if (!mode)
          return;
      msg = `strict mode: ${msg}`;
      if (mode === true)
          throw new Error(msg);
      it.self.logger.warn(msg);
  }
  exports.checkStrictMode = checkStrictMode;

  }(util));

  var names$1 = {};

  Object.defineProperty(names$1, "__esModule", { value: true });
  const codegen_1$t = codegen;
  const names = {
      // validation function arguments
      data: new codegen_1$t.Name("data"),
      // args passed from referencing schema
      valCxt: new codegen_1$t.Name("valCxt"),
      instancePath: new codegen_1$t.Name("instancePath"),
      parentData: new codegen_1$t.Name("parentData"),
      parentDataProperty: new codegen_1$t.Name("parentDataProperty"),
      rootData: new codegen_1$t.Name("rootData"),
      dynamicAnchors: new codegen_1$t.Name("dynamicAnchors"),
      // function scoped variables
      vErrors: new codegen_1$t.Name("vErrors"),
      errors: new codegen_1$t.Name("errors"),
      this: new codegen_1$t.Name("this"),
      // "globals"
      self: new codegen_1$t.Name("self"),
      scope: new codegen_1$t.Name("scope"),
      // JTD serialize/parse name for JSON string and position
      json: new codegen_1$t.Name("json"),
      jsonPos: new codegen_1$t.Name("jsonPos"),
      jsonLen: new codegen_1$t.Name("jsonLen"),
      jsonPart: new codegen_1$t.Name("jsonPart"),
  };
  names$1.default = names;

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.extendErrors = exports.resetErrorsCount = exports.reportExtraError = exports.reportError = exports.keyword$DataError = exports.keywordError = void 0;
  const codegen_1 = codegen;
  const util_1 = util;
  const names_1 = names$1;
  exports.keywordError = {
      message: ({ keyword }) => (0, codegen_1.str) `must pass "${keyword}" keyword validation`,
  };
  exports.keyword$DataError = {
      message: ({ keyword, schemaType }) => schemaType
          ? (0, codegen_1.str) `"${keyword}" keyword must be ${schemaType} ($data)`
          : (0, codegen_1.str) `"${keyword}" keyword is invalid ($data)`,
  };
  function reportError(cxt, error = exports.keywordError, errorPaths, overrideAllErrors) {
      const { it } = cxt;
      const { gen, compositeRule, allErrors } = it;
      const errObj = errorObjectCode(cxt, error, errorPaths);
      if (overrideAllErrors !== null && overrideAllErrors !== void 0 ? overrideAllErrors : (compositeRule || allErrors)) {
          addError(gen, errObj);
      }
      else {
          returnErrors(it, (0, codegen_1._) `[${errObj}]`);
      }
  }
  exports.reportError = reportError;
  function reportExtraError(cxt, error = exports.keywordError, errorPaths) {
      const { it } = cxt;
      const { gen, compositeRule, allErrors } = it;
      const errObj = errorObjectCode(cxt, error, errorPaths);
      addError(gen, errObj);
      if (!(compositeRule || allErrors)) {
          returnErrors(it, names_1.default.vErrors);
      }
  }
  exports.reportExtraError = reportExtraError;
  function resetErrorsCount(gen, errsCount) {
      gen.assign(names_1.default.errors, errsCount);
      gen.if((0, codegen_1._) `${names_1.default.vErrors} !== null`, () => gen.if(errsCount, () => gen.assign((0, codegen_1._) `${names_1.default.vErrors}.length`, errsCount), () => gen.assign(names_1.default.vErrors, null)));
  }
  exports.resetErrorsCount = resetErrorsCount;
  function extendErrors({ gen, keyword, schemaValue, data, errsCount, it, }) {
      /* istanbul ignore if */
      if (errsCount === undefined)
          throw new Error("ajv implementation error");
      const err = gen.name("err");
      gen.forRange("i", errsCount, names_1.default.errors, (i) => {
          gen.const(err, (0, codegen_1._) `${names_1.default.vErrors}[${i}]`);
          gen.if((0, codegen_1._) `${err}.instancePath === undefined`, () => gen.assign((0, codegen_1._) `${err}.instancePath`, (0, codegen_1.strConcat)(names_1.default.instancePath, it.errorPath)));
          gen.assign((0, codegen_1._) `${err}.schemaPath`, (0, codegen_1.str) `${it.errSchemaPath}/${keyword}`);
          if (it.opts.verbose) {
              gen.assign((0, codegen_1._) `${err}.schema`, schemaValue);
              gen.assign((0, codegen_1._) `${err}.data`, data);
          }
      });
  }
  exports.extendErrors = extendErrors;
  function addError(gen, errObj) {
      const err = gen.const("err", errObj);
      gen.if((0, codegen_1._) `${names_1.default.vErrors} === null`, () => gen.assign(names_1.default.vErrors, (0, codegen_1._) `[${err}]`), (0, codegen_1._) `${names_1.default.vErrors}.push(${err})`);
      gen.code((0, codegen_1._) `${names_1.default.errors}++`);
  }
  function returnErrors(it, errs) {
      const { gen, validateName, schemaEnv } = it;
      if (schemaEnv.$async) {
          gen.throw((0, codegen_1._) `new ${it.ValidationError}(${errs})`);
      }
      else {
          gen.assign((0, codegen_1._) `${validateName}.errors`, errs);
          gen.return(false);
      }
  }
  const E = {
      keyword: new codegen_1.Name("keyword"),
      schemaPath: new codegen_1.Name("schemaPath"),
      params: new codegen_1.Name("params"),
      propertyName: new codegen_1.Name("propertyName"),
      message: new codegen_1.Name("message"),
      schema: new codegen_1.Name("schema"),
      parentSchema: new codegen_1.Name("parentSchema"),
  };
  function errorObjectCode(cxt, error, errorPaths) {
      const { createErrors } = cxt.it;
      if (createErrors === false)
          return (0, codegen_1._) `{}`;
      return errorObject(cxt, error, errorPaths);
  }
  function errorObject(cxt, error, errorPaths = {}) {
      const { gen, it } = cxt;
      const keyValues = [
          errorInstancePath(it, errorPaths),
          errorSchemaPath(cxt, errorPaths),
      ];
      extraErrorProps(cxt, error, keyValues);
      return gen.object(...keyValues);
  }
  function errorInstancePath({ errorPath }, { instancePath }) {
      const instPath = instancePath
          ? (0, codegen_1.str) `${errorPath}${(0, util_1.getErrorPath)(instancePath, util_1.Type.Str)}`
          : errorPath;
      return [names_1.default.instancePath, (0, codegen_1.strConcat)(names_1.default.instancePath, instPath)];
  }
  function errorSchemaPath({ keyword, it: { errSchemaPath } }, { schemaPath, parentSchema }) {
      let schPath = parentSchema ? errSchemaPath : (0, codegen_1.str) `${errSchemaPath}/${keyword}`;
      if (schemaPath) {
          schPath = (0, codegen_1.str) `${schPath}${(0, util_1.getErrorPath)(schemaPath, util_1.Type.Str)}`;
      }
      return [E.schemaPath, schPath];
  }
  function extraErrorProps(cxt, { params, message }, keyValues) {
      const { keyword, data, schemaValue, it } = cxt;
      const { opts, propertyName, topSchemaRef, schemaPath } = it;
      keyValues.push([E.keyword, keyword], [E.params, typeof params == "function" ? params(cxt) : params || (0, codegen_1._) `{}`]);
      if (opts.messages) {
          keyValues.push([E.message, typeof message == "function" ? message(cxt) : message]);
      }
      if (opts.verbose) {
          keyValues.push([E.schema, schemaValue], [E.parentSchema, (0, codegen_1._) `${topSchemaRef}${schemaPath}`], [names_1.default.data, data]);
      }
      if (propertyName)
          keyValues.push([E.propertyName, propertyName]);
  }

  }(errors));

  Object.defineProperty(boolSchema, "__esModule", { value: true });
  boolSchema.boolOrEmptySchema = boolSchema.topBoolOrEmptySchema = void 0;
  const errors_1$2 = errors;
  const codegen_1$s = codegen;
  const names_1$6 = names$1;
  const boolError = {
      message: "boolean schema is false",
  };
  function topBoolOrEmptySchema(it) {
      const { gen, schema, validateName } = it;
      if (schema === false) {
          falseSchemaError(it, false);
      }
      else if (typeof schema == "object" && schema.$async === true) {
          gen.return(names_1$6.default.data);
      }
      else {
          gen.assign((0, codegen_1$s._) `${validateName}.errors`, null);
          gen.return(true);
      }
  }
  boolSchema.topBoolOrEmptySchema = topBoolOrEmptySchema;
  function boolOrEmptySchema(it, valid) {
      const { gen, schema } = it;
      if (schema === false) {
          gen.var(valid, false); // TODO var
          falseSchemaError(it);
      }
      else {
          gen.var(valid, true); // TODO var
      }
  }
  boolSchema.boolOrEmptySchema = boolOrEmptySchema;
  function falseSchemaError(it, overrideAllErrors) {
      const { gen, data } = it;
      // TODO maybe some other interface should be used for non-keyword validation errors...
      const cxt = {
          gen,
          keyword: "false schema",
          data,
          schema: false,
          schemaCode: false,
          schemaValue: false,
          params: {},
          it,
      };
      (0, errors_1$2.reportError)(cxt, boolError, undefined, overrideAllErrors);
  }

  var dataType = {};

  var rules = {};

  Object.defineProperty(rules, "__esModule", { value: true });
  rules.getRules = rules.isJSONType = void 0;
  const _jsonTypes = ["string", "number", "integer", "boolean", "null", "object", "array"];
  const jsonTypes = new Set(_jsonTypes);
  function isJSONType(x) {
      return typeof x == "string" && jsonTypes.has(x);
  }
  rules.isJSONType = isJSONType;
  function getRules() {
      const groups = {
          number: { type: "number", rules: [] },
          string: { type: "string", rules: [] },
          array: { type: "array", rules: [] },
          object: { type: "object", rules: [] },
      };
      return {
          types: { ...groups, integer: true, boolean: true, null: true },
          rules: [{ rules: [] }, groups.number, groups.string, groups.array, groups.object],
          post: { rules: [] },
          all: {},
          keywords: {},
      };
  }
  rules.getRules = getRules;

  var applicability = {};

  Object.defineProperty(applicability, "__esModule", { value: true });
  applicability.shouldUseRule = applicability.shouldUseGroup = applicability.schemaHasRulesForType = void 0;
  function schemaHasRulesForType({ schema, self }, type) {
      const group = self.RULES.types[type];
      return group && group !== true && shouldUseGroup(schema, group);
  }
  applicability.schemaHasRulesForType = schemaHasRulesForType;
  function shouldUseGroup(schema, group) {
      return group.rules.some((rule) => shouldUseRule(schema, rule));
  }
  applicability.shouldUseGroup = shouldUseGroup;
  function shouldUseRule(schema, rule) {
      var _a;
      return (schema[rule.keyword] !== undefined ||
          ((_a = rule.definition.implements) === null || _a === void 0 ? void 0 : _a.some((kwd) => schema[kwd] !== undefined)));
  }
  applicability.shouldUseRule = shouldUseRule;

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.reportTypeError = exports.checkDataTypes = exports.checkDataType = exports.coerceAndCheckDataType = exports.getJSONTypes = exports.getSchemaTypes = exports.DataType = void 0;
  const rules_1 = rules;
  const applicability_1 = applicability;
  const errors_1 = errors;
  const codegen_1 = codegen;
  const util_1 = util;
  var DataType;
  (function (DataType) {
      DataType[DataType["Correct"] = 0] = "Correct";
      DataType[DataType["Wrong"] = 1] = "Wrong";
  })(DataType = exports.DataType || (exports.DataType = {}));
  function getSchemaTypes(schema) {
      const types = getJSONTypes(schema.type);
      const hasNull = types.includes("null");
      if (hasNull) {
          if (schema.nullable === false)
              throw new Error("type: null contradicts nullable: false");
      }
      else {
          if (!types.length && schema.nullable !== undefined) {
              throw new Error('"nullable" cannot be used without "type"');
          }
          if (schema.nullable === true)
              types.push("null");
      }
      return types;
  }
  exports.getSchemaTypes = getSchemaTypes;
  function getJSONTypes(ts) {
      const types = Array.isArray(ts) ? ts : ts ? [ts] : [];
      if (types.every(rules_1.isJSONType))
          return types;
      throw new Error("type must be JSONType or JSONType[]: " + types.join(","));
  }
  exports.getJSONTypes = getJSONTypes;
  function coerceAndCheckDataType(it, types) {
      const { gen, data, opts } = it;
      const coerceTo = coerceToTypes(types, opts.coerceTypes);
      const checkTypes = types.length > 0 &&
          !(coerceTo.length === 0 && types.length === 1 && (0, applicability_1.schemaHasRulesForType)(it, types[0]));
      if (checkTypes) {
          const wrongType = checkDataTypes(types, data, opts.strictNumbers, DataType.Wrong);
          gen.if(wrongType, () => {
              if (coerceTo.length)
                  coerceData(it, types, coerceTo);
              else
                  reportTypeError(it);
          });
      }
      return checkTypes;
  }
  exports.coerceAndCheckDataType = coerceAndCheckDataType;
  const COERCIBLE = new Set(["string", "number", "integer", "boolean", "null"]);
  function coerceToTypes(types, coerceTypes) {
      return coerceTypes
          ? types.filter((t) => COERCIBLE.has(t) || (coerceTypes === "array" && t === "array"))
          : [];
  }
  function coerceData(it, types, coerceTo) {
      const { gen, data, opts } = it;
      const dataType = gen.let("dataType", (0, codegen_1._) `typeof ${data}`);
      const coerced = gen.let("coerced", (0, codegen_1._) `undefined`);
      if (opts.coerceTypes === "array") {
          gen.if((0, codegen_1._) `${dataType} == 'object' && Array.isArray(${data}) && ${data}.length == 1`, () => gen
              .assign(data, (0, codegen_1._) `${data}[0]`)
              .assign(dataType, (0, codegen_1._) `typeof ${data}`)
              .if(checkDataTypes(types, data, opts.strictNumbers), () => gen.assign(coerced, data)));
      }
      gen.if((0, codegen_1._) `${coerced} !== undefined`);
      for (const t of coerceTo) {
          if (COERCIBLE.has(t) || (t === "array" && opts.coerceTypes === "array")) {
              coerceSpecificType(t);
          }
      }
      gen.else();
      reportTypeError(it);
      gen.endIf();
      gen.if((0, codegen_1._) `${coerced} !== undefined`, () => {
          gen.assign(data, coerced);
          assignParentData(it, coerced);
      });
      function coerceSpecificType(t) {
          switch (t) {
              case "string":
                  gen
                      .elseIf((0, codegen_1._) `${dataType} == "number" || ${dataType} == "boolean"`)
                      .assign(coerced, (0, codegen_1._) `"" + ${data}`)
                      .elseIf((0, codegen_1._) `${data} === null`)
                      .assign(coerced, (0, codegen_1._) `""`);
                  return;
              case "number":
                  gen
                      .elseIf((0, codegen_1._) `${dataType} == "boolean" || ${data} === null
              || (${dataType} == "string" && ${data} && ${data} == +${data})`)
                      .assign(coerced, (0, codegen_1._) `+${data}`);
                  return;
              case "integer":
                  gen
                      .elseIf((0, codegen_1._) `${dataType} === "boolean" || ${data} === null
              || (${dataType} === "string" && ${data} && ${data} == +${data} && !(${data} % 1))`)
                      .assign(coerced, (0, codegen_1._) `+${data}`);
                  return;
              case "boolean":
                  gen
                      .elseIf((0, codegen_1._) `${data} === "false" || ${data} === 0 || ${data} === null`)
                      .assign(coerced, false)
                      .elseIf((0, codegen_1._) `${data} === "true" || ${data} === 1`)
                      .assign(coerced, true);
                  return;
              case "null":
                  gen.elseIf((0, codegen_1._) `${data} === "" || ${data} === 0 || ${data} === false`);
                  gen.assign(coerced, null);
                  return;
              case "array":
                  gen
                      .elseIf((0, codegen_1._) `${dataType} === "string" || ${dataType} === "number"
              || ${dataType} === "boolean" || ${data} === null`)
                      .assign(coerced, (0, codegen_1._) `[${data}]`);
          }
      }
  }
  function assignParentData({ gen, parentData, parentDataProperty }, expr) {
      // TODO use gen.property
      gen.if((0, codegen_1._) `${parentData} !== undefined`, () => gen.assign((0, codegen_1._) `${parentData}[${parentDataProperty}]`, expr));
  }
  function checkDataType(dataType, data, strictNums, correct = DataType.Correct) {
      const EQ = correct === DataType.Correct ? codegen_1.operators.EQ : codegen_1.operators.NEQ;
      let cond;
      switch (dataType) {
          case "null":
              return (0, codegen_1._) `${data} ${EQ} null`;
          case "array":
              cond = (0, codegen_1._) `Array.isArray(${data})`;
              break;
          case "object":
              cond = (0, codegen_1._) `${data} && typeof ${data} == "object" && !Array.isArray(${data})`;
              break;
          case "integer":
              cond = numCond((0, codegen_1._) `!(${data} % 1) && !isNaN(${data})`);
              break;
          case "number":
              cond = numCond();
              break;
          default:
              return (0, codegen_1._) `typeof ${data} ${EQ} ${dataType}`;
      }
      return correct === DataType.Correct ? cond : (0, codegen_1.not)(cond);
      function numCond(_cond = codegen_1.nil) {
          return (0, codegen_1.and)((0, codegen_1._) `typeof ${data} == "number"`, _cond, strictNums ? (0, codegen_1._) `isFinite(${data})` : codegen_1.nil);
      }
  }
  exports.checkDataType = checkDataType;
  function checkDataTypes(dataTypes, data, strictNums, correct) {
      if (dataTypes.length === 1) {
          return checkDataType(dataTypes[0], data, strictNums, correct);
      }
      let cond;
      const types = (0, util_1.toHash)(dataTypes);
      if (types.array && types.object) {
          const notObj = (0, codegen_1._) `typeof ${data} != "object"`;
          cond = types.null ? notObj : (0, codegen_1._) `!${data} || ${notObj}`;
          delete types.null;
          delete types.array;
          delete types.object;
      }
      else {
          cond = codegen_1.nil;
      }
      if (types.number)
          delete types.integer;
      for (const t in types)
          cond = (0, codegen_1.and)(cond, checkDataType(t, data, strictNums, correct));
      return cond;
  }
  exports.checkDataTypes = checkDataTypes;
  const typeError = {
      message: ({ schema }) => `must be ${schema}`,
      params: ({ schema, schemaValue }) => typeof schema == "string" ? (0, codegen_1._) `{type: ${schema}}` : (0, codegen_1._) `{type: ${schemaValue}}`,
  };
  function reportTypeError(it) {
      const cxt = getTypeErrorContext(it);
      (0, errors_1.reportError)(cxt, typeError);
  }
  exports.reportTypeError = reportTypeError;
  function getTypeErrorContext(it) {
      const { gen, data, schema } = it;
      const schemaCode = (0, util_1.schemaRefOrVal)(it, schema, "type");
      return {
          gen,
          keyword: "type",
          data,
          schema: schema.type,
          schemaCode,
          schemaValue: schemaCode,
          parentSchema: schema,
          params: {},
          it,
      };
  }

  }(dataType));

  var defaults = {};

  Object.defineProperty(defaults, "__esModule", { value: true });
  defaults.assignDefaults = void 0;
  const codegen_1$r = codegen;
  const util_1$p = util;
  function assignDefaults(it, ty) {
      const { properties, items } = it.schema;
      if (ty === "object" && properties) {
          for (const key in properties) {
              assignDefault(it, key, properties[key].default);
          }
      }
      else if (ty === "array" && Array.isArray(items)) {
          items.forEach((sch, i) => assignDefault(it, i, sch.default));
      }
  }
  defaults.assignDefaults = assignDefaults;
  function assignDefault(it, prop, defaultValue) {
      const { gen, compositeRule, data, opts } = it;
      if (defaultValue === undefined)
          return;
      const childData = (0, codegen_1$r._) `${data}${(0, codegen_1$r.getProperty)(prop)}`;
      if (compositeRule) {
          (0, util_1$p.checkStrictMode)(it, `default is ignored for: ${childData}`);
          return;
      }
      let condition = (0, codegen_1$r._) `${childData} === undefined`;
      if (opts.useDefaults === "empty") {
          condition = (0, codegen_1$r._) `${condition} || ${childData} === null || ${childData} === ""`;
      }
      // `${childData} === undefined` +
      // (opts.useDefaults === "empty" ? ` || ${childData} === null || ${childData} === ""` : "")
      gen.if(condition, (0, codegen_1$r._) `${childData} = ${(0, codegen_1$r.stringify)(defaultValue)}`);
  }

  var keyword = {};

  var code = {};

  Object.defineProperty(code, "__esModule", { value: true });
  code.validateUnion = code.validateArray = code.usePattern = code.callValidateCode = code.schemaProperties = code.allSchemaProperties = code.noPropertyInData = code.propertyInData = code.isOwnProperty = code.hasPropFunc = code.reportMissingProp = code.checkMissingProp = code.checkReportMissingProp = void 0;
  const codegen_1$q = codegen;
  const util_1$o = util;
  const names_1$5 = names$1;
  const util_2$1 = util;
  function checkReportMissingProp(cxt, prop) {
      const { gen, data, it } = cxt;
      gen.if(noPropertyInData(gen, data, prop, it.opts.ownProperties), () => {
          cxt.setParams({ missingProperty: (0, codegen_1$q._) `${prop}` }, true);
          cxt.error();
      });
  }
  code.checkReportMissingProp = checkReportMissingProp;
  function checkMissingProp({ gen, data, it: { opts } }, properties, missing) {
      return (0, codegen_1$q.or)(...properties.map((prop) => (0, codegen_1$q.and)(noPropertyInData(gen, data, prop, opts.ownProperties), (0, codegen_1$q._) `${missing} = ${prop}`)));
  }
  code.checkMissingProp = checkMissingProp;
  function reportMissingProp(cxt, missing) {
      cxt.setParams({ missingProperty: missing }, true);
      cxt.error();
  }
  code.reportMissingProp = reportMissingProp;
  function hasPropFunc(gen) {
      return gen.scopeValue("func", {
          // eslint-disable-next-line @typescript-eslint/unbound-method
          ref: Object.prototype.hasOwnProperty,
          code: (0, codegen_1$q._) `Object.prototype.hasOwnProperty`,
      });
  }
  code.hasPropFunc = hasPropFunc;
  function isOwnProperty(gen, data, property) {
      return (0, codegen_1$q._) `${hasPropFunc(gen)}.call(${data}, ${property})`;
  }
  code.isOwnProperty = isOwnProperty;
  function propertyInData(gen, data, property, ownProperties) {
      const cond = (0, codegen_1$q._) `${data}${(0, codegen_1$q.getProperty)(property)} !== undefined`;
      return ownProperties ? (0, codegen_1$q._) `${cond} && ${isOwnProperty(gen, data, property)}` : cond;
  }
  code.propertyInData = propertyInData;
  function noPropertyInData(gen, data, property, ownProperties) {
      const cond = (0, codegen_1$q._) `${data}${(0, codegen_1$q.getProperty)(property)} === undefined`;
      return ownProperties ? (0, codegen_1$q.or)(cond, (0, codegen_1$q.not)(isOwnProperty(gen, data, property))) : cond;
  }
  code.noPropertyInData = noPropertyInData;
  function allSchemaProperties(schemaMap) {
      return schemaMap ? Object.keys(schemaMap).filter((p) => p !== "__proto__") : [];
  }
  code.allSchemaProperties = allSchemaProperties;
  function schemaProperties(it, schemaMap) {
      return allSchemaProperties(schemaMap).filter((p) => !(0, util_1$o.alwaysValidSchema)(it, schemaMap[p]));
  }
  code.schemaProperties = schemaProperties;
  function callValidateCode({ schemaCode, data, it: { gen, topSchemaRef, schemaPath, errorPath }, it }, func, context, passSchema) {
      const dataAndSchema = passSchema ? (0, codegen_1$q._) `${schemaCode}, ${data}, ${topSchemaRef}${schemaPath}` : data;
      const valCxt = [
          [names_1$5.default.instancePath, (0, codegen_1$q.strConcat)(names_1$5.default.instancePath, errorPath)],
          [names_1$5.default.parentData, it.parentData],
          [names_1$5.default.parentDataProperty, it.parentDataProperty],
          [names_1$5.default.rootData, names_1$5.default.rootData],
      ];
      if (it.opts.dynamicRef)
          valCxt.push([names_1$5.default.dynamicAnchors, names_1$5.default.dynamicAnchors]);
      const args = (0, codegen_1$q._) `${dataAndSchema}, ${gen.object(...valCxt)}`;
      return context !== codegen_1$q.nil ? (0, codegen_1$q._) `${func}.call(${context}, ${args})` : (0, codegen_1$q._) `${func}(${args})`;
  }
  code.callValidateCode = callValidateCode;
  const newRegExp = (0, codegen_1$q._) `new RegExp`;
  function usePattern({ gen, it: { opts } }, pattern) {
      const u = opts.unicodeRegExp ? "u" : "";
      const { regExp } = opts.code;
      const rx = regExp(pattern, u);
      return gen.scopeValue("pattern", {
          key: rx.toString(),
          ref: rx,
          code: (0, codegen_1$q._) `${regExp.code === "new RegExp" ? newRegExp : (0, util_2$1.useFunc)(gen, regExp)}(${pattern}, ${u})`,
      });
  }
  code.usePattern = usePattern;
  function validateArray(cxt) {
      const { gen, data, keyword, it } = cxt;
      const valid = gen.name("valid");
      if (it.allErrors) {
          const validArr = gen.let("valid", true);
          validateItems(() => gen.assign(validArr, false));
          return validArr;
      }
      gen.var(valid, true);
      validateItems(() => gen.break());
      return valid;
      function validateItems(notValid) {
          const len = gen.const("len", (0, codegen_1$q._) `${data}.length`);
          gen.forRange("i", 0, len, (i) => {
              cxt.subschema({
                  keyword,
                  dataProp: i,
                  dataPropType: util_1$o.Type.Num,
              }, valid);
              gen.if((0, codegen_1$q.not)(valid), notValid);
          });
      }
  }
  code.validateArray = validateArray;
  function validateUnion(cxt) {
      const { gen, schema, keyword, it } = cxt;
      /* istanbul ignore if */
      if (!Array.isArray(schema))
          throw new Error("ajv implementation error");
      const alwaysValid = schema.some((sch) => (0, util_1$o.alwaysValidSchema)(it, sch));
      if (alwaysValid && !it.opts.unevaluated)
          return;
      const valid = gen.let("valid", false);
      const schValid = gen.name("_valid");
      gen.block(() => schema.forEach((_sch, i) => {
          const schCxt = cxt.subschema({
              keyword,
              schemaProp: i,
              compositeRule: true,
          }, schValid);
          gen.assign(valid, (0, codegen_1$q._) `${valid} || ${schValid}`);
          const merged = cxt.mergeValidEvaluated(schCxt, schValid);
          // can short-circuit if `unevaluatedProperties/Items` not supported (opts.unevaluated !== true)
          // or if all properties and items were evaluated (it.props === true && it.items === true)
          if (!merged)
              gen.if((0, codegen_1$q.not)(valid));
      }));
      cxt.result(valid, () => cxt.reset(), () => cxt.error(true));
  }
  code.validateUnion = validateUnion;

  Object.defineProperty(keyword, "__esModule", { value: true });
  keyword.validateKeywordUsage = keyword.validSchemaType = keyword.funcKeywordCode = keyword.macroKeywordCode = void 0;
  const codegen_1$p = codegen;
  const names_1$4 = names$1;
  const code_1$9 = code;
  const errors_1$1 = errors;
  function macroKeywordCode(cxt, def) {
      const { gen, keyword, schema, parentSchema, it } = cxt;
      const macroSchema = def.macro.call(it.self, schema, parentSchema, it);
      const schemaRef = useKeyword(gen, keyword, macroSchema);
      if (it.opts.validateSchema !== false)
          it.self.validateSchema(macroSchema, true);
      const valid = gen.name("valid");
      cxt.subschema({
          schema: macroSchema,
          schemaPath: codegen_1$p.nil,
          errSchemaPath: `${it.errSchemaPath}/${keyword}`,
          topSchemaRef: schemaRef,
          compositeRule: true,
      }, valid);
      cxt.pass(valid, () => cxt.error(true));
  }
  keyword.macroKeywordCode = macroKeywordCode;
  function funcKeywordCode(cxt, def) {
      var _a;
      const { gen, keyword, schema, parentSchema, $data, it } = cxt;
      checkAsyncKeyword(it, def);
      const validate = !$data && def.compile ? def.compile.call(it.self, schema, parentSchema, it) : def.validate;
      const validateRef = useKeyword(gen, keyword, validate);
      const valid = gen.let("valid");
      cxt.block$data(valid, validateKeyword);
      cxt.ok((_a = def.valid) !== null && _a !== void 0 ? _a : valid);
      function validateKeyword() {
          if (def.errors === false) {
              assignValid();
              if (def.modifying)
                  modifyData(cxt);
              reportErrs(() => cxt.error());
          }
          else {
              const ruleErrs = def.async ? validateAsync() : validateSync();
              if (def.modifying)
                  modifyData(cxt);
              reportErrs(() => addErrs(cxt, ruleErrs));
          }
      }
      function validateAsync() {
          const ruleErrs = gen.let("ruleErrs", null);
          gen.try(() => assignValid((0, codegen_1$p._) `await `), (e) => gen.assign(valid, false).if((0, codegen_1$p._) `${e} instanceof ${it.ValidationError}`, () => gen.assign(ruleErrs, (0, codegen_1$p._) `${e}.errors`), () => gen.throw(e)));
          return ruleErrs;
      }
      function validateSync() {
          const validateErrs = (0, codegen_1$p._) `${validateRef}.errors`;
          gen.assign(validateErrs, null);
          assignValid(codegen_1$p.nil);
          return validateErrs;
      }
      function assignValid(_await = def.async ? (0, codegen_1$p._) `await ` : codegen_1$p.nil) {
          const passCxt = it.opts.passContext ? names_1$4.default.this : names_1$4.default.self;
          const passSchema = !(("compile" in def && !$data) || def.schema === false);
          gen.assign(valid, (0, codegen_1$p._) `${_await}${(0, code_1$9.callValidateCode)(cxt, validateRef, passCxt, passSchema)}`, def.modifying);
      }
      function reportErrs(errors) {
          var _a;
          gen.if((0, codegen_1$p.not)((_a = def.valid) !== null && _a !== void 0 ? _a : valid), errors);
      }
  }
  keyword.funcKeywordCode = funcKeywordCode;
  function modifyData(cxt) {
      const { gen, data, it } = cxt;
      gen.if(it.parentData, () => gen.assign(data, (0, codegen_1$p._) `${it.parentData}[${it.parentDataProperty}]`));
  }
  function addErrs(cxt, errs) {
      const { gen } = cxt;
      gen.if((0, codegen_1$p._) `Array.isArray(${errs})`, () => {
          gen
              .assign(names_1$4.default.vErrors, (0, codegen_1$p._) `${names_1$4.default.vErrors} === null ? ${errs} : ${names_1$4.default.vErrors}.concat(${errs})`)
              .assign(names_1$4.default.errors, (0, codegen_1$p._) `${names_1$4.default.vErrors}.length`);
          (0, errors_1$1.extendErrors)(cxt);
      }, () => cxt.error());
  }
  function checkAsyncKeyword({ schemaEnv }, def) {
      if (def.async && !schemaEnv.$async)
          throw new Error("async keyword in sync schema");
  }
  function useKeyword(gen, keyword, result) {
      if (result === undefined)
          throw new Error(`keyword "${keyword}" failed to compile`);
      return gen.scopeValue("keyword", typeof result == "function" ? { ref: result } : { ref: result, code: (0, codegen_1$p.stringify)(result) });
  }
  function validSchemaType(schema, schemaType, allowUndefined = false) {
      // TODO add tests
      return (!schemaType.length ||
          schemaType.some((st) => st === "array"
              ? Array.isArray(schema)
              : st === "object"
                  ? schema && typeof schema == "object" && !Array.isArray(schema)
                  : typeof schema == st || (allowUndefined && typeof schema == "undefined")));
  }
  keyword.validSchemaType = validSchemaType;
  function validateKeywordUsage({ schema, opts, self, errSchemaPath }, def, keyword) {
      /* istanbul ignore if */
      if (Array.isArray(def.keyword) ? !def.keyword.includes(keyword) : def.keyword !== keyword) {
          throw new Error("ajv implementation error");
      }
      const deps = def.dependencies;
      if (deps === null || deps === void 0 ? void 0 : deps.some((kwd) => !Object.prototype.hasOwnProperty.call(schema, kwd))) {
          throw new Error(`parent schema must have dependencies of ${keyword}: ${deps.join(",")}`);
      }
      if (def.validateSchema) {
          const valid = def.validateSchema(schema[keyword]);
          if (!valid) {
              const msg = `keyword "${keyword}" value is invalid at path "${errSchemaPath}": ` +
                  self.errorsText(def.validateSchema.errors);
              if (opts.validateSchema === "log")
                  self.logger.error(msg);
              else
                  throw new Error(msg);
          }
      }
  }
  keyword.validateKeywordUsage = validateKeywordUsage;

  var subschema = {};

  Object.defineProperty(subschema, "__esModule", { value: true });
  subschema.extendSubschemaMode = subschema.extendSubschemaData = subschema.getSubschema = void 0;
  const codegen_1$o = codegen;
  const util_1$n = util;
  function getSubschema(it, { keyword, schemaProp, schema, schemaPath, errSchemaPath, topSchemaRef }) {
      if (keyword !== undefined && schema !== undefined) {
          throw new Error('both "keyword" and "schema" passed, only one allowed');
      }
      if (keyword !== undefined) {
          const sch = it.schema[keyword];
          return schemaProp === undefined
              ? {
                  schema: sch,
                  schemaPath: (0, codegen_1$o._) `${it.schemaPath}${(0, codegen_1$o.getProperty)(keyword)}`,
                  errSchemaPath: `${it.errSchemaPath}/${keyword}`,
              }
              : {
                  schema: sch[schemaProp],
                  schemaPath: (0, codegen_1$o._) `${it.schemaPath}${(0, codegen_1$o.getProperty)(keyword)}${(0, codegen_1$o.getProperty)(schemaProp)}`,
                  errSchemaPath: `${it.errSchemaPath}/${keyword}/${(0, util_1$n.escapeFragment)(schemaProp)}`,
              };
      }
      if (schema !== undefined) {
          if (schemaPath === undefined || errSchemaPath === undefined || topSchemaRef === undefined) {
              throw new Error('"schemaPath", "errSchemaPath" and "topSchemaRef" are required with "schema"');
          }
          return {
              schema,
              schemaPath,
              topSchemaRef,
              errSchemaPath,
          };
      }
      throw new Error('either "keyword" or "schema" must be passed');
  }
  subschema.getSubschema = getSubschema;
  function extendSubschemaData(subschema, it, { dataProp, dataPropType: dpType, data, dataTypes, propertyName }) {
      if (data !== undefined && dataProp !== undefined) {
          throw new Error('both "data" and "dataProp" passed, only one allowed');
      }
      const { gen } = it;
      if (dataProp !== undefined) {
          const { errorPath, dataPathArr, opts } = it;
          const nextData = gen.let("data", (0, codegen_1$o._) `${it.data}${(0, codegen_1$o.getProperty)(dataProp)}`, true);
          dataContextProps(nextData);
          subschema.errorPath = (0, codegen_1$o.str) `${errorPath}${(0, util_1$n.getErrorPath)(dataProp, dpType, opts.jsPropertySyntax)}`;
          subschema.parentDataProperty = (0, codegen_1$o._) `${dataProp}`;
          subschema.dataPathArr = [...dataPathArr, subschema.parentDataProperty];
      }
      if (data !== undefined) {
          const nextData = data instanceof codegen_1$o.Name ? data : gen.let("data", data, true); // replaceable if used once?
          dataContextProps(nextData);
          if (propertyName !== undefined)
              subschema.propertyName = propertyName;
          // TODO something is possibly wrong here with not changing parentDataProperty and not appending dataPathArr
      }
      if (dataTypes)
          subschema.dataTypes = dataTypes;
      function dataContextProps(_nextData) {
          subschema.data = _nextData;
          subschema.dataLevel = it.dataLevel + 1;
          subschema.dataTypes = [];
          it.definedProperties = new Set();
          subschema.parentData = it.data;
          subschema.dataNames = [...it.dataNames, _nextData];
      }
  }
  subschema.extendSubschemaData = extendSubschemaData;
  function extendSubschemaMode(subschema, { jtdDiscriminator, jtdMetadata, compositeRule, createErrors, allErrors }) {
      if (compositeRule !== undefined)
          subschema.compositeRule = compositeRule;
      if (createErrors !== undefined)
          subschema.createErrors = createErrors;
      if (allErrors !== undefined)
          subschema.allErrors = allErrors;
      subschema.jtdDiscriminator = jtdDiscriminator; // not inherited
      subschema.jtdMetadata = jtdMetadata; // not inherited
  }
  subschema.extendSubschemaMode = extendSubschemaMode;

  var resolve$1 = {};

  // do not edit .js files directly - edit src/index.jst



  var fastDeepEqual = function equal(a, b) {
    if (a === b) return true;

    if (a && b && typeof a == 'object' && typeof b == 'object') {
      if (a.constructor !== b.constructor) return false;

      var length, i, keys;
      if (Array.isArray(a)) {
        length = a.length;
        if (length != b.length) return false;
        for (i = length; i-- !== 0;)
          if (!equal(a[i], b[i])) return false;
        return true;
      }



      if (a.constructor === RegExp) return a.source === b.source && a.flags === b.flags;
      if (a.valueOf !== Object.prototype.valueOf) return a.valueOf() === b.valueOf();
      if (a.toString !== Object.prototype.toString) return a.toString() === b.toString();

      keys = Object.keys(a);
      length = keys.length;
      if (length !== Object.keys(b).length) return false;

      for (i = length; i-- !== 0;)
        if (!Object.prototype.hasOwnProperty.call(b, keys[i])) return false;

      for (i = length; i-- !== 0;) {
        var key = keys[i];

        if (!equal(a[key], b[key])) return false;
      }

      return true;
    }

    // true if both NaN, false otherwise
    return a!==a && b!==b;
  };

  var jsonSchemaTraverse = {exports: {}};

  var traverse$1 = jsonSchemaTraverse.exports = function (schema, opts, cb) {
    // Legacy support for v0.3.1 and earlier.
    if (typeof opts == 'function') {
      cb = opts;
      opts = {};
    }

    cb = opts.cb || cb;
    var pre = (typeof cb == 'function') ? cb : cb.pre || function() {};
    var post = cb.post || function() {};

    _traverse(opts, pre, post, schema, '', schema);
  };


  traverse$1.keywords = {
    additionalItems: true,
    items: true,
    contains: true,
    additionalProperties: true,
    propertyNames: true,
    not: true,
    if: true,
    then: true,
    else: true
  };

  traverse$1.arrayKeywords = {
    items: true,
    allOf: true,
    anyOf: true,
    oneOf: true
  };

  traverse$1.propsKeywords = {
    $defs: true,
    definitions: true,
    properties: true,
    patternProperties: true,
    dependencies: true
  };

  traverse$1.skipKeywords = {
    default: true,
    enum: true,
    const: true,
    required: true,
    maximum: true,
    minimum: true,
    exclusiveMaximum: true,
    exclusiveMinimum: true,
    multipleOf: true,
    maxLength: true,
    minLength: true,
    pattern: true,
    format: true,
    maxItems: true,
    minItems: true,
    uniqueItems: true,
    maxProperties: true,
    minProperties: true
  };


  function _traverse(opts, pre, post, schema, jsonPtr, rootSchema, parentJsonPtr, parentKeyword, parentSchema, keyIndex) {
    if (schema && typeof schema == 'object' && !Array.isArray(schema)) {
      pre(schema, jsonPtr, rootSchema, parentJsonPtr, parentKeyword, parentSchema, keyIndex);
      for (var key in schema) {
        var sch = schema[key];
        if (Array.isArray(sch)) {
          if (key in traverse$1.arrayKeywords) {
            for (var i=0; i<sch.length; i++)
              _traverse(opts, pre, post, sch[i], jsonPtr + '/' + key + '/' + i, rootSchema, jsonPtr, key, schema, i);
          }
        } else if (key in traverse$1.propsKeywords) {
          if (sch && typeof sch == 'object') {
            for (var prop in sch)
              _traverse(opts, pre, post, sch[prop], jsonPtr + '/' + key + '/' + escapeJsonPtr(prop), rootSchema, jsonPtr, key, schema, prop);
          }
        } else if (key in traverse$1.keywords || (opts.allKeys && !(key in traverse$1.skipKeywords))) {
          _traverse(opts, pre, post, sch, jsonPtr + '/' + key, rootSchema, jsonPtr, key, schema);
        }
      }
      post(schema, jsonPtr, rootSchema, parentJsonPtr, parentKeyword, parentSchema, keyIndex);
    }
  }


  function escapeJsonPtr(str) {
    return str.replace(/~/g, '~0').replace(/\//g, '~1');
  }

  Object.defineProperty(resolve$1, "__esModule", { value: true });
  resolve$1.getSchemaRefs = resolve$1.resolveUrl = resolve$1.normalizeId = resolve$1._getFullPath = resolve$1.getFullPath = resolve$1.inlineRef = void 0;
  const util_1$m = util;
  const equal$2 = fastDeepEqual;
  const traverse = jsonSchemaTraverse.exports;
  // TODO refactor to use keyword definitions
  const SIMPLE_INLINED = new Set([
      "type",
      "format",
      "pattern",
      "maxLength",
      "minLength",
      "maxProperties",
      "minProperties",
      "maxItems",
      "minItems",
      "maximum",
      "minimum",
      "uniqueItems",
      "multipleOf",
      "required",
      "enum",
      "const",
  ]);
  function inlineRef(schema, limit = true) {
      if (typeof schema == "boolean")
          return true;
      if (limit === true)
          return !hasRef(schema);
      if (!limit)
          return false;
      return countKeys(schema) <= limit;
  }
  resolve$1.inlineRef = inlineRef;
  const REF_KEYWORDS = new Set([
      "$ref",
      "$recursiveRef",
      "$recursiveAnchor",
      "$dynamicRef",
      "$dynamicAnchor",
  ]);
  function hasRef(schema) {
      for (const key in schema) {
          if (REF_KEYWORDS.has(key))
              return true;
          const sch = schema[key];
          if (Array.isArray(sch) && sch.some(hasRef))
              return true;
          if (typeof sch == "object" && hasRef(sch))
              return true;
      }
      return false;
  }
  function countKeys(schema) {
      let count = 0;
      for (const key in schema) {
          if (key === "$ref")
              return Infinity;
          count++;
          if (SIMPLE_INLINED.has(key))
              continue;
          if (typeof schema[key] == "object") {
              (0, util_1$m.eachItem)(schema[key], (sch) => (count += countKeys(sch)));
          }
          if (count === Infinity)
              return Infinity;
      }
      return count;
  }
  function getFullPath(resolver, id = "", normalize) {
      if (normalize !== false)
          id = normalizeId(id);
      const p = resolver.parse(id);
      return _getFullPath(resolver, p);
  }
  resolve$1.getFullPath = getFullPath;
  function _getFullPath(resolver, p) {
      const serialized = resolver.serialize(p);
      return serialized.split("#")[0] + "#";
  }
  resolve$1._getFullPath = _getFullPath;
  const TRAILING_SLASH_HASH = /#\/?$/;
  function normalizeId(id) {
      return id ? id.replace(TRAILING_SLASH_HASH, "") : "";
  }
  resolve$1.normalizeId = normalizeId;
  function resolveUrl(resolver, baseId, id) {
      id = normalizeId(id);
      return resolver.resolve(baseId, id);
  }
  resolve$1.resolveUrl = resolveUrl;
  const ANCHOR = /^[a-z_][-a-z0-9._]*$/i;
  function getSchemaRefs(schema, baseId) {
      if (typeof schema == "boolean")
          return {};
      const { schemaId, uriResolver } = this.opts;
      const schId = normalizeId(schema[schemaId] || baseId);
      const baseIds = { "": schId };
      const pathPrefix = getFullPath(uriResolver, schId, false);
      const localRefs = {};
      const schemaRefs = new Set();
      traverse(schema, { allKeys: true }, (sch, jsonPtr, _, parentJsonPtr) => {
          if (parentJsonPtr === undefined)
              return;
          const fullPath = pathPrefix + jsonPtr;
          let baseId = baseIds[parentJsonPtr];
          if (typeof sch[schemaId] == "string")
              baseId = addRef.call(this, sch[schemaId]);
          addAnchor.call(this, sch.$anchor);
          addAnchor.call(this, sch.$dynamicAnchor);
          baseIds[jsonPtr] = baseId;
          function addRef(ref) {
              // eslint-disable-next-line @typescript-eslint/unbound-method
              const _resolve = this.opts.uriResolver.resolve;
              ref = normalizeId(baseId ? _resolve(baseId, ref) : ref);
              if (schemaRefs.has(ref))
                  throw ambiguos(ref);
              schemaRefs.add(ref);
              let schOrRef = this.refs[ref];
              if (typeof schOrRef == "string")
                  schOrRef = this.refs[schOrRef];
              if (typeof schOrRef == "object") {
                  checkAmbiguosRef(sch, schOrRef.schema, ref);
              }
              else if (ref !== normalizeId(fullPath)) {
                  if (ref[0] === "#") {
                      checkAmbiguosRef(sch, localRefs[ref], ref);
                      localRefs[ref] = sch;
                  }
                  else {
                      this.refs[ref] = fullPath;
                  }
              }
              return ref;
          }
          function addAnchor(anchor) {
              if (typeof anchor == "string") {
                  if (!ANCHOR.test(anchor))
                      throw new Error(`invalid anchor "${anchor}"`);
                  addRef.call(this, `#${anchor}`);
              }
          }
      });
      return localRefs;
      function checkAmbiguosRef(sch1, sch2, ref) {
          if (sch2 !== undefined && !equal$2(sch1, sch2))
              throw ambiguos(ref);
      }
      function ambiguos(ref) {
          return new Error(`reference "${ref}" resolves to more than one schema`);
      }
  }
  resolve$1.getSchemaRefs = getSchemaRefs;

  Object.defineProperty(validate$1, "__esModule", { value: true });
  validate$1.getData = validate$1.KeywordCxt = validate$1.validateFunctionCode = void 0;
  const boolSchema_1 = boolSchema;
  const dataType_1$1 = dataType;
  const applicability_1 = applicability;
  const dataType_2 = dataType;
  const defaults_1 = defaults;
  const keyword_1 = keyword;
  const subschema_1 = subschema;
  const codegen_1$n = codegen;
  const names_1$3 = names$1;
  const resolve_1$2 = resolve$1;
  const util_1$l = util;
  const errors_1 = errors;
  // schema compilation - generates validation function, subschemaCode (below) is used for subschemas
  function validateFunctionCode(it) {
      if (isSchemaObj(it)) {
          checkKeywords(it);
          if (schemaCxtHasRules(it)) {
              topSchemaObjCode(it);
              return;
          }
      }
      validateFunction(it, () => (0, boolSchema_1.topBoolOrEmptySchema)(it));
  }
  validate$1.validateFunctionCode = validateFunctionCode;
  function validateFunction({ gen, validateName, schema, schemaEnv, opts }, body) {
      if (opts.code.es5) {
          gen.func(validateName, (0, codegen_1$n._) `${names_1$3.default.data}, ${names_1$3.default.valCxt}`, schemaEnv.$async, () => {
              gen.code((0, codegen_1$n._) `"use strict"; ${funcSourceUrl(schema, opts)}`);
              destructureValCxtES5(gen, opts);
              gen.code(body);
          });
      }
      else {
          gen.func(validateName, (0, codegen_1$n._) `${names_1$3.default.data}, ${destructureValCxt(opts)}`, schemaEnv.$async, () => gen.code(funcSourceUrl(schema, opts)).code(body));
      }
  }
  function destructureValCxt(opts) {
      return (0, codegen_1$n._) `{${names_1$3.default.instancePath}="", ${names_1$3.default.parentData}, ${names_1$3.default.parentDataProperty}, ${names_1$3.default.rootData}=${names_1$3.default.data}${opts.dynamicRef ? (0, codegen_1$n._) `, ${names_1$3.default.dynamicAnchors}={}` : codegen_1$n.nil}}={}`;
  }
  function destructureValCxtES5(gen, opts) {
      gen.if(names_1$3.default.valCxt, () => {
          gen.var(names_1$3.default.instancePath, (0, codegen_1$n._) `${names_1$3.default.valCxt}.${names_1$3.default.instancePath}`);
          gen.var(names_1$3.default.parentData, (0, codegen_1$n._) `${names_1$3.default.valCxt}.${names_1$3.default.parentData}`);
          gen.var(names_1$3.default.parentDataProperty, (0, codegen_1$n._) `${names_1$3.default.valCxt}.${names_1$3.default.parentDataProperty}`);
          gen.var(names_1$3.default.rootData, (0, codegen_1$n._) `${names_1$3.default.valCxt}.${names_1$3.default.rootData}`);
          if (opts.dynamicRef)
              gen.var(names_1$3.default.dynamicAnchors, (0, codegen_1$n._) `${names_1$3.default.valCxt}.${names_1$3.default.dynamicAnchors}`);
      }, () => {
          gen.var(names_1$3.default.instancePath, (0, codegen_1$n._) `""`);
          gen.var(names_1$3.default.parentData, (0, codegen_1$n._) `undefined`);
          gen.var(names_1$3.default.parentDataProperty, (0, codegen_1$n._) `undefined`);
          gen.var(names_1$3.default.rootData, names_1$3.default.data);
          if (opts.dynamicRef)
              gen.var(names_1$3.default.dynamicAnchors, (0, codegen_1$n._) `{}`);
      });
  }
  function topSchemaObjCode(it) {
      const { schema, opts, gen } = it;
      validateFunction(it, () => {
          if (opts.$comment && schema.$comment)
              commentKeyword(it);
          checkNoDefault(it);
          gen.let(names_1$3.default.vErrors, null);
          gen.let(names_1$3.default.errors, 0);
          if (opts.unevaluated)
              resetEvaluated(it);
          typeAndKeywords(it);
          returnResults(it);
      });
      return;
  }
  function resetEvaluated(it) {
      // TODO maybe some hook to execute it in the end to check whether props/items are Name, as in assignEvaluated
      const { gen, validateName } = it;
      it.evaluated = gen.const("evaluated", (0, codegen_1$n._) `${validateName}.evaluated`);
      gen.if((0, codegen_1$n._) `${it.evaluated}.dynamicProps`, () => gen.assign((0, codegen_1$n._) `${it.evaluated}.props`, (0, codegen_1$n._) `undefined`));
      gen.if((0, codegen_1$n._) `${it.evaluated}.dynamicItems`, () => gen.assign((0, codegen_1$n._) `${it.evaluated}.items`, (0, codegen_1$n._) `undefined`));
  }
  function funcSourceUrl(schema, opts) {
      const schId = typeof schema == "object" && schema[opts.schemaId];
      return schId && (opts.code.source || opts.code.process) ? (0, codegen_1$n._) `/*# sourceURL=${schId} */` : codegen_1$n.nil;
  }
  // schema compilation - this function is used recursively to generate code for sub-schemas
  function subschemaCode(it, valid) {
      if (isSchemaObj(it)) {
          checkKeywords(it);
          if (schemaCxtHasRules(it)) {
              subSchemaObjCode(it, valid);
              return;
          }
      }
      (0, boolSchema_1.boolOrEmptySchema)(it, valid);
  }
  function schemaCxtHasRules({ schema, self }) {
      if (typeof schema == "boolean")
          return !schema;
      for (const key in schema)
          if (self.RULES.all[key])
              return true;
      return false;
  }
  function isSchemaObj(it) {
      return typeof it.schema != "boolean";
  }
  function subSchemaObjCode(it, valid) {
      const { schema, gen, opts } = it;
      if (opts.$comment && schema.$comment)
          commentKeyword(it);
      updateContext(it);
      checkAsyncSchema(it);
      const errsCount = gen.const("_errs", names_1$3.default.errors);
      typeAndKeywords(it, errsCount);
      // TODO var
      gen.var(valid, (0, codegen_1$n._) `${errsCount} === ${names_1$3.default.errors}`);
  }
  function checkKeywords(it) {
      (0, util_1$l.checkUnknownRules)(it);
      checkRefsAndKeywords(it);
  }
  function typeAndKeywords(it, errsCount) {
      if (it.opts.jtd)
          return schemaKeywords(it, [], false, errsCount);
      const types = (0, dataType_1$1.getSchemaTypes)(it.schema);
      const checkedTypes = (0, dataType_1$1.coerceAndCheckDataType)(it, types);
      schemaKeywords(it, types, !checkedTypes, errsCount);
  }
  function checkRefsAndKeywords(it) {
      const { schema, errSchemaPath, opts, self } = it;
      if (schema.$ref && opts.ignoreKeywordsWithRef && (0, util_1$l.schemaHasRulesButRef)(schema, self.RULES)) {
          self.logger.warn(`$ref: keywords ignored in schema at path "${errSchemaPath}"`);
      }
  }
  function checkNoDefault(it) {
      const { schema, opts } = it;
      if (schema.default !== undefined && opts.useDefaults && opts.strictSchema) {
          (0, util_1$l.checkStrictMode)(it, "default is ignored in the schema root");
      }
  }
  function updateContext(it) {
      const schId = it.schema[it.opts.schemaId];
      if (schId)
          it.baseId = (0, resolve_1$2.resolveUrl)(it.opts.uriResolver, it.baseId, schId);
  }
  function checkAsyncSchema(it) {
      if (it.schema.$async && !it.schemaEnv.$async)
          throw new Error("async schema in sync schema");
  }
  function commentKeyword({ gen, schemaEnv, schema, errSchemaPath, opts }) {
      const msg = schema.$comment;
      if (opts.$comment === true) {
          gen.code((0, codegen_1$n._) `${names_1$3.default.self}.logger.log(${msg})`);
      }
      else if (typeof opts.$comment == "function") {
          const schemaPath = (0, codegen_1$n.str) `${errSchemaPath}/$comment`;
          const rootName = gen.scopeValue("root", { ref: schemaEnv.root });
          gen.code((0, codegen_1$n._) `${names_1$3.default.self}.opts.$comment(${msg}, ${schemaPath}, ${rootName}.schema)`);
      }
  }
  function returnResults(it) {
      const { gen, schemaEnv, validateName, ValidationError, opts } = it;
      if (schemaEnv.$async) {
          // TODO assign unevaluated
          gen.if((0, codegen_1$n._) `${names_1$3.default.errors} === 0`, () => gen.return(names_1$3.default.data), () => gen.throw((0, codegen_1$n._) `new ${ValidationError}(${names_1$3.default.vErrors})`));
      }
      else {
          gen.assign((0, codegen_1$n._) `${validateName}.errors`, names_1$3.default.vErrors);
          if (opts.unevaluated)
              assignEvaluated(it);
          gen.return((0, codegen_1$n._) `${names_1$3.default.errors} === 0`);
      }
  }
  function assignEvaluated({ gen, evaluated, props, items }) {
      if (props instanceof codegen_1$n.Name)
          gen.assign((0, codegen_1$n._) `${evaluated}.props`, props);
      if (items instanceof codegen_1$n.Name)
          gen.assign((0, codegen_1$n._) `${evaluated}.items`, items);
  }
  function schemaKeywords(it, types, typeErrors, errsCount) {
      const { gen, schema, data, allErrors, opts, self } = it;
      const { RULES } = self;
      if (schema.$ref && (opts.ignoreKeywordsWithRef || !(0, util_1$l.schemaHasRulesButRef)(schema, RULES))) {
          gen.block(() => keywordCode(it, "$ref", RULES.all.$ref.definition)); // TODO typecast
          return;
      }
      if (!opts.jtd)
          checkStrictTypes(it, types);
      gen.block(() => {
          for (const group of RULES.rules)
              groupKeywords(group);
          groupKeywords(RULES.post);
      });
      function groupKeywords(group) {
          if (!(0, applicability_1.shouldUseGroup)(schema, group))
              return;
          if (group.type) {
              gen.if((0, dataType_2.checkDataType)(group.type, data, opts.strictNumbers));
              iterateKeywords(it, group);
              if (types.length === 1 && types[0] === group.type && typeErrors) {
                  gen.else();
                  (0, dataType_2.reportTypeError)(it);
              }
              gen.endIf();
          }
          else {
              iterateKeywords(it, group);
          }
          // TODO make it "ok" call?
          if (!allErrors)
              gen.if((0, codegen_1$n._) `${names_1$3.default.errors} === ${errsCount || 0}`);
      }
  }
  function iterateKeywords(it, group) {
      const { gen, schema, opts: { useDefaults }, } = it;
      if (useDefaults)
          (0, defaults_1.assignDefaults)(it, group.type);
      gen.block(() => {
          for (const rule of group.rules) {
              if ((0, applicability_1.shouldUseRule)(schema, rule)) {
                  keywordCode(it, rule.keyword, rule.definition, group.type);
              }
          }
      });
  }
  function checkStrictTypes(it, types) {
      if (it.schemaEnv.meta || !it.opts.strictTypes)
          return;
      checkContextTypes(it, types);
      if (!it.opts.allowUnionTypes)
          checkMultipleTypes(it, types);
      checkKeywordTypes(it, it.dataTypes);
  }
  function checkContextTypes(it, types) {
      if (!types.length)
          return;
      if (!it.dataTypes.length) {
          it.dataTypes = types;
          return;
      }
      types.forEach((t) => {
          if (!includesType(it.dataTypes, t)) {
              strictTypesError(it, `type "${t}" not allowed by context "${it.dataTypes.join(",")}"`);
          }
      });
      it.dataTypes = it.dataTypes.filter((t) => includesType(types, t));
  }
  function checkMultipleTypes(it, ts) {
      if (ts.length > 1 && !(ts.length === 2 && ts.includes("null"))) {
          strictTypesError(it, "use allowUnionTypes to allow union type keyword");
      }
  }
  function checkKeywordTypes(it, ts) {
      const rules = it.self.RULES.all;
      for (const keyword in rules) {
          const rule = rules[keyword];
          if (typeof rule == "object" && (0, applicability_1.shouldUseRule)(it.schema, rule)) {
              const { type } = rule.definition;
              if (type.length && !type.some((t) => hasApplicableType(ts, t))) {
                  strictTypesError(it, `missing type "${type.join(",")}" for keyword "${keyword}"`);
              }
          }
      }
  }
  function hasApplicableType(schTs, kwdT) {
      return schTs.includes(kwdT) || (kwdT === "number" && schTs.includes("integer"));
  }
  function includesType(ts, t) {
      return ts.includes(t) || (t === "integer" && ts.includes("number"));
  }
  function strictTypesError(it, msg) {
      const schemaPath = it.schemaEnv.baseId + it.errSchemaPath;
      msg += ` at "${schemaPath}" (strictTypes)`;
      (0, util_1$l.checkStrictMode)(it, msg, it.opts.strictTypes);
  }
  class KeywordCxt {
      constructor(it, def, keyword) {
          (0, keyword_1.validateKeywordUsage)(it, def, keyword);
          this.gen = it.gen;
          this.allErrors = it.allErrors;
          this.keyword = keyword;
          this.data = it.data;
          this.schema = it.schema[keyword];
          this.$data = def.$data && it.opts.$data && this.schema && this.schema.$data;
          this.schemaValue = (0, util_1$l.schemaRefOrVal)(it, this.schema, keyword, this.$data);
          this.schemaType = def.schemaType;
          this.parentSchema = it.schema;
          this.params = {};
          this.it = it;
          this.def = def;
          if (this.$data) {
              this.schemaCode = it.gen.const("vSchema", getData(this.$data, it));
          }
          else {
              this.schemaCode = this.schemaValue;
              if (!(0, keyword_1.validSchemaType)(this.schema, def.schemaType, def.allowUndefined)) {
                  throw new Error(`${keyword} value must be ${JSON.stringify(def.schemaType)}`);
              }
          }
          if ("code" in def ? def.trackErrors : def.errors !== false) {
              this.errsCount = it.gen.const("_errs", names_1$3.default.errors);
          }
      }
      result(condition, successAction, failAction) {
          this.failResult((0, codegen_1$n.not)(condition), successAction, failAction);
      }
      failResult(condition, successAction, failAction) {
          this.gen.if(condition);
          if (failAction)
              failAction();
          else
              this.error();
          if (successAction) {
              this.gen.else();
              successAction();
              if (this.allErrors)
                  this.gen.endIf();
          }
          else {
              if (this.allErrors)
                  this.gen.endIf();
              else
                  this.gen.else();
          }
      }
      pass(condition, failAction) {
          this.failResult((0, codegen_1$n.not)(condition), undefined, failAction);
      }
      fail(condition) {
          if (condition === undefined) {
              this.error();
              if (!this.allErrors)
                  this.gen.if(false); // this branch will be removed by gen.optimize
              return;
          }
          this.gen.if(condition);
          this.error();
          if (this.allErrors)
              this.gen.endIf();
          else
              this.gen.else();
      }
      fail$data(condition) {
          if (!this.$data)
              return this.fail(condition);
          const { schemaCode } = this;
          this.fail((0, codegen_1$n._) `${schemaCode} !== undefined && (${(0, codegen_1$n.or)(this.invalid$data(), condition)})`);
      }
      error(append, errorParams, errorPaths) {
          if (errorParams) {
              this.setParams(errorParams);
              this._error(append, errorPaths);
              this.setParams({});
              return;
          }
          this._error(append, errorPaths);
      }
      _error(append, errorPaths) {
          (append ? errors_1.reportExtraError : errors_1.reportError)(this, this.def.error, errorPaths);
      }
      $dataError() {
          (0, errors_1.reportError)(this, this.def.$dataError || errors_1.keyword$DataError);
      }
      reset() {
          if (this.errsCount === undefined)
              throw new Error('add "trackErrors" to keyword definition');
          (0, errors_1.resetErrorsCount)(this.gen, this.errsCount);
      }
      ok(cond) {
          if (!this.allErrors)
              this.gen.if(cond);
      }
      setParams(obj, assign) {
          if (assign)
              Object.assign(this.params, obj);
          else
              this.params = obj;
      }
      block$data(valid, codeBlock, $dataValid = codegen_1$n.nil) {
          this.gen.block(() => {
              this.check$data(valid, $dataValid);
              codeBlock();
          });
      }
      check$data(valid = codegen_1$n.nil, $dataValid = codegen_1$n.nil) {
          if (!this.$data)
              return;
          const { gen, schemaCode, schemaType, def } = this;
          gen.if((0, codegen_1$n.or)((0, codegen_1$n._) `${schemaCode} === undefined`, $dataValid));
          if (valid !== codegen_1$n.nil)
              gen.assign(valid, true);
          if (schemaType.length || def.validateSchema) {
              gen.elseIf(this.invalid$data());
              this.$dataError();
              if (valid !== codegen_1$n.nil)
                  gen.assign(valid, false);
          }
          gen.else();
      }
      invalid$data() {
          const { gen, schemaCode, schemaType, def, it } = this;
          return (0, codegen_1$n.or)(wrong$DataType(), invalid$DataSchema());
          function wrong$DataType() {
              if (schemaType.length) {
                  /* istanbul ignore if */
                  if (!(schemaCode instanceof codegen_1$n.Name))
                      throw new Error("ajv implementation error");
                  const st = Array.isArray(schemaType) ? schemaType : [schemaType];
                  return (0, codegen_1$n._) `${(0, dataType_2.checkDataTypes)(st, schemaCode, it.opts.strictNumbers, dataType_2.DataType.Wrong)}`;
              }
              return codegen_1$n.nil;
          }
          function invalid$DataSchema() {
              if (def.validateSchema) {
                  const validateSchemaRef = gen.scopeValue("validate$data", { ref: def.validateSchema }); // TODO value.code for standalone
                  return (0, codegen_1$n._) `!${validateSchemaRef}(${schemaCode})`;
              }
              return codegen_1$n.nil;
          }
      }
      subschema(appl, valid) {
          const subschema = (0, subschema_1.getSubschema)(this.it, appl);
          (0, subschema_1.extendSubschemaData)(subschema, this.it, appl);
          (0, subschema_1.extendSubschemaMode)(subschema, appl);
          const nextContext = { ...this.it, ...subschema, items: undefined, props: undefined };
          subschemaCode(nextContext, valid);
          return nextContext;
      }
      mergeEvaluated(schemaCxt, toName) {
          const { it, gen } = this;
          if (!it.opts.unevaluated)
              return;
          if (it.props !== true && schemaCxt.props !== undefined) {
              it.props = util_1$l.mergeEvaluated.props(gen, schemaCxt.props, it.props, toName);
          }
          if (it.items !== true && schemaCxt.items !== undefined) {
              it.items = util_1$l.mergeEvaluated.items(gen, schemaCxt.items, it.items, toName);
          }
      }
      mergeValidEvaluated(schemaCxt, valid) {
          const { it, gen } = this;
          if (it.opts.unevaluated && (it.props !== true || it.items !== true)) {
              gen.if(valid, () => this.mergeEvaluated(schemaCxt, codegen_1$n.Name));
              return true;
          }
      }
  }
  validate$1.KeywordCxt = KeywordCxt;
  function keywordCode(it, keyword, def, ruleType) {
      const cxt = new KeywordCxt(it, def, keyword);
      if ("code" in def) {
          def.code(cxt, ruleType);
      }
      else if (cxt.$data && def.validate) {
          (0, keyword_1.funcKeywordCode)(cxt, def);
      }
      else if ("macro" in def) {
          (0, keyword_1.macroKeywordCode)(cxt, def);
      }
      else if (def.compile || def.validate) {
          (0, keyword_1.funcKeywordCode)(cxt, def);
      }
  }
  const JSON_POINTER = /^\/(?:[^~]|~0|~1)*$/;
  const RELATIVE_JSON_POINTER = /^([0-9]+)(#|\/(?:[^~]|~0|~1)*)?$/;
  function getData($data, { dataLevel, dataNames, dataPathArr }) {
      let jsonPointer;
      let data;
      if ($data === "")
          return names_1$3.default.rootData;
      if ($data[0] === "/") {
          if (!JSON_POINTER.test($data))
              throw new Error(`Invalid JSON-pointer: ${$data}`);
          jsonPointer = $data;
          data = names_1$3.default.rootData;
      }
      else {
          const matches = RELATIVE_JSON_POINTER.exec($data);
          if (!matches)
              throw new Error(`Invalid JSON-pointer: ${$data}`);
          const up = +matches[1];
          jsonPointer = matches[2];
          if (jsonPointer === "#") {
              if (up >= dataLevel)
                  throw new Error(errorMsg("property/index", up));
              return dataPathArr[dataLevel - up];
          }
          if (up > dataLevel)
              throw new Error(errorMsg("data", up));
          data = dataNames[dataLevel - up];
          if (!jsonPointer)
              return data;
      }
      let expr = data;
      const segments = jsonPointer.split("/");
      for (const segment of segments) {
          if (segment) {
              data = (0, codegen_1$n._) `${data}${(0, codegen_1$n.getProperty)((0, util_1$l.unescapeJsonPointer)(segment))}`;
              expr = (0, codegen_1$n._) `${expr} && ${data}`;
          }
      }
      return expr;
      function errorMsg(pointerType, up) {
          return `Cannot access ${pointerType} ${up} levels up, current level is ${dataLevel}`;
      }
  }
  validate$1.getData = getData;

  var validation_error = {};

  Object.defineProperty(validation_error, "__esModule", { value: true });
  class ValidationError extends Error {
      constructor(errors) {
          super("validation failed");
          this.errors = errors;
          this.ajv = this.validation = true;
      }
  }
  validation_error.default = ValidationError;

  var ref_error = {};

  Object.defineProperty(ref_error, "__esModule", { value: true });
  const resolve_1$1 = resolve$1;
  class MissingRefError extends Error {
      constructor(resolver, baseId, ref, msg) {
          super(msg || `can't resolve reference ${ref} from id ${baseId}`);
          this.missingRef = (0, resolve_1$1.resolveUrl)(resolver, baseId, ref);
          this.missingSchema = (0, resolve_1$1.normalizeId)((0, resolve_1$1.getFullPath)(resolver, this.missingRef));
      }
  }
  ref_error.default = MissingRefError;

  var compile = {};

  Object.defineProperty(compile, "__esModule", { value: true });
  compile.resolveSchema = compile.getCompilingSchema = compile.resolveRef = compile.compileSchema = compile.SchemaEnv = void 0;
  const codegen_1$m = codegen;
  const validation_error_1 = validation_error;
  const names_1$2 = names$1;
  const resolve_1 = resolve$1;
  const util_1$k = util;
  const validate_1$1 = validate$1;
  class SchemaEnv {
      constructor(env) {
          var _a;
          this.refs = {};
          this.dynamicAnchors = {};
          let schema;
          if (typeof env.schema == "object")
              schema = env.schema;
          this.schema = env.schema;
          this.schemaId = env.schemaId;
          this.root = env.root || this;
          this.baseId = (_a = env.baseId) !== null && _a !== void 0 ? _a : (0, resolve_1.normalizeId)(schema === null || schema === void 0 ? void 0 : schema[env.schemaId || "$id"]);
          this.schemaPath = env.schemaPath;
          this.localRefs = env.localRefs;
          this.meta = env.meta;
          this.$async = schema === null || schema === void 0 ? void 0 : schema.$async;
          this.refs = {};
      }
  }
  compile.SchemaEnv = SchemaEnv;
  // let codeSize = 0
  // let nodeCount = 0
  // Compiles schema in SchemaEnv
  function compileSchema(sch) {
      // TODO refactor - remove compilations
      const _sch = getCompilingSchema.call(this, sch);
      if (_sch)
          return _sch;
      const rootId = (0, resolve_1.getFullPath)(this.opts.uriResolver, sch.root.baseId); // TODO if getFullPath removed 1 tests fails
      const { es5, lines } = this.opts.code;
      const { ownProperties } = this.opts;
      const gen = new codegen_1$m.CodeGen(this.scope, { es5, lines, ownProperties });
      let _ValidationError;
      if (sch.$async) {
          _ValidationError = gen.scopeValue("Error", {
              ref: validation_error_1.default,
              code: (0, codegen_1$m._) `require("ajv/dist/runtime/validation_error").default`,
          });
      }
      const validateName = gen.scopeName("validate");
      sch.validateName = validateName;
      const schemaCxt = {
          gen,
          allErrors: this.opts.allErrors,
          data: names_1$2.default.data,
          parentData: names_1$2.default.parentData,
          parentDataProperty: names_1$2.default.parentDataProperty,
          dataNames: [names_1$2.default.data],
          dataPathArr: [codegen_1$m.nil],
          dataLevel: 0,
          dataTypes: [],
          definedProperties: new Set(),
          topSchemaRef: gen.scopeValue("schema", this.opts.code.source === true
              ? { ref: sch.schema, code: (0, codegen_1$m.stringify)(sch.schema) }
              : { ref: sch.schema }),
          validateName,
          ValidationError: _ValidationError,
          schema: sch.schema,
          schemaEnv: sch,
          rootId,
          baseId: sch.baseId || rootId,
          schemaPath: codegen_1$m.nil,
          errSchemaPath: sch.schemaPath || (this.opts.jtd ? "" : "#"),
          errorPath: (0, codegen_1$m._) `""`,
          opts: this.opts,
          self: this,
      };
      let sourceCode;
      try {
          this._compilations.add(sch);
          (0, validate_1$1.validateFunctionCode)(schemaCxt);
          gen.optimize(this.opts.code.optimize);
          // gen.optimize(1)
          const validateCode = gen.toString();
          sourceCode = `${gen.scopeRefs(names_1$2.default.scope)}return ${validateCode}`;
          // console.log((codeSize += sourceCode.length), (nodeCount += gen.nodeCount))
          if (this.opts.code.process)
              sourceCode = this.opts.code.process(sourceCode, sch);
          // console.log("\n\n\n *** \n", sourceCode)
          const makeValidate = new Function(`${names_1$2.default.self}`, `${names_1$2.default.scope}`, sourceCode);
          const validate = makeValidate(this, this.scope.get());
          this.scope.value(validateName, { ref: validate });
          validate.errors = null;
          validate.schema = sch.schema;
          validate.schemaEnv = sch;
          if (sch.$async)
              validate.$async = true;
          if (this.opts.code.source === true) {
              validate.source = { validateName, validateCode, scopeValues: gen._values };
          }
          if (this.opts.unevaluated) {
              const { props, items } = schemaCxt;
              validate.evaluated = {
                  props: props instanceof codegen_1$m.Name ? undefined : props,
                  items: items instanceof codegen_1$m.Name ? undefined : items,
                  dynamicProps: props instanceof codegen_1$m.Name,
                  dynamicItems: items instanceof codegen_1$m.Name,
              };
              if (validate.source)
                  validate.source.evaluated = (0, codegen_1$m.stringify)(validate.evaluated);
          }
          sch.validate = validate;
          return sch;
      }
      catch (e) {
          delete sch.validate;
          delete sch.validateName;
          if (sourceCode)
              this.logger.error("Error compiling schema, function code:", sourceCode);
          // console.log("\n\n\n *** \n", sourceCode, this.opts)
          throw e;
      }
      finally {
          this._compilations.delete(sch);
      }
  }
  compile.compileSchema = compileSchema;
  function resolveRef(root, baseId, ref) {
      var _a;
      ref = (0, resolve_1.resolveUrl)(this.opts.uriResolver, baseId, ref);
      const schOrFunc = root.refs[ref];
      if (schOrFunc)
          return schOrFunc;
      let _sch = resolve.call(this, root, ref);
      if (_sch === undefined) {
          const schema = (_a = root.localRefs) === null || _a === void 0 ? void 0 : _a[ref]; // TODO maybe localRefs should hold SchemaEnv
          const { schemaId } = this.opts;
          if (schema)
              _sch = new SchemaEnv({ schema, schemaId, root, baseId });
      }
      if (_sch === undefined)
          return;
      return (root.refs[ref] = inlineOrCompile.call(this, _sch));
  }
  compile.resolveRef = resolveRef;
  function inlineOrCompile(sch) {
      if ((0, resolve_1.inlineRef)(sch.schema, this.opts.inlineRefs))
          return sch.schema;
      return sch.validate ? sch : compileSchema.call(this, sch);
  }
  // Index of schema compilation in the currently compiled list
  function getCompilingSchema(schEnv) {
      for (const sch of this._compilations) {
          if (sameSchemaEnv(sch, schEnv))
              return sch;
      }
  }
  compile.getCompilingSchema = getCompilingSchema;
  function sameSchemaEnv(s1, s2) {
      return s1.schema === s2.schema && s1.root === s2.root && s1.baseId === s2.baseId;
  }
  // resolve and compile the references ($ref)
  // TODO returns AnySchemaObject (if the schema can be inlined) or validation function
  function resolve(root, // information about the root schema for the current schema
  ref // reference to resolve
  ) {
      let sch;
      while (typeof (sch = this.refs[ref]) == "string")
          ref = sch;
      return sch || this.schemas[ref] || resolveSchema.call(this, root, ref);
  }
  // Resolve schema, its root and baseId
  function resolveSchema(root, // root object with properties schema, refs TODO below SchemaEnv is assigned to it
  ref // reference to resolve
  ) {
      const p = this.opts.uriResolver.parse(ref);
      const refPath = (0, resolve_1._getFullPath)(this.opts.uriResolver, p);
      let baseId = (0, resolve_1.getFullPath)(this.opts.uriResolver, root.baseId, undefined);
      // TODO `Object.keys(root.schema).length > 0` should not be needed - but removing breaks 2 tests
      if (Object.keys(root.schema).length > 0 && refPath === baseId) {
          return getJsonPointer.call(this, p, root);
      }
      const id = (0, resolve_1.normalizeId)(refPath);
      const schOrRef = this.refs[id] || this.schemas[id];
      if (typeof schOrRef == "string") {
          const sch = resolveSchema.call(this, root, schOrRef);
          if (typeof (sch === null || sch === void 0 ? void 0 : sch.schema) !== "object")
              return;
          return getJsonPointer.call(this, p, sch);
      }
      if (typeof (schOrRef === null || schOrRef === void 0 ? void 0 : schOrRef.schema) !== "object")
          return;
      if (!schOrRef.validate)
          compileSchema.call(this, schOrRef);
      if (id === (0, resolve_1.normalizeId)(ref)) {
          const { schema } = schOrRef;
          const { schemaId } = this.opts;
          const schId = schema[schemaId];
          if (schId)
              baseId = (0, resolve_1.resolveUrl)(this.opts.uriResolver, baseId, schId);
          return new SchemaEnv({ schema, schemaId, root, baseId });
      }
      return getJsonPointer.call(this, p, schOrRef);
  }
  compile.resolveSchema = resolveSchema;
  const PREVENT_SCOPE_CHANGE = new Set([
      "properties",
      "patternProperties",
      "enum",
      "dependencies",
      "definitions",
  ]);
  function getJsonPointer(parsedRef, { baseId, schema, root }) {
      var _a;
      if (((_a = parsedRef.fragment) === null || _a === void 0 ? void 0 : _a[0]) !== "/")
          return;
      for (const part of parsedRef.fragment.slice(1).split("/")) {
          if (typeof schema === "boolean")
              return;
          const partSchema = schema[(0, util_1$k.unescapeFragment)(part)];
          if (partSchema === undefined)
              return;
          schema = partSchema;
          // TODO PREVENT_SCOPE_CHANGE could be defined in keyword def?
          const schId = typeof schema === "object" && schema[this.opts.schemaId];
          if (!PREVENT_SCOPE_CHANGE.has(part) && schId) {
              baseId = (0, resolve_1.resolveUrl)(this.opts.uriResolver, baseId, schId);
          }
      }
      let env;
      if (typeof schema != "boolean" && schema.$ref && !(0, util_1$k.schemaHasRulesButRef)(schema, this.RULES)) {
          const $ref = (0, resolve_1.resolveUrl)(this.opts.uriResolver, baseId, schema.$ref);
          env = resolveSchema.call(this, root, $ref);
      }
      // even though resolution failed we need to return SchemaEnv to throw exception
      // so that compileAsync loads missing schema.
      const { schemaId } = this.opts;
      env = env || new SchemaEnv({ schema, schemaId, root, baseId });
      if (env.schema !== env.root.schema)
          return env;
      return undefined;
  }

  var $id$1 = "https://raw.githubusercontent.com/ajv-validator/ajv/master/lib/refs/data.json#";
  var description = "Meta-schema for $data reference (JSON AnySchema extension proposal)";
  var type$1 = "object";
  var required$1 = [
  	"$data"
  ];
  var properties$2 = {
  	$data: {
  		type: "string",
  		anyOf: [
  			{
  				format: "relative-json-pointer"
  			},
  			{
  				format: "json-pointer"
  			}
  		]
  	}
  };
  var additionalProperties$1 = false;
  var require$$9 = {
  	$id: $id$1,
  	description: description,
  	type: type$1,
  	required: required$1,
  	properties: properties$2,
  	additionalProperties: additionalProperties$1
  };

  var uri$1 = {};

  var uri_all = {exports: {}};

  /** @license URI.js v4.4.1 (c) 2011 Gary Court. License: http://github.com/garycourt/uri-js */

  (function (module, exports) {
  (function (global, factory) {
  	factory(exports) ;
  }(commonjsGlobal, (function (exports) {
  function merge() {
      for (var _len = arguments.length, sets = Array(_len), _key = 0; _key < _len; _key++) {
          sets[_key] = arguments[_key];
      }

      if (sets.length > 1) {
          sets[0] = sets[0].slice(0, -1);
          var xl = sets.length - 1;
          for (var x = 1; x < xl; ++x) {
              sets[x] = sets[x].slice(1, -1);
          }
          sets[xl] = sets[xl].slice(1);
          return sets.join('');
      } else {
          return sets[0];
      }
  }
  function subexp(str) {
      return "(?:" + str + ")";
  }
  function typeOf(o) {
      return o === undefined ? "undefined" : o === null ? "null" : Object.prototype.toString.call(o).split(" ").pop().split("]").shift().toLowerCase();
  }
  function toUpperCase(str) {
      return str.toUpperCase();
  }
  function toArray(obj) {
      return obj !== undefined && obj !== null ? obj instanceof Array ? obj : typeof obj.length !== "number" || obj.split || obj.setInterval || obj.call ? [obj] : Array.prototype.slice.call(obj) : [];
  }
  function assign(target, source) {
      var obj = target;
      if (source) {
          for (var key in source) {
              obj[key] = source[key];
          }
      }
      return obj;
  }

  function buildExps(isIRI) {
      var ALPHA$$ = "[A-Za-z]",
          DIGIT$$ = "[0-9]",
          HEXDIG$$ = merge(DIGIT$$, "[A-Fa-f]"),
          PCT_ENCODED$ = subexp(subexp("%[EFef]" + HEXDIG$$ + "%" + HEXDIG$$ + HEXDIG$$ + "%" + HEXDIG$$ + HEXDIG$$) + "|" + subexp("%[89A-Fa-f]" + HEXDIG$$ + "%" + HEXDIG$$ + HEXDIG$$) + "|" + subexp("%" + HEXDIG$$ + HEXDIG$$)),
          //expanded
      GEN_DELIMS$$ = "[\\:\\/\\?\\#\\[\\]\\@]",
          SUB_DELIMS$$ = "[\\!\\$\\&\\'\\(\\)\\*\\+\\,\\;\\=]",
          RESERVED$$ = merge(GEN_DELIMS$$, SUB_DELIMS$$),
          UCSCHAR$$ = isIRI ? "[\\xA0-\\u200D\\u2010-\\u2029\\u202F-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFEF]" : "[]",
          //subset, excludes bidi control characters
      IPRIVATE$$ = isIRI ? "[\\uE000-\\uF8FF]" : "[]",
          //subset
      UNRESERVED$$ = merge(ALPHA$$, DIGIT$$, "[\\-\\.\\_\\~]", UCSCHAR$$);
          subexp(ALPHA$$ + merge(ALPHA$$, DIGIT$$, "[\\+\\-\\.]") + "*");
          subexp(subexp(PCT_ENCODED$ + "|" + merge(UNRESERVED$$, SUB_DELIMS$$, "[\\:]")) + "*");
          var DEC_OCTET_RELAXED$ = subexp(subexp("25[0-5]") + "|" + subexp("2[0-4]" + DIGIT$$) + "|" + subexp("1" + DIGIT$$ + DIGIT$$) + "|" + subexp("0?[1-9]" + DIGIT$$) + "|0?0?" + DIGIT$$),
          //relaxed parsing rules
      IPV4ADDRESS$ = subexp(DEC_OCTET_RELAXED$ + "\\." + DEC_OCTET_RELAXED$ + "\\." + DEC_OCTET_RELAXED$ + "\\." + DEC_OCTET_RELAXED$),
          H16$ = subexp(HEXDIG$$ + "{1,4}"),
          LS32$ = subexp(subexp(H16$ + "\\:" + H16$) + "|" + IPV4ADDRESS$),
          IPV6ADDRESS1$ = subexp(subexp(H16$ + "\\:") + "{6}" + LS32$),
          //                           6( h16 ":" ) ls32
      IPV6ADDRESS2$ = subexp("\\:\\:" + subexp(H16$ + "\\:") + "{5}" + LS32$),
          //                      "::" 5( h16 ":" ) ls32
      IPV6ADDRESS3$ = subexp(subexp(H16$) + "?\\:\\:" + subexp(H16$ + "\\:") + "{4}" + LS32$),
          //[               h16 ] "::" 4( h16 ":" ) ls32
      IPV6ADDRESS4$ = subexp(subexp(subexp(H16$ + "\\:") + "{0,1}" + H16$) + "?\\:\\:" + subexp(H16$ + "\\:") + "{3}" + LS32$),
          //[ *1( h16 ":" ) h16 ] "::" 3( h16 ":" ) ls32
      IPV6ADDRESS5$ = subexp(subexp(subexp(H16$ + "\\:") + "{0,2}" + H16$) + "?\\:\\:" + subexp(H16$ + "\\:") + "{2}" + LS32$),
          //[ *2( h16 ":" ) h16 ] "::" 2( h16 ":" ) ls32
      IPV6ADDRESS6$ = subexp(subexp(subexp(H16$ + "\\:") + "{0,3}" + H16$) + "?\\:\\:" + H16$ + "\\:" + LS32$),
          //[ *3( h16 ":" ) h16 ] "::"    h16 ":"   ls32
      IPV6ADDRESS7$ = subexp(subexp(subexp(H16$ + "\\:") + "{0,4}" + H16$) + "?\\:\\:" + LS32$),
          //[ *4( h16 ":" ) h16 ] "::"              ls32
      IPV6ADDRESS8$ = subexp(subexp(subexp(H16$ + "\\:") + "{0,5}" + H16$) + "?\\:\\:" + H16$),
          //[ *5( h16 ":" ) h16 ] "::"              h16
      IPV6ADDRESS9$ = subexp(subexp(subexp(H16$ + "\\:") + "{0,6}" + H16$) + "?\\:\\:"),
          //[ *6( h16 ":" ) h16 ] "::"
      IPV6ADDRESS$ = subexp([IPV6ADDRESS1$, IPV6ADDRESS2$, IPV6ADDRESS3$, IPV6ADDRESS4$, IPV6ADDRESS5$, IPV6ADDRESS6$, IPV6ADDRESS7$, IPV6ADDRESS8$, IPV6ADDRESS9$].join("|")),
          ZONEID$ = subexp(subexp(UNRESERVED$$ + "|" + PCT_ENCODED$) + "+");
          //RFC 6874, with relaxed parsing rules
      subexp("[vV]" + HEXDIG$$ + "+\\." + merge(UNRESERVED$$, SUB_DELIMS$$, "[\\:]") + "+");
          //RFC 6874
      subexp(subexp(PCT_ENCODED$ + "|" + merge(UNRESERVED$$, SUB_DELIMS$$)) + "*");
          var PCHAR$ = subexp(PCT_ENCODED$ + "|" + merge(UNRESERVED$$, SUB_DELIMS$$, "[\\:\\@]"));
          subexp(subexp(PCT_ENCODED$ + "|" + merge(UNRESERVED$$, SUB_DELIMS$$, "[\\@]")) + "+");
          subexp(subexp(PCHAR$ + "|" + merge("[\\/\\?]", IPRIVATE$$)) + "*");
      return {
          NOT_SCHEME: new RegExp(merge("[^]", ALPHA$$, DIGIT$$, "[\\+\\-\\.]"), "g"),
          NOT_USERINFO: new RegExp(merge("[^\\%\\:]", UNRESERVED$$, SUB_DELIMS$$), "g"),
          NOT_HOST: new RegExp(merge("[^\\%\\[\\]\\:]", UNRESERVED$$, SUB_DELIMS$$), "g"),
          NOT_PATH: new RegExp(merge("[^\\%\\/\\:\\@]", UNRESERVED$$, SUB_DELIMS$$), "g"),
          NOT_PATH_NOSCHEME: new RegExp(merge("[^\\%\\/\\@]", UNRESERVED$$, SUB_DELIMS$$), "g"),
          NOT_QUERY: new RegExp(merge("[^\\%]", UNRESERVED$$, SUB_DELIMS$$, "[\\:\\@\\/\\?]", IPRIVATE$$), "g"),
          NOT_FRAGMENT: new RegExp(merge("[^\\%]", UNRESERVED$$, SUB_DELIMS$$, "[\\:\\@\\/\\?]"), "g"),
          ESCAPE: new RegExp(merge("[^]", UNRESERVED$$, SUB_DELIMS$$), "g"),
          UNRESERVED: new RegExp(UNRESERVED$$, "g"),
          OTHER_CHARS: new RegExp(merge("[^\\%]", UNRESERVED$$, RESERVED$$), "g"),
          PCT_ENCODED: new RegExp(PCT_ENCODED$, "g"),
          IPV4ADDRESS: new RegExp("^(" + IPV4ADDRESS$ + ")$"),
          IPV6ADDRESS: new RegExp("^\\[?(" + IPV6ADDRESS$ + ")" + subexp(subexp("\\%25|\\%(?!" + HEXDIG$$ + "{2})") + "(" + ZONEID$ + ")") + "?\\]?$") //RFC 6874, with relaxed parsing rules
      };
  }
  var URI_PROTOCOL = buildExps(false);

  var IRI_PROTOCOL = buildExps(true);

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

  /** Highest positive signed 32-bit float value */

  var maxInt = 2147483647; // aka. 0x7FFFFFFF or 2^31-1

  /** Bootstring parameters */
  var base = 36;
  var tMin = 1;
  var tMax = 26;
  var skew = 38;
  var damp = 700;
  var initialBias = 72;
  var initialN = 128; // 0x80
  var delimiter = '-'; // '\x2D'

  /** Regular expressions */
  var regexPunycode = /^xn--/;
  var regexNonASCII = /[^\0-\x7E]/; // non-ASCII chars
  var regexSeparators = /[\x2E\u3002\uFF0E\uFF61]/g; // RFC 3490 separators

  /** Error messages */
  var errors = {
  	'overflow': 'Overflow: input needs wider integers to process',
  	'not-basic': 'Illegal input >= 0x80 (not a basic code point)',
  	'invalid-input': 'Invalid input'
  };

  /** Convenience shortcuts */
  var baseMinusTMin = base - tMin;
  var floor = Math.floor;
  var stringFromCharCode = String.fromCharCode;

  /*--------------------------------------------------------------------------*/

  /**
   * A generic error utility function.
   * @private
   * @param {String} type The error type.
   * @returns {Error} Throws a `RangeError` with the applicable error message.
   */
  function error$1(type) {
  	throw new RangeError(errors[type]);
  }

  /**
   * A generic `Array#map` utility function.
   * @private
   * @param {Array} array The array to iterate over.
   * @param {Function} callback The function that gets called for every array
   * item.
   * @returns {Array} A new array of values returned by the callback function.
   */
  function map(array, fn) {
  	var result = [];
  	var length = array.length;
  	while (length--) {
  		result[length] = fn(array[length]);
  	}
  	return result;
  }

  /**
   * A simple `Array#map`-like wrapper to work with domain name strings or email
   * addresses.
   * @private
   * @param {String} domain The domain name or email address.
   * @param {Function} callback The function that gets called for every
   * character.
   * @returns {Array} A new string of characters returned by the callback
   * function.
   */
  function mapDomain(string, fn) {
  	var parts = string.split('@');
  	var result = '';
  	if (parts.length > 1) {
  		// In email addresses, only the domain name should be punycoded. Leave
  		// the local part (i.e. everything up to `@`) intact.
  		result = parts[0] + '@';
  		string = parts[1];
  	}
  	// Avoid `split(regex)` for IE8 compatibility. See #17.
  	string = string.replace(regexSeparators, '\x2E');
  	var labels = string.split('.');
  	var encoded = map(labels, fn).join('.');
  	return result + encoded;
  }

  /**
   * Creates an array containing the numeric code points of each Unicode
   * character in the string. While JavaScript uses UCS-2 internally,
   * this function will convert a pair of surrogate halves (each of which
   * UCS-2 exposes as separate characters) into a single code point,
   * matching UTF-16.
   * @see `punycode.ucs2.encode`
   * @see <https://mathiasbynens.be/notes/javascript-encoding>
   * @memberOf punycode.ucs2
   * @name decode
   * @param {String} string The Unicode input string (UCS-2).
   * @returns {Array} The new array of code points.
   */
  function ucs2decode(string) {
  	var output = [];
  	var counter = 0;
  	var length = string.length;
  	while (counter < length) {
  		var value = string.charCodeAt(counter++);
  		if (value >= 0xD800 && value <= 0xDBFF && counter < length) {
  			// It's a high surrogate, and there is a next character.
  			var extra = string.charCodeAt(counter++);
  			if ((extra & 0xFC00) == 0xDC00) {
  				// Low surrogate.
  				output.push(((value & 0x3FF) << 10) + (extra & 0x3FF) + 0x10000);
  			} else {
  				// It's an unmatched surrogate; only append this code unit, in case the
  				// next code unit is the high surrogate of a surrogate pair.
  				output.push(value);
  				counter--;
  			}
  		} else {
  			output.push(value);
  		}
  	}
  	return output;
  }

  /**
   * Creates a string based on an array of numeric code points.
   * @see `punycode.ucs2.decode`
   * @memberOf punycode.ucs2
   * @name encode
   * @param {Array} codePoints The array of numeric code points.
   * @returns {String} The new Unicode string (UCS-2).
   */
  var ucs2encode = function ucs2encode(array) {
  	return String.fromCodePoint.apply(String, toConsumableArray(array));
  };

  /**
   * Converts a basic code point into a digit/integer.
   * @see `digitToBasic()`
   * @private
   * @param {Number} codePoint The basic numeric code point value.
   * @returns {Number} The numeric value of a basic code point (for use in
   * representing integers) in the range `0` to `base - 1`, or `base` if
   * the code point does not represent a value.
   */
  var basicToDigit = function basicToDigit(codePoint) {
  	if (codePoint - 0x30 < 0x0A) {
  		return codePoint - 0x16;
  	}
  	if (codePoint - 0x41 < 0x1A) {
  		return codePoint - 0x41;
  	}
  	if (codePoint - 0x61 < 0x1A) {
  		return codePoint - 0x61;
  	}
  	return base;
  };

  /**
   * Converts a digit/integer into a basic code point.
   * @see `basicToDigit()`
   * @private
   * @param {Number} digit The numeric value of a basic code point.
   * @returns {Number} The basic code point whose value (when used for
   * representing integers) is `digit`, which needs to be in the range
   * `0` to `base - 1`. If `flag` is non-zero, the uppercase form is
   * used; else, the lowercase form is used. The behavior is undefined
   * if `flag` is non-zero and `digit` has no uppercase form.
   */
  var digitToBasic = function digitToBasic(digit, flag) {
  	//  0..25 map to ASCII a..z or A..Z
  	// 26..35 map to ASCII 0..9
  	return digit + 22 + 75 * (digit < 26) - ((flag != 0) << 5);
  };

  /**
   * Bias adaptation function as per section 3.4 of RFC 3492.
   * https://tools.ietf.org/html/rfc3492#section-3.4
   * @private
   */
  var adapt = function adapt(delta, numPoints, firstTime) {
  	var k = 0;
  	delta = firstTime ? floor(delta / damp) : delta >> 1;
  	delta += floor(delta / numPoints);
  	for (; /* no initialization */delta > baseMinusTMin * tMax >> 1; k += base) {
  		delta = floor(delta / baseMinusTMin);
  	}
  	return floor(k + (baseMinusTMin + 1) * delta / (delta + skew));
  };

  /**
   * Converts a Punycode string of ASCII-only symbols to a string of Unicode
   * symbols.
   * @memberOf punycode
   * @param {String} input The Punycode string of ASCII-only symbols.
   * @returns {String} The resulting string of Unicode symbols.
   */
  var decode = function decode(input) {
  	// Don't use UCS-2.
  	var output = [];
  	var inputLength = input.length;
  	var i = 0;
  	var n = initialN;
  	var bias = initialBias;

  	// Handle the basic code points: let `basic` be the number of input code
  	// points before the last delimiter, or `0` if there is none, then copy
  	// the first basic code points to the output.

  	var basic = input.lastIndexOf(delimiter);
  	if (basic < 0) {
  		basic = 0;
  	}

  	for (var j = 0; j < basic; ++j) {
  		// if it's not a basic code point
  		if (input.charCodeAt(j) >= 0x80) {
  			error$1('not-basic');
  		}
  		output.push(input.charCodeAt(j));
  	}

  	// Main decoding loop: start just after the last delimiter if any basic code
  	// points were copied; start at the beginning otherwise.

  	for (var index = basic > 0 ? basic + 1 : 0; index < inputLength;) /* no final expression */{

  		// `index` is the index of the next character to be consumed.
  		// Decode a generalized variable-length integer into `delta`,
  		// which gets added to `i`. The overflow checking is easier
  		// if we increase `i` as we go, then subtract off its starting
  		// value at the end to obtain `delta`.
  		var oldi = i;
  		for (var w = 1, k = base;; /* no condition */k += base) {

  			if (index >= inputLength) {
  				error$1('invalid-input');
  			}

  			var digit = basicToDigit(input.charCodeAt(index++));

  			if (digit >= base || digit > floor((maxInt - i) / w)) {
  				error$1('overflow');
  			}

  			i += digit * w;
  			var t = k <= bias ? tMin : k >= bias + tMax ? tMax : k - bias;

  			if (digit < t) {
  				break;
  			}

  			var baseMinusT = base - t;
  			if (w > floor(maxInt / baseMinusT)) {
  				error$1('overflow');
  			}

  			w *= baseMinusT;
  		}

  		var out = output.length + 1;
  		bias = adapt(i - oldi, out, oldi == 0);

  		// `i` was supposed to wrap around from `out` to `0`,
  		// incrementing `n` each time, so we'll fix that now:
  		if (floor(i / out) > maxInt - n) {
  			error$1('overflow');
  		}

  		n += floor(i / out);
  		i %= out;

  		// Insert `n` at position `i` of the output.
  		output.splice(i++, 0, n);
  	}

  	return String.fromCodePoint.apply(String, output);
  };

  /**
   * Converts a string of Unicode symbols (e.g. a domain name label) to a
   * Punycode string of ASCII-only symbols.
   * @memberOf punycode
   * @param {String} input The string of Unicode symbols.
   * @returns {String} The resulting Punycode string of ASCII-only symbols.
   */
  var encode = function encode(input) {
  	var output = [];

  	// Convert the input in UCS-2 to an array of Unicode code points.
  	input = ucs2decode(input);

  	// Cache the length.
  	var inputLength = input.length;

  	// Initialize the state.
  	var n = initialN;
  	var delta = 0;
  	var bias = initialBias;

  	// Handle the basic code points.
  	var _iteratorNormalCompletion = true;
  	var _didIteratorError = false;
  	var _iteratorError = undefined;

  	try {
  		for (var _iterator = input[Symbol.iterator](), _step; !(_iteratorNormalCompletion = (_step = _iterator.next()).done); _iteratorNormalCompletion = true) {
  			var _currentValue2 = _step.value;

  			if (_currentValue2 < 0x80) {
  				output.push(stringFromCharCode(_currentValue2));
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

  	var basicLength = output.length;
  	var handledCPCount = basicLength;

  	// `handledCPCount` is the number of code points that have been handled;
  	// `basicLength` is the number of basic code points.

  	// Finish the basic string with a delimiter unless it's empty.
  	if (basicLength) {
  		output.push(delimiter);
  	}

  	// Main encoding loop:
  	while (handledCPCount < inputLength) {

  		// All non-basic code points < n have been handled already. Find the next
  		// larger one:
  		var m = maxInt;
  		var _iteratorNormalCompletion2 = true;
  		var _didIteratorError2 = false;
  		var _iteratorError2 = undefined;

  		try {
  			for (var _iterator2 = input[Symbol.iterator](), _step2; !(_iteratorNormalCompletion2 = (_step2 = _iterator2.next()).done); _iteratorNormalCompletion2 = true) {
  				var currentValue = _step2.value;

  				if (currentValue >= n && currentValue < m) {
  					m = currentValue;
  				}
  			}

  			// Increase `delta` enough to advance the decoder's <n,i> state to <m,0>,
  			// but guard against overflow.
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

  		var handledCPCountPlusOne = handledCPCount + 1;
  		if (m - n > floor((maxInt - delta) / handledCPCountPlusOne)) {
  			error$1('overflow');
  		}

  		delta += (m - n) * handledCPCountPlusOne;
  		n = m;

  		var _iteratorNormalCompletion3 = true;
  		var _didIteratorError3 = false;
  		var _iteratorError3 = undefined;

  		try {
  			for (var _iterator3 = input[Symbol.iterator](), _step3; !(_iteratorNormalCompletion3 = (_step3 = _iterator3.next()).done); _iteratorNormalCompletion3 = true) {
  				var _currentValue = _step3.value;

  				if (_currentValue < n && ++delta > maxInt) {
  					error$1('overflow');
  				}
  				if (_currentValue == n) {
  					// Represent delta as a generalized variable-length integer.
  					var q = delta;
  					for (var k = base;; /* no condition */k += base) {
  						var t = k <= bias ? tMin : k >= bias + tMax ? tMax : k - bias;
  						if (q < t) {
  							break;
  						}
  						var qMinusT = q - t;
  						var baseMinusT = base - t;
  						output.push(stringFromCharCode(digitToBasic(t + qMinusT % baseMinusT, 0)));
  						q = floor(qMinusT / baseMinusT);
  					}

  					output.push(stringFromCharCode(digitToBasic(q, 0)));
  					bias = adapt(delta, handledCPCountPlusOne, handledCPCount == basicLength);
  					delta = 0;
  					++handledCPCount;
  				}
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

  		++delta;
  		++n;
  	}
  	return output.join('');
  };

  /**
   * Converts a Punycode string representing a domain name or an email address
   * to Unicode. Only the Punycoded parts of the input will be converted, i.e.
   * it doesn't matter if you call it on a string that has already been
   * converted to Unicode.
   * @memberOf punycode
   * @param {String} input The Punycoded domain name or email address to
   * convert to Unicode.
   * @returns {String} The Unicode representation of the given Punycode
   * string.
   */
  var toUnicode = function toUnicode(input) {
  	return mapDomain(input, function (string) {
  		return regexPunycode.test(string) ? decode(string.slice(4).toLowerCase()) : string;
  	});
  };

  /**
   * Converts a Unicode string representing a domain name or an email address to
   * Punycode. Only the non-ASCII parts of the domain name will be converted,
   * i.e. it doesn't matter if you call it with a domain that's already in
   * ASCII.
   * @memberOf punycode
   * @param {String} input The domain name or email address to convert, as a
   * Unicode string.
   * @returns {String} The Punycode representation of the given domain name or
   * email address.
   */
  var toASCII = function toASCII(input) {
  	return mapDomain(input, function (string) {
  		return regexNonASCII.test(string) ? 'xn--' + encode(string) : string;
  	});
  };

  /*--------------------------------------------------------------------------*/

  /** Define the public API */
  var punycode = {
  	/**
    * A string representing the current Punycode.js version number.
    * @memberOf punycode
    * @type String
    */
  	'version': '2.1.0',
  	/**
    * An object of methods to convert from JavaScript's internal character
    * representation (UCS-2) to Unicode code points, and back.
    * @see <https://mathiasbynens.be/notes/javascript-encoding>
    * @memberOf punycode
    * @type Object
    */
  	'ucs2': {
  		'decode': ucs2decode,
  		'encode': ucs2encode
  	},
  	'decode': decode,
  	'encode': encode,
  	'toASCII': toASCII,
  	'toUnicode': toUnicode
  };

  /**
   * URI.js
   *
   * @fileoverview An RFC 3986 compliant, scheme extendable URI parsing/validating/resolving library for JavaScript.
   * @author <a href="mailto:gary.court@gmail.com">Gary Court</a>
   * @see http://github.com/garycourt/uri-js
   */
  /**
   * Copyright 2011 Gary Court. All rights reserved.
   *
   * Redistribution and use in source and binary forms, with or without modification, are
   * permitted provided that the following conditions are met:
   *
   *    1. Redistributions of source code must retain the above copyright notice, this list of
   *       conditions and the following disclaimer.
   *
   *    2. Redistributions in binary form must reproduce the above copyright notice, this list
   *       of conditions and the following disclaimer in the documentation and/or other materials
   *       provided with the distribution.
   *
   * THIS SOFTWARE IS PROVIDED BY GARY COURT ``AS IS'' AND ANY EXPRESS OR IMPLIED
   * WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
   * FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL GARY COURT OR
   * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
   * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
   * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
   * ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
   * NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
   * ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
   *
   * The views and conclusions contained in the software and documentation are those of the
   * authors and should not be interpreted as representing official policies, either expressed
   * or implied, of Gary Court.
   */
  var SCHEMES = {};
  function pctEncChar(chr) {
      var c = chr.charCodeAt(0);
      var e = void 0;
      if (c < 16) e = "%0" + c.toString(16).toUpperCase();else if (c < 128) e = "%" + c.toString(16).toUpperCase();else if (c < 2048) e = "%" + (c >> 6 | 192).toString(16).toUpperCase() + "%" + (c & 63 | 128).toString(16).toUpperCase();else e = "%" + (c >> 12 | 224).toString(16).toUpperCase() + "%" + (c >> 6 & 63 | 128).toString(16).toUpperCase() + "%" + (c & 63 | 128).toString(16).toUpperCase();
      return e;
  }
  function pctDecChars(str) {
      var newStr = "";
      var i = 0;
      var il = str.length;
      while (i < il) {
          var c = parseInt(str.substr(i + 1, 2), 16);
          if (c < 128) {
              newStr += String.fromCharCode(c);
              i += 3;
          } else if (c >= 194 && c < 224) {
              if (il - i >= 6) {
                  var c2 = parseInt(str.substr(i + 4, 2), 16);
                  newStr += String.fromCharCode((c & 31) << 6 | c2 & 63);
              } else {
                  newStr += str.substr(i, 6);
              }
              i += 6;
          } else if (c >= 224) {
              if (il - i >= 9) {
                  var _c = parseInt(str.substr(i + 4, 2), 16);
                  var c3 = parseInt(str.substr(i + 7, 2), 16);
                  newStr += String.fromCharCode((c & 15) << 12 | (_c & 63) << 6 | c3 & 63);
              } else {
                  newStr += str.substr(i, 9);
              }
              i += 9;
          } else {
              newStr += str.substr(i, 3);
              i += 3;
          }
      }
      return newStr;
  }
  function _normalizeComponentEncoding(components, protocol) {
      function decodeUnreserved(str) {
          var decStr = pctDecChars(str);
          return !decStr.match(protocol.UNRESERVED) ? str : decStr;
      }
      if (components.scheme) components.scheme = String(components.scheme).replace(protocol.PCT_ENCODED, decodeUnreserved).toLowerCase().replace(protocol.NOT_SCHEME, "");
      if (components.userinfo !== undefined) components.userinfo = String(components.userinfo).replace(protocol.PCT_ENCODED, decodeUnreserved).replace(protocol.NOT_USERINFO, pctEncChar).replace(protocol.PCT_ENCODED, toUpperCase);
      if (components.host !== undefined) components.host = String(components.host).replace(protocol.PCT_ENCODED, decodeUnreserved).toLowerCase().replace(protocol.NOT_HOST, pctEncChar).replace(protocol.PCT_ENCODED, toUpperCase);
      if (components.path !== undefined) components.path = String(components.path).replace(protocol.PCT_ENCODED, decodeUnreserved).replace(components.scheme ? protocol.NOT_PATH : protocol.NOT_PATH_NOSCHEME, pctEncChar).replace(protocol.PCT_ENCODED, toUpperCase);
      if (components.query !== undefined) components.query = String(components.query).replace(protocol.PCT_ENCODED, decodeUnreserved).replace(protocol.NOT_QUERY, pctEncChar).replace(protocol.PCT_ENCODED, toUpperCase);
      if (components.fragment !== undefined) components.fragment = String(components.fragment).replace(protocol.PCT_ENCODED, decodeUnreserved).replace(protocol.NOT_FRAGMENT, pctEncChar).replace(protocol.PCT_ENCODED, toUpperCase);
      return components;
  }

  function _stripLeadingZeros(str) {
      return str.replace(/^0*(.*)/, "$1") || "0";
  }
  function _normalizeIPv4(host, protocol) {
      var matches = host.match(protocol.IPV4ADDRESS) || [];

      var _matches = slicedToArray(matches, 2),
          address = _matches[1];

      if (address) {
          return address.split(".").map(_stripLeadingZeros).join(".");
      } else {
          return host;
      }
  }
  function _normalizeIPv6(host, protocol) {
      var matches = host.match(protocol.IPV6ADDRESS) || [];

      var _matches2 = slicedToArray(matches, 3),
          address = _matches2[1],
          zone = _matches2[2];

      if (address) {
          var _address$toLowerCase$ = address.toLowerCase().split('::').reverse(),
              _address$toLowerCase$2 = slicedToArray(_address$toLowerCase$, 2),
              last = _address$toLowerCase$2[0],
              first = _address$toLowerCase$2[1];

          var firstFields = first ? first.split(":").map(_stripLeadingZeros) : [];
          var lastFields = last.split(":").map(_stripLeadingZeros);
          var isLastFieldIPv4Address = protocol.IPV4ADDRESS.test(lastFields[lastFields.length - 1]);
          var fieldCount = isLastFieldIPv4Address ? 7 : 8;
          var lastFieldsStart = lastFields.length - fieldCount;
          var fields = Array(fieldCount);
          for (var x = 0; x < fieldCount; ++x) {
              fields[x] = firstFields[x] || lastFields[lastFieldsStart + x] || '';
          }
          if (isLastFieldIPv4Address) {
              fields[fieldCount - 1] = _normalizeIPv4(fields[fieldCount - 1], protocol);
          }
          var allZeroFields = fields.reduce(function (acc, field, index) {
              if (!field || field === "0") {
                  var lastLongest = acc[acc.length - 1];
                  if (lastLongest && lastLongest.index + lastLongest.length === index) {
                      lastLongest.length++;
                  } else {
                      acc.push({ index: index, length: 1 });
                  }
              }
              return acc;
          }, []);
          var longestZeroFields = allZeroFields.sort(function (a, b) {
              return b.length - a.length;
          })[0];
          var newHost = void 0;
          if (longestZeroFields && longestZeroFields.length > 1) {
              var newFirst = fields.slice(0, longestZeroFields.index);
              var newLast = fields.slice(longestZeroFields.index + longestZeroFields.length);
              newHost = newFirst.join(":") + "::" + newLast.join(":");
          } else {
              newHost = fields.join(":");
          }
          if (zone) {
              newHost += "%" + zone;
          }
          return newHost;
      } else {
          return host;
      }
  }
  var URI_PARSE = /^(?:([^:\/?#]+):)?(?:\/\/((?:([^\/?#@]*)@)?(\[[^\/?#\]]+\]|[^\/?#:]*)(?:\:(\d*))?))?([^?#]*)(?:\?([^#]*))?(?:#((?:.|\n|\r)*))?/i;
  var NO_MATCH_IS_UNDEFINED = "".match(/(){0}/)[1] === undefined;
  function parse(uriString) {
      var options = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};

      var components = {};
      var protocol = options.iri !== false ? IRI_PROTOCOL : URI_PROTOCOL;
      if (options.reference === "suffix") uriString = (options.scheme ? options.scheme + ":" : "") + "//" + uriString;
      var matches = uriString.match(URI_PARSE);
      if (matches) {
          if (NO_MATCH_IS_UNDEFINED) {
              //store each component
              components.scheme = matches[1];
              components.userinfo = matches[3];
              components.host = matches[4];
              components.port = parseInt(matches[5], 10);
              components.path = matches[6] || "";
              components.query = matches[7];
              components.fragment = matches[8];
              //fix port number
              if (isNaN(components.port)) {
                  components.port = matches[5];
              }
          } else {
              //IE FIX for improper RegExp matching
              //store each component
              components.scheme = matches[1] || undefined;
              components.userinfo = uriString.indexOf("@") !== -1 ? matches[3] : undefined;
              components.host = uriString.indexOf("//") !== -1 ? matches[4] : undefined;
              components.port = parseInt(matches[5], 10);
              components.path = matches[6] || "";
              components.query = uriString.indexOf("?") !== -1 ? matches[7] : undefined;
              components.fragment = uriString.indexOf("#") !== -1 ? matches[8] : undefined;
              //fix port number
              if (isNaN(components.port)) {
                  components.port = uriString.match(/\/\/(?:.|\n)*\:(?:\/|\?|\#|$)/) ? matches[4] : undefined;
              }
          }
          if (components.host) {
              //normalize IP hosts
              components.host = _normalizeIPv6(_normalizeIPv4(components.host, protocol), protocol);
          }
          //determine reference type
          if (components.scheme === undefined && components.userinfo === undefined && components.host === undefined && components.port === undefined && !components.path && components.query === undefined) {
              components.reference = "same-document";
          } else if (components.scheme === undefined) {
              components.reference = "relative";
          } else if (components.fragment === undefined) {
              components.reference = "absolute";
          } else {
              components.reference = "uri";
          }
          //check for reference errors
          if (options.reference && options.reference !== "suffix" && options.reference !== components.reference) {
              components.error = components.error || "URI is not a " + options.reference + " reference.";
          }
          //find scheme handler
          var schemeHandler = SCHEMES[(options.scheme || components.scheme || "").toLowerCase()];
          //check if scheme can't handle IRIs
          if (!options.unicodeSupport && (!schemeHandler || !schemeHandler.unicodeSupport)) {
              //if host component is a domain name
              if (components.host && (options.domainHost || schemeHandler && schemeHandler.domainHost)) {
                  //convert Unicode IDN -> ASCII IDN
                  try {
                      components.host = punycode.toASCII(components.host.replace(protocol.PCT_ENCODED, pctDecChars).toLowerCase());
                  } catch (e) {
                      components.error = components.error || "Host's domain name can not be converted to ASCII via punycode: " + e;
                  }
              }
              //convert IRI -> URI
              _normalizeComponentEncoding(components, URI_PROTOCOL);
          } else {
              //normalize encodings
              _normalizeComponentEncoding(components, protocol);
          }
          //perform scheme specific parsing
          if (schemeHandler && schemeHandler.parse) {
              schemeHandler.parse(components, options);
          }
      } else {
          components.error = components.error || "URI can not be parsed.";
      }
      return components;
  }

  function _recomposeAuthority(components, options) {
      var protocol = options.iri !== false ? IRI_PROTOCOL : URI_PROTOCOL;
      var uriTokens = [];
      if (components.userinfo !== undefined) {
          uriTokens.push(components.userinfo);
          uriTokens.push("@");
      }
      if (components.host !== undefined) {
          //normalize IP hosts, add brackets and escape zone separator for IPv6
          uriTokens.push(_normalizeIPv6(_normalizeIPv4(String(components.host), protocol), protocol).replace(protocol.IPV6ADDRESS, function (_, $1, $2) {
              return "[" + $1 + ($2 ? "%25" + $2 : "") + "]";
          }));
      }
      if (typeof components.port === "number" || typeof components.port === "string") {
          uriTokens.push(":");
          uriTokens.push(String(components.port));
      }
      return uriTokens.length ? uriTokens.join("") : undefined;
  }

  var RDS1 = /^\.\.?\//;
  var RDS2 = /^\/\.(\/|$)/;
  var RDS3 = /^\/\.\.(\/|$)/;
  var RDS5 = /^\/?(?:.|\n)*?(?=\/|$)/;
  function removeDotSegments(input) {
      var output = [];
      while (input.length) {
          if (input.match(RDS1)) {
              input = input.replace(RDS1, "");
          } else if (input.match(RDS2)) {
              input = input.replace(RDS2, "/");
          } else if (input.match(RDS3)) {
              input = input.replace(RDS3, "/");
              output.pop();
          } else if (input === "." || input === "..") {
              input = "";
          } else {
              var im = input.match(RDS5);
              if (im) {
                  var s = im[0];
                  input = input.slice(s.length);
                  output.push(s);
              } else {
                  throw new Error("Unexpected dot segment condition");
              }
          }
      }
      return output.join("");
  }

  function serialize(components) {
      var options = arguments.length > 1 && arguments[1] !== undefined ? arguments[1] : {};

      var protocol = options.iri ? IRI_PROTOCOL : URI_PROTOCOL;
      var uriTokens = [];
      //find scheme handler
      var schemeHandler = SCHEMES[(options.scheme || components.scheme || "").toLowerCase()];
      //perform scheme specific serialization
      if (schemeHandler && schemeHandler.serialize) schemeHandler.serialize(components, options);
      if (components.host) {
          //if host component is an IPv6 address
          if (protocol.IPV6ADDRESS.test(components.host)) ;
          //TODO: normalize IPv6 address as per RFC 5952

          //if host component is a domain name
          else if (options.domainHost || schemeHandler && schemeHandler.domainHost) {
                  //convert IDN via punycode
                  try {
                      components.host = !options.iri ? punycode.toASCII(components.host.replace(protocol.PCT_ENCODED, pctDecChars).toLowerCase()) : punycode.toUnicode(components.host);
                  } catch (e) {
                      components.error = components.error || "Host's domain name can not be converted to " + (!options.iri ? "ASCII" : "Unicode") + " via punycode: " + e;
                  }
              }
      }
      //normalize encoding
      _normalizeComponentEncoding(components, protocol);
      if (options.reference !== "suffix" && components.scheme) {
          uriTokens.push(components.scheme);
          uriTokens.push(":");
      }
      var authority = _recomposeAuthority(components, options);
      if (authority !== undefined) {
          if (options.reference !== "suffix") {
              uriTokens.push("//");
          }
          uriTokens.push(authority);
          if (components.path && components.path.charAt(0) !== "/") {
              uriTokens.push("/");
          }
      }
      if (components.path !== undefined) {
          var s = components.path;
          if (!options.absolutePath && (!schemeHandler || !schemeHandler.absolutePath)) {
              s = removeDotSegments(s);
          }
          if (authority === undefined) {
              s = s.replace(/^\/\//, "/%2F"); //don't allow the path to start with "//"
          }
          uriTokens.push(s);
      }
      if (components.query !== undefined) {
          uriTokens.push("?");
          uriTokens.push(components.query);
      }
      if (components.fragment !== undefined) {
          uriTokens.push("#");
          uriTokens.push(components.fragment);
      }
      return uriTokens.join(""); //merge tokens into a string
  }

  function resolveComponents(base, relative) {
      var options = arguments.length > 2 && arguments[2] !== undefined ? arguments[2] : {};
      var skipNormalization = arguments[3];

      var target = {};
      if (!skipNormalization) {
          base = parse(serialize(base, options), options); //normalize base components
          relative = parse(serialize(relative, options), options); //normalize relative components
      }
      options = options || {};
      if (!options.tolerant && relative.scheme) {
          target.scheme = relative.scheme;
          //target.authority = relative.authority;
          target.userinfo = relative.userinfo;
          target.host = relative.host;
          target.port = relative.port;
          target.path = removeDotSegments(relative.path || "");
          target.query = relative.query;
      } else {
          if (relative.userinfo !== undefined || relative.host !== undefined || relative.port !== undefined) {
              //target.authority = relative.authority;
              target.userinfo = relative.userinfo;
              target.host = relative.host;
              target.port = relative.port;
              target.path = removeDotSegments(relative.path || "");
              target.query = relative.query;
          } else {
              if (!relative.path) {
                  target.path = base.path;
                  if (relative.query !== undefined) {
                      target.query = relative.query;
                  } else {
                      target.query = base.query;
                  }
              } else {
                  if (relative.path.charAt(0) === "/") {
                      target.path = removeDotSegments(relative.path);
                  } else {
                      if ((base.userinfo !== undefined || base.host !== undefined || base.port !== undefined) && !base.path) {
                          target.path = "/" + relative.path;
                      } else if (!base.path) {
                          target.path = relative.path;
                      } else {
                          target.path = base.path.slice(0, base.path.lastIndexOf("/") + 1) + relative.path;
                      }
                      target.path = removeDotSegments(target.path);
                  }
                  target.query = relative.query;
              }
              //target.authority = base.authority;
              target.userinfo = base.userinfo;
              target.host = base.host;
              target.port = base.port;
          }
          target.scheme = base.scheme;
      }
      target.fragment = relative.fragment;
      return target;
  }

  function resolve(baseURI, relativeURI, options) {
      var schemelessOptions = assign({ scheme: 'null' }, options);
      return serialize(resolveComponents(parse(baseURI, schemelessOptions), parse(relativeURI, schemelessOptions), schemelessOptions, true), schemelessOptions);
  }

  function normalize(uri, options) {
      if (typeof uri === "string") {
          uri = serialize(parse(uri, options), options);
      } else if (typeOf(uri) === "object") {
          uri = parse(serialize(uri, options), options);
      }
      return uri;
  }

  function equal(uriA, uriB, options) {
      if (typeof uriA === "string") {
          uriA = serialize(parse(uriA, options), options);
      } else if (typeOf(uriA) === "object") {
          uriA = serialize(uriA, options);
      }
      if (typeof uriB === "string") {
          uriB = serialize(parse(uriB, options), options);
      } else if (typeOf(uriB) === "object") {
          uriB = serialize(uriB, options);
      }
      return uriA === uriB;
  }

  function escapeComponent(str, options) {
      return str && str.toString().replace(!options || !options.iri ? URI_PROTOCOL.ESCAPE : IRI_PROTOCOL.ESCAPE, pctEncChar);
  }

  function unescapeComponent(str, options) {
      return str && str.toString().replace(!options || !options.iri ? URI_PROTOCOL.PCT_ENCODED : IRI_PROTOCOL.PCT_ENCODED, pctDecChars);
  }

  var handler = {
      scheme: "http",
      domainHost: true,
      parse: function parse(components, options) {
          //report missing host
          if (!components.host) {
              components.error = components.error || "HTTP URIs must have a host.";
          }
          return components;
      },
      serialize: function serialize(components, options) {
          var secure = String(components.scheme).toLowerCase() === "https";
          //normalize the default port
          if (components.port === (secure ? 443 : 80) || components.port === "") {
              components.port = undefined;
          }
          //normalize the empty path
          if (!components.path) {
              components.path = "/";
          }
          //NOTE: We do not parse query strings for HTTP URIs
          //as WWW Form Url Encoded query strings are part of the HTML4+ spec,
          //and not the HTTP spec.
          return components;
      }
  };

  var handler$1 = {
      scheme: "https",
      domainHost: handler.domainHost,
      parse: handler.parse,
      serialize: handler.serialize
  };

  function isSecure(wsComponents) {
      return typeof wsComponents.secure === 'boolean' ? wsComponents.secure : String(wsComponents.scheme).toLowerCase() === "wss";
  }
  //RFC 6455
  var handler$2 = {
      scheme: "ws",
      domainHost: true,
      parse: function parse(components, options) {
          var wsComponents = components;
          //indicate if the secure flag is set
          wsComponents.secure = isSecure(wsComponents);
          //construct resouce name
          wsComponents.resourceName = (wsComponents.path || '/') + (wsComponents.query ? '?' + wsComponents.query : '');
          wsComponents.path = undefined;
          wsComponents.query = undefined;
          return wsComponents;
      },
      serialize: function serialize(wsComponents, options) {
          //normalize the default port
          if (wsComponents.port === (isSecure(wsComponents) ? 443 : 80) || wsComponents.port === "") {
              wsComponents.port = undefined;
          }
          //ensure scheme matches secure flag
          if (typeof wsComponents.secure === 'boolean') {
              wsComponents.scheme = wsComponents.secure ? 'wss' : 'ws';
              wsComponents.secure = undefined;
          }
          //reconstruct path from resource name
          if (wsComponents.resourceName) {
              var _wsComponents$resourc = wsComponents.resourceName.split('?'),
                  _wsComponents$resourc2 = slicedToArray(_wsComponents$resourc, 2),
                  path = _wsComponents$resourc2[0],
                  query = _wsComponents$resourc2[1];

              wsComponents.path = path && path !== '/' ? path : undefined;
              wsComponents.query = query;
              wsComponents.resourceName = undefined;
          }
          //forbid fragment component
          wsComponents.fragment = undefined;
          return wsComponents;
      }
  };

  var handler$3 = {
      scheme: "wss",
      domainHost: handler$2.domainHost,
      parse: handler$2.parse,
      serialize: handler$2.serialize
  };

  var O = {};
  //RFC 3986
  var UNRESERVED$$ = "[A-Za-z0-9\\-\\.\\_\\~" + ("\\xA0-\\u200D\\u2010-\\u2029\\u202F-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFEF" ) + "]";
  var HEXDIG$$ = "[0-9A-Fa-f]"; //case-insensitive
  var PCT_ENCODED$ = subexp(subexp("%[EFef]" + HEXDIG$$ + "%" + HEXDIG$$ + HEXDIG$$ + "%" + HEXDIG$$ + HEXDIG$$) + "|" + subexp("%[89A-Fa-f]" + HEXDIG$$ + "%" + HEXDIG$$ + HEXDIG$$) + "|" + subexp("%" + HEXDIG$$ + HEXDIG$$)); //expanded
  //RFC 5322, except these symbols as per RFC 6068: @ : / ? # [ ] & ; =
  //const ATEXT$$ = "[A-Za-z0-9\\!\\#\\$\\%\\&\\'\\*\\+\\-\\/\\=\\?\\^\\_\\`\\{\\|\\}\\~]";
  //const WSP$$ = "[\\x20\\x09]";
  //const OBS_QTEXT$$ = "[\\x01-\\x08\\x0B\\x0C\\x0E-\\x1F\\x7F]";  //(%d1-8 / %d11-12 / %d14-31 / %d127)
  //const QTEXT$$ = merge("[\\x21\\x23-\\x5B\\x5D-\\x7E]", OBS_QTEXT$$);  //%d33 / %d35-91 / %d93-126 / obs-qtext
  //const VCHAR$$ = "[\\x21-\\x7E]";
  //const WSP$$ = "[\\x20\\x09]";
  //const OBS_QP$ = subexp("\\\\" + merge("[\\x00\\x0D\\x0A]", OBS_QTEXT$$));  //%d0 / CR / LF / obs-qtext
  //const FWS$ = subexp(subexp(WSP$$ + "*" + "\\x0D\\x0A") + "?" + WSP$$ + "+");
  //const QUOTED_PAIR$ = subexp(subexp("\\\\" + subexp(VCHAR$$ + "|" + WSP$$)) + "|" + OBS_QP$);
  //const QUOTED_STRING$ = subexp('\\"' + subexp(FWS$ + "?" + QCONTENT$) + "*" + FWS$ + "?" + '\\"');
  var ATEXT$$ = "[A-Za-z0-9\\!\\$\\%\\'\\*\\+\\-\\^\\_\\`\\{\\|\\}\\~]";
  var QTEXT$$ = "[\\!\\$\\%\\'\\(\\)\\*\\+\\,\\-\\.0-9\\<\\>A-Z\\x5E-\\x7E]";
  var VCHAR$$ = merge(QTEXT$$, "[\\\"\\\\]");
  var SOME_DELIMS$$ = "[\\!\\$\\'\\(\\)\\*\\+\\,\\;\\:\\@]";
  var UNRESERVED = new RegExp(UNRESERVED$$, "g");
  var PCT_ENCODED = new RegExp(PCT_ENCODED$, "g");
  var NOT_LOCAL_PART = new RegExp(merge("[^]", ATEXT$$, "[\\.]", '[\\"]', VCHAR$$), "g");
  var NOT_HFNAME = new RegExp(merge("[^]", UNRESERVED$$, SOME_DELIMS$$), "g");
  var NOT_HFVALUE = NOT_HFNAME;
  function decodeUnreserved(str) {
      var decStr = pctDecChars(str);
      return !decStr.match(UNRESERVED) ? str : decStr;
  }
  var handler$4 = {
      scheme: "mailto",
      parse: function parse$$1(components, options) {
          var mailtoComponents = components;
          var to = mailtoComponents.to = mailtoComponents.path ? mailtoComponents.path.split(",") : [];
          mailtoComponents.path = undefined;
          if (mailtoComponents.query) {
              var unknownHeaders = false;
              var headers = {};
              var hfields = mailtoComponents.query.split("&");
              for (var x = 0, xl = hfields.length; x < xl; ++x) {
                  var hfield = hfields[x].split("=");
                  switch (hfield[0]) {
                      case "to":
                          var toAddrs = hfield[1].split(",");
                          for (var _x = 0, _xl = toAddrs.length; _x < _xl; ++_x) {
                              to.push(toAddrs[_x]);
                          }
                          break;
                      case "subject":
                          mailtoComponents.subject = unescapeComponent(hfield[1], options);
                          break;
                      case "body":
                          mailtoComponents.body = unescapeComponent(hfield[1], options);
                          break;
                      default:
                          unknownHeaders = true;
                          headers[unescapeComponent(hfield[0], options)] = unescapeComponent(hfield[1], options);
                          break;
                  }
              }
              if (unknownHeaders) mailtoComponents.headers = headers;
          }
          mailtoComponents.query = undefined;
          for (var _x2 = 0, _xl2 = to.length; _x2 < _xl2; ++_x2) {
              var addr = to[_x2].split("@");
              addr[0] = unescapeComponent(addr[0]);
              if (!options.unicodeSupport) {
                  //convert Unicode IDN -> ASCII IDN
                  try {
                      addr[1] = punycode.toASCII(unescapeComponent(addr[1], options).toLowerCase());
                  } catch (e) {
                      mailtoComponents.error = mailtoComponents.error || "Email address's domain name can not be converted to ASCII via punycode: " + e;
                  }
              } else {
                  addr[1] = unescapeComponent(addr[1], options).toLowerCase();
              }
              to[_x2] = addr.join("@");
          }
          return mailtoComponents;
      },
      serialize: function serialize$$1(mailtoComponents, options) {
          var components = mailtoComponents;
          var to = toArray(mailtoComponents.to);
          if (to) {
              for (var x = 0, xl = to.length; x < xl; ++x) {
                  var toAddr = String(to[x]);
                  var atIdx = toAddr.lastIndexOf("@");
                  var localPart = toAddr.slice(0, atIdx).replace(PCT_ENCODED, decodeUnreserved).replace(PCT_ENCODED, toUpperCase).replace(NOT_LOCAL_PART, pctEncChar);
                  var domain = toAddr.slice(atIdx + 1);
                  //convert IDN via punycode
                  try {
                      domain = !options.iri ? punycode.toASCII(unescapeComponent(domain, options).toLowerCase()) : punycode.toUnicode(domain);
                  } catch (e) {
                      components.error = components.error || "Email address's domain name can not be converted to " + (!options.iri ? "ASCII" : "Unicode") + " via punycode: " + e;
                  }
                  to[x] = localPart + "@" + domain;
              }
              components.path = to.join(",");
          }
          var headers = mailtoComponents.headers = mailtoComponents.headers || {};
          if (mailtoComponents.subject) headers["subject"] = mailtoComponents.subject;
          if (mailtoComponents.body) headers["body"] = mailtoComponents.body;
          var fields = [];
          for (var name in headers) {
              if (headers[name] !== O[name]) {
                  fields.push(name.replace(PCT_ENCODED, decodeUnreserved).replace(PCT_ENCODED, toUpperCase).replace(NOT_HFNAME, pctEncChar) + "=" + headers[name].replace(PCT_ENCODED, decodeUnreserved).replace(PCT_ENCODED, toUpperCase).replace(NOT_HFVALUE, pctEncChar));
              }
          }
          if (fields.length) {
              components.query = fields.join("&");
          }
          return components;
      }
  };

  var URN_PARSE = /^([^\:]+)\:(.*)/;
  //RFC 2141
  var handler$5 = {
      scheme: "urn",
      parse: function parse$$1(components, options) {
          var matches = components.path && components.path.match(URN_PARSE);
          var urnComponents = components;
          if (matches) {
              var scheme = options.scheme || urnComponents.scheme || "urn";
              var nid = matches[1].toLowerCase();
              var nss = matches[2];
              var urnScheme = scheme + ":" + (options.nid || nid);
              var schemeHandler = SCHEMES[urnScheme];
              urnComponents.nid = nid;
              urnComponents.nss = nss;
              urnComponents.path = undefined;
              if (schemeHandler) {
                  urnComponents = schemeHandler.parse(urnComponents, options);
              }
          } else {
              urnComponents.error = urnComponents.error || "URN can not be parsed.";
          }
          return urnComponents;
      },
      serialize: function serialize$$1(urnComponents, options) {
          var scheme = options.scheme || urnComponents.scheme || "urn";
          var nid = urnComponents.nid;
          var urnScheme = scheme + ":" + (options.nid || nid);
          var schemeHandler = SCHEMES[urnScheme];
          if (schemeHandler) {
              urnComponents = schemeHandler.serialize(urnComponents, options);
          }
          var uriComponents = urnComponents;
          var nss = urnComponents.nss;
          uriComponents.path = (nid || options.nid) + ":" + nss;
          return uriComponents;
      }
  };

  var UUID = /^[0-9A-Fa-f]{8}(?:\-[0-9A-Fa-f]{4}){3}\-[0-9A-Fa-f]{12}$/;
  //RFC 4122
  var handler$6 = {
      scheme: "urn:uuid",
      parse: function parse(urnComponents, options) {
          var uuidComponents = urnComponents;
          uuidComponents.uuid = uuidComponents.nss;
          uuidComponents.nss = undefined;
          if (!options.tolerant && (!uuidComponents.uuid || !uuidComponents.uuid.match(UUID))) {
              uuidComponents.error = uuidComponents.error || "UUID is not valid.";
          }
          return uuidComponents;
      },
      serialize: function serialize(uuidComponents, options) {
          var urnComponents = uuidComponents;
          //normalize UUID
          urnComponents.nss = (uuidComponents.uuid || "").toLowerCase();
          return urnComponents;
      }
  };

  SCHEMES[handler.scheme] = handler;
  SCHEMES[handler$1.scheme] = handler$1;
  SCHEMES[handler$2.scheme] = handler$2;
  SCHEMES[handler$3.scheme] = handler$3;
  SCHEMES[handler$4.scheme] = handler$4;
  SCHEMES[handler$5.scheme] = handler$5;
  SCHEMES[handler$6.scheme] = handler$6;

  exports.SCHEMES = SCHEMES;
  exports.pctEncChar = pctEncChar;
  exports.pctDecChars = pctDecChars;
  exports.parse = parse;
  exports.removeDotSegments = removeDotSegments;
  exports.serialize = serialize;
  exports.resolveComponents = resolveComponents;
  exports.resolve = resolve;
  exports.normalize = normalize;
  exports.equal = equal;
  exports.escapeComponent = escapeComponent;
  exports.unescapeComponent = unescapeComponent;

  Object.defineProperty(exports, '__esModule', { value: true });

  })));

  }(uri_all, uri_all.exports));

  Object.defineProperty(uri$1, "__esModule", { value: true });
  const uri = uri_all.exports;
  uri.code = 'require("ajv/dist/runtime/uri").default';
  uri$1.default = uri;

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.CodeGen = exports.Name = exports.nil = exports.stringify = exports.str = exports._ = exports.KeywordCxt = void 0;
  var validate_1 = validate$1;
  Object.defineProperty(exports, "KeywordCxt", { enumerable: true, get: function () { return validate_1.KeywordCxt; } });
  var codegen_1 = codegen;
  Object.defineProperty(exports, "_", { enumerable: true, get: function () { return codegen_1._; } });
  Object.defineProperty(exports, "str", { enumerable: true, get: function () { return codegen_1.str; } });
  Object.defineProperty(exports, "stringify", { enumerable: true, get: function () { return codegen_1.stringify; } });
  Object.defineProperty(exports, "nil", { enumerable: true, get: function () { return codegen_1.nil; } });
  Object.defineProperty(exports, "Name", { enumerable: true, get: function () { return codegen_1.Name; } });
  Object.defineProperty(exports, "CodeGen", { enumerable: true, get: function () { return codegen_1.CodeGen; } });
  const validation_error_1 = validation_error;
  const ref_error_1 = ref_error;
  const rules_1 = rules;
  const compile_1 = compile;
  const codegen_2 = codegen;
  const resolve_1 = resolve$1;
  const dataType_1 = dataType;
  const util_1 = util;
  const $dataRefSchema = require$$9;
  const uri_1 = uri$1;
  const defaultRegExp = (str, flags) => new RegExp(str, flags);
  defaultRegExp.code = "new RegExp";
  const META_IGNORE_OPTIONS = ["removeAdditional", "useDefaults", "coerceTypes"];
  const EXT_SCOPE_NAMES = new Set([
      "validate",
      "serialize",
      "parse",
      "wrapper",
      "root",
      "schema",
      "keyword",
      "pattern",
      "formats",
      "validate$data",
      "func",
      "obj",
      "Error",
  ]);
  const removedOptions = {
      errorDataPath: "",
      format: "`validateFormats: false` can be used instead.",
      nullable: '"nullable" keyword is supported by default.',
      jsonPointers: "Deprecated jsPropertySyntax can be used instead.",
      extendRefs: "Deprecated ignoreKeywordsWithRef can be used instead.",
      missingRefs: "Pass empty schema with $id that should be ignored to ajv.addSchema.",
      processCode: "Use option `code: {process: (code, schemaEnv: object) => string}`",
      sourceCode: "Use option `code: {source: true}`",
      strictDefaults: "It is default now, see option `strict`.",
      strictKeywords: "It is default now, see option `strict`.",
      uniqueItems: '"uniqueItems" keyword is always validated.',
      unknownFormats: "Disable strict mode or pass `true` to `ajv.addFormat` (or `formats` option).",
      cache: "Map is used as cache, schema object as key.",
      serialize: "Map is used as cache, schema object as key.",
      ajvErrors: "It is default now.",
  };
  const deprecatedOptions = {
      ignoreKeywordsWithRef: "",
      jsPropertySyntax: "",
      unicode: '"minLength"/"maxLength" account for unicode characters by default.',
  };
  const MAX_EXPRESSION = 200;
  // eslint-disable-next-line complexity
  function requiredOptions(o) {
      var _a, _b, _c, _d, _e, _f, _g, _h, _j, _k, _l, _m, _o, _p, _q, _r, _s, _t, _u, _v, _w, _x, _y, _z, _0;
      const s = o.strict;
      const _optz = (_a = o.code) === null || _a === void 0 ? void 0 : _a.optimize;
      const optimize = _optz === true || _optz === undefined ? 1 : _optz || 0;
      const regExp = (_c = (_b = o.code) === null || _b === void 0 ? void 0 : _b.regExp) !== null && _c !== void 0 ? _c : defaultRegExp;
      const uriResolver = (_d = o.uriResolver) !== null && _d !== void 0 ? _d : uri_1.default;
      return {
          strictSchema: (_f = (_e = o.strictSchema) !== null && _e !== void 0 ? _e : s) !== null && _f !== void 0 ? _f : true,
          strictNumbers: (_h = (_g = o.strictNumbers) !== null && _g !== void 0 ? _g : s) !== null && _h !== void 0 ? _h : true,
          strictTypes: (_k = (_j = o.strictTypes) !== null && _j !== void 0 ? _j : s) !== null && _k !== void 0 ? _k : "log",
          strictTuples: (_m = (_l = o.strictTuples) !== null && _l !== void 0 ? _l : s) !== null && _m !== void 0 ? _m : "log",
          strictRequired: (_p = (_o = o.strictRequired) !== null && _o !== void 0 ? _o : s) !== null && _p !== void 0 ? _p : false,
          code: o.code ? { ...o.code, optimize, regExp } : { optimize, regExp },
          loopRequired: (_q = o.loopRequired) !== null && _q !== void 0 ? _q : MAX_EXPRESSION,
          loopEnum: (_r = o.loopEnum) !== null && _r !== void 0 ? _r : MAX_EXPRESSION,
          meta: (_s = o.meta) !== null && _s !== void 0 ? _s : true,
          messages: (_t = o.messages) !== null && _t !== void 0 ? _t : true,
          inlineRefs: (_u = o.inlineRefs) !== null && _u !== void 0 ? _u : true,
          schemaId: (_v = o.schemaId) !== null && _v !== void 0 ? _v : "$id",
          addUsedSchema: (_w = o.addUsedSchema) !== null && _w !== void 0 ? _w : true,
          validateSchema: (_x = o.validateSchema) !== null && _x !== void 0 ? _x : true,
          validateFormats: (_y = o.validateFormats) !== null && _y !== void 0 ? _y : true,
          unicodeRegExp: (_z = o.unicodeRegExp) !== null && _z !== void 0 ? _z : true,
          int32range: (_0 = o.int32range) !== null && _0 !== void 0 ? _0 : true,
          uriResolver: uriResolver,
      };
  }
  class Ajv {
      constructor(opts = {}) {
          this.schemas = {};
          this.refs = {};
          this.formats = {};
          this._compilations = new Set();
          this._loading = {};
          this._cache = new Map();
          opts = this.opts = { ...opts, ...requiredOptions(opts) };
          const { es5, lines } = this.opts.code;
          this.scope = new codegen_2.ValueScope({ scope: {}, prefixes: EXT_SCOPE_NAMES, es5, lines });
          this.logger = getLogger(opts.logger);
          const formatOpt = opts.validateFormats;
          opts.validateFormats = false;
          this.RULES = (0, rules_1.getRules)();
          checkOptions.call(this, removedOptions, opts, "NOT SUPPORTED");
          checkOptions.call(this, deprecatedOptions, opts, "DEPRECATED", "warn");
          this._metaOpts = getMetaSchemaOptions.call(this);
          if (opts.formats)
              addInitialFormats.call(this);
          this._addVocabularies();
          this._addDefaultMetaSchema();
          if (opts.keywords)
              addInitialKeywords.call(this, opts.keywords);
          if (typeof opts.meta == "object")
              this.addMetaSchema(opts.meta);
          addInitialSchemas.call(this);
          opts.validateFormats = formatOpt;
      }
      _addVocabularies() {
          this.addKeyword("$async");
      }
      _addDefaultMetaSchema() {
          const { $data, meta, schemaId } = this.opts;
          let _dataRefSchema = $dataRefSchema;
          if (schemaId === "id") {
              _dataRefSchema = { ...$dataRefSchema };
              _dataRefSchema.id = _dataRefSchema.$id;
              delete _dataRefSchema.$id;
          }
          if (meta && $data)
              this.addMetaSchema(_dataRefSchema, _dataRefSchema[schemaId], false);
      }
      defaultMeta() {
          const { meta, schemaId } = this.opts;
          return (this.opts.defaultMeta = typeof meta == "object" ? meta[schemaId] || meta : undefined);
      }
      validate(schemaKeyRef, // key, ref or schema object
      data // to be validated
      ) {
          let v;
          if (typeof schemaKeyRef == "string") {
              v = this.getSchema(schemaKeyRef);
              if (!v)
                  throw new Error(`no schema with key or ref "${schemaKeyRef}"`);
          }
          else {
              v = this.compile(schemaKeyRef);
          }
          const valid = v(data);
          if (!("$async" in v))
              this.errors = v.errors;
          return valid;
      }
      compile(schema, _meta) {
          const sch = this._addSchema(schema, _meta);
          return (sch.validate || this._compileSchemaEnv(sch));
      }
      compileAsync(schema, meta) {
          if (typeof this.opts.loadSchema != "function") {
              throw new Error("options.loadSchema should be a function");
          }
          const { loadSchema } = this.opts;
          return runCompileAsync.call(this, schema, meta);
          async function runCompileAsync(_schema, _meta) {
              await loadMetaSchema.call(this, _schema.$schema);
              const sch = this._addSchema(_schema, _meta);
              return sch.validate || _compileAsync.call(this, sch);
          }
          async function loadMetaSchema($ref) {
              if ($ref && !this.getSchema($ref)) {
                  await runCompileAsync.call(this, { $ref }, true);
              }
          }
          async function _compileAsync(sch) {
              try {
                  return this._compileSchemaEnv(sch);
              }
              catch (e) {
                  if (!(e instanceof ref_error_1.default))
                      throw e;
                  checkLoaded.call(this, e);
                  await loadMissingSchema.call(this, e.missingSchema);
                  return _compileAsync.call(this, sch);
              }
          }
          function checkLoaded({ missingSchema: ref, missingRef }) {
              if (this.refs[ref]) {
                  throw new Error(`AnySchema ${ref} is loaded but ${missingRef} cannot be resolved`);
              }
          }
          async function loadMissingSchema(ref) {
              const _schema = await _loadSchema.call(this, ref);
              if (!this.refs[ref])
                  await loadMetaSchema.call(this, _schema.$schema);
              if (!this.refs[ref])
                  this.addSchema(_schema, ref, meta);
          }
          async function _loadSchema(ref) {
              const p = this._loading[ref];
              if (p)
                  return p;
              try {
                  return await (this._loading[ref] = loadSchema(ref));
              }
              finally {
                  delete this._loading[ref];
              }
          }
      }
      // Adds schema to the instance
      addSchema(schema, // If array is passed, `key` will be ignored
      key, // Optional schema key. Can be passed to `validate` method instead of schema object or id/ref. One schema per instance can have empty `id` and `key`.
      _meta, // true if schema is a meta-schema. Used internally, addMetaSchema should be used instead.
      _validateSchema = this.opts.validateSchema // false to skip schema validation. Used internally, option validateSchema should be used instead.
      ) {
          if (Array.isArray(schema)) {
              for (const sch of schema)
                  this.addSchema(sch, undefined, _meta, _validateSchema);
              return this;
          }
          let id;
          if (typeof schema === "object") {
              const { schemaId } = this.opts;
              id = schema[schemaId];
              if (id !== undefined && typeof id != "string") {
                  throw new Error(`schema ${schemaId} must be string`);
              }
          }
          key = (0, resolve_1.normalizeId)(key || id);
          this._checkUnique(key);
          this.schemas[key] = this._addSchema(schema, _meta, key, _validateSchema, true);
          return this;
      }
      // Add schema that will be used to validate other schemas
      // options in META_IGNORE_OPTIONS are alway set to false
      addMetaSchema(schema, key, // schema key
      _validateSchema = this.opts.validateSchema // false to skip schema validation, can be used to override validateSchema option for meta-schema
      ) {
          this.addSchema(schema, key, true, _validateSchema);
          return this;
      }
      //  Validate schema against its meta-schema
      validateSchema(schema, throwOrLogError) {
          if (typeof schema == "boolean")
              return true;
          let $schema;
          $schema = schema.$schema;
          if ($schema !== undefined && typeof $schema != "string") {
              throw new Error("$schema must be a string");
          }
          $schema = $schema || this.opts.defaultMeta || this.defaultMeta();
          if (!$schema) {
              this.logger.warn("meta-schema not available");
              this.errors = null;
              return true;
          }
          const valid = this.validate($schema, schema);
          if (!valid && throwOrLogError) {
              const message = "schema is invalid: " + this.errorsText();
              if (this.opts.validateSchema === "log")
                  this.logger.error(message);
              else
                  throw new Error(message);
          }
          return valid;
      }
      // Get compiled schema by `key` or `ref`.
      // (`key` that was passed to `addSchema` or full schema reference - `schema.$id` or resolved id)
      getSchema(keyRef) {
          let sch;
          while (typeof (sch = getSchEnv.call(this, keyRef)) == "string")
              keyRef = sch;
          if (sch === undefined) {
              const { schemaId } = this.opts;
              const root = new compile_1.SchemaEnv({ schema: {}, schemaId });
              sch = compile_1.resolveSchema.call(this, root, keyRef);
              if (!sch)
                  return;
              this.refs[keyRef] = sch;
          }
          return (sch.validate || this._compileSchemaEnv(sch));
      }
      // Remove cached schema(s).
      // If no parameter is passed all schemas but meta-schemas are removed.
      // If RegExp is passed all schemas with key/id matching pattern but meta-schemas are removed.
      // Even if schema is referenced by other schemas it still can be removed as other schemas have local references.
      removeSchema(schemaKeyRef) {
          if (schemaKeyRef instanceof RegExp) {
              this._removeAllSchemas(this.schemas, schemaKeyRef);
              this._removeAllSchemas(this.refs, schemaKeyRef);
              return this;
          }
          switch (typeof schemaKeyRef) {
              case "undefined":
                  this._removeAllSchemas(this.schemas);
                  this._removeAllSchemas(this.refs);
                  this._cache.clear();
                  return this;
              case "string": {
                  const sch = getSchEnv.call(this, schemaKeyRef);
                  if (typeof sch == "object")
                      this._cache.delete(sch.schema);
                  delete this.schemas[schemaKeyRef];
                  delete this.refs[schemaKeyRef];
                  return this;
              }
              case "object": {
                  const cacheKey = schemaKeyRef;
                  this._cache.delete(cacheKey);
                  let id = schemaKeyRef[this.opts.schemaId];
                  if (id) {
                      id = (0, resolve_1.normalizeId)(id);
                      delete this.schemas[id];
                      delete this.refs[id];
                  }
                  return this;
              }
              default:
                  throw new Error("ajv.removeSchema: invalid parameter");
          }
      }
      // add "vocabulary" - a collection of keywords
      addVocabulary(definitions) {
          for (const def of definitions)
              this.addKeyword(def);
          return this;
      }
      addKeyword(kwdOrDef, def // deprecated
      ) {
          let keyword;
          if (typeof kwdOrDef == "string") {
              keyword = kwdOrDef;
              if (typeof def == "object") {
                  this.logger.warn("these parameters are deprecated, see docs for addKeyword");
                  def.keyword = keyword;
              }
          }
          else if (typeof kwdOrDef == "object" && def === undefined) {
              def = kwdOrDef;
              keyword = def.keyword;
              if (Array.isArray(keyword) && !keyword.length) {
                  throw new Error("addKeywords: keyword must be string or non-empty array");
              }
          }
          else {
              throw new Error("invalid addKeywords parameters");
          }
          checkKeyword.call(this, keyword, def);
          if (!def) {
              (0, util_1.eachItem)(keyword, (kwd) => addRule.call(this, kwd));
              return this;
          }
          keywordMetaschema.call(this, def);
          const definition = {
              ...def,
              type: (0, dataType_1.getJSONTypes)(def.type),
              schemaType: (0, dataType_1.getJSONTypes)(def.schemaType),
          };
          (0, util_1.eachItem)(keyword, definition.type.length === 0
              ? (k) => addRule.call(this, k, definition)
              : (k) => definition.type.forEach((t) => addRule.call(this, k, definition, t)));
          return this;
      }
      getKeyword(keyword) {
          const rule = this.RULES.all[keyword];
          return typeof rule == "object" ? rule.definition : !!rule;
      }
      // Remove keyword
      removeKeyword(keyword) {
          // TODO return type should be Ajv
          const { RULES } = this;
          delete RULES.keywords[keyword];
          delete RULES.all[keyword];
          for (const group of RULES.rules) {
              const i = group.rules.findIndex((rule) => rule.keyword === keyword);
              if (i >= 0)
                  group.rules.splice(i, 1);
          }
          return this;
      }
      // Add format
      addFormat(name, format) {
          if (typeof format == "string")
              format = new RegExp(format);
          this.formats[name] = format;
          return this;
      }
      errorsText(errors = this.errors, // optional array of validation errors
      { separator = ", ", dataVar = "data" } = {} // optional options with properties `separator` and `dataVar`
      ) {
          if (!errors || errors.length === 0)
              return "No errors";
          return errors
              .map((e) => `${dataVar}${e.instancePath} ${e.message}`)
              .reduce((text, msg) => text + separator + msg);
      }
      $dataMetaSchema(metaSchema, keywordsJsonPointers) {
          const rules = this.RULES.all;
          metaSchema = JSON.parse(JSON.stringify(metaSchema));
          for (const jsonPointer of keywordsJsonPointers) {
              const segments = jsonPointer.split("/").slice(1); // first segment is an empty string
              let keywords = metaSchema;
              for (const seg of segments)
                  keywords = keywords[seg];
              for (const key in rules) {
                  const rule = rules[key];
                  if (typeof rule != "object")
                      continue;
                  const { $data } = rule.definition;
                  const schema = keywords[key];
                  if ($data && schema)
                      keywords[key] = schemaOrData(schema);
              }
          }
          return metaSchema;
      }
      _removeAllSchemas(schemas, regex) {
          for (const keyRef in schemas) {
              const sch = schemas[keyRef];
              if (!regex || regex.test(keyRef)) {
                  if (typeof sch == "string") {
                      delete schemas[keyRef];
                  }
                  else if (sch && !sch.meta) {
                      this._cache.delete(sch.schema);
                      delete schemas[keyRef];
                  }
              }
          }
      }
      _addSchema(schema, meta, baseId, validateSchema = this.opts.validateSchema, addSchema = this.opts.addUsedSchema) {
          let id;
          const { schemaId } = this.opts;
          if (typeof schema == "object") {
              id = schema[schemaId];
          }
          else {
              if (this.opts.jtd)
                  throw new Error("schema must be object");
              else if (typeof schema != "boolean")
                  throw new Error("schema must be object or boolean");
          }
          let sch = this._cache.get(schema);
          if (sch !== undefined)
              return sch;
          baseId = (0, resolve_1.normalizeId)(id || baseId);
          const localRefs = resolve_1.getSchemaRefs.call(this, schema, baseId);
          sch = new compile_1.SchemaEnv({ schema, schemaId, meta, baseId, localRefs });
          this._cache.set(sch.schema, sch);
          if (addSchema && !baseId.startsWith("#")) {
              // TODO atm it is allowed to overwrite schemas without id (instead of not adding them)
              if (baseId)
                  this._checkUnique(baseId);
              this.refs[baseId] = sch;
          }
          if (validateSchema)
              this.validateSchema(schema, true);
          return sch;
      }
      _checkUnique(id) {
          if (this.schemas[id] || this.refs[id]) {
              throw new Error(`schema with key or id "${id}" already exists`);
          }
      }
      _compileSchemaEnv(sch) {
          if (sch.meta)
              this._compileMetaSchema(sch);
          else
              compile_1.compileSchema.call(this, sch);
          /* istanbul ignore if */
          if (!sch.validate)
              throw new Error("ajv implementation error");
          return sch.validate;
      }
      _compileMetaSchema(sch) {
          const currentOpts = this.opts;
          this.opts = this._metaOpts;
          try {
              compile_1.compileSchema.call(this, sch);
          }
          finally {
              this.opts = currentOpts;
          }
      }
  }
  exports.default = Ajv;
  Ajv.ValidationError = validation_error_1.default;
  Ajv.MissingRefError = ref_error_1.default;
  function checkOptions(checkOpts, options, msg, log = "error") {
      for (const key in checkOpts) {
          const opt = key;
          if (opt in options)
              this.logger[log](`${msg}: option ${key}. ${checkOpts[opt]}`);
      }
  }
  function getSchEnv(keyRef) {
      keyRef = (0, resolve_1.normalizeId)(keyRef); // TODO tests fail without this line
      return this.schemas[keyRef] || this.refs[keyRef];
  }
  function addInitialSchemas() {
      const optsSchemas = this.opts.schemas;
      if (!optsSchemas)
          return;
      if (Array.isArray(optsSchemas))
          this.addSchema(optsSchemas);
      else
          for (const key in optsSchemas)
              this.addSchema(optsSchemas[key], key);
  }
  function addInitialFormats() {
      for (const name in this.opts.formats) {
          const format = this.opts.formats[name];
          if (format)
              this.addFormat(name, format);
      }
  }
  function addInitialKeywords(defs) {
      if (Array.isArray(defs)) {
          this.addVocabulary(defs);
          return;
      }
      this.logger.warn("keywords option as map is deprecated, pass array");
      for (const keyword in defs) {
          const def = defs[keyword];
          if (!def.keyword)
              def.keyword = keyword;
          this.addKeyword(def);
      }
  }
  function getMetaSchemaOptions() {
      const metaOpts = { ...this.opts };
      for (const opt of META_IGNORE_OPTIONS)
          delete metaOpts[opt];
      return metaOpts;
  }
  const noLogs = { log() { }, warn() { }, error() { } };
  function getLogger(logger) {
      if (logger === false)
          return noLogs;
      if (logger === undefined)
          return console;
      if (logger.log && logger.warn && logger.error)
          return logger;
      throw new Error("logger must implement log, warn and error methods");
  }
  const KEYWORD_NAME = /^[a-z_$][a-z0-9_$:-]*$/i;
  function checkKeyword(keyword, def) {
      const { RULES } = this;
      (0, util_1.eachItem)(keyword, (kwd) => {
          if (RULES.keywords[kwd])
              throw new Error(`Keyword ${kwd} is already defined`);
          if (!KEYWORD_NAME.test(kwd))
              throw new Error(`Keyword ${kwd} has invalid name`);
      });
      if (!def)
          return;
      if (def.$data && !("code" in def || "validate" in def)) {
          throw new Error('$data keyword must have "code" or "validate" function');
      }
  }
  function addRule(keyword, definition, dataType) {
      var _a;
      const post = definition === null || definition === void 0 ? void 0 : definition.post;
      if (dataType && post)
          throw new Error('keyword with "post" flag cannot have "type"');
      const { RULES } = this;
      let ruleGroup = post ? RULES.post : RULES.rules.find(({ type: t }) => t === dataType);
      if (!ruleGroup) {
          ruleGroup = { type: dataType, rules: [] };
          RULES.rules.push(ruleGroup);
      }
      RULES.keywords[keyword] = true;
      if (!definition)
          return;
      const rule = {
          keyword,
          definition: {
              ...definition,
              type: (0, dataType_1.getJSONTypes)(definition.type),
              schemaType: (0, dataType_1.getJSONTypes)(definition.schemaType),
          },
      };
      if (definition.before)
          addBeforeRule.call(this, ruleGroup, rule, definition.before);
      else
          ruleGroup.rules.push(rule);
      RULES.all[keyword] = rule;
      (_a = definition.implements) === null || _a === void 0 ? void 0 : _a.forEach((kwd) => this.addKeyword(kwd));
  }
  function addBeforeRule(ruleGroup, rule, before) {
      const i = ruleGroup.rules.findIndex((_rule) => _rule.keyword === before);
      if (i >= 0) {
          ruleGroup.rules.splice(i, 0, rule);
      }
      else {
          ruleGroup.rules.push(rule);
          this.logger.warn(`rule ${before} is not defined`);
      }
  }
  function keywordMetaschema(def) {
      let { metaSchema } = def;
      if (metaSchema === undefined)
          return;
      if (def.$data && this.opts.$data)
          metaSchema = schemaOrData(metaSchema);
      def.validateSchema = this.compile(metaSchema, true);
  }
  const $dataRef = {
      $ref: "https://raw.githubusercontent.com/ajv-validator/ajv/master/lib/refs/data.json#",
  };
  function schemaOrData(schema) {
      return { anyOf: [schema, $dataRef] };
  }

  }(core$2));

  var draft7 = {};

  var core$1 = {};

  var id = {};

  Object.defineProperty(id, "__esModule", { value: true });
  const def$s = {
      keyword: "id",
      code() {
          throw new Error('NOT SUPPORTED: keyword "id", use "$id" for schema ID');
      },
  };
  id.default = def$s;

  var ref = {};

  Object.defineProperty(ref, "__esModule", { value: true });
  ref.callRef = ref.getValidate = void 0;
  const ref_error_1 = ref_error;
  const code_1$8 = code;
  const codegen_1$l = codegen;
  const names_1$1 = names$1;
  const compile_1$1 = compile;
  const util_1$j = util;
  const def$r = {
      keyword: "$ref",
      schemaType: "string",
      code(cxt) {
          const { gen, schema: $ref, it } = cxt;
          const { baseId, schemaEnv: env, validateName, opts, self } = it;
          const { root } = env;
          if (($ref === "#" || $ref === "#/") && baseId === root.baseId)
              return callRootRef();
          const schOrEnv = compile_1$1.resolveRef.call(self, root, baseId, $ref);
          if (schOrEnv === undefined)
              throw new ref_error_1.default(it.opts.uriResolver, baseId, $ref);
          if (schOrEnv instanceof compile_1$1.SchemaEnv)
              return callValidate(schOrEnv);
          return inlineRefSchema(schOrEnv);
          function callRootRef() {
              if (env === root)
                  return callRef(cxt, validateName, env, env.$async);
              const rootName = gen.scopeValue("root", { ref: root });
              return callRef(cxt, (0, codegen_1$l._) `${rootName}.validate`, root, root.$async);
          }
          function callValidate(sch) {
              const v = getValidate(cxt, sch);
              callRef(cxt, v, sch, sch.$async);
          }
          function inlineRefSchema(sch) {
              const schName = gen.scopeValue("schema", opts.code.source === true ? { ref: sch, code: (0, codegen_1$l.stringify)(sch) } : { ref: sch });
              const valid = gen.name("valid");
              const schCxt = cxt.subschema({
                  schema: sch,
                  dataTypes: [],
                  schemaPath: codegen_1$l.nil,
                  topSchemaRef: schName,
                  errSchemaPath: $ref,
              }, valid);
              cxt.mergeEvaluated(schCxt);
              cxt.ok(valid);
          }
      },
  };
  function getValidate(cxt, sch) {
      const { gen } = cxt;
      return sch.validate
          ? gen.scopeValue("validate", { ref: sch.validate })
          : (0, codegen_1$l._) `${gen.scopeValue("wrapper", { ref: sch })}.validate`;
  }
  ref.getValidate = getValidate;
  function callRef(cxt, v, sch, $async) {
      const { gen, it } = cxt;
      const { allErrors, schemaEnv: env, opts } = it;
      const passCxt = opts.passContext ? names_1$1.default.this : codegen_1$l.nil;
      if ($async)
          callAsyncRef();
      else
          callSyncRef();
      function callAsyncRef() {
          if (!env.$async)
              throw new Error("async schema referenced by sync schema");
          const valid = gen.let("valid");
          gen.try(() => {
              gen.code((0, codegen_1$l._) `await ${(0, code_1$8.callValidateCode)(cxt, v, passCxt)}`);
              addEvaluatedFrom(v); // TODO will not work with async, it has to be returned with the result
              if (!allErrors)
                  gen.assign(valid, true);
          }, (e) => {
              gen.if((0, codegen_1$l._) `!(${e} instanceof ${it.ValidationError})`, () => gen.throw(e));
              addErrorsFrom(e);
              if (!allErrors)
                  gen.assign(valid, false);
          });
          cxt.ok(valid);
      }
      function callSyncRef() {
          cxt.result((0, code_1$8.callValidateCode)(cxt, v, passCxt), () => addEvaluatedFrom(v), () => addErrorsFrom(v));
      }
      function addErrorsFrom(source) {
          const errs = (0, codegen_1$l._) `${source}.errors`;
          gen.assign(names_1$1.default.vErrors, (0, codegen_1$l._) `${names_1$1.default.vErrors} === null ? ${errs} : ${names_1$1.default.vErrors}.concat(${errs})`); // TODO tagged
          gen.assign(names_1$1.default.errors, (0, codegen_1$l._) `${names_1$1.default.vErrors}.length`);
      }
      function addEvaluatedFrom(source) {
          var _a;
          if (!it.opts.unevaluated)
              return;
          const schEvaluated = (_a = sch === null || sch === void 0 ? void 0 : sch.validate) === null || _a === void 0 ? void 0 : _a.evaluated;
          // TODO refactor
          if (it.props !== true) {
              if (schEvaluated && !schEvaluated.dynamicProps) {
                  if (schEvaluated.props !== undefined) {
                      it.props = util_1$j.mergeEvaluated.props(gen, schEvaluated.props, it.props);
                  }
              }
              else {
                  const props = gen.var("props", (0, codegen_1$l._) `${source}.evaluated.props`);
                  it.props = util_1$j.mergeEvaluated.props(gen, props, it.props, codegen_1$l.Name);
              }
          }
          if (it.items !== true) {
              if (schEvaluated && !schEvaluated.dynamicItems) {
                  if (schEvaluated.items !== undefined) {
                      it.items = util_1$j.mergeEvaluated.items(gen, schEvaluated.items, it.items);
                  }
              }
              else {
                  const items = gen.var("items", (0, codegen_1$l._) `${source}.evaluated.items`);
                  it.items = util_1$j.mergeEvaluated.items(gen, items, it.items, codegen_1$l.Name);
              }
          }
      }
  }
  ref.callRef = callRef;
  ref.default = def$r;

  Object.defineProperty(core$1, "__esModule", { value: true });
  const id_1 = id;
  const ref_1 = ref;
  const core = [
      "$schema",
      "$id",
      "$defs",
      "$vocabulary",
      { keyword: "$comment" },
      "definitions",
      id_1.default,
      ref_1.default,
  ];
  core$1.default = core;

  var validation$1 = {};

  var limitNumber = {};

  Object.defineProperty(limitNumber, "__esModule", { value: true });
  const codegen_1$k = codegen;
  const ops = codegen_1$k.operators;
  const KWDs = {
      maximum: { okStr: "<=", ok: ops.LTE, fail: ops.GT },
      minimum: { okStr: ">=", ok: ops.GTE, fail: ops.LT },
      exclusiveMaximum: { okStr: "<", ok: ops.LT, fail: ops.GTE },
      exclusiveMinimum: { okStr: ">", ok: ops.GT, fail: ops.LTE },
  };
  const error$i = {
      message: ({ keyword, schemaCode }) => (0, codegen_1$k.str) `must be ${KWDs[keyword].okStr} ${schemaCode}`,
      params: ({ keyword, schemaCode }) => (0, codegen_1$k._) `{comparison: ${KWDs[keyword].okStr}, limit: ${schemaCode}}`,
  };
  const def$q = {
      keyword: Object.keys(KWDs),
      type: "number",
      schemaType: "number",
      $data: true,
      error: error$i,
      code(cxt) {
          const { keyword, data, schemaCode } = cxt;
          cxt.fail$data((0, codegen_1$k._) `${data} ${KWDs[keyword].fail} ${schemaCode} || isNaN(${data})`);
      },
  };
  limitNumber.default = def$q;

  var multipleOf = {};

  Object.defineProperty(multipleOf, "__esModule", { value: true });
  const codegen_1$j = codegen;
  const error$h = {
      message: ({ schemaCode }) => (0, codegen_1$j.str) `must be multiple of ${schemaCode}`,
      params: ({ schemaCode }) => (0, codegen_1$j._) `{multipleOf: ${schemaCode}}`,
  };
  const def$p = {
      keyword: "multipleOf",
      type: "number",
      schemaType: "number",
      $data: true,
      error: error$h,
      code(cxt) {
          const { gen, data, schemaCode, it } = cxt;
          // const bdt = bad$DataType(schemaCode, <string>def.schemaType, $data)
          const prec = it.opts.multipleOfPrecision;
          const res = gen.let("res");
          const invalid = prec
              ? (0, codegen_1$j._) `Math.abs(Math.round(${res}) - ${res}) > 1e-${prec}`
              : (0, codegen_1$j._) `${res} !== parseInt(${res})`;
          cxt.fail$data((0, codegen_1$j._) `(${schemaCode} === 0 || (${res} = ${data}/${schemaCode}, ${invalid}))`);
      },
  };
  multipleOf.default = def$p;

  var limitLength = {};

  var ucs2length$1 = {};

  Object.defineProperty(ucs2length$1, "__esModule", { value: true });
  // https://mathiasbynens.be/notes/javascript-encoding
  // https://github.com/bestiejs/punycode.js - punycode.ucs2.decode
  function ucs2length(str) {
      const len = str.length;
      let length = 0;
      let pos = 0;
      let value;
      while (pos < len) {
          length++;
          value = str.charCodeAt(pos++);
          if (value >= 0xd800 && value <= 0xdbff && pos < len) {
              // high surrogate, and there is a next character
              value = str.charCodeAt(pos);
              if ((value & 0xfc00) === 0xdc00)
                  pos++; // low surrogate
          }
      }
      return length;
  }
  ucs2length$1.default = ucs2length;
  ucs2length.code = 'require("ajv/dist/runtime/ucs2length").default';

  Object.defineProperty(limitLength, "__esModule", { value: true });
  const codegen_1$i = codegen;
  const util_1$i = util;
  const ucs2length_1 = ucs2length$1;
  const error$g = {
      message({ keyword, schemaCode }) {
          const comp = keyword === "maxLength" ? "more" : "fewer";
          return (0, codegen_1$i.str) `must NOT have ${comp} than ${schemaCode} characters`;
      },
      params: ({ schemaCode }) => (0, codegen_1$i._) `{limit: ${schemaCode}}`,
  };
  const def$o = {
      keyword: ["maxLength", "minLength"],
      type: "string",
      schemaType: "number",
      $data: true,
      error: error$g,
      code(cxt) {
          const { keyword, data, schemaCode, it } = cxt;
          const op = keyword === "maxLength" ? codegen_1$i.operators.GT : codegen_1$i.operators.LT;
          const len = it.opts.unicode === false ? (0, codegen_1$i._) `${data}.length` : (0, codegen_1$i._) `${(0, util_1$i.useFunc)(cxt.gen, ucs2length_1.default)}(${data})`;
          cxt.fail$data((0, codegen_1$i._) `${len} ${op} ${schemaCode}`);
      },
  };
  limitLength.default = def$o;

  var pattern = {};

  Object.defineProperty(pattern, "__esModule", { value: true });
  const code_1$7 = code;
  const codegen_1$h = codegen;
  const error$f = {
      message: ({ schemaCode }) => (0, codegen_1$h.str) `must match pattern "${schemaCode}"`,
      params: ({ schemaCode }) => (0, codegen_1$h._) `{pattern: ${schemaCode}}`,
  };
  const def$n = {
      keyword: "pattern",
      type: "string",
      schemaType: "string",
      $data: true,
      error: error$f,
      code(cxt) {
          const { data, $data, schema, schemaCode, it } = cxt;
          // TODO regexp should be wrapped in try/catchs
          const u = it.opts.unicodeRegExp ? "u" : "";
          const regExp = $data ? (0, codegen_1$h._) `(new RegExp(${schemaCode}, ${u}))` : (0, code_1$7.usePattern)(cxt, schema);
          cxt.fail$data((0, codegen_1$h._) `!${regExp}.test(${data})`);
      },
  };
  pattern.default = def$n;

  var limitProperties = {};

  Object.defineProperty(limitProperties, "__esModule", { value: true });
  const codegen_1$g = codegen;
  const error$e = {
      message({ keyword, schemaCode }) {
          const comp = keyword === "maxProperties" ? "more" : "fewer";
          return (0, codegen_1$g.str) `must NOT have ${comp} than ${schemaCode} properties`;
      },
      params: ({ schemaCode }) => (0, codegen_1$g._) `{limit: ${schemaCode}}`,
  };
  const def$m = {
      keyword: ["maxProperties", "minProperties"],
      type: "object",
      schemaType: "number",
      $data: true,
      error: error$e,
      code(cxt) {
          const { keyword, data, schemaCode } = cxt;
          const op = keyword === "maxProperties" ? codegen_1$g.operators.GT : codegen_1$g.operators.LT;
          cxt.fail$data((0, codegen_1$g._) `Object.keys(${data}).length ${op} ${schemaCode}`);
      },
  };
  limitProperties.default = def$m;

  var required = {};

  Object.defineProperty(required, "__esModule", { value: true });
  const code_1$6 = code;
  const codegen_1$f = codegen;
  const util_1$h = util;
  const error$d = {
      message: ({ params: { missingProperty } }) => (0, codegen_1$f.str) `must have required property '${missingProperty}'`,
      params: ({ params: { missingProperty } }) => (0, codegen_1$f._) `{missingProperty: ${missingProperty}}`,
  };
  const def$l = {
      keyword: "required",
      type: "object",
      schemaType: "array",
      $data: true,
      error: error$d,
      code(cxt) {
          const { gen, schema, schemaCode, data, $data, it } = cxt;
          const { opts } = it;
          if (!$data && schema.length === 0)
              return;
          const useLoop = schema.length >= opts.loopRequired;
          if (it.allErrors)
              allErrorsMode();
          else
              exitOnErrorMode();
          if (opts.strictRequired) {
              const props = cxt.parentSchema.properties;
              const { definedProperties } = cxt.it;
              for (const requiredKey of schema) {
                  if ((props === null || props === void 0 ? void 0 : props[requiredKey]) === undefined && !definedProperties.has(requiredKey)) {
                      const schemaPath = it.schemaEnv.baseId + it.errSchemaPath;
                      const msg = `required property "${requiredKey}" is not defined at "${schemaPath}" (strictRequired)`;
                      (0, util_1$h.checkStrictMode)(it, msg, it.opts.strictRequired);
                  }
              }
          }
          function allErrorsMode() {
              if (useLoop || $data) {
                  cxt.block$data(codegen_1$f.nil, loopAllRequired);
              }
              else {
                  for (const prop of schema) {
                      (0, code_1$6.checkReportMissingProp)(cxt, prop);
                  }
              }
          }
          function exitOnErrorMode() {
              const missing = gen.let("missing");
              if (useLoop || $data) {
                  const valid = gen.let("valid", true);
                  cxt.block$data(valid, () => loopUntilMissing(missing, valid));
                  cxt.ok(valid);
              }
              else {
                  gen.if((0, code_1$6.checkMissingProp)(cxt, schema, missing));
                  (0, code_1$6.reportMissingProp)(cxt, missing);
                  gen.else();
              }
          }
          function loopAllRequired() {
              gen.forOf("prop", schemaCode, (prop) => {
                  cxt.setParams({ missingProperty: prop });
                  gen.if((0, code_1$6.noPropertyInData)(gen, data, prop, opts.ownProperties), () => cxt.error());
              });
          }
          function loopUntilMissing(missing, valid) {
              cxt.setParams({ missingProperty: missing });
              gen.forOf(missing, schemaCode, () => {
                  gen.assign(valid, (0, code_1$6.propertyInData)(gen, data, missing, opts.ownProperties));
                  gen.if((0, codegen_1$f.not)(valid), () => {
                      cxt.error();
                      gen.break();
                  });
              }, codegen_1$f.nil);
          }
      },
  };
  required.default = def$l;

  var limitItems = {};

  Object.defineProperty(limitItems, "__esModule", { value: true });
  const codegen_1$e = codegen;
  const error$c = {
      message({ keyword, schemaCode }) {
          const comp = keyword === "maxItems" ? "more" : "fewer";
          return (0, codegen_1$e.str) `must NOT have ${comp} than ${schemaCode} items`;
      },
      params: ({ schemaCode }) => (0, codegen_1$e._) `{limit: ${schemaCode}}`,
  };
  const def$k = {
      keyword: ["maxItems", "minItems"],
      type: "array",
      schemaType: "number",
      $data: true,
      error: error$c,
      code(cxt) {
          const { keyword, data, schemaCode } = cxt;
          const op = keyword === "maxItems" ? codegen_1$e.operators.GT : codegen_1$e.operators.LT;
          cxt.fail$data((0, codegen_1$e._) `${data}.length ${op} ${schemaCode}`);
      },
  };
  limitItems.default = def$k;

  var uniqueItems = {};

  var equal$1 = {};

  Object.defineProperty(equal$1, "__esModule", { value: true });
  // https://github.com/ajv-validator/ajv/issues/889
  const equal = fastDeepEqual;
  equal.code = 'require("ajv/dist/runtime/equal").default';
  equal$1.default = equal;

  Object.defineProperty(uniqueItems, "__esModule", { value: true });
  const dataType_1 = dataType;
  const codegen_1$d = codegen;
  const util_1$g = util;
  const equal_1$2 = equal$1;
  const error$b = {
      message: ({ params: { i, j } }) => (0, codegen_1$d.str) `must NOT have duplicate items (items ## ${j} and ${i} are identical)`,
      params: ({ params: { i, j } }) => (0, codegen_1$d._) `{i: ${i}, j: ${j}}`,
  };
  const def$j = {
      keyword: "uniqueItems",
      type: "array",
      schemaType: "boolean",
      $data: true,
      error: error$b,
      code(cxt) {
          const { gen, data, $data, schema, parentSchema, schemaCode, it } = cxt;
          if (!$data && !schema)
              return;
          const valid = gen.let("valid");
          const itemTypes = parentSchema.items ? (0, dataType_1.getSchemaTypes)(parentSchema.items) : [];
          cxt.block$data(valid, validateUniqueItems, (0, codegen_1$d._) `${schemaCode} === false`);
          cxt.ok(valid);
          function validateUniqueItems() {
              const i = gen.let("i", (0, codegen_1$d._) `${data}.length`);
              const j = gen.let("j");
              cxt.setParams({ i, j });
              gen.assign(valid, true);
              gen.if((0, codegen_1$d._) `${i} > 1`, () => (canOptimize() ? loopN : loopN2)(i, j));
          }
          function canOptimize() {
              return itemTypes.length > 0 && !itemTypes.some((t) => t === "object" || t === "array");
          }
          function loopN(i, j) {
              const item = gen.name("item");
              const wrongType = (0, dataType_1.checkDataTypes)(itemTypes, item, it.opts.strictNumbers, dataType_1.DataType.Wrong);
              const indices = gen.const("indices", (0, codegen_1$d._) `{}`);
              gen.for((0, codegen_1$d._) `;${i}--;`, () => {
                  gen.let(item, (0, codegen_1$d._) `${data}[${i}]`);
                  gen.if(wrongType, (0, codegen_1$d._) `continue`);
                  if (itemTypes.length > 1)
                      gen.if((0, codegen_1$d._) `typeof ${item} == "string"`, (0, codegen_1$d._) `${item} += "_"`);
                  gen
                      .if((0, codegen_1$d._) `typeof ${indices}[${item}] == "number"`, () => {
                      gen.assign(j, (0, codegen_1$d._) `${indices}[${item}]`);
                      cxt.error();
                      gen.assign(valid, false).break();
                  })
                      .code((0, codegen_1$d._) `${indices}[${item}] = ${i}`);
              });
          }
          function loopN2(i, j) {
              const eql = (0, util_1$g.useFunc)(gen, equal_1$2.default);
              const outer = gen.name("outer");
              gen.label(outer).for((0, codegen_1$d._) `;${i}--;`, () => gen.for((0, codegen_1$d._) `${j} = ${i}; ${j}--;`, () => gen.if((0, codegen_1$d._) `${eql}(${data}[${i}], ${data}[${j}])`, () => {
                  cxt.error();
                  gen.assign(valid, false).break(outer);
              })));
          }
      },
  };
  uniqueItems.default = def$j;

  var _const = {};

  Object.defineProperty(_const, "__esModule", { value: true });
  const codegen_1$c = codegen;
  const util_1$f = util;
  const equal_1$1 = equal$1;
  const error$a = {
      message: "must be equal to constant",
      params: ({ schemaCode }) => (0, codegen_1$c._) `{allowedValue: ${schemaCode}}`,
  };
  const def$i = {
      keyword: "const",
      $data: true,
      error: error$a,
      code(cxt) {
          const { gen, data, $data, schemaCode, schema } = cxt;
          if ($data || (schema && typeof schema == "object")) {
              cxt.fail$data((0, codegen_1$c._) `!${(0, util_1$f.useFunc)(gen, equal_1$1.default)}(${data}, ${schemaCode})`);
          }
          else {
              cxt.fail((0, codegen_1$c._) `${schema} !== ${data}`);
          }
      },
  };
  _const.default = def$i;

  var _enum = {};

  Object.defineProperty(_enum, "__esModule", { value: true });
  const codegen_1$b = codegen;
  const util_1$e = util;
  const equal_1 = equal$1;
  const error$9 = {
      message: "must be equal to one of the allowed values",
      params: ({ schemaCode }) => (0, codegen_1$b._) `{allowedValues: ${schemaCode}}`,
  };
  const def$h = {
      keyword: "enum",
      schemaType: "array",
      $data: true,
      error: error$9,
      code(cxt) {
          const { gen, data, $data, schema, schemaCode, it } = cxt;
          if (!$data && schema.length === 0)
              throw new Error("enum must have non-empty array");
          const useLoop = schema.length >= it.opts.loopEnum;
          let eql;
          const getEql = () => (eql !== null && eql !== void 0 ? eql : (eql = (0, util_1$e.useFunc)(gen, equal_1.default)));
          let valid;
          if (useLoop || $data) {
              valid = gen.let("valid");
              cxt.block$data(valid, loopEnum);
          }
          else {
              /* istanbul ignore if */
              if (!Array.isArray(schema))
                  throw new Error("ajv implementation error");
              const vSchema = gen.const("vSchema", schemaCode);
              valid = (0, codegen_1$b.or)(...schema.map((_x, i) => equalCode(vSchema, i)));
          }
          cxt.pass(valid);
          function loopEnum() {
              gen.assign(valid, false);
              gen.forOf("v", schemaCode, (v) => gen.if((0, codegen_1$b._) `${getEql()}(${data}, ${v})`, () => gen.assign(valid, true).break()));
          }
          function equalCode(vSchema, i) {
              const sch = schema[i];
              return typeof sch === "object" && sch !== null
                  ? (0, codegen_1$b._) `${getEql()}(${data}, ${vSchema}[${i}])`
                  : (0, codegen_1$b._) `${data} === ${sch}`;
          }
      },
  };
  _enum.default = def$h;

  Object.defineProperty(validation$1, "__esModule", { value: true });
  const limitNumber_1 = limitNumber;
  const multipleOf_1 = multipleOf;
  const limitLength_1 = limitLength;
  const pattern_1 = pattern;
  const limitProperties_1 = limitProperties;
  const required_1 = required;
  const limitItems_1 = limitItems;
  const uniqueItems_1 = uniqueItems;
  const const_1 = _const;
  const enum_1 = _enum;
  const validation = [
      // number
      limitNumber_1.default,
      multipleOf_1.default,
      // string
      limitLength_1.default,
      pattern_1.default,
      // object
      limitProperties_1.default,
      required_1.default,
      // array
      limitItems_1.default,
      uniqueItems_1.default,
      // any
      { keyword: "type", schemaType: ["string", "array"] },
      { keyword: "nullable", schemaType: "boolean" },
      const_1.default,
      enum_1.default,
  ];
  validation$1.default = validation;

  var applicator = {};

  var additionalItems = {};

  Object.defineProperty(additionalItems, "__esModule", { value: true });
  additionalItems.validateAdditionalItems = void 0;
  const codegen_1$a = codegen;
  const util_1$d = util;
  const error$8 = {
      message: ({ params: { len } }) => (0, codegen_1$a.str) `must NOT have more than ${len} items`,
      params: ({ params: { len } }) => (0, codegen_1$a._) `{limit: ${len}}`,
  };
  const def$g = {
      keyword: "additionalItems",
      type: "array",
      schemaType: ["boolean", "object"],
      before: "uniqueItems",
      error: error$8,
      code(cxt) {
          const { parentSchema, it } = cxt;
          const { items } = parentSchema;
          if (!Array.isArray(items)) {
              (0, util_1$d.checkStrictMode)(it, '"additionalItems" is ignored when "items" is not an array of schemas');
              return;
          }
          validateAdditionalItems(cxt, items);
      },
  };
  function validateAdditionalItems(cxt, items) {
      const { gen, schema, data, keyword, it } = cxt;
      it.items = true;
      const len = gen.const("len", (0, codegen_1$a._) `${data}.length`);
      if (schema === false) {
          cxt.setParams({ len: items.length });
          cxt.pass((0, codegen_1$a._) `${len} <= ${items.length}`);
      }
      else if (typeof schema == "object" && !(0, util_1$d.alwaysValidSchema)(it, schema)) {
          const valid = gen.var("valid", (0, codegen_1$a._) `${len} <= ${items.length}`); // TODO var
          gen.if((0, codegen_1$a.not)(valid), () => validateItems(valid));
          cxt.ok(valid);
      }
      function validateItems(valid) {
          gen.forRange("i", items.length, len, (i) => {
              cxt.subschema({ keyword, dataProp: i, dataPropType: util_1$d.Type.Num }, valid);
              if (!it.allErrors)
                  gen.if((0, codegen_1$a.not)(valid), () => gen.break());
          });
      }
  }
  additionalItems.validateAdditionalItems = validateAdditionalItems;
  additionalItems.default = def$g;

  var prefixItems = {};

  var items = {};

  Object.defineProperty(items, "__esModule", { value: true });
  items.validateTuple = void 0;
  const codegen_1$9 = codegen;
  const util_1$c = util;
  const code_1$5 = code;
  const def$f = {
      keyword: "items",
      type: "array",
      schemaType: ["object", "array", "boolean"],
      before: "uniqueItems",
      code(cxt) {
          const { schema, it } = cxt;
          if (Array.isArray(schema))
              return validateTuple(cxt, "additionalItems", schema);
          it.items = true;
          if ((0, util_1$c.alwaysValidSchema)(it, schema))
              return;
          cxt.ok((0, code_1$5.validateArray)(cxt));
      },
  };
  function validateTuple(cxt, extraItems, schArr = cxt.schema) {
      const { gen, parentSchema, data, keyword, it } = cxt;
      checkStrictTuple(parentSchema);
      if (it.opts.unevaluated && schArr.length && it.items !== true) {
          it.items = util_1$c.mergeEvaluated.items(gen, schArr.length, it.items);
      }
      const valid = gen.name("valid");
      const len = gen.const("len", (0, codegen_1$9._) `${data}.length`);
      schArr.forEach((sch, i) => {
          if ((0, util_1$c.alwaysValidSchema)(it, sch))
              return;
          gen.if((0, codegen_1$9._) `${len} > ${i}`, () => cxt.subschema({
              keyword,
              schemaProp: i,
              dataProp: i,
          }, valid));
          cxt.ok(valid);
      });
      function checkStrictTuple(sch) {
          const { opts, errSchemaPath } = it;
          const l = schArr.length;
          const fullTuple = l === sch.minItems && (l === sch.maxItems || sch[extraItems] === false);
          if (opts.strictTuples && !fullTuple) {
              const msg = `"${keyword}" is ${l}-tuple, but minItems or maxItems/${extraItems} are not specified or different at path "${errSchemaPath}"`;
              (0, util_1$c.checkStrictMode)(it, msg, opts.strictTuples);
          }
      }
  }
  items.validateTuple = validateTuple;
  items.default = def$f;

  Object.defineProperty(prefixItems, "__esModule", { value: true });
  const items_1$1 = items;
  const def$e = {
      keyword: "prefixItems",
      type: "array",
      schemaType: ["array"],
      before: "uniqueItems",
      code: (cxt) => (0, items_1$1.validateTuple)(cxt, "items"),
  };
  prefixItems.default = def$e;

  var items2020 = {};

  Object.defineProperty(items2020, "__esModule", { value: true });
  const codegen_1$8 = codegen;
  const util_1$b = util;
  const code_1$4 = code;
  const additionalItems_1$1 = additionalItems;
  const error$7 = {
      message: ({ params: { len } }) => (0, codegen_1$8.str) `must NOT have more than ${len} items`,
      params: ({ params: { len } }) => (0, codegen_1$8._) `{limit: ${len}}`,
  };
  const def$d = {
      keyword: "items",
      type: "array",
      schemaType: ["object", "boolean"],
      before: "uniqueItems",
      error: error$7,
      code(cxt) {
          const { schema, parentSchema, it } = cxt;
          const { prefixItems } = parentSchema;
          it.items = true;
          if ((0, util_1$b.alwaysValidSchema)(it, schema))
              return;
          if (prefixItems)
              (0, additionalItems_1$1.validateAdditionalItems)(cxt, prefixItems);
          else
              cxt.ok((0, code_1$4.validateArray)(cxt));
      },
  };
  items2020.default = def$d;

  var contains = {};

  Object.defineProperty(contains, "__esModule", { value: true });
  const codegen_1$7 = codegen;
  const util_1$a = util;
  const error$6 = {
      message: ({ params: { min, max } }) => max === undefined
          ? (0, codegen_1$7.str) `must contain at least ${min} valid item(s)`
          : (0, codegen_1$7.str) `must contain at least ${min} and no more than ${max} valid item(s)`,
      params: ({ params: { min, max } }) => max === undefined ? (0, codegen_1$7._) `{minContains: ${min}}` : (0, codegen_1$7._) `{minContains: ${min}, maxContains: ${max}}`,
  };
  const def$c = {
      keyword: "contains",
      type: "array",
      schemaType: ["object", "boolean"],
      before: "uniqueItems",
      trackErrors: true,
      error: error$6,
      code(cxt) {
          const { gen, schema, parentSchema, data, it } = cxt;
          let min;
          let max;
          const { minContains, maxContains } = parentSchema;
          if (it.opts.next) {
              min = minContains === undefined ? 1 : minContains;
              max = maxContains;
          }
          else {
              min = 1;
          }
          const len = gen.const("len", (0, codegen_1$7._) `${data}.length`);
          cxt.setParams({ min, max });
          if (max === undefined && min === 0) {
              (0, util_1$a.checkStrictMode)(it, `"minContains" == 0 without "maxContains": "contains" keyword ignored`);
              return;
          }
          if (max !== undefined && min > max) {
              (0, util_1$a.checkStrictMode)(it, `"minContains" > "maxContains" is always invalid`);
              cxt.fail();
              return;
          }
          if ((0, util_1$a.alwaysValidSchema)(it, schema)) {
              let cond = (0, codegen_1$7._) `${len} >= ${min}`;
              if (max !== undefined)
                  cond = (0, codegen_1$7._) `${cond} && ${len} <= ${max}`;
              cxt.pass(cond);
              return;
          }
          it.items = true;
          const valid = gen.name("valid");
          if (max === undefined && min === 1) {
              validateItems(valid, () => gen.if(valid, () => gen.break()));
          }
          else if (min === 0) {
              gen.let(valid, true);
              if (max !== undefined)
                  gen.if((0, codegen_1$7._) `${data}.length > 0`, validateItemsWithCount);
          }
          else {
              gen.let(valid, false);
              validateItemsWithCount();
          }
          cxt.result(valid, () => cxt.reset());
          function validateItemsWithCount() {
              const schValid = gen.name("_valid");
              const count = gen.let("count", 0);
              validateItems(schValid, () => gen.if(schValid, () => checkLimits(count)));
          }
          function validateItems(_valid, block) {
              gen.forRange("i", 0, len, (i) => {
                  cxt.subschema({
                      keyword: "contains",
                      dataProp: i,
                      dataPropType: util_1$a.Type.Num,
                      compositeRule: true,
                  }, _valid);
                  block();
              });
          }
          function checkLimits(count) {
              gen.code((0, codegen_1$7._) `${count}++`);
              if (max === undefined) {
                  gen.if((0, codegen_1$7._) `${count} >= ${min}`, () => gen.assign(valid, true).break());
              }
              else {
                  gen.if((0, codegen_1$7._) `${count} > ${max}`, () => gen.assign(valid, false).break());
                  if (min === 1)
                      gen.assign(valid, true);
                  else
                      gen.if((0, codegen_1$7._) `${count} >= ${min}`, () => gen.assign(valid, true));
              }
          }
      },
  };
  contains.default = def$c;

  var dependencies = {};

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.validateSchemaDeps = exports.validatePropertyDeps = exports.error = void 0;
  const codegen_1 = codegen;
  const util_1 = util;
  const code_1 = code;
  exports.error = {
      message: ({ params: { property, depsCount, deps } }) => {
          const property_ies = depsCount === 1 ? "property" : "properties";
          return (0, codegen_1.str) `must have ${property_ies} ${deps} when property ${property} is present`;
      },
      params: ({ params: { property, depsCount, deps, missingProperty } }) => (0, codegen_1._) `{property: ${property},
    missingProperty: ${missingProperty},
    depsCount: ${depsCount},
    deps: ${deps}}`, // TODO change to reference
  };
  const def = {
      keyword: "dependencies",
      type: "object",
      schemaType: "object",
      error: exports.error,
      code(cxt) {
          const [propDeps, schDeps] = splitDependencies(cxt);
          validatePropertyDeps(cxt, propDeps);
          validateSchemaDeps(cxt, schDeps);
      },
  };
  function splitDependencies({ schema }) {
      const propertyDeps = {};
      const schemaDeps = {};
      for (const key in schema) {
          if (key === "__proto__")
              continue;
          const deps = Array.isArray(schema[key]) ? propertyDeps : schemaDeps;
          deps[key] = schema[key];
      }
      return [propertyDeps, schemaDeps];
  }
  function validatePropertyDeps(cxt, propertyDeps = cxt.schema) {
      const { gen, data, it } = cxt;
      if (Object.keys(propertyDeps).length === 0)
          return;
      const missing = gen.let("missing");
      for (const prop in propertyDeps) {
          const deps = propertyDeps[prop];
          if (deps.length === 0)
              continue;
          const hasProperty = (0, code_1.propertyInData)(gen, data, prop, it.opts.ownProperties);
          cxt.setParams({
              property: prop,
              depsCount: deps.length,
              deps: deps.join(", "),
          });
          if (it.allErrors) {
              gen.if(hasProperty, () => {
                  for (const depProp of deps) {
                      (0, code_1.checkReportMissingProp)(cxt, depProp);
                  }
              });
          }
          else {
              gen.if((0, codegen_1._) `${hasProperty} && (${(0, code_1.checkMissingProp)(cxt, deps, missing)})`);
              (0, code_1.reportMissingProp)(cxt, missing);
              gen.else();
          }
      }
  }
  exports.validatePropertyDeps = validatePropertyDeps;
  function validateSchemaDeps(cxt, schemaDeps = cxt.schema) {
      const { gen, data, keyword, it } = cxt;
      const valid = gen.name("valid");
      for (const prop in schemaDeps) {
          if ((0, util_1.alwaysValidSchema)(it, schemaDeps[prop]))
              continue;
          gen.if((0, code_1.propertyInData)(gen, data, prop, it.opts.ownProperties), () => {
              const schCxt = cxt.subschema({ keyword, schemaProp: prop }, valid);
              cxt.mergeValidEvaluated(schCxt, valid);
          }, () => gen.var(valid, true) // TODO var
          );
          cxt.ok(valid);
      }
  }
  exports.validateSchemaDeps = validateSchemaDeps;
  exports.default = def;

  }(dependencies));

  var propertyNames = {};

  Object.defineProperty(propertyNames, "__esModule", { value: true });
  const codegen_1$6 = codegen;
  const util_1$9 = util;
  const error$5 = {
      message: "property name must be valid",
      params: ({ params }) => (0, codegen_1$6._) `{propertyName: ${params.propertyName}}`,
  };
  const def$b = {
      keyword: "propertyNames",
      type: "object",
      schemaType: ["object", "boolean"],
      error: error$5,
      code(cxt) {
          const { gen, schema, data, it } = cxt;
          if ((0, util_1$9.alwaysValidSchema)(it, schema))
              return;
          const valid = gen.name("valid");
          gen.forIn("key", data, (key) => {
              cxt.setParams({ propertyName: key });
              cxt.subschema({
                  keyword: "propertyNames",
                  data: key,
                  dataTypes: ["string"],
                  propertyName: key,
                  compositeRule: true,
              }, valid);
              gen.if((0, codegen_1$6.not)(valid), () => {
                  cxt.error(true);
                  if (!it.allErrors)
                      gen.break();
              });
          });
          cxt.ok(valid);
      },
  };
  propertyNames.default = def$b;

  var additionalProperties = {};

  Object.defineProperty(additionalProperties, "__esModule", { value: true });
  const code_1$3 = code;
  const codegen_1$5 = codegen;
  const names_1 = names$1;
  const util_1$8 = util;
  const error$4 = {
      message: "must NOT have additional properties",
      params: ({ params }) => (0, codegen_1$5._) `{additionalProperty: ${params.additionalProperty}}`,
  };
  const def$a = {
      keyword: "additionalProperties",
      type: ["object"],
      schemaType: ["boolean", "object"],
      allowUndefined: true,
      trackErrors: true,
      error: error$4,
      code(cxt) {
          const { gen, schema, parentSchema, data, errsCount, it } = cxt;
          /* istanbul ignore if */
          if (!errsCount)
              throw new Error("ajv implementation error");
          const { allErrors, opts } = it;
          it.props = true;
          if (opts.removeAdditional !== "all" && (0, util_1$8.alwaysValidSchema)(it, schema))
              return;
          const props = (0, code_1$3.allSchemaProperties)(parentSchema.properties);
          const patProps = (0, code_1$3.allSchemaProperties)(parentSchema.patternProperties);
          checkAdditionalProperties();
          cxt.ok((0, codegen_1$5._) `${errsCount} === ${names_1.default.errors}`);
          function checkAdditionalProperties() {
              gen.forIn("key", data, (key) => {
                  if (!props.length && !patProps.length)
                      additionalPropertyCode(key);
                  else
                      gen.if(isAdditional(key), () => additionalPropertyCode(key));
              });
          }
          function isAdditional(key) {
              let definedProp;
              if (props.length > 8) {
                  // TODO maybe an option instead of hard-coded 8?
                  const propsSchema = (0, util_1$8.schemaRefOrVal)(it, parentSchema.properties, "properties");
                  definedProp = (0, code_1$3.isOwnProperty)(gen, propsSchema, key);
              }
              else if (props.length) {
                  definedProp = (0, codegen_1$5.or)(...props.map((p) => (0, codegen_1$5._) `${key} === ${p}`));
              }
              else {
                  definedProp = codegen_1$5.nil;
              }
              if (patProps.length) {
                  definedProp = (0, codegen_1$5.or)(definedProp, ...patProps.map((p) => (0, codegen_1$5._) `${(0, code_1$3.usePattern)(cxt, p)}.test(${key})`));
              }
              return (0, codegen_1$5.not)(definedProp);
          }
          function deleteAdditional(key) {
              gen.code((0, codegen_1$5._) `delete ${data}[${key}]`);
          }
          function additionalPropertyCode(key) {
              if (opts.removeAdditional === "all" || (opts.removeAdditional && schema === false)) {
                  deleteAdditional(key);
                  return;
              }
              if (schema === false) {
                  cxt.setParams({ additionalProperty: key });
                  cxt.error();
                  if (!allErrors)
                      gen.break();
                  return;
              }
              if (typeof schema == "object" && !(0, util_1$8.alwaysValidSchema)(it, schema)) {
                  const valid = gen.name("valid");
                  if (opts.removeAdditional === "failing") {
                      applyAdditionalSchema(key, valid, false);
                      gen.if((0, codegen_1$5.not)(valid), () => {
                          cxt.reset();
                          deleteAdditional(key);
                      });
                  }
                  else {
                      applyAdditionalSchema(key, valid);
                      if (!allErrors)
                          gen.if((0, codegen_1$5.not)(valid), () => gen.break());
                  }
              }
          }
          function applyAdditionalSchema(key, valid, errors) {
              const subschema = {
                  keyword: "additionalProperties",
                  dataProp: key,
                  dataPropType: util_1$8.Type.Str,
              };
              if (errors === false) {
                  Object.assign(subschema, {
                      compositeRule: true,
                      createErrors: false,
                      allErrors: false,
                  });
              }
              cxt.subschema(subschema, valid);
          }
      },
  };
  additionalProperties.default = def$a;

  var properties$1 = {};

  Object.defineProperty(properties$1, "__esModule", { value: true });
  const validate_1 = validate$1;
  const code_1$2 = code;
  const util_1$7 = util;
  const additionalProperties_1$1 = additionalProperties;
  const def$9 = {
      keyword: "properties",
      type: "object",
      schemaType: "object",
      code(cxt) {
          const { gen, schema, parentSchema, data, it } = cxt;
          if (it.opts.removeAdditional === "all" && parentSchema.additionalProperties === undefined) {
              additionalProperties_1$1.default.code(new validate_1.KeywordCxt(it, additionalProperties_1$1.default, "additionalProperties"));
          }
          const allProps = (0, code_1$2.allSchemaProperties)(schema);
          for (const prop of allProps) {
              it.definedProperties.add(prop);
          }
          if (it.opts.unevaluated && allProps.length && it.props !== true) {
              it.props = util_1$7.mergeEvaluated.props(gen, (0, util_1$7.toHash)(allProps), it.props);
          }
          const properties = allProps.filter((p) => !(0, util_1$7.alwaysValidSchema)(it, schema[p]));
          if (properties.length === 0)
              return;
          const valid = gen.name("valid");
          for (const prop of properties) {
              if (hasDefault(prop)) {
                  applyPropertySchema(prop);
              }
              else {
                  gen.if((0, code_1$2.propertyInData)(gen, data, prop, it.opts.ownProperties));
                  applyPropertySchema(prop);
                  if (!it.allErrors)
                      gen.else().var(valid, true);
                  gen.endIf();
              }
              cxt.it.definedProperties.add(prop);
              cxt.ok(valid);
          }
          function hasDefault(prop) {
              return it.opts.useDefaults && !it.compositeRule && schema[prop].default !== undefined;
          }
          function applyPropertySchema(prop) {
              cxt.subschema({
                  keyword: "properties",
                  schemaProp: prop,
                  dataProp: prop,
              }, valid);
          }
      },
  };
  properties$1.default = def$9;

  var patternProperties = {};

  Object.defineProperty(patternProperties, "__esModule", { value: true });
  const code_1$1 = code;
  const codegen_1$4 = codegen;
  const util_1$6 = util;
  const util_2 = util;
  const def$8 = {
      keyword: "patternProperties",
      type: "object",
      schemaType: "object",
      code(cxt) {
          const { gen, schema, data, parentSchema, it } = cxt;
          const { opts } = it;
          const patterns = (0, code_1$1.allSchemaProperties)(schema);
          const alwaysValidPatterns = patterns.filter((p) => (0, util_1$6.alwaysValidSchema)(it, schema[p]));
          if (patterns.length === 0 ||
              (alwaysValidPatterns.length === patterns.length &&
                  (!it.opts.unevaluated || it.props === true))) {
              return;
          }
          const checkProperties = opts.strictSchema && !opts.allowMatchingProperties && parentSchema.properties;
          const valid = gen.name("valid");
          if (it.props !== true && !(it.props instanceof codegen_1$4.Name)) {
              it.props = (0, util_2.evaluatedPropsToName)(gen, it.props);
          }
          const { props } = it;
          validatePatternProperties();
          function validatePatternProperties() {
              for (const pat of patterns) {
                  if (checkProperties)
                      checkMatchingProperties(pat);
                  if (it.allErrors) {
                      validateProperties(pat);
                  }
                  else {
                      gen.var(valid, true); // TODO var
                      validateProperties(pat);
                      gen.if(valid);
                  }
              }
          }
          function checkMatchingProperties(pat) {
              for (const prop in checkProperties) {
                  if (new RegExp(pat).test(prop)) {
                      (0, util_1$6.checkStrictMode)(it, `property ${prop} matches pattern ${pat} (use allowMatchingProperties)`);
                  }
              }
          }
          function validateProperties(pat) {
              gen.forIn("key", data, (key) => {
                  gen.if((0, codegen_1$4._) `${(0, code_1$1.usePattern)(cxt, pat)}.test(${key})`, () => {
                      const alwaysValid = alwaysValidPatterns.includes(pat);
                      if (!alwaysValid) {
                          cxt.subschema({
                              keyword: "patternProperties",
                              schemaProp: pat,
                              dataProp: key,
                              dataPropType: util_2.Type.Str,
                          }, valid);
                      }
                      if (it.opts.unevaluated && props !== true) {
                          gen.assign((0, codegen_1$4._) `${props}[${key}]`, true);
                      }
                      else if (!alwaysValid && !it.allErrors) {
                          // can short-circuit if `unevaluatedProperties` is not supported (opts.next === false)
                          // or if all properties were evaluated (props === true)
                          gen.if((0, codegen_1$4.not)(valid), () => gen.break());
                      }
                  });
              });
          }
      },
  };
  patternProperties.default = def$8;

  var not = {};

  Object.defineProperty(not, "__esModule", { value: true });
  const util_1$5 = util;
  const def$7 = {
      keyword: "not",
      schemaType: ["object", "boolean"],
      trackErrors: true,
      code(cxt) {
          const { gen, schema, it } = cxt;
          if ((0, util_1$5.alwaysValidSchema)(it, schema)) {
              cxt.fail();
              return;
          }
          const valid = gen.name("valid");
          cxt.subschema({
              keyword: "not",
              compositeRule: true,
              createErrors: false,
              allErrors: false,
          }, valid);
          cxt.failResult(valid, () => cxt.reset(), () => cxt.error());
      },
      error: { message: "must NOT be valid" },
  };
  not.default = def$7;

  var anyOf = {};

  Object.defineProperty(anyOf, "__esModule", { value: true });
  const code_1 = code;
  const def$6 = {
      keyword: "anyOf",
      schemaType: "array",
      trackErrors: true,
      code: code_1.validateUnion,
      error: { message: "must match a schema in anyOf" },
  };
  anyOf.default = def$6;

  var oneOf = {};

  Object.defineProperty(oneOf, "__esModule", { value: true });
  const codegen_1$3 = codegen;
  const util_1$4 = util;
  const error$3 = {
      message: "must match exactly one schema in oneOf",
      params: ({ params }) => (0, codegen_1$3._) `{passingSchemas: ${params.passing}}`,
  };
  const def$5 = {
      keyword: "oneOf",
      schemaType: "array",
      trackErrors: true,
      error: error$3,
      code(cxt) {
          const { gen, schema, parentSchema, it } = cxt;
          /* istanbul ignore if */
          if (!Array.isArray(schema))
              throw new Error("ajv implementation error");
          if (it.opts.discriminator && parentSchema.discriminator)
              return;
          const schArr = schema;
          const valid = gen.let("valid", false);
          const passing = gen.let("passing", null);
          const schValid = gen.name("_valid");
          cxt.setParams({ passing });
          // TODO possibly fail straight away (with warning or exception) if there are two empty always valid schemas
          gen.block(validateOneOf);
          cxt.result(valid, () => cxt.reset(), () => cxt.error(true));
          function validateOneOf() {
              schArr.forEach((sch, i) => {
                  let schCxt;
                  if ((0, util_1$4.alwaysValidSchema)(it, sch)) {
                      gen.var(schValid, true);
                  }
                  else {
                      schCxt = cxt.subschema({
                          keyword: "oneOf",
                          schemaProp: i,
                          compositeRule: true,
                      }, schValid);
                  }
                  if (i > 0) {
                      gen
                          .if((0, codegen_1$3._) `${schValid} && ${valid}`)
                          .assign(valid, false)
                          .assign(passing, (0, codegen_1$3._) `[${passing}, ${i}]`)
                          .else();
                  }
                  gen.if(schValid, () => {
                      gen.assign(valid, true);
                      gen.assign(passing, i);
                      if (schCxt)
                          cxt.mergeEvaluated(schCxt, codegen_1$3.Name);
                  });
              });
          }
      },
  };
  oneOf.default = def$5;

  var allOf = {};

  Object.defineProperty(allOf, "__esModule", { value: true });
  const util_1$3 = util;
  const def$4 = {
      keyword: "allOf",
      schemaType: "array",
      code(cxt) {
          const { gen, schema, it } = cxt;
          /* istanbul ignore if */
          if (!Array.isArray(schema))
              throw new Error("ajv implementation error");
          const valid = gen.name("valid");
          schema.forEach((sch, i) => {
              if ((0, util_1$3.alwaysValidSchema)(it, sch))
                  return;
              const schCxt = cxt.subschema({ keyword: "allOf", schemaProp: i }, valid);
              cxt.ok(valid);
              cxt.mergeEvaluated(schCxt);
          });
      },
  };
  allOf.default = def$4;

  var _if = {};

  Object.defineProperty(_if, "__esModule", { value: true });
  const codegen_1$2 = codegen;
  const util_1$2 = util;
  const error$2 = {
      message: ({ params }) => (0, codegen_1$2.str) `must match "${params.ifClause}" schema`,
      params: ({ params }) => (0, codegen_1$2._) `{failingKeyword: ${params.ifClause}}`,
  };
  const def$3 = {
      keyword: "if",
      schemaType: ["object", "boolean"],
      trackErrors: true,
      error: error$2,
      code(cxt) {
          const { gen, parentSchema, it } = cxt;
          if (parentSchema.then === undefined && parentSchema.else === undefined) {
              (0, util_1$2.checkStrictMode)(it, '"if" without "then" and "else" is ignored');
          }
          const hasThen = hasSchema(it, "then");
          const hasElse = hasSchema(it, "else");
          if (!hasThen && !hasElse)
              return;
          const valid = gen.let("valid", true);
          const schValid = gen.name("_valid");
          validateIf();
          cxt.reset();
          if (hasThen && hasElse) {
              const ifClause = gen.let("ifClause");
              cxt.setParams({ ifClause });
              gen.if(schValid, validateClause("then", ifClause), validateClause("else", ifClause));
          }
          else if (hasThen) {
              gen.if(schValid, validateClause("then"));
          }
          else {
              gen.if((0, codegen_1$2.not)(schValid), validateClause("else"));
          }
          cxt.pass(valid, () => cxt.error(true));
          function validateIf() {
              const schCxt = cxt.subschema({
                  keyword: "if",
                  compositeRule: true,
                  createErrors: false,
                  allErrors: false,
              }, schValid);
              cxt.mergeEvaluated(schCxt);
          }
          function validateClause(keyword, ifClause) {
              return () => {
                  const schCxt = cxt.subschema({ keyword }, schValid);
                  gen.assign(valid, schValid);
                  cxt.mergeValidEvaluated(schCxt, valid);
                  if (ifClause)
                      gen.assign(ifClause, (0, codegen_1$2._) `${keyword}`);
                  else
                      cxt.setParams({ ifClause: keyword });
              };
          }
      },
  };
  function hasSchema(it, keyword) {
      const schema = it.schema[keyword];
      return schema !== undefined && !(0, util_1$2.alwaysValidSchema)(it, schema);
  }
  _if.default = def$3;

  var thenElse = {};

  Object.defineProperty(thenElse, "__esModule", { value: true });
  const util_1$1 = util;
  const def$2 = {
      keyword: ["then", "else"],
      schemaType: ["object", "boolean"],
      code({ keyword, parentSchema, it }) {
          if (parentSchema.if === undefined)
              (0, util_1$1.checkStrictMode)(it, `"${keyword}" without "if" is ignored`);
      },
  };
  thenElse.default = def$2;

  Object.defineProperty(applicator, "__esModule", { value: true });
  const additionalItems_1 = additionalItems;
  const prefixItems_1 = prefixItems;
  const items_1 = items;
  const items2020_1 = items2020;
  const contains_1 = contains;
  const dependencies_1 = dependencies;
  const propertyNames_1 = propertyNames;
  const additionalProperties_1 = additionalProperties;
  const properties_1 = properties$1;
  const patternProperties_1 = patternProperties;
  const not_1 = not;
  const anyOf_1 = anyOf;
  const oneOf_1 = oneOf;
  const allOf_1 = allOf;
  const if_1 = _if;
  const thenElse_1 = thenElse;
  function getApplicator(draft2020 = false) {
      const applicator = [
          // any
          not_1.default,
          anyOf_1.default,
          oneOf_1.default,
          allOf_1.default,
          if_1.default,
          thenElse_1.default,
          // object
          propertyNames_1.default,
          additionalProperties_1.default,
          dependencies_1.default,
          properties_1.default,
          patternProperties_1.default,
      ];
      // array
      if (draft2020)
          applicator.push(prefixItems_1.default, items2020_1.default);
      else
          applicator.push(additionalItems_1.default, items_1.default);
      applicator.push(contains_1.default);
      return applicator;
  }
  applicator.default = getApplicator;

  var format$2 = {};

  var format$1 = {};

  Object.defineProperty(format$1, "__esModule", { value: true });
  const codegen_1$1 = codegen;
  const error$1 = {
      message: ({ schemaCode }) => (0, codegen_1$1.str) `must match format "${schemaCode}"`,
      params: ({ schemaCode }) => (0, codegen_1$1._) `{format: ${schemaCode}}`,
  };
  const def$1 = {
      keyword: "format",
      type: ["number", "string"],
      schemaType: "string",
      $data: true,
      error: error$1,
      code(cxt, ruleType) {
          const { gen, data, $data, schema, schemaCode, it } = cxt;
          const { opts, errSchemaPath, schemaEnv, self } = it;
          if (!opts.validateFormats)
              return;
          if ($data)
              validate$DataFormat();
          else
              validateFormat();
          function validate$DataFormat() {
              const fmts = gen.scopeValue("formats", {
                  ref: self.formats,
                  code: opts.code.formats,
              });
              const fDef = gen.const("fDef", (0, codegen_1$1._) `${fmts}[${schemaCode}]`);
              const fType = gen.let("fType");
              const format = gen.let("format");
              // TODO simplify
              gen.if((0, codegen_1$1._) `typeof ${fDef} == "object" && !(${fDef} instanceof RegExp)`, () => gen.assign(fType, (0, codegen_1$1._) `${fDef}.type || "string"`).assign(format, (0, codegen_1$1._) `${fDef}.validate`), () => gen.assign(fType, (0, codegen_1$1._) `"string"`).assign(format, fDef));
              cxt.fail$data((0, codegen_1$1.or)(unknownFmt(), invalidFmt()));
              function unknownFmt() {
                  if (opts.strictSchema === false)
                      return codegen_1$1.nil;
                  return (0, codegen_1$1._) `${schemaCode} && !${format}`;
              }
              function invalidFmt() {
                  const callFormat = schemaEnv.$async
                      ? (0, codegen_1$1._) `(${fDef}.async ? await ${format}(${data}) : ${format}(${data}))`
                      : (0, codegen_1$1._) `${format}(${data})`;
                  const validData = (0, codegen_1$1._) `(typeof ${format} == "function" ? ${callFormat} : ${format}.test(${data}))`;
                  return (0, codegen_1$1._) `${format} && ${format} !== true && ${fType} === ${ruleType} && !${validData}`;
              }
          }
          function validateFormat() {
              const formatDef = self.formats[schema];
              if (!formatDef) {
                  unknownFormat();
                  return;
              }
              if (formatDef === true)
                  return;
              const [fmtType, format, fmtRef] = getFormat(formatDef);
              if (fmtType === ruleType)
                  cxt.pass(validCondition());
              function unknownFormat() {
                  if (opts.strictSchema === false) {
                      self.logger.warn(unknownMsg());
                      return;
                  }
                  throw new Error(unknownMsg());
                  function unknownMsg() {
                      return `unknown format "${schema}" ignored in schema at path "${errSchemaPath}"`;
                  }
              }
              function getFormat(fmtDef) {
                  const code = fmtDef instanceof RegExp
                      ? (0, codegen_1$1.regexpCode)(fmtDef)
                      : opts.code.formats
                          ? (0, codegen_1$1._) `${opts.code.formats}${(0, codegen_1$1.getProperty)(schema)}`
                          : undefined;
                  const fmt = gen.scopeValue("formats", { key: schema, ref: fmtDef, code });
                  if (typeof fmtDef == "object" && !(fmtDef instanceof RegExp)) {
                      return [fmtDef.type || "string", fmtDef.validate, (0, codegen_1$1._) `${fmt}.validate`];
                  }
                  return ["string", fmtDef, fmt];
              }
              function validCondition() {
                  if (typeof formatDef == "object" && !(formatDef instanceof RegExp) && formatDef.async) {
                      if (!schemaEnv.$async)
                          throw new Error("async format in sync schema");
                      return (0, codegen_1$1._) `await ${fmtRef}(${data})`;
                  }
                  return typeof format == "function" ? (0, codegen_1$1._) `${fmtRef}(${data})` : (0, codegen_1$1._) `${fmtRef}.test(${data})`;
              }
          }
      },
  };
  format$1.default = def$1;

  Object.defineProperty(format$2, "__esModule", { value: true });
  const format_1$1 = format$1;
  const format = [format_1$1.default];
  format$2.default = format;

  var metadata = {};

  Object.defineProperty(metadata, "__esModule", { value: true });
  metadata.contentVocabulary = metadata.metadataVocabulary = void 0;
  metadata.metadataVocabulary = [
      "title",
      "description",
      "default",
      "deprecated",
      "readOnly",
      "writeOnly",
      "examples",
  ];
  metadata.contentVocabulary = [
      "contentMediaType",
      "contentEncoding",
      "contentSchema",
  ];

  Object.defineProperty(draft7, "__esModule", { value: true });
  const core_1 = core$1;
  const validation_1 = validation$1;
  const applicator_1 = applicator;
  const format_1 = format$2;
  const metadata_1 = metadata;
  const draft7Vocabularies = [
      core_1.default,
      validation_1.default,
      (0, applicator_1.default)(),
      format_1.default,
      metadata_1.metadataVocabulary,
      metadata_1.contentVocabulary,
  ];
  draft7.default = draft7Vocabularies;

  var discriminator = {};

  var types = {};

  (function (exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.DiscrError = void 0;
  (function (DiscrError) {
      DiscrError["Tag"] = "tag";
      DiscrError["Mapping"] = "mapping";
  })(exports.DiscrError || (exports.DiscrError = {}));

  }(types));

  Object.defineProperty(discriminator, "__esModule", { value: true });
  const codegen_1 = codegen;
  const types_1 = types;
  const compile_1 = compile;
  const util_1 = util;
  const error = {
      message: ({ params: { discrError, tagName } }) => discrError === types_1.DiscrError.Tag
          ? `tag "${tagName}" must be string`
          : `value of tag "${tagName}" must be in oneOf`,
      params: ({ params: { discrError, tag, tagName } }) => (0, codegen_1._) `{error: ${discrError}, tag: ${tagName}, tagValue: ${tag}}`,
  };
  const def = {
      keyword: "discriminator",
      type: "object",
      schemaType: "object",
      error,
      code(cxt) {
          const { gen, data, schema, parentSchema, it } = cxt;
          const { oneOf } = parentSchema;
          if (!it.opts.discriminator) {
              throw new Error("discriminator: requires discriminator option");
          }
          const tagName = schema.propertyName;
          if (typeof tagName != "string")
              throw new Error("discriminator: requires propertyName");
          if (schema.mapping)
              throw new Error("discriminator: mapping is not supported");
          if (!oneOf)
              throw new Error("discriminator: requires oneOf keyword");
          const valid = gen.let("valid", false);
          const tag = gen.const("tag", (0, codegen_1._) `${data}${(0, codegen_1.getProperty)(tagName)}`);
          gen.if((0, codegen_1._) `typeof ${tag} == "string"`, () => validateMapping(), () => cxt.error(false, { discrError: types_1.DiscrError.Tag, tag, tagName }));
          cxt.ok(valid);
          function validateMapping() {
              const mapping = getMapping();
              gen.if(false);
              for (const tagValue in mapping) {
                  gen.elseIf((0, codegen_1._) `${tag} === ${tagValue}`);
                  gen.assign(valid, applyTagSchema(mapping[tagValue]));
              }
              gen.else();
              cxt.error(false, { discrError: types_1.DiscrError.Mapping, tag, tagName });
              gen.endIf();
          }
          function applyTagSchema(schemaProp) {
              const _valid = gen.name("valid");
              const schCxt = cxt.subschema({ keyword: "oneOf", schemaProp }, _valid);
              cxt.mergeEvaluated(schCxt, codegen_1.Name);
              return _valid;
          }
          function getMapping() {
              var _a;
              const oneOfMapping = {};
              const topRequired = hasRequired(parentSchema);
              let tagRequired = true;
              for (let i = 0; i < oneOf.length; i++) {
                  let sch = oneOf[i];
                  if ((sch === null || sch === void 0 ? void 0 : sch.$ref) && !(0, util_1.schemaHasRulesButRef)(sch, it.self.RULES)) {
                      sch = compile_1.resolveRef.call(it.self, it.schemaEnv.root, it.baseId, sch === null || sch === void 0 ? void 0 : sch.$ref);
                      if (sch instanceof compile_1.SchemaEnv)
                          sch = sch.schema;
                  }
                  const propSch = (_a = sch === null || sch === void 0 ? void 0 : sch.properties) === null || _a === void 0 ? void 0 : _a[tagName];
                  if (typeof propSch != "object") {
                      throw new Error(`discriminator: oneOf subschemas (or referenced schemas) must have "properties/${tagName}"`);
                  }
                  tagRequired = tagRequired && (topRequired || hasRequired(sch));
                  addMappings(propSch, i);
              }
              if (!tagRequired)
                  throw new Error(`discriminator: "${tagName}" must be required`);
              return oneOfMapping;
              function hasRequired({ required }) {
                  return Array.isArray(required) && required.includes(tagName);
              }
              function addMappings(sch, i) {
                  if (sch.const) {
                      addMapping(sch.const, i);
                  }
                  else if (sch.enum) {
                      for (const tagValue of sch.enum) {
                          addMapping(tagValue, i);
                      }
                  }
                  else {
                      throw new Error(`discriminator: "properties/${tagName}" must have "const" or "enum"`);
                  }
              }
              function addMapping(tagValue, i) {
                  if (typeof tagValue != "string" || tagValue in oneOfMapping) {
                      throw new Error(`discriminator: "${tagName}" values must be unique strings`);
                  }
                  oneOfMapping[tagValue] = i;
              }
          }
      },
  };
  discriminator.default = def;

  var $schema = "http://json-schema.org/draft-07/schema#";
  var $id = "http://json-schema.org/draft-07/schema#";
  var title = "Core schema meta-schema";
  var definitions = {
  	schemaArray: {
  		type: "array",
  		minItems: 1,
  		items: {
  			$ref: "#"
  		}
  	},
  	nonNegativeInteger: {
  		type: "integer",
  		minimum: 0
  	},
  	nonNegativeIntegerDefault0: {
  		allOf: [
  			{
  				$ref: "#/definitions/nonNegativeInteger"
  			},
  			{
  				"default": 0
  			}
  		]
  	},
  	simpleTypes: {
  		"enum": [
  			"array",
  			"boolean",
  			"integer",
  			"null",
  			"number",
  			"object",
  			"string"
  		]
  	},
  	stringArray: {
  		type: "array",
  		items: {
  			type: "string"
  		},
  		uniqueItems: true,
  		"default": [
  		]
  	}
  };
  var type = [
  	"object",
  	"boolean"
  ];
  var properties = {
  	$id: {
  		type: "string",
  		format: "uri-reference"
  	},
  	$schema: {
  		type: "string",
  		format: "uri"
  	},
  	$ref: {
  		type: "string",
  		format: "uri-reference"
  	},
  	$comment: {
  		type: "string"
  	},
  	title: {
  		type: "string"
  	},
  	description: {
  		type: "string"
  	},
  	"default": true,
  	readOnly: {
  		type: "boolean",
  		"default": false
  	},
  	examples: {
  		type: "array",
  		items: true
  	},
  	multipleOf: {
  		type: "number",
  		exclusiveMinimum: 0
  	},
  	maximum: {
  		type: "number"
  	},
  	exclusiveMaximum: {
  		type: "number"
  	},
  	minimum: {
  		type: "number"
  	},
  	exclusiveMinimum: {
  		type: "number"
  	},
  	maxLength: {
  		$ref: "#/definitions/nonNegativeInteger"
  	},
  	minLength: {
  		$ref: "#/definitions/nonNegativeIntegerDefault0"
  	},
  	pattern: {
  		type: "string",
  		format: "regex"
  	},
  	additionalItems: {
  		$ref: "#"
  	},
  	items: {
  		anyOf: [
  			{
  				$ref: "#"
  			},
  			{
  				$ref: "#/definitions/schemaArray"
  			}
  		],
  		"default": true
  	},
  	maxItems: {
  		$ref: "#/definitions/nonNegativeInteger"
  	},
  	minItems: {
  		$ref: "#/definitions/nonNegativeIntegerDefault0"
  	},
  	uniqueItems: {
  		type: "boolean",
  		"default": false
  	},
  	contains: {
  		$ref: "#"
  	},
  	maxProperties: {
  		$ref: "#/definitions/nonNegativeInteger"
  	},
  	minProperties: {
  		$ref: "#/definitions/nonNegativeIntegerDefault0"
  	},
  	required: {
  		$ref: "#/definitions/stringArray"
  	},
  	additionalProperties: {
  		$ref: "#"
  	},
  	definitions: {
  		type: "object",
  		additionalProperties: {
  			$ref: "#"
  		},
  		"default": {
  		}
  	},
  	properties: {
  		type: "object",
  		additionalProperties: {
  			$ref: "#"
  		},
  		"default": {
  		}
  	},
  	patternProperties: {
  		type: "object",
  		additionalProperties: {
  			$ref: "#"
  		},
  		propertyNames: {
  			format: "regex"
  		},
  		"default": {
  		}
  	},
  	dependencies: {
  		type: "object",
  		additionalProperties: {
  			anyOf: [
  				{
  					$ref: "#"
  				},
  				{
  					$ref: "#/definitions/stringArray"
  				}
  			]
  		}
  	},
  	propertyNames: {
  		$ref: "#"
  	},
  	"const": true,
  	"enum": {
  		type: "array",
  		items: true,
  		minItems: 1,
  		uniqueItems: true
  	},
  	type: {
  		anyOf: [
  			{
  				$ref: "#/definitions/simpleTypes"
  			},
  			{
  				type: "array",
  				items: {
  					$ref: "#/definitions/simpleTypes"
  				},
  				minItems: 1,
  				uniqueItems: true
  			}
  		]
  	},
  	format: {
  		type: "string"
  	},
  	contentMediaType: {
  		type: "string"
  	},
  	contentEncoding: {
  		type: "string"
  	},
  	"if": {
  		$ref: "#"
  	},
  	then: {
  		$ref: "#"
  	},
  	"else": {
  		$ref: "#"
  	},
  	allOf: {
  		$ref: "#/definitions/schemaArray"
  	},
  	anyOf: {
  		$ref: "#/definitions/schemaArray"
  	},
  	oneOf: {
  		$ref: "#/definitions/schemaArray"
  	},
  	not: {
  		$ref: "#"
  	}
  };
  var require$$3 = {
  	$schema: $schema,
  	$id: $id,
  	title: title,
  	definitions: definitions,
  	type: type,
  	properties: properties,
  	"default": true
  };

  (function (module, exports) {
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.CodeGen = exports.Name = exports.nil = exports.stringify = exports.str = exports._ = exports.KeywordCxt = void 0;
  const core_1 = core$2;
  const draft7_1 = draft7;
  const discriminator_1 = discriminator;
  const draft7MetaSchema = require$$3;
  const META_SUPPORT_DATA = ["/properties"];
  const META_SCHEMA_ID = "http://json-schema.org/draft-07/schema";
  class Ajv extends core_1.default {
      _addVocabularies() {
          super._addVocabularies();
          draft7_1.default.forEach((v) => this.addVocabulary(v));
          if (this.opts.discriminator)
              this.addKeyword(discriminator_1.default);
      }
      _addDefaultMetaSchema() {
          super._addDefaultMetaSchema();
          if (!this.opts.meta)
              return;
          const metaSchema = this.opts.$data
              ? this.$dataMetaSchema(draft7MetaSchema, META_SUPPORT_DATA)
              : draft7MetaSchema;
          this.addMetaSchema(metaSchema, META_SCHEMA_ID, false);
          this.refs["http://json-schema.org/schema"] = META_SCHEMA_ID;
      }
      defaultMeta() {
          return (this.opts.defaultMeta =
              super.defaultMeta() || (this.getSchema(META_SCHEMA_ID) ? META_SCHEMA_ID : undefined));
      }
  }
  module.exports = exports = Ajv;
  Object.defineProperty(exports, "__esModule", { value: true });
  exports.default = Ajv;
  var validate_1 = validate$1;
  Object.defineProperty(exports, "KeywordCxt", { enumerable: true, get: function () { return validate_1.KeywordCxt; } });
  var codegen_1 = codegen;
  Object.defineProperty(exports, "_", { enumerable: true, get: function () { return codegen_1._; } });
  Object.defineProperty(exports, "str", { enumerable: true, get: function () { return codegen_1.str; } });
  Object.defineProperty(exports, "stringify", { enumerable: true, get: function () { return codegen_1.stringify; } });
  Object.defineProperty(exports, "nil", { enumerable: true, get: function () { return codegen_1.nil; } });
  Object.defineProperty(exports, "Name", { enumerable: true, get: function () { return codegen_1.Name; } });
  Object.defineProperty(exports, "CodeGen", { enumerable: true, get: function () { return codegen_1.CodeGen; } });

  }(ajv$1, ajv$1.exports));

  var Ajv = /*@__PURE__*/getDefaultExportFromCjs(ajv$1.exports);

  var ajv = new Ajv({
    strict: false
  }); // 校验器 key

  var VALIDATOR = 'validator';
  var OWN_RULE_PROPERTY = 'ui:rules';
  var DEBUG_PREFIX = '[bk-schema-form-validator]';

  var throwErr = function throwErr(err) {
    throw new Error("".concat(DEBUG_PREFIX, " ").concat(err));
  };

  var globalRules = new Map();
  /**
   * 注册全局校验规则
   * @param rules
   */

  var registryGlobalRules = function registryGlobalRules(rules) {
    try {
      if (!rules) return;
      if ((rules === null || rules === void 0 ? void 0 : rules.constructor) !== Object) throwErr('global rules must be an object');
      Object.keys(rules).forEach(function (ruleName) {
        var rule = rules[ruleName];
        if (!hasOwnProperty(rule, VALIDATOR)) throwErr("'".concat(ruleName, "' rule must have a validator property"));
        var validator = rule[VALIDATOR];
        if (!(isExpression(validator) || isRegExp(validator) || (validator === null || validator === void 0 ? void 0 : validator.constructor) === Function)) throwErr("'".concat(ruleName, "' must be one of expression or regexp or function"));
        globalRules.set(ruleName, rule);
      });
    } catch (error) {
      throwErr(error);
    }
  };
  var formItems = new Map();
  var registryFormItems = function registryFormItems(path, instance) {
    formItems.set(path, {
      instance: instance
    });
  };
  /**
   * 校验规则
   */

  var validate = function validate(rule, instance) {
    var theRule = rule;

    if (typeof rule === 'string') {
      theRule = globalRules.get(rule);
      if (!theRule) throwErr("'".concat(rule, " is not a valid global rule, you can registry it to global rules node or use form item own custom rules"));
    }

    var _theRule = theRule,
        validator = _theRule.validator,
        message = _theRule.message;
    var valid = true;

    if (isExpression(validator)) {
      valid = executeExpression(validator, instance);
    } else if (validator.constructor === Function) {
      valid = validator(instance);
    }

    return {
      valid: valid,
      message: message
    };
  };
  /**
   * 校验单个表单项
   * @param path 字段路径
   */


  var validateFormItem = function validateFormItem(path) {
    var _instance$schema;

    var formItem = formItems.get(path);
    if (!formItem) return true;
    var instance = formItem.instance;
    var ownSchema = instance.schema; // json schema validate

    var schemaValidate = ajv.compile(ownSchema);
    var schemaValid = schemaValidate(instance.value);

    if (!schemaValid) {
      var _schemaValidate$error;

      instance.setState('error', true);
      instance.setErrorTips((_schemaValidate$error = schemaValidate.errors) === null || _schemaValidate$error === void 0 ? void 0 : _schemaValidate$error.map(function (err) {
        return err.message;
      }));
      return false;
    }

    var customRules = (_instance$schema = instance.schema) === null || _instance$schema === void 0 ? void 0 : _instance$schema[OWN_RULE_PROPERTY]; // 自定义规则校验

    if (!customRules) return true;
    var isError = false;
    var errorMsg = '';

    var _iterator = _createForOfIteratorHelper(customRules),
        _step;

    try {
      for (_iterator.s(); !(_step = _iterator.n()).done;) {
        var rule = _step.value;
        var result = validate(rule, instance);

        if (!result.valid) {
          isError = true;
          errorMsg = result.message;
          break;
        }
      }
    } catch (err) {
      _iterator.e(err);
    } finally {
      _iterator.f();
    }

    instance.setState('error', isError);
    instance.setErrorTips(errorMsg);
    return !isError;
  };
  /**
   * 校验整个表单
   */

  var validateForm = function validateForm() {
    var isValid = true;
    formItems.forEach(function (value, path) {
      if (!validateFormItem(path)) isValid = false;
    });
    return isValid;
  };
  /**
   * 触发校验
   * @param path 字段路径
   */

  var dispatchValidate = function dispatchValidate(path) {
    return validateFormItem(path);
  };

  // 事件订阅器
  var FormEvent = /*#__PURE__*/function () {
    function FormEvent() {
      _classCallCheck(this, FormEvent);

      this.callbacks = void 0;
      this.callbacks = Object.create(null);
    }

    _createClass(FormEvent, [{
      key: "on",
      value: function on(path, type, cb) {
        if (!(path in this.callbacks)) {
          this.callbacks[path] = {};
        }

        if (!(type in this.callbacks[path])) {
          this.callbacks[path][type] = [];
        }

        this.callbacks[path][type].push(cb);
        return this;
      }
    }, {
      key: "off",
      value: function off(path, type, cb) {
        if (!(path && type)) {
          // 参数全部为空，清空callbacks
          this.callbacks = Object.create(null);
        } else if (path && !type) {
          // 清空当前path所有事件
          delete this.callbacks[path];
        } else if (path && type && !cb) {
          // 清空当前type的所有事件
          delete this.callbacks[path][type];
        } else {
          // 清空对应事件
          var events = this.callbacks[path][type]; // eslint-disable-next-line no-restricted-syntax

          for (var index in events) {
            if (cb === events[index]) {
              events.splice(Number(index), 1);
            }
          }
        }

        return this;
      }
    }, {
      key: "once",
      value: function once(path, type, cb) {
        var _this = this;

        function innerOnce() {
          cb.apply(void 0, arguments);

          _this.off(path, type, innerOnce);
        }

        innerOnce.fn = cb;
        this.on(path, type, innerOnce);
        return this;
      }
    }, {
      key: "emit",
      value: function emit(path, type) {
        if (!this.callbacks[path]) return;

        if (type in this.callbacks[path]) {
          var runs = _toConsumableArray(this.callbacks[path][type]);

          for (var _len = arguments.length, arg = new Array(_len > 2 ? _len - 2 : 0), _key = 2; _key < _len; _key++) {
            arg[_key - 2] = arguments[_key];
          }

          var _iterator = _createForOfIteratorHelper(runs),
              _step;

          try {
            for (_iterator.s(); !(_step = _iterator.n()).done;) {
              var cb = _step.value;
              cb.apply(void 0, arg);
            }
          } catch (err) {
            _iterator.e(err);
          } finally {
            _iterator.f();
          }
        }
      }
    }]);

    return FormEvent;
  }();
  var events = new FormEvent();

  var _excluded$2 = ["url", "params"],
      _excluded2 = ["name"];
  var Widget = Vue__default["default"].extend({
    name: 'Widget',
    props: props,
    data: function data() {
      return {
        loading: false,
        datasource: Schema.resolveDefaultDatasource(this.schema),
        formItemProps: {},
        state: {
          visible: true,
          disabled: false,
          readonly: false,
          error: false
        },
        errorTips: ''
      };
    },
    computed: {
      widgetNode: function widgetNode() {
        return widgetTree.widgetMap[this.path];
      }
    },
    watch: {
      value: {
        handler: function handler(newValue, oldValue) {
          var _this = this;

          if (!deepEquals(newValue, oldValue)) {
            setTimeout(function () {
              reactionDispatch(_this.path, 'valChange');
              dispatchValidate(_this.path);
            }, 0);
          }
        }
      }
    },
    created: function created() {
      var _this2 = this;

      // 表单项配置
      var uiOptions = Schema.getUiOptions(this.schema);
      this.formItemProps = _objectSpread2(_objectSpread2({}, uiOptions), {}, {
        // schema配置不存在title时默认用属性名作为title
        label: uiOptions.showTitle ? uiOptions.label || Path.getPathLastProp(this.path) : '',
        required: this.required
      }); // 设置widget初始化状态 ui:component优先级 > ui:props优先级

      var vNodeData = Schema.getUiComponent(this.schema);
      var defaultProps = Object.assign({}, this.formItemProps, vNodeData.props || {});
      Object.keys(defaultProps).forEach(function (key) {
        if (Reflect.has(_this2.state, key)) {
          _this2.setState(key, defaultProps[key]);
        }
      }); // 注册widget TreeNode

      widgetTree.addWidgetNode(this.path, this);
      registryFormItems(this.path, this);
    },
    mounted: function mounted() {
      // 注册联动
      reactionRegister(this.path, this.schema['ui:reactions']); // 首次联动

      reactionDispatch(this.path, 'valChange');
      reactionDispatch(this.path, 'lifetime/init');
    },
    beforeDestroy: function beforeDestroy() {
      widgetTree.removeWidgetNode(this.path, this);
    },
    methods: {
      setState: function setState(key, value) {
        if (Reflect.has(this.state, key)) {
          this.state[key] = value;
        } else if (key === 'value') {
          // 特殊处理value设置
          this.$emit('input', {
            path: this.path,
            value: value
          });
        } else {
          console.warn("Unsupported ".concat(key, " state, please check"));
        }
      },
      loadDataSource: function loadDataSource() {
        var _this3 = this;

        return _asyncToGenerator( /*#__PURE__*/regeneratorRuntime.mark(function _callee() {
          var _this3$schema, _this3$schema$uiComp, _this3$schema$uiComp$;

          var xhrConfig, url, params, reset, _this3$httpAdapter, _this3$httpAdapter$re, http, responseParse, remoteURL, requestParams;

          return regeneratorRuntime.wrap(function _callee$(_context) {
            while (1) {
              switch (_context.prev = _context.next) {
                case 0:
                  xhrConfig = (_this3$schema = _this3.schema) === null || _this3$schema === void 0 ? void 0 : (_this3$schema$uiComp = _this3$schema['ui:component']) === null || _this3$schema$uiComp === void 0 ? void 0 : (_this3$schema$uiComp$ = _this3$schema$uiComp.props) === null || _this3$schema$uiComp$ === void 0 ? void 0 : _this3$schema$uiComp$.remoteConfig;

                  if (!xhrConfig) {
                    _context.next = 18;
                    break;
                  }

                  url = xhrConfig.url, params = xhrConfig.params, reset = _objectWithoutProperties(xhrConfig, _excluded$2);
                  _this3$httpAdapter = _this3.httpAdapter, _this3$httpAdapter$re = _this3$httpAdapter.request, http = _this3$httpAdapter$re === void 0 ? request : _this3$httpAdapter$re, responseParse = _this3$httpAdapter.responseParse;
                  _context.prev = 4;
                  _this3.loading = true;
                  remoteURL = executeExpression(url, _this3);
                  requestParams = isObj(params) ? executeExpression(params, _this3) : params;
                  _context.next = 10;
                  return http(remoteURL, _objectSpread2(_objectSpread2({}, reset), {}, {
                    params: requestParams,
                    responseParse: responseParse
                  }));

                case 10:
                  _this3.datasource = _context.sent;
                  _this3.loading = false;
                  _context.next = 18;
                  break;

                case 14:
                  _context.prev = 14;
                  _context.t0 = _context["catch"](4);
                  _this3.loading = false;
                  console.error(_context.t0);

                case 18:
                case "end":
                  return _context.stop();
              }
            }
          }, _callee, null, [[4, 14]]);
        }))();
      },
      setErrorTips: function setErrorTips(tips) {
        this.errorTips = tips;
      },
      getValue: function getValue(path) {
        return Path.getPathVal(this.rootData, path);
      }
    },
    render: function render(h) {
      var _events$callbacks, _this$$scopedSlots$de, _this$$scopedSlots$de2, _this$$scopedSlots, _this$$scopedSlots$su, _this$$scopedSlots2;

      var _Schema$getUiComponen = Schema.getUiComponent(this.schema),
          name = _Schema$getUiComponen.name,
          uiVnodeData = _objectWithoutProperties(_Schema$getUiComponen, _excluded2); // 注意顺序！！！


      var widgetProps = _objectSpread2(_objectSpread2({}, this.$props), {}, {
        loading: this.loading,
        value: this.value
      });

      var _self = this;

      var widgetName = registry.getComponent(name) || name || Schema.getDefaultWidget(this.schema);
      var widgetEvents = ((_events$callbacks = events.callbacks) === null || _events$callbacks === void 0 ? void 0 : _events$callbacks[this.path]) || {}; // 当前state属性优先级最高

      var renderWidget = (_this$$scopedSlots$de = (_this$$scopedSlots$de2 = (_this$$scopedSlots = this.$scopedSlots).default) === null || _this$$scopedSlots$de2 === void 0 ? void 0 : _this$$scopedSlots$de2.call(_this$$scopedSlots, {
        path: this.path
      })) !== null && _this$$scopedSlots$de !== void 0 ? _this$$scopedSlots$de : h(widgetName, mergeDeep({
        props: _objectSpread2({}, widgetProps),
        attrs: _objectSpread2({}, uiVnodeData.props || {}),
        on: _objectSpread2(_objectSpread2({}, widgetEvents), {}, {
          input: [].concat(_toConsumableArray(widgetEvents.input || []), [function (value) {
            // 所有组件widget必须实现input事件，用于v-model时更新表单数据
            _self.$emit('input', {
              path: _self.path,
              value: value
            });
          }]),
          click: function click() {
            reactionDispatch(_self.path, 'effect/click');
          }
        })
      }, mergeDeep(uiVnodeData, {
        props: _objectSpread2(_objectSpread2({}, this.state), {}, {
          datasource: this.datasource
        }),
        attrs: _objectSpread2({}, this.state)
      }))); // 渲染删除按钮（用于数组类型widget删除）

      var renderDelete = function renderDelete() {
        return h('span', {
          class: ['bk-schema-form-group-delete'],
          style: {
            right: '-20px',
            top: '0px'
          },
          on: {
            click: function click() {
              _self.$emit('remove', _self.path);
            }
          }
        }, [h('i', {
          class: ['bk-icon icon-close3-shape']
        })]);
      };

      return h(registry.getBaseWidget('form-item'), {
        props: this.formItemProps,
        style: _objectSpread2(_objectSpread2(_objectSpread2({}, this.layout.item || {}), this.layout.container || {}), {}, {
          // 表单项显示和隐藏状态
          display: this.state.visible ? '' : 'none'
        }),
        class: {
          'bk-schema-form-item--error': this.state.error
        }
      }, [renderWidget, (_this$$scopedSlots$su = (_this$$scopedSlots2 = this.$scopedSlots).suffix) === null || _this$$scopedSlots$su === void 0 ? void 0 : _this$$scopedSlots$su.call(_this$$scopedSlots2, {
        path: this.path,
        schema: this.schema
      }), this.removeable && renderDelete(), this.state.error ? h('p', {
        class: 'bk-schema-form-item__error-tips'
      }, this.errorTips) : null, this.formItemProps.tips ? h('p', {
        slot: 'tip',
        class: ['mt5', 'mb0', 'f12'],
        style: {
          color: '#5e6d82',
          lineHeight: '1.5em'
        }
      }, this.formItemProps.tips) : null]);
    }
  });

  var StringField = Vue__default["default"].extend({
    name: 'StringField',
    functional: true,
    props: props,
    render: function render(h, ctx) {
      return h(Widget, _objectSpread2({}, ctx.data));
    }
  });

  var NumberField = Vue__default["default"].extend({
    name: 'NumberField',
    functional: true,
    props: props,
    render: function render(h, ctx) {
      var _ctx$props$schema;

      return h(StringField, _objectSpread2(_objectSpread2({}, ctx.data), {}, {
        props: _objectSpread2(_objectSpread2({}, ctx.props), {}, {
          schema: mergeDeep({
            'ui:component': {
              props: {
                type: 'number',
                min: ((_ctx$props$schema = ctx.props.schema) === null || _ctx$props$schema === void 0 ? void 0 : _ctx$props$schema.type) === 'integer' ? 0 : -Infinity
              }
            }
          }, ctx.props.schema)
        }),
        on: _objectSpread2(_objectSpread2({}, ctx.data.on || {}), {}, {
          input: function input(data) {
            var _ctx$data, _ctx$data$on;

            // 解决input组件number类型在输入时会变成字符串问题
            if (typeof ((_ctx$data = ctx.data) === null || _ctx$data === void 0 ? void 0 : (_ctx$data$on = _ctx$data.on) === null || _ctx$data$on === void 0 ? void 0 : _ctx$data$on.input) === 'function') {
              var _ctx$data2, _ctx$data2$on;

              (_ctx$data2 = ctx.data) === null || _ctx$data2 === void 0 ? void 0 : (_ctx$data2$on = _ctx$data2.on) === null || _ctx$data2$on === void 0 ? void 0 : _ctx$data2$on.input(_objectSpread2(_objectSpread2({}, data), {}, {
                value: Number(data.value) || 0
              }));
            }
          }
        })
      }));
    }
  });

  var _excluded$1 = ["name"];

  var ArrayWidget = Vue__default["default"].extend({
    name: 'ArrayWidget',
    props: props,
    mounted: function mounted() {
      this.handleFillItem();
    },
    methods: {
      // 补全minItems项
      handleFillItem: function handleFillItem() {
        var _this$schema$minItems = this.schema.minItems,
            minItems = _this$schema$minItems === void 0 ? 0 : _this$schema$minItems;

        if (this.value.length < minItems) {
          var data = Schema.getSchemaDefaultValue(this.schema.items);
          var remainData = new Array(minItems - this.value.length).fill(data);
          this.$emit('input', {
            path: this.path,
            value: [].concat(_toConsumableArray(this.value), _toConsumableArray(remainData))
          });
        }
      },
      // 添加item
      handleAddItem: function handleAddItem() {
        var data = Schema.getSchemaDefaultValue(this.schema.items);
        var value = JSON.parse(JSON.stringify(this.value || []));
        value.push(data);
        this.$emit('input', {
          path: this.path,
          value: value
        });
      },
      // 删除item
      handleDeleteItem: function handleDeleteItem(path) {
        var index = Number(Path.getPathLastProp(path));
        var value = JSON.parse(JSON.stringify(this.value || []));
        value.splice(index, 1);
        this.$emit('input', {
          path: this.path,
          value: value
        });
      }
    },
    render: function render(h) {
      var _this = this;

      var _self = this;

      var arrVnodeList = (Array.isArray(this.value) ? this.value : []).map(function (_, index) {
        var curPath = Path.getCurPath(_this.path, index);
        return h(SchemaField, {
          key: curPath,
          props: _objectSpread2(_objectSpread2({}, _this.$props), {}, {
            schema: _this.schema.items,
            path: curPath,
            layout: _objectSpread2(_objectSpread2({}, _this.layout), {}, {
              item: {} // todo: 暂时不支持数组项之间的布局

            }),
            removeable: true
          }),
          on: _objectSpread2(_objectSpread2({}, _this.$listeners), {}, {
            remove: function remove(path) {
              _self.handleDeleteItem(path);
            }
          })
        });
      });

      var _Schema$getGroupWrap = Schema.getGroupWrap(this.schema),
          name = _Schema$getGroupWrap.name,
          vnode = _objectWithoutProperties(_Schema$getGroupWrap, _excluded$1);

      return h(name, mergeDeep({
        props: _objectSpread2(_objectSpread2({}, this.$props), {}, {
          layout: {},
          showTitle: true // 数组类型默认展示分组title

        }),
        style: _objectSpread2({}, this.layout.item || {})
      }, vnode), [].concat(_toConsumableArray(arrVnodeList), [h('span', {
        class: ['bk-schema-form-group-add'],
        on: {
          click: function click() {
            _self.handleAddItem();
          }
        }
      }, [h('i', {
        class: ['bk-icon icon-plus-circle-shape mr5']
      }), '添加'])]));
    }
  });

  var _excluded = ["name"];
  var ArrayField = Vue__default["default"].extend({
    name: 'ArrayField',
    functional: true,
    props: props,
    render: function render(h, ctx) {
      var _ctx$props = ctx.props,
          schema = _ctx$props.schema,
          path = _ctx$props.path;

      if (Schema.isMultiSelect(schema) || Schema.isCustomArrayWidget(schema)) {
        // 多选类型 或 自定义数组类型（伪数组类型(只有一个FormItem，但是值为数组，值一般由自定义Widget控件决定)）
        return h(Widget, _objectSpread2(_objectSpread2({}, ctx.data), {}, {
          key: path,
          props: _objectSpread2(_objectSpread2({}, ctx.props), {}, {
            schema: mergeDeep({
              'ui:component': {
                props: {
                  multiple: true
                }
              }
            }, ctx.props.schema)
          })
        }));
      } // 元组类型


      if (Schema.isTupleArray(schema)) {
        var _Schema$getGroupWrap = Schema.getGroupWrap(schema),
            name = _Schema$getGroupWrap.name,
            vnode = _objectWithoutProperties(_Schema$getGroupWrap, _excluded);

        var tupleVnodeList = schema.items.map(function (item, index) {
          return h(SchemaField, _objectSpread2(_objectSpread2({}, ctx.data), {}, {
            key: Path.getCurPath(path, index),
            props: _objectSpread2(_objectSpread2({}, ctx.props), {}, {
              schema: item,
              path: Path.getCurPath(path, index)
            })
          }));
        });
        return h(name, mergeDeep({
          props: _objectSpread2(_objectSpread2({}, ctx.props), {}, {
            path: path,
            showTitle: true
          })
        }, vnode), _toConsumableArray(tupleVnodeList));
      } // 一般数组类型


      return h(ArrayWidget, _objectSpread2({}, ctx.data));
    }
  });

  var BooleanField = Vue__default["default"].extend({
    name: 'BooleanField',
    functional: true,
    props: props,
    render: function render(h, ctx) {
      var _ref = ctx.props.schema['ui:component'] || {},
          name = _ref.name;

      var widgetProps = {};

      if (['radio', 'select'].includes(name)) {
        // radioGroup、select类型需要默认数据源
        widgetProps = {
          datasource: [{
            label: 'False',
            value: false
          }, {
            label: 'True',
            value: true
          }]
        };
      } else if (name === 'checkbox') {
        var _ctx$props$schema;

        // boolean 类型checkbox
        widgetProps = {
          label: (_ctx$props$schema = ctx.props.schema) === null || _ctx$props$schema === void 0 ? void 0 : _ctx$props$schema.title
        };
      }

      return h(Widget, _objectSpread2(_objectSpread2({}, ctx.data), {}, {
        props: _objectSpread2(_objectSpread2({}, ctx.props), {}, {
          schema: mergeDeep({
            'ui:component': {
              props: widgetProps
            }
          }, ctx.props.schema)
        })
      }));
    }
  });

  var CheckboxWidget = Vue__default["default"].extend({
    name: 'CheckboxWidget',
    props: {
      datasource: {
        type: Array,
        default: function _default() {
          return [];
        }
      },
      value: {
        type: [Array, Boolean],
        default: function _default() {
          return [];
        }
      },
      // 单个checkbox时文案
      label: {
        type: String,
        default: ''
      }
    },
    methods: {
      handleChange: function handleChange(val) {
        this.$emit('input', val);
      }
    },
    render: function render(h) {
      var _this = this;

      return Array.isArray(this.value) ? h(registry.getBaseWidget('checkbox-group'), {
        on: {
          change: this.handleChange
        }
      }, this.datasource.map(function (item) {
        return h(registry.getBaseWidget('checkbox'), {
          key: item.value,
          class: ['mr24'],
          props: _objectSpread2({
            value: item.value
          }, _this.$attrs)
        }, item.label);
      })) : h(registry.getBaseWidget('checkbox'), {
        props: {
          value: this.value
        },
        on: {
          change: this.handleChange
        }
      }, this.label);
    }
  });

  var SelectWidget = Vue__default["default"].extend({
    name: 'SelectWidget',
    props: {
      datasource: {
        type: Array,
        default: function _default() {
          return [];
        }
      },
      value: {
        type: [Array, String, Number, Boolean],
        default: ''
      },
      loading: {
        type: Boolean,
        default: false
      }
    },
    methods: {
      handleSelectChange: function handleSelectChange(val) {
        this.$emit('input', val);
      }
    },
    render: function render(h) {
      return h(registry.getBaseWidget('select'), {
        props: _objectSpread2({
          loading: this.loading,
          value: this.value
        }, this.$attrs),
        on: {
          change: this.handleSelectChange
        }
      }, this.datasource.map(function (item) {
        return h(registry.getBaseWidget('option'), {
          key: item.value,
          props: {
            name: item.label,
            id: item.value
          }
        });
      }));
    }
  });

  var RadioWidget = Vue__default["default"].extend({
    name: 'RadioWidget',
    props: {
      datasource: {
        type: Array,
        default: function _default() {
          return [];
        }
      },
      value: {
        type: [String, Number, Boolean],
        default: ''
      }
    },
    methods: {
      handleChange: function handleChange(val) {
        this.$emit('input', val);
      }
    },
    render: function render(h) {
      return h(registry.getBaseWidget('radio-group'), {
        props: {
          value: this.value
        },
        on: {
          change: this.handleChange
        }
      }, this.datasource.map(function (item) {
        return h(registry.getBaseWidget('radio'), {
          key: item.value,
          class: ['mr24'],
          props: {
            value: item.value,
            label: item.label
          }
        }, item.label);
      }));
    }
  });

  var ButtonWidget = Vue__default["default"].extend({
    name: 'ButtonWidget',
    props: {
      word: {
        type: String,
        default: ''
      }
    },
    methods: {
      handleClick: function handleClick() {
        this.$emit('click');
      }
    },
    render: function render(h) {
      var _self = this;

      return h(registry.getBaseWidget('button'), {
        props: _objectSpread2({}, this.$attrs),
        on: {
          click: function click() {
            _self.handleClick();
          }
        }
      }, this.word);
    }
  });

  var TableWidget = Vue__default["default"].extend({
    name: 'TableWidget',
    render: function render(h) {
      return h(registry.getBaseWidget('table'), {
        props: _objectSpread2({}, this.$attrs)
      });
    }
  });

  var FieldGroupWrap = Vue__default["default"].extend({
    name: 'FieldGroupWrap',
    props: _objectSpread2(_objectSpread2({}, props), {}, {
      // 组类型
      type: {
        type: String,
        default: 'default',
        validator: function validator(value) {
          return ['default', 'normal', 'card'].includes(value);
        }
      },
      // 是否显示组title
      showTitle: {
        type: Boolean,
        default: false
      },
      // 是否显示border
      border: {
        type: Boolean,
        default: false
      }
    }),
    created: function created() {
      // 注册widget TreeNode
      widgetTree.addWidgetNode(this.path, this);
    },
    beforeDestroy: function beforeDestroy() {
      widgetTree.removeWidgetNode(this.path, this);
    },
    render: function render(h) {
      var _this$layout, _this$layout2, _this$schema, _this$layout3;

      var schemaFormStyle = _objectSpread2({
        position: 'relative',
        border: this.border ? '1px solid #dcdee5' : 'none'
      }, ((_this$layout = this.layout) === null || _this$layout === void 0 ? void 0 : _this$layout.item) || {});

      var groupContentStyle = _objectSpread2({}, ((_this$layout2 = this.layout) === null || _this$layout2 === void 0 ? void 0 : _this$layout2.container) || {
        display: 'grid',
        gridGap: '24px' // 未设置layout的布局组的默认样式

      });

      var _self = this;

      var renderDelete = function renderDelete() {
        return h('span', {
          class: ['bk-schema-form-group-delete'],
          style: {
            right: '10px',
            top: '10px'
          },
          on: {
            click: function click() {
              _self.$emit('remove', _self.path);
            }
          }
        }, [h('i', {
          class: ['bk-icon icon-close3-shape']
        })]);
      };

      var title = ((_this$schema = this.schema) === null || _this$schema === void 0 ? void 0 : _this$schema.title) || ((_this$layout3 = this.layout) === null || _this$layout3 === void 0 ? void 0 : _this$layout3.prop);
      return h("div", {
        "class": ['bk-schema-form-group', this.type],
        "style": schemaFormStyle
      }, [title && this.showTitle ? h("span", {
        "class": ['bk-schema-form-group-title', this.type]
      }, [title]) : null, h("div", {
        "style": groupContentStyle,
        "class": "bk-schema-form-group-content"
      }, [this.$slots.default]), this.removeable && renderDelete()]);
    }
  });

  function _extends(){return _extends=Object.assign||function(a){for(var b,c=1;c<arguments.length;c++)for(var d in b=arguments[c],b)Object.prototype.hasOwnProperty.call(b,d)&&(a[d]=b[d]);return a},_extends.apply(this,arguments)}var normalMerge=["attrs","props","domProps"],toArrayMerge=["class","style","directives"],functionalMerge=["on","nativeOn"],mergeJsxProps=function(a){return a.reduce(function(c,a){for(var b in a)if(!c[b])c[b]=a[b];else if(-1!==normalMerge.indexOf(b))c[b]=_extends({},c[b],a[b]);else if(-1!==toArrayMerge.indexOf(b)){var d=c[b]instanceof Array?c[b]:[c[b]],e=a[b]instanceof Array?a[b]:[a[b]];c[b]=d.concat(e);}else if(-1!==functionalMerge.indexOf(b)){for(var f in a[b])if(c[b][f]){var g=c[b][f]instanceof Array?c[b][f]:[c[b][f]],h=a[b][f]instanceof Array?a[b][f]:[a[b][f]];c[b][f]=g.concat(h);}else c[b][f]=a[b][f];}else if("hook"==b)for(var i in a[b])c[b][i]=c[b][i]?mergeFn(c[b][i],a[b][i]):a[b][i];else c[b]=a[b];return c},{})},mergeFn=function(a,b){return function(){a&&a.apply(this,arguments),b&&b.apply(this,arguments);}};var helper=mergeJsxProps;

  var NoTitleArray = Vue__default["default"].extend({
    name: 'NoTitleArray',
    props: props,
    mounted: function mounted() {
      var _this$schema$minItems = this.schema.minItems,
          minItems = _this$schema$minItems === void 0 ? 0 : _this$schema$minItems; // 补全minItems项

      if (this.value.length < minItems) {
        var data = Schema.getSchemaDefaultValue(this.schema.items);
        var remainData = new Array(minItems - this.value.length).fill(data);
        this.$emit('input', [].concat(_toConsumableArray(this.value), _toConsumableArray(remainData)));
      }
    },
    methods: {
      handleAddItem: function handleAddItem() {
        var data = Schema.getSchemaDefaultValue(this.schema.items);
        this.$emit('input', [].concat(_toConsumableArray(this.value), [data]));
      },
      handleRemoveItem: function handleRemoveItem(index) {
        var value = JSON.parse(JSON.stringify(this.value));
        value.splice(index, 1);
        this.$emit('input', value);
      },
      handleInput: function handleInput(_ref) {
        var path = _ref.path,
            value = _ref.value;
        // 捕获widget input事件，包装继续传给上一层处理
        var subPath = Path.getSubPath(this.path, path);
        var newValue = Path.setPathValue(this.value, subPath, value);
        this.$emit('input', newValue);
      }
    },
    render: function render(h) {
      var _this$schema,
          _this$schema$items,
          _this = this;

      var labelBtnStyle = {
        'font-size': '16px',
        color: '#979ba5',
        cursor: 'pointer'
      };
      var props = (_this$schema = this.schema) === null || _this$schema === void 0 ? void 0 : (_this$schema$items = _this$schema.items) === null || _this$schema$items === void 0 ? void 0 : _this$schema$items.properties; // props为空时，表示只有一个项

      var keysLen = Object.keys(props || {}).length;
      var defaultCols = props ? new Array(keysLen).fill('1fr').concat('24px').join(' ') : '1fr 24px';

      var defaultContainerLayout = _objectSpread2({}, this.layout.container || {
        display: 'grid',
        gridGap: '24px',
        'grid-template-columns': defaultCols // 默认配置

      });

      var _self = this;

      var dealSchema = function dealSchema(schema) {
        return (// 处理当前控件默认Schema配置逻辑
          mergeDeep({
            'ui:component': {
              props: {
                placeholder: schema.title
              }
            },
            'ui:props': {
              // 默认不展示标题
              showTitle: false,
              // 0.1 兼容formItem设置 labelWidth 0 不生效问题
              labelWidth: 0.1
            }
          }, schema)
        );
      };

      var renderSchemaField = function renderSchemaField(data) {
        var path = data.path,
            schema = data.schema,
            required = data.required,
            layout = data.layout;
        return h(SchemaField, {
          key: path,
          props: _objectSpread2(_objectSpread2({}, _this.$props), {}, {
            schema: schema,
            required: required,
            path: path,
            layout: layout
          }),
          on: {
            input: function input(data) {
              _self.handleInput(data);
            }
          }
        });
      };

      return h("div", [this.value.map(function (_, index) {
        var _this$schema2;

        var groupPath = Path.getCurPath(_this.path, "".concat(index));
        return h(FieldGroupWrap, helper([{}, {
          "props": _objectSpread2(_objectSpread2({}, _this.$props), {}, {
            path: groupPath,
            layout: _objectSpread2(_objectSpread2({}, _this.layout), {}, {
              container: _objectSpread2({}, defaultContainerLayout)
            })
          })
        }, {
          "class": "mb10"
        }]), [props ? Object.keys(props).map(function (prop) {
          var schemaItem = props[prop];
          var curPath = Path.getCurPath(_this.path, "".concat(index, ".").concat(prop));
          var lastProp = curPath.split('.').pop();
          var layoutConfig = Layout.findLayoutByProp(lastProp, _this.layout.group || []) || {};
          return renderSchemaField({
            path: curPath,
            schema: dealSchema(schemaItem),
            layout: layoutConfig,
            required: Schema.isRequired(schemaItem, prop)
          });
        }) : renderSchemaField({
          path: Path.getCurPath(_this.path, index),
          schema: dealSchema(((_this$schema2 = _this.schema) === null || _this$schema2 === void 0 ? void 0 : _this$schema2.items) || {}),
          layout: {},
          required: false
        }), h("span", {
          "style": labelBtnStyle,
          "on": {
            "click": function click() {
              return _this.handleRemoveItem(index);
            }
          }
        }, [h("i", {
          "class": "bk-icon icon-minus-line"
        })])]);
      }), h("span", {
        "on": {
          "click": this.handleAddItem
        },
        "style": labelBtnStyle
      }, [h("i", {
        "class": "bk-icon icon-plus-line"
      })])]);
    }
  });

  var TabGroupWidget = Vue__default["default"].extend({
    name: 'TabWidget',
    props: _objectSpread2(_objectSpread2({}, props), {}, {
      type: {
        type: String,
        default: 'default',
        validator: function validator(value) {
          return ['default', 'normal', 'card'].includes(value);
        }
      },
      showTitle: {
        type: Boolean,
        default: false
      },
      border: {
        type: Boolean,
        default: false
      }
    }),
    render: function render(h) {
      var _this$schema,
          _this = this;

      var groupWrapProps = _objectSpread2(_objectSpread2({}, this.$props), {}, {
        layout: _objectSpread2(_objectSpread2({}, this.layout), {}, {
          container: {} // Tab组的容器layout由panel内容控制

        }),
        title: this.schema.title
      });

      var _self = this;

      var properties = orderProperties(Object.keys(((_this$schema = this.schema) === null || _this$schema === void 0 ? void 0 : _this$schema.properties) || {}), this.schema['ui:order']);
      return h(FieldGroupWrap, {
        props: _objectSpread2({}, groupWrapProps),
        on: {
          remove: function remove(path) {
            _self.$emit('remove', path);
          }
        }
      }, [h(registry.getBaseWidget('tab'), {}, properties.map(function (key) {
        var _this$schema2, _this$schema2$propert;

        var schemaItem = (_this$schema2 = _this.schema) === null || _this$schema2 === void 0 ? void 0 : (_this$schema2$propert = _this$schema2.properties) === null || _this$schema2$propert === void 0 ? void 0 : _this$schema2$propert[key];
        var curPath = Path.getCurPath(_this.path, key);
        var lastProp = curPath.split('.').pop();
        var layoutConfig = Layout.findLayoutByProp(lastProp, _this.layout.group || []) || {};
        return h(registry.getBaseWidget('tab-panel'), {
          key: key,
          props: {
            name: key,
            label: schemaItem === null || schemaItem === void 0 ? void 0 : schemaItem.title
          }
        }, [h(SchemaField, {
          key: curPath,
          props: _objectSpread2(_objectSpread2({}, _this.$props), {}, {
            schema: schemaItem,
            required: Schema.isRequired(schemaItem, key),
            path: curPath,
            layout: layoutConfig,
            removeable: false // todo: 不往下传递可删除属性

          }),
          on: _objectSpread2({}, _this.$listeners)
        })]);
      }))]);
    }
  });

  var CollapseGroupWidget = Vue__default["default"].extend({
    name: 'CollapseWidget',
    props: _objectSpread2(_objectSpread2({}, props), {}, {
      type: {
        type: String,
        default: 'default',
        validator: function validator(value) {
          return ['default', 'normal', 'card'].includes(value);
        }
      },
      showTitle: {
        type: Boolean,
        default: false
      },
      border: {
        type: Boolean,
        default: false
      }
    }),
    data: function data() {
      return {
        activeName: []
      };
    },
    render: function render(h) {
      var _this$schema,
          _this = this;

      var collapseStyle = {};
      var collapseTitleStyle = {
        background: '#f5f7fa',
        'border-radius': '2px',
        padding: '0 14px',
        height: '100%',
        display: 'flex',
        'align-items': 'center'
      };
      var collapseIconStyle = {
        'font-size': '16px',
        display: 'inline-block',
        transition: 'all 0.5s ease'
      };

      var groupWrapProps = _objectSpread2(_objectSpread2({}, this.$props), {}, {
        layout: _objectSpread2(_objectSpread2({}, this.layout), {}, {
          container: {} // Tab组的容器layout由panel内容控制

        }),
        title: this.schema.title
      });

      var properties = orderProperties(Object.keys(((_this$schema = this.schema) === null || _this$schema === void 0 ? void 0 : _this$schema.properties) || {}), this.schema['ui:order']);
      var collapseItems = properties.map(function (key) {
        var _this$schema2, _this$schema2$propert;

        var schemaItem = (_this$schema2 = _this.schema) === null || _this$schema2 === void 0 ? void 0 : (_this$schema2$propert = _this$schema2.properties) === null || _this$schema2$propert === void 0 ? void 0 : _this$schema2$propert[key];
        var curPath = Path.getCurPath(_this.path, key);
        var lastProp = curPath.split('.').pop();
        var layoutConfig = Layout.findLayoutByProp(lastProp, _this.layout.group || []) || {};
        return h(registry.getBaseWidget('collapse-item'), {
          key: key,
          props: {
            hideArrow: true,
            name: key
          },
          class: ['mb15']
        }, [h('div', {
          style: collapseTitleStyle
        }, [h('i', {
          class: ['bk-icon icon-down-shape mr5'],
          style: _objectSpread2(_objectSpread2({}, collapseIconStyle), {}, {
            transform: _this.activeName.includes(key) ? 'rotate(0deg)' : 'rotate(-90deg)'
          })
        }), schemaItem.title]), h('template', {
          slot: 'content'
        }, [h(SchemaField, {
          key: curPath,
          props: _objectSpread2(_objectSpread2({}, _this.$props), {}, {
            schema: schemaItem,
            required: Schema.isRequired(schemaItem, key),
            path: curPath,
            layout: layoutConfig,
            removeable: false // todo: 不往下传递可删除属性

          }),
          on: _objectSpread2({}, _this.$listeners)
        })])]);
      });

      var _self = this;

      return h(FieldGroupWrap, {
        props: _objectSpread2({}, groupWrapProps)
      }, [h(registry.getBaseWidget('collapse'), {
        style: collapseStyle,
        props: {
          value: this.activeName
        },
        on: {
          input: function input(actives) {
            _self.activeName = actives;
          }
        }
      }, collapseItems)]);
    }
  });

  var SwitcherWidget = Vue__default["default"].extend({
    name: 'SwitchWidget',
    props: {
      value: Boolean
    },
    methods: {
      handleChange: function handleChange(v) {
        this.$emit('input', v);
      }
    },
    render: function render(h) {
      return h(registry.getBaseWidget('switcher'), {
        props: _objectSpread2({
          value: this.value
        }, this.$attrs),
        on: {
          change: this.handleChange
        }
      });
    }
  });

  var ColorWidget = Vue__default["default"].extend({
    name: 'ColorPicker',
    props: {
      value: String
    },
    methods: {
      handleChange: function handleChange(color) {
        this.$emit('input', color);
      }
    },
    render: function render(h) {
      return h(registry.getBaseWidget('color-picker'), {
        props: _objectSpread2({
          value: this.value
        }, this.$attrs),
        on: {
          change: this.handleChange
        }
      });
    }
  });

  var UnitInputWidget = Vue__default["default"].extend({
    name: 'UnitInput',
    props: {
      value: [String, Number],
      unit: {
        type: String,
        default: ''
      }
    },
    methods: {
      handleInput: function handleInput(v) {
        this.$emit('input', v);
      }
    },
    render: function render(h) {
      return h(registry.getBaseWidget('input'), {
        props: _objectSpread2({
          value: this.value
        }, this.$attrs),
        on: {
          input: this.handleInput
        }
      }, [this.unit && h('div', {
        slot: 'append',
        class: ['group-text']
      }, this.unit)]);
    }
  });

  function styleInject(css, ref) {
    if ( ref === void 0 ) ref = {};
    var insertAt = ref.insertAt;

    if (!css || typeof document === 'undefined') { return; }

    var head = document.head || document.getElementsByTagName('head')[0];
    var style = document.createElement('style');
    style.type = 'text/css';

    if (insertAt === 'top') {
      if (head.firstChild) {
        head.insertBefore(style, head.firstChild);
      } else {
        head.appendChild(style);
      }
    } else {
      head.appendChild(style);
    }

    if (style.styleSheet) {
      style.styleSheet.cssText = css;
    } else {
      style.appendChild(document.createTextNode(css));
    }
  }

  var css_248z = "";
  styleInject(css_248z);

  var defaultOptions = {
    namespace: 'bk',
    components: {
      button: ButtonWidget,
      select: SelectWidget,
      radio: RadioWidget,
      checkbox: CheckboxWidget,
      table: TableWidget,
      group: FieldGroupWrap,
      noTitleArray: NoTitleArray,
      tab: TabGroupWidget,
      collapse: CollapseGroupWidget,
      switcher: SwitcherWidget,
      color: ColorWidget,
      unitInput: UnitInputWidget
    },
    fields: {
      object: ObjectField,
      string: StringField,
      any: '',
      array: ArrayField,
      boolean: BooleanField,
      null: '',
      integer: NumberField,
      number: NumberField
    }
  };
  function createForm() {
    var opts = arguments.length > 0 && arguments[0] !== undefined ? arguments[0] : {};
    var options = mergeDeep(defaultOptions, opts);
    var namespace = options.namespace,
        components = options.components,
        fields = options.fields,
        _options$baseWidgets = options.baseWidgets,
        baseWidgets = _options$baseWidgets === void 0 ? {} : _options$baseWidgets;
    Registry.namespace = namespace;
    registry.addComponentsMap(components);
    registry.addFieldsMap(fields);
    registry.addBaseWidgets(baseWidgets);
    return Vue__default["default"].extend({
      name: 'BkuiForm',
      props: props$1,
      data: function data() {
        return {
          rootData: {}
        };
      },
      watch: {
        schema: function schema() {
          this.initFormData();
        },
        context: {
          handler: function handler(ctx) {
            Registry.context = ctx;

            if (hasOwnProperty(ctx, 'rules')) {
              registryGlobalRules(ctx.rules);
            }
          },
          immediate: true
        },
        rules: {
          immediate: true,
          handler: function handler(value) {
            registryGlobalRules(value);
          }
        },
        value: function value() {
          this.initFormData();
        }
      },
      created: function created() {
        this.initFormData();
      },
      methods: {
        initFormData: function initFormData() {
          this.rootData = merge(Schema.getSchemaDefaultValue(createProxy(this.schema, this)) || {}, this.value);
          this.emitFormValueChange(this.rootData, this.value);
        },
        emitFormValueChange: function emitFormValueChange(newValue, oldValue) {
          if (!deepEquals(newValue, oldValue)) {
            this.rootData = newValue;
            this.$emit('input', newValue);
            this.$emit('change', newValue, oldValue);
          }
        },
        validateForm: validateForm,
        validateFormItem: validateFormItem
      },
      render: function render(h) {
        var _self = this;

        return h(registry.getBaseWidget('form'), {
          ref: 'bkui-form',
          props: {
            model: this.value,
            formType: this.formType
          },
          class: {
            'bk-schema-form': true
          },
          style: {
            width: typeof this.width === 'number' ? "".concat(this.width, "px") : this.width
          }
        }, [h(SchemaField, {
          props: _objectSpread2(_objectSpread2({}, this.$props), {}, {
            schema: createProxy(this.schema, this),
            rootData: this.rootData,
            value: this.value,
            layout: new Layout(this.layout).layout
          }),
          scopedSlots: _objectSpread2({}, this.$scopedSlots),
          on: {
            input: function input(_ref) {
              var _ref$path = _ref.path,
                  path = _ref$path === void 0 ? '' : _ref$path,
                  value = _ref.value;

              if (!path) {
                console.warn('widget path is empty');
                return;
              } // 双向绑定逻辑


              var newValue = Path.setPathValue(_self.rootData, path, value);

              _self.emitFormValueChange(newValue, _self.value);
            }
          }
        })]);
      }
    });
  }

  Vue__default["default"].use(bkMagic__namespace);

  exports["default"] = createForm;
  exports.events = events;

  Object.defineProperty(exports, '__esModule', { value: true });

}));
//# sourceMappingURL=bkui-form-umd.js.map
