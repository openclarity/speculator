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
	"github.com/go-openapi/spec"
	"github.com/golang-jwt/jwt/v4"
	log "github.com/sirupsen/logrus"
)

type OAuth2Claims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

const (
	BasicAuthSecurityDefinitionKey  = "BasicAuth"
	APIKeyAuthSecurityDefinitionKey = "ApiKeyAuth"
	OAuth2SecurityDefinitionKey     = "OAuth2"

	BearerAuthPrefix = "Bearer "
	BasicAuthPrefix  = "Basic "

	AccessTokenParamKey = "access_token"

	tknURL           = "https://example.com/oauth/token"
	authorizationURL = "https://example.com/oauth/authorize"
)

var (
	// TODO: This should be runtime configurable, of course.
	// Note: keys should be lower case.
	APIKeyNames = map[string]bool{
		"key":     true, // Google
		"api_key": true,
	}
)

func updateSecurityDefinitionsFromOperation(sd spec.SecurityDefinitions, op *spec.Operation) spec.SecurityDefinitions {
	if op == nil {
		return sd
	}

	// Note: usage goes in the other direction; i.e., the defs do contain more detail, and operations
	// (security requirements) reference those definitions. The reference is required to be valid (i.e., the
	// name in the operation MUST be present in the security defs) for OAuth spec v2.0.  Here we assume
	// defs are generic to push the operation's security requirements into the general security definitions.
	for _, securityGroup := range op.Security {
		for sdKey := range securityGroup {
			var scheme *spec.SecurityScheme
			switch sdKey {
			case BasicAuthSecurityDefinitionKey:
				scheme = spec.BasicAuth()
			case OAuth2SecurityDefinitionKey:
				// we can't know the flow type (implicit, password, application or accessCode) so
				// we choose accessCode for now
				scheme = spec.OAuth2AccessToken(authorizationURL, tknURL)
			case APIKeyAuthSecurityDefinitionKey:
				// Use random key since it is not specified
				for key := range APIKeyNames {
					scheme = spec.APIKeyAuth(key, apiKeyInHeader)
					break
				}
			default:
				log.Warnf("Unsupported security definition key: %v", sdKey)
			}
			sd = updateSecurityDefinitions(sd, sdKey, scheme)
		}
	}

	return sd
}

func updateSecurityDefinitions(sd spec.SecurityDefinitions, sdKey string, secScheme *spec.SecurityScheme) spec.SecurityDefinitions {
	// we can override SecurityDefinitions if exists since it has the same key and value
	switch sdKey {
	case BasicAuthSecurityDefinitionKey, OAuth2SecurityDefinitionKey, APIKeyAuthSecurityDefinitionKey:
		sd[sdKey] = secScheme
	default:
		log.Warnf("Unsupported security definition key: %v", sdKey)
	}

	return sd
}

func updateSecuritySchemeScopes(sss *spec.SecurityScheme, scopes []string, descriptions []string) *spec.SecurityScheme {
	if sss.Scopes == nil {
		sss.Scopes = make(map[string]string)
	}
	if len(descriptions) > 0 {
		if len(descriptions) < len(scopes) {
			log.Errorf("too few descriptions (%v) supplied for security scheme scopes (%v)", len(descriptions), len(scopes))
		}
		for idx, scope := range scopes {
			sss.Scopes[scope] = descriptions[idx]
		}
	} else {
		for _, scope := range scopes {
			sss.Scopes[scope] = ""
		}
	}
	return sss
}
