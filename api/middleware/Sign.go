package middleware

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func Sign() gin.HandlerFunc {
	return func(c *gin.Context) {
		var session string
		c.Set("session", session)
		t := time.Now().Unix()
		t = t / 60
		d := strconv.Itoa(int(t))

		key := "af93e7c5f9f8567696dc2b2e677188af"
		sign := base64.StdEncoding.EncodeToString([]byte(MD5(MD5(d + key))))

		c.Set("sign", sign)
		c.Header("Content-Type", "text/json;charset=utf-8")
		c.Next()
	}
}

func MD5(str string) string {
	ctx := md5.New()
	ctx.Write([]byte(str))
	sign := hex.EncodeToString(ctx.Sum(nil))
	return sign
}