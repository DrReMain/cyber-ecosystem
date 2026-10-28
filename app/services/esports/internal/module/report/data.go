package report

import (
	"context"
	dbsql "database/sql"
	"errors"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"

	"cyber-ecosystem/app/services/esports/internal/ent"
	entmatchconfig "cyber-ecosystem/app/services/esports/internal/ent/matchconfig"
	entmatchsession "cyber-ecosystem/app/services/esports/internal/ent/matchsession"
	entprofile "cyber-ecosystem/app/services/esports/internal/ent/profile"
	entreportsnapshot "cyber-ecosystem/app/services/esports/internal/ent/reportsnapshot"
	enttest "cyber-ecosystem/app/services/esports/internal/ent/test"
	enttestrecord "cyber-ecosystem/app/services/esports/internal/ent/testrecord"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type boardRP struct {
	shared.RP
}

func NewRecordBoardRP(logger *slog.Logger, p *platform.Platform) RecordBoardRP {
	return &boardRP{RP: shared.NewRP(logger.With("module", "module/report_rp"), p)}
}

type snapshotRP struct {
	shared.RP
}

func NewSnapshotRP(logger *slog.Logger, p *platform.Platform) SnapshotRP {
	return &snapshotRP{RP: shared.NewRP(logger.With("module", "module/report_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *boardRP) Anchor(ctx context.Context, recordID string) (*Attempt, error) {
	row, err := rp.Platform.GetClient(ctx).TestRecord.Get(ctx, recordID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return &Attempt{
		ID:        row.ID,
		TestCode:  row.TestCode,
		UserID:    row.UserID,
		SessionID: row.SessionID,
		Valid:     row.Valid,
	}, nil
}

func (rp *boardRP) SessionBoard(ctx context.Context, sessionID string, anchor *Attempt) ([]BoardEntry, error) {
	// A session-less anchor is its own single-entry board; the join walks
	// two queries rather than one — the tests table is a dozen rows and
	// the ledger's read stays index-shaped.
	client := rp.Platform.GetClient(ctx)
	var rows []*ent.TestRecord
	if sessionID == "" {
		row, gerr := client.TestRecord.Get(ctx, anchor.ID)
		if gerr != nil {
			if ent.IsNotFound(gerr) {
				return nil, nil
			}
			return nil, rp.Platform.HandleEntError(gerr)
		}
		rows = []*ent.TestRecord{row}
	} else {
		var err error
		rows, err = client.TestRecord.Query().
			Where(enttestrecord.SessionIDEQ(sessionID), enttestrecord.ValidEQ(true)).
			Order(ent.Asc(enttestrecord.FieldCreatedAt)).
			All(ctx)
		if err != nil {
			return nil, rp.Platform.HandleEntError(err)
		}
	}
	return rp.boardOf(ctx, rows)
}

func (rp *boardRP) GenreOf(ctx context.Context, sessionID string, anchor *Attempt) (string, error) {
	client := rp.Platform.GetClient(ctx)
	if sessionID != "" {
		row, err := client.MatchSession.Query().
			Where(entmatchsession.IDEQ(sessionID)).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return "", nil
			}
			return "", rp.Platform.HandleEntError(err)
		}
		return row.Genre, nil
	}
	// A session-less anchor's genre is its single test's category.
	row, err := client.Test.Query().Where(enttest.CodeEQ(anchor.TestCode)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", rp.Platform.HandleEntError(err)
	}
	if row.Category == nil {
		return "", nil
	}
	return *row.Category, nil
}

func (rp *boardRP) GenreDimensions(ctx context.Context, genre string) (int, error) {
	// How many abilities the genre orders — read from match_config as the
	// distinct ability set, the completeness headline's denominator.
	dims, err := rp.Platform.GetClient(ctx).MatchConfig.Query().
		Where(entmatchconfig.GenreEQ(genre)).
		GroupBy(entmatchconfig.FieldAbility).
		Strings(ctx)
	if err != nil {
		return 0, rp.Platform.HandleEntError(err)
	}
	return len(dims), nil
}

func (rp *boardRP) Cohort(ctx context.Context, ability string, birthYear *int32, sex *string, ageBand *string) (CohortStats, string, error) {
	// The pool runs under the ruling's key: same birth year plus sex (an
	// undetermined or missing sex benches against the whole same-age pool),
	// degrading to the age band when the birth year is unknown — the
	// approximate rubric the basis flag names; the join rides ent's Modify.
	tr := sql.Table(enttestrecord.Table)
	t := sql.Table(enttest.Table)
	p := sql.Table(entprofile.Table)
	key := ""
	conds := []func(*sql.Selector){
		func(s *sql.Selector) { s.Where(sql.EQ(t.C(enttest.FieldAbility), ability)) },
	}
	if birthYear != nil {
		conds = append(conds, func(s *sql.Selector) { s.Where(sql.EQ(p.C(entprofile.FieldBirthYear), *birthYear)) })
		key = "same-age"
		if sex != nil && *sex != "" && *sex != shared.SexUndetermined {
			conds = append(conds, func(s *sql.Selector) { s.Where(sql.EQ(p.C(entprofile.FieldSex), *sex)) })
			key = "same-age-same-sex"
		}
	} else if ageBand != nil && *ageBand != "" {
		conds = append(conds, func(s *sql.Selector) { s.Where(sql.EQ(p.C(entprofile.FieldAgeBand), *ageBand)) })
		key = "band-approximate"
	}
	var stats CohortStats
	if key == "" {
		// No cohort key at all — no birth, no band: nothing to bench
		// against, and the caller renders score-only.
		return stats, "", nil
	}
	var rows []*struct {
		Avg   *float64 `json:"avg"`
		Std   *float64 `json:"std"`
		Count int      `json:"count"`
	}
	err := rp.Platform.GetClient(ctx).TestRecord.Query().
		Modify(func(s *sql.Selector) {
			s.Join(t).On(t.C(enttest.FieldCode), tr.C(enttestrecord.FieldTestCode)).
				Join(p).On(p.C(entprofile.FieldUserID), tr.C(enttestrecord.FieldUserID)).
				Where(sql.And(
					sql.EQ(tr.C(enttestrecord.FieldValid), true),
					sql.IsNull(tr.C(enttestrecord.FieldDeletedAt)),
					sql.IsNull(t.C(enttest.FieldDeletedAt)),
					sql.IsNull(p.C(entprofile.FieldDeletedAt)),
					sql.NotNull(tr.C(enttestrecord.FieldAuthoritativeScore)),
				))
			for _, c := range conds {
				c(s)
			}
			s.Select(
				sql.As("avg("+tr.C(enttestrecord.FieldAuthoritativeScore)+")", "avg"),
				sql.As("stddev_samp("+tr.C(enttestrecord.FieldAuthoritativeScore)+")", "std"),
				sql.As("count(*)", "count"),
			)
		}).
		Scan(ctx, &rows)
	if err != nil {
		return stats, key, rp.Platform.HandleEntError(err)
	}
	if len(rows) > 0 {
		stats.Avg, stats.Std, stats.SampleCount = rows[0].Avg, rows[0].Std, rows[0].Count
	}
	return stats, key, nil
}

func (rp *boardRP) PreviousSessionRecord(ctx context.Context, userID, ability string, before time.Time) (*BoardEntry, error) {
	// The last attempt from an earlier session, never a retry inside this one.
	client := rp.Platform.GetClient(ctx)
	codes, err := client.Test.Query().
		Where(enttest.AbilityEQ(ability)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if len(codes) == 0 {
		return nil, nil
	}
	codeList := make([]string, 0, len(codes))
	for _, t := range codes {
		codeList = append(codeList, t.Code)
	}
	row, err := client.TestRecord.Query().
		Where(
			enttestrecord.UserIDEQ(userID),
			enttestrecord.TestCodeIn(codeList...),
			enttestrecord.ValidEQ(true),
			enttestrecord.CreatedAtLT(before),
			enttestrecord.AuthoritativeScoreNotNil(),
		).
		Order(ent.Desc(enttestrecord.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return &BoardEntry{
		TestCode: row.TestCode,
		Ability:  ability,
		Score:    *row.AuthoritativeScore,
		At:       row.CreatedAt,
	}, nil
}

func (rp *boardRP) SessionMeta(ctx context.Context, sessionID string) (*SessionMetaRow, error) {
	row, err := rp.Platform.GetClient(ctx).MatchSession.Get(ctx, sessionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	// The done count rides the match face's own law: any stored record
	// marks its step done — validity never refunds a chance, and the two
	// faces must never disagree on a fraction.
	done, err := rp.doneCountOf(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	return &SessionMetaRow{
		ID: row.ID, UserID: row.UserID, Genre: row.Genre,
		Steps: row.Steps, DoneCount: done,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (rp *boardRP) SessionAttempts(ctx context.Context, sessionID string) ([]SessionAttemptRow, error) {
	// Every valid scored attempt under the session in submit order — the
	// receipt's raw truth; the ability rides the tests join the board
	// already uses.
	client := rp.Platform.GetClient(ctx)
	rows, err := client.TestRecord.Query().
		Where(
			enttestrecord.SessionIDEQ(sessionID),
			enttestrecord.ValidEQ(true),
			enttestrecord.AuthoritativeScoreNotNil(),
		).
		Order(ent.Asc(enttestrecord.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	abilityOf, err := rp.abilityOfCodes(ctx, rows)
	if err != nil {
		return nil, err
	}
	out := make([]SessionAttemptRow, 0, len(rows))
	for _, r := range rows {
		ability, ok := abilityOf[r.TestCode]
		if !ok {
			continue // a test outside the ability vocabulary feeds no line
		}
		var grade *string
		if r.Grade != nil {
			g := string(*r.Grade)
			grade = &g
		}
		out = append(out, SessionAttemptRow{
			TestCode: r.TestCode, Ability: ability, AttemptNo: r.AttemptNo,
			Score: *r.AuthoritativeScore, Grade: grade, CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}

func (rp *boardRP) AnchorOfSession(ctx context.Context, sessionID string) (*Attempt, error) {
	row, err := rp.Platform.GetClient(ctx).TestRecord.Query().
		Where(
			enttestrecord.SessionIDEQ(sessionID),
			enttestrecord.AuthoritativeScoreNotNil(),
		).
		Order(ent.Desc(enttestrecord.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return &Attempt{
		ID: row.ID, TestCode: row.TestCode, UserID: row.UserID,
		SessionID: row.SessionID, Valid: row.Valid,
	}, nil
}

func (rp *boardRP) AbilityBests(ctx context.Context, userID string) ([]DigestAbilityRow, error) {
	// Two rides over the records themselves: the ability ledger stays the
	// dormant E3 inheritance E5-b ruled (no business hand writes it), so
	// the cross-session bests aggregate the record-level authoritative
	// score the report board itself trusts — the ability rides the test
	// join, the same ownership shape the board's other reads use. Only
	// scored attempts carry a best (a played-but-unjudged round keeps its
	// ability honestly absent, the line law the session report holds).
	// First ride: per-ability best/sessions/last-at; second: the best's
	// own grade (the newest record at the best score wins). The
	// percentile stays honestly absent until the cohort face is a service
	// — the ledger's percentile column was its promised seat, never
	// written.
	tr := sql.Table(enttestrecord.Table)
	t := sql.Table(enttest.Table)
	client := rp.Platform.GetClient(ctx)

	var agg []*struct {
		Ability  string    `json:"ability"`
		Best     float64   `json:"best"`
		Sessions int       `json:"sessions"`
		LastAt   time.Time `json:"last_at"`
	}
	err := client.TestRecord.Query().
		Modify(func(s *sql.Selector) {
			s.Join(t).On(t.C(enttest.FieldCode), tr.C(enttestrecord.FieldTestCode)).
				Where(sql.And(
					sql.EQ(tr.C(enttestrecord.FieldUserID), userID),
					sql.EQ(tr.C(enttestrecord.FieldValid), true),
					sql.IsNull(tr.C(enttestrecord.FieldDeletedAt)),
					sql.NotNull(tr.C(enttestrecord.FieldAuthoritativeScore)),
				))
			s.Select(
				t.C(enttest.FieldAbility),
				sql.As("max("+tr.C(enttestrecord.FieldAuthoritativeScore)+")", "best"),
				sql.As("count(distinct "+tr.C(enttestrecord.FieldSessionID)+")", "sessions"),
				sql.As("max("+tr.C(enttestrecord.FieldCreatedAt)+")", "last_at"),
			).
				GroupBy(t.C(enttest.FieldAbility))
		}).
		Scan(ctx, &agg)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	if len(agg) == 0 {
		return nil, nil
	}

	out := make([]DigestAbilityRow, 0, len(agg))
	for _, a := range agg {
		row := DigestAbilityRow{Ability: a.Ability, Best: a.Best, Sessions: a.Sessions, LastAt: a.LastAt}
		var top []*struct {
			Grade *enttestrecord.Grade `json:"grade"`
		}
		err = client.TestRecord.Query().
			Modify(func(s *sql.Selector) {
				s.Join(t).On(t.C(enttest.FieldCode), tr.C(enttestrecord.FieldTestCode)).
					Where(sql.And(
						sql.EQ(tr.C(enttestrecord.FieldUserID), userID),
						sql.EQ(tr.C(enttestrecord.FieldValid), true),
						sql.IsNull(tr.C(enttestrecord.FieldDeletedAt)),
						sql.NotNull(tr.C(enttestrecord.FieldAuthoritativeScore)),
						sql.EQ(t.C(enttest.FieldAbility), a.Ability),
						sql.EQ(tr.C(enttestrecord.FieldAuthoritativeScore), a.Best),
					)).
					Select(tr.C(enttestrecord.FieldGrade)).
					OrderBy(sql.Desc(tr.C(enttestrecord.FieldCreatedAt)))
			}).
			Scan(ctx, &top)
		if err != nil {
			return nil, rp.Platform.HandleEntError(err)
		}
		if len(top) > 0 && top[0].Grade != nil {
			g := string(*top[0].Grade)
			row.BestGrade = &g
		}
		out = append(out, row)
	}
	return out, nil
}

func (rp *boardRP) SessionCounts(ctx context.Context, userID string) (int, int, int, error) {
	// Distinct sessions carrying a valid attempt, split per genre — the
	// join reads the session's own genre off the session row.
	tr := sql.Table(enttestrecord.Table)
	ms := sql.Table(entmatchsession.Table)
	var rows []*struct {
		Genre string `json:"genre"`
		Count int    `json:"count"`
	}
	err := rp.Platform.GetClient(ctx).TestRecord.Query().
		Modify(func(s *sql.Selector) {
			s.Join(ms).On(ms.C(entmatchsession.FieldID), tr.C(enttestrecord.FieldSessionID)).
				Where(sql.And(
					sql.EQ(tr.C(enttestrecord.FieldUserID), userID),
					sql.EQ(tr.C(enttestrecord.FieldValid), true),
					sql.IsNull(tr.C(enttestrecord.FieldDeletedAt)),
					sql.IsNull(ms.C(entmatchsession.FieldDeletedAt)),
				))
			s.Select(
				ms.C(entmatchsession.FieldGenre),
				sql.As("count(distinct "+tr.C(enttestrecord.FieldSessionID)+")", "count"),
			).
				GroupBy(ms.C(entmatchsession.FieldGenre))
		}).
		Scan(ctx, &rows)
	if err != nil {
		return 0, 0, 0, rp.Platform.HandleEntError(err)
	}
	total, fps, moba := 0, 0, 0
	for _, r := range rows {
		total += r.Count
		switch r.Genre {
		case shared.GenreFPS:
			fps = r.Count
		case shared.GenreMOBA:
			moba = r.Count
		}
	}
	return total, fps, moba, nil
}

func (rp *boardRP) LatestSessions(ctx context.Context, userID string) ([]SessionMetaRow, error) {
	// The newest session per genre that actually carried a valid attempt.
	// No window magic: the joined rows repeat per record, but the read
	// rides created_at descending, so the first row per genre already is
	// the newest — the seen-map cut below absorbs the repeats.
	tr := sql.Table(enttestrecord.Table)
	ms := sql.Table(entmatchsession.Table)
	var ids []*struct {
		ID    string `json:"id"`
		Genre string `json:"genre"`
	}
	err := rp.Platform.GetClient(ctx).MatchSession.Query().
		Modify(func(s *sql.Selector) {
			s.Join(tr).On(tr.C(enttestrecord.FieldSessionID), ms.C(entmatchsession.FieldID)).
				Where(sql.And(
					sql.EQ(ms.C(entmatchsession.FieldUserID), userID),
					sql.EQ(tr.C(enttestrecord.FieldUserID), userID),
					sql.EQ(tr.C(enttestrecord.FieldValid), true),
					sql.IsNull(tr.C(enttestrecord.FieldDeletedAt)),
					sql.IsNull(ms.C(entmatchsession.FieldDeletedAt)),
				)).
				Select(
					ms.C(entmatchsession.FieldID),
					ms.C(entmatchsession.FieldGenre),
				).
				OrderBy(sql.Desc(ms.C(entmatchsession.FieldCreatedAt)))
		}).
		Scan(ctx, &ids)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	seen := make(map[string]struct{}, 2)
	idList := make([]string, 0, 2)
	for _, r := range ids {
		if _, ok := seen[r.Genre]; ok {
			continue
		}
		seen[r.Genre] = struct{}{}
		idList = append(idList, r.ID)
	}
	if len(idList) == 0 {
		return nil, nil
	}
	rows, err := rp.Platform.GetClient(ctx).MatchSession.Query().
		Where(entmatchsession.IDIn(idList...)).
		All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := make([]SessionMetaRow, 0, len(rows))
	for _, row := range rows {
		done, derr := rp.doneCountOf(ctx, row.ID)
		if derr != nil {
			return nil, derr
		}
		out = append(out, SessionMetaRow{
			ID: row.ID, UserID: row.UserID, Genre: row.Genre,
			Steps: row.Steps, DoneCount: done, CreatedAt: row.CreatedAt,
		})
	}
	return out, nil
}

func (rp *boardRP) PreviousSessionScores(ctx context.Context, userID string, before time.Time) (map[string]float64, error) {
	// Per ability, the latest valid record older than the session's own
	// first attempt — everything this session wrote sits at or after it.
	// The records themselves carry the truth (the ability ledger stays the
	// dormant inheritance E5-b ruled); the rows are a member's whole
	// history (a handful), so the latest-per-ability cut runs in memory
	// over one ordered read.
	tr := sql.Table(enttestrecord.Table)
	t := sql.Table(enttest.Table)
	var rows []*struct {
		Ability string  `json:"ability"`
		Score   float64 `json:"score"`
	}
	err := rp.Platform.GetClient(ctx).TestRecord.Query().
		Modify(func(s *sql.Selector) {
			s.Join(t).On(t.C(enttest.FieldCode), tr.C(enttestrecord.FieldTestCode)).
				Where(sql.And(
					sql.EQ(tr.C(enttestrecord.FieldUserID), userID),
					sql.EQ(tr.C(enttestrecord.FieldValid), true),
					sql.IsNull(tr.C(enttestrecord.FieldDeletedAt)),
					sql.NotNull(tr.C(enttestrecord.FieldAuthoritativeScore)),
					sql.LT(tr.C(enttestrecord.FieldCreatedAt), before),
				)).
				Select(
					t.C(enttest.FieldAbility),
					sql.As(tr.C(enttestrecord.FieldAuthoritativeScore), "score"),
				).
				OrderBy(sql.Desc(tr.C(enttestrecord.FieldCreatedAt)))
		}).
		Scan(ctx, &rows)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := make(map[string]float64, len(rows))
	for _, r := range rows {
		if _, ok := out[r.Ability]; !ok {
			out[r.Ability] = r.Score
		}
	}
	return out, nil
}

func (rp *boardRP) doneCountOf(ctx context.Context, sessionID string) (int, error) {
	// Counts the session's steps that carry any stored record — the match
	// face's done law, so the two faces never disagree on a fraction.
	codes, err := rp.Platform.GetClient(ctx).TestRecord.Query().
		Where(
			enttestrecord.SessionIDEQ(sessionID),
		).
		GroupBy(enttestrecord.FieldTestCode).
		Strings(ctx)
	if err != nil {
		return 0, rp.Platform.HandleEntError(err)
	}
	return len(codes), nil
}

func (rp *boardRP) abilityOfCodes(ctx context.Context, rows []*ent.TestRecord) (map[string]string, error) {
	// Maps test codes to their ability — the tests table is a dozen rows,
	// the join rides two queries like the board's own.
	codes := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if _, ok := seen[r.TestCode]; !ok {
			seen[r.TestCode] = struct{}{}
			codes = append(codes, r.TestCode)
		}
	}
	tests, err := rp.Platform.GetClient(ctx).Test.Query().
		Where(enttest.CodeIn(codes...)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	out := make(map[string]string, len(tests))
	for _, t := range tests {
		if t.Ability != nil && *t.Ability != "" {
			out[t.Code] = *t.Ability
		}
	}
	return out, nil
}

func (rp *snapshotRP) FindByRecord(ctx context.Context, recordID string) (*Snapshot, error) {
	row, err := rp.Platform.GetClient(ctx).ReportSnapshot.Query().
		Where(entreportsnapshot.RecordIDEQ(recordID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return &Snapshot{
		RecordID:    row.RecordID,
		UserID:      row.UserID,
		Tier:        string(row.Tier),
		Composite:   row.CompositePercentile,
		Abilities:   row.Abilities,
		CohortBasis: row.CohortBasis,
	}, nil
}

func (rp *snapshotRP) Save(ctx context.Context, s *Snapshot) error {
	// ON CONFLICT DO NOTHING keeps the freeze idempotent: a racing freeze
	// (the event and a first read landing together) is a no-op statement,
	// never an error — ent surfaces the swallowed insert as ErrNoRows, and
	// that sentinel is the no-op answer itself, the Grant precedent.
	err := rp.Platform.GetClient(ctx).ReportSnapshot.Create().
		SetRecordID(s.RecordID).
		SetUserID(s.UserID).
		SetTier(entreportsnapshot.Tier(s.Tier)).
		SetNillableCompositePercentile(s.Composite).
		SetAbilities(s.Abilities).
		SetCohortBasis(s.CohortBasis).
		OnConflict().DoNothing().Exec(ctx)
	if err != nil && !errors.Is(err, dbsql.ErrNoRows) {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (rp *boardRP) boardOf(ctx context.Context, rows []*ent.TestRecord) ([]BoardEntry, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	client := rp.Platform.GetClient(ctx)
	codes := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if _, ok := seen[r.TestCode]; !ok {
			seen[r.TestCode] = struct{}{}
			codes = append(codes, r.TestCode)
		}
	}
	tests, err := client.Test.Query().Where(enttest.CodeIn(codes...)).All(ctx)
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	abilityOf := make(map[string]string, len(tests))
	for _, t := range tests {
		if t.Ability != nil {
			abilityOf[t.Code] = *t.Ability
		}
	}
	out := make([]BoardEntry, 0, len(rows))
	for _, r := range rows {
		if r.AuthoritativeScore == nil {
			continue // an unscored record feeds no axis
		}
		ability := abilityOf[r.TestCode]
		if ability == "" {
			continue // a test outside the ability vocabulary feeds no axis
		}
		out = append(out, BoardEntry{
			TestCode: r.TestCode,
			Ability:  ability,
			Score:    *r.AuthoritativeScore,
			At:       r.CreatedAt,
		})
	}
	return out, nil
}
