package main

import (
	"eat/cmd"
	"fmt"
	"os"
)

var (
	c string // how many cpu would you want eat
	m string // how many memory would you want eat
)

func main() {
	rootCmd := cmd.RootCmd

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&c, "cpu_usage", "c", "0", "How many cpu would you want eat")
	rootCmd.PersistentFlags().StringVarP(&m, "memory_usage", "m", "0m", "How many memory would you want eat(GB)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
