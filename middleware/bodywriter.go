package middleware

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// use BodyWriter to retrieve response body
func BodyWriter() gin.HandlerFunc {
	return func(c *gin.Context) {
		writer := &bodyWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
		}
		c.Writer = writer

		c.Set("bodyWriter", writer)
	}
}
