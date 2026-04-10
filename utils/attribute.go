package utils

import (
	"fmt"
	"strings"
)

type AttributeKeyType string

const (
	AttributeKeyNone AttributeKeyType = "nil"
	AttributeKeyPK   AttributeKeyType = "pk"
	AttributeKeyFK   AttributeKeyType = "fk"
)

type Attribute struct {
	Id int

	Name        string
	Description string
	Type        string
	KeyType     AttributeKeyType
	Domain      []string
}

func normalizeDomain(values []string) []string {
	if len(values) == 0 {
		return make([]string, 0)
	}

	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func defaultAttributeType(value string) string {
	if value == "" {
		return "Por definir"
	}
	return value
}

func normalizeAttributeKeyType(value AttributeKeyType) AttributeKeyType {
	switch value {
	case AttributeKeyPK, AttributeKeyNone:
		return value
	default:
		return AttributeKeyNone
	}
}

func countPrimaryKeys(attributes []Attribute) int {
	total := 0
	for _, attribute := range attributes {
		if normalizeAttributeKeyType(attribute.KeyType) == AttributeKeyPK {
			total++
		}
	}
	return total
}

func ensureStrongEntityApprovalConsistency(entity *Entity) {
	if entity == nil || entity.TableType != TableTypeStrong {
		return
	}
	if countPrimaryKeys(entity.Attributes) != 1 {
		approved := false
		entity.Status = &approved
	}
}

func validateAttributeKeyChange(attributes []Attribute, currentAttributeID int, nextKeyType AttributeKeyType) error {
	if normalizeAttributeKeyType(nextKeyType) != AttributeKeyPK {
		return nil
	}
	for _, attribute := range attributes {
		if attribute.Id == currentAttributeID {
			continue
		}
		if normalizeAttributeKeyType(attribute.KeyType) == AttributeKeyPK {
			return fmt.Errorf("solo se permite una PK por entidad fuerte")
		}
	}
	return nil
}

func (p *DbProject) DeleteFromDomain(entityId int, attId int, value string) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("Entity Not Found")
	}

	for idx := range ent.Attributes {
		if ent.Attributes[idx].Id != attId {
			continue
		}
		filtered := make([]string, 0, len(ent.Attributes[idx].Domain))
		for _, domainValue := range ent.Attributes[idx].Domain {
			if domainValue == value {
				continue
			}
			filtered = append(filtered, domainValue)
		}
		ent.Attributes[idx].Domain = normalizeDomain(filtered)
		return nil
	}
	return fmt.Errorf("Attribute Not Found")
}

func (p *DbProject) AddToDomain(entityId int, attributeId int, value string) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("Entity Not Found")
	}

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("Domain value is empty")
	}

	for idx := range ent.Attributes {
		if ent.Attributes[idx].Id != attributeId {
			continue
		}
		for _, domainValue := range ent.Attributes[idx].Domain {
			if domainValue == trimmed {
				return fmt.Errorf("Value already in domain")
			}
		}
		ent.Attributes[idx].Domain = append(ent.Attributes[idx].Domain, trimmed)
		return nil
	}
	return fmt.Errorf("Attribute Not Found")
}

func (p *DbProject) AddAttribute(entityId int, name string, description string, attType string, attKeyType AttributeKeyType, attDomain []string) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("Entity Not Found")
	}
	if ent.Attributes == nil {
		ent.Attributes = make([]Attribute, 0)
	}
	nextKeyType := normalizeAttributeKeyType(attKeyType)
	if ent.TableType == TableTypeStrong {
		if err := validateAttributeKeyChange(ent.Attributes, 0, nextKeyType); err != nil {
			return err
		}
	}

	p.AttributesLastMax++
	att := Attribute{
		Id:          p.AttributesLastMax,
		Name:        name,
		Description: description,
		Type:        defaultAttributeType(attType),
		KeyType:     nextKeyType,
		Domain:      normalizeDomain(attDomain),
	}
	ent.Attributes = append(ent.Attributes, att)
	ensureStrongEntityApprovalConsistency(ent)
	return nil
}

func (p *DbProject) AddIntersectionAttribute(relationID int, name string, description string, attType string, attDomain []string) error {
	item := p.GetIntersectionEntityByRelationID(relationID)
	if item == nil {
		return fmt.Errorf("Intersection Entity Not Found")
	}
	if item.Entity.Attributes == nil {
		item.Entity.Attributes = make([]Attribute, 0)
	}

	p.AttributesLastMax++
	att := Attribute{
		Id:          p.AttributesLastMax,
		Name:        name,
		Description: description,
		Type:        defaultAttributeType(attType),
		KeyType:     AttributeKeyNone,
		Domain:      normalizeDomain(attDomain),
	}
	item.Entity.Attributes = append(item.Entity.Attributes, att)
	return nil
}

func (p *DbProject) EditAttribute(entityId int, attributeId int, name string, description string, attType string, attKeyType AttributeKeyType, attDomain []string) error {
	ent := p.GetEntity(entityId)
	if ent == nil {
		return fmt.Errorf("Entity Not Found")
	}
	nextKeyType := normalizeAttributeKeyType(attKeyType)
	if ent.TableType == TableTypeStrong {
		if err := validateAttributeKeyChange(ent.Attributes, attributeId, nextKeyType); err != nil {
			return err
		}
	}
	for idx := range ent.Attributes {
		if ent.Attributes[idx].Id == attributeId {
			ent.Attributes[idx].Name = name
			ent.Attributes[idx].Description = description
			ent.Attributes[idx].Type = defaultAttributeType(attType)
			ent.Attributes[idx].KeyType = nextKeyType
			ent.Attributes[idx].Domain = normalizeDomain(attDomain)
			ensureStrongEntityApprovalConsistency(ent)
			return nil
		}
	}
	return fmt.Errorf("Attribute Not Found")
}

func (p *DbProject) EditIntersectionAttribute(relationID int, attributeId int, name string, description string, attType string, attDomain []string) error {
	item := p.GetIntersectionEntityByRelationID(relationID)
	if item == nil {
		return fmt.Errorf("Intersection Entity Not Found")
	}
	for idx := range item.Entity.Attributes {
		if item.Entity.Attributes[idx].Id == attributeId {
			item.Entity.Attributes[idx].Name = name
			item.Entity.Attributes[idx].Description = description
			item.Entity.Attributes[idx].Type = defaultAttributeType(attType)
			item.Entity.Attributes[idx].KeyType = AttributeKeyNone
			item.Entity.Attributes[idx].Domain = normalizeDomain(attDomain)
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
	ensureStrongEntityApprovalConsistency(ent)
	return nil
}

func (p *DbProject) RemoveIntersectionAttribute(relationID int, attId int) error {
	item := p.GetIntersectionEntityByRelationID(relationID)
	if item == nil {
		return fmt.Errorf("Intersection Entity Not Found")
	}

	atts := make([]Attribute, 0, len(item.Entity.Attributes))
	for _, attribute := range item.Entity.Attributes {
		if attribute.Id == attId {
			continue
		}
		atts = append(atts, attribute)
	}
	item.Entity.Attributes = atts
	return nil
}

func (p *DbProject) ensureAttributes() {
	for idx := range p.Entities {
		if p.Entities[idx].Attributes == nil {
			p.Entities[idx].Attributes = make([]Attribute, 0)
		}
		for attIdx := range p.Entities[idx].Attributes {
			p.Entities[idx].Attributes[attIdx].Type = defaultAttributeType(p.Entities[idx].Attributes[attIdx].Type)
			p.Entities[idx].Attributes[attIdx].KeyType = normalizeAttributeKeyType(p.Entities[idx].Attributes[attIdx].KeyType)
			p.Entities[idx].Attributes[attIdx].Domain = normalizeDomain(p.Entities[idx].Attributes[attIdx].Domain)
		}
		ensureStrongEntityApprovalConsistency(&p.Entities[idx])
	}
	for idx := range p.IntersectionEntities {
		if p.IntersectionEntities[idx].Entity.Attributes == nil {
			p.IntersectionEntities[idx].Entity.Attributes = make([]Attribute, 0)
		}
		for attIdx := range p.IntersectionEntities[idx].Entity.Attributes {
			p.IntersectionEntities[idx].Entity.Attributes[attIdx].Type = defaultAttributeType(p.IntersectionEntities[idx].Entity.Attributes[attIdx].Type)
			p.IntersectionEntities[idx].Entity.Attributes[attIdx].KeyType = AttributeKeyNone
			p.IntersectionEntities[idx].Entity.Attributes[attIdx].Domain = normalizeDomain(p.IntersectionEntities[idx].Entity.Attributes[attIdx].Domain)
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

func (p *DbProject) MoveIntersectionAttribute(relationID int, attributeId int, direction string) error {
	item := p.GetIntersectionEntityByRelationID(relationID)
	if item == nil {
		return fmt.Errorf("intersection entity not found")
	}
	if len(item.Entity.Attributes) <= 1 {
		return nil
	}

	attrIdx := -1
	for idx, att := range item.Entity.Attributes {
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
		item.Entity.Attributes[attrIdx], item.Entity.Attributes[attrIdx-1] = item.Entity.Attributes[attrIdx-1], item.Entity.Attributes[attrIdx]
	case "down":
		if attrIdx == len(item.Entity.Attributes)-1 {
			return nil
		}
		item.Entity.Attributes[attrIdx], item.Entity.Attributes[attrIdx+1] = item.Entity.Attributes[attrIdx+1], item.Entity.Attributes[attrIdx]
	default:
		return fmt.Errorf("unknown direction %s", direction)
	}
	return nil
}
