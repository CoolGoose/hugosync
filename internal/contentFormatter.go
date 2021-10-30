package internal

import (
	"github.com/Machiel/slugify"
	"github.com/matryer/anno"
	"strings"
)

func PostTitle(source string) string {
	var title = ""

	if len(source) > 25 {
		postTitleRaw := strings.Fields(source)

		for i := 0; i < len(postTitleRaw); i++ {
			if i == 10 {
				break
			}
			title += postTitleRaw[i] + " "
		}
	} else {
		title = source
	}

	title = strings.ReplaceAll(title, "\"", "")
	title = strings.Trim(title, " ")

	return title
}

func PostDescription(source string) string {
	var description = ""

	if len(source) > 150 {
		description = source[:150]
	} else {
		description = source
	}

	description = strings.ReplaceAll(description, "\"", "")
	description = strings.Trim(description, "")

	return description
}

func PostSlug(source string) string {
	var slug = ""

	slug = slugify.Slugify(source)

	return slug
}

func PostBody(rawBody string) (string, []string, error) {
	notes, err := anno.FindManyString(rawBody, anno.Hashtags)

	if err != nil {
		return "", []string{}, err
	}

	var tags []string
	body := rawBody
	for _, note := range notes {
		// log.Printf("Found a %s at [%d:%d]: \"%s\"", note.Kind, note.Start, note.End(), note.Val)
		tags = append(tags, string(note.Val[1:]))
		body = strings.ReplaceAll(body, string(note.Val), "")
	}

	return body, tags, err
}
