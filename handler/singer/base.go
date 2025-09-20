package singer

import (
	"github.com/EricWvi/sonora/handler"
	"github.com/gin-gonic/gin"
)

type Base struct{}

func DefaultHandler(c *gin.Context) {
	handler.Dispatch(c, Base{})
}
