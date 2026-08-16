import {
  LayoutDashboard,
  Users,
  Truck,
  Package,
  Warehouse,
  ShoppingCart,
  Receipt,
  Landmark,
  BookOpen,
  BarChart3,
  Settings,
  Shield,
  Tags,
  Search,
  Bell,
  Sun,
  Moon,
  Monitor,
  ChevronDown,
  ChevronUp,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  ChevronsUpDown,
  ChevronsUp,
  ChevronsDown,
  Check,
  X,
  Plus,
  Minus,
  Eye,
  EyeOff,
  Edit,
  Trash2,
  Save,
  Filter,
  Download,
  Upload,
  RefreshCw,
  MoreHorizontal,
  MoreVertical,
  ArrowUpRight,
  ArrowDownRight,
  ArrowUp,
  ArrowDown,
  ArrowLeft,
  ArrowRight,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  CheckCircle2,
  Banknote,
  Info,
  XCircle,
  HelpCircle,
  Inbox,
  Loader2,
  type LucideIcon,
} from 'lucide-react';

export const Icons = {
  Navigation: {
    Dashboard: LayoutDashboard,
    Customers: Users,
    Suppliers: Truck,
    Products: Package,
    Inventory: Warehouse,
    Purchases: ShoppingCart,
    Sales: Receipt,
    Treasury: Landmark,
    Accounting: BookOpen,
    Reports: BarChart3,
    Settings: Settings,
    Administration: Shield,
    Catalog: Tags,
  },
  Action: {
    Search,
    Save,
    Cancel: X,
    Delete: Trash2,
    Edit,
    Create: Plus,
    Close: X,
    Back: ArrowLeft,
    Next: ArrowRight,
    Refresh: RefreshCw,
    Filter,
    Download,
    Upload,
    More: MoreHorizontal,
    MoreVertical,
    Payment: Banknote,
  },
  Direction: {
    ChevronDown,
    ChevronUp,
    ChevronLeft,
    ChevronRight,
    ChevronsLeft,
    ChevronsRight,
    ChevronsUpDown,
    ChevronsUp,
    ChevronsDown,
    ArrowUp,
    ArrowDown,
    ArrowLeft,
    ArrowRight,
    ArrowUpRight,
    ArrowDownRight,
  },
  Status: {
    Success: CheckCircle2,
    Warning: AlertTriangle,
    Error: XCircle,
    Info,
    Help: HelpCircle,
    Empty: Inbox,
  },
  Theme: {
    Sun,
    Moon,
    System: Monitor,
  },
  Visibility: {
    Show: Eye,
    Hide: EyeOff,
  },
  Trend: {
    Up: TrendingUp,
    Down: TrendingDown,
  },
  Math: {
    Plus,
    Minus,
  },
  Check,
  Bell,
  Loading: Loader2,
} as const;

export type IconName =
  | `Navigation.${keyof typeof Icons.Navigation}`
  | `Action.${keyof typeof Icons.Action}`
  | `Direction.${keyof typeof Icons.Direction}`
  | `Status.${keyof typeof Icons.Status}`
  | `Theme.${keyof typeof Icons.Theme}`
  | `Visibility.${keyof typeof Icons.Visibility}`
  | `Trend.${keyof typeof Icons.Trend}`
  | `Math.${keyof typeof Icons.Math}`
  | 'Check'
  | 'Bell'
  | 'Loading';

export function getIcon(name: IconName): LucideIcon {
  if (name.includes('.')) {
    const [group, key] = name.split('.') as [keyof typeof Icons, string];
    const grp = Icons[group] as unknown as Record<string, LucideIcon>;
    return grp[key] ?? Inbox;
  }
  return ((Icons as unknown) as Record<string, LucideIcon>)[name] ?? Inbox;
}
