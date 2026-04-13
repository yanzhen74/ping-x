package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "生成示例 YAML 配置文件",
	Long:  `生成示例 YAML 配置文件，可输出到标准输出或指定文件。`,
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.Flags().StringP("output", "o", "", "输出文件路径 (默认输出到标准输出)")
}

const sampleConfig = `# ping-x 示例配置文件
# 每个测试定义一个探测任务，分别在 sender 和 receiver 端运行对应 mode 的配置

tests:
  - name: "tcp-basic"
    proto: tcp
    mode: send
    target: 192.168.1.100
    port: 9000
    count: 100
    interval: 1s
    timeout: 3s
    size: 64

  - name: "tcp-basic-recv"
    proto: tcp
    mode: recv
    port: 9000

  - name: "udp-test"
    proto: udp
    mode: send
    target: 192.168.1.100
    port: 9001
    count: 50
    interval: 1s

  - name: "udp-test-recv"
    proto: udp
    mode: recv
    port: 9001

  - name: "multicast-group1"
    proto: multicast
    mode: recv
    group: 239.1.1.1
    port: 9002
    iface: eth0

  - name: "multicast-group1-send"
    proto: multicast
    mode: send
    group: 239.1.1.1
    port: 9002
    iface: eth0
    count: 20

  - name: "ssm-source1"
    proto: ssm
    mode: recv
    group: 232.1.1.1
    source: 10.0.0.50
    port: 9003
    iface: eth0

  - name: "ssm-source1-send"
    proto: ssm
    mode: send
    group: 232.1.1.1
    port: 9003
    iface: eth0
    bind: 10.0.0.50
    count: 20
`

func runConfig(cmd *cobra.Command, args []string) error {
	output, _ := cmd.Flags().GetString("output")

	if output == "" {
		fmt.Print(sampleConfig)
		return nil
	}

	if err := os.WriteFile(output, []byte(sampleConfig), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	fmt.Printf("示例配置文件已写入: %s\n", output)
	return nil
}
