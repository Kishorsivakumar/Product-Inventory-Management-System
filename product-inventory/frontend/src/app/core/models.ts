export type ProductStatus = 'active' | 'inactive' | 'discontinued';
export type StockMode = 'set' | 'add' | 'remove';

export interface Category {
  id: number;
  name: string;
}

export interface Product {
  id: number;
  name: string;
  sku: string;
  category_id: number;
  category: string;
  price: number;
  quantity: number;
  status: ProductStatus;
  low_stock: boolean;
}

export interface ProductInput {
  name: string;
  sku: string;
  category_id: number;
  price: number;
  quantity: number;
  status: ProductStatus;
}

export interface ProductFilter {
  search?: string;
  category_id?: number | null;
  status?: string;
}

/** Shape of every error body returned by the Go API. */
export interface ApiError {
  error: string;
  fields?: Record<string, string>;
}
