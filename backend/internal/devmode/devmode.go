// Package devmode holds the fixed accounts of a development build.
//
// Development mode skips the setup token for the admin console and OIDC
// for the player. It is only compiled in with -tags dev, so a
// release binary cannot be configured into it.
package devmode

const (
	// AdminUsername and AdminPassword sign in to the admin console.
	AdminUsername = "admin"
	AdminPassword = "admin"
	// Username is the listener every player request signs in as (the
	// counterpart of homepage's dev_user).
	Username = "listener"
)

// Groups are the identity provider groups of the development listener.
var Groups = []string{"sonora-import"}

// Available reports whether this binary was built with -tags dev.
const Available = available
