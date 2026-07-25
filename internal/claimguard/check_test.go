package claimguard

import "testing"

func TestCheck_FailClosedOnUnclaimedNumber(t *testing.T) {
	allow := ClaimsFile{Claims: []string{"10%"}}
	rep := Check("We improved latency by 10% and shipped 999 features.", allow)
	if rep.OK {
		t.Fatal("expected fail")
	}
	found := false
	for _, v := range rep.Violations {
		if v.Token == "999" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected 999 violation, got %+v", rep.Violations)
	}
}

func TestCheck_OKWhenAllowlisted(t *testing.T) {
	allow := ClaimsFile{Claims: []string{"10%", "999"}}
	rep := Check("We improved latency by 10% and shipped 999 features.", allow)
	if !rep.OK {
		t.Fatalf("expected ok, got %+v", rep.Violations)
	}
}
