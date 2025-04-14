package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
)

//go:embed config.json
var defaultConfig []byte

type Config struct {
	Paths struct {
		Templates string `json:"templates"`
		Root      string `json:"root"`
		Build     string `json:"build"`
	} `json:"paths"`
	DefaultTemplates struct {
		Notes string `json:"notes"`
		Pages string `json:"pages"`
		Site  string `json:"site"`
	} `json:"default templates"`
	Scripts map[string]string `json:"scripts"`
	Emojis []struct {
		Path string `json:"path"`
		Name string `json:"name"`
	} `json:"emojis"`
}
func (cfg Config) String() string {
    var result string
    result = fmt.Sprintf("Paths:\n\tTemplates Directory: %s", cfg.Paths.Templates)
    result = fmt.Sprintf("%s\n\tRoot Directory: %s", result, cfg.Paths.Root)
    result = fmt.Sprintf("%s\n\tBuild Directory: %s", result, cfg.Paths.Build)
    result = fmt.Sprintf("%s\nDefault Templates:\n\tNotes: %s", result, cfg.DefaultTemplates.Notes)
    result = fmt.Sprintf("%s\n\tPages: %s", result, cfg.DefaultTemplates.Pages)
    result = fmt.Sprintf("%s\n\tSite: %s", result, cfg.DefaultTemplates.Site)
    result = fmt.Sprintf("%s\nEmojis:", result)
    for _, emoji := range cfg.Emojis {
        result = fmt.Sprintf("%s\n\t%-10s %s",
            result, emoji.Name, emoji.Path)
    }
    return result
}

func ReadConfig(data []byte) (cfg Config, err error) {
    err = json.Unmarshal(data, &cfg)
    return
}

func DefaultConfig() (cfg Config) {
    err := json.Unmarshal(defaultConfig, &cfg)
    if err != nil {
        log.Fatal(err)
    }
    return
}
