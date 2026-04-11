# PowerDesigner import scripts

Scripts para importar el JSON de `db_planner` dentro de SAP PowerDesigner.

Archivos:
- `import_dbplanner_conceptual.vbs`: crea un `CDM`.
- `import_dbplanner_logical.vbs`: crea un `LDM`.

JSON soportado:
- El `proyecto.json` completo guardado por la app.
- El JSON exportado desde la opción de exportación de entidades.

Reglas de modelado:
- Las entidades fuertes salen como entidades normales.
- Las relaciones `1:1`, `1:N` y `N:1` se crean como relaciones directas.
- Las relaciones `1:Np` y `Np:1` se crean como relaciones dependientes para que PowerDesigner migre el identificador al hijo.
- Las intersecciones `N:N` se materializan como entidad intermedia cuando existe una `IntersectionEntity` en el JSON. En el script lógico esto fuerza la herencia/migración de identificadores con la lógica nativa de PowerDesigner.

Uso:
1. Guarda o exporta el JSON desde `db_planner`.
2. Abre PowerDesigner.
3. Ejecuta el script VB correspondiente.
4. Cuando el script lo pida, pega la ruta completa del archivo JSON.

Notas:
- Los scripts crean un modelo nuevo; no modifican el modelo activo.
- El `Code` de entidades y atributos se normaliza para evitar caracteres problemáticos.
- Si un atributo tiene dominio enumerado (`Domain` en el JSON), el script lo agrega a la descripción/comentario del atributo.
