// Package devmode holds the fixed accounts of a development build.
//
// Development mode skips the setup token for the admin console and, from
// phase 4, OIDC for the player. It is only compiled in with -tags dev, so a
// release binary cannot be configured into it.
package devmode

const (
	// AdminUsername and AdminPassword sign in to the admin console.
	AdminUsername = "admin"
	AdminPassword = "admin"
	// Username is the ordinary player account, used once the player
	// signs in (the counterpart of homepage's dev_user).
	Username = "listener"
)

// Available reports whether this binary was built with -tags dev.
const Available = available
