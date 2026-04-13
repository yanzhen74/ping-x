package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// TestConfig 定义单个测试配置
type TestConfig struct {
	Name     string `yaml:"name" mapstructure:"name"`
	Proto    string `yaml:"proto" mapstructure:"proto"`
	Mode     string `yaml:"mode" mapstructure:"mode"`
	Target   string `yaml:"target" mapstructure:"target"`
	Group    string `yaml:"group" mapstructure:"group"`
	Source   string `yaml:"source" mapstructure:"source"`
	Port     int    `yaml:"port" mapstructure:"port"`
	Bind     string `yaml:"bind" mapstructure:"bind"`
	Iface    string `yaml:"iface" mapstructure:"iface"`
	Count    int    `yaml:"count" mapstructure:"count"`
	Interval string `yaml:"interval" mapstructure:"interval"`
	Timeout  string `yaml:"timeout" mapstructure:"timeout"`
	Size     int    `yaml:"size" mapstructure:"size"`
	TTL      int    `yaml:"ttl" mapstructure:"ttl"`
}

// FileConfig 定义配置文件结构
type FileConfig struct {
	Tests []TestConfig `yaml:"tests" mapstructure:"tests"`
}

// LoadFromFile 从 YAML 文件加载配置
func LoadFromFile(path string) (*FileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Validate 验证单个测试配置
func (c *TestConfig) Validate() error {
	// 验证协议
	validProtos := map[string]bool{"tcp": true, "udp": true, "multicast": true, "ssm": true}
	if !validProtos[c.Proto] {
		return fmt.Errorf("invalid proto: %s (must be tcp, udp, multicast, or ssm)", c.Proto)
	}

	// 验证模式
	if c.Mode != "send" && c.Mode != "recv" {
		return fmt.Errorf("invalid mode: %s (must be send or recv)", c.Mode)
	}

	// 验证端口
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", c.Port)
	}

	// send 模式下特定验证
	if c.Mode == "send" {
		// tcp/udp 需要 target
		if (c.Proto == "tcp" || c.Proto == "udp") && c.Target == "" {
			return fmt.Errorf("target is required for %s send mode", c.Proto)
		}
		// multicast/ssm 需要 group
		if (c.Proto == "multicast" || c.Proto == "ssm") && c.Group == "" {
			return fmt.Errorf("group is required for %s send mode", c.Proto)
		}
	}

	// recv 模式下特定验证
	if c.Mode == "recv" {
		// multicast/ssm 需要 group
		if (c.Proto == "multicast" || c.Proto == "ssm") && c.Group == "" {
			return fmt.Errorf("group is required for %s recv mode", c.Proto)
		}
	}

	// 验证 interval 和 timeout 格式
	if c.Interval != "" {
		if _, err := time.ParseDuration(c.Interval); err != nil {
			return fmt.Errorf("invalid interval format: %s", c.Interval)
		}
	}
	if c.Timeout != "" {
		if _, err := time.ParseDuration(c.Timeout); err != nil {
			return fmt.Errorf("invalid timeout format: %s", c.Timeout)
		}
	}

	return nil
}

// GetInterval 返回 time.Duration，如果未设置返回默认值
func (c *TestConfig) GetInterval() time.Duration {
	if c.Interval == "" {
		return time.Second
	}
	d, err := time.ParseDuration(c.Interval)
	if err != nil {
		return time.Second
	}
	return d
}

// GetTimeout 返回 time.Duration，如果未设置返回默认值
func (c *TestConfig) GetTimeout() time.Duration {
	if c.Timeout == "" {
		return 3 * time.Second
	}
	d, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 3 * time.Second
	}
	return d
}
