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

package spec

import (
	"fmt"
	"strings"

	oapi_spec "github.com/getkin/kin-openapi/openapi3"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/speculator/pkg/utils"
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

// CreateSuggestedReview group all paths that have suspect parameter (with a certain template),
// into one path which is parameterized, and then add this path params to the spec.
func (s *Spec) CreateSuggestedReview() *SuggestedSpecReview {
	s.lock.Lock()
	defer s.lock.Unlock()

	ret := &SuggestedSpecReview{
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

	for path := range s.LearningSpec.PathItems {
		parameterizedPath := createParameterizedPath(path)
		if _, ok := learningParametrizedPaths.Paths[parameterizedPath]; !ok {
			learningParametrizedPaths.Paths[parameterizedPath] = make(map[string]bool)
		}
		learningParametrizedPaths.Paths[parameterizedPath][path] = true
	}
	return &learningParametrizedPaths
}

func (s *Spec) ApplyApprovedReview(approvedReviews *ApprovedSpecReview, version OASVersion) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// first update the review into a copy of the state, in case the validation will fail
	clonedSpec, err := s.SpecInfoClone()
	if err != nil {
		return fmt.Errorf("failed to clone spec. %v", err)
	}

	for _, pathItemReview := range approvedReviews.PathItemsReview {
		mergedPathItem := &oapi_spec.PathItem{}
		for path := range pathItemReview.Paths {
			pathItem, ok := approvedReviews.PathToPathItem[path]
			if !ok {
				log.Errorf("path: %v was not found in learning spec", path)
				continue
			}
			mergedPathItem = MergePathItems(mergedPathItem, pathItem)

			// delete path from learning spec
			delete(clonedSpec.LearningSpec.PathItems, path)
		}

		addPathParamsToPathItem(mergedPathItem, pathItemReview.ParameterizedPath, pathItemReview.Paths)

		// add modified path and merged path item to ApprovedSpec
		clonedSpec.ApprovedSpec.PathItems[pathItemReview.ParameterizedPath] = mergedPathItem

		// add the modified path to the path tree
		isNewPath := clonedSpec.ApprovedPathTrie.Insert(pathItemReview.ParameterizedPath, pathItemReview.PathUUID)
		if !isNewPath {
			log.Warnf("Path was updated, a new path should be created in a normal case. path=%v, uuid=%v", pathItemReview.ParameterizedPath, pathItemReview.PathUUID)
		}

		// populate SecuritySchemes from the approved merged path item
		clonedSpec.ApprovedSpec.SecuritySchemes = updateSecuritySchemesFromPathItem(clonedSpec.ApprovedSpec.SecuritySchemes, mergedPathItem)
	}

	if _, err := clonedSpec.GenerateOASJson(version); err != nil {
		return fmt.Errorf("failed to generate Open API Spec. %w", err)
	}

	clonedSpec.ApprovedSpec.SpecVersion = version

	s.SpecInfo = clonedSpec.SpecInfo
	log.Debugf("Setting approved spec with version %q for %s:%s", s.ApprovedSpec.GetSpecVersion(), s.Host, s.Port)

	return nil
}

func updateSecuritySchemesFromPathItem(sd oapi_spec.SecuritySchemes, item *oapi_spec.PathItem) oapi_spec.SecuritySchemes {
	sd = updateSecuritySchemesFromOperation(sd, item.Get)
	sd = updateSecuritySchemesFromOperation(sd, item.Put)
	sd = updateSecuritySchemesFromOperation(sd, item.Post)
	sd = updateSecuritySchemesFromOperation(sd, item.Delete)
	sd = updateSecuritySchemesFromOperation(sd, item.Options)
	sd = updateSecuritySchemesFromOperation(sd, item.Head)
	sd = updateSecuritySchemesFromOperation(sd, item.Patch)

	return sd
}

func addPathParamsToPathItem(pathItem *oapi_spec.PathItem, suggestedPath string, paths map[string]bool) {
	// get all parameters names from path
	suggestedPathTrimed := strings.TrimPrefix(suggestedPath, "/")
	parts := strings.Split(suggestedPathTrimed, "/")

	for i, part := range parts {
		if !utils.IsPathParam(part) {
			continue
		}

		part = strings.TrimPrefix(part, utils.ParamPrefix)
		part = strings.TrimSuffix(part, utils.ParamSuffix)
		paramList := getOnlyIndexedPartFromPaths(paths, i)
		paramInfo := createPathParam(part, getParamSchema(paramList))
		pathItem.Parameters = append(pathItem.Parameters, &oapi_spec.ParameterRef{
			Value: paramInfo.Parameter,
		})
	}
}
