export type SaleStatus = 'pending' | 'paid' | 'partial' | 'cancelled';

export type CustomerStatus = 'active' | 'inactive' | 'blocked';

export interface Customer {
  id: string;
  documentType: 'DNI' | 'RUC' | 'CE';
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
  address?: string;
  creditLimit: number;
  currentDebt: number;
  status: CustomerStatus;
  totalPurchases: number;
  createdAt: string;
}

export interface Supplier {
  id: string;
  documentNumber: string;
  businessName: string;
  contactName?: string;
  phone?: string;
  email?: string;
  currentDebt: number;
  status: 'active' | 'inactive';
}

export interface Product {
  id: string;
  sku: string;
  barcode?: string;
  description: string;
  category: string;
  brand: string;
  cost: number;
  salePrice: number;
  minStock: number;
  maxStock: number;
  currentStock: number;
}

export interface InventoryItem {
  id: string;
  productId: string;
  productSku: string;
  productDescription: string;
  warehouse: string;
  quantity: number;
  arrivalDate: string;
  maxSaleDate: string;
  ageDays: number;
  daysRemaining: number;
  isClearance: boolean;
}

export interface Sale {
  id: string;
  number: string;
  customerId: string;
  customerName: string;
  date: string;
  status: SaleStatus;
  subtotal: number;
  tax: number;
  discount: number;
  total: number;
  cost: number;
  profit: number;
}

export interface Purchase {
  id: string;
  number: string;
  supplierId: string;
  supplierName: string;
  date: string;
  status: SaleStatus;
  total: number;
}

export interface ActivityItem {
  id: string;
  type: 'sale' | 'purchase' | 'payment' | 'customer' | 'product';
  description: string;
  amount?: number;
  date: string;
  user: string;
}

export interface ChartPoint {
  label: string;
  value: number;
}

function mulberry32(seed: number) {
  let a = seed;
  return () => {
    a |= 0;
    a = (a + 0x6d2b79f5) | 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

const rng = mulberry32(42);

function pick<T>(arr: readonly T[]): T {
  return arr[Math.floor(rng() * arr.length)];
}

function int(min: number, max: number): number {
  return Math.floor(rng() * (max - min + 1)) + min;
}

function money(min: number, max: number): number {
  return Math.round((rng() * (max - min) + min) * 100) / 100;
}

const FIRST = [
  'María',
  'José',
  'Carlos',
  'Ana',
  'Luis',
  'Rosa',
  'Jorge',
  'Lucía',
  'Miguel',
  'Elena',
  'Pedro',
  'Carmen',
  'Andrés',
  'Patricia',
  'Diego',
  'Sofía',
];
const LAST = [
  'García',
  'Rodríguez',
  'López',
  'Martínez',
  'González',
  'Pérez',
  'Sánchez',
  'Ramírez',
  'Torres',
  'Flores',
  'Rivera',
  'Gómez',
  'Díaz',
  'Morales',
  'Castro',
  'Ortiz',
];
const COMPANIES = [
  'Distribuidora',
  'Comercial',
  'Importaciones',
  'Representaciones',
  'Grupo',
  'Inversiones',
  'Servicios',
  'Industrias',
];
const SUFFIXES = ['S.A.C.', 'E.I.R.L.', 'S.R.L.', 'S.A.'];
const BRANDS = ['Acme', 'Genérica', 'ProMax', 'EcoLine', 'Nordic', 'Andina', 'Pacífico', 'Premium'];
const CATEGORIES = [
  'Abarrotes',
  'Bebidas',
  'Limpieza',
  'Panadería',
  'Lácteos',
  'Conservas',
  'Snacks',
  'Cuidado Personal',
];
const WAREHOUSES = ['Principal', 'Sucursal Norte', 'Sucursal Sur'];
const PRODUCTS: Array<[string, string]> = [
  ['Arroz', 'kg'],
  ['Azúcar', 'kg'],
  ['Aceite vegetal', 'L'],
  ['Fideos', 'kg'],
  ['Leche evaporada', 'L'],
  ['Atún en lata', 'und'],
  ['Galletas', 'und'],
  ['Detergente', 'kg'],
  ['Jabón de tocador', 'und'],
  ['Papel higiénico', 'pack'],
  ['Refresco de cola', 'L'],
  ['Agua mineral', 'L'],
  ['Café molido', 'g'],
  ['Chocolate', 'und'],
  ['Mantequilla', 'g'],
];

function pad(n: number, w = 2): string {
  return String(n).padStart(w, '0');
}

function isoDaysAgo(days: number): string {
  const d = new Date();
  d.setDate(d.getDate() - days);
  return d.toISOString();
}

function ruc(): string {
  return `${int(10000000000, 20999999999)}`;
}

function dni(): string {
  return `${int(10000000, 79999999)}`;
}

export const customers: Customer[] = Array.from({ length: 24 }, (_, i) => {
  const isCompany = rng() > 0.4;
  const businessName = isCompany
    ? `${pick(COMPANIES)} ${pick(LAST)} ${pick(SUFFIXES)}`
    : `${pick(FIRST)} ${pick(LAST)} ${pick(LAST)}`;
  const currentDebt = money(0, 18000);
  return {
    id: `c-${i + 1}`,
    documentType: isCompany ? 'RUC' : 'DNI',
    documentNumber: isCompany ? ruc() : dni(),
    businessName,
    contactName: `${pick(FIRST)} ${pick(LAST)}`,
    phone: `9${int(10000000, 99999999)}`,
    email: `cliente${i + 1}@empresa.pe`,
    address: `Av. ${pick(LAST)} ${int(100, 9999)}, Lima`,
    creditLimit: money(5000, 50000),
    currentDebt,
    status: currentDebt > 15000 ? 'blocked' : rng() > 0.1 ? 'active' : 'inactive',
    totalPurchases: money(currentDebt + 1000, currentDebt + 80000),
    createdAt: isoDaysAgo(int(30, 720)),
  };
});

export const suppliers: Supplier[] = Array.from({ length: 18 }, (_, i) => ({
  id: `s-${i + 1}`,
  documentNumber: ruc(),
  businessName: `${pick(COMPANIES)} ${pick(LAST)} ${pick(SUFFIXES)}`,
  contactName: `${pick(FIRST)} ${pick(LAST)}`,
  phone: `9${int(10000000, 99999999)}`,
  email: `proveedor${i + 1}@empresa.pe`,
  currentDebt: money(0, 22000),
  status: rng() > 0.1 ? 'active' : 'inactive',
}));

export const products: Product[] = PRODUCTS.flatMap(([name, unit], i) =>
  BRANDS.slice(0, 3).map((brand, j) => {
    const cost = money(2, 80);
    const salePrice = Math.round(cost * (1.2 + rng() * 0.6) * 100) / 100;
    const minStock = int(10, 30);
    const maxStock = int(200, 600);
    return {
      id: `p-${i + 1}-${j + 1}`,
      sku: `${pad(i + 1, 3)}${pad(j + 1, 2)}-${name.slice(0, 3).toUpperCase()}`,
      barcode: `775${int(100000000, 999999999)}`,
      description: `${name} ${brand} ${unit}`,
      category: pick(CATEGORIES),
      brand,
      cost,
      salePrice,
      minStock,
      maxStock,
      currentStock: int(0, maxStock + 20),
    };
  }),
);

export const inventory: InventoryItem[] = products.slice(0, 18).map((p, i) => {
  const ageDays = int(2, 40);
  const arrival = new Date();
  arrival.setDate(arrival.getDate() - ageDays);
  const maxSale = new Date(arrival);
  maxSale.setDate(maxSale.getDate() + 25);
  const today = new Date();
  const daysRemaining = Math.max(
    0,
    Math.ceil((maxSale.getTime() - today.getTime()) / (1000 * 60 * 60 * 24)),
  );
  return {
    id: `inv-${i + 1}`,
    productId: p.id,
    productSku: p.sku,
    productDescription: p.description,
    warehouse: pick(WAREHOUSES),
    quantity: int(0, 80),
    arrivalDate: arrival.toISOString(),
    maxSaleDate: maxSale.toISOString(),
    ageDays,
    daysRemaining,
    isClearance: daysRemaining === 0,
  };
});

export const sales: Sale[] = Array.from({ length: 30 }, (_, i) => {
  const c = pick(customers);
  const subtotal = money(80, 4500);
  const tax = Math.round(subtotal * 0.18 * 100) / 100;
  const discount = money(0, subtotal * 0.1);
  const total = Math.round((subtotal + tax - discount) * 100) / 100;
  const cost = Math.round(subtotal * (0.55 + rng() * 0.2) * 100) / 100;
  const status: SaleStatus =
    i % 7 === 0 ? 'pending' : i % 11 === 0 ? 'partial' : i % 17 === 0 ? 'cancelled' : 'paid';
  return {
    id: `sa-${i + 1}`,
    number: `V-${2024}-${pad(i + 1, 5)}`,
    customerId: c.id,
    customerName: c.businessName,
    date: isoDaysAgo(int(0, 30)),
    status,
    subtotal,
    tax,
    discount,
    total,
    cost,
    profit: Math.round((total - cost - tax) * 100) / 100,
  };
});

export const purchases: Purchase[] = Array.from({ length: 16 }, (_, i) => {
  const s = pick(suppliers);
  const total = money(500, 12000);
  return {
    id: `pu-${i + 1}`,
    number: `C-${2024}-${pad(i + 1, 5)}`,
    supplierId: s.id,
    supplierName: s.businessName,
    date: isoDaysAgo(int(0, 30)),
    status: i % 5 === 0 ? 'pending' : 'paid',
    total,
  };
});

export const activity: ActivityItem[] = [
  { id: 'a-1', type: 'sale', description: 'Venta V-2024-00012', amount: 1250.5, date: isoDaysAgo(0), user: 'Jorge Z.' },
  { id: 'a-2', type: 'payment', description: 'Pago recibido — Distribuidora García', amount: 3400, date: isoDaysAgo(0), user: 'Jorge Z.' },
  { id: 'a-3', type: 'purchase', description: 'Compra C-2024-00008', amount: 5800, date: isoDaysAgo(1), user: 'Jorge Z.' },
  { id: 'a-4', type: 'customer', description: 'Cliente creado: Comercial López S.A.C.', date: isoDaysAgo(1), user: 'Jorge Z.' },
  { id: 'a-5', type: 'product', description: 'Stock actualizado — Arroz Acme kg', date: isoDaysAgo(2), user: 'Jorge Z.' },
  { id: 'a-6', type: 'sale', description: 'Venta V-2024-00011', amount: 680, date: isoDaysAgo(2), user: 'Jorge Z.' },
  { id: 'a-7', type: 'payment', description: 'Pago a proveedor — Industrias Gómez', amount: 2150, date: isoDaysAgo(3), user: 'Jorge Z.' },
  { id: 'a-8', type: 'sale', description: 'Venta V-2024-00010', amount: 1980, date: isoDaysAgo(3), user: 'Jorge Z.' },
];

export const salesLast7Days: ChartPoint[] = Array.from({ length: 7 }, (_, i) => {
  const d = new Date();
  d.setDate(d.getDate() - (6 - i));
  const base = 3500 + Math.sin(i * 0.6) * 800 + rng() * 1200;
  return {
    label: d.toLocaleDateString('es-PE', { weekday: 'short' }),
    value: Math.round(base * 100) / 100,
  };
});

export const topProducts: ChartPoint[] = [
  { label: 'Arroz Acme kg', value: 12450 },
  { label: 'Aceite ProMax L', value: 9800 },
  { label: 'Fideos EcoLine kg', value: 7200 },
  { label: 'Detergente Nordic kg', value: 5400 },
  { label: 'Atún en lata und', value: 4900 },
];

export const salesByCategory: ChartPoint[] = [
  { label: 'Abarrotes', value: 42 },
  { label: 'Bebidas', value: 23 },
  { label: 'Limpieza', value: 14 },
  { label: 'Lácteos', value: 11 },
  { label: 'Snacks', value: 10 },
];

export const dashboardKpis = {
  monthSales: 142580.5,
  monthSalesChange: 12.4,
  monthPurchases: 87230.0,
  monthPurchasesChange: -3.1,
  profit: 28940.25,
  profitChange: 8.7,
  inventoryValue: 218700.0,
  accountsReceivable: 94620.5,
  accountsPayable: 52180.0,
  clearanceProducts: inventory.filter((i) => i.isClearance).length,
  customersWithDebt: customers.filter((c) => c.currentDebt > 0).length,
  lowStock: products.filter((p) => p.currentStock < p.minStock).length,
};
