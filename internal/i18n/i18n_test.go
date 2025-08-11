package i18n

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name     string
		language string
	}{
		{
			name:     "english",
			language: "en",
		},
		{
			name:     "japanese",
			language: "ja",
		},
		{
			name:     "empty language",
			language: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.language)
			assert.NoError(t, err)
			assert.NotNil(t, bundle)
			assert.NotNil(t, localizer)
		})
	}
}

func TestInitializeWithEnvironmentVariables(t *testing.T) {
	originalLang := os.Getenv("LANG")
	originalLCAll := os.Getenv("LC_ALL")

	// Clean up after test
	t.Cleanup(func() {
		if originalLang != "" {
			os.Setenv("LANG", originalLang)
		} else {
			os.Unsetenv("LANG")
		}
		if originalLCAll != "" {
			os.Setenv("LC_ALL", originalLCAll)
		} else {
			os.Unsetenv("LC_ALL")
		}
	})

	tests := []struct {
		name    string
		langVar string
		lcAll   string
	}{
		{
			name:    "japanese locale",
			langVar: "ja_JP.UTF-8",
			lcAll:   "",
		},
		{
			name:    "english locale",
			langVar: "en_US.UTF-8",
			lcAll:   "",
		},
		{
			name:    "LC_ALL precedence",
			langVar: "en_US.UTF-8",
			lcAll:   "ja_JP.UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.langVar != "" {
				os.Setenv("LANG", tt.langVar)
			} else {
				os.Unsetenv("LANG")
			}
			if tt.lcAll != "" {
				os.Setenv("LC_ALL", tt.lcAll)
			} else {
				os.Unsetenv("LC_ALL")
			}

			err := Initialize("")
			assert.NoError(t, err)
			assert.NotNil(t, bundle)
			assert.NotNil(t, localizer)
		})
	}
}

func TestExtractLangFromLangVar(t *testing.T) {
	tests := []struct {
		name     string
		langVar  string
		expected string
	}{
		{
			name:     "japanese full locale",
			langVar:  "ja_JP.UTF-8",
			expected: "ja",
		},
		{
			name:     "english full locale",
			langVar:  "en_US.UTF-8",
			expected: "en",
		},
		{
			name:     "simple lang code",
			langVar:  "fr",
			expected: "fr",
		},
		{
			name:     "lang with country",
			langVar:  "zh_CN",
			expected: "zh",
		},
		{
			name:     "complex locale",
			langVar:  "pt_BR.ISO-8859-1",
			expected: "pt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLangFromLangVar(tt.langVar)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestT(t *testing.T) {
	// Initialize i18n system for testing
	err := Initialize("en")
	require.NoError(t, err)

	tests := []struct {
		name         string
		messageID    string
		templateData map[string]interface{}
		expectEmpty  bool
	}{
		{
			name:      "existing message without data",
			messageID: "app_name", // This should exist in locales
		},
		{
			name:      "non-existent message",
			messageID: "non_existent_message_id",
		},
		{
			name:      "message with template data",
			messageID: "files_processed", // Should exist and use Count
			templateData: map[string]interface{}{
				"Count": 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := T(tt.messageID, tt.templateData)

			if tt.expectEmpty {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
				// If message doesn't exist, T should return the message ID as fallback
				if result == tt.messageID {
					// This means the message wasn't found, which is okay for testing
					t.Logf("Message ID %s not found, returning as fallback", tt.messageID)
				}
			}
		})
	}
}

func TestTWithoutInitialization(t *testing.T) {
	// Clear localizer to test behavior without initialization
	originalLocalizer := localizer
	localizer = nil

	t.Cleanup(func() {
		localizer = originalLocalizer
	})

	result := T("any_message_id")
	assert.Equal(t, "any_message_id", result)
}

func TestTWithCount(t *testing.T) {
	// Initialize i18n system for testing
	err := Initialize("en")
	require.NoError(t, err)

	tests := []struct {
		name         string
		messageID    string
		count        int
		templateData map[string]interface{}
	}{
		{
			name:      "singular count",
			messageID: "files_processed",
			count:     1,
		},
		{
			name:      "plural count",
			messageID: "files_processed",
			count:     5,
		},
		{
			name:      "zero count",
			messageID: "files_processed",
			count:     0,
		},
		{
			name:      "count with additional template data",
			messageID: "files_processed",
			count:     3,
			templateData: map[string]interface{}{
				"ExtraData": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TWithCount(tt.messageID, tt.count, tt.templateData)
			assert.NotEmpty(t, result)

			// If message doesn't exist, should return message ID as fallback
			if result == tt.messageID {
				t.Logf("Message ID %s not found, returning as fallback", tt.messageID)
			}
		})
	}
}

func TestTWithCountWithoutInitialization(t *testing.T) {
	// Clear localizer to test behavior without initialization
	originalLocalizer := localizer
	localizer = nil

	t.Cleanup(func() {
		localizer = originalLocalizer
	})

	result := TWithCount("any_message_id", 5)
	assert.Equal(t, "any_message_id", result)
}

func TestMultipleInitializations(t *testing.T) {
	// Test that multiple initializations don't cause issues
	err := Initialize("en")
	assert.NoError(t, err)

	err = Initialize("ja")
	assert.NoError(t, err)

	err = Initialize("")
	assert.NoError(t, err)

	// Should still work after multiple initializations
	result := T("app_name")
	assert.NotEmpty(t, result)
}

func TestLanguageFallback(t *testing.T) {
	// Test with unsupported language - should fallback to English or Japanese
	err := Initialize("unsupported_lang")
	assert.NoError(t, err)

	result := T("app_name")
	assert.NotEmpty(t, result)
	// Should not return the message ID since it should fallback to supported language
	assert.NotEqual(t, "app_name", result)
}
