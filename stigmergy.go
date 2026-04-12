package fluxstigmergy

import (
	"strings"
	"time"
)

type TraceType int

const (
	InfoTrace      TraceType = iota
	WarningTrace
	ClaimTrace
	WaypointTrace
	BoundaryTrace
)

type Trace struct {
	AuthorId  uint16
	Key       string
	Value     string
	Timestamp time.Time
	Strength  uint32 // 0-1000, decays
	Reads     uint32
	Type      TraceType
}

type SharedEnvironment struct {
	traces []*Trace
}

func NewEnv() *SharedEnvironment {
	return &SharedEnvironment{}
}

func (e *SharedEnvironment) Deposit(author uint16, key, value string, t TraceType, strength uint32) int {
	e.traces = append(e.traces, &Trace{
		AuthorId:  author,
		Key:       key,
		Value:     value,
		Timestamp: time.Now(),
		Strength:  strength,
		Type:      t,
	})
	return len(e.traces) - 1
}

func (e *SharedEnvironment) Read(key string) *Trace {
	for _, t := range e.traces {
		if t.Key == key {
			t.Reads++
			return t
		}
	}
	return nil
}

func (e *SharedEnvironment) ReadAll(prefix string, max int) []*Trace {
	var result []*Trace
	for _, t := range e.traces {
		if strings.HasPrefix(t.Key, prefix) {
			t.Reads++
			result = append(result, t)
			if max > 0 && len(result) >= max {
				break
			}
		}
	}
	return result
}

func (e *SharedEnvironment) Modify(author uint16, key, newValue string, strengthAdd uint32) bool {
	for _, t := range e.traces {
		if t.Key == key && t.AuthorId == author {
			t.Value = newValue
			t.Strength = min(t.Strength+strengthAdd, 1000)
			return true
		}
	}
	return false
}

func (e *SharedEnvironment) Erase(author uint16, key string) bool {
	for i, t := range e.traces {
		if t.Key == key && t.AuthorId == author {
			e.traces = append(e.traces[:i], e.traces[i+1:]...)
			return true
		}
	}
	return false
}

func (e *SharedEnvironment) Decay(halfLife time.Duration, readBoost uint32, minStrength uint32) int {
	if halfLife == 0 {
		halfLife = time.Hour
	}
	now := time.Now()
	var keep []*Trace
	removed := 0
	for _, t := range e.traces {
		elapsed := now.Sub(t.Timestamp)
		factor := float64(elapsed) / float64(halfLife)
		boost := float64(t.Reads) * float64(readBoost)
		newStrength := float64(t.Strength)/pow2(factor) + boost
		if newStrength > 1000 {
			newStrength = 1000
		}
		t.Strength = uint32(newStrength)
		if t.Strength < minStrength {
			removed++
		} else {
			keep = append(keep, t)
		}
	}
	e.traces = keep
	return removed
}

func pow2(x float64) float64 {
	if x <= 0 {
		return 1
	}
	return pow2(x/2) * 2
}

func (e *SharedEnvironment) ByAuthor(author uint16) []*Trace {
	var result []*Trace
	for _, t := range e.traces {
		if t.AuthorId == author {
			result = append(result, t)
		}
	}
	return result
}

func (e *SharedEnvironment) ByType(t TraceType) []*Trace {
	var result []*Trace
	for _, tr := range e.traces {
		if tr.Type == t {
			result = append(result, tr)
		}
	}
	return result
}

func (e *SharedEnvironment) Strongest(n int) []*Trace {
	cp := make([]*Trace, len(e.traces))
	copy(cp, e.traces)
	for i := 0; i < len(cp)-1; i++ {
		for j := i + 1; j < len(cp); j++ {
			if cp[j].Strength > cp[i].Strength {
				cp[i], cp[j] = cp[j], cp[i]
			}
		}
	}
	if n > 0 && n < len(cp) {
		return cp[:n]
	}
	return cp
}

func (e *SharedEnvironment) Oldest(n int) []*Trace {
	cp := make([]*Trace, len(e.traces))
	copy(cp, e.traces)
	for i := 0; i < len(cp)-1; i++ {
		for j := i + 1; j < len(cp); j++ {
			if cp[j].Timestamp.Before(cp[i].Timestamp) {
				cp[i], cp[j] = cp[j], cp[i]
			}
		}
	}
	if n > 0 && n < len(cp) {
		return cp[:n]
	}
	return cp
}

type DecayConfig struct {
	HalfLife    time.Duration
	ReadBoost   uint32
	MinStrength uint32
}

type Stats struct {
	Total          int
	Active         int
	ByType         [5]int
	TotalReads     uint32
	TotalStrength  uint32
	UniqueAuthors  int
}

func (e *SharedEnvironment) Stats(minStrength uint32) Stats {
	var s Stats
	authors := make(map[uint16]bool)
	s.Total = len(e.traces)
	for _, t := range e.traces {
		if t.Strength >= minStrength {
			s.Active++
			s.TotalReads += t.Reads
			s.TotalStrength += t.Strength
			s.ByType[t.Type]++
			authors[t.AuthorId] = true
		}
	}
	s.UniqueAuthors = len(authors)
	return s
}
