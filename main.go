package main

import (
	"fmt"
	"os"

	"eat/cmd"
)

func main() {
	rootCmd := cmd.RootCmd

	rootCmd.PersistentFlags().StringP("cpu", "c", "0", "How many cpu would you want eat")
	rootCmd.PersistentFlags().StringP("memory", "m", "0m", "How many memory would you want eat(GB)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
