package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sgdevelopers29-afk/GoKit-Lite/config"
)

func BenchmarkGet(b *testing.B) {
	key := "BENCHMARK_KEY"
	val := "benchmark_value"
	os.Setenv(key, val)
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Get(key)
	}
}

func BenchmarkLoadEnv(b *testing.B) {
	tempDir := b.TempDir()
	path := filepath.Join(tempDir, ".env")
	content := "PORT=8080\nAPP_ENV=production\nDB_HOST=localhost\n"
	os.WriteFile(path, []byte(content), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Load(path)
	}
}

func BenchmarkRepeatedLookup(b *testing.B) {
	os.Setenv("K1", "V1")
	os.Setenv("K2", "V2")
	os.Setenv("K3", "V3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Get("K1")
		_ = config.Get("K2")
		_ = config.Get("K3")
	}
}
