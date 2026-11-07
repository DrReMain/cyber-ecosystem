package main

import (
	"database/sql"
	"testing"
	"time"
)

func ns(s string) sql.NullString { return sql.NullString{String: s, Valid: true} }

// MD5 users avoid the BCRYPT carry path entirely — these tests exercise
// the account-collision rule, not the password plane.
func acctUser(id int64, username, phone string, updated time.Time, live bool) *srcUser {
	return &srcUser{
		ID: id, Username: ns(username), Phone: ns(phone), PasswordAlgorithm: "MD5",
		Status: 1, IsDeleted: !live,
		CreatedAt: updated.Add(-time.Hour), UpdatedAt: updated,
	}
}

func planOf(t *testing.T, plans []*userPlan, id int64) *userPlan {
	t.Helper()
	for _, p := range plans {
		if p.src.ID == id {
			return p
		}
	}
	t.Fatalf("no plan for legacy id %d", id)
	return nil
}

// Ruling D-MIG-2's mechanical rule, pinned shape by shape: exactly one row
// of a collision group keeps the original account — live beats deleted,
// then latest updated_at, then the smaller legacy id — the rest take
// `<account>#<legacy_id>`, and the renamed plans write after every
// winner (the credential inserts are first-wins, so the account's owner
// must land its credential first).
func TestAccountCollisionRule(t *testing.T) {
	t0 := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		users   []*srcUser
		keepID  int64
		keepAcc string
		renamed map[int64]string // legacy id -> expected renamed handle
	}{
		{
			name:    "latest live row keeps the shared username",
			users:   []*srcUser{acctUser(10, "test", "", t0, true), acctUser(20, "test", "", t0.Add(time.Hour), true)},
			keepID:  20,
			keepAcc: "test",
			renamed: map[int64]string{10: "test#10"},
		},
		{
			name:    "live beats a newer-updated deleted row",
			users:   []*srcUser{acctUser(10, "test", "", t0, true), acctUser(11, "test", "", t0.Add(2*time.Hour), false)},
			keepID:  10,
			keepAcc: "test",
			renamed: map[int64]string{11: "test#11"},
		},
		{
			name:    "all-dead group: latest updated keeps the handle (pure bookkeeping)",
			users:   []*srcUser{acctUser(30, "todeleteuser", "", t0, false), acctUser(31, "todeleteuser", "", t0.Add(time.Hour), false)},
			keepID:  31,
			keepAcc: "todeleteuser",
			renamed: map[int64]string{30: "todeleteuser#30"},
		},
		{
			name:    "updated_at tie: smaller legacy id keeps",
			users:   []*srcUser{acctUser(41, "test2", "", t0, true), acctUser(40, "test2", "", t0, true)},
			keepID:  40,
			keepAcc: "test2",
			renamed: map[int64]string{41: "test2#41"},
		},
		{
			name: "phone-derived accounts collide the same way",
			users: []*srcUser{
				acctUser(50, "", "13100008167", t0, true),
				acctUser(51, "", "13100008167", t0.Add(time.Minute), true),
			},
			keepID:  51,
			keepAcc: "13100008167",
			renamed: map[int64]string{50: "13100008167#50"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plans, c, err := buildPlans(tc.users)
			if err != nil {
				t.Fatalf("buildPlans: %v", err)
			}
			winner := planOf(t, plans, tc.keepID)
			if winner.account != tc.keepAcc || winner.accountRenamed {
				t.Fatalf("winner id %d: account=%q renamed=%v, want %q unnamed", tc.keepID, winner.account, winner.accountRenamed, tc.keepAcc)
			}
			for id, want := range tc.renamed {
				p := planOf(t, plans, id)
				if p.account != want || !p.accountRenamed {
					t.Fatalf("loser id %d: account=%q renamed=%v, want %q renamed", id, p.account, p.accountRenamed, want)
				}
			}
			if c.renamedAccounts != len(tc.renamed) {
				t.Fatalf("census renamedAccounts = %d, want %d", c.renamedAccounts, len(tc.renamed))
			}
		})
	}
}

func TestNoCollisionLeavesAccountsAlone(t *testing.T) {
	t0 := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	users := []*srcUser{
		acctUser(1, "alice", "", t0, true),
		acctUser(2, "bob", "", t0, true),
		acctUser(3, "", "13900000001", t0, false),
	}
	plans, c, err := buildPlans(users)
	if err != nil {
		t.Fatalf("buildPlans: %v", err)
	}
	if c.renamedAccounts != 0 {
		t.Fatalf("renamedAccounts = %d, want 0", c.renamedAccounts)
	}
	if plans[0].account != "alice" || plans[1].account != "bob" || plans[2].account != "13900000001" {
		t.Fatalf("accounts reordered or rewritten: %q %q %q", plans[0].account, plans[1].account, plans[2].account)
	}
}

// The write order the credential inserts rely on: every renamed plan
// lands after every plan that kept its account.
func TestRenamedPlansWriteLast(t *testing.T) {
	t0 := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	users := []*srcUser{
		acctUser(1, "alice", "", t0, true),
		acctUser(10, "test", "", t0, true),
		acctUser(2, "bob", "", t0, true),
		acctUser(20, "test", "", t0.Add(time.Hour), true),
	}
	plans, _, err := buildPlans(users)
	if err != nil {
		t.Fatalf("buildPlans: %v", err)
	}
	lastWinner := -1
	firstRenamed := len(plans)
	for i, p := range plans {
		if p.accountRenamed {
			if i < firstRenamed {
				firstRenamed = i
			}
		} else {
			lastWinner = i
		}
	}
	if lastWinner > firstRenamed {
		t.Fatalf("renamed plan at %d writes before winner at %d", firstRenamed, lastWinner)
	}
}
