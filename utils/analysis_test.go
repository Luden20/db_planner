package utils

import (
	"testing"
)

func TestAnalyzeProjectDependencies_Cycle(t *testing.T) {
	p := &DbProject{
		Entities: []Entity{
			{Id: 1, Name: "A"}, // Parent
			{Id: 2, Name: "B"}, // Child
		},
		Relations: []Relation{
			{Id: 1, IdEntity1: 1, IdEntity2: 2, Relation: RelationType1N}, // A -> B
			{Id: 2, IdEntity1: 1, IdEntity2: 2, Relation: RelationTypeN1}, // B -> A
		},
	}

	report := p.AnalyzeProjectDependencies()

	hasCircular := false
	for _, issue := range report.Issues {
		if issue.Type == "Circular" {
			hasCircular = true
			break
		}
	}

	if !hasCircular {
		t.Errorf("Expected a circular issue, but none found")
	}
}

func TestAnalyzeProjectDependencies_Ambiguity(t *testing.T) {
	// A -> B -> D
	// A -> C -> D
	// Direction: Parent -> Child
	p := &DbProject{
		Entities: []Entity{
			{Id: 1, Name: "A"},
			{Id: 2, Name: "B"},
			{Id: 3, Name: "C"},
			{Id: 4, Name: "D"},
		},
		Relations: []Relation{
			{Id: 1, IdEntity1: 1, IdEntity2: 2, Relation: RelationType1N}, // A -> B
			{Id: 2, IdEntity1: 1, IdEntity2: 3, Relation: RelationType1N}, // A -> C
			{Id: 3, IdEntity1: 2, IdEntity2: 4, Relation: RelationType1N}, // B -> D
			{Id: 4, IdEntity1: 3, IdEntity2: 4, Relation: RelationType1N}, // C -> D
		},
	}

	report := p.AnalyzeProjectDependencies()

	found := false
	for _, issue := range report.Issues {
		if issue.Type == "Ambiguity" && issue.Entities[0] == "A" && issue.Entities[1] == "D" {
			if len(issue.Paths) == 2 {
				found = true
				break
			}
		}
	}

	if !found {
		t.Errorf("Expected ambiguity issue with 2 paths from A to D, but none found")
	}
}

func TestAnalyzeProjectDependencies_InheritedAmbiguity(t *testing.T) {
    // Z -> A -> B -> D
    // Z -> A -> C -> D
    // Ambigüedad nace en A -> D.
    // Z -> D tiene ambigüedad pero es heredada de A.
    // Debe reportarse A -> D y suprimirse Z -> D.
    p := &DbProject{
		Entities: []Entity{
            {Id: 0, Name: "Z"},
			{Id: 1, Name: "A"},
			{Id: 2, Name: "B"},
			{Id: 3, Name: "C"},
			{Id: 4, Name: "D"},
		},
		Relations: []Relation{
            {Id: 0, IdEntity1: 0, IdEntity2: 1, Relation: RelationType1N}, // Z -> A
			{Id: 1, IdEntity1: 1, IdEntity2: 2, Relation: RelationType1N}, // A -> B
			{Id: 2, IdEntity1: 1, IdEntity2: 3, Relation: RelationType1N}, // A -> C
			{Id: 3, IdEntity1: 2, IdEntity2: 4, Relation: RelationType1N}, // B -> D
			{Id: 4, IdEntity1: 3, IdEntity2: 4, Relation: RelationType1N}, // C -> D
		},
	}

    report := p.AnalyzeProjectDependencies()

    foundAD := false
    foundZD := false
    for _, issue := range report.Issues {
        if issue.Type == "Ambiguity" {
            if issue.Entities[0] == "A" && issue.Entities[1] == "D" {
                foundAD = true
            }
            if issue.Entities[0] == "Z" && issue.Entities[1] == "D" {
                foundZD = true
            }
        }
    }

    if !foundAD {
        t.Errorf("Expected to find A -> D as the root cause of ambiguity")
    }
    if foundZD {
        t.Errorf("Expected Z -> D to be suppressed as it inherits ambiguity from A")
    }
}

func TestAnalyzeProjectDependencies_IDCollision(t *testing.T) {
	p := &DbProject{
		Entities: []Entity{
			{Id: 1, Name: "A"},
			{Id: 2, Name: "B"},
			{Id: 3, Name: "C"},
		},
		Relations: []Relation{
			{Id: 10, IdEntity1: 2, IdEntity2: 3, Relation: RelationTypeNN},
		},
		IntersectionEntities: []IntersectionEntity{
			{RelationID: 10, Entity: Entity{Id: 1, Name: "IE"}}, // MISMO ID QUE ENTIDAD "A"
		},
	}

	g, _ := p.buildDependencyGraph()
	if len(g.nodeMap) != 4 {
		t.Errorf("Expected 4 distinct nodes, got %d", len(g.nodeMap))
	}
}
