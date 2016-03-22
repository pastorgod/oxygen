package base

import (
	"math"
)

func PowInt32(base, num int32) int32 {
	return int32(math.Pow(float64(base), float64(num)))
}

// Fast Inverse Square Root
func InvSqrt(x float32) float32 {
	var xhalf float32 = 0.5 * x // get bits for floating VALUE
	i := math.Float32bits(x)    // gives initial guess y0
	i = 0x5f375a86 - (i >> 1)   // convert bits BACK to float
	x = math.Float32frombits(i) // Newton step, repeating increases accuracy
	x = x * (1.5 - xhalf*x*x)
	x = x * (1.5 - xhalf*x*x)
	x = x * (1.5 - xhalf*x*x)
	return 1 / x
}

func Sin(x float32) float32 {
	var index int
	if x >= 0 {
		index = int(x*TrigTableFactor) % TrigTableSize
	} else {
		index = TrigTableSize - (int(-x*TrigTableFactor) % TrigTableSize) - 1
	}
	return msSinTable[index]
}

func Cos(x float32) float32 {
	return Sin(x + HalfPI)
}

func Rotate(x, z, angle float32) (lx, lz float32) {
	sin := Sin(angle)
	cos := Cos(angle)
	return sin*z + cos*x, cos*z + sin*x
}

func DistanceSqr(p0, p1 Vector3) float32 {
	x, z := p0.X-p1.X, p0.Z-p1.Z
	return x*x + z*z
}

func ApproxDistanceVec(p0, p1 *Vector3) int {
	return ApproxDistance((p0.X-p1.X)*DistanceScale, (p0.Z-p1.Z)*DistanceScale)
}

// reference: http://www.flipcode.com/archives/Fast_Approximate_Distance_Functions.shtml
func ApproxDistance(dx, dy float32) int {
	var min, max, approx int
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx < dy {
		min = int(dx)
		max = int(dy)
	} else {
		min = int(dy)
		max = int(dx)
	}

	approx = max*1007 + min*441
	if max < (min << 4) {
		approx -= max * 40
	}

	// add 512 for proper rounding
	return (approx + 512) >> 10
}

func DistanceMaxVec(p0, p1 *Vector3) float32 {
	return float32(math.Max(math.Abs(float64(p0.X-p1.X)), math.Abs(float64(p0.Z-p1.Z))))
}

func DistanceMax(x1, y1, x2, y2 float32) float32 {
	return float32(math.Max(math.Abs(float64(x1-x2)), math.Abs(float64(y1-y2))))
}

func CrossProduct(x0, y0, x1, y1 float32) float32 {
	return x0*y1 - y0*x1
}

func Clamp(value, min, max float32) float32 {
	if value < min {
		value = min
	} else {
		if value > max {
			value = max
		}
	}
	return value
}

func ClampInt(value, min, max int) int {
	if value < min {
		value = min
	} else {
		if value > max {
			value = max
		}
	}
	return value
}

func Clamp01(value float32) float32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func Repeat(t, length float32) float32 {
	return t - float32(math.Floor(float64(t/length)))*length
}

func Lerp(from, to, t float32) float32 {
	return from + (to-from)*Clamp01(t)
}

func LerpAngle(a, b, t float32) float32 {
	num := Repeat(b-a, 360)
	if num > 180 {
		num -= 180
	}
	return a + num*Clamp01(t)
}

// 求两个点之间的距离
func Distance(x1, y1, x2, y2 float32) float32 {
	deletax := x1 - x2
	deletay := y1 - y2
	return float32(math.Sqrt(float64(deletax*deletax + deletay*deletay)))
}

// 求点到线段的距离
func PointToLineSegmentDistance(x, y, x1, y1, x2, y2 float32) float32 {
	v1_x := x2 - x1
	v1_y := y2 - y1

	a := (x2-x1)*v1_x + (y2-y1)*v1_y
	b := (x1-x)*v1_x + (y1-y)*v1_y

	if a == 0. {
		return 0
	}

	t := -b / a

	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	point_x := x1 + t*(x2-x1)
	point_y := y1 + t*(y2-y1)

	return Distance(point_x, point_y, x, y)
}

// 四边形求交
func RectangleHitDefineCollision(
	HitDefPos *Vector3, HitDefOrientation float32,
	HitDef *Vector3,
	AttackeePos *Vector3, AttackeeOrientation float32,
	AttackeeBounding *Vector3) bool {

	//排除高度影响，以XZ平面坐标作为判定基准
	if HitDefPos.Y > AttackeePos.Y+AttackeeBounding.Y || AttackeePos.Y > HitDefPos.Y+HitDef.Y {
		return false
	}

	// 计算出第一个四边形的四个定点
	x0, z0 := -HitDef.X*0.5, -HitDef.Z*0.5
	x1, z1 := -HitDef.X*0.5, HitDef.Z*0.5

	x0, z0 = Rotate(x0, z0, HitDefOrientation)
	x1, z1 = Rotate(x1, z1, HitDefOrientation)

	maxHit := &Vector2{
		X: float32(math.Max(math.Abs(float64(x0)), math.Abs(float64(x1)))),
		Y: float32(math.Max(math.Abs(float64(z0)), math.Abs(float64(z1)))),
	}

	HitDefPointX := []float32{
		HitDefPos.X - x0,
		HitDefPos.X - x1,
		HitDefPos.X + x0,
		HitDefPos.X + x1}

	HitDefPointZ := []float32{
		HitDefPos.Z - z0,
		HitDefPos.Z - z1,
		HitDefPos.Z + z0,
		HitDefPos.Z + z1}

	// 计算出第二个四边形的四个顶点
	x0 = -AttackeeBounding.X * 0.5
	z0 = -AttackeeBounding.Z * 0.5
	x1 = -AttackeeBounding.X * 0.5
	z1 = AttackeeBounding.Z * 0.5

	x0, z0 = Rotate(x0, z0, AttackeeOrientation)
	x1, z1 = Rotate(x1, z1, AttackeeOrientation)

	maxAtk := &Vector2{
		X: float32(math.Max(math.Abs(float64(x0)), math.Abs(float64(x1)))),
		Y: float32(math.Max(math.Abs(float64(z0)), math.Abs(float64(z1)))),
	}

	AttackeePointX := []float32{
		AttackeePos.X - x0,
		AttackeePos.X - x1,
		AttackeePos.X + x0,
		AttackeePos.X + x1}

	AttackeePointZ := []float32{
		AttackeePos.Z - z0,
		AttackeePos.Z - z1,
		AttackeePos.Z + z0,
		AttackeePos.Z + z1}

	if HitDefPos.X > AttackeePos.X+maxHit.X+maxAtk.X ||
		HitDefPos.X < AttackeePos.X-maxHit.X-maxAtk.X ||
		HitDefPos.Z > AttackeePos.Z+maxHit.Y+maxAtk.Y ||
		HitDefPos.Z < AttackeePos.Z-maxHit.Y-maxAtk.Y {
		return false
	}

	// 拿四边形的四个顶点判断，是否在另外一个四边形的四条边的一侧
	for i := 0; i < 4; i++ {
		x0 = HitDefPointX[i]
		x1 = HitDefPointX[(i+1)%4]
		z0 = HitDefPointZ[i]
		z1 = HitDefPointZ[(i+1)%4]

		hasSameSidePoint := false

		for j := 0; j < 4; j++ {
			if v := CrossProduct(x1-x0, z1-z0, AttackeePointX[j]-x0, AttackeePointZ[j]-z0); v < 0 {
				hasSameSidePoint = true
				break
			}
		}

		// 如果4个定点都在其中一条边的另外一侧，说明没有交点
		if !hasSameSidePoint {
			return false
		}
	}

	// 所有边可以分割另外一个四边形，说明有焦点。
	return true
}

// 圆柱体求交
func CylinderHitDefineCollision(
	HitDefPos *Vector3, HitDefOrientation,
	HitRadius, HitDefHeight float32,
	AttackeePos *Vector3, AttackeeOrientation float32,
	AttackeeBounding *Vector3) bool {

	//排除高度影响，以XZ平面坐标作为判定基准
	if (HitDefPos.Y > (AttackeePos.Y + AttackeeBounding.Y)) || (AttackeePos.Y > (HitDefPos.Y + HitDefHeight)) {
		return false
	}

	vectx := HitDefPos.X - AttackeePos.X
	vectz := HitDefPos.Z - AttackeePos.Z

	if vectx != 0 || vectz != 0 {
		vectx, vectz = Rotate(vectx, vectz, -AttackeeOrientation)
	}

	if (math.Abs(float64(vectx)) > float64(HitRadius+AttackeeBounding.Z)) ||
		(math.Abs(float64(vectz)) > float64(HitRadius+AttackeeBounding.X)) {
		return false
	}

	return true
}

// 圆环求交
func RingHitDefineCollision(
	HitDefPos *Vector3, HitDefOrientation,
	HitInnerRadius, HitDefHeight, HitOutRadius float32,
	AttackeePos *Vector3, AttackeeOrientation float32,
	AttackeeBounding *Vector3) bool {

	//排除高度影响，以XZ平面坐标作为判定基准
	if HitDefPos.Y > AttackeePos.Y+AttackeeBounding.Y || AttackeePos.Y > HitDefPos.Y+HitDefHeight {
		return false
	}

	radius := float32(math.Min(float64(AttackeeBounding.X), float64(AttackeeBounding.Z)))
	distance := Vector3Zero().Sub(AttackeePos, HitDefPos).Mag()

	if distance+radius < HitInnerRadius || distance-radius > HitOutRadius {
		return false
	}

	return true
}

// 扇形求交
func FanDefineCollision(
	HitDefPos *Vector3, HitDefOrientation,
	HitRadius, HitDefHeight, StartAngle, EndAngle float32,
	AttackeePos *Vector3, AttackeeOrientation float32,
	AttackeeBounding *Vector3) bool {

	//排除高度影响，以XZ平面坐标作为判定基准
	if HitDefPos.Y > AttackeePos.Y+AttackeeBounding.Y || AttackeePos.Y > HitDefPos.Y+HitDefHeight {
		return false
	}

	//圆心的坐标转化到被攻击者的坐标系去
	vectz := AttackeePos.Z - HitDefPos.Z
	vectx := AttackeePos.X - HitDefPos.X
	vectz, vectx = Rotate(vectz, vectx, HitDefOrientation)

	hitRadius := HitRadius

	attackCenter_x := vectz
	attackCenter_y := vectx

	attackRadius := AttackeeBounding.X

	if AttackeeBounding.X > AttackeeBounding.Z {
		attackRadius = AttackeeBounding.Z
	}

	centerDis := Distance(0, 0, attackCenter_x, attackCenter_y)

	if centerDis > (hitRadius + attackRadius) { //相离
		return false
	}

	if centerDis <= attackRadius {
		return true
	}

	start_rad := StartAngle * Deg2Rad
	end_rad := EndAngle * Deg2Rad

	center_angle := 0.5 * (start_rad + end_rad)
	axis_x := Cos(center_angle)
	axis_y := Sin(center_angle)

	length := float32(math.Sqrt(float64(attackCenter_x*attackCenter_x + attackCenter_y*attackCenter_y)))
	temp_x := attackCenter_x / length
	temp_y := attackCenter_y / length

	dot := axis_x*temp_x + temp_y*axis_y

	value := Cos(0.5 * (end_rad - start_rad))

	if dot >= value {
		return true
	}

	dis1 := PointToLineSegmentDistance(attackCenter_x, attackCenter_y, 0, 0,
		hitRadius*Cos(start_rad), hitRadius*Sin(start_rad))

	dis2 := PointToLineSegmentDistance(attackCenter_x, attackCenter_y, 0, 0,
		hitRadius*Cos(end_rad), hitRadius*Sin(end_rad))

	if (dis1 <= attackRadius) || (dis2 <= attackRadius) {
		return true
	}

	return false
}

///////////////////////////////////////////////////////////////////////////////////////
const TrigTableSize = 4096
const TrigTableFactor = TrigTableSize / (math.Pi * 2)
const HalfPI = math.Pi * 0.5
const DistanceScale = 100
const Deg2Rad = 0.0174533
const Rad2Deg = 57.2958

var msSinTable []float32

func InitMath() {
	msSinTable = make([]float32, TrigTableSize)
	for i := 0; i < TrigTableSize; i++ {
		angle := math.Pi * 2.0 * float64(i) / float64(TrigTableSize)
		msSinTable[i] = float32(math.Sin(angle))
	}
}

func init() {
	//	InitMath()
}
