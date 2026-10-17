import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { ApiService, parseError } from '../core/api.service';
import { Category, ProductInput } from '../core/models';

@Component({
  selector: 'app-product-form',
  standalone: true,
  imports: [ReactiveFormsModule, RouterLink],
  templateUrl: './product-form.component.html',
})
export class ProductFormComponent implements OnInit {
  private fb = inject(FormBuilder);
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);

  categories = signal<Category[]>([]);
  loading = signal(false);
  saving = signal(false);
  error = signal('');
  serverFields = signal<Record<string, string>>({});
  productId: number | null = null;

  form = this.fb.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(200)]],
    sku: ['', [Validators.required, Validators.maxLength(64)]],
    category_id: [0, [Validators.min(1)]],
    price: [0, [Validators.required, Validators.min(0.01)]],
    quantity: [0, [Validators.required, Validators.min(0), Validators.pattern(/^\d+$/)]],
    status: ['active' as ProductInput['status'], Validators.required],
  });

  get isEdit(): boolean {
    return this.productId !== null;
  }

  ngOnInit(): void {
    this.api.getCategories().subscribe({
      next: (c) => this.categories.set(c),
      error: (e) => this.error.set(parseError(e).error),
    });

    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.productId = Number(id);
      this.loading.set(true);
      this.api.getProduct(this.productId).subscribe({
        next: (p) => {
          this.form.patchValue(p);
          this.loading.set(false);
        },
        error: (e) => {
          this.error.set(parseError(e).error);
          this.loading.set(false);
        },
      });
    }

    // Clear a server-side message as soon as the user edits that field.
    this.form.valueChanges.subscribe(() => {
      if (Object.keys(this.serverFields()).length) this.serverFields.set({});
    });
  }

  /** Message for a field: server error first, then client validation. */
  fieldError(name: keyof ProductInput): string {
    const server = this.serverFields()[name];
    if (server) return server;
    const c = this.form.controls[name];
    if (!(c.touched || c.dirty) || !c.errors) return '';
    if (c.errors['required']) return 'This field is required.';
    if (name === 'category_id') return 'Choose a category.';
    if (name === 'price') return 'Price must be greater than zero.';
    if (name === 'quantity') return 'Quantity must be a whole number, zero or higher.';
    if (c.errors['maxlength']) return 'This value is too long.';
    return 'Invalid value.';
  }

  invalid(name: keyof ProductInput): boolean {
    return !!this.fieldError(name);
  }

  submit(): void {
    this.error.set('');
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }
    const input = this.form.getRawValue() as ProductInput;
    input.category_id = Number(input.category_id);
    this.saving.set(true);

    const request = this.isEdit
      ? this.api.updateProduct(this.productId!, input)
      : this.api.createProduct(input);

    request.subscribe({
      next: () => this.router.navigate(['/products']),
      error: (e) => {
        const err = parseError(e);
        this.saving.set(false);
        this.error.set(err.error);
        this.serverFields.set(err.fields ?? {});
      },
    });
  }
}
