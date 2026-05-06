package scheduler

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/consi/grosz/internal/tariff"
)

func TestComputeScheduleCheapestHours(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	deadline := now.Add(24 * time.Hour)

	// 24 hours with varied prices
	rates := make([]tariff.Rate, 24)
	prices := []float64{
		0.80, 0.75, 0.70, 0.65, // 0-3: moderate
		0.30, 0.25, 0.20, 0.22, // 4-7: cheap
		0.90, 0.95, 1.00, 0.85, // 8-11: expensive
		0.60, 0.55, 0.50, 0.45, // 12-15: moderate
		0.70, 0.75, 0.80, 0.85, // 16-19: moderate-high
		0.40, 0.35, 0.30, 0.50, // 20-23: cheap-moderate
	}
	for i := 0; i < 24; i++ {
		rates[i] = tariff.Rate{
			Start: now.Add(time.Duration(i) * time.Hour),
			End:   now.Add(time.Duration(i+1) * time.Hour),
			Price: prices[i],
		}
	}

	cfg := Config{
		TargetEnergy:   32, // kWh — with 3% headroom: ceil(32*1.03/11) = 3 hours
		Deadline:       deadline,
		MaxPower:       11000, // W = 11 kW
		MinPower:       1380,
		ChargeHeadroom: 3,
	}

	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched)
	require.Len(t, sched.Slots, 1)

	slot := sched.Slots[0]
	assert.Len(t, slot.Periods, 3)

	// Should pick the 3 cheapest: 0.20, 0.22, 0.25
	for _, p := range slot.Periods {
		assert.LessOrEqual(t, p.Price, 0.30, "should pick cheapest hours")
		assert.Equal(t, float64(11000), p.Power)
	}

	// Verify sorted chronologically
	for i := 1; i < len(slot.Periods); i++ {
		assert.True(t, slot.Periods[i].Start.After(slot.Periods[i-1].Start))
	}
}

func TestComputeScheduleSlowerCharger(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	deadline := now.Add(12 * time.Hour)

	rates := make([]tariff.Rate, 12)
	for i := 0; i < 12; i++ {
		rates[i] = tariff.Rate{
			Start: now.Add(time.Duration(i) * time.Hour),
			End:   now.Add(time.Duration(i+1) * time.Hour),
			Price: float64(i+1) * 0.10, // 0.10 to 1.20
		}
	}

	cfg := Config{
		TargetEnergy:   13.5, // kWh — with 3% headroom: ceil(13.5*1.03/3.6) = 4 hours
		Deadline:       deadline,
		MaxPower:       3600, // W = 3.6 kW (single phase)
		ChargeHeadroom: 3,
	}

	sched := ComputeSchedule(rates, cfg, 3600)
	require.NotNil(t, sched)
	require.Len(t, sched.Slots, 1)

	slot := sched.Slots[0]
	assert.Len(t, slot.Periods, 4)

	// Should pick hours 0,1,2,3 (cheapest)
	assert.Equal(t, 0.10, slot.Periods[0].Price)
	assert.Equal(t, 0.40, slot.Periods[3].Price)
}

func TestComputeScheduleDeadlinePressure(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	deadline := now.Add(6 * time.Hour)

	rates := make([]tariff.Rate, 6)
	for i := 0; i < 6; i++ {
		rates[i] = tariff.Rate{
			Start: now.Add(time.Duration(i) * time.Hour),
			End:   now.Add(time.Duration(i+1) * time.Hour),
			Price: 1.00, // all expensive
		}
	}

	cfg := Config{
		TargetEnergy:   60, // kWh — needs all 6 hours at 11kW = 66 kWh
		Deadline:       deadline,
		MaxPower:       11000,
		ChargeHeadroom: 3,
	}

	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched)
	require.Len(t, sched.Slots, 1)

	// Must use all 6 hours (ceil(60/11) = 6)
	assert.Len(t, sched.Slots[0].Periods, 6)
}

func TestComputeScheduleNoRates(t *testing.T) {
	cfg := Config{TargetEnergy: 30, Deadline: time.Now().Add(24 * time.Hour), MaxPower: 11000, ChargeHeadroom: 3}
	sched := ComputeSchedule(nil, cfg, 11000)
	assert.Nil(t, sched)
}

func TestComputeScheduleZeroTarget(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	rates := []tariff.Rate{{Start: now, End: now.Add(time.Hour), Price: 0.5}}
	cfg := Config{TargetEnergy: 0, Deadline: now.Add(24 * time.Hour), MaxPower: 11000, ChargeHeadroom: 3}
	sched := ComputeSchedule(rates, cfg, 11000)
	assert.Nil(t, sched)
}

func TestComputeScheduleNegativePrices(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	rates := []tariff.Rate{
		{Start: now, End: now.Add(time.Hour), Price: 0.50},
		{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour), Price: -0.10},
		{Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour), Price: 0.30},
		{Start: now.Add(3 * time.Hour), End: now.Add(4 * time.Hour), Price: -0.05},
	}

	// Need 32 kWh — with 3% headroom: ceil(32*1.03/11) = 3 hours
	// Should pick the 3 cheapest: -0.10, -0.05, 0.30 (chronological order)
	cfg := Config{
		TargetEnergy:   32,
		Deadline:       now.Add(5 * time.Hour),
		MaxPower:       11000,
		ChargeHeadroom: 3,
	}
	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched)
	require.Len(t, sched.Slots, 1)
	assert.Len(t, sched.Slots[0].Periods, 3)
	assert.Equal(t, -0.10, sched.Slots[0].Periods[0].Price)
	assert.Equal(t, 0.30, sched.Slots[0].Periods[1].Price)
	assert.Equal(t, -0.05, sched.Slots[0].Periods[2].Price)
}

func TestComputeScheduleNegativePricesLimitedCapacity(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	rates := []tariff.Rate{
		{Start: now, End: now.Add(time.Hour), Price: -0.05},
		{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour), Price: -0.20},
		{Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour), Price: -0.15},
		{Start: now.Add(3 * time.Hour), End: now.Add(4 * time.Hour), Price: -0.10},
	}

	// Only 21 kWh capacity — with 3% headroom: ceil(21*1.03/11) = 2 hours
	// Should pick the 2 most negative: -0.20 and -0.15
	cfg := Config{
		TargetEnergy:   21,
		Deadline:       now.Add(5 * time.Hour),
		MaxPower:       11000,
		ChargeHeadroom: 3,
	}
	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched)
	require.Len(t, sched.Slots, 1)
	assert.Len(t, sched.Slots[0].Periods, 2)
	assert.Equal(t, -0.20, sched.Slots[0].Periods[0].Price)
	assert.Equal(t, -0.15, sched.Slots[0].Periods[1].Price)
	assert.True(t, sched.Cost < 0, "cost should be negative (earnings)")
}

func TestComputeScheduleNegativePricesZeroTarget(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	rates := []tariff.Rate{
		{Start: now, End: now.Add(time.Hour), Price: 0.50},
		{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour), Price: -0.10},
	}

	// No energy needed (battery at target SoC) — no schedule even with negative prices.
	// SoC limits are respected; the battery has no headroom to absorb energy.
	cfg := Config{
		TargetEnergy:   0,
		Deadline:       now.Add(3 * time.Hour),
		MaxPower:       11000,
		ChargeHeadroom: 3,
	}
	sched := ComputeSchedule(rates, cfg, 11000)
	assert.Nil(t, sched)
}

func TestComputeScheduleCost(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	rates := []tariff.Rate{
		{Start: now, End: now.Add(time.Hour), Price: 0.50},
		{Start: now.Add(time.Hour), End: now.Add(2 * time.Hour), Price: 0.30},
		{Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour), Price: 0.80},
	}

	cfg := Config{
		TargetEnergy:   21, // with 3% headroom: ceil(21*1.03/11) = 2 hours
		Deadline:       now.Add(4 * time.Hour),
		MaxPower:       11000,
		ChargeHeadroom: 3,
	}

	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched)
	require.Len(t, sched.Slots, 1)
	assert.Len(t, sched.Slots[0].Periods, 2)

	// Should pick 0.30 and 0.50 hours
	expectedCost := math.Round((0.30*11+0.50*11)*100) / 100
	assert.Equal(t, expectedCost, sched.Cost)
}

func TestComputeScheduleMultiDay(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	// Deadline at 4 hours from now — first window has only 4 hours
	firstDeadline := now.Add(4 * time.Hour)

	// 48 hours of rates spanning two day windows
	rates := make([]tariff.Rate, 48)
	for i := 0; i < 48; i++ {
		rates[i] = tariff.Rate{
			Start: now.Add(time.Duration(i) * time.Hour),
			End:   now.Add(time.Duration(i+1) * time.Hour),
			Price: 0.50 + float64(i%12)*0.05, // repeating price pattern
		}
	}

	cfg := Config{
		TargetEnergy:   50, // with 3% headroom: ceil(50*1.03/11) = 5 hours, exceeds first 4h window
		Deadline:       firstDeadline,
		MaxPower:       11000,
		ChargeHeadroom: 3,
	}

	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched)
	assert.GreaterOrEqual(t, len(sched.Slots), 2, "should spill into second day slot")

	// Energy should be at least the target (headroom may add slightly more)
	assert.GreaterOrEqual(t, sched.Energy, 50.0, "total energy should cover target")

	// Each slot should have cheapest hour(s) for its window
	for _, slot := range sched.Slots {
		assert.NotEmpty(t, slot.Periods)
		assert.NotEmpty(t, slot.Date)
	}
}

// TestComputeScheduleSpillsToTomorrowWhenTodayBlockedByMaxPrice covers the
// late/missed plug-in case: today's window has rates but all are above
// MaxPrice, so today's slot is empty and the full target must spill into
// tomorrow's deadline window.
func TestComputeScheduleSpillsToTomorrowWhenTodayBlockedByMaxPrice(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	firstDeadline := now.Add(4 * time.Hour)

	rates := make([]tariff.Rate, 28)
	for i := 0; i < 4; i++ {
		rates[i] = tariff.Rate{
			Start: now.Add(time.Duration(i) * time.Hour),
			End:   now.Add(time.Duration(i+1) * time.Hour),
			Price: 1.50, // all above MaxPrice
		}
	}
	for i := 4; i < 28; i++ {
		rates[i] = tariff.Rate{
			Start: now.Add(time.Duration(i) * time.Hour),
			End:   now.Add(time.Duration(i+1) * time.Hour),
			Price: 0.40,
		}
	}

	cfg := Config{
		TargetEnergy:   22, // ceil(22*1.03/11) = 3 hours
		Deadline:       firstDeadline,
		MaxPower:       11000,
		MaxPrice:       1.00,
		ChargeHeadroom: 3,
	}

	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched, "should produce a schedule by spilling into tomorrow")
	require.Len(t, sched.Slots, 1, "today's window had no eligible rates; only tomorrow's slot should appear")
	assert.GreaterOrEqual(t, sched.Slots[0].Energy, 22.0, "tomorrow's slot should cover the full target")
	assert.True(t, sched.Slots[0].Deadline.After(firstDeadline), "spilled slot should target the next deadline")
}

// TestComputeScheduleCapsAtTwoCycles ensures we don't iterate beyond
// tomorrow even when remainingEnergy is still positive after two windows.
func TestComputeScheduleCapsAtTwoCycles(t *testing.T) {
	now := time.Now().Truncate(time.Hour)
	firstDeadline := now.Add(2 * time.Hour)

	// Rates for 4 days, all expensive enough that one slot can't fit a huge target.
	rates := make([]tariff.Rate, 96)
	for i := 0; i < 96; i++ {
		rates[i] = tariff.Rate{
			Start: now.Add(time.Duration(i) * time.Hour),
			End:   now.Add(time.Duration(i+1) * time.Hour),
			Price: 0.50,
		}
	}

	cfg := Config{
		TargetEnergy:   500, // unrealistic huge target; would otherwise iterate 4 days
		Deadline:       firstDeadline,
		MaxPower:       11000,
		ChargeHeadroom: 3,
	}

	sched := ComputeSchedule(rates, cfg, 11000)
	require.NotNil(t, sched)
	assert.LessOrEqual(t, len(sched.Slots), 2, "should never plan beyond tomorrow")
}

// --- mergeActiveSlotPreservingActive ---

func mockNow(t *testing.T, at time.Time) {
	t.Helper()
	orig := timeNow
	timeNow = func() time.Time { return at }
	t.Cleanup(func() { timeNow = orig })
}

// Mid-session, the next adjacent recomputed hour should extend the active
// period's End so the running OCPP transaction continues seamlessly.
func TestMergeActiveSlotExtendsAdjacent(t *testing.T) {
	day := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	now := day.Add(2*time.Hour + 30*time.Minute) // 02:30, mid-active-period
	mockNow(t, now)

	active := SchedulePeriod{
		Start: day.Add(2 * time.Hour),
		End:   day.Add(3 * time.Hour),
		Power: 11000, Price: 0.30,
	}
	inProgress := ScheduleSlot{
		Date:     "2026-05-06",
		Deadline: day.Add(7 * time.Hour),
		Periods:  []SchedulePeriod{active},
	}

	// Recomputed today picks 03:00–04:00 (adjacent to active.End=03:00) at
	// the same power — should merge into the active period.
	recomputed := &Schedule{
		Slots: []ScheduleSlot{
			{
				Date:     "2026-05-06",
				Deadline: day.Add(7 * time.Hour),
				Periods: []SchedulePeriod{
					{Start: day.Add(3 * time.Hour), End: day.Add(4 * time.Hour), Power: 11000, Price: 0.40},
				},
			},
		},
	}

	merged := mergeActiveSlotPreservingActive(recomputed, cloneSlot(inProgress))
	require.NotNil(t, merged)
	require.Len(t, merged.Slots, 1)
	require.Len(t, merged.Slots[0].Periods, 1, "adjacent period should be merged into active, not appended")

	p := merged.Slots[0].Periods[0]
	assert.Equal(t, active.Start, p.Start, "active Start unchanged")
	assert.Equal(t, day.Add(4*time.Hour), p.End, "active End extended through adjacent hour")
	assert.Equal(t, float64(11000), p.Power)
	// Volume-weighted price: (0.30*11 + 0.40*11) / 22 = 0.35
	assert.InDelta(t, 0.35, p.Price, 0.001)
}

// Mid-session with no adjacent recomputed hour — non-adjacent later periods
// are appended as separate windows; active period left intact.
func TestMergeActiveSlotAppendsNonAdjacent(t *testing.T) {
	day := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	now := day.Add(2*time.Hour + 30*time.Minute)
	mockNow(t, now)

	active := SchedulePeriod{
		Start: day.Add(2 * time.Hour),
		End:   day.Add(3 * time.Hour),
		Power: 11000, Price: 0.30,
	}
	inProgress := ScheduleSlot{
		Date:     "2026-05-06",
		Deadline: day.Add(7 * time.Hour),
		Periods:  []SchedulePeriod{active},
	}

	recomputed := &Schedule{
		Slots: []ScheduleSlot{
			{
				Date:     "2026-05-06",
				Deadline: day.Add(7 * time.Hour),
				Periods: []SchedulePeriod{
					{Start: day.Add(5 * time.Hour), End: day.Add(6 * time.Hour), Power: 11000, Price: 0.20},
				},
			},
		},
	}

	merged := mergeActiveSlotPreservingActive(recomputed, cloneSlot(inProgress))
	require.NotNil(t, merged)
	require.Len(t, merged.Slots, 1)
	require.Len(t, merged.Slots[0].Periods, 2)

	assert.Equal(t, active.Start, merged.Slots[0].Periods[0].Start)
	assert.Equal(t, active.End, merged.Slots[0].Periods[0].End, "active End preserved (no adjacent)")
	assert.Equal(t, day.Add(5*time.Hour), merged.Slots[0].Periods[1].Start)
}

// Mid-session, recompute today is empty (nothing eligible). Active period
// preserved verbatim; tomorrow's slot stays.
func TestMergeActiveSlotEmptyRecomputeToday(t *testing.T) {
	day := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	now := day.Add(2*time.Hour + 30*time.Minute)
	mockNow(t, now)

	active := SchedulePeriod{
		Start: day.Add(2 * time.Hour),
		End:   day.Add(3 * time.Hour),
		Power: 11000, Price: 0.30,
	}
	inProgress := ScheduleSlot{
		Date:     "2026-05-06",
		Deadline: day.Add(7 * time.Hour),
		Periods:  []SchedulePeriod{active},
	}

	tomorrow := ScheduleSlot{
		Date:     "2026-05-07",
		Deadline: day.Add(31 * time.Hour),
		Periods: []SchedulePeriod{
			{Start: day.Add(26 * time.Hour), End: day.Add(28 * time.Hour), Power: 11000, Price: 0.25},
		},
	}
	recomputed := &Schedule{Slots: []ScheduleSlot{tomorrow}}

	merged := mergeActiveSlotPreservingActive(recomputed, cloneSlot(inProgress))
	require.NotNil(t, merged)
	require.Len(t, merged.Slots, 2)

	today := merged.Slots[0]
	require.Len(t, today.Periods, 1)
	assert.Equal(t, active, today.Periods[0], "active preserved verbatim when no recomputed today-slot")

	tom := merged.Slots[1]
	assert.Equal(t, "2026-05-07", tom.Date)
	require.Len(t, tom.Periods, 1)
}

// When the recomputed period overlaps the active period (would-be conflict),
// it must be filtered out — the active period owns its time range exclusively.
func TestMergeActiveSlotIgnoresOverlappingRecomputedPeriods(t *testing.T) {
	day := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	now := day.Add(2*time.Hour + 30*time.Minute)
	mockNow(t, now)

	active := SchedulePeriod{
		Start: day.Add(2 * time.Hour),
		End:   day.Add(4 * time.Hour),
		Power: 11000, Price: 0.30,
	}
	inProgress := ScheduleSlot{
		Date:     "2026-05-06",
		Deadline: day.Add(7 * time.Hour),
		Periods:  []SchedulePeriod{active},
	}

	// Recomputed picks 03:00-04:00 (overlaps active) and 04:00-05:00 (adjacent).
	recomputed := &Schedule{
		Slots: []ScheduleSlot{
			{
				Date:     "2026-05-06",
				Deadline: day.Add(7 * time.Hour),
				Periods: []SchedulePeriod{
					{Start: day.Add(3 * time.Hour), End: day.Add(4 * time.Hour), Power: 11000, Price: 0.10},
					{Start: day.Add(4 * time.Hour), End: day.Add(5 * time.Hour), Power: 11000, Price: 0.40},
				},
			},
		},
	}

	merged := mergeActiveSlotPreservingActive(recomputed, cloneSlot(inProgress))
	require.NotNil(t, merged)
	require.Len(t, merged.Slots, 1)
	require.Len(t, merged.Slots[0].Periods, 1, "overlapping period dropped, adjacent merged")
	p := merged.Slots[0].Periods[0]
	assert.Equal(t, active.Start, p.Start)
	assert.Equal(t, day.Add(5*time.Hour), p.End, "extended through adjacent 04:00-05:00")
}

func TestNextDeadline(t *testing.T) {
	d := nextDeadline("07:00")
	assert.Equal(t, 7, d.Hour())
	assert.Equal(t, 0, d.Minute())
	assert.True(t, d.After(time.Now()) || d.Equal(time.Now()))
}
