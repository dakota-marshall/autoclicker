package main

import (
	"fmt"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

// Autoclicker Related Functions

func autoClicker(keystate <-chan bool) {

	for {

		keystate := <-keystate

		if keystate {
			robotgo.Click()
			// robotgo.KeyTap("w")
		}
	}
}

func updateKeyState(keystate chan<- bool, state bool) func(e hook.Event) {
	return func(e hook.Event) {
		// fmt.Println("Setting keystate to: " + strconv.FormatBool(state))
		keystate <- state
	}
}

func eventHooks(keystate chan<- bool, keybind string) {

	fmt.Println("Registering key events for: " + keybind)

	// Register our keystate hooks
	hook.Register(hook.KeyDown, []string{keybind}, updateKeyState(keystate, true))

	hook.Register(hook.KeyUp, []string{keybind}, updateKeyState(keystate, false))
	key_checker := hook.Start()
	<-hook.Process(key_checker)

}

// UI Related Functions

func updateKeybind(new_key string, keybind *string) {

	keys := [46]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "q", "w", "e", "r", "t", "y", "u", "i", "o", "p", "[", "]", "a", "s", "d", "f", "g", "h", "j", "k", "l", ";", "z", "x", "c", "v", "b", "n", "m", ",", ".", "/", "-", "="}

	fmt.Println("Updating keybind...")
	for _, key := range keys {

		if key == new_key {
			*keybind = new_key
			fmt.Println("Keybind updated to: " + new_key)

		}
	}
}

func startEventHooks(keystate chan<- bool, keybind *string, status_label *widget.Label) func() {
	return func() {
		if status_label.Text == "Off" {
			go eventHooks(keystate, *keybind)
			status_label.SetText("On")
		} else {
			fmt.Println("Hooks already enabled...")
		}
	}
}

func stopEventHooks(status_label *widget.Label) func() {
	return func() {
		if status_label.Text == "On" {
			hook.End()
			status_label.SetText("Off")
		} else {
			fmt.Println("Hooks already disabled...")
		}
	}
}

func main() {

	fmt.Println("Starting autoclicker...")

	// Create channels
	keystate := make(chan bool, 2)
	keystate <- false

	keybind := "p"

	// Run our key event thread
	go eventHooks(keystate, keybind)

	// Run our autoclicker thread
	go autoClicker(keystate)

	// Create the window
	application := app.New()
	window := application.NewWindow("Autoclicker")

	status_label := widget.NewLabel("Status: ")
	clicker_status_label := widget.NewLabel("On")
	status_text := container.New(layout.NewHBoxLayout(), status_label, clicker_status_label)

	keybind_label := widget.NewLabel("Button to auto-click:")
	// keybind_key_label := widget.NewLabel(keybind)
	keybind_key := widget.NewEntry()
	keybind_key.SetText(keybind)
	keybind_key.OnChanged = func(input string) { updateKeybind(input, &keybind) }
	keybind_text := container.New(layout.NewHBoxLayout(), keybind_label, keybind_key)

	hooks_start_button := widget.NewButton("Start Autoclicker", startEventHooks(keystate, &keybind, clicker_status_label))
	hooks_stop_button := widget.NewButton("Stop Autoclicker", stopEventHooks(clicker_status_label))

	content := container.New(layout.NewVBoxLayout(), status_text, keybind_text, hooks_start_button, hooks_stop_button)

	window.SetContent(content)
	window.ShowAndRun()

}
