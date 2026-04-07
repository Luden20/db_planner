package utils

import (
	"fmt"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
)

func (p *DbProject) ExportToExcel(filename string) error {
	return p.exportToExcel(filename, true)
}

func (p *DbProject) ExportCombinationsToExcel(filename string) error {
	return p.exportToExcel(filename, false)
}

func (p *DbProject) exportToExcel(filename string, includeRelationTypes bool) error {
	f := excelize.NewFile()
	columnWidths := make(map[string]map[int]float64)
	defer func(f *excelize.File) {
		err := f.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(f)

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
		trackCellWidth(columnWidths, "Entities", col, value.Name)

		defCord, err := excelize.CoordinatesToCellName(col+1, row)
		if err != nil {
			return err
		}
		if err := f.SetCellValue("Entities", defCord, value.Description); err != nil {
			return err
		}
		trackCellWidth(columnWidths, "Entities", col+1, value.Description)
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
			trackCellWidth(columnWidths, "Relations", startCol, value.PrincipalEntity)
			if includeRelationTypes {
				startCol += 4
			} else {
				startCol += 3
			}
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
				trackCellWidth(columnWidths, "Relations", startCol, value.PrincipalEntity)
			}

			nameCell, err := excelize.CoordinatesToCellName(startCol+1, relRow)
			if err != nil {
				return err
			}
			if err := f.SetCellValue("Relations", nameCell, rel.Entity2); err != nil {
				return err
			}
			trackCellWidth(columnWidths, "Relations", startCol+1, rel.Entity2)

			if includeRelationTypes {
				relationCell, err := excelize.CoordinatesToCellName(startCol+2, relRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Relations", relationCell, relationValue); err != nil {
					return err
				}
				trackCellWidth(columnWidths, "Relations", startCol+2, relationValue)
			}
			relRow++
		}
		if includeRelationTypes {
			startCol += 4
		} else {
			startCol += 3
		}
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
			trackCellWidth(columnWidths, "Attributes", startCol, ent.Name)

			nameCell, err := excelize.CoordinatesToCellName(startCol+1, attrRow)
			if err != nil {
				return err
			}
			if err := f.SetCellValue("Attributes", nameCell, "Sin atributos definidos"); err != nil {
				return err
			}
			trackCellWidth(columnWidths, "Attributes", startCol+1, "Sin atributos definidos")
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
					trackCellWidth(columnWidths, "Attributes", startCol, ent.Name)
					named = true
				}

				nameCell, err := excelize.CoordinatesToCellName(startCol+1, attrRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Attributes", nameCell, att.Name); err != nil {
					return err
				}
				trackCellWidth(columnWidths, "Attributes", startCol+1, att.Name)

				descCell, err := excelize.CoordinatesToCellName(startCol+2, attrRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Attributes", descCell, att.Description); err != nil {
					return err
				}
				trackCellWidth(columnWidths, "Attributes", startCol+2, att.Description)

				typeCell, err := excelize.CoordinatesToCellName(startCol+3, attrRow)
				if err != nil {
					return err
				}
				if err := f.SetCellValue("Attributes", typeCell, defaultAttributeType(att.Type)); err != nil {
					return err
				}
				trackCellWidth(columnWidths, "Attributes", startCol+3, defaultAttributeType(att.Type))
				attrRow++
			}
		}
		startCol += 5
	}

	if err := applyTrackedWidths(f, columnWidths); err != nil {
		return err
	}
	err = f.DeleteSheet("Sheet1")
	if err != nil {
		return err
	}
	return f.SaveAs(filename)
}

func trackCellWidth(sheetWidths map[string]map[int]float64, sheet string, column int, value string) {
	if _, ok := sheetWidths[sheet]; !ok {
		sheetWidths[sheet] = make(map[int]float64)
	}
	width := estimatedExcelWidth(value)
	if width > sheetWidths[sheet][column] {
		sheetWidths[sheet][column] = width
	}
}

func estimatedExcelWidth(value string) float64 {
	length := utf8.RuneCountInString(value)
	if length < 12 {
		return 12
	}
	width := float64(length + 2)
	if width > 60 {
		return 60
	}
	return width
}

func applyTrackedWidths(f *excelize.File, sheetWidths map[string]map[int]float64) error {
	for sheet, widths := range sheetWidths {
		for column, width := range widths {
			colName, err := excelize.ColumnNumberToName(column)
			if err != nil {
				return err
			}
			if err := f.SetColWidth(sheet, colName, colName, width); err != nil {
				return err
			}
		}
	}
	return nil
}
