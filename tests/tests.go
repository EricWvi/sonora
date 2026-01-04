package tests

import (
	"bytes"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

func TestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware to extract and set Action from query parameters
	router.Use(func(c *gin.Context) {
		action := c.Request.URL.Query().Get("Action")
		if action != "" {
			c.Set("Action", action)
		}
		c.Next()
	})
	return router
}

func ServeJSON(router *gin.Engine, method, url string, body []byte) (int, []byte) {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	return w.Code, w.Body.Bytes()
}
