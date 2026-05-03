import { APIClient } from '../api/client';

export class DriveView {
  private api: APIClient;
  private currentFolder: string | null = null;
  private breadcrumb: any[] = [];
  private files: any[] = [];

  constructor(api: APIClient) {
    this.api = api;
  }

  async render(): Promise<HTMLElement> {
    const el = document.createElement('div');
    el.className = 'drive-view';

    el.innerHTML = `
      <div class="drive-header">
        <h2>云盘 / Drive</h2>
        <div class="drive-actions">
          <button class="btn btn-primary btn-sm" id="btn-upload">上传 / Upload</button>
          <button class="btn btn-ghost btn-sm" id="btn-new-folder">新建文件夹 / New Folder</button>
        </div>
      </div>
      <div class="drive-toolbar">
        <div class="drive-breadcrumb" id="breadcrumb">
          <span class="breadcrumb-item" data-id="">根目录 / Root</span>
        </div>
        <div class="drive-search">
          <input type="text" id="drive-search-input" placeholder="搜索文件... / Search files..." />
        </div>
      </div>
      <div class="drive-content scrollbar" id="drive-content"></div>
      <input type="file" id="file-upload-input" style="display:none" multiple />
    `;

    this.bindEvents(el);
    await this.loadFiles();

    return el;
  }

  private bindEvents(el: HTMLElement) {
    el.querySelector('#btn-upload')?.addEventListener('click', () => {
      (el.querySelector('#file-upload-input') as HTMLInputElement).click();
    });

    el.querySelector('#file-upload-input')?.addEventListener('change', async (e) => {
      const files = (e.target as HTMLInputElement).files;
      if (!files?.length) return;

      for (const file of files) {
        try {
          await this.api.driveUpload(file, this.currentFolder);
        } catch (err: any) {
          alert('Upload failed: ' + err.message);
        }
      }

      (e.target as HTMLInputElement).value = '';
      await this.loadFiles();
    });

    el.querySelector('#btn-new-folder')?.addEventListener('click', async () => {
      const name = prompt('文件夹名称 / Folder name:');
      if (!name?.trim()) return;

      try {
        await this.api.createDriveFolder(name.trim(), this.currentFolder);
        await this.loadFiles();
      } catch (err: any) {
        alert(err.message);
      }
    });

    el.querySelector('#drive-search-input')?.addEventListener('input', async (e) => {
      const q = (e.target as HTMLInputElement).value.trim();
      if (q) {
        await this.searchFiles(q);
      } else {
        await this.loadFiles();
      }
    });
  }

  async loadFiles() {
    try {
      const result = await this.api.listDriveFiles(this.currentFolder);
      this.files = result?.files || [];
      this.breadcrumb = result?.path || [];
      this.renderFiles();
      this.renderBreadcrumb();
    } catch (err: any) {
      console.error('Load files failed:', err);
    }
  }

  async searchFiles(query: string) {
    try {
      const files = await this.api.searchDrive(query);
      this.files = files || [];
      this.breadcrumb = [];
      this.renderFiles();
      this.renderBreadcrumb();
    } catch (err: any) {
      console.error('Search failed:', err);
    }
  }

  private renderBreadcrumb() {
    const container = document.getElementById('breadcrumb');
    if (!container) return;

    let html = '<span class="breadcrumb-item" data-id="">根目录 / Root</span>';
    for (const item of this.breadcrumb) {
      html += '<span class="breadcrumb-sep">/</span>';
      html += '<span class="breadcrumb-item" data-id="' + item.id + '">' + this.escapeHtml(item.name) + '</span>';
    }

    container.innerHTML = html;

    container.querySelectorAll('.breadcrumb-item').forEach(item => {
      item.addEventListener('click', async () => {
        const id = (item as HTMLElement).dataset.id || null;
        this.currentFolder = id === '' ? null : id;
        await this.loadFiles();
      });
    });
  }

  private renderFiles() {
    const container = document.getElementById('drive-content');
    if (!container) return;

    if (this.files.length === 0) {
      container.innerHTML = '<div class="drive-empty">暂无文件 / No files</div>';
      return;
    }

    container.innerHTML = this.files.map((f: any) => this.renderFileItem(f)).join('');

    container.querySelectorAll('.drive-item').forEach(item => {
      const id = (item as HTMLElement).dataset.id;
      const type = (item as HTMLElement).dataset.type;

      item.addEventListener('click', () => {
        if (type === 'folder') {
          this.currentFolder = id || null;
          this.loadFiles();
        } else {
          this.previewFile(id || '');
        }
      });

      item.addEventListener('contextmenu', (e: Event) => {
        e.preventDefault();
        this.showContextMenu(e as MouseEvent, id || '', type || '', item as HTMLElement);
      });
    });
  }

  private renderFileItem(f: any): string {
    const icon = this.getFileIcon(f);
    const size = f.type === 'folder' ? '' : this.formatSize(f.size);
    const thumb = f.thumb_path && f.type === 'file' && f.mime_type?.startsWith('image/')
      ? '<img class="drive-thumb" src="/api/notes/drive/files/' + f.id + '/thumbnail" />'
      : '';

    return `
      <div class="drive-item" data-id="${f.id}" data-type="${f.type}">
        ${thumb || '<div class="drive-icon">' + icon + '</div>'}
        <div class="drive-name">${this.escapeHtml(f.name)}</div>
        <div class="drive-meta">${size}</div>
      </div>
    `;
  }

  private getFileIcon(f: any): string {
    if (f.type === 'folder') return '&#x1F4C1;';

    const mime = f.mime_type || '';
    if (mime.startsWith('image/')) return '&#x1F4F7;';
    if (mime.startsWith('video/')) return '&#x1F3AC;';
    if (mime.startsWith('audio/')) return '&#x1F3B5;';
    if (mime === 'application/pdf') return '&#x1F4C4;';
    if (mime.includes('word') || mime.includes('document')) return '&#x1F4DD;';
    if (mime.includes('excel') || mime.includes('spreadsheet')) return '&#x1F4CA;';
    if (mime === 'application/zip') return '&#x1F4E6;';

    return '&#x1F4C4;';
  }

  private formatSize(size: number): string {
    if (size < 1024) return size + ' B';
    if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB';
    if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(1) + ' MB';
    return (size / 1024 / 1024 / 1024).toFixed(1) + ' GB';
  }

  private previewFile(id: string) {
    const file = this.files.find((f: any) => f.id === id);
    if (!file) return;

    const mime = file.mime_type || '';
    const url = '/api/notes/drive/files/' + id + '/preview';

    if (mime.startsWith('image/')) {
      this.showImagePreview(url, file.name);
    } else if (mime.startsWith('video/')) {
      this.showVideoPreview(url, file.name);
    } else if (mime === 'application/pdf') {
      this.showPdfPreview(url, file.name);
    } else {
      this.downloadFile(id, file.name);
    }
  }

  private showImagePreview(url: string, name: string) {
    const modal = document.createElement('div');
    modal.className = 'preview-modal';
    modal.innerHTML = `
      <div class="preview-content">
        <div class="preview-header">
          <span>${this.escapeHtml(name)}</span>
          <button class="btn btn-ghost btn-sm" id="close-preview">&times;</button>
        </div>
        <div class="preview-body">
          <img src="${url}" style="max-width:100%;max-height:80vh;" />
        </div>
      </div>
    `;

    document.body.appendChild(modal);
    modal.querySelector('#close-preview')?.addEventListener('click', () => modal.remove());
    modal.addEventListener('click', (e) => {
      if (e.target === modal) modal.remove();
    });
  }

  private showVideoPreview(url: string, name: string) {
    const modal = document.createElement('div');
    modal.className = 'preview-modal';
    modal.innerHTML = `
      <div class="preview-content">
        <div class="preview-header">
          <span>${this.escapeHtml(name)}</span>
          <button class="btn btn-ghost btn-sm" id="close-preview">&times;</button>
        </div>
        <div class="preview-body">
          <video src="${url}" controls style="max-width:100%;max-height:80vh;"></video>
        </div>
      </div>
    `;

    document.body.appendChild(modal);
    modal.querySelector('#close-preview')?.addEventListener('click', () => modal.remove());
    modal.addEventListener('click', (e) => {
      if (e.target === modal) modal.remove();
    });
  }

  private showPdfPreview(url: string, name: string) {
    window.open(url, '_blank');
  }

  private downloadFile(id: string, name: string) {
    const url = '/api/notes/drive/files/' + id + '/download';
    const a = document.createElement('a');
    a.href = url;
    a.download = name;
    a.click();
  }

  private showContextMenu(e: MouseEvent, id: string, type: string, item: HTMLElement) {
    const existing = document.querySelector('.context-menu');
    if (existing) existing.remove();

    const menu = document.createElement('div');
    menu.className = 'context-menu';
    menu.style.left = e.clientX + 'px';
    menu.style.top = e.clientY + 'px';

    if (type === 'folder') {
      menu.innerHTML = `
        <div class="context-menu-item" data-action="rename">重命名 / Rename</div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item danger" data-action="delete">删除 / Delete</div>
      `;
    } else {
      menu.innerHTML = `
        <div class="context-menu-item" data-action="download">下载 / Download</div>
        <div class="context-menu-item" data-action="rename">重命名 / Rename</div>
        <div class="context-menu-divider"></div>
        <div class="context-menu-item danger" data-action="delete">删除 / Delete</div>
      `;
    }

    document.body.appendChild(menu);

    menu.querySelectorAll('.context-menu-item').forEach(mi => {
      mi.addEventListener('click', async () => {
        const action = (mi as HTMLElement).dataset.action;
        menu.remove();

        const file = this.files.find((f: any) => f.id === id);

        if (action === 'rename') {
          const newName = prompt('新名称 / New name:', file?.name);
          if (newName?.trim()) {
            await this.api.renameDriveFile(id, newName.trim());
            await this.loadFiles();
          }
        } else if (action === 'download') {
          this.downloadFile(id, file?.name || '');
        } else if (action === 'delete') {
          if (confirm('确定删除? / Are you sure?')) {
            await this.api.deleteDriveFile(id);
            await this.loadFiles();
          }
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