// antha/AnthaStandardLibrary/Packages/enzymes/Find.go: Part of the Antha language
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

// Utility package providing functions useful for searches
package search

import (
	"strings"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
)

// Returns the positions of any matching instances of a sequence in a slice of sequences.
// If checkSeqs is set to false, only the name will be checked;
// if checkSeqs is set to true, matching sequences with different names will also be checked.
func ContainsSeq(seqs []wtype.DNASequence, seq wtype.DNASequence, checkSeqs bool) (bool, []int) {

	var positionsFound []int

	for i := range seqs {
		if !checkSeqs {
			if seqs[i].Name() == seq.Name() {
				positionsFound = append(positionsFound, i)
			}
		} else {
			if seqs[i].Name() == seq.Name() {
				positionsFound = append(positionsFound, i)
			} else if strings.ToUpper(seqs[i].Sequence()) == strings.ToUpper(seq.Sequence()) && seqs[i].Plasmid == seq.Plasmid {
				positionsFound = append(positionsFound, i)
			}
		}
	}

	if len(positionsFound) > 0 {
		return true, positionsFound
	}

	return false, positionsFound
}
