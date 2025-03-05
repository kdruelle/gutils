package containers

// ensure List implement Container interface
var _ Container = (*Deque[int])(nil)
var _ SequenceContainer[int] = (*Deque[int])(nil)

// Deque (double-ended queue) is an indexed sequence container that allows fast insertion and deletion at both its beginning and its end.
// In addition, insertion and deletion at either end of a deque never invalidates pointers or references to the rest of the elements.
// The elements of a Deque are not stored contiguously: implementation use a sequence of individually allocated fixed-size arrays.
// The storage of Deque is automatically expanded and contracted as needed. Expansion of a Deque does not involve copying of the existing elements to a new memory location.
// The complexity (efficiency) of common operations on Deque is as follows:
//   - Random access: constant O(1).
//   - Insertion or removal of elements at the end or beginning: constant O(1).
//   - Insertion or removal of elements at specific index - linear O(n).
//
// Deque implements [Container] and [SequenceContainer].
type Deque[T any] struct {
	blocks       [][]T // Tableau de pointeurs vers blocs
	first        int   // Index du premier élément logique
	last         int   // Index du dernier élément logique
	growFactor   float64
	shrinkFactor float64
	nbBlocks     int
	chunkSize    int
}

// NewDeque Create a new empty Deque container
func NewDeque[T any]() *Deque[T] {

	d := &Deque[T]{
		nbBlocks:     64,
		chunkSize:    256,
		growFactor:   0.25,
		shrinkFactor: 0.2,
	}

	alloc := make([]T, d.nbBlocks*d.chunkSize)
	d.blocks = make([][]T, d.nbBlocks)
	for i := range d.blocks {
		d.blocks[i] = alloc[i*d.chunkSize : (i+1)*d.chunkSize]
	}
	d.first = (d.nbBlocks * d.chunkSize) / 2
	d.last = (d.nbBlocks * d.chunkSize) / 2

	return d
}

/****************************************
** Container implementation
****************************************/

// Len returns the number of elements in the Deque
func (d *Deque[T]) Len() int {
	return int(d.last - d.first)
}

// Cap returns the number of elements that can be held in currently allocated storage
func (d *Deque[T]) Cap() int {
	return len(d.blocks) * d.chunkSize
}

// IsEmpty checks if the container is empty
func (d *Deque[T]) IsEmpty() bool {
	return d.Len() == 0
}

// Clear Empties the container
func (d *Deque[T]) Clear() {
	d.first = (d.nbBlocks * d.chunkSize) / 2
	d.last = (d.nbBlocks * d.chunkSize) / 2
	d.checkSize()
}

/****************************************
** SequenceContainer implementation
****************************************/

// Get returns the value of the element at specified location index, with bounds checking.
// If index is not within the range of the container, it returns a zero value and a [ErrEmpty] if the Deque is empty or [ErrIndexOutOfRange] if it is not.
//
// Complexity is constant O(1)
func (d *Deque[T]) Get(index int) (T, error) {
	if d.Len() == 0 {
		var zero T
		return zero, ErrEmpty
	}
	if index < 0 || index >= d.Len() {
		var zero T
		return zero, ErrOutOfRange
	}
	blockIdx := (int(d.first) + index) / d.chunkSize
	elemIdx := (int(d.first) + index) % d.chunkSize
	return d.blocks[blockIdx][elemIdx], nil
}

func (d *Deque[T]) GetPointer(index int) (*T, error) {
	if d.Len() == 0 {
		return nil, ErrEmpty
	}
	if index < 0 || index >= d.Len() {
		return nil, ErrOutOfRange
	}
	blockIdx := (int(d.first) + index) / d.chunkSize
	elemIdx := (int(d.first) + index) % d.chunkSize
	return &d.blocks[blockIdx][elemIdx], nil
}

// Set modify the value of the element at specified location index, with bounds checking.
// If index is not within the range of the container, it returns a [ErrIndexOutOfRange] error.
//
// Complexity is constant O(1)
func (d *Deque[T]) Set(index int, value T) error {
	if index < 0 || index >= d.Len() {
		return ErrOutOfRange
	}
	blockIdx := (int(d.first) + index) / d.chunkSize
	elemIdx := (int(d.first) + index) % d.chunkSize
	d.blocks[blockIdx][elemIdx] = value
	return nil
}

// Insert inserts element at the specified index location in the container.
// All iterators (including the End() iterator) are invalidated.
// If index is not within the range of the container, it returns a [ErrIndexOutOfRange] error unless index refers to the end, in such case this function behave like [Deque.PushBack].
//
// If index is first of last element, complexity is constant O(1) and O(n/2) for random access
func (d *Deque[T]) Insert(index int, value T) error {
	if index < 0 || index > d.Len() {
		return ErrOutOfRange
	}

	d.resize()

	absIndex := int(d.first) + index
	blockIdx := absIndex / d.chunkSize
	elemIdx := absIndex % d.chunkSize

	firstBlockIdx := int(d.first) / d.chunkSize
	lastBlockIdx := int(d.last) / d.chunkSize

	if index < d.Len()/2 {

		for i := firstBlockIdx; i < blockIdx; i++ {
			if i >= firstBlockIdx {
				d.blocks[i-1][d.chunkSize-1] = d.blocks[i][0]
			}
			copy(d.blocks[i], d.blocks[i][1:])
		}
		d.blocks[blockIdx-1][d.chunkSize-1] = d.blocks[blockIdx][0]
		copy(d.blocks[blockIdx][:elemIdx], d.blocks[blockIdx][1:elemIdx])
		d.blocks[blockIdx][elemIdx-1] = value
		d.first--
	} else {
		for i := lastBlockIdx; i > blockIdx; i-- {
			if i < lastBlockIdx {
				d.blocks[i+1][0] = d.blocks[i][d.chunkSize-1]
			}
			copy(d.blocks[i][1:], d.blocks[i][:d.chunkSize-1])
		}
		d.blocks[blockIdx+1][0] = d.blocks[blockIdx][d.chunkSize-1]
		copy(d.blocks[blockIdx][elemIdx+1:], d.blocks[blockIdx][elemIdx:])
		d.blocks[blockIdx][elemIdx] = value
		d.last++
	}

	return nil
}

// Remove removes the element at the specified index location in the container.
// All iterators (including the End() iterator) are invalidated.
// If index is not within the range of the container, it returns a [ErrIndexOutOfRange] error.
//
// If index is first of last element, complexity is constant O(1) and O(n/2) for random access
func (d *Deque[T]) Remove(index int) error {
	if index < 0 || index >= d.Len() {
		return ErrOutOfRange
	}

	absIndex := int(d.first) + index
	blockIdx := absIndex / d.chunkSize
	elemIdx := absIndex % d.chunkSize

	firstBlockIdx := int(d.first) / d.chunkSize
	lastBlockIdx := int(d.last) / d.chunkSize

	if index < d.Len()/2 {
		copy(d.blocks[blockIdx][elemIdx:], d.blocks[blockIdx][elemIdx+1:])
		for i := blockIdx - 1; i >= firstBlockIdx; i-- {
			copy(d.blocks[i][1:], d.blocks[i][:d.chunkSize-1])
			d.blocks[i][0] = d.blocks[i-1][d.chunkSize-1]
		}
		d.first++
	} else {
		copy(d.blocks[blockIdx][elemIdx:], d.blocks[blockIdx][elemIdx+1:])
		for i := blockIdx + 1; i <= lastBlockIdx; i++ {
			copy(d.blocks[i][:d.chunkSize-1], d.blocks[i][1:])
			d.blocks[i][d.chunkSize-1] = d.blocks[i+1][0]
		}
		d.last--
	}

	d.checkSize()
	return nil
}

// PushBack adds an element to the end of the container
//
// Complexity is constant O(1)
func (d *Deque[T]) PushBack(val T) {
	d.resize()
	blockIdx := d.last / d.chunkSize
	elemIdx := d.last % d.chunkSize
	d.blocks[blockIdx][elemIdx] = val
	d.last++
}

// PushFront adds an element to the begining of the container
//
// Complexity is constant O(1)
func (d *Deque[T]) PushFront(val T) {
	d.resize()
	blockIdx := d.first / d.chunkSize
	elemIdx := d.first % d.chunkSize
	if elemIdx == 0 {
		blockIdx--
		elemIdx = d.chunkSize
	}
	d.blocks[blockIdx][elemIdx-1] = val
	d.first--
}

// PopBack gets and removes the last element
// It returns an [ErrEmpty] if the container is empty
// The function is equivalant to [Deque.Get(0)]
//
// Complexity is constant O(1)
func (d *Deque[T]) PopBack() (T, error) {
	if d.Len() == 0 {
		var zero T
		return zero, ErrEmpty
	}
	blockIdx := d.last / d.chunkSize
	elemIdx := d.last % d.chunkSize
	if elemIdx == 0 {
		blockIdx--
		elemIdx = d.chunkSize
	}
	val := d.blocks[blockIdx][elemIdx-1]
	d.last--
	d.checkSize()
	return val, nil
}

// PopFront gets and removes the first element
// // It returns an [ErrEmpty] if the container is empty
// The function is equivalent to Deque.Get(Deque.Len() - 1)
//
// Complexity is constant O(1)
func (d *Deque[T]) PopFront() (T, error) {
	if d.Len() == 0 {
		var zero T
		return zero, ErrEmpty
	}
	blockIdx := d.first / d.chunkSize
	elemIdx := d.first % d.chunkSize
	val := d.blocks[blockIdx][elemIdx]
	d.first++
	d.checkSize()
	return val, nil
}

// ToSlice return a slice holding a copy of the data hold by the container
func (d *Deque[T]) ToSlice() []T {
	var result []T = make([]T, 0)

	if d.Len() == 0 {
		return result
	}

	firstBlockIdx := d.first / d.chunkSize
	firstElemIdx := d.first % d.chunkSize
	lastBlockIdx := (d.last) / d.chunkSize
	lastElemIdx := (d.last) % d.chunkSize

	if firstBlockIdx == lastBlockIdx {
		return d.blocks[firstBlockIdx][firstElemIdx:lastElemIdx]
	}

	result = append(result, d.blocks[firstBlockIdx][firstElemIdx:]...)
	for i := firstBlockIdx + 1; i < lastBlockIdx; i++ {
		result = append(result, d.blocks[i]...)
	}
	result = append(result, d.blocks[lastBlockIdx][:lastElemIdx]...)

	return result
}

func (d *Deque[T]) resize() {
	firstBlockIdx := d.first / d.chunkSize

	if firstBlockIdx == 0 {
		if d.Len() < d.Cap()/2 {
			d.balance()
			return
		}
		d.expandStart()
	}

	lastBlockIdx := (d.last) / d.chunkSize
	if lastBlockIdx == len(d.blocks)-1 {
		if d.Len() < d.Cap()/2 {
			d.balance()
			return
		}
		d.expandEnd()
	}
}

func (d *Deque[T]) expandStart() {

	blockCount := len(d.blocks)
	newBlockCount := blockCount + max(d.nbBlocks/2, int(float64(blockCount)*d.growFactor))
	newBlocks := make([][]T, newBlockCount)

	copy(newBlocks[newBlockCount-blockCount:], d.blocks)

	alloc := make([]T, blockCount*d.chunkSize)
	for i := range newBlockCount - blockCount {
		newBlocks[i] = alloc[i*d.chunkSize : (i+1)*d.chunkSize]
	}

	offset := (newBlockCount - blockCount) * d.chunkSize
	d.first += offset
	d.last += offset
	d.blocks = newBlocks
}

func (d *Deque[T]) expandEnd() {
	blockCount := len(d.blocks)
	newBlockCount := blockCount + max(d.nbBlocks/2, int(float64(blockCount)*d.growFactor))
	newBlocks := make([][]T, newBlockCount)

	copy(newBlocks, d.blocks)

	alloc := make([]T, (newBlockCount-blockCount)*d.chunkSize)
	for i := blockCount; i < newBlockCount; i++ {
		newBlocks[i] = alloc[:d.chunkSize]
		alloc = alloc[d.chunkSize:]
	}

	d.blocks = newBlocks
}

func (d *Deque[T]) balance() {

	firstBlockIdx := d.first / d.chunkSize
	lastBlockIdx := (d.last) / d.chunkSize

	nbEndEmptyBlocks := len(d.blocks) - 1 - int(lastBlockIdx)
	nbStartEmptyBlocks := firstBlockIdx

	if nbStartEmptyBlocks > nbEndEmptyBlocks {
		ShiftLeft(d.blocks, int(nbStartEmptyBlocks)/2)
		d.first -= (nbStartEmptyBlocks / 2) * d.chunkSize
		d.last -= (nbStartEmptyBlocks / 2) * d.chunkSize
	} else if nbStartEmptyBlocks < nbEndEmptyBlocks {
		ShiftRight(d.blocks, nbEndEmptyBlocks/2)
		d.first += (nbEndEmptyBlocks / 2) * d.chunkSize
		d.last += (nbEndEmptyBlocks / 2) * d.chunkSize
	}

}

func (d *Deque[T]) checkSize() {
	blockCount := len(d.blocks)
	firstBlockIdx := d.first / d.chunkSize
	lastBlockIdx := (d.last) / d.chunkSize
	usedBlocks := lastBlockIdx - firstBlockIdx

	usedBlocksPct := float64(usedBlocks) / float64(blockCount)
	if usedBlocksPct >= d.shrinkFactor {
		return
	}

	freeBlocks := blockCount - usedBlocks
	blocksToRemove := freeBlocks / 2

	if blocksToRemove == 0 || blockCount <= d.nbBlocks {
		return
	}

	removeFront := blocksToRemove / 2
	removeBack := blocksToRemove - removeFront

	if firstBlockIdx < blockCount/4 {
		removeBack = min(blocksToRemove, freeBlocks)
		removeFront = blocksToRemove - removeBack
	} else if lastBlockIdx > (3*blockCount)/4 {
		removeFront = min(blocksToRemove, freeBlocks)
		removeBack = blocksToRemove - removeFront
	}

	newBlockCount := blockCount - removeBack - removeFront
	newBlocks := make([][]T, newBlockCount)

	copy(newBlocks, d.blocks[removeFront:blockCount-removeBack])

	d.first -= removeFront * d.chunkSize
	d.last -= removeFront * d.chunkSize
	d.blocks = newBlocks
}
