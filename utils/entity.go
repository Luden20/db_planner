package utils

import "fmt"

type Entity struct {
	Id          int
	Name        string
	Description string
	Attributes  []Attribute
	Status      *bool
}

func (p *DbProject) entityIndex(id int) int {
	for idx, entity := range p.Entities {
		if entity.Id == id {
			return idx
		}
	}
	return -1
}

func (p *DbProject) AddEntity(name string, description string) error {
	p.Entities = append(p.Entities, Entity{Id: p.EntitiesLastMax + 1, Name: name, Description: description, Attributes: make([]Attribute, 0)})
	p.EntitiesLastMax = p.EntitiesLastMax + 1
	return nil
}

func (p *DbProject) EditEntity(id int, name string, description string) error {
	entity := p.GetEntity(id)
	if entity == nil {
		return fmt.Errorf("entity not founded")
	}
	entity.Name = name
	entity.Description = description
	return nil
}

func (p *DbProject) MarkEntityStatus(id int, status bool) error {
	entity := p.GetEntity(id)
	if entity == nil {
		return fmt.Errorf("entity not founded")
	}
	entity.Status = &status
	return nil
}

func (p *DbProject) RemoveEntity(id int) error {
	newEnts := make([]Entity, 0)
	for _, entity := range p.Entities {
		if entity.Id == id {
			continue
		}
		newEnts = append(newEnts, entity)
	}
	p.Entities = newEnts

	var newRelations []Relation
	for _, relation := range p.Relations {
		if relation.IdEntity1 == id || relation.IdEntity2 == id {
			continue
		}
		newRelations = append(newRelations, relation)
	}
	p.Relations = newRelations
	p.removeRolePermissionsByTableID(id)
	return nil
}

func (p *DbProject) GetEntity(id int) *Entity {
	idx := p.entityIndex(id)
	if idx >= 0 {
		if p.Entities[idx].Attributes == nil {
			p.Entities[idx].Attributes = make([]Attribute, 0)
		}
		return &p.Entities[idx]
	}
	return nil
}

func (p *DbProject) MoveEntity(id int, direction string) error {
	idx := p.entityIndex(id)
	if idx < 0 {
		return fmt.Errorf("entity not found")
	}
	if len(p.Entities) <= 1 {
		return nil
	}

	switch direction {
	case "up":
		if idx == 0 {
			return nil
		}
		p.Entities[idx], p.Entities[idx-1] = p.Entities[idx-1], p.Entities[idx]
	case "down":
		if idx == len(p.Entities)-1 {
			return nil
		}
		p.Entities[idx], p.Entities[idx+1] = p.Entities[idx+1], p.Entities[idx]
	default:
		return fmt.Errorf("unknown direction %s", direction)
	}
	return nil
}
