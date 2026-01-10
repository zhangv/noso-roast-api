package roast

import (
	"hahacking/models/bean"
	"log"
)

func NewRoast(b *bean.Bean, batch *bean.BeanBatch) (*Roast, *bean.BeanBatch, error) {
	items := []RoastItem{
		{
			BeanBatchID: batch.ID,
			Qty:         50,
		},
	}
	startForm := StartForm{
		Name:   "test coffee roast",
		BeanID: b.ID,
		InQty:  50,
		Items:  items,
	}
	roasto, err := Start(startForm)
	if err != nil {
		log.Printf("烘焙开始失败: %v", err)
		return nil, nil, err
	}

	roasto, err = GetById(roasto.ID)
	if err != nil {
		log.Printf("获取烘焙记录失败: %v", err)
		return nil, nil, err
	}

	completeForm := CompleteForm{
		ID:     roasto.ID,
		OutQty: 40,
	}
	err = roasto.Complete(completeForm)
	if err != nil {
		log.Printf("完成烘焙失败: %v", err)
		return nil, nil, err
	}

	roastedBatch, _ := bean.GetBatchById(roasto.BeanBatchID)
	return roasto, roastedBatch, nil
}
