package signal

import (
	"os"
	"os/signal"
)

// interruptChannel is used to receive SIGINT (Ctrl+C) signals.
var interruptChannel chan os.Signal

// addHandlerChannel is used to add an interrupt handler to the list of handlers
// to be invoked on SIGINT (Ctrl+C) signals.
var addHandlerChannel = make(chan func())

// InterruptHandlersDone is closed after all interrupt handlers run the first
// time an interrupt is signaled.
var InterruptHandlersDone = make(chan struct{})

var SimulateInterruptChannel = make(chan struct{}, 1)

// signals defines the signals that are handled to do a clean shutdown.
// Conditional compilation is used to also include SIGTERM and KILL on Unix.
var signals = []os.Signal{os.Interrupt, os.Kill}

// SimulateInterrupt requests invoking the clean termination process by an
// internal component instead of a SIGINT.
func SimulateInterrupt() {
	select {
	case SimulateInterruptChannel <- struct{}{}:
	default:
	}
}

// mainInterruptHandler listens for SIGINT (Ctrl+C) signals on the
// interruptChannel and invokes the registered interruptCallbacks accordingly.
// It also listens for callback registration.  It must be run as a goroutine.
func mainInterruptHandler() {
	// interruptCallbacks is a list of callbacks to invoke when a
	// SIGINT (Ctrl+C) is received.
	var interruptCallbacks []func()
	invokeCallbacks := func() {
		// run handlers in LIFO order.
		for i := range interruptCallbacks {
			idx := len(interruptCallbacks) - 1 - i
			interruptCallbacks[idx]()
		}
		close(InterruptHandlersDone)
	}

	for {
		select {
		case <-interruptChannel:
			invokeCallbacks()
			return
		case <-SimulateInterruptChannel:
			invokeCallbacks()
			return

		case handler := <-addHandlerChannel:
			interruptCallbacks = append(interruptCallbacks, handler)
		}
	}
}

// AddInterruptHandler adds a handler to call when a SIGINT (Ctrl+C) is
// received.
func AddInterruptHandler(handler func()) {
	// Create the channel and start the main interrupt handler which invokes
	// all other callbacks and exits if not already done.
	if interruptChannel == nil {
		interruptChannel = make(chan os.Signal, 1)
		signal.Notify(interruptChannel, signals...)
		go mainInterruptHandler()
	}

	addHandlerChannel <- handler
}

// InterruptRequested returns true when the channel returned by
// interruptListener was closed.  This simplifies early shutdown slightly since
// the caller can just use an if statement instead of a select.
func InterruptRequested() bool {
	select {
	case <-InterruptHandlersDone:
		return true
	default:
	}

	return false
}
