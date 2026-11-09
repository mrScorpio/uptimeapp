package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr          string
	DatabaseURL   string
	JWTSecret     string
	AllowedOrigin string
	CookieSecure  bool
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	AvatarDir     string
}

func Load() (Config, error) {
	c := Config{Addr: env("API_ADDR", ":8080"), DatabaseURL: os.Getenv("DATABASE_URL"), JWTSecret: os.Getenv("JWT_SECRET"), AllowedOrigin: env("ALLOWED_ORIGIN", "http://localhost:3000"), AccessTTL: 15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour, AvatarDir: env("AVATAR_DIR", "./avatars")}
	if c.DatabaseURL == "" || len(c.JWTSecret) < 32 {
		return c, errors.New("DATABASE_URL and JWT_SECRET (at least 32 characters) are required")
	}
	var err error
	if c.CookieSecure, err = strconv.ParseBool(env("COOKIE_SECURE", "false")); err != nil {
		return c, err
	}
	if c.AccessTTL, err = time.ParseDuration(env("ACCESS_TTL", "15m")); err != nil {
		return c, err
	}
	if c.RefreshTTL, err = time.ParseDuration(env("REFRESH_TTL", "720h")); err != nil {
		return c, err
	}
	if c.AccessTTL <= 0 || c.RefreshTTL <= 0 {
		return c, errors.New("token lifetimes must be positive")
	}
	return c, nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
