package utils

import (
	"fmt"
	"slices"
)

type AnalysisIssue struct {
	Type      string     `json:"type"`      // "Circular" o "Ambiguity"
	Entities  []string   `json:"entities"`  // [Origen, Destino] en Ambiguity, o lista en Circular
	PathCount int        `json:"pathCount"`
	Paths     [][]string `json:"paths"`     // Caminos concretos (nombres de tablas)
}

type AnalysisReport struct {
	Issues   []AnalysisIssue `json:"issues"`
	EdgeList []string        `json:"edgeList"` // Lista de aristas interpretadas para validación
}

// dependencyGraph usa IDs internos secuenciales (0..N) para evitar colisiones
// entre los IDs de Entidades Fuertes y Entidades de Intersección.
type dependencyGraph struct {
	adj     map[int][]int
	nodeMap map[int]string
}

// nodeKey se utiliza para el mapeo único de identidades
type nodeKey struct {
	Type string // "Strong" o "Intersection"
	ID   int
}

func (p *DbProject) AnalyzeProjectDependencies() AnalysisReport {
	g, edges := p.buildDependencyGraph()
	report := AnalysisReport{
		Issues:   make([]AnalysisIssue, 0),
		EdgeList: edges,
	}

	// 1. Detectar dependencias circulares críticas
	circularities := g.detectCycles()
	for _, cycle := range circularities {
		names := make([]string, 0, len(cycle))
		for _, id := range cycle {
			names = append(names, g.nodeMap[id])
		}
		report.Issues = append(report.Issues, AnalysisIssue{
			Type:     "Circular",
			Entities: names,
		})
	}

	// 2. DETECTAR AMBIGÜEDAD ESTRUCTURAL (Padre -> Hijo)
	// Usamos un caché para evitar recalcular rutas en el filtro de supresión
	pathCache := make(map[string][][]int)
	getPathPaths := func(u, v int) [][]int {
		key := fmt.Sprintf("%d:%d", u, v)
		if p, ok := pathCache[key]; ok {
			return p
		}
		p := g.findAllSimplePaths(u, v)
		pathCache[key] = p
		return p
	}

	nodes := make([]int, 0, len(g.nodeMap))
	for id := range g.nodeMap {
		nodes = append(nodes, id)
	}
	slices.Sort(nodes)

	// Candidatos iniciales
	type cand struct{ u, v int }
	candidates := make([]cand, 0)

	for i := 0; i < len(nodes); i++ {
		for j := 0; j < len(nodes); j++ {
			if i == j {
				continue
			}
			u, v := nodes[i], nodes[j]
			if len(getPathPaths(u, v)) > 1 {
				candidates = append(candidates, cand{u, v})
			}
		}
	}

	// FILTRADO DE SUPRESIÓN:
	// Identificar el punto más bajo (más cercano al destino) donde nace la ambigüedad.
	for _, c := range candidates {
		isInherited := false
		for _, child := range g.adj[c.u] {
			if len(getPathPaths(child, c.v)) > 1 {
				isInherited = true
				break
			}
		}

		if !isInherited {
			simplePaths := getPathPaths(c.u, c.v)
			namedPaths := make([][]string, 0, len(simplePaths))
			for _, path := range simplePaths {
				namedPath := make([]string, 0, len(path))
				for _, nodeID := range path {
					namedPath = append(namedPath, g.nodeMap[nodeID])
				}
				namedPaths = append(namedPaths, namedPath)
			}

			report.Issues = append(report.Issues, AnalysisIssue{
				Type:      "Ambiguity",
				Entities:  []string{g.nodeMap[c.u], g.nodeMap[c.v]},
				PathCount: len(simplePaths),
				Paths:     namedPaths,
			})
		}
	}

	return report
}

func (p *DbProject) buildDependencyGraph() (*dependencyGraph, []string) {
	g := &dependencyGraph{
		adj:     make(map[int][]int),
		nodeMap: make(map[int]string),
	}
	edgeList := make([]string, 0)
	
	// Mapeador de Identidades Únicas para evitar colisiones de IDs
	internalMap := make(map[nodeKey]int)
	nextInternalID := 0

	getInternalID := func(t string, originalID int, name string) int {
		key := nodeKey{Type: t, ID: originalID}
		if id, ok := internalMap[key]; ok {
			// Si el nombre no estaba registrado (llamada desde relación), actualizarlo
			if name != "" && g.nodeMap[id] == "" {
				g.nodeMap[id] = name
			}
			return id
		}
		id := nextInternalID
		internalMap[key] = id
		g.nodeMap[id] = name
		g.adj[id] = []int{}
		nextInternalID++
		return id
	}

	// Registrar Entidades Fuertes primero para asegurar nombres
	for _, ent := range p.Entities {
		getInternalID("Strong", ent.Id, ent.Name)
	}

	// Registrar Entidades de Intersección
	for _, ie := range p.IntersectionEntities {
		getInternalID("Intersection", ie.Entity.Id, ie.Entity.Name)
	}

	// Construir aristas basadas en el flujo Padre -> Hijo
	for _, rel := range p.Relations {
		// Validar si la relación es real (omitir "-", "", etc.)
		if !isAllowedRelationValue(rel.Relation) {
			continue
		}

		switch rel.Relation {
		case RelationTypeNN:
			ie := p.GetIntersectionEntityByRelationID(rel.Id)
			if ie != nil {
				// El flujo semántico va de los padres a la intersección (foso)
				u1ID := getInternalID("Strong", rel.IdEntity1, "") 
				u2ID := getInternalID("Strong", rel.IdEntity2, "")
				vID := getInternalID("Intersection", ie.Entity.Id, ie.Entity.Name)

				g.adj[u1ID] = append(g.adj[u1ID], vID)
				g.adj[u2ID] = append(g.adj[u2ID], vID)
				
				edgeList = append(edgeList, fmt.Sprintf("%s -> %s", g.nodeMap[u1ID], g.nodeMap[vID]))
				edgeList = append(edgeList, fmt.Sprintf("%s -> %s", g.nodeMap[u2ID], g.nodeMap[vID]))
			}
		case RelationType11, RelationType1N, RelationType1Np:
			// Ent1 es Padre, Ent2 es Hijo/Dependiente
			uID := getInternalID("Strong", rel.IdEntity1, "")
			vID := getInternalID("Strong", rel.IdEntity2, "")
			g.adj[uID] = append(g.adj[uID], vID)
			edgeList = append(edgeList, fmt.Sprintf("%s -> %s", g.nodeMap[uID], g.nodeMap[vID]))
		case RelationTypeN1, RelationTypeNp1:
			// Ent2 es Padre, Ent1 es Hijo/Dependiente
			uID := getInternalID("Strong", rel.IdEntity2, "")
			vID := getInternalID("Strong", rel.IdEntity1, "")
			g.adj[uID] = append(g.adj[uID], vID)
			edgeList = append(edgeList, fmt.Sprintf("%s -> %s", g.nodeMap[uID], g.nodeMap[vID]))
		}
	}

	for id := range g.adj {
		slices.Sort(g.adj[id])
		g.adj[id] = slices.Compact(g.adj[id])
	}

	slices.Sort(edgeList)
	edgeList = slices.Compact(edgeList)

	return g, edgeList
}

func (g *dependencyGraph) detectCycles() [][]int {
	var cycles [][]int
	visited := make(map[int]bool)
	onStack := make(map[int]bool)
	var stack []int

	var dfs func(int)
	dfs = func(u int) {
		visited[u] = true
		onStack[u] = true
		stack = append(stack, u)

		for _, v := range g.adj[u] {
			if onStack[v] {
				var cycle []int
				found := false
				for _, node := range stack {
					if node == v {
						found = true
					}
					if found {
						cycle = append(cycle, node)
					}
				}
				cycles = append(cycles, cycle)
			} else if !visited[v] {
				dfs(v)
			}
		}

		onStack[u] = false
		stack = stack[:len(stack)-1]
	}

	for id := range g.nodeMap {
		if !visited[id] {
			dfs(id)
		}
	}

	return cycles
}

func (g *dependencyGraph) findAllSimplePaths(start, end int) [][]int {
	var paths [][]int
	visited := make(map[int]bool)
	var currentPath []int

	var dfs func(int)
	dfs = func(u int) {
		visited[u] = true
		currentPath = append(currentPath, u)

		if u == end {
			pathCopy := make([]int, len(currentPath))
			copy(pathCopy, currentPath)
			paths = append(paths, pathCopy)
		} else {
			for _, v := range g.adj[u] {
				if !visited[v] {
					dfs(v)
				}
			}
		}

		currentPath = currentPath[:len(currentPath)-1]
		visited[u] = false
	}

	dfs(start)
	return paths
}
