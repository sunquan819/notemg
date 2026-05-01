import Vditor from 'vditor';
import 'vditor/dist/index.css';
import { APIClient } from '../api/client';

let vditorInstance: Vditor | null = null;
let saveTimeout: ReturnType<typeof setTimeout> | null = null;

export class EditorView {
  private api: APIClient;
  private noteId?: string;
  private note: any = null;

  constructor(api: APIClient, noteId?: string) {
    this.api = api;
    this.noteId = noteId;
  }

  async render(): Promise<HTMLElement> {
    const el = document.createElement('div');
    el.className = 'editor-view';

    if (!this.noteId) {
      el.innerHTML = `<div class="editor-empty">Select a note or create a new one</div>`;
      return el;
    }

    try {
      this.note = await this.api.getNote(this.noteId);
    } catch (err: any) {
      el.innerHTML = `<div class="editor-empty">Failed to load note: ${err.message}</div>`;
      return el;
    }

    const toolbar = document.createElement('div');
    toolbar.className = 'editor-toolbar';
    toolbar.innerHTML = `
      <input type="text" class="note-title" value="${this.escapeHtml(this.note.title || '')}" placeholder="Untitled" />
      <span style="font-size:12px;color:var(--text-muted)">${this.note.word_count || 0} words</span>
      <button class="btn btn-ghost btn-sm" id="btn-export" title="Export">&#x2B07;</button>
      <button class="btn btn-ghost btn-sm" id="btn-delete" title="Delete" style="color:var(--danger)">&#x1F5D1;</button>
    `;

    const container = document.createElement('div');
    container.className = 'editor-container';
    container.id = 'vditor';

    el.appendChild(toolbar);
    el.appendChild(container);

    requestAnimationFrame(() => {
      this.initVditor(container);
      this.bindToolbarEvents(toolbar, el);
    });

    return el;
  }

  private initVditor(container: HTMLElement) {
    if (vditorInstance) {
      try {
        vditorInstance.destroy();
      } catch {}
    }

    vditorInstance = new Vditor(container.id, {
      height: '100%',
      mode: 'ir',
      theme: 'dark',
      icon: 'ant',
      placeholder: '开始写作... / Start writing...',
      value: this.note?.content || '',
      cache: {
        enable: false,
      },
      toolbar: [
        'headings', 'bold', 'italic', 'strike', '|',
        'list', 'ordered-list', 'check', 'outdent', 'indent', '|',
        'quote', 'code', 'inline-code', 'inline-math', '|',
        'link', 'upload', 'table', 'math', '|',
        'undo', 'redo', '|',
        'fullscreen', 'edit-mode', 'preview',
      ],
      options: {
        markdown: {
          toc: true,
          mark: true,
          footnotes: true,
          autoSpace: true,
        },
        math: {
          inlineDigit: true,
          engine: 'KaTeX',
        },
      },
      upload: {
        url: '/api/attachments/upload',
        headers: {
          Authorization: `Bearer ${localStorage.getItem('access_token')}`,
        },
        accept: 'image/*,.pdf,.doc,.docx',
        handler: async (files: File[]) => {
          for (const file of files) {
            try {
              const result = await this.api.upload('/attachments/upload', file);
              const imgMd = `![${file.name}](/api/attachments/${result.id})\n`;
              vditorInstance?.insertValue(imgMd);
            } catch (err: any) {
              console.error('Upload failed:', err);
            }
          }
          return null;
        },
      },
      input: (value: string) => {
        this.scheduleSave(value);
      },
      after: () => {
        vditorInstance?.setValue(this.note?.content || '', true);
      },
    });
  }

  private scheduleSave(content: string) {
    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(() => {
      this.doSave(content);
    }, 1000);
  }

  private async doSave(content: string) {
    if (!this.noteId) return;
    try {
      const title = this.extractTitle(content);
      this.note = await this.api.updateNote(this.noteId, { content, title });
    } catch (err) {
      console.error('Auto-save failed:', err);
    }
  }

  private extractTitle(content: string): string {
    const lines = content.split('\n');
    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed.startsWith('# ')) {
        return trimmed.slice(2).trim();
      }
    }
    return '';
  }

  private bindToolbarEvents(toolbar: HTMLElement, el: HTMLElement) {
    const titleInput = toolbar.querySelector('.note-title') as HTMLInputElement;
    titleInput?.addEventListener('change', async () => {
      if (!this.noteId) return;
      try {
        this.note = await this.api.updateNote(this.noteId, { title: titleInput.value });
      } catch (err) {
        console.error('Title update failed:', err);
      }
    });

    toolbar.querySelector('#btn-export')?.addEventListener('click', async () => {
      if (!this.noteId) return;
      try {
        await this.api.exportNote(this.noteId, 'markdown');
      } catch (err: any) {
        alert('Export failed: ' + err.message);
      }
    });

    toolbar.querySelector('#btn-delete')?.addEventListener('click', async () => {
      if (!this.noteId) return;
      if (!confirm('Are you sure you want to delete this note?')) return;
      try {
        await this.api.deleteNote(this.noteId);
        window.location.hash = '#/editor';
      } catch (err: any) {
        alert('Delete failed: ' + err.message);
      }
    });
  }

  private escapeHtml(str: string): string {
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }
}
