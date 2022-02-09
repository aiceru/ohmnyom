package i18n

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aiceru/protonyom/gonyom"
	"ohmnyom/internal/path"
)

var supportedLocales = map[string]struct{}{
	"en": {},
	"ko": {},
}

var SupportedFamilies map[string]map[string]*gonyom.Family

func init() {
	speciesJsonPath := filepath.Join(path.Root(), "assets", "translation-species.json")
	speciesJsonFile, err := os.Open(speciesJsonPath)
	if err != nil {
		log.Fatal(err)
	}
	defer speciesJsonFile.Close()

	speciesJsonBytes, err := io.ReadAll(speciesJsonFile)
	if err != nil {
		log.Fatal(err)
	}
	SupportedFamilies = make(map[string]map[string]*gonyom.Family)
	if err = json.Unmarshal(speciesJsonBytes, &SupportedFamilies); err != nil {
		log.Fatal(err)
	}
}

func SupportOrFallback(lo string) string {
	_, ok := supportedLocales[lo]
	if ok {
		return lo
	}
	return "en"
}
