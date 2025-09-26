package coord

import (
	"fmt"
	"math"
	"strings"

	"github.com/xyzj/toolbox"
)

var georep = strings.NewReplacer("(", "", ")", "", ", ", ",", "POINT", "", "POLYGON", "", "LINESTRING", "", "POINT ", "", "POLYGON ", "", "LINESTRING ", "") // 经纬度字符串处理替换器

// Point point struct
type Point struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// String return lng, lat
func (p *Point) String() string {
	return fmt.Sprintf("%.12f %.12f", p.Lng, p.Lat)
}

// GeoText return mysql geotext
func (p *Point) GeoText() string {
	return fmt.Sprintf("POINT (%.12f %.12f)", p.Lng, p.Lat)
}

// Value return the lon and lat value
func (p *Point) Value() (float64, float64) {
	return p.Lng, p.Lat
}

// Equals this point is equivalent to the other point
func (p *Point) Equals(other *Point) bool {
	return p.Lng == other.Lng && p.Lat == other.Lat
}

// Round round this point to the nearest
func (p *Point) Round(l int) *Point {
	var a float64 = 1
	for i := 0; i < l; i++ {
		a *= 10
	}
	return &Point{
		Lng: float64(int(p.Lng*a+0.5)) / a,
		Lat: float64(int(p.Lat*a+0.5)) / a,
	}
}

// RoundString limit number after dot
func (p *Point) RoundString(l int) string {
	if l < 0 {
		l = 12
	}

	s := fmt.Sprintf("%%.%df %%.%df", l, l)
	return fmt.Sprintf(s, p.Lng, p.Lat)
}

// InPolygonRay determines whether the point p lies inside the polygon defined by the slice poly.
// The polygon is represented as a slice of *Point, and must have at least 3 vertices.
// The function uses the ray-casting algorithm to test for point inclusion.
// Returns true if the point is inside the polygon, false otherwise.
func (p *Point) InPolygonRayCasting(poly []*Point) bool {
	n := len(poly)
	if n < 3 {
		return false
	}
	var inside bool
	j := n - 1
	for i := range n {
		pi := poly[i]
		pj := poly[j]
		intersect := ((pi.Lat > p.Lat) != (pj.Lat > p.Lat)) &&
			(p.Lng < (pj.Lng-pi.Lng)*(p.Lat-pi.Lat)/(pj.Lat-pi.Lat)+pi.Lng)
		if intersect {
			inside = !inside
		}
		j = i
	}
	return inside
}

// InPolygon determines whether the point p lies inside the given polygon on a sphere.
// The polygon is defined as a slice of *Point, and each point is assumed to be on the sphere's surface.
// The algorithm works by converting points to 3D Cartesian coordinates, then summing the signed angles
// between consecutive polygon edges as seen from p. If the total angle sum is approximately 2π, p is inside.
// Returns true if p is inside the polygon, false otherwise.
//
// Note: The polygon must have at least 3 points. The function assumes the polygon is non-self-intersecting
// and the points are ordered (either clockwise or counterclockwise).
func (p *Point) InPolygon(polygon []*Point) bool {
	pv := toXYZ(p)
	n := len(polygon)
	if n < 3 {
		return false
	}

	sum := 0.0
	for i := range n {
		a := toXYZ(polygon[i])
		b := toXYZ(polygon[(i+1)%n])

		// 向量点积 & 叉积求夹角
		dot := a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
		cross := [3]float64{
			a[1]*b[2] - a[2]*b[1],
			a[2]*b[0] - a[0]*b[2],
			a[0]*b[1] - a[1]*b[0],
		}
		// 夹角
		angle := math.Atan2(
			math.Sqrt(cross[0]*cross[0]+cross[1]*cross[1]+cross[2]*cross[2]),
			dot,
		)

		// 判断夹角朝向 (点是否在大圆的哪一侧)
		sign := 1.0
		if pv[0]*cross[0]+pv[1]*cross[1]+pv[2]*cross[2] < 0 {
			sign = -1.0
		}
		sum += sign * angle
	}

	// 判断角和是否接近 2π
	return math.Abs(math.Abs(sum)-2*math.Pi) < 1e-6
}

// InCircle determines whether the point p lies within or on the boundary of a circle
// defined by the specified center and radius. It returns true if the distance between
// p and center is less than or equal to radius.
func (p *Point) InCircle(center *Point, radius float64) bool {
	return geoDistance(p, center) <= radius
}

// InLineBuffer checks if the point p lies within a specified buffer distance
// from any segment of the given polyline. The polyline is represented as a
// slice of Point pointers, and the buffer is a float64 value specifying the
// maximum allowed distance. The function returns true if p is within the
// buffer distance of any segment, and false otherwise.
func (p *Point) InLineBuffer(line []*Point, buffer float64) bool {
	for i := 0; i < len(line)-1; i++ {
		if pointToSegmentDistance(p, line[i], line[i+1]) <= buffer {
			return true
		}
	}
	return false
}
func Text2Geo(s string) []*Point {
	geostr := strings.Split(georep.Replace(s), ",")
	gp := make([]*Point, 0, len(geostr))
	for _, v := range geostr {
		vv := strings.Split(v, " ")
		gp = append(gp, &Point{
			Lng: toolbox.String2Float64(vv[0]),
			Lat: toolbox.String2Float64(vv[1]),
		})
	}
	return gp
}

func Geo2Text(gp []*Point) string {
	geostr := "POINT(0 0)" // 默认值，上海
	switch len(gp) {
	case 0: // 没有位置
	case 1: // 点
		geostr = gp[0].GeoText()
	default: // 线或者面
		pts := make([]string, len(gp))
		for k, v := range gp {
			pts[k] = v.String() // fmt.Sprintf("%f %f", v.Lng, v.Lat)
		}
		if pts[0] == pts[len(gp)-1] { // 前后2点一致，表示面
			geostr = fmt.Sprintf("POLYGON((%s))", strings.Join(pts, ","))
		} else {
			geostr = fmt.Sprintf("LINESTRING(%s)", strings.Join(pts, ","))
		}
	}
	return geostr
}

// 度 -> 弧度
func rad(d float64) float64 {
	return d * math.Pi / 180.0
}

// 球面两点间的距离 (米)
func geoDistance(p1, p2 *Point) float64 {
	lat1 := rad(p1.Lat)
	lng1 := rad(p1.Lng)
	lat2 := rad(p2.Lat)
	lng2 := rad(p2.Lng)

	dlat := lat2 - lat1
	dlng := lng2 - lng1

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dlng/2)*math.Sin(dlng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

// 点转单位球向量
func toXYZ(pt *Point) [3]float64 {
	lat := rad(pt.Lat)
	lng := rad(pt.Lng)
	x := math.Cos(lat) * math.Cos(lng)
	y := math.Cos(lat) * math.Sin(lng)
	z := math.Sin(lat)
	return [3]float64{x, y, z}
}

// 点到线段球面距离 (米)
func pointToSegmentDistance(p, a, b *Point) float64 {
	pv := toXYZ(p)
	av := toXYZ(a)
	bv := toXYZ(b)

	// 叉乘
	cross := func(u, v [3]float64) [3]float64 {
		return [3]float64{
			u[1]*v[2] - u[2]*v[1],
			u[2]*v[0] - u[0]*v[2],
			u[0]*v[1] - u[1]*v[0],
		}
	}
	// 点积
	dot := func(u, v [3]float64) float64 {
		return u[0]*v[0] + u[1]*v[1] + u[2]*v[2]
	}
	// 单位化
	norm := func(v [3]float64) [3]float64 {
		l := math.Sqrt(dot(v, v))
		return [3]float64{v[0] / l, v[1] / l, v[2] / l}
	}

	// 大圆平面法向量
	gcNormal := cross(av, bv)
	gcNormal = norm(gcNormal)

	// 投影点
	pn := cross(pv, gcNormal)
	foot := cross(gcNormal, pn)
	foot = norm(foot)

	// 判断垂足是否在线段弧上
	angleAB := math.Acos(dot(av, bv))
	angleAF := math.Acos(dot(av, foot))
	angleBF := math.Acos(dot(bv, foot))

	if angleAF+angleBF-angleAB < 1e-9 {
		// 在弧段上
		d := math.Acos(dot(pv, foot))
		return d * earthRadius
	}

	// 否则取端点距离
	da := geoDistance(p, a)
	db := geoDistance(p, b)
	if da < db {
		return da
	}
	return db
}
