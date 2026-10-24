package math

import "math"

// Vec3 is a plain vector, used for positions, velocities and directions.
type Vec3 = [3]float32

func Floor(x float32) float32 { return float32(math.Floor(float64(x))) }
func Ceil(x float32) float32  { return float32(math.Ceil(float64(x))) }
func Sin(x float32) float32   { return float32(math.Sin(float64(x))) }
func Cos(x float32) float32   { return float32(math.Cos(float64(x))) }
func Sqrt(x float32) float32  { return float32(math.Sqrt(float64(x))) }
func Abs(x float32) float32   { return float32(math.Abs(float64(x))) }

const PI = float32(math.Pi)

func Lerp(a, b, t float32) float32 { return a + (b-a)*t }

func Clamp(x, lo, hi float32) float32 { return max(lo, min(hi, x)) }

func Rad(deg float32) float32 { return deg * PI / 180 }

// wrap maps v into 0..n, also for negative values.
func Wrap(v, n int) int { return ((v % n) + n) % n }
