package configs

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed language_extensions.json
var languageExtensionsJSON []byte

type languageExtension struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Extensions []string `json:"extensions"`
}

var languageExtensions map[string][]string

func init() {
	var languages []languageExtension
	err := json.Unmarshal(languageExtensionsJSON, &languages)
	if err != nil {
		panic(err)
	}

	languageExtensions = make(map[string][]string)
	for _, lang := range languages {
		languageExtensions[strings.ToLower(lang.Name)] = lang.Extensions
	}
}

func GetLanguageExtensions() map[string][]string {
	return languageExtensions
}
