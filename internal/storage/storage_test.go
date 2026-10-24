package storage

import (
	"path/filepath"
	"testing"
	"time"

	"cryptomon/internal/model"
)

func open(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestAppendQueryAndTiers(t *testing.T) {
	st := open(t)
	now := time.Now().UnixMilli()
	// Old daily backfill in the hour tier and recent live samples.
	var daily []model.Point
	for i := 30; i >= 3; i-- {
		daily = append(daily, model.Point{T: now - int64(i)*86400_000, V: float64(100 + i)})
	}
	if err := st.Put(Hour, "agg:USD", daily); err != nil {
		t.Fatal(err)
	}
	var live []Sample
	for i := 0; i < 50; i++ {
		live = append(live, Sample{Series: "agg:USD", T: now - int64(50-i)*10_000, V: float64(i)})
	}
	if err := st.Append(live); err != nil {
		t.Fatal(err)
	}

	// 1h query: only live data from the raw tier.
	pts, err := st.Query("agg:USD", now-3600_000, now, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) != 50 || pts[0].V != 0 || pts[49].V != 49 {
		t.Fatalf("1h query returned %d points (%v…%v)", len(pts), pts[0], pts[len(pts)-1])
	}

	// 90d query: daily history followed by recent data, strictly ordered.
	pts, err = st.Query("agg:USD", now-90*86400_000, now, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if len(pts) < len(daily) {
		t.Fatalf("90d query returned %d points", len(pts))
	}
	for i := 1; i < len(pts); i++ {
		if pts[i].T <= pts[i-1].T {
			t.Fatalf("points not ordered at %d: %v then %v", i, pts[i-1], pts[i])
		}
	}

	// Downsampling keeps the end points.
	small, _ := st.Query("agg:USD", now-90*86400_000, now, 10)
	if len(small) != 10 || small[0] != pts[0] || small[9] != pts[len(pts)-1] {
		t.Fatalf("LTTB result %v", small)
	}
}

func TestValueAtAndLast(t *testing.T) {
	st := open(t)
	base := time.Now().Add(-time.Hour).UnixMilli()
	_ = st.Append([]Sample{{"s", base, 1}, {"s", base + 60_000, 2}, {"s", base + 120_000, 3}})
	if p, ok := st.ValueAt("s", base+90_000); !ok || p.V != 2 {
		t.Fatalf("ValueAt = %v %v", p, ok)
	}
	if _, ok := st.ValueAt("s", base-1); ok {
		t.Fatal("no value expected before the first sample")
	}
	if p, ok := st.Last("s"); !ok || p.V != 3 {
		t.Fatalf("Last = %v", p)
	}
	if _, ok := st.Last("missing"); ok {
		t.Fatal("missing series must report !ok")
	}
}

func TestPrune(t *testing.T) {
	st := open(t)
	old := time.Now().Add(-RawRetention - time.Hour).UnixMilli()
	_ = st.Append([]Sample{{"s", old, 1}, {"s", time.Now().UnixMilli(), 2}})
	if err := st.Prune(time.Now()); err != nil {
		t.Fatal(err)
	}
	pts, _ := st.Query("s", old-1, old+1, 10)
	// The raw sample is gone; the minute/hour aggregates stay.
	for _, p := range pts {
		if p.T == old {
			t.Fatal("expired raw sample survived pruning")
		}
	}
	if p, ok := st.ValueAt("s", old+60_000); !ok || p.V != 1 {
		t.Fatalf("minute aggregate lost: %v %v", p, ok)
	}
}

func TestMeta(t *testing.T) {
	st := open(t)
	if st.GetMeta("x") != 0 {
		t.Fatal("unset meta must be 0")
	}
	_ = st.SetMeta("x", 42)
	if st.GetMeta("x") != 42 {
		t.Fatal("meta not stored")
	}
}

func TestLTTBSmallInput(t *testing.T) {
	in := []model.Point{{T: 1, V: 1}, {T: 2, V: 2}}
	if out := LTTB(in, 100); len(out) != 2 {
		t.Fatal("short input must be returned as is")
	}
}
