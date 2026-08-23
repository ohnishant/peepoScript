package token

import "testing"

func TestKeywordsReturnsEveryKeywordSorted(t *testing.T) {
	tests := []string{
		"Bedge",
		"Hmmge",
		"NODDERS",
		"NOPERS",
		"PepegaCredit",
		"PepoG",
		"SadgeBusiness",
		"Scoots",
		"Thinking1",
		"Thinking2",
		"Wokege",
		"mitosis",
		"peepoBye",
		"peepoCookie",
		"peepoFriendship",
		"peepoGreaterThan",
		"peepoJuice",
		"peepoLessThan",
		"peepoShrug",
	}

	got := Keywords()

	if len(got) != len(tests) {
		t.Fatalf("keyword count mismatch: want %d, got %d", len(tests), len(got))
	}
	for i, want := range tests {
		if got[i] != want {
			t.Errorf("Keywords()[%d] = %q, want %q", i, got[i], want)
		}
	}
}
