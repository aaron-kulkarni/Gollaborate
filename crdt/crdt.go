package crdt

type Document struct {
	Lines []Line `json:"lines"`
}

type Line struct {
	Characters []Character `json:"characters"`
}

type Identifier struct {
	Digit int `json:"digit"`
	Node  int `json:"node"`
}

func FromIdentifierList(identifiers []Identifier) []int {
	var digits []int
	for _, ident := range identifiers {
		digits = append(digits, ident.Digit)
	}
	return digits
}

func Increment(n1 []int, delta []int) []int {
	// Find the first non-zero digit in delta
	firstNonZeroDigit := -1
	for i, x := range delta {
		if x != 0 {
			firstNonZeroDigit = i
			break
		}
	}

	if firstNonZeroDigit == -1 {
		panic("Delta must contain at least one non-zero digit")
	}

	// Create the increment array
	inc := append(delta[:firstNonZeroDigit], 0, 1)

	// Add increment to n1
	v1 := Add(n1, inc)

	// Check if the last digit is zero, and adjust if necessary
	if v1[len(v1)-1] == 0 {
		v1 = Add(v1, inc)
	}

	return v1
}

type Character struct {
	Pos   []Identifier `json:"pos"`
	Clock int          `json:"clock"`
	Value rune         `json:"value"`
}

const BASE = 256

func SubtractGreaterThan(n1 []int, n2 []int) []int {
	carry := 0
	diff := make([]int, max(len(n1), len(n2)))
	for i := len(diff) - 1; i >= 0; i-- {
		d1 := 0
		if i < len(n1) {
			d1 = n1[i] - carry
		}
		d2 := 0
		if i < len(n2) {
			d2 = n2[i]
		}
		if d1 < d2 {
			carry = 1
			diff[i] = d1 + BASE - d2
		} else {
			carry = 0
			diff[i] = d1 - d2
		}
	}
	return diff
}

func Add(n1 []int, n2 []int) []int {
	carry := 0
	sum := make([]int, max(len(n1), len(n2)))
	for i := len(sum) - 1; i >= 0; i-- {
		s := carry
		if i < len(n1) {
			s += n1[i]
		}
		if i < len(n2) {
			s += n2[i]
		}
		carry = s / BASE
		sum[i] = s % BASE
	}
	if carry != 0 {
		panic("sum is greater than one, cannot be represented by this type")
	}
	return sum
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func generatePositionBetween(position1 []Identifier, position2 []Identifier, node int) []Identifier {
	// Get either the head of the position, or fallback to default value
	var head1 Identifier
	if len(position1) > 0 {
		head1 = position1[0]
	} else {
		head1 = Identifier{Digit: 0, Node: node}
	}

	var head2 Identifier
	if len(position2) > 0 {
		head2 = position2[0]
	} else {
		head2 = Identifier{Digit: BASE, Node: node}
	}

	if head1.Digit != head2.Digit {
		// Case 1: Head digits are different
		n1 := FromIdentifierList(position1)
		n2 := FromIdentifierList(position2)
		delta := SubtractGreaterThan(n2, n1)
		// Increment n1 by some amount less than delta
		next := Increment(n1, delta)
		return ToIdentifierList(next, position1, position2, node)
	} else {
		if head1.Node < head2.Node {
			// Case 2: Head digits are the same, nodes are different
			return append([]Identifier{head1}, generatePositionBetween(position1[1:], []Identifier{}, node)...)
		} else if head1.Node == head2.Node {
			// Case 3: Head digits and nodes are the same
			return append([]Identifier{head1}, generatePositionBetween(position1[1:], position2[1:], node)...)
		} else {
			panic("invalid node ordering")
		}
	}
}

func ToIdentifierList(n []int, before []Identifier, after []Identifier, creationNode int) []Identifier {
	identifiers := make([]Identifier, len(n))
	for index, digit := range n {
		if index == len(n)-1 {
			identifiers[index] = Identifier{Digit: digit, Node: creationNode}
		} else if index < len(before) && digit == before[index].Digit {
			identifiers[index] = Identifier{Digit: digit, Node: before[index].Node}
		} else if index < len(after) && digit == after[index].Digit {
			identifiers[index] = Identifier{Digit: digit, Node: after[index].Node}
		} else {
			identifiers[index] = Identifier{Digit: digit, Node: creationNode}
		}
	}
	return identifiers
}

// func (d *Document) InsertCharacter(character Character, lineNum int, linePos int) error {
// 	var identity Identifier
// }
