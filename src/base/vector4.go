package base

import (
	"encoding/json"
	"math"
)

type Vector4 struct {
	W float32
	X float32
	Y float32
	Z float32
}

func (v Vector4) String() string {
	return Sprintf("W: %v, X: %v, Y: %v, Z: %v", v.W, v.X, v.Y, v.Z)
}

func NewVector4(w, x, y, z float32) *Vector4 {
	return &Vector4{w, x, y, z}
}

func Vector4Zero() *Vector4 {
	return &Vector4{0, 0, 0, 0}
}

func Vector4One() *Vector4 {
	return &Vector4{1, 1, 1, 1}
}

func Vector4Up() *Vector4 {
	return &Vector4{0, 0, 1, 0}
}

func Vector4Forward() *Vector4 {
	return &Vector4{0, 0, 0, 1}
}

func Vector4Left() *Vector4 {
	return &Vector4{0, 1, 0, 0}
}

func (v *Vector4) Copy() *Vector4 {
	return &Vector4{v.W, v.X, v.Y, v.Z}
}

func (v *Vector4) Add(a, b *Vector4) *Vector4 {
	v.W = a.W + b.W
	v.X = a.X + b.X
	v.Y = a.Y + b.Y
	v.Z = a.Z + b.Z
	return v
}

func (v *Vector4) Sub(a, b *Vector4) *Vector4 {
	v.W = a.W - b.W
	v.X = a.X - b.X
	v.Y = a.Y - b.Y
	v.Z = a.Z - b.Z
	return v
}

// Div sets v to be u / x for some scalar x and returns v.
func (v *Vector4) Div(u *Vector4, x float32) *Vector4 {
	v.W = u.W / x
	v.X = u.X / x
	v.Y = u.Y / x
	v.Z = u.Z / x
	return v
}

// Mul sets v to be u * x for some scalar x and returns v.
func (v *Vector4) Mul(u *Vector4, x float32) *Vector4 {
	v.W = u.W * x
	v.X = u.X * x
	v.Y = u.Y * x
	v.Z = u.Z * x
	return v
}

// Inv sets v to be the inverse of u and returns v.
func (v *Vector4) Inv(u *Vector4) *Vector4 {
	return v.Mul(v, -1)
}

// Mag returns the magnitude of v.
func (v *Vector4) Mag() float32 {
	return float32(math.Sqrt(float64((v.W * v.W) + (v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z))))
}

// Normalize sets v to be the normalized (unit) vector of u and returns v.
func (v *Vector4) Normalize(u *Vector4) *Vector4 {
	return v.Div(u, u.Mag())
}

// Cross returns the cross product of u and v as a newly allocated vector.
// (This function does not follow the math/big pattern because it wouldn't work if the result vector were also
// one of the operands.)
//func (v *Vector4) Cross(u *Vector4) *Vector4 {
//	return &Vector4{
//		W: ???
//		X: ???
//		Y: ???
//		Z: ???
//	}
//}

// Dot returns the dot product of u and v.
func (v *Vector4) Dot(u *Vector4) float32 {
	return (v.W * u.W) + (v.X * u.X) + (v.Y * u.Y) + (v.Z * u.Z)
}

func (v *Vector4) UnmarshalJSON(b []byte) error {
	a := [4]float32{}
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	v.W, v.X, v.Y, v.Z = a[0], a[1], a[2], a[3]
	return nil
}
