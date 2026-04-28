// Package config는 서버 설정을 로드한다.
// 우선순위: 환경변수(CLEDYU_*) > config.yaml > 기본값.
// config.yaml이 없으면 기본값으로 동작 (로컬 개발 편의).
// 전체 항목은 config.example.yaml 참조.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	VM       VMConfig       `mapstructure:"vm"`
	AI       AIConfig       `mapstructure:"ai"`
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	KeycloakURL   string `mapstructure:"keycloak_url"`
	KeycloakRealm string `mapstructure:"keycloak_realm"`
	ClientID      string `mapstructure:"client_id"`
	ClientSecret  string `mapstructure:"client_secret"`
	RedirectURL   string `mapstructure:"redirect_url"`
}

type VMConfig struct {
	Provider        string `mapstructure:"provider"`
	KubeAPIServer   string `mapstructure:"kube_api_server"`
	KubeToken       string `mapstructure:"kube_token"`
	KubeVirtNS      string `mapstructure:"kubevirt_namespace"`
	EC2Region       string `mapstructure:"ec2_region"`
	EC2AMI          string `mapstructure:"ec2_ami"`
	EC2InstanceType string `mapstructure:"ec2_instance_type"`
	EC2LaunchTpl    string `mapstructure:"ec2_launch_template"`
}

type AIConfig struct {
	BFFURL string `mapstructure:"bff_url"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetEnvPrefix("CLEDYU")
	v.AutomaticEnv()

	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("database.url", "postgres://postgres:postgres@localhost:5432/cledyu?sslmode=disable")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("vm.provider", "hybrid")
	v.SetDefault("vm.kubevirt_namespace", "lab-sessions")
	v.SetDefault("vm.ec2_region", "ap-northeast-2")
	v.SetDefault("vm.ec2_instance_type", "t3.medium")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
