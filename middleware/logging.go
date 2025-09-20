package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var noLoggingActions = []string{"UpdateTiptap"}

func shouldLogging(action string) bool {
	return !slices.Contains(noLoggingActions, action)
}

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now().UTC()

		// set RequestId and Action
		// Get existing ID or generate new
		if c.GetString(log.RequestIDCtxKey) == "" {
			c.Set(log.RequestIDCtxKey, uuid.NewString())
		}
		action := c.Request.URL.Query().Get("Action")
		if len(action) == 0 {
			handler.ReplyError(c, http.StatusBadRequest, "request action is missing")
			c.Abort()
			return
		}
		c.Set("Action", action)

		// log request fields
		fields := []any{
			"ip", c.ClientIP(),
		}

		if shouldLogging(action) {
			// log request body
			requestBody, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Error(c, "failed to read request body")
				handler.ReplyError(c, http.StatusInternalServerError, "failed to read request body")
				c.Abort()
				return
			}
			// c.Set("RequestBody", string(requestBody))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody)) // restore

			if c.ContentType() == "application/json" {
				fields = append(fields, "body", string(requestBody))
			}
		}

		log.Info(c, c.Request.Method+" "+c.Request.URL.String(), fields...)

		c.Next()

		// Calculates the latency.
		end := time.Now().UTC()
		latency := end.Sub(start)

		// get code and message
		rsp := handler.Response{}
		blw, _ := c.Get("bodyWriter")
		if err := json.Unmarshal(blw.(*bodyWriter).body.Bytes(), &rsp); err != nil {
			log.Errorf(c, "response body can not unmarshal to handler.Response struct, body: `%s`", blw.(*bodyWriter).body.Bytes())
			c.Abort()
		} else {
			fields := []any{
				"code", rsp.Code,
				"latency", latency,
			}
			if rsp.Code >= 400 {
				log.Info(c, rsp.Message.(string), fields...)
			} else {
				log.Info(c, "ok", fields...)
			}
		}
	}
}
