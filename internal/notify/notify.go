package notify

import "github.com/gen2brain/beeep"

func Send(title, message string) {
	_ = beeep.Notify(title, message, "")
}
