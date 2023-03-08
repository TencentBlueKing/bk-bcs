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

export default class RequestQueue {
  constructor() {
    this.queue = [];
  }

  /**
     * 根据 id 获取请求对象，如果不传入 id，则获取整个地列
     *
     * @param {string?} id id
     *
     * @return {Array|Object} 队列集合或队列对象
     */
  get(id = undefined) {
    if (typeof id === 'undefined') {
      return this.queue;
    }
    return this.queue.filter(request => request.requestId === id);
  }

  /**
     * 设置新的请求对象到请求队列中
     *
     * @param {Object} newRequest 请求对象
     */
  set(newRequest) {
    this.queue.push(newRequest);
  }

  /**
     * 根据 id 删除请求对象
     *
     * @param {string} id id
     */
  delete(id) {
    this.queue = [...this.queue.filter(request => request.requestId !== id)];
  }

  /**
     * cancel 请求队列中的请求
     *
     * @param {string|Array?} requestIds 要 cancel 的请求 id，如果不传，则 cancel 所有请求
     * @param {string?} msg cancel 时的信息
     *
     * @return {Promise} promise 对象
     */
  cancel(requestIds, msg = 'request canceled') {
    let cancelQueue = [];
    if (typeof requestIds === 'undefined') {
      cancelQueue = [...this.queue];
    } else if (requestIds instanceof Array) {
      requestIds.forEach((requestId) => {
        const cancelRequest = this.get(requestId);
        if (cancelRequest) {
          cancelQueue = [...cancelQueue, ...cancelRequest];
        }
      });
    } else {
      const cancelRequest = this.get(requestIds);
      if (cancelRequest) {
        cancelQueue = [...cancelQueue, ...cancelRequest];
      }
    }

    try {
      cancelQueue.forEach((request) => {
        const { requestId } = request;
        this.delete(requestId);
        request.cancelExcutor({ type: 'cancel', msg: `${msg}: ${requestId}` });
      });
      return Promise.resolve(requestIds);
    } catch (error) {
      return Promise.reject(error);
    }
  }
}
