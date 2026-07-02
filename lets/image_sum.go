package lets

type imageSum struct {
	Repo   string
	Tag    string
	ID     string
	Origin string `json:",omitempty"`
}

func newImageSum(repo, tag, id string) *imageSum {
	return &imageSum{
		Repo: repo,
		Tag:  tag,
		ID:   id,
	}
}

func imageSumOut(name string) string { return name + ".imgsum" }

func imageTarOut(name string) string { return name + ".tar.gz" }

func loadImageSum(f string) (*imageSum, error) {
	sum := new(imageSum)
	if err := readJSONFile(f, sum); err != nil {
		return nil, err
	}
	return sum, nil
}
