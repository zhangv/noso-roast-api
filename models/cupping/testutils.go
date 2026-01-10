package cupping

import (
	"hahacking/database/datatypes"
	"hahacking/models/bean"
	"hahacking/models/roast"
)

func NewCupping(b *bean.Bean, roastedBatch *bean.BeanBatch, ro *roast.Roast) (*Cupping, error) {
	cp := Cupping{
		Description: "test cupping",
		Location:    "test cuping location",
		Status:      CuppingStatusValid,
		Method:      CuppingMethodStandard,
		Samples: &[]CuppingSample{
			{
				BeanID:          b.ID,
				BeanBatchID:     roastedBatch.ID,
				RoastID:         ro.ID,
				BeanQty:         10,
				ExtractionYield: "0.5",
				GrindSize:       "1",
				WaterTemp:       20,
				Method:          CuppingMethodStandard,
			},
		},
	}
	err := Update(&cp)
	if err != nil {
		return nil, err
	}
	return &cp, nil
}
func NewCuppingSample(cp *Cupping, b *bean.Bean, roastedBatch *bean.BeanBatch, roasto *roast.Roast) (*CuppingSample, error) {
	sample := CuppingSample{
		CuppingID:       cp.ID,
		BeanID:          b.ID,
		BeanBatchID:     roastedBatch.ID,
		RoastID:         roasto.ID,
		BeanQty:         10,
		ExtractionYield: "0.5",
		GrindSize:       "1",
		WaterTemp:       20,
		Method:          CuppingMethodStandard,
	}
	err := UpdateSample(&sample)
	if err != nil {
		return nil, err
	}
	return &sample, nil
}
func NewCuppingScore(cp *Cupping, sample *CuppingSample, overrideScore *CuppingScore) (*CuppingScore, error) {
	jsonDefects := Defects{
		{
			Name:  "霉味",
			Level: "full",
			Desc:  "霉味",
		},
	}
	var cuppingScore CuppingScore
	if overrideScore != nil {
		cuppingScore = *overrideScore
	} else {
		cuppingScore = CuppingScore{
			CuppingID:       cp.ID,
			CuppingSampleID: sample.ID,
			CuppingScoreItem: CuppingScoreItem{
				Acidity:       9,
				Fragrance:     9,
				FlavorTags:    datatypes.Slot{"test", "flavor", "tag1", "tag2"},
				FlavorNotes:   "test flavor notes",
				Body:          9,
				BodyTags:      datatypes.Slot{"test", "body", "tag1", "tag2"},
				PrimaryTastes: datatypes.Slot{"test", "primary", "tag1", "tag2"},
				Aroma:         9,
				Aftertaste:    9,
				Sweetness:     9,
				Overall:       9,
				Defects:       &jsonDefects,
				Notes:         "test notes",
			},
		}
	}
	err := UpdateScore(&cuppingScore)
	if err != nil {
		return nil, err
	}
	return &cuppingScore, nil
}
