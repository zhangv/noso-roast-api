package order

import (
	"hahacking/common/app"
	"hahacking/middleware"

	"hahacking/models/order"
	models "hahacking/models/order"
	"hahacking/utils"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup) {
	APIs := r.Group("/order")
	{
		APIs.GET("meta/:ids", Meta)
	}
	userAPI := APIs.Group("")
	userAPI.Use(middleware.JWTAuth())
	userAPI.Use(middleware.BaseAuth("user"))
	{
		userAPI.GET("", Search)
		userAPI.GET(":id", GetById)
		userAPI.GET(":id/items", GetItems)
		userAPI.GET(":id/payment", GetPayment)
		userAPI.GET(":id/refunds", GetRefunds)
		userAPI.DELETE(":id", Remove)
		userAPI.POST("update", Update)
		userAPI.POST("create", Create)
		userAPI.POST("place", Place)
		userAPI.POST("settle", Settle)
		userAPI.POST("refund", Refund)
		// userAPI.POST("refundUndo/:id", RefundUndo) //TODO 需要退款撤销吗？
		userAPI.POST("afterSettle/:id", AfterSettle) //结算后更新库存 - 临时
		userAPI.POST("settleUndo/:id", SettleUndo)
		userAPI.GET("getPriceLevelDiscount", GetPriceLevelDiscount)
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
	idparam := c.Params.ByName("id")
	id, err := strconv.Atoi(idparam)
	var o *models.Order
	if err != nil {
		o, err = models.GetBySn(idparam)
	} else {
		o, err = models.GetById(uint(id))
	}

	if err != nil {
		appG.NotFound("抱歉未找到订单")
		return
	}
	appG.Success("", o)
}

func GetItems(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	orderid, _ := strconv.Atoi(c.Params.ByName("id"))
	order, _ := models.GetById(uint(orderid))
	items := order.Items()
	appG.Success("", items)
}

func GetPayment(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	orderid, _ := strconv.Atoi(c.Params.ByName("id"))
	order, _ := models.GetById(uint(orderid))
	payment := order.Payment()
	appG.Success("", payment)
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

/*
* 主要是对订单Item的确认
 */
func Create(c *gin.Context) {
	Update(c)
}

func Update(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var f models.UpdateForm
	if err := c.ShouldBindJSON(&f); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	if err := f.Validate(); err != nil {
		appG.Response(200, 1000, err)
		return
	}

	o, err := models.Update(&f)
	if err != nil {
		appG.Error(400, "更新失败", err)
		return
	}
	appG.Success("", o)
}

/*
* 主要是订单支付方式（组合）的确认
 */
func Place(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.PlaceForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if err != nil {
		appG.Response(200, 1000, err)
		return
	}
	order, _ := models.GetById(form.OrderID)
	if order == nil {
		appG.Error(400, "订单不存在", nil)
		return
	}

	orderPayment, err2 := order.Place(form)
	if err2 != nil {
		appG.Error(400, "订单确认失败", err2)
		return
	}
	appG.Success("", orderPayment)
}

func Settle(c *gin.Context) {
	appG := app.Gin{Cxt: c}

	var form models.SettleForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if err != nil {
		appG.Response(200, 1000, err)
		return
	}
	order, _ := models.GetById(form.OrderID)

	o, err2 := order.Settle(form)
	if err2 != nil {
		appG.Error(400, "订单结算失败", err2)
		return
	}

	// err3 := o.OnSettled()
	// if err3 != nil {
	// 	appG.Error(400, "结算后库存更新失败，需要修复", err)
	// 	return
	// }
	appG.Success("", o)
}
func SettleUndo(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	orderid, _ := strconv.Atoi(c.Params.ByName("id"))
	order, _ := models.GetById(uint(orderid))

	err2 := order.SettleUndo()
	if err2 != nil {
		appG.Error(400, "订单结算撤销失败", err2)
		return
	}
	appG.Success("", nil)
}

func AfterSettle(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	orderid, _ := strconv.Atoi(c.Params.ByName("id"))
	order, _ := models.GetById(uint(orderid))
	err := order.OnSettled()
	if err != nil {
		log.Println("结算后更新库存失败", err)
		appG.Error(400, err.Error(), nil)
		return
	}
	appG.Success("", "success")
}

func GetRefunds(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	orderid, _ := strconv.Atoi(c.Params.ByName("id"))
	order, _ := models.GetById(uint(orderid))
	refunds := order.GetRefunds()
	appG.Success("", refunds)
}

func Refund(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	var form models.RefundForm
	if err := c.ShouldBindJSON(&form); err != nil {
		appG.Error(400, "绑定请求中的json到form失败，可能是格式问题", err)
		return
	}
	err := form.Validate()
	if err != nil {
		appG.Response(200, 1000, err)
		return
	}
	order, _ := models.GetById(form.OrderID)
	if order == nil {
		appG.Error(400, "订单不存在", nil)
		return
	}
	refund, err2 := order.Refund(form)
	if err2 != nil {
		appG.Error(400, "订单退款失败", err2)
		return
	}
	appG.Success("", refund)
}
func Remove(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	orderid, _ := strconv.Atoi(c.Params.ByName("id"))
	o, _ := models.GetById(uint(orderid))
	if o.Status != order.OrderStatusCreated {
		appG.Error(400, "订单状态不正确", nil)
		return
	}
	err2 := o.Remove()
	if err2 != nil {
		appG.Error(400, "订单删除失败", err2)
		return
	}
	appG.Success("", nil)
}

func GetPriceLevelDiscount(c *gin.Context) {
	appG := app.Gin{Cxt: c}
	orderid, _ := strconv.Atoi(c.Query("orderid"))
	levelid, _ := strconv.Atoi(c.Query("levelid"))
	order, err := models.GetById(uint(orderid))
	if err != nil {
		appG.Error(400, "获取订单失败", err)
		return
	}

	priceLevelDiscount, err := order.GetPriceLevelDiscount(uint(levelid))
	if err != nil {
		appG.Error(400, "获取价格等级折扣失败", err)
		return
	}
	appG.Success("", priceLevelDiscount)
}
