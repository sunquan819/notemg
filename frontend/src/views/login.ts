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
    
    const title = 'NoteMG';
    const hint = this.initialized ? '请输入密码继续 / Enter password to continue' : '设置密码开始使用 / Set password to get started';
    const pwdLabel = this.initialized ? '密码 / Password' : '设置密码 / Set Password';
    const pwdPlaceholder = this.initialized ? '输入密码 / Enter password' : '至少6个字符 / At least 6 characters';
    const confirmLabel = '确认密码 / Confirm Password';
    const confirmPlaceholder = '再次输入密码 / Confirm password';
    const btnText = this.initialized ? '登录 / Sign In' : '初始化 / Initialize';
    const pwdMinError = '密码至少6个字符 / Password must be at least 6 characters';
    const pwdMismatchError = '密码不一致 / Passwords do not match';
    const authFailedError = '认证失败 / Authentication failed';

    el.innerHTML = `
      <div class="login-card">
        <h1>${title}</h1>
        <p>${hint}</p>
        <form id="login-form">
          ${!this.initialized ? `
            <div class="form-group">
              <label>${pwdLabel}</label>
              <input type="password" id="password" placeholder="${pwdPlaceholder}" autofocus />
            </div>
            <div class="form-group">
              <label>${confirmLabel}</label>
              <input type="password" id="password-confirm" placeholder="${confirmPlaceholder}" />
            </div>
          ` : `
            <div class="form-group">
              <label>${pwdLabel}</label>
              <input type="password" id="password" placeholder="${pwdPlaceholder}" autofocus />
            </div>
          `}
          <button type="submit" class="btn btn-primary">${btnText}</button>
        </form>
      </div>
    `;

    el.querySelector('#login-form')!.addEventListener('submit', async (e) => {
      e.preventDefault();
      const password = (el.querySelector('#password') as HTMLInputElement).value;
      if (!password || password.length < 6) {
        alert(pwdMinError);
        return;
      }

      try {
        if (!this.initialized) {
          const confirm = (el.querySelector('#password-confirm') as HTMLInputElement)?.value;
          if (password !== confirm) {
            alert(pwdMismatchError);
            return;
          }
          await this.api.init(password);
        } else {
          await this.api.login(password);
        }
        this.router.navigate('editor');
      } catch (err: any) {
        alert(err.message || authFailedError);
      }
    });

    return el;
  }
}