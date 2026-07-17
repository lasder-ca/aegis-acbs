package i18n

import "testing"

func TestCatalogCompleteness(t *testing.T) {
	base := Catalog(EN)
	for _, lang := range Supported() {
		c := Catalog(lang)
		for k := range base {
			if c[k] == "" {
				t.Fatalf("missing %s in %s", k, lang)
			}
		}
	}
}
