package order

import (
	"hahacking/models/bean"
	"hahacking/models/shop"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRefund(t *testing.T) {
	setupTestDB()
	o := newTestRefundOrder()
	if o.Status != OrderStatusSettled {
		t.Errorf("测试订单状态不正确，状态为：%s", o.Status)
	}
	opayment, _ := GetPaymentById(o.OrderPaymentID)
	opitems, _ := opayment.GetItems()
	assert.Equal(t, opitems[0].Amt, 5.0)
	assert.Equal(t, opitems[1].Amt, 50.0)
	cashaccid, _ := strconv.Atoi(opitems[0].Ids)
	shopacc1id, _ := strconv.Atoi(opitems[1].Ids)

	cashacc, _ := shop.GetAccountById(cashaccid)
	shopacc1, _ := shop.GetAccountById(shopacc1id)

	assert.Equal(t, float64(195), cashacc.Balance)
	assert.Equal(t, float64(400), shopacc1.Balance)
	oitems := o.Items()
	refund, err := o.Refund(RefundForm{
		OrderID:  o.ID,
		Date:     time.Now().Format("2006-01-02"),
		TotalQty: 110,
		TotalAmt: 55.0,
		Items: []RefundItem{
			{OrderItemID: oitems[0].ID, Qty: 10},
			{OrderItemID: oitems[1].ID, Qty: 100},
		},
		Amts: []RefundAmt{
			{OrderPaymentItemID: opitems[0].ID, Amt: 5.0},
			{OrderPaymentItemID: opitems[1].ID, Amt: 50.0},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, refund.OrderID, o.ID)
	assert.Equal(t, refund.TotalQty, 110)
	assert.Equal(t, refund.TotalAmt, 55.0)
	refunditems := refund.GetItems()
	assert.Equal(t, len(refunditems), 2)
	assert.Equal(t, refunditems[0].OrderItemID, oitems[0].ID)
	assert.Equal(t, refunditems[0].Qty, 10)
	assert.Equal(t, refunditems[1].OrderItemID, oitems[1].ID)
	assert.Equal(t, refunditems[1].Qty, 100)

	//检查库存是否恢复
	err = refund.StockUndo()
	assert.NoError(t, err)
	b1, _ := bean.GetById(oitems[0].BeanID)
	b2, _ := bean.GetById(oitems[1].BeanID)
	assert.Equal(t, 200, b2.RoastedStock) //100+100
	assert.Equal(t, 60, b1.RoastedStock)  //50+10
	//TODO 检查批次是否回复

	//检查账户是否恢复
	// refund.AmtUndo()
	// shopacc1, _ = shop.GetAccountById(shopacc1id)
	// cashacc, _ = shop.GetAccountById(cashaccid)
	// assert.Equal(t, 450.0, shopacc1.Balance)
	// assert.Equal(t, 200.0, cashacc.Balance)

}
func newTestRefundOrder() *Order {
	setupTestDB()
	b, _ := bean.Update(&bean.Bean{Name: "Test Bean"})
	bb, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeRoasted})
	b2, _ := bean.Update(&bean.Bean{Name: "Test Bean 2"})
	bb2, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b2.ID, TotalQty: 200, BeanType: bean.BeanTypeRoasted})
	form := UpdateForm{
		Date:     time.Now().Format("2006-01-02"),
		TotalAmt: 100.0,
		TotalQty: 10,
		Items: []OrderItem{
			{BeanID: b.ID, BeanBatchID: bb.ID, BeanName: b.Name, Qty: 50, Price: 100.0, Amt: 5.0},
			{BeanID: b2.ID, BeanBatchID: bb2.ID, BeanName: b2.Name, Qty: 100, Price: 500.0, Amt: 50.0},
		},
	}
	o, _ := Update(&form)

	cashacc, _ := shop.UpdateAccount(&shop.ShopAccount{Name: "Test Cash Account", Balance: 200.0})
	shopacc1, _ := shop.UpdateAccount(&shop.ShopAccount{Name: "Test Account", Balance: 450.0})
	payment, _ := o.Place(PlaceForm{
		OrderID: o.ID,
		PaymentItems: []OrderPaymentItem{
			{Amt: 5.0, Type: OrderPaymentTypeShopAccount, Ids: strconv.Itoa(int(cashacc.ID))},
			{Amt: 50.0, Type: OrderPaymentTypeShopAccount, Ids: strconv.Itoa(int(shopacc1.ID))},
		},
	})
	o.Settle(SettleForm{
		OrderID:        o.ID,
		OrderPaymentID: payment.ID,
	})
	return o
}
