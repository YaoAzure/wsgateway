package config

import (
	"github.com/samber/do/v2"
)

func RegisterConfigService(injector do.Injector, config Config) {
	do.ProvideValue(injector, config.App)	
	do.ProvideValue(injector, config.JWT)	
}
