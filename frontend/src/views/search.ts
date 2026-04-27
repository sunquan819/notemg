import { APIClient } from '../api/client';

export class SearchView {
  private api: APIClient;
  private query?: string;

  constructor(api: APIClient, query?: string) {
    this.api = api;
    this.query = query;
  }

  async render(): Promise<HTMLElement> {
    const el = document.createElement('div');
    el.className = 'search-view';
    el.innerHTML = `
      <h2>Search</h2>
      <input type="text" class="search-input" placeholder="Search notes..." value="${this.escapeAttr(this.query || '')}" autofocus />
      <div class="search-results scrollbar"></div>
    `;

    const input = el.querySelector('.search-input') as HTMLInputElement;
    const results = el.querySelector('.search-results') as HTMLElement;

    let searchTimeout: ReturnType<typeof setTimeout>;

    input.addEventListener('input', () => {
      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(async () => {
        const q = input.value.trim();
        if (!q) {
          results.innerHTML = '';
          return;
        }
        try {
          const data = await this.api.search(q);
          this.renderResults(results, data || []);
        } catch (err: any) {
          results.innerHTML = `<p style="color:var(--text-muted)">${err.message}</p>`;
        }
      }, 300);
    });

    if (this.query) {
      input.dispatchEvent(new Event('input'));
    }

    return el;
  }

  private renderResults(container: HTMLElement, results: any[]) {
    if (results.length === 0) {
      container.innerHTML = '<p style="color:var(--text-muted)">No results found</p>';
      return;
    }

    container.innerHTML = results.map((r: any) => `
      <div class="search-result-item" data-id="${r.id}">
        <div class="title">${this.escapeHtml(r.title || 'Untitled')}</div>
        <div class="meta">Score: ${r.score?.toFixed(2) || 'N/A'}</div>
      </div>
    `).join('');

    container.querySelectorAll('.search-result-item').forEach(item => {
      item.addEventListener('click', () => {
        const id = (item as HTMLElement).dataset.id;
        if (id) window.location.hash = `#/editor/${id}`;
      });
    });
  }

  private escapeHtml(str: string): string {
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  private escapeAttr(str: string): string {
    return str.replace(/&/g, '&amp;').replace(/"/g, '&quot;');
  }
}
