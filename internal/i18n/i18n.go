package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/en.json
var enMessages []byte

//go:embed locales/ja.json
var jaMessages []byte

var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
)

// Initialize initializes the i18n system with the given language preference
func Initialize(lang string) error {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load embedded message files
	if err := loadMessages("en", enMessages); err != nil {
		return err
	}
	if err := loadMessages("ja", jaMessages); err != nil {
		return err
	}

	// Determine language from various sources
	var acceptLanguages []string
	if lang != "" {
		acceptLanguages = append(acceptLanguages, lang)
	}

	// Check environment variables
	if envLang := os.Getenv("LANG"); envLang != "" {
		acceptLanguages = append(acceptLanguages, extractLangFromLangVar(envLang))
	}
	if envLang := os.Getenv("LC_ALL"); envLang != "" {
		acceptLanguages = append(acceptLanguages, extractLangFromLangVar(envLang))
	}

	// Add fallbacks
	acceptLanguages = append(acceptLanguages, "ja", "en")

	localizer = i18n.NewLocalizer(bundle, acceptLanguages...)

	return nil
}

// loadMessages loads messages from embedded JSON data
func loadMessages(lang string, data []byte) error {
	fileName := fmt.Sprintf("%s.json", lang)
	msgFile, err := bundle.ParseMessageFileBytes(data, fileName)
	if err != nil {
		return err
	}
	return bundle.AddMessages(language.Make(lang), msgFile.Messages...)
}

// extractLangFromLangVar extracts language code from LANG environment variable
// e.g., "ja_JP.UTF-8" -> "ja"
func extractLangFromLangVar(langVar string) string {
	parts := strings.Split(langVar, ".")
	langCode := strings.Split(parts[0], "_")[0]
	return langCode
}

// T returns a localized message for the given message ID
func T(messageID string, templateData ...map[string]any) string {
	if localizer == nil {
		// Fallback if not initialized
		return messageID
	}

	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		// Return message ID as fallback
		return messageID
	}
	return msg
}

// TWithCount returns a localized message with count for plural handling
func TWithCount(messageID string, count int, templateData ...map[string]any) string {
	if localizer == nil {
		return messageID
	}

	data := map[string]any{"Count": count}
	if len(templateData) > 0 {
		for k, v := range templateData[0] {
			data[k] = v
		}
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
		PluralCount:  count,
	})
	if err != nil {
		return messageID
	}
	return msg
}
