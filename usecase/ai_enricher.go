package usecase

import (
	"time"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
)

type MockAIEnricher struct{}

func NewMockAIEnricher() *MockAIEnricher {
	return &MockAIEnricher{}
}

func (e *MockAIEnricher) Enrich(report *domain.Report) *domain.ReportEnrichment {
	return &domain.ReportEnrichment{
		Category:    mockDetectCategory(report),
		Urgency:     mockDetectUrgency(report),
		Summary:     mockGenerateSummary(report),
		Confidence:  90,
		ExtractedAt: time.Now().Unix(),
	}
}

func mockDetectCategory(r *domain.Report) string {
	// rule-based fake AI
	if contains(r.Description, "cháy") || contains(r.Detail, "fire") {
		return "fire"
	}
	if contains(r.Description, "lũ") || contains(r.Detail, "flood") {
		return "flood"
	}
	if contains(r.Description, "tai nạn") {
		return "accident"
	}
	return "unknown"
}

func mockDetectUrgency(r *domain.Report) string {
	if contains(r.Description, "khẩn cấp") {
		return "HIGH"
	}
	if contains(r.Description, "nguy hiểm") {
		return "MEDIUM"
	}
	return "LOW"
}

func mockGenerateSummary(r *domain.Report) string {
	return "This is a generated fake summary for development."
}

func contains(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && (stringContainsInsensitive(s, sub))
}

func stringContainsInsensitive(s, substr string) bool {
	s = normalize(s)
	substr = normalize(substr)
	return findIndex(s, substr) >= 0
}

func normalize(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		// lowercase và loại bỏ dấu tiếng Việt đơn giản
		if r >= 'A' && r <= 'Z' {
			r = r + 32
		}
		out = append(out, r)
	}
	return string(out)
}

func findIndex(s, sub string) int {
	n, m := len(s), len(sub)
	if m == 0 {
		return 0
	}
	if m > n {
		return -1
	}
	for i := 0; i <= n-m; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}
