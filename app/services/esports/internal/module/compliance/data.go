package compliance

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/app/services/esports/internal/ent"
	"cyber-ecosystem/app/services/esports/internal/ent/complianceconfig"
	"cyber-ecosystem/app/services/esports/internal/platform"
	"cyber-ecosystem/app/services/esports/internal/shared"
)

// Repo ----------------------------------------------------------------------------------------------------------------

type complianceRP struct {
	shared.RP
}

func NewComplianceRP(logger *slog.Logger, p *platform.Platform) ComplianceRP {
	return &complianceRP{RP: shared.NewRP(logger.With("module", "module/compliance_rp"), p)}
}

// Method -------------------------------------------------------------------------------------------------------------

func (rp *complianceRP) SetBirthCollection(ctx context.Context, collect bool, updatedBy string) (*ComplianceConfig, error) {
	// Column-targeted upsert: the conflict update rewrites only this
	// knob's column, so two admins flipping different knobs concurrently
	// never clobber each other's answer. The insert path leans on the
	// schema's column defaults for the knobs this call never mentions.
	if err := rp.Platform.GetClient(ctx).ComplianceConfig.Create().
		SetSlot(complianceconfig.SlotSingleton).
		SetBirthCollect(collect).
		SetUpdatedBy(updatedBy).
		OnConflictColumns(complianceconfig.FieldSlot).
		Update(func(u *ent.ComplianceConfigUpsert) {
			u.SetUpdatedAt(time.Now())
			u.SetBirthCollect(collect)
			u.SetUpdatedBy(updatedBy)
		}).
		Exec(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return rp.Get(ctx)
}

func (rp *complianceRP) SetCompliancePrompts(ctx context.Context, minor, missing *string, updatedBy string) (*ComplianceConfig, error) {
	create := rp.Platform.GetClient(ctx).ComplianceConfig.Create().
		SetSlot(complianceconfig.SlotSingleton).
		SetUpdatedBy(updatedBy)
	if minor != nil {
		create = create.SetMinorPrompt(*minor)
	}
	if missing != nil {
		create = create.SetMissingBirthPrompt(*missing)
	}
	if err := create.
		OnConflictColumns(complianceconfig.FieldSlot).
		Update(func(u *ent.ComplianceConfigUpsert) {
			u.SetUpdatedAt(time.Now())
			u.SetUpdatedBy(updatedBy)
			if minor != nil {
				u.SetMinorPrompt(*minor)
			}
			if missing != nil {
				u.SetMissingBirthPrompt(*missing)
			}
		}).
		Exec(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return rp.Get(ctx)
}

func (rp *complianceRP) SetPayAgeLine(ctx context.Context, age int32, updatedBy string) (*ComplianceConfig, error) {
	if err := rp.Platform.GetClient(ctx).ComplianceConfig.Create().
		SetSlot(complianceconfig.SlotSingleton).
		SetPayAgeLine(int(age)).
		SetUpdatedBy(updatedBy).
		OnConflictColumns(complianceconfig.FieldSlot).
		Update(func(u *ent.ComplianceConfigUpsert) {
			u.SetUpdatedAt(time.Now())
			u.SetPayAgeLine(int(age))
			u.SetUpdatedBy(updatedBy)
		}).
		Exec(ctx); err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return rp.Get(ctx)
}

func (rp *complianceRP) Get(ctx context.Context) (*ComplianceConfig, error) {
	row, err := rp.Platform.GetClient(ctx).ComplianceConfig.Query().
		Where(complianceconfig.SlotEQ(complianceconfig.SlotSingleton)).
		Only(ctx)
	if ent.IsNotFound(err) {
		// Absent row = the shipped defaults answering; the table only
		// ever carries what an operator actually set.
		return &ComplianceConfig{BirthCollect: DefaultBirthCollect, PayAgeLine: DefaultPayAgeLine}, nil
	}
	if err != nil {
		return nil, rp.Platform.HandleEntError(err)
	}
	return mapConfig(row), nil
}

// Private ------------------------------------------------------------------------------------------------------------

func mapConfig(row *ent.ComplianceConfig) *ComplianceConfig {
	return &ComplianceConfig{
		BirthCollect:       row.BirthCollect,
		PayAgeLine:         utils.ConvNum[int32](row.PayAgeLine),
		MinorPrompt:        promptPtr(row.MinorPrompt),
		MissingBirthPrompt: promptPtr(row.MissingBirthPrompt),
	}
}

func promptPtr(v string) *string {
	// Empty is the column's "no override" — surfaced as nil so every face
	// falls back to its own localized default copy.
	if v == "" {
		return nil
	}
	return &v
}
