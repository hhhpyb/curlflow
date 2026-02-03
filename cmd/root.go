package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// StartApp 是启动 Wails GUI 应用的回调函数
// 在 main.go 中赋值
var StartApp func()

var rootCmd = &cobra.Command{
	Use:   "curlflow",
	Short: "A desktop tool to convert curl commands to code",
	Long:  `CurlFlow is a desktop application built with Wails to convert curl commands into various programming language requests.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 默认行为：启动 GUI
		if StartApp != nil {
			StartApp()
		} else {
			fmt.Println("No GUI application handler defined.")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
