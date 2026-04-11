Option Explicit

Const ForReading = 1

Dim gJsonText
Dim gJsonPos

Dim jsonPath, root, entities, relations, intersections
jsonPath = WScript.Arguments.Item(0)

Set root = ParseJSON(ReadAllText(jsonPath))
Set entities = GetCollectionValue(root, Array("entities", "Entities"))
Set relations = GetCollectionValue(root, Array("relations", "Relations"))
Set intersections = GetCollectionValue(root, Array("intersection_entities", "IntersectionEntities"))

WScript.Echo "Root type=" & TypeName(root)
WScript.Echo "Entities count=" & CStr(ListCount(entities))
WScript.Echo "Relations count=" & CStr(ListCount(relations))
WScript.Echo "Intersections count=" & CStr(ListCount(intersections))
WScript.Echo "Root has Name? " & CStr(HasAnyKey(root, Array("Name")))
WScript.Echo "Root has Entities? " & CStr(HasAnyKey(root, Array("Entities")))

Function ReadAllText(filePath)
    Dim fs, handle, content
    Set fs = CreateObject("Scripting.FileSystemObject")
    Set handle = fs.OpenTextFile(filePath, ForReading)
    content = handle.ReadAll
    handle.Close
    ReadAllText = content
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
    If obj Is Nothing Or TypeName(obj) <> "Dictionary" Then
        Set GetCollectionValue = CreateList()
    Else
        Set GetCollectionValue = obj
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
    Dim dict, key, ch, scalarValue, objValue
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
    Dim arr, ch, scalarValue, objValue
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
    If Mid(gJsonText, gJsonPos, Len(expectedText)) <> expectedText Then
        Err.Raise vbObjectError + 1100, "ParseLiteral", "Literal invalido"
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
        Err.Raise vbObjectError + 1101, "ConsumeChar", "Se esperaba " & expectedChar
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
