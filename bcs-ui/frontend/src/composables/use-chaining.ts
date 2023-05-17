import { customRef } from 'vue';

const handlerFactory = configs => ({
  get(target, property, receiver) {
    if (!(property in target) && (property in configs)) {
      target[property] = Object.prototype.toString.call(configs[property]) === '[object Object]'
        ? proxify({}, configs)
        : configs[property];
    }

    return Reflect.get(target, property, receiver);
  },
  set(target, property, value, receiver) {
    if (({}).toString.call(value) === '[object Object]') value = deepApply(value, configs);

    return Reflect.set(target, property, value, receiver);
  },
});

const deepApply = function (data, configs) {
  const proxy = proxify({}, configs);
  Object.keys(data).forEach((property) => {
    proxy[property] = data[property];
  });
  return proxy;
};

function proxify(defprop = {}, configs) {
  return new Proxy(defprop, handlerFactory(configs));
};

export type IChainingConfig =  {
  path: string;
  type: 'string' | 'number' | 'object' | 'null' | 'undefined' | 'array'
} | string;

export default function useChainingRef<T>(value, config: IChainingConfig[]) {
  const valueMap = {
    string: '',
    number: 0,
    object: {},
    null: null,
    undefined,
    array: [],
  };
  const _config = config.reduce<Record<string, any>>((pre, current) => {
    if (typeof current === 'string') {
      current.split('.').forEach((key) => {
        pre[key] = {};
      });
    } else {
      const paths = current.path.split('.');
      paths.forEach((key, index) => {
        pre[key] = index < (paths.length - 1) ? {} : valueMap[current.type];
      });
    }

    return pre;
  }, {});

  let innerValue = proxify(value, _config);

  return customRef<T>((track, trigger) => ({
    get() {
      track();
      return innerValue;
    },
    set(newValue: any) {
      innerValue = newValue;
      trigger();
    },
  }));
}
