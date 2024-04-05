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

// mrca is the top-level function aside from main
func mrca(treeFile, re string) error {
	// Open the file containing the tree
	f, err := os.Open(treeFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Parse it into a gotree Tree (t)
	t, err := newick.NewParser(f).Parse()
	if err != nil {
		return err
	}

	// The input tree needs to be rooted for the routines below to work
	if !t.Rooted() {
		return errors.New("input tree is not rooted")
	}

	// tb is a tipbag which contains all the tips whose names match the
	// regular expression provided on the command line
	tb, err := relevantTips(t, re)
	if err != nil {
		return err
	}
	if len(tb) == 0 {
		return errors.New("no tips found matching regex")
	}

	// get the most recent common ancestor of the relevant tips
	mrca := getMRCA(t, tb)
	if len(mrca.Name()) > 0 {
		os.Stdout.WriteString(mrca.Name() + "\n")
	} else if mrca == t.Root() {
		os.Stdout.WriteString("root\n")
	} else {
		return errors.New("MRCA node has no name")
	}

	return nil
}

// Given a (gotree) tree and the string representation of a regular expression,
// return a tipbag containing all the tips that match the regexp.
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

// getMRCA does all the work given the tree and the bag of relevant tips.
// It populations a map of each interior node: its child tips (as long as
// they contain all of the tips we want in tb), and finally returns the node
// with the fewest number of child tips overall: this is the MRCA of the
// tipbag.
func getMRCA(t *tree.Tree, tb tipbag) *tree.Node {
	// an empty map to populate with candidate MRCAs
	childMap := make(map[*tree.Node][]*tree.Node)
	// starting at the root, traverse every node (recursively) and add it to
	// childMap if all its child tips are a superset of the tips in tb
	childTips(t.Root(), nil, tb, childMap)
	// populate a final map from node -> number of child tips
	finalMap := make(map[*tree.Node]int)
	for parent, children := range childMap {
		finalMap[parent] = len(children)
	}
	// the MRCA is the node with the fewest number of children overall
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

// (Recursively) check each node in the tree for whether its descendant nodes contain all
// the relevant tips, and add them to a map if they do
func childTips(cur, prev *tree.Node, tb tipbag, m map[*tree.Node][]*tree.Node) []*tree.Node {
	children := make([]*tree.Node, 0)
	for _, neighbour := range cur.Neigh() {
		if neighbour != prev {
			if neighbour.Tip() {
				children = append(children, neighbour)
			} else {
				ch := childTips(neighbour, cur, tb, m)
				children = append(children, ch...)
			}
		}
	}
	if isSubset(tb, children) {
		m[cur] = children
	}
	return children
}

// isSubset returns true/false the tips in the tipbag are a subset
// of the tips in others
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
