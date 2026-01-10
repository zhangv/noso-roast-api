package roast

import (
	"hahacking/models/bean"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveRecipe(t *testing.T) {
	setupTestDB()
	b, batch, _ := bean.NewBean()

	assert.Equal(t, 100, batch.TotalQty)
	assert.Equal(t, 100, batch.CurrentQty)

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
	roast, err := Start(startForm)
	roast, err = GetById(roast.ID)

	completeForm := CompleteForm{
		ID:           roast.ID,
		OutQty:       40,
		Agtron:       100,
		AgtronGround: 90,
		LossRate:     11.9,
		Degree:       string(RoastDegreeMedium),
	}
	err = roast.Complete(completeForm)

	recipe, err := SaveRecipeByRoastId(roast.ID)
	if err != nil {
		t.Fatalf("保存配方失败: %v", err)
	}
	assert.Equal(t, roast.BeanID, recipe.BeanID)
	assert.Equal(t, roast.InQty, recipe.BeanQty)
	profile, _ := GetProfileByRoastId(roast.ID)
	assert.Equal(t, profile.ID, recipe.RoastProfileID)
	assert.Equal(t, profile.Agtron, recipe.Agtron)
	assert.Equal(t, profile.AgtronGround, recipe.AgtronGround)
	assert.Equal(t, 11.9, recipe.LossRate)
	assert.Equal(t, profile.LossRate, recipe.LossRate)
	assert.Equal(t, profile.Degree, recipe.Degree)
	assert.Equal(t, profile.StartTemp, recipe.StartTemp)
	assert.Equal(t, profile.TpTemp, recipe.TpTemp)
	assert.Equal(t, profile.TpTime, recipe.TpTime)
	assert.Equal(t, profile.YellowTemp, recipe.YellowTemp)
	assert.Equal(t, profile.YellowTime, recipe.YellowTime)
	assert.Equal(t, profile.FcTemp, recipe.FcTemp)
	assert.Equal(t, profile.FcTime, recipe.FcTime)
	assert.Equal(t, profile.ScTemp, recipe.ScTemp)
	assert.Equal(t, profile.ScTime, recipe.ScTime)
	assert.Equal(t, profile.TotalTime, recipe.TotalTime)
}
