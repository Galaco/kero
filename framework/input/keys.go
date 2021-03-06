package input

import (
	"github.com/galaco/tinygametools"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Key tinygametools.Key

type KeyAction glfw.Action
type ModifierKey glfw.ModifierKey

const (
	KeyPress   = KeyAction(glfw.Press)
	KeyRepeat  = KeyAction(glfw.Repeat)
	KeyRelease = KeyAction(glfw.Release)
)

const (
	KeyUnknown      = Key(tinygametools.KeyUnknown)
	KeySpace        = Key(tinygametools.KeySpace)
	KeyApostrophe   = Key(tinygametools.KeyApostrophe)
	KeyComma        = Key(tinygametools.KeyComma)
	KeyMinus        = Key(tinygametools.KeyMinus)
	KeyPeriod       = Key(tinygametools.KeyPeriod)
	KeySlash        = Key(tinygametools.KeySlash)
	Key0            = Key(tinygametools.Key0)
	Key1            = Key(tinygametools.Key1)
	Key2            = Key(tinygametools.Key2)
	Key3            = Key(tinygametools.Key3)
	Key4            = Key(tinygametools.Key4)
	Key5            = Key(tinygametools.Key5)
	Key6            = Key(tinygametools.Key6)
	Key7            = Key(tinygametools.Key7)
	Key8            = Key(tinygametools.Key8)
	Key9            = Key(tinygametools.Key9)
	KeySemicolon    = Key(tinygametools.KeySemicolon)
	KeyEqual        = Key(tinygametools.KeyEqual)
	KeyA            = Key(tinygametools.KeyA)
	KeyB            = Key(tinygametools.KeyB)
	KeyC            = Key(tinygametools.KeyC)
	KeyD            = Key(tinygametools.KeyD)
	KeyE            = Key(tinygametools.KeyE)
	KeyF            = Key(tinygametools.KeyF)
	KeyG            = Key(tinygametools.KeyG)
	KeyH            = Key(tinygametools.KeyH)
	KeyI            = Key(tinygametools.KeyI)
	KeyJ            = Key(tinygametools.KeyJ)
	KeyK            = Key(tinygametools.KeyK)
	KeyL            = Key(tinygametools.KeyL)
	KeyM            = Key(tinygametools.KeyM)
	KeyN            = Key(tinygametools.KeyN)
	KeyO            = Key(tinygametools.KeyO)
	KeyP            = Key(tinygametools.KeyP)
	KeyQ            = Key(tinygametools.KeyQ)
	KeyR            = Key(tinygametools.KeyR)
	KeyS            = Key(tinygametools.KeyS)
	KeyT            = Key(tinygametools.KeyT)
	KeyU            = Key(tinygametools.KeyU)
	KeyV            = Key(tinygametools.KeyV)
	KeyW            = Key(tinygametools.KeyW)
	KeyX            = Key(tinygametools.KeyX)
	KeyY            = Key(tinygametools.KeyY)
	KeyZ            = Key(tinygametools.KeyZ)
	KeyLeftBracket  = Key(tinygametools.KeyLeftBracket)
	KeyBackslash    = Key(tinygametools.KeyBackslash)
	KeyRightBracket = Key(tinygametools.KeyRightBracket)
	KeyGraveAccent  = Key(tinygametools.KeyGraveAccent)
	KeyWorld1       = Key(tinygametools.KeyWorld1)
	KeyWorld2       = Key(tinygametools.KeyWorld2)
	KeyEscape       = Key(tinygametools.KeyEscape)
	KeyEnter        = Key(tinygametools.KeyEnter)
	KeyTab          = Key(tinygametools.KeyTab)
	KeyBackspace    = Key(tinygametools.KeyBackspace)
	KeyInsert       = Key(tinygametools.KeyInsert)
	KeyDelete       = Key(tinygametools.KeyDelete)
	KeyRight        = Key(tinygametools.KeyRight)
	KeyLeft         = Key(tinygametools.KeyLeft)
	KeyDown         = Key(tinygametools.KeyDown)
	KeyUp           = Key(tinygametools.KeyUp)
	KeyPageUp       = Key(tinygametools.KeyPageUp)
	KeyPageDown     = Key(tinygametools.KeyPageDown)
	KeyHome         = Key(tinygametools.KeyHome)
	KeyEnd          = Key(tinygametools.KeyEnd)
	KeyCapsLock     = Key(tinygametools.KeyCapsLock)
	KeyScrollLock   = Key(tinygametools.KeyScrollLock)
	KeyNumLock      = Key(tinygametools.KeyNumLock)
	KeyPrintScreen  = Key(tinygametools.KeyPrintScreen)
	KeyPause        = Key(tinygametools.KeyPause)
	KeyF1           = Key(tinygametools.KeyF1)
	KeyF2           = Key(tinygametools.KeyF2)
	KeyF3           = Key(tinygametools.KeyF3)
	KeyF4           = Key(tinygametools.KeyF4)
	KeyF5           = Key(tinygametools.KeyF5)
	KeyF6           = Key(tinygametools.KeyF6)
	KeyF7           = Key(tinygametools.KeyF7)
	KeyF8           = Key(tinygametools.KeyF8)
	KeyF9           = Key(tinygametools.KeyF9)
	KeyF10          = Key(tinygametools.KeyF10)
	KeyF11          = Key(tinygametools.KeyF11)
	KeyF12          = Key(tinygametools.KeyF12)
	KeyF13          = Key(tinygametools.KeyF13)
	KeyF14          = Key(tinygametools.KeyF14)
	KeyF15          = Key(tinygametools.KeyF15)
	KeyF16          = Key(tinygametools.KeyF16)
	KeyF17          = Key(tinygametools.KeyF17)
	KeyF18          = Key(tinygametools.KeyF18)
	KeyF19          = Key(tinygametools.KeyF19)
	KeyF20          = Key(tinygametools.KeyF20)
	KeyF21          = Key(tinygametools.KeyF21)
	KeyF22          = Key(tinygametools.KeyF22)
	KeyF23          = Key(tinygametools.KeyF23)
	KeyF24          = Key(tinygametools.KeyF24)
	KeyF25          = Key(tinygametools.KeyF25)
	KeyKP0          = Key(tinygametools.KeyKP0)
	KeyKP1          = Key(tinygametools.KeyKP1)
	KeyKP2          = Key(tinygametools.KeyKP2)
	KeyKP3          = Key(tinygametools.KeyKP3)
	KeyKP4          = Key(tinygametools.KeyKP4)
	KeyKP5          = Key(tinygametools.KeyKP5)
	KeyKP6          = Key(tinygametools.KeyKP6)
	KeyKP7          = Key(tinygametools.KeyKP7)
	KeyKP8          = Key(tinygametools.KeyKP8)
	KeyKP9          = Key(tinygametools.KeyKP9)
	KeyKPDecimal    = Key(tinygametools.KeyKPDecimal)
	KeyKPDivide     = Key(tinygametools.KeyKPDivide)
	KeyKPMultiply   = Key(tinygametools.KeyKPMultiply)
	KeyKPSubtract   = Key(tinygametools.KeyKPSubtract)
	KeyKPAdd        = Key(tinygametools.KeyKPAdd)
	KeyKPEnter      = Key(tinygametools.KeyKPEnter)
	KeyKPEqual      = Key(tinygametools.KeyKPEqual)
	KeyLeftShift    = Key(tinygametools.KeyLeftShift)
	KeyLeftControl  = Key(tinygametools.KeyLeftCtrl)
	KeyLeftAlt      = Key(tinygametools.KeyLeftAlt)
	KeyLeftSuper    = Key(tinygametools.KeyLeftSuper)
	KeyRightShift   = Key(tinygametools.KeyRightShift)
	KeyRightControl = Key(tinygametools.KeyRightCtrl)
	KeyRightAlt     = Key(tinygametools.KeyRightAlt)
	KeyRightSuper   = Key(tinygametools.KeyRightSuper)
	KeyMenu         = Key(tinygametools.KeyMenu)
	KeyLast         = Key(tinygametools.KeyLast)
)
