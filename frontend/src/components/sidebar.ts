import { APIClient } from '../api/client';
import { Router } from '../router';

export class Sidebar {
  private api: APIClient;
  private router: Router;
  private folders: any[] = [];
  private notes: any[] = [];

  constructor(api: APIClient, router: Router) {
    this.api = api;
    this.router = router;
  }

  render(): HTMLElement {
    const el = document.createElement('aside');
    el.className = 'sidebar scrollbar';

    const header = document.createElement('div');
    header.className = 'sidebar-header';
    header.innerHTML = '<h2>NoteMG</h2><div class="sidebar-actions"><button class="btn btn-ghost btn-sm" id="btn-new-note" title="New Note">+</button><button class="btn btn-ghost btn-sm" id="btn-new-folder" title="New Folder">++</button></div>';

    const search = document.createElement('div');
    search.className = 'sidebar-search';
    search.innerHTML = '<input type="text" id="sidebar-search-input" placeholder="Search..." />';

    const tree = document.createElement('div');
    tree.className = 'sidebar-tree scrollbar';
    tree.id = 'sidebar-tree';

    const bottom = document.createElement('div');
    bottom.className = 'sidebar-bottom';
    bottom.innerHTML = '<button class="btn btn-ghost btn-sm" id="nav-drive" title="Drive">&#x1F4BE;</button><button class="btn btn-ghost btn-sm" id="nav-tags">Tags</button><button class="btn btn-ghost btn-sm" id="nav-search">Search</button><button class="btn btn-ghost btn-sm" id="nav-settings">Settings</button><button class="btn btn-ghost btn-sm" id="nav-logout" style="color:var(--danger);margin-left:auto">Logout</button>';

    el.appendChild(header);
    el.appendChild(search);
    el.appendChild(tree);
    el.appendChild(bottom);

    this.bindEvents(el);
    this.loadData();

    return el;
  }

  private bindEvents(el: HTMLElement) {
    el.querySelector('#btn-new-note')?.addEventListener('click', async () => {
      try {
        const note = await this.api.createNote({ title: '', content: '' });
        if (note?.id) {
          window.location.hash = '#/editor/' + note.id;
          this.loadData();
        }
      } catch (err: any) {
        alert(err.message);
      }
    });

    el.querySelector('#btn-new-folder')?.addEventListener('click', async () => {
      const name = prompt('Folder name:');
      if (!name?.trim()) return;
      try {
        await this.api.createFolder({ name: name.trim() });
        this.loadData();
      } catch (err: any) {
        alert(err.message);
      }
    });

    el.querySelector('#sidebar-search-input')?.addEventListener('input', (e) => {
      const q = (e.target as HTMLInputElement).value.trim();
      if (q) {
        window.location.hash = '#/search?q=' + encodeURIComponent(q);
      }
    });

    el.querySelector('#nav-drive')?.addEventListener('click', () => {
      window.location.hash = '#/drive';
    });

    el.querySelector('#nav-tags')?.addEventListener('click', () => {
      window.location.hash = '#/tags';
    });

    el.querySelector('#nav-search')?.addEventListener('click', () => {
      window.location.hash = '#/search';
    });

    el.querySelector('#nav-settings')?.addEventListener('click', () => {
      window.location.hash = '#/settings';
    });

    el.querySelector('#nav-logout')?.addEventListener('click', () => {
      if (confirm('Sign out?')) this.api.logout();
    });
  }

  async loadData() {
    try {
      const [folders, notes] = await Promise.all([
        this.api.listFolders(),
        this.api.listNotes(),
      ]);
      this.folders = folders || [];
      this.notes = (notes?.notes || notes) || [];
      this.renderTree();
    } catch (err) {
      console.error('Failed to load sidebar data:', err);
    }
  }

  private renderTree() {
    const tree = document.getElementById('sidebar-tree');
    if (!tree) return;

    let html = '';

    const renderFolder = (folder: any, depth: number = 0): string => {
      const indent = 16 + depth * 16;
      let result = '<div class="tree-item tree-folder" data-id="' + folder.id + '" data-type="folder" style="padding-left:' + indent + 'px"><span class="icon">&#x1F4C1;</span><span class="name">' + this.escapeHtml(folder.name) + '</span></div>';
      if (folder.children) {
        for (const child of folder.children) {
          result += renderFolder(child, depth + 1);
        }
      }
      const folderNotes = this.notes.filter((n: any) => n.folder_id === folder.id);
      for (const note of folderNotes) {
        result += '<div class="tree-item" data-id="' + note.id + '" data-type="note" style="padding-left:' + (indent + 16) + 'px"><span class="icon">&#x1F4DD;</span><span class="name">' + this.escapeHtml(note.title || 'Untitled') + '</span></div>';
      }
      return result;
    };

    const rootFolders = this.folders;
    for (const folder of rootFolders) {
      html += renderFolder(folder);
    }

    const unfiled = this.notes.filter((n: any) => !n.folder_id);
    for (const note of unfiled) {
      html += '<div class="tree-item" data-id="' + note.id + '" data-type="note" style="padding-left:16px"><span class="icon">&#x1F4DD;</span><span class="name">' + this.escapeHtml(note.title || 'Untitled') + '</span></div>';
    }

    tree.innerHTML = html;

    tree.querySelectorAll('.tree-item').forEach(item => {
      item.addEventListener('click', () => {
        const id = (item as HTMLElement).dataset.id;
        const type = (item as HTMLElement).dataset.type;
        if (type === 'note' && id) {
          window.location.hash = '#/editor/' + id;
        }
      });

      item.addEventListener('contextmenu', ((e: Event) => {
        e.preventDefault();
        this.showContextMenu(e as MouseEvent, item as HTMLElement);
      }) as EventListener);
    });

    this.highlightActive();
  }

  private highlightActive() {
    const hash = window.location.hash;
    const match = hash.match(/#\/editor\/(.+)$/);
    document.querySelectorAll('.tree-item.active').forEach(el => el.classList.remove('active'));
    if (match) {
      const activeItem = document.querySelector('.tree-item[data-id="' + match[1] + '"]');
      activeItem?.classList.add('active');
    }
  }

  private showContextMenu(e: MouseEvent, item: HTMLElement) {
    const existing = document.querySelector('.context-menu');
    if (existing) existing.remove();

    const id = item.dataset.id;
    const type = item.dataset.type;

    const menu = document.createElement('div');
    menu.className = 'context-menu';
    menu.style.left = e.clientX + 'px';
    menu.style.top = e.clientY + 'px';

    if (type === 'folder') {
      menu.innerHTML = '<div class="context-menu-item" data-action="rename">Rename</div><div class="context-menu-item" data-action="new-note">New Note Here</div><div class="context-menu-divider"></div><div class="context-menu-item danger" data-action="delete">Delete</div>';
    } else {
      menu.innerHTML = '<div class="context-menu-item" data-action="duplicate">Duplicate</div><div class="context-menu-divider"></div><div class="context-menu-item danger" data-action="delete">Delete</div>';
    }

    document.body.appendChild(menu);

    menu.querySelectorAll('.context-menu-item').forEach(mi => {
      mi.addEventListener('click', async () => {
        const action = (mi as HTMLElement).dataset.action;
        menu.remove();

        if (!id) return;

        try {
          if (action === 'rename') {
            const newName = prompt('New name:');
            if (newName?.trim()) {
              await this.api.updateFolder(id, { name: newName.trim() });
              this.loadData();
            }
          } else if (action === 'new-note') {
            const note = await this.api.createNote({ title: '', content: '', folder_id: id });
            if (note?.id) window.location.hash = '#/editor/' + note.id;
            this.loadData();
          } else if (action === 'duplicate') {
            await this.api.duplicateNote(id);
            this.loadData();
          } else if (action === 'delete') {
            if (!confirm('Are you sure?')) return;
            if (type === 'folder') {
              await this.api.deleteFolder(id);
            } else {
              await this.api.deleteNote(id);
            }
            this.loadData();
            window.location.hash = '#/editor';
          }
        } catch (err: any) {
          alert(err.message);
        }
      });
    });

    const closeMenu = () => {
      menu.remove();
      document.removeEventListener('click', closeMenu);
    };
    setTimeout(() => document.addEventListener('click', closeMenu), 0);
  }

  private escapeHtml(str: string): string {
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }
}