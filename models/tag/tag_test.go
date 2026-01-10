package tag

import (
	"hahacking/config"
	"hahacking/database"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	config.InitViper("test", "../../")
	database.Setup()
	db := database.GetConn()
	// 自动迁移模型
	db.AutoMigrate(&Tag{}, &TagGroup{}, &Tagging{})
	return db
}

func TestAddTag(t *testing.T) {
	setupTestDB()
	uniquetag := time.Now().Format("20060102150405")
	tag, err := AddTag(uniquetag, "冲煮方式", "1", "beans")
	assert.NoError(t, err)
	assert.Equal(t, uniquetag, tag.Name)

	tag, err = AddTag(uniquetag+"#2", "冲煮方式", "1", "beans")
	assert.NoError(t, err)
}

func TestGetByTarget(t *testing.T) {
	targetid := "1"
	targettype := "beans" + time.Now().Format("200601021504")
	currenttags, _ := GetByTarget(targetid, targettype)
	log.Println("current tag count", (currenttags))

	uniquetag := time.Now().Format("20060102150405")
	AddTag(uniquetag, "冲煮方式", targetid, targettype)
	AddTag(uniquetag+"#2", "冲煮方式", targetid, targettype)

	newtags, _ := GetByTarget(targetid, targettype)
	log.Println("new tag count", (newtags))

	assert.Equal(t, len(currenttags)+2, len(newtags))

}

func TestGetGroupByTarget(t *testing.T) {
	setupTestDB()

	groups, _ := GetGroupByTarget("1", "beans")
	// log.Println("group count", len(groups))
	// log.Println("group", groups)
	assert.Equal(t, 1, len(groups))

}
