package authz

import (
	"context"
	"fmt"
	"maps"

	kauthz "cyber-ecosystem/shared-go/kratos/security/authz"

	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/ent/dept"
)

func (r *authzRP) hydrateScopes(ctx context.Context, userID string, views []GrantView) ([]GrantView, error) {
	needsTree := false
	for _, v := range views {
		if v.ScopeKind == kauthz.ScopeKindDeptTree {
			needsTree = true
			break
		}
	}
	if !needsTree {
		return views, nil
	}
	client := r.Platform.GetClient(ctx)
	u, err := client.User.Get(ctx, userID)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz hydrate: load subject: %w", err))
	}
	var subtree []string
	if u.DeptID != nil && *u.DeptID != "" {
		if subtree, err = r.resolveDeptSubtree(ctx, client, *u.DeptID); err != nil {
			return nil, err
		}
	}
	for i := range views {
		if views[i].ScopeKind != kauthz.ScopeKindDeptTree {
			continue
		}
		// Fresh map per view: ScopeParams may alias the snapshot's grantMeta
		// map, and the subtree is subject-specific.
		params := make(map[string]any, len(views[i].ScopeParams)+1)
		maps.Copy(params, views[i].ScopeParams)
		params[kauthz.ScopeParamDeptIDs] = subtree
		views[i].ScopeParams = params
	}
	return views, nil
}

func (r *authzRP) resolveDeptSubtree(ctx context.Context, client *ent.Client, deptID string) ([]string, error) {
	rows, err := client.Dept.Query().Select(dept.FieldID, dept.FieldParentID).All(ctx)
	if err != nil {
		return nil, r.Platform.HandleEntError(fmt.Errorf("authz hydrate: load dept tree: %w", err))
	}
	children := make(map[string][]string, len(rows))
	for _, d := range rows {
		if d.ParentID != nil && *d.ParentID != "" {
			children[*d.ParentID] = append(children[*d.ParentID], d.ID)
		}
	}
	subtree := []string{deptID}
	seen := map[string]struct{}{deptID: {}}
	for i := 0; i < len(subtree); i++ {
		for _, child := range children[subtree[i]] {
			if _, ok := seen[child]; !ok {
				seen[child] = struct{}{}
				subtree = append(subtree, child)
			}
		}
	}
	return subtree, nil
}
