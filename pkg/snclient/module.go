package snclient

import (
	"fmt"
)

// Module is a generic module interface to abstract optional agent functionality
type Module interface {
	Defaults() ConfigData
	Init(snc *Agent, section *ConfigSection, cfg *Config, set *ModuleSet) error
	Start() error
	Stop()
}

// ModuleSet is a list of modules sharing a common type
type ModuleSet struct {
	noCopy  noCopy
	name    string
	modules map[string]Module
}

func NewModuleSet(name string) *ModuleSet {
	ms := &ModuleSet{
		name:    name,
		modules: make(map[string]Module),
	}

	return ms
}

func (ms *ModuleSet) Stop() {
	for _, t := range ms.modules {
		t.Stop()
	}
}

func (ms *ModuleSet) StopRemove() {
	for name, t := range ms.modules {
		t.Stop()
		delete(ms.modules, name)
	}
}

func (ms *ModuleSet) Start() {
	for name := range ms.modules {
		ms.startModule(name)
	}
}

func (ms *ModuleSet) Get(name string) (task Module) {
	if task, ok := ms.modules[name]; ok {
		return task
	}

	return nil
}

func (ms *ModuleSet) Add(name string, task Module) error {
	if _, ok := ms.modules[name]; ok {
		if mod, ok := task.(RequestHandler); ok {
			name = name + ":" + mod.Type()
		}
		if _, ok := ms.modules[name]; ok {
			return fmt.Errorf("duplicate %s module with name: %s", ms.name, name)
		}
	}
	ms.modules[name] = task

	return nil
}

func (ms *ModuleSet) startModule(name string) {
	module, ok := ms.modules[name]
	if !ok {
		log.Errorf("no %s module with name: %s", ms.name, name)

		return
	}

	err := module.Start()
	if err != nil {
		log.Errorf("failed to start %s %s module: %s", name, ms.name, err.Error())
		module.Stop()
		delete(ms.modules, name)

		return
	}

	log.Tracef("module %s started", name)
}

// LoadableModule is a module which can be enabled by config
type LoadableModule struct {
	noCopy    noCopy
	ModuleKey string
	ConfigKey string
	Creator   func() Module
	Handler   *Module
}

// RegisterModule creates a new Module object and puts it on the list of available modules.
func RegisterModule(list *[]*LoadableModule, moduleKey, confKey string, creator func() Module) {
	module := LoadableModule{
		ModuleKey: moduleKey,
		ConfigKey: confKey,
		Creator:   creator,
	}
	*list = append(*list, &module)
}

// Name returns name of this module
func (lm *LoadableModule) Name() string {
	return lm.ModuleKey
}

// Init creates the actual TaskHandler for this task
func (lm *LoadableModule) Init(snc *Agent, conf *Config, set *ModuleSet) (Module, error) {
	handler := lm.Creator()

	modConf := conf.Section(lm.ConfigKey)
	modConf.MergeSection(conf.Section("/settings/default"))
	modConf.MergeData(handler.Defaults())
	conf.ReplaceMacrosDefault(modConf)

	err := handler.Init(snc, modConf, conf, set)
	if err != nil {
		return nil, fmt.Errorf("%s init failed: %s", lm.Name(), err.Error())
	}

	return handler, nil
}
