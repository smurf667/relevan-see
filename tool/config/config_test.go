package config

import (
	"io/fs"
	"os"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	config, err := Load("nowhere")
	if err != nil {
		t.Fatal("Default loading failed.")
	}
	if config.Message != " [skip ci]" {
		t.Fatalf("Unexpected default message is configured %s", config.Message)
	}
}

func TestLoadCustom(t *testing.T) {
	config, err := Load(".")
	if err != nil {
		t.Fatal("Custom loading failed.")
	}
	if config.Message != "hello" {
		t.Fatalf("Unexpected message is configured %s", config.Message)
	}
}

func TestDumpConfig(t *testing.T) {
	if os.Mkdir("temp", fs.ModeDir) != nil {
		t.Fatal("Cannot create temporary folder")
	}
	defer os.RemoveAll("temp")
	DumpConfig("temp")
	if _, err := os.Stat("temp/" + configFileName); err != nil {
		t.Fatal("Dumped file does not exit")
	}
}
