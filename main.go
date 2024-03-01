package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

const app_version = "v1.1.1"

// Autoclicker Related Functions

func autoClicker(keystate <-chan bool, delay_chan <-chan int64) {

	var state bool
	var delay int64
	// var toggle bool

	state = false

	for {
		// Do a non blocking read on toggle state
		// select {
		// case toggle_read := <-togglestate:
		// 	toggle = toggle_read
		// default:
		// }

		// fmt.Println(strconv.FormatBool(toggle))

		select {
		case state_read := <-keystate:
			fmt.Println("Value found in keystate channel:")
			state = state_read
			fmt.Println(strconv.FormatBool(state))
		default:
		}

		// Dont block on delay read
		select {
		case delay_read := <-delay_chan:
			delay = delay_read
		default:
		}

		if state {
			robotgo.Click()
		}
		if delay != 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

func updateKeyState(keystate chan<- bool, state bool) func(e hook.Event) {
	return func(e hook.Event) {

		fmt.Println("Keystate Event")
		fmt.Println(strconv.FormatBool(state))

		if state {
			// Dont block to add true's to the queue
			select {
			case keystate <- state:
			default:
			}
		} else {
			// Do block to add false to queue no matter what
			fmt.Println("Block adding false to channel")
			keystate <- state
		}
	}
}

func toggleKeyState(keystate chan<- bool, togglestate chan bool) func(e hook.Event) {
	return func(e hook.Event) {

		var toggle bool
		fmt.Println("Toggling Keystate")
		// toggle state tracker
		toggle = !<-togglestate
		togglestate <- toggle
		// Send state update to autoclicker
		keystate <- toggle
	}
}

func eventHooks(keystate chan<- bool, keybind string, is_toggle bool, togglestate chan bool) {

	fmt.Println("Registering key events for: " + keybind)

	if !is_toggle {
		// Register our keystate hooks
		hook.Register(hook.KeyDown, []string{keybind}, updateKeyState(keystate, true))

		// This seems like it is currently broken
		// It never calls this function on a KeyUp event
		// https://github.com/robotn/gohook/issues/31
		// hook.Register(hook.KeyUp, []string{keybind}, updateKeyState(keystate, false))
	} else {
		hook.Register(hook.KeyDown, []string{keybind}, toggleKeyState(keystate, togglestate))
	}

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

func startEventHooks(keystate chan<- bool, keybind *string, status_label *widget.Label, is_toggle bool, togglestate chan bool) func() {
	return func() {
		if status_label.Text == "Off" {
			go eventHooks(keystate, *keybind, is_toggle, togglestate)
			status_label.SetText("On")
		} else {
			fmt.Println("Hooks already enabled...")
		}
	}
}

func stopEventHooks(status_label *widget.Label, keystate chan<- bool) func() {
	return func() {
		if status_label.Text == "On" {
			keystate <- false
			hook.End()
			status_label.SetText("Off")
		} else {
			fmt.Println("Hooks already disabled...")
		}
	}
}

func updateTogglestate(togglestate chan<- bool) func(value bool) {
	return func(value bool) {
		fmt.Println("Toggling state")
		togglestate <- value
	}
}

func main() {

	fmt.Println("Starting autoclicker...")

	// Create channels
	keystate := make(chan bool, 2)
	keystate <- false

	delay_chan := make(chan int64, 2)
	delay_chan <- 0

	// Toggle state tracker
	togglestate := make(chan bool, 2)
	togglestate <- false

	var delay int64
	delay = 0
	keybind := "p"

	// Run our key event thread
	go eventHooks(keystate, keybind, true, togglestate)

	// Run our autoclicker thread
	go autoClicker(keystate, delay_chan)

	// Create the window
	application := app.New()
	window := application.NewWindow("Autoclicker " + app_version)

	// Status box
	status_label := widget.NewLabel("Status: ")
	clicker_status_label := widget.NewLabel("On")
	status_text := container.New(layout.NewHBoxLayout(), status_label, clicker_status_label)

	// Keybind box
	keybind_label := widget.NewLabel("Button to auto-click:")
	keybind_key := widget.NewEntry()
	keybind_key.SetText(keybind)
	keybind_key.OnChanged = func(input string) { updateKeybind(input, &keybind) }
	//Toggle Box
	// Because KeyUp is broken, setting toggle to the only option
	// toggle_box := widget.NewCheck("Toggle instead of hold?", updateTogglestate(togglestate))
	keybind_text := container.New(layout.NewHBoxLayout(), keybind_label, keybind_key)

	// Delay box
	delay_label := widget.NewLabel("Click Delay (Milliseconds):")
	delay_key := widget.NewEntry()
	delay_key.SetText(strconv.FormatInt(0, 10))
	delay_key.OnChanged = func(input string) { delay, _ = strconv.ParseInt(input, 10, 64) }
	delay_button := widget.NewButton("Update", func() { delay_chan <- delay })
	delay_text := container.New(layout.NewHBoxLayout(), delay_label, delay_key, delay_button)

	// Buttons
	hooks_start_button := widget.NewButton("Start Autoclicker", startEventHooks(keystate, &keybind, clicker_status_label, true, togglestate))
	hooks_stop_button := widget.NewButton("Stop Autoclicker", stopEventHooks(clicker_status_label, keystate))

	// Entire window
	content := container.New(layout.NewVBoxLayout(), status_text, keybind_text, delay_text, hooks_start_button, hooks_stop_button)

	window.SetContent(content)
	window.ShowAndRun()

}
