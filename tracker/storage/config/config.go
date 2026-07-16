//Содердимое файла .env
//# --- Переменные для подключения к PostgreSQL ---
//DB_HOST=localhost
//DB_PORT=5432
//DB_USER=логин
//DB_PASSWORD=пароль
//DB_NAME=taskmanager (имя бд)

package repository

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {

	// Настройки PostgreSQL
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	SSLMode    string // disable, require, verify-ca, verify-full
}

// LoadConfig загружает конфигурацию из .env и переменных окружения.
// Если файл .env не найден – не ошибка, используются системные переменные.
// Возвращает заполненную структуру Config или ошибку (если критическая переменная отсутствует).
func LoadConfig() (*Config, error) {
	// Пытаемся загрузить .env, но не паникуем при отсутствии
	if err := godotenv.Load(); err != nil {
		log.Println("config: .env файл не найден, используются системные переменные окружения")
	}

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnvAsInt("DB_PORT", 5432),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "postgres"),
		SSLMode:    getEnv("DB_SSL_MODE", "disable"),
	}

	// Обязательные проверки (например, пароль не может быть пустым)
	if cfg.DBPassword == "" {
		log.Fatal("DB_PASSWORD не задан. Приложение не может запуститься.")
	}

	return cfg, nil
}

// DSN возвращает строку подключения к PostgreSQL в формате, понятном драйверу lib/pq.
func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + strconv.Itoa(c.DBPort) +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode
}

// DatabaseURL возвращает строку подключения в формате URL для драйвера migrate.
func (c *Config) DatabaseURL() string {
	// Формируем URL: postgres://user:pass@host:port/dbname?sslmode=mode
	// Пароль нужно экранировать, если там есть спецсимволы (например, @)
	password := url.QueryEscape(c.DBPassword)
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, password, c.DBHost, c.DBPort, c.DBName, c.SSLMode)
}

// Вспомогательные функции для чтения переменных с типизацией и значениями по умолчанию

// getEnv возвращает значение переменной окружения или defaultVal, если она не установлена.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvAsInt возвращает целое число из переменной окружения или defaultVal при ошибке/отсутствии.
func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
		log.Printf("config: неверный формат %s = %s, используется значение по умолчанию %d", key, value, defaultVal)
	}
	return defaultVal
}
