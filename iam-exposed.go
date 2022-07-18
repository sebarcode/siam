package siam

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

func (a *Manager) Get(ctx *kaos.Context, sessionID string) (*Session, error) {
	var err error
	if sessionID == "" {
		return nil, errors.New("sessionID is mandatory")
	}

	session, ok := a.pool.GetBySessionID(sessionID)
	if !ok {
		if a.opts.Storage == nil {
			return nil, errors.New("Sesssion not found")
		}

		session, err = a.opts.Storage.Get(sessionID)
		if err != nil {
			return nil, errors.New("Session not found. " + err.Error())
		}

		// session found from storage, update session pool
		func() {
			a.pool.mtx.Lock()
			defer a.pool.mtx.Unlock()
			a.pool.sessions[session.SessionID] = session
			a.pool.refs[session.ReferenceID] = session.SessionID
		}()
	}

	a.pool.Update(session.SessionID, 0)
	return session, nil
}

func (a *Manager) Renew(ctx *kaos.Context, parm codekit.M) (*Session, error) {
	id := parm.GetString("ID")
	duration := parm.GetInt("Second")
	if id == "" {
		return nil, errors.New("ID is mandatory")
	}

	seid, e := a.pool.Renew(id, duration)
	if e != nil {
		return nil, e
	}
	se, _ := a.pool.GetBySessionID(seid)
	if a.opts.Storage != nil {
		go a.opts.Storage.Write(se)
	}
	return se, nil
}

func (a *Manager) Remove(ctx *kaos.Context, parm codekit.M) (string, error) {
	id := parm.GetString("ID")

	se, _ := a.pool.GetBySessionID(id)
	if se == nil {
		return "", nil
	}

	delete(a.pool.refs, se.ReferenceID)
	delete(a.pool.sessions, se.SessionID)

	if a.opts.Storage != nil {
		go a.opts.Storage.Remove(se.SessionID)
	}
	return "", nil
}
