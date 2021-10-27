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

"""module for creating long-running polling tasks via celery"""
import logging
import time
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from typing import Any, Dict, Optional, Type

from celery import shared_task

logger = logging.getLogger(__name__)


class PollingStatus(int, Enum):
    """Status of a single polling action"""

    DOING = 1
    DONE = 2


class PollingResult:
    """A single polling result object

    :param status: status of current polling action, such as: DOING, DONE
    :param data: extra data of current polling
    """

    def __init__(self, status: PollingStatus, data: Optional[Any] = None):
        self.status = status
        self.data = data

    def __str__(self):
        return f'stauts={self.status} data={self.data}'

    @classmethod
    def doing(cls, *args, **kwargs):
        """Shortcut for creating doing result"""
        return cls(PollingStatus.DOING, *args, **kwargs)

    @classmethod
    def done(cls, *args, **kwargs):
        """Shortcut for creating done result"""
        return cls(PollingStatus.DONE, *args, **kwargs)


@dataclass
class PollingMetadata:
    """Metadata of a polling process"""

    retries: int
    # unix timestamp of when current polling has been started
    query_started_at: float
    queried_count: int
    # data attribute of last polling action
    last_polling_data: Optional[Dict] = None


class TaskPoller(ABC):
    """task status poller

    :param params: params to perform polling
    :param metadata: metadata object of current poller
    """

    _registered_pollers: Dict[str, Type['TaskPoller']] = {}

    max_retries_on_error = 10
    overall_timeout_seconds = 3600 * 24 * 7
    default_retry_delay_seconds = 10

    def __init__(self, params: Dict, metadata: PollingMetadata):
        self.params = params
        self.metadata = metadata

    def __init_subclass__(cls, *args, **kwargs):
        cls._registered_pollers[cls.__name__] = cls

    @classmethod
    def get_poller_cls(cls, name: str) -> Type['TaskPoller']:
        return cls._registered_pollers[name]

    @classmethod
    def start(cls, params: Dict, callback_handler_cls: Optional[Type] = None):
        """Start a new polling task

        :param params: params for starting polling, must be Json compatible in order to use celery
        :param callback_handler_cls: type to handle poll result
        """
        handler_name = None
        if callback_handler_cls is not None:
            assert issubclass(callback_handler_cls, CallbackHandler)
            handler_name = callback_handler_cls.__name__

        # Start background task
        cls.get_async_task().delay(cls.__name__, handler_name, params)

    def make_next_metadata(self, has_error: bool = False, last_polling_data: Optional[Dict] = None) -> PollingMetadata:
        """Make the metadata object for next polling

        :param has_error: if has error, will affect `retries` attribute
        :param last_polling_data: current polling data, "last" data for next polling
        """
        # Reset retries when no exception occurs
        if not has_error:
            retries = 0
        else:
            retries = self.metadata.retries + 1
        return PollingMetadata(
            retries=retries,
            query_started_at=self.metadata.query_started_at,
            queried_count=self.metadata.queried_count + 1,
            last_polling_data=last_polling_data,
        )

    @abstractmethod
    def query(self) -> PollingResult:
        """Start a polling action, subclasses must override this method"""
        raise NotImplementedError()

    def get_retry_delay(self) -> int:
        """Get delay of next retry"""
        return self.default_retry_delay_seconds

    def exceeded_timeout(self) -> bool:
        """Check if current polling procedure has exceeded max timeout"""
        return (time.time() - self.metadata.query_started_at) > self.get_overall_timeout_seconds()

    def get_overall_timeout_seconds(self) -> int:
        """The overall timeout seconds for a complete polling procedure"""
        return self.overall_timeout_seconds

    def exceeded_max_retries(self) -> bool:
        """Check if current polling has retried too many times"""
        return (self.metadata.retries + 1) > self.max_retries_on_error

    def __str__(self):
        return '<%s: params=%s>' % (self.__class__.__name__, self.params)

    @classmethod
    def get_async_task(self) -> Any:
        """Return the async celery task object for polling in backend"""
        return check_status_until_finished


class CallbackStatus(int, Enum):
    """Status of a finished polling"""

    NORMAL = 0
    TIMEOUT = 1
    EXCEPTION = 2

    def is_exception(self):
        return self in (self.TIMEOUT, self.EXCEPTION)


class CallbackResult:
    """The final result of a polling procedure

    :param status: status of current result
    :param data: data of current result
    :param message: extra message of current result
    """

    def __init__(self, status: CallbackStatus, data: Optional[Dict] = None, message: str = ""):
        self.status = status
        self.data = data or {}
        self.message = message

    @property
    def is_exception(self):
        return self.status.is_exception()

    def to_dict(self):
        return {'status': self.status.value, 'message': self.message, 'data': self.data}

    def __str__(self):
        return '<%s: %s is_exception=%s>' % (self.__class__.__name__, self.to_dict(), self.is_exception)


class CallbackHandler(ABC):
    """handle callback result

    :params: params of current polling result
    """

    _registered_handlers: Dict[str, Type['CallbackHandler']] = {}

    def __init_subclass__(cls, *args, **kwargs):
        cls._registered_handlers[cls.__name__] = cls

    @classmethod
    def get_handler_cls(cls, name: str) -> Type['CallbackHandler']:
        return cls._registered_handlers[name]

    @abstractmethod
    def handle(self, result: CallbackResult, poller: TaskPoller):
        """Handle final callback result

        :param result: CallbackResult instance
        :param poller: Current TaskPoller instance
        """
        raise NotImplementedError()


class NullResultHandler(CallbackHandler):
    """A null implementation of callback result handler"""

    def handle(self, result: CallbackResult, poller: TaskPoller):
        pass


@shared_task(acks_late=True, name='poll_task.check_status_until_finished')
def check_status_until_finished(poller_name: str, handler_name: str, params: Dict, queue: Optional[str] = None):
    """Main async task for polling

    :param poller_name: name of poller class
    :param handler_name: name of result handler
    :param params: params for performing polling
    :param queue: dedicated queue name
    """
    req = check_status_until_finished.request
    metadata = PollingMetadata(
        retries=req.retries,
        query_started_at=req.get('query_started_at', time.time()),
        queried_count=req.get('queried_count', 0),
        last_polling_data=req.get('last_polling_data'),
    )

    # Make handler and poller by name
    poller = TaskPoller.get_poller_cls(poller_name)(params, metadata)
    if handler_name is not None:
        handler_cls = CallbackHandler.get_handler_cls(handler_name)
    else:
        handler_cls = NullResultHandler

    scheduler = PollTaskScheduler(poller, handler_cls)
    next_metadata = scheduler.run()

    if next_metadata:
        # Start next polling
        countdown = poller.get_retry_delay()
        logger.debug('Will retry query status for %s after %s seconds. metadata=%s', poller, countdown, metadata)
        poller.get_async_task().subtask(
            args=(poller_name, handler_name, params),
            kwargs={'queue': queue},
            countdown=countdown,
            retries=next_metadata.retries,
            queue=queue,
        ).apply_async(
            headers={
                'queried_count': next_metadata.queried_count,
                'query_started_at': next_metadata.query_started_at,
                'last_polling_data': next_metadata.last_polling_data,
            }
        )


class PollingQueryError(Exception):
    """Error when perform poller.query method"""


class PollTaskScheduler:
    """Schedule poll tasks"""

    def __init__(self, poller: TaskPoller, handler_cls: Type[CallbackHandler]):
        self.poller = poller
        self.handler_cls = handler_cls

    def run(self) -> Optional[PollingMetadata]:
        """Start schedule process"""
        if self.poller.exceeded_timeout():
            logger.info('exceeded total timeout, ts_query_started=%s' % self.poller.metadata.query_started_at)
            self._callback_timeout()
            return None

        try:
            polling_result = self._safe_query(self.poller)
        except PollingQueryError as e:
            if self.poller.exceeded_max_retries():
                self._callback_exception(e)
                return None

            # Retry next polling, set `last_polling_data` field to the value of last succeeded call
            metadata = self.poller.make_next_metadata(
                has_error=True, last_polling_data=self.poller.metadata.last_polling_data
            )
            return metadata

        if polling_result.status == PollingStatus.DONE:
            ret = CallbackResult(status=CallbackStatus.NORMAL, data=polling_result.data)
            self._callback(ret)
            return None

        metadata = self.poller.make_next_metadata(has_error=False, last_polling_data=polling_result.data)
        return metadata

    @staticmethod
    def _safe_query(poller: TaskPoller) -> PollingResult:
        """call poller's query method with exception handling

        :raises: PollingQueryError
        """
        try:
            polling_result = poller.query()
        except Exception as e:
            logger.exception('Exception when query status, poll_class=%s' % poller)
            raise PollingQueryError(str(e))

        logger.debug('Query status result, poll_class=%s, polling result: %s' % (poller, polling_result))
        return polling_result

    def _callback(self, result: CallbackResult):
        """Callback handler"""
        self.handler_cls().handle(result, self.poller)

    def _callback_timeout(self):
        """Callback handler with timeout result"""
        ret = CallbackResult(status=CallbackStatus.TIMEOUT, message='exceeded total timeout')
        self.handler_cls().handle(ret, self.poller)

    def _callback_exception(self, e: Exception):
        """Callback handler when exceeds max retries"""
        ret = CallbackResult(status=CallbackStatus.EXCEPTION, message=f'exception: {e}')
        self.handler_cls().handle(ret, self.poller)
