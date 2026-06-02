package lets

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type download struct {
	name   string
	url    *url.URL
	rule   *Download
	sha256 string
	out    string
}

func newDownload(_ *env, p string, r *Download) (*download, error) {
	name := makeRelPath(p, r.Name)

	const sha256Prefix = "sha256:"
	if !strings.HasPrefix(r.Checksum, sha256Prefix) {
		return nil, errors.New("checksum is not sha256")
	}

	u, err := url.Parse(r.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	if r.Output == "" {
		return nil, errors.New("output not specified")
	}

	return &download{
		name:   name,
		url:    u,
		rule:   r,
		sha256: strings.TrimPrefix(r.Checksum, sha256Prefix),
		out:    makeRelPath(p, r.Output),
	}, nil
}

func (d *download) meta(env *env) (*buildRuleMeta, error) {
	dat := struct {
		Sha256 string
		Out    string
	}{
		Sha256: d.sha256,
		Out:    d.out,
	}
	digest, err := makeDigest(ruleDownload, d.name, &dat)
	if err != nil {
		return nil, fmt.Errorf("digest: %w", err)
	}

	return &buildRuleMeta{
		name:   d.name,
		outs:   []string{d.out},
		digest: digest,
	}, nil
}

func downloadToFile(f string, r io.Reader) (string, error) {
	out, err := os.Create(f)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	defer out.Close()

	h := sha256.New()
	mw := io.MultiWriter(h, out)

	if _, err := io.Copy(mw, r); err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	if err := out.Sync(); err != nil {
		return "", fmt.Errorf("filesystem sync: %w", err)
	}

	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:]), nil
}

func (d *download) build(env *env, opts *buildOpts) error {
	out, err := env.prepareOut(d.out)
	if err != nil {
		return fmt.Errorf("prepare out: %w", err)
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    d.url,
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	sum, err := downloadToFile(out, resp.Body)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	if sum != d.sha256 {
		return fmt.Errorf(
			"incorrect sha256, want %s, got %s",
			d.sha256, sum,
		)
	}

	return nil
}
