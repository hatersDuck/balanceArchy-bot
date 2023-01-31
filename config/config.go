package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	TelegramToken    string
	Messages         Messages
	DatabasePassword string
}

type Messages struct {
	Responses
	Errors
	Buttons
}

type Responses struct {
	Start             string `mapstructure:"start"`
	NewEvent          string `mapstructure:"new_event"`
	FirstEvent        string `mapstructure:"first_event"`
	FirstEventFailed  string `mapstructure:"first_event_failed"`
	SecondEvent       string `mapstructure:"second_event"`
	SecondEventFailed string `mapstructure:"second_event_failed"`
}

type Errors struct {
	UndefinedCommand string `mapstructure:"undefined_command"`
}

type Buttons struct {
	BtnSkip   string `mapstructure:"skip"`
	BtnAdd    string `mapstructure:"add"`
	BtnCancel string `mapstructure:"cancel"`
}

func Init() (*Config, error) {
	viper.AddConfigPath("config")
	viper.SetConfigName("main")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := viper.UnmarshalKey("messages.responses", &cfg.Messages.Responses); err != nil {
		return nil, err
	}
	if err := viper.UnmarshalKey("messages.errors", &cfg.Messages.Errors); err != nil {
		return nil, err
	}
	if err := viper.UnmarshalKey("messages.buttons", &cfg.Messages.Buttons); err != nil {
		return nil, err
	}

	if err := ParseEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ParseEnv(cfg *Config) error {
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	cfg.TelegramToken = viper.GetString("BOT_TOKEN")
	cfg.DatabasePassword = viper.GetString("DATABASE_PASS")
	return nil
}
