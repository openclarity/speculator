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
	"net/http"

	oapi_spec "github.com/go-openapi/spec"
)

func MergePathItems(dst, src *oapi_spec.PathItem) *oapi_spec.PathItem {
	dst.Get, _ = mergeOperation(dst.Get, src.Get)
	dst.Put, _ = mergeOperation(dst.Put, src.Put)
	dst.Post, _ = mergeOperation(dst.Post, src.Post)
	dst.Delete, _ = mergeOperation(dst.Delete, src.Delete)
	dst.Options, _ = mergeOperation(dst.Options, src.Options)
	dst.Head, _ = mergeOperation(dst.Head, src.Head)
	dst.Patch, _ = mergeOperation(dst.Patch, src.Patch)

	// TODO what about merging parameters?

	return dst
}

func CopyPathItemWithNewOperation(item *oapi_spec.PathItem, method string, operation *oapi_spec.Operation) *oapi_spec.PathItem {
	// TODO - do we want to do : ret = *item?
	ret := oapi_spec.PathItem{}
	ret.Get = item.Get
	ret.Put = item.Put
	ret.Patch = item.Patch
	ret.Post = item.Post
	ret.Head = item.Head
	ret.Delete = item.Delete
	ret.Options = item.Options
	ret.Parameters = item.Parameters

	AddOperationToPathItem(&ret, method, operation)
	return &ret
}

func GetOperationFromPathItem(item *oapi_spec.PathItem, method string) *oapi_spec.Operation {
	switch method {
	case http.MethodGet:
		return item.Get
	case http.MethodDelete:
		return item.Delete
	case http.MethodOptions:
		return item.Options
	case http.MethodPatch:
		return item.Patch
	case http.MethodHead:
		return item.Head
	case http.MethodPost:
		return item.Post
	case http.MethodPut:
		return item.Put
	}
	return nil
}

func AddOperationToPathItem(item *oapi_spec.PathItem, method string, operation *oapi_spec.Operation) {
	switch method {
	case http.MethodGet:
		item.Get = operation
	case http.MethodDelete:
		item.Delete = operation
	case http.MethodOptions:
		item.Options = operation
	case http.MethodPatch:
		item.Patch = operation
	case http.MethodHead:
		item.Head = operation
	case http.MethodPost:
		item.Post = operation
	case http.MethodPut:
		item.Put = operation
	}
	return
}
