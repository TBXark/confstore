package confstore

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
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

func Test_isRemoteURL_CaseAndHost(t *testing.T) {
	cases := map[string]bool{
		"HTTP://example.com/a": true,
		"https://EXAMPLE.com":  true,
		"http://":              false,
		"https://":             false,
		"HtTp://example.com":   true,
	}
	for in, want := range cases {
		if got := isRemoteURL(in); got != want {
			t.Fatalf("isRemoteURL(%q)=%v, want %v", in, got, want)
		}
	}
}

func Test_isLocalPath_EmptyFalse(t *testing.T) {
	if isLocalPath("") {
		t.Fatalf("empty path should not be local")
	}
}

type sample struct {
	Name string `json:"name"`
}

func Test_HttpProvider_Load_StatusAndUnmarshal(t *testing.T) {
	good := sample{Name: "ok"}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(good)
	}))
	defer ts.Close()

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer bad.Close()

	p := NewHttpProvider(JsonCodec{})

	var got sample
	if err := p.Load(ts.URL, &got); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != good.Name {
		t.Fatalf("unexpected payload: %+v", got)
	}

	if err := p.Load(bad.URL, &got); err == nil {
		t.Fatalf("expected error on non-2xx status")
	}
}

func Test_HttpProvider_Save_StatusAndContentType(t *testing.T) {
	var contentType atomic.Value
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType.Store(r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "reject", http.StatusBadRequest)
	}))
	defer bad.Close()

	p := NewHttpProvider(JsonCodec{})
	payload := sample{Name: "post"}

	if err := p.Save(ts.URL, &payload); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := p.Save(bad.URL, &payload); err == nil {
		t.Fatalf("expected error on non-2xx status")
	}
}

func Test_LocalProvider_FileURI_Load_and_Save(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "in.json")
	if err := os.WriteFile(srcPath, []byte("{\n  \"name\": \"alice\"\n}"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	fileURL := "file://" + srcPath

	var got struct {
		Name string `json:"name"`
	}
	lp := NewLocalProvider(JsonCodec{})
	if err := lp.Load(fileURL, &got); err != nil {
		t.Fatalf("load via file URI: %v", err)
	}
	if got.Name != "alice" {
		t.Fatalf("unexpected value: %+v", got)
	}

	nested := filepath.Join(dir, "a", "b", "out.json")
	outURL := "file://" + nested
	want := struct {
		Name string `json:"name"`
	}{Name: "bob"}
	if err := lp.Save(outURL, &want); err != nil {
		t.Fatalf("save via file URI: %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Fatalf("expected output file created: %v", err)
	}
}
