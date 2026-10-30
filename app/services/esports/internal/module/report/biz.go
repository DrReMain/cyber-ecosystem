package report

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"cyber-ecosystem/shared-go/kratos/security"
	"cyber-ecosystem/shared-go/utils"

	esportspb "cyber-ecosystem/gen/go/cyber/esports/v1"
	errorspb "cyber-ecosystem/gen/go/cyber/shared/errors/v1"

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

var abilityOrder = []string{ // the six-word vocabulary's display order — the tests table's own keys, the order every board and series renders by
	"reactionSpeed", "aimingAccuracy", "trackingStability",
	"skillReleaseAccuracy", "resourceAllocation", "decisionEfficiency",
}

var cognitiveModel = []cognitiveDimWeights{ // the frozen matrix's one transcription into code, values unchanged; a source ability with no cohort answer sinks its whole dimension, never a partial axis
	{Dimension: "neuralResponse", Weights: map[string]float64{
		"reactionSpeed": 0.6, "aimingAccuracy": 0.3, "trackingStability": 0.1,
	}},
	{Dimension: "sustainedAttention", Weights: map[string]float64{
		"trackingStability": 0.5, "resourceAllocation": 0.2, "decisionEfficiency": 0.3,
	}},
	{Dimension: "motorCoordination", Weights: map[string]float64{
		"aimingAccuracy": 0.5, "trackingStability": 0.3, "skillReleaseAccuracy": 0.2,
	}},
	{Dimension: "stressResilience", Weights: map[string]float64{
		"reactionSpeed": 0.2, "resourceAllocation": 0.3, "decisionEfficiency": 0.5,
	}},
	{Dimension: "resourceAllocation", Weights: map[string]float64{
		"trackingStability": 0.2, "resourceAllocation": 0.5, "decisionEfficiency": 0.3,
	}},
	{Dimension: "systemCognition", Weights: map[string]float64{
		"skillReleaseAccuracy": 0.2, "resourceAllocation": 0.4, "decisionEfficiency": 0.4,
	}},
	{Dimension: "logicDecision", Weights: map[string]float64{
		"reactionSpeed": 0.1, "resourceAllocation": 0.3, "decisionEfficiency": 0.6,
	}},
	{Dimension: "creativeThinking", Weights: map[string]float64{
		"skillReleaseAccuracy": 0.3, "resourceAllocation": 0.4, "decisionEfficiency": 0.3,
	}},
}

type cognitiveDimWeights struct { // Weights' keys are ability keys while Dimension is a dimension key — the shared spelling of resourceAllocation is two names in two namespaces, kept apart by structure
	Dimension string
	Weights   map[string]float64
}

var careerAreas = []careerAreaDef{ // the seeded area library: the legacy gates and advice transcribed with dimension keys carried onto the frozen matrix's own spelling; the two library facts are the industry-level statistics the corpus carries (youth-academy washout, the pro training load) — per-area figures wait on product data, and the values are rows here, not config
	{Area: "STEM", Name: "STEM领域",
		Gates:             map[string]float64{"logicDecision": 75, "resourceAllocation": 75},
		Score:             []string{"logicDecision", "resourceAllocation"},
		EliminationRate:   95,
		TrainingIntensity: "日均12-14小时",
		Suggestion:        "逻辑推理和认知资源分配能力突出，适合复杂问题解决和技术创新"},
	{Area: "CREATIVE_DESIGN", Name: "创意设计",
		Gates:             map[string]float64{"motorCoordination": 75, "creativeThinking": 70},
		Score:             []string{"motorCoordination"},
		EliminationRate:   95,
		TrainingIntensity: "日均12-14小时",
		Suggestion:        "视觉运动协调与创新思维结合，适合需要精细操作与创意的设计工作"},
	{Area: "MANAGEMENT", Name: "管理决策",
		Gates:             map[string]float64{"stressResilience": 75, "resourceAllocation": 75},
		Score:             []string{"stressResilience", "resourceAllocation"},
		EliminationRate:   95,
		TrainingIntensity: "日均12-14小时",
		Suggestion:        "心理韧性和多任务管理能力强，适合高压环境下的决策与领导工作"},
	{Area: "PRECISION_OPERATION", Name: "精密操作",
		Gates:             map[string]float64{"neuralResponse": 75, "motorCoordination": 75},
		Score:             []string{"neuralResponse", "motorCoordination"},
		EliminationRate:   95,
		TrainingIntensity: "日均12-14小时",
		Suggestion:        "神经反应速度和手眼协调优秀，适合需要快速精准响应的职业"},
}

type careerAreaDef struct { // Gates opens the area (every basis dimension at or past its own threshold); Score names the dimensions whose floor is the match — two separate shapes because one legacy area rides a single lead dimension rather than the gate's floor
	Area, Name        string
	Gates             map[string]float64
	Score             []string
	EliminationRate   float64
	TrainingIntensity string
	Suggestion        string
}

var careerPrescriptions = map[string][]string{ // the legacy per-dimension prescription copy, verbatim; resourceAllocation and systemCognition share their rows because the source's two rows are themselves identical
	"neuralResponse": {
		"每天进行10分钟快速反应训练（如红点触碰游戏）",
		"参加乒乓球、羽毛球等需要快速反应的体育运动",
		"增加物理/化学实验课的动手操作机会",
	},
	"sustainedAttention": {
		"使用番茄工作法逐步延长专注时间（25分钟专注+5分钟休息）",
		"每天阅读30分钟长篇文学作品，培养深度阅读习惯",
		"创造安静、无干扰的学习环境，减少电子设备干扰",
	},
	"motorCoordination": {
		"练习工整书写，每天抄写一页优美文章",
		"参加几何作图、美术绘画等需要精细操作的活动",
		"增加实验课程中精密仪器操作的练习机会",
	},
	"stressResilience": {
		"进行模拟考试训练，逐步适应考试压力环境",
		"学习深呼吸、正念冥想等压力管理技巧",
		"建立成长型思维，将错误视为学习机会而非失败",
	},
	"resourceAllocation": {
		"使用思维导图整理学科知识体系",
		"练习制定学习计划并合理分配时间和精力",
		"阅读历史、地理类书籍，培养宏观视野和系统观",
	},
	"systemCognition": {
		"使用思维导图整理学科知识体系",
		"练习制定学习计划并合理分配时间和精力",
		"阅读历史、地理类书籍，培养宏观视野和系统观",
	},
	"logicDecision": {
		"每天完成5道数学逻辑推理题，强化演绎思维",
		"尝试编程学习（如Python），培养算法思维",
		"参加辩论社或模拟法庭，训练逻辑论证能力",
	},
	"creativeThinking": {
		"每周写一篇创意作文，鼓励新颖独特的表达",
		"参加美术、音乐、戏剧等艺术类课程和活动",
		"鼓励'一题多解'，从多角度思考问题",
	},
}

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

type SessionAttemptRow struct {
	TestCode  string
	Ability   string
	AttemptNo int
	Score     float64 // authoritative — only a scored attempt lands a row, the settlement receipt's raw truth
	Grade     *string
	CreatedAt time.Time
}

type SessionAbilityRow struct {
	Ability     string
	GameCode    string
	Best        float64   // this session's best of the retries
	Attempts    []float64 // every valid attempt, in order — the receipt beside the verdict
	BestGrade   *string
	Percentile  *float64
	PrevSession *float64 // the same ability's last-session score, not this session's prior try
	HistoryBest *float64 // the cross-session best, the third number the settlement face keeps distinct
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

type SessionMetaRow struct {
	ID        string
	UserID    string
	Genre     string
	Steps     []types.MatchStep
	DoneCount int
	CreatedAt time.Time
}

type DigestAbilityRow struct {
	Ability    string
	Best       float64
	BestGrade  *string
	Percentile *float64
	Sessions   int
	LastAt     time.Time
}

type DigestSessionRow struct {
	SessionID string
	Genre     string
	Finished  bool
	DoneCount int
	StepCount int
	CreatedAt time.Time
	Anchor    *string // carried for the composite consumer — a digest line renders no verdict of its own
}

type AbilityDigest struct {
	SessionsTotal int
	SessionsFPS   int
	SessionsMOBA  int
	Abilities     []DigestAbilityRow
	Sessions      []DigestSessionRow
}

type CognitiveReport struct {
	Allowed    bool
	Reason     esportspb.ReportAccessReason
	Covered    int
	Tested     int
	Total      int
	Dimensions []CognitiveDimensionRow // only fully-fed dimensions; a missing dimension is an absent row, never a zero axis
	Composite  *float64
	Tier       string
}

type CognitiveDimensionRow struct {
	Dimension string
	Value     float64
	Tier      string
}

type TrendReport struct {
	Sessions        int
	Abilities       []TrendSeriesRow
	AdvancedAllowed bool // the advanced tier's gate — the base tier is free for anyone who has tested, the advanced content a later face
	AdvancedReason  esportspb.ReportAccessReason
}

type TrendSeriesRow struct {
	Ability string
	Points  []TrendPointRow
}

type TrendPointRow struct {
	At         time.Time
	Percentile float64
	Genre      string
}

type ScoredSessionRow struct { // one trend chronology entry — a session carrying at least one valid scored record; practice runs never qualify
	ID        string
	Genre     string
	CreatedAt time.Time
}

type judgement struct { // the live computation's whole answer — the freeze serializes it, the free tier renders it
	Lines     []AbilityLine
	Composite *float64
	Tier      string
	Basis     string
}

type CareerReport struct {
	Allowed bool
	Reason  esportspb.ReportAccessReason
	// Ordered by match, descending; areas whose basis dimensions went
	// unrendered (an unfed source sank them) state nothing — no gate,
	// no made-up number.
	Recommendations []CareerRecommendationRow
	Improvements    []CareerImprovementRow
}

type CareerRecommendationRow struct {
	Area, Name        string
	Matched           float64
	Basis             []string
	EliminationRate   float64
	TrainingIntensity string
	Suggestion        string
}

type CareerImprovementRow struct { // the weakest rendered dimensions first — the excellence line is the system's own tierA, the gap measured against it, never against a private constant
	Dimension   string
	Current     float64
	Target      float64
	GapPercent  int32
	Suggestions []string
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
	// ScoredSessions is every session of the member's carrying at least
	// one valid scored record, oldest first — trend's chronology; the
	// scored-record join is the gate, practice runs never qualify.
	ScoredSessions(ctx context.Context, userID string) ([]ScoredSessionRow, error)
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
	policyRP   entitle.ReportPolicyRP
	profileRP  profile.ProfileRP
	eventBus   shared.DomainEventBus
}

func NewReportUC(logger *slog.Logger, tm shared.Transaction, boardRP RecordBoardRP, snapshotRP SnapshotRP, entRP entitle.EntitlementRP, policyRP entitle.ReportPolicyRP, profileRP profile.ProfileRP, eventBus shared.DomainEventBus) *ReportUC {
	uc := &ReportUC{
		UC:         shared.NewUC(logger.With("module", "module/report"), tm),
		boardRP:    boardRP,
		snapshotRP: snapshotRP,
		entRP:      entRP,
		policyRP:   policyRP,
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

type MergeUC struct {
	shared.UC
	systemRP profile.SystemRP // who holds the credential, and the hard-gate verdict
	boardRP  RecordBoardRP    // the emptiness bar's session facts, this module's own
	entRP    entitle.EntitlementRP
}

func NewMergeUC(logger *slog.Logger, tm shared.Transaction, systemRP profile.SystemRP, boardRP RecordBoardRP, entRP entitle.EntitlementRP) *MergeUC {
	return &MergeUC{
		UC:       shared.NewUC(logger.With("module", "module/report"), tm),
		systemRP: systemRP,
		boardRP:  boardRP,
		entRP:    entRP,
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

func (uc *ReportUC) GetSessionReport(ctx context.Context, sessionID string) (*SessionReport, error) {
	// The session read-back face: the lineup's own order, every ability's
	// three-number verdict, and the anchor's composite under the same law
	// the record board follows — frozen when a snapshot exists, live
	// otherwise. A session with no scored attempt is an honest empty
	// board, never an error.
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

func (uc *ReportUC) GetAbilityDigest(ctx context.Context) (*AbilityDigest, error) {
	// The cross-session aggregate: session counts by genre, the per-ability
	// best lines, and the latest session line per genre. An empty history
	// is an honest zero-and-nothing answer.
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
	// The percentile column's promised seat, filled at last: the best
	// score benched read-time under the same cohort keys the live
	// judgement uses. The honest null stays whenever the cohort has
	// nothing to say about the best — no made-up percentile.
	if len(d.Abilities) > 0 {
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
		for i := range d.Abilities {
			pct, _, _, _, _ := uc.percentileFor(ctx, BoardEntry{Ability: d.Abilities[i].Ability, Score: d.Abilities[i].Best}, birthYear, sex, ageBand)
			d.Abilities[i].Percentile = pct
		}
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

func (uc *ReportUC) GetCognitiveReport(ctx context.Context) (*CognitiveReport, error) {
	// The eight-dimension aggregate: each dimension the weighted sum of
	// its source abilities' history-best percentiles. The verdict rides
	// in first — cognitive ships membership-gated, so an unentitled
	// member gets the covered count and the gate's own reason, and the
	// numbers behind the gate never travel.
	subject, allowed, reason, err := uc.reportAccessOf(ctx, shared.ReportCognitive)
	if err != nil {
		return nil, err
	}
	r := &CognitiveReport{Total: len(cognitiveModel), Allowed: allowed, Reason: reason}
	bests, err := uc.boardRP.AbilityBests(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	r.Covered = len(bests) // the digest's own public fact, gate or no gate
	if !r.Allowed {
		return r, nil
	}
	dims, err := uc.renderedDims(ctx, subject, bests)
	if err != nil {
		return nil, err
	}
	r.Dimensions = dims
	r.Tested = len(r.Dimensions)
	if len(r.Dimensions) > 0 {
		vals := make([]float64, 0, len(r.Dimensions))
		for _, d := range r.Dimensions {
			vals = append(vals, d.Value)
		}
		if comp := compositeOf(vals); comp != nil {
			c := math.Round(*comp)
			r.Composite, r.Tier = &c, tierOf(c)
		}
	}
	return r, nil
}

func (uc *ReportUC) GetTrendReport(ctx context.Context) (*TrendReport, error) {
	// The base tier's chronology: one series per ability, each point the
	// session-anchor's cohort answer at that time. The advanced gate is
	// judged unconditionally — an empty history still renders its locked
	// state from one call.
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, esportspb.ErrorEsportsReportRecordNotOwned("").WithCause(fmt.Errorf("trend report read requires a session"))
	}
	r := &TrendReport{}
	if active, aerr := uc.entRP.HasActiveMembership(ctx, subject.UserID); aerr != nil {
		return nil, aerr
	} else if active {
		r.AdvancedAllowed, r.AdvancedReason = true, esportspb.ReportAccessReason_REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE
	} else {
		r.AdvancedReason = esportspb.ReportAccessReason_REPORT_ACCESS_REASON_NO_ENTITLEMENT
	}
	if total, _, _, serr := uc.boardRP.SessionCounts(ctx, subject.UserID); serr != nil {
		return nil, serr
	} else {
		r.Sessions = total
	}
	if r.Sessions == 0 {
		return r, nil // the honest empty page
	}

	metas, err := uc.boardRP.ScoredSessions(ctx, subject.UserID)
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
	byAbility := make(map[string][]TrendPointRow)
	for _, meta := range metas {
		for ability, p := range uc.anchorPercentilesOf(ctx, meta, birthYear, sex, ageBand) {
			if p == nil {
				continue // a point exists only where the cohort spoke
			}
			byAbility[ability] = append(byAbility[ability], TrendPointRow{At: meta.CreatedAt, Percentile: *p, Genre: meta.Genre})
		}
	}
	for _, ability := range abilityOrder {
		if pts, ok := byAbility[ability]; ok && len(pts) > 0 {
			r.Abilities = append(r.Abilities, TrendSeriesRow{Ability: ability, Points: pts})
		}
	}
	return r, nil
}

func (uc *ReportUC) GetCareerReport(ctx context.Context) (*CareerReport, error) {
	// Career rides open by the standing default: the seeded area library
	// judged against the same rendered dimensions the cognitive face
	// shows, the weakest of those dimensions carried onto the
	// prescription board against the system's own excellence line.
	// Behind a closed gate neither the areas nor the prescriptions
	// travel — the verdict alone.
	subject, allowed, reason, err := uc.reportAccessOf(ctx, shared.ReportCareer)
	if err != nil {
		return nil, err
	}
	r := &CareerReport{Allowed: allowed, Reason: reason}
	if !r.Allowed {
		return r, nil
	}
	bests, err := uc.boardRP.AbilityBests(ctx, subject.UserID)
	if err != nil {
		return nil, err
	}
	dims, err := uc.renderedDims(ctx, subject, bests)
	if err != nil {
		return nil, err
	}
	if len(dims) == 0 {
		return r, nil // never-tested renders the honest empty page
	}
	byDim := make(map[string]float64, len(dims))
	for _, d := range dims {
		byDim[d.Dimension] = d.Value
	}
	for _, area := range careerAreas {
		open := true
		for dim, threshold := range area.Gates {
			if v, ok := byDim[dim]; !ok || v < threshold {
				open = false // an unrendered basis dimension is an unjudgeable area, never a zero axis
				break
			}
		}
		if !open {
			continue
		}
		matched := math.Inf(1)
		for _, dim := range area.Score {
			matched = math.Min(matched, byDim[dim])
		}
		r.Recommendations = append(r.Recommendations, CareerRecommendationRow{
			Area: area.Area, Name: area.Name, Matched: matched, Basis: area.Score,
			EliminationRate: area.EliminationRate, TrainingIntensity: area.TrainingIntensity,
			Suggestion: area.Suggestion,
		})
	}
	sort.SliceStable(r.Recommendations, func(i, j int) bool { // the contract's own ordering: by match, descending
		return r.Recommendations[i].Matched > r.Recommendations[j].Matched
	})
	weakest := append([]CognitiveDimensionRow(nil), dims...)
	sort.SliceStable(weakest, func(i, j int) bool { return weakest[i].Value < weakest[j].Value })
	if len(weakest) > 4 {
		weakest = weakest[:4]
	}
	for _, d := range weakest {
		short := d.Value < targetPercentile
		gap := 0.0
		if short { // an already-excellent row states a zero gap, never a negative one
			gap = math.Round((targetPercentile - d.Value) / targetPercentile * 100)
		}
		suggestions := []string{"已达优秀水平"}
		if short {
			if rows, ok := careerPrescriptions[d.Dimension]; ok {
				suggestions = rows
			}
		}
		r.Improvements = append(r.Improvements, CareerImprovementRow{
			Dimension: d.Dimension, Current: d.Value, Target: targetPercentile,
			GapPercent: utils.ConvNum[int32](gap), Suggestions: suggestions,
		})
	}
	return r, nil
}

func (uc *ReportUC) reportAccessOf(ctx context.Context, family string) (*security.Subject, bool, esportspb.ReportAccessReason, error) {
	// The gate, folded from the three facts through the shared order —
	// the same verdict the entitle face answers, so the faces cannot
	// drift apart. Cross-module reads ride the owning module's repo
	// ports; the order itself lives in the kernel, not here.
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok || subject == nil {
		return nil, false, 0, esportspb.ErrorEsportsReportRecordNotOwned("").WithCause(fmt.Errorf("report access requires a session"))
	}
	ref := family
	grant, gerr := uc.entRP.ByScope(ctx, subject.UserID, shared.ScopeReportType, &ref)
	if gerr != nil {
		return nil, false, 0, gerr
	}
	codeActive := grant != nil && (grant.ExpiresAt == nil || grant.ExpiresAt.After(time.Now()))
	policy, rowPresent, perr := uc.policyRP.Policy(ctx, family)
	if perr != nil {
		return nil, false, 0, perr
	}
	memberActive, merr := uc.entRP.HasActiveMembership(ctx, subject.UserID)
	if merr != nil {
		return nil, false, 0, merr
	}
	allowed, reason := shared.ReportAccessVerdict(family, codeActive, rowPresent, policy, memberActive)
	return subject, allowed, reason, nil
}

func (uc *ReportUC) renderedDims(ctx context.Context, subject *security.Subject, bests []DigestAbilityRow) ([]CognitiveDimensionRow, error) {
	// The frozen matrix's input: each measured ability's best score
	// benched read-time — the snapshot stays the only frozen source, and
	// an unbenchable best (no cohort key, thin pool) feeds nothing.
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
	pctByAbility := make(map[string]float64, len(bests))
	for _, b := range bests {
		pct, _, _, _, _ := uc.percentileFor(ctx, BoardEntry{Ability: b.Ability, Score: b.Best}, birthYear, sex, ageBand)
		if pct != nil {
			pctByAbility[b.Ability] = *pct
		}
	}
	var dims []CognitiveDimensionRow
	for _, dim := range cognitiveModel {
		sum, fed := 0.0, true
		for ability, w := range dim.Weights {
			p, ok := pctByAbility[ability]
			if !ok {
				fed = false
				break // one unfed source sinks the whole dimension, never a partial axis
			}
			sum += p * w
		}
		if !fed {
			continue
		}
		v := math.Round(sum)
		dims = append(dims, CognitiveDimensionRow{Dimension: dim.Dimension, Value: v, Tier: tierOf(v)})
	}
	return dims, nil
}

func (uc *ReportUC) anchorPercentilesOf(ctx context.Context, meta ScoredSessionRow, birthYear *int32, sex, ageBand *string) map[string]*float64 {
	// One session's anchor verdict read the way the read-back faces do:
	// the frozen snapshot when one exists — the numbers the unlock
	// bought, pinned at unlock time — and the live judgement otherwise,
	// the free tier's own law in aggregate form. Trend pins nothing; a
	// snapshot that held nulls keeps its nulls, because a later fatter
	// pool is today's answer, not the answer at that time.
	anchor, err := uc.boardRP.AnchorOfSession(ctx, meta.ID)
	if err != nil || anchor == nil {
		return nil
	}
	if snap, serr := uc.snapshotRP.FindByRecord(ctx, anchor.ID); serr == nil && snap != nil {
		out := make(map[string]*float64, len(snap.Abilities))
		for _, a := range snap.Abilities {
			if a.Percentile != nil {
				out[a.Ability] = a.Percentile
			}
		}
		return out
	}
	attempts, err := uc.boardRP.SessionAttempts(ctx, meta.ID)
	if err != nil {
		return nil
	}
	j := uc.computeFor(ctx, sessionBoardOf(attempts), birthYear, sex, ageBand)
	out := make(map[string]*float64, len(j.Lines))
	for _, l := range j.Lines {
		out[l.Ability] = l.Percentile
	}
	return out
}

func (uc *ReportUC) sessionPercentiles(ctx context.Context, anchorID string) map[string]*float64 {
	// Reads the anchor's per-ability cohort answer from the frozen
	// snapshot when one exists; the live branch of the caller fills the
	// lines itself.
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

func sessionBoardOf(attempts []SessionAttemptRow) []BoardEntry {
	// Flattens the attempt rows into the board shape the live judgement
	// consumes — folded to one entry per ability by the settlement's own
	// law, so a retaken game's retries never each weigh on the composite
	// (0/0/100 would average as a 33rd percentile otherwise — the same
	// doubled-axis fault the record board's dedupe fixed).
	out := make([]BoardEntry, 0, len(attempts))
	for _, a := range attempts {
		out = append(out, BoardEntry{TestCode: a.TestCode, Ability: a.Ability, Score: a.Score, At: a.CreatedAt})
	}
	return dedupeBoard(out)
}

func dedupeBoard(board []BoardEntry) []BoardEntry {
	// Folds the board to one entry per ability — the settlement's own law
	// (the best attempt of the retries is the one that scores), so a
	// retaken game never draws a doubled radar axis or a duplicated
	// ability row on the record board either. Equal scores keep the newer
	// attempt.
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

func (uc *MergeUC) Preview(ctx context.Context, cred *profile.MergeCredential) (*profile.MergeOffer, error) {
	// Hosted here, not in profile: the emptiness bar reads this module's
	// session facts, and report already imports profile + entitle — the
	// seam back is the profile.MergePreviewer interface. The verdict names
	// the direction the client must execute.
	subject, ok := security.SubjectFromCtx(ctx)
	if !ok {
		return nil, errorspb.ErrorGeneralErrorUnauthenticated("").WithCause(fmt.Errorf("merge preview requires a session"))
	}
	sec, err := uc.systemRP.SecurityStatus(ctx)
	if err != nil {
		return nil, err
	}
	caller, err := uc.summary(ctx, subject.UserID, sec.PhoneMask, sec.Wechat)
	if err != nil {
		return nil, err
	}
	other, err := uc.systemRP.ResolveCredential(ctx, cred)
	if err != nil {
		return nil, err
	}
	switch {
	case other == nil:
		return &profile.MergeOffer{Status: esportspb.MergePreviewStatus_MERGE_PREVIEW_STATUS_FREE, Caller: caller}, nil
	case other.UserID == subject.UserID:
		return &profile.MergeOffer{Status: esportspb.MergePreviewStatus_MERGE_PREVIEW_STATUS_SELF, Caller: caller, Other: caller}, nil
	}
	otherSummary, err := uc.summary(ctx, other.UserID, other.PhoneMask, other.Wechat)
	if err != nil {
		return nil, err
	}
	// The corpus bar is a data-safety bar: a side may only be absorbed
	// when it carries no business weight. A doubly empty pair rides the
	// absorb-other branch — two empty sides have no data to dispute, and
	// the caller keeps (the seat the user is sitting in); the blocked
	// shape is reserved for the genuinely both-weighted pair its copy
	// describes. The system weight gate is direction-aware: it binds the
	// side being absorbed, never the survivor — a password on the rich
	// side is its normal attribute.
	callerEmpty, otherEmpty := mergeEmpty(caller), mergeEmpty(otherSummary)
	blocked := &profile.MergeOffer{Status: esportspb.MergePreviewStatus_MERGE_PREVIEW_STATUS_BLOCKED, Caller: caller, Other: otherSummary}
	switch {
	case otherEmpty:
		// The other side is absorbed: its system weight must pass.
		if !other.Mergeable {
			return blocked, nil
		}
		return &profile.MergeOffer{Status: esportspb.MergePreviewStatus_MERGE_PREVIEW_STATUS_KEEP_CALLER, Caller: caller, Other: otherSummary}, nil
	case callerEmpty:
		// The caller is absorbed: a password or a second credential would
		// fail the merge at the cash. Seat weight stays the executor's own
		// check — this is UX honesty, not the law.
		if sec.HasPassword || sec.CredentialCount > 1 {
			return blocked, nil
		}
		return &profile.MergeOffer{Status: esportspb.MergePreviewStatus_MERGE_PREVIEW_STATUS_KEEP_OTHER, Caller: caller, Other: otherSummary}, nil
	default:
		return blocked, nil
	}
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

func (uc *MergeUC) summary(ctx context.Context, userID string, phoneMask *string, wechat bool) (*profile.MergeSummary, error) {
	sessions, _, _, err := uc.boardRP.SessionCounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	window, err := uc.entRP.MembershipWindow(ctx, userID)
	if err != nil {
		return nil, err
	}
	unlocked, err := uc.entRP.HasAnyScope(ctx, userID, shared.ScopeReportType)
	if err != nil {
		return nil, err
	}
	s := &profile.MergeSummary{UserID: userID, PhoneMask: phoneMask, Wechat: wechat, Sessions: sessions, Unlocked: unlocked}
	if window != nil {
		s.MemberUntil = window.ExpiresAt
	}
	return s, nil
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

func mergeEmpty(s *profile.MergeSummary) bool {
	// The corpus emptiness bar over business records: no sessions, no live
	// membership, no unlocked family. The invite leg of the corpus test
	// has no table yet — vacuously true until the growth slice lands.
	return s.Sessions == 0 && s.MemberUntil == nil && !s.Unlocked
}
