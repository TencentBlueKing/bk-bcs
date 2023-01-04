
// TODO 替换为表单化库的Schema解析

export const valueType = (value) => {
  if (Array.isArray(value)) {
    return 'array';
  } if (typeof value === 'string') {
    return 'string';
  } if (typeof value === 'boolean') {
    return 'boolean';
  } if (!isNaN(value)) {
    return 'number';
  } if (value === null) {
    return 'null';
  } if (typeof value === 'object') {
    return 'object';
  }
  return typeof value;
};

export const isObj = (val: unknown): val is object => Object.prototype.toString.call(val) === '[object Object]';

export function initializationValue(
  type,
  defaultInitValue = { integer: undefined, number: undefined },
) {
  switch (type) {
    case 'any':
      return undefined;
    case 'array':
      return [];
    case 'boolean':
      return false;
    case 'integer':
      return defaultInitValue.integer;
    case 'null':
      return null;
    case 'number':
      return defaultInitValue.number;
    case 'object':
      return {};
    case 'string':
      return '';
  }
}

export default class Schema {
  // 获取Schema字段默认值
  static getSchemaDefaultValue(_schema) {
    const schema: any = isObj(_schema) ? _schema : {};

    switch (this.getSchemaType(schema)) {
      case 'null':
        return null;
      case 'object':
        return Object.keys(schema.properties || {}).reduce((pre, key) => {
          const defaultValue = this.getSchemaDefaultValue(schema.properties?.[key]);
          pre[key] = defaultValue;
          return pre;
        }, {});
      case 'array':
        // todo
        return Array.isArray(schema.items)
          ? schema.items.map(item => this.getSchemaDefaultValue(item))
          : [];
    }

    return schema.default !== undefined
      ? schema.default
      : initializationValue(schema.type || 'any');
  }
  static getSchemaType(schema): string {
    const { type } = schema;

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
  static getSchemaByProp(schema, props: string) {
    if (!schema || !Object.keys(schema).length) return {};
    const data = props.split('.');
    return data.reduce((data, prop) => {
      if (data.type === 'array') {
        return data?.items?.properties?.[prop];
      }

      return data?.properties?.[prop];
    }, schema);
  }
}
