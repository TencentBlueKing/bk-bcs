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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// This document defines the attributes used to perform database client calls.
const (
	// An identifier for the database management system (DBMS) product being used. See
	// below for a list of well-known identifiers.
	//
	// Type: Enum
	// Required: Always
	// Stability: stable
	DBSystemKey = attribute.Key("db.system")
	// The connection string used to connect to the database. It is recommended to
	// remove embedded credentials.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'Server=(localdb)\\v11.0;Integrated Security=true;'
	DBConnectionStringKey = attribute.Key("db.connection_string")
	// Username for accessing the database.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'readonly_user', 'reporting_user'
	DBUserKey = attribute.Key("db.user")
	// The fully-qualified class name of the [Java Database Connectivity
	// (JDBC)](https://docs.oracle.com/javase/8/docs/technotes/guides/jdbc/) driver
	// used to connect.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'org.postgresql.Driver',
	// 'com.microsoft.sqlserver.jdbc.SQLServerDriver'
	DBJDBCDriverClassnameKey = attribute.Key("db.jdbc.driver_classname")
	// If no [tech-specific attribute](#call-level-attributes-for-specific-
	// technologies) is defined, this attribute is used to report the name of the
	// database being accessed. For commands that switch the database, this should be
	// set to the target database (even if the command fails).
	//
	// Type: string
	// Required: Required, if applicable and no more-specific attribute is defined.
	// Stability: stable
	// Examples: 'customers', 'main'
	// Note: In some SQL databases, the database name to be used is called "schema
	// name".
	DBNameKey = attribute.Key("db.name")
	// The database statement being executed.
	//
	// Type: string
	// Required: Required if applicable and not explicitly disabled via
	// instrumentation configuration.
	// Stability: stable
	// Examples: 'SELECT * FROM wuser_table', 'SET mykey "WuValue"'
	// Note: The value may be sanitized to exclude sensitive information.
	DBStatementKey = attribute.Key("db.statement")
	// The name of the operation being executed, e.g. the [MongoDB command
	// name](https://docs.mongodb.com/manual/reference/command/#database-operations)
	// such as `findAndModify`, or the SQL keyword.
	//
	// Type: string
	// Required: Required, if `db.statement` is not applicable.
	// Stability: stable
	// Examples: 'findAndModify', 'HMSET', 'SELECT'
	// Note: When setting this to an SQL keyword, it is not recommended to attempt any
	// client-side parsing of `db.statement` just to get this property, but it should
	// be set if the operation name is provided by the library being instrumented. If
	// the SQL statement has an ambiguous operation, or performs more than one
	// operation, this value may be omitted.
	DBOperationKey = attribute.Key("db.operation")

	DBTableKey = attribute.Key("db.table")
)

// SetDBSpanAttributes sets DB tags for span
func SetDBSpanAttributes(span trace.Span, key attribute.Key, value string) {
	if span == nil {
		return
	}
	switch key {
	case DBSystemKey:
		span.SetAttributes(DBSystemKey.String(value))
	case DBConnectionStringKey:
		span.SetAttributes(DBConnectionStringKey.String(value))
	case DBStatementKey:
		span.SetAttributes(DBStatementKey.String(value))
	case DBJDBCDriverClassnameKey:
		span.SetAttributes(DBJDBCDriverClassnameKey.String(value))
	case DBUserKey:
		span.SetAttributes(DBUserKey.String(value))
	case DBNameKey:
		span.SetAttributes(DBNameKey.String(value))
	case DBOperationKey:
		span.SetAttributes(DBOperationKey.String(value))
	case DBTableKey:
		span.SetAttributes(DBTableKey.String(value))
	default:
	}
	return
}

// This document defines semantic conventions for HTTP client and server Spans.
const (
	// HTTP request method.
	//
	// Type: string
	// Required: Always
	// Stability: stable
	// Examples: 'GET', 'POST', 'HEAD'
	HTTPMethodKey = attribute.Key("http.method")
	// Full HTTP request URL in the form `scheme://host[:port]/path?query[#fragment]`.
	// Usually the fragment is not transmitted over HTTP, but if it is known, it
	// should be included nevertheless.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'https://www.foo.bar/search?q=OpenTelemetry#SemConv'
	// Note: `http.url` MUST NOT contain credentials passed via URL in form of
	// `https://username:password@www.example.com/`. In such case the attribute's
	// value should be `https://www.example.com/`.
	HTTPURLKey = attribute.Key("http.url")
	// The full request target as passed in a HTTP request line or equivalent.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: '/path/12314/?q=ddds#123'
	HTTPTargetKey = attribute.Key("http.target")
	// The value of the [HTTP host
	// header](https://tools.ietf.org/html/rfc7230#section-5.4). An empty Host header
	// should also be reported, see note.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'www.example.org'
	// Note: When the header is present but empty the attribute SHOULD be set to the
	// empty string. Note that this is a valid situation that is expected in certain
	// cases, according the aforementioned [section of RFC
	// 7230](https://tools.ietf.org/html/rfc7230#section-5.4). When the header is not
	// set the attribute MUST NOT be set.
	HTTPHostKey = attribute.Key("http.host")
	// The URI scheme identifying the used protocol.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'http', 'https'
	HTTPSchemeKey = attribute.Key("http.scheme")
	// [HTTP response status code](https://tools.ietf.org/html/rfc7231#section-6).
	//
	// Type: int
	// Required: If and only if one was received/sent.
	// Stability: stable
	// Examples: 200
	HTTPStatusCodeKey = attribute.Key("http.status_code")
	// Kind of HTTP protocol used.
	//
	// Type: Enum
	// Required: No
	// Stability: stable
	// Note: If `net.transport` is not specified, it can be assumed to be `IP.TCP`
	// except if `http.flavor` is `QUIC`, in which case `IP.UDP` is assumed.
	HTTPFlavorKey = attribute.Key("http.flavor")
	// Value of the [HTTP User-
	// Agent](https://tools.ietf.org/html/rfc7231#section-5.5.3) header sent by the
	// client.
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'CERN-LineMode/2.15 libwww/2.17b3'
	HTTPUserAgentKey = attribute.Key("http.user_agent")
	// The size of the request payload body in bytes. This is the number of bytes
	// transferred excluding headers and is often, but not always, present as the
	// [Content-Length](https://tools.ietf.org/html/rfc7230#section-3.3.2) header. For
	// requests using transport encoding, this should be the compressed size.
	//
	// Type: int
	// Required: No
	// Stability: stable
	// Examples: 3495
	HTTPRequestContentLengthKey = attribute.Key("http.request_content_length")
	// The size of the uncompressed request payload body after transport decoding. Not
	// set if transport encoding not used.
	//
	// Type: int
	// Required: No
	// Stability: stable
	// Examples: 5493
	HTTPRequestContentLengthUncompressedKey = attribute.Key("http.request_content_length_uncompressed")
	// The size of the response payload body in bytes. This is the number of bytes
	// transferred excluding headers and is often, but not always, present as the
	// [Content-Length](https://tools.ietf.org/html/rfc7230#section-3.3.2) header. For
	// requests using transport encoding, this should be the compressed size.
	//
	// Type: int
	// Required: No
	// Stability: stable
	// Examples: 3495
	HTTPResponseContentLengthKey = attribute.Key("http.response_content_length")
	// The size of the uncompressed response payload body after transport decoding.
	// Not set if transport encoding not used.
	//
	// Type: int
	// Required: No
	// Stability: stable
	// Examples: 5493
	HTTPResponseContentLengthUncompressedKey = attribute.Key("http.response_content_length_uncompressed")
)

var (
	// HTTP 1.0
	HTTPFlavorHTTP10 = HTTPFlavorKey.String("1.0")
	// HTTP 1.1
	HTTPFlavorHTTP11 = HTTPFlavorKey.String("1.1")
	// HTTP 2
	HTTPFlavorHTTP20 = HTTPFlavorKey.String("2.0")
	// SPDY protocol
	HTTPFlavorSPDY = HTTPFlavorKey.String("SPDY")
	// QUIC protocol
	HTTPFlavorQUIC = HTTPFlavorKey.String("QUIC")
)

// Semantic Convention for HTTP Server
const (
	// The primary server name of the matched virtual host. This should be obtained
	// via configuration. If no such configuration can be obtained, this attribute
	// MUST NOT be set ( `net.host.name` should be used instead).
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: 'example.com'
	// Note: `http.url` is usually not readily available on the server side but would
	// have to be assembled in a cumbersome and sometimes lossy process from other
	// information (see e.g. open-telemetry/opentelemetry-python/pull/148). It is thus
	// preferred to supply the raw data that is available.
	HTTPServerNameKey = attribute.Key("http.server_name")
	// The matched route (path template).
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: '/users/:userID?'
	HTTPRouteKey = attribute.Key("http.route")
	// The IP address of the original client behind all proxies, if known (e.g. from
	// [X-Forwarded-For](https://developer.mozilla.org/en-
	// US/docs/Web/HTTP/Headers/X-Forwarded-For)).
	//
	// Type: string
	// Required: No
	// Stability: stable
	// Examples: '83.164.160.102'
	// Note: This is not necessarily the same as `net.peer.ip`, which would
	// identify the network-level peer, which may be a proxy.

	// This attribute should be set when a source of information different
	// from the one used for `net.peer.ip`, is available even if that other
	// source just confirms the same value as `net.peer.ip`.
	// Rationale: For `net.peer.ip`, one typically does not know if it
	// comes from a proxy, reverse proxy, or the actual client. Setting
	// `http.client_ip` when it's the same as `net.peer.ip` means that
	// one is at least somewhat confident that the address is not that of
	// the closest proxy.
	HTTPClientIPKey = attribute.Key("http.client_ip")

	HTTPHandlerKey = attribute.Key("handler")
)

// SetHTTPSpanAttributes sets HTTP tags for span
func SetHTTPSpanAttributes(span trace.Span, key attribute.Key, value string) {
	if span == nil {
		return
	}
	switch key {
	case HTTPMethodKey:
		span.SetAttributes(HTTPMethodKey.String(value))
	case HTTPURLKey:
		span.SetAttributes(HTTPURLKey.String(value))
	case HTTPTargetKey:
		span.SetAttributes(HTTPTargetKey.String(value))
	case HTTPHostKey:
		span.SetAttributes(HTTPHostKey.String(value))
	case HTTPSchemeKey:
		span.SetAttributes(HTTPSchemeKey.String(value))
	case HTTPStatusCodeKey:
		span.SetAttributes(HTTPStatusCodeKey.String(value))
	case HTTPHandlerKey:
		span.SetAttributes(HTTPHandlerKey.String(value))
	default:
	}
	return
}
