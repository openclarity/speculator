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
	spec "github.com/getkin/kin-openapi/openapi3"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

type OAuth2Claims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

const (
	BasicAuthSecuritySchemeKey  = "BasicAuth"
	APIKeyAuthSecuritySchemeKey = "ApiKeyAuth"
	OAuth2SecuritySchemeKey     = "OAuth2"
	BearerAuthSecuritySchemeKey = "BearerAuth"

	BearerAuthPrefix = "Bearer "
	BasicAuthPrefix  = "Basic "

	AccessTokenParamKey = "access_token"

	tknURL           = "https://example.com/oauth2/token"
	authorizationURL = "https://example.com/oauth2/authorize"

	apiKeyType      = "apiKey"
	basicAuthType   = "http"
	basicAuthScheme = "basic"
	oauth2Type      = "oauth2"
)

// APIKeyNames is set of names of headers or query params defining API keys.
// This should be runtime configurable, of course.
// Note: keys should be lowercase.
var APIKeyNames = map[string]bool{
	"key":     true, // Google
	"api_key": true,
}

func newAPIKeySecurityScheme(name string) *spec.SecurityScheme {
	// https://swagger.io/docs/specification/authentication/api-keys/
	return &spec.SecurityScheme{
		Type: apiKeyType,
		Name: name,
	}
}

func NewAPIKeySecuritySchemeInHeader(name string) *spec.SecurityScheme {
	return newAPIKeySecurityScheme(name).WithIn(spec.ParameterInHeader)
}

func NewAPIKeySecuritySchemeInQuery(name string) *spec.SecurityScheme {
	return newAPIKeySecurityScheme(name).WithIn(spec.ParameterInQuery)
}

func NewBasicAuthSecurityScheme() *spec.SecurityScheme {
	// https://swagger.io/docs/specification/authentication/basic-authentication/
	return &spec.SecurityScheme{
		Type:   basicAuthType,
		Scheme: basicAuthScheme,
	}
}

func NewOAuth2SecurityScheme(scopes []string) *spec.SecurityScheme {
	// https://swagger.io/docs/specification/authentication/oauth2/
	// we can't know the flow type (implicit, password, clientCredentials or authorizationCode)
	// so we choose authorizationCode for now
	return &spec.SecurityScheme{
		Type: oauth2Type,
		Flows: &spec.OAuthFlows{
			AuthorizationCode: &spec.OAuthFlow{
				AuthorizationURL: authorizationURL,
				TokenURL:         tknURL,
				Scopes:           createOAuthFlowScopes(scopes, []string{}),
			},
		},
	}
}

func updateSecuritySchemesFromOperation(securitySchemes spec.SecuritySchemes, op *spec.Operation) spec.SecuritySchemes {
	if op == nil || op.Security == nil {
		return securitySchemes
	}

	// Note: usage goes in the other direction; i.e., the security schemes do contain more detail, and operations
	// (security requirements) reference those schemes. The reference is required to be valid (i.e., the
	// name in the operation MUST be present in the security schemes) for OAuth spec v2.0.  Here we assume
	// schemes are generic to push the operation's security requirements into the general security schemes.
	for _, securityGroup := range *op.Security {
		for key := range securityGroup {
			var scheme *spec.SecurityScheme
			switch key {
			case BasicAuthSecuritySchemeKey:
				scheme = NewBasicAuthSecurityScheme()
			case OAuth2SecuritySchemeKey:
				// we can't know the flow type (implicit, password, clientCredentials or authorizationCode) so
				// we choose authorizationCode for now
				scheme = NewOAuth2SecurityScheme(nil)
			case BearerAuthSecuritySchemeKey:
				scheme = spec.NewJWTSecurityScheme()
			case APIKeyAuthSecuritySchemeKey:
				// Use random key since it is not specified
				for apiKeyName := range APIKeyNames {
					scheme = NewAPIKeySecuritySchemeInHeader(apiKeyName)
					break
				}
			default:
				log.Warnf("Unsupported security definition key: %v", key)
			}
			securitySchemes = updateSecuritySchemes(securitySchemes, key, scheme)
		}
	}

	return securitySchemes
}

func updateSecuritySchemes(securitySchemes spec.SecuritySchemes, key string, securityScheme *spec.SecurityScheme) spec.SecuritySchemes {
	// we can override SecuritySchemes if exists since it has the same key and value
	switch key {
	case BasicAuthSecuritySchemeKey, OAuth2SecuritySchemeKey, APIKeyAuthSecuritySchemeKey:
		securitySchemes[key] = &spec.SecuritySchemeRef{Value: securityScheme}
	default:
		log.Warnf("Unsupported security definition key: %v", key)
	}

	return securitySchemes
}

func createOAuthFlowScopes(scopes []string, descriptions []string) map[string]string {
	flowScopes := make(map[string]string)
	if len(descriptions) > 0 {
		if len(descriptions) < len(scopes) {
			log.Errorf("too few descriptions (%v) supplied for security scheme scopes (%v)", len(descriptions), len(scopes))
		}
		for idx, scope := range scopes {
			if idx < len(descriptions) {
				flowScopes[scope] = descriptions[idx]
			} else {
				flowScopes[scope] = ""
			}
		}
	} else {
		for _, scope := range scopes {
			flowScopes[scope] = ""
		}
	}
	return flowScopes
}
