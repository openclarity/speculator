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
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/speculator/pkg/utils"

	oapi_spec "github.com/go-openapi/spec"
)

type SuggestedSpecReview struct {
	PathItemsReview []*SuggestedSpecReviewPathItem
	PathToPathItem  map[string]*oapi_spec.PathItem
}

type ApprovedSpecReview struct {
	PathItemsReview []*ApprovedSpecReviewPathItem
	PathToPathItem  map[string]*oapi_spec.PathItem
}

type ApprovedSpecReviewPathItem struct {
	ReviewPathItem
	PathUUID string
}

type SuggestedSpecReviewPathItem struct {
	ReviewPathItem
}

type ReviewPathItem struct {
	// ParameterizedPath represents the parameterized path grouping Paths
	ParameterizedPath string
	// Paths group of paths ParametrizedPath is representing
	Paths map[string]bool
}

// this function should group all paths that have suspect parameter (with a certain template), into one path which is parameterized, and then add this path params to the spec
func (s *Spec) CreateSuggestedReview() *SuggestedSpecReview {
	s.lock.Lock()
	defer s.lock.Unlock()

	var ret = &SuggestedSpecReview{
		PathToPathItem: s.LearningSpec.PathItems,
	}

	learningParametrizedPaths := s.createLearningParametrizedPaths()

	for parametrizedPath, paths := range learningParametrizedPaths.Paths {
		pathReview := &SuggestedSpecReviewPathItem{}
		pathReview.ParameterizedPath = parametrizedPath

		pathReview.Paths = paths

		ret.PathItemsReview = append(ret.PathItemsReview, pathReview)
	}
	return ret
}

func (s *Spec) createLearningParametrizedPaths() *LearningParametrizedPaths {
	var learningParametrizedPaths LearningParametrizedPaths

	learningParametrizedPaths.Paths = make(map[string]map[string]bool)

	for path, _ := range s.LearningSpec.PathItems {
		parameterizedPath := createParameterizedPath(path)
		if _, ok := learningParametrizedPaths.Paths[parameterizedPath]; !ok {
			learningParametrizedPaths.Paths[parameterizedPath] = make(map[string]bool)
		}
		learningParametrizedPaths.Paths[parameterizedPath][path] = true
	}
	return &learningParametrizedPaths
}

func (s *Spec) ApplyApprovedReview(approvedReviews *ApprovedSpecReview) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, pathItemReview := range approvedReviews.PathItemsReview {
		mergedPathItem := &oapi_spec.PathItem{}
		for path, _ := range pathItemReview.Paths {
			pathItem, ok := approvedReviews.PathToPathItem[path]
			if !ok {
				log.Errorf("path: %v was not found in learning spec", path)
				continue
			}
			mergedPathItem = MergePathItems(mergedPathItem, pathItem)

			// delete path from learning spec
			delete(s.LearningSpec.PathItems, path)
		}

		addPathParamsToPathItem(mergedPathItem, pathItemReview.ParameterizedPath, pathItemReview.Paths)

		// add modified path and merged path item to ApprovedSpec
		s.ApprovedSpec.PathItems[pathItemReview.ParameterizedPath] = mergedPathItem

		// add the modified path to the path tree
		isNewNode := s.PathTrie.Insert(pathItemReview.ParameterizedPath, pathItemReview.PathUUID)
		if !isNewNode {
			log.Warnf("Node path was updated, a new node should be created in a normal case. path=%v, uuid=%v", pathItemReview.ParameterizedPath, pathItemReview.PathUUID)
		}

		// populate SecurityDefinitions from the approved merged path item
		s.ApprovedSpec.SecurityDefinitions = updateSecurityDefinitionsFromPathItem(s.ApprovedSpec.SecurityDefinitions, mergedPathItem)
	}
}

func updateSecurityDefinitionsFromPathItem(sd oapi_spec.SecurityDefinitions, item *oapi_spec.PathItem) oapi_spec.SecurityDefinitions {
	sd = updateSecurityDefinitionsFromOperation(sd, item.Get)
	sd = updateSecurityDefinitionsFromOperation(sd, item.Put)
	sd = updateSecurityDefinitionsFromOperation(sd, item.Post)
	sd = updateSecurityDefinitionsFromOperation(sd, item.Delete)
	sd = updateSecurityDefinitionsFromOperation(sd, item.Options)
	sd = updateSecurityDefinitionsFromOperation(sd, item.Head)
	sd = updateSecurityDefinitionsFromOperation(sd, item.Patch)

	return sd
}

func addPathParamsToPathItem(pathItem *oapi_spec.PathItem, suggestedPath string, paths map[string]bool) {
	// get all parameters names from path
	suggestedPathTrimed := strings.TrimPrefix(suggestedPath, "/")
	parts := strings.Split(suggestedPathTrimed, "/")

	for i, part := range parts {
		if utils.IsPathParam(part) {
			part = strings.TrimPrefix(part, utils.ParamPrefix)
			part = strings.TrimSuffix(part, utils.ParamSuffix)
			paramList := getOnlyIndexedPartFromPaths(paths, i)
			tpe, format := getParamTypeAndFormat(paramList)
			paramInfo := createPathParam(part, tpe, format)
			pathItem.Parameters = append(pathItem.Parameters, *paramInfo.Parameter)
		}
	}
}
