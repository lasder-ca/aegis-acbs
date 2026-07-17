package bench

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
)

//go:embed report_template.html
var reportTemplate string

// WriteHTML creates a self-contained, offline benchmark dashboard.
func WriteHTML(path string, report Report) error {
	b, err := json.Marshal(report)
	if err != nil {
		return err
	}
	payload := base64.StdEncoding.EncodeToString(b)
	html := strings.Replace(reportTemplate, "__AEGIS_REPORT_BASE64__", payload, 1)
	return os.WriteFile(path, []byte(html), 0644)
}
