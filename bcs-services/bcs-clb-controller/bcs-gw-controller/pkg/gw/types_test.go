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

package gw

import (
	"testing"
)

func TestLocationForwardStrategyDiff(t *testing.T) {
	tests := []struct {
		refLFS    *LocationForwardStrategy
		diffLFS   *LocationForwardStrategy
		refResult bool
	}{
		{
			refLFS:    nil,
			diffLFS:   nil,
			refResult: false,
		},
		{
			refLFS: &LocationForwardStrategy{
				Type:   "type1",
				Detail: "detail1",
			},
			diffLFS:   nil,
			refResult: true,
		},
		{
			refLFS: nil,
			diffLFS: &LocationForwardStrategy{
				Type:   "type1",
				Detail: "detail1",
			},
			refResult: true,
		},
		{
			refLFS: &LocationForwardStrategy{
				Type:   "type1",
				Detail: "detail1",
			},
			diffLFS: &LocationForwardStrategy{
				Type:   "type2",
				Detail: "detail1",
			},
			refResult: true,
		},
		{
			refLFS: &LocationForwardStrategy{
				Type:   "type1",
				Detail: "detail1",
			},
			diffLFS: &LocationForwardStrategy{
				Type:   "type1",
				Detail: "detail1",
			},
			refResult: false,
		},
	}
	for _, test := range tests {
		re := test.refLFS.Diff(test.diffLFS)
		if re != test.refResult {
			t.Errorf("expect %v but get %v", test.refResult, re)
		}
	}
}

func TestLocationSessionPersistenceDiff(t *testing.T) {
	tests := []struct {
		refLsp    *LocationSessionPersistence
		diffLsp   *LocationSessionPersistence
		refResult bool
	}{
		{
			refLsp:    nil,
			diffLsp:   nil,
			refResult: false,
		},
		{
			refLsp: &LocationSessionPersistence{
				Type:           "type1",
				CookieTimeMode: 1,
				Timeout:        1,
				CookieKey:      "cookie1",
			},
			diffLsp:   nil,
			refResult: true,
		},
		{
			refLsp: &LocationSessionPersistence{
				Type:           "type1",
				CookieTimeMode: 1,
				Timeout:        1,
				CookieKey:      "cookie1",
			},
			diffLsp: &LocationSessionPersistence{
				Type:           "type2",
				CookieTimeMode: 1,
				Timeout:        1,
				CookieKey:      "cookie1",
			},
			refResult: true,
		},
		{
			refLsp: &LocationSessionPersistence{
				Type:           "type1",
				CookieTimeMode: 1,
				Timeout:        1,
				CookieKey:      "cookie1",
			},
			diffLsp: &LocationSessionPersistence{
				Type:           "type1",
				CookieTimeMode: 1,
				Timeout:        1,
				CookieKey:      "cookie1",
			},
			refResult: false,
		},
	}
	for _, test := range tests {
		re := test.refLsp.Diff(test.diffLsp)
		if re != test.refResult {
			t.Errorf("expect %v but get %v", test.refResult, re)
		}
	}
}

func TestLocationDiff(t *testing.T) {
	tests := []struct {
		refLocation  *Location
		diffLocation *Location
		refResult    bool
	}{
		{
			refLocation:  nil,
			diffLocation: nil,
			refResult:    false,
		},
		{
			refLocation: &Location{
				LocationID:             "id1",
				URL:                    "/",
				LocationCustomizedConf: "",
				LocLimitRate:           10,
				LocLimitStatusCode:     10,
				ForwardStrategy: &LocationForwardStrategy{
					Type:   "type1",
					Detail: "detail1",
				},
				SessionPersistence: &LocationSessionPersistence{
					Type:           "type1",
					CookieTimeMode: 1,
					Timeout:        1,
					CookieKey:      "cooke1",
				},
				HealthCheck: &LocationHealthCheck{
					OP:            "op1",
					Protocol:      "http",
					AliveNum:      1,
					KickNum:       1,
					ProbeInterval: 10,
					AliveCode:     200,
					ProbeURL:      "/",
					Method:        "POST",
					ServerName:    "www.test.com",
				},
				Rewrite: &LocationRewrite{
					OP:   "op1",
					Type: "type1",
					URL:  "url1",
				},
				RSList: []*RealServer{
					{
						IP:     "127.0.0.1",
						Port:   8080,
						Weight: 100,
					},
					{
						IP:     "127.0.0.2",
						Port:   8080,
						Weight: 100,
					},
					{
						IP:     "127.0.0.2",
						Port:   8081,
						Weight: 100,
					},
				},
			},
			diffLocation: &Location{
				LocationID:             "id1",
				URL:                    "/",
				LocationCustomizedConf: "",
				LocLimitRate:           10,
				LocLimitStatusCode:     10,
				ForwardStrategy: &LocationForwardStrategy{
					Type:   "type1",
					Detail: "detail1",
				},
				SessionPersistence: &LocationSessionPersistence{
					Type:           "type1",
					CookieTimeMode: 1,
					Timeout:        1,
					CookieKey:      "cooke1",
				},
				HealthCheck: &LocationHealthCheck{
					OP:            "op1",
					Protocol:      "http",
					AliveNum:      1,
					KickNum:       1,
					ProbeInterval: 10,
					AliveCode:     200,
					ProbeURL:      "/",
					Method:        "POST",
					ServerName:    "www.test.com",
				},
				Rewrite: &LocationRewrite{
					OP:   "op1",
					Type: "type1",
					URL:  "url1",
				},
				RSList: []*RealServer{
					{
						IP:     "127.0.0.2",
						Port:   8081,
						Weight: 100,
					},
					{
						IP:     "127.0.0.1",
						Port:   8080,
						Weight: 100,
					},
					{
						IP:     "127.0.0.2",
						Port:   8080,
						Weight: 100,
					},
				},
			},
			refResult: false,
		},
		{
			refLocation: &Location{
				RSList: []*RealServer{
					{
						IP:     "127.0.0.1",
						Port:   8080,
						Weight: 100,
					},
					{
						IP:     "127.0.0.2",
						Port:   8080,
						Weight: 100,
					},
					{
						IP:     "127.0.0.2",
						Port:   8081,
						Weight: 100,
					},
				},
			},
			diffLocation: &Location{
				RSList: []*RealServer{
					{
						IP:     "127.0.0.2",
						Port:   8081,
						Weight: 100,
					},
					{
						IP:     "127.0.0.1",
						Port:   8080,
						Weight: 10,
					},
					{
						IP:     "127.0.0.2",
						Port:   8080,
						Weight: 100,
					},
				},
			},
			refResult: true,
		},
	}
	for _, test := range tests {
		re := test.refLocation.Diff(test.diffLocation)
		if re != test.refResult {
			t.Errorf("expect %v but get %v", test.refResult, re)
		}
	}
}

func TestServiceDiff(t *testing.T) {
	tests := []struct {
		refService  *Service
		diffService *Service
		refResult   bool
	}{
		{
			refService:  nil,
			diffService: nil,
			refResult:   false,
		},
		{
			refService: &Service{
				BizID:                   "id1",
				VIPList:                 []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
				Domain:                  "www.test1.com",
				VPort:                   8080,
				VpcID:                   1,
				Type:                    "HTTP",
				SSLEnable:               true,
				SSLVerifyClientEnable:   true,
				CertID:                  "xxxxx",
				DefaultServer:           true,
				ServerCustomizedConf:    "",
				VIPProtoLimitRate:       10,
				VIPProtoLimitStatusCode: 200,
				VSLimitRate:             20,
				VSLimitStatusCode:       200,
				LocationList: []*Location{
					{
						LocationID:             "id1",
						URL:                    "/",
						LocationCustomizedConf: "",
						LocLimitRate:           10,
						LocLimitStatusCode:     10,
						ForwardStrategy: &LocationForwardStrategy{
							Type:   "type1",
							Detail: "detail1",
						},
						SessionPersistence: &LocationSessionPersistence{
							Type:           "type1",
							CookieTimeMode: 1,
							Timeout:        1,
							CookieKey:      "cooke1",
						},
						HealthCheck: &LocationHealthCheck{
							OP:            "op1",
							Protocol:      "http",
							AliveNum:      1,
							KickNum:       1,
							ProbeInterval: 10,
							AliveCode:     200,
							ProbeURL:      "/",
							Method:        "POST",
							ServerName:    "www.test.com",
						},
						Rewrite: &LocationRewrite{
							OP:   "op1",
							Type: "type1",
							URL:  "url1",
						},
						RSList: []*RealServer{
							{
								IP:     "127.0.0.1",
								Port:   8080,
								Weight: 100,
							},
							{
								IP:     "127.0.0.2",
								Port:   8080,
								Weight: 100,
							},
							{
								IP:     "127.0.0.2",
								Port:   8081,
								Weight: 100,
							},
						},
					},
					{
						LocationID:             "id1",
						URL:                    "/path2",
						LocationCustomizedConf: "",
						LocLimitRate:           10,
						LocLimitStatusCode:     10,
						ForwardStrategy: &LocationForwardStrategy{
							Type:   "type1",
							Detail: "detail1",
						},
						SessionPersistence: &LocationSessionPersistence{
							Type:           "type1",
							CookieTimeMode: 1,
							Timeout:        1,
							CookieKey:      "cooke1",
						},
						HealthCheck: &LocationHealthCheck{
							OP:            "op1",
							Protocol:      "http",
							AliveNum:      1,
							KickNum:       1,
							ProbeInterval: 10,
							AliveCode:     200,
							ProbeURL:      "/",
							Method:        "POST",
							ServerName:    "www.test.com",
						},
						Rewrite: &LocationRewrite{
							OP:   "op1",
							Type: "type1",
							URL:  "url1",
						},
						RSList: []*RealServer{
							{
								IP:     "127.0.0.1",
								Port:   8080,
								Weight: 100,
							},
							{
								IP:     "127.0.0.2",
								Port:   8080,
								Weight: 100,
							},
							{
								IP:     "127.0.0.2",
								Port:   8081,
								Weight: 100,
							},
						},
					},
				},
			},
			diffService: &Service{
				BizID:                   "id1",
				VIPList:                 []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
				Domain:                  "www.test1.com",
				VPort:                   8080,
				VpcID:                   1,
				Type:                    "HTTP",
				SSLEnable:               true,
				SSLVerifyClientEnable:   true,
				CertID:                  "xxxxx",
				DefaultServer:           true,
				ServerCustomizedConf:    "",
				VIPProtoLimitRate:       10,
				VIPProtoLimitStatusCode: 200,
				VSLimitRate:             20,
				VSLimitStatusCode:       200,
				LocationList: []*Location{
					{
						LocationID:             "id1",
						URL:                    "/path2",
						LocationCustomizedConf: "",
						LocLimitRate:           10,
						LocLimitStatusCode:     10,
						ForwardStrategy: &LocationForwardStrategy{
							Type:   "type1",
							Detail: "detail1",
						},
						SessionPersistence: &LocationSessionPersistence{
							Type:           "type1",
							CookieTimeMode: 1,
							Timeout:        1,
							CookieKey:      "cooke1",
						},
						HealthCheck: &LocationHealthCheck{
							OP:            "op1",
							Protocol:      "http",
							AliveNum:      1,
							KickNum:       1,
							ProbeInterval: 10,
							AliveCode:     200,
							ProbeURL:      "/",
							Method:        "POST",
							ServerName:    "www.test.com",
						},
						Rewrite: &LocationRewrite{
							OP:   "op1",
							Type: "type1",
							URL:  "url1",
						},
						RSList: []*RealServer{
							{
								IP:     "127.0.0.2",
								Port:   8080,
								Weight: 100,
							},
							{
								IP:     "127.0.0.2",
								Port:   8081,
								Weight: 100,
							},
							{
								IP:     "127.0.0.1",
								Port:   8080,
								Weight: 100,
							},
						},
					},
					{
						LocationID:             "id1",
						URL:                    "/",
						LocationCustomizedConf: "",
						LocLimitRate:           10,
						LocLimitStatusCode:     10,
						ForwardStrategy: &LocationForwardStrategy{
							Type:   "type1",
							Detail: "detail1",
						},
						SessionPersistence: &LocationSessionPersistence{
							Type:           "type1",
							CookieTimeMode: 1,
							Timeout:        1,
							CookieKey:      "cooke1",
						},
						HealthCheck: &LocationHealthCheck{
							OP:            "op1",
							Protocol:      "http",
							AliveNum:      1,
							KickNum:       1,
							ProbeInterval: 10,
							AliveCode:     200,
							ProbeURL:      "/",
							Method:        "POST",
							ServerName:    "www.test.com",
						},
						Rewrite: &LocationRewrite{
							OP:   "op1",
							Type: "type1",
							URL:  "url1",
						},
						RSList: []*RealServer{
							{
								IP:     "127.0.0.2",
								Port:   8081,
								Weight: 100,
							},
							{
								IP:     "127.0.0.1",
								Port:   8080,
								Weight: 100,
							},
							{
								IP:     "127.0.0.2",
								Port:   8080,
								Weight: 100,
							},
						},
					},
				},
			},
			refResult: false,
		},
		{
			refService: &Service{
				BizID:                   "id1",
				VIPList:                 []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
				Domain:                  "www.test1.com",
				VPort:                   8080,
				VpcID:                   1,
				Type:                    "HTTP",
				SSLEnable:               true,
				SSLVerifyClientEnable:   true,
				CertID:                  "xxxxx",
				DefaultServer:           true,
				ServerCustomizedConf:    "",
				VIPProtoLimitRate:       10,
				VIPProtoLimitStatusCode: 200,
				VSLimitRate:             20,
				VSLimitStatusCode:       200,
				LocationList: []*Location{
					{
						LocationID:             "id1",
						URL:                    "/",
						LocationCustomizedConf: "",
						LocLimitRate:           10,
						LocLimitStatusCode:     10,
						ForwardStrategy: &LocationForwardStrategy{
							Type:   "type1",
							Detail: "detail1",
						},
						SessionPersistence: &LocationSessionPersistence{
							Type:           "type1",
							CookieTimeMode: 1,
							Timeout:        1,
							CookieKey:      "cooke1",
						},
						HealthCheck: &LocationHealthCheck{
							OP:            "op1",
							Protocol:      "http",
							AliveNum:      1,
							KickNum:       1,
							ProbeInterval: 10,
							AliveCode:     200,
							ProbeURL:      "/",
							Method:        "POST",
							ServerName:    "www.test.com",
						},
						Rewrite: &LocationRewrite{
							OP:   "op1",
							Type: "type1",
							URL:  "url1",
						},
						RSList: []*RealServer{
							{
								IP:     "127.0.0.1",
								Port:   8080,
								Weight: 100,
							},
							{
								IP:     "127.0.0.2",
								Port:   8081,
								Weight: 100,
							},
						},
					},
				},
			},
			diffService: &Service{
				BizID:                   "id1",
				VIPList:                 []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
				Domain:                  "www.test1.com",
				VPort:                   8080,
				VpcID:                   1,
				Type:                    "HTTP",
				SSLEnable:               true,
				SSLVerifyClientEnable:   true,
				CertID:                  "xxxxx",
				DefaultServer:           true,
				ServerCustomizedConf:    "",
				VIPProtoLimitRate:       10,
				VIPProtoLimitStatusCode: 200,
				VSLimitRate:             20,
				VSLimitStatusCode:       300,
				LocationList: []*Location{
					{
						LocationID:             "id1",
						URL:                    "/",
						LocationCustomizedConf: "",
						LocLimitRate:           10,
						LocLimitStatusCode:     10,
						ForwardStrategy: &LocationForwardStrategy{
							Type:   "type1",
							Detail: "detail1",
						},
						SessionPersistence: &LocationSessionPersistence{
							Type:           "type1",
							CookieTimeMode: 1,
							Timeout:        1,
							CookieKey:      "cooke1",
						},
						HealthCheck: &LocationHealthCheck{
							OP:            "op1",
							Protocol:      "http",
							AliveNum:      1,
							KickNum:       1,
							ProbeInterval: 10,
							AliveCode:     200,
							ProbeURL:      "/",
							Method:        "POST",
							ServerName:    "www.test.com",
						},
						Rewrite: &LocationRewrite{
							OP:   "op1",
							Type: "type1",
							URL:  "url1",
						},
						RSList: []*RealServer{
							{
								IP:     "127.0.0.2",
								Port:   8081,
								Weight: 100,
							},
							{
								IP:     "127.0.0.1",
								Port:   8080,
								Weight: 100,
							},
						},
					},
				},
			},
			refResult: true,
		},
	}

	for _, test := range tests {
		re := test.refService.Diff(test.diffService)
		if re != test.refResult {
			t.Errorf("expect %v but get %v", test.refResult, re)
		}
	}
}
