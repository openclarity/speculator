// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pathtrie

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"gotest.tools/assert"
)

func TestPathTrie_getNode(t *testing.T) {
	pt := New()
	assert.Equal(t, pt.Insert("/api/{param1}/items", 1), true)
	assert.Equal(t, pt.Insert("/api/items", 2), true)
	assert.Equal(t, pt.Insert("/api/{param1}/{param2}", 3), true)
	assert.Equal(t, pt.Insert("/api/{param1}/cat", 4), true)
	assert.Equal(t, pt.Insert("/api/items/cat", 5), true)
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want *TrieNode
	}{
		{
			name: "most accurate match - will match both `/api/{param1}/items` and `/api/{param1}/{param2}`",
			args: args{
				path: "/api/1/items",
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "items",
				FullPath:         "/api/{param1}/items",
				PathParamCounter: 1,
				Value:            1,
			},
		},
		{
			name: "exact match with path param",
			args: args{
				path: "/api/{param1}/items",
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "items",
				FullPath:         "/api/{param1}/items",
				PathParamCounter: 1,
				Value:            1,
			},
		},
		{
			name: "short match - not continue to `/api/items/cat`",
			args: args{
				path: "/api/items",
			},
			want: &TrieNode{
				Children: map[string]*TrieNode{
					"cat": {
						Children:         make(PathToTrieNode),
						Name:             "cat",
						FullPath:         "/api/items/cat",
						PathParamCounter: 0,
						Value:            5,
					},
				},
				Name:             "items",
				FullPath:         "/api/items",
				PathParamCounter: 0,
				Value:            2,
			},
		},
		{
			name: "simple path param match",
			args: args{
				path: "/api/1/2",
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "{param2}",
				FullPath:         "/api/{param1}/{param2}",
				PathParamCounter: 2,
				Value:            3,
			},
		},
		{
			name: "most accurate match - will match both `/api/{param1}/cat` and `/api/{param1}/{param2}`",
			args: args{
				path: "/api/1/cat",
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "cat",
				FullPath:         "/api/{param1}/cat",
				PathParamCounter: 1,
				Value:            4,
			},
		},
		{
			name: "exact match with no path param",
			args: args{
				path: "/api/items/cat",
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "cat",
				FullPath:         "/api/items/cat",
				PathParamCounter: 0,
				Value:            5,
			},
		},
		{
			name: "no match",
			args: args{
				path: "api/items/cat",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pt.getNode(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTrie_GetValue(t *testing.T) {
	pt := New()
	assert.Equal(t, pt.Insert("/api/{param1}/items", 1), true)
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "match",
			args: args{
				path: "/api/1/items",
			},
			want: 1,
		},
		{
			name: "no match",
			args: args{
				path: "api/items/cat",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pt.GetValue(tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTrie_GetPathAndValue(t *testing.T) {
	pt := New()
	assert.Equal(t, pt.Insert("/api/{param1}/items", 1), true)
	type args struct {
		path string
	}
	tests := []struct {
		name      string
		args      args
		wantPath  string
		wantValue interface{}
		wantFound bool
	}{
		{
			name: "match",
			args: args{
				path: "/api/1/items",
			},
			wantPath:  "/api/{param1}/items",
			wantValue: 1,
			wantFound: true,
		},
		{
			name: "no match",
			args: args{
				path: "api/items/cat",
			},
			wantPath:  "",
			wantValue: nil,
			wantFound: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotValue, gotFound := pt.GetPathAndValue(tt.args.path)
			if gotPath != tt.wantPath {
				t.Errorf("GetPathAndValue() gotPath = %v, wantPath %v", gotPath, tt.wantPath)
			}
			if !reflect.DeepEqual(gotValue, tt.wantValue) {
				t.Errorf("GetPathAndValue() gotValue = %v, wantValue %v", gotValue, tt.wantValue)
			}
			if gotFound != tt.wantFound {
				t.Errorf("GetPathAndValue() gotFound = %v, wantFound %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestPathTrieMap_getMatchNodes(t *testing.T) {
	type args struct {
		segments []string
		idx      int
	}
	tests := []struct {
		name string
		trie PathToTrieNode
		args args
		want []*TrieNode
	}{
		{
			name: "return 2 matches nodes",
			trie: map[string]*TrieNode{
				"api": {
					Children: map[string]*TrieNode{
						"{param1}": {
							Children: map[string]*TrieNode{
								"test": {
									Children:         make(PathToTrieNode),
									Name:             "test",
									FullPath:         "/api/{param1}/test",
									PathParamCounter: 1,
									Value:            1,
								},
								"{param2}": {
									Children:         make(PathToTrieNode),
									Name:             "{param2}",
									FullPath:         "/api/{param1}/{param2}",
									PathParamCounter: 2,
									Value:            2,
								},
							},
							Name:             "{param1}",
							FullPath:         "/api/{param1}",
							PathParamCounter: 1,
						},
					},
					Name:     "api",
					FullPath: "/api",
				},
			},
			args: args{
				segments: []string{"api", "123", "test"},
				idx:      0,
			},
			want: []*TrieNode{
				{
					Children:         make(PathToTrieNode),
					Name:             "test",
					FullPath:         "/api/{param1}/test",
					PathParamCounter: 1,
					Value:            1,
				},
				{
					Children:         make(PathToTrieNode),
					Name:             "{param2}",
					FullPath:         "/api/{param1}/{param2}",
					PathParamCounter: 2,
					Value:            2,
				},
			},
		},
		{
			name: "last path segment has nil value - return only 1 matches nodes (/api/{param1}/{param2})",
			trie: map[string]*TrieNode{
				"api": {
					Children: map[string]*TrieNode{
						"{param1}": {
							Children: map[string]*TrieNode{
								"test": {
									Children: map[string]*TrieNode{
										"cats": {
											Children:         make(PathToTrieNode),
											Name:             "cats",
											FullPath:         "/api/{param1}/test/cats",
											PathParamCounter: 1,
											Value:            1,
										},
									},
									Name:             "test",
									FullPath:         "/api/{param1}/test",
									PathParamCounter: 1,
									Value:            nil,
								},
								"{param2}": {
									Children:         make(PathToTrieNode),
									Name:             "{param2}",
									FullPath:         "/api/{param1}/{param2}",
									PathParamCounter: 2,
									Value:            2,
								},
							},
							Name:             "{param1}",
							FullPath:         "/api/{param1}",
							PathParamCounter: 1,
						},
					},
					Name:     "api",
					FullPath: "/api",
				},
			},
			args: args{
				segments: []string{"api", "123", "test"},
				idx:      0,
			},
			want: []*TrieNode{
				{
					Children:         make(PathToTrieNode),
					Name:             "{param2}",
					FullPath:         "/api/{param1}/{param2}",
					PathParamCounter: 2,
					Value:            2,
				},
			},
		},
		{
			name: "0 nodes match",
			trie: map[string]*TrieNode{
				"api": {
					Children: map[string]*TrieNode{
						"{param1}": {
							Children: map[string]*TrieNode{
								"test": {
									Children:         make(PathToTrieNode),
									Name:             "test",
									FullPath:         "/api/{param1}/test",
									PathParamCounter: 1,
									Value:            1,
								},
								"{param2}": {
									Children:         make(PathToTrieNode),
									Name:             "{param2}",
									FullPath:         "/api/{param1}/{param2}",
									PathParamCounter: 2,
									Value:            2,
								},
							},
							Name:             "{param1}",
							FullPath:         "/api/{param1}",
							PathParamCounter: 1,
							Value:            nil,
						},
					},
					Name:     "api",
					FullPath: "/api",
				},
			},
			args: args{
				segments: []string{"api", "cats", "dogs", "test"},
				idx:      0,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.trie.getMatchNodes(tt.args.segments, tt.args.idx)
			sort.Slice(got, func(i, j int) bool {
				return got[i].FullPath < got[j].FullPath
			})
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].FullPath < tt.want[j].FullPath
			})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMatchNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMostAccurateNode(t *testing.T) {
	pt := New()
	type args struct {
		nodes       []*TrieNode
		path        string
		segmentsLen int
	}
	tests := []struct {
		name string
		args args
		want *TrieNode
	}{
		{
			name: "exact prefix match",
			args: args{
				nodes: []*TrieNode{
					pt.createPathTrieNode([]string{"", "api", "{param1}", "test"}, 3, true, 1),
					pt.createPathTrieNode([]string{"", "api", "{param1}", "{param2}"}, 3, true, 2),
				},
				path:        "/api/{param1}/test",
				segmentsLen: 4,
			},
			want: pt.createPathTrieNode([]string{"", "api", "{param1}", "test"}, 3, true, 1),
		},
		{
			name: "less path params match",
			args: args{
				nodes: []*TrieNode{
					pt.createPathTrieNode([]string{"", "api", "{param1}", "test", "{param2}"}, 4, true, 1),
					pt.createPathTrieNode([]string{"", "api", "{param1}", "{param2}", "{param3}"}, 4, true, 2),
				},
				path:        "/api/cats/test/dogs",
				segmentsLen: 5,
			},
			want: pt.createPathTrieNode([]string{"", "api", "{param1}", "test", "{param2}"}, 4, true, 1),
		},
		{
			name: "single match",
			args: args{
				nodes: []*TrieNode{
					pt.createPathTrieNode([]string{"", "api", "{param1}", "test"}, 3, true, 1),
				},
				path:        "/api/cats/test",
				segmentsLen: 4,
			},
			want: pt.createPathTrieNode([]string{"", "api", "{param1}", "test"}, 3, true, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMostAccurateNode(tt.args.nodes, tt.args.path, tt.args.segmentsLen); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMostAccurateNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTrieNode_isNameMatch(t *testing.T) {
	type fields struct {
		Name string
	}
	type args struct {
		segment string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "path param match",
			fields: fields{
				Name: "{param}",
			},
			args: args{
				segment: "match",
			},
			want: true,
		},
		{
			name: "segment name match",
			fields: fields{
				Name: "match",
			},
			args: args{
				segment: "match",
			},
			want: true,
		},
		{
			name: "segment name not match",
			fields: fields{
				Name: "match",
			},
			args: args{
				segment: "not-match",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &TrieNode{
				Name: tt.fields.Name,
			}
			if got := node.isNameMatch(tt.args.segment); got != tt.want {
				t.Errorf("isNameMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTrieNode_isFullPathMatch(t *testing.T) {
	type fields struct {
		Prefix string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "match",
			fields: fields{
				Prefix: "/api/{param1}/test",
			},
			args: args{
				path: "/api/{param1}/test",
			},
			want: true,
		},
		{
			name: "no match",
			fields: fields{
				Prefix: "/api/{param1}/test",
			},
			args: args{
				path: "/api/{param1}/{param2}",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &TrieNode{
				FullPath: tt.fields.Prefix,
			}
			if got := node.isFullPathMatch(tt.args.path); got != tt.want {
				t.Errorf("isFullPathMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_countPathParam(t *testing.T) {
	type args struct {
		segments []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "no path param",
			args: args{
				segments: []string{"", "api", "cat", "test"},
			},
			want: 0,
		},
		{
			name: "single path param",
			args: args{
				segments: []string{"", "api", "{param1}", "test"},
			},
			want: 1,
		},
		{
			name: "multiple path params",
			args: args{
				segments: []string{"", "api", "{param1}", "test", "{param2}"},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countPathParam(tt.args.segments); got != tt.want {
				t.Errorf("countPathParam() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTrie_createPathTrieNode(t *testing.T) {
	type args struct {
		segments      []string
		idx           int
		isLastSegment bool
		val           interface{}
	}
	tests := []struct {
		name string
		args args
		want *TrieNode
	}{
		{
			name: "last segment with 1 path param",
			args: args{
				segments:      []string{"", "api", "{param}"},
				idx:           2,
				isLastSegment: true,
				val:           1,
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "{param}",
				FullPath:         "/api/{param}",
				PathParamCounter: 1,
				Value:            1,
			},
		},
		{
			name: "not last segment with 3 path param",
			args: args{
				segments:      []string{"", "api", "{param1}", "{param2}", "{param3}", "test"},
				idx:           4,
				isLastSegment: false,
				val:           1,
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "{param3}",
				FullPath:         "/api/{param1}/{param2}/{param3}",
				PathParamCounter: 3,
				Value:            nil,
			},
		},
		{
			name: "not last segment with no path param",
			args: args{
				segments:      []string{"", "api", "{param1}", "{param2}", "{param3}", "test"},
				idx:           1,
				isLastSegment: false,
				val:           1,
			},
			want: &TrieNode{
				Children:         make(PathToTrieNode),
				Name:             "api",
				FullPath:         "/api",
				PathParamCounter: 0,
				Value:            nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := New()
			if got := pt.createPathTrieNode(tt.args.segments, tt.args.idx, tt.args.isLastSegment, tt.args.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createPathTrieNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathTrie_InsertMerge(t *testing.T) {
	swapMerge := func(existing, newV *interface{}) {
		*existing = *newV
	}
	shouldNotBeCalledMergeFunc := func(existing, newV *interface{}) {
		panic(fmt.Sprintf("merge should not be called. existing=%+v, newV=%+v", *existing, *newV))
	}
	type fields struct {
		Trie          PathToTrieNode
		PathSeparator string
	}
	type args struct {
		path  string
		val   interface{}
		merge ValueMergeFunc
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantIsNewPath bool
		expectedTrie  PathToTrieNode
	}{
		{
			name: "new path",
			fields: fields{
				Trie:          PathToTrieNode{},
				PathSeparator: "/",
			},
			args: args{
				path:  "/api",
				val:   1,
				merge: shouldNotBeCalledMergeFunc,
			},
			wantIsNewPath: true,
			expectedTrie: PathToTrieNode{
				"": &TrieNode{
					Children: map[string]*TrieNode{
						"api": {
							Children:         make(PathToTrieNode),
							Name:             "api",
							FullPath:         "/api",
							PathParamCounter: 0,
							Value:            1,
						},
					},
					Name:             "",
					FullPath:         "",
					PathParamCounter: 0,
					Value:            nil,
				},
			},
		},
		{
			name: "existing path",
			fields: fields{
				Trie: PathToTrieNode{
					"": &TrieNode{
						Children: map[string]*TrieNode{
							"api": {
								Children:         make(PathToTrieNode),
								Name:             "api",
								FullPath:         "/api",
								PathParamCounter: 0,
								Value:            1,
							},
						},
						Name:             "",
						FullPath:         "",
						PathParamCounter: 0,
						Value:            nil,
					},
				},
				PathSeparator: "/",
			},
			args: args{
				path:  "/api",
				val:   2,
				merge: swapMerge,
			},
			wantIsNewPath: false,
			expectedTrie: PathToTrieNode{
				"": &TrieNode{
					Children: map[string]*TrieNode{
						"api": {
							Children:         make(PathToTrieNode),
							Name:             "api",
							FullPath:         "/api",
							PathParamCounter: 0,
							Value:            2,
						},
					},
					Name:             "",
					FullPath:         "",
					PathParamCounter: 0,
					Value:            nil,
				},
			},
		},
		{
			name: "path with separator at the end - expected new path",
			fields: fields{
				Trie: PathToTrieNode{
					"": &TrieNode{
						Children: map[string]*TrieNode{
							"api": {
								Children:         make(PathToTrieNode),
								Name:             "api",
								FullPath:         "/api",
								PathParamCounter: 0,
								Value:            1,
							},
						},
						Name:             "",
						FullPath:         "",
						PathParamCounter: 0,
						Value:            nil,
					},
				},
				PathSeparator: "/",
			},
			args: args{
				path:  "/api/",
				val:   2,
				merge: shouldNotBeCalledMergeFunc,
			},
			wantIsNewPath: true,
			expectedTrie: PathToTrieNode{
				"": &TrieNode{
					Children: map[string]*TrieNode{
						"api": {
							Children: map[string]*TrieNode{
								"": {
									Children:         make(PathToTrieNode),
									Name:             "",
									FullPath:         "/api/",
									PathParamCounter: 0,
									Value:            2,
								},
							},
							Name:             "api",
							FullPath:         "/api",
							PathParamCounter: 0,
							Value:            1,
						},
					},
					Name:             "",
					FullPath:         "",
					PathParamCounter: 0,
					Value:            nil,
				},
			},
		},
		{
			name: "path param addition",
			fields: fields{
				Trie: PathToTrieNode{
					"": &TrieNode{
						Children: map[string]*TrieNode{
							"api": {
								Children:         make(PathToTrieNode),
								Name:             "api",
								FullPath:         "/api",
								PathParamCounter: 0,
								Value:            1,
							},
						},
						Name:             "",
						FullPath:         "",
						PathParamCounter: 0,
						Value:            nil,
					},
				},
				PathSeparator: "/",
			},
			args: args{
				path:  "/api/{param}",
				val:   2,
				merge: swapMerge,
			},
			wantIsNewPath: true,
			expectedTrie: PathToTrieNode{
				"": &TrieNode{
					Children: map[string]*TrieNode{
						"api": {
							Children: map[string]*TrieNode{
								"{param}": {
									Children:         make(PathToTrieNode),
									Name:             "{param}",
									FullPath:         "/api/{param}",
									PathParamCounter: 1,
									Value:            2,
								},
							},
							Name:             "api",
							FullPath:         "/api",
							PathParamCounter: 0,
							Value:            1,
						},
					},
					Name:             "",
					FullPath:         "",
					PathParamCounter: 0,
					Value:            nil,
				},
			},
		},
		{
			name: "new path for existing node",
			fields: fields{
				// /carts/{customerID}/items/{itemID}
				Trie: PathToTrieNode{
					"": &TrieNode{
						Children: map[string]*TrieNode{
							"carts": {
								Children: map[string]*TrieNode{
									"{customerID}": {
										Children: map[string]*TrieNode{
											"items": {
												Children: map[string]*TrieNode{
													"{itemID}": {
														Children:         make(PathToTrieNode),
														Name:             "{itemID}",
														FullPath:         "/carts/{customerID}/items/{itemID}",
														PathParamCounter: 1,
														Value:            "1",
													},
												},
												Name:             "items",
												FullPath:         "/carts/{customerID}/items",
												PathParamCounter: 1,
											},
										},
										Name:             "{customerID}",
										FullPath:         "/carts/{customerID}",
										PathParamCounter: 0,
									},
								},
								Name:             "carts",
								FullPath:         "/carts",
								PathParamCounter: 0,
							},
						},
						Name:             "",
						FullPath:         "",
						PathParamCounter: 0,
						Value:            nil,
					},
				},
				PathSeparator: "/",
			},
			args: args{
				path:  "/carts/{customerID}/items",
				val:   "2",
				merge: swapMerge,
			},
			wantIsNewPath: true,
			expectedTrie: PathToTrieNode{
				"": &TrieNode{
					Children: map[string]*TrieNode{
						"carts": {
							Children: map[string]*TrieNode{
								"{customerID}": {
									Children: map[string]*TrieNode{
										"items": {
											Children: map[string]*TrieNode{
												"{itemID}": {
													Children:         map[string]*TrieNode{},
													Name:             "{itemID}",
													FullPath:         "/carts/{customerID}/items/{itemID}",
													PathParamCounter: 1,
													Value:            "1",
												},
											},
											Name:             "items",
											FullPath:         "/carts/{customerID}/items",
											PathParamCounter: 1,
											Value:            "2",
										},
									},
									Name:             "{customerID}",
									FullPath:         "/carts/{customerID}",
									PathParamCounter: 0,
								},
							},
							Name:             "carts",
							FullPath:         "/carts",
							PathParamCounter: 0,
						},
					},
					Name:             "",
					FullPath:         "",
					PathParamCounter: 0,
					Value:            nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := &PathTrie{
				Trie:          tt.fields.Trie,
				PathSeparator: tt.fields.PathSeparator,
			}
			if gotIsNewPath := pt.InsertMerge(tt.args.path, tt.args.val, tt.args.merge); gotIsNewPath != tt.wantIsNewPath {
				t.Errorf("InsertMerge() = %v, want %v", gotIsNewPath, tt.wantIsNewPath)
			}
			if !reflect.DeepEqual(pt.Trie, tt.expectedTrie) {
				t.Errorf("InsertMerge() Trie = %+v, want %+v", marshal(pt.Trie), marshal(tt.expectedTrie))
			}
		})
	}
}

func marshal(obj interface{}) string {
	objB, _ := json.Marshal(obj)
	return string(objB)
}
