package config

type Config struct {
	EnableDebug   bool   `env:"ENABLE_DEBUG" envDefault:"false"`
	SlackAppToken string `env:"SLACK_APP_TOKEN"`
	SlackBotToken string `env:"SLACK_BOT_TOKEN"`

	AzureOpenAPIKey       string `env:"AZURE_OPEN_API_KEY"`
	AzureOpenAPIEndpoint  string `env:"AZURE_OPEN_API_ENDPOINT"`
	AzureOpenAPIVersion   string `env:"AZURE_OPEN_API_VERSION" envDefault:"2024-06-01"`
	AzureOpenAPIModelName string `env:"AZURE_OPEN_API_MODEL_NAME" envDefault:"gpt-4o"`
}
