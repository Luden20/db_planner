'******************************************************************************
'* Importa JSON de db_planner y genera un CDM en PowerDesigner.
'* Soporta proyecto.json completo y JSON exportado.
'******************************************************************************
Option Explicit

Const ForReading = 1
Const PdCDM_Model = 509178224
Const ExtractorScriptPath = "C:\DevStuff\SideUnrelatedProjects\db_planner\powerdesigner\dbplanner_extract.ps1"

Dim gJsonText
Dim gJsonPos

Main

Sub Main()
    Dim jsonPath, root, model, diagram, projectName
    Dim entityIndex, entitySymbolIndex
    Dim intersections, intersectionByRelation, entities, relations
    Dim item

    jsonPath = Trim(InputBox("Ruta completa del JSON exportado por db_planner:", "Importar db_planner a CDM"))
    jsonPath = Trim(Replace(jsonPath, """", ""))
    If jsonPath = "" Then
        Output "Importacion cancelada."
        Exit Sub
    End If

    Set root = LoadDbPlannerData(jsonPath)
    projectName = CStr(GetValue(root, Array("name", "Name"), "db_planner"))

    Set model = CreateModel(PdCDM_Model, "|Diagram=ConceptualDiagram")
    model.Name = projectName & " CDM"
    model.Code = BuildCode(model.Name, "DB_PLANNER_CDM")

    Set diagram = model.DefaultDiagram
    If diagram Is Nothing Then
        Set diagram = ActiveDiagram
    End If

    Set entities = GetCollectionValue(root, Array("entities", "Entities"))
    Set relations = GetCollectionValue(root, Array("relations", "Relations"))
    Set intersections = NormalizeIntersectionCollection(root)
    Set intersectionByRelation = BuildIntersectionMap(intersections)

    Output "JSON cargado. Entidades=" & CStr(ListCount(entities)) & ", intersecciones=" & CStr(ListCount(intersections)) & ", relaciones=" & CStr(ListCount(relations))
    MsgBox "JSON cargado." & vbCrLf & "Entidades=" & CStr(ListCount(entities)) & vbCrLf & "Intersecciones=" & CStr(ListCount(intersections)) & vbCrLf & "Relaciones=" & CStr(ListCount(relations))

    Set entityIndex = CreateObject("Scripting.Dictionary")
    Set entitySymbolIndex = CreateObject("Scripting.Dictionary")

    For Each item In ListValues(entities)
        CreateCDMEntity model, diagram, item, entityIndex, entitySymbolIndex
    Next

    For Each item In ListValues(intersections)
        CreateCDMEntity model, diagram, GetObject(item, Array("entity", "Entity")), entityIndex, entitySymbolIndex
    Next

    For Each item In ListValues(relations)
        CreateCDMRelation model, diagram, item, entityIndex, entitySymbolIndex, intersectionByRelation
    Next

    On Error Resume Next
    diagram.LayoutDiagram
    Err.Clear
    On Error GoTo 0

    Output "CDM generado: " & model.Name
    MsgBox "CDM generado: " & model.Name
End Sub

Sub CreateCDMEntity(model, diagram, entityData, entityIndex, entitySymbolIndex)
    Dim entityId, entity, attrs, attr, symbol

    If entityData Is Nothing Then Exit Sub

    entityId = CStr(GetValue(entityData, Array("id", "Id"), ""))
    If entityId = "" Then Exit Sub
    If entityIndex.Exists(entityId) Then Exit Sub

    Set entity = model.Entities.CreateNew
    entity.Name = CStr(GetValue(entityData, Array("name", "Name"), "Entidad"))
    entity.Code = BuildCode(GetValue(entityData, Array("name", "Name"), ""), "ENTITY_" & entityId)
    SafeSetText entity, "Comment", GetValue(entityData, Array("description", "Description"), "")
    SafeSetText entity, "Description", GetValue(entityData, Array("description", "Description"), "")

    Set attrs = GetCollectionValue(entityData, Array("attributes", "Attributes"))
    For Each attr In ListValues(attrs)
        CreateCDMAttribute entity, attr
    Next

    Set entityIndex(entityId) = entity
    If Not diagram Is Nothing Then
        Set symbol = diagram.AttachObject(entity)
        If Not symbol Is Nothing Then
            Set entitySymbolIndex(entityId) = symbol
        End If
    End If
End Sub

Sub CreateCDMAttribute(entity, attrData)
    Dim attr, descriptionText

    If attrData Is Nothing Then Exit Sub

    Set attr = entity.Attributes.CreateNew
    attr.Name = CStr(GetValue(attrData, Array("name", "Name"), "Atributo"))
    attr.Code = BuildCode(GetValue(attrData, Array("name", "Name"), ""), "ATTR_" & CStr(GetValue(attrData, Array("id", "Id"), "0")))
    SafeSetText attr, "DataType", GetValue(attrData, Array("type", "Type"), "")
    SafeSetBoolean attr, "Mandatory", Not CBool(GetValue(attrData, Array("optional", "Optional"), False))
    If LCase(CStr(GetValue(attrData, Array("key_type", "KeyType"), "nil"))) = "pk" Then
        SafeSetBoolean attr, "PrimaryIdentifier", True
    End If

    descriptionText = JoinDescriptionWithDomain( _
        CStr(GetValue(attrData, Array("description", "Description"), "")), _
        GetCollectionValue(attrData, Array("domain", "Domain")) _
    )
    SafeSetText attr, "Comment", descriptionText
    SafeSetText attr, "Description", descriptionText
End Sub

Sub CreateCDMRelation(model, diagram, relationData, entityIndex, entitySymbolIndex, intersectionByRelation)
    Dim relationType, relationId, entity1Id, entity2Id
    Dim relationObject, intersectionData, intersectionEntity, intersectionId

    relationType = UCase(CStr(GetValue(relationData, Array("relation", "Relation"), "")))
    relationId = CStr(GetValue(relationData, Array("id", "Id"), ""))
    entity1Id = CStr(GetValue(relationData, Array("id_entity1", "IdEntity1"), ""))
    entity2Id = CStr(GetValue(relationData, Array("id_entity2", "IdEntity2"), ""))

    If entity1Id = "" Or entity2Id = "" Then Exit Sub
    If (Not entityIndex.Exists(entity1Id)) Or (Not entityIndex.Exists(entity2Id)) Then Exit Sub

    If relationType = "N:N" And intersectionByRelation.Exists(relationId) Then
        Set intersectionData = intersectionByRelation(relationId)
        Set intersectionEntity = GetObject(intersectionData, Array("entity", "Entity"))
        If Not intersectionEntity Is Nothing Then
            intersectionId = CStr(GetValue(intersectionEntity, Array("id", "Id"), ""))
            If intersectionId <> "" And entityIndex.Exists(intersectionId) Then
                CreateDependentCDMRelation model, diagram, entityIndex(entity1Id), entityIndex(intersectionId), entitySymbolIndex, entity1Id, intersectionId, "A", "REL_" & relationId & "_1"
                CreateDependentCDMRelation model, diagram, entityIndex(entity2Id), entityIndex(intersectionId), entitySymbolIndex, entity2Id, intersectionId, "A", "REL_" & relationId & "_2"
                Exit Sub
            End If
        End If
    End If

    Set relationObject = model.Relationships.CreateNew
    relationObject.Name = "Rel_" & relationId
    relationObject.Code = BuildCode(relationObject.Name, "REL_" & relationId)
    Set relationObject.Entity1 = entityIndex(entity1Id)
    Set relationObject.Entity2 = entityIndex(entity2Id)
    ApplyRelationSemantics relationObject, relationType
    AttachRelationSymbol diagram, relationObject, entitySymbolIndex, entity1Id, entity2Id
End Sub

Sub CreateDependentCDMRelation(model, diagram, parentEntity, childEntity, entitySymbolIndex, parentId, childId, dependentRole, relationName)
    Dim relationObject

    Set relationObject = model.Relationships.CreateNew
    relationObject.Name = relationName
    relationObject.Code = relationName
    Set relationObject.Entity1 = parentEntity
    Set relationObject.Entity2 = childEntity

    SafeSetText relationObject, "Entity2ToEntity1RoleMaximumCardinality", "1"
    SafeSetText relationObject, "Entity2ToEntity1RoleMinimumCardinality", "1"
    SafeSetText relationObject, "Entity1ToEntity2RoleMaximumCardinality", "n"
    SafeSetText relationObject, "Entity1ToEntity2RoleMinimumCardinality", "1"
    SafeSetBoolean relationObject, "Entity1ToEntity2RoleMandatory", True
    SafeSetText relationObject, "DependentRole", dependentRole

    AttachRelationSymbol diagram, relationObject, entitySymbolIndex, parentId, childId
End Sub

Sub ApplyRelationSemantics(relationObject, relationType)
    Select Case relationType
        Case "1:1"
            SafeSetText relationObject, "Entity2ToEntity1RoleMaximumCardinality", "1"
            SafeSetText relationObject, "Entity2ToEntity1RoleMinimumCardinality", "1"
            SafeSetText relationObject, "Entity1ToEntity2RoleMaximumCardinality", "1"
            SafeSetText relationObject, "Entity1ToEntity2RoleMinimumCardinality", "1"
        Case "1:N"
            SafeSetText relationObject, "Entity2ToEntity1RoleMaximumCardinality", "1"
            SafeSetText relationObject, "Entity2ToEntity1RoleMinimumCardinality", "1"
            SafeSetText relationObject, "Entity1ToEntity2RoleMaximumCardinality", "n"
            SafeSetText relationObject, "Entity1ToEntity2RoleMinimumCardinality", "0"
        Case "N:1"
            SafeSetText relationObject, "Entity2ToEntity1RoleMaximumCardinality", "n"
            SafeSetText relationObject, "Entity2ToEntity1RoleMinimumCardinality", "0"
            SafeSetText relationObject, "Entity1ToEntity2RoleMaximumCardinality", "1"
            SafeSetText relationObject, "Entity1ToEntity2RoleMinimumCardinality", "1"
        Case "N:N"
            SafeSetText relationObject, "Entity2ToEntity1RoleMaximumCardinality", "n"
            SafeSetText relationObject, "Entity2ToEntity1RoleMinimumCardinality", "0"
            SafeSetText relationObject, "Entity1ToEntity2RoleMaximumCardinality", "n"
            SafeSetText relationObject, "Entity1ToEntity2RoleMinimumCardinality", "0"
        Case "1:NP"
            SafeSetText relationObject, "Entity2ToEntity1RoleMaximumCardinality", "1"
            SafeSetText relationObject, "Entity2ToEntity1RoleMinimumCardinality", "1"
            SafeSetText relationObject, "Entity1ToEntity2RoleMaximumCardinality", "n"
            SafeSetText relationObject, "Entity1ToEntity2RoleMinimumCardinality", "1"
            SafeSetBoolean relationObject, "Entity1ToEntity2RoleMandatory", True
            SafeSetText relationObject, "DependentRole", "A"
        Case "NP:1"
            SafeSetText relationObject, "Entity2ToEntity1RoleMaximumCardinality", "n"
            SafeSetText relationObject, "Entity2ToEntity1RoleMinimumCardinality", "1"
            SafeSetText relationObject, "Entity1ToEntity2RoleMaximumCardinality", "1"
            SafeSetText relationObject, "Entity1ToEntity2RoleMinimumCardinality", "1"
            SafeSetBoolean relationObject, "Entity2ToEntity1RoleMandatory", True
            SafeSetText relationObject, "DependentRole", "B"
    End Select
End Sub

Sub AttachRelationSymbol(diagram, relationObject, entitySymbolIndex, entity1Id, entity2Id)
    Dim sym1, sym2

    If diagram Is Nothing Then Exit Sub
    If (Not entitySymbolIndex.Exists(entity1Id)) Or (Not entitySymbolIndex.Exists(entity2Id)) Then Exit Sub

    Set sym1 = entitySymbolIndex(entity1Id)
    Set sym2 = entitySymbolIndex(entity2Id)

    On Error Resume Next
    diagram.AttachLinkObject relationObject, sym1, sym2
    If Err.Number <> 0 Then
        Err.Clear
        diagram.AttachLinkObject relationObject
    End If
    On Error GoTo 0
End Sub

Function NormalizeIntersectionCollection(root)
    Dim rawItems, normalized, item, wrapper

    Set rawItems = GetCollectionValue(root, Array("intersection_entities", "IntersectionEntities"))
    Set normalized = CreateList()

    For Each item In ListValues(rawItems)
        If TypeName(item) = "Dictionary" Then
            If HasAnyKey(item, Array("entity", "Entity")) Then
                ListAdd normalized, item
            Else
                Set wrapper = CreateObject("Scripting.Dictionary")
                wrapper.Add "RelationID", GetValue(item, Array("relation_id", "RelationID"), 0)
                DictSetObject wrapper, "Entity", item
                ListAdd normalized, wrapper
            End If
        End If
    Next

    Set NormalizeIntersectionCollection = normalized
End Function

Function BuildIntersectionMap(intersections)
    Dim dict, item, relationId

    Set dict = CreateObject("Scripting.Dictionary")
    For Each item In ListValues(intersections)
        relationId = CStr(GetValue(item, Array("relation_id", "RelationID"), ""))
        If relationId <> "" Then
            Set dict(relationId) = item
        End If
    Next
    Set BuildIntersectionMap = dict
End Function

Function JoinDescriptionWithDomain(baseDescription, domainItems)
    Dim textValue, piece

    textValue = Trim(CStr(baseDescription))
    piece = JoinCollection(domainItems, ", ")
    If piece <> "" Then
        If textValue <> "" Then
            textValue = textValue & vbCrLf & "Dominio: " & piece
        Else
            textValue = "Dominio: " & piece
        End If
    End If
    JoinDescriptionWithDomain = textValue
End Function

Function BuildCode(rawValue, fallbackValue)
    Dim inputText, resultText, i, ch, prevUnderscore

    inputText = Trim(CStr(rawValue))
    If inputText = "" Then inputText = fallbackValue
    resultText = ""
    prevUnderscore = False

    For i = 1 To Len(inputText)
        ch = Mid(inputText, i, 1)
        If IsAlphaNumeric(ch) Then
            resultText = resultText & UCase(ch)
            prevUnderscore = False
        ElseIf Not prevUnderscore Then
            resultText = resultText & "_"
            prevUnderscore = True
        End If
    Next

    Do While Len(resultText) > 0 And Left(resultText, 1) = "_"
        resultText = Mid(resultText, 2)
    Loop
    Do While Len(resultText) > 0 And Right(resultText, 1) = "_"
        resultText = Left(resultText, Len(resultText) - 1)
    Loop

    If resultText = "" Then resultText = fallbackValue
    BuildCode = resultText
End Function

Function IsAlphaNumeric(ch)
    Dim code
    If Len(ch) = 0 Then
        IsAlphaNumeric = False
        Exit Function
    End If
    code = AscW(UCase(ch))
    IsAlphaNumeric = (code >= 48 And code <= 57) Or (code >= 65 And code <= 90)
End Function

Sub SafeSetText(target, propertyName, propertyValue)
    If Trim(CStr(propertyValue)) = "" Then Exit Sub
    On Error Resume Next
    Execute "target." & propertyName & " = propertyValue"
    Err.Clear
    On Error GoTo 0
End Sub

Sub SafeSetBoolean(target, propertyName, propertyValue)
    On Error Resume Next
    Execute "target." & propertyName & " = CBool(propertyValue)"
    Err.Clear
    On Error GoTo 0
End Sub

Function ReadAllText(filePath)
    Dim fs, handle, content

    Set fs = CreateObject("Scripting.FileSystemObject")
    If Not fs.FileExists(filePath) Then
        Err.Raise vbObjectError + 1000, "ReadAllText", "No existe el archivo: " & filePath
    End If

    Set handle = fs.OpenTextFile(filePath, ForReading)
    content = handle.ReadAll
    handle.Close
    ReadAllText = content
End Function

Function LoadDbPlannerData(jsonPath)
    Dim shell, exec, commandText, stdoutText, stderrText
    Dim root, entityList, relationList, intersectionList
    Dim entityIndex, intersectionIndex
    Dim lines, lineText, fields, tag
    Dim entityData, attrList, relationData, wrapper, intersectionEntity
    Dim i

    Set root = CreateObject("Scripting.Dictionary")
    Set entityList = CreateList()
    Set relationList = CreateList()
    Set intersectionList = CreateList()
    DictSetObject root, "Entities", entityList
    DictSetObject root, "Relations", relationList
    DictSetObject root, "IntersectionEntities", intersectionList

    Set entityIndex = CreateObject("Scripting.Dictionary")
    Set intersectionIndex = CreateObject("Scripting.Dictionary")

    commandText = "powershell.exe -NoProfile -ExecutionPolicy Bypass -File " & QuoteForCmd(ExtractorScriptPath) & " -JsonPath " & QuoteForCmd(jsonPath)
    Set shell = CreateObject("WScript.Shell")
    Set exec = shell.Exec(commandText)
    stdoutText = exec.StdOut.ReadAll
    stderrText = exec.StdErr.ReadAll

    If exec.ExitCode <> 0 Then
        MsgBox "Fallo leyendo JSON con PowerShell:" & vbCrLf & stderrText
        Err.Raise vbObjectError + 1200, "LoadDbPlannerData", stderrText
    End If

    lines = Split(Replace(stdoutText, vbCrLf, vbLf), vbLf)
    For i = 0 To UBound(lines)
        lineText = Trim(lines(i))
        If lineText <> "" Then
            fields = Split(lineText, vbTab)
            tag = fields(0)

            Select Case tag
                Case "PROJECT"
                    If UBound(fields) >= 1 Then root("Name") = fields(1)

                Case "ENTITY"
                    Set entityData = CreateObject("Scripting.Dictionary")
                    entityData("Id") = ToLong(GetField(fields, 1))
                    entityData("Name") = GetField(fields, 2)
                    entityData("Description") = GetField(fields, 3)
                    Set attrList = CreateList()
                    DictSetObject entityData, "Attributes", attrList
                    DictSetObject entityIndex, CStr(entityData("Id")), entityData
                    ListAdd entityList, entityData

                Case "EATTR"
                    If entityIndex.Exists(CStr(ToLong(GetField(fields, 1)))) Then
                        Set entityData = entityIndex(CStr(ToLong(GetField(fields, 1))))
                        Set attrList = GetCollectionValue(entityData, Array("Attributes"))
                        ListAdd attrList, BuildAttributeRecord(fields)
                    End If

                Case "INTERSECTION"
                    Set wrapper = CreateObject("Scripting.Dictionary")
                    wrapper("RelationID") = ToLong(GetField(fields, 1))
                    Set intersectionEntity = CreateObject("Scripting.Dictionary")
                    intersectionEntity("Id") = ToLong(GetField(fields, 2))
                    intersectionEntity("Name") = GetField(fields, 3)
                    intersectionEntity("Description") = GetField(fields, 4)
                    Set attrList = CreateList()
                    DictSetObject intersectionEntity, "Attributes", attrList
                    DictSetObject wrapper, "Entity", intersectionEntity
                    DictSetObject intersectionIndex, CStr(wrapper("RelationID")), wrapper
                    ListAdd intersectionList, wrapper

                Case "IATTR"
                    If intersectionIndex.Exists(CStr(ToLong(GetField(fields, 1)))) Then
                        Set wrapper = intersectionIndex(CStr(ToLong(GetField(fields, 1))))
                        Set intersectionEntity = GetObject(wrapper, Array("Entity"))
                        Set attrList = GetCollectionValue(intersectionEntity, Array("Attributes"))
                        ListAdd attrList, BuildAttributeRecord(fields)
                    End If

                Case "REL"
                    Set relationData = CreateObject("Scripting.Dictionary")
                    relationData("Id") = ToLong(GetField(fields, 1))
                    relationData("IdEntity1") = ToLong(GetField(fields, 2))
                    relationData("IdEntity2") = ToLong(GetField(fields, 3))
                    relationData("Relation") = GetField(fields, 4)
                    ListAdd relationList, relationData
            End Select
        End If
    Next

    Set LoadDbPlannerData = root
End Function

Function BuildAttributeRecord(fields)
    Dim attrData
    Set attrData = CreateObject("Scripting.Dictionary")
    attrData("Id") = ToLong(GetField(fields, 2))
    attrData("Name") = GetField(fields, 3)
    attrData("Description") = GetField(fields, 4)
    attrData("Type") = GetField(fields, 5)
    attrData("KeyType") = GetField(fields, 6)
    attrData("Optional") = (GetField(fields, 7) = "1")
    DictSetObject attrData, "Domain", BuildDomainList(GetField(fields, 8))
    Set BuildAttributeRecord = attrData
End Function

Function BuildDomainList(domainText)
    Dim resultList, parts, i
    Set resultList = CreateList()
    If Trim(domainText) <> "" Then
        parts = Split(domainText, " | ")
        For i = 0 To UBound(parts)
            ListAdd resultList, parts(i)
        Next
    End If
    Set BuildDomainList = resultList
End Function

Function GetField(fields, idx)
    If UBound(fields) >= idx Then
        GetField = fields(idx)
    Else
        GetField = ""
    End If
End Function

Function ToLong(textValue)
    If Trim(CStr(textValue)) = "" Then
        ToLong = 0
    Else
        ToLong = CLng(textValue)
    End If
End Function

Function QuoteForCmd(textValue)
    QuoteForCmd = """" & Replace(CStr(textValue), """", """""") & """"
End Function

Function GetValue(source, keyNames, defaultValue)
    Dim idx, keyName

    GetValue = defaultValue
    If TypeName(source) <> "Dictionary" Then Exit Function

    For idx = 0 To UBound(keyNames)
        keyName = CStr(keyNames(idx))
        If source.Exists(keyName) Then
            If Not IsObject(source(keyName)) Then
                GetValue = source(keyName)
            End If
            Exit Function
        End If
    Next
End Function

Function GetObject(source, keyNames)
    Dim idx, keyName

    Set GetObject = Nothing
    If TypeName(source) <> "Dictionary" Then Exit Function

    For idx = 0 To UBound(keyNames)
        keyName = CStr(keyNames(idx))
        If source.Exists(keyName) Then
            If IsObject(source(keyName)) Then
                Set GetObject = source(keyName)
            End If
            Exit Function
        End If
    Next
End Function

Function GetCollectionValue(source, keyNames)
    Dim obj
    Set obj = GetObject(source, keyNames)
    If obj Is Nothing Or TypeName(obj) <> "Collection" Then
        Set GetCollectionValue = CreateList()
    ElseIf TypeName(obj) = "Dictionary" Then
        Set GetCollectionValue = obj
    Else
        Set GetCollectionValue = CreateList()
    End If
End Function

Function HasAnyKey(source, keyNames)
    Dim idx

    HasAnyKey = False
    If TypeName(source) <> "Dictionary" Then Exit Function

    For idx = 0 To UBound(keyNames)
        If source.Exists(CStr(keyNames(idx))) Then
            HasAnyKey = True
            Exit Function
        End If
    Next
End Function

Function JoinCollection(items, separator)
    Dim resultText, item

    resultText = ""
    For Each item In ListValues(items)
        If resultText <> "" Then resultText = resultText & separator
        resultText = resultText & CStr(item)
    Next
    JoinCollection = resultText
End Function

Function CreateList()
    Set CreateList = CreateObject("Scripting.Dictionary")
End Function

Sub ListAdd(listObject, itemValue)
    Dim nextKey
    nextKey = CStr(listObject.Count)
    If IsObject(itemValue) Then
        DictSetObject listObject, nextKey, itemValue
    Else
        listObject.Add nextKey, itemValue
    End If
End Sub

Function ListValues(listObject)
    If TypeName(listObject) = "Dictionary" Then
        ListValues = listObject.Items
    Else
        ListValues = Array()
    End If
End Function

Function ListCount(listObject)
    If TypeName(listObject) = "Dictionary" Then
        ListCount = listObject.Count
    Else
        ListCount = 0
    End If
End Function

Sub DictSetObject(dictObject, keyName, objectValue)
    If dictObject.Exists(keyName) Then
        Set dictObject(keyName) = objectValue
    Else
        dictObject.Add keyName, Nothing
        Set dictObject(keyName) = objectValue
    End If
End Sub

'========================
' JSON parser
'========================

Function ParseJSON(jsonText)
    gJsonText = jsonText
    gJsonPos = 1
    SkipWhitespace
    If PeekChar() <> "{" Then
        Err.Raise vbObjectError + 1102, "ParseJSON", "El JSON raiz debe ser un objeto."
    End If
    Set ParseJSON = ParseObject()
End Function

Function ParseObject()
    Dim dict, key, ch, scalarValue
    Dim objValue

    Set dict = CreateObject("Scripting.Dictionary")
    ConsumeChar "{"
    SkipWhitespace

    If PeekChar() = "}" Then
        ConsumeChar "}"
        Set ParseObject = dict
        Exit Function
    End If

    Do
        SkipWhitespace
        key = ParseString()
        SkipWhitespace
        ConsumeChar ":"
        SkipWhitespace

        ch = PeekChar()
        If ch = "{" Then
            Set objValue = ParseObject()
            DictSetObject dict, key, objValue
        ElseIf ch = "[" Then
            Set objValue = ParseArray()
            DictSetObject dict, key, objValue
        Else
            scalarValue = ParseScalar()
            dict(key) = scalarValue
        End If

        SkipWhitespace
        ch = PeekChar()
        If ch = "}" Then
            ConsumeChar "}"
            Exit Do
        End If
        ConsumeChar ","
    Loop

    Set ParseObject = dict
End Function

Function ParseArray()
    Dim arr, ch, scalarValue
    Dim objValue

    Set arr = CreateList()
    ConsumeChar "["
    SkipWhitespace

    If PeekChar() = "]" Then
        ConsumeChar "]"
        Set ParseArray = arr
        Exit Function
    End If

    Do
        ch = PeekChar()
        If ch = "{" Then
            Set objValue = ParseObject()
            ListAdd arr, objValue
        ElseIf ch = "[" Then
            Set objValue = ParseArray()
            ListAdd arr, objValue
        Else
            scalarValue = ParseScalar()
            ListAdd arr, scalarValue
        End If

        SkipWhitespace
        ch = PeekChar()
        If ch = "]" Then
            ConsumeChar "]"
            Exit Do
        End If
        ConsumeChar ","
        SkipWhitespace
    Loop

    Set ParseArray = arr
End Function

Function ParseScalar()
    Dim ch

    ch = PeekChar()
    Select Case ch
        Case """"
            ParseScalar = ParseString()
        Case "t"
            ParseLiteral "true"
            ParseScalar = True
        Case "f"
            ParseLiteral "false"
            ParseScalar = False
        Case "n"
            ParseLiteral "null"
            ParseScalar = Null
        Case Else
            ParseScalar = ParseNumber()
    End Select
End Function

Function ParseString()
    Dim resultText, ch, hexValue

    resultText = ""
    ConsumeChar """"

    Do While gJsonPos <= Len(gJsonText)
        ch = Mid(gJsonText, gJsonPos, 1)
        gJsonPos = gJsonPos + 1

        If ch = """" Then Exit Do
        If ch = "\" Then
            ch = Mid(gJsonText, gJsonPos, 1)
            gJsonPos = gJsonPos + 1
            Select Case ch
                Case """": resultText = resultText & """"
                Case "\": resultText = resultText & "\"
                Case "/": resultText = resultText & "/"
                Case "b": resultText = resultText & Chr(8)
                Case "f": resultText = resultText & Chr(12)
                Case "n": resultText = resultText & vbLf
                Case "r": resultText = resultText & vbCr
                Case "t": resultText = resultText & vbTab
                Case "u"
                    hexValue = Mid(gJsonText, gJsonPos, 4)
                    gJsonPos = gJsonPos + 4
                    resultText = resultText & ChrW(CLng("&H" & hexValue))
            End Select
        Else
            resultText = resultText & ch
        End If
    Loop

    ParseString = resultText
End Function

Function ParseNumber()
    Dim startPos, token

    startPos = gJsonPos
    Do While gJsonPos <= Len(gJsonText)
        token = Mid(gJsonText, gJsonPos, 1)
        If InStr("0123456789+-.eE", token) = 0 Then Exit Do
        gJsonPos = gJsonPos + 1
    Loop

    token = Mid(gJsonText, startPos, gJsonPos - startPos)
    If InStr(token, ".") > 0 Or InStr(UCase(token), "E") > 0 Then
        ParseNumber = CDbl(token)
    Else
        ParseNumber = CLng(token)
    End If
End Function

Sub ParseLiteral(expectedText)
    Dim actualText

    actualText = Mid(gJsonText, gJsonPos, Len(expectedText))
    If actualText <> expectedText Then
        Err.Raise vbObjectError + 1100, "ParseLiteral", "Literal invalido en posicion " & CStr(gJsonPos)
    End If
    gJsonPos = gJsonPos + Len(expectedText)
End Sub

Sub SkipWhitespace()
    Dim ch

    Do While gJsonPos <= Len(gJsonText)
        ch = Mid(gJsonText, gJsonPos, 1)
        If ch <> " " And ch <> vbTab And ch <> vbCr And ch <> vbLf Then Exit Do
        gJsonPos = gJsonPos + 1
    Loop
End Sub

Sub ConsumeChar(expectedChar)
    If PeekChar() <> expectedChar Then
        Err.Raise vbObjectError + 1101, "ConsumeChar", "Se esperaba '" & expectedChar & "' en posicion " & CStr(gJsonPos)
    End If
    gJsonPos = gJsonPos + 1
End Sub

Function PeekChar()
    If gJsonPos > Len(gJsonText) Then
        PeekChar = ""
    Else
        PeekChar = Mid(gJsonText, gJsonPos, 1)
    End If
End Function
