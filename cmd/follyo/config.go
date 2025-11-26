package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Short:   "Manage user preferences",
	Long: `Manage user preferences for Follyo.

Available settings:
  prices    - Enable/disable live price fetching by default (on/off)
  colors    - Enable/disable colored output (on/off)
  platform  - Set default platform for new entries`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [SETTING]",
	Short: "Get current configuration values",
	Long: `Get current configuration values.

If no setting is specified, shows all settings.

Available settings: prices, colors, platform`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()

		if len(args) == 0 {
			// Show all settings
			fmt.Fprintln(osStdout, "Current Configuration:")
			fmt.Fprintln(osStdout)
			fmt.Fprintf(osStdout, "  prices:   %s\n", boolToOnOff(cfg.GetFetchPrices()))
			fmt.Fprintf(osStdout, "  colors:   %s\n", boolToOnOff(cfg.GetColorOutput()))
			platform := cfg.GetDefaultPlatform()
			if platform == "" {
				platform = "(not set)"
			}
			fmt.Fprintf(osStdout, "  platform: %s\n", platform)
			return
		}

		setting := strings.ToLower(args[0])
		switch setting {
		case "prices":
			fmt.Fprintf(osStdout, "prices: %s\n", boolToOnOff(cfg.GetFetchPrices()))
		case "colors":
			fmt.Fprintf(osStdout, "colors: %s\n", boolToOnOff(cfg.GetColorOutput()))
		case "platform":
			platform := cfg.GetDefaultPlatform()
			if platform == "" {
				fmt.Fprintln(osStdout, "platform: (not set)")
			} else {
				fmt.Fprintf(osStdout, "platform: %s\n", platform)
			}
		default:
			fmt.Fprintf(osStderr, "Unknown setting: %s\n", setting)
			fmt.Fprintln(osStderr, "Available settings: prices, colors, platform")
			osExit(1)
		}
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set SETTING VALUE",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available settings:
  prices    - Enable/disable live price fetching (on/off)
  colors    - Enable/disable colored output (on/off)
  platform  - Set default platform for new entries (any string, or "clear" to remove)

Examples:
  follyo config set prices off
  follyo config set colors on
  follyo config set platform Coinbase
  follyo config set platform clear`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		setting := strings.ToLower(args[0])
		value := args[1]

		switch setting {
		case "prices":
			boolVal, ok := parseOnOff(value)
			if !ok {
				fmt.Fprintf(osStderr, "Invalid value for prices: %s (use on/off)\n", value)
				osExit(1)
			}
			if err := cfg.SetFetchPrices(boolVal); err != nil {
				fmt.Fprintf(osStderr, "Error saving config: %v\n", err)
				osExit(1)
			}
			fmt.Fprintf(osStdout, "Set prices to %s\n", boolToOnOff(boolVal))

		case "colors":
			boolVal, ok := parseOnOff(value)
			if !ok {
				fmt.Fprintf(osStderr, "Invalid value for colors: %s (use on/off)\n", value)
				osExit(1)
			}
			if err := cfg.SetColorOutput(boolVal); err != nil {
				fmt.Fprintf(osStderr, "Error saving config: %v\n", err)
				osExit(1)
			}
			fmt.Fprintf(osStdout, "Set colors to %s\n", boolToOnOff(boolVal))

		case "platform":
			if strings.ToLower(value) == "clear" {
				if err := cfg.ClearDefaultPlatform(); err != nil {
					fmt.Fprintf(osStderr, "Error saving config: %v\n", err)
					osExit(1)
				}
				fmt.Fprintln(osStdout, "Cleared default platform")
			} else {
				if err := cfg.SetDefaultPlatform(value); err != nil {
					fmt.Fprintf(osStderr, "Error saving config: %v\n", err)
					osExit(1)
				}
				fmt.Fprintf(osStdout, "Set default platform to %s\n", value)
			}

		default:
			fmt.Fprintf(osStderr, "Unknown setting: %s\n", setting)
			fmt.Fprintln(osStderr, "Available settings: prices, colors, platform")
			osExit(1)
		}
	},
}

// boolToOnOff converts a bool to "on" or "off"
func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// parseOnOff parses "on"/"off"/"true"/"false"/"1"/"0" to bool
func parseOnOff(s string) (bool, bool) {
	switch strings.ToLower(s) {
	case "on", "true", "1", "yes":
		return true, true
	case "off", "false", "0", "no":
		return false, true
	default:
		return false, false
	}
}
