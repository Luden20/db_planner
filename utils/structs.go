package utils

import (
	"fmt"
	"slices"

	"github.com/xuri/excelize/v2"
)

type DbProject struct {
	Name             string
	Entities         []Entity
	EntitiesLastMax  int
	Relations        []Relation
	RelationsLastMax int
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
	_ = relationsIndex // keep Entities as active sheet

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
	return f.SaveAs(filename)
}
func (p *DbProject) AddRelation(idEnt1 int, idEnt2 int, relation string) error {
	//check if exist
	rel := p.GetRelationByEntities(idEnt1, idEnt2)
	if rel != nil {
		err := p.RemoveRelation(rel.Id)
		if err != nil {
			return err
		}
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
func (p *DbProject) AddEntity(name string, description string) error {
	p.Entities = append(p.Entities, Entity{Id: p.EntitiesLastMax + 1, Name: name, Description: description})
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
	for idx, relation := range p.Relations {
		if relation.IdEntity1 == idEntity1 && relation.IdEntity2 == idEntity2 {
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
