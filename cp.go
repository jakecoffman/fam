package fam

import (
	"math"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type CPRenderer struct {
	shader   *Shader
	vao, vbo uint32

	triangles []Triangle
}

func NewCPRenderer(shader *Shader, projection mgl32.Mat4) *CPRenderer {
	shader.Use().SetMat4("projection", projection)

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	CheckGLErrors()

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	CheckGLErrors()

	SetAttribute(shader.ID, "vertex", 2, gl.FLOAT, 48, 0)
	SetAttribute(shader.ID, "aa_coord", 2, gl.FLOAT, 48, 8)
	SetAttribute(shader.ID, "fill_color", 4, gl.FLOAT, 48, 16)
	SetAttribute(shader.ID, "outline_color", 4, gl.FLOAT, 48, 32)

	gl.BindVertexArray(0)

	CheckGLErrors()
	return &CPRenderer{
		shader: shader,
		vao:    vao,
		vbo:    vbo,
	}
}

func (cpr *CPRenderer) SetProjection(projection mgl32.Mat4) {
	cpr.shader.Use().SetMat4("projection", projection)
}

const (
	DrawPointLineScale = 1
	DrawOutlineWidth   = 1
)

type FColor struct {
	R, G, B, A float32
}

func (cpr *CPRenderer) DrawSpace(space *cp.Space, drawConstraints bool) {
	cpr.ClearRenderer()
	space.EachShape(func(obj *cp.Shape) {
		cpr.DrawShape(obj)
	})
	cpr.FlushRenderer()

	if drawConstraints {
		space.EachConstraint(func(constraint *cp.Constraint) {
			// TODO
			cp.DrawConstraint(constraint, nil)
		})
	}
}

func (cpr *CPRenderer) DrawShape(shape *cp.Shape) {
	body := shape.Body()

	outline := FColor{1, 1, 1, 1}
	fill := FColor{1, 1, 1, 1}

	switch shape.Class.(type) {
	case *cp.Circle:
		circle := shape.Class.(*cp.Circle)
		cpr.DrawCircle(circle.TransformC(), body.Angle(), circle.Radius(), outline, fill)
	case *cp.Segment:
		seg := shape.Class.(*cp.Segment)
		cpr.DrawFatSegment(seg.TransformA(), seg.TransformB(), seg.Radius(), outline, fill)
	case *cp.PolyShape:
		poly := shape.Class.(*cp.PolyShape)

		count := poly.Count()
		verts := make([]cp.Vector, count)

		for i := 0; i < count; i++ {
			verts[i] = poly.TransformVert(i)
		}
		cpr.DrawPolygon(count, verts, poly.Radius(), outline, fill)
	default:
		panic("Unknown shape type")
	}
}

// 8 bytes
type vf struct {
	x, y float32
}

func v2f(v cp.Vector) vf {
	return vf{float32(v.X), float32(v.Y)}
}

// 8*2 + 16*2 bytes = 48 bytes
type Vertex struct {
	vertex, aa_coord          vf
	fill_color, outline_color FColor
}

type Triangle struct {
	a, b, c Vertex
}

func (cpr *CPRenderer) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill FColor) {
	r := radius + 1/DrawPointLineScale
	a := Vertex{
		vf{float32(pos.X - r), float32(pos.Y - r)},
		vf{-1, -1},
		fill,
		outline,
	}
	b := Vertex{
		vf{float32(pos.X - r), float32(pos.Y + r)},
		vf{-1, 1},
		fill,
		outline,
	}
	c := Vertex{
		vf{float32(pos.X + r), float32(pos.Y + r)},
		vf{1, 1},
		fill,
		outline,
	}
	d := Vertex{
		vf{float32(pos.X + r), float32(pos.Y - r)},
		vf{1, -1},
		fill,
		outline,
	}

	t0 := Triangle{a, b, c}
	t1 := Triangle{a, c, d}

	cpr.triangles = append(cpr.triangles, t0)
	cpr.triangles = append(cpr.triangles, t1)

	cpr.DrawFatSegment(pos, pos.Add(cp.ForAngle(angle).Mult(radius-DrawPointLineScale*0.5)), 0, outline, fill)
}

func (cpr *CPRenderer) DrawSegment(a, b cp.Vector, fill FColor) {
	cpr.DrawFatSegment(a, b, 0, fill, fill)
}

func (cpr *CPRenderer) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill FColor) {
	n := b.Sub(a).ReversePerp().Normalize()
	t := n.ReversePerp()

	var half = 1.0 / DrawPointLineScale
	r := radius + half

	if r <= half {
		r = half
		fill = outline
	}

	nw := n.Mult(r)
	tw := t.Mult(r)
	v0 := v2f(b.Sub(nw.Add(tw)))
	v1 := v2f(b.Add(nw.Sub(tw)))
	v2 := v2f(b.Sub(nw))
	v3 := v2f(b.Add(nw))
	v4 := v2f(a.Sub(nw))
	v5 := v2f(a.Add(nw))
	v6 := v2f(a.Sub(nw.Sub(tw)))
	v7 := v2f(a.Add(nw.Add(tw)))

	t0 := Triangle{
		Vertex{v0, vf{1, -1}, fill, outline},
		Vertex{v1, vf{1, 1}, fill, outline},
		Vertex{v2, vf{0, -1}, fill, outline},
	}
	t1 := Triangle{
		Vertex{v3, vf{0, 1}, fill, outline},
		Vertex{v1, vf{1, 1}, fill, outline},
		Vertex{v2, vf{0, -1}, fill, outline},
	}
	t2 := Triangle{
		Vertex{v3, vf{0, 1}, fill, outline},
		Vertex{v4, vf{0, -1}, fill, outline},
		Vertex{v2, vf{0, -1}, fill, outline},
	}
	t3 := Triangle{
		Vertex{v3, vf{0, 1}, fill, outline},
		Vertex{v4, vf{0, -1}, fill, outline},
		Vertex{v5, vf{0, 1}, fill, outline},
	}
	t4 := Triangle{
		Vertex{v6, vf{-1, -1}, fill, outline},
		Vertex{v4, vf{0, -1}, fill, outline},
		Vertex{v5, vf{0, 1}, fill, outline},
	}
	t5 := Triangle{
		Vertex{v6, vf{-1, -1}, fill, outline},
		Vertex{v7, vf{-1, 1}, fill, outline},
		Vertex{v5, vf{0, 1}, fill, outline},
	}

	cpr.triangles = append(cpr.triangles, t0)
	cpr.triangles = append(cpr.triangles, t1)
	cpr.triangles = append(cpr.triangles, t2)
	cpr.triangles = append(cpr.triangles, t3)
	cpr.triangles = append(cpr.triangles, t4)
	cpr.triangles = append(cpr.triangles, t5)
}

func (cpr *CPRenderer) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill FColor) {
	type ExtrudeVerts struct {
		offset, n cp.Vector
	}
	extrude := make([]ExtrudeVerts, count)

	for i := 0; i < count; i++ {
		v0 := verts[(i-1+count)%count]
		v1 := verts[i]
		v2 := verts[(i+1)%count]

		n1 := v1.Sub(v0).ReversePerp().Normalize()
		n2 := v2.Sub(v1).ReversePerp().Normalize()

		offset := n1.Add(n2).Mult(1.0 / (n1.Dot(n2) + 1.0))
		extrude[i] = ExtrudeVerts{offset, n2}
	}

	inset := -math.Max(0, 1.0/DrawPointLineScale-radius)
	for i := 0; i < count-2; i++ {
		v0 := v2f(verts[0].Add(extrude[0].offset.Mult(inset)))
		v1 := v2f(verts[i+1].Add(extrude[i+1].offset.Mult(inset)))
		v2 := v2f(verts[i+2].Add(extrude[i+2].offset.Mult(inset)))

		cpr.triangles = append(cpr.triangles, Triangle{
			Vertex{v0, vf{}, fill, fill},
			Vertex{v1, vf{}, fill, fill},
			Vertex{v2, vf{}, fill, fill},
		})
	}

	outset := 1.0/DrawPointLineScale + radius - inset
	j := count - 1
	for i := 0; i < count; {
		vA := verts[i]
		vB := verts[j]

		nA := extrude[i].n
		nB := extrude[j].n

		offsetA := extrude[i].offset
		offsetB := extrude[j].offset

		innerA := vA.Add(offsetA.Mult(inset))
		innerB := vB.Add(offsetB.Mult(inset))

		inner0 := v2f(innerA)
		inner1 := v2f(innerB)
		outer0 := v2f(innerA.Add(nB.Mult(outset)))
		outer1 := v2f(innerB.Add(nB.Mult(outset)))
		outer2 := v2f(innerA.Add(offsetA.Mult(outset)))
		outer3 := v2f(innerA.Add(nA.Mult(outset)))

		n0 := v2f(nA)
		n1 := v2f(nB)
		offset0 := v2f(offsetA)

		cpr.triangles = append(cpr.triangles, Triangle{
			Vertex{inner0, vf{}, fill, outline},
			Vertex{inner1, vf{}, fill, outline},
			Vertex{outer1, n1, fill, outline},
		})
		cpr.triangles = append(cpr.triangles, Triangle{
			Vertex{inner0, vf{}, fill, outline},
			Vertex{outer0, n1, fill, outline},
			Vertex{outer1, n1, fill, outline},
		})
		cpr.triangles = append(cpr.triangles, Triangle{
			Vertex{inner0, vf{}, fill, outline},
			Vertex{outer0, n1, fill, outline},
			Vertex{outer2, offset0, fill, outline},
		})
		cpr.triangles = append(cpr.triangles, Triangle{
			Vertex{inner0, vf{}, fill, outline},
			Vertex{outer2, offset0, fill, outline},
			Vertex{outer3, n0, fill, outline},
		})

		j = i
		i++
	}
}

func (cpr *CPRenderer) DrawDot(size float64, pos cp.Vector, fill FColor) {
	r := size * 0.5 / DrawPointLineScale
	a := Vertex{vf{float32(pos.X - r), float32(pos.Y - r)}, vf{-1, -1}, fill, fill}
	b := Vertex{vf{float32(pos.X - r), float32(pos.Y + r)}, vf{-1, 1}, fill, fill}
	c := Vertex{vf{float32(pos.X + r), float32(pos.Y + r)}, vf{1, 1}, fill, fill}
	d := Vertex{vf{float32(pos.X + r), float32(pos.Y - r)}, vf{1, -1}, fill, fill}

	cpr.triangles = append(cpr.triangles, Triangle{a, b, c})
	cpr.triangles = append(cpr.triangles, Triangle{a, c, d})
}

func (cpr *CPRenderer) DrawBB(bb cp.BB, outline FColor) {
	verts := []cp.Vector{
		{bb.R, bb.B},
		{bb.R, bb.T},
		{bb.L, bb.T},
		{bb.L, bb.B},
	}
	cpr.DrawPolygon(4, verts, 0, outline, FColor{})
}

func (cpr *CPRenderer) FlushRenderer() {
	gl.BindBuffer(gl.ARRAY_BUFFER, cpr.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(cpr.triangles)*(48*3), gl.Ptr(cpr.triangles), gl.STREAM_DRAW)

	gl.UseProgram(cpr.shader.ID)
	gl.Uniform1f(gl.GetUniformLocation(cpr.shader.ID, gl.Str("u_outline_coef\x00")), DrawPointLineScale)

	gl.BindVertexArray(cpr.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(cpr.triangles)*3))
	CheckGLErrors()
}

func (cpr *CPRenderer) ClearRenderer() {
	cpr.triangles = cpr.triangles[:0]
}
