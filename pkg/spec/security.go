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
	log "github.com/sirupsen/logrus"
)

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

func updateSecurityDefinitionsFromOperation(sd spec.SecurityDefinitions, op *spec.Operation) spec.SecurityDefinitions {
	if op == nil {
		return sd
	}

	for _, securityGroup := range op.Security {
		for sdKey := range securityGroup {
			sd = updateSecurityDefinitions(sd, sdKey)
		}
	}

	return sd
}

func updateSecurityDefinitions(sd spec.SecurityDefinitions, sdKey string) spec.SecurityDefinitions {
	// we can override SecurityDefinitions if exists since it has the same key and value
	switch sdKey {
	case BasicAuthSecurityDefinitionKey:
		sd[BasicAuthSecurityDefinitionKey] = spec.BasicAuth()
	case OAuth2SecurityDefinitionKey:
		// we can't know the flow type (implicit, password, application or accessCode) so we choose accessCode for now
		sd[OAuth2SecurityDefinitionKey] = spec.OAuth2AccessToken(authorizationURL, tknURL)
	// TODO: Add support for API Key
	// case APIKeyAuthSecurityDefinitionKey:
	//	spec.APIKeyAuth()
	default:
		log.Warnf("Unsupported security definition key: %v", sdKey)
	}

	return sd
}
