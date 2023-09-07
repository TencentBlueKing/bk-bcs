package main

import (
	"context"
	"os"
)

func main() {
	ctx := context.Background()
	if err := execute(ctx); err != nil {
		os.Exit(1)
	}
}
