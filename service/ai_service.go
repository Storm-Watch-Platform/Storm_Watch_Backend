package service

import (
	"strings"

	"github.com/Storm-Watch-Platform/Storm_Watch_Backend/domain"
)

type AISystem interface {
	ProcessReport(r *domain.Report) (*domain.ReportEnrichment, error)
}

type HardcodeAI struct{}

func NewHardcodeAI() *HardcodeAI {
	return &HardcodeAI{}
}

func (ai *HardcodeAI) ProcessReport(r *domain.Report) (*domain.ReportEnrichment, error) {

	// --- RULE BASED AI FAKE ---
	desc := strings.ToLower(r.Description + " " + r.Detail)

	out := &domain.ReportEnrichment{
		Confidence: 90,
	}

	switch {
	case strings.Contains(desc, "flood") || strings.Contains(desc, "ngập"):
		out.Category = "FLOOD"
		out.Urgency = "HIGH"
		out.Summary = "Detected flooding situation based on user report."

	case strings.Contains(desc, "fire") || strings.Contains(desc, "cháy"):
		out.Category = "FIRE"
		out.Urgency = "HIGH"
		out.Summary = "Detected fire incident requiring urgent attention."

	case strings.Contains(desc, "accident") || strings.Contains(desc, "tai nạn"):
		out.Category = "ACCIDENT"
		out.Urgency = "MEDIUM"
		out.Summary = "Traffic accident reported by user."

	default:
		out.Category = "UNKNOWN"
		out.Urgency = "LOW"
		out.Summary = "Unable to identify clear emergency category."
	}

	return out, nil
}
