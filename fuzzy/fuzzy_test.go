package fuzzy

import (
	"reflect"
	"testing"
)

func TestMatchValueRewardsConsecutiveCharacters(t *testing.T) {
	consecutive, ok := MatchValue("Platform API", "plat")
	if !ok {
		t.Fatal("expected consecutive match")
	}
	scattered, ok := MatchValue("Product labels and templates", "plat")
	if !ok {
		t.Fatal("expected scattered subsequence match")
	}
	if consecutive.Score <= scattered.Score {
		t.Fatalf("consecutive score %d should exceed scattered score %d", consecutive.Score, scattered.Score)
	}
}

func TestMatchValueChoosesLaterConsecutiveRunOverEarlyScatteredMatch(t *testing.T) {
	match, ok := MatchValue("aqqqbqqqc abc", "abc")
	if !ok {
		t.Fatal("expected fuzzy match")
	}
	if want := []int{10, 11, 12}; !reflect.DeepEqual(match.Positions, want) {
		t.Fatalf("positions = %v, want optimal consecutive run %v", match.Positions, want)
	}
}

func TestMatchValueRequiresEveryToken(t *testing.T) {
	if _, ok := MatchValue("Platform API", "plat zzz"); ok {
		t.Fatal("a token without match must reject the value")
	}
	if m, ok := MatchValue("anything", "  "); !ok || m.Score != 0 {
		t.Fatalf("an empty query matches with score 0, got %+v %v", m, ok)
	}
}

func TestScoreFieldsAllowTokensAcrossFieldsAndWeights(t *testing.T) {
	score, ok := ScoreFields("platform api",
		Field{Value: "Platform", Weight: 3},
		Field{Value: "api", Weight: 2},
	)
	if !ok || score <= 0 {
		t.Fatalf("multi-field score = %d, matched = %v", score, ok)
	}
	if _, ok := ScoreFields("platform missing", Field{Value: "Platform API", Weight: 3}); ok {
		t.Fatal("every query token should be required")
	}
	heavy, _ := ScoreFields("plat", Field{Value: "Platform", Weight: 3})
	light, _ := ScoreFields("plat", Field{Value: "Platform", Weight: 1})
	if heavy <= light {
		t.Fatalf("weight 3 score %d should exceed weight 1 score %d", heavy, light)
	}
}

func TestScoreFieldsAgreesWithMatchValue(t *testing.T) {
	values := []string{"Platform API", "invoices.list", "auth.login", "Product labels and templates", "x"}
	queries := []string{"plat", "inv list", "login", "pt", "zz"}
	for _, v := range values {
		for _, q := range queries {
			m, ok1 := MatchValue(v, q)
			s, ok2 := ScoreFields(q, Field{Value: v, Weight: 1})
			if ok1 != ok2 || (ok1 && m.Score != s) {
				t.Fatalf("%q / %q: MatchValue=(%d,%v) ScoreFields=(%d,%v)", v, q, m.Score, ok1, s, ok2)
			}
		}
	}
}

func TestExactAndPrefixRankFirst(t *testing.T) {
	exact, _ := MatchValue("list", "list")
	prefix, _ := MatchValue("listing", "list")
	inner, _ := MatchValue("blacklist", "list")
	if !(exact.Score > prefix.Score && prefix.Score > inner.Score) {
		t.Fatalf("scores exact=%d prefix=%d inner=%d", exact.Score, prefix.Score, inner.Score)
	}
	if !Contains("invoices.list", "inv") || Contains("invoices.list", "xyz") {
		t.Fatal("Contains disagrees with matching")
	}
}
