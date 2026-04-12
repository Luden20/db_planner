<script lang="ts">
  import { onMount } from "svelte";
  import type { utils } from "../../wailsjs/go/models";
  import { GetCombinatory } from "../../wailsjs/go/main/App";
  import ButtonIcon from "./ButtonIcon.svelte";
  import { showToast } from "../lib/toast";

  type GraphNode = { id: number; name: string; status: boolean; degree: number; width: number; height: number; x: number; y: number; component: number; depth: number; };
  type GraphEdge = { key: string; source: number; target: number; relation: string; sourceName: string; targetName: string; };
  type ComponentLayout = { nodes: GraphNode[]; width: number; height: number; };
  type ViewTransitionDocument = Document & { startViewTransition?: (update: () => void | Promise<void>) => { finished: Promise<void>; }; };

  let { 
    entities = [], 
    onJumpTo = () => {} 
  } = $props<{
    entities?: utils.Entity[];
    onJumpTo?: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void;
  }>();

  const BASE_NODE_WIDTH = 178;
  const ROOT_NODE_WIDTH = 212;
  const NODE_HEIGHT = 84;
  const ROOT_NODE_HEIGHT = 92;
  const COMPONENT_GAP_X = 186;
  const COMPONENT_GAP_Y = 152;
  const CANVAS_PADDING = 72;
  const MAX_CANVAS_ROW_WIDTH = 2120;
  const RING_STEP_X = 254;
  const RING_STEP_Y = 176;
  const ZOOM_MIN = 0.4;
  const ZOOM_MAX = 1.85;

  let loading = $state(true);
  let loadSignature = $state("");
  let relationViews = $state<utils.RelationView[]>([]);
  let nodes = $state<GraphNode[]>([]);
  let edges = $state<GraphEdge[]>([]);
  let boardWidth = $state(1320);
  let boardHeight = $state(900);
  let selectedId = $state<number | null>(null);
  let hoverId = $state<number | null>(null);
  let searchQuery = $state("");
  let zoom = $state(1);
  let panX = $state(0);
  let panY = $state(0);
  let viewportWidth = $state(0);
  let viewportHeight = $state(0);
  let viewportEl = $state<HTMLDivElement | null>(null);
  let minimapEl = $state<SVGSVGElement | null>(null);
  let isStageDragging = $state(false);
  let dragStartX = 0;
  let dragStartY = 0;
  let dragOriginPanX = 0;
  let dragOriginPanY = 0;
  let userHasMovedCamera = $state(false);
  let cameraAnimating = $state(false);
  let animationTimeout: number | null = null;

  const currentFocusId = $derived(hoverId ?? selectedId);
  const nodeLookup = $derived(new Map(nodes.map((node) => [node.id, node])));
  const connectedIds = $derived(relatedNodeIds(currentFocusId));
  const selectedNode = $derived(nodeById(selectedId));
  const searchResults = $derived(searchResultsFromNodes(searchQuery));
  const minimapAspect = $derived(boardWidth > 0 && boardHeight > 0 ? (boardHeight / boardWidth) : 0.6);
  const viewportRect = $derived(zoom === 0 ? {x: 0, y: 0, width: boardWidth, height: boardHeight} : { x: Math.max(0, -panX / zoom), y: Math.max(0, -panY / zoom), width: viewportWidth / zoom, height: viewportHeight / zoom });
  const relationLegend = $derived(Object.entries(edges.reduce<Record<string, number>>((acc, edge) => { acc[edge.relation] = (acc[edge.relation] ?? 0) + 1; return acc; }, {})).sort(([l], [r]) => l.localeCompare(r)));

  const relationPalette: Record<string, {line: string; fill: string; label: string}> = {
    "1:1": {line: "var(--accent)", fill: "color-mix(in srgb, var(--accent) 16%, transparent)", label: "Uno a uno"},
    "1:N": {line: "var(--success)", fill: "color-mix(in srgb, var(--success) 16%, transparent)", label: "Uno a muchos"},
    "N:1": {line: "var(--danger)", fill: "color-mix(in srgb, var(--danger) 16%, transparent)", label: "Muchos a uno"},
    "N:N": {line: "var(--ink-soft)", fill: "color-mix(in srgb, var(--ink) 12%, transparent)", label: "Muchos a muchos"}
  };

  const normalizeText = (value: string) => (value || "").normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLowerCase().trim();
  const entitySignature = () => entities.map((entity) => `${entity.Id}:${entity.Status ? 1 : 0}:${entity.Name}`).join("|");
  const edgeKey = (a: number, b: number) => { const first = Math.min(a, b); const second = Math.max(a, b); return `${first}-${second}`; };
  const relationMeta = (relation: string) => relationPalette[relation] ?? relationPalette["N:N"];
  const degreeMapFromEdges = (graphEdges: GraphEdge[]) => { const degreeMap = new Map<number, number>(); graphEdges.forEach((edge) => { degreeMap.set(edge.source, (degreeMap.get(edge.source) ?? 0) + 1); degreeMap.set(edge.target, (degreeMap.get(edge.target) ?? 0) + 1); }); return degreeMap; };
  const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));
  const nodeDimensions = (name: string, isRoot: boolean) => { const width = Math.min(isRoot ? 248 : 228, Math.max(isRoot ? ROOT_NODE_WIDTH : BASE_NODE_WIDTH, name.length * 8 + (isRoot ? 84 : 72))); return { width, height: isRoot ? ROOT_NODE_HEIGHT : NODE_HEIGHT }; };
  const componentSeed = (ids: number[]) => ids.reduce((acc, val, idx) => acc + val * (idx + 1), 0);
  const nodeById = (id: number | null) => (id == null ? null : nodeLookup.get(id) ?? null);

  const relatedNodeIds = (nodeId: number | null) => {
    const nextIds = new Set<number>();
    if (nodeId == null) return nextIds;
    edges.forEach((edge) => {
      if (edge.source === nodeId) nextIds.add(edge.target);
      if (edge.target === nodeId) nextIds.add(edge.source);
    });
    return nextIds;
  };

  const searchResultsFromNodes = (query: string) => {
    const normalized = normalizeText(query);
    if (!normalized) return [];
    return nodes.filter((node) => normalizeText(node.name).includes(normalized));
  };

  const connectedComponents = (projectEntities: utils.Entity[], adjacency: Map<number, number[]>, degreeMap: Map<number, number>) => {
    const visited = new Set<number>();
    const orderedEntities = [...projectEntities].sort((l, r) => (degreeMap.get(r.Id) ?? 0) - (degreeMap.get(l.Id) ?? 0) || l.Name.localeCompare(r.Name));
    const components: utils.Entity[][] = [];
    for (const entity of orderedEntities) {
      if (visited.has(entity.Id)) continue;
      const stack = [entity.Id];
      const componentIds: number[] = [];
      visited.add(entity.Id);
      while (stack.length) {
        const current = stack.pop();
        if (current == null) continue;
        componentIds.push(current);
        (adjacency.get(current) || []).forEach((nextId) => {
          if (!visited.has(nextId)) { visited.add(nextId); stack.push(nextId); }
        });
      }
      components.push(componentIds.map((id) => projectEntities.find((c) => c.Id === id)).filter(Boolean) as utils.Entity[]);
    }
    return components;
  };

  const layoutComponent = (componentEntities: utils.Entity[], componentIndex: number, adjacency: Map<number, number[]>, degreeMap: Map<number, number>): ComponentLayout => {
    const root = [...componentEntities].sort((l, r) => (degreeMap.get(r.Id) ?? 0) - (degreeMap.get(l.Id) ?? 0) || l.Name.localeCompare(r.Name))[0];
    const depthMap = new Map<number, number>();
    const bfsQueue = [root.Id];
    depthMap.set(root.Id, 0);
    while (bfsQueue.length) {
      const current = bfsQueue.shift();
      if (current == null) continue;
      const currentDepth = depthMap.get(current) ?? 0;
      (adjacency.get(current) || []).sort((l, r) => (degreeMap.get(r) ?? 0) - (degreeMap.get(l) ?? 0) || `${l}`.localeCompare(`${r}`)).forEach((nextId) => {
        if (!depthMap.has(nextId)) { depthMap.set(nextId, currentDepth + 1); bfsQueue.push(nextId); }
      });
    }
    componentEntities.forEach((entity) => { if (!depthMap.has(entity.Id)) depthMap.set(entity.Id, 0); });
    const groups = new Map<number, utils.Entity[]>();
    componentEntities.forEach((entity) => { const depth = depthMap.get(entity.Id) ?? 0; const group = groups.get(depth) || []; group.push(entity); groups.set(depth, group); });
    const seed = componentSeed(componentEntities.map((e) => e.Id));
    const componentDensity = componentEntities.reduce((acc, e) => acc + (degreeMap.get(e.Id) ?? 0), 0) / Math.max(componentEntities.length, 1);
    const densityBoost = clamp(componentDensity * 6.2, 0, 72);
    const sizeBoost = clamp(componentEntities.length * 5.4, 0, 84);
    const orderedDepths = Array.from(groups.keys()).sort((l, r) => l - r);
    const placedNodes: GraphNode[] = [];
    orderedDepths.forEach((depth) => {
      const ringEntities = (groups.get(depth) || []).sort((l, r) => (degreeMap.get(r.Id) ?? 0) - (degreeMap.get(l.Id) ?? 0) || l.Name.localeCompare(r.Name));
      if (depth === 0) {
        const size = nodeDimensions(root.Name, true);
        placedNodes.push({ id: root.Id, name: root.Name, status: root.Status === true, degree: degreeMap.get(root.Id) ?? 0, width: size.width, height: size.height, x: -size.width / 2, y: -size.height / 2, component: componentIndex, depth });
        return;
      }
      const ringCount = ringEntities.length;
      const angleSpan = ringCount <= 2 ? Math.PI * 1.08 : ringCount <= 5 ? Math.PI * 1.52 : Math.min(Math.PI * 2.12, Math.PI * (1.72 + ringCount * 0.072));
      const angleOffset = (seed % 17) * 0.09 + depth * 0.35 + componentIndex * 0.24;
      const startAngle = angleOffset - angleSpan / 2;
      const step = ringCount === 1 ? 0 : angleSpan / (ringCount - 1);
      const ringSpreadBoost = Math.max(0, ringCount - 3);
      const radiusX = (RING_STEP_X + densityBoost + sizeBoost * 0.96) * depth + ringSpreadBoost * 24;
      const radiusY = (RING_STEP_Y + densityBoost * 0.84 + sizeBoost * 0.78) * depth + ringSpreadBoost * 18;
      ringEntities.forEach((entity, index) => {
        const angle = startAngle + step * index;
        const jitter = ((entity.Id % 9) - 4) * (7 + depth * 1.4);
        const centerX = Math.cos(angle) * (radiusX + jitter);
        const centerY = Math.sin(angle) * (radiusY + jitter * 0.4);
        const size = nodeDimensions(entity.Name, false);
        placedNodes.push({ id: entity.Id, name: entity.Name, status: entity.Status === true, degree: degreeMap.get(entity.Id) ?? 0, width: size.width, height: size.height, x: centerX - size.width / 2, y: centerY - size.height / 2, component: componentIndex, depth });
      });
    });
    const minX = Math.min(...placedNodes.map((n) => n.x));
    const maxX = Math.max(...placedNodes.map((n) => n.x + n.width));
    const minY = Math.min(...placedNodes.map((n) => n.y));
    const maxY = Math.max(...placedNodes.map((n) => n.y + n.height));
    return { nodes: placedNodes.map((n) => ({ ...n, x: n.x - minX, y: n.y - minY })), width: Math.max(420, maxX - minX), height: Math.max(320, maxY - minY) };
  };

  const buildGraph = (projectEntities: utils.Entity[], combinatory: utils.RelationView[]) => {
    const edgesMap = new Map<string, GraphEdge>();
    combinatory.forEach((view) => {
      (view.Relations || []).forEach((relation) => {
        if (!relation.Relation || relation.IdEntity2 === view.IdPrincipalEntity) return;
        const key = edgeKey(view.IdPrincipalEntity, relation.IdEntity2);
        if (!edgesMap.has(key)) {
          edgesMap.set(key, { key, source: view.IdPrincipalEntity, target: relation.IdEntity2, relation: relation.Relation, sourceName: view.PrincipalEntity, targetName: relation.Entity2 });
        }
      });
    });
    const graphEdges = Array.from(edgesMap.values()).sort((l, r) => l.relation.localeCompare(r.relation) || l.sourceName.localeCompare(r.sourceName));
    const degreeMap = degreeMapFromEdges(graphEdges);
    const adjacency = new Map<number, number[]>();
    projectEntities.forEach((e) => adjacency.set(e.Id, []));
    graphEdges.forEach((e) => { adjacency.get(e.source)?.push(e.target); adjacency.get(e.target)?.push(e.source); });
    const layouts = connectedComponents(projectEntities, adjacency, degreeMap).map((ent, idx) => layoutComponent(ent, idx, adjacency, degreeMap)).sort((l, r) => r.nodes.length - l.nodes.length || r.width - l.width);
    const placedNodes: GraphNode[] = [];
    let cursorX = CANVAS_PADDING, cursorY = CANVAS_PADDING, rowHeight = 0, maxWidth = 0;
    layouts.forEach((layout) => {
      if (cursorX > CANVAS_PADDING && cursorX + layout.width > MAX_CANVAS_ROW_WIDTH) { cursorX = CANVAS_PADDING; cursorY += rowHeight + COMPONENT_GAP_Y + clamp(rowHeight * 0.12, 34, 140); rowHeight = 0; }
      layout.nodes.forEach((node) => { placedNodes.push({ ...node, x: node.x + cursorX, y: node.y + cursorY }); });
      cursorX += layout.width + COMPONENT_GAP_X + clamp(layout.nodes.length * 7.2 + layout.width * 0.045, 36, 180);
      rowHeight = Math.max(rowHeight, layout.height); maxWidth = Math.max(maxWidth, cursorX);
    });
    return { nodes: placedNodes, edges: graphEdges, boardWidth: Math.max(maxWidth + CANVAS_PADDING, 1180), boardHeight: Math.max(cursorY + rowHeight + CANVAS_PADDING + clamp(rowHeight * 0.06, 0, 84), 760) };
  };

  const applyCamera = (nextPanX: number, nextPanY: number, nextZoom: number, animate = true) => {
    panX = nextPanX; panY = nextPanY; zoom = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, nextZoom)); cameraAnimating = animate;
    if (animationTimeout !== null) { window.clearTimeout(animationTimeout); animationTimeout = null; }
    if (animate) { animationTimeout = window.setTimeout(() => { cameraAnimating = false; }, 320); }
  };

  const fitBoard = (animate = true) => {
    if (!viewportEl || !boardWidth || !boardHeight) return;
    viewportWidth = viewportEl.clientWidth; viewportHeight = viewportEl.clientHeight;
    const availableWidth = Math.max(viewportWidth - 44, 320); const availableHeight = Math.max(viewportHeight - 44, 260);
    const scaleFactor = Math.min(availableWidth / boardWidth, availableHeight / boardHeight, 1);
    const nextZoom = Math.min(1.02, Math.max(ZOOM_MIN, scaleFactor));
    applyCamera((viewportWidth - boardWidth * nextZoom) / 2, (viewportHeight - boardHeight * nextZoom) / 2, nextZoom, animate);
    userHasMovedCamera = false;
  };

  const focusNode = (nodeId: number, recenter = false) => {
    selectedId = nodeId; if (!recenter || !viewportEl) return;
    const node = nodeById(nodeId); if (!node) return;
    viewportWidth = viewportEl.clientWidth; viewportHeight = viewportEl.clientHeight;
    applyCamera(viewportWidth / 2 - (node.x + node.width / 2) * zoom, viewportHeight / 2 - (node.y + node.height / 2) * zoom, zoom, true);
    userHasMovedCamera = true;
  };

  const handleStageClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (target.closest(".graph-node") || target.closest(".graph-sidepanel")) return;
    selectedId = null;
  };

  const handleStageMouseDown = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (target.closest(".graph-node") || target.closest(".graph-sidepanel") || event.button !== 0) return;
    isStageDragging = true; dragStartX = event.clientX; dragStartY = event.clientY; dragOriginPanX = panX; dragOriginPanY = panY; cameraAnimating = false;
  };

  const handleWindowMouseMove = (event: MouseEvent) => { if (isStageDragging) { panX = dragOriginPanX + (event.clientX - dragStartX); panY = dragOriginPanY + (event.clientY - dragStartY); userHasMovedCamera = true; } };
  const stopDragging = () => { isStageDragging = false; };
  const handleNodeKeydown = (nodeId: number, event: KeyboardEvent) => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); focusNode(nodeId, true); } };
  const handleStageKeydown = (event: KeyboardEvent) => { if (event.key === "Escape") { selectedId = null; } else if (event.key.toLowerCase() === "f" || event.key === "0") { fitBoard(true); } };
  const handleWindowResize = () => { if (!viewportEl) return; viewportWidth = viewportEl.clientWidth; viewportHeight = viewportEl.clientHeight; if (!userHasMovedCamera) fitBoard(false); };
  const handleWheel = (event: WheelEvent) => {
    if (!viewportEl) return;
    const rect = viewportEl.getBoundingClientRect(); const pointerX = event.clientX - rect.left, pointerY = event.clientY - rect.top;
    const worldX = (pointerX - panX) / zoom, worldY = (pointerY - panY) / zoom;
    const factor = event.deltaY > 0 ? 0.92 : 1.08; const nextZoom = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, zoom * factor));
    applyCamera(pointerX - worldX * nextZoom, pointerY - worldY * nextZoom, nextZoom, false);
    userHasMovedCamera = true;
  };

  const centerCameraOnPoint = (targetX: number, targetY: number) => {
    if (!viewportEl) return;
    viewportWidth = viewportEl.clientWidth; viewportHeight = viewportEl.clientHeight;
    applyCamera(viewportWidth / 2 - targetX * zoom, viewportHeight / 2 - targetY * zoom, zoom, true);
    userHasMovedCamera = true;
  };

  const handleMinimapClick = (event: MouseEvent) => {
    if (!minimapEl) return;
    const rect = minimapEl.getBoundingClientRect();
    centerCameraOnPoint(boardWidth * ((event.clientX - rect.left) / rect.width), boardHeight * ((event.clientY - rect.top) / rect.height));
  };
  const handleMinimapKeydown = (event: KeyboardEvent) => { if (event.key === "Enter" || event.key === " ") { event.preventDefault(); centerCameraOnPoint(boardWidth / 2, boardHeight / 2); } };

  const edgeControlPoint = (edge: GraphEdge) => {
    const source = nodeById(edge.source), target = nodeById(edge.target); if (!source || !target) return {x: 0, y: 0};
    const startX = source.x + source.width / 2, startY = source.y + source.height / 2, endX = target.x + target.width / 2, endY = target.y + target.height / 2;
    const dx = endX - startX, dy = endY - startY, dist = Math.max(1, Math.hypot(dx, dy));
    const bend = Math.min(108, Math.max(24, dist * 0.18));
    return { x: (startX + endX) / 2 + (-dy / dist) * bend, y: (startY + endY) / 2 + (dx / dist) * bend };
  };

  const connectionPath = (edge: GraphEdge) => { const source = nodeById(edge.source), target = nodeById(edge.target); if (!source || !target) return ""; const sX = source.x + source.width / 2, sY = source.y + source.height / 2, eX = target.x + target.width / 2, eY = target.y + target.height / 2, c = edgeControlPoint(edge); return `M ${sX} ${sY} Q ${c.x} ${c.y} ${eX} ${eY}`; };
  const edgeMidpoint = (edge: GraphEdge) => { const source = nodeById(edge.source), target = nodeById(edge.target); if (!source || !target) return {x: 0, y: 0}; const sX = source.x + source.width / 2, sY = source.y + source.height / 2, eX = target.x + target.width / 2, eY = target.y + target.height / 2, c = edgeControlPoint(edge); return { x: 0.25 * sX + 0.5 * c.x + 0.25 * eX, y: 0.25 * sY + 0.5 * c.y + 0.25 * eY }; };

  const applyGraphLayout = (graph: {nodes: GraphNode[]; edges: GraphEdge[]; boardWidth: number; boardHeight: number}, recenter = true) => {
    nodes = graph.nodes; edges = graph.edges; boardWidth = graph.boardWidth; boardHeight = graph.boardHeight;
    if (selectedId === null || !graph.nodes.some((n) => n.id === selectedId)) selectedId = graph.nodes[0]?.id ?? null;
    requestAnimationFrame(() => { if (recenter) { fitBoard(true); } else if (viewportEl && !userHasMovedCamera) { fitBoard(false); } });
  };

  const relayoutGraph = (recenter = true) => applyGraphLayout(buildGraph(entities, relationViews || []), recenter);
  const loadGraph = async () => { loading = true; try { relationViews = await GetCombinatory(); relayoutGraph(true); } catch (err) { showToast(`No se pudo construir el diagrama: ${extractError(err)}`, "error"); } finally { loading = false; } };

  $effect(() => {
    const nextSignature = entitySignature();
    if (nextSignature !== loadSignature) {
      loadSignature = nextSignature;
      if (entities.length === 0) { nodes = []; edges = []; selectedId = null; loading = false; }
      else { loadGraph(); }
    }
  });

  $effect(() => {
    if (viewportEl && !loading && !userHasMovedCamera) {
      requestAnimationFrame(() => {
        if (viewportEl && !userHasMovedCamera) { viewportWidth = viewportEl.clientWidth; viewportHeight = viewportEl.clientHeight; fitBoard(false); }
      });
    }
  });

  onMount(() => () => { if (animationTimeout !== null) window.clearTimeout(animationTimeout); });
</script>

<svelte:window onmousemove={handleWindowMouseMove} onmouseup={stopDragging} onmouseleave={stopDragging} onresize={handleWindowResize}/>

<section class="graph-tab">
  <div class="graph-toolbar">
    <div>
      <p class="label">Diagrama vivo</p>
      <h3>Mapa relacional de arrastre libre</h3>
      <p class="muted">Layout automático para arrancar rápido, drag manual para acomodar, y una vista mucho más limpia: solo nombres y conexiones.</p>
    </div>

    <div class="toolbar-controls">
      <label class="search-field">
        <span>Buscar entidad</span>
        <input
          type="search"
          bind:value={searchQuery}
          placeholder="Cliente, Factura, Producto..."
          aria-label="Buscar entidad en el diagrama"
        />
      </label>

      <button class="control control--soft" type="button" onclick={() => relayoutGraph(true)}>
        <ButtonIcon name="layout"/>
        <span>Reordenar</span>
      </button>
      <button class="control control--ghost" type="button" onclick={() => fitBoard(true)}>
        <ButtonIcon name="frame"/>
        <span>Reencuadrar</span>
      </button>
    </div>
  </div>

  {#if loading}
    <div class="empty-panel">Construyendo el diagrama...</div>
  {:else if !nodes.length}
    <div class="empty-panel">No hay entidades suficientes para dibujar el mapa.</div>
  {:else}
    <div class="graph-shell">
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div
        class="graph-stage"
        bind:this={viewportEl}
        aria-label="Diagrama relacional"
        onmousedown={handleStageMouseDown}
        onclick={handleStageClick}
        onwheel={(e) => { e.preventDefault(); handleWheel(e); }}
        onkeydown={handleStageKeydown}
      >
        <div class:graph-camera={true} class:graph-camera--smooth={cameraAnimating && !isStageDragging} style={`transform: translate(${panX}px, ${panY}px) scale(${zoom}); width: ${boardWidth}px; height: ${boardHeight}px;`}>
          <svg class="edge-layer" viewBox={`0 0 ${boardWidth} ${boardHeight}`} aria-hidden="true">
            {#each edges as edge}
              <g class:edge-group={true} class:edge-group--active={currentFocusId !== null && (edge.source === currentFocusId || edge.target === currentFocusId)}>
                <path
                  class:edge={true}
                  class:edge--active={currentFocusId !== null && (edge.source === currentFocusId || edge.target === currentFocusId)}
                  class:edge--muted={currentFocusId !== null && edge.source !== currentFocusId && edge.target !== currentFocusId}
                  d={connectionPath(edge)}
                  stroke={relationMeta(edge.relation).line}
                />
                <circle
                  cx={edgeMidpoint(edge).x}
                  cy={edgeMidpoint(edge).y}
                  r="14"
                  fill="var(--background)"
                  stroke={relationMeta(edge.relation).line}
                  stroke-width="1.5"
                />
                <text
                  x={edgeMidpoint(edge).x}
                  y={edgeMidpoint(edge).y}
                  dy="0.32em"
                  text-anchor="middle"
                  class="edge-label"
                  fill={relationMeta(edge.relation).line}
                >
                  {edge.relation}
                </text>
              </g>
            {/each}
          </svg>

          <div class="node-layer">
            {#each nodes as node (node.id)}
              <button
                type="button"
                class:graph-node={true}
                class:graph-node--root={node.depth === 0}
                class:graph-node--active={selectedId === node.id}
                class:graph-node--hover={hoverId === node.id}
                class:graph-node--related={connectedIds.has(node.id)}
                class:graph-node--muted={currentFocusId !== null && selectedId !== node.id && hoverId !== node.id && !connectedIds.has(node.id)}
                class:graph-node--inactive={!node.status}
                style={`left: ${node.x}px; top: ${node.y}px; width: ${node.width}px; height: ${node.height}px; view-transition-name: graph-node-${node.id};`}
                onclick={() => focusNode(node.id, true)}
                onmouseenter={() => { hoverId = node.id; }}
                onmouseleave={() => { hoverId = null; }}
                onkeydown={(event) => handleNodeKeydown(node.id, event)}
              >
                <div class="graph-node__content">
                  <div class="graph-node__icon">
                    <ButtonIcon name={node.depth === 0 ? "home" : "entities"}/>
                  </div>
                  <div class="graph-node__copy">
                    <strong>{node.name}</strong>
                    <span class="graph-node__meta">Grado: {node.degree}</span>
                  </div>
                </div>
                {#if !node.status}
                  <div class="graph-node__status" title="Tabla inactiva">
                    <ButtonIcon name="clear"/>
                  </div>
                {/if}
              </button>
            {/each}
          </div>
        </div>
      </div>

      <aside class:graph-sidepanel={true} class:graph-sidepanel--visible={selectedId !== null}>
        {#if selectedNode}
          <div class="sidepanel-head">
            <div class="sidepanel-kicker">
              <ButtonIcon name="entities"/>
              <span>Detalle de entidad</span>
            </div>
            <h2>{selectedNode.name}</h2>
            <p class="sidepanel-desc">{entityDescription(selectedNode.id)}</p>
          </div>

          <div class="sidepanel-stats">
            <div class="sidepanel-stat">
              <span>Conexiones</span>
              <strong>{selectedNode.degree}</strong>
            </div>
            <div class="sidepanel-stat">
              <span>Estado</span>
              <strong class={selectedNode.status ? "text-success" : "text-danger"}>
                {selectedNode.status ? "Activa" : "Inactiva"}
              </strong>
            </div>
          </div>

          <div class="sidepanel-section">
            <header>
              <h4>Tablas relacionadas</h4>
              <span>{connectedIds.size} encontradas</span>
            </header>
            <div class="related-list">
              {#each edges.filter((e) => e.source === selectedId || e.target === selectedId) as edge}
                {@const otherId = edge.source === selectedId ? edge.target : edge.source}
                {@const otherName = edge.source === selectedId ? edge.targetName : edge.sourceName}
                <button type="button" class="related-item" onclick={() => focusNode(otherId, true)}>
                  <div class="related-item__copy">
                    <strong>{otherName}</strong>
                    <span>{relationMeta(edge.relation).label}</span>
                  </div>
                  <span class="related-item__type" style={`color: ${relationMeta(edge.relation).line}; background: ${relationMeta(edge.relation).fill};`}>
                    {edge.relation}
                  </span>
                </button>
              {:else}
                <p class="empty-hint">Esta tabla no tiene relaciones directas.</p>
              {/each}
            </div>
          </div>

          <div class="sidepanel-footer">
            <button class="jump-btn" type="button" onclick={() => onJumpTo("entities", selectedId)}>
              <span>Abrir configurador</span>
              <ButtonIcon name="chevron-right"/>
            </button>
          </div>
        {/if}
      </aside>

      <div class="graph-overlay graph-overlay--left">
        <div class="legend-panel">
          <h4>Leyenda</h4>
          {#each relationLegend as [relation]}
            <div class="legend-item">
              <span class="legend-swatch" style={`background: ${relationPalette[relation]?.line ?? relationPalette["N:N"].line};`}></span>
              <span>{relationPalette[relation]?.label ?? relation}</span>
            </div>
          {/each}
        </div>
      </div>

      <div class="graph-overlay graph-overlay--right">
        <div class="viewport-meta">
          <span>Zoom: {Math.round(zoom * 100)}%</span>
          <span class="separator">/</span>
          <span>{nodes.length} Nodos</span>
          <span class="separator">/</span>
          <span>{edges.length} Ligas</span>
        </div>

        <div class="minimap-container">
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <svg
            bind:this={minimapEl}
            class="minimap"
            viewBox={`0 0 ${boardWidth} ${boardHeight}`}
            style={`aspect-ratio: ${1 / minimapAspect};`}
            onclick={handleMinimapClick}
            onkeydown={handleMinimapKeydown}
          >
            {#each nodes as node}
              <rect
                x={node.x}
                y={node.y}
                width={node.width}
                height={node.height}
                rx="4"
                fill={selectedId === node.id ? "var(--accent)" : "var(--ink-faint)"}
                opacity={selectedId === node.id || connectedIds.has(node.id) ? "1" : "0.5"}
              />
            {/each}
            <rect
              x={viewportRect.x}
              y={viewportRect.y}
              width={viewportRect.width}
              height={viewportRect.height}
              fill="none"
              stroke="var(--accent)"
              stroke-width={Math.max(4, 32 / zoom)}
              opacity="0.8"
            />
          </svg>
        </div>
      </div>

      {#if searchQuery && searchResults.length}
        <div class="search-results-floating">
          {#each searchResults as node}
            <button type="button" onclick={() => { focusNode(node.id, true); searchQuery = ""; }}>
              <ButtonIcon name="entities"/>
              <span>{node.name}</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</section>

<style>
  .graph-tab {
    display: flex;
    flex-direction: column;
    height: 100%;
    gap: 1.5rem;
  }

  .graph-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 2rem;
  }

  .graph-toolbar h3 {
    margin: 0.25rem 0 0;
    font-size: 1.5rem;
  }

  .toolbar-controls {
    display: flex;
    align-items: flex-end;
    gap: 0.75rem;
  }

  .search-field {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .search-field span {
    font-size: 0.65rem;
    font-weight: 800;
    text-transform: uppercase;
    color: var(--ink-soft);
    letter-spacing: 0.05em;
  }

  .search-field input {
    height: 2.75rem;
    width: 16rem;
    background: var(--muted-soft);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    padding: 0 1rem;
    font-size: 0.9rem;
    transition: all 0.2s ease;
  }

  .search-field input:focus {
    background: var(--background);
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-ghost);
    outline: none;
  }

  .control {
    height: 2.75rem;
    padding: 0 1.25rem;
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    gap: 0.6rem;
    font-size: 0.85rem;
    font-weight: 700;
    transition: all 0.2s ease;
    cursor: pointer;
  }

  .control--soft {
    background: var(--accent-ghost);
    color: var(--accent);
    border: 1px solid var(--accent-soft);
  }

  .control--soft:hover {
    background: var(--accent-soft);
  }

  .control--ghost {
    background: transparent;
    color: var(--ink-soft);
    border: 1px solid var(--border-card);
  }

  .control--ghost:hover {
    background: var(--muted-soft);
    color: var(--ink);
  }

  .graph-shell {
    flex: 1;
    position: relative;
    background: var(--muted-ghost);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg);
    overflow: hidden;
    min-height: 500px;
  }

  .graph-stage {
    position: absolute;
    inset: 0;
    cursor: grab;
    overflow: hidden;
    outline: none;
  }

  .graph-stage:active {
    cursor: grabbing;
  }

  .graph-camera {
    position: absolute;
    transform-origin: 0 0;
    will-change: transform;
  }

  .graph-camera--smooth {
    transition: transform 0.32s cubic-bezier(0.19, 1, 0.22, 1);
  }

  .edge-layer {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    overflow: visible;
  }

  .edge {
    fill: none;
    stroke-width: 1.5;
    transition: all 0.3s ease;
  }

  .edge--active {
    stroke-width: 3;
    filter: drop-shadow(0 0 4px currentColor);
  }

  .edge--muted {
    opacity: 0.1;
  }

  .edge-label {
    font-size: 10px;
    font-weight: 900;
    pointer-events: none;
  }

  .node-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }

  .graph-node {
    position: absolute;
    background: var(--background);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: 0 1rem;
    cursor: pointer;
    box-shadow: var(--shadow-sm);
    transition: all 0.3s cubic-bezier(0.19, 1, 0.22, 1);
    pointer-events: auto;
    text-align: left;
  }

  .graph-node:hover {
    transform: translateY(-4px) scale(1.02);
    box-shadow: var(--shadow-md);
    border-color: var(--accent);
    z-index: 10;
  }

  .graph-node--root {
    border-width: 2px;
    border-color: var(--accent-soft);
    background: color-mix(in srgb, var(--accent-ghost) 10%, var(--background));
  }

  .graph-node--active {
    border-color: var(--accent);
    box-shadow: 0 0 0 3px var(--accent-ghost), var(--shadow-lg);
    z-index: 20;
    transform: scale(1.05);
  }

  .graph-node--muted {
    opacity: 0.3;
    filter: grayscale(0.8);
  }

  .graph-node--related {
    opacity: 1;
    filter: none;
    border-color: var(--accent-soft);
  }

  .graph-node--inactive { opacity: 0.6; }
  .graph-node__content { display: flex; align-items: center; gap: 0.75rem; overflow: hidden; }
  .graph-node__icon { width: 2rem; height: 2rem; display: flex; align-items: center; justify-content: center; background: var(--muted-soft); color: var(--accent); border-radius: 99px; flex-shrink: 0; }
  .graph-node__copy { display: flex; flex-direction: column; gap: 0.1rem; overflow: hidden; }
  .graph-node__copy strong { font-size: 0.85rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .graph-node__meta { font-size: 0.65rem; color: var(--ink-soft); font-weight: 700; }
  .graph-node__status { position: absolute; top: -0.5rem; right: -0.5rem; width: 1.25rem; height: 1.25rem; background: var(--danger); color: white; border-radius: 99px; display: flex; align-items: center; justify-content: center; padding: 0.2rem; box-shadow: var(--shadow-sm); }

  .graph-sidepanel { position: absolute; top: 1rem; left: 1rem; bottom: 1rem; width: 22rem; background: var(--background); border: 1px solid var(--border-card); border-radius: var(--radius-lg); box-shadow: var(--shadow-lg); display: flex; flex-direction: column; z-index: 100; transform: translateX(-120%); transition: transform 0.4s cubic-bezier(0.19, 1, 0.22, 1); }
  .graph-sidepanel--visible { transform: translateX(0); }
  .sidepanel-head { padding: 1.75rem; border-bottom: 1px solid var(--border-card); background: var(--muted-ghost); border-radius: var(--radius-lg) var(--radius-lg) 0 0; }
  .sidepanel-kicker { display: flex; align-items: center; gap: 0.5rem; font-size: 0.65rem; font-weight: 900; text-transform: uppercase; color: var(--accent); margin-bottom: 0.75rem; }
  .sidepanel-head h2 { font-size: 1.5rem; margin: 0; }
  .sidepanel-desc { margin: 0.75rem 0 0; font-size: 0.85rem; color: var(--ink-soft); line-height: 1.5; }
  .sidepanel-stats { display: flex; padding: 1.25rem 1.75rem; border-bottom: 1px solid var(--border-card); gap: 2rem; }
  .sidepanel-stat { display: flex; flex-direction: column; gap: 0.25rem; }
  .sidepanel-stat span { font-size: 0.6rem; font-weight: 800; text-transform: uppercase; color: var(--ink-faint); }
  .sidepanel-stat strong { font-size: 1.1rem; }
  .sidepanel-section { flex: 1; overflow-y: auto; padding: 1.75rem; }
  .sidepanel-section header { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 1rem; }
  .sidepanel-section h4 { margin: 0; font-size: 0.9rem; }
  .sidepanel-section header span { font-size: 0.7rem; color: var(--ink-soft); font-weight: 700; }
  .related-list { display: flex; flex-direction: column; gap: 0.6rem; }
  .related-item { width: 100%; display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 1rem; background: var(--muted-ghost); border: 1px solid var(--border-card); border-radius: var(--radius-md); text-align: left; transition: all 0.2s ease; cursor: pointer; }
  .related-item:hover { transform: translateX(4px); border-color: var(--accent-soft); background: var(--accent-ghost); }
  .related-item__copy strong { display: block; font-size: 0.85rem; }
  .related-item__copy span { font-size: 0.65rem; color: var(--ink-soft); }
  .related-item__type { font-size: 0.6rem; font-weight: 900; padding: 0.15rem 0.4rem; border-radius: 4px; }
  .sidepanel-footer { padding: 1.5rem; border-top: 1px solid var(--border-card); }
  .jump-btn { width: 100%; height: 3rem; background: var(--accent); color: white; border-radius: var(--radius-md); display: flex; align-items: center; justify-content: center; gap: 0.75rem; font-weight: 700; font-size: 0.9rem; cursor: pointer; transition: all 0.2s ease; }
  .jump-btn:hover { background: var(--accent-strong); transform: translateY(-2px); box-shadow: var(--shadow-md); }

  .graph-overlay { position: absolute; display: flex; flex-direction: column; gap: 0.75rem; pointer-events: none; }
  .graph-overlay--left { bottom: 1.5rem; left: 1.5rem; z-index: 50; }
  .graph-overlay--right { bottom: 1.5rem; right: 1.5rem; align-items: flex-end; z-index: 50; }
  .legend-panel { background: var(--background); border: 1px solid var(--border-card); border-radius: var(--radius-md); padding: 1rem; box-shadow: var(--shadow-md); pointer-events: auto; }
  .legend-panel h4 { margin: 0 0 0.75rem; font-size: 0.7rem; text-transform: uppercase; color: var(--ink-faint); }
  .legend-item { display: flex; align-items: center; gap: 0.6rem; font-size: 0.75rem; font-weight: 700; margin-bottom: 0.5rem; }
  .legend-swatch { width: 0.75rem; height: 0.75rem; border-radius: 2px; }
  .viewport-meta { background: var(--ink); color: white; padding: 0.4rem 1rem; border-radius: 99px; font-size: 0.65rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em; display: flex; align-items: center; gap: 0.6rem; box-shadow: var(--shadow-md); }
  .separator { opacity: 0.3; }
  .minimap-container { width: 12rem; background: var(--background); border: 1px solid var(--border-card); border-radius: var(--radius-md); overflow: hidden; box-shadow: var(--shadow-md); pointer-events: auto; }
  .minimap { width: 100%; height: auto; cursor: crosshair; }
  .empty-panel { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; font-size: 0.95rem; color: var(--ink-soft); background: var(--muted-ghost); }
  .search-results-floating { position: absolute; top: 1rem; right: 1rem; width: 16rem; background: var(--background); border: 1px solid var(--border-card); border-radius: var(--radius-md); box-shadow: var(--shadow-lg); z-index: 1000; display: flex; flex-direction: column; overflow: hidden; }
  .search-results-floating button { width: 100%; padding: 0.75rem 1rem; display: flex; align-items: center; gap: 0.75rem; font-size: 0.85rem; font-weight: 700; text-align: left; border-bottom: 1px solid var(--border-card); transition: background 0.2s ease; }
  .search-results-floating button:hover { background: var(--muted-soft); }
</style>
