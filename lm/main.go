package main

import (
	"fmt"

	"github.com/organicveggie/livemusic/lm/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
