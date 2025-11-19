package domain

type AIEnricher interface {
	Enrich(report *Report) *ReportEnrichment
}
