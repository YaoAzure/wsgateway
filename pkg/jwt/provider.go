package jwt

import (
	"github.com/samber/do/v2"
)

func RegisterJWTService(injector do.Injector) {
	do.Provide(injector, NewToken)
	do.Provide(injector, NewUserToken)	
}