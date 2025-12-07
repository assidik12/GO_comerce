package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

// Config struct untuk menampung semua konfigurasi dari environment.
type Config struct {
	DBUser        string `mapstructure:"MYSQL_USER"`
	DBHost        string `mapstructure:"MYSQL_HOST"`
	DBPassword    string `mapstructure:"MYSQL_PASSWORD"`
	DBName        string `mapstructure:"MYSQL_DATABASE"`
	DBPort        string `mapstructure:"MYSQL_PORT"`
	AppPort       string `mapstructure:"APP_PORT"`
	JWTSecret     string `mapstructure:"JWT_SECRET"`
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	KafkaHost     string `mapstructure:"KAFKA_HOST"`
	KafkaPort     string `mapstructure:"KAFKA_PORT"`
}

// cfg adalah instance singleton dari Config.
var (
	cfg  *Config
	once sync.Once
)

// GetConfig memuat konfigurasi hanya sekali dan mengembalikannya.
// Fungsi ini menjadi satu-satunya cara untuk mengakses konfigurasi di seluruh aplikasi.
func GetConfig() *Config {
	once.Do(func() {
		viper.AddConfigPath(".")
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			// Abaikan jika file .env tidak ditemukan, karena env var tetap bisa dibaca.
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				log.Fatalf("Gagal membaca file konfigurasi: %s", err)
			}
		}

		// Unmarshal konfigurasi ke dalam struct cfg.
		var config Config
		if err := viper.Unmarshal(&config); err != nil {
			log.Fatalf("Gagal unmarshal konfigurasi: %s", err)
		}
		cfg = &config
	})
	return cfg
}
