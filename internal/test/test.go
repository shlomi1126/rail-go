package heap

// MinHeap represents a min heap data structure
type MinHeap struct {
	array []int
}

// NewMinHeap creates a new empty min heap
func NewMinHeap() *MinHeap {
	return &MinHeap{
		array: []int{},
	}
}

// NewMinHeapFromArray creates a min heap from an existing array
func NewMinHeapFromArray(arr []int) *MinHeap {
	h := &MinHeap{
		array: make([]int, len(arr)),
	}
	copy(h.array, arr)
	h.BuildHeap()
	return h
}

// Size returns the number of elements in the heap
func (h *MinHeap) Size() int {
	return len(h.array)
}

// IsEmpty returns true if the heap is empty
func (h *MinHeap) IsEmpty() bool {
	return len(h.array) == 0
}

// GetMin returns the minimum element without removing it
func (h *MinHeap) GetMin() (int, bool) {
	if h.IsEmpty() {
		return 0, false
	}
	return h.array[0], true
}

// Insert adds a new element to the heap
func (h *MinHeap) Insert(key int) {
	h.array = append(h.array, key)
	h.siftUp(len(h.array) - 1)
}

// ExtractMin removes and returns the minimum element
func (h *MinHeap) ExtractMin() (int, bool) {
	if h.IsEmpty() {
		return 0, false
	}

	min := h.array[0]
	lastIndex := len(h.array) - 1
	
	// Replace root with last element
	h.array[0] = h.array[lastIndex]
	h.array = h.array[:lastIndex]
	
	// Restore heap property if there are elements left
	if !h.IsEmpty() {
		h.siftDown(0)
	}
	
	return min, true
}

// BuildHeap constructs a heap from an unordered array
func (h *MinHeap) BuildHeap() {
	// Start from the last non-leaf node
	for i := (len(h.array) / 2) - 1; i >= 0; i-- {
		h.siftDown(i)
	}
}

// GetArray returns the underlying array (mainly for testing)
func (h *MinHeap) GetArray() []int {
	result := make([]int, len(h.array))
	copy(result, h.array)
	return result
}

// Private helper methods

// siftUp moves an element up to its correct position
func (h *MinHeap) siftUp(index int) {
	parent := (index - 1) / 2
	
	// If we're at root or the parent is already smaller, we're done
	if index == 0 || h.array[parent] <= h.array[index] {
		return
	}
	
	// Swap with parent and continue sifting up
	h.array[parent], h.array[index] = h.array[index], h.array[parent]
	h.siftUp(parent)
}

// siftDown moves an element down to its correct position
func (h *MinHeap) siftDown(index int) {
	smallest := index
	leftChild := 2*index + 1
	rightChild := 2*index + 2
	
	// Find the smallest among node, left child and right child
	if leftChild < len(h.array) && h.array[leftChild] < h.array[smallest] {
		smallest = leftChild
	}
	
	if rightChild < len(h.array) && h.array[rightChild] < h.array[smallest] {
		smallest = rightChild
	}
	
	// If the smallest is not the current node, swap and continue sifting down
	if smallest != index {
		h.array[index], h.array[smallest] = h.array[smallest], h.array[index]
		h.siftDown(smallest)
	}
}