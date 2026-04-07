package utils

import "fmt"

var allowedRelationValues = map[string]struct{}{
	"1:1": {},
	"1:N": {},
	"N:1": {},
	"N:N": {},
}

type Relation struct {
	Id        int
	IdEntity1 int
	IdEntity2 int
	Relation  string
}

type RelationViewItem struct {
	Id        *int
	Entity2   string
	IdEntity2 int
	Relation  *string
}

type RelationView struct {
	PrincipalEntity   string
	IdPrincipalEntity int
	Relations         []RelationViewItem
}

func normalizePair(id1 int, id2 int) (int, int) {
	if id1 <= id2 {
		return id1, id2
	}
	return id2, id1
}

func canonicalPair(id1 int, id2 int) (int, int, bool) {
	if id1 <= id2 {
		return id1, id2, false
	}
	return id2, id1, true
}

func invertRelationValue(rel string) string {
	switch rel {
	case "1:1":
		return "1:1"
	case "1:N":
		return "N:1"
	case "N:1":
		return "1:N"
	default:
		return rel
	}
}

func isAllowedRelationValue(rel string) bool {
	_, ok := allowedRelationValues[rel]
	return ok
}

func (p *DbProject) normalizeRelations() {
	for idx := range p.Relations {
		if p.Relations[idx].IdEntity1 > p.Relations[idx].IdEntity2 {
			p.Relations[idx].IdEntity1, p.Relations[idx].IdEntity2 = p.Relations[idx].IdEntity2, p.Relations[idx].IdEntity1
			p.Relations[idx].Relation = invertRelationValue(p.Relations[idx].Relation)
		}
	}
}

func (p *DbProject) AddRelation(idEnt1 int, idEnt2 int, relation string) error {
	if !isAllowedRelationValue(relation) {
		return fmt.Errorf("invalid relation type: %s", relation)
	}

	idEnt1, idEnt2, swapped := canonicalPair(idEnt1, idEnt2)
	if swapped {
		relation = invertRelationValue(relation)
	}

	rel := p.GetRelationByEntities(idEnt1, idEnt2)
	if rel != nil {
		rel.IdEntity1 = idEnt1
		rel.IdEntity2 = idEnt2
		rel.Relation = relation
		return nil
	}

	id := p.RelationsLastMax
	p.RelationsLastMax = p.RelationsLastMax + 1
	newRelation := Relation{
		Id:        id,
		IdEntity1: idEnt1,
		IdEntity2: idEnt2,
		Relation:  relation,
	}
	p.Relations = append(p.Relations, newRelation)
	return nil
}

func (p *DbProject) RemoveRelation(id int) error {
	newRels := make([]Relation, 0)
	for _, relation := range p.Relations {
		if relation.Id == id {
			continue
		}
		newRels = append(newRels, relation)
	}
	p.Relations = newRels
	return nil
}

func (p *DbProject) GetRelationByEntities(idEntity1 int, idEntity2 int) *Relation {
	idEntity1, idEntity2 = normalizePair(idEntity1, idEntity2)
	for idx, relation := range p.Relations {
		r1, r2 := normalizePair(relation.IdEntity1, relation.IdEntity2)
		if r1 == idEntity1 && r2 == idEntity2 {
			return &p.Relations[idx]
		}
	}
	return nil
}

func (p *DbProject) GetCombinatoryModel() []RelationView {
	var combinatory []RelationView
	for idx, entity := range p.Entities {
		var relations []RelationViewItem
		comb := p.Entities[idx+1:]
		for _, rEntity := range comb {
			id1, id2, swapped := canonicalPair(entity.Id, rEntity.Id)
			relation := p.GetRelationByEntities(id1, id2)
			var rItem RelationViewItem
			if relation != nil {
				rItem.Id = &relation.Id
				relValue := relation.Relation
				if swapped {
					relValue = invertRelationValue(relValue)
				}
				rItem.Relation = &relValue
			} else {
				rItem.Id = nil
				rItem.Relation = nil
			}
			rItem.IdEntity2 = rEntity.Id
			rItem.Entity2 = rEntity.Name
			relations = append(relations, rItem)
		}
		combinatory = append(combinatory, RelationView{
			PrincipalEntity:   entity.Name,
			IdPrincipalEntity: entity.Id,
			Relations:         relations,
		})
		relations = make([]RelationViewItem, 0)
	}
	return combinatory
}
