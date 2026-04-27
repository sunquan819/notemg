import { APIClient } from '../api/client';
import { Router } from '../router';

export class TagsView {
  private api: APIClient;
  private router: Router;

  constructor(api: APIClient, router: Router) {
    this.api = api;
    this.router = router;
  }

  async render(): Promise<HTMLElement> {
    const el = document.createElement('div');
    el.className = 'tags-view';

    let tags: any[] = [];
    try {
      tags = await this.api.listTags();
    } catch (err: any) {
      el.innerHTML = `<p style="color:var(--danger)">Failed to load tags: ${err.message}</p>`;
      return el;
    }

    el.innerHTML = `
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:16px">
        <h2>Tags</h2>
        <button class="btn btn-primary btn-sm" id="btn-new-tag">+ New Tag</button>
      </div>
      <div class="tags-list"></div>
    `;

    const listEl = el.querySelector('.tags-list') as HTMLElement;
    this.renderTags(listEl, tags);

    el.querySelector('#btn-new-tag')?.addEventListener('click', async () => {
      const name = prompt('Tag name:');
      if (!name?.trim()) return;
      try {
        await this.api.createTag({ name: name.trim() });
        tags = await this.api.listTags();
        this.renderTags(listEl, tags);
      } catch (err: any) {
        alert(err.message);
      }
    });

    return el;
  }

  private renderTags(container: HTMLElement, tags: any[]) {
    if (tags.length === 0) {
      container.innerHTML = '<p style="color:var(--text-muted)">No tags yet. Create one to get started.</p>';
      return;
    }

    container.innerHTML = tags.map((t: any) => `
      <div class="tag-badge" data-id="${t.id}">
        <span>${this.escapeHtml(t.name)}</span>
        <span class="count">${t.note_count || 0}</span>
        <button class="btn-ghost btn-sm tag-delete" data-id="${t.id}" title="Delete" style="font-size:10px;padding:0 2px;color:var(--danger)">&times;</button>
      </div>
    `).join('');

    container.querySelectorAll('.tag-badge').forEach(badge => {
      badge.addEventListener('click', (e) => {
        if ((e.target as HTMLElement).classList.contains('tag-delete')) return;
        const id = (badge as HTMLElement).dataset.id;
        if (id) window.location.hash = `#/editor?tag_id=${id}`;
      });
    });

    container.querySelectorAll('.tag-delete').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        e.stopPropagation();
        const id = (e.target as HTMLElement).dataset.id;
        if (!id || !confirm('Delete this tag?')) return;
        try {
          await this.api.deleteTag(id);
          const freshTags = await this.api.listTags();
          this.renderTags(container, freshTags);
        } catch (err: any) {
          alert(err.message);
        }
      });
    });
  }

  private escapeHtml(str: string): string {
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }
}
