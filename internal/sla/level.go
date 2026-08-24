package sla

// LevelMinutes is the SLA window added per escalation level.
const LevelMinutes = 30

// MaxLevel is the highest escalation level.
const MaxLevel = 3

// NextLevel returns the level after the current one.
func NextLevel(current int) int {
	if current >= MaxLevel {
		return MaxLevel
	}
	return current + 1
}
