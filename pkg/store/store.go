package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kubeplay/gameserver/pkg/types"
	bolt "go.etcd.io/bbolt"
)

type Store struct {
	dbfile     string
	pathPrefix string

	path         string
	resourceName string
	kind         string
	objType      types.Object
	into         func(obj interface{}) error
}

func New(dbfile, pathPrefix string) *Store {
	store := &Store{
		dbfile:     dbfile,
		pathPrefix: pathPrefix,
	}
	// TODO: check if dbfile is not a folder
	return store
}

func (s *Store) Kind(kind string) *Store {
	// s.kind = strings.ToLower(kind)
	for _, obj := range types.RegisteredTypes {
		if obj.GetObjectKind() == kind {
			s.objType = obj
		}
	}
	if s.objType == nil {
		// TODO: it must contains a type!
	}
	return s
}

func (s *Store) Resources(keys ...string) *Store {
	s.path = fmt.Sprintf(strings.Join(keys, "/"))
	return s
}

func (s *Store) SaveObject(obj types.Object) (types.Object, error) {
	db, err := s.DB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	// TODO: deep-copy instead of mutating the object
	meta := obj.GetObjectMeta()
	meta.UID = NewUUID()
	meta.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	err = db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		b, err := tx.CreateBucketIfNotExists([]byte(s.pathPrefix))
		if err != nil {
			return err
		}
		objectKey := []byte(s.GetResourcePath())
		if o := b.Get(objectKey); o != nil {
			return fmt.Errorf("object %q already exists", string(objectKey))
		}
		return b.Put(objectKey, data)
	})
	return obj, err
}

func (s *Store) Update(old, new types.Object) (types.Object, error) {
	db, err := s.DB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	newMeta := new.GetObjectMeta()
	oldMeta := old.GetObjectMeta()

	newMeta.Name = oldMeta.Name
	newMeta.CreatedAt = oldMeta.CreatedAt
	newMeta.UID = oldMeta.UID
	if !reflect.DeepEqual(newMeta.Annotations, oldMeta.Annotations) {
		newMeta.Annotations = oldMeta.Annotations
	}
	return new, db.Update(func(tx *bolt.Tx) error {
		data, err := json.Marshal(new)
		if err != nil {
			return err
		}
		b := tx.Bucket([]byte(s.pathPrefix))
		if b == nil {
			return fmt.Errorf("bucket %q doesn't exists", s.pathPrefix)
		}
		objectKey := []byte(s.GetResourcePath())
		return b.Put(objectKey, data)
	})
}

func (s *Store) Get(name string) (types.Object, error) {
	db, err := s.DB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	obj := s.newObject()
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.pathPrefix))
		if b == nil {
			return fmt.Errorf("bucket %q doesn't exists", s.pathPrefix)
		}
		s.path = path.Join(s.path, name)
		objKey := []byte(s.GetResourcePath())
		data := b.Get(objKey)

		if data == nil {
			return fmt.Errorf("obj %q not found", string(objKey))
		}
		return json.Unmarshal(data, obj)
	})
	return obj, err
}

func (s *Store) List(re *regexp.Regexp) ([]types.Object, error) {
	db, err := s.DB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var items []types.Object
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.pathPrefix))
		if b == nil {
			return fmt.Errorf("bucket %q doesn't exists", s.pathPrefix)
		}
		c := b.Cursor()
		prefix := []byte(s.GetResourcePath())
		for k, v := c.Seek(prefix); k != nil && re.Match(k); k, v = c.Next() {
			obj := s.newObject()
			if err := json.Unmarshal(v, obj); err != nil {
				return err
			}
			items = append(items, obj)
		}
		return nil
	})
	return items, err
}

func (s *Store) Delete(name string) error {
	db, err := s.DB()
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(s.pathPrefix))
		if b == nil {
			return fmt.Errorf("bucket %q doesn't exists", s.pathPrefix)
		}
		s.path = path.Join(s.path, name)
		c := b.Cursor()
		prefix := []byte(s.GetResourcePath())
		// Lookup and delete all child keys
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return b.Delete(prefix)
	})
}

func (s *Store) GetResourcePath() string {
	return path.Join("/", s.path)
}

func (s *Store) newObject() types.Object {
	if s.objType == nil {
		// TODO: raise error
	}
	return s.objType.New()
}

func (s *Store) DB() (*bolt.DB, error) {
	db, err := bolt.Open(s.dbfile, 0600, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}
