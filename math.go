package main

import "math"

// Vec3 is a plain vector, used for positions, velocities and directions.
type Vec3 = [3]float32

func floor(x float32) float32 { return float32(math.Floor(float64(x))) }
func ceil(x float32) float32  { return float32(math.Ceil(float64(x))) }
func sin(x float32) float32   { return float32(math.Sin(float64(x))) }
func cos(x float32) float32   { return float32(math.Cos(float64(x))) }
func sqrt(x float32) float32  { return float32(math.Sqrt(float64(x))) }
func abs(x float32) float32   { return float32(math.Abs(float64(x))) }

const pi = float32(math.Pi)

func lerp(a, b, t float32) float32 { return a + (b-a)*t }

func clamp(x, lo, hi float32) float32 { return max(lo, min(hi, x)) }

func rad(deg float32) float32 { return deg * pi / 180 }

// wrap maps v into 0..n, also for negative values.
func wrap(v, n int) int { return ((v % n) + n) % n }
