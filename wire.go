//go:generate wire
//go:build wireinject
// +build wireinject

package main

/**
参考实例：https://github.com/DuanJiaNing/thewaytowire
https://medium.com/@dche423/master-wire-cn-d57de86caa1b
https://github.com/DrmagicE/wire-examples
*/
import (
	"context"
	"hahacking/config"
	"log"
	"net/http"

	"github.com/google/wire"
)

var configSet = wire.NewSet(
	// 该 provider 只为演示 wire.Value 用法，具体项目这样做可能不是最好的方式。
	wire.Value(config.BasePath(".")),
	config.NewConfig,
)

// var datasourceSet = wire.NewSet(
// 	db.NewPostgresClient,
// 	wire.Bind(new(db.Client), new(*db.PostgresClient)),
// )

// var serviceSet = wire.NewSet(
// 	image.NewService,
// 	user.NewService, wire.Bind(new(user.ImageServiceProvider), new(*image.Service)),
// )

// var handlerSet = wire.NewSet(
// 	user.NewHandler,
// 	tests.NewHandler,
// 	// ...
// 	wire.Struct(new(Handlers), "*"),
// )

func setup(ctx context.Context) (sv *http.Server, clean func(), err error) {
	wire.Build(
		configSet,
		// datasourceSet,
		// serviceSet,
		// handlerSet,

		NewServer,
	)

	return nil, func() {
		log.Println("cleanup everything")
	}, nil
}
