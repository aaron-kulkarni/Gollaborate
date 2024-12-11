package crdt

type Document struct {
	lines []Line
}

type Line struct {
	characters []Character
}

type Identifier struct {
	digit int
	node  int
}

type Character struct {
	pos   []Identifier
	clock float64
	value string
}

type Node struct {
	nodeID int
}

func NewNode(nodeID int) *Node {
	return &Node{
		nodeID: nodeID,
	}
}

// func (d *Document) InsertCharacter(character Character, lineNum int, linePos int) error {
// 	var identity Identifier
// }
