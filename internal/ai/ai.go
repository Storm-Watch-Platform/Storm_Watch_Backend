package ai

type AIResult struct {
	Category   string `json:"category"`
	Urgency    string `json:"urgency"`
	Summary    string `json:"summary"`
	Confidence int    `json:"confidence"`
}

// AnalyzeReport is a fake AI function for prototyping
func AnalyzeReport(text string) (*AIResult, error) {
	if len(text) == 0 {
		return &AIResult{Category: "UNKNOWN", Urgency: "LOW", Summary: "No detail", Confidence: 0}, nil
	}

	// fake rules
	if contains(text, "lụt") || contains(text, "ngập") {
		return &AIResult{Category: "FLOOD", Urgency: "HIGH", Summary: "Flood reported", Confidence: 95}, nil
	}
	if contains(text, "cháy") {
		return &AIResult{Category: "FIRE", Urgency: "MEDIUM", Summary: "Fire reported", Confidence: 90}, nil
	}

	return &AIResult{Category: "OTHER", Urgency: "LOW", Summary: "Other incident", Confidence: 50}, nil
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && stringIndex(s, sub) >= 0
}

func stringIndex(a, b string) int {
	for i := 0; i+len(b) <= len(a); i++ {
		if a[i:i+len(b)] == b {
			return i
		}
	}
	return -1
}
