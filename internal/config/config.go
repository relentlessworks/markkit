package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

// Config holds all service configuration.
type Config struct {
	Addr   string
	Secret string
}

// Load reads configuration from defaults < env < flags.
func Load() Config {
	c := Config{
		Addr:   ":8470",
		Secret: "",
	}

	// Env
	if v := os.Getenv("MARKKIT_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("MARKKIT_SECRET"); v != "" {
		c.Secret = v
	}

	// Flags
	flag.StringVar(&c.Addr, "addr", c.Addr, "listen address")
	flag.StringVar(&c.Secret, "secret", c.Secret, "auth token signing secret (auto-generated if empty)")
	flag.Parse()

	// Auto-generate secret if not provided
	if c.Secret == "" {
		c.Secret = generateSecret()
	}

	return c
}

func generateSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("markkit-default-secret-%d", os.Getpid())
	}
	return hex.EncodeToString(b)
}
