package report

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"

	"cyber-ecosystem/app/services/esports/internal/module/entitle"
	"cyber-ecosystem/app/services/esports/internal/module/profile"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

const (
	// The confidence ladder's bands, the API three-band ruling: below the
	// floor the cohort says nothing (a null percentile, never a made-up
	// number); between floor and ceiling it says something provisional;
	// above the ceiling it has settled.
	cohortFloor = 10
	cohortCeil  = 50

	// The tier ladder judged by composite percentile — the prototype's own
	// thresholds, the frozen tier table in the other's vocabulary (95/85/60
	// is top-5/15/40-percent).
	tierS = 95.0
	tierA = 85.0
	tierB = 60.0

	// The improvement board's target is the ladder's excellence line, in
	// the same percentile currency as everything else.
	targetPercentile = tierA
)

// DO ------------------------------------------------------------------------------------------------------------------

type Attempt struct {
	ID        string
	TestCode  string
	UserID    string
	SessionID *string
	Valid     bool
}

type AbilityLine struct {
	Ability     string  // the six-word vocabulary
	GameCode    string  // the game whose score fed the axis
	Score       float64 // the record's own authoritative score
	Percentile  *float64
	Confidence  esportspb.PeerConfidence
	SampleCount int
	PeerAverage *float64
}

type Report struct {
	RecordID     string
	Unlocked     bool
	Tier         string
	Composite    *float64
	Abilities    []AbilityLine
	Comparison   []AbilityDelta
	Improvements []ReportImprovement
	Tested       int
	Total        int
}

type AbilityDelta struct {
	Ability  string
	Previous *float64
	Delta    *float64
}

type ReportImprovement struct {
	Ability string
	Current float64
	Gap     float64
}

type BoardEntry struct {
	TestCode string
	Ability  string
	Score    float64
	At       time.Time
}

type CohortStats struct {
	Avg         *float64
	Std         *float64
	SampleCount int
}

// Port ---------------------------------------------------------------------------------------------------------------

type RecordBoardRP interface {
	Anchor(ctx context.Context, recordID string) (*Attempt, error) // nil, nil = no such record
	SessionBoard(ctx context.Context, sessionID string, anchor *Attempt) ([]BoardEntry, error)
	// GenreOf resolves the board's genre — the session's own draw when one
	// exists, the single test's category otherwise.
	GenreOf(ctx context.Context, sessionID string, anchor *Attempt) (string, error)
	// GenreDimensions is how many abilities the genre orders — the
	// completeness headline's denominator, read from match_config, never
	// counted off the measured board (which would make tested always
	// equal total and the headline meaningless).
	GenreDimensions(ctx context.Context, genre string) (int, error)
	Cohort(ctx context.Context, ability string, birthYear *int32, sex *string, ageBand *string) (stats CohortStats, basis string, err error)
	// PreviousSessionRecord finds the same member's latest record for the
	// ability from a session older than the anchor's — the last attempt,
	// never a retry inside this one.
	PreviousSessionRecord(ctx context.Context, userID, ability string, before time.Time) (*BoardEntry, error) // nil, nil = none
}

// UC ------------------------------------------------------------------------------------------------------------------

type ReportUC struct {
	shared.UC
	boardRP   RecordBoardRP
	entRP     entitle.EntitlementRP
	profileRP profile.ProfileRP
}

func NewReportUC(logger *slog.Logger, tm shared.Transaction, boardRP RecordBoardRP, entRP entitle.EntitlementRP, profileRP profile.ProfileRP) *ReportUC {
	return &ReportUC{
		UC:        shared.NewUC(logger.With("module", "module/report"), tm),
		boardRP:   boardRP,
		entRP:     entRP,
		profileRP: profileRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ReportUC) GetReport(ctx context.Context, recordID string) (*Report, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, esportspb.ErrorEsportsReportRecordNotOwned("").WithCause(fmt.Errorf("report read requires a session"))
	}
	rec, err := uc.boardRP.Anchor(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if rec == nil {
		return nil, esportspb.ErrorEsportsReportRecordNotFound("")
	}
	if rec.UserID != subject.UserID {
		return nil, esportspb.ErrorEsportsReportRecordNotOwned("")
	}
	if !rec.Valid {
		return nil, esportspb.ErrorEsportsReportRecordInvalid("")
	}

	board, err := uc.boardRP.SessionBoard(ctx, sessionIDOf(rec), rec)
	if err != nil {
		return nil, err
	}

	pf, err := uc.profileRP.FindByUserID(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	var birthYear *int32
	var sex *string
	var ageBand *string
	if pf != nil {
		birthYear, sex, ageBand = pf.BirthYear, pf.Sex, pf.AgeBand
	}

	// One cohort read per measured ability; the pool is the same members'
	// distribution under the ruling's key.
	lines := make([]AbilityLine, 0, len(board))
	percentiles := make([]float64, 0, len(board))
	byAbility := make(map[string]AbilityLine, len(board))
	for _, e := range board {
		pct, conf, avg, n := uc.percentileFor(ctx, e, birthYear, sex, ageBand)
		line := AbilityLine{
			Ability: e.Ability, GameCode: e.TestCode, Score: e.Score,
			Percentile: pct, Confidence: conf, SampleCount: n, PeerAverage: avg,
		}
		lines = append(lines, line)
		byAbility[e.Ability] = line
		if pct != nil {
			percentiles = append(percentiles, *pct)
		}
	}

	r := &Report{
		RecordID:  recordID,
		Abilities: lines,
		Tested:    len(lines),
	}
	if comp := compositeOf(percentiles); comp != nil {
		r.Composite = comp
		r.Tier = tierOf(*comp)
	}

	if unlocked, err := uc.unlockedFor(ctx, subject.UserID, recordID); err != nil {
		return nil, err
	} else {
		r.Unlocked = unlocked
	}

	if r.Unlocked {
		r.Comparison, err = uc.comparisonFor(ctx, subject.UserID, board, byAbility, birthYear, sex, ageBand)
		if err != nil {
			return nil, err
		}
		r.Improvements = improvementsOf(lines)
	}
	if total, err := uc.totalDimensions(ctx, sessionIDOf(rec), rec); err != nil {
		return nil, err
	} else {
		r.Total = total
	}
	return r, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func sessionIDOf(rec *Attempt) string {
	if rec.SessionID == nil {
		return ""
	}
	return *rec.SessionID
}

func compositeOf(percentiles []float64) *float64 {
	// Equal-weight mean of the percentiles that exist — a missing ability
	// never counts as zero, the PRD's own law.
	if len(percentiles) == 0 {
		return nil
	}
	sum := 0.0
	for _, p := range percentiles {
		sum += p
	}
	v := sum / float64(len(percentiles))
	return &v
}

func tierOf(pct float64) string {
	switch {
	case pct >= tierS:
		return shared.TierS
	case pct >= tierA:
		return shared.TierA
	case pct >= tierB:
		return shared.TierB
	default:
		return shared.TierC
	}
}

func improvementsOf(lines []AbilityLine) []ReportImprovement {
	// Weakest first, each carrying its gap to the excellence line as a
	// percent of the target.
	out := make([]ReportImprovement, 0, len(lines))
	for _, l := range lines {
		if l.Percentile == nil {
			continue // a score-only axis has no percentile gap to close
		}
		out = append(out, ReportImprovement{
			Ability: l.Ability,
			Current: *l.Percentile,
			Gap:     math.Round((targetPercentile - *l.Percentile) / targetPercentile * 100),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Current < out[j].Current })
	return out
}

func (uc *ReportUC) totalDimensions(ctx context.Context, sessionID string, anchor *Attempt) (int, error) {
	// The denominator is the ordered count from match_config, one axis
	// for a genre-less single test — never the measured board's own
	// count, which would make tested always equal total.
	genre, err := uc.boardRP.GenreOf(ctx, sessionID, anchor)
	if err != nil {
		return 0, err
	}
	if genre == "" || genre == "universal" {
		return 1, nil
	}
	return uc.boardRP.GenreDimensions(ctx, genre)
}

func (uc *ReportUC) percentileFor(ctx context.Context, e BoardEntry, birthYear *int32, sex *string, ageBand *string) (*float64, esportspb.PeerConfidence, *float64, int) {
	// Z = (score - avg) / std through the exact normal CDF; below the
	// floor the pool says nothing (null, LOW); a degenerate pool benches
	// at 51/50 by side. The whys ride the legacy method this follows.
	stats, _, err := uc.boardRP.Cohort(ctx, e.Ability, birthYear, sex, ageBand)
	if err != nil {
		return nil, esportspb.PeerConfidence_PEER_CONFIDENCE_LOW, nil, 0 // a broken pool read degrades to score-only, never fails the report
	}
	if stats.SampleCount < cohortFloor || stats.Avg == nil || stats.Std == nil {
		return nil, esportspb.PeerConfidence_PEER_CONFIDENCE_LOW, stats.Avg, stats.SampleCount
	}
	conf := esportspb.PeerConfidence_PEER_CONFIDENCE_HIGH
	if stats.SampleCount <= cohortCeil {
		conf = esportspb.PeerConfidence_PEER_CONFIDENCE_MEDIUM
	}
	if *stats.Std < 0.01 {
		var pct float64
		if e.Score > *stats.Avg {
			pct = 51
		} else {
			pct = 50
		}
		return &pct, conf, stats.Avg, stats.SampleCount
	}
	z := (e.Score - *stats.Avg) / *stats.Std
	pct := math.Round(50 * math.Erfc(-z/math.Sqrt2))
	return &pct, conf, stats.Avg, stats.SampleCount
}

func (uc *ReportUC) unlockedFor(ctx context.Context, userID, recordID string) (bool, error) {
	// The ledger's three facts: the record-scope grant (the pricing
	// ruling's unlock unit), the membership window, a redeemed family
	// code — the same facts the family access judgement reads.
	now := time.Now()
	if g, err := uc.entRP.ByScope(ctx, userID, shared.ScopeTrial, &recordID); err != nil {
		return false, err
	} else if g != nil && (g.ExpiresAt == nil || g.ExpiresAt.After(now)) {
		return true, nil
	}
	if g, err := uc.entRP.ByScope(ctx, userID, shared.ScopeMembership, nil); err != nil {
		return false, err
	} else if g != nil && g.ExpiresAt != nil && g.ExpiresAt.After(now) {
		return true, nil
	}
	if g, err := uc.entRP.ByScope(ctx, userID, shared.ScopeReportType, nil); err != nil {
		return false, err
	} else if g != nil && (g.ExpiresAt == nil || g.ExpiresAt.After(now)) {
		return true, nil
	}
	return false, nil
}

func (uc *ReportUC) comparisonFor(ctx context.Context, userID string, board []BoardEntry, current map[string]AbilityLine, birthYear *int32, sex *string, ageBand *string) ([]AbilityDelta, error) {
	// The delta is current minus previous, so a positive move is
	// improvement; a previous without a percentile renders as
	// first-measurement, never a fall from zero.
	out := make([]AbilityDelta, 0, len(board))
	for _, e := range board {
		d := AbilityDelta{Ability: e.Ability}
		prev, err := uc.boardRP.PreviousSessionRecord(ctx, userID, e.Ability, e.At)
		if err != nil {
			return nil, err
		}
		if prev != nil {
			if pct, _, _, _ := uc.percentileFor(ctx, *prev, birthYear, sex, ageBand); pct != nil {
				d.Previous = pct
				if cur, ok := current[e.Ability]; ok && cur.Percentile != nil {
					delta := *cur.Percentile - *pct
					d.Delta = &delta
				}
			}
		}
		out = append(out, d)
	}
	return out, nil
}
