# db_planner

Herramienta de escritorio (Wails + Go + Svelte) para diseñar entidades y sus relaciones de forma visual.

## Requisitos
- Go 1.21+
- Node 18+ y pnpm/npm
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## Desarrollo
1) Instala dependencias del frontend:
   ```bash
   cd frontend
   npm install
   ```
2) Vuelve a la raíz y levanta en modo dev con hot reload:
   ```bash
   wails dev
   ```
   - UI: Vite sirve la app.
   - Métodos Go expuestos por Wails se pueden consumir desde la UI.

## Build
Genera el binario empaquetado:
```bash
wails build
```
El artefacto queda en `build/`.

## Funcionalidades destacadas
- Gestión de entidades con tabla editable.
- Pestaña de relaciones con navegación por entidad principal y controles sticky.
- Persistencia de proyectos en JSON y recarga desde la app.

## Estructura rápida
- `frontend/`: Svelte + Vite.
- `main.go`, `app.go`: arranque Wails y lógica de backend Go.
- `wails.json`: configuración de build.

## Licencia
MIT con atribución. Si reutilizas código o assets, conserva el aviso de copyright y la referencia al proyecto.
