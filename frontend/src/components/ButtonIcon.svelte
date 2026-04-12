<script lang="ts">
  import {
    ArrowUp, ArrowLeft, ArrowDown, Save, FolderOpen, Plus, Download,
    LogOut, X, Trash2, Pencil, Check, Circle, Database, Share2,
    List, Users, GitBranch, Search, Eraser, Copy, Maximize,
    LayoutDashboard, Eye, ChevronLeft, ChevronRight, Layers, Sun, Moon,
    Link, Zap, Home
  } from 'lucide-svelte';

  let { 
    name = "spark", 
    title = "" 
  } = $props<{
    name?: string;
    title?: string;
  }>();

  const icons: Record<string, any> = {
    'home': Home,
    'arrow-up': ArrowUp,
    'arrow-left': ArrowLeft,
    'arrow-down': ArrowDown,
    'save': Save,
    'folder': FolderOpen,
    'plus': Plus,
    'download': Download,
    'exit': LogOut,
    'close': X,
    'trash': Trash2,
    'edit': Pencil,
    'check': Check,
    'check-off': Circle,
    'database': Database,
    'relations': Share2,
    'attributes': List,
    'roles': Users,
    'flows': GitBranch,
    'search': Search,
    'clear': Eraser,
    'copy': Copy,
    'frame': Maximize,
    'layout': LayoutDashboard,
    'eye': Eye,
    'chevron-left': ChevronLeft,
    'chevron-right': ChevronRight,
    'stack': Layers,
    'theme-light': Sun,
    'theme-dark': Moon,
    'link': Link,
    'spark': Zap
  };

  const ActiveIcon = $derived(icons[name] || Zap);
</script>

<ActiveIcon
  class="button-glyph"
  aria-hidden={title ? undefined : "true"}
  aria-label={title || undefined}
/>
