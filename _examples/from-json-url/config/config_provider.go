package config

import (
	"sync"
)

var (
	service     *ConfigSvc
	serviceOnce sync.Once
)

func ProvideConfigSvc() *ConfigSvc {
	serviceOnce.Do(func() {
		service = &ConfigSvc{}
	})
	return service
}
