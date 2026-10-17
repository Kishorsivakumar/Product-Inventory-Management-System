import { HttpClient, HttpErrorResponse, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { ApiError, Category, Product, ProductFilter, ProductInput, StockMode } from './models';

@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);
  private base = environment.apiUrl;

  getCategories(): Observable<Category[]> {
    return this.http.get<Category[]>(`${this.base}/categories`);
  }

  getProducts(filter: ProductFilter = {}): Observable<Product[]> {
    let params = new HttpParams();
    if (filter.search?.trim()) params = params.set('search', filter.search.trim());
    if (filter.category_id) params = params.set('category_id', filter.category_id);
    if (filter.status) params = params.set('status', filter.status);
    return this.http.get<Product[]>(`${this.base}/products`, { params });
  }

  getProduct(id: number): Observable<Product> {
    return this.http.get<Product>(`${this.base}/products/${id}`);
  }

  createProduct(input: ProductInput): Observable<Product> {
    return this.http.post<Product>(`${this.base}/products`, input);
  }

  updateProduct(id: number, input: ProductInput): Observable<Product> {
    return this.http.put<Product>(`${this.base}/products/${id}`, input);
  }

  deleteProduct(id: number): Observable<void> {
    return this.http.delete<void>(`${this.base}/products/${id}`);
  }

  updateStock(id: number, mode: StockMode, quantity: number): Observable<Product> {
    return this.http.patch<Product>(`${this.base}/products/${id}/stock`, { mode, quantity });
  }
}

/** Normalises any HTTP failure into a message plus optional per-field messages. */
export function parseError(err: unknown): ApiError {
  if (err instanceof HttpErrorResponse) {
    if (err.status === 0) {
      return { error: 'Cannot reach the server. Check that the API is running.' };
    }
    const body = err.error as Partial<ApiError> | null;
    if (body && typeof body === 'object' && body.error) {
      return { error: body.error, fields: body.fields };
    }
    return { error: `Request failed (${err.status}).` };
  }
  return { error: 'Something went wrong.' };
}
