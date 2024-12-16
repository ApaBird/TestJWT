package app

import (
	"app/pkg/config"
	"app/pkg/httpserver"
)

func main() {
	config.LoadConfig(".env")

	err := httpserver.StartHttpServer(config.GetEnv("PORT", ":8080"))
	if err != nil {
		panic(err)
	}
}
