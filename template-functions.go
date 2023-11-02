package files

import "html/template"

func imageTPLHelper(image *ImageModel, style, class, width, attrs string) template.HTML {
	html := ""

	url := image.GetUrl(style)

	if url != "" {
		html += `<img`

		if image.Description != nil {
			html += ` alt="` + *image.Description + `"`
		}

		html += ` src="` + url + `"`

		if class != "" {
			html += ` class="` + class + `"`
		}

		if width != "" {
			html += ` width="` + width + `"`
		}

		if attrs != "" {
			html += ` ` + attrs + `"`
		}

		html += `>`
	}

	return template.HTML(html)
}

func imagesTPLHelper(images []*ImageModel, style, class, width, attrs string) template.HTML {

	if len(images) == 0 {
		return template.HTML("")
	}

	return imageTPLHelper(images[0], style, class, width, attrs)
}
