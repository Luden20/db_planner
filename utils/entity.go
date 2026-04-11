package utils

import (
	"fmt"
	"strings"
	"unicode"
)

type TableType string

const (
	TableTypeStrong       TableType = "strong"
	TableTypeIntersection TableType = "intersection"
)

type Coordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Entity struct {
	Id          int
	Name        string
	Description string
	Attributes  []Attribute
	Status      *bool
	TableType   TableType
	Coords      *Coordinates `json:"coords"`
}

type IntersectionEntity struct {
	RelationID int
	Entity     Entity
}

func normalizeTableType(value TableType, fallback TableType) TableType {
	switch value {
	case TableTypeStrong, TableTypeIntersection:
		return value
	default:
		return fallback
	}
}

func ensureEntityDefaults(entity *Entity, fallback TableType) {
	if entity.Attributes == nil {
		entity.Attributes = make([]Attribute, 0)
	}
	entity.TableType = normalizeTableType(entity.TableType, fallback)
}

func (p *DbProject) ensureEntities() {
	if p.Entities == nil {
		p.Entities = make([]Entity, 0)
	}
	for idx := range p.Entities {
		ensureEntityDefaults(&p.Entities[idx], TableTypeStrong)
	}

	if p.IntersectionEntities == nil {
		p.IntersectionEntities = make([]IntersectionEntity, 0)
	}
	for idx := range p.IntersectionEntities {
		ensureEntityDefaults(&p.IntersectionEntities[idx].Entity, TableTypeIntersection)
	}
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
	p.Entities = append(p.Entities, Entity{
		Id:          p.EntitiesLastMax + 1,
		Name:        name,
		Description: description,
		Attributes:  make([]Attribute, 0),
		TableType:   TableTypeStrong,
	})
	p.EntitiesLastMax = p.EntitiesLastMax + 1
	return nil
}

func (p *DbProject) intersectionEntityIndexByRelationID(relationID int) int {
	for idx, item := range p.IntersectionEntities {
		if item.RelationID == relationID {
			return idx
		}
	}
	return -1
}

func compactIntersectionNamePart(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), "")
}

func buildIntersectionEntityName(leftName string, rightName string) string {
	return compactIntersectionNamePart(leftName) + compactIntersectionNamePart(rightName)
}

func normalizeIntersectionLookupName(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for _, char := range strings.ToLower(value) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func (p *DbProject) inferIntersectionRelationID(entity Entity, usedRelationIDs map[int]struct{}) int {
	lookupName := normalizeIntersectionLookupName(entity.Name)
	if lookupName == "" {
		return 0
	}

	for _, relation := range p.Relations {
		if relation.Relation != "N:N" {
			continue
		}
		if _, alreadyUsed := usedRelationIDs[relation.Id]; alreadyUsed {
			continue
		}
		left := p.GetEntity(relation.IdEntity1)
		right := p.GetEntity(relation.IdEntity2)
		if left == nil || right == nil {
			continue
		}
		expectedName := normalizeIntersectionLookupName(buildIntersectionEntityName(left.Name, right.Name))
		if lookupName == expectedName {
			return relation.Id
		}
	}

	return 0
}

func (p *DbProject) relinkIntersectionEntities() {
	p.ensureEntities()

	usedRelationIDs := make(map[int]struct{}, len(p.IntersectionEntities))
	for idx := range p.IntersectionEntities {
		relationID := p.IntersectionEntities[idx].RelationID
		if relationID == 0 {
			continue
		}
		relation := p.GetRelationByID(relationID)
		if relation == nil || relation.Relation != "N:N" {
			p.IntersectionEntities[idx].RelationID = 0
			continue
		}
		if _, duplicated := usedRelationIDs[relationID]; duplicated {
			p.IntersectionEntities[idx].RelationID = 0
			continue
		}
		usedRelationIDs[relationID] = struct{}{}
	}

	for idx := range p.IntersectionEntities {
		if p.IntersectionEntities[idx].RelationID != 0 {
			continue
		}
		inferredRelationID := p.inferIntersectionRelationID(p.IntersectionEntities[idx].Entity, usedRelationIDs)
		if inferredRelationID == 0 {
			continue
		}
		p.IntersectionEntities[idx].RelationID = inferredRelationID
		usedRelationIDs[inferredRelationID] = struct{}{}
	}
}

func (p *DbProject) buildIntersectionEntity(relation *Relation, existing *IntersectionEntity) *IntersectionEntity {
	if relation == nil {
		return nil
	}

	left := p.GetEntity(relation.IdEntity1)
	right := p.GetEntity(relation.IdEntity2)
	if left == nil || right == nil {
		return nil
	}

	entity := Entity{
		Name:        buildIntersectionEntityName(left.Name, right.Name),
		Description: fmt.Sprintf("Interseccion generada por la relacion N:N entre %s y %s.", left.Name, right.Name),
		Attributes:  make([]Attribute, 0),
		TableType:   TableTypeIntersection,
	}

	if existing != nil {
		entity.Id = existing.Entity.Id
		entity.Attributes = existing.Entity.Attributes
		entity.Status = existing.Entity.Status
		if strings.TrimSpace(existing.Entity.Description) != "" {
			entity.Description = existing.Entity.Description
		}
		entity.TableType = TableTypeIntersection
		ensureEntityDefaults(&entity, TableTypeIntersection)
	}

	return &IntersectionEntity{
		RelationID: relation.Id,
		Entity:     entity,
	}
}

func (p *DbProject) ensureIntersectionEntityForRelation(relation *Relation) {
	if relation == nil {
		return
	}

	index := p.intersectionEntityIndexByRelationID(relation.Id)
	var existing *IntersectionEntity
	if index >= 0 {
		existing = &p.IntersectionEntities[index]
	}

	if relation.Relation != RelationTypeNN {
		if index >= 0 {
			p.IntersectionEntities = append(p.IntersectionEntities[:index], p.IntersectionEntities[index+1:]...)
		}
		return
	}

	next := p.buildIntersectionEntity(relation, existing)
	if next == nil {
		return
	}

	if existing == nil {
		p.IntersectionEntitiesLastMax++
		next.Entity.Id = p.IntersectionEntitiesLastMax
		p.IntersectionEntities = append(p.IntersectionEntities, *next)
		return
	}

	p.IntersectionEntities[index] = *next
}

func (p *DbProject) removeIntersectionEntityByRelationID(relationID int) {
	index := p.intersectionEntityIndexByRelationID(relationID)
	if index < 0 {
		return
	}
	p.IntersectionEntities = append(p.IntersectionEntities[:index], p.IntersectionEntities[index+1:]...)
}

func (p *DbProject) GetIntersectionEntityByRelationID(relationID int) *IntersectionEntity {
	index := p.intersectionEntityIndexByRelationID(relationID)
	if index < 0 {
		return nil
	}
	ensureEntityDefaults(&p.IntersectionEntities[index].Entity, TableTypeIntersection)
	return &p.IntersectionEntities[index]
}

func (p *DbProject) EditIntersectionEntityDescription(relationID int, description string) error {
	item := p.GetIntersectionEntityByRelationID(relationID)
	if item == nil {
		return fmt.Errorf("intersection entity not found")
	}
	item.Entity.Description = description
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
	if status && entity.TableType == TableTypeStrong {
		if countPrimaryKeys(entity.Attributes) != 1 {
			return fmt.Errorf("una entidad fuerte debe tener exactamente una PK para aprobarse")
		}
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
			p.removeIntersectionEntityByRelationID(relation.Id)
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
		ensureEntityDefaults(&p.Entities[idx], TableTypeStrong)
		return &p.Entities[idx]
	}
	return nil
}

func (p *DbProject) IntersectionHasAttributes(relationID int) bool {
	intersection := p.GetIntersectionEntityByRelationID(relationID)
	if intersection == nil {
		return false
	}
	return len(intersection.Entity.Attributes) > 0
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

func (p *DbProject) UpdateEntityCoords(id int, x float64, y float64, isIntersection bool) error {
	if isIntersection {
		for idx := range p.IntersectionEntities {
			if p.IntersectionEntities[idx].Entity.Id == id {
				p.IntersectionEntities[idx].Entity.Coords = &Coordinates{X: x, Y: y}
				return nil
			}
		}
		return fmt.Errorf("intersection entity not found")
	}

	idx := p.entityIndex(id)
	if idx < 0 {
		return fmt.Errorf("entity not found")
	}
	p.Entities[idx].Coords = &Coordinates{X: x, Y: y}
	return nil
}

