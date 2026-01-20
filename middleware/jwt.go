package middleware

import (
	"os"
	"sync"

	"github.com/EricWvi/sonora/config"
	"github.com/EricWvi/sonora/log"
	"github.com/EricWvi/sonora/model"
	"github.com/EricWvi/sonora/service"
	"github.com/gin-gonic/gin"
)

var emailToID map[string]uint

var lock sync.RWMutex

func InitJWTMap() {
	m, err := model.CreateEmailToIDMap(config.ContextDB(log.WorkerCtx))
	if err != nil {
		log.Error(log.WorkerCtx, err.Error())
		os.Exit(1)
	}
	emailToID = m
}

func readMap(email string) (uint, bool) {
	lock.RLock()
	defer lock.RUnlock()
	id, ok := emailToID[email]
	return id, ok
}

func writeMap(email string) uint {
	lock.Lock()
	defer lock.Unlock()
	if id, ok := emailToID[email]; ok {
		return id
	} else {
		id, err := model.CreateUser(config.ContextDB(log.WorkerCtx), email)
		if err != nil {
			log.Error(log.WorkerCtx, err.Error())
		}
		emailToID[email] = id
		return id
	}
}

func getId(email string) uint {
	if id, ok := readMap(email); ok {
		return id
	} else {
		return writeMap(email)
	}
}

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Request.Header.Get("Remote-Email")
		if len(email) == 0 {
			token := c.Request.Header.Get("TAURI_TOKEN")
			if token == "" {
				c.Set("Action", "Auth")
				c.Set("UserId", uint(0))
				return
			}
			decryptedEmail, err := service.Decrypt(service.CryptKey, token)
			if err != nil {
				c.Set("UserId", uint(0))
				return
			}
			email = decryptedEmail
			if len(email) == 0 {
				c.Set("UserId", uint(0))
				return
			} else {
				c.Set("UserId", getId(email))
			}
		} else {
			c.Set("UserId", getId(email))
		}
	}
}

func GetUserId(c *gin.Context) uint {
	return c.GetUint("UserId")
}
