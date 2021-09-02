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

package spec

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"testing"

	"gotest.tools/assert"

	oapi_spec "github.com/go-openapi/spec"
	uuid "github.com/satori/go.uuid"

	"github.com/apiclarity/speculator/pkg/path_trie"
)

func TestSpec_ApplyApprovedReview(t *testing.T) {
	type fields struct {
		ID           uuid.UUID
		ApprovedSpec *ApprovedSpec
		LearningSpec *LearningSpec
		Mutex        sync.Mutex
	}
	type args struct {
		approvedReviews *ApprovedSpecReview
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantSpec *Spec
	}{
		{
			name: "1 reviewed path item. modified path param. same path item. 2 Paths",
			fields: fields{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
					},
				},
			},
			args: args{
				approvedReviews: &ApprovedSpecReview{
					PathToPathItem: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
					},
					PathItemsReview: []*ApprovedSpecReviewPathItem{
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/{param1}",
								Paths: map[string]bool{
									"/api/1": true,
									"/api/2": true,
								},
							},
							PathUUID: "1",
						},
					},
				},
			},
			wantSpec: &Spec{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/{param1}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, DataCombined).Op).WithPathParams("param1", schemaTypeInteger, "").PathItem,
					},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				PathTrie: createPathTrie(map[string]string{
					"/api/{param1}": "1",
				}),
			},
		},
		{
			name: "user took out one path out of the parameterized path, and also one more path has learned between review and approve (should ignore it and not delete)",
			fields: fields{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"api/3/foo": &NewTestPathItem().PathItem,
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
					},
				},
			},
			args: args{
				approvedReviews: &ApprovedSpecReview{
					PathToPathItem: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodPost, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
					},
					PathItemsReview: []*ApprovedSpecReviewPathItem{
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/{param1}",
								Paths: map[string]bool{
									"/api/2": true,
								},
							},
							PathUUID: "1",
						},
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/1",
								Paths: map[string]bool{
									"/api/1": true,
								},
							},
							PathUUID: "2",
						},
					},
				},
			},
			wantSpec: &Spec{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/{param1}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).WithPathParams("param1", schemaTypeInteger, "").PathItem,
						"/api/1":        &NewTestPathItem().WithOperation(http.MethodPost, NewOperation(t, Data).Op).PathItem,
					},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"api/3/foo": &NewTestPathItem().PathItem,
					},
				},
				PathTrie: createPathTrie(map[string]string{
					"/api/{param1}": "1",
					"/api/1":        "2",
				}),
			},
		},
		{
			name: "multiple methods",
			fields: fields{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/anything": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).
							WithOperation(http.MethodPost, NewOperation(t, Data).Op).PathItem,
						"/headers": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user-agent": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
			},
			args: args{
				approvedReviews: &ApprovedSpecReview{
					PathToPathItem: map[string]*oapi_spec.PathItem{
						"/anything": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).
							WithOperation(http.MethodPost, NewOperation(t, Data).Op).PathItem,
						"/headers": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user-agent": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
					PathItemsReview: []*ApprovedSpecReviewPathItem{
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/{test}",
								Paths: map[string]bool{
									"/anything":   true,
									"/headers":    true,
									"/user-agent": true,
								},
							},
							PathUUID: "1",
						},
					},
				},
			},
			wantSpec: &Spec{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/{test}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).WithOperation(http.MethodPost, NewOperation(t, Data).Op).WithPathParams("test", schemaTypeString, "").PathItem,
					},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				PathTrie: createPathTrie(map[string]string{
					"/api/{test}": "1",
				}),
			},
		},
		{
			name: "new parameterized path, unmerge of path item",
			fields: fields{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/foo": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user/1/bar/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
			},
			args: args{
				approvedReviews: &ApprovedSpecReview{
					PathToPathItem: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/foo": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user/1/bar/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
					PathItemsReview: []*ApprovedSpecReviewPathItem{
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/{param1}",
								Paths: map[string]bool{
									"/api/1": true,
									"/api/2": true,
								},
							},
							PathUUID: "1",
						},
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/foo",
								Paths: map[string]bool{
									"/api/foo": true,
								},
							},
							PathUUID: "2",
						},
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/user/{param1}/bar/{param2}",
								Paths: map[string]bool{
									"/user/1/bar/2": true,
								},
							},
							PathUUID: "3",
						},
					},
				},
			},
			wantSpec: &Spec{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/{param1}":               &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).WithPathParams("param1", schemaTypeInteger, "").PathItem,
						"/api/foo":                    &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user/{param1}/bar/{param2}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).WithPathParams("param1", schemaTypeInteger, "").WithPathParams("param2", schemaTypeInteger, "").PathItem,
					},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				PathTrie: createPathTrie(map[string]string{
					"/api/{param1}":               "1",
					"/api/foo":                    "2",
					"/user/{param1}/bar/{param2}": "3",
				}),
			},
		},
		{
			name: "new parameterized path, unmerge of path item with security",
			fields: fields{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems:           map[string]*oapi_spec.PathItem{},
					SecurityDefinitions: oapi_spec.SecurityDefinitions{},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op.
								SecuredWith(OAuth2SecurityDefinitionKey, []string{}...)).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/foo": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user/1/bar/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
			},
			args: args{
				approvedReviews: &ApprovedSpecReview{
					PathToPathItem: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op.
								SecuredWith(BasicAuthSecurityDefinitionKey, []string{}...)).PathItem,
						"/api/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/foo": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user/1/bar/2": &NewTestPathItem().
							WithOperation(http.MethodGet, NewOperation(t, Data).Op.
								SecuredWith(OAuth2SecurityDefinitionKey, []string{}...)).PathItem,
					},
					PathItemsReview: []*ApprovedSpecReviewPathItem{
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/{param1}",
								Paths: map[string]bool{
									"/api/1": true,
									"/api/2": true,
								},
							},
							PathUUID: "1",
						},
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/api/foo",
								Paths: map[string]bool{
									"/api/foo": true,
								},
							},
							PathUUID: "2",
						},
						{
							ReviewPathItem: ReviewPathItem{
								ParameterizedPath: "/user/{param1}/bar/{param2}",
								Paths: map[string]bool{
									"/user/1/bar/2": true,
								},
							},
							PathUUID: "3",
						},
					},
				},
			},
			wantSpec: &Spec{
				ID: uuid.UUID{},
				ApprovedSpec: &ApprovedSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/{param1}":               &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op.SecuredWith(BasicAuthSecurityDefinitionKey, []string{}...)).WithPathParams("param1", schemaTypeInteger, "").PathItem,
						"/api/foo":                    &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/user/{param1}/bar/{param2}": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op.SecuredWith(OAuth2SecurityDefinitionKey, []string{}...)).WithPathParams("param1", schemaTypeInteger, "").WithPathParams("param2", schemaTypeInteger, "").PathItem,
					},
					SecurityDefinitions: map[string]*oapi_spec.SecurityScheme{
						BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
						OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
					},
				},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{},
				},
				PathTrie: createPathTrie(map[string]string{
					"/api/{param1}":               "1",
					"/api/foo":                    "2",
					"/user/{param1}/bar/{param2}": "3",
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				ID:           tt.fields.ID,
				ApprovedSpec: tt.fields.ApprovedSpec,
				LearningSpec: tt.fields.LearningSpec,
				PathTrie:     path_trie.New(),
			}
			s.ApplyApprovedReview(tt.args.approvedReviews)

			specB, err := json.Marshal(s)
			assert.NilError(t, err)
			wantB, err := json.Marshal(tt.wantSpec)
			assert.NilError(t, err)

			if bytes.Compare(specB, wantB) != 0 {
				t.Errorf("ApplyApprovedReview() = %v, want %v", string(specB), string(wantB))
			}
		})
	}
}

func TestSpec_CreateSuggestedReview(t *testing.T) {
	type fields struct {
		ID                        uuid.UUID
		ApprovedSpec              *ApprovedSpec
		LearningSpec              *LearningSpec
		LearningParametrizedPaths *LearningParametrizedPaths
		Mutex                     sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   *SuggestedSpecReview
	}{
		{
			name: "2 paths - map to one parameterized path",
			fields: fields{
				ID: uuid.UUID{},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
					},
				},
				LearningParametrizedPaths: &LearningParametrizedPaths{
					Paths: map[string]map[string]bool{
						"/api/{param1}": {"/api/1": true, "/api/2": true},
					},
				},
			},
			want: &SuggestedSpecReview{
				PathToPathItem: map[string]*oapi_spec.PathItem{
					"/api/1": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					"/api/2": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
				},
				PathItemsReview: []*SuggestedSpecReviewPathItem{
					{
						ReviewPathItem: ReviewPathItem{
							ParameterizedPath: "/api/{param1}",
							Paths: map[string]bool{
								"/api/1": true,
								"/api/2": true,
							},
						},
					},
				},
			},
		},
		{
			name: "4 paths - 2 under one parameterized path with one param, one is not parameterized, one is parameterized path with 2 params",
			fields: fields{
				ID: uuid.UUID{},
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/1":           &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/2":           &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
						"/api/foo":         &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
						"/api/foo/1/bar/2": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					},
				},
				LearningParametrizedPaths: &LearningParametrizedPaths{
					Paths: map[string]map[string]bool{
						"/api/{param1}":                  {"/api/1": true, "/api/2": true},
						"/api/foo/{param1}/bar/{param2}": {"/api/foo/1/bar/2": true},
						"/api/foo":                       {"/api/foo": true},
					},
				},
			},
			want: &SuggestedSpecReview{
				PathToPathItem: map[string]*oapi_spec.PathItem{
					"/api/1":           &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					"/api/2":           &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data2).Op).PathItem,
					"/api/foo":         &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
					"/api/foo/1/bar/2": &NewTestPathItem().WithOperation(http.MethodGet, NewOperation(t, Data).Op).PathItem,
				},
				PathItemsReview: []*SuggestedSpecReviewPathItem{
					{
						ReviewPathItem: ReviewPathItem{
							ParameterizedPath: "/api/{param1}",
							Paths: map[string]bool{
								"/api/1": true,
								"/api/2": true,
							},
						},
					},
					{
						ReviewPathItem: ReviewPathItem{
							ParameterizedPath: "/api/foo/{param1}/bar/{param2}",
							Paths: map[string]bool{
								"/api/foo/1/bar/2": true,
							},
						},
					},
					{
						ReviewPathItem: ReviewPathItem{
							ParameterizedPath: "/api/foo",
							Paths: map[string]bool{
								"/api/foo": true,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				ID:           tt.fields.ID,
				ApprovedSpec: tt.fields.ApprovedSpec,
				LearningSpec: tt.fields.LearningSpec,
			}
			got := s.CreateSuggestedReview()
			sort.Slice(got.PathItemsReview, func(i, j int) bool {
				return got.PathItemsReview[i].ParameterizedPath > got.PathItemsReview[j].ParameterizedPath
			})
			sort.Slice(tt.want.PathItemsReview, func(i, j int) bool {
				return tt.want.PathItemsReview[i].ParameterizedPath > tt.want.PathItemsReview[j].ParameterizedPath
			})
			gotB := marshal(got)
			wantB := marshal(tt.want)
			if gotB != wantB {
				t.Errorf("CreateSuggestedReview() got = %v, want %v", gotB, wantB)
			}
		})
	}
}

func TestSpec_createLearningParametrizedPaths(t *testing.T) {
	type fields struct {
		Host         string
		ID           uuid.UUID
		ApprovedSpec *ApprovedSpec
		LearningSpec *LearningSpec
		lock         sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   *LearningParametrizedPaths
	}{
		{
			name: "",
			fields: fields{
				LearningSpec: &LearningSpec{
					PathItems: map[string]*oapi_spec.PathItem{
						"/api/1": &NewTestPathItem().PathItem,
					},
				},
			},
			want: &LearningParametrizedPaths{
				Paths: map[string]map[string]bool{
					"/api/{param1}": {"/api/1": true},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Spec{
				Host:         tt.fields.Host,
				ID:           tt.fields.ID,
				ApprovedSpec: tt.fields.ApprovedSpec,
				LearningSpec: tt.fields.LearningSpec,
				lock:         tt.fields.lock,
			}
			if got := s.createLearningParametrizedPaths(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createLearningParametrizedPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addPathParamsToPathItem(t *testing.T) {
	type args struct {
		pathItem      *oapi_spec.PathItem
		suggestedPath string
		paths         map[string]bool
	}
	tests := []struct {
		name         string
		args         args
		wantPathItem *oapi_spec.PathItem
	}{
		{
			name: "1 param",
			args: args{
				pathItem:      &NewTestPathItem().PathItem,
				suggestedPath: "/api/{param1}/foo",
				paths: map[string]bool{
					"api/1/foo": true,
					"api/2/foo": true,
				},
			},
			wantPathItem: &NewTestPathItem().WithPathParams("param1", schemaTypeInteger, "").PathItem,
		},
		{
			name: "2 params",
			args: args{
				pathItem:      &NewTestPathItem().PathItem,
				suggestedPath: "/api/{param1}/foo/{param2}",
				paths: map[string]bool{
					"api/1/foo/2":   true,
					"api/2/foo/345": true,
				},
			},
			wantPathItem: &NewTestPathItem().WithPathParams("param1", schemaTypeInteger, "").WithPathParams("param2", schemaTypeInteger, "").PathItem,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addPathParamsToPathItem(tt.args.pathItem, tt.args.suggestedPath, tt.args.paths)
			assert.Assert(t, reflect.DeepEqual(tt.args.pathItem, tt.wantPathItem))
		})
	}
}

func Test_updateSecurityDefinitionsFromPathItem(t *testing.T) {
	type args struct {
		sd   oapi_spec.SecurityDefinitions
		item *oapi_spec.PathItem
	}
	tests := []struct {
		name string
		args args
		want oapi_spec.SecurityDefinitions
	}{
		{
			name: "Get operation",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Get: createOperationWithSecurity([]map[string][]string{
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
		{
			name: "Put operation",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Put: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"admin"},
							},
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
		{
			name: "Post operation",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Post: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"admin"},
							},
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
		{
			name: "Delete operation",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Delete: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"admin"},
							},
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
		{
			name: "Options operation",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Options: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"admin"},
							},
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
		{
			name: "Head operation",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Head: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"admin"},
							},
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
		{
			name: "Patch operation",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Patch: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"admin"},
							},
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
		{
			name: "Multiple operations",
			args: args{
				sd: oapi_spec.SecurityDefinitions{},
				item: &oapi_spec.PathItem{
					PathItemProps: oapi_spec.PathItemProps{
						Get: createOperationWithSecurity([]map[string][]string{
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
						Put: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"read"},
							},
						}),
						Post: createOperationWithSecurity([]map[string][]string{
							{
								"unsupported": {"read"},
							},
						}),
						Delete: createOperationWithSecurity([]map[string][]string{
							{
								OAuth2SecurityDefinitionKey: {"admin"},
							},
							{
								BasicAuthSecurityDefinitionKey: {},
							},
						}),
						Options: createOperationWithSecurity(nil),
					},
				},
			},
			want: oapi_spec.SecurityDefinitions{
				OAuth2SecurityDefinitionKey:    oapi_spec.OAuth2AccessToken(authorizationURL, tokenURL),
				BasicAuthSecurityDefinitionKey: oapi_spec.BasicAuth(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateSecurityDefinitionsFromPathItem(tt.args.sd, tt.args.item); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateSecurityDefinitionsFromPathItem() = %v, want %v", got, tt.want)
			}
		})
	}
}
