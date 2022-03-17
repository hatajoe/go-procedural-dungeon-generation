package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/fogleman/delaunay"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

// Room is room
type Room struct {
	Color    mgl32.Vec4
	Shape    *cp.Shape
	Selected bool
}

var (
	space         *cp.Space
	rooms         []*Room
	selected      []*Room
	triangulation *delaunay.Triangulation
	edges         graph.WeightedEdges
)

func drawRoom(room *Room) {
	h := float64(room.Shape.BB().B - room.Shape.BB().T)
	w := float64(room.Shape.BB().R - room.Shape.BB().L)
	x1 := float64(0 - w*0.5)
	y1 := float64(0 - h*0.5)
	x2 := float64(0 + w*0.5)
	y2 := float64(0 - h*0.5)
	x3 := float64(0 + w*0.5)
	y3 := float64(0 + h*0.5)
	x4 := float64(0 - w*0.5)
	y4 := float64(0 + h*0.5)

	gl.Begin(gl.POLYGON)
	gl.Color4f(
		room.Color.X(),
		room.Color.Y(),
		room.Color.Z(),
		room.Color.W(),
	)
	gl.Vertex2d(x1, y1)
	gl.Vertex2d(x2, y2)
	gl.Vertex2d(x3, y3)
	gl.Vertex2d(x4, y4)
	gl.End()
	gl.Begin(gl.LINE_LOOP)
	gl.Color4f(
		room.Color.X()*0.5,
		room.Color.Y()*0.5,
		room.Color.Z()*0.5,
		room.Color.W(),
	)
	gl.LineWidth(1.0)
	gl.Vertex2d(x1, y1)
	gl.Vertex2d(x2, y2)
	gl.Vertex2d(x3, y3)
	gl.Vertex2d(x4, y4)
	gl.End()
}

func drawTriangles() {
	if edges != nil {
		// draw edges
		hashMap := map[int64]*Room{}
		for _, room := range selected {
			hashMap[int64(room.Shape.HashId())] = room
		}
		for edges.Next() {
			p := hashMap[edges.WeightedEdge().From().ID()].Shape.BB().Center()
			q := hashMap[edges.WeightedEdge().To().ID()].Shape.BB().Center()
			gl.Begin(gl.LINE_LOOP)
			gl.Color4f(.9, .9, 0, .9)
			gl.LineWidth(1.0)
			gl.Vertex2d(p.X, p.Y)
			gl.Vertex2d(q.X, q.Y)
			gl.End()
		}
		edges.Reset()
	} else if triangulation != nil {
		// draw triagulation
		ts := triangulation.Triangles
		hs := triangulation.Halfedges
		for i, h := range hs {
			if i > h {
				p := selected[ts[i]].Shape.BB().Center()
				q := selected[ts[nextHalfEdge(i)]].Shape.BB().Center()
				gl.Begin(gl.LINE_LOOP)
				gl.Color4f(.9, .9, 0, .9)
				gl.LineWidth(1.0)
				gl.Vertex2d(p.X, p.Y)
				gl.Vertex2d(q.X, q.Y)
				gl.End()
			}
		}
	}
}

// OpenGL draw function
func draw(window *glfw.Window) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Enable(gl.BLEND)
	gl.Enable(gl.POINT_SMOOTH)
	gl.Enable(gl.LINE_SMOOTH)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.LoadIdentity()

	gl.PushMatrix()

	gl.Disable(gl.LIGHTING)

	width, height := window.GetSize()
	x := float64(width)
	y := float64(height)
	h := 0

	gl.Color4f(.1, .1, .1, .4)
	gl.LineWidth(1.0)

	// x方向
	var x0, x1, y0, y1 float64
	var deltaX, deltaY float64
	d := width / 2

	x0 = -x
	x1 = -x
	y0 = -y
	y1 = y
	deltaX = ((2 * x) / float64(d))

	for i := 0; i < d; i++ {
		x0 = x0 + deltaX
		gl.Begin(gl.LINES)
		gl.Vertex3f(float32(x0), float32(y0), float32(h))
		gl.Vertex3f(float32(x0), float32(y1), float32(h))
		gl.End()
	}

	// y方向
	x0 = -x
	x1 = x
	deltaY = ((2 * y) / float64(d))

	for i := 0; i < d; i++ {
		y0 = y0 + deltaY
		gl.Begin(gl.LINES)
		gl.Vertex3f(float32(x0), float32(y0), float32(h))
		gl.Vertex3f(float32(x1), float32(y0), float32(h))
		gl.End()
	}

	gl.PopMatrix()

	// draw boxes
	for _, room := range rooms {
		gl.PushMatrix()

		x := roundm(float64(room.Shape.Body().Position().X), 4.0)
		y := roundm(float64(room.Shape.Body().Position().Y), 4.0)
		gl.Translated(x, y, 0.0)
		if room.Selected {
			room.Color = mgl32.Vec4{.8, .0, 0, .6}
		}
		drawRoom(room)
		gl.PopMatrix()
	}
	drawTriangles()
}

func nextHalfEdge(e int) int {
	if e%3 == 2 {
		return e - 2
	}
	return e + 1
}

func addRoom(pos cp.Vector, w float64, h float64) {
	body := space.AddBody(cp.NewBody(1.0, cp.INFINITY))
	body.SetPosition(pos)

	shape := space.AddShape(cp.NewBox(body, w+0.5, h+0.5, 0))
	shape.SetElasticity(1)
	shape.SetFriction(0.0)

	room := Room{
		Color: mgl32.Vec4{.3, .3, 1, .2},
		Shape: shape,
	}
	rooms = append(rooms, &room)
}

func waitForSleep() bool {
	sleeping := false
	for _, v := range rooms {
		if v.Shape.Body().IsSleeping() {
			v.Color = mgl32.Vec4{.6, .6, .6, .2}
			sleeping = true
		} else {
			v.Color = mgl32.Vec4{.3, .3, 1, .6}
			sleeping = false
		}
	}
	return sleeping
}

func step(dt float64) {
	space.Step(dt)
}

func selectRoom() bool {
	selected := false
	for _, room := range rooms {
		if room.Selected {
			continue
		}
		if room.Shape.Area() > 3500.0 {
			room.Selected = true
			selected = true
			break
		}
	}
	if !selected {
		return false
	}
	return true
}

func triangulate() (err error) {
	points := []delaunay.Point{}
	for _, room := range rooms {
		if !room.Selected {
			continue
		}
		points = append(points, delaunay.Point{
			X: room.Shape.BB().Center().X,
			Y: room.Shape.BB().Center().Y,
		})
		selected = append(selected, room)
	}
	triangulation, err = delaunay.Triangulate(points)
	return err
}

// createSpace sets up the chipmunk space and static bodies
func createSpace() {
	space = cp.NewSpace()
	space.SleepTimeThreshold = 3
}

// onResize sets up a simple 2d ortho context based on the window size
func onResize(window *glfw.Window, w, h int) {
	w, h = window.GetSize() // query window to get screen pixels
	width, height := window.GetFramebufferSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, float64(w), 0, float64(h), -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.ClearColor(1, 1, 1, 1)
}

func roundm(n, m float64) float64 {
	return math.Floor(((n + m - 1.0) / m)) * m
}

func getRandomPointInCircle(radius float64) cp.Vector {
	t := 2 * math.Pi * rand.Float64()
	u := rand.Float64() + rand.Float64()
	r := .0
	if u > 1 {
		r = 2 - u
	} else {
		r = u
	}
	return cp.Vector{
		X: roundm(radius*r*math.Cos(t), 4.0) + 300,
		Y: roundm(radius*r*math.Sin(t), 4.0) + 300,
	}
}

func main() {
	runtime.LockOSThread()

	rand.Seed(time.Now().UnixNano())

	// initialize glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize GLFW: ", err)
	}
	defer glfw.Terminate()

	// create window
	window, err := glfw.CreateWindow(600, 600, os.Args[0], nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.SetFramebufferSizeCallback(onResize)
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatal(err)
	}

	// set up opengl context
	onResize(window, 600, 600)

	// set up physics
	createSpace()

	glfw.SwapInterval(1)

	phase := 0
	ticker := time.NewTicker(time.Second / 30)
	for !window.ShouldClose() {
		switch phase {
		case 0:
			pos := getRandomPointInCircle(100.0)
			w := roundm(float64(rand.Intn(28)+8.0), 4.0) * 2.0
			h := roundm(float64(rand.Intn(28)+8.0), 4.0) * 2.0
			addRoom(pos, w, h)
			if len(rooms) > 50 {
				phase = 1
				fmt.Println("phase1")
			}
		case 1:
			bufio.NewScanner(os.Stdin).Scan()
			phase = 2
			fmt.Println("phase2")
		case 2:
			step(1.0 / 60.0)
			if waitForSleep() {
				phase = 3
				fmt.Println("phase3")
			}
		case 3:
			if !selectRoom() {
				phase = 4
				bufio.NewScanner(os.Stdin).Scan()
				fmt.Println("phase4")
			}
		case 4:
			if err := triangulate(); err != nil {
				panic(err)
			}
			phase = 5
			fmt.Println("phase5")
		case 5:
			bufio.NewScanner(os.Stdin).Scan()
			g := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
			dst := simple.NewWeightedUndirectedGraph(0, math.Inf(1))
			ts := triangulation.Triangles
			hs := triangulation.Halfedges
			for i, h := range hs {
				if i > h {
					p := selected[ts[i]]
					q := selected[ts[nextHalfEdge(i)]]
					g.SetWeightedEdge(simple.WeightedEdge{
						F: simple.Node(p.Shape.HashId()),
						T: simple.Node(q.Shape.HashId()),
						W: p.Shape.BB().Center().DistanceSq(q.Shape.BB().Center()),
					})
				}
			}
			path.Prim(dst, g)
			edges = dst.WeightedEdges()
			phase = 6
			fmt.Println("phase6")
		case 6:
		}
		draw(window)
		window.SwapBuffers()
		glfw.PollEvents()
		<-ticker.C // wait up to 1/60th of a second
	}
}
