// Package assets embeds Recuerdo's icon.
package assets

import _ "embed"

// Icon is Recuerdo's icon as a 512×512 PNG.
//
//go:embed recuerdo.png
var Icon []byte
