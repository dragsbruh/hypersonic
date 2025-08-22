package config

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

var (
	DatabaseUrl      string
	MusicRoot        string
	DataDir          string
	JWTSecret        []byte
	SelfRegistration bool
)

func LoadConfig() error {
	var ok bool

	DatabaseUrl, ok = os.LookupEnv("DATABASE_URL")
	if !ok {
		return fmt.Errorf("DATABASE_URL missing")
	}

	MusicRoot, ok = os.LookupEnv("MUSIC_ROOT")
	if !ok {
		return fmt.Errorf("MUSIC_ROOT missing")
	}

	info, err := os.Stat(MusicRoot)
	if err != nil {
		return fmt.Errorf("check MUSIC_ROOT: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("MUSIC_ROOT is not a dir")
	}

	DataDir, ok = os.LookupEnv("DATA_DIR")
	if !ok {
		return fmt.Errorf("DATA_DIR missing")
	}

	if err := os.MkdirAll(DataDir, fs.ModePerm); err != nil {
		return fmt.Errorf("make DATA_DIR: %w", err)
	}

	jwtSecret, ok := os.LookupEnv("JWT_SECRET")
	if !ok {
		return fmt.Errorf("JWT_SECRET missing")
	}
	jwtBytes := sha256.Sum256([]byte(jwtSecret))
	JWTSecret = jwtBytes[:]

	SelfRegistration = os.Getenv("SELF_REGISTRATION") == "enabled"

	return nil
}
