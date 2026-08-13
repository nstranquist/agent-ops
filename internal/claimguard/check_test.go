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

func TestExtractClaimsPreservesGroupedNumbers(t *testing.T) {
	got := ExtractClaims("Served 1,000+ developers from 50,000 accounts with 10k+ jobs.")
	want := []string{"1,000+", "50,000", "10k+"}
	if len(got) != len(want) {
		t.Fatalf("claims = %q, want %q", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("claims = %q, want %q", got, want)
		}
	}
	report := Check("Served 1,000+ developers.", ClaimsFile{Allowed: []string{"Reviewed: served 1,000+ developers."}})
	if !report.OK {
		t.Fatalf("grouped allowlisted claim was rejected: %+v", report.Violations)
	}
	short := Check("Improved by 33%.", ClaimsFile{Allowed: []string{"Improved by 133%."}})
	if short.OK {
		t.Fatal("short token matched inside a longer allowlisted number")
	}
}

func TestCheckKeepsQuantifiedClaimShapesDistinct(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		allowed string
	}{
		{name: "percent does not allow bare count", text: "Served 100 users.", allowed: "Reviewed: 100% uptime."},
		{name: "money does not allow bare count", text: "Served 100 users.", allowed: "Reviewed: $100 cost."},
		{name: "lower bound does not allow exact count", text: "Served 1,000 users.", allowed: "Reviewed: 1,000+ users."},
		{name: "bare count does not allow percentage", text: "Reached 100% uptime.", allowed: "Reviewed: 100 users."},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			report := Check(test.text, ClaimsFile{Allowed: []string{test.allowed}})
			if report.OK {
				t.Fatalf("mismatched claim shape was allowed: text=%q allowed=%q", test.text, test.allowed)
			}
		})
	}
}

func TestCheckAllowsExactTokenInsideReviewedSentence(t *testing.T) {
	for _, test := range []struct {
		text    string
		allowed string
	}{
		{text: "Reached 100% uptime.", allowed: "Reviewed: reached 100% uptime."},
		{text: "Reduced cost by $100.", allowed: "Reviewed: reduced cost by $100."},
		{text: "Served 1,000+ users.", allowed: "Reviewed: served 1,000+ users."},
	} {
		if report := Check(test.text, ClaimsFile{Allowed: []string{test.allowed}}); !report.OK {
			t.Fatalf("exact reviewed claim was rejected: %+v", report.Violations)
		}
	}
}
