import { Warehouse } from 'lucide-react';
import { ModulePage } from './ModulePage';
import { inventory, products } from '@/data/mock';
import { formatCurrency } from '@/lib/utils';

export function InventoryPage() {
  const totalValue = inventory.reduce((s, i) => {
    const p = products.find((p) => p.id === i.productId);
    return s + (p ? p.salePrice * i.quantity : 0);
  }, 0);
  const clearance = inventory.filter((i) => i.isClearance).length;
  return (
    <ModulePage
      title="Inventario"
      subtitle="Stock, movimientos y rotación"
      icon={Warehouse}
      description="Consulte el stock actual, movimientos y productos próximos a vencer."
      phase="Fase 6"
      features={[
        'Stock actual por almacén',
        'Movimientos (compras, ventas, ajustes, transferencias)',
        'Lotes con fecha de ingreso',
        'Regla de 25 días — productos en remate',
        'Alertas de stock bajo y sin stock',
        'Valorización por método PEPS, UEPS, promedio',
        'Toma de inventario física',
        'Rotación de productos',
      ]}
      mockStats={[
        { label: 'Items en stock', value: String(inventory.length) },
        { label: 'En remate', value: String(clearance) },
        { label: 'Stock bajo', value: String(products.filter((p) => p.currentStock < p.minStock).length) },
        { label: 'Valor total', value: formatCurrency(totalValue) },
      ]}
    />
  );
}
