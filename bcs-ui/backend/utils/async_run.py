# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
import asyncio
import concurrent
from dataclasses import dataclass
from typing import Any, List, Optional


class AsyncRunException(BaseException):
    pass


@dataclass
class AsyncResult:
    ret: Any
    exc: Optional[Exception]


def get_or_create_loop():
    try:
        # 主线程
        loop = asyncio.get_event_loop()
        if loop.is_closed():
            raise RuntimeError('Event loop is closed')
        return loop, False
    except RuntimeError:
        # 非主线程
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        return loop, True


def async_run(tasks, raise_exception: bool = True) -> List[AsyncResult]:
    """
    run a group of tasks async(仅适用于IO密集型)
    Requires the tasks arg to be a list of functools.partial()
    """
    if not tasks:
        return []

    loop, created = get_or_create_loop()
    # https://github.com/python/asyncio/issues/258
    executor = concurrent.futures.ThreadPoolExecutor(8)
    loop.set_default_executor(executor)

    async_tasks = [asyncio.ensure_future(async_task(task, loop)) for task in tasks]
    # run tasks in parallel
    loop.run_until_complete(asyncio.wait(async_tasks))

    # 获取 Task 结果
    results = []
    for task in async_tasks:
        exc = task.exception()
        ret = task.result() if exc is None else None
        results.append(AsyncResult(ret=ret, exc=exc))

    executor.shutdown(wait=True)
    if created:
        loop.close()

    if raise_exception:
        exceptions = filter(None, [result.exc for result in results])
        err_msg = ';'.join([str(exc) for exc in exceptions])
        if err_msg:
            raise AsyncRunException(err_msg)

    return results


async def async_task(params, loop):
    """
    Perform a task asynchronously.
    """
    # get the calling function
    # This executes a task in its own thread (in parallel)
    result = await loop.run_in_executor(None, params)
    return result
