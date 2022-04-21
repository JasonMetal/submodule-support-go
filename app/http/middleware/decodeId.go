package middleware

import (
	"github.com/gin-gonic/gin"
	strLib "idea-go/helper/strings"
)

// ParseRoundId 解析round_id
func ParseRoundId() gin.HandlerFunc {
	return func(c *gin.Context) {
		roundId := 0
		encryptId := c.GetHeader("encryptid")
		if encryptId != "" {

			encryptId = strLib.Base64Decode(encryptId)
			//strings.Cut()
		}
		c.Set("id", roundId)
		c.Next()
	}
}

// ParseRoundIdWithErr 解析round_id,不存在或错误则报错
func ParseRoundIdWithErr() gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
