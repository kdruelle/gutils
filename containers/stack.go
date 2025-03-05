package containers

func NewStack[T any]() *Stack[T] {
	s := &Stack[T]{
		data: make([]T, 0, 1),
	}
	return s
}

type Stack[T any] struct {
	data []T
}

func (s *Stack[T]) Push(value T) {
	s.data = append(s.data, value)
}

// Pop retire et retourne l'élément en haut de la pile
func (s *Stack[T]) Pop() (T, error) {
	if len(s.data) == 0 {
		var zero T
		return zero, ErrEmpty
	}
	last := len(s.data) - 1
	value := s.data[last]
	s.data = s.data[:last] // Réduction efficace du slice (O(1))
	return value, nil
}

// Peek retourne l'élément en haut sans le retirer
func (s *Stack[T]) Peek() (T, error) {
	if len(s.data) == 0 {
		var zero T
		return zero, ErrEmpty
	}
	return s.data[len(s.data)-1], nil
}

// Len retourne le nombre d'éléments
func (s *Stack[T]) Len() int {
	return len(s.data)
}

func (s *Stack[T]) Cap() int {
	return cap(s.data)
}

// IsEmpty vérifie si la pile est vide
func (s *Stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// Clear vide la pile
func (s *Stack[T]) Clear() {
	s.data = s.data[:0] // Réinitialisation efficace (évite réallocation mémoire)
}
