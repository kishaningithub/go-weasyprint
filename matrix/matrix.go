package matrix

import (
	"errors"
	"math"
)

type fl = float32

// Transform encode a (2D) linear transformation (Y = AX + B)
// The encoded transformation is given by :
// 		x_new = a * x + c * y + e
//		y_new = b * x + d * y + f
// which is equivalent to the vector notation
// 	A = | A C | B = | E |
//		| B	D |		| F |
type Transform struct {
	a, b, c, d, e, f fl
}

func New(a, b, c, d, e, f fl) Transform {
	return Transform{a: a, b: b, c: c, d: d, e: e, f: f}
}

func Identity() Transform {
	return New(1, 0, 0, 1, 0, 0)
}

func Translation(tx, ty fl) Transform {
	mt := Identity()
	mt.Translate(tx, ty)
	return mt
}

func Scaling(sx, sy fl) Transform {
	mt := Identity()
	mt.Scale(sx, sy)
	return mt
}

// Copy() method returns a copy of M
func (T Transform) Copy() *Transform {
	return &T
}

func (T Transform) Data() (a, b, c, d, e, f fl) {
	return T.a, T.b, T.c, T.d, T.e, T.f
}

// write T_ * T in out
func mult(T_, T Transform, out *Transform) {
	out.a = T_.a*T.a + T_.c*T.b
	out.b = T_.b*T.a + T_.d*T.b
	out.c = T_.a*T.c + T_.c*T.d
	out.d = T_.b*T.c + T_.d*T.d
	out.e = T_.a*T.e + T_.c*T.f + T_.e
	out.f = T_.b*T.e + T_.d*T.f + T_.f
}

// Mul returns the transform T * U,
// which apply U then T.
func Mul(T, U Transform) Transform {
	out := Transform{}
	mult(T, U, &out)
	return out
}

// Invert modify the matrix in place. Return an error
// if the transformation is not bijective.
func (T *Transform) Invert() error {
	det := T.a*T.d - T.b*T.c
	if det == 0 {
		return errors.New("The transformation is not invertible.")
	}
	T.a, T.d = T.d/det, T.a/det
	T.b = -T.b / det
	T.c = -T.c / det
	T.e = -(T.a*T.e + T.c*T.f)
	T.f = -(T.b*T.e + T.d*T.f)
	return nil
}

// Transforms the point `(x, y)` by this matrix
func (T Transform) TransformPoint(x, y fl) (outX, outY fl) {
	tmpX, tmpY := T.TransformDistance(x, y)
	return tmpX + T.e, tmpY + T.f
}

// Transforms the distance vector ``(dx, dy)`` by this matrix.
// This is similar to `TransformPoint` except that the translation components
// of the transformation are ignored.
// The calculation of the returned vector is as follows::
// 	dx2 = dx1 * xx + dy1 * xy
// 	dy2 = dx1 * yx + dy1 * yy
func (T Transform) TransformDistance(x, y fl) (outX, outY fl) {
	return T.a*x + T.c*y, T.b*x + T.d*y
}

// Applies a translation by `tx`, `ty`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first translate the coordinates by `tx` and `ty`,
// then apply the original transformation to the coordinates.
//
// 	This changes the matrix in-place.
func (T *Transform) Translate(tx, ty fl) {
	T.e, T.f = T.TransformPoint(tx, ty)
}

// Applies scaling by `sx`, `sy`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first scale the coordinates by `sx` and `sy`,
// then apply the original transformation to the coordinates.
//
// This changes the matrix in-place.
func (T *Transform) Scale(sx, sy fl) {
	mult(*T, Transform{sx, 0, 0, sy, 0, 0}, T)
}

// Applies a rotation by `radians`
// to the transformation in this matrix.
//
// The effect of the new transformation is to
// first rotate the coordinates by `radians`,
// then apply the original transformation to the coordinates.
//
// 	This changes the matrix in-place.
//
// `radians` is the angle of rotation, in radians.
// 	The direction of rotation is defined such that positive angles
// 	rotate in the direction from the positive X axis
// 	toward the positive Y axis.
func (T *Transform) Rotate(radians fl) {
	cos, sin := fl(math.Cos(float64(radians))), fl(math.Sin(float64(radians)))
	mult(*T, Transform{cos, sin, -sin, cos, 0, 0}, T)
}
