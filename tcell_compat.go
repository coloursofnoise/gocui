// Copyright 2020 The TCell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// This is copy from original github.com/gdamore/tcell/termbox package
// for easier adoption of tcell.
// There are some changes made, to make it work with termbox keys (Ctrl modifier)

package gocui

import (
	"github.com/gdamore/tcell/v2"
)

var screen tcell.Screen
var outMode OutputMode

// Init initializes the screen for use.
func Init() error {
	outMode = OutputNormal
	if s, e := tcell.NewScreen(); e != nil {
		return e
	} else if e = s.Init(); e != nil {
		return e
	} else {
		screen = s
		return nil
	}
}

// Close cleans up the terminal, restoring terminal modes, etc.
func Close() {
	screen.Fini()
}

// Flush updates the screen.
func Flush() error {
	screen.Show()
	return nil
}

// SetCursor displays the terminal cursor at the given location.
func SetCursor(x, y int) {
	screen.ShowCursor(x, y)
}

// HideCursor hides the terminal cursor.
func HideCursor() {
	SetCursor(-1, -1)
}

// Size returns the screen size as width, height in character cells.
func Size() (int, int) {
	return screen.Size()
}

// Attribute affects the presentation of characters, such as color, boldness,
// and so forth.
type Attribute uint64

const (
	// ColorDefault is used to leave the Color unchanged from whatever system or teminal default may exist.
	ColorDefault = Attribute(tcell.ColorValid) // We are going to use reverse method to see if color is setup than tcell does

	// AttrIsRGBColor is used to indicate that the Attribute value is RGB value of color.
	// The lower order 3 bytes are RGB.
	// (It's not a color in basic ANSI range 256).
	AttrIsRGBColor = Attribute(tcell.ColorIsRGB)

	// AttrColorBits is a mask where color is located in Attribute
	AttrColorBits = 0xffffffffff // roughly 5 bytes, tcell uses 4 bytes and half-byte as a special flags for color (rest is reserved for future)

	// AttrStyleBits is a mask where character attributes (e.g.: bold, italic, underline) are located in Attribute
	AttrStyleBits = 0xffffff0000000000 // remaining 3 bytes in the 8 bytes Attribute (tcell is not using it, so we should be fine)
)

// Colors first.  The order here is significant.
const (
	ColorBlack Attribute = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite
)

// Attributes are not colors, but affect the display of text.  They can
// be combined.
const (
	AttrBold Attribute = 1 << (40 + iota)
	AttrBlink
	AttrReverse
	AttrUnderline
	AttrDim
	AttrItalic
	AttrNone Attribute = 0 // Just normal text.
)

// AttrAll is all the attributes turned on
const AttrAll = AttrBold | AttrBlink | AttrReverse | AttrUnderline | AttrDim | AttrItalic

// fixColor transform  Attribute into tcell.Color
func fixColor(c Attribute) tcell.Color {
	c = c & AttrColorBits
	if c&ColorDefault != 0 {
		return tcell.ColorDefault
	}

	tc := tcell.Color(c)
	switch outMode {
	case OutputTrue:
		return tc | tcell.ColorValid
	case OutputNormal:
		tc %= tcell.Color(16)
	case Output256:
		tc %= tcell.Color(256)
	case Output216:
		tc %= tcell.Color(216)
		tc += tcell.Color(16)
	case OutputGrayscale:
		tc %= tcell.Color(24)
		tc += tcell.Color(232)
	default:
		return tcell.ColorDefault
	}
	return tc | tcell.ColorValid
}

func mkStyle(fg, bg Attribute) tcell.Style {
	st := tcell.StyleDefault

	// extract colors and attributes
	if fg != ColorDefault {
		st = st.Foreground(fixColor(fg))
		st = setAttr(st, fg)
	}
	if bg != ColorDefault {
		st = st.Background(fixColor(bg))
		st = setAttr(st, bg)
	}

	return st
}

func setAttr(st tcell.Style, attr Attribute) tcell.Style {
	if attr&AttrBold != 0 {
		st = st.Bold(true)
	}
	if attr&AttrUnderline != 0 {
		st = st.Underline(true)
	}
	if attr&AttrReverse != 0 {
		st = st.Reverse(true)
	}
	if attr&AttrBlink != 0 {
		st = st.Blink(true)
	}
	if attr&AttrDim != 0 {
		st = st.Dim(true)
	}
	if attr&AttrItalic != 0 {
		st = st.Italic(true)
	}
	return st
}

// GetColor creates a Color from a color name (W3C name). A hex value may
// be supplied as a string in the format "#ffffff".
func GetColor(color string) Attribute {
	c := tcell.GetColor(color)
	if c&tcell.ColorValid != 0 {
		// reverse ColorValid for Attribute
		c &= ^tcell.ColorValid
		return Attribute(c)
	}

	return ColorDefault
}

// Clear clears the screen with the given attributes.
func Clear(fg, bg Attribute) {
	st := mkStyle(fg, bg)
	w, h := screen.Size()
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			screen.SetContent(col, row, ' ', nil, st)
		}
	}
}

// InputMode is not used.
type InputMode int

// Unused input modes; here for compatibility.
const (
	InputEsc InputMode = 1 << iota
	InputAlt
	InputMouse
	InputCurrent InputMode = 0
)

// SetInputMode does not do anything in this version.
func SetInputMode(mode InputMode) InputMode {
	if mode&InputMouse != 0 {
		screen.EnableMouse()
		return InputEsc | InputMouse
	}
	// We don't do anything else right now
	return InputEsc
}

// OutputMode represents an output mode, which determines how colors
// are used.  See the termbox documentation for an explanation.
type OutputMode int

// OutputMode values.
const (
	OutputCurrent OutputMode = iota
	OutputNormal
	Output256
	Output216
	OutputGrayscale
	OutputTrue
)

// SetOutputMode is used to set the color palette used.
func SetOutputMode(mode OutputMode) OutputMode {
	if screen.Colors() < 256 {
		mode = OutputNormal
	}
	switch mode {
	case OutputCurrent:
		return outMode
	case OutputNormal, Output256, Output216, OutputGrayscale, OutputTrue:
		outMode = mode
		return mode
	default:
		return outMode
	}
}

// Sync forces a resync of the screen.
func Sync() error {
	screen.Sync()
	return nil
}

// SetCell sets the character cell at a given location to the given
// content (rune) and attributes.
func SetCell(x, y int, ch rune, fg, bg Attribute) {
	st := mkStyle(fg, bg)
	screen.SetContent(x, y, ch, nil, st)
}

// EventType represents the type of event.
type EventType uint8

// Modifier represents the possible modifier keys.
type Modifier tcell.ModMask

// Key is a key press.
type Key tcell.Key

// Event represents an event like a key press, mouse action, or window resize.
type Event struct {
	Type   EventType
	Mod    Modifier
	Key    Key
	Ch     rune
	Width  int
	Height int
	Err    error
	MouseX int
	MouseY int
	N      int
}

// Event types.
const (
	EventNone EventType = iota
	EventKey
	EventResize
	EventMouse
	EventInterrupt
	EventError
	EventRaw
)

// Keys codes.
const (
	KeyF1             = Key(tcell.KeyF1)
	KeyF2             = Key(tcell.KeyF2)
	KeyF3             = Key(tcell.KeyF3)
	KeyF4             = Key(tcell.KeyF4)
	KeyF5             = Key(tcell.KeyF5)
	KeyF6             = Key(tcell.KeyF6)
	KeyF7             = Key(tcell.KeyF7)
	KeyF8             = Key(tcell.KeyF8)
	KeyF9             = Key(tcell.KeyF9)
	KeyF10            = Key(tcell.KeyF10)
	KeyF11            = Key(tcell.KeyF11)
	KeyF12            = Key(tcell.KeyF12)
	KeyInsert         = Key(tcell.KeyInsert)
	KeyDelete         = Key(tcell.KeyDelete)
	KeyHome           = Key(tcell.KeyHome)
	KeyEnd            = Key(tcell.KeyEnd)
	KeyArrowUp        = Key(tcell.KeyUp)
	KeyArrowDown      = Key(tcell.KeyDown)
	KeyArrowRight     = Key(tcell.KeyRight)
	KeyArrowLeft      = Key(tcell.KeyLeft)
	KeyCtrlA          = Key(tcell.KeyCtrlA)
	KeyCtrlB          = Key(tcell.KeyCtrlB)
	KeyCtrlC          = Key(tcell.KeyCtrlC)
	KeyCtrlD          = Key(tcell.KeyCtrlD)
	KeyCtrlE          = Key(tcell.KeyCtrlE)
	KeyCtrlF          = Key(tcell.KeyCtrlF)
	KeyCtrlG          = Key(tcell.KeyCtrlG)
	KeyCtrlH          = Key(tcell.KeyCtrlH)
	KeyCtrlI          = Key(tcell.KeyCtrlI)
	KeyCtrlJ          = Key(tcell.KeyCtrlJ)
	KeyCtrlK          = Key(tcell.KeyCtrlK)
	KeyCtrlL          = Key(tcell.KeyCtrlL)
	KeyCtrlM          = Key(tcell.KeyCtrlM)
	KeyCtrlN          = Key(tcell.KeyCtrlN)
	KeyCtrlO          = Key(tcell.KeyCtrlO)
	KeyCtrlP          = Key(tcell.KeyCtrlP)
	KeyCtrlQ          = Key(tcell.KeyCtrlQ)
	KeyCtrlR          = Key(tcell.KeyCtrlR)
	KeyCtrlS          = Key(tcell.KeyCtrlS)
	KeyCtrlT          = Key(tcell.KeyCtrlT)
	KeyCtrlU          = Key(tcell.KeyCtrlU)
	KeyCtrlV          = Key(tcell.KeyCtrlV)
	KeyCtrlW          = Key(tcell.KeyCtrlW)
	KeyCtrlX          = Key(tcell.KeyCtrlX)
	KeyCtrlY          = Key(tcell.KeyCtrlY)
	KeyCtrlZ          = Key(tcell.KeyCtrlZ)
	KeyCtrlUnderscore = Key(tcell.KeyCtrlUnderscore)
	KeyBackspace      = Key(tcell.KeyBackspace)
	KeyBackspace2     = Key(tcell.KeyBackspace2)
	KeyTab            = Key(tcell.KeyTab)
	KeyEnter          = Key(tcell.KeyEnter)
	KeyEsc            = Key(tcell.KeyEscape)
	KeyPgdn           = Key(tcell.KeyPgDn)
	KeyPgup           = Key(tcell.KeyPgUp)
	KeyCtrlSpace      = Key(tcell.KeyCtrlSpace)
	// KeySpace          = Key(tcell.Key(' '))
	KeySpace = ' '
	// KeyTilde = Key(tcell.Key('~'))
	KeyTilde = '~'

	// The following assignments were used in termbox implementation.
	// In tcell, these are not keys per se. But in gocui we have them
	// mapped to the keys so we have to use placeholder keys.
	MouseLeft         = Key(tcell.KeyF63) // arbitrary assignments
	MouseRight        = Key(tcell.KeyF62)
	MouseMiddle       = Key(tcell.KeyF61)
	MouseRelease      = Key(tcell.KeyF60)
	MouseWheelUp      = Key(tcell.KeyF59)
	MouseWheelDown    = Key(tcell.KeyF58)
	MouseWheelLeft    = Key(tcell.KeyF57)
	MouseWheelRight   = Key(tcell.KeyF56)
	KeyCtrl2          = Key(tcell.KeyNUL) // termbox defines theses
	KeyCtrl3          = Key(tcell.KeyEscape)
	KeyCtrl4          = Key(tcell.KeyCtrlBackslash)
	KeyCtrl5          = Key(tcell.KeyCtrlRightSq)
	KeyCtrl6          = Key(tcell.KeyCtrlCarat)
	KeyCtrl7          = Key(tcell.KeyCtrlUnderscore)
	KeyCtrlSlash      = Key(tcell.KeyCtrlUnderscore)
	KeyCtrlRsqBracket = Key(tcell.KeyCtrlRightSq)
	KeyCtrlBackslash  = Key(tcell.KeyCtrlBackslash)
	KeyCtrlLsqBracket = Key(tcell.KeyCtrlLeftSq)
)

var (
	lastMouseKey tcell.ButtonMask = tcell.ButtonNone
	lastMouseMod tcell.ModMask    = tcell.ModNone
)

// Modifiers.
const (
	ModAlt  = Modifier(tcell.ModAlt)
	ModNone = Modifier(0)
	// ModCtrl doesn't work with keyboard keys. Use CtrlKey in Key and ModNone. This is for mouse clicks only (tcell.v1)
	// ModCtrl = Modifier(tcell.ModCtrl)
)

func makeEvent(tev tcell.Event) Event {
	switch tev := tev.(type) {
	case *tcell.EventInterrupt:
		return Event{Type: EventInterrupt}
	case *tcell.EventResize:
		w, h := tev.Size()
		return Event{Type: EventResize, Width: w, Height: h}
	case *tcell.EventKey:
		k := tev.Key()
		ch := rune(0)
		if k == tcell.KeyRune {
			k = 0 // if rune remove key (so it can match, for example spacebar)
			ch = tev.Rune()
		}
		mod := tev.Modifiers()
		// remove control modifier and setup special handling of ctrl+spacebar, etc.
		if mod == tcell.ModCtrl && k == 0 && ch == 0 {
			mod = 0
			k = tcell.KeyCtrlSpace
		} else if mod == tcell.ModCtrl {
			mod = 0
		}
		return Event{
			Type: EventKey,
			Key:  Key(k),
			Ch:   ch,
			Mod:  Modifier(mod),
		}
	case *tcell.EventMouse:
		x, y := tev.Position()
		button := tev.Buttons()
		mouseKey := MouseRelease
		mouseMod := ModNone
		// process mouse wheel
		if button&tcell.WheelUp != 0 {
			mouseKey = MouseWheelUp
		}
		if button&tcell.WheelDown != 0 {
			mouseKey = MouseWheelDown
		}
		if button&tcell.WheelLeft != 0 {
			mouseKey = MouseWheelLeft
		}
		if button&tcell.WheelRight != 0 {
			mouseKey = MouseWheelRight
		}

		// process button events (not wheel events)
		button &= tcell.ButtonMask(0xff)
		if button != tcell.ButtonNone && lastMouseKey == tcell.ButtonNone {
			lastMouseKey = button
			lastMouseMod = tev.Modifiers()
		}

		switch tev.Buttons() {
		case tcell.ButtonNone:
			if lastMouseKey != tcell.ButtonNone {
				switch lastMouseKey {
				case tcell.ButtonPrimary:
					mouseKey = MouseLeft
				case tcell.ButtonSecondary:
					mouseKey = MouseRight
				case tcell.ButtonMiddle:
					mouseKey = MouseMiddle
				}
				mouseMod = Modifier(lastMouseMod)
				lastMouseMod = tcell.ModNone
				lastMouseKey = tcell.ButtonNone
			}
		}

		return Event{
			Type:   EventMouse,
			MouseX: x,
			MouseY: y,
			Key:    mouseKey,
			Ch:     0,
			Mod:    mouseMod,
		}
	default:
		return Event{Type: EventNone}
	}
}

// PollEvent blocks until an event is ready, and then returns it.
func PollEvent() Event {
	ev := screen.PollEvent()
	return makeEvent(ev)
}
