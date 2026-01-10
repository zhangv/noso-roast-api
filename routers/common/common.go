package common

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"hahacking/common/app"
	"hahacking/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/common")
	{
		APIs.GET("upconfig/:name", UpConfig)
	}
}

func UpConfig(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Param("name")
	cfg := config.Get("upyun")[name]

	if cfg == nil {
		appG.Error(404, "配置不存在", nil)
		return
	}
	cfg2 := make(map[string]interface{})
	cfgMap := cfg.(map[string]interface{})
	cfg2["bucket"] = cfgMap["bucket"]
	expiration, _ := strconv.Atoi(cfgMap["formexpiration"].(string))
	cfg2["expiration"] = time.Now().Unix() + int64(expiration)
	cfg2["save-key"] = cfgMap["save-key"]
	cfg2["allow-file-type"] = cfgMap["allow-file-type"]
	cfg2["content-length-range"] = cfgMap["content-length-range"]

	policy, _ := json.Marshal(cfg2)
	policyStr := base64.StdEncoding.EncodeToString(policy)
	formapisecret := cfgMap["formapisecret"].(string)
	signature := md5.Sum([]byte(policyStr + "&" + formapisecret))
	signatureStr := hex.EncodeToString(signature[:])
	cfg2["policy"] = policyStr
	cfg2["signature"] = signatureStr
	cfg2["cdn"] = cfgMap["cdn"]
	appG.Success("", cfg2)
}
