package utils

import "testing"

func TestInvertRelationValueKeepsOneToOne(t *testing.T) {
	if got := invertRelationValue("1:1"); got != "1:1" {
		t.Fatalf("expected 1:1 to remain unchanged, got %q", got)
	}
}

func TestAddRelationAcceptsOneToOneWhenIdsAreSwapped(t *testing.T) {
	project := &DbProject{
		Relations: make([]Relation, 0),
	}

	if err := project.AddRelation(5, 2, "1:1"); err != nil {
		t.Fatalf("expected AddRelation to accept 1:1, got error: %v", err)
	}

	rel := project.GetRelationByEntities(2, 5)
	if rel == nil {
		t.Fatal("expected relation to be stored")
	}
	if rel.Relation != "1:1" {
		t.Fatalf("expected stored relation to be 1:1, got %q", rel.Relation)
	}
}

func TestAddRelationRejectsUnknownType(t *testing.T) {
	project := &DbProject{
		Relations: make([]Relation, 0),
	}

	if err := project.AddRelation(1, 2, "2:3"); err == nil {
		t.Fatal("expected invalid relation type to be rejected")
	}
}
