package tag

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchGroup(t *testing.T) {
	setupTestDB()
	uniquetag := time.Now().Format("2006-01-02 15:04:05")
	form := SearchGroupForm{
		Names: []string{uniquetag},
	}
	os := SearchGroup(form)

	assert.Equal(t, 0, len(os))

	tg, err := AddGroup(uniquetag)
	assert.NoError(t, err)
	os = SearchGroup(form)
	assert.Equal(t, 1, len(os))

	tg.Delete()

}
