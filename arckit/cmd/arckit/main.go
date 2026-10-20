package main

import (
	"os"

	"github.com/algorandfoundation/ARCs/arckit/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
