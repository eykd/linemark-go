package domain

import (
	"errors"
	"testing"
)

// --- helpers for building occupied slices ---

// rangeInts returns integers from start to end inclusive.
func rangeInts(start, end int) []int {
	nums := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		nums = append(nums, i)
	}
	return nums
}

// stepInts returns integers from start to end inclusive, stepping by step.
func stepInts(start, end, step int) []int {
	nums := make([]int, 0)
	for i := start; i <= end; i += step {
		nums = append(nums, i)
	}
	return nums
}

// concatInts concatenates multiple int slices.
func concatInts(slices ...[]int) []int {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	result := make([]int, 0, total)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func TestNextSiblingNumber(t *testing.T) {
	tests := []struct {
		name     string
		occupied []int
		want     int
		wantErr  error
	}{
		// Initial spacing at 100s
		{
			name:     "empty parent gets 100",
			occupied: nil,
			want:     100,
		},
		{
			name:     "empty slice gets 100",
			occupied: []int{},
			want:     100,
		},
		{
			name:     "second child gets 200",
			occupied: []int{100},
			want:     200,
		},
		{
			name:     "third child gets 300",
			occupied: []int{100, 200},
			want:     300,
		},
		{
			name:     "ninth child gets 900",
			occupied: []int{100, 200, 300, 400, 500, 600, 700, 800},
			want:     900,
		},
		// Falls to 10s tier when no 100-multiple available after last
		{
			name:     "after 900 falls to 10s tier",
			occupied: stepInts(100, 900, 100),
			want:     910,
		},
		{
			name:     "continues 10s tier after 910",
			occupied: concatInts(stepInts(100, 900, 100), []int{910}),
			want:     920,
		},
		// Falls to 1s tier when no 10-multiple available after last
		{
			name:     "uses 1s tier when 10s after last exhausted",
			occupied: concatInts(stepInts(100, 900, 100), stepInts(910, 990, 10)),
			want:     991,
		},
		// Handles unsorted input
		{
			name:     "unsorted input is sorted internally",
			occupied: []int{300, 100, 200},
			want:     400,
		},
		// Non-standard spacing still finds next 100-multiple
		{
			name:     "non-standard spacing finds next 100 multiple",
			occupied: []int{50, 150},
			want:     200,
		},
		// Slot exhaustion: all 999 positions occupied
		{
			name:     "all 999 occupied returns max siblings error",
			occupied: rangeInts(1, 999),
			wantErr:  ErrMaxSiblingsReached,
		},
		// No gap after last but gaps exist elsewhere (suggest compact)
		{
			name:     "no gap after last suggests compact",
			occupied: concatInts([]int{1}, rangeInts(991, 999)),
			wantErr:  ErrNoSlotAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextSiblingNumber(tt.occupied)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("NextSiblingNumber() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSiblingNumberBefore(t *testing.T) {
	tests := []struct {
		name     string
		occupied []int
		target   int
		want     int
		wantErr  error
	}{
		// 10s tier between predecessor and target
		{
			name:     "before 200 finds 10s gap",
			occupied: []int{100, 200, 300},
			target:   200,
			want:     110,
		},
		// Before first sibling (predecessor is 0)
		{
			name:     "before first sibling uses 10s tier",
			occupied: []int{100, 200},
			target:   100,
			want:     10,
		},
		// 1s tier when 10s exhausted
		{
			name:     "before 110 uses 1s tier",
			occupied: []int{100, 110, 200},
			target:   110,
			want:     101,
		},
		// 100s tier when wide gap available
		{
			name:     "before 500 with gap at 100s",
			occupied: []int{100, 500},
			target:   500,
			want:     200,
		},
		// No gap between adjacent values
		{
			name:     "no gap between adjacent values",
			occupied: []int{100, 101, 102},
			target:   101,
			wantErr:  ErrNoSlotAvailable,
		},
		// All 999 occupied
		{
			name:     "all 999 occupied returns max error",
			occupied: rangeInts(1, 999),
			target:   500,
			wantErr:  ErrMaxSiblingsReached,
		},
		// Before first when first is 1 (no room below)
		{
			name:     "before 1 has no room",
			occupied: []int{1, 2, 3},
			target:   1,
			wantErr:  ErrNoSlotAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SiblingNumberBefore(tt.occupied, tt.target)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("SiblingNumberBefore() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSiblingNumberAfter(t *testing.T) {
	tests := []struct {
		name     string
		occupied []int
		target   int
		want     int
		wantErr  error
	}{
		// 10s tier between target and successor
		{
			name:     "after 100 finds 10s gap before 200",
			occupied: []int{100, 200, 300},
			target:   100,
			want:     110,
		},
		// After last sibling uses 100s tier into open space
		{
			name:     "after last sibling finds next 100 multiple",
			occupied: []int{100, 200},
			target:   200,
			want:     300,
		},
		// 1s tier when 10s gaps occupied
		{
			name:     "after 100 with 110 occupied uses next 10s gap",
			occupied: []int{100, 110, 200},
			target:   100,
			want:     120,
		},
		// 1s tier when all 10s in range occupied
		{
			name:     "uses 1s tier when all 10s occupied",
			occupied: concatInts([]int{100}, stepInts(110, 190, 10), []int{200}),
			target:   100,
			want:     101,
		},
		// 100s tier when wide gap available after target
		{
			name:     "after 100 with next at 500 finds 200",
			occupied: []int{100, 500},
			target:   100,
			want:     200,
		},
		// No gap between adjacent values
		{
			name:     "no gap between adjacent values",
			occupied: []int{100, 101, 102},
			target:   100,
			wantErr:  ErrNoSlotAvailable,
		},
		// After last sibling at 999 (no room above)
		{
			name:     "after 999 has no room",
			occupied: []int{100, 999},
			target:   999,
			wantErr:  ErrNoSlotAvailable,
		},
		// All 999 occupied
		{
			name:     "all 999 occupied returns max error",
			occupied: rangeInts(1, 999),
			target:   500,
			wantErr:  ErrMaxSiblingsReached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SiblingNumberAfter(tt.occupied, tt.target)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("SiblingNumberAfter() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCompactNumbers(t *testing.T) {
	tests := []struct {
		name    string
		count   int
		want    []int
		wantErr error
	}{
		{
			name:  "zero siblings returns empty",
			count: 0,
			want:  []int{},
		},
		{
			name:  "one sibling",
			count: 1,
			want:  []int{100},
		},
		{
			name:  "three siblings",
			count: 3,
			want:  []int{100, 200, 300},
		},
		{
			name:  "nine siblings fills capacity at 100-spacing",
			count: 9,
			want:  []int{100, 200, 300, 400, 500, 600, 700, 800, 900},
		},
		{
			name:    "ten siblings exceeds 100-spacing capacity",
			count:   10,
			wantErr: ErrMaxSiblingsReached,
		},
		{
			name:    "999 siblings exceeds 100-spacing capacity",
			count:   999,
			wantErr: ErrMaxSiblingsReached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompactNumbers(tt.count)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("CompactNumbers(%d) length = %d, want %d", tt.count, len(got), len(tt.want))
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("CompactNumbers(%d)[%d] = %d, want %d", tt.count, i, v, tt.want[i])
				}
			}
		})
	}
}

func TestFindGap(t *testing.T) {
	tests := []struct {
		name   string
		low    int
		high   int
		want   int
		wantOK bool
	}{
		// 100s tier
		{
			name:   "finds 100-multiple in wide gap",
			low:    0,
			high:   200,
			want:   100,
			wantOK: true,
		},
		{
			name:   "finds 200 between 100 and 300",
			low:    100,
			high:   300,
			want:   200,
			wantOK: true,
		},
		// 10s tier
		{
			name:   "finds 10-multiple between 100s",
			low:    100,
			high:   200,
			want:   110,
			wantOK: true,
		},
		{
			name:   "finds 10 between 0 and 100",
			low:    0,
			high:   100,
			want:   10,
			wantOK: true,
		},
		// 1s tier
		{
			name:   "finds unit between 10s",
			low:    100,
			high:   110,
			want:   101,
			wantOK: true,
		},
		{
			name:   "finds unit in tight gap",
			low:    100,
			high:   102,
			want:   101,
			wantOK: true,
		},
		// No gap
		{
			name:   "no gap between adjacent values",
			low:    100,
			high:   101,
			wantOK: false,
		},
		{
			name:   "no gap when low equals high",
			low:    100,
			high:   100,
			wantOK: false,
		},
		{
			name:   "no gap when low greater than high",
			low:    200,
			high:   100,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FindGap(tt.low, tt.high)

			if ok != tt.wantOK {
				t.Errorf("FindGap(%d, %d) ok = %v, want %v", tt.low, tt.high, ok, tt.wantOK)
				return
			}
			if tt.wantOK && got != tt.want {
				t.Errorf("FindGap(%d, %d) = %d, want %d", tt.low, tt.high, got, tt.want)
			}
		})
	}
}
