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
 *
 */

package utils

import (
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// file span.go aim to collect common span tag and log field to make it easier to call

// The tag names are defined as typed strings, so that in addition to the usual use
//     span.setTag(TagName, value)
// they also support value type validation via this additional syntax:
//    TagName.Set(span, value)

// SetSpanKindTag set tag span.kind by kind, SpanKind (client/server or producer/consumer)
func SetSpanKindTag(span opentracing.Span, kind ext.SpanKindEnum) {
	if span == nil {
		return
	}

	switch kind {
	case ext.SpanKindRPCClientEnum:
		ext.SpanKindRPCClient.Set(span)
	case ext.SpanKindRPCServerEnum:
		ext.SpanKindRPCServer.Set(span)
	case ext.SpanKindProducerEnum:
		ext.SpanKindProducer.Set(span)
	case ext.SpanKindConsumerEnum:
		ext.SpanKindConsumer.Set(span)
	default:
	}

	return
}

// SetSpanComponentTag is a identifier of the module, library, or package that is generating a span.
func SetSpanComponentTag(span opentracing.Span, name string) {
	if span == nil {
		return
	}
	ext.Component.Set(span, name)
}

//////////////////////////////////////////////////////////////////////
// Peer tags. These tags can be emitted by either client-side or
// server-side to describe the other side/service in a peer-to-peer
// communications, like an RPC call.
//////////////////////////////////////////////////////////////////////

// PeerKind show peerTag type
type PeerKind string

const (
	// PeerService records the service name of the peer.
	PeerService PeerKind = "service"
	// PeerAddress records the address name of the peer.
	// This may be a "ip:port", a bare "hostname" or even a database DSN substring
	PeerAddress PeerKind = "address"
	// PeerIP records IP v4 host address of the peer
	PeerIP PeerKind = "ip"
	// PeerPort records port number of the peer
	PeerPort PeerKind = "port"
	// PeerHandler record interface type of the peer
	PeerHandler PeerKind = "handler"
)

var (
	// PeerTagHandler is handler name tag for span
	PeerTagHandler = ext.StringTagName("handler")
)

// SetSpanPeerTag set peer tag for span
func SetSpanPeerTag(span opentracing.Span, kind PeerKind, value string) {
	if span == nil {
		return
	}
	switch kind {
	case PeerService:
		ext.PeerService.Set(span, value)
	case PeerAddress:
		ext.PeerAddress.Set(span, value)
	case PeerIP:
		ext.PeerHostIPv4.SetString(span, value)
	case PeerPort:
		port, _ := strconv.Atoi(value)
		ext.PeerPort.Set(span, uint16(port))
	case PeerHandler:
		PeerTagHandler.Set(span, value)
	default:
	}

	return
}

//////////////////////////////////////////////////////////////////////
// HTTP Tags
//////////////////////////////////////////////////////////////////////

// HTTPKind for http tag type
type HTTPKind string

var (
	// HTTPTagHandler is handler name tag for span
	HTTPTagHandler = ext.StringTagName("handler")
)

const (
	// HTTPUrl should be the URL of the request being handled in this segment of the trace,
	// in standard URI format.
	HTTPUrl HTTPKind = "url"
	// HTTPMethod is the HTTP method of the request, and is case-insensitive.
	HTTPMethod HTTPKind = "method"
	// HTTPStatusCode is the numeric HTTP status code
	HTTPStatusCode HTTPKind = "status_code"
)

// SetSpanHTTPTag set http tag for span
func SetSpanHTTPTag(span opentracing.Span, kind HTTPKind, value string) {
	if span == nil {
		return
	}
	switch kind {
	case HTTPUrl:
		ext.HTTPUrl.Set(span, value)
	case HTTPMethod:
		ext.HTTPMethod.Set(span, value)
	case HTTPStatusCode:
		port, _ := strconv.Atoi(value)
		ext.HTTPStatusCode.Set(span, uint16(port))
	default:
	}

	return
}

//////////////////////////////////////////////////////////////////////
// DB Tags
//////////////////////////////////////////////////////////////////////

// DBKind for DB tag type
type DBKind string

const (
	// DBInstance is database instance name.
	DBInstance DBKind = "instance"
	// DBType is a database type. For any SQL database, "sql".
	// For others, the lower-case database category, e.g. "redis"
	DBType DBKind = "type"
	// DBUser is a username for accessing database.
	DBUser DBKind = "user"
	// DBDatabase is database name for access
	DBDatabase DBKind = "database"
	// DBTable for database table
	DBTable DBKind = "table"
	// DbOperation for database table operation(create/delete/update/get/list)
	DbOperation DBKind = "operation"
)

var (
	// DBTagDatabase is database name tag for span
	DBTagDatabase = ext.StringTagName("db.database")
	// DBTagTable is table tag for span
	DBTagTable = ext.StringTagName("db.table")
	// DBTagOperation is a operation tag for span
	DBTagOperation = ext.StringTagName("db.operation")
)

// SetSpanDBTag for set DB tag for span
func SetSpanDBTag(span opentracing.Span, kind DBKind, value string) {
	if span == nil {
		return
	}
	switch kind {
	case DBInstance:
		ext.DBInstance.Set(span, value)
	case DBType:
		ext.DBType.Set(span, value)
	case DBUser:
		ext.DBUser.Set(span, value)
	case DBDatabase:
		DBTagDatabase.Set(span, value)
	case DBTable:
		DBTagTable.Set(span, value)
	case DbOperation:
		DBTagOperation.Set(span, value)
	default:
	}

	return
}

//////////////////////////////////////////////////////////////////////
// Error Tag
//////////////////////////////////////////////////////////////////////

// SetSpanErrorTag indicates that operation represented by the span resulted in an error.
func SetSpanErrorTag(span opentracing.Span, value bool) {
	if span == nil {
		return
	}

	ext.Error.Set(span, value)
}

//////////////////////////////////////////////////////////////////////
// Common Tag
//////////////////////////////////////////////////////////////////////

// SetSpanCommonTag set common tags for span
func SetSpanCommonTag(span opentracing.Span, key string, value interface{}) {
	if span == nil {
		return
	}

	span.SetTag(key, value)
}

//////////////////////////////////////////////////////////////////////
// Span Log
// log.String(key, val string) Field
// log.Bool(key string, val bool) Field
//////////////////////////////////////////////////////////////////////

// SetSpanLogTagError sets the error=true tag on the Span and logs err as an "error" event when error is not nil
func SetSpanLogTagError(span opentracing.Span, err error, fields ...log.Field) {
	if span == nil {
		return
	}
	SetSpanErrorTag(span, true)

	ef := []log.Field{
		log.Event("error"),
		log.Error(err),
	}
	ef = append(ef, fields...)
	span.LogFields(ef...)
}

// SetSpanLogMessage set span message log and other log fields(string, num, bool)
func SetSpanLogMessage(span opentracing.Span, value string, fields ...log.Field) {
	if span == nil {
		return
	}

	ef := []log.Field{
		log.Message(value),
	}

	ef = append(ef, fields...)
	span.LogFields(ef...)
}

// SetSpanLogError set span error log and other log fields(string, num, bool)
func SetSpanLogError(span opentracing.Span, err error, fields ...log.Field) {
	if span == nil {
		return
	}
	ef := []log.Field{
		log.Error(err),
	}

	ef = append(ef, fields...)
	span.LogFields(ef...)
}

// SetSpanLogEvent set span event log and other log fields(string, num, bool)
func SetSpanLogEvent(span opentracing.Span, value string, fields ...log.Field) {
	if span == nil {
		return
	}
	ef := []log.Field{
		log.Event(value),
	}

	ef = append(ef, fields...)
	span.LogFields(ef...)
}

// SetSpanLogObject set span object log and other log fields(string, num, bool)
func SetSpanLogObject(span opentracing.Span, key string, object interface{}, fields ...log.Field) {
	if span == nil {
		return
	}
	ef := []log.Field{
		log.Object(key, object),
	}

	ef = append(ef, fields...)
	span.LogFields(ef...)
}

// SetSpanLogFields set common log fields for span
func SetSpanLogFields(span opentracing.Span, fields ...log.Field) {
	if span == nil {
		return
	}

	span.LogFields(fields...)
}
