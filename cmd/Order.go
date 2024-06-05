package cmd

import "github.com/urfave/cli/v2"

var OrderCmd = &cli.Command{
	Name:  "order",
	Usage: "Manager orders",
	Action: func(cctx *cli.Context) error {

		return nil
	},
}
