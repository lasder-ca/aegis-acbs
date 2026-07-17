package bench

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
)

//go:embed replay_template.html
var regretReplayTemplate string

func WriteRegretReplayHTML(path string, report RegretReplayReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	payload := base64.StdEncoding.EncodeToString(data)
	html := strings.Replace(regretReplayTemplate, "__AEGIS_REGRET_REPLAY_BASE64__", payload, 1)
	return os.WriteFile(path, []byte(html), 0o644)
}
