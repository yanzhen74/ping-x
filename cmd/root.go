package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	version string
)

// rootCmd 表示根命令
var rootCmd = &cobra.Command{
	Use:   "ping-x",
	Short: "跨平台网络协议联通检测工具",
	Long: `ping-x 是一个跨平台网络协议联通检测工具，
支持 TCP、UDP、组播(Multicast)和 SSM 等多种协议的连通性测试。`,
}

// Execute 执行根命令
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// SetVersion 设置版本号
func SetVersion(v string) {
	version = v
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局 flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出模式")

	// 绑定 viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// initConfig 初始化配置
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			// 配置文件读取失败不是致命错误
			// 命令可以正常执行
		}
	}
}
