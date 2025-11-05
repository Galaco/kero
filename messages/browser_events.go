package messages

// BrowserQueryStartedEvent is dispatched when a server query begins
// This allows the UI to show loading indicators or disable buttons
type BrowserQueryStartedEvent struct {
	Address string // The server address being queried
}

// BrowserQueryCompletedEvent is dispatched when a server query succeeds
// The UI should call browser.GetLastResult() to retrieve the server information
type BrowserQueryCompletedEvent struct {
	Address string // The server address that was queried
}

// BrowserQueryFailedEvent is dispatched when a server query fails
// The UI should call browser.GetLastResult() to retrieve the error details
type BrowserQueryFailedEvent struct {
	Address string // The server address that failed
}
