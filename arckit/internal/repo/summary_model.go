package repo

import (
	"time"

	"github.com/algorandfoundation/ARCs/arckit/internal/diag"
)

const (
	upcomingLastCallWindowDays = 14
	staleAdoptionThresholdDays = 30
	dateLayout                 = "2006-01-02"
)

var (
	statusOrder = []string{"Draft", "Review", "Last Call", "Final", "Stagnant", "Withdrawn", "Idle", "Deprecated", "Living"}
	typeOrder   = []string{"Standards Track", "Meta"}
)

type RepoSummary struct {
	Root                                string
	GeneratedAt                         time.Time
	ValidationSummary                   diag.Summary
	TotalARCs                           int
	StatusCounts                        []SummaryCount
	TypeCounts                          []SummaryCount
	ARCsWithAssets                      int
	AdoptionSummaryCoverage             int
	TotalAdoptionFiles                  int
	TotalAdopterEntries                 int
	ARCsWithTrackedAdopters             int
	AdoptionReadinessCounts             []SummaryCount
	ReferenceImplementationStatusCounts []SummaryCount
	OverdueLastCall                     []LastCallSummaryRow
	UpcomingLastCall                    []LastCallSummaryRow
	IdleARCs                            []IdleSummaryRow
	ImplementationRequiredNotShipped    []ImplementationSummaryRow
	FinalZeroAdopters                   []AdoptionSummaryRow
	FinalLowAdopters                    []AdoptionSummaryRow
	StaleAdoptionReviews                []StaleAdoptionSummaryRow
	AdopterEntriesByCategory            []SummaryCount
	AdopterEntriesByStatus              []SummaryCount
	TopAdoptersByCoverage               []NamedCount
	TopARCsByAdopters                   []ARCCount
	TopRequiresTargets                  []ReferencedARCCount
	TopExtendsTargets                   []ReferencedARCCount
	SupersessionRows                    []SupersessionSummaryRow
}

type SummaryCount struct {
	Label string
	Count int
}

type NamedCount struct {
	Name  string
	Count int
}

type ARCCount struct {
	ARC    int
	Title  string
	Status string
	Count  int
}

type ReferencedARCCount struct {
	ARC   int
	Title string
	Count int
}

type LastCallSummaryRow struct {
	ARC    int
	Title  string
	Date   string
	Days   int
	Action string
}

type IdleSummaryRow struct {
	ARC    int
	Title  string
	Date   string
	Days   int
	Action string
}

type ImplementationSummaryRow struct {
	ARC                     int
	Title                   string
	Status                  string
	ReferenceImplementation string
	Action                  string
}

type AdoptionSummaryRow struct {
	ARC               int
	Title             string
	AdoptionReadiness string
	LastReviewed      string
	AdopterCount      int
	Action            string
}

type StaleAdoptionSummaryRow struct {
	ARC          int
	Title        string
	Status       string
	LastReviewed string
	AgeDays      int
	Action       string
}

type SupersessionSummaryRow struct {
	ARC          int
	Title        string
	Supersedes   string
	SupersededBy string
}
