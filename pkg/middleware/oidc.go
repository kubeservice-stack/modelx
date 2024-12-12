/*
Copyright 2024 The KubeService-Stack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package middleware

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"kubegems.io/modelx/pkg/config"
	"kubegems.io/modelx/pkg/response"
)

type Username struct{}

func UsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value(Username{}).(string); ok {
		return username
	}
	return ""
}

func NewUsernameContext(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, Username{}, username)
}

const (
	OIDCAUTH       = "OIDC"
	OIDCAUTHWEIGHT = 999
	OIDCProvider
)

func OIDCAuthFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.GlobalModelxdOptions.OIDC != nil && config.GlobalModelxdOptions.OIDC.Issuer != "" {
			reqCxt := c.Request.Context()
			ctx := oidc.InsecureIssuerURLContext(context.TODO(), config.GlobalModelxdOptions.OIDC.Issuer)
			provider, err := oidc.NewProvider(ctx, config.GlobalModelxdOptions.OIDC.Issuer)
			if err != nil {
				ginLogger.Error("oidc new provider error", zap.Error(err))
				response.ResponseError(c.Writer, response.NewUnauthorizedError("oidc new provider error"))
				c.Abort()
				return
			}

			verifier := provider.Verifier(&oidc.Config{
				SkipClientIDCheck: true,
				SkipIssuerCheck:   true,
			})

			headerAuthorzation := c.GetHeader("Authorization")
			token := strings.TrimPrefix(headerAuthorzation, "Bearer ")
			if token == "" {
				queries := c.Request.URL.Query()
				for _, k := range []string{"token", "access_token"} {
					if token = queries.Get(k); token != "" {
						break
					}
				}
			}
			if len(token) == 0 {
				response.ResponseError(c.Writer, response.NewUnauthorizedError("missing access token"))
				c.Abort()
				return
			}
			idtoken, err := verifier.Verify(reqCxt, token)
			if err != nil {
				response.ResponseError(c.Writer, response.NewUnauthorizedError("invalid access token"))
				c.Abort()
				return
			}

			reqCxt = NewUsernameContext(reqCxt, idtoken.Subject)
			c.Request = c.Request.WithContext(reqCxt)
		}
		// 处理请求
		c.Next()
	}
}

func init() {
	Register(&Instance{Name: OIDCAUTH, F: OIDCAuthFunc, Weight: OIDCAUTHWEIGHT})
}
