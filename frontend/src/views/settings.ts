import { APIClient } from '../api/client';

export class SettingsView {
  private api: APIClient;

  constructor(api: APIClient) {
    this.api = api;
  }

  async render(): Promise<HTMLElement> {
    const el = document.createElement('div');
    el.className = 'settings-view';

    el.innerHTML = `
      <h2>Settings</h2>

      <div class="settings-section">
        <h3>Change Password</h3>
        <div class="form-group">
          <label>Current Password</label>
          <input type="password" id="old-password" />
        </div>
        <div class="form-group">
          <label>New Password</label>
          <input type="password" id="new-password" />
        </div>
        <div class="form-group">
          <label>Confirm New Password</label>
          <input type="password" id="confirm-password" />
        </div>
        <button class="btn btn-primary" id="btn-change-password">Change Password</button>
      </div>

      <div class="settings-section">
        <h3>Import</h3>
        <p style="color:var(--text-secondary);font-size:13px;margin-bottom:8px">Import Markdown files or ZIP archives</p>
        <input type="file" id="import-file" accept=".md,.zip" />
        <button class="btn btn-ghost" id="btn-import" style="margin-top:8px">Import</button>
      </div>

      <div class="settings-section">
        <h3>Account</h3>
        <button class="btn btn-danger" id="btn-logout">Sign Out</button>
      </div>
    `;

    el.querySelector('#btn-change-password')?.addEventListener('click', async () => {
      const oldPw = (el.querySelector('#old-password') as HTMLInputElement).value;
      const newPw = (el.querySelector('#new-password') as HTMLInputElement).value;
      const confirmPw = (el.querySelector('#confirm-password') as HTMLInputElement).value;

      if (newPw.length < 6) {
        alert('Password must be at least 6 characters');
        return;
      }
      if (newPw !== confirmPw) {
        alert('Passwords do not match');
        return;
      }

      try {
        await this.api.changePassword(oldPw, newPw);
        alert('Password changed successfully');
        (el.querySelector('#old-password') as HTMLInputElement).value = '';
        (el.querySelector('#new-password') as HTMLInputElement).value = '';
        (el.querySelector('#confirm-password') as HTMLInputElement).value = '';
      } catch (err: any) {
        alert(err.message);
      }
    });

    el.querySelector('#btn-import')?.addEventListener('click', async () => {
      const fileInput = el.querySelector('#import-file') as HTMLInputElement;
      if (!fileInput.files?.length) {
        alert('Please select a file');
        return;
      }
      const file = fileInput.files[0];
      try {
        if (file.name.endsWith('.zip')) {
          const result = await this.api.importZip(file);
          alert(`Imported ${result.count} notes`);
        } else {
          await this.api.importMarkdown(file);
          alert('Note imported');
        }
      } catch (err: any) {
        alert('Import failed: ' + err.message);
      }
    });

    el.querySelector('#btn-logout')?.addEventListener('click', () => {
      if (confirm('Sign out?')) {
        this.api.logout();
      }
    });

    return el;
  }
}
