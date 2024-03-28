package main

import (
	"errors"
	"math"
	"os"
	"regexp"

	"github.com/evolbioinfo/gotree/io/newick"
	"github.com/evolbioinfo/gotree/tree"
	"github.com/spf13/cobra"
)

type tipbag map[*tree.Node]bool

var flagsTree string
var flagsRegex string

func init() {
	mainCmd.Flags().StringVarP(&flagsTree, "tree", "t", "", "tree file to read (in Newick format)")
	mainCmd.Flags().StringVarP(&flagsRegex, "regex", "r", "", "regex of tip names to parse")
}

var mainCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := mrca(flagsTree, flagsRegex)
		return err
	},
}

func main() {
	err := mainCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func mrca(treeFile, re string) error {
	var t *tree.Tree
	f, err := os.Open(treeFile)
	if err != nil {
		return err
	}
	defer f.Close()

	t, err = newick.NewParser(f).Parse()
	if err != nil {
		return err
	}

	if !t.Rooted() {
		return errors.New("input tree is not rooted")
	}

	tb, err := relevantTips(t, re)
	if err != nil {
		return err
	}
	if len(tb) == 0 {
		return errors.New("no tips found matching regex")
	}

	mrca := MRCA(t, tb)
	if len(mrca.Name()) > 0 {
		os.Stdout.WriteString(mrca.Name() + "\n")
	} else {
		return errors.New("MRCA node has no name")
	}

	return nil
}

func relevantTips(t *tree.Tree, re string) (tipbag, error) {
	tb := make(map[*tree.Node]bool)
	for _, t := range t.Tips() {
		m, err := regexp.MatchString(re, t.Name())
		if err != nil {
			return tb, err
		}
		if m {
			tb[t] = true
		}
	}
	return tb, nil
}

func MRCA(t *tree.Tree, tb tipbag) *tree.Node {
	childMap := make(map[*tree.Node][]*tree.Node)
	childTips(t.Root(), nil, childMap)
	finalMap := make(map[*tree.Node]int)
	for parent, children := range childMap {
		isSubset := isSubset(tb, children)
		if isSubset {
			finalMap[parent] = len(children)
		}
	}
	min := math.MaxInt
	var mrca *tree.Node
	for parent, nchildren := range finalMap {
		if nchildren < min {
			mrca = parent
			min = nchildren
		}
	}
	return mrca
}

func childTips(cur, prev *tree.Node, m map[*tree.Node][]*tree.Node) []*tree.Node {
	children := make([]*tree.Node, 0)
	for _, neighbour := range cur.Neigh() {
		if neighbour != prev {
			if neighbour.Tip() {
				children = append(children, neighbour)
			} else {
				ch := childTips(neighbour, cur, m)
				children = append(children, ch...)
			}
		}
	}
	m[cur] = children
	return children
}

func isSubset(relevant tipbag, others []*tree.Node) bool {
	o := make(map[*tree.Node]bool)
	for _, node := range others {
		o[node] = true
	}
	for node, _ := range relevant {
		_, ok := o[node]
		if !ok {
			return false
		}
	}
	return true
}
