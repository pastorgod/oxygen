package base

import (
	"encoding/json"
	"math"
)

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

func (v Vector3) String() string {
	return Sprintf("X: %v, Y: %v, Z: %v", v.X, v.Y, v.Z)
}

func NewVector3(x, y, z float32) *Vector3 {
	return &Vector3{x, y, z}
}

func Vector3Zero() *Vector3 {
	return &Vector3{0, 0, 0}
}

func Vector3One() *Vector3 {
	return &Vector3{1, 1, 1}
}

func Vector3Up() *Vector3 {
	return &Vector3{0, 1, 0}
}

func Vector3Forward() *Vector3 {
	return &Vector3{0, 0, 1}
}

func Vector3Left() *Vector3 {
	return &Vector3{1, 0, 0}
}

func (v *Vector3) Copy() *Vector3 {
	return &Vector3{v.X, v.Y, v.Z}
}

func (v *Vector3) Add(a, b *Vector3) *Vector3 {
	v.X = a.X + b.X
	v.Y = a.Y + b.Y
	v.Z = a.Z + b.Z
	return v
}

func (v *Vector3) Sub(a, b *Vector3) *Vector3 {
	v.X = a.X - b.X
	v.Y = a.Y - b.Y
	v.Z = a.Z - b.Z
	return v
}

// Div sets v to be u / x for some scalar x and returns v.
func (v *Vector3) Div(u *Vector3, x float32) *Vector3 {
	v.X = u.X / x
	v.Y = u.Y / x
	v.Z = u.Z / x
	return v
}

// Mul sets v to be u * x for some scalar x and returns v.
func (v *Vector3) Mul(u *Vector3, x float32) *Vector3 {
	v.X = u.X * x
	v.Y = u.Y * x
	v.Z = u.Z * x
	return v
}

// Inv sets v to be the inverse of u and returns v.
func (v *Vector3) Inv(u *Vector3) *Vector3 {
	return v.Mul(v, -1)
}

// Mag returns the magnitude of v.
func (v *Vector3) Mag() float32 {
	return float32(math.Sqrt(float64((v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z))))
}

// Normalize sets v to be the normalized (unit) vector of u and returns v.
func (v *Vector3) Normalize(u *Vector3) *Vector3 {
	return v.Div(u, u.Mag())
}

// Cross returns the cross product of u and v as a newly allocated vector.
// (This function does not follow the math/big pattern because it wouldn't work if the result vector were also
// one of the operands.)
func (v *Vector3) Cross(u *Vector3) *Vector3 {
	return &Vector3{
		X: (v.Y * u.Z) - (v.Z * u.Y),
		Y: (v.Z * u.X) - (v.X * u.Z),
		Z: (v.X * u.Y) - (v.Y * u.X),
	}
}

// Dot returns the dot product of u and v.
func (v *Vector3) Dot(u *Vector3) float32 {
	return (v.X * u.X) + (v.Y * u.Y) + (v.Z * u.Z)
}

func (v *Vector3) UnmarshalJSON(b []byte) error {
	a := [3]float32{}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	v.X, v.Y, v.Z = a[0], a[1], a[2]
	return nil
}
