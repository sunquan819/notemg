import { APIClient } from '../api/client';
import { Router } from '../router';

export class LoginView {
  private api: APIClient;
  private router: Router;
  private initialized: boolean = false;

  constructor(api: APIClient, router: Router) {
    this.api = api;
    this.router = router;
  }

  async render(): Promise<HTMLElement> {
    const status = await this.api.getAuthStatus().catch(() => ({ initialized: false }));
    this.initialized = status?.initialized ?? false;

    const el = document.createElement('div');
    el.className = 'login-view';
    el.innerHTML = `
      <div class="login-card">
        <h1>NoteMG</h1>
        <p>${this.initialized ? 'Enter your password to continue' : 'Set your password to get started'}</p>
        <form id="login-form">
          ${!this.initialized ? `
            <div class="form-group">
              <label>Set Password</label>
              <input type="password" id="password" placeholder="At least 6 characters" autofocus />
            </div>
            <div class="form-group">
              <label>Confirm Password</label>
              <input type="password" id="password-confirm" placeholder="Confirm password" />
            </div>
          ` : `
            <div class="form-group">
              <label>Password</label>
              <input type="password" id="password" placeholder="Enter password" autofocus />
            </div>
          `}
          <button type="submit" class="btn btn-primary">
            ${this.initialized ? 'Sign In' : 'Initialize'}
          </button>
        </form>
      </div>
    `;

    el.querySelector('#login-form')!.addEventListener('submit', async (e) => {
      e.preventDefault();
      const password = (el.querySelector('#password') as HTMLInputElement).value;
      if (!password || password.length < 6) {
        alert('Password must be at least 6 characters');
        return;
      }

      try {
        if (!this.initialized) {
          const confirm = (el.querySelector('#password-confirm') as HTMLInputElement)?.value;
          if (password !== confirm) {
            alert('Passwords do not match');
            return;
          }
          await this.api.init(password);
        } else {
          await this.api.login(password);
        }
        this.router.navigate('editor');
      } catch (err: any) {
        alert(err.message || 'Authentication failed');
      }
    });

    return el;
  }
}
