import { Router } from './router';
import { APIClient } from './api/client';
import { EditorView } from './views/editor';
import { LoginView } from './views/login';
import { SearchView } from './views/search';
import { TagsView } from './views/tags';
import { SettingsView } from './views/settings';
import { DriveView } from './views/drive';
import { Sidebar } from './components/sidebar';
import { Toast } from './components/toast';

export class App {
  private el: HTMLElement;
  private router: Router;
  private api: APIClient;
  private sidebar: Sidebar;
  private mainContent: HTMLElement;
  private toast: Toast;

  constructor(el: HTMLElement) {
    this.el = el;
    this.api = new APIClient();
    this.router = new Router();
    this.toast = new Toast();
    this.sidebar = new Sidebar(this.api, this.router);
    this.mainContent = document.createElement('main');
    this.mainContent.className = 'main-content';
  }

  async start() {
    this.el.innerHTML = '';
    this.el.className = 'app';

    const layout = document.createElement('div');
    layout.className = 'app-layout';

    const sidebarEl = this.sidebar.render();
    layout.appendChild(sidebarEl);
    layout.appendChild(this.mainContent);

    this.el.appendChild(layout);
    this.el.appendChild(this.toast.render());

    this.setupRoutes();
    this.router.start();
  }

  private setupRoutes() {
    this.router.addRoute('login', async () => {
      const view = new LoginView(this.api, this.router);
      this.renderView(view);
    });

    this.router.addRoute('editor', async (params) => {
      if (!this.api.isAuthenticated()) {
        this.router.navigate('login');
        return;
      }
      const view = new EditorView(this.api, params?.id);
      this.renderView(view);
    });

    this.router.addRoute('search', async (params) => {
      if (!this.api.isAuthenticated()) {
        this.router.navigate('login');
        return;
      }
      const view = new SearchView(this.api, params?.q);
      this.renderView(view);
    });

    this.router.addRoute('tags', async () => {
      if (!this.api.isAuthenticated()) {
        this.router.navigate('login');
        return;
      }
      const view = new TagsView(this.api, this.router);
      this.renderView(view);
    });

    this.router.addRoute('settings', async () => {
      if (!this.api.isAuthenticated()) {
        this.router.navigate('login');
        return;
      }
      const view = new SettingsView(this.api);
      this.renderView(view);
    });

    this.router.addRoute('drive', async () => {
      if (!this.api.isAuthenticated()) {
        this.router.navigate('login');
        return;
      }
      const view = new DriveView(this.api);
      this.renderView(view);
    });

    this.router.addRoute('default', async () => {
      if (this.api.isAuthenticated()) {
        this.router.navigate('editor');
      } else {
        this.router.navigate('login');
      }
    });
  }

  private async renderView(view: any) {
    this.mainContent.innerHTML = '';
    try {
      const el = await view.render();
      this.mainContent.appendChild(el);
    } catch (err: any) {
      this.toast.show(err.message || 'Failed to load view', 'error');
    }
  }
}
