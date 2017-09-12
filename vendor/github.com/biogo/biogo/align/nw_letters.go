// This file is automatically generated. Do not edit - make changes to relevant got file.

// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"github.com/biogo/biogo/alphabet"
	"github.com/biogo/biogo/feat"

	"fmt"
	"os"
	"text/tabwriter"
)

//line nw_type.got:17
func drawNWTableLetters(rSeq, qSeq alphabet.Letters, index alphabet.Index, table []int, a [][]int) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Printf("rSeq: %s\n", rSeq)
	fmt.Printf("qSeq: %s\n", qSeq)
	fmt.Fprint(tw, "\tqSeq\t")
	for _, l := range qSeq {
		fmt.Fprintf(tw, "%c\t", l)
	}
	fmt.Fprintln(tw)

	r, c := rSeq.Len()+1, qSeq.Len()+1
	fmt.Fprint(tw, "rSeq\t")
	for i := 0; i < r; i++ {
		if i != 0 {
			fmt.Fprintf(tw, "%c\t", rSeq[i-1])
		}

		for j := 0; j < c; j++ {
			p := pointerNWLetters(rSeq, qSeq, i, j, table, index, a, c)
			if p != "" {
				fmt.Fprintf(tw, "%s % 3v\t", p, table[i*c+j])
			} else {
				fmt.Fprintf(tw, "%v\t", table[i*c+j])
			}
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

func pointerNWLetters(rSeq, qSeq alphabet.Letters, i, j int, table []int, index alphabet.Index, a [][]int, c int) string {
	switch {
	case i == 0 && j == 0:
		return ""
	case i == 0:
		return "⬅"
	case j == 0:
		return "⬆"
	}
	rVal := index[rSeq[i-1]]
	qVal := index[qSeq[j-1]]
	if rVal < 0 || qVal < 0 {
		return ""
	}
	switch p := i*c + j; table[p] {
	case table[p-c-1] + a[rVal][qVal]:
		return "⬉"
	case table[p-c] + a[rVal][gap]:
		return "⬆"
	case table[p-1] + a[gap][qVal]:
		return "⬅"
	default:
		return ""
	}
}

func (a NW) alignLetters(rSeq, qSeq alphabet.Letters, alpha alphabet.Alphabet) ([]feat.Pair, error) {
	let := len(a)
	if let < alpha.Len() {
		return nil, ErrMatrixWrongSize{Size: let, Len: alpha.Len()}
	}
	la := make([]int, 0, let*let)
	for _, row := range a {
		if len(row) != let {
			return nil, ErrMatrixNotSquare
		}
		la = append(la, row...)
	}

	index := alpha.LetterIndex()
	r, c := rSeq.Len()+1, qSeq.Len()+1
	table := make([]int, r*c)
	for j := range table[1:c] {
		table[j+1] = table[j] + la[index[qSeq[j]]]
	}
	for i := 1; i < r; i++ {
		table[i*c] = table[(i-1)*c] + la[index[rSeq[i-1]]*let]
	}

	for i := 1; i < r; i++ {
		for j := 1; j < c; j++ {
			var (
				rVal = index[rSeq[i-1]]
				qVal = index[qSeq[j-1]]
			)
			if rVal < 0 {
				return nil, fmt.Errorf("align: illegal letter %q at position %d in rSeq", rSeq[i-1], i-1)
			}
			if qVal < 0 {
				return nil, fmt.Errorf("align: illegal letter %q at position %d in qSeq", qSeq[j-1], j-1)
			}
			p := i*c + j

			diagScore := table[p-c-1] + la[rVal*let+qVal]
			upScore := table[p-c] + la[rVal*let]
			leftScore := table[p-1] + la[qVal]

			table[p] = max3(diagScore, upScore, leftScore)
		}
	}
	if debugNeedle {
		drawNWTableLetters(rSeq, qSeq, index, table, a)
	}

	var aln []feat.Pair
	score, last := 0, diag
	i, j := r-1, c-1
	maxI, maxJ := i, j
	for i > 0 && j > 0 {
		var (
			rVal = index[rSeq[i-1]]
			qVal = index[qSeq[j-1]]
		)
		switch p := i*c + j; table[p] {
		case table[p-c-1] + la[rVal*let+qVal]:
			if last != diag {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p] - table[p-c-1]
			i--
			j--
			last = diag
		case table[p-c] + la[rVal*let]:
			if last != up && p != len(table)-1 {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p] - table[p-c]
			i--
			last = up
		case table[p-1] + la[qVal]:
			if last != left && p != len(table)-1 {
				aln = append(aln, &featPair{
					a:     feature{start: i, end: maxI},
					b:     feature{start: j, end: maxJ},
					score: score,
				})
				maxI, maxJ = i, j
				score = 0
			}
			score += table[p] - table[p-1]
			j--
			last = left
		default:
			panic(fmt.Sprintf("align: nw internal error: no path at row: %d col:%d\n", i, j))
		}
	}

	aln = append(aln, &featPair{
		a:     feature{start: i, end: maxI},
		b:     feature{start: j, end: maxJ},
		score: score,
	})
	if i != j {
		aln = append(aln, &featPair{
			a:     feature{start: 0, end: i},
			b:     feature{start: 0, end: j},
			score: table[i*c+j],
		})
	}

	for i, j := 0, len(aln)-1; i < j; i, j = i+1, j-1 {
		aln[i], aln[j] = aln[j], aln[i]
	}

	return aln, nil
}
