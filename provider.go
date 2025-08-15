package confstore

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Provider interface {
	IsValid(path string) bool
	Load(path string, value any) error
	Save(path string, value any) error
}

func isRemoteURL(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	s := strings.ToLower(u.Scheme)
	return (s == "http" || s == "https") && u.Host != ""
}

func isLocalPath(path string) bool {
	if path == "" {
		return false
	}
	if filepath.IsAbs(path) {
		return true
	}
	if u, err := url.Parse(path); err == nil && u.Scheme != "" {
		return u.Scheme == "file"
	}
	return true
}

type LocalProvider struct {
	codec Codec
}

func NewLocalProvider(codec Codec) *LocalProvider {
	return &LocalProvider{codec: codec}
}

func (p *LocalProvider) IsValid(path string) bool {
	return isLocalPath(path)
}

func (p *LocalProvider) Load(path string, value any) error {
	if u, err := url.Parse(path); err == nil && u.Scheme == "file" {
		path = u.Path
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return p.codec.Unmarshal(data, value)
}

func (p *LocalProvider) Save(path string, value any) error {
	if u, err := url.Parse(path); err == nil && u.Scheme == "file" {
		path = u.Path
	}
	data, err := p.codec.Marshal(value)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return mkErr
		}
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()
	_, err = file.Write(data)
	return err
}

type HttpProvider struct {
	codec  Codec
	client *http.Client
}

type HttpProviderOption func(*HttpProvider)

func WithHTTPClient(client *http.Client) HttpProviderOption {
	return func(p *HttpProvider) {
		if client != nil {
			p.client = client
		}
	}
}

func NewHttpProvider(codec Codec, options ...HttpProviderOption) *HttpProvider {
	provider := &HttpProvider{
		codec:  codec,
		client: nil,
	}
	for _, opt := range options {
		opt(provider)
	}
	if provider.client == nil {
		provider.client = &http.Client{Timeout: 30 * time.Second}
	}
	return provider
}

func (p *HttpProvider) IsValid(path string) bool {
	return isRemoteURL(path)
}

func (p *HttpProvider) Load(path string, value any) error {
	resp, err := p.client.Get(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(data))
	}
	return p.codec.Unmarshal(data, value)
}

func (p *HttpProvider) Save(path string, value any) error {
	data, err := p.codec.Marshal(value)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	//req.Header.Set("Content-Type", "application/binary")
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

type ProviderGroup struct {
	providers []Provider
}

func NewProviderGroup(providers ...Provider) *ProviderGroup {
	return &ProviderGroup{providers: providers}
}

func (g *ProviderGroup) IsValid(path string) bool {
	for _, provider := range g.providers {
		if provider.IsValid(path) {
			return true
		}
	}
	return false
}

func (g *ProviderGroup) Load(path string, value any) error {
	for _, provider := range g.providers {
		if !provider.IsValid(path) {
			continue
		}
		err := provider.Load(path, value)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("provider %s not found", path)
}

func (g *ProviderGroup) Save(path string, value any) error {
	for _, provider := range g.providers {
		if !provider.IsValid(path) {
			continue
		}
		err := provider.Save(path, value)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("provider %s not found", path)
}
