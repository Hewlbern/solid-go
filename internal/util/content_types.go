package util

// ContentTypes provides constants for common content types
type ContentTypes struct{}

// NewContentTypes creates a new ContentTypes
func NewContentTypes() *ContentTypes {
	return &ContentTypes{}
}

// Common content type constants
const (
	// Text content types
	TextPlain      = "text/plain"
	TextHTML       = "text/html"
	TextCSS        = "text/css"
	TextJavaScript = "text/javascript"
	TextXML        = "text/xml"
	TextCSV        = "text/csv"
	TextCalendar   = "text/calendar"
	TextMarkdown   = "text/markdown"
	TextYAML       = "text/yaml"
	TextJSON       = "text/json"

	// Application content types
	ApplicationJSON        = "application/json"
	ApplicationXML         = "application/xml"
	ApplicationYAML        = "application/yaml"
	ApplicationJavaScript  = "application/javascript"
	ApplicationPDF         = "application/pdf"
	ApplicationZip         = "application/zip"
	ApplicationOctetStream = "application/octet-stream"
	ApplicationForm        = "application/x-www-form-urlencoded"
	ApplicationMultipart   = "multipart/form-data"

	// Image content types
	ImageJPEG = "image/jpeg"
	ImagePNG  = "image/png"
	ImageGIF  = "image/gif"
	ImageSVG  = "image/svg+xml"
	ImageWebP = "image/webp"

	// Audio content types
	AudioMPEG = "audio/mpeg"
	AudioWAV  = "audio/wav"
	AudioOGG  = "audio/ogg"

	// Video content types
	VideoMP4  = "video/mp4"
	VideoWebM = "video/webm"
	VideoOGG  = "video/ogg"

	// Font content types
	FontTTF   = "font/ttf"
	FontOTF   = "font/otf"
	FontWOFF  = "font/woff"
	FontWOFF2 = "font/woff2"

	// RDF content types
	RDFXML   = "application/rdf+xml"
	NTriples = "application/n-triples"
	NQuads   = "application/n-quads"
	Turtle   = "text/turtle"
	TriG     = "application/trig"
	JSONLD   = "application/ld+json"
)

// IsText checks if a content type is a text type
func (c *ContentTypes) IsText(contentType string) bool {
	return contentType == TextPlain ||
		contentType == TextHTML ||
		contentType == TextCSS ||
		contentType == TextJavaScript ||
		contentType == TextXML ||
		contentType == TextCSV ||
		contentType == TextCalendar ||
		contentType == TextMarkdown ||
		contentType == TextYAML ||
		contentType == TextJSON
}

// IsApplication checks if a content type is an application type
func (c *ContentTypes) IsApplication(contentType string) bool {
	return contentType == ApplicationJSON ||
		contentType == ApplicationXML ||
		contentType == ApplicationYAML ||
		contentType == ApplicationJavaScript ||
		contentType == ApplicationPDF ||
		contentType == ApplicationZip ||
		contentType == ApplicationOctetStream ||
		contentType == ApplicationForm ||
		contentType == ApplicationMultipart
}

// IsImage checks if a content type is an image type
func (c *ContentTypes) IsImage(contentType string) bool {
	return contentType == ImageJPEG ||
		contentType == ImagePNG ||
		contentType == ImageGIF ||
		contentType == ImageSVG ||
		contentType == ImageWebP
}

// IsAudio checks if a content type is an audio type
func (c *ContentTypes) IsAudio(contentType string) bool {
	return contentType == AudioMPEG ||
		contentType == AudioWAV ||
		contentType == AudioOGG
}

// IsVideo checks if a content type is a video type
func (c *ContentTypes) IsVideo(contentType string) bool {
	return contentType == VideoMP4 ||
		contentType == VideoWebM ||
		contentType == VideoOGG
}

// IsFont checks if a content type is a font type
func (c *ContentTypes) IsFont(contentType string) bool {
	return contentType == FontTTF ||
		contentType == FontOTF ||
		contentType == FontWOFF ||
		contentType == FontWOFF2
}

// IsRDF checks if a content type is an RDF type
func (c *ContentTypes) IsRDF(contentType string) bool {
	return contentType == RDFXML ||
		contentType == NTriples ||
		contentType == NQuads ||
		contentType == Turtle ||
		contentType == TriG ||
		contentType == JSONLD
}
