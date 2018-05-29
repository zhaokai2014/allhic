/**
 * Filename: /Users/htang/code/allhic/allhic/partition.go
 * Path: /Users/htang/code/allhic/allhic
 * Created Date: Wednesday, January 3rd 2018, 11:21:45 am
 * Author: htang
 *
 * Copyright (c) 2018 Haibao Tang
 */

package allhic

import (
	"fmt"
	"strconv"
	"strings"
)

// Partitioner converts the bamfile into a matrix of link counts
type Partitioner struct {
	Contigsfile   string
	Distfile      string
	K             int
	contigs       []ContigInfo
	contigToIdx   map[string]int
	matrix        [][]float64
	longestLength int
}

// Run is the main function body of partition
func (r *Partitioner) Run() {
	r.ParseContigLines()
	r.contigToIdx = make(map[string]int)
	for i, ci := range r.contigs {
		r.contigToIdx[ci.name] = i
	}
	dists := r.ParseDist()
	M := r.MakeMatrix(dists)
	clusters := Cluster(M, r.K)

	for _, ids := range clusters {
		names := make([]string, len(ids))
		for i, id := range ids {
			names[i] = r.contigs[id].name
		}
		fmt.Println(len(names), strings.Join(names, ","))
	}

	log.Notice("Success")
}

// MakeMatrix creates an adjacency matrix containing normalized score
func (r *Partitioner) MakeMatrix(edges []ContigPair) [][]float64 {
	M := Make2DSliceFloat64(len(r.contigs), len(r.contigs))
	longestSquared := float64(r.longestLength) * float64(r.longestLength)

	// Load up all the contig pairs
	for _, e := range edges {
		a, _ := r.contigToIdx[e.at]
		b, _ := r.contigToIdx[e.bt]
		if a == b {
			continue
		}

		w := float64(e.nObservedLinks) * longestSquared / (float64(e.L1) * float64(e.L2))
		// fmt.Printf("%s %s %d %.6f\n", e.at, e.bt, e.nObservedLinks, w)
		M[a][b] = w
		M[b][a] = w
	}

	return M
}

// ParseDist imports the edges of the contig linkage graph
func (r *Partitioner) ParseDist() []ContigPair {
	pairs := ParseDistLines(r.Distfile)
	goodPairs := FilterEdges(pairs)
	log.Noticef("Edge filtering keeps %s edges",
		Percentage(len(goodPairs), len(pairs)))
	return goodPairs
}

// FilterEdges implements rules to keep edges between close contigs and remove distant or weak contig pairs
func FilterEdges(edges []ContigPair) []ContigPair {
	var goodEdges []ContigPair

	for _, e := range edges {
		if e.mleDistance >= EffLinkDist {
			continue
		}
		goodEdges = append(goodEdges, e)
	}

	return goodEdges
}

// ParseContigLines imports the contig infor into a slice of ContigInfo
// ContigInfo stores the data struct of the contigfile
// #Contig Length  Expected        Observed        LDE
// jpcChr1.ctg249  25205   2.3     4       1.7391
// jpcChr1.ctg344  82275   15.4    17      1.1068
func (r *Partitioner) ParseContigLines() {
	recs := ReadCSVLines(r.Contigsfile)
	for _, rec := range recs {
		name := rec[0]
		length, _ := strconv.Atoi(rec[1])
		if length > r.longestLength {
			r.longestLength = length
		}
		nExpectedLinks, _ := strconv.ParseFloat(rec[2], 64)
		nObservedLinks, _ := strconv.Atoi(rec[3])
		lde, _ := strconv.ParseFloat(rec[4], 64)

		ci := ContigInfo{
			name: name, length: length,
			nExpectedLinks: nExpectedLinks, nObservedLinks: nObservedLinks,
			lde: lde,
		}
		r.contigs = append(r.contigs, ci)
	}
}

// ParseDistLines imports the edges of the contig into a slice of DistLine
// DistLine stores the data structure of the distfile
// #Contig1        Contig2 Length1 Length2 LDE1    LDE2    LDE     ObservedLinks   ExpectedLinksIfAdjacent MLEdistance
// jpcChr1.ctg199  jpcChr1.ctg257  124567  274565  0.3195  2.0838  1.1607  2       27.4    1617125
// idcChr1.ctg353  idcChr1.ctg382  143105  270892  2.1577  1.0544  1.3505  2       34.2    2190000
func ParseDistLines(distfile string) []ContigPair {
	var edges []ContigPair
	recs := ReadCSVLines(distfile)

	for _, rec := range recs {
		at, bt := rec[0], rec[1]
		L1, _ := strconv.Atoi(rec[2])
		L2, _ := strconv.Atoi(rec[3])
		lde1, _ := strconv.ParseFloat(rec[4], 64)
		lde2, _ := strconv.ParseFloat(rec[5], 64)
		localLDE, _ := strconv.ParseFloat(rec[6], 64)
		nObservedLinks, _ := strconv.Atoi(rec[7])
		nExpectedLinks, _ := strconv.ParseFloat(rec[8], 64)
		mleDistance, _ := strconv.Atoi(rec[9])
		score, _ := strconv.ParseFloat(rec[10], 64)

		cp := ContigPair{
			at: at, bt: bt,
			L1: L1, L2: L2,
			lde1: lde1, lde2: lde2, localLDE: localLDE,
			nObservedLinks: nObservedLinks, nExpectedLinks: nExpectedLinks,
			mleDistance: mleDistance, score: score,
		}

		edges = append(edges, cp)
	}

	return edges
}
