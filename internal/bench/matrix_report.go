package bench

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
)

//go:embed matrix_template.html
var matrixTemplate string

func WriteMatrixHTML(path string, report MatrixReport) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	payload := base64.StdEncoding.EncodeToString(data)
	html := strings.Replace(matrixTemplate, "__AEGIS_MATRIX_BASE64__", payload, 1)
	return os.WriteFile(path, []byte(html), 0o644)
}
