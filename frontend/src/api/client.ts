interface APIResponse {
  success: boolean;
  data?: any;
  error?: { code: number; message: string };
}

export class APIClient {
  private baseUrl: string;

  constructor() {
    this.baseUrl = '/api';
  }

  isAuthenticated(): boolean {
    return !!localStorage.getItem('access_token');
  }

  private getToken(): string {
    return localStorage.getItem('access_token') || '';
  }

  private async request(method: string, path: string, body?: any): Promise<any> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    const token = this.getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const opts: RequestInit = {
      method,
      headers,
    };

    if (body && method !== 'GET') {
      if (body instanceof FormData) {
        delete headers['Content-Type'];
        opts.body = body;
      } else {
        opts.body = JSON.stringify(body);
      }
    }

    const res = await fetch(`${this.baseUrl}${path}`, opts);
    const data: APIResponse = await res.json();

    if (!data.success && data.error) {
      if (data.error.code === 401) {
        localStorage.removeItem('access_token');
        window.location.hash = '#/login';
      }
      throw new Error(data.error.message);
    }

    return data.data;
  }

  async get(path: string): Promise<any> {
    return this.request('GET', path);
  }

  async post(path: string, body: any): Promise<any> {
    return this.request('POST', path, body);
  }

  async put(path: string, body: any): Promise<any> {
    return this.request('PUT', path, body);
  }

  async delete(path: string): Promise<any> {
    return this.request('DELETE', path);
  }

  async upload(path: string, file: File): Promise<any> {
    const formData = new FormData();
    formData.append('file', file);
    return this.request('POST', path, formData);
  }

  async login(password: string): Promise<any> {
    const data = await this.request('POST', '/auth/login', { password });
    if (data?.access_token) {
      localStorage.setItem('access_token', data.access_token);
      if (data.refresh_token) {
        localStorage.setItem('refresh_token', data.refresh_token);
      }
    }
    return data;
  }

  async init(password: string): Promise<any> {
    const data = await this.request('POST', '/auth/init', { password });
    if (data?.access_token) {
      localStorage.setItem('access_token', data.access_token);
      if (data.refresh_token) {
        localStorage.setItem('refresh_token', data.refresh_token);
      }
    }
    return data;
  }

  async getAuthStatus(): Promise<any> {
    return this.request('GET', '/auth/status');
  }

  async changePassword(oldPassword: string, newPassword: string): Promise<any> {
    return this.request('PUT', '/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    });
  }

  async listNotes(params?: Record<string, string>): Promise<any> {
    const qs = params ? '?' + new URLSearchParams(params).toString() : '';
    return this.request('GET', '/notes' + qs);
  }

  async getNote(id: string): Promise<any> {
    return this.request('GET', `/notes/${id}`);
  }

  async createNote(data: { title?: string; content?: string; folder_id?: string; tags?: string[] }): Promise<any> {
    return this.request('POST', '/notes', data);
  }

  async updateNote(id: string, data: { title?: string; content?: string; folder_id?: string | null; tags?: string[] }): Promise<any> {
    return this.request('PUT', `/notes/${id}`, data);
  }

  async deleteNote(id: string): Promise<any> {
    return this.request('DELETE', `/notes/${id}`);
  }

  async moveNote(id: string, folderId: string | null): Promise<any> {
    return this.request('POST', `/notes/${id}/move`, { folder_id: folderId });
  }

  async duplicateNote(id: string): Promise<any> {
    return this.request('POST', `/notes/${id}/duplicate`);
  }

  async restoreNote(id: string): Promise<any> {
    return this.request('POST', `/notes/${id}/restore`);
  }

  async permanentDeleteNote(id: string): Promise<any> {
    return this.request('DELETE', `/notes/${id}/permanent`);
  }

  async listFolders(): Promise<any> {
    return this.request('GET', '/folders');
  }

  async createFolder(data: { name: string; parent_id?: string }): Promise<any> {
    return this.request('POST', '/folders', data);
  }

  async updateFolder(id: string, data: { name?: string; parent_id?: string | null }): Promise<any> {
    return this.request('PUT', `/folders/${id}`, data);
  }

  async deleteFolder(id: string): Promise<any> {
    return this.request('DELETE', `/folders/${id}`);
  }

  async listTags(): Promise<any> {
    return this.request('GET', '/tags');
  }

  async createTag(data: { name: string }): Promise<any> {
    return this.request('POST', '/tags', data);
  }

  async deleteTag(id: string): Promise<any> {
    return this.request('DELETE', `/tags/${id}`);
  }

  async search(query: string): Promise<any> {
    return this.request('GET', `/search?q=${encodeURIComponent(query)}`);
  }

  async renderMarkdown(content: string): Promise<any> {
    return this.request('POST', '/markdown/render', { content });
  }

  async exportNote(id: string, format: string = 'markdown'): Promise<void> {
    const token = this.getToken();
    const url = `${this.baseUrl}/export/notes/${id}?format=${format}`;
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) throw new Error('Export failed');
    const blob = await res.blob();
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `note.${format === 'html' ? 'html' : 'md'}`;
    a.click();
    URL.revokeObjectURL(a.href);
  }

  async importMarkdown(file: File, folderId?: string): Promise<any> {
    const formData = new FormData();
    formData.append('file', file);
    if (folderId) formData.append('folder_id', folderId);
    return this.upload('/import/markdown', file);
  }

  async importZip(file: File): Promise<any> {
    const formData = new FormData();
    formData.append('file', file);
    return this.request('POST', '/import/zip', formData);
  }

  async batchExport(noteIds: string[], format: string = 'markdown'): Promise<void> {
    const token = this.getToken();
    const res = await fetch(`${this.baseUrl}/export/batch`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ note_ids: noteIds, format }),
    });
    if (!res.ok) throw new Error('Export failed');
    const blob = await res.blob();
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = 'notemg-export.zip';
    a.click();
    URL.revokeObjectURL(a.href);
  }

  logout() {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    window.location.hash = '#/login';
  }
}
