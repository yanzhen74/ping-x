package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yanzhen74/ping-x/internal/config"
	"github.com/yanzhen74/ping-x/internal/proto"
	"github.com/spf13/cobra"
)

var recvCmd = &cobra.Command{
	Use:   "recv",
	Short: "以 receiver 模式运行，监听并响应探测包",
	Long:  `以 receiver 模式运行，监听指定端口或多播组，接收探测包并回复确认。`,
	RunE:  runRecv,
}

func init() {
	rootCmd.AddCommand(recvCmd)

	recvCmd.Flags().StringP("proto", "p", "tcp", "协议类型: tcp/udp/multicast/ssm")
	recvCmd.Flags().StringP("bind", "b", "0.0.0.0", "监听绑定地址")
	recvCmd.Flags().Int("port", 0, "监听端口 (必填)")
	recvCmd.Flags().StringP("group", "g", "", "多播组地址 (multicast/ssm 必填)")
	recvCmd.Flags().String("source", "", "SSM 源地址 (ssm 必填)")
	recvCmd.Flags().StringP("iface", "i", "", "网络接口名 (多播时使用)")
}

func runRecv(cmd *cobra.Command, args []string) error {
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
			if strings.ToLower(tc.Mode) != "recv" {
				continue
			}
			if err := tc.Validate(); err != nil {
				return fmt.Errorf("配置 [%s] 验证失败: %w", tc.Name, err)
			}
			receiver, err := proto.NewReceiver(tc.Proto)
			if err != nil {
				return fmt.Errorf("配置 [%s] 创建 receiver 失败: %w", tc.Name, err)
			}
			tcCopy := tc
			if err := receiver.Receive(ctx, &tcCopy); err != nil {
				return fmt.Errorf("配置 [%s] 接收失败: %w", tc.Name, err)
			}
		}
		return nil
	}

	protoName, _ := cmd.Flags().GetString("proto")
	bind, _ := cmd.Flags().GetString("bind")
	port, _ := cmd.Flags().GetInt("port")
	group, _ := cmd.Flags().GetString("group")
	source, _ := cmd.Flags().GetString("source")
	iface, _ := cmd.Flags().GetString("iface")

	tc := &config.TestConfig{
		Proto:  protoName,
		Mode:   "recv",
		Bind:   bind,
		Port:   port,
		Group:  group,
		Source: source,
		Iface:  iface,
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

	receiver, err := proto.NewReceiver(tc.Proto)
	if err != nil {
		return fmt.Errorf("创建 receiver 失败: %w", err)
	}
	return receiver.Receive(ctx, tc)
}
