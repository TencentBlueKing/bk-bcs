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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// TotalAddWS xxx
	TotalAddWS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_add_websocket_session",
			Help:      "Total count of added websocket sessions",
		},
		[]string{"clientkey", "peer"})

	// TotalRemoveWS xxx
	TotalRemoveWS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_remove_websocket_session",
			Help:      "Total count of removed websocket sessions",
		},
		[]string{"clientkey", "peer"})

	// TotalAddConnectionsForWS xxx
	TotalAddConnectionsForWS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_add_connections",
			Help:      "Total count of added connections",
		},
		[]string{"clientkey", "proto", "addr"},
	)

	// TotalRemoveConnectionsForWS xxx
	TotalRemoveConnectionsForWS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_remove_connections",
			Help:      "Total count of removed connections",
		},
		[]string{"clientkey", "proto", "addr"},
	)

	// TotalTransmitBytesOnWS xxx
	TotalTransmitBytesOnWS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_transmit_bytes",
			Help:      "Total bytes transmitted",
		},
		[]string{"clientkey"},
	)

	// TotalTransmitErrorBytesOnWS xxx
	TotalTransmitErrorBytesOnWS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_transmit_error_bytes",
			Help:      "Total error bytes transmitted",
		},
		[]string{"clientkey"},
	)

	// TotalReceiveBytesOnWS xxx
	TotalReceiveBytesOnWS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_receive_bytes",
			Help:      "Total bytes received",
		},
		[]string{"clientkey"},
	)

	// TotalAddPeerAttempt xxx
	TotalAddPeerAttempt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_peer_ws_attempt",
			Help:      "Total count of attempts to establish websocket session to other bcs-api",
		},
		[]string{"peer"},
	)
	// TotalPeerConnected xxx
	TotalPeerConnected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_peer_ws_connected",
			Help:      "Total count of connected websocket sessions to other bcs-api",
		},
		[]string{"peer"},
	)
	// TotalPeerDisConnected xxx
	TotalPeerDisConnected = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "bcs_api",
			Subsystem: "session_server",
			Name:      "total_peer_ws_disconnected",
			Help:      "Total count of disconnected websocket sessions from other bcs-api",
		},
		[]string{"peer"},
	)
)

func init() {
	// Session metrics
	prometheus.MustRegister(TotalAddWS)
	prometheus.MustRegister(TotalRemoveWS)
	prometheus.MustRegister(TotalAddConnectionsForWS)
	prometheus.MustRegister(TotalRemoveConnectionsForWS)
	prometheus.MustRegister(TotalTransmitBytesOnWS)
	prometheus.MustRegister(TotalTransmitErrorBytesOnWS)
	prometheus.MustRegister(TotalReceiveBytesOnWS)
	prometheus.MustRegister(TotalAddPeerAttempt)
	prometheus.MustRegister(TotalPeerConnected)
	prometheus.MustRegister(TotalPeerDisConnected)
}

// IncSMTotalAddWS xxx
func IncSMTotalAddWS(clientKey string, peer bool) {
	var peerStr string
	if peer {
		peerStr = "true"
	} else {
		peerStr = "false"
	}

	TotalAddWS.With(
		prometheus.Labels{
			"clientkey": clientKey,
			"peer":      peerStr,
		}).Inc()
}

// IncSMTotalRemoveWS xxx
func IncSMTotalRemoveWS(clientKey string, peer bool) {
	var peerStr string
	if peer {
		peerStr = "true"
	} else {
		peerStr = "false"
	}
	TotalRemoveWS.With(
		prometheus.Labels{
			"clientkey": clientKey,
			"peer":      peerStr,
		}).Inc()
}

// AddSMTotalTransmitErrorBytesOnWS xxx
func AddSMTotalTransmitErrorBytesOnWS(clientKey string, size float64) {
	TotalTransmitErrorBytesOnWS.With(
		prometheus.Labels{
			"clientkey": clientKey,
		}).Add(size)
}

// AddSMTotalTransmitBytesOnWS xxx
func AddSMTotalTransmitBytesOnWS(clientKey string, size float64) {
	TotalTransmitBytesOnWS.With(
		prometheus.Labels{
			"clientkey": clientKey,
		}).Add(size)
}

// AddSMTotalReceiveBytesOnWS xxx
func AddSMTotalReceiveBytesOnWS(clientKey string, size float64) {
	TotalReceiveBytesOnWS.With(
		prometheus.Labels{
			"clientkey": clientKey,
		}).Add(size)
}

// IncSMTotalAddConnectionsForWS xxx
func IncSMTotalAddConnectionsForWS(clientKey, proto, addr string) {
	TotalAddConnectionsForWS.With(
		prometheus.Labels{
			"clientkey": clientKey,
			"proto":     proto,
			"addr":      addr,
		}).Inc()
}

// IncSMTotalRemoveConnectionsForWS xxx
func IncSMTotalRemoveConnectionsForWS(clientKey, proto, addr string) {
	TotalRemoveConnectionsForWS.With(
		prometheus.Labels{
			"clientkey": clientKey,
			"proto":     proto,
			"addr":      addr,
		}).Inc()
}

// IncSMTotalAddPeerAttempt xxx
func IncSMTotalAddPeerAttempt(peer string) {
	TotalAddPeerAttempt.With(
		prometheus.Labels{
			"peer": peer,
		}).Inc()
}

// IncSMTotalPeerConnected xxx
func IncSMTotalPeerConnected(peer string) {
	TotalPeerConnected.With(
		prometheus.Labels{
			"peer": peer,
		}).Inc()
}

// IncSMTotalPeerDisConnected xxx
func IncSMTotalPeerDisConnected(peer string) {
	TotalPeerDisConnected.With(
		prometheus.Labels{
			"peer": peer,
		}).Inc()

}
