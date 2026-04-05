package i18n

import (
	"strings"
	"sync"
)

var (
	translations map[string]string
	currentLang  = "zh"
	mu           sync.RWMutex
)

func init() {
	translations = zhMessages
}

func T(key string) string {
	mu.RLock()
	defer mu.RUnlock()
	if msg, ok := translations[key]; ok {
		return msg
	}
	return key
}

func SetLanguage(lang string) {
	mu.Lock()
	defer mu.Unlock()
	lang = strings.ToLower(lang)
	if lang == "en" || lang == "en-US" || lang == "en-GB" {
		currentLang = "en"
		translations = enMessages
	} else {
		currentLang = "zh"
		translations = zhMessages
	}
}

func GetLanguage() string {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}