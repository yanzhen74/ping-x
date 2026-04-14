package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/yanzhen74/ping-x/internal/config"
	"github.com/yanzhen74/ping-x/internal/proto"
	"github.com/yanzhen74/ping-x/internal/stats"
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "以 sender 模式运行，向目标发送探测包",
	Long:  `以 sender 模式运行，根据指定的协议向目标地址发送探测包，并统计联通性。`,
	RunE:  runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)

	sendCmd.Flags().StringP("proto", "p", "tcp", "协议类型: tcp/udp/multicast/ssm")
	sendCmd.Flags().StringP("target", "t", "", "目标地址 (tcp/udp 必填)")
	sendCmd.Flags().StringP("group", "g", "", "多播组地址 (multicast/ssm 必填)")
	sendCmd.Flags().Int("port", 0, "目标端口 (必填)")
	sendCmd.Flags().StringP("bind", "b", "0.0.0.0", "本地绑定地址")
	sendCmd.Flags().StringP("iface", "i", "", "网络接口名 (多播时使用)")
	sendCmd.Flags().IntP("count", "n", 0, "发送次数 (0=持续发送)")
	sendCmd.Flags().String("interval", "1s", "发送间隔")
	sendCmd.Flags().String("timeout", "3s", "单包超时时间")
	sendCmd.Flags().IntP("size", "s", 64, "探测包大小 (bytes)")
	sendCmd.Flags().Int("ttl", 0, "TTL (0=自动选择: unicast=64, multicast=16)")
}

func runSend(cmd *cobra.Command, args []string) error {
	// 如果指定了配置文件，从文件加载
	if cfgFile != "" {
		fileCfg, err := config.LoadFromFile(cfgFile)
		if err != nil {
			return fmt.Errorf("加载配置文件失败: %w", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()
		defer cancel()

		for _, tc := range fileCfg.Tests {
			if strings.ToLower(tc.Mode) != "send" {
				continue
			}
			if err := tc.Validate(); err != nil {
				return fmt.Errorf("配置 [%s] 验证失败: %w", tc.Name, err)
			}
			sender, err := proto.NewSender(tc.Proto)
			if err != nil {
				return fmt.Errorf("配置 [%s] 创建 sender 失败: %w", tc.Name, err)
			}
			stat := stats.NewCollector()
			tcCopy := tc
			if err := sender.Send(ctx, &tcCopy, stat); err != nil {
				return fmt.Errorf("配置 [%s] 发送失败: %w", tc.Name, err)
			}
		}
		return nil
	}

	// 从命令行参数构建配置
	protoName, _ := cmd.Flags().GetString("proto")
	target, _ := cmd.Flags().GetString("target")
	group, _ := cmd.Flags().GetString("group")
	port, _ := cmd.Flags().GetInt("port")
	bind, _ := cmd.Flags().GetString("bind")
	iface, _ := cmd.Flags().GetString("iface")
	count, _ := cmd.Flags().GetInt("count")
	interval, _ := cmd.Flags().GetString("interval")
	timeout, _ := cmd.Flags().GetString("timeout")
	size, _ := cmd.Flags().GetInt("size")
	ttl, _ := cmd.Flags().GetInt("ttl")

	tc := &config.TestConfig{
		Proto:    protoName,
		Mode:     "send",
		Target:   target,
		Group:    group,
		Port:     port,
		Bind:     bind,
		Iface:    iface,
		Count:    count,
		Interval: interval,
		Timeout:  timeout,
		Size:     size,
		TTL:      ttl,
	}

	if err := tc.Validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()
	defer cancel()

	sender, err := proto.NewSender(tc.Proto)
	if err != nil {
		return fmt.Errorf("创建 sender 失败: %w", err)
	}
	stat := stats.NewCollector()
	return sender.Send(ctx, tc, stat)
}
