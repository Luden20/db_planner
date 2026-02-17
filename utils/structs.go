package utils

import "slices"

type DbProject struct {
	Name      string
	Entities  []Entity
	Relations []Relation
}
type Entity struct {
	Id          int
	Name        string
	Description string
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

func (p *DbProject) GetCombinatoryModel() []RelationView {
	var combinatory []RelationView
	for idx, entity := range p.Entities {
		var relations []RelationViewItem
		var comb = p.Entities[idx:]
		for _, r_entity := range comb {
			relation := p.GetRelationByEntities(entity.Id, r_entity.Id)
			var r_item RelationViewItem
			if relation != nil {
				r_item.Id = &relation.Id
				r_item.Relation = &relation.Relation
			} else {
				r_item.Id = nil
				r_item.Relation = nil
			}
			r_item.IdEntity2 = r_entity.Id
			r_item.Entity2 = r_entity.Name
			relations = append(relations, r_item)
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
func (p *DbProject) AddEntity(name string, description string) {
	p.Entities = append(p.Entities, Entity{Id: len(p.Entities) + 1, Name: name, Description: description})
}
func (p *DbProject) GetRelationByEntities(idEntity1 int, idEntity2 int) *Relation {
	for idx, relation := range p.Relations {
		if relation.IdEntity1 == idEntity1 && relation.IdEntity1 == idEntity2 {
			return &p.Relations[idx]
		}
	}
	return nil
}
func (p *DbProject) GetEntity(id int) *Entity {
	idx, found := slices.BinarySearchFunc(p.Entities, id, func(e Entity, target int) int {
		if e.Id < target {
			return -1
		}
		if e.Id > target {
			return 1
		}
		return 0
	})
	if found {
		return &p.Entities[idx]
	}
	return nil
}
