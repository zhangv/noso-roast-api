package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// r, _ := Init()
	c := viper.GetStringSlice("db")
	t.Log(c)
}

func TestGetEnvInfo(t *testing.T) {
	c := viper.GetStringSlice("db")
	t.Log(c)
}
