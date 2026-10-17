import { CurrencyPipe } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { Subject, debounceTime } from 'rxjs';
import { ApiService, parseError } from '../core/api.service';
import { Category, Product, StockMode } from '../core/models';

@Component({
  selector: 'app-product-list',
  standalone: true,
  imports: [FormsModule, RouterLink, CurrencyPipe],
  templateUrl: './product-list.component.html',
})
export class ProductListComponent implements OnInit {
  private api = inject(ApiService);

  products = signal<Product[]>([]);
  categories = signal<Category[]>([]);
  loading = signal(true);
  error = signal('');
  notice = signal('');

  search = '';
  categoryId: number | null = null;
  status = '';
  private search$ = new Subject<void>();

  // delete confirmation
  toDelete = signal<Product | null>(null);
  deleting = signal(false);

  // stock dialog
  stockTarget = signal<Product | null>(null);
  stockMode: StockMode = 'add';
  stockQty: number | null = null;
  stockTouched = false;
  stockError = signal('');
  stockSaving = signal(false);

  ngOnInit(): void {
    this.api.getCategories().subscribe({
      next: (c) => this.categories.set(c),
      error: (e) => this.error.set(parseError(e).error),
    });
    this.search$.pipe(debounceTime(300)).subscribe(() => this.load());
    this.load();
  }

  onSearchInput(): void {
    this.search$.next();
  }

  load(): void {
    this.loading.set(true);
    this.api
      .getProducts({ search: this.search, category_id: this.categoryId, status: this.status })
      .subscribe({
        next: (p) => {
          this.products.set(p);
          this.error.set('');
          this.loading.set(false);
        },
        error: (e) => {
          this.error.set(parseError(e).error);
          this.loading.set(false);
        },
      });
  }

  setCategory(id: number | null): void {
    this.categoryId = id;
    this.load();
  }

  setStatus(value: string): void {
    this.status = value;
    this.load();
  }

  // ---- category images (src/assets/categories/<slug>.<ext>) ----
  private static readonly IMG_EXTS = ['jpg', 'png', 'webp', 'svg'];

  slug(name: string): string {
    return name.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '');
  }

  imgSrc(category: string): string {
    return `assets/categories/${this.slug(category)}.${ProductListComponent.IMG_EXTS[0]}`;
  }

  /** Tries the next extension, then falls back to default.svg. */
  onImgError(event: Event, category: string): void {
    const img = event.target as HTMLImageElement;
    const step = Number(img.dataset['step'] ?? 0) + 1;
    img.dataset['step'] = String(step);
    const exts = ProductListComponent.IMG_EXTS;
    if (step < exts.length) {
      img.src = `assets/categories/${this.slug(category)}.${exts[step]}`;
    } else if (!img.src.endsWith('default.svg')) {
      img.src = 'assets/categories/default.svg';
    }
  }

  stockLabel(p: Product): string {
    if (p.quantity === 0) return 'Out of stock';
    if (p.low_stock) return `Only ${p.quantity} left in stock`;
    return `In stock (${p.quantity})`;
  }

  get categoryName(): string {
    return this.categories().find((c) => c.id === this.categoryId)?.name ?? '';
  }

  clearFilters(): void {
    this.search = '';
    this.categoryId = null;
    this.status = '';
    this.load();
  }

  get hasFilters(): boolean {
    return !!(this.search || this.categoryId || this.status);
  }

  // ---- delete ----
  askDelete(p: Product): void {
    this.toDelete.set(p);
  }

  confirmDelete(): void {
    const p = this.toDelete();
    if (!p) return;
    this.deleting.set(true);
    this.api.deleteProduct(p.id).subscribe({
      next: () => {
        this.deleting.set(false);
        this.toDelete.set(null);
        this.flash(`Deleted ${p.name}.`);
        this.load();
      },
      error: (e) => {
        this.deleting.set(false);
        this.toDelete.set(null);
        this.error.set(parseError(e).error);
      },
    });
  }

  // ---- stock ----
  openStock(p: Product): void {
    this.stockTarget.set(p);
    this.stockMode = 'add';
    this.stockQty = null;
    this.stockTouched = false;
    this.stockError.set('');
  }

  /** Client-side message for the quantity field ('' when valid). */
  get stockQtyMessage(): string {
    const q = this.stockQty;
    if (q === null || (q as unknown) === '') return 'This field is required.';
    if (!Number.isInteger(q)) return 'Quantity must be a whole number.';
    if (q < 0) return 'Quantity cannot be negative.';
    return '';
  }

  saveStock(): void {
    const p = this.stockTarget();
    if (!p) return;
    this.stockTouched = true;
    if (this.stockQtyMessage) return;
    this.stockSaving.set(true);
    this.api.updateStock(p.id, this.stockMode, this.stockQty as number).subscribe({
      next: (updated) => {
        this.stockSaving.set(false);
        this.stockTarget.set(null);
        this.products.update((list) => list.map((x) => (x.id === updated.id ? updated : x)));
        this.flash(`Stock for ${updated.name} is now ${updated.quantity}.`);
      },
      error: (e) => {
        const err = parseError(e);
        this.stockSaving.set(false);
        this.stockError.set(err.fields?.['quantity'] ?? err.error);
      },
    });
  }

  private flash(msg: string): void {
    this.notice.set(msg);
    setTimeout(() => this.notice.set(''), 4000);
  }
}
