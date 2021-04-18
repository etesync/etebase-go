package etebase

import (
	"time"

	"github.com/etesync/etebase-go/internal/codec"
	"github.com/google/uuid"
)

const (
	CurrentVersion uint8 = 1
)

type CollectionAccessLevel uint32

const (
	ReadOnly CollectionAccessLevel = iota
	Admin
	ReadWrite
)

type EncryptedCollection struct {
	Item        EncryptedItem         `msgpack:"item"`
	AccessLevel CollectionAccessLevel `msgpack:"accessLevel"`
	Type        []byte                `msgpack:"collectionType"`
	Key         []byte                `msgpack:"collectionKey"`
	Stoken      string                `msgpack:"stoken"`
}

func NewEncryptedCollection(cType string, meta interface{}, content []byte) (*EncryptedCollection, error) {
	metaBytes, err := codec.Marshal(meta)
	if err != nil {
		return nil, err
	}

	uid1 := uuid.New().String()
	uid2 := uuid.New().String()
	return &EncryptedCollection{
		Item: EncryptedItem{
			UID:     uid1,
			Version: CurrentVersion,
			Content: EncryptedRevision{
				UID:    uid2,
				Meta:   metaBytes,
				Chunks: []ChunkArrayItem{},
			},
		},
		Type: []byte(time.Now().String()),
		Key:  []byte("key"),
	}, nil
}

type EncryptedItem struct {
	UID     string            `msgpack:"uid"`
	Version uint8             `msgpack:"version"`
	Etag    []byte            `msgpack:"etag"`
	Content EncryptedRevision `msgpack:"content"`
}

type EncryptedRevision struct {
	UID     string           `msgpack:"uid"`
	Meta    []byte           `msgpack:"meta"`
	Deleted bool             `msgpack:"deleted"`
	Chunks  []ChunkArrayItem `msgpack:"chunks"`
}

type ChunkArrayItem struct {
	Data   string
	Option []byte
}

func (c ChunkArrayItem) MarshalMsgpack() ([]byte, error) {
	return codec.Marshal([]interface{}{
		c.Data, c.Option,
	})
}
