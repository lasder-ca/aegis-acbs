package bench

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProfileSchedulerTriggersBuildsCheckpointRules(t *testing.T) {
	g := replayGridGraph(t, 20, 20)
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "seed-9"), 0o755); err != nil {
		t.Fatal(err)
	}
	report := Report{
		Version: "test", GraphName: g.Name, Metric: g.Metric, AllCorrect: true,
		Config: Config{Seed: 9},
		Queries: []Query{
			{Source: 0, Target: len(g.Nodes) - 1, Class: "regional"},
			{Source: 0, Target: 2, Class: "local"},
		},
	}
	writeJSONFile(t, filepath.Join(dir, "seed-9", "report.json"), report)
	validation := RegretValidationReport{
		Version: "test", Runs: []RegretValidationRun{{Path: "seed-9/report.json", Seed: 9, Queries: 2, AllCorrect: true}},
		TotalQueries: 2, AllCorrect: true,
	}
	validationPath := filepath.Join(dir, "regret-validation.json")
	writeJSONFile(t, validationPath, validation)
	replay := RegretReplayReport{
		Version: "test", Cases: []RegretReplayCase{{SourceReport: "seed-9/report.json", QueryIndex: 0, Classification: "adaptive-scheduler-tail"}},
		AllCorrect: true,
	}
	replayPath := filepath.Join(dir, "regret-replay.json")
	writeJSONFile(t, replayPath, replay)

	got, err := ProfileSchedulerTriggers(context.Background(), g, validationPath, replayPath, dir, TriggerProfileConfig{
		Checkpoints: []uint64{1, 2}, Timeout: 5 * time.Second, MaxMatches: 2, TopRules: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Queries != 2 || got.SchedulerTails != 1 || !got.AllCorrect || len(got.Rows) != 2 {
		t.Fatalf("unexpected profile summary: %+v", got)
	}
	if len(got.Rules) == 0 {
		t.Fatal("expected ranked trigger rules")
	}
	if len(got.Rows[0].Checkpoints) != 2 || !got.Rows[0].Checkpoints[0].Reached {
		t.Fatalf("missing checkpoint features: %+v", got.Rows[0])
	}
	if err := WriteTriggerProfileJSON(filepath.Join(dir, "profile.json"), got); err != nil {
		t.Fatal(err)
	}
	if err := WriteTriggerProfileCSV(filepath.Join(dir, "profile.csv"), got); err != nil {
		t.Fatal(err)
	}
	if err := WriteTriggerProfileHTML(filepath.Join(dir, "profile.html"), got); err != nil {
		t.Fatal(err)
	}
}

func writeJSONFile(t testing.TB, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
