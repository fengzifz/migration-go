package main

import (
	"os"
	"testing"
)

var fileName = "create_dummy_table"
var dirPath string

func TestCreate(t *testing.T) {
	path, err := Create(fileName)
	if err != nil {
		t.Fatalf("Create migration failed: %v", err)
	}

	dirPath = path
}

func TestMigrate(t *testing.T) {
	err := Migrate()
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
}

func TestRollback(t *testing.T) {
	err := Rollback("1")
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// 测试结束后，删除测试所创建的文件
	defer func(dir string) {
		os.RemoveAll(dir)
	}(dirPath)
}
