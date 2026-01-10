package mail

import (
	"hahacking/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup() {
	config.Setup("../../", "test")
}
func TestSend(t *testing.T) {
	setup()
	success := Send("derekzhangv@gmail.com", "Test Email", "This is a test email sent from Go using QQ SMTP.")

	assert.Equal(t, true, success)
}
