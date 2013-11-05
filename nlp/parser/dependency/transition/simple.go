package transition

import (
	"chukuparser/algorithm/graph"
	. "chukuparser/algorithm/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"fmt"
	// "log"
	// "reflect"
	"strings"
	"sync"
)

type SimpleConfiguration struct {
	sync.Mutex
	InternalStack                    Stack
	InternalQueue                    Queue
	InternalArcs                     ArcSet
	Nodes                            []*ArcCachedDepNode
	InternalPrevious                 *SimpleConfiguration
	Last                             Transition
	Pointers                         int
	EWord, EPOS, EWPOS, ERel, ETrans *util.EnumSet
	// test zpar parity
	NumHeadStack int
}

func (c *SimpleConfiguration) IncrementPointers() {
	// c.Lock()
	// c.Pointers++
	// c.Unlock()
}

func (c *SimpleConfiguration) DecrementPointers() {
	// c.Lock()
	// c.Pointers--
	// c.Unlock()
	// if c.Pointers <= 0 {
	// 	c.Clear()
	// }
}

func (c *SimpleConfiguration) Conf() Configuration {
	return Configuration(c)
}

func (c *SimpleConfiguration) Graph() nlp.LabeledDependencyGraph {
	return nlp.LabeledDependencyGraph(c)
}

// Verify that SimpleConfiguration is a Configuration
var _ DependencyConfiguration = &SimpleConfiguration{}
var _ nlp.DependencyGraph = &SimpleConfiguration{}

func (c *SimpleConfiguration) ID() int {
	return 0
}

func (c *SimpleConfiguration) Init(abstractSentence interface{}) {
	sent := abstractSentence.(nlp.EnumTaggedSentence)
	// var exists bool
	sentLength := len(sent.TaggedTokens())
	// Nodes is always the same slice to the same token array
	c.Nodes = make([]*ArcCachedDepNode, 0, sentLength)
	for i, enumToken := range sent.EnumTaggedTokens() {
		node := &TaggedDepNode{
			i,
			enumToken.EToken,
			enumToken.EPOS,
			enumToken.ETPOS,
			enumToken.Token,
			enumToken.POS,
		}
		c.Nodes = append(c.Nodes, NewArcCachedDepNode(nlp.DepNode(node)))
	}

	c.InternalStack = NewStackArray(sentLength)
	c.InternalQueue = NewQueueSlice(sentLength)
	c.InternalArcs = NewArcSetSimple(sentLength)

	// push index of ROOT node to Stack
	// c.Stack().Push(0) // TODO: note switch to zpar's PopRoot

	// push indexes of statement nodes to Queue, in reverse order (first word at the top of the queue)
	for i := 0; i < sentLength; i++ {
		c.Queue().Enqueue(i)
	}
	// explicit resetting of zero-valued properties
	// in case of reuse
	c.Last = 0
	c.InternalPrevious = nil
	c.NumHeadStack = 0
	// c.Pointers = 0
}

func (c *SimpleConfiguration) Clear() {
	// c.Lock()
	// defer c.Unlock()
	// if c.Pointers > 0 {
	// 	return
	// }
	// c.InternalStack = nil
	// c.InternalQueue = nil
	// c.InternalArcs = nil
	// if c.InternalPrevious != nil {
	// 	c.InternalPrevious.DecrementPointers()
	// 	c.InternalPrevious.Clear()
	// 	c.InternalPrevious = nil
	// }

}

func (c *SimpleConfiguration) Terminal() bool {
	return c.Queue().Size() == 0 && c.Stack().Size() == 0
}

func (c *SimpleConfiguration) Stack() Stack {
	return c.InternalStack
}

func (c *SimpleConfiguration) Queue() Queue {
	return c.InternalQueue
}

func (c *SimpleConfiguration) Arcs() ArcSet {
	return c.InternalArcs
}

func (c *SimpleConfiguration) Copy() Configuration {
	newConf := new(SimpleConfiguration)

	if c.Stack() != nil {
		newConf.InternalStack = c.Stack().Copy()
	}
	if c.Queue() != nil {
		newConf.InternalQueue = c.Queue().Copy()
	}
	if c.Arcs() != nil {
		newConf.InternalArcs = c.Arcs().Copy()
	}
	newConf.Nodes = make([]*ArcCachedDepNode, len(c.Nodes), cap(c.Nodes))
	copy(newConf.Nodes[0:len(c.Nodes)], c.Nodes[0:len(c.Nodes)])

	newConf.Last = c.Last
	newConf.NumHeadStack = c.NumHeadStack
	// store a pointer to the previous configuration
	newConf.InternalPrevious = c
	// explicit setting of pointer counter
	// newConf.Pointers = 0

	// c.Lock()
	// c.Pointers += 1
	// c.Unlock()
	newConf.EWord, newConf.EPOS, newConf.EWPOS, newConf.ERel, newConf.ETrans = c.EWord, c.EPOS, c.EWPOS, c.ERel, c.ETrans

	return newConf
}

func (c *SimpleConfiguration) AddArc(arc *BasicDepArc) {
	c.Arcs().Add(arc)
	if c.Nodes[arc.Modifier].ELabel >= 0 {
		panic("Tried to change the label of a labeled node")
	}
	c.Nodes[arc.Modifier] = c.Nodes[arc.Modifier].Copy()
	c.Nodes[arc.Modifier].Head = arc.Head
	c.Nodes[arc.Modifier].ELabel = arc.Relation
	c.Nodes[arc.Modifier].ArcId = c.Arcs().Size() - 1
	c.Nodes[arc.Head] = c.Nodes[arc.Head].Copy()
	c.Nodes[arc.Head].AddModifier(arc.Modifier, arc.Relation)
}

func (c *SimpleConfiguration) Equal(otherEq util.Equaler) bool {
	// log.Println("Testing equality for")
	// log.Println("\t", c)
	// log.Println("\t", otherEq)
	if (otherEq == nil && c != nil) || (c == nil && otherEq != nil) {
		return false
	}
	switch other := otherEq.(type) {
	case *SimpleConfiguration:
		if (other == nil && c != nil) || (c == nil && other != nil) {
			return false
		}
		if other.Last != c.Last {
			return false
		}
		if c.InternalPrevious == nil && other.InternalPrevious == nil {
			return true
		}
		if c.InternalPrevious != nil && other.InternalPrevious != nil {
			return c.Previous().Equal(other.Previous())
		} else {
			return false
		}
		// return other.Last == c.Last &&
		// 	((c.InternalPrevious == nil && other.InternalPrevious == nil) ||
		// 		(c.InternalPrevious != nil && other.InternalPrevious != nil && c.Previous().Equal(other.Previous())))

		// return c.NumberOfArcs() == other.NumberOfArcs() &&
		// 	c.NumberOfNodes() == other.NumberOfNodes() &&
		// 	c.Stack().Equal(other.Stack()) &&
		// 	c.Queue().Equal(other.Queue()) &&
		// 	c.Arcs().Equal(other.Arcs()) &&
		// 	reflect.DeepEqual(c.Nodes, other.Nodes)
	case *BasicDepGraph:
		return other.Equal(c)
	}
	return false
}

func (c *SimpleConfiguration) Previous() DependencyConfiguration {
	return c.InternalPrevious
}

func (c *SimpleConfiguration) SetLastTransition(t Transition) {
	c.Last = t
}

func (c *SimpleConfiguration) GetLastTransition() Transition {
	return Transition(c.Last)
}

func (c *SimpleConfiguration) GetSequence() ConfigurationSequence {
	if c.Arcs() == nil {
		return make(ConfigurationSequence, 0)
	}
	retval := make(ConfigurationSequence, 0, c.Arcs().Size())
	currentConf := c
	for currentConf != nil {
		retval = append(retval, currentConf)
		currentConf = currentConf.InternalPrevious
	}
	return retval
}

// GRAPH FUNCTIONS
func (c *SimpleConfiguration) GetVertices() []int {
	return util.RangeInt(len(c.Nodes))
}

func (c *SimpleConfiguration) GetEdges() []int {
	return util.RangeInt(c.Arcs().Size())
}

func (c *SimpleConfiguration) GetVertex(vertexID int) graph.Vertex {
	return graph.Vertex(c.Nodes[vertexID])
}

func (c *SimpleConfiguration) GetEdge(edgeID int) graph.Edge {
	arcPtr := c.Arcs().Index(edgeID)
	return graph.Edge(arcPtr)
}

func (c *SimpleConfiguration) GetDirectedEdge(edgeID int) graph.DirectedEdge {
	arcPtr := c.Arcs().Index(edgeID)
	return graph.DirectedEdge(arcPtr)
}

func (c *SimpleConfiguration) NumberOfVertices() int {
	return c.NumberOfNodes()
}

func (c *SimpleConfiguration) NumberOfEdges() int {
	return c.NumberOfArcs()
}

func (c *SimpleConfiguration) NumberOfNodes() int {
	return len(c.Nodes)
}

func (c *SimpleConfiguration) NumberOfArcs() int {
	return c.Arcs().Size()
}

func (c *SimpleConfiguration) GetNode(nodeID int) nlp.DepNode {
	return c.Nodes[nodeID].Node
}

func (c *SimpleConfiguration) GetRawNode(nodeID int) *TaggedDepNode {
	return c.Nodes[nodeID].Node.(*TaggedDepNode)
}

func (c *SimpleConfiguration) GetArc(nodeID int) nlp.DepArc {
	if c.Nodes[nodeID].ArcId > -1 {
		arcPtr := c.Arcs().Index(c.Nodes[nodeID].ArcId)
		return nlp.DepArc(arcPtr)
	} else {
		return nil
	}
}

func (c *SimpleConfiguration) GetLabeledArc(nodeID int) nlp.LabeledDepArc {
	if c.Nodes[nodeID].ArcId > -1 {
		arcPtr := c.Arcs().Index(c.Nodes[nodeID].ArcId)
		return nlp.LabeledDepArc(arcPtr)
	} else {
		return nil
	}
}

// OUTPUT FUNCTIONS

func (c *SimpleConfiguration) String() string {
	var (
		transitionVal string = ""
		transInt      int    = int(c.Last)
	)
	if transInt >= 0 {
		transitionVal = c.ETrans.ValueOf(transInt).(string)
	}
	return fmt.Sprintf("%s\t=>([%s],\t[%s],\t%s)",
		transitionVal, c.StringStack(), c.StringQueue(),
		c.StringArcs())
}

func (c *SimpleConfiguration) StringStack() string {
	stackSize := c.Stack().Size()
	switch {
	case stackSize > 0 && stackSize <= 3:
		var stackStrings []string = make([]string, 0, 3)
		for i := c.Stack().Size() - 1; i >= 0; i-- {
			atI, _ := c.Stack().Index(i)
			stackStrings = append(stackStrings, c.GetRawNode(atI).RawToken)
		}
		return strings.Join(stackStrings, ",")
	case stackSize > 3:
		headID, _ := c.Stack().Index(0)
		tailID, _ := c.Stack().Index(c.Stack().Size() - 1)
		head := c.GetRawNode(headID)
		tail := c.GetRawNode(tailID)
		return strings.Join([]string{tail.RawToken, "...", head.RawToken}, ",")
	default:
		return ""
	}
}

func (c *SimpleConfiguration) StringQueue() string {
	queueSize := c.Queue().Size()
	switch {
	case queueSize > 0 && queueSize <= 3:
		var queueStrings []string = make([]string, 0, 3)
		for i := 0; i < c.Queue().Size(); i++ {
			atI, _ := c.Queue().Index(i)
			queueStrings = append(queueStrings, c.GetRawNode(atI).RawToken)
		}
		return strings.Join(queueStrings, ",")
	case queueSize > 3:
		headID, _ := c.Queue().Index(0)
		tailID, _ := c.Queue().Index(c.Queue().Size() - 1)
		head := c.GetRawNode(headID)
		tail := c.GetRawNode(tailID)
		return strings.Join([]string{head.RawToken, "...", tail.RawToken}, ",")
	default:
		return ""
	}
}

func (c *SimpleConfiguration) StringArcs() string {
	var transInt int = int(c.Last)
	if transInt < 0 {
		return ""
	}
	last := c.ETrans.ValueOf(transInt).(string)
	if len(last) < 2 {
		return fmt.Sprintf("A%d", c.Arcs().Size())
	}
	switch last[:2] {
	case "LA", "RA":
		lastArc := c.Arcs().Last()
		head := c.GetRawNode(lastArc.GetHead())
		mod := c.GetRawNode(lastArc.GetModifier())
		arcStr := fmt.Sprintf("(%s,%s,%s)", head.RawToken, lastArc.GetRelation().String(), mod.RawToken)
		return fmt.Sprintf("A%d=A%d+{%s}", c.Arcs().Size(), c.Arcs().Size()-1, arcStr)
	default:
		return fmt.Sprintf("A%d", c.Arcs().Size())
	}
}

func (c *SimpleConfiguration) StringGraph() string {
	return fmt.Sprintf("%v %v", c.Nodes, c.InternalArcs)
}

func (c *SimpleConfiguration) Sentence() nlp.Sentence {
	return nlp.Sentence(c.TaggedSentence())
}

func (c *SimpleConfiguration) TaggedSentence() nlp.TaggedSentence {
	var sent nlp.BasicETaggedSentence = make([]nlp.EnumTaggedToken, c.NumberOfNodes()-1)
	for i, _ := range c.Nodes {
		taggedNode := c.GetRawNode(i)
		sent[i] = nlp.EnumTaggedToken{
			nlp.TaggedToken{taggedNode.RawToken, taggedNode.RawPOS},
			taggedNode.Token, taggedNode.POS, taggedNode.TokenPOS}
	}
	return sent
}

func NewSimpleConfiguration() Configuration {
	return Configuration(new(SimpleConfiguration))
}