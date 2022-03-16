package main

import (
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"
)

// Room is room
type Room struct {
	Color mgl32.Vec4
	Shape *cp.Shape
}

var (
	space *cp.Space
	rooms []*Room
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
	gl.Color4f(.3, .3, 1, .9)
	gl.LineWidth(1.0)
	gl.Vertex2d(x1, y1)
	gl.Vertex2d(x2, y2)
	gl.Vertex2d(x3, y3)
	gl.Vertex2d(x4, y4)
	gl.End()
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

	gl.Color4f(.1, .1, .1, .8)
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
		drawRoom(room)
		gl.PopMatrix()
	}
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
			v.Color = mgl32.Vec4{.3, .0, .0, .2}
			sleeping = true
		} else {
			v.Color = mgl32.Vec4{.3, .3, 1, .2}
			sleeping = false
		}
	}
	return sleeping
}

func step(dt float64) {
	space.Step(dt)
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
	ticker := time.NewTicker(time.Second / 60)
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
			phase = 2
			fmt.Println("phase2")
		case 2:
			step(1.0 / 60.0)
			if waitForSleep() {
				phase = 3
				fmt.Println("phase3")
			}
		case 3:
		}
		draw(window)
		window.SwapBuffers()
		glfw.PollEvents()
		<-ticker.C // wait up to 1/60th of a second
	}
}
