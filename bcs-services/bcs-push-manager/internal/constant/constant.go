/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package constant defines various constants used throughout the application.
package constant

// ResponseCode defines the custom response code.
type ResponseCode int

// Response codes for API responses.
const (
	// ResponseCodeSuccess indicates a successful API request.
	ResponseCodeSuccess ResponseCode = 0
	// ResponseCodeBadRequest indicates that the request parameters were invalid.
	ResponseCodeBadRequest ResponseCode = 400
	// ResponseCodeNotFound indicates that the requested resource was not found.
	ResponseCodeNotFound ResponseCode = 404
	// ResponseCodeInternalError indicates a server-side internal error.
	ResponseCodeInternalError ResponseCode = 500
)

// Response messages for API responses.
const (
	// ResponseMsgSuccess is the message for a successful request.
	ResponseMsgSuccess = "success"
	// ResponseMsgDomainRequired is the message when the domain is missing.
	ResponseMsgDomainRequired = "domain is required"
	// ResponseMsgEventRequired is the message when the event data is missing.
	ResponseMsgEventRequired = "event is required"
	// ResponseMsgEventIDRequired is the message when the event_id is missing.
	ResponseMsgEventIDRequired = "event_id is required"
	// ResponseMsgTemplateRequired is the message when the template data is missing.
	ResponseMsgTemplateRequired = "template is required"
	// ResponseMsgTemplateIDRequired is the message when the template_id is missing.
	ResponseMsgTemplateIDRequired = "template_id is required"
	// ResponseMsgWhitelistRequired is the message when the whitelist data is missing.
	ResponseMsgWhitelistRequired = "whitelist is required"
	// ResponseMsgWhitelistIDRequired is the message when the whitelist_id is missing.
	ResponseMsgWhitelistIDRequired = "whitelist_id is required"
	// ResponseMsgPushEventNotFound is the message when a push event is not found.
	ResponseMsgPushEventNotFound = "push event not found"
	// ResponseMsgPushTemplateNotFound is the message when a push template is not found.
	ResponseMsgPushTemplateNotFound = "push template not found"
	// ResponseMsgPushWhitelistNotFound is the message when a push whitelist is not found.
	ResponseMsgPushWhitelistNotFound = "push whitelist not found"
	// ResponseMsgNoFieldsToUpdate is the message when no fields are provided for an update.
	ResponseMsgNoFieldsToUpdate = "no fields to update"
	// ResponseMsgEventDetailRequired is the message when event detail is missing.
	ResponseMsgEventDetailRequired = "event detail is required"
	// ResponseMsgDimensionFieldsRequired is the message when dimension fields are missing.
	ResponseMsgDimensionFieldsRequired = "dimension fields is required"
	// ResponseMsgDimensionRequired is the message when dimension is missing.
	ResponseMsgDimensionRequired = "dimension is required"
	// ResponseMsgMetricDataRequired is the message when metric data is missing.
	ResponseMsgMetricDataRequired = "metric data is required"
	// ResponseMsgStartTimeRequired is the message when start time is missing.
	ResponseMsgStartTimeRequired = "start time is required"
	// ResponseMsgEndTimeRequired is the message when end time is missing.
	ResponseMsgEndTimeRequired = "end time is required"
	// ResponseMsgTitleRequired is the message when the template title is missing.
	ResponseMsgTitleRequired = "template title is required"
	// ResponseMsgBodyRequired is the message when the template body is missing.
	ResponseMsgBodyRequired = "template body is required"
	// ResponseMsgDomainMismatch is the message when the domain of the resource does not match the requested domain.
	ResponseMsgDomainMismatch = "domain mismatch"
)

// Push event status constants.
const (
	// EventStatusPending indicates the push event is pending.
	EventStatusPending = 0
	// EventStatusSuccess indicates the push event was successful.
	EventStatusSuccess = 1
	// EventStatusFailed indicates the push event failed.
	EventStatusFailed = 2
	// EventStatusWhitelisted indicates the push event was whitelisted.
	EventStatusWhitelisted = 3
)

// Whitelist status constants.
const (
	// WhitelistStatusNone indicates the whitelist is not active.
	WhitelistStatusNone = 0
	// WhitelistStatusActive indicates the whitelist is active.
	WhitelistStatusActive = 1
	// WhitelistStatusExpired indicates the whitelist has expired.
	WhitelistStatusExpired = 2
)

// Approval status constants.
const (
	// ApprovalStatusPending indicates the approval is pending.
	ApprovalStatusPending = 0
	// ApprovalStatusApproved indicates the request has been approved.
	ApprovalStatusApproved = 1
	// ApprovalStatusRejected indicates the request has been rejected.
	ApprovalStatusRejected = 2
)

// Alert level constants.
const (
	// AlertLevelFatal represents a fatal alert level.
	AlertLevelFatal = "fatal"
	// AlertLevelWarning represents a warning alert level.
	AlertLevelWarning = "warning"
	// AlertLevelReminder represents a reminder alert level.
	AlertLevelReminder = "reminder"
)

// Notification constants.
const (
	// NotificationActionQueueName is the name of queue.
	NotificationActionQueueName = "textqueuename"
	// PushTypeRtx represents the RTX push type.
	PushTypeRtx = "rtx"
	// PushTypeMail represents the Mail push type.
	PushTypeMail = "mail"
	// NotificationResultSuccess represents a successful notification.
	NotificationResultSuccess = "success"
	// NotificationResultFailed represents a failed notification.
	NotificationResultFailed = "failed"
	// NotificationDefaultSender specifies the default sender.
	NotificationDefaultSender = "bcs"
	// EmailBodyFormat specifies the format of the email body, using HTML format.
	NotificationMailFormat = "html"
	// EventDetailKeyContent is the key for event content.
	EventDetailKeyContent = "content"
	// EventDetailKeyReceivers is the key for event receivers.
	EventDetailKeyReceivers = "receivers"
	// EventDetailKeyTypes is the key for event types.
	EventDetailKeyTypes = "types"
	// EventDetailKeyTitle is the key for event title.
	EventDetailKeyTitle = "title"
	// MQRoutingKeyFormat is the format string for MQ routing key, e.g. "*.push.%s"
	MQRoutingKeyFormat = "*.push.%s"
	// MQRoutingKeyBindPattern is the routing key pattern for queue binding, e.g. "*.push.#"
	MQRoutingKeyBindPattern = "*.push.#"
)

const (
	// MicroMetaKeyHTTPPort http port in micro service meta
	MicroMetaKeyHTTPPort = "httpport"
	// DefaultPage specifies the default page number.
	DefaultPage = 1
	// DefaultPageSize specifies the default number of items per page.
	DefaultPageSize = 10
	// DefaultExchangeName xxx
	DefaultExchangeName = "push"
	// ModuleThirdpartyServiceManager helm manager discovery name
	ModuleThirdpartyServiceManager = "bcsthirdpartyservice.bkbcs.tencent.com"
	// ModulePushManager helm manager discovery name
	ModulePushManager = "pushmanager.bkbcs.tencent.com"
)
