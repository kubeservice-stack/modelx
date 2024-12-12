package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"kubegems.io/modelx/pkg/config"
)

func TestOIDCAuthEmpty(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(OIDCAuthFunc()) //Metricd
	router.GET("/test1", func(c *gin.Context) {
		c.String(http.StatusOK, "dongjiang")
	})

	req := httptest.NewRequest(http.MethodGet, "/test1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("dongjiang", w.Body.String())
}

func TestOIDCAuth(t *testing.T) {
	assert := assert.New(t)
	config.GlobalModelxdOptions.OIDC = &config.OIDCOptions{
		Issuer: "http:/127.0.0.1/",
	}
	router := gin.New()
	router.Use(OIDCAuthFunc()) //Metricd
	router.GET("/test1", func(c *gin.Context) {
		c.String(http.StatusOK, "dongjiang")
	})

	req := httptest.NewRequest(http.MethodGet, "/test1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusUnauthorized, w.Code)
}
