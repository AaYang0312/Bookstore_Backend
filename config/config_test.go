package config

import "testing"

func TestApplyEnvironmentOverrides(t *testing.T) {
	AppConfig = Config{
		Server:   ServerConfig{Host: "localhost", Port: 8080},
		Database: DatabaseConfig{Host: "127.0.0.1", Port: 3306, User: "root", Password: "local", Name: "bookstore"},
		Redis:    RedisConfig{Host: "127.0.0.1", Port: 6379, DB: 0},
	}

	t.Setenv("BOOKSTORE_SERVER_HOST", "0.0.0.0")
	t.Setenv("BOOKSTORE_SERVER_PORT", "18080")
	t.Setenv("BOOKSTORE_DATABASE_HOST", "mysql")
	t.Setenv("BOOKSTORE_DATABASE_PASSWORD", "docker-secret")
	t.Setenv("BOOKSTORE_REDIS_HOST", "redis")
	t.Setenv("BOOKSTORE_REDIS_DB", "2")

	applyEnvironmentOverrides()

	if AppConfig.Server.Host != "0.0.0.0" || AppConfig.Server.Port != 18080 {
		t.Fatalf("unexpected server config: %+v", AppConfig.Server)
	}
	if AppConfig.Database.Host != "mysql" || AppConfig.Database.Password != "docker-secret" {
		t.Fatalf("unexpected database config: %+v", AppConfig.Database)
	}
	if AppConfig.Redis.Host != "redis" || AppConfig.Redis.DB != 2 {
		t.Fatalf("unexpected redis config: %+v", AppConfig.Redis)
	}
}

func TestApplyEnvironmentOverridesUsesDefaultHost(t *testing.T) {
	AppConfig = Config{Server: ServerConfig{Port: 8080}}
	applyEnvironmentOverrides()

	if AppConfig.Server.Host != "localhost" {
		t.Fatalf("expected localhost, got %q", AppConfig.Server.Host)
	}
}
