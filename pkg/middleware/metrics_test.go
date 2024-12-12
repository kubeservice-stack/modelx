package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeservice-stack/common/pkg/metrics"
	"github.com/stretchr/testify/assert"
)

func Test_Metrics(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(MetricsFunc()) //Metricd
	router.GET("/metrics", gin.WrapH(metrics.DefaultTallyScope.Reporter.HTTPHandler()))
	router.GET("/test1", func(c *gin.Context) {
		c.String(http.StatusOK, "dongjiang")
	})

	req := httptest.NewRequest(http.MethodGet, "/test1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("dongjiang", w.Body.String())

	time.Sleep(6 * time.Second)

	req1 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(http.StatusOK, w1.Code)
	log.Println(w1.Body.String())
}

func Test_Metrics_CallMuti(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(MetricsFunc()) //Metricd
	router.GET("/metrics", gin.WrapH(metrics.DefaultTallyScope.Reporter.HTTPHandler()))
	router.GET("/test1", func(c *gin.Context) {
		c.String(http.StatusOK, "dongjiang")
	})

	for index := 0; index < 10; index++ {
		req := httptest.NewRequest(http.MethodGet, "/test1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(http.StatusOK, w.Code)
		assert.Equal("dongjiang", w.Body.String())

	}
	time.Sleep(6 * time.Second)

	req1 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(http.StatusOK, w1.Code)
	log.Println(w1.Body.String())
}
