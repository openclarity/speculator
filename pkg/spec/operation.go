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
	"encoding/json"
	"fmt"
	"mime"
	"net/url"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/xeipuuv/gojsonschema"

	"github.com/apiclarity/speculator/pkg/utils"
)

var (
	defaultSchema   = &spec.Schema{}
	defaultResponse = spec.NewResponse().
			WithDescription("Default Response").
			WithSchema(defaultSchema.AddType(schemaTypeObject, "").SetProperty("message", *spec.StringProperty()))
)

func getSchema(value interface{}) (schema *spec.Schema, err error) {
	switch value.(type) {
	case bool:
		schema = spec.BooleanProperty()
	case string:
		schema = getStringSchema(value)
	case json.Number:
		schema = getNumberSchema(value)
	case map[string]interface{}:
		schema, err = getObjectSchema(value)
		if err != nil {
			return nil, err
		}
	case []interface{}:
		schema, err = getArraySchema(value)
		if err != nil {
			return nil, err
		}
	case nil:
		// TODO: Not sure how to handle null. ex: {"size":3,"err":null}
		schema = spec.StringProperty()
	default:
		// TODO:
		// I've tested additionalProperties and it seems like properties - we will might have problems in the diff logic
		// spec.MapProperty()
		// spec.RefProperty()
		// spec.RefSchema()
		// spec.ComposedSchema() - discriminator?
		return nil, fmt.Errorf("unexpected value type. value=%v, type=%T", value, value)
	}

	return schema, nil
}

func getStringSchema(value interface{}) (schema *spec.Schema) {
	return spec.StrFmtProperty(getStringFormat(value))
}

func getNumberSchema(value interface{}) (schema *spec.Schema) {
	// https://swagger.io/docs/specification/data-models/data-types/#numbers

	// It is important to try first convert it to int
	if _, err := value.(json.Number).Int64(); err != nil {
		// if failed to convert to int it's a double
		// TODO: we will set a 'double' and not a 'float' - is that ok?
		schema = spec.Float64Property()
	} else {
		schema = spec.Int64Property()
	}
	// TODO: Format
	// spec.Int8Property()
	// spec.Int16Property()
	// spec.Int32Property()
	// spec.Float64Property()
	// spec.Float32Property()
	return schema /*.WithExample(value)*/
}

func getObjectSchema(value interface{}) (schema *spec.Schema, err error) {
	schema = &spec.Schema{}
	stringMapE, err := cast.ToStringMapE(value)
	if err != nil {
		return nil, fmt.Errorf("failed to cast to string map. value=%v: %w", value, err)
	}

	schema.AddType(schemaTypeObject, "")
	for key, val := range stringMapE {
		if s, err := getSchema(val); err != nil {
			return nil, fmt.Errorf("failed to get schema from string map. key=%v, value=%v: %w", key, val, err)
		} else {
			schema.SetProperty(escapeString(key), *s)
		}
	}

	return schema, nil
}

func escapeString(key string) string {
	// need to escape double quotes if exists
	if strings.Contains(key, "\"") {
		key = strings.ReplaceAll(key, "\"", "\\\"")
	}
	return key
}

func getArraySchema(value interface{}) (schema *spec.Schema, err error) {
	sliceE, err := cast.ToSliceE(value)
	if err != nil {
		return nil, fmt.Errorf("failed to cast to slice. value=%v: %w", value, err)
	}

	// in order to support mixed type array we ...
	schemaTypeToSchema := make(map[string]*spec.Schema)
	for i := range sliceE {
		item, err := getSchema(sliceE[i])
		if err != nil {
			return nil, fmt.Errorf("failed to get items schema from slice. value=%v: %w", sliceE[i], err)
		}
		t := []string(item.Type)[0]
		if _, ok := schemaTypeToSchema[t]; !ok {
			schemaTypeToSchema[t] = item
		}
	}
	if len(schemaTypeToSchema) == 0 {
		// array is empty but we can't create an empty array property (Schemas with 'type: array', require a sibling 'items:' field)
		// we will create string type items as a default value
		return spec.ArrayProperty(spec.StringProperty()), nil
	}
	if len(schemaTypeToSchema) == 1 {
		for _, s := range schemaTypeToSchema {
			schema = spec.ArrayProperty(s)
			break
		}
	} else {
		return nil, fmt.Errorf("oneOf is not supported by swagger 2.0, only by swagger 3.0. slice=%v", sliceE)
	}

	// TODO: Should we support an array of arbitrary types? (https://swagger.io/docs/specification/data-models/data-types/#array)
	// type: array
	// items: {}
	//	 # [ "hello", -2, true, [5.7], {"id": 5} ]

	// schema.CollectionOf(*items)
	return schema, nil
}

type HTTPInteractionData struct {
	ReqBody, RespBody       string
	ReqHeaders, RespHeaders map[string]string
	QueryParams             url.Values
	statusCode              int
}

func (h *HTTPInteractionData) getReqContentType() string {
	return h.ReqHeaders[contentTypeHeaderName]
}

func (h *HTTPInteractionData) getRespContentType() string {
	return h.RespHeaders[contentTypeHeaderName]
}

type OperationGeneratorConfig struct {
	ResponseHeadersToIgnore []string
	RequestHeadersToIgnore  []string
}

type OperationGenerator struct {
	ResponseHeadersToIgnore map[string]struct{}
	RequestHeadersToIgnore  map[string]struct{}
}

func NewOperationGenerator(config OperationGeneratorConfig) *OperationGenerator {
	return &OperationGenerator{
		ResponseHeadersToIgnore: createHeadersToIgnore(config.ResponseHeadersToIgnore),
		RequestHeadersToIgnore:  createHeadersToIgnore(config.RequestHeadersToIgnore),
	}
}

// Note: securityDefinitions might be updated.
func (o *OperationGenerator) GenerateSpecOperation(data *HTTPInteractionData, securityDefinitions spec.SecurityDefinitions) (*spec.Operation, error) {
	operation := spec.NewOperation("")

	if len(data.ReqBody) > 0 {
		reqContentType := data.getReqContentType()
		if reqContentType == "" {
			log.Infof("Missing Content-Type header, ignoring request body. (%v)", data.ReqBody)
		} else {
			operation.Consumes = append(operation.Consumes, reqContentType)
			mediaType, mediaTypeParams, err := mime.ParseMediaType(reqContentType)
			if err != nil {
				return nil, fmt.Errorf("failed to parse request media type. Content-Type=%v: %w", reqContentType, err)
			}
			switch true {
			case utils.IsApplicationJSONMediaType(mediaType):
				reqBodyJSON, err := gojsonschema.NewStringLoader(data.ReqBody).LoadJSON()
				if err != nil {
					return nil, fmt.Errorf("failed to load json from request body. body=%v: %w", data.ReqBody, err)
				}

				reqSchema, err := getSchema(reqBodyJSON)
				if err != nil {
					return nil, fmt.Errorf("failed to get schema from request body. body=%v: %w", data.ReqBody, err)
				}

				// all operation have to hold the same in body name parameter (inBodyParameterName)
				operation.AddParam(spec.BodyParam(inBodyParameterName, reqSchema))
			case mediaType == mediaTypeApplicationForm:
				operation, securityDefinitions = addApplicationFormParams(operation, securityDefinitions, data.ReqBody)
			case mediaType == mediaTypeMultipartFormData:
				// multipart/form-data (used to upload files or a combination of files and primitive data).
				// https://swagger.io/docs/specification/2-0/file-upload/
				operation, err = addMultipartFormDataParams(operation, data.ReqBody, mediaTypeParams)
				if err != nil {
					return nil, fmt.Errorf("failed to add multipart formData params from request body. body=%v: %v", data.ReqBody, err)
				}
			default:
				log.Infof("Treating %v as default request content type (no schema)", reqContentType)
			}
		}
	}

	for key, value := range data.ReqHeaders {
		lowerKey := strings.ToLower(key)
		if lowerKey == authorizationTypeHeaderName {
			operation, securityDefinitions = handleAuthReqHeader(operation, securityDefinitions, value)
		} else if APIKeyNames[lowerKey] {
			sdName := APIKeyAuthSecurityDefinitionKey
			operation = addSecurity(operation, sdName)
			scheme := spec.APIKeyAuth(key, apiKeyInHeader)
			securityDefinitions = updateSecurityDefinitions(securityDefinitions, sdName, scheme)
		} else {
			operation = o.addHeaderParam(operation, key, value)
		}
	}

	for key, values := range data.QueryParams {
		lowerKey := strings.ToLower(key)
		if lowerKey == AccessTokenParamKey {
			// Use scheme as security definition name
			sdName := OAuth2SecurityDefinitionKey
			var scheme *spec.SecurityScheme

			if hasSecurity(operation, sdName) {
				// RFC 6750 states multiple methods (form, uri query, header) cannot be used.
				log.Errorf("OAuth tokens supplied with multiple methods, ignoring URI query param: %v", key)
				continue
			}
			if len(values) > 1 {
				// RFC 6750 does not prohibit multiple tokens, but we do not know whether
				// they would be AND or OR so we just pick the latest.
				log.Warnf("Found %v tokens in query parameters, using only the last", len(values))
				values = values[len(values)-1:]
			}

			operation, scheme = generateAuthBearerScheme(operation, values[0], sdName)
			securityDefinitions = updateSecurityDefinitions(securityDefinitions, sdName, scheme)
		} else if APIKeyNames[lowerKey] {
			sdName := APIKeyAuthSecurityDefinitionKey
			operation = addSecurity(operation, sdName)
			scheme := spec.APIKeyAuth(key, apiKeyInQuery)
			securityDefinitions = updateSecurityDefinitions(securityDefinitions, sdName, scheme)
		} else {
			operation = addQueryParam(operation, key, values)
		}
	}

	response := spec.NewResponse()
	if len(data.RespBody) > 0 {
		respContentType := data.getRespContentType()
		if respContentType == "" {
			log.Infof("Missing Content-Type header, ignoring response body. (%v)", data.RespBody)
		} else {
			operation.Produces = append(operation.Produces, respContentType)
			mediaType, _, err := mime.ParseMediaType(respContentType)
			if err != nil {
				return nil, fmt.Errorf("failed to parse response media type. Content-Type=%v: %w", respContentType, err)
			}
			switch true {
			case utils.IsApplicationJSONMediaType(mediaType):
				respBodyJSON, err := gojsonschema.NewStringLoader(data.RespBody).LoadJSON()
				if err != nil {
					return nil, fmt.Errorf("failed to load json from response body. body=%v: %w", data.RespBody, err)
				}

				respSchema, err := getSchema(respBodyJSON)
				if err != nil {
					return nil, fmt.Errorf("failed to get schema from response body. body=%v: %w", respBodyJSON, err)
				}

				response.WithSchema(respSchema)
			// WithDescription("some response").
			// AddExample("application/json", respBody)
			default:
				log.Infof("Treating %v as default response content type (no schema)", respContentType)
			}
		}
	}

	for key, value := range data.RespHeaders {
		response = o.addResponseHeader(response, key, value)
	}

	operation.RespondsWith(data.statusCode, response).
		WithDefaultResponse(defaultResponse)

	return operation, nil
}

func CloneOperation(op *spec.Operation) (*spec.Operation, error) {
	var out spec.Operation

	opB, err := json.Marshal(op)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operation (%+v): %v", op, err)
	}

	if err := json.Unmarshal(opB, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %v", err)
	}

	return &out, nil
}

func generateAuthBearerScheme(operation *spec.Operation, bearerToken string, sdName string) (*spec.Operation, *spec.SecurityScheme) {
	// Parse the claims without validating (since we don't want to bother downloading a key)
	parser := jwt.Parser{}
	// we can't know the flow type (implicit, password, application or accessCode) so we choose accessCode for now
	scheme := spec.OAuth2AccessToken(authorizationURL, tknURL)

	if len(bearerToken) == 0 {
		log.Warnf("authZ token provided with no value, assuming OAuth required anyway")
		return addSecurity(operation, sdName), scheme
	}

	token, _, err := parser.ParseUnverified(bearerToken, jwt.MapClaims{})
	if err != nil {
		// Note: it is not a JWT so we just mark generic OAuth
		log.Warnf("authZ token is not a JWT, assuming OAuth required anyway")
		return addSecurity(operation, sdName), scheme
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// Claims are unintelligible so we just mark generic OAuth
		log.Warnf("authZ token had unintelligble claims, assuming OAuth required anyway")
		return addSecurity(operation, sdName), scheme
	}
	scopes := []string{}
	if scope, ok := claims["scope"]; ok {
		scopes = strings.Split(scope.(string), " ")
		log.Debugf("found OAuth token scopes: %v", scopes)
	} else {
		log.Warnf("no scopes defined in this token")
	}
	updateSecuritySchemeScopes(scheme, scopes, []string{})
	return addSecurity(operation, sdName, scopes...), scheme
}

func handleAuthReqHeader(operation *spec.Operation, sd spec.SecurityDefinitions, value string) (*spec.Operation, spec.SecurityDefinitions) {
	if strings.HasPrefix(value, BasicAuthPrefix) {
		// Use scheme as security definition name
		sdName := BasicAuthSecurityDefinitionKey
		operation = addSecurity(operation, sdName)
		sd = updateSecurityDefinitions(sd, sdName, spec.BasicAuth())
	} else if strings.HasPrefix(value, BearerAuthPrefix) {
		// Use scheme as security definition name. For OAuth, we should consider checking
		// supported scopes to allow multiple defs.
		sdName := OAuth2SecurityDefinitionKey

		if hasSecurity(operation, sdName) {
			// RFC 6750 states multiple methods (form, uri query, header) cannot be used.
			log.Error("OAuth tokens supplied with multiple methods, ignoring header")
			return operation, sd
		}

		var scheme *spec.SecurityScheme
		operation, scheme = generateAuthBearerScheme(operation, strings.TrimPrefix(value, BearerAuthPrefix), sdName)
		sd = updateSecurityDefinitions(sd, sdName, scheme)
	} else {
		log.Warnf("ignoring unknown authorization header value (%v)", value)
	}
	return operation, sd
}

func addSecurity(op *spec.Operation, name string, scopes ...string) *spec.Operation {
	// https://swagger.io/docs/specification/2-0/authentication/
	// We will treat multiple authentication types as an OR
	// (Security schemes combined via OR are alternatives – any one can be used in the given context)
	if len(scopes) > 0 {
		return op.SecuredWith(name, scopes...)
	} else {
		// We must use an empty array as the scopes, otherwise it will create invalid swagger
		return op.SecuredWith(name, []string{}...)
	}
}

func hasSecurity(op *spec.Operation, name string) bool {
	for _, securityScheme := range op.Security {
		if _, ok := securityScheme[name]; ok {
			return true
		}
	}
	return false
}
