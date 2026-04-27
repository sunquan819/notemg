type RouteHandler = (params?: Record<string, string>) => Promise<void>;

export class Router {
  private routes: Map<string, RouteHandler> = new Map();
  private defaultHandler: RouteHandler | null = null;

  addRoute(name: string, handler: RouteHandler) {
    if (name === 'default') {
      this.defaultHandler = handler;
    } else {
      this.routes.set(name, handler);
    }
  }

  start() {
    window.addEventListener('hashchange', () => this.handleRoute());
    window.addEventListener('popstate', () => this.handleRoute());
    this.handleRoute();
  }

  navigate(route: string, params?: Record<string, string>) {
    let hash = '#/' + route;
    if (params) {
      const qs = new URLSearchParams(params).toString();
      if (qs) hash += '?' + qs;
    }
    window.location.hash = hash;
  }

  private handleRoute() {
    const hash = window.location.hash.slice(1) || '/';
    const [path, queryString] = hash.split('?');
    const parts = path.split('/').filter(Boolean);

    const routeName = parts[0] || '';
    const params: Record<string, string> = {};

    if (queryString) {
      const searchParams = new URLSearchParams(queryString);
      searchParams.forEach((value, key) => {
        params[key] = value;
      });
    }

    if (parts.length > 1) {
      params['id'] = parts.slice(1).join('/');
    }

    const handler = this.routes.get(routeName);
    if (handler) {
      handler(params);
    } else if (this.defaultHandler) {
      this.defaultHandler(params);
    }
  }

  getCurrentRoute(): string {
    const hash = window.location.hash.slice(1) || '/';
    const parts = hash.split('/').filter(Boolean);
    return parts[0] || '';
  }
}
