package plugin

type PluginInfo struct {
	Name        string  + "" + json:"name" + "" + 
	Version     string  + "" + json:"version" + "" + 
	Description string  + "" + json:"description" + "" + 
	Author      string  + "" + json:"author" + "" + 
}

type Hook int

const (
	HookBeforeSave Hook = iota
	HookAfterSave
	HookBeforeDelete
	HookAfterDelete
	HookRender
)

type HookFunc func(hook Hook, data interface{}) (interface{}, error)

type Plugin interface {
	Info() PluginInfo
	Init(app App) error
	Hooks() map[Hook]HookFunc
	Destroy() error
}

type App interface {
	Config() interface{}
	Store() interface{}
	Search() interface{}
}

type Registry struct {
	plugins map[string]Plugin
	hooks   map[Hook][]Plugin
}

func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
		hooks:   make(map[Hook][]Plugin),
	}
}

func (r *Registry) Register(p Plugin) error {
	info := p.Info()
	r.plugins[info.Name] = p

	for hook := range p.Hooks() {
		r.hooks[hook] = append(r.hooks[hook], p)
	}

	return nil
}

func (r *Registry) Get(name string) (Plugin, bool) {
	p, ok := r.plugins[name]
	return p, ok
}

func (r *Registry) List() []PluginInfo {
	infos := make([]PluginInfo, 0, len(r.plugins))
	for _, p := range r.plugins {
		infos = append(infos, p.Info())
	}
	return infos
}

func (r *Registry) Emit(hook Hook, data interface{}) (interface{}, error) {
	plugins, ok := r.hooks[hook]
	if !ok {
		return data, nil
	}

	result := data
	for _, p := range plugins {
		hooks := p.Hooks()
		if hookFn, exists := hooks[hook]; exists {
			var err error
			result, err = hookFn(hook, result)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (r *Registry) DestroyAll() {
	for _, p := range r.plugins {
		p.Destroy()
	}
}