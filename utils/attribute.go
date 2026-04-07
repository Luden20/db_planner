package utils

import "fmt"

type Attribute struct {
	Id          int
	Name        string
	Description string
	Type        string
}

func defaultAttributeType(value string) string {
	if value == "" {
		return "Por definir"
	}
	return value
}

func (p *DbProject) AddAttribute(entityId int, name string, description string, attType string) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("Entity Not Found")
	}
	if ent.Attributes == nil {
		ent.Attributes = make([]Attribute, 0)
	}

	p.AttributesLastMax++
	att := Attribute{
		Id:          p.AttributesLastMax,
		Name:        name,
		Description: description,
		Type:        defaultAttributeType(attType),
	}
	ent.Attributes = append(ent.Attributes, att)
	return nil
}

func (p *DbProject) EditAttribute(entityId int, attributeId int, name string, description string, attType string) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("Entity Not Found")
	}
	for idx := range ent.Attributes {
		if ent.Attributes[idx].Id == attributeId {
			ent.Attributes[idx].Name = name
			ent.Attributes[idx].Description = description
			ent.Attributes[idx].Type = defaultAttributeType(attType)
			return nil
		}
	}
	return fmt.Errorf("Attribute Not Found")
}

func (p *DbProject) RemoveAttribute(entityId int, attId int) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("Entity Not Found")
	}

	atts := make([]Attribute, 0)
	for _, attribute := range ent.Attributes {
		if attribute.Id == attId {
			continue
		}
		atts = append(atts, attribute)
	}
	ent.Attributes = atts
	return nil
}

func (p *DbProject) ensureAttributes() {
	for idx := range p.Entities {
		if p.Entities[idx].Attributes == nil {
			p.Entities[idx].Attributes = make([]Attribute, 0)
		}
		for attIdx := range p.Entities[idx].Attributes {
			p.Entities[idx].Attributes[attIdx].Type = defaultAttributeType(p.Entities[idx].Attributes[attIdx].Type)
		}
	}
}

func (p *DbProject) MoveAttribute(entityId int, attributeId int, direction string) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("entity not found")
	}
	if len(ent.Attributes) <= 1 {
		return nil
	}

	attrIdx := -1
	for idx, att := range ent.Attributes {
		if att.Id == attributeId {
			attrIdx = idx
			break
		}
	}
	if attrIdx == -1 {
		return fmt.Errorf("attribute not found")
	}

	switch direction {
	case "up":
		if attrIdx == 0 {
			return nil
		}
		ent.Attributes[attrIdx], ent.Attributes[attrIdx-1] = ent.Attributes[attrIdx-1], ent.Attributes[attrIdx]
	case "down":
		if attrIdx == len(ent.Attributes)-1 {
			return nil
		}
		ent.Attributes[attrIdx], ent.Attributes[attrIdx+1] = ent.Attributes[attrIdx+1], ent.Attributes[attrIdx]
	default:
		return fmt.Errorf("unknown direction %s", direction)
	}
	return nil
}
