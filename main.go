package main

import (
	"fmt"
	"os"

	"eat/cmd"
)

func main() {
	rootCmd := cmd.RootCmd

	// Add global flags
	rootCmd.PersistentFlags().StringP("cpu_usage", "c", "0", "How many cpu would you want eat")
	rootCmd.PersistentFlags().StringP("memory_usage", "m", "0m", "How many memory would you want eat(GB)")
	rootCmd.PersistentFlags().StringP("time_deadline", "t", "24h", "deadline to quit eat process")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
