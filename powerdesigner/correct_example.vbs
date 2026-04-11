Option Explicit

' =========================================================
' MODELO LOGICO: VENTAS
'
' Entidades:
'   CLIENTE
'   PEDIDO
'   PRODUCTO
'   DETALLE_PEDIDO   -> tabla intersección
'
' Relaciones:
'   CLIENTE 1 --- N PEDIDO
'   PEDIDO  1 --- N DETALLE_PEDIDO
'   PRODUCTO 1 --- N DETALLE_PEDIDO
'
' La N:N entre PEDIDO y PRODUCTO se resuelve con DETALLE_PEDIDO.
' =========================================================

Dim mdl
Set mdl = CreateModel(PdLDM.cls_Model)

If mdl Is Nothing Then
    MsgBox "No se pudo crear el modelo."
    WScript.Quit
End If

mdl.SetNameAndCode "ModeloVentas", "MODELO_VENTAS"
Call TrySetComment(mdl, "Modelo lógico de ejemplo con dominios, comentarios, relación 1:N y N:N mediante tabla de intersección.")

' =========================================================
' HELPERS GENERALES
' =========================================================

Sub TrySetComment(obj, txt)
    On Error Resume Next
    obj.Comment = txt
    obj.Description = txt
    On Error GoTo 0
End Sub

Function AddEntity(model, code, name, commentText)
    Dim ent
    Set ent = model.Entities.CreateNew()
    ent.SetNameAndCode name, code
    Call TrySetComment(ent, commentText)
    Set AddEntity = ent
End Function

Function AddDomain(model, code, name, dataType, lengthValue, commentText)
    Dim dom
    Set dom = model.Domains.CreateNew()
    dom.SetNameAndCode name, code

    On Error Resume Next
    dom.DataType = dataType
    dom.Length = lengthValue
    dom.DefaultDataType = dataType
    On Error GoTo 0

    Call TrySetComment(dom, commentText)
    Set AddDomain = dom
End Function

Function AddAttribute(ent, code, name, dom, mandatory, commentText)
    Dim att
    Set att = ent.Attributes.CreateNew()
    att.SetNameAndCode name, code
    att.Mandatory = mandatory

    On Error Resume Next
    Set att.Domain = dom
    If Err.Number <> 0 Then
        Err.Clear
        att.DataType = dom.DataType
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

' =========================================================
' DOMINIOS
' =========================================================

Dim dId, dNombre, dFecha, dCantidad

Set dId = AddDomain( _
    mdl, _
    "DOM_ID", _
    "Dominio Id", _
    "Integer", _
    0, _
    "Dominio para identificadores numéricos." _
)

Set dNombre = AddDomain( _
    mdl, _
    "DOM_NOMBRE_100", _
    "Dominio Nombre 100", _
    "String", _
    100, _
    "Dominio para nombres de hasta 100 caracteres." _
)

Set dFecha = AddDomain( _
    mdl, _
    "DOM_FECHA", _
    "Dominio Fecha", _
    "Date", _
    0, _
    "Dominio para fechas." _
)

Set dCantidad = AddDomain( _
    mdl, _
    "DOM_CANTIDAD", _
    "Dominio Cantidad", _
    "Integer", _
    0, _
    "Dominio para cantidades enteras." _
)

' =========================================================
' ENTIDADES
' =========================================================

Dim eCliente, ePedido, eProducto, eDetalle

Set eCliente = AddEntity( _
    mdl, _
    "CLIENTE", _
    "Cliente", _
    "Entidad que almacena la información básica de los clientes." _
)

Set ePedido = AddEntity( _
    mdl, _
    "PEDIDO", _
    "Pedido", _
    "Entidad que registra los pedidos realizados por los clientes." _
)

Set eProducto = AddEntity( _
    mdl, _
    "PRODUCTO", _
    "Producto", _
    "Entidad que almacena los productos disponibles." _
)

Set eDetalle = AddEntity( _
    mdl, _
    "DETALLE_PEDIDO", _
    "Detalle Pedido", _
    "Entidad asociativa que resuelve la relación N:N entre pedido y producto." _
)

' =========================================================
' ATRIBUTOS
' =========================================================

Dim aCliId, aCliNombre
Dim aPedId, aPedFecha, aPedClienteId
Dim aProId, aProNombre
Dim aDetPedidoId, aDetProductoId, aDetCantidad

' -------------------------
' CLIENTE
' -------------------------
Set aCliId = AddAttribute( _
    eCliente, _
    "ID_CLIENTE", _
    "Id Cliente", _
    dId, _
    True, _
    "Clave primaria del cliente." _
)
Call TryMarkAsPK(aCliId)

Set aCliNombre = AddAttribute( _
    eCliente, _
    "NOMBRE", _
    "Nombre", _
    dNombre, _
    True, _
    "Nombre del cliente." _
)

' -------------------------
' PEDIDO
' -------------------------
Set aPedId = AddAttribute( _
    ePedido, _
    "ID_PEDIDO", _
    "Id Pedido", _
    dId, _
    True, _
    "Clave primaria del pedido." _
)
Call TryMarkAsPK(aPedId)

Set aPedFecha = AddAttribute( _
    ePedido, _
    "FECHA", _
    "Fecha", _
    dFecha, _
    True, _
    "Fecha en la que se registra el pedido." _
)

Set aPedClienteId = AddAttribute( _
    ePedido, _
    "ID_CLIENTE", _
    "Id Cliente", _
    dId, _
    True, _
    "Clave foránea que referencia al cliente que realiza el pedido." _
)
Call TryMarkAsFK(aPedClienteId)

' -------------------------
' PRODUCTO
' -------------------------
Set aProId = AddAttribute( _
    eProducto, _
    "ID_PRODUCTO", _
    "Id Producto", _
    dId, _
    True, _
    "Clave primaria del producto." _
)
Call TryMarkAsPK(aProId)

Set aProNombre = AddAttribute( _
    eProducto, _
    "NOMBRE", _
    "Nombre", _
    dNombre, _
    True, _
    "Nombre del producto." _
)

' -------------------------
' DETALLE_PEDIDO
' -------------------------
Set aDetPedidoId = AddAttribute( _
    eDetalle, _
    "ID_PEDIDO", _
    "Id Pedido", _
    dId, _
    True, _
    "Parte de la clave primaria compuesta y clave foránea hacia PEDIDO." _
)
Call TryMarkAsPK(aDetPedidoId)
Call TryMarkAsFK(aDetPedidoId)

Set aDetProductoId = AddAttribute( _
    eDetalle, _
    "ID_PRODUCTO", _
    "Id Producto", _
    dId, _
    True, _
    "Parte de la clave primaria compuesta y clave foránea hacia PRODUCTO." _
)
Call TryMarkAsPK(aDetProductoId)
Call TryMarkAsFK(aDetProductoId)

Set aDetCantidad = AddAttribute( _
    eDetalle, _
    "CANTIDAD", _
    "Cantidad", _
    dCantidad, _
    True, _
    "Cantidad del producto incluida en el pedido." _
)

' =========================================================
' RELACIONES
' =========================================================

Dim rClientePedido, rPedidoDetalle, rProductoDetalle

' CLIENTE 1 --- N PEDIDO
Set rClientePedido = AddRelationship( _
    mdl, _
    eCliente, _
    ePedido, _
    "R_CLIENTE_PEDIDO", _
    "Cliente realiza Pedido", _
    "Un cliente puede realizar muchos pedidos; cada pedido pertenece a un cliente." _
)
Call TrySetCardinality(rClientePedido, "1", "N")
Call TrySetIdentifying(rClientePedido, False)

' PEDIDO 1 --- N DETALLE_PEDIDO
Set rPedidoDetalle = AddRelationship( _
    mdl, _
    ePedido, _
    eDetalle, _
    "R_PEDIDO_DETALLE", _
    "Pedido tiene Detalle", _
    "Un pedido puede tener muchos detalles; cada detalle pertenece a un pedido." _
)
Call TrySetCardinality(rPedidoDetalle, "1", "N")
Call TrySetIdentifying(rPedidoDetalle, True)

' PRODUCTO 1 --- N DETALLE_PEDIDO
Set rProductoDetalle = AddRelationship( _
    mdl, _
    eProducto, _
    eDetalle, _
    "R_PRODUCTO_DETALLE", _
    "Producto aparece en Detalle", _
    "Un producto puede aparecer en muchos detalles; cada detalle referencia un producto." _
)
Call TrySetCardinality(rProductoDetalle, "1", "N")
Call TrySetIdentifying(rProductoDetalle, True)

' =========================================================
' DIAGRAMA
' =========================================================

Dim diag
Set diag = ActiveDiagram

If diag Is Nothing Then
    On Error Resume Next
    Set diag = mdl.DefaultDiagram
    On Error GoTo 0
End If

If Not diag Is Nothing Then
    Call AttachEntity(diag, eCliente, 100, 100)
    Call AttachEntity(diag, ePedido, 360, 100)
    Call AttachEntity(diag, eProducto, 360, 320)
    Call AttachEntity(diag, eDetalle, 670, 210)

    Call AttachRelationship(diag, rClientePedido)
    Call AttachRelationship(diag, rPedidoDetalle)
    Call AttachRelationship(diag, rProductoDetalle)

    On Error Resume Next
    diag.AutoLayoutWithOptions 2
    On Error GoTo 0
End If

MsgBox "Modelo lógico creado correctamente con dominios y comentarios."