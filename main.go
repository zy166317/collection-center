package main

import (
	"collection-center/cmd"
	"collection-center/internal/logger"
	"github.com/urfave/cli/v2"
	"os"
)

// @title collection-center 接口
// @version 1.0.0
// @description collection-center 接口
// @host localhost:8080
// @BasePath /
func main() {
	local := []*cli.Command{
		cmd.RunCmd,
	}
	app := &cli.App{
		Name:  "collection-center",
		Usage: "collection-center",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "conf",
				Value: "./resources",
			},
		},

		Commands: local,
	}
	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}
