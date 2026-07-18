package bench

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
)

//go:embed trigger_profile_template.html
var triggerProfileTemplate string

func WriteTriggerProfileHTML(path string, report TriggerProfileReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	payload := base64.StdEncoding.EncodeToString(data)
	html := strings.Replace(triggerProfileTemplate, "__AEGIS_TRIGGER_PROFILE_BASE64__", payload, 1)
	return os.WriteFile(path, []byte(html), 0o644)
}
