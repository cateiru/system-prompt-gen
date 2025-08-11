package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupI18n(t *testing.T, inputLang ...string) {
	lang := "en"

	if len(inputLang) > 0 {
		lang = inputLang[0]
	}

	err := Initialize(lang)
	require.NoError(t, err)
}
