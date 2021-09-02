/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package path_trie

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
		want *PathTrieNode
	}{
		{
			name: "most accurate match - will match both `/api/{param1}/items` and `/api/{param1}/{param2}`",
			args: args{
				path: "/api/1/items",
			},
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
			want: &PathTrieNode{
				Children: map[string]*PathTrieNode{
					"cat": {
						Children:         make(PathTrieMap, 0),
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
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
		trie PathTrieMap
		args args
		want []*PathTrieNode
	}{
		{
			name: "return 2 matches nodes",
			trie: map[string]*PathTrieNode{
				"api": {
					Children: map[string]*PathTrieNode{
						"{param1}": {
							Children: map[string]*PathTrieNode{
								"test": {
									Children:         make(PathTrieMap, 0),
									Name:             "test",
									FullPath:         "/api/{param1}/test",
									PathParamCounter: 1,
									Value:            1,
								},
								"{param2}": {
									Children:         make(PathTrieMap, 0),
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
			want: []*PathTrieNode{
				{
					Children:         make(PathTrieMap, 0),
					Name:             "test",
					FullPath:         "/api/{param1}/test",
					PathParamCounter: 1,
					Value:            1,
				},
				{
					Children:         make(PathTrieMap, 0),
					Name:             "{param2}",
					FullPath:         "/api/{param1}/{param2}",
					PathParamCounter: 2,
					Value:            2,
				},
			},
		},
		{
			name: "last path segment has nil value - return only 1 matches nodes (/api/{param1}/{param2})",
			trie: map[string]*PathTrieNode{
				"api": {
					Children: map[string]*PathTrieNode{
						"{param1}": {
							Children: map[string]*PathTrieNode{
								"test": {
									Children: map[string]*PathTrieNode{
										"cats": {
											Children:         make(PathTrieMap, 0),
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
									Children:         make(PathTrieMap, 0),
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
			want: []*PathTrieNode{
				{
					Children:         make(PathTrieMap, 0),
					Name:             "{param2}",
					FullPath:         "/api/{param1}/{param2}",
					PathParamCounter: 2,
					Value:            2,
				},
			},
		},
		{
			name: "0 nodes match",
			trie: map[string]*PathTrieNode{
				"api": {
					Children: map[string]*PathTrieNode{
						"{param1}": {
							Children: map[string]*PathTrieNode{
								"test": {
									Children:         make(PathTrieMap, 0),
									Name:             "test",
									FullPath:         "/api/{param1}/test",
									PathParamCounter: 1,
									Value:            1,
								},
								"{param2}": {
									Children:         make(PathTrieMap, 0),
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
		nodes       []*PathTrieNode
		path        string
		segmentsLen int
	}
	tests := []struct {
		name string
		args args
		want *PathTrieNode
	}{
		{
			name: "exact prefix match",
			args: args{
				nodes: []*PathTrieNode{
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
				nodes: []*PathTrieNode{
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
				nodes: []*PathTrieNode{
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
			node := &PathTrieNode{
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
			node := &PathTrieNode{
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
		want *PathTrieNode
	}{
		{
			name: "last segment with 1 path param",
			args: args{
				segments:      []string{"", "api", "{param}"},
				idx:           2,
				isLastSegment: true,
				val:           1,
			},
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
			want: &PathTrieNode{
				Children:         make(PathTrieMap, 0),
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
	swapMerge := func(existing, new *interface{}) {
		*existing = *new
	}
	shouldNotBeCalledMergeFunc := func(existing, new *interface{}) {
		panic(fmt.Sprintf("merge should not be called. existing=%+v, new=%+v", *existing, *new))
	}
	type fields struct {
		Trie          PathTrieMap
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
		wantIsNewNode bool
		expectedTrie  PathTrieMap
	}{
		{
			name: "new node",
			fields: fields{
				Trie:          PathTrieMap{},
				PathSeparator: "/",
			},
			args: args{
				path:  "/api",
				val:   1,
				merge: shouldNotBeCalledMergeFunc,
			},
			wantIsNewNode: true,
			expectedTrie: PathTrieMap{
				"": &PathTrieNode{
					Children: map[string]*PathTrieNode{
						"api": {
							Children:         make(PathTrieMap, 0),
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
			name: "existing node",
			fields: fields{
				Trie: PathTrieMap{
					"": &PathTrieNode{
						Children: map[string]*PathTrieNode{
							"api": {
								Children:         make(PathTrieMap, 0),
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
			wantIsNewNode: false,
			expectedTrie: PathTrieMap{
				"": &PathTrieNode{
					Children: map[string]*PathTrieNode{
						"api": {
							Children:         make(PathTrieMap, 0),
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
			name: "path with separator at the end - expected new node",
			fields: fields{
				Trie: PathTrieMap{
					"": &PathTrieNode{
						Children: map[string]*PathTrieNode{
							"api": {
								Children:         make(PathTrieMap, 0),
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
			wantIsNewNode: true,
			expectedTrie: PathTrieMap{
				"": &PathTrieNode{
					Children: map[string]*PathTrieNode{
						"api": {
							Children: map[string]*PathTrieNode{
								"": {
									Children:         make(PathTrieMap, 0),
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
				Trie: PathTrieMap{
					"": &PathTrieNode{
						Children: map[string]*PathTrieNode{
							"api": {
								Children:         make(PathTrieMap, 0),
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
			wantIsNewNode: true,
			expectedTrie: PathTrieMap{
				"": &PathTrieNode{
					Children: map[string]*PathTrieNode{
						"api": {
							Children: map[string]*PathTrieNode{
								"{param}": {
									Children:         make(PathTrieMap, 0),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := &PathTrie{
				Trie:          tt.fields.Trie,
				PathSeparator: tt.fields.PathSeparator,
			}
			if gotIsNewNode := pt.InsertMerge(tt.args.path, tt.args.val, tt.args.merge); gotIsNewNode != tt.wantIsNewNode {
				t.Errorf("InsertMerge() = %v, want %v", gotIsNewNode, tt.wantIsNewNode)
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
