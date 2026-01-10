package bean

import (
	"time"
)

func NewBean() (*Bean, *BeanBatch, error) {
	b := Bean{
		Name:    "test coffee bean" + time.Now().Format("2006-01-02 15:04"),
		Code:    "code",
		Country: "china"}
	Update(&b)

	batch, err := UpdateBatch(&BeanBatch{
		BeanID:     b.ID,
		Type:       StockLogTypePurchase,
		BeanType:   BeanTypeGreen,
		TotalQty:   100,
		TotalPrice: 100,
	})
	if err != nil {
		return nil, nil, err
	}
	return &b, batch, nil
}
