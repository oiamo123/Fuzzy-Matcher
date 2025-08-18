package fuzzymatchertypes

// RecurseParameters contains all parameters needed for recursive matching
type RecurseParameters struct {
    Word              []rune
    Index             int
    Node              *FuzzyMatcherNode
    Path              []rune
    MaxDepth          int
    Depth             int
    DepthIncrement    int
    NumEdits          int
    MaxEdits          int
    NumEditsIncrement int
    EditableFields    []bool
    Visited           map[VisitKey]struct{}
    CalculationMethod CalculationMethod
}

func (rp *RecurseParameters) Clone() RecurseParameters {
    newPath := append([]rune{}, rp.Path...)
    newVisited := make(map[VisitKey]struct{}, len(rp.Visited))
    for k, v := range rp.Visited {
        newVisited[k] = v
    }

    return RecurseParameters{
        Word:              rp.Word,
        Index:             rp.Index,
        Node:              rp.Node,
        Path:              newPath,
        MaxDepth:          rp.MaxDepth,
        Depth:             rp.Depth,
        DepthIncrement:    rp.DepthIncrement,
        NumEdits:          rp.NumEdits,
        MaxEdits:          rp.MaxEdits,
        NumEditsIncrement: rp.NumEditsIncrement,
        EditableFields:    rp.EditableFields,
        Visited:           newVisited,
    }
}