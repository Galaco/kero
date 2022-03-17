package client

import (
	"github.com/galaco/kero/client/gui"
	"github.com/galaco/kero/client/input"
	"github.com/galaco/kero/client/renderer"
	"github.com/galaco/kero/internal/framework/event"
	"github.com/galaco/kero/internal/framework/graphics/adapter"
	input2 "github.com/galaco/kero/internal/framework/input"
	"github.com/galaco/kero/internal/framework/window"
	"github.com/galaco/kero/shared/messages"
	"github.com/go-gl/mathgl/mgl32"
)

type Client struct {
	scene *sceneEntities

	renderer *renderer.Renderer
	ui       *gui.Gui
	input    *input.Input

	isInMenu bool
}

func (c *Client) FixedUpdate(dt float64) {
	c.scene.Update(dt)
}

func (c *Client) Update() {
	c.input.Poll()
	c.renderer.Render()
	c.ui.Render()

	window.CurrentWindow().SwapBuffers()
	c.renderer.FinishFrame()
}

func (c *Client) onKeyRelease(message interface{}) {
	key := message.(input2.Key)
	if key == input2.KeyEscape {
		c.isInMenu = !c.isInMenu
	}
}

func (c *Client) onMouseMove(message interface{}) {
	if c.isInMenu {
		return
	}
	if c.scene == nil || c.scene.activeCamera == nil {
		return
	}
	msg := message.(mgl32.Vec2)
	c.scene.activeCamera.Rotate(msg[0], 0, msg[1])
}

func (c *Client) Initialize() error {
	// Creates the Client Game Window

	win, err := window.CreateWindow(1920, 1080, "Kero: A Source Engine Implementation")
	if err != nil {
		return err
	}
	win.SetActive()
	input2.SetBoundWindow(win)
	win.Handle().Handle().Focus()
	if err = adapter.Init(); err != nil {
		return err
	}

	// Bind to the input library for window handling
	input.InputMiddleware().AddListener(messages.TypeKeyRelease, c.onKeyRelease)
	input.InputMiddleware().AddListener(messages.TypeMouseMove, c.onMouseMove)
	event.Get().AddListener(messages.TypeEngineQuit, func(interface{}) {
		window.CurrentWindow().Close()
	})

	// Initialize our rendering system
	c.renderer.Initialize()
	c.ui.Initialize()

	// Bind our client data to shared/server resources
	c.scene.BindSharedResources()
	c.renderer.BindSharedResources()

	return nil
}

func (c *Client) ShouldClose() bool {
	return window.CurrentWindow() == nil || window.CurrentWindow().ShouldClose()
}

func (c *Client) Cleanup() {
	c.renderer.Cleanup()
}

func NewClient() *Client {
	return &Client{
		renderer: renderer.NewRenderer(),
		ui:       gui.NewGui(),
		input:    input.InitializeInput(),
		scene:    &sceneEntities{},
		isInMenu: true, // Client always starts in menu
	}
}
