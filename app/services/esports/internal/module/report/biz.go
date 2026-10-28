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

	"cyber-ecosystem/app/services/esports/internal/ent/schema/types"
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

const basisNone = "none" // cohort_basis is NotEmpty — a score-only freeze still stamps its rubric honestly: no cohort key existed at pin time

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

type Snapshot struct {
	RecordID    string
	UserID      string
	Tier        string
	Composite   *float64
	Abilities   []types.AbilitySnapshot
	CohortBasis string
}

// SessionAttemptRow is one valid scored attempt under a session — the
// settlement receipt's raw truth, grade attached.
type SessionAttemptRow struct {
	TestCode  string
	Ability   string
	AttemptNo int
	Score     float64
	Grade     *string
	CreatedAt time.Time
}

// SessionAbilityRow is one ability's session verdict — the three numbers
// the settlement face keeps distinct.
type SessionAbilityRow struct {
	Ability      string
	GameCode     string
	Best         float64
	Attempts     []float64
	BestGrade    *string
	Percentile   *float64
	PrevSession  *float64
	HistoryBest  *float64
}

type SessionReport struct {
	SessionID string
	Genre     string
	StartedAt time.Time
	Anchor    *string
	Tier      string // "" while no anchor exists
	Composite *float64
	Abilities []SessionAbilityRow
	Tested    int
	Total     int
}

// SessionMetaRow is the session's own row — the read-back faces'
// ownership check, lineup order and liveness.
type SessionMetaRow struct {
	ID        string
	UserID    string
	Genre     string
	Steps     []types.MatchStep
	DoneCount int
	CreatedAt time.Time
}

// DigestAbilityRow is the member's per-ability cross-session aggregate.
type DigestAbilityRow struct {
	Ability    string
	Best       float64
	BestGrade  *string
	Percentile *float64
	Sessions   int
	LastAt     time.Time
}

// DigestSessionRow is the latest session line of one genre — no verdict
// of its own; the composite face reads GetSessionReport with the id.
type DigestSessionRow struct {
	SessionID string
	Genre     string
	Finished  bool
	DoneCount int
	StepCount int
	CreatedAt time.Time
	Anchor    *string
}

type AbilityDigest struct {
	SessionsTotal int
	SessionsFPS   int
	SessionsMOBA  int
	Abilities     []DigestAbilityRow
	Sessions      []DigestSessionRow
}

type judgement struct { // the live computation's whole answer — the freeze serializes it, the free tier renders it
	Lines     []AbilityLine
	Composite *float64
	Tier      string
	Basis     string
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
	// SessionMeta reads the session's own row — the read-back faces'
	// ownership check, lineup order and liveness. nil, nil = no such session.
	SessionMeta(ctx context.Context, sessionID string) (*SessionMetaRow, error)
	// SessionAttempts is every valid scored attempt under the session in
	// ordinal order, grade attached — the receipt's raw truth.
	SessionAttempts(ctx context.Context, sessionID string) ([]SessionAttemptRow, error)
	// AnchorOfSession is the session's latest scored record — the
	// composite verdict's owner. nil, nil = nothing scored yet.
	AnchorOfSession(ctx context.Context, sessionID string) (*Attempt, error)
	// AbilityBests is the member's per-ability cross-session aggregate:
	// the best score with its grade and percentile, how many sessions
	// carried a valid attempt, and the last time. A never-measured
	// ability is an absent row.
	AbilityBests(ctx context.Context, userID string) ([]DigestAbilityRow, error)
	// SessionCounts is how many distinct sessions carried a valid
	// attempt, split per genre — practice runs never count.
	SessionCounts(ctx context.Context, userID string) (total int, fps int, moba int, err error)
	// LatestSessions is the latest session per genre that carried at
	// least one valid attempt — at most two rows.
	LatestSessions(ctx context.Context, userID string) ([]SessionMetaRow, error)
	// PreviousSessionScores is, per ability, the score from the member's
	// latest scored record older than the given time — the settlement
	// face's "vs last session" word.
	PreviousSessionScores(ctx context.Context, userID string, before time.Time) (map[string]float64, error)
}

type SnapshotRP interface {
	FindByRecord(ctx context.Context, recordID string) (*Snapshot, error) // nil, nil = not frozen yet
	Save(ctx context.Context, s *Snapshot) error                          // insert-only; a racing freeze lands as a no-op
}

// UC ------------------------------------------------------------------------------------------------------------------

type ReportUC struct {
	shared.UC
	boardRP    RecordBoardRP
	snapshotRP SnapshotRP
	entRP      entitle.EntitlementRP
	profileRP  profile.ProfileRP
	eventBus   shared.DomainEventBus
}

func NewReportUC(logger *slog.Logger, tm shared.Transaction, boardRP RecordBoardRP, snapshotRP SnapshotRP, entRP entitle.EntitlementRP, profileRP profile.ProfileRP, eventBus shared.DomainEventBus) *ReportUC {
	uc := &ReportUC{
		UC:         shared.NewUC(logger.With("module", "module/report"), tm),
		boardRP:    boardRP,
		snapshotRP: snapshotRP,
		entRP:      entRP,
		profileRP:  profileRP,
		eventBus:   eventBus,
	}
	if eventBus != nil {
		// The subscription is this module's whole coupling to the unlock
		// fact: the freeze runs on the bus, the grant writers stay where
		// they are, and no module imports another.
		eventBus.SubscribeReportUnlocked(uc.onReportUnlocked)
	}
	return uc
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
	board = dedupeBoard(board)

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

	r := &Report{RecordID: recordID}
	if unlocked, err := uc.unlockedFor(ctx, subject.UserID, recordID); err != nil {
		return nil, err
	} else {
		r.Unlocked = unlocked
	}

	var lines []AbilityLine
	if r.Unlocked {
		// An unlocked report reads its frozen judgement — the numbers the
		// unlock bought never move with the rubric. The membership and
		// family sources name no record to freeze on, and an event can be
		// lost: both land here, freezing once, idempotently.
		snap, ferr := uc.freezeFor(ctx, subject.UserID, recordID)
		if ferr != nil {
			return nil, ferr
		}
		if snap != nil {
			r.Tier, r.Composite = snap.Tier, snap.Composite
			lines = linesOfSnapshot(snap.Abilities)
		}
	}
	if lines == nil {
		// The free tier's live judgement — also the unlocked fallback when
		// the freeze found nothing to pin (defensive: the anchor checks
		// above already rule the real case out).
		j := uc.computeFor(ctx, board, birthYear, sex, ageBand)
		r.Tier, r.Composite = j.Tier, j.Composite
		lines = j.Lines
	}
	r.Abilities = lines
	r.Tested = len(lines)

	if r.Unlocked {
		byAbility := make(map[string]AbilityLine, len(lines))
		for _, l := range lines {
			byAbility[l.Ability] = l
		}
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

// GetSessionReport is the session read-back face: the lineup's own
// order, every ability's three-number verdict, and the anchor's
// composite under the same law the record board follows — frozen when a
// snapshot exists, live otherwise. A session with no scored attempt is
// an honest empty board, never an error.
func (uc *ReportUC) GetSessionReport(ctx context.Context, sessionID string) (*SessionReport, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, esportspb.ErrorEsportsReportRecordNotOwned("").WithCause(fmt.Errorf("session report read requires a session"))
	}
	meta, err := uc.boardRP.SessionMeta(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, esportspb.ErrorEsportsMatchSessionNotFound("")
	}
	if meta.UserID != subject.UserID {
		return nil, esportspb.ErrorEsportsReportSessionNotOwned("")
	}

	attempts, err := uc.boardRP.SessionAttempts(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	r := &SessionReport{SessionID: meta.ID, Genre: meta.Genre, StartedAt: meta.CreatedAt}

	// Per ability in the lineup's order: the best score with its
	// attempts receipt. An attempt outside the lineup's codes has no
	// line to sit on — dropped, not invented a slot for.
	bestOf := make(map[string]SessionAttemptRow, len(meta.Steps))
	scoresOf := make(map[string][]float64, len(meta.Steps))
	for _, a := range attempts {
		scoresOf[a.TestCode] = append(scoresOf[a.TestCode], a.Score)
		if b, ok := bestOf[a.TestCode]; !ok || a.Score > b.Score {
			bestOf[a.TestCode] = a
		}
	}
	for _, s := range meta.Steps {
		scores, ok := scoresOf[s.TestCode]
		if !ok {
			continue // an unmeasured step never lands a line
		}
		r.Abilities = append(r.Abilities, SessionAbilityRow{
			Ability:   s.Ability,
			GameCode:  s.TestCode,
			Best:      bestOf[s.TestCode].Score,
			Attempts:  scores,
			BestGrade: bestOf[s.TestCode].Grade,
		})
	}
	r.Tested = len(r.Abilities)

	// The anchor's verdict: the frozen snapshot when one exists, the
	// live judgement otherwise — the free tier's own law.
	var firstAt time.Time
	if len(attempts) > 0 {
		firstAt = attempts[0].CreatedAt
	}
	if anchor, aerr := uc.boardRP.AnchorOfSession(ctx, sessionID); aerr != nil {
		return nil, aerr
	} else if anchor != nil {
		anchorID := anchor.ID
		r.Anchor = &anchorID
		pctByAbility := uc.sessionPercentiles(ctx, anchor.ID)
		for i := range r.Abilities {
			if pct, ok := pctByAbility[r.Abilities[i].Ability]; ok {
				r.Abilities[i].Percentile = pct
			}
		}
		if snap, ferr := uc.snapshotRP.FindByRecord(ctx, anchor.ID); ferr != nil {
			return nil, ferr
		} else if snap != nil {
			r.Tier, r.Composite = snap.Tier, snap.Composite
		} else {
			pf, perr := uc.profileRP.FindByUserID(ctx, subject.UserID)
			if perr != nil {
				return nil, perr
			}
			var birthYear *int32
			var sex *string
			var ageBand *string
			if pf != nil {
				birthYear, sex, ageBand = pf.BirthYear, pf.Sex, pf.AgeBand
			}
			j := uc.computeFor(ctx, sessionBoardOf(attempts), birthYear, sex, ageBand)
			r.Tier, r.Composite = j.Tier, j.Composite
			// The live lines carry the per-ability percentiles the
			// frozen branch already pinned above.
			for _, l := range j.Lines {
				for i := range r.Abilities {
					if r.Abilities[i].Ability == l.Ability && r.Abilities[i].Percentile == nil {
						r.Abilities[i].Percentile = l.Percentile
					}
				}
			}
		}
	}

	// The cross-session words: the previous session's score and the
	// history best — both feed only lines that exist.
	if len(attempts) > 0 {
		prev, perr := uc.boardRP.PreviousSessionScores(ctx, subject.UserID, firstAt)
		if perr != nil {
			return nil, perr
		}
		for i := range r.Abilities {
			if v, ok := prev[r.Abilities[i].Ability]; ok {
				pv := v
				r.Abilities[i].PrevSession = &pv
			}
		}
	}
	if bests, berr := uc.boardRP.AbilityBests(ctx, subject.UserID); berr != nil {
		return nil, berr
	} else {
		byAbility := make(map[string]DigestAbilityRow, len(bests))
		for _, b := range bests {
			byAbility[b.Ability] = b
		}
		for i := range r.Abilities {
			if b, ok := byAbility[r.Abilities[i].Ability]; ok {
				v := math.Max(b.Best, r.Abilities[i].Best)
				r.Abilities[i].HistoryBest = &v
			} else {
				v := r.Abilities[i].Best
				r.Abilities[i].HistoryBest = &v
			}
		}
	}

	if total, terr := uc.boardRP.GenreDimensions(ctx, meta.Genre); terr != nil {
		return nil, terr
	} else {
		r.Total = total
	}
	return r, nil
}

// GetAbilityDigest is the cross-session aggregate: session counts by
// genre, the per-ability best lines, and the latest session line per
// genre. An empty history is an honest zero-and-nothing answer.
func (uc *ReportUC) GetAbilityDigest(ctx context.Context) (*AbilityDigest, error) {
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, esportspb.ErrorEsportsReportRecordNotOwned("").WithCause(fmt.Errorf("digest read requires a session"))
	}
	d := &AbilityDigest{}
	var err error
	if d.SessionsTotal, d.SessionsFPS, d.SessionsMOBA, err = uc.boardRP.SessionCounts(ctx, subject.UserID); err != nil {
		return nil, err
	}
	if d.Abilities, err = uc.boardRP.AbilityBests(ctx, subject.UserID); err != nil {
		return nil, err
	}
	lines, err := uc.boardRP.LatestSessions(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	for _, meta := range lines {
		row := DigestSessionRow{
			SessionID: meta.ID, Genre: meta.Genre,
			DoneCount: meta.DoneCount, StepCount: len(meta.Steps),
			Finished: meta.DoneCount >= len(meta.Steps), CreatedAt: meta.CreatedAt,
		}
		if anchor, aerr := uc.boardRP.AnchorOfSession(ctx, meta.ID); aerr != nil {
			return nil, aerr
		} else if anchor != nil {
			id := anchor.ID
			row.Anchor = &id
		}
		d.Sessions = append(d.Sessions, row)
	}
	return d, nil
}

// sessionPercentiles reads the anchor's per-ability cohort answer from
// the frozen snapshot when one exists; the live branch of the caller
// fills the lines itself.
func (uc *ReportUC) sessionPercentiles(ctx context.Context, anchorID string) map[string]*float64 {
	snap, err := uc.snapshotRP.FindByRecord(ctx, anchorID)
	if err != nil || snap == nil {
		return nil
	}
	out := make(map[string]*float64, len(snap.Abilities))
	for _, a := range snap.Abilities {
		if _, dup := out[a.Ability]; !dup && a.Percentile != nil {
			p := *a.Percentile
			out[a.Ability] = &p
		}
	}
	return out
}

// sessionBoardOf flattens the attempt rows into the board shape the
// live judgement consumes.
func sessionBoardOf(attempts []SessionAttemptRow) []BoardEntry {
	out := make([]BoardEntry, 0, len(attempts))
	for _, a := range attempts {
		out = append(out, BoardEntry{TestCode: a.TestCode, Ability: a.Ability, Score: a.Score, At: a.CreatedAt})
	}
	return out
}

// dedupeBoard folds the session board to one entry per ability — the
// settlement's own law (the best attempt of the retries is the one that
// scores), so a retaken game never draws a doubled radar axis or a
// duplicated ability row on the record board either. Equal scores keep
// the newer attempt.
func dedupeBoard(board []BoardEntry) []BoardEntry {
	best := make(map[string]BoardEntry, len(board))
	order := make([]string, 0, len(board))
	for _, e := range board {
		if cur, ok := best[e.Ability]; !ok || e.Score > cur.Score || (e.Score == cur.Score && e.At.After(cur.At)) {
			if !ok {
				order = append(order, e.Ability)
			}
			best[e.Ability] = e
		}
	}
	out := make([]BoardEntry, 0, len(order))
	for _, ability := range order {
		out = append(out, best[ability])
	}
	return out
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

func (uc *ReportUC) percentileFor(ctx context.Context, e BoardEntry, birthYear *int32, sex, ageBand *string) (*float64, esportspb.PeerConfidence, *float64, int, string) {
	// Z = (score - avg) / std through the exact normal CDF; below the
	// floor the pool says nothing (null, LOW); a degenerate pool benches
	// at 51/50 by side. The whys ride the legacy method this follows. The
	// basis rides out with the answer so a freeze can stamp it.
	stats, basis, err := uc.boardRP.Cohort(ctx, e.Ability, birthYear, sex, ageBand)
	if err != nil {
		return nil, esportspb.PeerConfidence_PEER_CONFIDENCE_LOW, nil, 0, basis // a broken pool read degrades to score-only, never fails the report
	}
	if stats.SampleCount < cohortFloor || stats.Avg == nil || stats.Std == nil {
		return nil, esportspb.PeerConfidence_PEER_CONFIDENCE_LOW, stats.Avg, stats.SampleCount, basis
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
		return &pct, conf, stats.Avg, stats.SampleCount, basis
	}
	z := (e.Score - *stats.Avg) / *stats.Std
	pct := math.Round(50 * math.Erfc(-z/math.Sqrt2))
	return &pct, conf, stats.Avg, stats.SampleCount, basis
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
			if pct, _, _, _, _ := uc.percentileFor(ctx, *prev, birthYear, sex, ageBand); pct != nil {
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

func (uc *ReportUC) computeFor(ctx context.Context, board []BoardEntry, birthYear *int32, sex, ageBand *string) *judgement {
	// One cohort read per measured ability; the pool is the same members'
	// distribution under the ruling's key. Pure computation — a broken
	// pool read degrades to score-only inside percentileFor, never an
	// error here.
	j := &judgement{Lines: make([]AbilityLine, 0, len(board))}
	percentiles := make([]float64, 0, len(board))
	for _, e := range board {
		pct, conf, avg, n, basis := uc.percentileFor(ctx, e, birthYear, sex, ageBand)
		j.Lines = append(j.Lines, AbilityLine{
			Ability: e.Ability, GameCode: e.TestCode, Score: e.Score,
			Percentile: pct, Confidence: conf, SampleCount: n, PeerAverage: avg,
		})
		if pct != nil {
			percentiles = append(percentiles, *pct)
		}
		if basis != "" {
			j.Basis = basis
		}
	}
	if comp := compositeOf(percentiles); comp != nil {
		j.Composite = comp
		j.Tier = tierOf(*comp)
	}
	return j
}

func (uc *ReportUC) freezeFor(ctx context.Context, userID, recordID string) (*Snapshot, error) {
	// The freeze is idempotent: an existing row is the answer itself, and
	// a racing freeze (the event and a first read landing together) is a
	// no-op at the insert, the Grant precedent. A nil answer means the
	// record would not freeze (absent, not the member's, invalid) — the
	// caller falls back to the live judgement.
	if s, err := uc.snapshotRP.FindByRecord(ctx, recordID); err != nil {
		return nil, err
	} else if s != nil {
		return s, nil
	}
	rec, err := uc.boardRP.Anchor(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.UserID != userID || !rec.Valid {
		return nil, nil
	}
	board, err := uc.boardRP.SessionBoard(ctx, sessionIDOf(rec), rec)
	if err != nil {
		return nil, err
	}
	board = dedupeBoard(board)
	pf, err := uc.profileRP.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var birthYear *int32
	var sex *string
	var ageBand *string
	if pf != nil {
		birthYear, sex, ageBand = pf.BirthYear, pf.Sex, pf.AgeBand
	}
	j := uc.computeFor(ctx, board, birthYear, sex, ageBand)
	basis := j.Basis
	if basis == "" {
		basis = basisNone
	}
	tier := j.Tier
	if tier == "" {
		tier = shared.TierUnjudged
	}
	s := &Snapshot{
		RecordID: recordID, UserID: userID,
		Tier: tier, Composite: j.Composite,
		Abilities: snapshotsOf(j.Lines), CohortBasis: basis,
	}
	if err := uc.snapshotRP.Save(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (uc *ReportUC) onReportUnlocked(ctx context.Context, e shared.ReportUnlocked) {
	// Best-effort by contract: the grant is the truth, a failed freeze
	// degrades to the lazy read path, and the error stays a log line.
	if _, err := uc.freezeFor(ctx, e.UserID, e.RecordID); err != nil {
		uc.Log.Error("unlock freeze failed", "user_id", e.UserID, "record_id", e.RecordID, "err", err)
	}
}

func linesOfSnapshot(abilities []types.AbilitySnapshot) []AbilityLine {
	out := make([]AbilityLine, 0, len(abilities))
	for _, a := range abilities {
		out = append(out, AbilityLine{
			Ability: a.Ability, GameCode: a.GameCode, Score: a.Score,
			Percentile: a.Percentile, Confidence: esportspb.PeerConfidence(esportspb.PeerConfidence_value[a.Confidence]),
			SampleCount: a.SampleCount, PeerAverage: a.PeerAverage,
		})
	}
	return out
}

func snapshotsOf(lines []AbilityLine) []types.AbilitySnapshot {
	out := make([]types.AbilitySnapshot, 0, len(lines))
	for _, l := range lines {
		out = append(out, types.AbilitySnapshot{
			Ability: l.Ability, GameCode: l.GameCode, Score: l.Score,
			Percentile: l.Percentile, Confidence: l.Confidence.String(),
			SampleCount: l.SampleCount, PeerAverage: l.PeerAverage,
		})
	}
	return out
}
