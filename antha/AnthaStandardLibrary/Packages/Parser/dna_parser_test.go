// antha/AnthaStandardLibrary/Packages/Parser/RebaseParser.go: Part of the Antha language
// Copyright (C) 2015 The Antha authors. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
//
// For more information relating to the software or licensing issues please
// contact license@antha-lang.org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 2 Royal College St, London NW1 0NH UK

package parser

import (
	//"fmt"

	//	"github.com/Synthace/antha/antha/AnthaStandardLibrary/Packages/sequences"
	//	"github.com/Synthace/antha/antha/anthalib/wtype"
	//"github.com/Synthace/antha/antha/AnthaStandardLibrary/Packages/text"
	//"io/ioutil"
	//	"bufio"
	//	"log"
	//	"os"
	//	"strconv"
	//	"strings"
	"testing"
)

type testpair struct {
	filename string
	lines    []string
	sequence string
}

var testpairs = []testpair{

	{filename: "GFP_dash.dna",
		lines:    []string{"|", "GCTCTTCTatgacggcattgacggaaggtgcaaaactgtttgagaaagagatcccgtatatcaccgaactggaaggcgacgtcgaaggtatgaaatttatcattaaaggcgagggtaccggtgacgcgaccacgggtaccattaaagcgaaatacatctgcactacgggcgacctgccggtcccgtgggcaaccctggtgagcaccctgagctacggtgttcagtgtttcgccaagtacccgagccacatcaaggatttctttaagagcgccatgccggaaggttatacccaagagcgtaccatcagcttcgaaggcgacggcgtgtacaagacgcgtgctatggttacctacgaacgcggttctatctacaatcgtgtcacgctgactggtgagaactttaagaaagacggtcacattctgcgtaagaacgttgcattccaatgcccgccaagcattctgtatattctgcctgacaccgttaacaatggcatccgcgttgagttcaaccaggcgtacgatattgaaggtgtgaccgaaaaactggttaccaaatgcagccaaatgaatcgtccgttggcgggctccgcggcagtgcatatcccgcgttatcatcacattacctaccacaccaaactgagcaaagaccgcgacgagcgccgtgatcacatgtgtctggtagaggtcgtgaaagcggttgatctggacacgtatcagtaactcgagtaaggatctccaggcatcaaataaaacgaaaggctcagtcgaaagactgggcctttcgttttatctgttgtttgtcggtgaacgctctctactagagtcacactggctcaccttcgggtggGCCTTTCTGCGTTTtactagagtcacactgatgaGGTAGAAGAGC", "˘"},
		sequence: "GCTCTTCTatgacggcattgacggaaggtgcaaaactgtttgagaaagagatcccgtatatcaccgaactggaaggcgacgtcgaaggtatgaaatttatcattaaaggcgagggtaccggtgacgcgaccacgggtaccattaaagcgaaatacatctgcactacgggcgacctgccggtcccgtgggcaaccctggtgagcaccctgagctacggtgttcagtgtttcgccaagtacccgagccacatcaaggatttctttaagagcgccatgccggaaggttatacccaagagcgtaccatcagcttcgaaggcgacggcgtgtacaagacgcgtgctatggttacctacgaacgcggttctatctacaatcgtgtcacgctgactggtgagaactttaagaaagacggtcacattctgcgtaagaacgttgcattccaatgcccgccaagcattctgtatattctgcctgacaccgttaacaatggcatccgcgttgagttcaaccaggcgtacgatattgaaggtgtgaccgaaaaactggttaccaaatgcagccaaatgaatcgtccgttggcgggctccgcggcagtgcatatcccgcgttatcatcacattacctaccacaccaaactgagcaaagaccgcgacgagcgccgtgatcacatgtgtctggtagaggtcgtgaaagcggttgatctggacacgtatcagtaactcgagtaaggatctccaggcatcaaataaaacgaaaggctcagtcgaaagactgggcctttcgttttatctgttgtttgtcggtgaacgctctctactagagtcacactggctcaccttcgggtggGCCTTTCTGCGTTTtactagagtcacactgatgaGGTAGAAGAGC"},
}

/*
func TestSnapGenetoSimpleSeq(t *testing.T) {
	for _, seq := range testpairs {
		result, err := SnapGenetoSimpleSeq(seq.filename)
		if result != seq.sequence {
			t.Error(
				"For", seq.filename, "/n",
				"expected", seq.sequence, "\n",
				"got", result, "\n",
			)
		}
			if err != nil {
			t.Error(
				"Error = ", err.Error(), "/n",
			)
		}
	}
}
*/
func TestHandlesnapgenelines(t *testing.T) {
	for _, seq := range testpairs {
		result := Handlesnapgenelines(seq.lines)
		if result != seq.sequence {
			t.Error(
				"For", seq.filename, "/n",
				"expected", seq.sequence, "\n",
				"got", result, "\n",
			)
		}

	}
}
