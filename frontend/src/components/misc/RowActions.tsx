import { Ellipsis, type LucideIcon } from 'lucide-react';
import { Button } from '@/components/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/misc';

export interface RowAction {
  label: string;
  icon?: LucideIcon;
  danger?: boolean;
  disabled?: boolean;
  onSelect: () => void;
}

interface RowActionsProps {
  actions: RowAction[];
  label?: string;
}

export function RowActions({ actions, label = 'Acciones' }: RowActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={<Button variant="ghost" size="icon-sm" aria-label={label} />}>
        <Ellipsis />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {actions.map((action) => (
          <DropdownMenuItem
            key={action.label}
            danger={action.danger}
            disabled={action.disabled}
            onSelect={action.onSelect}
          >
            {action.icon && <action.icon className="menu-item-icon" />}
            {action.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
