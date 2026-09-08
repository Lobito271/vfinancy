import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Building2, Check, Pencil, Plus } from 'lucide-react';
import { Button } from '@/components/button';
import { Spinner } from '@/components/feedback';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/misc';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { useNotificationStore } from '@/stores/notification';
import { CompanyDialog } from '@/features/workspace/components/CompanyDialog';
import type { CompanyDTO } from '@/services/wails-types';

export function CompanySwitcher() {
  const queryClient = useQueryClient();
  const push = useNotificationStore((s) => s.push);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editCompany, setEditCompany] = useState<CompanyDTO | null>(null);

  const companies = useQuery({ queryKey: queryKeys.companies.all, queryFn: () => wailsClient.listCompanies() });
  const active = useQuery({ queryKey: queryKeys.companies.active, queryFn: () => wailsClient.getActiveCompany() });

  const switchCompany = useMutation({
    mutationFn: async (id: string) => {
      await wailsClient.setActiveCompany(id);
    },
    onSuccess: async () => {
      push({ title: 'Empresa activa actualizada', variant: 'success' });
      await queryClient.clear();
      await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo cambiar de empresa', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const list = companies.data ?? [];
  const currentId = active.data?.id;

  const openEdit = (c: CompanyDTO | null) => {
    setEditCompany(c);
    setDialogOpen(true);
  };

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" className="topbar__company" aria-label="Cambiar empresa">
            <Building2 strokeWidth={2.5} />
            {companies.isLoading ? <Spinner size="sm" /> : <span className="truncate">{active.data?.legalName ?? active.data?.tradeName ?? 'Empresa'}</span>}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" style={{ width: '18rem' }}>
          <DropdownMenuLabel>Empresas</DropdownMenuLabel>
          <DropdownMenuSeparator />
          {list.length === 0 ? (
            <DropdownMenuItem disabled>No hay empresas activas</DropdownMenuItem>
          ) : (
            list.map((c) => (
              <DropdownMenuItem
                key={c.id}
                onSelect={() => {
                  if (c.id !== currentId) switchCompany.mutate(c.id);
                }}
              >
                {c.id === currentId && <Check className="menu-item-icon" />}
                <span className="truncate">{c.legalName}</span>
                <Button
                  variant="ghost"
                  size="sm"
                  className="company-switch__edit"
                  onClick={(e) => {
                    e.stopPropagation();
                    openEdit(c);
                  }}
                >
                  <Pencil aria-hidden="true" />
                </Button>
              </DropdownMenuItem>
            ))
          )}
          <DropdownMenuSeparator />
          <DropdownMenuItem onSelect={() => openEdit(null)}>
            <Plus className="menu-item-icon" /> Nueva empresa
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <CompanyDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        company={editCompany}
        isActive={Boolean(editCompany && editCompany.id === currentId)}
      />
    </>
  );
}