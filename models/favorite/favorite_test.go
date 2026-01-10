package favorite

import (
	"hahacking/config"
	"hahacking/database"
	"hahacking/models"
	"hahacking/models/item"
	"hahacking/models/report"
	"hahacking/models/user"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var setupOnce sync.Once

func setup() {
	setupOnce.Do(func() {
		config.Setup("../../", "test") //相对当前路径
		// logger.Setup()                 // 初始化logger
		// database.SetLogLevel(logger.Silent)
		database.Setup()
		db := database.GetConn()
		db.AutoMigrate(
			&Favorite{}, &report.Report{}, &item.Item{},
		)
	})
}

func TestSearch(t *testing.T) {
	setup()

	it, _ := item.Create(&item.Item{
		Name: "testing",
		Code: "testing",
	})
	u := user.User{
		Nickname: "testnick",
		Mobile:   nil,
	}
	user.Update(&u)
	uid := u.ID

	err := Update(&Favorite{
		ItemID: strconv.Itoa(int(it.ID)),
		Type:   "Item",
		UserID: &uid,
		Status: models.StatusValid,
	})

	assert.Nil(t, err)

	results, _ := Search(SearchForm{
		ItemID: strconv.Itoa(int(it.ID)),
		UserID: strconv.Itoa(int(uid)),
		Type:   "Item",
		Status: string(models.StatusValid),
	})

	assert.NotNil(t, results)
	assert.Equal(t, 1, len(results))
	ii := results[0].(*item.Item)

	assert.Equal(t, it.ID, uint(ii.ID))
}
