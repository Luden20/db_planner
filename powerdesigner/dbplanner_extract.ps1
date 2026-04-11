param(
    [Parameter(Mandatory = $true)]
    [string]$JsonPath
)

$ErrorActionPreference = 'Stop'

function Clean([object]$Value) {
    if ($null -eq $Value) { return '' }
    return (($Value.ToString() -replace "`r?`n", ' ' -replace "`t", ' ').Trim())
}

function Emit([string[]]$Fields) {
    Write-Output (($Fields | ForEach-Object { Clean $_ }) -join "`t")
}

$raw = Get-Content -Raw -LiteralPath $JsonPath
$root = $raw | ConvertFrom-Json

$entities = @()
if ($null -ne $root.Entities) { $entities = @($root.Entities) }

$relations = @()
if ($null -ne $root.Relations) { $relations = @($root.Relations) }

$intersections = @()
if ($null -ne $root.IntersectionEntities) { $intersections = @($root.IntersectionEntities) }

Emit @('PROJECT', $root.Name)

foreach ($entity in $entities) {
    Emit @('ENTITY', $entity.Id, $entity.Name, $entity.Description)
    foreach ($attr in @($entity.Attributes)) {
        $domain = @($attr.Domain) -join ' | '
        Emit @('EATTR', $entity.Id, $attr.Id, $attr.Name, $attr.Description, $attr.Type, $attr.KeyType, [int][bool]$attr.Optional, $domain)
    }
}

foreach ($item in $intersections) {
    $relationId = 0
    $entity = $item
    if ($null -ne $item.Entity) {
        $relationId = $item.RelationID
        $entity = $item.Entity
    }

    Emit @('INTERSECTION', $relationId, $entity.Id, $entity.Name, $entity.Description)
    foreach ($attr in @($entity.Attributes)) {
        $domain = @($attr.Domain) -join ' | '
        Emit @('IATTR', $relationId, $attr.Id, $attr.Name, $attr.Description, $attr.Type, $attr.KeyType, [int][bool]$attr.Optional, $domain)
    }
}

foreach ($rel in $relations) {
    Emit @('REL', $rel.Id, $rel.IdEntity1, $rel.IdEntity2, $rel.Relation)
}
