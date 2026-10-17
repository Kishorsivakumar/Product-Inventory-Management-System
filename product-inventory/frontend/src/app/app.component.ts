import { Component } from '@angular/core';
import { RouterLink, RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink],
  template: `
    <header class="topbar">
      <a routerLink="/products" class="brand">Stockroom<span class="dot">.</span></a>
      <span class="tagline">Product inventory</span>
      <nav class="topnav">
        <a routerLink="/products">All products</a>
        <a routerLink="/products/new" class="nav-cta">+ Add product</a>
      </nav>
    </header>
    <main class="page"><router-outlet /></main>
  `,
})
export class AppComponent {}
