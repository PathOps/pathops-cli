package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

type Profile struct {
	ControlPlaneBaseURL string `json:"controlPlaneBaseUrl"`
	Issuer              string `json:"issuer"`
	ClientID            string `json:"clientId"`
}

type Config struct {
	ActiveProfile string             `json:"activeProfile"`
	Profiles      map[string]Profile `json:"profiles"`
}

func DefaultConfig() Config {
	return Config{
		ActiveProfile: "default",
		Profiles: map[string]Profile{
			"default": {},
		},
	}
}

func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "pathops"), nil
}

func ConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func EnsureDir() error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o700)
}

func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig()
		if err := Save(cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Save(cfg Config) error {
	if err := EnsureDir(); err != nil {
		return err
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func Active(cfg Config) (Profile, error) {
	p, ok := cfg.Profiles[cfg.ActiveProfile]
	if !ok {
		return Profile{}, errors.New("active profile not found")
	}
	return p, nil
}

func OSName() string {
	return runtime.GOOS
}

func SaveActiveProfile(cfg Config, profile Profile) error {
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}

	active := cfg.ActiveProfile
	if active == "" {
		active = "default"
		cfg.ActiveProfile = active
	}

	cfg.Profiles[active] = profile
	return Save(cfg)
}
