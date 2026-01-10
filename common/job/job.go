package job

import (
	"log"

	cron "github.com/robfig/cron/v3"
)

func Setup() {
	c := cron.New()
	//run per minute
	c.AddFunc("0 */1 * * * ?", func() {
		log.Println("Check if order is expired...")
	})
	c.Start()
}
