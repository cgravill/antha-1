package inventory

import (
	"context"
	"errors"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
)

var (
	// ErrUnknownType is returned if type is not in inventory
	ErrUnknownType = errors.New("unknown type")

	errCannotListPlates = errors.New("cannot list plates")
)

const (
	// WaterType is the component type of water
	WaterType = "water"
)

const (
	theCtxKey ctxKey = "inventory"
)

type ctxKey string

func fromContext(ctx context.Context) Inventory {
	return ctx.Value(theCtxKey).(Inventory)
}

// An Inventory returns items by name
type Inventory interface {
	NewComponent(ctx context.Context, typ string) (*wtype.LHComponent, error)
	NewPlate(ctx context.Context, typ string) (*wtype.LHPlate, error)
	NewTipwaste(ctx context.Context, typ string) (*wtype.LHTipwaste, error)
	NewTipbox(ctx context.Context, typ string) (*wtype.LHTipbox, error)
}

type hasXXXGetPlates interface {
	XXXGetPlates(ctx context.Context) ([]*wtype.LHPlate, error)
}

// NewContext returns a context with the given inventory
func NewContext(ctx context.Context, inv Inventory) context.Context {
	return context.WithValue(ctx, theCtxKey, inv)
}

func GetInventory(ctx context.Context) Inventory {
	return fromContext(ctx)
}

// NewComponent returns a new component of the given type
func NewComponent(ctx context.Context, typ string) (*wtype.LHComponent, error) {
	return fromContext(ctx).NewComponent(ctx, typ)
}

// NewPlate returns a new plate of the given type
func NewPlate(ctx context.Context, typ string) (*wtype.LHPlate, error) {
	return fromContext(ctx).NewPlate(ctx, typ)
}

// NewTipwaste returns a new tipwaste of the given type
func NewTipwaste(ctx context.Context, typ string) (*wtype.LHTipwaste, error) {
	return fromContext(ctx).NewTipwaste(ctx, typ)
}

// NewTipbox returns a new tipbox of the given type
func NewTipbox(ctx context.Context, typ string) (*wtype.LHTipbox, error) {
	return fromContext(ctx).NewTipbox(ctx, typ)
}

// XXXNewPlates is a transitional call that will be removed once the planner
// requires input plate types to be explicitly set.
func XXXNewPlates(ctx context.Context) ([]*wtype.LHPlate, error) {
	inv := GetInventory(ctx)
	xInv, ok := inv.(hasXXXGetPlates)
	if !ok {
		return nil, errCannotListPlates
	}

	return xInv.XXXGetPlates(ctx)
}