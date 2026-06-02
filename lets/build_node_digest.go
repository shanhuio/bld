package lets

import (
	"fmt"
)

func buildNodeDigest(
	env *env, n *buildNode, deps map[string]string,
) (string, error) {
	switch n.typ {
	case nodeRule:
		action := &buildAction{
			Deps:     deps,
			RuleType: n.ruleType,
		}
		if meta := n.ruleMeta; meta != nil {
			if meta.digest == "" {
				return "", nil
			}
			action.Rule = meta.digest
			action.Outs = meta.outs
			action.DockerOut = meta.dockerOut
		}
		d, err := makeDigest("build_action", "", action)
		if err != nil {
			return "", fmt.Errorf("digest build action: %w", err)
		}
		return d, nil
	case nodeSrc:
		stat, err := newSrcFileStat(env, n.name)
		if err != nil {
			return "", fmt.Errorf("stat file %q: %w", n.name, err)
		}
		d, err := makeDigest("src", "", stat)
		if err != nil {
			return "", fmt.Errorf("digest source file: %w", err)
		}
		return d, nil
	case nodeOut:
		action := &buildAction{
			Deps:     deps,
			OutputOf: n.name,
		}
		d, err := makeDigest("out", "", action)
		if err != nil {
			return "", fmt.Errorf("digest output-of: %w", err)
		}
		return d, nil
	default:
		return "", nil
	}
}
