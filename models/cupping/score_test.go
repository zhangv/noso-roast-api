package cupping

import (
	"hahacking/models/bean"
	"hahacking/models/roast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateScore(t *testing.T) {
	setupTestDB()
	b, batch, _ := bean.NewBean()
	roasto, roastedBatch, _ := roast.NewRoast(b, batch)
	cp, _ := NewCupping(b, roastedBatch, roasto)
	assert.NotNil(t, cp.ID)
	samples, _ := SearchSample(SearchSampleForm{
		CuppingID: cp.ID,
	})
	assert.Equal(t, len(samples), 1)

	jsonDefects := Defects{
		{
			Name:  "霉味",
			Level: "full",
			Desc:  "霉味",
		},
		{
			Name:  "霉味",
			Level: "quark",
			Desc:  "霉味",
		},
	}
	cuppingScore := CuppingScore{
		CuppingID:       cp.ID,
		CuppingSampleID: samples[0].ID,
		BeanID:          &b.ID,
		CuppingScoreItem: CuppingScoreItem{
			Acidity:    9,
			Fragrance:  9,
			Flavor:     9,
			Body:       9,
			Aftertaste: 9,
			Sweetness:  9,
			Overall:    9,
			Defects:    &jsonDefects,
			Notes:      "test notes",
		},
	}
	err := UpdateScore(&cuppingScore)
	if err != nil {
		t.Fatalf("更新杯测评分失败: %v", err)
	}
	assert.NotNil(t, cuppingScore.ID)
	cuppingScorePtr, err := GetScoreById(cuppingScore.ID)
	if err != nil {
		t.Fatalf("获取杯测评分失败: %v", err)
	}
	assert.Equal(t, float64(63), cuppingScorePtr.Total)
	assert.Equal(t, b.ID, *cuppingScorePtr.BeanID)
	assert.Equal(t, 2, len(*cuppingScorePtr.Defects))
	assert.Equal(t, (*cuppingScorePtr.Defects)[0].Name, "霉味")
	assert.Equal(t, (*cuppingScorePtr.Defects)[0].Level, "full")
	assert.Equal(t, (*cuppingScorePtr.Defects)[0].Desc, "霉味")
	assert.Equal(t, (*cuppingScorePtr.Defects)[1].Name, "霉味")
	assert.Equal(t, (*cuppingScorePtr.Defects)[1].Level, "quark")
	assert.Equal(t, (*cuppingScorePtr.Defects)[1].Desc, "霉味")
	assert.Equal(t, 6, int(cuppingScorePtr.DefectsTotal))
}

func TestGetSampleScoreSummary(t *testing.T) {
	setupTestDB()
	b, batch, _ := bean.NewBean()
	roasto, roastedBatch, _ := roast.NewRoast(b, batch)
	cupping, _ := NewCupping(b, roastedBatch, roasto)         //默认加一个样本
	sample2, _ := NewCuppingSample(cupping, b, batch, roasto) //加入第二个样本
	cuppingScore, _ := NewCuppingScore(cupping, sample2, &CuppingScore{
		CuppingID:       cupping.ID,
		CuppingSampleID: sample2.ID,
		CuppingScoreItem: CuppingScoreItem{
			Acidity:    7,
			Fragrance:  7,
			Flavor:     7,
			Body:       7,
			Aftertaste: 7,
			Sweetness:  7,
			Overall:    7,
			Defects: &Defects{
				{
					Name:  "霉味",
					Level: "full",
					Desc:  "霉味",
				},
				{
					Name:  "霉味",
					Level: "quark",
					Desc:  "霉味",
				},
			},
			Notes: "test notes",
		},
	})
	assert.Equal(t, float64(6), cuppingScore.DefectsTotal)
	cuppingScore2, _ := NewCuppingScore(cupping, sample2, &CuppingScore{
		CuppingID:       cupping.ID,
		CuppingSampleID: sample2.ID,
		CuppingScoreItem: CuppingScoreItem{
			Acidity:    9,
			Fragrance:  9,
			Flavor:     9,
			Body:       9,
			Aftertaste: 9,
			Sweetness:  9,
			Overall:    9,
			Defects: &Defects{
				{
					Name:  "霉味",
					Level: "quark",
					Desc:  "霉味",
				},
			},
			Notes: "test notes",
		},
	})
	assert.Equal(t, float64(2), cuppingScore2.DefectsTotal)
	cuppingScoreSummary, err := GetSampleScoreSummary(sample2.ID)
	if err != nil {
		t.Fatalf("获取杯测评分汇总失败: %v", err)
	}
	assert.Equal(t, cuppingScore.CuppingSampleID, cuppingScoreSummary.CuppingSampleID)
	assert.Equal(t, float64(56), cuppingScoreSummary.Total)
	assert.Equal(t, 2, cuppingScoreSummary.Count)
	assert.Equal(t, float64(4), cuppingScoreSummary.DefectsTotal)
}
