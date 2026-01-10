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
	"time"

	"gorm.io/gorm"
)

type Refund struct {
	models.Model
	SN       string        `json:"sn" gorm:"column:sn;unique"` //退款流水号
	Date     string        `json:"date"`                       //退款日期
	OrderID  uint          `json:"order_id"`
	Order    *Order        `json:"order" gorm:"foreignKey:OrderID;references:ID"`
	Status   models.Status `json:"status"`
	TotalAmt float64       `json:"totalprice" gorm:"column:totalamt;type:decimal(10,2)"`
	TotalQty int           `json:"totalqty" gorm:"column:totalqty;default:0;not null;type:int"`
	Items    []RefundItem  `json:"items" gorm:"foreignKey:RefundID"`
	Amts     []RefundAmt   `json:"amts" gorm:"foreignKey:RefundID"`
}

// 记录退款商品（豆子）
type RefundItem struct {
	models.Model
	RefundID    uint    `json:"refund_id" gorm:"column:refund_id"`
	Refund      *Refund `json:"refund" gorm:"foreignKey:RefundID;references:ID"`
	OrderItemID uint    `json:"order_item_id" gorm:"column:order_item_id"`
	// OrderItem   *OrderItem      `json:"order_item" gorm:"foreignKey:OrderItemID;references:ID"`
	//这些豆子信息似乎没用了
	// BeanID      uint            `json:"bean_id" gorm:"column:bean_id"`
	// BeanName    string          `json:"bean_name"`
	// BeanBatchID uint            `json:"beanbatch_id" gorm:"column:beanbatch_id"`
	// BeanBatch   *bean.BeanBatch `json:"beanbatch" gorm:"foreignKey:BeanBatchID;references:ID"`
	// Bean        *bean.Bean      `json:"bean" gorm:"foreignKey:BeanID;references:ID"`
	// BeanType    bean.BeanType   `json:"bean_type" gorm:"column:bean_type"`
	Qty int `json:"qty" gorm:"column:qty"`
	// Price       float64         `json:"price" gorm:"column:price;type:decimal(10,2)"`
	// Amt         float64         `json:"amt" gorm:"column:amt;type:decimal(10,2)"`
	Status models.Status `json:"status" gorm:"column:status"`
}

// 记录退款金额和渠道来源
type RefundAmt struct {
	models.Model
	RefundID uint    `json:"refund_id" gorm:"column:refund_id"`
	Refund   *Refund `json:"refund" gorm:"foreignKey:RefundID;references:ID"`
	//支付项ID
	OrderPaymentItemID uint          `json:"order_payment_item_id"`
	Amt                float64       `json:"amt"`
	Status             models.Status `json:"status" gorm:"column:status"`
}

func (refund *Refund) BeforeCreate(tx *gorm.DB) (err error) {
	// 生成流水号：前缀 + 时间戳 + 纳秒级时间戳
	prefix := "RF"
	timestamp := time.Now().Format("20060102")
	nanos := time.Now().UnixNano() % 1000000 // 取纳秒的后6位
	serialNumber := fmt.Sprintf("%s%s%06d", prefix, timestamp, nanos)

	// 如果SN已存在，重试生成
	for i := 0; i < 10; i++ {
		var count int64
		tx.Model(&Refund{}).Where("SN = ?", serialNumber).Count(&count)
		if count == 0 {
			break
		}
		// 如果存在冲突，添加随机数
		nanos = time.Now().UnixNano() % 1000000
		serialNumber = fmt.Sprintf("%s%s%06d", prefix, timestamp, nanos)
	}

	refund.SN = serialNumber
	return nil
}

func RefundRuleCheck(form RefundForm) error {
	//1. 检查订单是否存在足够的退款金额
	order, _ := GetById(form.OrderID)
	refunds := order.GetRefunds()
	currentRefundAmount := 0.0
	for _, refund := range refunds {
		currentRefundAmount += refund.TotalAmt
	}
	if currentRefundAmount+form.TotalAmt > order.TotalAmt {
		return errors.New("订单退款金额超过订单总金额,已退金额:" + strconv.FormatFloat(currentRefundAmount, 'f', 2, 64) + ",可退金额:" + strconv.FormatFloat(order.TotalAmt-currentRefundAmount, 'f', 2, 64))
	}
	//2. 检查每个订单项目(batch)是否还有足够的退货数量
	//订单原始数量
	orderitems := order.Items()
	orderitemqty := make(map[uint]int)
	for _, item := range orderitems {
		orderitemqty[item.ID] += item.Qty
	}
	//已退总数量
	refunditemqty := make(map[uint]int)
	for _, refund := range refunds {
		refunditems := refund.GetItems()
		for _, item := range refunditems {
			refunditemqty[item.OrderItemID] += item.Qty
		}
	}
	for _, item := range form.Items {
		if orderitemqty[item.OrderItemID]-refunditemqty[item.OrderItemID] < item.Qty {
			return errors.New("订单项目" + strconv.FormatUint(uint64(item.OrderItemID), 10) +
				"退款数量超过订单项目总数量(" + strconv.FormatUint(uint64(orderitemqty[item.OrderItemID]), 10) + "),已退数量:" + strconv.Itoa(refunditemqty[item.OrderItemID]) +
				",可退数量:" + strconv.Itoa(item.Qty))
		}
	}
	//3. 检查退款
	// 先获取各个支付子项的已退合计金额
	refundpaymentamts := make(map[uint]float64)
	for _, refund := range refunds {
		refunditems := refund.GetAmts()
		for _, item := range refunditems {
			refundpaymentamts[item.OrderPaymentItemID] += item.Amt
		}
	}
	orderpayments, _ := SearchPayment(SearchPaymentForm{
		Pager: utils.Pager{
			PageNo:   1, //这里应该只搜索出一个
			PageSize: 1,
		},
		Status:  models.StatusValid,
		OrderID: form.OrderID,
	})
	if len(orderpayments) == 0 {
		return errors.New("支付项不存在")
	}
	// 再获取各个支付子项的合计金额
	totalpaymentamts := make(map[uint]float64)
	orderpayment := orderpayments[0]
	orderpaymentitems, _ := orderpayment.GetItems()
	for _, pitem := range orderpaymentitems {
		totalpaymentamts[pitem.ID] += pitem.Amt
	}
	// 检查各个支付项的已退合计金额是否超过合计金额
	for _, refundamt := range form.Amts {
		if totalpaymentamts[refundamt.OrderPaymentItemID]-refundpaymentamts[refundamt.OrderPaymentItemID] < refundamt.Amt {
			return errors.New("支付项" + strconv.FormatUint(uint64(refundamt.OrderPaymentItemID), 10) +
				"退款金额超过支付项总金额,已退金额:" + strconv.FormatFloat(refundpaymentamts[refundamt.OrderPaymentItemID], 'f', 2, 64) +
				",可退金额:" + strconv.FormatFloat(totalpaymentamts[refundamt.OrderPaymentItemID]-refundpaymentamts[refundamt.OrderPaymentItemID], 'f', 2, 64))
		}
	}

	return nil
}

func (refund *Refund) GetItems() []RefundItem {
	db := database.Conn
	var items []RefundItem
	db.Where("refund_id = ?", refund.ID).Find(&items)
	return items
}

func (refund *Refund) GetAmts() []RefundAmt {
	db := database.Conn
	var amts []RefundAmt
	db.Where("refund_id = ?", refund.ID).Find(&amts)
	return amts
}

func GetRefundBySn(sn string) (*Refund, error) {
	var refund Refund
	db := database.Conn
	err := db.Where("sn = ?", sn).First(&refund).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return &refund, nil
}

// 撤销 beanbatch、bean 库存、 account 账户记账记录
func (refund *Refund) StockUndo() error {
	db := database.Conn
	tx := db.Begin()
	refunditems := refund.GetItems()
	log.Printf("退款商品撤销 %v", refunditems)
	for _, item := range refunditems {
		log.Printf("退款商品撤销 %v, qty: %v", item.ID, item.Qty)
		orderitem, _ := GetItemById(item.OrderItemID)
		itembean, err := bean.GetById(orderitem.BeanID)
		if err != nil {
			return err
		}
		itembatch, err := bean.GetBatchById(orderitem.BeanBatchID)
		if err != nil {
			return err
		}
		itembean.Stocking(item.Qty, itembatch.BeanType, bean.StockLogTypeReturn,
			"refund", refund.SN+"|"+strconv.FormatUint(uint64(item.ID), 10),
			fmt.Sprintf("订单%v退款%v", refund.OrderID, refund.SN+"|"+strconv.FormatUint(uint64(item.ID), 10)))
		itembatch.Log(item.Qty, "refund", refund.SN+"|"+strconv.FormatUint(uint64(item.ID), 10),
			"订单退货"+refund.SN+"|"+strconv.FormatUint(uint64(item.ID), 10))
		if err := db.Model(&item).Where("id = ?", item.ID).Update("status", models.StatusInvalid).Error; err != nil {
			return err
		}
	}
	tx.Commit()
	return nil
}

// 撤销退款金额
func (refund *Refund) AmtUndo() error {
	db := database.Conn
	tx := db.Begin()
	refundamts := refund.GetAmts()
	log.Printf("退款金额撤销 %v", refundamts)
	for _, ritem := range refundamts {
		opitem, _ := GetPaymentItemById(ritem.OrderPaymentItemID)
		switch opitem.Type {
		case OrderPaymentTypeShopAccount:
			accid, _ := strconv.Atoi(opitem.Ids)
			acc, _ := shop.GetAccountById(accid)
			log.Printf("账户 %v 退款金额撤销 %v", acc.Name, ritem.Amt)
			acc.Balancing(ritem.Amt, "refund", refund.SN)
		case OrderPaymentTypeCash:
			//TODO: 现金退款 - 应该也用账户处理
		case OrderPaymentTypeAlipay:
			//TODO: 支付宝退款
		case OrderPaymentTypeWeChat:
			//TODO: 微信退款
		case OrderPaymentTypeBank:
			//TODO: 银行退款
		case OrderPaymentTypeCredit:
			//TODO: 积分退款
		case OrderPaymentTypeGiftCard:
			//TODO: 礼品卡退款
		case OrderPaymentTypeOther:
			//TODO: 其他退款
		}
		if err := db.Model(&ritem).Where("id = ?", ritem.ID).Update("status", models.StatusInvalid).Error; err != nil {
			return err
		}
	}
	tx.Commit()
	return nil

}

type SearchRefundForm struct {
	utils.Pager
	OrderID uint `form:"order_id"`
}

func SearchRefund(q SearchRefundForm) ([]Refund, int) {
	db := database.Conn
	var refunds []Refund
	var statement = db.Model(&Refund{})
	if q.OrderID != 0 {
		statement = statement.Where("order_id = ?", q.OrderID)
	}
	_, pagesize, start, sort := q.Pager.Paging()
	var count int64
	statement.Where(&Refund{}).Count(&count).Order(sort).Offset(start).Limit(pagesize).Find(&refunds)
	return refunds, len(refunds)
}
