package config

import (
	"os"
	"testing"
)

// printHelp 测试
func TestPrintHelp(t *testing.T) {
	// PrintHelp 只是打印到 stdout，测试它不会 panic
	PrintHelp() // 如果 panic，测试会失败
}

// 版本常量测试
func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("version should not be empty")
	}
}

// ListenAddress 测试
func TestListenAddress(t *testing.T) {
	cfg := &Config{Port: "9090", Addr: ""}
	addr := cfg.ListenAddress()
	if addr != ":9090" {
		t.Errorf("expected ':9090', got %s", addr)
	}

	cfg = &Config{Port: "8080", Addr: "127.0.0.1"}
	addr = cfg.ListenAddress()
	if addr != "127.0.0.1:8080" {
		t.Errorf("expected '127.0.0.1:8080', got %s", addr)
	}
}

// Parse 测试
func TestParse(t *testing.T) {
	// 测试默认配置
	cfg := Parse()
	if cfg.Port != "9090" {
		t.Errorf("expected default port 9090, got %s", cfg.Port)
	}
}

// PrintVersion 测试
func TestPrintVersion(t *testing.T) {
	PrintVersion()
}

func TestParse_CustomPort(t *testing.T) {
	os.Args = []string{"app", "-p", "8080"}
	cfg := Parse()
	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
}

func TestParse_CustomAddr(t *testing.T) {
	os.Args = []string{"app", "-addr", "127.0.0.1"}
	cfg := Parse()
	if cfg.Addr != "127.0.0.1" {
		t.Errorf("expected addr 127.0.0.1, got %s", cfg.Addr)
	}
}

func TestParse_CombinedFlags(t *testing.T) {
	os.Args = []string{"app", "-p", "3000", "-addr", "localhost"}
	cfg := Parse()
	if cfg.Port != "3000" {
		t.Errorf("expected port 3000, got %s", cfg.Port)
	}
	if cfg.Addr != "localhost" {
		t.Errorf("expected addr localhost, got %s", cfg.Addr)
	}
}

func TestParse_HelpFlag(t *testing.T) {
	os.Args = []string{"app", "-h"}
	cfg := Parse()
	if !cfg.Help {
		t.Error("expected Help to be true")
	}
}

func TestParse_VersionFlag(t *testing.T) {
	os.Args = []string{"app", "-v"}
	cfg := Parse()
	if !cfg.Version {
		t.Error("expected Version to be true")
	}
}

func TestParse_HelpLongFlag(t *testing.T) {
	os.Args = []string{"app", "-help"}
	cfg := Parse()
	if !cfg.Help {
		t.Error("expected Help to be true for -help flag")
	}
}

func TestParse_VersionLongFlag(t *testing.T) {
	os.Args = []string{"app", "-version"}
	cfg := Parse()
	if !cfg.Version {
		t.Error("expected Version to be true for -version flag")
	}
}

func TestParse_PortWithEquals(t *testing.T) {
	os.Args = []string{"app", "-p=9091"}
	cfg := Parse()
	if cfg.Port != "9091" {
		t.Errorf("expected port 9091, got %s", cfg.Port)
	}
}

func TestParse_AddrWithEquals(t *testing.T) {
	os.Args = []string{"app", "-addr=192.168.1.1"}
	cfg := Parse()
	if cfg.Addr != "192.168.1.1" {
		t.Errorf("expected addr 192.168.1.1, got %s", cfg.Addr)
	}
}

func TestListenAddress_WithCustomAddr(t *testing.T) {
	cfg := &Config{Port: "9090", Addr: "0.0.0.0"}
	addr := cfg.ListenAddress()
	if addr != "0.0.0.0:9090" {
		t.Errorf("expected '0.0.0.0:9090', got %s", addr)
	}
}

func TestListenAddress_EmptyAddr(t *testing.T) {
	cfg := &Config{Port: "9090", Addr: ""}
	addr := cfg.ListenAddress()
	if addr != ":9090" {
		t.Errorf("expected ':9090', got %s", addr)
	}
}

func TestVersion_Value(t *testing.T) {
	if Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", Version)
	}
}
