package order

import (
	"hahacking/common/logger"
	"hahacking/config"
	"hahacking/database"
	"hahacking/models/bean"
	"hahacking/models/price"
	"hahacking/models/shop"
	"hahacking/utils"
	"strconv"
	"testing"
	"time"

	"sync"

	"github.com/stretchr/testify/assert"
	gormlogger "gorm.io/gorm/logger"
)

var once sync.Once

func setupTestDB() {
	once.Do(func() {
		config.Setup("../../", "test")
		database.SetLogLevel(gormlogger.Silent)
		database.Setup()
		logger.Setup()
		db := database.Conn
		db.AutoMigrate(&Order{}, &OrderItem{}, &OrderPayment{}, &OrderPaymentItem{},
			&shop.ShopAccount{}, &shop.ShopAccountLog{},
			&Refund{}, &RefundItem{}, &RefundAmt{},
			&bean.Bean{}, &bean.BeanBatch{}, &bean.BeanBatchLog{}, &bean.BeanStockLog{},
		)
	})
}

func TestGetById(t *testing.T) {
	setupTestDB()

	// 创建测试数据
	b, _ := bean.Update(&bean.Bean{Name: "Test Bean"})
	bb, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeRoasted})
	form := UpdateForm{
		Date:     time.Now().Format("2006-01-02"),
		TotalAmt: 100.0,
		TotalQty: 10,
		Items: []OrderItem{
			{BeanID: b.ID, BeanBatchID: bb.ID, BeanName: b.Name, Qty: 5, Price: 10.0, Amt: 50.0},
		},
	}
	o, err := Update(&form)
	assert.NoError(t, err)

	order, err := GetById(o.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.SN, o.SN)
	assert.Equal(t, len(order.OrderItems), len(o.OrderItems))

	order, err = GetBySn(o.SN)
	assert.NoError(t, err)
	assert.Equal(t, order.SN, o.SN)
	assert.Equal(t, len(order.OrderItems), len(o.OrderItems))
}

func TestSearch(t *testing.T) {
	setupTestDB()
	b, _ := bean.Update(&bean.Bean{Name: "Test Bean"})
	bb, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeRoasted})

	form := UpdateForm{
		Date:     "2022-01-01",
		TotalAmt: 100.0,
		TotalQty: 10,
		Items: []OrderItem{
			{BeanID: b.ID, BeanBatchID: bb.ID, BeanName: b.Name, Qty: 5, Price: 10.0, Amt: 50.0},
		},
	}
	o, err := Update(&form)
	assert.NoError(t, err)
	// 测试 Search
	q := SearchForm{
		Status: string(OrderStatusCreated),
		Date:   "2022-01-01",
		Pager: utils.Pager{
			PageNo:   1,
			PageSize: 1,
			SortBy:   "id",
			SortDesc: "desc",
		},
	}

	orders, count := Search(q)
	assert.LessOrEqual(t, 1, count)
	assert.Equal(t, o.ID, orders[0].ID)
}

func TestPlace(t *testing.T) {
	setupTestDB()
	b, _ := bean.Update(&bean.Bean{Name: "Test Bean"})
	bb, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeRoasted})

	form := UpdateForm{
		Date:     "2022-01-01",
		TotalAmt: 80.0,
		TotalQty: 10,
		Items: []OrderItem{
			{BeanID: b.ID, BeanBatchID: bb.ID, BeanName: b.Name, Qty: 8, Price: 10.0, Amt: 80.0},
		},
	}
	o, err := Update(&form)
	assert.NoError(t, err)

	// 测试 Place
	form2 := PlaceForm{
		OrderID: o.ID,
		PaymentItems: []OrderPaymentItem{
			{Amt: 50.0, Type: OrderPaymentTypeCash, Ids: ""},
			{Amt: 30.0, Type: OrderPaymentTypeShopAccount, Ids: ""},
		},
	}
	payment, err := o.Place(form2)
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusPlaced, o.Status)

	pitems, err := payment.GetItems()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(pitems))

	//settle by paymentid - 这里用已有的payment结算
	form3 := SettleForm{
		OrderID:        o.ID,
		OrderPaymentID: payment.ID,
	}
	o.Settle(form3)
	assert.Equal(t, OrderStatusSettled, o.Status)
}

func TestSettle(t *testing.T) {
	setupTestDB()
	b, _ := bean.Update(&bean.Bean{Name: "Test Bean", PurchasePrice: 100, CPrice: 150, BPrice: 120})
	bb, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeRoasted})
	b2, _ := bean.Update(&bean.Bean{Name: "test bean 2", PurchasePrice: 100, CPrice: 150, BPrice: 120})
	bb2, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b2.ID, TotalQty: 200, BeanType: bean.BeanTypeRoasted})

	// 创建测试数据
	form := UpdateForm{
		Date:     "2022-01-01",
		TotalAmt: 680.0,
		TotalQty: 128,
		Items: []OrderItem{
			{BeanID: b.ID, BeanBatchID: bb.ID, BeanName: b.Name, Qty: 8, Price: 10.0, Amt: 80.0},
			{BeanID: b2.ID, BeanBatchID: bb2.ID, BeanName: b2.Name, Qty: 120, Price: 5.0, Amt: 600.0},
		},
	}
	o, _ := Update(&form)

	//创建一个测试account
	shopacc, _ := shop.UpdateAccount(&shop.ShopAccount{Name: "Test Account", Balance: 0.0})
	// Settle
	form2 := SettleForm{
		OrderID: o.ID,
		PaymentItems: []OrderPaymentItem{
			{Amt: 50.0, Type: OrderPaymentTypeCash, Ids: ""},
			{Amt: 630.0, Type: OrderPaymentTypeShopAccount, Ids: strconv.Itoa(int(shopacc.ID)), Memo: "Test Account"},
		},
	}
	o, err := o.Settle(form2)
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusSettled, o.Status)

	// 检查账户结算 ShopAccountLog
	shopacc, _ = shop.GetAccountById(int(shopacc.ID))
	assert.NotNil(t, shopacc)
	assert.Equal(t, -630.0, shopacc.Balance)

	//检查库存变动、批次库存变动、roastqty
	b, _ = bean.GetById(b.ID)
	assert.Equal(t, 92, b.RoastedStock)
	bb, _ = bean.GetBatchById(bb.ID)
	assert.Equal(t, 100, bb.TotalQty)
	assert.Equal(t, 92, bb.CurrentQty)
	b2, _ = bean.GetById(b2.ID)
	assert.Equal(t, 80, b2.RoastedStock)
	bb2, _ = bean.GetBatchById(bb2.ID)
	assert.Equal(t, 200, bb2.TotalQty)
	assert.Equal(t, 80, bb2.CurrentQty)

	oitems := o.Items()
	batchLogs, _ := bean.SearchBatchLog(bean.SearchBatchLogForm{RefType: "order", RefID: o.SN + "|" + strconv.Itoa(int(oitems[0].ID))})
	assert.Equal(t, 1, len(batchLogs))
	assert.Equal(t, -oitems[0].Qty, batchLogs[0].Qty)
	batchLogs, _ = bean.SearchBatchLog(bean.SearchBatchLogForm{RefType: "order", RefID: o.SN + "|" + strconv.Itoa(int(oitems[1].ID))})
	assert.Equal(t, 1, len(batchLogs))
	assert.Equal(t, -oitems[1].Qty, batchLogs[0].Qty)

	stockLogs, _ := bean.SearchStockLog(bean.SearchStockLogForm{SrcType: "order", SrcID: o.SN + "|" + strconv.Itoa(int(oitems[0].ID))})
	assert.Equal(t, 1, len(stockLogs))
	assert.Equal(t, -oitems[0].Qty, stockLogs[0].Qty)
	stockLogs, _ = bean.SearchStockLog(bean.SearchStockLogForm{SrcType: "order", SrcID: o.SN + "|" + strconv.Itoa(int(oitems[1].ID))})
	assert.Equal(t, 1, len(stockLogs))
	assert.Equal(t, -oitems[1].Qty, stockLogs[0].Qty)

	//由于Settle中已经调用，这里在调用会重复，需要报错
	err = o.OnSettled()
	assert.Error(t, err)

	//撤销结算!!!!
	o.SettleUndo()
	assert.Equal(t, OrderStatusSetttledUndo, o.Status)
	//账户恢复
	shopacc, _ = shop.GetAccountById(int(shopacc.ID))
	assert.Equal(t, 0.0, shopacc.Balance)

	//检查库存变动、批次库存变动、roastqty
	b, _ = bean.GetById(b.ID)
	assert.Equal(t, 100, b.RoastedStock)
	bb, _ = bean.GetBatchById(bb.ID)
	assert.Equal(t, 100, bb.TotalQty)
	bb2, _ = bean.GetBatchById(bb2.ID)
	assert.Equal(t, 200, bb2.TotalQty)
}

func TestGetPriceLevelDiscount(t *testing.T) {
	setupTestDB()
	b, _ := bean.Update(&bean.Bean{Name: "Test Bean", PurchasePrice: 100,
		CPrice: 150, BPrice: 120})
	bb, _ := bean.UpdateBatch(&bean.BeanBatch{BeanID: b.ID, TotalQty: 100, BeanType: bean.BeanTypeRoasted})

	// 创建测试数据
	form := UpdateForm{
		Date:     "2022-01-01",
		TotalAmt: 80.0,
		TotalQty: 8,
		Items: []OrderItem{
			{BeanID: b.ID, BeanBatchID: bb.ID, BeanName: b.Name, Qty: 8, Price: 10.0, Amt: 80.0},
		},
	}
	order, _ := Update(&form)
	// db.Create(&order)
	order, _ = GetById(order.ID)

	pricelevels, _ := price.Search(price.SearchForm{
		BeanID:    b.ID,
		LevelType: string(price.LevelTypeC),
	})
	thepricelevel := pricelevels[0]
	expecteddiscount := 0.0
	for _, item := range form.Items {
		expecteddiscount += (b.CPrice - thepricelevel.Price) * float64(item.Qty) / 1000
	}
	discountamt, err := order.GetPriceLevelDiscount(thepricelevel.PriceLevelID)
	assert.NoError(t, err)
	assert.Equal(t, expecteddiscount, discountamt) // 假设没有折扣
}
