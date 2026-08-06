import { Package } from 'lucide-react';
import { ModulePage } from './ModulePage';
import { products } from '@/data/mock';
import { formatCurrency } from '@/lib/utils';

export function ProductsPage() {
  return (
    <ModulePage
      title="Productos"
      subtitle="Catálogo de productos y precios"
      icon={Package}
      description="Administre su catálogo de productos, precios, categorías y marcas."
      phase="Fase 3"
      features={[
        'SKU y código de barras',
        'Costo y precio de venta',
        'Stock mínimo y máximo',
        'Categorías y marcas',
        'Unidades de medida',
        'Márgenes por producto',
        'Imágenes y descripciones',
        'Importación masiva',
      ]}
      mockStats={[
        { label: 'Total productos', value: String(products.length) },
        { label: 'Categorías', value: '8' },
        { label: 'Marcas', value: '8' },
        { label: 'Valor del catálogo', value: formatCurrency(products.reduce((s, p) => s + p.salePrice, 0)) },
      ]}
    />
  );
}
