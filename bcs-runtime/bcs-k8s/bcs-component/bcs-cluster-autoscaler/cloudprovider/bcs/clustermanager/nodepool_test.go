package clustermanager

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// var (
// 	testNodeGroup = ""
// 	operator      = ""
// 	apiUrl        = ""
// 	token         = ""
// 	ips           = []string{""}
// )

// func TestGetPool(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}
// 	ng, err := client.GetPool(testNodeGroup)
// 	if err != nil {
// 		t.Errorf("GetPool failed. err: %v", err)
// 	}
// 	_, ok := reflect.ValueOf(ng).Interface().(*NodeGroup)
// 	if !ok {
// 		t.Errorf("GetPool returns bad values")
// 	}
// }

// func TestGetPoolConfig(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}
// 	config, err := client.GetPoolConfig(testNodeGroup)
// 	if err != nil {
// 		t.Errorf("GetPoolConfig failed. err: %v", err)
// 	}
// 	_, ok := reflect.ValueOf(config).Interface().(*AutoScalingGroup)
// 	if !ok {
// 		t.Errorf("GetPoolConfig returns bad values")
// 	}
// }

// func TestGetPoolNodeTemplate(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}
// 	template, err := client.GetPoolNodeTemplate(testNodeGroup)
// 	if err != nil {
// 		t.Errorf("GetPoolNodeTemplate failed. err: %v", err)
// 	}
// 	_, ok := reflect.ValueOf(template).Interface().(*LaunchConfiguration)
// 	if !ok {
// 		t.Errorf("GetPoolNodeTemplate returns bad values")
// 	}
// }

// func TestGetNodes(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}
// 	nodes, err := client.GetNodes(testNodeGroup)
// 	if err != nil {
// 		t.Errorf("GetNodes failed. err: %v", err)
// 	}
// 	_, ok := reflect.ValueOf(nodes).Interface().([]*Node)
// 	if !ok {
// 		t.Errorf("GetNodes returns bad values")
// 	}
// }

// func TestGetAutoScalingNodes(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}
// 	nodes, err := client.GetAutoScalingNodes(testNodeGroup)
// 	if err != nil {
// 		t.Errorf("GetAutoScalingNodes failed. err: %v", err)
// 	}
// 	_, ok := reflect.ValueOf(nodes).Interface().([]*Node)
// 	if !ok {
// 		t.Errorf("GetAutoScalingNodes returns bad values")
// 	}
// }

// func TestGetNode(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}
// 	node, err := client.GetNode(ips[0])
// 	if err != nil {
// 		t.Errorf("GetNodes failed. err: %v", err)
// 	}
// 	_, ok := reflect.ValueOf(node).Interface().(*Node)
// 	if !ok {
// 		t.Errorf("GetNodes returns bad values")
// 	}
// }

// func TestUpdateDesiredNode(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}
// 	err = client.UpdateDesiredNode(testNodeGroup, 2)
// 	if err != nil {
// 		t.Errorf("UpdateDesiredNode failed. err: %v", err)
// 	}
// }

// func TestRemoveNodes(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}

// 	err = client.RemoveNodes(testNodeGroup, ips)
// 	if err != nil {
// 		t.Errorf("RemoveNodes failed. err: %v", err)
// 	}
// }

// func TestUpdateDesiredSize(t *testing.T) {
// 	client, err := NewNodePoolClient(operator, apiUrl, token)
// 	if err != nil {
// 		t.Errorf("NewPoolClient failed. err: %v", err)
// 	}

// 	err = client.UpdateDesiredSize(testNodeGroup, 3)
// 	if err != nil {
// 		t.Errorf("UpdateDesiredSize failed. err: %v", err)
// 	}
// }

func TestNodePoolClient_GetPool(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test1" {
			res := GetNodeGroupResponse{
				Code: 0,
				Data: &NodeGroup{
					NodeGroupID: "test1",
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test2" {
			w.WriteHeader(404)
			res := GetNodeGroupResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test3" {
			res := GetNodeGroupResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()
	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *NodeGroup
		wantErr bool
	}{
		{
			name: "get pool normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test1",
			},
			want: &NodeGroup{
				NodeGroupID: "test1",
			},
			wantErr: false,
		},
		{
			name: "get pool, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test2",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get pool, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test3",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			got, err := npc.GetPool(tt.args.np)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.GetPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodePoolClient.GetPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodePoolClient_GetPoolConfig(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test1" {
			res := GetNodeGroupResponse{
				Code: 0,
				Data: &NodeGroup{
					NodeGroupID: "test1",
					AutoScaling: &AutoScalingGroup{
						MinSize: 0,
						MaxSize: 10,
					},
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test2" {
			w.WriteHeader(404)
			res := GetNodeGroupResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test3" {
			res := GetNodeGroupResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()
	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AutoScalingGroup
		wantErr bool
	}{
		{
			name: "get pool config normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test1",
			},
			want: &AutoScalingGroup{
				MinSize: 0,
				MaxSize: 10,
			},
			wantErr: false,
		},
		{
			name: "get pool config, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test2",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get pool config, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test3",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			got, err := npc.GetPoolConfig(tt.args.np)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.GetPoolConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodePoolClient.GetPoolConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodePoolClient_GetPoolNodeTemplate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test1" {
			res := GetNodeGroupResponse{
				Code: 0,
				Data: &NodeGroup{
					NodeGroupID: "test",
					LaunchTemplate: &LaunchConfiguration{
						CPU: 10,
						Mem: 1024,
					},
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test2" {
			w.WriteHeader(404)
			res := GetNodeGroupResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test3" {
			res := GetNodeGroupResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()
	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *LaunchConfiguration
		wantErr bool
	}{
		{
			name: "get pool node template normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test1",
			},
			want: &LaunchConfiguration{
				CPU: 10,
				Mem: 1024,
			},
			wantErr: false,
		},
		{
			name: "get pool node template, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test2",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get pool node template, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test3",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			got, err := npc.GetPoolNodeTemplate(tt.args.np)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.GetPoolNodeTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodePoolClient.GetPoolNodeTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodePoolClient_GetNodes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test1/node" {
			res := ListNodesInGroupResponse{
				Code: 0,
				Data: []*Node{
					&Node{
						NodeID:      "n1",
						NodeGroupID: "test1",
					},
					&Node{
						NodeID:      "n2",
						NodeGroupID: "test1",
					},
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test2/node" {
			w.WriteHeader(404)
			res := ListNodesInGroupResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test3/node" {
			res := ListNodesInGroupResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()
	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Node
		wantErr bool
	}{
		{
			name: "get nodes normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test1",
			},
			want: []*Node{
				&Node{
					NodeID:      "n1",
					NodeGroupID: "test1",
				},
				&Node{
					NodeID:      "n2",
					NodeGroupID: "test1",
				},
			},
			wantErr: false,
		},
		{
			name: "get nodes, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test2",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get nodes, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test3",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			got, err := npc.GetNodes(tt.args.np)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.GetNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodePoolClient.GetNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodePoolClient_GetAutoScalingNodes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test1" {
			res := GetNodeGroupResponse{
				Code: 0,
				Data: &NodeGroup{
					NodeGroupID: "test1",
					AutoScaling: &AutoScalingGroup{
						MinSize: 0,
						MaxSize: 10,
					},
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test2" {
			w.WriteHeader(404)
			res := GetNodeGroupResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test3" {
			res := GetNodeGroupResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test1/node" {
			res := ListNodesInGroupResponse{
				Code: 0,
				Data: []*Node{
					&Node{
						NodeID:      "n1",
						NodeGroupID: "test1",
					},
					&Node{
						NodeID:      "n2",
						NodeGroupID: "test1",
					},
					&Node{
						NodeID:      "n3",
						NodeGroupID: "",
					},
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test2/node" {
			w.WriteHeader(404)
			res := ListNodesInGroupResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test3/node" {
			res := ListNodesInGroupResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}

	}))
	defer ts.Close()

	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Node
		wantErr bool
	}{
		{
			name: "get autoscaling node normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test1",
			},
			want: []*Node{
				&Node{
					NodeID:      "n1",
					NodeGroupID: "test1",
				},
				&Node{
					NodeID:      "n2",
					NodeGroupID: "test1",
				},
			},
			wantErr: false,
		},
		{
			name: "get autoscaling node, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test2",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get autoscaling node, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np: "test3",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			got, err := npc.GetAutoScalingNodes(tt.args.np)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.GetAutoScalingNodes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodePoolClient.GetAutoScalingNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodePoolClient_GetNode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.EscapedPath() == "/node/127.0.0.1" {
			res := GetNodeResponse{
				Code: 0,
				Data: []*Node{
					&Node{
						NodeID:      "n1",
						NodeGroupID: "test1",
					},
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/node/127.0.0.2" {
			w.WriteHeader(404)
			res := GetNodeResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "GET" && r.URL.EscapedPath() == "/node/127.0.0.3" {
			res := GetNodeResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}

	}))
	defer ts.Close()
	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Node
		wantErr bool
	}{
		{
			name: "get node normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				ip: "127.0.0.1",
			},
			want: &Node{
				NodeID:      "n1",
				NodeGroupID: "test1",
			},
			wantErr: false,
		},
		{
			name: "get node, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				ip: "127.0.0.2",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "get node, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				ip: "127.0.0.3",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			got, err := npc.GetNode(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.GetNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NodePoolClient.GetNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNodePoolClient_UpdateDesiredNode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.EscapedPath() == "/nodegroup/test1/desirednode" {
			res := UpdateGroupDesiredNodeResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
			param := &UpdateGroupDesiredNodeRequest{}
			json.NewDecoder(r.Body).Decode(param)
			if param.DesiredNode != 5 {
				t.Errorf("Except rquest to have 'desirednode=5',got '%d'", param.DesiredNode)
			}
			if param.NodeGroupID != "test1" {
				t.Errorf("Except rquest to have 'nodegroupID=test1',got '%s'", param.NodeGroupID)
			}
		} else if r.Method == "POST" && r.URL.EscapedPath() == "/nodegroup/test2/desirednode" {
			w.WriteHeader(404)
			res := UpdateGroupDesiredNodeResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "POST" && r.URL.EscapedPath() == "/nodegroup/test3/desirednode" {
			res := UpdateGroupDesiredNodeResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()
	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np          string
		desiredNode int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update desired node normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:          "test1",
				desiredNode: 5,
			},
			wantErr: false,
		},
		{
			name: "update desired node, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:          "test2",
				desiredNode: 5,
			},
			wantErr: true,
		},
		{
			name: "update desired node, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:          "test3",
				desiredNode: 5,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			if err := npc.UpdateDesiredNode(tt.args.np, tt.args.desiredNode); (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.UpdateDesiredNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodePoolClient_RemoveNodes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.EscapedPath() == "/nodegroup/test1" {
			res := GetNodeGroupResponse{
				Code: 0,
				Data: &NodeGroup{
					NodeGroupID: "test1",
					ClusterID:   "c1",
				},
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "DELETE" && r.URL.EscapedPath() == "/nodegroup/test1/groupnode" {
			param := &CleanNodesInGroupRequest{}
			json.NewDecoder(r.Body).Decode(param)
			res := CleanNodesInGroupResponse{}
			if param.Nodes[0] == "127.0.0.1" {
				res.Code = 0
			} else if param.Nodes[0] == "127.0.0.2" {
				w.WriteHeader(404)
			} else if param.Nodes[0] == "127.0.0.3" {
				res.Code = 1
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}

	}))
	defer ts.Close()
	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np  string
		ips []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "remove node normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:  "test1",
				ips: []string{"127.0.0.1"},
			},
			wantErr: false,
		},
		{
			name: "remove node, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:  "test1",
				ips: []string{"127.0.0.2"},
			},
			wantErr: true,
		},
		{
			name: "remove node, return code 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:  "test1",
				ips: []string{"127.0.0.3"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			if err := npc.RemoveNodes(tt.args.np, tt.args.ips); (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.RemoveNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodePoolClient_UpdateDesiredSize(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.EscapedPath() == "/nodegroup/test1/desiredsize" {
			res := UpdateGroupDesiredSizeResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
			param := &UpdateGroupDesiredSizeRequest{}
			json.NewDecoder(r.Body).Decode(param)
			if param.DesiredSize != 5 {
				t.Errorf("Except rquest to have 'desiredSize=5',got '%d'", param.DesiredSize)
			}
		} else if r.Method == "POST" && r.URL.EscapedPath() == "/nodegroup/test2/desiredsize" {
			w.WriteHeader(404)
			res := UpdateGroupDesiredSizeResponse{
				Code: 0,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else if r.Method == "POST" && r.URL.EscapedPath() == "/nodegroup/test3/desiredsize" {
			res := UpdateGroupDesiredSizeResponse{
				Code: 1,
			}
			resBytes, _ := json.Marshal(res)
			w.Write(resBytes)
		} else {
			t.Errorf("Got unexpected acton '%v' and path '%v'", r.Method, r.URL.EscapedPath())
		}
	}))
	defer ts.Close()

	type fields struct {
		operator string
		url      string
		header   http.Header
	}
	type args struct {
		np          string
		desiredSize int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "update desired size normal",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:          "test1",
				desiredSize: 5,
			},
			wantErr: false,
		},
		{
			name: "update desired size, return 404",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:          "test2",
				desiredSize: 5,
			},
			wantErr: true,
		},
		{
			name: "update desired size, code is 1",
			fields: fields{
				operator: "bcs",
				url:      ts.URL,
				header: func() http.Header {
					tmp := make(http.Header)
					tmp.Add("Accept", "application/json")
					return tmp
				}(),
			},
			args: args{
				np:          "test3",
				desiredSize: 5,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			npc := &NodePoolClient{
				operator: tt.fields.operator,
				url:      tt.fields.url,
				header:   tt.fields.header,
			}
			if err := npc.UpdateDesiredSize(tt.args.np, tt.args.desiredSize); (err != nil) != tt.wantErr {
				t.Errorf("NodePoolClient.UpdateDesiredSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
