package config

import (
	"github.com/samber/do/v2"
)

func RegisterConfigService(injector do.Injector) {
	do.Provide(injector, NewAppConfig)	
	do.Provide(injector, NewJWTConfig)	
}
