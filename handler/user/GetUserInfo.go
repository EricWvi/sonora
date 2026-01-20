package user

import (
	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/handler"
	"github.com/EricWvi/sonora/middleware"
	"github.com/EricWvi/sonora/model"
	"github.com/gin-gonic/gin"
)

func (b Base) GetUserInfo(c *gin.Context, req *GetUserInfoRequest) *GetUserInfoResponse {
	// Get user ID from JWT middleware
	userId := middleware.GetUserId(c)
	if userId == 0 {
		handler.Errorf(c, "user not authenticated")
		return nil
	}

	// Query user from database
	user := &model.User{}
	m := model.WhereMap{model.Id: userId}

	if err := user.Get(config.ContextDB(c), m); err != nil {
		handler.Errorf(c, "%s", err.Error())
		return nil
	}

	return &GetUserInfoResponse{
		Email:    user.Email,
		Avatar:   user.Avatar,
		Username: user.Username,
		Language: user.Language,
	}
}

type GetUserInfoRequest struct {
	// No parameters needed - user ID comes from JWT
}

type GetUserInfoResponse struct {
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
	Language string `json:"language"`
}
