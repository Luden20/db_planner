package utils

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

var allowedRelationValues = map[string]struct{}{
	"1:1": {},
	"1:N": {},
	"N:1": {},
	"N:N": {},
}

type DbProject struct {
	Name              string
	Entities          []Entity
	EntitiesLastMax   int
	Relations         []Relation
	RelationsLastMax  int
	AttributesLastMax int
}
type Entity struct {
	Id          int
	Name        string
	Description string
	Attributes  []Attribute
	Status      *bool
}
type Attribute struct {
	Id          int
	Name        string
	Description string
	Type        string
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
	var atts = make([]Attribute, 0)
	for _, attribute := range ent.Attributes {
		if attribute.Id == attId {
			continue
		}
		atts = append(atts, attribute)
	}
	ent.Attributes = atts
	return nil
}
func (p *DbProject) entityIndex(id int) int {
	for idx, entity := range p.Entities {
		if entity.Id == id {
			return idx
		}
	}
	return -1
}

func (p *DbProject) syncCounters() {
	maxEntityID := 0
	for _, entity := range p.Entities {
		if entity.Id > maxEntityID {
			maxEntityID = entity.Id
		}
	}
	if p.EntitiesLastMax < maxEntityID {
		p.EntitiesLastMax = maxEntityID
	}

	maxRelationID := -1
	for _, relation := range p.Relations {
		if relation.Id > maxRelationID {
			maxRelationID = relation.Id
		}
	}
	nextRelationID := maxRelationID + 1
	if p.RelationsLastMax < nextRelationID {
		p.RelationsLastMax = nextRelationID
	}

	maxAttributeID := 0
	p.ensureAttributes()
	for _, entity := range p.Entities {
		for _, att := range entity.Attributes {
			if att.Id > maxAttributeID {
				maxAttributeID = att.Id
			}
		}
	}
	if p.AttributesLastMax < maxAttributeID {
		p.AttributesLastMax = maxAttributeID
	}
}

func (p *DbProject) normalizeRelations() {
	for idx := range p.Relations {
		if p.Relations[idx].IdEntity1 > p.Relations[idx].IdEntity2 {
			p.Relations[idx].IdEntity1, p.Relations[idx].IdEntity2 = p.Relations[idx].IdEntity2, p.Relations[idx].IdEntity1
			p.Relations[idx].Relation = invertRelationValue(p.Relations[idx].Relation)
		}
	}
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

func (p *DbProject) ExportToExcel(filename string) error {
	f := excelize.NewFile()
	defer func(f *excelize.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)
	//Write entities
	entitiesIndex, err := f.NewSheet("Entities")
	if err != nil {
		return err
	}
	f.SetActiveSheet(entitiesIndex)
	col, row, err := excelize.CellNameToCoordinates("A1")
	if err != nil {
		return err
	}
	for _, value := range p.Entities {
		nameCord, err := excelize.CoordinatesToCellName(col, row)
		if err != nil {
			return err
		}
		if err := f.SetCellValue("Entities", nameCord, value.Name); err != nil {
			return err
		}
		defCord, err := excelize.CoordinatesToCellName(col+1, row)
		if err != nil {
			return err
		}
		if err := f.SetCellValue("Entities", defCord, value.Description); err != nil {
			return err
		}
		row = row + 1
	}
	relationsIndex, err := f.NewSheet("Relations")
	if err != nil {
		return err
	}
	f.SetActiveSheet(relationsIndex)
	col, row, err = excelize.CellNameToCoordinates("A1")
	if err != nil {
		return err
	}
	relations := p.GetCombinatoryModel()
	startCol := col
	for _, value := range relations {
		relRow := row
		if len(value.Relations) == 0 {
			principalCell, err := excelize.CoordinatesToCellName(startCol, relRow)
			if err != nil {
				return err
			}
			if err := f.SetCellValue("Relations", principalCell, value.PrincipalEntity); err != nil {
				return err
			}
			startCol += 4
			continue
		}
		for _, rel := range value.Relations {
			relationValue := ""
			if rel.Relation != nil {
				relationValue = *rel.Relation
			}
			if relRow == row {
				principalCell, err := excelize.CoordinatesToCellName(startCol, relRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Relations", principalCell, value.PrincipalEntity); err != nil {
					return err
				}
			}

			nameCell, err := excelize.CoordinatesToCellName(startCol+1, relRow)
			if err != nil {
				return err
			}
			if err := f.SetCellValue("Relations", nameCell, rel.Entity2); err != nil {
				return err
			}

			relationCell, err := excelize.CoordinatesToCellName(startCol+2, relRow)
			if err != nil {
				return err
			}
			if err := f.SetCellValue("Relations", relationCell, relationValue); err != nil {
				return err
			}
			relRow++
		}
		startCol += 4
	}
	_, err = f.NewSheet("Attributes")
	if err != nil {
		return err
	}
	col, row, err = excelize.CellNameToCoordinates("A1")
	if err != nil {
		return err
	}
	startCol = col
	for _, ent := range p.Entities {
		attrRow := row
		if len(ent.Attributes) == 0 {
			entityCell, err := excelize.CoordinatesToCellName(startCol, attrRow)
			if err != nil {
				return err
			}
			if err := f.SetCellValue("Attributes", entityCell, ent.Name); err != nil {
				return err
			}
			nameCell, err := excelize.CoordinatesToCellName(startCol+1, attrRow)
			if err != nil {
				return err
			}
			if err := f.SetCellValue("Attributes", nameCell, "Sin atributos definidos"); err != nil {
				return err
			}
			attrRow++
		} else {
			named := false
			for _, att := range ent.Attributes {
				entityCell, err := excelize.CoordinatesToCellName(startCol, attrRow)
				if err != nil {
					return err
				}
				if !named {
					if err := f.SetCellValue("Attributes", entityCell, ent.Name); err != nil {
						return err
					}
					named = true
				}
				nameCell, err := excelize.CoordinatesToCellName(startCol+1, attrRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Attributes", nameCell, att.Name); err != nil {
					return err
				}
				descCell, err := excelize.CoordinatesToCellName(startCol+2, attrRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Attributes", descCell, att.Description); err != nil {
					return err
				}
				typeCell, err := excelize.CoordinatesToCellName(startCol+3, attrRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Attributes", typeCell, defaultAttributeType(att.Type)); err != nil {
					return err
				}
				attrRow++
			}
		}
		startCol += 5
	}
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return err
	}
	return f.SaveAs(filename)
}
func (p *DbProject) AddRelation(idEnt1 int, idEnt2 int, relation string) error {
	if !isAllowedRelationValue(relation) {
		return fmt.Errorf("invalid relation type: %s", relation)
	}
	idEnt1, idEnt2, swapped := canonicalPair(idEnt1, idEnt2)
	if swapped {
		relation = invertRelationValue(relation)
	}
	// check if exist and update in place
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
func (p *DbProject) GetCombinatoryModel() []RelationView {
	var combinatory []RelationView
	for idx, entity := range p.Entities {
		var relations []RelationViewItem
		var comb = p.Entities[idx+1:]
		for _, r_entity := range comb {
			id1, id2, swapped := canonicalPair(entity.Id, r_entity.Id)
			relation := p.GetRelationByEntities(id1, id2)
			var r_item RelationViewItem
			if relation != nil {
				r_item.Id = &relation.Id
				relValue := relation.Relation
				if swapped {
					relValue = invertRelationValue(relValue)
				}
				r_item.Relation = &relValue
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
