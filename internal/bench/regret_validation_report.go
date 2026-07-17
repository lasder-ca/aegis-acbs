package bench

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
)

//go:embed regret_validation_template.html
var regretValidationTemplate string

func WriteRegretValidationHTML(path string, report RegretValidationReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	payload := base64.StdEncoding.EncodeToString(data)
	html := strings.Replace(regretValidationTemplate, "__AEGIS_REGRET_VALIDATION_BASE64__", payload, 1)
	return os.WriteFile(path, []byte(html), 0o644)
}
