package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = Name
	app.Version = Version
	app.Author = "yukimemi"
	app.Email = "yukimemi@gmail.com"
	app.Usage = "Run git status recursively"

	app.Flags = GlobalFlags
	app.Action = GlobalAction
	app.Run(os.Args)
}
