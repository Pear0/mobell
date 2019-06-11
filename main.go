package main

import (
	"encoding/json"
	"fmt"
	"github.com/gregdel/pushover"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "mobell",
	Short: "Send mobile bells",
	Long: `Send mobile notifications from the command line with pushover.net`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here

		mustLoadConfig()
		sendNotification(pushover.NewMessageWithTitle("Triggered at "+time.Now().Format(time.RFC822), "Mobell 🔔"))

	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a config",
	Long: `Create a config in the user's home config directory`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here

		configPath := absPathify(defaultUserConfigPath)
		if configPath == "" {
			fmt.Printf("failed to resolve the config location: %s", defaultUserConfigPath)
		}

		fmt.Println("To use Mobell, you must setup an account at https://pushover.net.")
		fmt.Println("Enter your API key:")

		var apiKey string
		_, err := fmt.Scanln(&apiKey)
		if err != nil {
			fmt.Printf("failed to read input: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Enter your User key (comma separated if multiple):")

		var userKeys string
		_, err = fmt.Scanln(&userKeys)
		if err != nil {
			fmt.Printf("failed to read input: %s\n", err.Error())
			os.Exit(1)
		}

		users := strings.Split(userKeys, ",")

		err = os.MkdirAll(configPath, 0777)
		if err != nil {
			fmt.Printf("failed to create config directory: %s\n", err.Error())
			os.Exit(1)
		}

		f, err := os.Create(configPath + "/config.json")
		if err != nil {
			fmt.Printf("failed to open file for writing: %s\n", err.Error())
			os.Exit(1)
		}

		err = json.NewEncoder(f).Encode(map[string]interface{}{
			"pushover_api_key": apiKey,
			"pushover_user_key": users,
		})
		if err != nil {
			fmt.Printf("error occurred while writing file: %s\n", err.Error())
			os.Exit(1)
		}

		err = f.Close()
		if err != nil {
			fmt.Printf("error occurred while closing file: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Config saved.")
		fmt.Println()

		if len(apiKey) != 30 {
			fmt.Printf("Note: pushover.net api keys are usually 32 characters but the one you provided is %d characters.\n", len(apiKey))
		}

		for i, userKey := range users {
			if len(userKey) != 30 {
				fmt.Printf("Note: pushover.net user keys are usually 32 characters but #%d is %d characters.\n", i+1, len(apiKey))
			}
		}

	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Send a custom notification",
	Long: `Send a custom notification`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here

		mustLoadConfig()

		message := "Triggered at " + time.Now().Format(time.RFC822)
		if len(args) > 0 {
			message = args[0]
		}

		sendNotification(pushover.NewMessageWithTitle(message, "Mobell 🔔"))
	},
}

var defaultUserConfigPath string

func init() {
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		defaultUserConfigPath = "$XDG_CONFIG_HOME/mobell"
	} else {
		defaultUserConfigPath = "$HOME/.config/mobell"
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(pushCmd)
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func absPathify(inPath string) string {
	if strings.HasPrefix(inPath, "$HOME") {
		inPath = userHomeDir() + inPath[5:]
	}

	if strings.HasPrefix(inPath, "$") {
		end := strings.Index(inPath, string(os.PathSeparator))
		inPath = os.Getenv(inPath[1:end]) + inPath[end:]
	}

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath)
	}

	p, err := filepath.Abs(inPath)
	if err == nil {
		return filepath.Clean(p)
	}

	return ""
}

func mustLoadConfig() {
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Could not config.json: %s\n", err.Error())
			fmt.Printf("Use `%s init` to create a config file in %s\n", os.Args[0], defaultUserConfigPath)
		} else {
			fmt.Printf("Failed to read config.json file: %s\n", err.Error())
		}

		os.Exit(1)
	}

	requiredKeys := []string{"pushover_user_key", "pushover_api_key"}
	missingKeys := false

	for _, key := range requiredKeys {
		if viper.Get(key) == nil {
			fmt.Printf("config file is missing '%s'\n", key)
			missingKeys = true
		}
	}

	if missingKeys {
		fmt.Printf("Use `%s init` to regenerate the config file in %s\n", os.Args[0], defaultUserConfigPath)
		os.Exit(1)
	}
}

func sendNotification(message *pushover.Message) {
	message.Priority = pushover.PriorityHigh


	app := pushover.New(viper.GetString("pushover_api_key"))
	recipientKeys := viper.GetStringSlice("pushover_user_key")

	fmt.Println(viper.GetString("pushover_api_key"))

	var wg sync.WaitGroup

	fmt.Printf("Sending to %d users\n", len(recipientKeys))

	for _, recipientKey := range recipientKeys {
		fmt.Println(recipientKey)
		wg.Add(1)
		go func(recipientKey string) {
			recipient := pushover.NewRecipient(recipientKey)

			_, err := app.SendMessage(message, recipient)
			if err != nil {
				fmt.Printf("failed to send to %s: %s\n", recipientKey, err.Error())
			}

			wg.Done()
		}(recipientKey)
	}

	wg.Wait()
	fmt.Printf("Sent to %d users.\n", len(recipientKeys))
}

func main() {

	viper.SetConfigType("json")
	viper.SetConfigName("config") // name of config file (without extension)

	if os.Getenv("MOBELL_CONF_PATH") != "" {
		viper.AddConfigPath("$MOBELL_CONF_PATH")
	}
	viper.AddConfigPath(defaultUserConfigPath)
	viper.AddConfigPath("/etc/mobell/")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
