package utils

import (
	"strings"
	"testing"
)

func TestGeneratePowerDesignerScriptMarksExplicitFKAttributes(t *testing.T) {
	entities := []Entity{
		{
			Id:   1,
			Name: "Pedido",
			Attributes: []Attribute{
				{Name: "Id", Type: "Int", KeyType: AttributeKeyPK},
				{Name: "ClienteId", Type: "Int", KeyType: AttributeKeyFK},
			},
		},
	}

	script, err := GeneratePowerDesignerScript(entities, nil, nil)
	if err != nil {
		t.Fatalf("GeneratePowerDesignerScript() error = %v", err)
	}

	if !strings.Contains(script, "Call TryMarkAsFK(a_l_Pedido_ClienteId)") {
		t.Fatalf("expected explicit FK attribute to be marked as FK in the generated VBS")
	}
}

func TestGeneratePowerDesignerScriptMakesIntersectionFKsMandatoryAndCompositeIdentifier(t *testing.T) {
	entities := []Entity{
		{
			Id:   1,
			Name: "Alumno",
			Attributes: []Attribute{
				{Name: "Id", Type: "Int", KeyType: AttributeKeyPK},
			},
		},
		{
			Id:   2,
			Name: "Curso",
			Attributes: []Attribute{
				{Name: "Id", Type: "Int", KeyType: AttributeKeyPK},
			},
		},
	}
	intersections := []IntersectionEntity{
		{
			RelationID: 10,
			Entity: Entity{
				Id:   3,
				Name: "AlumnoCurso",
			},
		},
	}
	relations := []Relation{
		{Id: 10, IdEntity1: 1, IdEntity2: 2, Relation: RelationTypeNN},
	}

	script, err := GeneratePowerDesignerScript(entities, intersections, relations)
	if err != nil {
		t.Fatalf("GeneratePowerDesignerScript() error = %v", err)
	}

	requiredSnippets := []string{
		"Dim matched_l_AlumnoCurso",
		"Dim id_l_AlumnoCurso",
		"Set id_l_AlumnoCurso = GetOrCreatePrimaryIdentifier(i_l_AlumnoCurso, \"ID_PRIMARY\", \"Identificador Primario\")",
		"Call TrySetMandatory(rel_l_1_1, True, True)",
		"Call TrySetMandatory(rel_l_1_2, True, True)",
		"Call TrySetIdentifying(rel_l_1_1, True)",
		"Call TrySetIdentifying(rel_l_1_2, True)",
		"matched_l_AlumnoCurso = matched_l_AlumnoCurso + TryAddRelationshipJoinChildAttributesToIdentifier(rel_l_1_1, id_l_AlumnoCurso)",
		"matched_l_AlumnoCurso = matched_l_AlumnoCurso + TryAddRelationshipJoinChildAttributesToIdentifier(rel_l_1_2, id_l_AlumnoCurso)",
		"If matched_l_AlumnoCurso > 0 Then Set i_l_AlumnoCurso.PrimaryIdentifier = id_l_AlumnoCurso",
		"If matched_l_AlumnoCurso > 0 Then Call TryPromoteIdentifier(i_l_AlumnoCurso, \"ID_PRIMARY\", \"Identificador Primario\")",
		"Function GetOrCreatePrimaryIdentifier(ent, code, name)",
		"id.Attributes.Insert -1, att",
		"Function TryAddRelationshipJoinChildAttributesToIdentifier(rel, id)",
		"Set childAtt = joinObj.ChildAttribute",
		"If TryAddAttributeToIdentifier(id, childAtt) Then",
		"If att.ForeignIdentifier = True Then",
		"If att.ForeignKey = True Then",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(script, snippet) {
			t.Fatalf("expected generated VBS to contain %q", snippet)
		}
	}

	unexpectedSnippets := []string{
		"a_fk_l_AlumnoCurso_Alumno_Id",
		"a_fk_l_AlumnoCurso_Curso_Id",
	}

	for _, snippet := range unexpectedSnippets {
		if strings.Contains(script, snippet) {
			t.Fatalf("expected generated VBS to avoid manual duplicated intersection key artifact %q", snippet)
		}
	}
}

func TestGeneratePowerDesignerScriptMakesGeneratedChildFKMandatory(t *testing.T) {
	entities := []Entity{
		{
			Id:   1,
			Name: "Cliente",
			Attributes: []Attribute{
				{Name: "Id", Type: "Int", KeyType: AttributeKeyPK},
			},
		},
		{
			Id:   2,
			Name: "Pedido",
			Attributes: []Attribute{
				{Name: "Id", Type: "Int", KeyType: AttributeKeyPK},
			},
		},
	}
	relations := []Relation{
		{Id: 1, IdEntity1: 1, IdEntity2: 2, Relation: RelationType1N},
	}

	script, err := GeneratePowerDesignerScript(entities, nil, relations)
	if err != nil {
		t.Fatalf("GeneratePowerDesignerScript() error = %v", err)
	}

	if !strings.Contains(script, "Clave foranea obligatoria heredada de Cliente.") {
		t.Fatalf("expected generated child FK to be mandatory for 1:N relationships")
	}
}
