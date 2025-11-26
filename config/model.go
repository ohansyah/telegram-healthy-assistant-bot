package config

type Config struct {
	Server   Server
	Telegram Telegram
	Gemini   Gemini
}

type Server struct {
	Host string
	Port string
}

type Telegram struct {
	Token string
}

type Gemini struct {
	Key   string
	Model string
}
