package base

import (
	"encoding/json"
	"math"
)

type Vector2 struct {
	X float32
	Y float32
}

func (v Vector2) String() string {
	return Sprintf("X: %v, Y: %v", v.X, v.Y)
}

func NewVector2(x, y float32) *Vector2 {
	return &Vector2{x, y}
}

func Vector2Zero() *Vector2 {
	return &Vector2{0, 0}
}

func Vector2One() *Vector2 {
	return &Vector2{1, 1}
}

func Vector2Up() *Vector2 {
	return &Vector2{0, 1}
}

func Vector2Forward() *Vector2 {
	return &Vector2{1, 0}
}

func (v *Vector2) Copy() *Vector2 {
	return &Vector2{v.X, v.Y}
}

func (v *Vector2) Add(a, b *Vector2) *Vector2 {
	v.X = a.X + b.X
	v.Y = a.Y + b.Y
	return v
}

func (v *Vector2) Sub(a, b *Vector2) *Vector2 {
	v.X = a.X - b.X
	v.Y = a.Y - b.Y
	return v
}

// Div sets v to be u / x for some scalar x and returns v.
func (v *Vector2) Div(u *Vector2, x float32) *Vector2 {
	v.X = u.X / x
	v.Y = u.Y / x
	return v
}

// Mul sets v to be u * x for some scalar x and returns v.
func (v *Vector2) Mul(u *Vector2, x float32) *Vector2 {
	v.X = u.X * x
	v.Y = u.Y * x
	return v
}

// Inv sets v to be the inverse of u and returns v.
func (v *Vector2) Inv(u *Vector2) *Vector2 {
	return v.Mul(v, -1)
}

// Mag returns the magnitude of v.
func (v *Vector2) Mag() float32 {
	return float32(math.Sqrt(float64((v.X * v.X) + (v.Y * v.Y))))
}

// Normalize sets v to be the normalized (unit) vector of u and returns v.
func (v *Vector2) Normalize(u *Vector2) *Vector2 {
	return v.Div(u, u.Mag())
}

// Cross returns the cross product of u and v as a newly allocated vector.
// (This function does not follow the math/big pattern because it wouldn't work if the result vector were also
// one of the operands.)
func (v *Vector2) Cross(u *Vector2) *Vector2 {
	return &Vector2{
		X: (v.Y * u.X) - (v.X * u.Y),
		Y: (v.Y * u.X) - (v.X * u.Y),
	}
}

// Dot returns the dot product of u and v.
func (v *Vector2) Dot(u *Vector2) float32 {
	return (v.X * u.X) + (v.Y * u.Y)
}

func (v *Vector2) UnmarshalJSON(b []byte) error {
	a := [2]float32{}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	v.X, v.Y = a[0], a[1]
	return nil
}
