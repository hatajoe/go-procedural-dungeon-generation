package main

import (
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"
)

var (
	ballRadius = 25
	ballMass   = 1

	space       *chipmunk.Space
	balls       []*chipmunk.Shape
	boxes       []*chipmunk.Shape
	staticLines []*chipmunk.Shape
	deg2rad     = math.Pi / 180
)

func drawRoom(box *chipmunk.Shape) {
	h := float64(box.GetAsBox().Height)
	w := float64(box.GetAsBox().Width)
	x1 := float64(0 - w*0.5)
	y1 := float64(0 - h*0.5)
	x2 := float64(0 + w*0.5)
	y2 := float64(0 - h*0.5)
	x3 := float64(0 + w*0.5)
	y3 := float64(0 + h*0.5)
	x4 := float64(0 - w*0.5)
	y4 := float64(0 + h*0.5)

	gl.Begin(gl.POLYGON)
	gl.Color4f(.3, .3, 1, .2)
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

// drawCircle draws a circle for the specified radius, rotation angle, and the specified number of sides
func drawCircle(radius float64, sides int) {
	gl.Begin(gl.LINE_LOOP)
	for a := 0.0; a < 2*math.Pi; a += (2 * math.Pi / float64(sides)) {
		gl.Vertex2d(math.Sin(a)*radius, math.Cos(a)*radius)
	}
	gl.Vertex3f(0, 0, 0)
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
	//gl.Begin(gl.LINE_LOOP)
	//gl.Vertex3d(-x, y, float64(h))
	//gl.Vertex3d(x, y, float64(h))
	//gl.Vertex3d(x, -y, float64(h))
	//gl.Vertex3d(-x, -y, float64(h))
	//gl.End()
	//gl.LineWidth(1.0)

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

	//gl.LineWidth(1.0)
	gl.PopMatrix()

	//gl.Begin(gl.LINES)
	//gl.Color3f(.2, .2, .2)
	//for i := range staticLines {
	//	x := staticLines[i].GetAsSegment().A.X
	//	y := staticLines[i].GetAsSegment().A.Y
	//	gl.Vertex3f(float32(x), float32(y), 0)
	//	x = staticLines[i].GetAsSegment().B.X
	//	y = staticLines[i].GetAsSegment().B.Y
	//	gl.Vertex3f(float32(x), float32(y), 0)
	//}
	//gl.End()

	// draw boxes
	for _, box := range boxes {
		gl.PushMatrix()
		rot := box.Body.Angle() * chipmunk.DegreeConst
		gl.Rotatef(float32(rot), 0, 0, 1.0)
		//gl.Translatef(float32(box.Body.Position().X), float32(box.Body.Position().Y), 0.0)
		x := roundm(float64(box.Body.Position().X), 4.0)
		y := roundm(float64(box.Body.Position().Y), 4.0)
		gl.Translated(x, y, 0.0)
		drawRoom(box)
		gl.PopMatrix()
	}

	//// draw balls
	//for _, ball := range balls {
	//	gl.PushMatrix()
	//	pos := ball.Body.Position()
	//	rot := ball.Body.Angle() * chipmunk.DegreeConst
	//	gl.Translatef(float32(pos.X), float32(pos.Y), 0.0)
	//	gl.Rotatef(float32(rot), 0, 0, 1)
	//	drawCircle(float64(ballRadius), 60)
	//	gl.PopMatrix()
	//}
}

func addRoom(pos vect.Vect, w vect.Float, h vect.Float) {
	box := chipmunk.NewBox(vect.Vector_Zero, w, h)
	box.SetElasticity(0.5)
	body := chipmunk.NewBody(1.0, box.Moment(float32(1.0)))
	body.SetPosition(pos)
	body.AddShape(box)
	//space.AddBody(body)
	boxes = append(boxes, box)
}

func addBall() {
	x := rand.Intn(350-115) + 115
	ball := chipmunk.NewCircle(vect.Vector_Zero, float32(ballRadius))
	ball.SetElasticity(0.95)

	body := chipmunk.NewBody(vect.Float(ballMass), ball.Moment(float32(ballMass)))
	body.SetPosition(vect.Vect{vect.Float(x), 600.0})
	body.SetAngle(0.0)

	body.AddShape(ball)
	space.AddBody(body)
	balls = append(balls, ball)
}

func setRoomToSpace() {
	for _, v := range boxes {
		space.AddBody(v.Body)
	}
}

func waitForSleep() bool {
	for _, v := range boxes {
		if !v.Body.IsSleeping() {
			return false
		}
	}
	return true
}

// step advances the physics engine and cleans up any balls that are off-screen
func step(dt float32) {
	space.Step(vect.Float(dt))

	for i := 0; i < len(boxes); i++ {
		//x := boxes[i].Body.Position().X
		//y := boxes[i].Body.Position().Y
		//x = vect.Float(roundm(float64(x), 4.0))
		//y = vect.Float(roundm(float64(y), 4.0))
		//boxes[i].Body.SetPosition(vect.Vect{
		//	X: x,
		//	Y: y,
		//})
		boxes[i].Body.SetAngle(vect.Float(0))
	}

	//for i := 0; i < len(balls); i++ {
	//	p := balls[i].Body.Position()
	//	if p.Y < -100 {
	//		space.RemoveBody(balls[i].Body)
	//		balls[i] = nil
	//		balls = append(balls[:i], balls[i+1:]...)
	//		i-- // consider same index again
	//	}
	//}
}

// createBodies sets up the chipmunk space and static bodies
func createBodies() {
	space = chipmunk.NewSpace()
	//space.Gravity = vect.Vect{0, -900}
	//
	//	staticBody := chipmunk.NewBodyStatic()
	//	staticLines = []*chipmunk.Shape{
	//		chipmunk.NewSegment(vect.Vect{111.0, 280.0}, vect.Vect{407.0, 246.0}, 0),
	//		chipmunk.NewSegment(vect.Vect{407.0, 246.0}, vect.Vect{407.0, 343.0}, 0),
	//	}
	//	for _, segment := range staticLines {
	//		segment.SetElasticity(0.6)
	//		staticBody.AddShape(segment)
	//	}
	//	space.AddBody(staticBody)
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
func getRandomPointInCircle(radius float64) vect.Vect {
	t := 2 * math.Pi * rand.Float64()
	u := rand.Float64() + rand.Float64()
	r := .0
	if u > 1 {
		r = 2 - u
	} else {
		r = u
	}
	return vect.Vect{
		X: vect.Float(roundm(radius*r*math.Cos(t), 4.0)) + 300,
		Y: vect.Float(roundm(radius*r*math.Sin(t), 4.0)) + 300,
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
	createBodies()

	glfw.SwapInterval(1)

	phase := 0
	//ticksToNextBall := 10
	ticker := time.NewTicker(time.Second / 60)
	for !window.ShouldClose() {
		//ticksToNextBall--
		//if ticksToNextBall == 0 {
		//	ticksToNextBall = rand.Intn(100) + 1
		//addBall()
		switch phase {
		case 0:
			pos := getRandomPointInCircle(100.0)
			w := vect.Float(roundm(float64(rand.Intn(28)+8.0), 4.0) * 2.0)
			h := vect.Float(roundm(float64(rand.Intn(28)+8.0), 4.0) * 2.0)
			addRoom(pos, w, h)
			if len(boxes) > 50 {
				phase = 1
				fmt.Println("phase1")
			}
		case 1:
			setRoomToSpace()
			phase = 2
			fmt.Println("phase2")
		case 2:
			if waitForSleep() {
				fmt.Println("phase3")
				phase = 3
			}
		case 3:
		}
		//}
		draw(window)
		step(1.0 / 60.0)
		window.SwapBuffers()
		glfw.PollEvents()
		<-ticker.C // wait up to 1/60th of a second
	}
}
