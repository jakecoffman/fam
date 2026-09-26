package gui

import (
	"testing"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/inkyblackness/imgui-go"
)

func TestKeyboardEventsReachImGui(t *testing.T) {
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	platform := &GLFW{imguiIO: imgui.CurrentIO()}
	platform.setKeyMapping()
	platform.keyChange(nil, glfw.KeyA, 0, glfw.Press, 0)
	if !imgui.IsKeyDown(int(glfw.KeyA)) {
		t.Error("key press did not reach ImGui")
	}
	platform.keyChange(nil, glfw.KeyA, 0, glfw.Release, 0)
	if imgui.IsKeyDown(int(glfw.KeyA)) {
		t.Error("key release did not reach ImGui")
	}
	platform.keyChange(nil, glfw.KeyUnknown, 0, glfw.Press, 0)
}

func TestShortMouseClicksRemainLatched(t *testing.T) {
	platform := &GLFW{}
	platform.mouseButtonChange(nil, glfw.MouseButton1, glfw.Press, 0)
	platform.mouseButtonChange(nil, glfw.MouseButton1, glfw.Release, 0)
	if !platform.mouseJustPressed[0] {
		t.Error("press and release between frames lost the click")
	}
}
