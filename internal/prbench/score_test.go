package prbench

import "testing"

func TestScore_PrecisionRecall(t *testing.T) {
	gold := []Finding{{ID: "nil-deref"}, {ID: "sql-inject"}, {ID: "race"}}
	pred := []Finding{{ID: "nil-deref"}, {ID: "typo"}, {ID: "sql-inject"}}
	r := Score(gold, pred)
	if r.TruePositives != 2 || r.FalsePositives != 1 || r.FalseNegatives != 1 {
		t.Fatalf("%+v", r)
	}
	if r.Precision < 0.66 || r.Precision > 0.67 {
		t.Fatalf("precision %v", r.Precision)
	}
	if r.Recall < 0.66 || r.Recall > 0.67 {
		t.Fatalf("recall %v", r.Recall)
	}
	if r.F1 <= 0 {
		t.Fatal("f1")
	}
}
