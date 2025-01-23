package confstore

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Provider interface {
	IsValid(path string) bool
	Load(path string, value any) error
	Save(path string, value any) error
}

type LocalProvider struct {
	codec Codec
}

func NewLocalProvider(codec Codec) *LocalProvider {
	return &LocalProvider{codec: codec}
}

func (p *LocalProvider) IsValid(path string) bool {
	return filepath.IsLocal(path)
}

func (p *LocalProvider) Load(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return p.codec.Unmarshal(data, value)
}

func (p *LocalProvider) Save(path string, value any) error {
	data, err := p.codec.Marshal(value)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

type HttpProvider struct {
	codec Codec
}

func NewHttpProvider(codec Codec) *HttpProvider {
	return &HttpProvider{codec: codec}
}

func (p *HttpProvider) IsValid(path string) bool {
	return strings.HasPrefix(path, "http")
}

func (p *HttpProvider) Load(path string, value any) error {
	resp, err := http.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to load config from %s", path)
}

func (g *ProviderGroup) Save(path string, value any) error {
	for _, provider := range g.providers {
		if !provider.IsValid(path) {
			continue
		}
		err := provider.Save(path, value)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to save config to %s", path)
}
