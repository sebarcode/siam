package siam

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
	"github.com/sebarcode/logger"
)

type Options struct {
	Storage           Storage
	AllowMultiSession bool
	MultiSession      int
	SecondLifeTime    int
}

type Manager struct {
	pool           *SessionPool
	secondLifeTime int
	opts           Options
}

func New(logger *logger.LogEngine, secondLifeTime int, opt *Options) *Manager {
	ae := new(Manager)
	ae.pool = NewSessionPool(logger)
	if secondLifeTime == 0 {
		secondLifeTime = 60 * 60 * 24 * 7
	}
	ae.secondLifeTime = secondLifeTime
	if opt == nil {
		ae.opts = Options{}
	} else {
		ae.opts = *opt
	}
	if ae.opts.SecondLifeTime == 0 {
		ae.opts.SecondLifeTime = ae.secondLifeTime
	}
	return ae
}

func (m *Manager) Options() Options {
	return m.opts
}

func (a *Manager) Create(ctx *kaos.Context, parm codekit.M, data codekit.M) (*Session, error) {
	id := parm.GetString("ID")
	duration := parm.GetInt("Second")
	if duration == 0 {
		duration = a.secondLifeTime
	}

	if id == "" {
		return nil, errors.New("ID is mandatory")
	}

	s, e := a.pool.Create(id, data, duration)
	if e == nil && a.opts.Storage != nil {
		go a.opts.Storage.Write(s)
	}
	return s, e
}

func (a *Manager) FindOrCreate(ctx *kaos.Context, parm codekit.M, data codekit.M) (*Session, error) {
	id := parm.GetString("ID")
	duration := parm.GetInt("Second")
	if duration == 0 {
		duration = a.secondLifeTime
	}

	if id == "" {
		return nil, errors.New("ID is mandatory")
	}

	s, ok := a.pool.GetByReferenceID(id)
	if !ok {
		s, e := a.pool.Create(id, data, duration)
		if e == nil && a.opts.Storage != nil {
			go a.opts.Storage.Write(s)
		}
		return s, e
	} else {
		if data != nil && len(data) > 0 {
			for k, v := range data {
				s.Data.Set(k, v)
			}
			go a.opts.Storage.Write(s)
		}
	}
	return s, nil
}

func (a *Manager) Store() error {
	if a.opts.Storage != nil {
		return a.opts.Storage.Store(a.pool)
	}
	return nil
}

func (a *Manager) Load() error {
	if a.opts.Storage != nil {
		return a.opts.Storage.Load(a.pool)
	}
	return nil
}

func (a *Manager) Close() {
	a.Store()
	// do nothing for now
}
