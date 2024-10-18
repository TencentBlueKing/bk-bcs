// @ts-nocheck
/* eslint-disable */
import { createApp, getCurrentInstance, ref, watch } from 'vue';

import instanceStore from './instance-store';

let callbackSeed = 0;
function JSONP(api: string, params = {}, options: any = {}) {
  return new Promise((resolve, reject) => {
    let timer: number;
    const callbackName = `USER_LIST_CALLBACK_${(callbackSeed += 1)}`;
    window[callbackName] = (response: any) => {
      timer && clearTimeout(timer);
      document.body.removeChild(script);
      delete window[callbackName];
      resolve(response);
    };
    const script = document.createElement('script');
    script.onerror = (_event) => {
      document.body.removeChild(script);
      delete window[callbackName];
      reject('Get user list failed.');
    };
    const query = [];
    // eslint-disable-next-line no-restricted-syntax
    for (const key in params) {
      query.push(`${key}=${params[key]}`);
    }
    script.src = `${api}?${query.join('&')}&callback=${callbackName}`;
    if (options.timeout) {
      setTimeout(() => {
        document.body.removeChild(script);
        delete window[callbackName];
        reject('Get user list timeout.');
      }, options.timeout);
    }
    document.body.appendChild(script);
  });
}

// 缓存已经加载过的人员
// 以api为key，存储不同数据源的用户
const userMap = new Map();

function getMap(api: string) {
  if (userMap.has(api)) {
    return userMap.get(api);
  }
  const map = new Map();
  userMap.set(api, map);
  return map;
}

function storeUsers(api: string, users: any) {
  const map = getMap(api);
  users.forEach((user: any) => map.set(user.username, user));
}

function getUsers(api: string, usernames: any) {
  const map = getMap(api);
  const users: string[] = [];
  usernames.forEach((username: string) => {
    if (map.has(username)) {
      users.push(map.get(username));
    }
  });
  return users;
}

// 接口最大支持100条记录，超过100条需要将请求拆分为多个
async function handleBatchSearch(api: string, usernames: string[], options: any) {
  const map = getMap(api);
  const unique = [...new Set(usernames)].filter((username) => !map.has(username));
  if (!unique.length) {
    return Promise.resolve(getUsers(api, usernames));
  }
  const slices: string[][] = [];
  unique.reduce((slice, username, index) => {
    if (slice.length < 100) {
      slice.push(username);
      if (index === unique.length - 1) {
        slices.push(slice);
      }
      return slice;
    }
    slices.push(slice);
    return [];
  }, []);
  try {
    const responses = await Promise.all(
      slices.map((slice) =>
        JSONP(
          api,
          {
            app_code: 'bk-magicbox',
            exact_lookups: slice.join(','),
            page_size: 100,
            page: 1,
          },
          options,
        ),
      ),
    );
    responses.forEach((response: any) => {
      if (response.code !== 0) return;
      storeUsers(api, response.data.results || []);
    });
  } catch (error) {
    console.error(error);
  }
  return Promise.resolve(getUsers(api, usernames));
}

function createVm(apiStr: string) {
  const app = {
    setup() {
      const { proxy } = getCurrentInstance();
      instanceStore.setInstance('exactSearch', apiStr, proxy);

      const api = ref('');
      api.value = apiStr;

      const queue = ref([]);

      watch(
        () => queue,
        (q: any) => {
          q.value.length && dispatchSeach();
        },
        { deep: true },
      );

      const search = (usernames) =>
        new Promise((resolve) => {
          queue.value.push({
            resolve,
            usernames,
          });
        });

      const dispatchSeach = async () => {
        const currentQueue = [...queue.value];
        queue.value = [];
        try {
          const allNames = currentQueue.reduce((all, { usernames }) => all.concat(usernames), []);
          const users = await request.exactSearch(api.value, allNames);
          const map = {};
          users.forEach((user) => {
            map[user.username] = user;
          });
          currentQueue.forEach(({ resolve, usernames }) => {
            const resolveData = [];
            usernames.forEach((username) => {
              // eslint-disable-next-line no-prototype-builtins
              if (map.hasOwnProperty(username)) {
                resolveData.push(map[username]);
              }
            });
            resolve(resolveData);
          });
        } catch (error) {
          currentQueue.forEach(({ resolve }) => {
            resolve([]);
          });
          console.error(error);
        }
      };

      return {
        api,
        queue,
        search,
        dispatchSeach,
      };
    },
    render() {
      return null; // 渲染为空
    },
  };

  const vm = createApp(app);
  const vmContainer = document.createElement('div');
  vm.mount(vmContainer);
  // document.body.appendChild(vmContainer);
  // vmMap.set(apiStr, vm);
  return vm;
}

const request = {
  // 模糊搜索，失败时返回空
  async fuzzySearch(api, params, options) {
    const data = {};
    try {
      const response = await JSONP(api, params, options);
      if (response.code !== 0) {
        throw new Error(response);
      }
      data.count = response.data.count;
      data.results = response.data.results || [];
      storeUsers(api, data.results);
    } catch (error) {
      console.error(error.message);
      data.count = 0;
      data.results = [];
    }
    return data;
  },
  // 精确搜索，不在此处捕获异常，在上层捕获并显示在tooltips中
  async exactSearch(api, username, options) {
    const isArray = Array.isArray(username);
    const usernames = isArray ? username : [username];
    const users = await handleBatchSearch(api, usernames, options);
    if (isArray) {
      return users;
    }
    return users[0];
  },
  // 粘贴时对粘贴的用户进行校验
  pasteValidate(api, usernames, options) {
    return handleBatchSearch(api, usernames, options);
  },
  // 队列式查询，用于多个组件共存时，批量拉取已存在的用户信息
  scheduleExactSearch(api, username) {
    const usernames = Array.isArray(username) ? username : [username];
    createVm(api);
    const vm = instanceStore.getInstance('exactSearch', api);
    return vm.search(usernames);
  },
};

export default request;
