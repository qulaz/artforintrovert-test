package config

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T, env map[string]string) func() {
	t.Helper()

	for key, value := range env {
		t.Setenv(key, value)
	}

	// teardown test
	return func() {
		once = sync.Once{}
		config = nil
	}
}

func TestGetConfig_Defaults(t *testing.T) {
	requiredVars := map[string]string{
		"DATABASE_NAME": "test_db",
		"DATABASE_DSN":  "mongodb://test@test:localhost:27017/?replicaSet=rs0",
	}

	defer setupTest(t, requiredVars)()

	c, err := GetConfig()
	require.NoError(t, err)

	assert.Equal(t, false, c.API.Debug)                    // default
	assert.Equal(t, "0.0.0.0", c.API.Host)                 // default
	assert.Equal(t, "8000", c.API.RestPort)                // default
	assert.Equal(t, "50051", c.API.GrpcPort)               // default
	assert.Equal(t, time.Minute*1, c.API.ProductsCacheTtl) // default

	assert.Equal(t, requiredVars["DATABASE_DSN"], c.Database.DSN)   // from env
	assert.Equal(t, requiredVars["DATABASE_NAME"], c.Database.Name) // from env

	assert.Equal(t, "", c.Tracing.ExporterAddress) // default
	assert.Equal(t, "", c.Tracing.ExporterPort)    // default

	assert.Equal(t, "", c.Sentry.DSN)    // default
	assert.Equal(t, "dev", c.Sentry.Env) // default
}

func TestGetConfig_RewriteDefaults(t *testing.T) {
	env := map[string]string{
		"DEBUG":                    "true",
		"HOST":                     "localhost",
		"REST_PORT":                "5555",
		"GRPC_PORT":                "6666",
		"PRODUCTS_CACHE_TTL":       "12h",
		"DATABASE_DSN":             "mongodb://test@test:localhost:27017/?replicaSet=rs0",
		"DATABASE_NAME":            "test",
		"TRACING_EXPORTER_ADDRESS": "localhost",
		"TRACING_EXPORTER_PORT":    "16686",
		"SENTRY_DSN":               "https://sentry.com/test",
		"SENTRY_ENV":               "stage",
	}

	defer setupTest(t, env)()

	c, err := GetConfig()
	require.NoError(t, err)

	assert.Equal(t, true, c.API.Debug)
	assert.Equal(t, env["HOST"], c.API.Host)
	assert.Equal(t, env["GRPC_PORT"], c.API.GrpcPort)
	assert.Equal(t, env["REST_PORT"], c.API.RestPort)
	assert.Equal(t, time.Hour*12, c.API.ProductsCacheTtl)

	assert.Equal(t, env["DATABASE_DSN"], c.Database.DSN)
	assert.Equal(t, env["DATABASE_NAME"], c.Database.Name)

	assert.Equal(t, env["TRACING_EXPORTER_ADDRESS"], c.Tracing.ExporterAddress)
	assert.Equal(t, env["TRACING_EXPORTER_PORT"], c.Tracing.ExporterPort)

	assert.Equal(t, env["SENTRY_DSN"], c.Sentry.DSN)
	assert.Equal(t, env["SENTRY_ENV"], c.Sentry.Env)
}

func TestGetConfig_MissingRequired(t *testing.T) {
	defer setupTest(t, map[string]string{})()

	c, err := GetConfig()
	require.Error(t, err)
	require.Nil(t, c)
}
