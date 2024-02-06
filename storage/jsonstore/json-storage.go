package jsonstore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sebarcode/siam"
)

type store struct {
	folderPath string
	lock       *sync.RWMutex
}

func NewStorage(loc string) *store {
	s := new(store)
	s.folderPath = loc
	s.lock = new(sync.RWMutex)
	return s
}

func (s *store) Load(pool *siam.SessionPool) error {
	fis, e := ioutil.ReadDir(s.folderPath)
	if e != nil {
		return e
	}

	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}

		bs, e := ioutil.ReadFile(filepath.Join(s.folderPath, fi.Name()))
		if e != nil {
			return fmt.Errorf("fail read file %s. %s", fi.Name(), e.Error())
		}

		sess := new(siam.Session)
		if e = json.Unmarshal(bs, sess); e != nil {
			return fmt.Errorf("fail serializing file %s. %s", fi.Name(), e.Error())
		}

		if e = pool.RegisterSession(sess); e != nil {
			return fmt.Errorf("fail register session. %s", e.Error())
		}
	}

	return nil
}

func (s *store) Store(pool *siam.SessionPool) error {
	ids := pool.GetIDs()
	for _, id := range ids {
		if sess, ok := pool.GetBySessionID(id); ok {
			e := s.Write(sess)
			if e != nil {
				return fmt.Errorf("fail to store session %s. %s", id, e.Error())
			}
		}
	}
	return nil
}

func (s *store) Get(id string) (*siam.Session, error) {
	locPath := filepath.Join(s.folderPath, id+".json")
	bs, e := os.ReadFile(locPath)
	if e != nil {
		return nil, fmt.Errorf("fail read file %s. %s", id, e.Error())
	}

	sess := new(siam.Session)
	if e = json.Unmarshal(bs, sess); e != nil {
		return nil, fmt.Errorf("fail serializing file %s. %s", id, e.Error())
	}

	return sess, nil
}

func (s *store) Write(sess *siam.Session) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	bs, e := json.Marshal(sess)
	if e != nil {
		return fmt.Errorf("fail serializing session %s. %s", sess.SessionID, e.Error())
	}

	fileName := filepath.Join(s.folderPath, sess.SessionID+".json")
	e = os.WriteFile(fileName, bs, 0644)
	if e != nil {
		return fmt.Errorf("fail write file %s. %s", sess.SessionID, e.Error())
	}
	return nil
}

func (s *store) Remove(id string) {
	s.lock.Lock()
	s.lock.Unlock()

	locPath := filepath.Join(s.folderPath, id+".json")
	os.Remove(locPath)
}

func (s *store) Close() {
}
