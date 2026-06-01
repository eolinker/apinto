package pricing

import (
	"testing"
)

func TestLocalTaskExecutor_SubmitAndGet(t *testing.T) {
	exec := NewLocalTaskExecutor(0)
	defer exec.Stop()

	pre := &CalculateResult{TotalCost: 1.5}
	if err := exec.Submit("t-1", "openai", "gpt-4", PricingPhaseSubmit, pre, map[string]interface{}{"k": "v"}); err != nil {
		t.Fatalf("submit err: %v", err)
	}
	entry, ok := exec.Get("t-1")
	if !ok {
		t.Fatalf("expect entry exists")
	}
	if entry.Provider != "openai" || entry.Resource != "gpt-4" {
		t.Fatalf("entry meta mismatch: %+v", entry)
	}
	if entry.Status != TaskStatusPending {
		t.Fatalf("initial status want pending, got %s", entry.Status)
	}
	if entry.PreResult == nil || entry.PreResult.TotalCost != 1.5 {
		t.Fatalf("pre result mismatch: %+v", entry.PreResult)
	}
}

func TestLocalTaskExecutor_SubmitIdempotent(t *testing.T) {
	exec := NewLocalTaskExecutor(0)
	defer exec.Stop()

	_ = exec.Submit("t-1", "p", "r", PricingPhaseSubmit, &CalculateResult{TotalCost: 1}, nil)
	_ = exec.Submit("t-1", "p", "r", PricingPhaseSubmit, &CalculateResult{TotalCost: 99}, nil)
	entry, _ := exec.Get("t-1")
	if entry.PreResult.TotalCost != 1 {
		t.Fatalf("submit should be idempotent, got %.2f", entry.PreResult.TotalCost)
	}
}

func TestLocalTaskExecutor_MarkBilled(t *testing.T) {
	exec := NewLocalTaskExecutor(0)
	defer exec.Stop()

	_ = exec.Submit("t-1", "p", "r", PricingPhaseQuery, nil, nil)

	if exec.IsBilled("t-1") {
		t.Fatalf("freshly submitted task should not be billed")
	}
	if !exec.MarkBilled("t-1") {
		t.Fatalf("first MarkBilled should succeed")
	}
	if !exec.IsBilled("t-1") {
		t.Fatalf("after mark, IsBilled should be true")
	}
	if exec.MarkBilled("t-1") {
		t.Fatalf("second MarkBilled should return false")
	}
	if exec.MarkBilled("t-not-exists") {
		t.Fatalf("MarkBilled on missing task should return false")
	}
}

func TestLocalTaskExecutor_MarkFailed(t *testing.T) {
	exec := NewLocalTaskExecutor(0)
	defer exec.Stop()

	_ = exec.Submit("t-1", "p", "r", PricingPhaseSubmit, nil, nil)
	if !exec.MarkFailed("t-1") {
		t.Fatalf("MarkFailed should succeed")
	}
	entry, _ := exec.Get("t-1")
	if entry.Status != TaskStatusFailed {
		t.Fatalf("status want failed, got %s", entry.Status)
	}
}

func TestLocalTaskExecutor_Remove(t *testing.T) {
	exec := NewLocalTaskExecutor(0)
	defer exec.Stop()

	_ = exec.Submit("t-1", "p", "r", PricingPhaseSubmit, nil, nil)
	exec.Remove("t-1")
	if _, ok := exec.Get("t-1"); ok {
		t.Fatalf("entry should be removed")
	}
}

func TestExtractTaskID(t *testing.T) {
	cases := []struct {
		name   string
		fields map[string]interface{}
		want   string
	}{
		{"task_id first", map[string]interface{}{"task_id": "T1", "id": "I1"}, "T1"},
		{"id second", map[string]interface{}{"id": "I1", "request_id": "R1"}, "I1"},
		{"request_id last", map[string]interface{}{"request_id": "R1"}, "R1"},
		{"empty", map[string]interface{}{}, ""},
		{"non-string skipped", map[string]interface{}{"task_id": 123, "id": "I1"}, "I1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ExtractTaskID(c.fields); got != c.want {
				t.Fatalf("want %q, got %q", c.want, got)
			}
		})
	}
}

func TestMarshalUnmarshalTaskEntry(t *testing.T) {
	entry := &TaskEntry{
		TaskID:        "T-1",
		Provider:      "p",
		Resource:      "r",
		Phase:         PricingPhaseSubmit,
		Status:        TaskStatusPending,
		FreezeID:      "fz-1",
		EstimatedCost: 1.23,
		AccountID:     "acc-1",
	}
	data, err := MarshalTaskEntry(entry)
	if err != nil {
		t.Fatalf("marshal err: %v", err)
	}
	got, err := UnmarshalTaskEntry(data)
	if err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}
	if got.TaskID != entry.TaskID || got.FreezeID != entry.FreezeID || got.EstimatedCost != 1.23 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestTaskKeys_BillingPrefix(t *testing.T) {
	if got := TaskCacheKey("xyz"); got != "apinto:billing:task:xyz" {
		t.Fatalf("TaskCacheKey want apinto:billing:task:xyz, got %s", got)
	}
	if got := TaskBilledKey("xyz"); got != "apinto:billing:task-billed:xyz" {
		t.Fatalf("TaskBilledKey want apinto:billing:task-billed:xyz, got %s", got)
	}
}
