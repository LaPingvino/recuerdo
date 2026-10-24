package media

import "github.com/LaPingvino/recuerdo/internal/lesson/mediakind"

// The kinds of media and their helpers, shared with the web page
// (internal/lesson/mediakind).
const (
	Image       = mediakind.Image
	Text        = mediakind.Text
	Audio       = mediakind.Audio
	Video       = mediakind.Video
	YouTube     = mediakind.YouTube
	Vimeo       = mediakind.Vimeo
	Dailymotion = mediakind.Dailymotion
	Website     = mediakind.Website
	Unknown     = mediakind.Unknown
)

var (
	Kind         = mediakind.Kind
	Shown        = mediakind.Shown
	AddressOK    = mediakind.AddressOK
	ResourceName = mediakind.ResourceName
	NextID       = mediakind.NextID
	Teachable    = mediakind.Teachable
)
