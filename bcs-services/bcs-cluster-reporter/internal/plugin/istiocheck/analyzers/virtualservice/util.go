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

package virtualservice

import (
	"istio.io/api/networking/v1alpha3"
)

// AnnotatedDestination holds metadata about a Destination object that is used for analyzing
type AnnotatedDestination struct {
	RouteRule        string
	ServiceIndex     int
	DestinationIndex int
	Destination      *v1alpha3.Destination
}

func getRouteDestinations(vs *v1alpha3.VirtualService) []*AnnotatedDestination {
	destinations := make([]*AnnotatedDestination, 0)
	for i, r := range vs.GetTcp() {
		for j, rd := range r.GetRoute() {
			destinations = append(destinations, &AnnotatedDestination{
				RouteRule:        "tcp",
				ServiceIndex:     i,
				DestinationIndex: j,
				Destination:      rd.GetDestination(),
			})
		}
	}
	for i, r := range vs.GetTls() {
		for j, rd := range r.GetRoute() {
			destinations = append(destinations, &AnnotatedDestination{
				RouteRule:        "tls",
				ServiceIndex:     i,
				DestinationIndex: j,
				Destination:      rd.GetDestination(),
			})
		}
	}
	for i, r := range vs.GetHttp() {
		for j, rd := range r.GetRoute() {
			destinations = append(destinations, &AnnotatedDestination{
				RouteRule:        "http",
				ServiceIndex:     i,
				DestinationIndex: j,
				Destination:      rd.GetDestination(),
			})
		}
	}

	return destinations
}

func getHTTPMirrorDestinations(vs *v1alpha3.VirtualService) []*AnnotatedDestination {
	var destinations []*AnnotatedDestination

	for i, r := range vs.GetHttp() {
		if m := r.GetMirror(); m != nil {
			destinations = append(destinations, &AnnotatedDestination{
				RouteRule:    "http.mirror",
				ServiceIndex: i,
				Destination:  m,
			})
		}
	}

	return destinations
}
