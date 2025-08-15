package config

import (
	"crypto/sha256"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

var MusicDir, DatabaseURL, DataDir, Addr string
var EncryptionKey [32]byte
var JWTSecret []byte
var AdminAddr struct {
	Network string
	Address string
}

func LoadConfig() {
	MusicDir = ensureEnv("MUSIC_DIR")
	DatabaseURL = ensureEnv("DATABASE_URL")
	DataDir = ensureEnv("DATA_DIR")
	EncryptionKey = sha256.Sum256([]byte(ensureEnv("PASSWORD_ENC_KEY")))
	JWTSecret = []byte(ensureEnv("JWT_SECRET"))
	Addr = defaultEnv("ADDR", ":8080")

	adminAddr := defaultEnv("ADMIN_ADDR", "unix;/var/run/hypersonic/admin.sock")

	var found bool
	AdminAddr.Network, AdminAddr.Address, found = strings.Cut(adminAddr, ";")
	if !found {
		AdminAddr.Address = AdminAddr.Network
		AdminAddr.Network = "tcp"
	}
}

func ensureEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		logrus.Fatalf("missing env variable: %s", key)
	}

	return value
}

func defaultEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	return value
}
