package business

import "github.com/pnnh/neutron/models"

func IsSupportedLanguage(lang string) bool {
	if lang == models.LangEn || lang == models.LangZh {
		return true
	}
	return false
}
