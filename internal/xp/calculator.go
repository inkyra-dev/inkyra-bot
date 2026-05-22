package xp

import "strings"

// XPToNextLevel retourne l'XP nécessaire pour passer du niveau N au niveau N+1.
// Formule identique à MEE6 : 5n² + 50n + 100
func XPToNextLevel(level int) int64 {
	return int64(5*level*level + 50*level + 100)
}

// TotalXPForLevel retourne l'XP cumulée nécessaire pour atteindre le niveau donné.
func TotalXPForLevel(level int) int64 {
	var total int64
	for i := 0; i < level; i++ {
		total += XPToNextLevel(i)
	}
	return total
}

// LevelFromXP retourne le niveau correspondant à une quantité d'XP totale.
func LevelFromXP(totalXP int64) int {
	level := 0
	for level < MaxLevel && TotalXPForLevel(level+1) <= totalXP {
		level++
	}
	return level
}

// ProgressInLevel retourne (xp actuel dans le niveau, xp requis pour passer au suivant).
func ProgressInLevel(totalXP int64, level int) (current, needed int64) {
	base := TotalXPForLevel(level)
	return totalXP - base, XPToNextLevel(level)
}

// ProgressBar génère une barre de progression unicode de `width` caractères.
func ProgressBar(current, max int64, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := int(float64(current) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)
}

// MaxLevel est le niveau maximum avant de pouvoir prestigier.
const MaxLevel = 100
