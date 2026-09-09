import { useMemo, useState } from 'react';
import { Pencil, Trash2, Plus } from 'lucide-react';
import { PageContainer, PageHeader } from '@/components/layout';
import { DataTable, type Column } from '@/components/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/tabs';
import { Button } from '@/components/button';
import { EmptyState } from '@/components/feedback';
import { AlertDialog } from '@/components/dialog';
import { RowActions } from '@/components/misc';
import {
  useCatalogBrands,
  useCatalogCategories,
  useDeleteBrand,
  useDeleteCategory,
} from '@/features/catalog/hooks/useCatalog';
import {
  CatalogItemFormDialog,
  type CatalogItem,
  type CatalogKind,
} from '@/features/catalog/components/CatalogItemFormDialog';
import { useNotificationStore } from '@/stores/notification';

const columnSet: Column<CatalogItem>[] = [
  {
    id: 'code',
    header: 'Código',
    sortable: true,
    sticky: true,
    cell: (row) => <span className="fw-medium tabular">{row.code}</span>,
  },
  {
    id: 'name',
    header: 'Nombre',
    sortable: true,
    cell: (row) => row.name,
  },
];

function CatalogTable({ kind }: { kind: CatalogKind }) {
  const isCategory = kind === 'category';
  const categoriesQuery = useCatalogCategories();
  const brandsQuery = useCatalogBrands();
  const deleteCategory = useDeleteCategory();
  const deleteBrand = useDeleteBrand();

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<CatalogItem | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<CatalogItem | null>(null);

  const push = useNotificationStore((s) => s.push);

  const query = isCategory ? categoriesQuery : brandsQuery;
  const data = query.data ?? [];
  const singular = isCategory ? 'categoría' : 'marca';
  const plural = isCategory ? 'categorías' : 'marcas';
  const noun = isCategory ? 'Categoría' : 'Marca';
  const deleteMutation = isCategory ? deleteCategory : deleteBrand;

  const openCreate = () => {
    setEditing(null);
    setFormOpen(true);
  };

  const columns = useMemo<Column<CatalogItem>[]>(() => {
    return [
      ...columnSet,
      {
        id: 'actions',
        header: '',
        width: 72,
        cell: (row) => (
          <RowActions
            actions={[
              {
                label: 'Editar',
                icon: Pencil,
                onSelect: () => {
                  setEditing(row);
                  setFormOpen(true);
                },
              },
              {
                label: 'Eliminar',
                icon: Trash2,
                danger: true,
                onSelect: () => setDeleteTarget(row),
              },
            ]}
            label={`Acciones de ${row.name}`}
          />
        ),
      },
    ];
  }, []);

  return (
    <>
      <DataTable
        columns={columns}
        data={data}
        keyField="id"
        loading={query.isLoading}
        error={query.isError ? (query.error as Error) : null}
        onRetry={() => query.refetch()}
        toolbarRight={
          <Button onClick={openCreate}>
            <Plus /> Nueva {singular}
          </Button>
        }
        empty={
          <EmptyState
            title={`No hay ${plural} registradas`}
            description={`Crea ${plural} para clasificar tus productos.`}
            action={{ label: `Nueva ${singular}`, onClick: openCreate }}
          />
        }
      />

      <CatalogItemFormDialog open={formOpen} onOpenChange={setFormOpen} kind={kind} item={editing} />

      <AlertDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        variant="destructive"
        title={`Eliminar ${singular}`}
        description={`¿Eliminar la ${singular} ${deleteTarget?.name ?? ''} (${deleteTarget?.code ?? ''})? Los productos existentes conservarán sus datos.`}
        confirmLabel="Eliminar"
        loading={deleteMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              push({ title: `${noun} eliminada`, variant: 'success' });
              setDeleteTarget(null);
            },
            onError: (err: unknown) => {
              push({
                title: `No se pudo eliminar la ${singular}`,
                description: err instanceof Error ? err.message : undefined,
                variant: 'destructive',
              });
              setDeleteTarget(null);
            },
          });
        }}
      />
    </>
  );
}

export function CatalogSettingsPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Catálogo"
        subtitle="Administra categorías y marcas de productos"
      />

      <Tabs defaultValue="categories">
        <TabsList aria-label="Secciones del catálogo">
          <TabsTrigger value="categories">Categorías</TabsTrigger>
          <TabsTrigger value="brands">Marcas</TabsTrigger>
        </TabsList>
        <TabsContent value="categories">
          <CatalogTable kind="category" />
        </TabsContent>
        <TabsContent value="brands">
          <CatalogTable kind="brand" />
        </TabsContent>
      </Tabs>
    </PageContainer>
  );
}
