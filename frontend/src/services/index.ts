export { customersService, type CustomerQuery, type CustomerCreateInput, type CustomerUpdateInput } from './customers';
export { suppliersService, type SupplierQuery, type SupplierCreateInput, type SupplierUpdateInput } from './suppliers';
export { productsService, type ProductQuery, type ProductCreateInput, type ProductUpdateInput } from './products';
export { catalogService } from './catalog';
export { salesService, type SaleCreateInput } from './sales';
export { purchasingService, type PurchaseCreateInput } from './purchasing';
export { inventoryService, type InventoryMovement } from './inventory';
export {
  treasuryService,
  type BankAccount,
  type BankAccountInput,
  type BankAccountUpdateInput,
  type BankTransaction,
  type BankTransactionInput,
} from './treasury';
export { reportsService, type ReportType, type ReportRunInput, type ReportResult } from './reports';
export { settingsService } from './settings';
