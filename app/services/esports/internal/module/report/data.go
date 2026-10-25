package report

import (
	"context"
	"log/slog"
	"time"

	"entgo.io/ent/dialect/sql"

	"cyber-ecosystem/app/services/esports/internal/ent"
	enttest "cyber-ecosystem/app/services/esports/internal/ent/test"
	enttestrecord "cyber-ecosystem/app/services/esports/internal/ent/testrecord"
	"cyber-ecosystem/app/services/esports/internal/module/test"
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

// Method --------------------------------------------------------------------------------------------------------------

func (rp *boardRP) Anchor(ctx context.Context, recordID string) (*test.AttemptRecord, error) {
	row, err := rp.Platform.GetClient(ctx).TestRecord.Get(ctx, recordID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, rp.Platform.HandleEntError(err)
	}
	return attemptOf(row), nil
}

func (rp *boardRP) SessionBoard(ctx context.Context, sessionID string, anchor *test.AttemptRecord) ([]BoardEntry, error) {
	client := rp.Platform.GetClient(ctx)
	var (
		rows []*ent.TestRecord
		err  error
	)
	// A session-less anchor is its own single-entry board; the join walks
	// two queries rather than one — the tests table is a dozen rows and
	// the ledger's read stays index-shaped.
	if sessionID == "" {
		rows, err = client.TestRecord.Query().
			Where(enttestrecord.IDEQ(anchor.ID)).All(ctx)
	} else {
		rows, err = client.TestRecord.Query().
			Where(enttestrecord.SessionIDEQ(sessionID), enttestrecord.ValidEQ(true)).
			Order(ent.Asc(enttestrecord.FieldCreatedAt)).
			All(ctx)
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return rp.boardOf(ctx, rows)
}

func (rp *boardRP) Cohort(ctx context.Context, ability string, birthYear *int32, sex *string, ageBand *string) (CohortStats, string, error) {
	// The pool runs under the ruling's key: same birth year plus sex (an
	// undetermined or missing sex benches against the whole same-age pool),
	// degrading to the age band when the birth year is unknown — the
	// approximate rubric the basis flag names; the join rides ent's Modify.
	key := ""
	conds := []func(*sql.Selector){
		func(s *sql.Selector) { s.Where(sql.EQ(sql.Table("test").C("ability"), ability)) },
	}
	if birthYear != nil {
		conds = append(conds, func(s *sql.Selector) { s.Where(sql.EQ(sql.Table("profile").C("birth_year"), *birthYear)) })
		key = "same-age"
		if sex != nil && *sex != "" && *sex != "undetermined" {
			conds = append(conds, func(s *sql.Selector) { s.Where(sql.EQ(sql.Table("profile").C("sex"), *sex)) })
			key = "same-age-same-sex"
		}
	} else if ageBand != nil && *ageBand != "" {
		conds = append(conds, func(s *sql.Selector) { s.Where(sql.EQ(sql.Table("profile").C("age_band"), *ageBand)) })
		key = "band-approximate"
	}
	var stats CohortStats
	if key == "" {
		// No cohort key at all — no birth, no band: nothing to bench
		// against, and the caller renders score-only.
		return stats, "", nil
	}
	client := rp.Platform.GetClient(ctx)
	var rows []*struct {
		Avg   *float64 `json:"avg"`
		Std   *float64 `json:"std"`
		Count int      `json:"count"`
	}
	err := client.TestRecord.Query().
		Modify(func(s *sql.Selector) {
			tr := sql.Table("test_record")
			t := sql.Table("test")
			p := sql.Table("profile")
			s.Join(t).On(t.C("code"), tr.C("test_code")).
				Join(p).On(p.C("user_id"), tr.C("user_id")).
				Where(sql.And(
					sql.EQ(tr.C("valid"), true),
					sql.IsNull(tr.C("deleted_at")),
					sql.IsNull(t.C("deleted_at")),
					sql.IsNull(p.C("deleted_at")),
					sql.NotNull(tr.C("authoritative_score")),
				))
			for _, c := range conds {
				c(s)
			}
			s.Select(
				sql.As("avg("+tr.C("authoritative_score")+")", "avg"),
				sql.As("stddev_samp("+tr.C("authoritative_score")+")", "std"),
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

func attemptOf(row *ent.TestRecord) *test.AttemptRecord {
	grade := (*string)(nil)
	if row.Grade != nil {
		g := string(*row.Grade)
		grade = &g
	}
	return &test.AttemptRecord{
		ID:            row.ID,
		TestCode:      row.TestCode,
		UserID:        row.UserID,
		AttemptNo:     row.AttemptNo,
		SessionID:     row.SessionID,
		StartedAt:     row.StartedAt,
		EndedAt:       row.EndedAt,
		Backgrounds:   row.Backgrounds,
		RawDetail:     row.RawDetail,
		RawScore:      row.RawScore,
		Authoritative: row.AuthoritativeScore,
		Grade:         grade,
		Valid:         row.Valid,
		InvalidReason: row.InvalidReason,
	}
}
