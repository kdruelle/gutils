package containers



import "slices"

// ShiftRight décale le slice vers la droite de `shift` positions
func ShiftRight[T any](s []T, shift int) {
	n := len(s)
	if n == 0 || shift%n == 0 {
		return
	}
	shift = shift % n // Évite les rotations inutiles
	tmp := slices.Clone(s[n-shift:]) // Copie des derniers `shift` éléments
	copy(s[shift:], s[:n-shift]) // Décalage interne
	copy(s[:shift], tmp) // Copie des éléments sauvegardés au début
}

// ShiftLeft décale le slice vers la gauche de `shift` positions
func ShiftLeft[T any](s []T, shift int) {
	n := len(s)
	if n == 0 || shift%n == 0 {
		return
	}
	shift = shift % n // Évite les rotations inutiles
	tmp := slices.Clone(s[:shift]) // Copie des `shift` premiers éléments
	copy(s, s[shift:]) // Décalage interne
	copy(s[n-shift:], tmp) // Copie des éléments sauvegardés à la fin
}