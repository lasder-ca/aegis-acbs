package bench

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
)

//go:embed regret_template.html
var regretTemplate string

func WriteRegretHTML(path string, report RegretReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	payload := base64.StdEncoding.EncodeToString(data)
	html := strings.Replace(regretTemplate, "__AEGIS_REGRET_BASE64__", payload, 1)
	return os.WriteFile(path, []byte(html), 0o644)
}
