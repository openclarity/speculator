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

	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/xeipuuv/gojsonschema"

	"github.com/openclarity/speculator/pkg/utils"
)

//var (
//	defaultSchema   = &spec.Schema{}
//	defaultResponse = spec.NewResponse().
//			WithDescription("Default Response").
//			WithSchema(defaultSchema.AddType(schemaTypeObject, "").SetProperty("message", *spec.StringProperty()))
//)

func getSchema(value interface{}) (schema *spec.Schema, err error) {
	switch value.(type) {
	case bool:
		schema = spec.NewBoolSchema()
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
		schema = spec.NewStringSchema()
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
	return spec.NewStringSchema().WithFormat(getStringFormat(value))
}

func getNumberSchema(value interface{}) (schema *spec.Schema) {
	// https://swagger.io/docs/specification/data-models/data-types/#numbers

	// It is important to try first convert it to int
	if _, err := value.(json.Number).Int64(); err != nil {
		// if failed to convert to int it's a double
		// TODO: we will set a 'double' and not a 'float' - is that ok?
		schema = spec.NewFloat64Schema()
	} else {
		schema = spec.NewInt64Schema()
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
	schema = spec.NewObjectSchema()
	stringMapE, err := cast.ToStringMapE(value)
	if err != nil {
		return nil, fmt.Errorf("failed to cast to string map. value=%v: %w", value, err)
	}

	for key, val := range stringMapE {
		if s, err := getSchema(val); err != nil {
			return nil, fmt.Errorf("failed to get schema from string map. key=%v, value=%v: %w", key, val, err)
		} else {
			schema.WithProperty(escapeString(key), s)
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

	// in order to support mixed type array we will map all schemas by schema type
	schemaTypeToSchema := make(map[string]*spec.Schema)
	for i := range sliceE {
		item, err := getSchema(sliceE[i])
		if err != nil {
			return nil, fmt.Errorf("failed to get items schema from slice. value=%v: %w", sliceE[i], err)
		}
		if _, ok := schemaTypeToSchema[item.Type]; !ok {
			schemaTypeToSchema[item.Type] = item
		}
	}

	switch len(schemaTypeToSchema) {
	case 0:
		// array is empty, but we can't create an empty array property (Schemas with 'type: array', require a sibling 'items:' field)
		// we will create string type items as a default value
		schema = spec.NewArraySchema().WithItems(spec.NewStringSchema())
	case 1:
		for _, s := range schemaTypeToSchema {
			schema = spec.NewArraySchema().WithItems(s)
			break
		}
	default:
		// oneOf
		// https://swagger.io/docs/specification/data-models/oneof-anyof-allof-not/
		var schemas []*spec.Schema
		for _, s := range schemaTypeToSchema {
			schemas = append(schemas, s)
		}
		schema = spec.NewOneOfSchema(schemas...)
	}

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

// Note: SecuritySchemes might be updated.
func (o *OperationGenerator) GenerateSpecOperation(data *HTTPInteractionData, securitySchemes spec.SecuritySchemes) (*spec.Operation, error) {
	operation := spec.NewOperation()

	if len(data.ReqBody) > 0 {
		reqContentType := data.getReqContentType()
		if reqContentType == "" {
			log.Infof("Missing Content-Type header, ignoring request body. (%v)", data.ReqBody)
		} else {
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

				operation.RequestBody.Value.WithJSONSchema(reqSchema)
			case mediaType == mediaTypeApplicationForm:
				operation, securitySchemes, err = handleApplicationFormURLEncodedBody(operation, securitySchemes, data.ReqBody)
				if err != nil {
					return nil, fmt.Errorf("failed to handle %s body: %v", mediaTypeApplicationForm, err)
				}
			case mediaType == mediaTypeMultipartFormData:
				// Multipart requests combine one or more sets of data into a single body, separated by boundaries.
				// You typically use these requests for file uploads and for transferring data of several types
				// in a single request (for example, a file along with a JSON object).
				// https://swagger.io/docs/specification/describing-request-body/multipart-requests/
				schema, err := getMultipartFormDataSchema(data.ReqBody, mediaTypeParams)
				if err != nil {
					return nil, fmt.Errorf("failed to get multipart form-data schema from request body. body=%v: %v", data.ReqBody, err)
				}
				operation.RequestBody.Value.WithFormDataSchema(schema)
			default:
				log.Infof("Treating %v as default request content type (no schema)", reqContentType)
			}
		}
	}

	for key, value := range data.ReqHeaders {
		lowerKey := strings.ToLower(key)
		if lowerKey == authorizationTypeHeaderName {
			// https://datatracker.ietf.org/doc/html/rfc6750#section-2.1
			operation, securitySchemes = handleAuthReqHeader(operation, securitySchemes, value)
		} else if APIKeyNames[lowerKey] {
			schemeKey := APIKeyAuthSecuritySchemeKey
			operation = addSecurity(operation, schemeKey)
			securitySchemes = updateSecuritySchemes(securitySchemes, schemeKey, NewAPIKeySecuritySchemeInHeader(key))
		} else if lowerKey == cookieTypeHeaderName {
			operation = o.addCookieParam(operation, value)
		} else {
			operation = o.addHeaderParam(operation, key, value)
		}
	}

	for key, values := range data.QueryParams {
		lowerKey := strings.ToLower(key)
		if lowerKey == AccessTokenParamKey {
			// https://datatracker.ietf.org/doc/html/rfc6750#section-2.3
			operation, securitySchemes = handleAuthQueryParam(operation, securitySchemes, values)
		} else if APIKeyNames[lowerKey] {
			schemeKey := APIKeyAuthSecuritySchemeKey
			operation = addSecurity(operation, schemeKey)
			securitySchemes = updateSecuritySchemes(securitySchemes, schemeKey, NewAPIKeySecuritySchemeInQuery(key))
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

				response.WithJSONSchema(respSchema)
			default:
				log.Infof("Treating %v as default response content type (no schema)", respContentType)
			}
		}
	}

	for key, value := range data.RespHeaders {
		response = o.addResponseHeader(response, key, value)
	}

	operation.AddResponse(data.statusCode, response)

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

func getBearerAuthClaims(bearerToken string) (jwt.MapClaims, bool) {
	if len(bearerToken) == 0 {
		log.Warnf("authZ token provided with no value, assuming authentication required anyway")
		return nil, false
	}

	// Parse the claims without validating (since we don't want to bother downloading a key)
	parser := jwt.Parser{}
	token, _, err := parser.ParseUnverified(bearerToken, jwt.MapClaims{})
	if err != nil {
		log.Warnf("authZ token is not a JWT, assuming authentication required anyway")
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Infof("authZ token had unintelligble claims, assuming authentication required anyway")
		return nil, false
	}

	return claims, true
}

func generateBearerAuthScheme(operation *spec.Operation, claims jwt.MapClaims, key string) (*spec.Operation, *spec.SecurityScheme) {
	switch key {
	case BearerAuthSecuritySchemeKey:
		// https://swagger.io/docs/specification/authentication/bearer-authentication/
		return addSecurity(operation, key), spec.NewJWTSecurityScheme()
	case OAuth2SecuritySchemeKey:
		// https://swagger.io/docs/specification/authentication/oauth2/
		// we can't know the flow type (implicit, password, clientCredentials or authorizationCode) so we choose authorizationCode for now
		scopes := getScopesFromJWTClaims(claims)
		oAuth2SecurityScheme := NewOAuth2SecurityScheme(scopes)
		return addSecurity(operation, key, scopes...), oAuth2SecurityScheme
	default:
		log.Warnf("Unsupported BearerAuth key: %v", key)
		return operation, nil
	}
}

func getScopesFromJWTClaims(claims jwt.MapClaims) []string {
	var scopes []string
	if claims == nil {
		return scopes
	}

	if scope, ok := claims["scope"]; ok {
		scopes = strings.Split(scope.(string), " ")
		log.Debugf("found OAuth token scopes: %v", scopes)
	} else {
		log.Warnf("no scopes defined in this token")
	}
	return scopes
}

func handleAuthQueryParam(operation *spec.Operation, securitySchemes spec.SecuritySchemes, values []string) (*spec.Operation, spec.SecuritySchemes) {
	if len(values) > 1 {
		// RFC 6750 does not prohibit multiple tokens, but we do not know whether
		// they would be AND or OR so we just pick the latest.
		log.Warnf("Found %v tokens in query parameters, using only the last", len(values))
		values = values[len(values)-1:]
	}

	// Use scheme as security scheme name
	securitySchemeKey := OAuth2SecuritySchemeKey
	claims, _ := getBearerAuthClaims(values[0])

	if hasSecurity(operation, securitySchemeKey) {
		// RFC 6750 states multiple methods (form, uri query, header) cannot be used.
		log.Errorf("OAuth tokens supplied with multiple methods, ignoring query param")
		return operation, securitySchemes
	}

	var scheme *spec.SecurityScheme
	operation, scheme = generateBearerAuthScheme(operation, claims, securitySchemeKey)
	if scheme != nil {
		securitySchemes = updateSecuritySchemes(securitySchemes, securitySchemeKey, scheme)
	}
	return operation, securitySchemes
}

func handleAuthReqHeader(operation *spec.Operation, securitySchemes spec.SecuritySchemes, value string) (*spec.Operation, spec.SecuritySchemes) {
	if strings.HasPrefix(value, BasicAuthPrefix) {
		// https://swagger.io/docs/specification/authentication/basic-authentication/
		// Use scheme as security scheme name
		key := BasicAuthSecuritySchemeKey
		operation = addSecurity(operation, key)
		securitySchemes = updateSecuritySchemes(securitySchemes, key, NewBasicAuthSecurityScheme())
	} else if strings.HasPrefix(value, BearerAuthPrefix) {
		// https://swagger.io/docs/specification/authentication/bearer-authentication/
		// https://datatracker.ietf.org/doc/html/rfc6750#section-2.1
		// Use scheme as security scheme name. For OAuth, we should consider checking
		// supported scopes to allow multiple defs.
		key := BearerAuthSecuritySchemeKey
		claims, found := getBearerAuthClaims(strings.TrimPrefix(value, BearerAuthPrefix))
		if found {
			key = OAuth2SecuritySchemeKey
		}

		if hasSecurity(operation, key) {
			// RFC 6750 states multiple methods (form, uri query, header) cannot be used.
			log.Error("OAuth tokens supplied with multiple methods, ignoring header")
			return operation, securitySchemes
		}

		var scheme *spec.SecurityScheme
		operation, scheme = generateBearerAuthScheme(operation, claims, key)
		if scheme != nil {
			securitySchemes = updateSecuritySchemes(securitySchemes, key, scheme)
		}
	} else {
		log.Warnf("ignoring unknown authorization header value (%v)", value)
	}
	return operation, securitySchemes
}

func addSecurity(op *spec.Operation, name string, scopes ...string) *spec.Operation {
	// https://swagger.io/docs/specification/authentication/
	// We will treat multiple authentication types as an OR
	// (Security schemes combined via OR are alternatives – any one can be used in the given context)
	securityRequirement := spec.NewSecurityRequirement()

	if len(scopes) > 0 {
		securityRequirement[name] = scopes
	} else {
		// We must use an empty array as the scopes, otherwise it will create invalid swagger
		securityRequirement[name] = []string{}
	}

	if op.Security == nil {
		op.Security = spec.NewSecurityRequirements()
	}
	op.Security.With(securityRequirement)

	return op
}

func hasSecurity(op *spec.Operation, name string) bool {
	if op.Security == nil {
		return false
	}

	for _, securityScheme := range *op.Security {
		if _, ok := securityScheme[name]; ok {
			return true
		}
	}
	return false
}
