package asset

import _ "embed"

//go:embed favicon.png
var Favicon []byte

const FaviconMimeType = "image/png"
