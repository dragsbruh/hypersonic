package config

import (
	"crypto/sha256"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

var DATABASE_URL, MUSIC_DIR, DATA_DIR, ADDR, ADMIN_NETWORK, ADMIN_ADDR string
var JWT_SECRET []byte

func Load() {
	DATABASE_URL = ensureEnv("DATABASE_URL")
	MUSIC_DIR = ensureEnv("MUSIC_DIR")
	DATA_DIR = ensureEnv("DATA_DIR")
	ADDR = fallbackEnv("ADDR", ":8080")

	adminAddr := fallbackEnv("ADMIN_ADDR", "unix;/var/run/hypersonic/admin.sock")
	var both bool
	ADMIN_NETWORK, ADMIN_ADDR, both = strings.Cut(adminAddr, ";")
	if !both {
		ADMIN_ADDR = ADMIN_NETWORK
		ADMIN_NETWORK = "tcp"
	}

	jwtSecret := sha256.Sum256([]byte(ensureEnv("JWT_SECRET")))
	JWT_SECRET = jwtSecret[:]
}

func ensureEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		logrus.Fatalf("missing env variable: %q", key)
	}
	return value
}

func fallbackEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	return value
}
