package confstore

import (
	"testing"
)

func Test_isLocalPath(t *testing.T) {
	paths := map[string]bool{
		// 常见的本地路径
		"config.yaml":      true, // 相对路径
		"./config.yaml":    true, // 相对路径
		"../config.yaml":   true, // 相对路径
		"/etc/config.yaml": true, // Unix 绝对路径
		//"C:\\Users\\config.yaml":     true, // Windows 绝对路径
		`\\server\share\config.yaml`: true, // Windows UNC 路径 (url.Parse 可能解析 Scheme 为空)

		// 文件 URI
		"file:///etc/config.yaml":     true, // 标准文件 URI
		"file://C:/Users/config.yaml": true, // Windows 文件 URI

		// 常见的远程 URL
		"http://example.com/config.yaml":  false,
		"https://example.com/config.yaml": false,
		"ftp://example.com/config.yaml":   false,
		"s3://mybucket/config.yaml":       false,

		"/": true,
	}
	for path, expected := range paths {
		t.Run(path, func(t *testing.T) {
			result := isLocalPath(path)
			if result != expected {
				t.Errorf("[ %s ] expected %v, got %v", path, expected, result)
			}
		})
	}
}
