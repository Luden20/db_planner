<script lang="ts">
  import {onMount} from "svelte";
  import type {utils} from "../../wailsjs/go/models";
  import {GetCombinatory} from "../../wailsjs/go/main/App";
  import ButtonIcon from "./ButtonIcon.svelte";
  import {showToast} from "../lib/toast";

  type GraphNode = {
    id: number;
    name: string;
    status: boolean;
    degree: number;
    width: number;
    height: number;
    x: number;
    y: number;
    component: number;
    depth: number;
  };

  type GraphEdge = {
    key: string;
    source: number;
    target: number;
    relation: string;
    sourceName: string;
    targetName: string;
  };

  type ComponentLayout = {
    nodes: GraphNode[];
    width: number;
    height: number;
  };

  export let entities: utils.Entity[] = [];
  export let onJumpTo: (tab: "entities" | "relations" | "tertiary", entityId?: number | null) => void = () => {};

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

  let loading = true;
  let loadSignature = "";
  let relationViews: utils.RelationView[] = [];
  let nodes: GraphNode[] = [];
  let edges: GraphEdge[] = [];
  let boardWidth = 1320;
  let boardHeight = 900;
  let selectedId: number | null = null;
  let hoverId: number | null = null;
  let searchQuery = "";
  let zoom = 1;
  let panX = 0;
  let panY = 0;
  let viewportWidth = 0;
  let viewportHeight = 0;
  let viewportEl: HTMLDivElement | null = null;
  let minimapEl: SVGSVGElement | null = null;
  let isStageDragging = false;
  let dragStartX = 0;
  let dragStartY = 0;
  let dragOriginPanX = 0;
  let dragOriginPanY = 0;
  let userHasMovedCamera = false;
  let cameraAnimating = false;
  let animationTimeout: number | null = null;
  let nodeLookup = new Map<number, GraphNode>();

  const relationPalette: Record<string, {line: string; fill: string; label: string}> = {
    "1:1": {line: "var(--accent)", fill: "color-mix(in srgb, var(--accent) 16%, transparent)", label: "Uno a uno"},
    "1:N": {line: "var(--success)", fill: "color-mix(in srgb, var(--success) 16%, transparent)", label: "Uno a muchos"},
    "N:1": {line: "var(--danger)", fill: "color-mix(in srgb, var(--danger) 16%, transparent)", label: "Muchos a uno"},
    "N:N": {line: "var(--ink-soft)", fill: "color-mix(in srgb, var(--ink) 12%, transparent)", label: "Muchos a muchos"}
  };

  const normalizeText = (value: string) =>
    (value || "")
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .toLowerCase()
      .trim();

  const entitySignature = () =>
    entities
      .map((entity) => `${entity.Id}:${entity.Status ? 1 : 0}:${entity.Name}`)
      .join("|");

  const edgeKey = (a: number, b: number) => {
    const first = Math.min(a, b);
    const second = Math.max(a, b);
    return `${first}-${second}`;
  };

  const relationMeta = (relation: string) => relationPalette[relation] ?? relationPalette["N:N"];

  const degreeMapFromEdges = (graphEdges: GraphEdge[]) => {
    const degreeMap = new Map<number, number>();
    graphEdges.forEach((edge) => {
      degreeMap.set(edge.source, (degreeMap.get(edge.source) ?? 0) + 1);
      degreeMap.set(edge.target, (degreeMap.get(edge.target) ?? 0) + 1);
    });
    return degreeMap;
  };

  const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

  const nodeDimensions = (name: string, isRoot: boolean) => {
    const width = Math.min(
      isRoot ? 248 : 228,
      Math.max(isRoot ? ROOT_NODE_WIDTH : BASE_NODE_WIDTH, name.length * 8 + (isRoot ? 84 : 72))
    );
    return {
      width,
      height: isRoot ? ROOT_NODE_HEIGHT : NODE_HEIGHT
    };
  };

  const componentSeed = (ids: number[]) =>
    ids.reduce((accumulator, value, index) => accumulator + value * (index + 1), 0);

  const connectedComponents = (
    projectEntities: utils.Entity[],
    adjacency: Map<number, number[]>,
    degreeMap: Map<number, number>
  ) => {
    const visited = new Set<number>();
    const orderedEntities = [...projectEntities].sort((left, right) =>
      (degreeMap.get(right.Id) ?? 0) - (degreeMap.get(left.Id) ?? 0) || left.Name.localeCompare(right.Name)
    );
    const components: utils.Entity[][] = [];

    for (const entity of orderedEntities) {
      if (visited.has(entity.Id)) {
        continue;
      }

      const stack = [entity.Id];
      const componentIds: number[] = [];
      visited.add(entity.Id);

      while (stack.length) {
        const current = stack.pop();
        if (current == null) {
          continue;
        }
        componentIds.push(current);
        (adjacency.get(current) || []).forEach((nextId) => {
          if (!visited.has(nextId)) {
            visited.add(nextId);
            stack.push(nextId);
          }
        });
      }

      components.push(
        componentIds
          .map((id) => projectEntities.find((candidate) => candidate.Id === id))
          .filter(Boolean) as utils.Entity[]
      );
    }

    return components;
  };

  const layoutComponent = (
    componentEntities: utils.Entity[],
    componentIndex: number,
    adjacency: Map<number, number[]>,
    degreeMap: Map<number, number>
  ): ComponentLayout => {
    const root = [...componentEntities].sort((left, right) =>
      (degreeMap.get(right.Id) ?? 0) - (degreeMap.get(left.Id) ?? 0) || left.Name.localeCompare(right.Name)
    )[0];

    const depthMap = new Map<number, number>();
    const bfsQueue = [root.Id];
    depthMap.set(root.Id, 0);

    while (bfsQueue.length) {
      const current = bfsQueue.shift();
      if (current == null) {
        continue;
      }

      const currentDepth = depthMap.get(current) ?? 0;
      (adjacency.get(current) || [])
        .sort((left, right) =>
          (degreeMap.get(right) ?? 0) - (degreeMap.get(left) ?? 0)
          || `${left}`.localeCompare(`${right}`)
        )
        .forEach((nextId) => {
          if (!depthMap.has(nextId)) {
            depthMap.set(nextId, currentDepth + 1);
            bfsQueue.push(nextId);
          }
        });
    }

    componentEntities.forEach((entity) => {
      if (!depthMap.has(entity.Id)) {
        depthMap.set(entity.Id, 0);
      }
    });

    const groups = new Map<number, utils.Entity[]>();
    componentEntities.forEach((entity) => {
      const depth = depthMap.get(entity.Id) ?? 0;
      const group = groups.get(depth) || [];
      group.push(entity);
      groups.set(depth, group);
    });

    const seed = componentSeed(componentEntities.map((entity) => entity.Id));
    const componentDensity = componentEntities.reduce(
      (accumulator, entity) => accumulator + (degreeMap.get(entity.Id) ?? 0),
      0
    ) / Math.max(componentEntities.length, 1);
    const densityBoost = clamp(componentDensity * 6.2, 0, 72);
    const sizeBoost = clamp(componentEntities.length * 5.4, 0, 84);
    const orderedDepths = Array.from(groups.keys()).sort((left, right) => left - right);
    const placedNodes: GraphNode[] = [];

    orderedDepths.forEach((depth) => {
      const ringEntities = (groups.get(depth) || []).sort((left, right) =>
        (degreeMap.get(right.Id) ?? 0) - (degreeMap.get(left.Id) ?? 0) || left.Name.localeCompare(right.Name)
      );

      if (depth === 0) {
        const size = nodeDimensions(root.Name, true);
        placedNodes.push({
          id: root.Id,
          name: root.Name,
          status: root.Status === true,
          degree: degreeMap.get(root.Id) ?? 0,
          width: size.width,
          height: size.height,
          x: -size.width / 2,
          y: -size.height / 2,
          component: componentIndex,
          depth
        });
        return;
      }

      const ringCount = ringEntities.length;
      const angleSpan = ringCount <= 2
        ? Math.PI * 1.08
        : ringCount <= 5
          ? Math.PI * 1.52
          : Math.min(Math.PI * 2.12, Math.PI * (1.72 + ringCount * 0.072));
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

        placedNodes.push({
          id: entity.Id,
          name: entity.Name,
          status: entity.Status === true,
          degree: degreeMap.get(entity.Id) ?? 0,
          width: size.width,
          height: size.height,
          x: centerX - size.width / 2,
          y: centerY - size.height / 2,
          component: componentIndex,
          depth
        });
      });
    });

    const minX = Math.min(...placedNodes.map((node) => node.x));
    const maxX = Math.max(...placedNodes.map((node) => node.x + node.width));
    const minY = Math.min(...placedNodes.map((node) => node.y));
    const maxY = Math.max(...placedNodes.map((node) => node.y + node.height));

    return {
      nodes: placedNodes.map((node) => ({
        ...node,
        x: node.x - minX,
        y: node.y - minY
      })),
      width: Math.max(420, maxX - minX),
      height: Math.max(320, maxY - minY)
    };
  };

  const buildGraph = (projectEntities: utils.Entity[], combinatory: utils.RelationView[]) => {
    const edgesMap = new Map<string, GraphEdge>();

    combinatory.forEach((view) => {
      (view.Relations || []).forEach((relation) => {
        if (!relation.Relation || relation.IdEntity2 === view.IdPrincipalEntity) {
          return;
        }

        const key = edgeKey(view.IdPrincipalEntity, relation.IdEntity2);
        if (!edgesMap.has(key)) {
          edgesMap.set(key, {
            key,
            source: view.IdPrincipalEntity,
            target: relation.IdEntity2,
            relation: relation.Relation,
            sourceName: view.PrincipalEntity,
            targetName: relation.Entity2
          });
        }
      });
    });

    const graphEdges = Array.from(edgesMap.values()).sort((left, right) =>
      left.relation.localeCompare(right.relation) || left.sourceName.localeCompare(right.sourceName)
    );
    const degreeMap = degreeMapFromEdges(graphEdges);
    const adjacency = new Map<number, number[]>();

    projectEntities.forEach((entity) => adjacency.set(entity.Id, []));
    graphEdges.forEach((edge) => {
      adjacency.get(edge.source)?.push(edge.target);
      adjacency.get(edge.target)?.push(edge.source);
    });

    const layouts = connectedComponents(projectEntities, adjacency, degreeMap)
      .map((componentEntities, componentIndex) => layoutComponent(componentEntities, componentIndex, adjacency, degreeMap))
      .sort((left, right) => right.nodes.length - left.nodes.length || right.width - left.width);

    const placedNodes: GraphNode[] = [];
    let cursorX = CANVAS_PADDING;
    let cursorY = CANVAS_PADDING;
    let rowHeight = 0;
    let maxWidth = 0;

    layouts.forEach((layout) => {
      if (cursorX > CANVAS_PADDING && cursorX + layout.width > MAX_CANVAS_ROW_WIDTH) {
        const adaptiveGapY = COMPONENT_GAP_Y + clamp(rowHeight * 0.12, 34, 140);
        cursorX = CANVAS_PADDING;
        cursorY += rowHeight + adaptiveGapY;
        rowHeight = 0;
      }

      layout.nodes.forEach((node) => {
        placedNodes.push({
          ...node,
          x: node.x + cursorX,
          y: node.y + cursorY
        });
      });

      const adaptiveGapX = COMPONENT_GAP_X + clamp(layout.nodes.length * 7.2 + layout.width * 0.045, 36, 180);
      cursorX += layout.width + adaptiveGapX;
      rowHeight = Math.max(rowHeight, layout.height);
      maxWidth = Math.max(maxWidth, cursorX);
    });

    return {
      nodes: placedNodes,
      edges: graphEdges,
      boardWidth: Math.max(maxWidth + CANVAS_PADDING, 1180),
      boardHeight: Math.max(cursorY + rowHeight + CANVAS_PADDING + clamp(rowHeight * 0.06, 0, 84), 760)
    };
  };

  const nodeById = (id: number | null) => (id == null ? null : nodeLookup.get(id) ?? null);

  const relatedNodeIds = (nodeId: number | null) => {
    const nextIds = new Set<number>();
    if (nodeId == null) {
      return nextIds;
    }

    edges.forEach((edge) => {
      if (edge.source === nodeId) {
        nextIds.add(edge.target);
      }
      if (edge.target === nodeId) {
        nextIds.add(edge.source);
      }
    });

    return nextIds;
  };

  const searchResultsFromNodes = (query: string) => {
    const normalized = normalizeText(query);
    if (!normalized) {
      return [];
    }
    return nodes.filter((node) => normalizeText(node.name).includes(normalized));
  };

  const applyCamera = (nextPanX: number, nextPanY: number, nextZoom: number, animate = true) => {
    panX = nextPanX;
    panY = nextPanY;
    zoom = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, nextZoom));
    cameraAnimating = animate;

    if (animationTimeout !== null) {
      window.clearTimeout(animationTimeout);
      animationTimeout = null;
    }

    if (animate) {
      animationTimeout = window.setTimeout(() => {
        cameraAnimating = false;
      }, 320);
    }
  };

  const fitBoard = (animate = true) => {
    if (!viewportEl || !boardWidth || !boardHeight) {
      return;
    }

    viewportWidth = viewportEl.clientWidth;
    viewportHeight = viewportEl.clientHeight;

    const availableWidth = Math.max(viewportWidth - 44, 320);
    const availableHeight = Math.max(viewportHeight - 44, 260);
    const scale = Math.min(availableWidth / boardWidth, availableHeight / boardHeight, 1);
    const nextZoom = Math.min(1.02, Math.max(ZOOM_MIN, scale));
    const nextPanX = (viewportWidth - boardWidth * nextZoom) / 2;
    const nextPanY = (viewportHeight - boardHeight * nextZoom) / 2;
    applyCamera(nextPanX, nextPanY, nextZoom, animate);
    userHasMovedCamera = false;
  };

  const focusNode = (nodeId: number, recenter = false) => {
    selectedId = nodeId;
    if (!recenter || !viewportEl) {
      return;
    }

    const node = nodeById(nodeId);
    if (!node) {
      return;
    }

    viewportWidth = viewportEl.clientWidth;
    viewportHeight = viewportEl.clientHeight;
    const nextPanX = viewportWidth / 2 - (node.x + node.width / 2) * zoom;
    const nextPanY = viewportHeight / 2 - (node.y + node.height / 2) * zoom;
    applyCamera(nextPanX, nextPanY, zoom, true);
    userHasMovedCamera = true;
  };

  const resetHover = () => {
    hoverId = null;
  };

  const clearSelection = () => {
    selectedId = null;
  };

  const handleStageClick = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (target.closest(".graph-node") || target.closest(".graph-sidepanel")) {
      return;
    }
    clearSelection();
  };

  const handleStageMouseDown = (event: MouseEvent) => {
    const target = event.target as HTMLElement;
    if (target.closest(".graph-node") || target.closest(".graph-sidepanel") || event.button !== 0) {
      return;
    }

    isStageDragging = true;
    dragStartX = event.clientX;
    dragStartY = event.clientY;
    dragOriginPanX = panX;
    dragOriginPanY = panY;
    cameraAnimating = false;
  };

  const handleWindowMouseMove = (event: MouseEvent) => {
    if (!isStageDragging) {
      return;
    }

    panX = dragOriginPanX + (event.clientX - dragStartX);
    panY = dragOriginPanY + (event.clientY - dragStartY);
    userHasMovedCamera = true;
  };

  const stopDragging = () => {
    isStageDragging = false;
  };

  const handleNodeClick = (nodeId: number) => {
    focusNode(nodeId, true);
  };

  const handleNodeKeydown = (nodeId: number, event: KeyboardEvent) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      focusNode(nodeId, true);
    }
  };

  const handleStageKeydown = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      clearSelection();
      return;
    }

    if (event.key.toLowerCase() === "f") {
      fitBoard(true);
      return;
    }

    if (event.key === "0") {
      fitBoard(true);
    }
  };

  const handleWindowResize = () => {
    if (!viewportEl) {
      return;
    }
    viewportWidth = viewportEl.clientWidth;
    viewportHeight = viewportEl.clientHeight;
    if (!userHasMovedCamera) {
      fitBoard(false);
    }
  };

  const handleWheel = (event: WheelEvent) => {
    if (!viewportEl) {
      return;
    }

    viewportWidth = viewportEl.clientWidth;
    viewportHeight = viewportEl.clientHeight;
    const rect = viewportEl.getBoundingClientRect();
    const pointerX = event.clientX - rect.left;
    const pointerY = event.clientY - rect.top;
    const worldX = (pointerX - panX) / zoom;
    const worldY = (pointerY - panY) / zoom;
    const factor = event.deltaY > 0 ? 0.92 : 1.08;
    const nextZoom = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, zoom * factor));
    const nextPanX = pointerX - worldX * nextZoom;
    const nextPanY = pointerY - worldY * nextZoom;
    applyCamera(nextPanX, nextPanY, nextZoom, false);
    userHasMovedCamera = true;
  };

  const centerCameraOnPoint = (targetX: number, targetY: number) => {
    if (!viewportEl) {
      return;
    }
    viewportWidth = viewportEl.clientWidth;
    viewportHeight = viewportEl.clientHeight;
    const nextPanX = viewportWidth / 2 - targetX * zoom;
    const nextPanY = viewportHeight / 2 - targetY * zoom;
    applyCamera(nextPanX, nextPanY, zoom, true);
    userHasMovedCamera = true;
  };

  const handleMinimapClick = (event: MouseEvent) => {
    if (!minimapEl) {
      return;
    }
    const rect = minimapEl.getBoundingClientRect();
    const ratioX = (event.clientX - rect.left) / rect.width;
    const ratioY = (event.clientY - rect.top) / rect.height;
    centerCameraOnPoint(boardWidth * ratioX, boardHeight * ratioY);
  };

  const handleMinimapKeydown = (event: KeyboardEvent) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      centerCameraOnPoint(boardWidth / 2, boardHeight / 2);
    }
  };

  const edgeControlPoint = (edge: GraphEdge) => {
    const source = nodeById(edge.source);
    const target = nodeById(edge.target);
    if (!source || !target) {
      return {x: 0, y: 0};
    }

    const startX = source.x + source.width / 2;
    const startY = source.y + source.height / 2;
    const endX = target.x + target.width / 2;
    const endY = target.y + target.height / 2;
    const dx = endX - startX;
    const dy = endY - startY;
    const distance = Math.max(1, Math.hypot(dx, dy));
    const bend = Math.min(108, Math.max(24, distance * 0.18));
    const normalX = -dy / distance;
    const normalY = dx / distance;

    return {
      x: (startX + endX) / 2 + normalX * bend,
      y: (startY + endY) / 2 + normalY * bend
    };
  };

  const connectionPath = (edge: GraphEdge) => {
    const source = nodeById(edge.source);
    const target = nodeById(edge.target);
    if (!source || !target) {
      return "";
    }

    const startX = source.x + source.width / 2;
    const startY = source.y + source.height / 2;
    const endX = target.x + target.width / 2;
    const endY = target.y + target.height / 2;
    const control = edgeControlPoint(edge);

    return `M ${startX} ${startY} Q ${control.x} ${control.y} ${endX} ${endY}`;
  };

  const edgeMidpoint = (edge: GraphEdge) => {
    const source = nodeById(edge.source);
    const target = nodeById(edge.target);
    if (!source || !target) {
      return {x: 0, y: 0};
    }

    const startX = source.x + source.width / 2;
    const startY = source.y + source.height / 2;
    const endX = target.x + target.width / 2;
    const endY = target.y + target.height / 2;
    const control = edgeControlPoint(edge);

    return {
      x: 0.25 * startX + 0.5 * control.x + 0.25 * endX,
      y: 0.25 * startY + 0.5 * control.y + 0.25 * endY
    };
  };

  const applyGraphLayout = (graph: {nodes: GraphNode[]; edges: GraphEdge[]; boardWidth: number; boardHeight: number}, recenter = true) => {
    nodes = graph.nodes;
    edges = graph.edges;
    boardWidth = graph.boardWidth;
    boardHeight = graph.boardHeight;

    if (selectedId === null || !graph.nodes.some((node) => node.id === selectedId)) {
      selectedId = graph.nodes[0]?.id ?? null;
    }

    requestAnimationFrame(() => {
      if (recenter) {
        fitBoard(true);
      } else if (viewportEl && !userHasMovedCamera) {
        fitBoard(false);
      }
    });
  };

  const relayoutGraph = (recenter = true) => {
    applyGraphLayout(buildGraph(entities, relationViews || []), recenter);
  };

  const loadGraph = async () => {
    loading = true;

    try {
      relationViews = await GetCombinatory();
      relayoutGraph(true);
    } catch (err) {
      const message = err?.error ?? err?.message ?? err ?? "Error desconocido";
      showToast(`No se pudo construir el diagrama: ${message}`, "error");
    } finally {
      loading = false;
    }
  };

  $: {
    const nextSignature = entitySignature();
    if (nextSignature !== loadSignature) {
      loadSignature = nextSignature;
      if (entities.length === 0) {
        nodes = [];
        edges = [];
        selectedId = null;
        loading = false;
      } else {
        loadGraph();
      }
    }
  }

  $: currentFocusId = hoverId ?? selectedId;
  $: nodeLookup = new Map(nodes.map((node) => [node.id, node]));
  $: connectedIds = relatedNodeIds(currentFocusId);
  $: selectedNode = nodeById(selectedId);
  $: searchResults = searchResultsFromNodes(searchQuery);
  $: relationLegend = Object.entries(
    edges.reduce<Record<string, number>>((accumulator, edge) => {
      accumulator[edge.relation] = (accumulator[edge.relation] ?? 0) + 1;
      return accumulator;
    }, {})
  ).sort(([left], [right]) => left.localeCompare(right));

  $: viewportRect = zoom === 0
    ? {x: 0, y: 0, width: boardWidth, height: boardHeight}
    : {
      x: Math.max(0, -panX / zoom),
      y: Math.max(0, -panY / zoom),
      width: viewportWidth / zoom,
      height: viewportHeight / zoom
    };

  $: minimapAspect = boardWidth > 0 && boardHeight > 0 ? (boardHeight / boardWidth) : 0.6;

  $: if (viewportEl && !loading && !userHasMovedCamera) {
    requestAnimationFrame(() => {
      if (viewportEl && !userHasMovedCamera) {
        viewportWidth = viewportEl.clientWidth;
        viewportHeight = viewportEl.clientHeight;
        fitBoard(false);
      }
    });
  }

  onMount(() => () => {
    if (animationTimeout !== null) {
      window.clearTimeout(animationTimeout);
    }
  });
</script>

<svelte:window on:mousemove={handleWindowMouseMove} on:mouseup={stopDragging} on:mouseleave={stopDragging} on:resize={handleWindowResize}/>

<section class="graph-tab">
  <div class="graph-toolbar">
    <div>
      <p class="label">Diagrama vivo</p>
      <h3>Mapa relacional de arrastre libre</h3>
      <p class="muted">Layout automatico para arrancar rapido, drag manual para acomodar, y una vista mucho mas limpia: solo nombres y conexiones.</p>
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

      <button class="control control--soft" type="button" on:click={() => relayoutGraph(true)}>
        <ButtonIcon name="layout"/>
        <span>Reordenar</span>
      </button>
      <button class="control control--ghost" type="button" on:click={() => fitBoard(true)}>
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
      <div
        class="graph-stage"
        bind:this={viewportEl}
        aria-label="Diagrama relacional"
        on:mousedown={handleStageMouseDown}
        on:click={handleStageClick}
        on:wheel|preventDefault={handleWheel}
        on:keydown={handleStageKeydown}
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
                {#if currentFocusId === null || edge.source === currentFocusId || edge.target === currentFocusId || zoom > 0.92}
                  {@const midpoint = edgeMidpoint(edge)}
                  <g transform={`translate(${midpoint.x}, ${midpoint.y})`}>
                    <rect
                      x="-24"
                      y="-10"
                      rx="999"
                      ry="999"
                      width="48"
                      height="20"
                      fill={relationMeta(edge.relation).fill}
                      stroke="color-mix(in srgb, var(--ink) 10%, transparent)"
                    />
                    <text class="edge-label" text-anchor="middle" dominant-baseline="middle">{edge.relation}</text>
                  </g>
                {/if}
              </g>
            {/each}
          </svg>

          <div class="node-layer">
            {#each nodes as node (node.id)}
              <div
                class:graph-node={true}
                class:graph-node--selected={selectedId === node.id}
                class:graph-node--hovered={hoverId === node.id}
                class:graph-node--muted={currentFocusId !== null && node.id !== currentFocusId && !connectedIds.has(node.id)}
                class:graph-node--approved={node.status}
                role="button"
                tabindex="0"
                aria-label={`Entidad ${node.name}`}
                style={`left: ${node.x}px; top: ${node.y}px; width: ${node.width}px; height: ${node.height}px;`}
                on:click|stopPropagation={() => handleNodeClick(node.id)}
                on:keydown={(event) => handleNodeKeydown(node.id, event)}
                on:mouseenter={() => hoverId = node.id}
                on:mouseleave={resetHover}
              >
                <div class="node-topline">
                  <span class="node-degree">{node.degree}</span>
                  <span class={`node-state ${node.status ? 'node-state--approved' : ''}`}>
                    {node.status ? "OK" : "Draft"}
                  </span>
                </div>
                <h4>{node.name}</h4>
                <p class="node-meta">{node.degree} conexion{node.degree === 1 ? "" : "es"}</p>
              </div>
            {/each}
          </div>
        </div>
      </div>

      <aside class="graph-sidepanel">
        <section class="side-card">
          <p class="side-label">Foco actual</p>
          {#if selectedNode}
            <h4>{selectedNode.name}</h4>
            <div class="side-stats">
              <div>
                <span class="stat-value">{selectedNode.degree}</span>
                <span class="stat-label">relaciones</span>
              </div>
              <div>
                <span class="stat-value">{selectedNode.status ? "OK" : "Draft"}</span>
                <span class="stat-label">estado</span>
              </div>
            </div>
            <div class="focus-actions">
              <button class="control control--ghost control--sm control--block" type="button" on:click={() => onJumpTo("entities", selectedNode.id)}>
                <ButtonIcon name="database"/>
                <span>Definicion</span>
              </button>
              <button class="control control--accent control--sm control--block" type="button" on:click={() => onJumpTo("relations", selectedNode.id)}>
                <ButtonIcon name="relations"/>
                <span>Relaciones</span>
              </button>
              <button class="control control--soft control--sm control--block" type="button" on:click={() => onJumpTo("tertiary", selectedNode.id)}>
                <ButtonIcon name="attributes"/>
                <span>Atributos</span>
              </button>
            </div>
          {:else}
            <p class="side-copy">Haz click en una entidad para enfocarla. Usa "Reordenar" para regenerar el layout con mas aire y "Reencuadrar" para volver a centrar el mapa.</p>
          {/if}
        </section>

        <section class="side-card">
          <div class="mini-head">
            <p class="side-label">Overview</p>
            <span>{Math.round(zoom * 100)}%</span>
          </div>
          <div class="minimap-shell">
            <svg
              bind:this={minimapEl}
              class="minimap"
              viewBox={`0 0 ${boardWidth} ${boardHeight}`}
              preserveAspectRatio="xMidYMid meet"
              tabindex="0"
              role="button"
              aria-label="Minimapa del diagrama"
              on:click={handleMinimapClick}
              on:keydown={handleMinimapKeydown}
              style={`aspect-ratio: 1 / ${Math.max(0.7, minimapAspect)};`}
            >
              {#each edges as edge}
                <path class="minimap-edge" d={connectionPath(edge)} />
              {/each}
              {#each nodes as node}
                <rect
                  class:minimap-node={true}
                  class:minimap-node--selected={selectedId === node.id}
                  x={node.x}
                  y={node.y}
                  rx="18"
                  ry="18"
                  width={node.width}
                  height={node.height}
                />
              {/each}
              <rect
                class="minimap-viewport"
                x={viewportRect.x}
                y={viewportRect.y}
                width={viewportRect.width}
                height={viewportRect.height}
                rx="26"
                ry="26"
              />
            </svg>
          </div>
        </section>

        <section class="side-card">
          <p class="side-label">Tipos de relacion</p>
          <div class="legend-list">
            {#each relationLegend as [relation, count]}
              <div class="legend-item">
                <span class="legend-pill" style={`background: ${relationMeta(relation).fill}; color: ${relationMeta(relation).line};`}>
                  {relation}
                </span>
                <span class="legend-copy">{relationMeta(relation).label}</span>
                <span class="legend-count">{count}</span>
              </div>
            {/each}
            {#if relationLegend.length === 0}
              <p class="side-copy">Todavia no hay relaciones definidas.</p>
            {/if}
          </div>
        </section>

        {#if searchQuery.trim()}
          <section class="side-card">
            <p class="side-label">Resultados</p>
            <div class="result-list">
              {#each searchResults.slice(0, 8) as result}
                <button class="control control--sm control--block control--ghost" type="button" on:click={() => focusNode(result.id, true)}>
                  <ButtonIcon name="search"/>
                  <span>{result.name}</span>
                </button>
              {/each}
              {#if searchResults.length === 0}
                <p class="side-copy">Sin coincidencias para esa busqueda.</p>
              {/if}
            </div>
          </section>
        {/if}
      </aside>
    </div>
  {/if}
</section>

<style>
  .graph-tab {
    display: grid;
    gap: 1rem;
  }

  .graph-toolbar {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 1rem;
    align-items: end;
    padding: 1rem 1.05rem;
    border-radius: calc(var(--radius-md) - 4px);
    border: 1px solid var(--border);
    background: var(--panel-surface);
    box-shadow: var(--shadow-sm);
  }

  .label,
  .side-label {
    margin: 0;
    color: var(--accent);
    font-size: 0.72rem;
    letter-spacing: 0.17em;
    text-transform: uppercase;
    font-weight: 800;
  }

  .graph-toolbar h3,
  .side-card h4,
  .graph-node h4 {
    margin: 0.35rem 0 0;
    font-size: clamp(1.35rem, 3vw, 2rem);
    line-height: 0.96;
  }

  .muted,
  .side-copy,
  .legend-copy,
  .stat-label {
    color: var(--ink-faint);
    line-height: 1.55;
  }

  .toolbar-controls {
    display: flex;
    gap: 0.75rem;
    align-items: end;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .search-field {
    display: grid;
    gap: 0.35rem;
    min-width: min(320px, 100%);
    color: var(--ink-soft);
    font-size: 0.86rem;
    font-weight: 700;
  }

  .search-field input {
    min-height: 3rem;
    padding: 0 0.95rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--field-surface);
    color: var(--ink);
    outline: none;
    transition: border-color 150ms ease, box-shadow 150ms ease, background 150ms ease;
  }

  .search-field input:focus {
    border-color: var(--focus-border);
    box-shadow: var(--focus-ring);
    background: var(--field-surface-focus);
  }

  .graph-shell {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 310px;
    gap: 1rem;
    min-height: 40rem;
  }

  .graph-stage {
    position: relative;
    min-height: 40rem;
    border-radius: calc(var(--radius-lg) - 4px);
    border: 1px solid var(--border);
    background:
      radial-gradient(circle at 16% 10%, color-mix(in srgb, var(--accent) 8%, transparent), transparent 24%),
      linear-gradient(180deg, color-mix(in srgb, var(--surface) 88%, transparent), color-mix(in srgb, var(--surface-quiet) 90%, transparent));
    box-shadow: var(--shadow-lg);
    overflow: hidden;
    cursor: grab;
    user-select: none;
    outline: none;
  }

  .graph-stage:focus-visible {
    box-shadow: var(--shadow-lg), var(--focus-ring);
  }

  .graph-stage:active {
    cursor: grabbing;
  }

  .graph-stage::before {
    content: "";
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(var(--grid-line) 1px, transparent 1px),
      linear-gradient(90deg, var(--grid-line) 1px, transparent 1px);
    background-size: 36px 36px;
    opacity: 0.8;
    pointer-events: none;
  }

  .graph-stage::after {
    content: "";
    position: absolute;
    inset: 0;
    background: radial-gradient(circle at center, transparent 54%, color-mix(in srgb, var(--surface-quiet) 16%, transparent) 100%);
    pointer-events: none;
  }

  .graph-camera {
    position: absolute;
    inset: 0 auto auto 0;
    transform-origin: 0 0;
  }

  .graph-camera--smooth {
    transition: transform 280ms cubic-bezier(.19, 1, .22, 1);
  }

  .edge-layer,
  .node-layer {
    position: absolute;
    inset: 0;
  }

  .edge {
    fill: none;
    stroke-width: 2.2;
    opacity: 0.28;
    transition: opacity 180ms ease, stroke-width 180ms ease;
  }

  .edge--active {
    opacity: 0.9;
    stroke-width: 2.8;
  }

  .edge--muted {
    opacity: 0.08;
  }

  .edge-label {
    fill: var(--ink);
    font-size: 0.68rem;
    font-weight: 800;
    letter-spacing: 0.06em;
  }

  .graph-node {
    position: absolute;
    display: grid;
    align-content: start;
    gap: 0.42rem;
    padding: 0.9rem 0.95rem;
    border-radius: 1.2rem;
    border: 1px solid var(--border);
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 96%, transparent), color-mix(in srgb, var(--surface) 92%, transparent)),
      linear-gradient(135deg, color-mix(in srgb, var(--accent) 8%, transparent), transparent 50%);
    box-shadow: 0 18px 30px color-mix(in srgb, var(--ink) 9%, transparent);
    cursor: pointer;
    transition: transform 180ms ease, box-shadow 180ms ease, border-color 180ms ease, opacity 180ms ease;
    outline: none;
  }

  .graph-node:hover,
  .graph-node--selected,
  .graph-node:focus-visible {
    transform: translateY(-3px);
    border-color: color-mix(in srgb, var(--accent) 24%, var(--border));
    box-shadow: 0 22px 34px color-mix(in srgb, var(--ink) 14%, transparent);
  }

  .graph-node--approved {
    background:
      linear-gradient(180deg, color-mix(in srgb, var(--surface-strong) 96%, transparent), color-mix(in srgb, var(--surface) 92%, transparent)),
      linear-gradient(135deg, color-mix(in srgb, var(--success) 9%, transparent), transparent 54%);
  }

  .graph-node--muted {
    opacity: 0.28;
  }

  .node-topline {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
    align-items: center;
  }

  .node-degree,
  .node-state,
  .legend-pill {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 1.65rem;
    padding: 0 0.65rem;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--chip-surface);
    font-size: 0.76rem;
    font-weight: 800;
  }

  .node-degree {
    color: var(--ink-soft);
  }

  .node-state {
    color: var(--accent);
  }

  .node-state--approved {
    color: var(--success);
  }

  .graph-node h4 {
    font-size: 1.15rem;
    line-height: 1.05;
    color: var(--ink);
    word-break: break-word;
  }

  .node-meta {
    margin: 0;
    color: var(--ink-faint);
    font-size: 0.84rem;
    line-height: 1.35;
  }

  .graph-sidepanel {
    display: grid;
    gap: 0.9rem;
    align-content: start;
  }

  .side-card {
    display: grid;
    gap: 0.8rem;
    padding: 1rem;
    border-radius: calc(var(--radius-md) - 4px);
    border: 1px solid var(--border);
    background: var(--panel-surface-strong);
    box-shadow: var(--shadow-sm);
  }

  .side-card h4 {
    margin: 0;
    font-size: 1.32rem;
  }

  .side-stats {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.75rem;
  }

  .side-stats > div {
    display: grid;
    gap: 0.18rem;
    padding: 0.8rem 0.9rem;
    border-radius: 1rem;
    background: color-mix(in srgb, var(--surface-strong) 76%, transparent);
    border: 1px solid color-mix(in srgb, var(--ink) 8%, transparent);
  }

  .stat-value {
    color: var(--ink);
    font-size: 1.24rem;
    font-weight: 800;
  }

  .focus-actions,
  .legend-list,
  .result-list {
    display: grid;
    gap: 0.55rem;
  }

  .mini-head {
    display: flex;
    justify-content: space-between;
    gap: 0.6rem;
    align-items: center;
    color: var(--ink-faint);
    font-size: 0.85rem;
    font-weight: 700;
  }

  .minimap-shell {
    padding: 0.5rem;
    border-radius: 1rem;
    border: 1px solid color-mix(in srgb, var(--ink) 8%, transparent);
    background: color-mix(in srgb, var(--surface-quiet) 86%, transparent);
  }

  .minimap {
    width: 100%;
    display: block;
    cursor: crosshair;
    outline: none;
  }

  .minimap:focus-visible {
    box-shadow: var(--focus-ring);
  }

  .minimap-edge {
    fill: none;
    stroke: color-mix(in srgb, var(--ink) 18%, transparent);
    stroke-width: 8;
    stroke-linecap: round;
    opacity: 0.56;
  }

  .minimap-node {
    fill: color-mix(in srgb, var(--surface-strong) 92%, transparent);
    stroke: color-mix(in srgb, var(--ink) 10%, transparent);
    stroke-width: 8;
  }

  .minimap-node--selected {
    stroke: var(--accent);
    fill: color-mix(in srgb, var(--accent) 12%, var(--surface-strong));
  }

  .minimap-viewport {
    fill: none;
    stroke: var(--accent);
    stroke-width: 12;
    stroke-dasharray: 20 12;
  }

  .legend-item {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 0.55rem;
    align-items: center;
  }

  .legend-count {
    color: var(--ink);
    font-weight: 800;
  }

  @media (prefers-reduced-motion: reduce) {
    .graph-camera--smooth,
    .graph-node,
    .edge {
      transition: none;
    }
  }

  @media (max-width: 1000px) {
    .graph-shell {
      grid-template-columns: 1fr;
    }

    .graph-sidepanel {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      align-items: start;
    }
  }

  @media (max-width: 720px) {
    .graph-toolbar {
      grid-template-columns: 1fr;
      align-items: stretch;
    }

    .toolbar-controls {
      justify-content: stretch;
    }

    .search-field {
      min-width: 0;
    }

    .graph-stage {
      min-height: 34rem;
    }

    .graph-sidepanel {
      grid-template-columns: 1fr;
    }
  }
</style>
