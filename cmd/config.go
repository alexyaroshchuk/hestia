package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"hestia/pkg/auth"
)

// httpConfig is the configuration for the HTTP server.
type httpConfig struct {
	addr            string
	readTimeout     time.Duration
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	shutdownTimeout time.Duration
}

// dbConfig is the database configuration.
type dbConfig struct {
	connection string
	migrate    bool
}

// config is the configuration for the server command.
type config struct {
	http httpConfig
	db   dbConfig
	auth auth.JWTConfig
}

// defaultConfig returns a config with sane default values.
func defaultConfig() config {
	return config{
		http: httpConfig{
			addr:            ":8080",
			readTimeout:     time.Second * 5,
			writeTimeout:    time.Second * 10,
			idleTimeout:     time.Second * 120,
			shutdownTimeout: time.Second * 15,
		},
		db: dbConfig{
			connection: "host=localhost port=5432 user=test password=password dbname=hestia sslmode=disable",
			migrate:    false,
		},
		auth: auth.JWTConfig{
			SecretKey:     "secret",
			TokenDuration: time.Minute * 120,
		},
	}
}

type envVariable struct {
	required bool
	mapFunc  func(v string, c *config) error
}

// envMap maps environment variable names to fields in the config struct.
var envMap = map[string]envVariable{
	"HTTP_ADDR": {
		mapFunc: func(v string, c *config) error {
			c.http.addr = v
			return nil
		},
	},
	"HTTP_READ_TIMEOUT": {
		mapFunc: func(v string, c *config) error {
			return confDuration(v, &c.http.readTimeout, 0, math.MaxInt64)
		},
	},
	"HTTP_WRITE_TIMEOUT": {
		mapFunc: func(v string, c *config) error {
			return confDuration(v, &c.http.writeTimeout, 0, math.MaxInt64)
		},
	},
	"HTTP_IDLE_TIMEOUT": {
		mapFunc: func(v string, c *config) error {
			return confDuration(v, &c.http.idleTimeout, 0, math.MaxInt64)
		},
	},
	"HTTP_SHUTDOWN_TIMEOUT": {
		mapFunc: func(v string, c *config) error {
			return confDuration(v, &c.http.shutdownTimeout, 0, math.MaxInt64)
		},
	},
	"DB_CONNECTION": {
		mapFunc: func(v string, c *config) error {
			return confString(v, &c.db.connection, 1, math.MaxInt64)
		},
	},
	"DB_MIGRATE": {
		mapFunc: func(v string, c *config) error {
			return confBool(v, &c.db.migrate)
		},
	},
	"AUTH_SECRET_KEY": {
		mapFunc: func(v string, c *config) error {
			return confString(v, &c.auth.SecretKey, 1, math.MaxInt64)
		},
	},
	"AUTH_TOKEN_DURATION": {
		mapFunc: func(v string, c *config) error {
			return confDuration(v, &c.auth.TokenDuration, 0, math.MaxInt64)
		},
	},
}

// configFromEnv returns a config with values from the environment.
func configFromEnv() (config, error) {
	c := defaultConfig()

	var errSum error
	for key, envVar := range envMap {
		val, ok := os.LookupEnv(key)
		if !ok {
			if envVar.required {
				errSum = errors.Join(errSum, fmt.Errorf("missing required env variable %s", key))
			}
			continue
		}

		if err := envVar.mapFunc(val, &c); err != nil {
			errSum = errors.Join(errSum, fmt.Errorf("invalid env variable %s: %w", key, err))
		}
	}

	return c, errSum
}

// confDuration attempts to parse v into tgt and checks if the result is in
// the provided range (inclusive).
func confDuration(v string, tgt *time.Duration, min, max time.Duration) error {
	dur, err := time.ParseDuration(v)
	if err != nil {
		return err
	}

	if dur < min || dur > max {
		return fmt.Errorf("duration %s not in range [%s, %s] (inclusive)", dur, min, max)
	}

	*tgt = dur

	return nil
}

func confString(v string, tgt *string, minLen, maxLen int) error {
	if len(v) < minLen || len(v) > maxLen {
		return fmt.Errorf("string length %d not in range [%d, %d] (inclusive)", len(v), minLen, maxLen)
	}

	*tgt = v

	return nil
}

func confBool(v string, tgt *bool) error {
	b, err := strconv.ParseBool(v)
	if err != nil {
		return err
	}

	*tgt = b

	return nil
}
