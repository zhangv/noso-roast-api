package auth

import (
	"hahacking/common/app"
	"hahacking/database"
	"hahacking/i18n"
	"hahacking/middleware"
	"hahacking/models/user"
	"hahacking/service"
	userservice "hahacking/service/user"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/auth")
	{
		APIs.POST("login", Login)
		APIs.POST("register", Register)
		APIs.POST("refresh", RefreshToken)
	}
	userAPI := r.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{

	}
}

// RefreshTokenRequest 刷新令牌请求结构
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func RefreshToken(c *gin.Context) {
	appG := app.Gin{Cxt: c}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.Error(400, "绑定请求中的json失败，可能是格式问题", err)
		return
	}

	refreshToken := req.RefreshToken
	if refreshToken == "" {
		appG.Error(400, "refreshToken is required", nil)
		return
	}

	userAgent := c.GetHeader("User-Agent")

	// 验证refresh token是否在redis中
	_, err := userservice.GetTokenStore().Get(refreshToken)
	if err != nil {
		appG.Error(400, "refreshToken can not found:"+err.Error(), nil)
		return
	}
	userService := service.NewUserService()
	claims, err := userService.ParseJWT(refreshToken)
	if err != nil {
		appG.Error(400, "refreshToken parse failed:"+err.Error(), nil)
		return
	}
	//导致token无法刷新 - useragent很容易变化
	// if feature.UserAgent != userAgent {
	// 	appG.Error(400, "user agent mismatch", nil)
	// 	// 吊销全部token
	// 	userservice.GetTokenStore().Revoke(claims.Uid)
	// }

	token, refreshToken, err := userService.RefreshToken(refreshToken, userAgent, &userservice.Claims{
		Username: claims.Username,
		Uid:      claims.Uid,
	})
	if err != nil {
		appG.Error(500, "刷新token失败", err)
		return
	}
	appG.Success("", map[string]any{
		"token":        token,
		"refreshToken": refreshToken,
	})
}

type LoginForm struct {
	Email    string `form:"email" validate:"required,email"`
	Password string `validate:"required"`
}

func (form *LoginForm) Validate() (*user.User, map[string]string) {
	var result = make(map[string]string)
	if form.Email == "" {
		result["email"] = "请填写邮箱或手机号码"
	}

	idtype := "nickname"
	var mobileRegex = `^1[3-9]\d{9}$`
	if matched, _ := regexp.MatchString(mobileRegex, form.Email); matched {
		idtype = "mobile"
	}
	var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if matched, _ := regexp.MatchString(emailRegex, form.Email); matched {
		idtype = "email"
	}

	var u *user.User
	if idtype == "nickname" {
		u, _ = user.GetByNickname(form.Email)
	}
	if idtype == "email" {
		u, _ = user.GetByEmail(form.Email)
	}
	if idtype == "mobile" {
		u, _ = user.GetByMobile(form.Email)
	}

	if u == nil {
		result["email"] = "用户(" + form.Email + ")不存在"
		return nil, result
	}
	var userpass = user.CreatePassword(form.Password)
	if userpass != u.Password {
		result["password"] = "密码错误"
		return nil, result
	}
	return u, nil
}

func Login(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form LoginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	u, err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	userService := service.NewUserService()
	userAgent := c.GetHeader("User-Agent")
	token, refreshToken, err2 := userService.GenerateJWT(u.ID, u.Nickname, userAgent)
	if err2 != nil {
		appG.Error(500, "生成token失败", err2)
		return
	}
	appG.Success("", map[string]any{
		"token":        token,
		"refreshToken": refreshToken,
		"userInfo":     u})
}

type RegisterForm struct {
	// Nickname string `json:"nickname" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	// Mobile   string `json:"mobile" validate:"required"`
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"passwordconfirm" validate:"required,eqfield=Password"`
}

func (form *RegisterForm) Validate() map[string]string {
	validate := validator.New()
	result := make(map[string]string)
	if err := validate.Struct(form); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			switch err.Field() {
			// case "Nickname":
			// 	result["nickname"] = i18n.T("auth.required", i18n.T("auth.nickname"))
			case "Email":
				switch err.Tag() {
				case "required":
					result["email"] = i18n.T("auth.required", i18n.T("auth.email"))
				case "email":
					result["email"] = i18n.T("auth.invalid", i18n.T("auth.email"))
				}
			case "Password":
				result["password"] = i18n.T("auth.required", i18n.T("auth.password"))
			case "PasswordConfirm":
				switch err.Tag() {
				case "required":
					result["passwordconfirm"] = i18n.T("auth.required", i18n.T("auth.passwordconfirm"))
				case "eqfield":
					result["passwordconfirm"] = i18n.T("auth.mismatch", i18n.T("auth.passwordconfirm"))
				}
			}
		}
	}

	if len(result) > 0 {
		return result
	}

	// Check if email already exists
	if u, _ := user.GetByEmail(form.Email); u != nil {
		result["email"] = "该邮箱已被注册"
	}

	// Check if mobile already exists (if you add mobile field)
	// if u, _ := user.GetByMobile(form.Mobile); u != nil {
	// 	result["mobile"] = "该手机号已被注册"
	// }

	return result
}

func Register(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form RegisterForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}

	err := form.Validate()
	if len(err) > 0 {
		appG.Response(200, 1000, err)
		return
	}

	// Create new user
	newUser := &user.User{
		// Nickname: form.Nickname,
		Email: &form.Email,
		// Mobile:       &form.Mobile,
		//这里从表单里得到的是 sha256加密后的结果
		Password:     user.CreatePassword(form.Password),
		Type:         "user",
		Status:       "active",
		RegisterTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	db := database.GetConn()
	if err := db.Create(newUser).Error; err != nil {
		appG.Error(500, "创建用户失败", err)
		return
	}

	// Generate token for the new user
	userService := service.NewUserService()
	userAgent := c.GetHeader("User-Agent")
	token, refreshToken, _ := userService.GenerateJWT(newUser.ID, newUser.Nickname, userAgent)

	appG.Success("注册成功", map[string]any{
		"token":        token,
		"refreshToken": refreshToken,
		"userInfo":     newUser,
	})
}
