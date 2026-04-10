package utils

import "fmt"

const (
	RelationType11  = "1:1"
	RelationType1N  = "1:N"
	RelationTypeN1  = "N:1"
	RelationTypeNN  = "N:N"
	RelationType1Np = "1:Np"
	RelationTypeNp1 = "Np:1"
)

var allowedRelationValues = map[string]struct{}{
	RelationType11:  {},
	RelationType1N:  {},
	RelationTypeN1:  {},
	RelationTypeNN:  {},
	RelationType1Np: {},
	RelationTypeNp1: {},
}

func GetAllowedRelationTypes() []string {
	return []string{
		"",
		RelationType11,
		RelationType1N,
		RelationTypeN1,
		RelationTypeNN,
		RelationType1Np,
		RelationTypeNp1,
	}
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

func (p *DbProject) GetRelationByID(id int) *Relation {
	for idx := range p.Relations {
		if p.Relations[idx].Id == id {
			return &p.Relations[idx]
		}
	}
	return nil
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
	case RelationType11:
		return RelationType11
	case RelationType1N:
		return RelationTypeN1
	case RelationTypeN1:
		return RelationType1N
	case RelationTypeNN:
		return RelationTypeNN
	case RelationType1Np:
		return RelationTypeNp1
	case RelationTypeNp1:
		return RelationType1Np
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

func (p *DbProject) syncIntersectionEntitiesFromRelations() {
	p.ensureEntities()
	for idx := range p.Relations {
		p.ensureIntersectionEntityForRelation(&p.Relations[idx])
	}
	validRelationIDs := make(map[int]struct{}, len(p.Relations))
	for _, relation := range p.Relations {
		if relation.Relation == RelationTypeNN {
			validRelationIDs[relation.Id] = struct{}{}
		}
	}
	filtered := make([]IntersectionEntity, 0, len(p.IntersectionEntities))
	for _, item := range p.IntersectionEntities {
		if _, ok := validRelationIDs[item.RelationID]; ok {
			filtered = append(filtered, item)
		}
	}
	p.IntersectionEntities = filtered
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
		p.ensureIntersectionEntityForRelation(rel)
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
	p.ensureIntersectionEntityForRelation(&p.Relations[len(p.Relations)-1])
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
	p.removeIntersectionEntityByRelationID(id)
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
