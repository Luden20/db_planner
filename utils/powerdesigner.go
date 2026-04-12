package utils

import (
	"fmt"
	"strings"
)

var pdTypeMap = map[string]string{
	// DB Planner & SQL Aliases -> PD Native Types
	"Cadena":           "Variable Characters",
	"String":           "Variable Characters",
	"Varchar":          "Variable Characters",
	"Varchar2":         "Variable Characters",
	"Nvarchar":         "Variable Characters",
	"Nvarchar2":        "Variable Characters",
	"Nchar":            "Characters",
	"Entero":           "Integer",
	"Int":              "Integer",
	"Long":             "Long Integer",
	"Bigint":           "Long Integer",
	"Short":            "Short Integer",
	"Smallint":         "Short Integer",
	"Tinyint":          "Byte",
	"Mediumint":        "Integer",
	"Decimal":          "Decimal",
	"Dec":              "Decimal",
	"Numerico":         "Number",
	"Numeric":          "Number",
	"Number":           "Number",
	"Booleano":         "Boolean",
	"Bool":             "Boolean",
	"Bit":              "Boolean",
	"Fecha":            "Date",
	"Date":             "Date",
	"Fecha y Hora":     "Date & Time",
	"DateTime":         "Date & Time",
	"Timestamp":        "Timestamp",
	"Hora":             "Time",
	"Time":             "Time",
	"Texto":            "Text",
	"Text":             "Text",
	"Clob":             "Text",
	"Nclob":            "Text",
	"Xml":              "Text",
	"Json":             "Text",
	"Jsonb":            "Text",
	"Flotante":         "Float",
	"Float":            "Float",
	"Real":             "Float",
	"Doble":            "Double",
	"Double":           "Double",
	"Double Precision": "Double",
	"Dinero":           "Money",
	"Money":            "Money",
	"Binario":          "Binary",
	"Binary":           "Binary",
	"Imagen":           "Image",
	"Image":            "Image",
	"Byte":             "Byte",
	"Caracter":         "Characters",
	"Char":             "Characters",
	"Characters":       "Characters",
	"Lote":             "Long Binary",
	"Blob":             "Long Binary",
	"Long Varchar":     "Long Varchar",
	"Variable Binary":  "Variable Binary",
	"Uniqueid":         "Variable Characters (36)",
	"Guid":             "Variable Characters (36)",
	"Uuid":             "Variable Characters (36)",
	"Serial":           "Integer",
	"Bigserial":        "Long Integer",

	// PowerDesigner Native Names (Direct entry - keeping only non-duplicates)
	"Variable Characters":            "Variable Characters",
	"Multi-line Variable Characters": "Multi-line Variable Characters",
	"Date & Time":                    "Date & Time",
	"UniqueID":                       "Variable Characters (36)",
}

// PDAttributeError represents a validation error for an attribute in PowerDesigner context
type PDAttributeError struct {
	EntityName    string `json:"entity_name"`
	AttributeName string `json:"attribute_name"`
	Type          string `json:"type"`
	Message       string `json:"message"`
}

func mapToPowerDesignerType(rawType string) string {
	rawType = strings.TrimSpace(rawType)
	if rawType == "" || rawType == "Por definir" {
		return "Variable Characters (255)"
	}

	baseType := rawType
	param := ""
	if paramIdx := strings.Index(rawType, "("); paramIdx != -1 {
		baseType = strings.TrimSpace(rawType[:paramIdx])
		param = rawType[paramIdx:]
	}

	// Case insensitive match
	var mapped string
	var exists bool
	for k, v := range pdTypeMap {
		if strings.EqualFold(k, baseType) {
			mapped = v
			exists = true
			break
		}
	}

	if !exists {
		return rawType
	}

	if param != "" {
		btLower := strings.ToLower(baseType)
		// Allow parameters for types that reasonably support them (lengths, precision, scale)
		if strings.Contains(btLower, "char") || strings.Contains(btLower, "cadena") ||
			strings.Contains(btLower, "varchar") || strings.Contains(btLower, "string") ||
			strings.Contains(btLower, "text") || strings.Contains(btLower, "decimal") ||
			strings.Contains(btLower, "dec") || strings.Contains(btLower, "numeric") ||
			strings.Contains(btLower, "number") || strings.Contains(btLower, "float") ||
			strings.Contains(btLower, "double") || strings.Contains(btLower, "binary") {
			return mapped + " " + param
		}
	} else {
		btLower := strings.ToLower(baseType)
		// Default sizes for common types without specified params
		if btLower == "cadena" || btLower == "string" || btLower == "varchar" ||
			btLower == "varchar2" || btLower == "nvarchar" || btLower == "nvarchar2" {
			return mapped + " (255)"
		}
		if btLower == "decimal" || btLower == "dec" || btLower == "numeric" {
			return mapped + " (10,2)"
		}
	}
	return mapped
}

func ValidateEntitiesForPowerDesigner(entities []Entity, intersections []IntersectionEntity) []PDAttributeError {
	var errors []PDAttributeError

	checkAttrs := func(entName string, attrs []Attribute) {
		for _, att := range attrs {
			raw := strings.TrimSpace(att.Type)
			if raw == "" || raw == "Por definir" {
				errors = append(errors, PDAttributeError{
					EntityName:    entName,
					AttributeName: att.Name,
					Type:          raw,
					Message:       "No tiene tipo definido.",
				})
				continue
			}

			base := raw
			if idx := strings.Index(raw, "("); idx != -1 {
				base = strings.TrimSpace(raw[:idx])
			}

			found := false
			for k, v := range pdTypeMap {
				if strings.EqualFold(k, base) {
					found = true
					break
				}
				if strings.EqualFold(v, base) {
					found = true
					break
				}
			}

			if !found {
				errors = append(errors, PDAttributeError{
					EntityName:    entName,
					AttributeName: att.Name,
					Type:          raw,
					Message:       fmt.Sprintf("Tipo '%s' no es válido para PowerDesigner.", raw),
				})
			}
		}
	}

	for _, ent := range entities {
		checkAttrs(ent.Name, ent.Attributes)

		hasPK := false
		for _, att := range ent.Attributes {
			if att.KeyType == AttributeKeyPK {
				hasPK = true
				break
			}
		}
		if !hasPK {
			errors = append(errors, PDAttributeError{
				EntityName:    ent.Name,
				AttributeName: "(Falta PK)",
				Type:          "-",
				Message:       "La entidad debe tener al menos una Clave Primaria (PK) definida.",
			})
		}
	}
	for _, inter := range intersections {
		checkAttrs(inter.Entity.Name, inter.Entity.Attributes)

		// Las tablas de intersección NO requieren PK explícita en el validador
		// ya que se generan automáticamente a partir de las FKs heredadas
		// en el script de PowerDesigner.
	}

	return errors
}

func GeneratePowerDesignerScript(entities []Entity, intersections []IntersectionEntity, relations []Relation) (string, error) {
	var sb strings.Builder

	sb.WriteString(`` +
		`Option Explicit

' =========================================================
' HELPERS GENERALES
' =========================================================

Sub TrySetComment(obj, txt)
    On Error Resume Next
    If txt <> "" Then
        obj.Comment = txt
        obj.Description = txt
    End If
    On Error GoTo 0
End Sub

Function AddEntity(model, code, name, commentText)
    Dim ent
    Set ent = model.Entities.CreateNew()
    ent.SetNameAndCode name, code
    Call TrySetComment(ent, commentText)
    Set AddEntity = ent
End Function

Function AddDomain(model, code, name, dataType, commentText, validValues)
    Dim dom
    Set dom = model.Domains.CreateNew()
    dom.SetNameAndCode name, code
    On Error Resume Next
    dom.DataType = dataType
    dom.DefaultDataType = dataType
    If validValues <> "" Then
        dom.Values = validValues
    End If
    On Error GoTo 0
    Call TrySetComment(dom, commentText)
    Set AddDomain = dom
End Function

Function AddAttribute(ent, code, name, dom, mandatory, dataType, commentText)
    Dim att
    Set att = ent.Attributes.CreateNew()
    att.SetNameAndCode name, code
    att.Mandatory = mandatory
    On Error Resume Next
    If Not dom Is Nothing Then
        Set att.Domain = dom
    Else
        att.DataType = dataType
    End If
    On Error GoTo 0
    Call TrySetComment(att, commentText)
    Set AddAttribute = att
End Function

Sub TryMarkAsPK(att)
    On Error Resume Next
    att.PrimaryIdentifier = True
    att.Primary = True
    att.Identifier = True
    att.Mandatory = True
    On Error GoTo 0
End Sub

Sub TryMarkAsFK(att)
    On Error Resume Next
    att.ForeignKey = True
    att.Migrated = True
    On Error GoTo 0
End Sub

Function AddRelationship(model, obj1, obj2, code, name, commentText)
    Dim rel
    Set rel = model.Relationships.CreateNew()
    rel.SetNameAndCode name, code
    Set rel.Object1 = obj1
    Set rel.Object2 = obj2
    Call TrySetComment(rel, commentText)
    Set AddRelationship = rel
End Function

Sub TrySetCardinality(rel, card1, card2)
    On Error Resume Next
    rel.Cardinality1 = card1
    rel.Cardinality2 = card2
    rel.ParentCardinality = card1
    rel.ChildCardinality = card2
    On Error GoTo 0
End Sub

Sub TrySetIdentifying(rel, value)
    On Error Resume Next
    rel.Identifying = value
    rel.IdentifyingRelationship = value
    On Error GoTo 0
End Sub

Sub TrySetMandatory(rel, m1, m2)
    On Error Resume Next
    ' Propiedades para CDM
    rel.Entity1ToEntity2Mandatory = m1
    rel.Entity2ToEntity1Mandatory = m2
    rel.Entity1ToEntity2RoleMandatory = m1
    rel.Entity2ToEntity1RoleMandatory = m2
    ' Propiedades para LDM (pueden variar según la versión de PD)
    rel.Role1Mandatory = m1
    rel.Role2Mandatory = m2
    rel.ParentRoleMandatory = m1
    rel.ChildRoleMandatory = m2
    On Error GoTo 0
End Sub

Sub TryPromoteIdentifier(ent, code, name)
    On Error Resume Next
    Dim id
    Dim ia
    Dim hasAttributes

	    For Each id In ent.Identifiers
	        hasAttributes = False
	        For Each ia In id.IdentifierAttributes
	            hasAttributes = True
	            Exit For
        Next

	        If hasAttributes Then
	            id.SetNameAndCode name, code
	            id.Primary = True
	            id.PrimaryIdentifier = True
	            Set ent.PrimaryIdentifier = id
	            Exit For
	        End If
	    Next

	    On Error GoTo 0
End Sub

Function IsIdentifierForeignKeyAttribute(att)
    On Error Resume Next
    Dim isForeignKey
    isForeignKey = False

    Err.Clear
    If att.ForeignIdentifier = True Then
        isForeignKey = True
    End If
    If Err.Number <> 0 Then
        Err.Clear
    End If

    If Not isForeignKey Then
        If att.ForeignKey = True Then
            isForeignKey = True
        End If
        If Err.Number <> 0 Then
            Err.Clear
        End If
    End If

    IsIdentifierForeignKeyAttribute = isForeignKey
    On Error GoTo 0
End Function

Function GetOrCreatePrimaryIdentifier(ent, code, name)
    On Error Resume Next
    Dim id
    Dim hasAttributes
    Dim candidate
    Dim att

    Set candidate = Nothing

    For Each id In ent.Identifiers
        hasAttributes = False
        For Each att In id.Attributes
            hasAttributes = True
            Exit For
        Next

        If candidate Is Nothing Then
            Set candidate = id
        End If

        If hasAttributes Then
            Set candidate = id
            Exit For
        End If
    Next

    If candidate Is Nothing Then
        Set candidate = ent.Identifiers.CreateNew()
    End If

    If Not candidate Is Nothing Then
        candidate.SetNameAndCode name, code
        candidate.Primary = True
        candidate.PrimaryIdentifier = True
        Set ent.PrimaryIdentifier = candidate
    End If
    If Err.Number <> 0 Then
        Err.Clear
    End If

    Set GetOrCreatePrimaryIdentifier = candidate
    On Error GoTo 0
End Function

Function IdentifierContainsAttribute(id, att)
    On Error Resume Next
    Dim existingAtt
    IdentifierContainsAttribute = False

    If id Is Nothing Or att Is Nothing Then
        On Error GoTo 0
        Exit Function
    End If

    For Each existingAtt In id.Attributes
        If existingAtt Is att Then
            IdentifierContainsAttribute = True
            Exit Function
        End If
    Next

    On Error GoTo 0
End Function

Function TryAddAttributeToIdentifier(id, att)
    On Error Resume Next
    TryAddAttributeToIdentifier = False

    If id Is Nothing Or att Is Nothing Then
        On Error GoTo 0
        Exit Function
    End If

    If Not IsIdentifierForeignKeyAttribute(att) Then
        On Error GoTo 0
        Exit Function
    End If

    If IdentifierContainsAttribute(id, att) Then
        On Error GoTo 0
        Exit Function
    End If

    id.Attributes.Insert -1, att
    att.PrimaryIdentifier = True
    att.Primary = True
    att.Identifier = True
    att.Mandatory = True
    TryAddAttributeToIdentifier = True

    On Error GoTo 0
End Function

Function TryAddRelationshipJoinChildAttributesToIdentifier(rel, id)
    On Error Resume Next
    Dim joinObj
    Dim childAtt

    TryAddRelationshipJoinChildAttributesToIdentifier = 0

    If rel Is Nothing Or id Is Nothing Then
        On Error GoTo 0
        Exit Function
    End If

    For Each joinObj In rel.Joins
        Set childAtt = Nothing

        Err.Clear
        Set childAtt = joinObj.ChildAttribute
        If Err.Number <> 0 Then
            Err.Clear
            Set childAtt = Nothing
        End If

        If TryAddAttributeToIdentifier(id, childAtt) Then
            TryAddRelationshipJoinChildAttributesToIdentifier = TryAddRelationshipJoinChildAttributesToIdentifier + 1
        End If
    Next

    On Error GoTo 0
End Function

Sub AttachEntity(diag, obj, x, y)
    On Error Resume Next
    Dim sym
    Set sym = diag.AttachObject(obj)
    If Not sym Is Nothing Then
        sym.RectX = x
        sym.RectY = y
    End If
    On Error GoTo 0
End Sub

Sub AttachRelationship(diag, rel)
    On Error Resume Next
    Dim sym
    Set sym = diag.AttachLinkObject(rel)
    On Error GoTo 0
End Sub

`)

	generateModel := func(isCDM bool) {
		modelType := "PdLDM.cls_Model"
		modelName := "Modelo_Logico"
		modelCode := "LDM_DBP"
		if isCDM {
			modelType = "PdCDM.cls_Model"
			modelName = "Modelo_Conceptual"
			modelCode = "CDM_DBP"
			sb.WriteString(fmt.Sprintf("' =========================================================\n' CREANDO MODELO CONCEPTUAL\n' =========================================================\n"))
			sb.WriteString(fmt.Sprintf("Dim mdlC\nSet mdlC = CreateModel(%s)\nIf mdlC Is Nothing Then WScript.Quit\nmdlC.SetNameAndCode \"%s\", \"%s\"\n\n", modelType, modelName, modelCode))
		} else {
			sb.WriteString(fmt.Sprintf("' =========================================================\n' CREANDO MODELO LOGICO\n' =========================================================\n"))
			sb.WriteString(fmt.Sprintf("Dim mdlL\nSet mdlL = CreateModel(%s)\nIf mdlL Is Nothing Then WScript.Quit\nmdlL.SetNameAndCode \"%s\", \"%s\"\n\n", modelType, modelName, modelCode))
		}

		mdlVar := "mdlL"
		if isCDM {
			mdlVar = "mdlC"
		}
		prefix := "l"
		if isCDM {
			prefix = "c"
		}

		domainVars := make(map[string]string)
		for _, ent := range entities {
			for _, att := range ent.Attributes {
				if len(att.Domain) > 0 {
					attKey := fmt.Sprintf("%s_%s", ent.Name, att.Name)
					domVarName := "dSpec_" + prefix + "_" + safeVBSName(attKey)
					domCode := "DOM_SPEC_" + prefix + "_" + strings.ToUpper(safeVBSName(attKey))
					domName := "Dominio " + att.Name
					domValuesInfo := "Valores permitidos: " + strings.Join(att.Domain, ", ")
					t := mapToPowerDesignerType(att.Type)
					sb.WriteString(fmt.Sprintf("Dim %s\n", domVarName))
					domValuesJoined := strings.Join(att.Domain, ";")
					sb.WriteString(fmt.Sprintf("Set %s = AddDomain(%s, \"%s\", \"%s\", \"%s\", \"%s\", \"%s\")\n\n", domVarName, mdlVar, domCode, domName, t, domValuesInfo, domValuesJoined))
					domainVars[attKey] = domVarName
				}
			}
		}
		for _, inter := range intersections {
			for _, att := range inter.Entity.Attributes {
				if len(att.Domain) > 0 {
					attKey := fmt.Sprintf("%s_%s", inter.Entity.Name, att.Name)
					domVarName := "dSpec_" + prefix + "_" + safeVBSName(attKey)
					domCode := "DOM_SPEC_" + prefix + "_" + strings.ToUpper(safeVBSName(attKey))
					domName := "Dominio " + att.Name
					domValuesInfo := "Valores permitidos: " + strings.Join(att.Domain, ", ")
					t := mapToPowerDesignerType(att.Type)
					sb.WriteString(fmt.Sprintf("Dim %s\n", domVarName))
					domValuesJoined := strings.Join(att.Domain, ";")
					sb.WriteString(fmt.Sprintf("Set %s = AddDomain(%s, \"%s\", \"%s\", \"%s\", \"%s\", \"%s\")\n\n", domVarName, mdlVar, domCode, domName, t, domValuesInfo, domValuesJoined))
					domainVars[attKey] = domVarName
				}
			}
		}

		entityVarNames := make(map[int]string)
		for _, ent := range entities {
			varName := "e_" + prefix + "_" + safeVBSName(ent.Name)
			entityVarNames[ent.Id] = varName
			desc := escapeVBSString(ent.Description)
			sb.WriteString(fmt.Sprintf("Dim %s\n", varName))
			sb.WriteString(fmt.Sprintf("Set %s = AddEntity(%s, \"%s\", \"%s\", \"%s\")\n", varName, mdlVar, strings.ToUpper(safeVBSName(ent.Name)), escapeVBSString(ent.Name), desc))
		}
		intersectionVarNames := make(map[int]string)
		intersectionMatchVarNames := make(map[int]string)
		intersectionIdentifierVarNames := make(map[int]string)
		for _, inter := range intersections {
			varName := "i_" + prefix + "_" + safeVBSName(inter.Entity.Name)
			intersectionVarNames[inter.RelationID] = varName
			matchVarName := "matched_" + prefix + "_" + safeVBSName(inter.Entity.Name)
			intersectionMatchVarNames[inter.RelationID] = matchVarName
			identifierVarName := "id_" + prefix + "_" + safeVBSName(inter.Entity.Name)
			intersectionIdentifierVarNames[inter.RelationID] = identifierVarName
			desc := escapeVBSString(inter.Entity.Description)
			sb.WriteString(fmt.Sprintf("Dim %s\n", varName))
			sb.WriteString(fmt.Sprintf("Set %s = AddEntity(%s, \"%s\", \"%s\", \"%s\")\n", varName, mdlVar, strings.ToUpper(safeVBSName(inter.Entity.Name)), escapeVBSString(inter.Entity.Name), desc))
			if !isCDM {
				sb.WriteString(fmt.Sprintf("Dim %s\n%s = 0\n", matchVarName, matchVarName))
				sb.WriteString(fmt.Sprintf("Dim %s\nSet %s = GetOrCreatePrimaryIdentifier(%s, \"ID_PRIMARY\", \"Identificador Primario\")\n", identifierVarName, identifierVarName, varName))
			}
		}
		sb.WriteString("\n")

		writeAttributes := func(entName string, entVar string, attributes []Attribute) []string {
			pkVars := []string{}
			for _, att := range attributes {
				isPK := normalizeAttributeKeyType(att.KeyType) == AttributeKeyPK
				isFK := normalizeAttributeKeyType(att.KeyType) == AttributeKeyFK
				if isCDM && !isPK && !isFK {
					continue
				}

				attVar := "a_" + prefix + "_" + safeVBSName(entName+"_"+att.Name)
				attCode := strings.ToUpper(safeVBSName(att.Name))
				attNameSafe := escapeVBSString(att.Name)
				mandatory := "True"
				if att.Optional {
					mandatory = "False"
				}

				t := mapToPowerDesignerType(att.Type)

				domOrTypeVar := "Nothing"
				attKey := fmt.Sprintf("%s_%s", entName, att.Name)
				if specDom, ok := domainVars[attKey]; ok {
					domOrTypeVar = specDom
				}

				desc := escapeVBSString(att.Description)
				if len(att.Domain) > 0 {
					desc = desc + " | Valores: " + strings.Join(att.Domain, ", ")
				}

				sb.WriteString(fmt.Sprintf("Dim %s\n", attVar))
				sb.WriteString(fmt.Sprintf("Set %s = AddAttribute(%s, \"%s\", \"%s\", %s, %s, \"%s\", \"%s\")\n",
					attVar, entVar, attCode, attNameSafe, domOrTypeVar, mandatory, t, desc))

				if isPK {
					sb.WriteString(fmt.Sprintf("Call TryMarkAsPK(%s)\n", attVar))
					pkVars = append(pkVars, attVar)
				} else if isFK {
					sb.WriteString(fmt.Sprintf("Call TryMarkAsFK(%s)\n", attVar))
				}
			}
			return pkVars
		}

		for _, ent := range entities {
			pks := writeAttributes(ent.Name, entityVarNames[ent.Id], ent.Attributes)
			if len(pks) > 0 {
				sb.WriteString(fmt.Sprintf("Call TryPromoteIdentifier(%s, \"ID_PRIMARY\", \"Identificador Primario\")\n",
					entityVarNames[ent.Id]))
			}
		}

		for _, inter := range intersections {
			interVar := intersectionVarNames[inter.RelationID]
			if !isCDM {
				writeAttributes(inter.Entity.Name, interVar, inter.Entity.Attributes)
			}
		}

		for _, rel := range relations {
			isIntersection := false
			for _, inter := range intersections {
				if inter.RelationID == rel.Id {
					isIntersection = true
					break
				}
			}
			if isIntersection {
				continue
			}

			var e1, e2 *Entity
			for _, e := range entities {
				if e.Id == rel.IdEntity1 {
					e1 = &e
				}
				if e.Id == rel.IdEntity2 {
					e2 = &e
				}
			}
			if e1 != nil && e2 != nil {
				var childEnt, parentEnt *Entity
				childVar := entityVarNames[e2.Id]
				parentEnt = e1
				childEnt = e2

				if rel.Relation == "N:1" || rel.Relation == "Np:1" {
					parentEnt = e2
					childEnt = e1
					childVar = entityVarNames[e1.Id]
				}

				for _, pAtt := range parentEnt.Attributes {
					if normalizeAttributeKeyType(pAtt.KeyType) == AttributeKeyPK {
						attVar := "a_fk_" + prefix + "_" + safeVBSName(childEnt.Name+"_"+parentEnt.Name+"_"+pAtt.Name)
						attCode := "ID_" + strings.ToUpper(safeVBSName(parentEnt.Name))
						attNameSafe := escapeVBSString("Id " + parentEnt.Name)
						t := mapToPowerDesignerType(pAtt.Type)
						sb.WriteString(fmt.Sprintf("Dim %s\n", attVar))
						sb.WriteString(fmt.Sprintf("Set %s = AddAttribute(%s, \"%s\", \"%s\", Nothing, True, \"%s\", \"Clave foranea obligatoria heredada de %s.\")\n",
							attVar, childVar, attCode, attNameSafe, t, parentEnt.Name))
						sb.WriteString(fmt.Sprintf("Call TryMarkAsFK(%s)\n", attVar))
					}
				}
			}
		}
		sb.WriteString("\n")

		relCounter := 1
		for _, rel := range relations {
			var e1, e2 *Entity
			for _, e := range entities {
				if e.Id == rel.IdEntity1 {
					e1 = &e
				}
				if e.Id == rel.IdEntity2 {
					e2 = &e
				}
			}

			if e1 != nil && e2 != nil {
				isIntersection := false
				for _, inter := range intersections {
					if inter.RelationID == rel.Id {
						isIntersection = true
						break
					}
				}

				if isIntersection {
					interVar := intersectionVarNames[rel.Id]
					matchVar := intersectionMatchVarNames[rel.Id]
					identifierVar := intersectionIdentifierVarNames[rel.Id]

					relName1 := fmt.Sprintf("R_%s_%s", strings.ToUpper(safeVBSName(e1.Name)), strings.ToUpper(safeVBSName("INTER")))
					rVar1 := fmt.Sprintf("rel_%s_%d_1", prefix, relCounter)
					sb.WriteString(fmt.Sprintf("Dim %s\n", rVar1))
					sb.WriteString(fmt.Sprintf("Set %s = AddRelationship(%s, %s, %s, \"%s\", \"%s asociado a interseccion\", \"Relacion 1:N hacia la interseccion.\")\n",
						rVar1, mdlVar, entityVarNames[e1.Id], interVar, relName1, e1.Name))
					sb.WriteString(fmt.Sprintf("Call TrySetCardinality(%s, \"1,1\", \"1,n\")\n", rVar1))
					sb.WriteString(fmt.Sprintf("Call TrySetMandatory(%s, True, True)\n", rVar1))
					sb.WriteString(fmt.Sprintf("Call TrySetIdentifying(%s, True)\n\n", rVar1))
					if !isCDM {
						sb.WriteString(fmt.Sprintf("%s = %s + TryAddRelationshipJoinChildAttributesToIdentifier(%s, %s)\n\n", matchVar, matchVar, rVar1, identifierVar))
					}

					relName2 := fmt.Sprintf("R_%s_%s", strings.ToUpper(safeVBSName(e2.Name)), strings.ToUpper(safeVBSName("INTER")))
					rVar2 := fmt.Sprintf("rel_%s_%d_2", prefix, relCounter)
					sb.WriteString(fmt.Sprintf("Dim %s\n", rVar2))
					sb.WriteString(fmt.Sprintf("Set %s = AddRelationship(%s, %s, %s, \"%s\", \"%s asociado a interseccion\", \"Relacion 1:N hacia la interseccion.\")\n",
						rVar2, mdlVar, entityVarNames[e2.Id], interVar, relName2, e2.Name))
					sb.WriteString(fmt.Sprintf("Call TrySetCardinality(%s, \"1,1\", \"1,n\")\n", rVar2))
					sb.WriteString(fmt.Sprintf("Call TrySetMandatory(%s, True, True)\n", rVar2))
					sb.WriteString(fmt.Sprintf("Call TrySetIdentifying(%s, True)\n\n", rVar2))
					if !isCDM {
						sb.WriteString(fmt.Sprintf("%s = %s + TryAddRelationshipJoinChildAttributesToIdentifier(%s, %s)\n\n", matchVar, matchVar, rVar2, identifierVar))
					}
				} else {
					relName := fmt.Sprintf("R_%s_%s", strings.ToUpper(safeVBSName(e1.Name)), strings.ToUpper(safeVBSName(e2.Name)))
					rVar := fmt.Sprintf("rel_%s_%d", prefix, relCounter)

					card1 := "0,n"
					card2 := "1,1"
					m1 := "False"
					m2 := "True"

					switch rel.Relation {
					case RelationType11:
						card1, card2 = "1,1", "0,1" // child->parent(e1) = 1,1. parent(e1)->child(e2)=0,1
						m1, m2 = "False", "True"
					case RelationType1N:
						card1, card2 = "1,1", "1,n" // child>parent=1,1. parent>child=1,n.
						m1, m2 = "True", "True"
					case RelationTypeN1:
						card1, card2 = "1,n", "1,1" // child(e1)>parent(e2)=1,n (wait, N children to 1 parent). Card2=1,1
						m1, m2 = "True", "True"
					case RelationType1Np:
						card1, card2 = "1,1", "0,n" // Child must have parent, parent can have 0..n
						m1, m2 = "False", "True"
					case RelationTypeNp1:
						card1, card2 = "0,n", "1,1" // child(e1)>parent(e2)=0,n
						m1, m2 = "True", "False"
					default:
						card1, card2 = "1,1", "0,n"
						m1, m2 = "False", "True"
					}

					sb.WriteString(fmt.Sprintf("Dim %s\n", rVar))
					sb.WriteString(fmt.Sprintf("Set %s = AddRelationship(%s, %s, %s, \"%s\", \"Relacion %s y %s\", \"Relacion tipo %s.\")\n",
						rVar, mdlVar, entityVarNames[e1.Id], entityVarNames[e2.Id], relName, e1.Name, e2.Name, rel.Relation))
					sb.WriteString(fmt.Sprintf("Call TrySetCardinality(%s, \"%s\", \"%s\")\n", rVar, card1, card2))
					sb.WriteString(fmt.Sprintf("Call TrySetMandatory(%s, %s, %s)\n", rVar, m1, m2))
					sb.WriteString(fmt.Sprintf("Call TrySetIdentifying(%s, False)\n\n", rVar))
				}
				relCounter++
			}
		}

		if !isCDM {
			for _, inter := range intersections {
				interVar := intersectionVarNames[inter.RelationID]
				matchVar := intersectionMatchVarNames[inter.RelationID]
				identifierVar := intersectionIdentifierVarNames[inter.RelationID]
				sb.WriteString(fmt.Sprintf("If %s > 0 Then Set %s.PrimaryIdentifier = %s\n", matchVar, interVar, identifierVar))
				sb.WriteString(fmt.Sprintf("If %s > 0 Then Call TryPromoteIdentifier(%s, \"ID_PRIMARY\", \"Identificador Primario\")\n", matchVar, interVar))
			}
			sb.WriteString("\n")
		}

		diagVar := "diag_" + prefix
		sb.WriteString(fmt.Sprintf("Dim %s\nSet %s = ActiveDiagram\nIf %s Is Nothing Then\n    On Error Resume Next\n    Set %s = %s.DefaultDiagram\n    On Error GoTo 0\nEnd If\n\nIf Not %s Is Nothing Then\n", diagVar, diagVar, diagVar, diagVar, mdlVar, diagVar))

		x, y := 100, 100
		for _, ent := range entities {
			sb.WriteString(fmt.Sprintf("    Call AttachEntity(%s, %s, %d, %d)\n", diagVar, entityVarNames[ent.Id], x, y))
			x += 400
			if x > 1500 {
				x, y = 100, y+300
			}
		}
		for _, inter := range intersections {
			sb.WriteString(fmt.Sprintf("    Call AttachEntity(%s, %s, %d, %d)\n", diagVar, intersectionVarNames[inter.RelationID], x, y))
			x += 400
			if x > 1500 {
				x, y = 100, y+300
			}
		}

		for i := 1; i < relCounter; i++ {
			sb.WriteString(fmt.Sprintf("    On Error Resume Next\n    If Not IsEmpty(rel_%s_%d) Then Call AttachRelationship(%s, rel_%s_%d)\n    If Not IsEmpty(rel_%s_%d_1) Then Call AttachRelationship(%s, rel_%s_%d_1)\n    If Not IsEmpty(rel_%s_%d_2) Then Call AttachRelationship(%s, rel_%s_%d_2)\n    On Error GoTo 0\n",
				prefix, i, diagVar, prefix, i,
				prefix, i, diagVar, prefix, i,
				prefix, i, diagVar, prefix, i))
		}
		sb.WriteString(fmt.Sprintf("\n    On Error Resume Next\n    %s.AutoLayoutWithOptions 2\n    On Error GoTo 0\nEnd If\n\n", diagVar))
	}

	generateModel(true)
	generateModel(false)

	sb.WriteString("\nMsgBox \"Modelos Conceptual y Lógico creados exitosamente.\"\n")

	return sb.String(), nil
}

func safeVBSName(s string) string {
	var sb strings.Builder
	for _, r := range s {
		char := r
		switch {
		case 'a' <= char && char <= 'z':
			sb.WriteRune(char)
		case 'A' <= char && char <= 'Z':
			sb.WriteRune(char)
		case '0' <= char && char <= '9':
			sb.WriteRune(char)
		default:
			sb.WriteRune('_')
		}
	}
	return sb.String()
}

func escapeVBSString(s string) string {
	s = strings.ReplaceAll(s, "\"", "\"\"")
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
