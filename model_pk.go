package xbun

import (
	"github.com/google/uuid"
	"golang.org/x/exp/constraints"
)

var (
	_ HasPK[string]    = (*PK[string])(nil)
	_ HasPK[int]       = (*PKAutoIncrement[int])(nil)
	_ HasPK[uuid.UUID] = (*PKUUID)(nil)
)

type (
	IID interface {
		constraints.Ordered | uuid.UUID
	}

	IIDAutoIncrement interface {
		constraints.Integer | constraints.Float
	}

	HasPK[ID IID] interface {
		GetPK() ID
	}
)

type PK[ID IID] struct {
	ID ID `bun:"id,pk,notnull"`
}

func (p *PK[ID]) GetPK() ID { return p.ID }

type PKAutoIncrement[ID IIDAutoIncrement] struct {
	ID ID `bun:"id,pk,autoincrement"`
}

func (p *PKAutoIncrement[ID]) GetPK() ID { return p.ID }

type PKUUID struct {
	ID uuid.UUID `bun:"id,pk,type:uuid"`
}

func (p *PKUUID) GetPK() uuid.UUID { return p.ID }
