package main

import (
	"os"
	"testing"
)

var fileName = "create_dummy_table"

func TestCreate(t *testing.T) {
	path, err := Create(fileName)
	if err != nil {
		t.Fatalf("Create migration failed: %v", err)
	}

	// 测试结束后，删除测试所创建的文件
	defer func(dir string) {
		os.RemoveAll(dir)
	}(path)
}
