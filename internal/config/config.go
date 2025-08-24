package config

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"github.com/sirupsen/logrus"
)

var (
	DatabaseUrl      string
	MusicRoot        string
	DataDir          string
	JWTSecret        []byte
	SelfRegistration bool
	Production       bool
	Administrators   []string
	Addr             string
	Concurrency      int
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

	admins := os.Getenv("ADMINS")
	for userID := range strings.SplitSeq(admins, ",") {
		userID = strings.TrimSpace(userID)

		Administrators = append(Administrators, userID)
	}

	Production = os.Getenv("PRODUCTION") == "true"
	Addr, ok = os.LookupEnv("ADDR")
	if !ok {
		Addr = ":8080"
	}

	c, ok := os.LookupEnv("CONCURRENCY")
	if ok {
		Concurrency, err = strconv.Atoi(c)
		if err != nil {
			return fmt.Errorf("parse CONCURRENCY: %w", err)
		}
	} else {
		logrus.Warnf("CONCURRENCY is not set, using 2")
		Concurrency = 2
	}

	return nil
}
