package order

import (
	"errors"
	"fmt"
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/bean"
	"hahacking/models/shop"
	"hahacking/utils"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

type Order struct {
	models.Model
	UserID         *uint       `json:"user_id"`
	SN             string      `json:"sn" gorm:"column:sn;unique"`
	Date           string      `json:"date"`
	TotalAmt       float64     `json:"totalamt" gorm:"type:decimal(10,2);column:totalamt"` // 总金额
	PayAmt         float64     `json:"payamt" gorm:"type:decimal(10,2);column:payamt"`     // 实付金额 - 或者是计入收入账户 - 暂时不用也可以 2025-4-9
	TotalQty       int         `json:"totalqty" gorm:"column:totalqty"`                    // 总数量
	Memo           string      `json:"memo"`
	Pics           string      `json:"pics"`
	Status         OrderStatus `json:"status" `
	OrderItems     []OrderItem `json:"items" gorm:"foreignKey:OrderID"`                            // Add foreignKey tag to establish the relationship
	OrderPaymentID uint        `json:"orderpayment_id" gorm:"column:orderpayment_id"`              // 关联的支付单ID
	Refunds        []Refund    `json:"refunds" gorm:"foreignKey:OrderID"`                          // 退款
	RefundAmount   float64     `json:"refundamount" gorm:"type:decimal(10,2);column:refundamount"` // 退款金额
}
type OrderStatus string

const (
	OrderStatusCreated      OrderStatus = "created"     // 已创建
	OrderStatusPlaced       OrderStatus = "placed"      // 已下单
	OrderStatusSettled      OrderStatus = "settled"     // 已结算
	OrderStatusInvalid      OrderStatus = "invalid"     // 无效
	OrderStatusSetttledUndo OrderStatus = "settledundo" //已结算撤销
)

var OrderStatuses = map[OrderStatus]string{
	OrderStatusCreated:      "已创建",
	OrderStatusPlaced:       "已下单",
	OrderStatusSettled:      "已结算",
	OrderStatusInvalid:      "无效",
	OrderStatusSetttledUndo: "结算撤销",
}

func Meta(name string) map[string]any {
	names := strings.Split(name, ",")
	result := make(map[string]any)
	for _, n := range names {
		if n == "OrderStatus" {
			result[n] = models.ConvertMapToArray(OrderStatuses)
		}
	}
	return result
}

func GetById(id uint) (*Order, error) {
	var o Order
	db := database.Conn
	err := db.Where("id = ?", id).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

func GetBySn(sn string) (*Order, error) {
	var o Order
	db := database.Conn
	err := db.Where("sn = ?", sn).First(&o).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &o, nil
}

type SearchForm struct {
	utils.Pager
	UserID   uint    `form:"user_id"`
	SN       string  `form:"sn"`
	Date     string  `form:"date"`
	TotalAmt float64 `form:"totalamt"`
	TotalQty int     `form:"totalqty"`
	Status   string  `form:"status"`
}

func Search(q SearchForm) ([]Order, int) {
	db := database.Conn
	var os []Order
	var statement = db.Model(&Order{})
	if q.Status != "" {
		statement = statement.Where("status = ?", q.Status)
	}
	if q.SN != "" {
		statement = statement.Where("sn = ?", q.SN)
	}
	if q.UserID != 0 {
		statement = statement.Where("user_id = ?", q.UserID)
	}
	if q.Date != "" {
		statement = statement.Where("date like ?", "%"+q.Date+"%")
	}
	_, pagesize, start, sort := q.Pager.Paging()
	var count int64
	statement.Where(&Order{}).Preload("OrderItems").
		Count(&count).
		Order(sort).
		Offset(start).
		Limit(pagesize).
		Find(&os)
	return os, int(count)
}

type CreateForm struct {
	Date     string      `json:"date"`
	TotalAmt float64     `json:"totalamt"`
	TotalQty int         `json:"totalqty"`
	Status   string      `json:"status"`
	Items    []OrderItem `json:"items"`
	Memo     string      `json:"memo"`
}

func (form *CreateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if len(form.Items) == 0 {
		result["items"] = "订单明细不能为空"
	}
	if form.Date == "" {
		result["date"] = "日期不能为空"
	}
	if form.TotalAmt <= 0 {
		result["totalamt"] = "总金额不能为空"
	}
	if form.TotalQty <= 0 {
		result["totalqty"] = "总数量不能为空"
	}
	return nil
}

type UpdateForm struct {
	ID       uint        `json:"id" form:"id"`
	Date     string      `json:"date"`
	TotalAmt float64     `json:"totalamt"`
	TotalQty int         `json:"totalqty"`
	Status   string      `json:"status"`
	Items    []OrderItem `json:"items"`
	Memo     string      `json:"memo"`
}

func (form *UpdateForm) Validate() map[string]string {
	var result = make(map[string]string)
	if len(form.Items) == 0 {
		result["items"] = "订单明细不能为空"
	}
	if form.Date == "" {
		result["date"] = "日期不能为空"
	}
	if form.TotalAmt <= 0 {
		result["totalamt"] = "总金额不能为空"
	}
	if form.TotalQty <= 0 {
		result["totalqty"] = "总数量不能为空"
	}
	return nil
}

func Update(f *UpdateForm) (*Order, error) {
	db := database.Conn
	var o Order
	if err := mapstructure.Decode(f, &o); err != nil {
		log.Println("Decode order err", err)
		return nil, err
	}
	if f.ID != 0 { //修改
		if err := db.Model(&Order{}).Where("id = ?", f.ID).Updates(&o).Error; err != nil {
			log.Println("Update order err", err)
			return nil, err
		}
		var existingItems []OrderItem
		if err := db.Where("order_id = ?", f.ID).Find(&existingItems).Error; err != nil {
			log.Println("查询订单明细失败", err)
			return nil, err
		}
		// Delete items that are not in the new list
		var itemsToDelete []OrderItem
		for _, existingItem := range existingItems {
			found := false
			for _, item := range f.Items {
				if existingItem.ID == item.ID {
					found = true
					break
				}
			}
			if !found {
				itemsToDelete = append(itemsToDelete, existingItem)
			}
		}
		tx := db.Begin()
		for _, item := range itemsToDelete {
			if err := tx.Where("id = ?", item.ID).Delete(&OrderItem{}).Error; err != nil {
				log.Println("Delete order item err", err)
				tx.Rollback()
				return nil, err
			}
		}
		for _, item := range f.Items {
			if item.ID != 0 {
				if err := tx.Model(&OrderItem{}).Where("id = ?", item.ID).Updates(&item).Error; err != nil {
					log.Println("Update order item err", err)
					tx.Rollback()
					return nil, err
				}
			} else {
				var oitem OrderItem
				if err := mapstructure.Decode(item, &oitem); err != nil {
					log.Println("Decode order item err", err)
					tx.Rollback()
					return nil, err
				}
				oitem.OrderID = f.ID
				if err := tx.Create(&oitem).Error; err != nil {
					log.Println("Create order item err", err)
					tx.Rollback()
					return nil, err
				}
			}
		}
		tx.Commit()
	} else {
		tx := db.Begin()
		o.Status = "created"
		if err := tx.Create(&o).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for _, item := range f.Items {
			item.OrderID = o.ID
			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
		tx.Commit()
	}
	return &o, nil
}
func (o *Order) AfterCreate(tx *gorm.DB) (err error) {
	// 生成流水号：前缀 + 时间戳 + 自增ID
	prefix := ""
	timestamp := time.Now().Format("20060102")
	serialNumber := fmt.Sprintf("%s%s%06d", prefix, timestamp, o.ID)
	if err := tx.Model(&o).Update("SN", serialNumber).Error; err != nil {
		tx.Rollback()
		return err
	}
	return
}

type PlaceForm struct {
	UserID       uint               `json:"user_id"`
	OrderID      uint               `json:"order_id"`
	PaymentItems []OrderPaymentItem `json:"payitems,omitempty"`
}

func (form *PlaceForm) Validate() map[string]string {
	if len(form.PaymentItems) == 0 {
		return map[string]string{"payitems": "结算单明细不能为空"}
	}
	if form.OrderID == 0 {
		return map[string]string{"order_id": "订单ID不能为空"}
	}
	return nil
}

/**
* 订单支付项确认 - 计算实付金额
* 如果支付项校验通过：则创建并关联Order和OrderPayment，之后让用户确认后Settle
* 如果校验未通过：则给用户提示
 */
func (order *Order) Place(form PlaceForm) (*OrderPayment, error) {
	db := database.Conn
	var err error
	//保存payment并计算剩余应付金额
	var totalPayAmt float64
	orderPayment := OrderPayment{
		OrderID: form.OrderID,
	}
	if err := db.Create(&orderPayment).Error; err != nil {
		// tx.Rollback()
		return nil, err
	}
	for _, paymentItem := range form.PaymentItems {
		paymentItem.OrderPaymentID = orderPayment.ID
		if err = db.Create(&paymentItem).Error; err != nil {
			// tx.Rollback()
			return nil, err
		}
		totalPayAmt += paymentItem.Amt
	}
	order.PayAmt = order.TotalAmt - totalPayAmt
	order.Status = "placed"
	order.OrderPaymentID = orderPayment.ID
	if err := db.Save(&order).Error; err != nil {
		// tx.Rollback()
		return nil, err
	}
	// tx.Commit()
	// Attach payment items to the payment object
	var paymentItems []OrderPaymentItem
	if err := db.Where("orderpayment_id = ?", orderPayment.ID).Find(&paymentItems).Error; err != nil {
		return nil, err
	}
	orderPayment.Items = paymentItems
	return &orderPayment, nil
}

type SettleForm struct {
	UserID  uint `json:"user_id,omitempty"`
	OrderID uint `json:"order_id,omitempty"`
	//模式1:支付单ID - 和支付单明细同时存在时，优先选择订单支付单ID
	OrderPaymentID uint `json:"orderpayment_id,omitempty"`
	//模式2:支付单明细（按需创建支付单）
	PaymentItems []OrderPaymentItem `json:"payitems,omitempty"`
}

func (form *SettleForm) Validate() map[string]string {
	if form.OrderID == 0 {
		return map[string]string{"order_id": "订单ID不能为空"}
	}
	if form.OrderPaymentID == 0 {
		if len(form.PaymentItems) == 0 {
			return map[string]string{"payitems": "结算单明细不能为空"}
		}
	}
	return nil
}

/*
*  结算订单
* case1: 传入OrderPaymentID，根据paymentID结算 （需要校验支付单是否还有效可用过期）
* case2: 传入payitems数组，此时需要重新调用place进行校验（可能通过可能不通过）
 */
func (order *Order) Settle(form SettleForm) (*Order, error) {
	db := database.GetConn()
	log.Println("开始结算订单", order.ID)
	log.Printf("结算单 %v", form)
	// tx := db.Begin()
	order.Status = OrderStatusSettled
	if err := db.Save(&order).Error; err != nil {
		// tx.Rollback()
		return nil, err
	}
	var payment *OrderPayment
	var err error
	//校验OrderPayment状态、合并的金额是否正确
	if form.OrderPaymentID == 0 { //如果没有paymentid，则检查payitems是否正确
		payment, err = order.Place(PlaceForm{
			UserID:       form.UserID,
			OrderID:      form.OrderID,
			PaymentItems: form.PaymentItems,
		})
		if err != nil {
			log.Println("支付单校验失败", err)
			return nil, err
		}
	} else {
		payment, err = GetPaymentById(form.OrderPaymentID)
		if err != nil {
			log.Println("获取支付单失败", err)
			return nil, err
		}
	}

	payment.Status = models.StatusValid
	db.Save(payment)
	log.Printf("支付单 %v", payment)
	//修改payment的状态和items的状态
	pitems := payment.Items
	log.Printf("支付单明细 %v", pitems)
	for _, item := range pitems {
		// item.Status = models.StatusValid
		// if err := db.Model(&item).Updates(map[string]any{"status": item.Status}).Error; err != nil {
		// 	log.Println("更新支付单明细状态失败", err)
		// 	// tx.Rollback()
		// 	return nil, err
		// }
		switch item.Type {
		case OrderPaymentTypeShopAccount: //账户结算（包括现金）
			log.Println("账户结算", item.Ids, item.Amt)
			accid, _ := strconv.Atoi(item.Ids)
			acc, _ := shop.GetAccountById(accid)
			acc.Balancing(-item.Amt, "OrderPayment", order.SN+"|"+strconv.Itoa(int(payment.ID)))
		case OrderPaymentTypeCash: //现金结算 - 暂时兼容，后续现金以账户处理
			// accid, _ := strconv.Atoi(item.Ids)
			// acc, _ := shop.GetAccountById(accid)
			// acc.Balancing(-item.Amt, "OrderPayment", order.SN+"|"+strconv.Itoa(int(payment.ID)))
		case OrderPaymentTypeAlipay:
			//TODO: 支付宝结算
		case OrderPaymentTypeWeChat:
			//TODO: 微信结算
		case OrderPaymentTypeBank:
			//TODO: 银行结算
		case OrderPaymentTypeCredit:
			//TODO: 积分结算
		case OrderPaymentTypeGiftCard:
			//TODO: 礼品卡结算
		case OrderPaymentTypeOther:
			//TODO: 其他结算
		}

	}
	order.Status = OrderStatusSettled
	if err := db.Save(&order).Error; err != nil {
		// tx.Rollback()
		return nil, err
	}
	//熟豆库存更新（批次剩余量、stocklog，bean.roastqty）
	err = order.OnSettled()
	if err != nil {
		log.Println("结算后库存更新失败", err)
		return nil, err
	}

	return order, nil
}
func (order *Order) OnSettled() error {
	items := order.Items()
	var err error
	for _, item := range items {
		//批次日志需要一个一个写入，库存日志可以将相同的豆子（豆子id、生/熟）数量合并
		b, _ := bean.GetById(item.BeanID)
		_, err = b.Stocking(-item.Qty, bean.BeanTypeRoasted, bean.StockLogTypeSale, "order", order.SN+"|"+strconv.FormatUint(uint64(item.ID), 10),
			"订单"+order.SN+"|"+strconv.FormatUint(uint64(item.ID), 10))
		if err != nil {
			errmsg := "库存更新失败 -" + err.Error() + " 订单" + order.SN + " 豆子" + b.Name + " 数量" + strconv.Itoa(item.Qty) +
				"refid:" + order.SN + "|" + strconv.FormatUint(uint64(item.ID), 10)
			log.Println(errmsg)
			err = errors.New(errmsg)
			return err
		}
		batch, _ := bean.GetBatchById(item.BeanBatchID)
		// batch.CurrentQty -= item.Qty 移入log方法
		// database.Conn.Save(batch)
		err = batch.Log(-item.Qty, "order", order.SN+"|"+strconv.FormatUint(uint64(item.ID), 10),
			"订单结算"+order.SN+"|"+strconv.FormatUint(uint64(item.ID), 10))
		if err != nil {
			errmsg := "批次库存更新失败 - " + err.Error() + " 订单" + order.SN + " 批次" + strconv.Itoa(int(batch.ID)) + " 数量" + strconv.Itoa(item.Qty) +
				"refid:" + order.SN + "|" + strconv.FormatUint(uint64(item.ID), 10)
			log.Println(errmsg)
			err = errors.New(errmsg)
			return err
		}
	}
	return nil
}
func (order *Order) Items() []OrderItem {
	db := database.Conn
	var items []OrderItem
	db.Where("order_id = ?", order.ID).Find(&items)
	return items
}
func (order *Order) Payment() OrderPayment {
	db := database.Conn
	var payment OrderPayment
	db.Where("order_id = ? and status = ?", order.ID, models.StatusValid).Preload("Items").Last(&payment)
	return payment
}

func (order *Order) SettleUndo() error {
	db := database.Conn
	tx := db.Begin()

	if err := tx.Model(&order).Update("status", OrderStatusSetttledUndo).Error; err != nil {
		return err
	}
	//1. 撤销支付单 状态, 撤销账户记录
	payment, err := GetPaymentById(order.OrderPaymentID)
	if err != nil {
		return err
	}
	payment.Status = models.StatusInvalid
	tx.Save(payment)
	//恢复结算（如结算账户扣款、退款、券等）

	pitems, _ := payment.GetItems()
	// fmt.Println("pitemsxxxxxxxxxxxx", pitems)
	for _, item := range pitems {
		if err := tx.Model(&item).Updates(map[string]any{"status": OrderStatusInvalid}).Error; err != nil {
			log.Println("更新支付单明细状态失败", err)
			// tx.Rollback()
			return err
		}
		// fmt.Println("itemxxxxxxxxxxxx", item)
		if item.Type == OrderPaymentTypeShopAccount { //账户结算（包括现金）
			accid, _ := strconv.Atoi(item.Ids)
			acc, _ := shop.GetAccountById(accid)
			acc.Balancing(item.Amt, "OrderPayment", strconv.Itoa(int(payment.ID)))
		}
		//如果是用券， 需要记录log Later
	}

	//恢复库存
	items := order.Items()
	for _, item := range items {
		b, _ := bean.GetById(item.BeanID)
		b.Stocking(item.Qty, bean.BeanTypeRoasted, bean.StockLogTypeUndo, "order", order.SN, "订单取消"+order.SN)
		batch, _ := bean.GetBatchById(item.BeanBatchID)
		// batch.CurrentQty += item.Qty 移入log
		// database.Conn.Save(batch)
		batch.Log(item.Qty, "order", strconv.FormatUint(uint64(order.ID), 10), "订单取消"+order.SN+"|"+strconv.FormatUint(uint64(item.ID), 10))
	}

	tx.Commit()

	return nil
}
func (order *Order) Remove() error {
	db := database.Conn
	if err := db.Model(&order).Update("status", OrderStatusInvalid).Error; err != nil {
		return err
	}
	return nil
}

// 获取订单的折扣金额(子项中每个bean都采用对应的等级价格的和)
func (order Order) GetPriceLevelDiscount(levelid uint) (float64, error) {
	db := database.Conn
	var orderitems []OrderItem
	err := db.Where("order_id = ?", order.ID).Find(&orderitems).Error
	if err != nil {
		return 0, err
	}
	var beanids []uint
	for _, item := range orderitems {
		beanids = append(beanids, item.BeanID)
	}
	prices, _ := bean.GetPrices(beanids, levelid)
	var totaldiscount float64
	for _, item := range orderitems {
		beanlevelprice := prices[item.BeanID]
		bean, _ := bean.GetById(item.BeanID)
		//一律以零售基础价为折扣基础
		totaldiscount += (bean.CPrice - beanlevelprice) * float64(item.Qty) / 1000

	}
	return totaldiscount, nil
}

type RefundForm struct {
	OrderID  uint         `json:"order_id"`
	Date     string       `json:"date"`
	TotalQty int          `json:"totalqty"`
	TotalAmt float64      `json:"totalamt"`
	Items    []RefundItem `json:"items"`
	Amts     []RefundAmt  `json:"amts"`
}

func (refundform *RefundForm) Validate() map[string]string {
	if refundform.OrderID == 0 {
		return map[string]string{"order_id": "订单ID不能为空"}
	}
	if refundform.Date == "" {
		return map[string]string{"date": "日期不能为空"}
	}
	if refundform.TotalQty <= 0 && refundform.TotalAmt <= 0 {
		return map[string]string{"totalqty": "总数量和总金额不能同时为空"}
	}
	if len(refundform.Items) == 0 && len(refundform.Amts) == 0 {
		return map[string]string{"items": "退货或者退款项目不能同时为空"}
	}
	for _, item := range refundform.Items {
		if item.OrderItemID == 0 {
			return map[string]string{"items": "订单项目ID不能为空"}
		}
	}
	err := RefundRuleCheck(*refundform)
	if err != nil {
		return map[string]string{"items": err.Error()}
	}
	return nil
}

func (order *Order) Refund(form RefundForm) (*Refund, error) {
	db := database.Conn
	//先检查够不够退 -- 移入form.Validate()
	err := RefundRuleCheck(form)
	if err != nil {
		return nil, err
	}
	refund := Refund{
		OrderID:  form.OrderID,
		Date:     form.Date,
		TotalQty: form.TotalQty,
		TotalAmt: form.TotalAmt,
		Status:   models.StatusValid,
	}
	tx := db.Begin()
	if err := tx.Create(&refund).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range form.Items {
		if err := tx.Create(&RefundItem{
			RefundID:    refund.ID,
			OrderItemID: item.OrderItemID,
			Qty:         item.Qty, //克
			// Price:       item.Price, //价格是公斤价格
			// Amt:         float64(item.Qty) * item.Price / 1000,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	for _, amt := range form.Amts {
		if err := tx.Create(&RefundAmt{
			RefundID:           refund.ID,
			OrderPaymentItemID: amt.OrderPaymentItemID,
			Amt:                amt.Amt,
		}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()
	return &refund, nil
}

func (order *Order) GetRefunds() []Refund {
	db := database.Conn
	var refunds []Refund
	db.Where("order_id = ?", order.ID).Preload("Items").Preload("Amts").Find(&refunds)
	return refunds
}
