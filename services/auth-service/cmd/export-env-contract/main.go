package main

import (
	"fmt"
	"os"

	"csort.ru/auth-service/internal/config"
)

func main() {
	for _, k := range config.EnvKeys() {
		if _, err := fmt.Fprintln(os.Stdout, k); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
