package config

import "os"

type config struct {
	DBConnectionStr  string
	AmqConnectionStr string
	Port             string
	JWTSecret        string
}

func getEnv(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func LoadConfig() *config {
	return &config{
		DBConnectionStr:  getEnv("DB_URL", "host=localhost user=postgres dbname=postgres password=jomum port=5432 sslmode=disable"),
		AmqConnectionStr: getEnv("AMQ_URL", "amqp://guest:guest@localhost:5672/"),
		Port:             getEnv("PORT", ":"+"8080"),
		JWTSecret:        getEnv("JWT_SECRET", "SHHHHHHHH"),
	}
}
