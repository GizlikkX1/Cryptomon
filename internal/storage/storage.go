// Package storage is a small tiered time-series store on top of bbolt
// (pure Go, no CGO). Each microservice owns its own database file.
//
// Every sample is written to three tiers:
//
//	raw — every sample, kept for RawRetention
//	m1  — last value per minute, kept for MinuteRetention
//	h1  — last value per hour, kept forever (daily backfill lives here too)
//
// Queries read the finest tier that suits the requested span and fall back
// to coarser tiers for older data, then downsample with LTTB.
package storage

import (
	"encoding/binary"
	"errors"
	"math"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"

	"cryptomon/internal/model"
)

type Tier int

const (
	Raw Tier = iota
	Minute
	Hour
)

var tierBuckets = [...][]byte{[]byte("raw"), []byte("m1"), []byte("h1")}

// tierStep is the bucket width of each tier in milliseconds.
var tierStep = [...]int64{1, 60_000, 3_600_000}

const (
	RawRetention    = 48 * time.Hour
	MinuteRetention = 62 * 24 * time.Hour
)

var metaBucket = []byte("meta")

type Store struct {
	db *bolt.DB
}

// Sample is one value of one series.
type Sample struct {
	Series string
	T      int64 // unix ms
	V      float64
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return nil, errors.New("database " + path + " is locked by another process")
		}
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range tierBuckets {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		_, err := tx.CreateBucketIfNotExists(metaBucket)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func key(t int64) []byte {
	var k [8]byte
	binary.BigEndian.PutUint64(k[:], uint64(t))
	return k[:]
}

func val(v float64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], math.Float64bits(v))
	return b[:]
}

func decode(k, v []byte) model.Point {
	return model.Point{T: int64(binary.BigEndian.Uint64(k)), V: math.Float64frombits(binary.BigEndian.Uint64(v))}
}

func floor(t int64, step time.Duration) int64 {
	ms := step.Milliseconds()
	return t - t%ms
}

// Append writes live samples into all tiers in a single transaction.
func (s *Store) Append(samples []Sample) error {
	if len(samples) == 0 {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, smp := range samples {
			if math.IsNaN(smp.V) || math.IsInf(smp.V, 0) {
				continue
			}
			keys := [...]int64{smp.T, floor(smp.T, time.Minute), floor(smp.T, time.Hour)}
			for tier, t := range keys {
				b, err := tx.Bucket(tierBuckets[tier]).CreateBucketIfNotExists([]byte(smp.Series))
				if err != nil {
					return err
				}
				if err := b.Put(key(t), val(smp.V)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Put writes historical points (backfill) into one tier.
func (s *Store) Put(tier Tier, series string, pts []model.Point) error {
	if len(pts) == 0 {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.Bucket(tierBuckets[tier]).CreateBucketIfNotExists([]byte(series))
		if err != nil {
			return err
		}
		for _, p := range pts {
			if p.V > 0 && !math.IsInf(p.V, 0) {
				if err := b.Put(key(p.T), val(p.V)); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func rangeTier(tx *bolt.Tx, tier Tier, series string, from, to int64) []model.Point {
	b := tx.Bucket(tierBuckets[tier]).Bucket([]byte(series))
	if b == nil {
		return nil
	}
	var pts []model.Point
	c := b.Cursor()
	for k, v := c.Seek(key(from)); k != nil; k, v = c.Next() {
		p := decode(k, v)
		if p.T > to {
			break
		}
		pts = append(pts, p)
	}
	return pts
}

// Query returns up to maxPoints points of series in [from, to].
func (s *Store) Query(series string, from, to int64, maxPoints int) ([]model.Point, error) {
	span := time.Duration(to-from) * time.Millisecond
	start := Hour
	switch {
	case span <= RawRetention:
		start = Raw
	case span <= MinuteRetention:
		start = Minute
	}
	var out []model.Point
	err := s.db.View(func(tx *bolt.Tx) error {
		for tier := start; tier <= Hour; tier++ {
			end := to
			if len(out) > 0 {
				// A coarser bucket is keyed by its start but holds its last
				// value, so skip buckets that overlap the finer data.
				end = out[0].T - tierStep[tier]
			}
			if end < from {
				break
			}
			out = append(rangeTier(tx, tier, series, from, end), out...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return LTTB(out, maxPoints), nil
}

// ValueAt returns the latest value known at time t, looking through all
// tiers. A minute/hour bucket only counts once it has closed before t,
// because it holds the last value of its whole interval.
func (s *Store) ValueAt(series string, t int64) (model.Point, bool) {
	var best model.Point
	var bestEnd int64
	found := false
	_ = s.db.View(func(tx *bolt.Tx) error {
		for tier := Raw; tier <= Hour; tier++ {
			b := tx.Bucket(tierBuckets[tier]).Bucket([]byte(series))
			if b == nil {
				continue
			}
			step := tierStep[tier]
			limit := t
			if t < math.MaxInt64-step {
				limit = t - step + 1
			}
			c := b.Cursor()
			k, v := c.Seek(key(limit))
			if k == nil || int64(binary.BigEndian.Uint64(k)) > limit {
				k, v = c.Prev()
			}
			if k == nil {
				continue
			}
			p := decode(k, v)
			if end := p.T + step - 1; !found || end > bestEnd {
				best, bestEnd, found = p, end, true
			}
		}
		return nil
	})
	return best, found
}

// Last returns the most recent point of a series.
func (s *Store) Last(series string) (model.Point, bool) {
	return s.ValueAt(series, math.MaxInt64)
}

// Prune drops expired raw and minute samples.
func (s *Store) Prune(now time.Time) error {
	limits := map[Tier]int64{
		Raw:    now.Add(-RawRetention).UnixMilli(),
		Minute: now.Add(-MinuteRetention).UnixMilli(),
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		for tier, limit := range limits {
			root := tx.Bucket(tierBuckets[tier])
			var names [][]byte
			_ = root.ForEachBucket(func(k []byte) error {
				names = append(names, append([]byte(nil), k...))
				return nil
			})
			for _, name := range names {
				c := root.Bucket(name).Cursor()
				for k, _ := c.First(); k != nil && int64(binary.BigEndian.Uint64(k)) < limit; k, _ = c.Next() {
					if err := c.Delete(); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

// GetMeta / SetMeta keep small bookkeeping values (e.g. last backfill time).
func (s *Store) GetMeta(k string) int64 {
	var v int64
	_ = s.db.View(func(tx *bolt.Tx) error {
		if b := tx.Bucket(metaBucket).Get([]byte(k)); len(b) == 8 {
			v = int64(binary.BigEndian.Uint64(b))
		}
		return nil
	})
	return v
}

func (s *Store) SetMeta(k string, v int64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(metaBucket).Put([]byte(k), key(v))
	})
}

// LTTB downsamples with the Largest-Triangle-Three-Buckets algorithm, which
// keeps the visual shape (peaks and dips) of a line chart.
func LTTB(data []model.Point, threshold int) []model.Point {
	n := len(data)
	if threshold <= 2 || n <= threshold {
		return data
	}
	out := make([]model.Point, 0, threshold)
	every := float64(n-2) / float64(threshold-2)
	a := 0
	out = append(out, data[0])
	for i := 0; i < threshold-2; i++ {
		avgStart := int(math.Floor(float64(i+1)*every)) + 1
		avgEnd := int(math.Floor(float64(i+2)*every)) + 1
		if avgEnd > n {
			avgEnd = n
		}
		var avgX, avgY float64
		cnt := float64(avgEnd - avgStart)
		for j := avgStart; j < avgEnd; j++ {
			avgX += float64(data[j].T)
			avgY += data[j].V
		}
		if cnt > 0 {
			avgX /= cnt
			avgY /= cnt
		}
		rangeStart := int(math.Floor(float64(i)*every)) + 1
		rangeEnd := int(math.Floor(float64(i+1)*every)) + 1
		ax, ay := float64(data[a].T), data[a].V
		maxArea := -1.0
		next := rangeStart
		for j := rangeStart; j < rangeEnd; j++ {
			area := math.Abs((ax-avgX)*(data[j].V-ay)-(ax-float64(data[j].T))*(avgY-ay)) * 0.5
			if area > maxArea {
				maxArea, next = area, j
			}
		}
		out = append(out, data[next])
		a = next
	}
	return append(out, data[n-1])
}
