package application

import (
	"math/rand"
	"testing"
	"time"

	"multicliente-backend/internal/features/rifa/domain"
)

// Helper to build weighted pool following the exact algorithm in SortearRifa
func buildWeightedPool(
	participaciones []domain.ParticipacionRifa,
	esPremioMayor bool,
	refMap map[uint]int64,
) []domain.ParticipacionRifa {
	if !esPremioMayor {
		return participaciones
	}

	var weightedPool []domain.ParticipacionRifa
	for _, p := range participaciones {
		// Base chance
		weightedPool = append(weightedPool, p)
		// Extra chances per referral in the period
		extra := refMap[p.UsuarioID]
		for i := int64(0); i < extra; i++ {
			weightedPool = append(weightedPool, p)
		}
	}
	return weightedPool
}

func TestPremioMayorWeightedPool(t *testing.T) {
	// Participant A: User 1 with 9 referrals in period
	// Participant B: User 2 with 0 referrals in period
	participaciones := []domain.ParticipacionRifa{
		{ID: 101, UsuarioID: 1, NumeroParticipacion: "000001"},
		{ID: 102, UsuarioID: 2, NumeroParticipacion: "000002"},
	}

	refMap := map[uint]int64{
		1: 9, // 9 referrals
		2: 0, // 0 referrals
	}

	t.Run("Standard Draw (esPremioMayor = false) should have equal chances", func(t *testing.T) {
		pool := buildWeightedPool(participaciones, false, refMap)
		if len(pool) != 2 {
			t.Fatalf("Expected pool size 2, got %d", len(pool))
		}
	})

	t.Run("Premio Mayor (esPremioMayor = true) should weight by referrals", func(t *testing.T) {
		pool := buildWeightedPool(participaciones, true, refMap)
		// User 1 gets 1 base + 9 referrals = 10 entries
		// User 2 gets 1 base + 0 referrals = 1 entry
		// Total pool = 11 entries
		if len(pool) != 11 {
			t.Fatalf("Expected pool size 11, got %d", len(pool))
		}

		user1Entries := 0
		user2Entries := 0
		for _, p := range pool {
			if p.UsuarioID == 1 {
				user1Entries++
			} else if p.UsuarioID == 2 {
				user2Entries++
			}
		}

		if user1Entries != 10 {
			t.Errorf("Expected User 1 to have 10 entries in pool, got %d", user1Entries)
		}
		if user2Entries != 1 {
			t.Errorf("Expected User 2 to have 1 entry in pool, got %d", user2Entries)
		}

		// Statistical verification: over 10,000 draws, User 1 should win roughly 90.9% of the time (10/11)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		winsUser1 := 0
		const iterations = 10000
		for i := 0; i < iterations; i++ {
			picked := pool[r.Intn(len(pool))]
			if picked.UsuarioID == 1 {
				winsUser1++
			}
		}

		winRateUser1 := float64(winsUser1) / float64(iterations)
		// Expected win rate is ~90.9% (between 88% and 94%)
		if winRateUser1 < 0.88 || winRateUser1 > 0.94 {
			t.Errorf("User 1 win rate %f was outside expected statistical bounds (~0.909)", winRateUser1)
		}
	})
}

func TestDetermineFechaCorteInicio(t *testing.T) {
	now := time.Date(2026, 10, 1, 21, 30, 0, 0, time.UTC)

	t.Run("No previous Premio Mayor - uses start of FechaInicio day", func(t *testing.T) {
		rifa := domain.Rifa{
			FechaInicio: time.Date(2026, 10, 1, 15, 0, 0, 0, time.UTC),
		}
		corte := determineFechaCorteInicio(rifa, nil, now)
		expected := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		if !corte.Equal(expected) {
			t.Errorf("Expected %v, got %v", expected, corte)
		}
	})

	t.Run("Previous draw in the past - uses start of previous draw date", func(t *testing.T) {
		prevTime := time.Date(2026, 9, 20, 14, 30, 0, 0, time.UTC)
		prev := domain.Rifa{
			ID:       10,
			UpdateAt: &prevTime,
		}
		rifa := domain.Rifa{
			FechaInicio: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
		}
		corte := determineFechaCorteInicio(rifa, &prev, now)
		expected := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
		if !corte.Equal(expected) {
			t.Errorf("Expected %v, got %v", expected, corte)
		}
	})

	t.Run("Scheduled future date in previous draw or current raffle - guards against inversion", func(t *testing.T) {
		futureTime := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
		prev := domain.Rifa{
			ID:          10,
			FechaSorteo: futureTime,
			UpdateAt:    &futureTime,
		}
		rifa := domain.Rifa{
			FechaInicio: futureTime,
		}
		corte := determineFechaCorteInicio(rifa, &prev, now)
		// Must not be after now
		if corte.After(now) {
			t.Errorf("corte %v should not be after now %v", corte, now)
		}
		expected := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		if !corte.Equal(expected) {
			t.Errorf("Expected %v, got %v", expected, corte)
		}
	})
}

