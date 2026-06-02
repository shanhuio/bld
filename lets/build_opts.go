package lets

import (
	"io"
)

type dockerOpts struct {
	useBuildCache bool
}

type buildOpts struct {
	log    io.Writer
	docker *dockerOpts

	alwaysRebuild bool
}
