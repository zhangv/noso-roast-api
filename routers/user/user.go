package user

import (
	"hahacking/common/app"
	"hahacking/middleware"
	models "hahacking/models/user"
	"hahacking/utils"
	"hahacking/utils/api"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/user")
	{
		APIs.GET("meta/:ids", Meta)
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.GET("", Search)
		userAPI.GET(":id", GetById)
		userAPI.POST("", Update)
		userAPI.DELETE(":id", Remove)
		userAPI.POST("/modifypassword", ModifyPassword)
	}
}

func Meta(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	name := c.Params.ByName("ids")
	result := models.Meta(name)
	appG.Success("", result)
}

func GetById(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id, err := strconv.Atoi(c.Params.ByName("id"))
	o, err := models.GetById(uint(id))

	if err != nil {
		appG.NotFound("抱歉未找到相关信息")
		return
	}
	appG.Success("", o)
}

func Search(c *gin.Context) {
	var f models.SearchForm
	appG := app.Gin{Cxt: c}

	if c.ShouldBindQuery(&f) == nil {
		log.Println(f)
	}
	result, count := models.Search(f)
	appG.Success("", utils.Response{
		"objects": result,
		"total":   count,
	})
}

type UpdateForm struct {
	ID           uint    `json:"id"`
	Nickname     string  `json:"nickname"`
	Email        *string `json:"email"`
	Mobile       *string `json:"mobile"`
	Type         string  `json:"type"`
	Status       string  `json:"status"`
	PriceLevelID uint    `json:"pricelevel_id" form:"pricelevel_id"`
	Password     string  `json:"password"`
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.Nickname == "" {
		result["nickname"] = "请填写昵称"
	}
	if *form.Email == "" {
		form.Email = nil
	}
	if *form.Mobile == "" {
		form.Mobile = nil
	}
	if form.Email == nil && form.Mobile == nil {
		result["email"] = "请填写邮箱或手机号"
	} else {
		if form.Email != nil {
			if !utils.IsEmail(*form.Email) {
				result["email"] = "请填写正确的邮箱"
			}
		}
		if form.Mobile != nil {
			if !utils.IsMobile(*form.Mobile) {
				result["mobile"] = "请填写正确的手机号"
			}
		}
	}

	return result
}

func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form UpdateForm
	var o models.User
	formErrors := api.Bind(&form, &o, c)
	if len(formErrors) > 0 {
		appG.Error(400, "表单验证失败", formErrors)
		return
	}
	if form.PriceLevelID == 0 {
		o.PriceLevelID = nil
	}
	o.ID = form.ID //注意：需要手动设置一下，因为model中的id是嵌入的models.Model，在绑定的时候不会自动设置
	err := models.Update(&o)
	if err != nil {
		appG.Error(500, "更新失败", err)
		return
	}
	appG.Success("", o)
}

func Remove(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	id := c.Params.ByName("id")
	idInt, err := strconv.Atoi(id)
	user, err := models.GetById(uint(idInt))
	err = user.Remove()
	if err != nil {
		appG.Error(500, "删除失败", err)
		return
	}
	appG.Success("", nil)
}

type ModifyPasswordForm struct {
	UserID uint `json:"user_id"`
	//前端输入+sha256后的串
	OldPassword     string `json:"oldpassword"`
	NewPassword     string `json:"newpassword"`
	ConfirmPassword string `json:"passwordconfirm"`
}

func (form *ModifyPasswordForm) Validate() map[string]string {
	var result = make(map[string]string)
	if form.UserID == 0 {
		result["user_id"] = "用户ID不能为空"
		return result
	}
	if form.OldPassword == "" {
		result["oldpassword"] = "请填写旧密码"
	}
	if form.NewPassword == "" {
		result["newpassword"] = "请填写新密码"
	}
	if form.ConfirmPassword == "" {
		result["passwordconfirm"] = "请填写确认密码"
	}
	if form.NewPassword != form.ConfirmPassword {
		result["passwordconfirm"] = "新密码和确认密码不一致"
	}
	//检查密码是否正确
	user, err := models.GetById(form.UserID)
	if err != nil {
		result["user_id"] = "用户不存在"
		return result
	}
	if user.Password != models.CreatePassword(form.OldPassword) {
		result["oldpassword"] = "旧密码不正确"
	}
	return result
}
func ModifyPassword(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form ModifyPasswordForm

	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	formErrors := form.Validate()
	if len(formErrors) > 0 {
		appG.Error(400, "表单验证失败", formErrors)
		return
	}

	u, err := models.GetById(uint(form.UserID))
	if err != nil {
		appG.Error(500, "修改密码失败", err)
		return
	}
	u.Password = models.CreatePassword(form.NewPassword)
	err = models.Update(u)
	if err != nil {
		appG.Error(500, "修改密码失败", err)
		return
	}
	appG.Success("", nil)
}
