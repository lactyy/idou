package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"
)

func LoadDefault[K comparable, V any](l Loader[K, V], key K, d V) (value V, err error) {
	value, err = l.Load(key)
	if errors.Is(err, ErrNotFound) {
		value, err = d, nil
	}
	return value, err
}

var ErrNotFound = errors.New("internal: Loader: not found")

type Loader[K comparable, V any] interface {
	Load(key K) (value V, err error)
}

type Storer[K comparable, V any] interface {
	Store(key K, value V) error
}

type LoadStorer[K comparable, V any] interface {
	Loader[K, V]
	Storer[K, V]
}

type LevelDBProvider[K comparable, V any] struct {
	DB *leveldb.DB

	ReadOptions  *opt.ReadOptions
	WriteOptions *opt.WriteOptions
}

func (prov LevelDBProvider[K, V]) Load(key K) (value V, err error) {
	k, err := json.Marshal(key)
	if err != nil {
		return value, fmt.Errorf("encode key: %w", err)
	}
	v, err := prov.DB.Get(k, prov.ReadOptions)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			err = ErrNotFound
		}
		return value, err
	}
	if err := json.Unmarshal(v, &value); err != nil {
		return value, fmt.Errorf("decode: %w", err)
	}
	return value, nil
}

func (prov LevelDBProvider[K, V]) Store(key K, value V) error {
	k, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("encode key: %w", err)
	}
	v, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode value: %w", err)
	}
	return prov.DB.Put(k, v, prov.WriteOptions)
}
