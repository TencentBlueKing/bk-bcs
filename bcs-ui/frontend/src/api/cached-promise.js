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

export default class CachedPromise {
  constructor() {
    this.cache = {};
  }

  /**
     * 根据 id 获取缓存对象，如果不传 id，则获取所有缓存
     *
     * @param {string?} id id
     *
     * @return {Array|Promise} 缓存集合或promise 缓存对象
     */
  get(id) {
    if (typeof id === 'undefined') {
      return Object.keys(this.cache).map(requestId => this.cache[requestId]);
    }
    return this.cache[id];
  }

  /**
     * 设置 promise 缓存对象
     *
     * @param {string} id id
     * @param {Promise} promise 要缓存的 promise 对象
     *
     * @return {Promise} promise 对象
     */
  set(id, promise) {
    this.cache = Object.assign({}, this.cache, { [id]: promise });
  }

  /**
     * 删除 promise 缓存对象
     *
     * @param {string|Array?} deleteIds 要删除的缓存对象的 id，如果不传则删除所有
     *
     * @return {Promise} 以成功的状态返回 Promise 对象
     */
  delete(deleteIds) {
    let requestIds = [];
    if (typeof deleteIds === 'undefined') {
      requestIds = Object.keys(this.cache);
    } else if (deleteIds instanceof Array) {
      deleteIds.forEach((id) => {
        if (this.get(id)) {
          requestIds.push(id);
        }
      });
    } else if (this.get(deleteIds)) {
      requestIds.push(deleteIds);
    }

    requestIds.forEach((requestId) => {
      delete this.cache[requestId];
    });

    return Promise.resolve(deleteIds);
  }
}
