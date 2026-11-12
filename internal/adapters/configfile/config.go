package configfile

import (
	"fmt"

	"github.com/Byroncl/exposedwatch/internal/domain"
)

// Config is the content of the JSON file.
type Config struct {
	Version  int              `json:"version"`
	Telegram TelegramSettings `json:"telegram"`
	Alerts   AlertSettings    `json:"alerts"`
}

// TelegramSettings holds the bot credentials. The token is a secret.
type TelegramSettings struct {
	Enabled  bool            `json:"enabled"`
	BotToken domain.BotToken `json:"bot_token"`
	ChatID   domain.ChatID   `json:"chat_id"`
}

// AlertSettings tune when and how often alerts are produced.
type AlertSettings struct {
	MinSeverity     string `json:"min_severity"`
	CooldownMinutes int    `json:"cooldown_minutes"`
	IntervalSeconds int    `json:"interval_seconds"`
}

// Defaults returns the configuration used when nothing is configured.
func Defaults() Config {
	return Config{
		Version: CurrentVersion,
		Alerts: AlertSettings{
			MinSeverity:     defaultMinSeverity,
			CooldownMinutes: defaultCooldownMinutes,
			IntervalSeconds: defaultIntervalSeconds,
		},
	}
}

// ParsedMinSeverity parses the configured threshold.
func (settings AlertSettings) ParsedMinSeverity() (domain.Severity, error) {
	return domain.ParseSeverity(settings.MinSeverity)
}

// Validate checks the version and the bounds of every setting. Its errors
// wrap ErrInvalidConfig and never include the token.
func (config Config) Validate() error {
	if config.Version != CurrentVersion {
		return invalid(fmt.Sprintf(invalidVersionFormat, config.Version, CurrentVersion))
	}
	if _, err := config.Alerts.ParsedMinSeverity(); err != nil {
		return invalid(invalidSeverityMessage)
	}
	if config.Alerts.CooldownMinutes < MinCooldownMinutes || config.Alerts.CooldownMinutes > MaxCooldownMinutes {
		return invalid(fmt.Sprintf(invalidCooldownFormat, MinCooldownMinutes, MaxCooldownMinutes))
	}
	if config.Alerts.IntervalSeconds < MinIntervalSeconds || config.Alerts.IntervalSeconds > MaxIntervalSeconds {
		return invalid(fmt.Sprintf(invalidIntervalFormat, MinIntervalSeconds, MaxIntervalSeconds))
	}
	return nil
}
